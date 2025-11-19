package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"bck/auth/authpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func AuthMiddleware(jwtSecret string, _ any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("Authorization")
			if err != nil {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			token := cookie.Value

			// grpc connection with auth-service
			conn, err := grpc.NewClient("localhost:42070", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("failed to connect to auth service: %v", err)
				http.Error(w, "authentication service unavailable", http.StatusServiceUnavailable)
				return
			}
			defer conn.Close()

			client := authpb.NewAuthServiceClient(conn)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
				Token: token,
			})
			if err != nil {
				log.Printf("token validation failed: %v", err)
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			if !resp.Valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

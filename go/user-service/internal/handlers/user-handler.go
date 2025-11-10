package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"bck/auth/authpb"
	"bck/user/userpb"
	"user-service/internal/database"
	"user-service/internal/kafka"
	"user-service/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	userRepo     database.UserRepository
	userProducer *kafka.UserProducer
	authClient   authpb.AuthServiceClient
}

func NewUserHandler(userRepo database.UserRepository, authClient authpb.AuthServiceClient, userProducer *kafka.UserProducer) *UserHandler {
	return &UserHandler{
		userRepo:   userRepo,
		authClient: authClient,
		userProducer: kafka.NewUserProducer(
			[]string{"localhost:9092"},
			"user-events",
		),
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invaild json", http.StatusBadRequest)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "couldn't create hash", http.StatusInternalServerError)
		return
	}
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
	}
	err = h.userRepo.CreateUser(r.Context(), user)
	if err != nil {
		log.Printf("CreateUser error: %v", err)
		http.Error(w, "couldn't create user", http.StatusInternalServerError)
		return
	}

	event := models.UserCreatedEvent{
		ID:    user.ID.Hex(),
		Email: user.Email,
		Name:  user.Name,
		Time:  time.Now(),
	}

	if err := h.userProducer.PublishUserCreated(r.Context(), event); err != nil {
		log.Printf("failed to publish user created event: %v", err)
	}
	log.Printf("new user created: %s", user.ID.Hex())
	w.Header().Set("Cotent-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("user created!"))
}
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		log.Println("error repo: ", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	authResp, err := h.authClient.Authenticate(context.Background(), &authpb.AuthRequest{
		UserId:         user.ID.Hex(),
		Email:          req.Email,
		Password:       req.Password,
		HashedPassword: user.Password,
	})
	if err != nil || !authResp.Valid {
		log.Println("error authclient grpc: ", err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	token := authResp.Token

	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // using localhost btw atleast for now
		SameSite: http.SameSiteLaxMode,
		MaxAge:   36000,
	})
	log.Printf("user authenticated with id: %s", authResp.UserId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user logged in!"))
}
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Overwriting to expire the token
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // expires immediately
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("logged out"))
}

func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("Authorization")
	if cookie == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{
		Token: cookie.Value,
	})
	if err != nil || !resp.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	user, _ := h.userRepo.GetUserById(r.Context(), resp.UserId)
	json.NewEncoder(w).Encode(map[string]string{
		"id":    user.ID.Hex(),
		"email": user.Email,
	})
}

// grpc handlers
func (h *UserHandler) VerifyCredentials(ctx context.Context, req *userpb.VerifyCredentialsRequest) (*userpb.VerifyCredentialsResponse, error) {
	user, err := h.userRepo.GetUserByEmail(ctx, req.Email)

	// just debugging
	log.Printf("lookup email=%s  user=%+v  err=%v", req.Email, user, err)

	if err != nil || user == nil {
		return &userpb.VerifyCredentialsResponse{
			Valid: false,
			// Token:  "",
			UserId: "",
		}, nil
	}

	authReq := &authpb.AuthRequest{
		Email:          req.Email,
		Password:       req.Password,
		HashedPassword: user.Password,
		UserId:         user.ID.Hex(),
	}

	authResp, err := h.authClient.Authenticate(ctx, authReq)
	if err != nil {
		log.Printf("error authenticating via authservice: %v", err)
		return &userpb.VerifyCredentialsResponse{
			Valid: false,
			// Token:  "",
			UserId: "",
		}, nil
	}

	return &userpb.VerifyCredentialsResponse{
		Valid: authResp.Valid,
		// Token:  authResp.Token,
		UserId: user.ID.Hex(),
	}, nil
}

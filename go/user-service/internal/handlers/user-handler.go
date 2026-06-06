package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"grpc_module/auth/authpb"
	"grpc_module/user/userpb"
	"user-service/internal/database"
	"user-service/internal/kafka"
	"user-service/internal/models"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	userRepo     database.UserRepository
	logger       *zap.Logger
	userProducer *kafka.UserProducer
	authClient   authpb.AuthServiceClient
}

func NewUserHandler(userRepo database.UserRepository, logger *zap.Logger, authClient authpb.AuthServiceClient, userProducer *kafka.UserProducer) *UserHandler {
	return &UserHandler{
		userRepo:   userRepo,
		logger:     logger,
		authClient: authClient,
		userProducer: kafka.NewUserProducer(
			[]string{"localhost:9092"},
			"user-events",
		),
	}
}

func (h *UserHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/profile", h.Profile)
	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/logout", h.Logout)
	return mux
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Warn("method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invaild json", http.StatusBadRequest)
		h.logger.Error("err in decoding json", zap.Error(err))
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		h.logger.Error("missing required fields")
		return
	}

	authResp, err := h.authClient.GeneratePassword(r.Context(), &authpb.BcryptPasswordRequest{Password: req.Password})
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		h.logger.Error("auth-service grpc err", zap.Error(err), zap.String("email", req.Email))
		return
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: authResp.HashedPassword,
	}
	err = h.userRepo.CreateUser(r.Context(), user)
	if err != nil {
		http.Error(w, "couldn't create user", http.StatusInternalServerError)
		h.logger.Error("err creating user", zap.Error(err), zap.String("email", user.Email))
		return
	}

	event := models.UserCreatedEvent{
		ID:    user.ID.Hex(),
		Email: user.Email,
		Name:  user.Name,
		Time:  time.Now(),
	}

	if err := h.userProducer.PublishUserCreated(r.Context(), event); err != nil {
		h.logger.Error("failed to publish user-created event", zap.Error(err))
	}

	h.logger.Info("User created", zap.String("ID", user.ID.Hex()), zap.String("Email", user.Email))

	w.Header().Set("Cotent-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("user created!"))
}
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.logger.Warn("method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "interval server error", http.StatusBadRequest)
		h.logger.Error("err in decoding json", zap.Error(err))
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		h.logger.Error("missing required fields")
	}

	user, err := h.userRepo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		h.logger.Error("db error while fetching user",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		h.logger.Warn("login attempt with non existent email",
			zap.String("email", req.Email),
		)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	authResp, err := h.authClient.Authenticate(context.Background(), &authpb.AuthRequest{
		UserId:         user.ID.Hex(),
		Email:          req.Email,
		Password:       req.Password,
		HashedPassword: user.Password,
	})
	if err != nil {
		h.logger.Error("auth service grpc failure",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		http.Error(w, "invalid server error", http.StatusUnauthorized)
		return
	}
	if !authResp.Valid {
		h.logger.Warn("invalid login credentials",
			zap.String("email", req.Email),
		)
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
		h.logger.Warn("method not allowed")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := r.Cookie("Authorization")
	if err != nil {
		h.logger.Warn("missing auth cookie")

		http.Error(w, "unauthorized", http.StatusUnauthorized)
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

	h.logger.Info("logged out")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("logged out"))
}

func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		h.logger.Warn("missing authorization cookie")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{
		Token: cookie.Value,
	})
	if err != nil {
		h.logger.Error("token validation grpc failure",
			zap.Error(err),
		)
		http.Error(w, "internal server error", http.StatusUnauthorized)
		return
	}
	if !resp.Valid {
		h.logger.Warn("invalid token provided")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetUserById(r.Context(), resp.UserId)
	if err != nil {
		h.logger.Error("failed to fetch user profile", zap.Error(err), zap.String("user_id", user.ID.Hex()))
		http.Error(w, "internal server error", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"id":    user.ID.Hex(),
		"email": user.Email,
	})
	h.logger.Info("profile fetched",
		zap.String("user_id", user.ID.Hex()),
	)
}

// grpc handlers
func (h *UserHandler) VerifyCredentials(ctx context.Context, req *userpb.VerifyCredentialsRequest) (*userpb.VerifyCredentialsResponse, error) {
	user, err := h.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.logger.Error(
			"db error fetching user",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		return nil, status.Error(
			codes.Internal,
			"internal server error",
		)
	}
	if user == nil {
		h.logger.Warn("verifycreds attempt with non existent email", zap.String("email", req.Email))
		return &userpb.VerifyCredentialsResponse{
			Valid: false,
			// Token: "",
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
		h.logger.Error(
			"auth service authenticate failure",
			zap.Error(err),
		)
		return nil, status.Error(
			codes.Internal,
			"authentication service unavailable",
		)
	}

	return &userpb.VerifyCredentialsResponse{
		Valid: authResp.Valid,
		// Token:  authResp.Token,
		UserId: user.ID.Hex(),
	}, nil
}

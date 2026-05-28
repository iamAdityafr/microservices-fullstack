package handler

import (
	"auth-service/internal/service"
	"context"
	"grpc_module/auth/authpb"

	"go.uber.org/zap"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	logger      *zap.Logger
	authService *service.AuthService
}

func NewAuthHandler(logger *zap.Logger, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		logger:      logger,
		authService: authService,
	}
}

func (h *AuthHandler) Authenticate(ctx context.Context, req *authpb.AuthRequest) (*authpb.AuthResponse, error) {
	token, err := h.authService.Authenticate(req.Email, req.Password, req.HashedPassword, req.UserId)
	if err != nil {
		h.logger.Error("err in authenticating user", zap.String("email", req.Email), zap.Error(err))
		return &authpb.AuthResponse{
			Token:  "",
			UserId: "",
			Valid:  false,
		}, nil
	}

	return &authpb.AuthResponse{
		Token:  token,
		UserId: req.UserId,
		Valid:  true,
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	valid, userid := h.authService.ValidateToken(req.Token)

	return &authpb.ValidateTokenResponse{
		Valid:  valid,
		UserId: userid,
	}, nil
}
func (h *AuthHandler) GeneratePassword(ctx context.Context, req *authpb.BcryptPasswordRequest) (*authpb.BcryptPasswordResponse, error) {
	hashed, err := h.authService.GeneratePassword(req.Password)
	if err != nil || req.Password == "" {
		h.logger.Error("err generating password", zap.Error(err))
		return nil, err
	}
	return &authpb.BcryptPasswordResponse{HashedPassword: hashed}, nil
}

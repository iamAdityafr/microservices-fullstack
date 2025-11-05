package handler

import (
	"auth-service/internal/service"
	"bck/auth/authpb"
	"context"
	"log"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Authenticate(ctx context.Context, req *authpb.AuthRequest) (*authpb.AuthResponse, error) {
	token, err := h.authService.Authenticate(req.Email, req.Password, req.HashedPassword, req.UserId)
	if err != nil {
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
	log.Printf("[auth-service] ValidateTokenRequest token=%s with userId=%s", req.Token, userid)
	return &authpb.ValidateTokenResponse{
		Valid:  valid,
		UserId: userid,
	}, nil
}

package service

import (
	"auth-service/utils"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	Authenticate(email, password, hashedPassword string) (token string, err error)
	ValidateToken(token string) (valid bool, email string)
}

type AuthService struct{}

func NewAuthService() (*AuthService, error) {
	return &AuthService{}, nil
}

func (s *AuthService) Authenticate(email, password, hashedPassword string, userid string) (token string, err error) {
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	token, err = utils.GenerateJWT(email, userid)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) ValidateToken(token string) (valid bool, userId string) {
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		return false, ""
	}
	return true, claims.UserID
}

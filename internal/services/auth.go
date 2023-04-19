package services

import (
	"context"
	"errors"
	"github.com/LorezV/go-diploma.git/internal/models"
	"github.com/LorezV/go-diploma.git/internal/repository"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var ErrInvalidCredentials = errors.New("credentials don't match")

type AuthService struct {
	userRepo repository.Users
	secret   string
}

func MakeAuthService(repo repository.Users, secret string) *AuthService {
	return &AuthService{
		userRepo: repo,
		secret:   secret,
	}
}

func (as *AuthService) GetSecret() string {
	return as.secret
}

func (as *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := as.userRepo.FindByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err = as.checkPassword(user.Password, password); err != nil {
		return "", ErrInvalidCredentials
	}

	return as.GenerateToken(user)
}

func (as *AuthService) GenerateToken(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(5 * 24 * time.Hour).Unix()
	claims["iat"] = time.Now().Unix()
	claims["sub"] = user.Login

	tokenString, err := token.SignedString([]byte(as.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (as *AuthService) checkPassword(hashedPassword, providedPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword)); err != nil {
		return err
	}

	return nil
}

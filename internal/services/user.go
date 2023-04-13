package services

import (
	"context"
	"errors"
	"github.com/LorezV/go-diploma.git/internal/models"
	"github.com/LorezV/go-diploma.git/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrLoginTaken = errors.New("this login is already taken")

type UserService struct {
	r repository.Users
}

func MakeUserService(repo repository.Users) *UserService {
	return &UserService{
		r: repo,
	}
}

func (us *UserService) Create(ctx context.Context, login, password string) (*models.User, error) {
	user, err := us.r.FindByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, ErrLoginTaken
	}

	hashedPassword, err := us.hashPassword(password)
	if err != nil {
		return nil, err
	}

	if err = us.r.Create(ctx, login, hashedPassword); err != nil {
		return nil, err
	}

	return us.r.FindByLogin(ctx, login)
}

func (us *UserService) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	return us.r.FindByLogin(ctx, login)
}

func (us *UserService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

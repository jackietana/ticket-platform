package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/auth-service/internal/domain"
	"github.com/jackietana/ticket-platform/auth-service/pkg/hash"
)

type Repository interface {
	CreateUser(ctx context.Context, usr domain.User) error
	GetUser(ctx context.Context, inp domain.UserInput) (domain.User, error)
}

type Cache interface {
	AddSession(ctx context.Context, session domain.Session) error
	GetUserId(ctx context.Context, token string) (string, error)
}

type AuthService struct {
	hasher *hash.SHA1Hasher
	repo   Repository
	cache  Cache
}

func NewAuthService(hasher *hash.SHA1Hasher, repo Repository, cache Cache) *AuthService {
	return &AuthService{
		hasher: hasher,
		repo:   repo,
		cache:  cache,
	}
}

func (s *AuthService) SignUp(ctx context.Context, usr domain.User) error {
	hpass, err := s.hasher.Hash(usr.Password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	usr.Password = hpass

	return s.repo.CreateUser(ctx, usr)
}

func (s *AuthService) SignIn(ctx context.Context, inp domain.UserInput) (string, error) {
	hpass, err := s.hasher.Hash(inp.Password)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}

	inp.Password = hpass
	usr, err := s.repo.GetUser(ctx, inp)
	if err != nil {
		return "", fmt.Errorf("error getting user: %w", err)
	}

	token := uuid.New().String()
	err = s.cache.AddSession(ctx, domain.Session{
		Token:  token,
		UserId: usr.ID,
	})

	return token, err
}

func (s *AuthService) GetUserIdByToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", errors.New("empty token provided")
	}

	userId, err := s.cache.GetUserId(ctx, token)
	if err != nil {
		return "", fmt.Errorf("error validating token: %w", err)
	}

	return userId, nil
}

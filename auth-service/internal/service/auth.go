package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackietana/ticket-platform/auth-service/internal/domain"
	"github.com/jackietana/ticket-platform/auth-service/pkg/hash"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenExpired       = errors.New("token expired or invalid")
)

type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
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

func (s *AuthService) SignUp(ctx context.Context, usr domain.User) (string, error) {
	passHash, err := s.hasher.Hash(usr.Password)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}

	id, err := s.repo.CreateUser(ctx, usr.Email, passHash)
	if err != nil {
		return "", fmt.Errorf("error creating user in db: %w", err)
	}

	return id, nil
}

func (s *AuthService) SignIn(ctx context.Context, inp domain.User) (string, error) {
	usr, err := s.repo.GetUserByEmail(ctx, inp.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	passHash, err := s.hasher.Hash(inp.Password)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}

	if usr.Password != passHash {
		return "", ErrInvalidCredentials
	}

	token := uuid.New().String()
	err = s.cache.AddSession(ctx, domain.Session{
		Token:  token,
		UserId: usr.ID,
	})
	if err != nil {
		return "", fmt.Errorf("error saving session to cache: %w", err)
	}

	return token, nil
}

func (s *AuthService) GetUserIdByToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", errors.New("empty token provided")
	}

	userId, err := s.cache.GetUserId(ctx, token)
	if err != nil {
		return "", ErrTokenExpired
	}

	return userId, nil
}

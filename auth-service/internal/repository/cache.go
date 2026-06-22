package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackietana/ticket-platform/auth-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const SESSION_TTL = time.Hour * 24

type Cache struct {
	db *redis.Client
}

func NewCache(db *redis.Client) *Cache {
	return &Cache{db}
}

func (c *Cache) AddSession(ctx context.Context, session domain.Session) error {
	key := fmt.Sprintf("session:%s", session.Token)

	return c.db.Set(ctx, key, session.UserId.String(), SESSION_TTL).Err()
}

func (c *Cache) GetUserId(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf("session:%s", token)
	usrId := c.db.Get(ctx, key).Val()
	if usrId == "" {
		return "", errors.New("user by token not found")
	}

	return usrId, nil
}

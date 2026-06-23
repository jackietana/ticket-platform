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

	usrId, err := c.db.Get(ctx, key).Result()
	if nonexistent := errors.Is(err, redis.Nil); nonexistent {
		return "", errors.New("session expires or not found")
	} else if err != nil {
		return "", fmt.Errorf("redis error: %w", err)
	}

	return usrId, nil
}

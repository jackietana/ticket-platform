package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackietana/ticket-platform/auth-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	SESSION_TTL          = time.Hour * 24
	TOKENS_PLACEHOLDER   = "session:%s"
	SESSIONS_PLACEHOLDER = "user:sessions:%s"
	MAX_SESSIONS         = 5
)

type Cache struct {
	db *redis.Client
}

func NewCache(db *redis.Client) *Cache {
	return &Cache{db}
}

func (c *Cache) AddSession(ctx context.Context, session domain.Session) error {
	userIdStr := session.UserId.String()
	sessionsKey := fmt.Sprintf(SESSIONS_PLACEHOLDER, userIdStr)
	tokensKey := fmt.Sprintf(TOKENS_PLACEHOLDER, session.Token)

	if err := c.clearExpiredSessions(ctx, session.UserId.String()); err != nil {
		return fmt.Errorf("failed to clear expired sessions: %w", err)
	}

	sessionsCount, err := c.db.ZCard(ctx, sessionsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get sessions count: %w", err)
	}

	if sessionsCount >= MAX_SESSIONS {
		lastSession, err := c.db.ZPopMin(ctx, sessionsKey, 1).Result()
		if err != nil {
			return fmt.Errorf("failed to pop oldest session: %w", err)
		}

		if len(lastSession) > 0 {
			oldToken := lastSession[0].Member.(string)
			oldTokenKey := fmt.Sprintf(TOKENS_PLACEHOLDER, oldToken)

			if err := c.db.Del(ctx, oldTokenKey).Err(); err != nil {
				return fmt.Errorf("failed to delete old token key: %w", err)
			}
		}
	}

	if err := c.db.Set(ctx, tokensKey, session.UserId.String(), SESSION_TTL).Err(); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	if err := c.db.ZAdd(ctx, sessionsKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: session.Token,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add session to zset: %w", err)
	}

	return nil
}

func (c *Cache) GetUserId(ctx context.Context, token string) (string, error) {
	key := fmt.Sprintf(TOKENS_PLACEHOLDER, token)

	usrId, err := c.db.Get(ctx, key).Result()
	if nonexistent := errors.Is(err, redis.Nil); nonexistent {
		return "", errors.New("session expires or not found")
	} else if err != nil {
		return "", fmt.Errorf("redis error: %w", err)
	}

	return usrId, nil
}

func (c *Cache) clearExpiredSessions(ctx context.Context, userId string) error {
	key := fmt.Sprintf(SESSIONS_PLACEHOLDER, userId)
	expirationTime := time.Now().Add(time.Minute).Add(-SESSION_TTL).Unix()

	return c.db.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(expirationTime, 10)).Err()
}

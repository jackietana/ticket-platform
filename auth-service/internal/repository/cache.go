package repository

import (
	"context"
	"encoding/json"
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
	sessionsKey := fmt.Sprintf(SESSIONS_PLACEHOLDER, session.UserID)
	tokensKey := fmt.Sprintf(TOKENS_PLACEHOLDER, session.Token)

	sessionBytes, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := c.clearExpiredSessions(ctx, session.UserID); err != nil {
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

	if err := c.db.Set(ctx, tokensKey, string(sessionBytes), SESSION_TTL).Err(); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	if err := c.db.ZAdd(ctx, sessionsKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: session.Token,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add session to zset: %w", err)
	}

	return nil
}

func (c *Cache) DeleteSession(ctx context.Context, session domain.Session) error {
	sessionsKey := fmt.Sprintf(SESSIONS_PLACEHOLDER, session.UserID)
	tokensKey := fmt.Sprintf(TOKENS_PLACEHOLDER, session.Token)

	if err := c.db.ZRem(ctx, sessionsKey, session.Token).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if err := c.db.Del(ctx, tokensKey).Err(); err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

func (c *Cache) GetSessionContext(ctx context.Context, token string) (domain.Session, error) {
	var session domain.Session
	key := fmt.Sprintf(TOKENS_PLACEHOLDER, token)

	sessionStr, err := c.db.Get(ctx, key).Result()
	if nonexistent := errors.Is(err, redis.Nil); nonexistent {
		return domain.Session{}, errors.New("session expires or not found")
	} else if err != nil {
		return domain.Session{}, fmt.Errorf("redis error: %w", err)
	}

	if err := json.Unmarshal([]byte(sessionStr), &session); err != nil {
		return domain.Session{}, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return session, nil
}

func (c *Cache) clearExpiredSessions(ctx context.Context, userId string) error {
	key := fmt.Sprintf(SESSIONS_PLACEHOLDER, userId)
	expirationTime := time.Now().Add(time.Minute).Add(-SESSION_TTL).Unix()

	return c.db.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(expirationTime, 10)).Err()
}

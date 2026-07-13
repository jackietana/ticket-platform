package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const bookSlotsScript = `
local occupied_key = KEYS[1]
local booking_key = KEYS[2]

local total_slots = tonumber(ARGV[1])
local slots_count = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

local current_occupied = tonumber(redis.call('GET', occupied_key) or "0")

if current_occupied + slots_count > total_slots then
    return 0
end

redis.call('INCRBY', occupied_key, slots_count)
redis.call('SET', booking_key, "active", 'EX', ttl)

return 1
`

type RedisRepository struct {
	rdb *redis.Client
}

func NewRedisRepository(rdb *redis.Client) *RedisRepository {
	return &RedisRepository{rdb}
}

func (r *RedisRepository) BookSlots(ctx context.Context, eventID, orderID string, totalSlots, slotsCount int,
	ttl time.Duration) (bool, error) {

	occupiedKey := fmt.Sprintf("event:%s:occupied_slots", eventID)
	bookingKey := fmt.Sprintf("booking:%s", orderID)

	res, err := r.rdb.Eval(ctx, bookSlotsScript, []string{occupiedKey, bookingKey}, totalSlots, slotsCount,
		int(ttl.Seconds())).Result()
	if err != nil {
		return false, fmt.Errorf("failed to execute lua script: %w", err)
	}

	status, ok := res.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected script result type: %T", res)
	}

	return status == 1, nil
}

func (r *RedisRepository) ConfirmBooking(ctx context.Context, orderID string) error {
	bookingKey := fmt.Sprintf("booking:%s", orderID)
	if err := r.rdb.Del(ctx, bookingKey).Err(); err != nil {
		return fmt.Errorf("failed to remove booking key: %w", err)
	}

	return nil
}

func (r *RedisRepository) CancelBooking(ctx context.Context, eventID, orderID string, slotsCount int) error {
	bookingKey := fmt.Sprintf("booking:%s", orderID)
	occupiedKey := fmt.Sprintf("event:%s:occupied_slots", eventID)

	pipe := r.rdb.Pipeline()
	pipe.Del(ctx, bookingKey)
	pipe.DecrBy(ctx, occupiedKey, int64(slotsCount))

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute redis pipeline on cancel booking: %w", err)
	}

	return nil
}

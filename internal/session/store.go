package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/c-annabel/NomVoxBranding/internal/models"
	"github.com/redis/go-redis/v9"
)

const sessionTTL = 2 * time.Hour

// Store wraps a Redis client and provides typed read/write for BrandSession.
type Store struct {
	client *redis.Client
}

// New creates a new Store from a Redis URL (e.g. from REDIS_URL env var).
func New(redisURL string) (*Store, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("session.New: invalid redis URL: %w", err)
	}
	return &Store{client: redis.NewClient(opts)}, nil
}

func key(sessionID string) string {
	return "session:" + sessionID
}

// Get retrieves a BrandSession by ID. Returns nil, nil if not found.
func (s *Store) Get(ctx context.Context, sessionID string) (*models.BrandSession, error) {
	val, err := s.client.Get(ctx, key(sessionID)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("session.Get: %w", err)
	}
	var sess models.BrandSession
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		return nil, fmt.Errorf("session.Get: unmarshal: %w", err)
	}
	return &sess, nil
}

// Set writes a BrandSession to Redis, resetting the TTL to 2 hours.
func (s *Store) Set(ctx context.Context, sess *models.BrandSession) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("session.Set: marshal: %w", err)
	}
	return s.client.Set(ctx, key(sess.SessionID), data, sessionTTL).Err()
}

// Ping checks the Redis connection. Called during startup.
func (s *Store) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

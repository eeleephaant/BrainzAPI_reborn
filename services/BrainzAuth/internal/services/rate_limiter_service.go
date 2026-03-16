package services

import (
	"brainz/auth/internal/config"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RateLimitDecision struct {
	Allowed    bool
	Limit      int64
	Remaining  int64
	RetryAfter time.Duration
	ResetAfter time.Duration
	Scope      string
}

type RateLimiterService struct {
	rc     *redis.Client
	cfg    config.RateLimitConfig
	script *redis.Script
}

func NewRateLimiterService(rc *redis.Client, cfg config.RateLimitConfig) *RateLimiterService {
	return &RateLimiterService{
		rc:  rc,
		cfg: cfg,
		script: redis.NewScript(`
		local current = redis.call("INCR", KEYS[1])
		if current == 1 then
		redis.call("PEXPIRE", KEYS[1], ARGV[1])
		end
		local ttl = redis.call("PTTL", KEYS[1])
		return {current, ttl}
		`),
	}
}

func (rls *RateLimiterService) Enabled() bool {
	return rls != nil && rls.cfg.Enabled && rls.rc != nil
}

func (rls *RateLimiterService) ShouldFailOpen() bool {
	if rls == nil {
		return true
	}
	return rls.cfg.FailOpen
}

func (rls *RateLimiterService) AllowIP(ctx context.Context, ip string) (RateLimitDecision, error) {
	windowSeconds := maxInt64(1, rls.cfg.IPWindowSeconds)
	window := time.Duration(windowSeconds) * time.Second
	key := fmt.Sprintf("rl:ip:%s:%d", ip, time.Now().Unix()/windowSeconds)
	return rls.allow(ctx, key, rls.cfg.IPMaxRequests, window, "ip")
}

func (rls *RateLimiterService) AllowKey(ctx context.Context, keyID uuid.UUID) (RateLimitDecision, error) {
	windowSeconds := maxInt64(1, rls.cfg.KeyWindowSeconds)
	window := time.Duration(windowSeconds) * time.Second
	key := fmt.Sprintf("rl:key:%s:%d", keyID.String(), time.Now().Unix()/windowSeconds)
	return rls.allow(ctx, key, rls.cfg.KeyMaxRequests, window, "api_key")
}

func (rls *RateLimiterService) allow(ctx context.Context, key string, limit int64, window time.Duration, scope string) (RateLimitDecision, error) {
	if !rls.Enabled() {
		return RateLimitDecision{Allowed: true, Scope: scope}, nil
	}
	if limit <= 0 || window <= 0 {
		return RateLimitDecision{Allowed: true, Scope: scope}, nil
	}

	out, err := rls.script.Run(ctx, rls.rc, []string{key}, window.Milliseconds()).Result()
	if err != nil {
		return RateLimitDecision{}, fmt.Errorf("rate limiter script run: %w", err)
	}

	values, ok := out.([]interface{})
	if !ok || len(values) != 2 {
		return RateLimitDecision{}, fmt.Errorf("rate limiter script unexpected payload")
	}

	current, err := toInt64(values[0])
	if err != nil {
		return RateLimitDecision{}, fmt.Errorf("rate limiter parse current: %w", err)
	}
	ttlMs, err := toInt64(values[1])
	if err != nil {
		return RateLimitDecision{}, fmt.Errorf("rate limiter parse ttl: %w", err)
	}

	if ttlMs < 0 {
		ttlMs = int64(window / time.Millisecond)
	}
	remaining := limit - current
	if remaining < 0 {
		remaining = 0
	}

	decision := RateLimitDecision{
		Allowed:    current <= limit,
		Limit:      limit,
		Remaining:  remaining,
		ResetAfter: time.Duration(ttlMs) * time.Millisecond,
		Scope:      scope,
	}
	if !decision.Allowed {
		decision.RetryAfter = decision.ResetAfter
	}

	return decision, nil
}

func toInt64(v interface{}) (int64, error) {
	switch n := v.(type) {
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	default:
		return 0, fmt.Errorf("unexpected numeric type %T", v)
	}
}

func maxInt64(a int64, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

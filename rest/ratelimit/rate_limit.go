package ratelimit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type Key string

type rlKey [2]string

type RateLimiter struct {
	lastRequest map[rlKey]time.Time
	config      map[Key]time.Duration
}

func NewRateLimiter(config map[Key]time.Duration) *RateLimiter {
	return &RateLimiter{
		config:      config,
		lastRequest: make(map[rlKey]time.Time),
	}
}

func (r *RateLimiter) CheckAndSet(key Key, args string) error {
	lrKey := rlKey{string(key), args}
	lastTime, ok := r.lastRequest[lrKey]
	if !ok {
		goto set
	}
	if time.Since(lastTime) < r.config[key] {
		return echo.NewHTTPError(http.StatusTooManyRequests, fmt.Sprintf("too many %s requests, try again in %v", key, r.config[key]-(time.Since(lastTime))))
	}
set:
	r.lastRequest[lrKey] = time.Now()
	return nil
}

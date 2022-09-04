package rest

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"

	"query-adventure/cfg"
)

type rateLimitKind string

const (
	rlQuery rateLimitKind = "query"
	rlCheck rateLimitKind = "check"
)

type rateLimitState map[rateLimitKind]time.Time

func init() {
	gob.Register(make(rateLimitState))
}

const rateLimitSessionKey = "ratelimit"

func checkAndSetRateLimit(c echo.Context, kind rateLimitKind, g *cfg.Globals) error {
	limit, ok := g.RateLimits[string(kind)]
	if !ok {
		return fmt.Errorf("no rate limit configured for %q", kind)
	}
	sess, err := session.Get(rateLimitSessionKey, c)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	rls, ok := sess.Values[rateLimitSessionKey].(rateLimitState)
	if !ok {
		rls = make(rateLimitState)
	}
	lastTime, ok := rls[kind]
	if !ok {
		goto set
	}
	if time.Since(lastTime) < limit {
		return echo.NewHTTPError(http.StatusTooManyRequests, fmt.Sprintf("rate limit exceeded, try again in %v", limit-(time.Since(lastTime))))
	}
set:
	rls[kind] = time.Now()
	sess.Values[rateLimitSessionKey] = rls
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

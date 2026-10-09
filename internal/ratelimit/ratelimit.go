// Package ratelimit provides an in-memory per-IP rate limiter.
//
// State lives in the process, so limits apply per instance. That is enough
// for a single container; a shared store would be needed to scale out.
package ratelimit

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// idleTTL is how long an IP is remembered after its last request.
const idleTTL = 10 * time.Minute

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type Limiter struct {
	mu        sync.Mutex
	clients   map[string]*client
	limit     rate.Limit
	burst     int
	lastSweep time.Time
	now       func() time.Time
}

// New allows `burst` requests at once per IP, refilling at `limit` per second.
func New(limit rate.Limit, burst int) *Limiter {
	return &Limiter{
		clients: map[string]*client{},
		limit:   limit,
		burst:   burst,
		now:     time.Now,
	}
}

// allow reports whether ip may proceed; otherwise how long it should wait.
func (l *Limiter) allow(ip string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if now.Sub(l.lastSweep) > idleTTL {
		for key, c := range l.clients {
			if now.Sub(c.lastSeen) > idleTTL {
				delete(l.clients, key)
			}
		}
		l.lastSweep = now
	}

	c, ok := l.clients[ip]
	if !ok {
		c = &client{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.clients[ip] = c
	}
	c.lastSeen = now

	reservation := c.limiter.ReserveN(now, 1)
	if delay := reservation.DelayFrom(now); !reservation.OK() || delay > 0 {
		reservation.CancelAt(now)
		return false, delay
	}

	return true, 0
}

// Middleware answers 429 with Retry-After once an IP exceeds its limit.
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ok, wait := l.allow(ctx.ClientIP())
		if ok {
			ctx.Next()
			return
		}

		seconds := int(math.Ceil(wait.Seconds()))
		if seconds < 1 {
			seconds = 1
		}
		ctx.Header("Retry-After", strconv.Itoa(seconds))
		ctx.String(http.StatusTooManyRequests, "too many requests")
		ctx.Abort()
	}
}

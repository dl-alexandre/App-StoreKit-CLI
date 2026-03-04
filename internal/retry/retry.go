package retry

import (
	"context"
	"crypto/rand"
	"math"
	"math/big"
	"time"
)

type Config struct {
	MaxRetries int
	Backoff    time.Duration
	MaxBackoff time.Duration
}

type Result struct {
	Retry bool
	Err   error
	Value any
	Wait  time.Duration
}

func Do(ctx context.Context, cfg Config, fn func() Result) Result {
	attempts := cfg.MaxRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	backoff := cfg.Backoff
	if backoff <= 0 {
		backoff = 500 * time.Millisecond
	}

	maxBackoff := cfg.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 8 * time.Second
	}

	var last Result
	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return Result{Err: ctx.Err()}
		}
		last = fn()
		if !last.Retry || i == attempts-1 {
			return last
		}
		sleep := last.Wait
		if sleep <= 0 {
			sleep = jitter(backoff)
			if sleep > maxBackoff {
				sleep = maxBackoff
			}
		}
		timer := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			timer.Stop()
			return Result{Err: ctx.Err()}
		case <-timer.C:
		}
		backoff = time.Duration(float64(backoff) * 2)
	}

	return last
}

func jitter(base time.Duration) time.Duration {
	// Use crypto/rand for secure random jitter
	jitterInt, _ := rand.Int(rand.Reader, big.NewInt(100))
	jitterFraction := float64(jitterInt.Int64())/100.0*0.4 + 0.8
	return time.Duration(math.Round(float64(base) * jitterFraction))
}

package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoRetriesConfiguredNumberOfTimes(t *testing.T) {
	calls := 0
	got := Do(context.Background(), Config{MaxRetries: 2, Backoff: time.Nanosecond}, func() Result {
		calls++
		if calls < 3 {
			return Result{Retry: true, Err: errors.New("temporary")}
		}
		return Result{Value: "ok"}
	})
	if calls != 3 || got.Value != "ok" || got.Err != nil {
		t.Fatalf("calls = %d, result = %+v", calls, got)
	}
}

func TestDoStopsWithoutRetry(t *testing.T) {
	calls := 0
	got := Do(context.Background(), Config{MaxRetries: 4}, func() Result {
		calls++
		return Result{Err: errors.New("permanent")}
	})
	if calls != 1 || got.Err == nil {
		t.Fatalf("calls = %d, result = %+v", calls, got)
	}
}

func TestDoHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	got := Do(ctx, Config{MaxRetries: 2}, func() Result {
		calls++
		return Result{Value: "unexpected"}
	})
	if calls != 0 || !errors.Is(got.Err, context.Canceled) {
		t.Fatalf("calls = %d, result = %+v", calls, got)
	}
}

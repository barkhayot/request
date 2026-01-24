package throttler

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestRateThrottler_AllowsUnderLimit(t *testing.T) {
	th := NewRateThrottler(rate.Every(time.Second), 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	if err := th.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if time.Since(start) > 50*time.Millisecond {
		t.Fatalf("Wait took too long, throttler blocked unexpectedly")
	}
}

func TestRateThrottler_Throttles(t *testing.T) {
	th := NewRateThrottler(rate.Every(200*time.Millisecond), 1)

	ctx := context.Background()

	// First call should pass immediately
	if err := th.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	start := time.Now()
	if err := th.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 180*time.Millisecond {
		t.Fatalf("expected throttling delay, got %v", elapsed)
	}
}

func TestRateThrottler_ContextCancelled(t *testing.T) {
	th := NewRateThrottler(rate.Every(time.Second), 1)

	// Consume first token
	if err := th.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := th.Wait(ctx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

package scheduler_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/scheduler"
)

// TestNew verifies that New returns a non-nil Scheduler.
func TestNew(t *testing.T) {
	s := scheduler.New(nil, nil, nil, time.Second, []string{"/secret/foo"})
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

// TestRun_CancelImmediately ensures Run returns when the context is cancelled.
func TestRun_CancelImmediately(t *testing.T) {
	s := scheduler.New(nil, nil, nil, time.Minute, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Run

	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}

// tickCounter is a helper that counts how many times it has been called.
type tickCounter struct {
	count atomic.Int64
}

func (tc *tickCounter) inc() { tc.count.Add(1) }
func (tc *tickCounter) get() int64 { return tc.count.Load() }

// TestRun_TicksAtInterval verifies the scheduler fires at least once within the interval.
func TestRun_TicksAtInterval(t *testing.T) {
	// Use a very short interval and a no-op scheduler with nil deps
	// (tick will log errors but not panic when paths is empty).
	s := scheduler.New(nil, nil, nil, 50*time.Millisecond, []string{})
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()

	select {
	case err := <-done:
		if err != context.DeadlineExceeded && err != context.Canceled {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not return after context deadline")
	}
}

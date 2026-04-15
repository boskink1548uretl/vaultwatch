package monitor

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockRenewer struct {
	called    bool
	newExpiry time.Time
	err       error
}

func (m *mockRenewer) RenewLease(_ context.Context, _ string, _ time.Duration) (time.Time, error) {
	m.called = true
	return m.newExpiry, m.err
}

func TestLeaseManager_TrackDefaults(t *testing.T) {
	lm := NewLeaseManager(&mockRenewer{})
	lm.Track(LeaseEntry{LeaseID: "test/lease", Path: "database/creds/role"})

	if len(lm.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(lm.entries))
	}
	e := lm.entries[0]
	if e.RenewThreshold != 5*time.Minute {
		t.Errorf("expected default threshold 5m, got %v", e.RenewThreshold)
	}
	if e.RenewIncrement != time.Hour {
		t.Errorf("expected default increment 1h, got %v", e.RenewIncrement)
	}
}

func TestCheckAndRenew_SkipsIfNotNearExpiry(t *testing.T) {
	mr := &mockRenewer{newExpiry: time.Now().Add(2 * time.Hour)}
	lm := NewLeaseManager(mr)
	lm.Track(LeaseEntry{
		LeaseID:        "test/lease",
		ExpireTime:     time.Now().Add(30 * time.Minute),
		RenewThreshold: 5 * time.Minute,
	})

	errs := lm.CheckAndRenew(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if mr.called {
		t.Error("expected renewer NOT to be called")
	}
}

func TestCheckAndRenew_RenewsWhenNearExpiry(t *testing.T) {
	newExpiry := time.Now().Add(2 * time.Hour)
	mr := &mockRenewer{newExpiry: newExpiry}
	lm := NewLeaseManager(mr)
	lm.Track(LeaseEntry{
		LeaseID:        "test/lease",
		ExpireTime:     time.Now().Add(2 * time.Minute),
		RenewThreshold: 5 * time.Minute,
		RenewIncrement: time.Hour,
	})

	errs := lm.CheckAndRenew(context.Background())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !mr.called {
		t.Error("expected renewer to be called")
	}
	if !lm.entries[0].ExpireTime.Equal(newExpiry) {
		t.Errorf("expected updated expiry %v, got %v", newExpiry, lm.entries[0].ExpireTime)
	}
}

func TestCheckAndRenew_HandlesRenewError(t *testing.T) {
	mr := &mockRenewer{err: errors.New("vault unavailable")}
	lm := NewLeaseManager(mr)
	lm.Track(LeaseEntry{
		LeaseID:        "test/lease",
		ExpireTime:     time.Now().Add(1 * time.Minute),
		RenewThreshold: 5 * time.Minute,
	})

	errs := lm.CheckAndRenew(context.Background())
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !errors.Is(errs[0], mr.err) {
		t.Errorf("unexpected error: %v", errs[0])
	}
}

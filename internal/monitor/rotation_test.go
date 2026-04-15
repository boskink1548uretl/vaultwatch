package monitor

import (
	"context"
	"testing"
	"time"
)

func TestMaybeRotate_SkipsWhenAutoRotateDisabled(t *testing.T) {
	rm := NewRotationManager(nil, RotationPolicy{AutoRotate: false})
	result := CheckResult{
		Path:  "secret/app/creds",
		Level: LevelCritical,
	}
	rr, err := rm.MaybeRotate(context.Background(), result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rr != nil {
		t.Errorf("expected nil RotateResult when AutoRotate=false, got %+v", rr)
	}
}

func TestMaybeRotate_SkipsNonCriticalLevels(t *testing.T) {
	rm := NewRotationManager(nil, RotationPolicy{AutoRotate: true})

	levels := []AlertLevel{LevelNone, LevelWarning}
	for _, lvl := range levels {
		result := CheckResult{
			Path:  "secret/app/token",
			Level: lvl,
		}
		rr, err := rm.MaybeRotate(context.Background(), result)
		if err != nil {
			t.Fatalf("level=%s: unexpected error: %v", lvl, err)
		}
		if rr != nil {
			t.Errorf("level=%s: expected nil RotateResult, got %+v", lvl, rr)
		}
	}
}

func TestMaybeRotate_PolicyNewDataFnCalled(t *testing.T) {
	called := false
	policy := RotationPolicy{
		AutoRotate: true,
		NewDataFn: func(path string) map[string]interface{} {
			called = true
			return map[string]interface{}{"rotated": true, "path": path}
		},
	}

	// We can't call RotateSecret without a real client; verify NewDataFn is
	// invoked by checking the flag after a nil-client call.
	rm := NewRotationManager(nil, policy)
	result := CheckResult{
		Path:      "secret/svc/key",
		Level:     LevelExpired,
		ExpiresAt: time.Now().Add(-time.Hour),
	}

	// This will error because client is nil, but NewDataFn should be called
	// before the client call.
	_, _ = rm.MaybeRotate(context.Background(), result)

	if !called {
		t.Error("expected NewDataFn to be called for expired secret")
	}
}

func TestRotationPolicy_DefaultNewDataFn(t *testing.T) {
	policy := RotationPolicy{AutoRotate: true}
	if policy.NewDataFn != nil {
		t.Error("expected NewDataFn to be nil by default")
	}
}

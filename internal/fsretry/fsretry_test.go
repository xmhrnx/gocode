package fsretry

import (
	"errors"
	"testing"
)

func TestDoReturnsAtOnceOnSuccessOrLastingErrors(t *testing.T) {
	calls := 0
	if err := Do(func() error { calls++; return nil }); err != nil || calls != 1 {
		t.Fatalf("success: err=%v calls=%d", err, calls)
	}
	lasting := errors.New("disk full")
	calls = 0
	if err := Do(func() error { calls++; return lasting }); !errors.Is(err, lasting) || calls != 1 {
		t.Fatalf("lasting error: err=%v calls=%d", err, calls)
	}
}

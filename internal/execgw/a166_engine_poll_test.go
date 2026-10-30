package execgw

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
)

func TestA166EnginePollRejectsUnbound(t *testing.T) {
	r := &Retrier{AccountRef: "fixture"}
	_, err := NewEnginePollRetrier(r, nil, nil, "fixture", 1)
	if !errors.Is(err, accountbudget.ErrUnavailable) {
		t.Fatalf("unbound: %v", err)
	}
	called := 0
	if err := r.Query(context.Background(), QueryPrice, func(context.Context) error { called++; return nil }); err != nil || called != 1 {
		t.Fatalf("OFF err=%v calls=%d", err, called)
	}
}

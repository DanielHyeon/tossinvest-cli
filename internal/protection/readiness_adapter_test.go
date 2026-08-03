package protection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/protectionreadiness"
)

type readinessProviderStub struct {
	snapshot protectionreadiness.ReadinessSnapshot
	err      error
	calls    int
}

func (provider *readinessProviderStub) Current(context.Context) (protectionreadiness.ReadinessSnapshot, error) {
	provider.calls++
	return provider.snapshot, provider.err
}

func TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot(t *testing.T) {
	now := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	if _, err := NewReadinessAdapter(nil, "acct", "production"); err == nil {
		t.Fatal("nil provider created an adapter")
	}
	provider := &readinessProviderStub{snapshot: protectionreadiness.DefaultSnapshot()}
	adapter, err := NewReadinessAdapter(provider, "acct", "production")
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, refusal := adapter.Check(context.Background(), "us", now, Checkpoint{})
	if refusal == nil || refusal.Code != protectionreadiness.RefusalMissingEvidence || checkpoint.Valid() {
		t.Fatalf("default snapshot admitted checkpoint=%+v refusal=%+v", checkpoint, refusal)
	}
	if provider.calls != 1 {
		t.Fatalf("provider calls=%d", provider.calls)
	}

	provider.err = errors.New("read failed")
	if _, refusal = adapter.Check(context.Background(), "kr", now, Checkpoint{}); refusal == nil || refusal.Code != RefusalProviderUnavailable {
		t.Fatalf("provider failure=%+v", refusal)
	}
}

func TestReadinessAdapterRejectsInvalidMarketWithoutCallingProvider(t *testing.T) {
	provider := &readinessProviderStub{snapshot: protectionreadiness.DefaultSnapshot()}
	adapter, err := NewReadinessAdapter(provider, "acct", "production")
	if err != nil {
		t.Fatal(err)
	}
	if _, refusal := adapter.Check(context.Background(), "cn", time.Now(), Checkpoint{}); refusal == nil || refusal.Code != protectionreadiness.RefusalInvalid {
		t.Fatalf("invalid market refusal=%+v", refusal)
	}
	if provider.calls != 0 {
		t.Fatalf("invalid market called provider %d times", provider.calls)
	}
}

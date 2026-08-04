package candidate

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestOpenReadOnlyAssessesButCannotMutateTheDiscoveryStore(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), DBFileName)
	clk := clock.NewFake(t0)
	writable, err := Open(ctx, Options{Path: path, Clock: clk, FSProber: FixedFSProber(FSInfo{Name: "ext4"})})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writable.Promote(ctx, MarketKR, "005930", t0); err != nil {
		t.Fatal(err)
	}
	if err := writable.Close(); err != nil {
		t.Fatal(err)
	}

	readOnly, err := OpenReadOnly(ctx, Options{Path: path, Clock: clk, FSProber: FixedFSProber(FSInfo{Name: "ext4"})})
	if err != nil {
		t.Fatal(err)
	}
	defer readOnly.Close()
	if summaries, err := readOnly.Summaries(ctx, t0); err != nil || len(summaries) != 1 {
		t.Fatalf("Summaries = (%d, %v), want one", len(summaries), err)
	}
	if _, err := readOnly.Promote(ctx, MarketUS, "AAPL", t0.Add(time.Second)); err == nil {
		t.Fatal("read-only store accepted Promote")
	}
}

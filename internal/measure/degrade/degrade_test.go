package degrade_test

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/measure/degrade"
)

// degradation_test.go is change add-net-rr-measurement task 2.8, the counter's own
// half: the two tiers, the honesty about which one is in use, and the arithmetic.
//
// The other half — that the count lands when the *observation write* fails with a
// real SQLITE_FULL — is in internal/journal/observation_degradation_test.go, next
// to the fixture that can fill a database. Keeping it there is deliberate: that is
// the test that would fail if somebody ever moved the count back into the
// observation table, and it should live where they would be working.

func newCounter(t *testing.T, dir string) (*degrade.Counter, *bytes.Buffer) {
	t.Helper()
	var logged bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logged, nil))
	return degrade.NewCounter(filepath.Join(dir, "observation-degradation.json"), log), &logged
}

// TestRefusalAndIssuedLossesAreCountedApart is design D6's asymmetry: an issued
// verdict's lost observation is rebuildable from the preimage and a refusal's is
// not, so collapsing them into one number would report a recoverable gap and a
// permanent one as the same fact.
func TestRefusalAndIssuedLossesAreCountedApart(t *testing.T) {
	dir := t.TempDir()
	counter, _ := newCounter(t, dir)

	counter.Record(degrade.LossObservationWrite)
	counter.Record(degrade.LossRefusalUnrecoverable)
	counter.Record(degrade.LossRefusalUnrecoverable)
	counter.Record(degrade.LossLapsedBeyondHorizon)

	snap := counter.Snapshot()
	if snap.Counts[degrade.LossObservationWrite] != 1 ||
		snap.Counts[degrade.LossRefusalUnrecoverable] != 2 ||
		snap.Counts[degrade.LossLapsedBeyondHorizon] != 1 {
		t.Fatalf("counts = %v, want the three kinds tallied apart", snap.Counts)
	}
	if snap.Total() != 4 {
		t.Errorf("total = %d, want 4", snap.Total())
	}
	want := strings.Join([]string{
		degrade.LossObservationWrite,
		degrade.LossLapsedBeyondHorizon,
		degrade.LossRefusalUnrecoverable,
	}, ",")
	if got := strings.Join(snap.Kinds(), ","); got != want {
		t.Errorf("kinds must be listed in a stable order:\n got %s\nwant %s", got, want)
	}
}

// TestCountsAccumulateAcrossRestarts is what "durable" is claiming. Without it the
// tier distinction would be decoration.
func TestCountsAccumulateAcrossRestarts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "observation-degradation.json")

	first := degrade.NewCounter(path, nil)
	first.Record(degrade.LossObservationWrite)
	first.Record(degrade.LossObservationWrite)
	if !first.Snapshot().Durable {
		t.Fatal("a counter with a writable store is durable")
	}

	second := degrade.NewCounter(path, nil)
	second.Record(degrade.LossObservationWrite)

	if got := second.Snapshot().Counts[degrade.LossObservationWrite]; got != 3 {
		t.Fatalf("count after a restart = %d, want 3", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm&0o077 != 0 {
		t.Errorf("counter permissions = %o, want no group/other access", perm)
	}
}

// TestDegradesToLogAndMemoryWhenEvenTheFileFails is the second tier, and the
// honesty requirement attached to it (SHALL NOT — 그 강등된 계수를 재시작 후에도
// durable하다고 주장해서는 안 된다). The directory is made unwritable, so the atomic
// rewrite cannot happen at all; the process must keep counting and must say the
// counts are no longer durable.
func TestDegradesToLogAndMemoryWhenEvenTheFileFails(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions, so the write cannot be made to fail this way")
	}
	dir := t.TempDir()
	locked := filepath.Join(dir, "locked")
	if err := os.Mkdir(locked, 0o700); err != nil {
		t.Fatal(err)
	}
	var logged bytes.Buffer
	counter := degrade.NewCounter(filepath.Join(locked, "counter.json"),
		slog.New(slog.NewTextHandler(&logged, nil)))
	if err := os.Chmod(locked, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })

	counter.Record(degrade.LossObservationWrite)
	counter.Record(degrade.LossObservationWrite)

	snap := counter.Snapshot()
	if snap.Counts[degrade.LossObservationWrite] != 2 {
		t.Fatalf("in-process count = %d, want 2", snap.Counts[degrade.LossObservationWrite])
	}
	if snap.Durable {
		t.Error("a counter that cannot write its file must not claim its counts survive a restart")
	}
	out := logged.String()
	if strings.Count(out, "no longer durable") != 1 {
		t.Errorf("the demotion is announced once, not per record:\n%s", out)
	}
	if !strings.Contains(out, "loss_kind="+degrade.LossObservationWrite) {
		t.Errorf("every loss must produce a structured line, which is the floor "+
			"the demoted tier stands on:\n%s", out)
	}
}

// TestRecordNeverPanicsWithoutAStore covers the wiring that supplies no path at
// all. The counter is the last thing standing when storage has already failed, so
// no configuration of it may take the process down.
func TestRecordNeverPanicsWithoutAStore(t *testing.T) {
	counter := degrade.NewCounter("", nil)
	for i := 0; i < 3; i++ {
		counter.Record(degrade.LossObservationWrite, "i", strconv.Itoa(i))
	}
	snap := counter.Snapshot()
	if snap.Counts[degrade.LossObservationWrite] != 3 {
		t.Fatalf("count = %d, want 3", snap.Counts[degrade.LossObservationWrite])
	}
	if snap.Durable {
		t.Error("a counter with no store must not claim durability")
	}
}

// TestCorruptStoreStartsCleanRatherThanRefusing: a measurement fault must not be
// able to stop the engine starting. The counter reports itself degraded instead.
func TestCorruptStoreStartsCleanRatherThanRefusing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "counter.json")
	if err := os.WriteFile(path, []byte("this is not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	var logged bytes.Buffer
	counter := degrade.NewCounter(path, slog.New(slog.NewTextHandler(&logged, nil)))
	counter.Record(degrade.LossObservationWrite)

	snap := counter.Snapshot()
	if snap.Counts[degrade.LossObservationWrite] != 1 {
		t.Fatalf("count = %d, want 1", snap.Counts[degrade.LossObservationWrite])
	}
	if snap.Durable {
		t.Error("counts that started from an unreadable file are not a total since installation")
	}
	if !strings.Contains(logged.String(), "unreadable") {
		t.Errorf("the unreadable store must be announced:\n%s", logged.String())
	}
}

// TestCounterIsConcurrencySafe: the trading path can produce losses from more than
// one goroutine, and a lost increment would understate exactly the number an
// operator uses to decide whether the measurement can be trusted.
func TestCounterIsConcurrencySafe(t *testing.T) {
	dir := t.TempDir()
	counter, _ := newCounter(t, dir)

	const writers, each = 8, 25
	done := make(chan struct{})
	for i := 0; i < writers; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < each; j++ {
				counter.Record(degrade.LossObservationWrite)
			}
		}()
	}
	for i := 0; i < writers; i++ {
		<-done
	}
	if got := counter.Snapshot().Counts[degrade.LossObservationWrite]; got != writers*each {
		t.Fatalf("count = %d, want %d", got, writers*each)
	}
}

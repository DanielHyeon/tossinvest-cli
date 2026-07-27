package audit_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

var at = time.Date(2026, 7, 26, 9, 30, 0, 0, time.UTC)

func openLog(t *testing.T) (*audit.Log, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(at)
	log, err := audit.Open(audit.Options{
		Path:    filepath.Join(t.TempDir(), "sub", audit.FileName),
		Clock:   clk,
		Subject: "operator",
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return log, clk
}

// TestRecordsBeforeAndAfter is the §0.5 requirement in one test: the entry has
// the old value, the new value, the time and the subject.
func TestRecordsBeforeAndAfter(t *testing.T) {
	log, clk := openLog(t)

	wrote, err := log.RecordChange(audit.ActionGateToggle, "engine.automation_gate.enabled", "false", "")
	if err != nil {
		t.Fatalf("RecordChange: %v", err)
	}
	if !wrote {
		t.Fatal("the first observation of a setting must be recorded")
	}

	clk.Advance(time.Hour)
	wrote, err = log.RecordChange(audit.ActionGateToggle, "engine.automation_gate.enabled", "true", "operator turned it on")
	if err != nil {
		t.Fatalf("RecordChange: %v", err)
	}
	if !wrote {
		t.Fatal("a changed setting must be recorded")
	}

	entries, err := log.Entries()
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2: %+v", len(entries), entries)
	}
	second := entries[1]
	if second.Old != "false" || second.New != "true" {
		t.Errorf("old/new = %q/%q, want false/true", second.Old, second.New)
	}
	if second.Subject != "operator" {
		t.Errorf("subject = %q", second.Subject)
	}
	if !second.At.Equal(at.Add(time.Hour)) {
		t.Errorf("at = %v, want %v", second.At, at.Add(time.Hour))
	}
	if second.Action != audit.ActionGateToggle {
		t.Errorf("action = %q", second.Action)
	}
}

// TestUnchangedSettingIsNotRecorded: an audit log nobody can read is not an
// audit log, and recording every startup would bury the one line that matters.
func TestUnchangedSettingIsNotRecorded(t *testing.T) {
	log, _ := openLog(t)

	if _, err := log.RecordChange(audit.ActionLimitChange, "limit", "10", ""); err != nil {
		t.Fatalf("RecordChange: %v", err)
	}
	for i := 0; i < 5; i++ {
		wrote, err := log.RecordChange(audit.ActionLimitChange, "limit", "10", "")
		if err != nil {
			t.Fatalf("RecordChange: %v", err)
		}
		if wrote {
			t.Fatal("an unchanged setting must not be recorded again")
		}
	}
	entries, _ := log.Entries()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
}

// TestSettingsAreTrackedIndependently: two settings changing at different times
// must not shadow each other's history.
func TestSettingsAreTrackedIndependently(t *testing.T) {
	log, _ := openLog(t)

	mustRecord(t, log, "a", "1")
	mustRecord(t, log, "b", "2")
	mustRecord(t, log, "a", "3")

	latest, found, err := log.Latest("b")
	if err != nil || !found {
		t.Fatalf("Latest(b) = %v, %v, %v", latest, found, err)
	}
	if latest.New != "2" {
		t.Errorf("latest b = %q, want 2", latest.New)
	}
	latest, _, _ = log.Latest("a")
	if latest.New != "3" || latest.Old != "1" {
		t.Errorf("latest a = %q (old %q), want 3 (old 1)", latest.New, latest.Old)
	}
}

// TestAppendOnlyAcrossReopens: the log survives a process restart and nothing
// rewrites what is already there.
func TestAppendOnlyAcrossReopens(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, audit.FileName)

	first, err := audit.Open(audit.Options{Path: path, Clock: clock.NewFake(at), Subject: "op"})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	mustRecord(t, first, "engine.automation_gate.enabled", "false")

	second, err := audit.Open(audit.Options{Path: path, Clock: clock.NewFake(at.Add(time.Hour)), Subject: "op"})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	mustRecord(t, second, "engine.automation_gate.enabled", "true")

	entries, err := second.Entries()
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2 — the first process's record must survive", len(entries))
	}
	if entries[0].New != "false" || entries[1].New != "true" {
		t.Errorf("history = %q → %q, want false → true", entries[0].New, entries[1].New)
	}
}

// TestTruncatedTailDoesNotHideHistory: a half-written final line after a power
// loss must not make the preceding record unreadable.
func TestTruncatedTailDoesNotHideHistory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, audit.FileName)
	log, err := audit.Open(audit.Options{Path: path, Clock: clock.NewFake(at)})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	mustRecord(t, log, "limit", "10")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("open for append: %v", err)
	}
	if _, err := f.WriteString(`{"at":"2026-07-26T10:00:0`); err != nil {
		t.Fatalf("write partial: %v", err)
	}
	f.Close()

	entries, err := log.Entries()
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	if len(entries) != 1 || entries[0].New != "10" {
		t.Fatalf("entries = %+v, want the one complete record", entries)
	}
}

// TestFileIsOwnerOnly: the log names accounts and limits.
func TestFileIsOwnerOnly(t *testing.T) {
	log, _ := openLog(t)
	mustRecord(t, log, "limit", "10")

	info, err := os.Stat(log.Path())
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("mode = %v, want 0600", mode)
	}
}

// TestConcurrentRecordsAreAllWritten covers the -race requirement: the log is
// reachable from the engine's startup path and from whatever later changes a
// limit, and a torn line would corrupt the record either way.
func TestConcurrentRecordsAreAllWritten(t *testing.T) {
	log, _ := openLog(t)

	const writers = 8
	const each = 10
	var wg sync.WaitGroup
	wg.Add(writers)
	for w := 0; w < writers; w++ {
		go func(w int) {
			defer wg.Done()
			for i := 0; i < each; i++ {
				if err := log.Record(audit.Entry{
					Action:  audit.ActionLimitChange,
					Setting: "limit",
					New:     strings.Repeat("x", w+1),
				}); err != nil {
					t.Errorf("Record: %v", err)
					return
				}
			}
		}(w)
	}
	wg.Wait()

	entries, err := log.Entries()
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	if len(entries) != writers*each {
		t.Fatalf("entries = %d, want %d — a lost or torn line is a lost audit record",
			len(entries), writers*each)
	}
}

// TestOpenRequiresAPath: there is no implicit default, because the engine and the
// CLI resolve their data directories differently and a silent default would put
// the record somewhere neither of them looks.
func TestOpenRequiresAPath(t *testing.T) {
	if _, err := audit.Open(audit.Options{}); err == nil {
		t.Fatal("Open with no path must fail")
	}
}

// TestSubjectFallsBackToTheOSUser so an unattended change still carries a name.
func TestSubjectFallsBackToTheOSUser(t *testing.T) {
	log, err := audit.Open(audit.Options{
		Path:  filepath.Join(t.TempDir(), audit.FileName),
		Clock: clock.NewFake(at),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	mustRecord(t, log, "limit", "10")

	entries, _ := log.Entries()
	if strings.TrimSpace(entries[0].Subject) == "" {
		t.Error("an entry with no subject answers none of the questions an audit log exists for")
	}
}

func mustRecord(t *testing.T, log *audit.Log, setting, value string) {
	t.Helper()
	if _, err := log.RecordChange(audit.ActionLimitChange, setting, value, ""); err != nil {
		t.Fatalf("RecordChange(%s=%s): %v", setting, value, err)
	}
}

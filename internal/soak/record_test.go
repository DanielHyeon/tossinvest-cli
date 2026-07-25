package soak_test

// record_test.go covers the durable half: the soak runs for days, and a record
// that loses a day, or silently drops a line it cannot parse, would let an
// attestation be built on evidence nobody has.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

func sampleCycle(at time.Time) soak.Cycle {
	return soak.Cycle{
		FormatVersion: soak.RecordFormatVersion,
		Kind:          "cycle",
		StartedAt:     at,
		FinishedAt:    at.Add(time.Second),
		AccountRef:    "123-45-678901",
		Credential:    soak.Credential{OK: true, Observed: true, TokenExpiresAt: at.Add(time.Hour)},
		Endpoints: []soak.EndpointResult{
			{Endpoint: soak.EndpointAccounts, OK: true, Requests: 1},
		},
		Completeness: soak.Completeness{Evaluated: true, OK: true},
	}
}

func TestRecorderAppendsOneObjectPerLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", soak.FileName)
	rec, err := soak.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	for i := 0; i < 3; i++ {
		if err := rec.Append(sampleCycle(soakStart.Add(time.Duration(i) * time.Hour))); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the record: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("record has %d line(s), want 3", len(lines))
	}
	for i, line := range lines {
		var c soak.Cycle
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			t.Fatalf("line %d is not a JSON object: %v", i+1, err)
		}
	}
}

// TestRecorderReopensWithoutTruncating. A soak is restarted — by a reboot, by a
// user, by a crash — and a record that started over each time could never show
// three consecutive days.
func TestRecorderReopensWithoutTruncating(t *testing.T) {
	path := filepath.Join(t.TempDir(), soak.FileName)

	for i := 0; i < 2; i++ {
		rec, err := soak.OpenRecorder(path)
		if err != nil {
			t.Fatalf("OpenRecorder: %v", err)
		}
		if err := rec.Append(sampleCycle(soakStart.Add(time.Duration(i) * 24 * time.Hour))); err != nil {
			t.Fatalf("Append: %v", err)
		}
		if err := rec.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	}

	cycles, err := soak.LoadCycles(path)
	if err != nil {
		t.Fatalf("LoadCycles: %v", err)
	}
	if len(cycles) != 2 {
		t.Fatalf("loaded %d cycles, want 2 — the second run truncated the first", len(cycles))
	}
}

// TestRecorderWritesOwnerOnly. The record carries an account number and a token
// expiry; it is not a secret, but it is nobody else's business either.
func TestRecorderWritesOwnerOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not meaningful here")
	}
	path := filepath.Join(t.TempDir(), "sub", soak.FileName)
	rec, err := soak.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("record mode = %o, want 600", perm)
	}
	dir, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if perm := dir.Mode().Perm(); perm != 0o700 {
		t.Errorf("record directory mode = %o, want 700", perm)
	}
}

// TestLoadCyclesSkipsATruncatedFinalLine. A power cut during an append leaves a
// half-written last line. Everything before it is still evidence.
func TestLoadCyclesSkipsATruncatedFinalLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), soak.FileName)
	rec, err := soak.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	if err := rec.Append(sampleCycle(soakStart)); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := rec.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("reopening: %v", err)
	}
	if _, err := f.WriteString(`{"kind":"cycle","started_at":"2026-07-2`); err != nil {
		t.Fatalf("writing the partial line: %v", err)
	}
	f.Close()

	cycles, err := soak.LoadCycles(path)
	if err != nil {
		t.Fatalf("LoadCycles refused a record with a torn final line: %v", err)
	}
	if len(cycles) != 1 {
		t.Fatalf("loaded %d cycles, want the 1 complete one", len(cycles))
	}
}

// TestLoadCyclesRefusesAMalformedInteriorLine. A broken line in the middle is
// not a torn write — something else damaged the file, and quietly dropping a day
// would shorten a streak or, worse, hide a failure.
func TestLoadCyclesRefusesAMalformedInteriorLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), soak.FileName)
	good, err := json.Marshal(sampleCycle(soakStart))
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	body := string(good) + "\nnot json at all\n" + string(good) + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("writing: %v", err)
	}

	if _, err := soak.LoadCycles(path); err == nil {
		t.Fatal("LoadCycles accepted a record with a corrupt interior line")
	} else if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("err = %v, want it to name the line", err)
	}
}

// TestLoadCyclesRefusesANewerFormat, for the same reason attest.Load does:
// reading a field we do not know about as absent turns "this failed" into
// "this was never measured".
func TestLoadCyclesRefusesANewerFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), soak.FileName)
	c := sampleCycle(soakStart)
	c.FormatVersion = soak.RecordFormatVersion + 1
	line, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	if err := os.WriteFile(path, append(line, '\n'), 0o600); err != nil {
		t.Fatalf("writing: %v", err)
	}

	if _, err := soak.LoadCycles(path); err == nil {
		t.Fatal("LoadCycles accepted a record written by a newer build")
	}
}

// TestLoadCyclesOnAMissingFileIsEmptyNotAnError: "the soak has not been started"
// is a normal state and every surface asks about it.
func TestLoadCyclesOnAMissingFileIsEmptyNotAnError(t *testing.T) {
	cycles, err := soak.LoadCycles(filepath.Join(t.TempDir(), "absent.jsonl"))
	if err != nil {
		t.Fatalf("LoadCycles on a missing record: %v", err)
	}
	if len(cycles) != 0 {
		t.Fatalf("loaded %d cycles from a file that does not exist", len(cycles))
	}
}

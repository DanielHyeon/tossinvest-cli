// Package degrade counts measurement losses somewhere the thing that lost them
// cannot reach (change add-net-rr-measurement task 2.8).
//
// # Why this is a package of its own
//
// Round 3 found the round-2 disposition self-contradictory: it put the count of
// failed observation writes *in the observation table*. Disk exhaustion, an I/O
// error or a schema fault fails the observation INSERT and the degradation INSERT
// alike, because they are the same file. A counter that only works when the thing
// it counts did not happen counts nothing.
//
// So the counter writes its own file — and it lives here, alone, importing nothing
// from internal/journal. No database handle is reachable from this code, so no
// future edit can quietly route the count back through the connection whose
// failure it exists to record. That is a property of the package boundary rather
// than of anyone's care, and it is why the counter is not simply a type inside
// internal/measure (which does use internal/journal, to read decisions).
//
// Two tests hold it: internal/measure/isolation_static_test.go scans this
// package's imports, and internal/journal/observation_degradation_test.go proves
// the behaviour with a real SQLITE_FULL.
package degrade

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Loss kinds. Stable strings: they are what a count is grouped by, and renaming
// one silently resets the series.
const (
	// LossObservationWrite is an entry observation that could not be written. The
	// verdict itself stands — this is the measurement being lost, not the trade.
	LossObservationWrite = "observation_write_failed"
	// LossRefusalUnrecoverable is a refusal observation that could not be
	// written. It is counted separately because a refusal has no preimage, so
	// unlike an issued verdict it can never be rebuilt (design D6).
	LossRefusalUnrecoverable = "refusal_observation_lost"
	// LossLapsedBeyondHorizon is an issued decision whose observation was never
	// written and which has aged past the pruning horizon.
	LossLapsedBeyondHorizon = "reconstruction_lapsed"
)

// Counter records measurement losses somewhere the observation table's failure
// cannot reach.
//
// It has two tiers and is explicit about which one it is in:
//
//	durable    a small JSON file, rewritten via a temp file and a rename, in a
//	           directory the caller chose. Independent of the journal database.
//	in-process a monotonic map, when even that file cannot be written.
//
// The second tier is a real degradation and Snapshot says so. A count that
// silently became process-local would be worse than no count: an operator reading
// "3 lost observations" after a restart would believe it was three since boot,
// when it might be three since the last of an unknown number of restarts.
//
// Record never returns an error. It is the last thing standing when the storage
// already failed, so there is nobody left to hand an error to — the failure to
// count is itself logged and reflected in Snapshot.Durable.
type Counter struct {
	mu      sync.Mutex
	path    string
	log     *slog.Logger
	counts  map[string]int64
	durable bool
}

// Snapshot is the counter's current answer.
type Snapshot struct {
	// Counts is loss kind → total.
	Counts map[string]int64 `json:"counts"`
	// Durable reports whether those totals survived the last write to the
	// independent store. False means they are this process's tally only and must
	// not be reported as a total since installation.
	Durable bool `json:"-"`
}

// Total sums every kind.
func (s Snapshot) Total() int64 {
	var n int64
	for _, v := range s.Counts {
		n += v
	}
	return n
}

// Kinds lists the recorded kinds in a stable order.
func (s Snapshot) Kinds() []string {
	out := make([]string, 0, len(s.Counts))
	for k := range s.Counts {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// NewCounter opens (or starts) the counter at path.
//
// An unreadable or corrupt file is not fatal: the counter starts at zero in
// degraded mode and says so, because refusing to start would turn a measurement
// fault into an engine that will not run — the exact escalation this whole change
// is arranged to avoid.
func NewCounter(path string, log *slog.Logger) *Counter {
	if log == nil {
		log = slog.Default()
	}
	c := &Counter{path: path, log: log, counts: map[string]int64{}, durable: path != ""}
	if path == "" {
		return c
	}
	data, err := os.ReadFile(path)
	switch {
	case err != nil && !os.IsNotExist(err):
		c.durable = false
		log.Warn("the observation degradation counter could not be read; "+
			"counting continues in process memory only",
			"path", path, "error", err.Error())
	case err == nil:
		var stored Snapshot
		if err := json.Unmarshal(data, &stored); err != nil {
			c.durable = false
			log.Warn("the observation degradation counter is unreadable; "+
				"counting continues in process memory only",
				"path", path, "error", err.Error())
			break
		}
		for k, v := range stored.Counts {
			c.counts[k] = v
		}
	}
	return c
}

// Record adds one to a loss kind and persists the tally.
//
// The structured line is emitted whether or not the file write worked: it is the
// floor of the two-tier arrangement, and it is what makes a degraded counter still
// auditable from the log stream.
func (c *Counter) Record(kind string, attrs ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[kind]++
	total := c.counts[kind]

	args := append([]any{"loss_kind", kind, "count", total}, attrs...)
	c.log.Warn("an entry observation was lost; the verdict it measured is unaffected", args...)

	if c.path == "" {
		return
	}
	if err := c.persist(); err != nil {
		if c.durable {
			// Announce the transition once. After this the log line above is the
			// only durable record, which the operator needs to know.
			c.log.Warn("the observation degradation counter is no longer durable; "+
				"counts are now this process's tally only",
				"path", c.path, "error", err.Error())
		}
		c.durable = false
	}
}

// Snapshot returns the current totals.
func (c *Counter) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	counts := make(map[string]int64, len(c.counts))
	for k, v := range c.counts {
		counts[k] = v
	}
	return Snapshot{Counts: counts, Durable: c.durable}
}

// persist rewrites the file atomically. The caller holds the lock.
func (c *Counter) persist() error {
	data, err := json.Marshal(Snapshot{Counts: c.counts})
	if err != nil {
		return fmt.Errorf("encoding the degradation counts: %w", err)
	}
	dir := filepath.Dir(c.path)
	tmp, err := os.CreateTemp(dir, filepath.Base(c.path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("creating the degradation counter temp file: %w", err)
	}
	name := tmp.Name()
	defer os.Remove(name)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing the degradation counter: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing the degradation counter: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing the degradation counter: %w", err)
	}
	if err := os.Chmod(name, 0o600); err != nil {
		return fmt.Errorf("securing the degradation counter: %w", err)
	}
	if err := os.Rename(name, c.path); err != nil {
		return fmt.Errorf("replacing the degradation counter: %w", err)
	}
	return nil
}

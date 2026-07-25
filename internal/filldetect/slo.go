package filldetect

// slo.go measures fill-detection freshness and classifies the two ways it can go
// wrong.
//
// # The measurement point
//
// The fill-detection spec fixes it: from "the fill became observable at the
// broker" to "the local durable commit returned". Both ends are deliberate.
//
// Starting at the broker's execution.filledAt rather than at our own read means
// the poll interval is inside the number — a 60s poll loop cannot look fast by
// measuring only the part after it noticed. Ending at the commit rather than at
// the parse means the fsync is inside it too: an in-memory "we know" that a crash
// erases is not detection.
//
// # Why an outage is not an SLO violation
//
// A 429 or a dead transport means we did not measure anything, not that we
// measured something slow. Folding the two together would make the SLO fire for a
// condition no amount of local work fixes, and would hide the condition that
// local work does fix. They are therefore counted separately — and a failed poll
// still blocks new entries, just through the staleness path the retry matrix
// already owns rather than through this one.

import (
	"sort"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

// SLO is the fill-detection freshness objective.
type SLO struct {
	// Target is the latency the percentile must stay under.
	Target time.Duration
	// Percentile is the fraction of samples that must meet Target, in (0,1].
	Percentile float64
	// Window is how far back samples are considered.
	Window time.Duration
	// MinSamples is how many samples the window needs before a verdict is
	// possible. Below it the SLO is "unmeasured", which is not a violation:
	// blocking on one slow fill would make the engine flap.
	MinSamples int
	// Grace is how long a violation must persist before it blocks new entries.
	Grace time.Duration
}

// DefaultSLO is the provisional objective. Like the retry matrix's numbers these
// are conservative placeholders; verify-execution-capability measures the real
// ones.
//
//	Target 10s      — a 3s poll plus an 8s query budget is the worst normal case
//	p95             — one slow sample in twenty is noise, not a broken pipeline
//	Window 5m       — long enough to smooth a single bad minute
//	MinSamples 3    — a verdict from one fill is a verdict about nothing
//	Grace 30s       — ten poll intervals: a blip does not stop trading
func DefaultSLO() SLO {
	return SLO{
		Target:     10 * time.Second,
		Percentile: 0.95,
		Window:     5 * time.Minute,
		MinSamples: 3,
		Grace:      30 * time.Second,
	}
}

func (s SLO) withDefaults() SLO {
	d := DefaultSLO()
	if s.Target <= 0 {
		s.Target = d.Target
	}
	if s.Percentile <= 0 || s.Percentile > 1 {
		s.Percentile = d.Percentile
	}
	if s.Window <= 0 {
		s.Window = d.Window
	}
	if s.MinSamples <= 0 {
		s.MinSamples = d.MinSamples
	}
	if s.Grace <= 0 {
		s.Grace = d.Grace
	}
	return s
}

// SLOStatus is the current verdict.
type SLOStatus struct {
	Target     time.Duration
	Percentile float64
	// Samples is how many measurements are inside the window.
	Samples int
	// Observed is the measured percentile latency, zero when unmeasured.
	Observed time.Duration
	// Worst is the slowest sample in the window.
	Worst time.Duration
	// Violated reports Observed > Target with enough samples to say so.
	Violated bool
	// ViolatedFor is how long the violation has been continuous.
	ViolatedFor time.Duration
	// Measured reports whether the window held enough samples for a verdict.
	Measured bool
}

type sloSample struct {
	at      time.Time
	latency time.Duration
}

type sloTracker struct {
	slo SLO

	mu             sync.Mutex
	samples        []sloSample
	violatingSince time.Time
}

func newSLOTracker(slo SLO) *sloTracker {
	return &sloTracker{slo: slo.withDefaults()}
}

// observe records one measurement: a fill that became visible at the broker and
// was committed locally at at.
func (t *sloTracker) observe(at time.Time, latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if latency < 0 {
		latency = 0
	}
	t.samples = append(t.samples, sloSample{at: at, latency: latency})
}

// status prunes the window and returns the verdict, updating how long a violation
// has been running.
func (t *sloTracker) status(now time.Time) SLOStatus {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := now.Add(-t.slo.Window)
	kept := t.samples[:0]
	for _, s := range t.samples {
		if !s.at.Before(cutoff) {
			kept = append(kept, s)
		}
	}
	t.samples = kept

	out := SLOStatus{Target: t.slo.Target, Percentile: t.slo.Percentile, Samples: len(kept)}
	if len(kept) < t.slo.MinSamples {
		// Not enough evidence to say anything. A running violation ends here
		// rather than freezing at its last value: an unmeasured pipeline is the
		// staleness path's problem, not this one's.
		t.violatingSince = time.Time{}
		return out
	}
	out.Measured = true

	latencies := make([]time.Duration, len(kept))
	for i, s := range kept {
		latencies[i] = s.latency
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	out.Worst = latencies[len(latencies)-1]
	out.Observed = percentile(latencies, t.slo.Percentile)
	out.Violated = out.Observed > t.slo.Target

	switch {
	case out.Violated && t.violatingSince.IsZero():
		t.violatingSince = now
	case !out.Violated:
		t.violatingSince = time.Time{}
	}
	if !t.violatingSince.IsZero() {
		out.ViolatedFor = now.Sub(t.violatingSince)
	}
	return out
}

// percentile returns the p-th percentile of a sorted slice using the
// nearest-rank method: the smallest value at or above which p of the samples
// fall. Nearest-rank is chosen over interpolation because it always returns a
// latency that was actually observed — an interpolated "p95" that never happened
// is a poor thing to block trading on.
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	rank := int(float64(len(sorted))*p + 0.999999999)
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}

// Outage is the classification of a failing poll.
type Outage struct {
	// Active reports whether the last cycle failed.
	Active bool
	// Class is how the retry matrix classifies the failure.
	Class execgw.ErrorClass
	// Since is when the current run of failures started.
	Since time.Time
	// Consecutive counts the failures in the current run.
	Consecutive int
	// LastError is the most recent failure.
	LastError error
}

// outageTracker classifies poll failures using the retry matrix's own
// classification, so "the broker is throttling us" and "the network is down" read
// the same here as they do in the query path.
type outageTracker struct {
	mu  sync.Mutex
	cur Outage
}

func (t *outageTracker) failure(now time.Time, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	class := execgw.ClassifyQueryError(err)
	if !t.cur.Active {
		t.cur = Outage{Active: true, Since: now}
	}
	t.cur.Class = class
	t.cur.Consecutive++
	t.cur.LastError = err
}

// success ends an outage. Only a *complete* cycle calls it: a partial read is not
// evidence that the broker is reachable for everything the engine needs.
func (t *outageTracker) success(time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cur = Outage{}
}

func (t *outageTracker) status() Outage {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cur
}

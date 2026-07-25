package soak_test

// summary_test.go covers the reading of the record: how many consecutive days of
// unattended refresh it proves, which endpoints it proves, and what it does not
// prove. Everything the attestation claims is derived here, so the interesting
// cases are the ones where the answer has to be "less than you might think".

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

// cycleAt builds a fully successful cycle at an instant, then lets a test spoil
// exactly one thing about it.
func cycleAt(at time.Time, mutate func(*soak.Cycle)) soak.Cycle {
	c := soak.Cycle{
		FormatVersion: soak.RecordFormatVersion,
		Kind:          "cycle",
		StartedAt:     at,
		FinishedAt:    at.Add(2 * time.Second),
		AccountRef:    "123-45-678901",
		Credential:    soak.Credential{OK: true, Observed: true, TokenExpiresAt: at.Add(time.Hour)},
		Completeness:  soak.Completeness{Evaluated: true, OK: true},
	}
	for _, e := range soak.RequiredEndpoints() {
		c.Endpoints = append(c.Endpoints, soak.EndpointResult{
			Endpoint: e, OK: true, Requests: 1, LatencyMS: 100,
		})
	}
	if mutate != nil {
		mutate(&c)
	}
	return c
}

// threeCleanDays is the shape of a soak that has run to completion.
func threeCleanDays() []soak.Cycle {
	var out []soak.Cycle
	for day := 0; day < 3; day++ {
		for hour := 0; hour < 2; hour++ {
			at := soakStart.AddDate(0, 0, day).Add(time.Duration(hour) * 6 * time.Hour)
			out = append(out, cycleAt(at, func(c *soak.Cycle) {
				// The token expiry moves forward on the second cycle of each day:
				// that is what an unattended refresh looks like from outside.
				if day > 0 || hour > 0 {
					c.Credential.Refreshed = true
				}
			}))
		}
	}
	return out
}

func setEndpoint(c *soak.Cycle, endpoint string, mutate func(*soak.EndpointResult)) {
	for i := range c.Endpoints {
		if c.Endpoints[i].Endpoint == endpoint {
			mutate(&c.Endpoints[i])
			return
		}
	}
}

func TestSummarizeCountsConsecutiveCredentialDays(t *testing.T) {
	s := soak.Summarize(threeCleanDays())

	if s.StreakDays != 3 {
		t.Errorf("StreakDays = %d, want 3", s.StreakDays)
	}
	if len(s.Days) != 3 {
		t.Fatalf("Days = %d, want 3", len(s.Days))
	}
	for _, d := range s.Days {
		if !d.Pass {
			t.Errorf("day %s did not pass although every cycle succeeded", d.Date)
		}
	}
	if s.AccountRef != "123-45-678901" {
		t.Errorf("AccountRef = %q", s.AccountRef)
	}
	if s.All.Cycles != 6 {
		t.Errorf("All.Cycles = %d, want 6", s.All.Cycles)
	}
}

// TestSummarizeBreaksTheStreakOnAMissingDay. A soak that was not running on
// Tuesday did not prove anything about Tuesday.
func TestSummarizeBreaksTheStreakOnAMissingDay(t *testing.T) {
	cycles := []soak.Cycle{
		cycleAt(soakStart, nil),
		cycleAt(soakStart.AddDate(0, 0, 1), nil),
		// no day 2
		cycleAt(soakStart.AddDate(0, 0, 3), nil),
	}
	s := soak.Summarize(cycles)
	if s.StreakDays != 1 {
		t.Errorf("StreakDays = %d, want 1 — the gap must reset the run", s.StreakDays)
	}
}

// TestSummarizeBreaksTheStreakOnAnAuthFailure. The streak is a claim about
// unattended credential refresh specifically, so this is the failure that must
// reset it.
func TestSummarizeBreaksTheStreakOnAnAuthFailure(t *testing.T) {
	cycles := threeCleanDays()
	// The first day of three. One refused token is enough to disqualify the day,
	// even though the day's other cycle succeeded.
	cycles[0].Credential.OK = false
	cycles[0].Credential.Class = soak.ClassAuth

	s := soak.Summarize(cycles)
	if s.StreakDays != 2 {
		t.Errorf("StreakDays = %d, want 2 — the day with the refused token cannot count", s.StreakDays)
	}
	if s.Days[0].Pass {
		t.Error("the day carrying the auth failure was marked as passing")
	}
}

// TestSummarizeKeepsTheStreakThroughATransportFailure. A dropped connection is
// not a credential failure, and treating it as one would make the three-day bar
// unreachable on any real network.
func TestSummarizeKeepsTheStreakThroughATransportFailure(t *testing.T) {
	cycles := threeCleanDays()
	cycles[0].Credential.OK = false
	cycles[0].Credential.Class = soak.ClassTransport
	setEndpoint(&cycles[0], soak.EndpointAccounts, func(e *soak.EndpointResult) {
		e.OK = false
		e.Class = soak.ClassTransport
	})

	s := soak.Summarize(cycles)
	if s.StreakDays != 3 {
		t.Errorf("StreakDays = %d, want 3 — a network blip is not a refused credential", s.StreakDays)
	}
}

// TestSummarizeCountsEndpointOutcomes.
func TestSummarizeCountsEndpointOutcomes(t *testing.T) {
	cycles := threeCleanDays()
	setEndpoint(&cycles[0], soak.EndpointPrices, func(e *soak.EndpointResult) {
		e.OK = false
		e.Class = soak.ClassRateLimited
		e.Error = "official: rate limited"
		e.LatencyMS = 900
	})
	setEndpoint(&cycles[1], soak.EndpointOrderByID, func(e *soak.EndpointResult) {
		e.OK = false
		e.Skipped = true
		e.SkipReason = "the account has no orders"
	})

	s := soak.Summarize(cycles)

	prices := statFor(t, s.All.Endpoints, soak.EndpointPrices)
	if prices.Attempts != 6 || prices.Successes != 5 || prices.Failures != 1 {
		t.Errorf("prices stat = %+v, want 6 attempts / 5 successes / 1 failure", prices)
	}
	if prices.RateLimited != 1 {
		t.Errorf("prices RateLimited = %d, want 1", prices.RateLimited)
	}
	if prices.SuccessRate < 0.83 || prices.SuccessRate > 0.84 {
		t.Errorf("prices SuccessRate = %v, want ~0.833", prices.SuccessRate)
	}
	if prices.LastError == "" {
		t.Error("prices LastError is empty although a call failed")
	}

	byID := statFor(t, s.All.Endpoints, soak.EndpointOrderByID)
	if byID.Skipped != 1 {
		t.Errorf("orders/{id} Skipped = %d, want 1", byID.Skipped)
	}
	if byID.Attempts != 5 {
		t.Errorf("orders/{id} Attempts = %d, want 5 — a skip is not an attempt", byID.Attempts)
	}

	if s.All.RateLimited != 1 {
		t.Errorf("All.RateLimited = %d, want 1", s.All.RateLimited)
	}
}

// TestSummarizeReportsMedianAndTailLatency: the polling SLO in
// harden-execution-base is stated in latency, so the soak has to measure it.
func TestSummarizeReportsMedianAndTailLatency(t *testing.T) {
	var cycles []soak.Cycle
	for i := 0; i < 20; i++ {
		at := soakStart.Add(time.Duration(i) * time.Hour)
		latency := int64(100 + i*10) // 100..290
		cycles = append(cycles, cycleAt(at, func(c *soak.Cycle) {
			setEndpoint(c, soak.EndpointAccounts, func(e *soak.EndpointResult) { e.LatencyMS = latency })
		}))
	}
	s := soak.Summarize(cycles)
	stat := statFor(t, s.All.Endpoints, soak.EndpointAccounts)
	if stat.MedianLatencyMS < 190 || stat.MedianLatencyMS > 210 {
		t.Errorf("MedianLatencyMS = %d, want ~200", stat.MedianLatencyMS)
	}
	if stat.P95LatencyMS < stat.MedianLatencyMS {
		t.Errorf("P95 (%d) below the median (%d)", stat.P95LatencyMS, stat.MedianLatencyMS)
	}
}

// TestSummarizeReportsEveryDistinctAccount. Two accounts in one record means the
// measurements describe neither of them, and the attestation names exactly one.
func TestSummarizeReportsEveryDistinctAccount(t *testing.T) {
	cycles := threeCleanDays()
	cycles[4].AccountRef = "987-65-432109"

	s := soak.Summarize(cycles)
	if len(s.AccountRefs) != 2 {
		t.Fatalf("AccountRefs = %v, want both accounts", s.AccountRefs)
	}
}

// TestSummarizeScopesTheWindowToTheStreak. What the attestation claims is what
// the streak proved, so the window's statistics must exclude everything before
// it — including a failure from a run that was later restarted.
func TestSummarizeScopesTheWindowToTheStreak(t *testing.T) {
	old := cycleAt(soakStart.AddDate(0, 0, -10), func(c *soak.Cycle) {
		c.Completeness.OK = false
		c.Completeness.Detail = "ancient history"
	})
	cycles := append([]soak.Cycle{old}, threeCleanDays()...)

	s := soak.Summarize(cycles)
	if s.StreakDays != 3 {
		t.Fatalf("StreakDays = %d, want 3", s.StreakDays)
	}
	if s.All.CompletenessFailures != 1 {
		t.Errorf("All.CompletenessFailures = %d, want 1", s.All.CompletenessFailures)
	}
	if s.Window.CompletenessFailures != 0 {
		t.Errorf("Window.CompletenessFailures = %d, want 0 — the failure predates the streak",
			s.Window.CompletenessFailures)
	}
	if !s.WindowStart.Equal(soakStart.Truncate(24 * time.Hour)) {
		t.Errorf("WindowStart = %s, want the first day of the streak", s.WindowStart)
	}
}

// TestSummarizeOnAnEmptyRecord.
func TestSummarizeOnAnEmptyRecord(t *testing.T) {
	s := soak.Summarize(nil)
	if s.StreakDays != 0 || s.All.Cycles != 0 || len(s.Days) != 0 {
		t.Errorf("empty summary = %+v, want zeroes", s)
	}
}

func statFor(t *testing.T, stats []soak.EndpointStat, endpoint string) soak.EndpointStat {
	t.Helper()
	for _, s := range stats {
		if s.Endpoint == endpoint {
			return s
		}
	}
	t.Fatalf("no statistics for %s", endpoint)
	return soak.EndpointStat{}
}

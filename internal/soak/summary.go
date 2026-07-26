package soak

// summary.go reads the record back and turns it into the few numbers a decision
// is made on.
//
// The one that matters is StreakDays: how many consecutive calendar days ended
// with the credentials still working, with nobody having touched them. Two rules
// shape it and both are deliberately strict:
//
//	a missing day breaks it   a soak that was not running on Tuesday proved
//	                          nothing about Tuesday
//	an auth failure breaks it but a transport or server failure does not — the
//	                          claim is about credentials, and a dropped
//	                          connection is not a refused token
//
// Everything is measured twice: over the whole record (All), which is what an
// operator wants to see while it runs, and over the streak window (Window),
// which is what the attestation is allowed to claim. A failure from a run that
// was later restarted must not be attested away, and a failure from before the
// current streak must not block it forever.

import (
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
)

// Day is one calendar day (UTC) of the soak.
type Day struct {
	// Date is the day, as YYYY-MM-DD in UTC.
	Date string `json:"date"`
	// Cycles is how many cycles ran that day.
	Cycles int `json:"cycles"`
	// CredentialOK is how many of them authenticated successfully.
	CredentialOK int `json:"credential_ok"`
	// AuthFailures is how many were refused. One is enough to fail the day.
	AuthFailures int `json:"auth_failures"`
	// Pass reports that the day counts toward the streak.
	Pass bool `json:"pass"`
}

// EndpointStat aggregates one endpoint over a set of cycles.
type EndpointStat struct {
	// Endpoint is the call, as "METHOD /path".
	Endpoint string `json:"endpoint"`
	// Attempts counts the cycles in which the call was actually made.
	Attempts int `json:"attempts"`
	// Successes and Failures partition Attempts.
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
	// Skipped counts the cycles in which there was nothing to call.
	Skipped int `json:"skipped"`
	// RateLimited counts 429s — the input to the retry-matrix work in task 1.3.
	RateLimited int `json:"rate_limited"`
	// SuccessRate is Successes/Attempts, or 0 when nothing was attempted.
	SuccessRate float64 `json:"success_rate"`
	// MedianLatencyMS and P95LatencyMS describe the successful calls.
	MedianLatencyMS int64 `json:"median_latency_ms"`
	P95LatencyMS    int64 `json:"p95_latency_ms"`
	// LastError is the most recent failure message.
	LastError string `json:"last_error,omitempty"`
	// LastSuccessAt is when the call last worked.
	LastSuccessAt time.Time `json:"last_success_at,omitempty"`
}

// Stats aggregates a set of cycles.
type Stats struct {
	// Cycles is how many cycles are in the set.
	Cycles int `json:"cycles"`
	// Endpoints is one entry per endpoint seen, ordered by RequiredEndpoints.
	Endpoints []EndpointStat `json:"endpoints"`
	// CompletenessFailures counts evaluated cycles whose completeness check
	// failed.
	CompletenessFailures int `json:"completeness_failures"`
	// LastCompletenessDetail is the most recent such failure's detail.
	LastCompletenessDetail string `json:"last_completeness_detail,omitempty"`
	// RateLimited is the total number of throttled calls.
	RateLimited int `json:"rate_limited"`
	// TokenObservations is how many cycles could read the token expiry at all.
	TokenObservations int `json:"token_observations"`
	// TokenRefreshes is how many cycles saw the expiry move forward.
	TokenRefreshes int `json:"token_refreshes"`
	// MaxSustainedRatePerSecond is the highest request rate observed in a cycle
	// that was never throttled. It is a lower bound on what the API tolerates,
	// not a measured ceiling: the soak is a survey, not a load test.
	MaxSustainedRatePerSecond float64 `json:"max_sustained_rate_per_second"`
}

// Summary is the whole record, read.
type Summary struct {
	// FirstAt and LastAt bound the record.
	FirstAt time.Time `json:"first_at"`
	LastAt  time.Time `json:"last_at"`

	// AccountRef is the account the record describes, and AccountRefs is every
	// distinct one it contains. More than one entry means the record describes no
	// single account and cannot be attested.
	AccountRef  string   `json:"account_ref"`
	AccountRefs []string `json:"account_refs"`

	// Days is every calendar day with at least one cycle, oldest first.
	Days []Day `json:"days"`
	// StreakDays is the run of consecutive passing days ending at the last day.
	StreakDays int `json:"streak_days"`
	// WindowStart is midnight UTC on the streak's first day.
	WindowStart time.Time `json:"window_start"`

	// All covers the whole record; Window covers the streak only.
	All    Stats `json:"all"`
	Window Stats `json:"window"`

	// Binary is the executable the survey said it was running, as of the most
	// recent cycle that recorded one.
	//
	// It is the survey's own statement about itself, which is what makes it worth
	// having: a reader can tell whether the process appending to this file is the
	// build that is installed without scraping a process table or guessing at a
	// pid. A record written before the survey stamped itself leaves it zero, which
	// reads as "not known" everywhere and never as "out of date".
	Binary binstamp.Stamp `json:"binary,omitzero"`
}

// Summarize reads a set of cycles.
func Summarize(cycles []Cycle) Summary {
	var s Summary
	if len(cycles) == 0 {
		return s
	}

	ordered := append([]Cycle(nil), cycles...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].StartedAt.Before(ordered[j].StartedAt) })

	s.FirstAt = ordered[0].StartedAt.UTC()
	s.LastAt = ordered[len(ordered)-1].StartedAt.UTC()
	s.AccountRefs = distinctAccounts(ordered)
	if len(s.AccountRefs) > 0 {
		// The most recent one: if the operator moved the soak to another account
		// the conflict is reported separately, and the newest is the one they
		// meant.
		s.AccountRef = latestAccount(ordered)
	}

	for i := len(ordered) - 1; i >= 0; i-- {
		if ordered[i].Binary.Known() {
			s.Binary = ordered[i].Binary
			break
		}
	}

	s.Days = daysOf(ordered)
	s.StreakDays = streakOf(s.Days)
	if s.StreakDays > 0 {
		first := s.Days[len(s.Days)-s.StreakDays].Date
		if t, err := time.Parse("2006-01-02", first); err == nil {
			s.WindowStart = t.UTC()
		}
	}

	s.All = statsOf(ordered)
	s.Window = statsOf(inWindow(ordered, s.WindowStart, s.StreakDays > 0))
	return s
}

func inWindow(cycles []Cycle, start time.Time, have bool) []Cycle {
	if !have {
		return nil
	}
	var out []Cycle
	for _, c := range cycles {
		if !c.StartedAt.UTC().Before(start) {
			out = append(out, c)
		}
	}
	return out
}

func distinctAccounts(cycles []Cycle) []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range cycles {
		ref := strings.TrimSpace(c.AccountRef)
		if ref == "" || seen[ref] {
			continue
		}
		seen[ref] = true
		out = append(out, ref)
	}
	sort.Strings(out)
	return out
}

func latestAccount(cycles []Cycle) string {
	for i := len(cycles) - 1; i >= 0; i-- {
		if ref := strings.TrimSpace(cycles[i].AccountRef); ref != "" {
			return ref
		}
	}
	return ""
}

// daysOf groups cycles by UTC date and decides which days count.
func daysOf(cycles []Cycle) []Day {
	byDate := map[string]*Day{}
	var order []string
	for _, c := range cycles {
		date := c.StartedAt.UTC().Format("2006-01-02")
		d, ok := byDate[date]
		if !ok {
			d = &Day{Date: date}
			byDate[date] = d
			order = append(order, date)
		}
		d.Cycles++
		if c.Credential.OK {
			d.CredentialOK++
		}
		if c.Credential.Class == ClassAuth {
			d.AuthFailures++
		}
	}
	sort.Strings(order)

	out := make([]Day, 0, len(order))
	for _, date := range order {
		d := byDate[date]
		// A day counts when the credentials worked at least once and were never
		// refused. "Never refused" is the strict half: one ClassAuth means the
		// unattended renewal failed that day, whatever happened afterwards.
		d.Pass = d.CredentialOK > 0 && d.AuthFailures == 0
		out = append(out, *d)
	}
	return out
}

// streakOf counts consecutive passing days ending at the most recent one. A
// calendar gap breaks the run even if both sides passed.
func streakOf(days []Day) int {
	streak := 0
	var next time.Time
	for i := len(days) - 1; i >= 0; i-- {
		d := days[i]
		if !d.Pass {
			break
		}
		at, err := time.Parse("2006-01-02", d.Date)
		if err != nil {
			break
		}
		if streak > 0 && !at.AddDate(0, 0, 1).Equal(next) {
			break // a missing day: the soak was not running, so nothing was proven
		}
		streak++
		next = at
	}
	return streak
}

// statsOf aggregates a set of cycles.
func statsOf(cycles []Cycle) Stats {
	stats := Stats{Cycles: len(cycles)}
	if len(cycles) == 0 {
		return stats
	}

	type acc struct {
		stat      EndpointStat
		latencies []int64
	}
	byEndpoint := map[string]*acc{}
	var order []string

	for _, c := range cycles {
		if c.Credential.Observed {
			stats.TokenObservations++
		}
		if c.Credential.Refreshed {
			stats.TokenRefreshes++
		}
		if c.Completeness.Evaluated && !c.Completeness.OK {
			stats.CompletenessFailures++
			stats.LastCompletenessDetail = c.Completeness.Detail
		}

		requests, throttled := 0, false
		for _, e := range c.Endpoints {
			a, ok := byEndpoint[e.Endpoint]
			if !ok {
				a = &acc{stat: EndpointStat{Endpoint: e.Endpoint}}
				byEndpoint[e.Endpoint] = a
				order = append(order, e.Endpoint)
			}
			requests += e.Requests
			switch {
			case e.Skipped:
				a.stat.Skipped++
			case e.OK:
				a.stat.Attempts++
				a.stat.Successes++
				a.stat.LastSuccessAt = c.StartedAt.UTC()
				a.latencies = append(a.latencies, e.LatencyMS)
			default:
				a.stat.Attempts++
				a.stat.Failures++
				if e.Error != "" {
					a.stat.LastError = e.Error
				}
			}
			if e.Class == ClassRateLimited {
				a.stat.RateLimited++
				stats.RateLimited++
				throttled = true
			}
		}

		if !throttled && requests > 0 {
			if elapsed := c.FinishedAt.Sub(c.StartedAt).Seconds(); elapsed > 0 {
				if rate := float64(requests) / elapsed; rate > stats.MaxSustainedRatePerSecond {
					stats.MaxSustainedRatePerSecond = rate
				}
			}
		}
	}

	// Ordered by RequiredEndpoints first so the operator's view is stable, then
	// anything else the record happens to carry.
	sort.Slice(order, func(i, j int) bool {
		ri, rj := requiredIndex(order[i]), requiredIndex(order[j])
		if ri != rj {
			return ri < rj
		}
		return order[i] < order[j]
	})

	for _, endpoint := range order {
		a := byEndpoint[endpoint]
		if a.stat.Attempts > 0 {
			a.stat.SuccessRate = float64(a.stat.Successes) / float64(a.stat.Attempts)
		}
		a.stat.MedianLatencyMS = percentile(a.latencies, 0.50)
		a.stat.P95LatencyMS = percentile(a.latencies, 0.95)
		stats.Endpoints = append(stats.Endpoints, a.stat)
	}
	return stats
}

func requiredIndex(endpoint string) int {
	for i, e := range RequiredEndpoints() {
		if e == endpoint {
			return i
		}
	}
	return len(RequiredEndpoints())
}

// percentile returns the p-th percentile of values by nearest rank. It sorts a
// copy: the caller's slice is accumulated in arrival order and a later reader
// may want it that way.
func percentile(values []int64, p float64) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(p * float64(len(sorted)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// SuccessfulEndpoints returns the endpoints that succeeded at least once in the
// set, sorted. This is the set an attestation may claim, and nothing else.
func (s Stats) SuccessfulEndpoints() []string {
	var out []string
	for _, e := range s.Endpoints {
		if e.Successes > 0 {
			out = append(out, e.Endpoint)
		}
	}
	sort.Strings(out)
	return out
}

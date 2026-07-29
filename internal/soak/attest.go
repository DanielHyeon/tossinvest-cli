package soak

// attest.go turns a finished soak into the file the engine's automation gate is
// interlocked on — or, far more often, explains why it will not.
//
// The asymmetry is the design. Writing an attestation that overstates what was
// measured has a cost measured in somebody's account; refusing to write one that
// was almost good enough costs a day of waiting. So every check here fails
// closed, the endpoint set is exactly what succeeded and nothing else, and there
// is no flag anywhere in this package that relaxes a criterion.

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
)

// ErrIncomplete means the soak has not proven enough to be attested. It wraps
// the individual reasons so a caller can print them.
var ErrIncomplete = errors.New("soak: the capability soak is not complete")

// Criteria is the bar a soak has to clear.
type Criteria struct {
	// MinConsecutiveDays is how many consecutive days of unattended credential
	// refresh are required. The spec says three.
	MinConsecutiveDays int
	// RequiredEndpoints must each have succeeded at least once inside the window.
	RequiredEndpoints []string
	// Validity is how long the resulting attestation is trusted for.
	Validity time.Duration
	// MaxRecordAge is how old the record's newest cycle may be.
	//
	// Without it, a soak that finished in March would still mint a
	// freshly-dated attestation in July: the streak is a property of the record,
	// and the expiry is a property of the moment the file is written. The
	// attestation says "this account's API behaved like this recently", and a
	// record nobody has added to for weeks cannot support the word recently.
	MaxRecordAge time.Duration
}

// DefaultCriteria is the execution-verification spec's bar.
//
// The 30-day validity is a judgement, not a spec number: the measurements are
// about a broker that can change its rate limits or its token lifetime without
// telling anybody, and a month is short enough that a stale attestation is
// noticed while the account is still recognisable.
func DefaultCriteria() Criteria {
	return Criteria{
		MinConsecutiveDays: 3,
		RequiredEndpoints:  RequiredEndpoints(),
		Validity:           30 * 24 * time.Hour,
		MaxRecordAge:       48 * time.Hour,
	}
}

func (c Criteria) withDefaults() Criteria {
	d := Criteria{
		MinConsecutiveDays: 3,
		RequiredEndpoints:  RequiredEndpoints(),
		Validity:           30 * 24 * time.Hour,
		MaxRecordAge:       48 * time.Hour,
	}
	if c.MinConsecutiveDays <= 0 {
		c.MinConsecutiveDays = d.MinConsecutiveDays
	}
	if len(c.RequiredEndpoints) == 0 {
		c.RequiredEndpoints = d.RequiredEndpoints
	}
	if c.Validity <= 0 {
		c.Validity = d.Validity
	}
	if c.MaxRecordAge <= 0 {
		c.MaxRecordAge = d.MaxRecordAge
	}
	return c
}

// Evaluate reports whether the soak has finished, and every reason it has not.
//
// It returns all the reasons rather than the first, because each one costs the
// operator a day to find out about separately.
func (s Summary) Evaluate(now time.Time, c Criteria) (bool, []string) {
	c = c.withDefaults()
	var reasons []string

	if s.All.Cycles == 0 {
		return false, []string{"the soak record is empty — run `tossctl soak run` first"}
	}

	if age := now.UTC().Sub(s.LastAt); age > c.MaxRecordAge {
		reasons = append(reasons, fmt.Sprintf(
			"the newest cycle in the record is %s old and at most %s is accepted; the attestation would be "+
				"dated today while describing an API nobody has looked at since %s — restart `tossctl soak run`",
			humanDays(age), humanDays(c.MaxRecordAge), s.LastAt.Format("2006-01-02")))
	}

	if len(s.AccountRefs) > 1 {
		reasons = append(reasons, fmt.Sprintf(
			"the record covers %d accounts (%s); an attestation names one account, so start a fresh "+
				"record for the account you mean to trade",
			len(s.AccountRefs), strings.Join(maskAll(s.AccountRefs), ", ")))
	}
	if strings.TrimSpace(s.AccountRef) == "" {
		reasons = append(reasons, "no account was resolved from the credentials, so there is nothing to attest about")
	}

	if s.StreakDays < c.MinConsecutiveDays {
		reasons = append(reasons, fmt.Sprintf(
			"unattended credential refresh is proven for %d consecutive day(s); %d are required",
			s.StreakDays, c.MinConsecutiveDays))
	}

	switch {
	case s.Window.TokenObservations < 2:
		reasons = append(reasons, fmt.Sprintf(
			"the access-token expiry was readable in %d cycle(s) of the window; without at least two "+
				"observations an unattended refresh cannot be told from a long-lived token",
			s.Window.TokenObservations))
	case s.Window.TokenRefreshes == 0:
		reasons = append(reasons,
			"no token refresh was observed: the expiry never moved forward, so the soak has watched a "+
				"token that never needed renewing rather than a credential that renews itself")
	}

	for _, want := range c.RequiredEndpoints {
		if statOf(s.Window.Endpoints, want).Successes == 0 {
			reasons = append(reasons, endpointReason(want, statOf(s.Window.Endpoints, want)))
		}
	}

	if s.Window.CompletenessFailures > 0 {
		detail := s.Window.LastCompletenessDetail
		if detail == "" {
			detail = "see the record"
		}
		reasons = append(reasons, fmt.Sprintf(
			"%d cycle(s) inside the window failed the read-completeness check (%s)",
			s.Window.CompletenessFailures, detail))
	}

	return len(reasons) == 0, reasons
}

// endpointReason explains a required endpoint that never worked, including the
// one case where the reason is not a failure at all.
func endpointReason(endpoint string, stat EndpointStat) string {
	if stat.Attempts == 0 && stat.Skipped > 0 && endpoint == EndpointOrderByID {
		return fmt.Sprintf(
			"%s was never exercised because the account had no orders to read. The engine calls it, so it "+
				"has to be proven: the supervised live check (verify-execution-capability task 2.2) places "+
				"and immediately cancels one minimum-size limit order, which leaves an order the soak can read",
			endpoint)
	}
	if stat.Attempts == 0 {
		return fmt.Sprintf("%s was never exercised inside the window", endpoint)
	}
	last := stat.LastError
	if last == "" {
		last = "no successful call"
	}
	return fmt.Sprintf("%s never succeeded inside the window (%d attempt(s); last error: %s)",
		endpoint, stat.Attempts, last)
}

// BuildAttestation produces the attestation for a finished soak.
//
// verifiedBy names who ran it and notes is free-form operator context; both end
// up in the audit trail the engine prints at startup.
func BuildAttestation(s Summary, c Criteria, now time.Time, verifiedBy, notes string,
	supervised []attest.Proof) (attest.Attestation, error) {
	c = c.withDefaults()

	ok, reasons := s.Evaluate(now, c)
	if !ok {
		return attest.Attestation{}, fmt.Errorf("%w:\n  - %s", ErrIncomplete, strings.Join(reasons, "\n  - "))
	}

	endpoints := s.Window.SuccessfulEndpoints()
	// A read-only tool cannot have exercised a write, so a write in this set means
	// the record was edited or the survey grew a method it should not have. Either
	// way the attestation must not carry it: the engine reads this set as "these
	// calls were proven to work".
	for _, e := range endpoints {
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(e)), "GET ") {
			return attest.Attestation{}, fmt.Errorf(
				"soak: refusing to attest %q — the soak issues no mutating request, so a non-GET endpoint in "+
					"its record did not come from a measurement", e)
		}
	}

	accepted, err := acceptSupervised(s.AccountRef, now, c.Validity, supervised)
	if err != nil {
		return attest.Attestation{}, err
	}
	for _, p := range accepted {
		endpoints = append(endpoints, p.Endpoint)
	}

	note := fmt.Sprintf(
		"read-only capability soak: %d cycle(s) over %d consecutive day(s) (%s..%s), %d token refresh(es) "+
			"observed, %d throttled call(s). %s",
		s.Window.Cycles, s.StreakDays,
		s.WindowStart.UTC().Format("2006-01-02"), s.LastAt.UTC().Format("2006-01-02"),
		s.Window.TokenRefreshes, s.Window.RateLimited,
		mutationNote(accepted))
	if strings.TrimSpace(notes) != "" {
		note = strings.TrimSpace(notes) + " | " + note
	}

	return attest.Attestation{
		FormatVersion:      attest.FormatVersion,
		AccountRef:         s.AccountRef,
		IssuedAt:           now.UTC(),
		ExpiresAt:          now.UTC().Add(c.Validity),
		SoakDays:           s.StreakDays,
		Endpoints:          endpoints,
		RateLimitPerSecond: s.Window.MaxSustainedRatePerSecond,
		VerifiedBy:         strings.TrimSpace(verifiedBy),
		Notes:              note,
		SupervisedBy:       accepted,
	}, nil
}

// acceptSupervised decides which of the supervised proofs may be attested.
//
// The set it can draw from is closed to LiveOnlyEndpoints() — the calls this
// read-only tool structurally cannot make. That closure is the whole safety
// argument, and it cuts two ways:
//
//	a read           is refused. One supervised success is not what days of
//	                 unattended operation prove, and accepting it here would let
//	                 a supervised run launder a soak that failed the same read.
//	another mutation is refused. The verification record also exercises
//	                 conditional orders; the gate does not require them, and an
//	                 attestation that claims what nobody asked for is a claim
//	                 waiting to be relied on.
//
// It is the mirror of the refusal above it: that one keeps mutations out of the
// soak's own evidence, this one keeps everything else out of the supervised
// evidence. Together they make each source prove only its own half.
//
// Two failure modes are told apart deliberately:
//
//	wrong account  is an error. A record for a different account sitting on the
//	               expected path is a misconfiguration, and skipping it would make
//	               that indistinguishable from having no evidence at all.
//	stale/future   is a skip. Evidence from months ago is an ordinary state — the
//	               check was run and then time passed — and the gate's own
//	               MissingEndpoints already reports the consequence.
func acceptSupervised(accountRef string, now time.Time, validity time.Duration,
	supervised []attest.Proof) ([]attest.Proof, error) {

	allowed := map[string]bool{}
	for _, e := range LiveOnlyEndpoints() {
		allowed[normaliseEndpoint(e)] = true
	}

	var accepted []attest.Proof
	seen := map[string]bool{}
	for _, p := range supervised {
		endpoint := strings.Join(strings.Fields(strings.TrimSpace(p.Endpoint)), " ")
		if endpoint == "" {
			continue
		}
		key := normaliseEndpoint(endpoint)
		if !allowed[key] {
			return nil, fmt.Errorf(
				"soak: refusing to attest %q from the supervised record — only %s may come from there, "+
					"because the reads are what the unattended soak proves and a single supervised success "+
					"does not establish the same property",
				endpoint, strings.Join(LiveOnlyEndpoints(), ", "))
		}
		if !attest.SameAccountMasked(accountRef, p.AccountRef) {
			return nil, fmt.Errorf(
				"soak: refusing to attest %q — the supervised evidence names account %s but this soak is "+
					"for %s. A record for another account on the expected path is a misconfiguration, not "+
					"an absence of evidence",
				endpoint, p.AccountRef, attest.Mask(accountRef))
		}
		age := now.Sub(p.At)
		if age < 0 || age >= validity {
			continue
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		p.Endpoint = endpoint
		accepted = append(accepted, p)
	}
	return accepted, nil
}

// mutationNote is the sentence an operator reads before deciding whether the gate
// can be turned on: which mutations are covered, and which are not.
func mutationNote(accepted []attest.Proof) string {
	have := map[string]bool{}
	for _, p := range accepted {
		have[normaliseEndpoint(p.Endpoint)] = true
	}
	var covered, missing []string
	for _, e := range LiveOnlyEndpoints() {
		if have[normaliseEndpoint(e)] {
			covered = append(covered, e)
		} else {
			missing = append(missing, e)
		}
	}
	switch {
	case len(missing) == 0:
		return fmt.Sprintf("Mutation endpoints covered by the supervised live check: %s.",
			strings.Join(covered, ", "))
	case len(covered) == 0:
		return fmt.Sprintf("Mutation endpoints are NOT covered: %s remain to be proven by the "+
			"supervised live check (verify-execution-capability task 2.2).", strings.Join(missing, ", "))
	default:
		return fmt.Sprintf("Mutation endpoints covered by the supervised live check: %s. Still NOT "+
			"covered: %s.", strings.Join(covered, ", "), strings.Join(missing, ", "))
	}
}

func normaliseEndpoint(e string) string {
	return strings.Join(strings.Fields(strings.ToUpper(strings.TrimSpace(e))), " ")
}

// MissingForEngine returns the endpoints an engine requires that this
// attestation does not carry, so the operator is told what is still outstanding
// rather than discovering it when the engine refuses to start.
func MissingForEngine(a attest.Attestation, engineRequired []string) []string {
	return a.MissingEndpoints(engineRequired)
}

// --- operator output --------------------------------------------------------

// WriteText renders the progress summary `tossctl soak status` prints.
func (s Summary) WriteText(w io.Writer, now time.Time, c Criteria) {
	c = c.withDefaults()

	if s.All.Cycles == 0 {
		fmt.Fprintln(w, "No soak has been recorded yet. Start one with `tossctl soak run`.")
		return
	}

	fmt.Fprintf(w, "account            %s\n", attest.Mask(s.AccountRef))
	if len(s.AccountRefs) > 1 {
		fmt.Fprintf(w, "  ⚠ the record also covers %s\n", strings.Join(maskAll(s.AccountRefs), ", "))
	}
	fmt.Fprintf(w, "cycles             %d (%s .. %s)\n", s.All.Cycles,
		s.FirstAt.Format(time.RFC3339), s.LastAt.Format(time.RFC3339))
	fmt.Fprintf(w, "elapsed            %s\n", humanDays(s.LastAt.Sub(s.FirstAt)))
	fmt.Fprintf(w, "credential streak  %d of %d consecutive day(s)\n", s.StreakDays, c.MinConsecutiveDays)
	fmt.Fprintf(w, "token refreshes    %d observed in %d readable cycle(s)\n",
		s.All.TokenRefreshes, s.All.TokenObservations)

	fmt.Fprintln(w, "\ndays")
	for _, d := range s.Days {
		mark := "pass"
		if !d.Pass {
			mark = "FAIL"
		}
		fmt.Fprintf(w, "  %s  %-4s  cycles %-4d credential ok %-4d auth failures %d\n",
			d.Date, mark, d.Cycles, d.CredentialOK, d.AuthFailures)
	}

	fmt.Fprintln(w, "\nendpoints (whole record)")
	fmt.Fprintf(w, "  %-28s %8s %8s %8s %9s %9s\n", "endpoint", "ok", "fail", "429", "p50 ms", "p95 ms")
	for _, e := range s.All.Endpoints {
		fmt.Fprintf(w, "  %-28s %8d %8d %8d %9d %9d\n",
			e.Endpoint, e.Successes, e.Failures, e.RateLimited, e.MedianLatencyMS, e.P95LatencyMS)
		if e.Skipped > 0 {
			fmt.Fprintf(w, "  %-28s skipped in %d cycle(s)\n", "", e.Skipped)
		}
		if e.LastError != "" {
			fmt.Fprintf(w, "  %-28s last error: %s\n", "", e.LastError)
		}
	}

	fmt.Fprintf(w, "\ncompleteness       %d failure(s) in the record, %d inside the window\n",
		s.All.CompletenessFailures, s.Window.CompletenessFailures)
	if s.All.LastCompletenessDetail != "" {
		fmt.Fprintf(w, "  last: %s\n", s.All.LastCompletenessDetail)
	}
	if s.All.MaxSustainedRatePerSecond > 0 {
		fmt.Fprintf(w, "observed rate      %.2f req/s sustained without a 429 (a lower bound, not a ceiling)\n",
			s.All.MaxSustainedRatePerSecond)
	}

	ok, reasons := s.Evaluate(now, c)
	fmt.Fprintln(w)
	if ok {
		fmt.Fprintln(w, "READY — run `tossctl soak attest` to write the capability attestation.")
		return
	}
	fmt.Fprintln(w, "NOT READY — no attestation will be written:")
	for _, r := range reasons {
		fmt.Fprintf(w, "  - %s\n", r)
	}
}

func statOf(stats []EndpointStat, endpoint string) EndpointStat {
	for _, s := range stats {
		if s.Endpoint == endpoint {
			return s
		}
	}
	return EndpointStat{Endpoint: endpoint}
}

func maskAll(refs []string) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, attest.Mask(r))
	}
	sort.Strings(out)
	return out
}

func humanDays(d time.Duration) string {
	if d <= 0 {
		return "0h"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if days == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dd %dh", days, hours)
}

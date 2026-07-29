// Package attest reads and verifies the capability attestation that the engine's
// automation gate is interlocked on.
//
// # What an attestation is for
//
// TossOS's automation gate is a claim that unattended order placement is safe on
// this machine, for this account, right now. That claim rests on things a build
// cannot check for itself: that credentials refresh unattended for days at a
// time, that the observed rate limits leave room for the polling loop, that the
// order/fill/balance reads are actually complete. The separate change
// `verify-execution-capability` measures those and, only if every one of them
// passed, writes the result here as a durable local file.
//
// This package is the reading half of that contract. It is deliberately in a
// package of its own rather than inside the engine: the writer (a soak tool that
// must contain no mutation transport at all) and the reader (the engine) need the
// same format, and a shared format that lives in the consumer is a format that
// drifts.
//
// # Why verification is total, not partial
//
// Verify returns on the first failure and never returns a "mostly valid"
// attestation. Every field it checks answers a question whose wrong answer is a
// live-account incident:
//
//	expired          the measurements are stale; the broker may have changed
//	wrong account    the soak proved something about a different portfolio
//	missing endpoint the engine needs a call nobody ever proved works
//
// An unreadable, unparseable or absent file is the same as a failed one. There is
// no path through this package that treats "we could not check" as "it is fine".
package attest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// FileName is the attestation's default name inside the config directory.
const FileName = "capability-attestation.json"

// FormatVersion is the attestation format this build writes and understands.
//
// A newer file is refused rather than partially read, for the same reason the
// journal refuses a newer schema: reading a field we do not know about as absent
// would turn an attestation that says "US orders were never verified" into one
// that says nothing at all.
const FormatVersion = 1

// Errors verification can produce. They are sentinels because the engine's
// startup message differs per cause — an operator who sees "expired" needs to
// re-run the soak, one who sees "wrong account" has a configuration mistake.
var (
	// ErrMissing means there is no attestation file.
	ErrMissing = errors.New("attest: no capability attestation")
	// ErrMalformed means the file exists but could not be read as one.
	ErrMalformed = errors.New("attest: capability attestation is malformed")
	// ErrFormatTooNew means the file was written by a newer TossOS.
	ErrFormatTooNew = errors.New("attest: capability attestation format is newer than this build")
	// ErrExpired means the attestation's validity window has passed.
	ErrExpired = errors.New("attest: capability attestation has expired")
	// ErrAccountMismatch means the attestation is about a different account.
	ErrAccountMismatch = errors.New("attest: capability attestation is for a different account")
	// ErrEndpointNotAttested means a required endpoint is not in the proven set.
	ErrEndpointNotAttested = errors.New("attest: capability attestation does not cover a required endpoint")
	// ErrIncomplete means a mandatory field is empty.
	ErrIncomplete = errors.New("attest: capability attestation is incomplete")
)

// Attestation is the durable record verify-execution-capability produces.
type Attestation struct {
	// FormatVersion is the format of this file.
	FormatVersion int `json:"format_version"`

	// AccountRef is the account the measurements were taken against. The engine
	// compares this to the account its credentials actually resolve to.
	AccountRef string `json:"account_ref"`

	// IssuedAt is when the verification completed.
	IssuedAt time.Time `json:"issued_at"`
	// ExpiresAt is when the measurements stop being trusted. The engine refuses
	// to start with the gate on past this instant.
	ExpiresAt time.Time `json:"expires_at"`

	// SoakDays is how many consecutive days of unattended credential refresh were
	// observed.
	SoakDays int `json:"soak_days"`

	// Endpoints is the set of broker calls that were exercised successfully,
	// written as "METHOD /path".
	Endpoints []string `json:"endpoints"`

	// RateLimitPerSecond is the observed sustainable request rate, if measured.
	RateLimitPerSecond float64 `json:"rate_limit_per_second,omitempty"`

	// VerifiedBy names who ran the verification, for the audit trail.
	VerifiedBy string `json:"verified_by,omitempty"`
	// Notes is free-form operator context.
	Notes string `json:"notes,omitempty"`

	// SupervisedBy records what proved each endpoint the read-only soak cannot
	// exercise. Additive and optional: an attestation written before this field
	// existed reads unchanged, so FormatVersion does not move.
	//
	// It exists because the interlock's refusal says what is *missing* and
	// nothing says what a gate that started was allowed to start on. An auditor
	// a month later is asking exactly that.
	SupervisedBy []Proof `json:"supervised_by,omitempty"`
}

// Proof is one endpoint and the supervised measurement that established it.
//
// The read-only soak proves reads by running unattended for days. It structurally
// cannot prove a write — it contains no mutation transport — so those come from
// the supervised one-off live check, where a person approved every request. This
// type is what carries that second kind of evidence, and it lives here because
// both the writer and the reader already depend on this package and neither
// should have to depend on the other.
type Proof struct {
	// Endpoint is "METHOD /path", spelled the way internal/soak,
	// internal/verifylive and internal/app/engine all spell it.
	Endpoint string `json:"endpoint"`
	// At is when the call succeeded.
	At time.Time `json:"at"`
	// AccountRef is the account the evidence names, as its record wrote it —
	// which for the verification record is masked. Carried per proof rather than
	// once, so the account check is local to the evidence it is checking.
	AccountRef string `json:"account_ref"`
	// Source is where the evidence was read from, and Market which market's
	// record it was. An operator reading this a month later needs to be able to
	// go and look at the file.
	Source string `json:"source,omitempty"`
	Market string `json:"market,omitempty"`
}

// SameAccountMasked reports whether a reference written in full and one that may
// have been masked describe the same account.
//
// The soak record keeps the account unmasked; the verification record never does
// (record.go: "the unmasked number never enters this file"). Binding the two is
// therefore comparing a number with its own mask, and doing it in one place stops
// a caller from inventing a second rule.
func SameAccountMasked(full, other string) bool {
	if strings.TrimSpace(full) == "" || strings.TrimSpace(other) == "" {
		return false
	}
	return sameAccount(full, other) || Mask(full) == strings.TrimSpace(other)
}

// Load reads an attestation from disk.
//
// A missing file is ErrMissing rather than a wrapped os error, because "there is
// no attestation" is a normal, expected state (it is what every machine looks
// like before the verification is run) and the caller's message for it is
// different from the one for a corrupt file.
func Load(path string) (Attestation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Attestation{}, fmt.Errorf("%w at %s", ErrMissing, path)
		}
		return Attestation{}, fmt.Errorf("%w: reading %s: %v", ErrMalformed, path, err)
	}

	var a Attestation
	if err := json.Unmarshal(data, &a); err != nil {
		return Attestation{}, fmt.Errorf("%w: parsing %s: %v", ErrMalformed, path, err)
	}
	if a.FormatVersion > FormatVersion {
		return Attestation{}, fmt.Errorf("%w: %s is format %d, this build understands %d",
			ErrFormatTooNew, path, a.FormatVersion, FormatVersion)
	}
	return a, nil
}

// Save writes an attestation atomically, owner-only.
//
// It exists here so the verification tool and the engine cannot disagree about
// the format. It is not used by the engine.
func Save(path string, a Attestation) error {
	if a.FormatVersion == 0 {
		a.FormatVersion = FormatVersion
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("attest: encoding the attestation: %w", err)
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("attest: writing %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("attest: publishing %s: %w", path, err)
	}
	return nil
}

// Verify checks an attestation against the engine's actual situation.
//
// accountRef is the account the engine's credentials resolve to — not the one the
// config names, because the point of the check is to catch the case where those
// two differ. required lists the endpoints the engine will actually call; an
// empty list skips that check.
func (a Attestation) Verify(now time.Time, accountRef string, required []string) error {
	switch {
	case strings.TrimSpace(a.AccountRef) == "":
		return fmt.Errorf("%w: it names no account", ErrIncomplete)
	case a.ExpiresAt.IsZero():
		return fmt.Errorf("%w: it has no expiry, and an attestation without one cannot be trusted indefinitely",
			ErrIncomplete)
	case len(a.Endpoints) == 0:
		return fmt.Errorf("%w: it lists no verified endpoints", ErrIncomplete)
	}

	if !now.Before(a.ExpiresAt) {
		return fmt.Errorf("%w: it expired at %s and it is now %s — re-run the capability verification "+
			"(`verify-execution-capability`) before enabling the automation gate",
			ErrExpired, a.ExpiresAt.UTC().Format(time.RFC3339), now.UTC().Format(time.RFC3339))
	}

	if !sameAccount(a.AccountRef, accountRef) {
		return fmt.Errorf("%w: it attests %s but the credentials resolve to %s",
			ErrAccountMismatch, Mask(a.AccountRef), Mask(accountRef))
	}

	if missing := a.MissingEndpoints(required); len(missing) > 0 {
		return fmt.Errorf("%w: %s were never verified — re-run the capability verification covering them",
			ErrEndpointNotAttested, strings.Join(missing, ", "))
	}
	return nil
}

// MissingEndpoints returns the required endpoints the attestation does not cover,
// sorted, so the message an operator sees is stable.
func (a Attestation) MissingEndpoints(required []string) []string {
	if len(required) == 0 {
		return nil
	}
	have := make(map[string]bool, len(a.Endpoints))
	for _, e := range a.Endpoints {
		have[normaliseEndpoint(e)] = true
	}
	var missing []string
	for _, want := range required {
		if !have[normaliseEndpoint(want)] {
			missing = append(missing, want)
		}
	}
	sort.Strings(missing)
	return missing
}

// Expired reports whether the attestation's window has passed.
func (a Attestation) Expired(now time.Time) bool {
	return a.ExpiresAt.IsZero() || !now.Before(a.ExpiresAt)
}

// ExpiresWithin reports whether the attestation lapses inside d. The alerting
// path (task 4.3) treats an imminent expiry as a critical event, because an
// attestation that lapses while the engine is running stops it at the next
// restart, and finding that out at the next restart is finding it out too late.
func (a Attestation) ExpiresWithin(now time.Time, d time.Duration) bool {
	if a.ExpiresAt.IsZero() {
		return true
	}
	return now.Add(d).After(a.ExpiresAt)
}

// Mask renders an account reference for logs and error messages.
//
// Account numbers end up in error output, alerts and support transcripts. Showing
// the last four digits is enough for an operator to tell two of their own
// accounts apart and not enough for a reader of the log to use one.
func Mask(ref string) string {
	trimmed := strings.TrimSpace(ref)
	if trimmed == "" {
		return "(none)"
	}
	digits := []rune(trimmed)
	if len(digits) <= 4 {
		return strings.Repeat("*", len(digits))
	}
	return strings.Repeat("*", len(digits)-4) + string(digits[len(digits)-4:])
}

// sameAccount compares account references tolerantly of formatting: brokers write
// the same account as "12345678" and "123-45-678901" depending on the endpoint,
// and a hyphen must not be the reason an engine refuses to start.
func sameAccount(a, b string) bool {
	return accountDigits(a) != "" && accountDigits(a) == accountDigits(b)
}

func accountDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normaliseEndpoint(e string) string {
	return strings.Join(strings.Fields(strings.ToUpper(strings.TrimSpace(e))), " ")
}

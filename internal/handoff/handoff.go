// Package handoff carries one already-authenticated browser session across a
// console restart (openspec change verify-execution-capability, task 1.8 ①).
//
// # The problem it exists for, stated narrowly
//
// `tossctl console` authenticates by possession of the terminal that started it:
// a per-process session token is printed in a URL and exchanged for a cookie —
// valid for that process's lifetime, and not the single-use thing this package
// mints. The
// restart button re-executes the binary, so the new process mints a new session
// token — and the browser, which is where the operator actually is, holds a cookie
// that is now worthless and has no way to learn the new one without going back to
// the terminal. Task 1.8's whole premise is that the operator never goes back to
// the terminal.
//
// So the old process leaves exactly one thing behind for the new one: a token in a
// file, good once, good briefly.
//
// # Why this is not a weakening of the console's authentication
//
// Four properties, and all four are load-bearing:
//
//	only an authenticated session mints one   Mint is reached from a POST that has
//	                                          already cleared the session cookie and
//	                                          the CSRF token. There is no other caller.
//	the file is owner-only, in an owner-only  0600 inside the data directory the
//	directory                                 rest of this change already writes
//	                                          0700. Reading it means being the user.
//	one consumption, ever                     Consume removes the file before it
//	                                          answers. A replay finds nothing.
//	seconds, not sessions                     TTL is a window for a browser to follow
//	                                          a redirect, not a window for anything
//	                                          to be discovered and used.
//
// The terminal remains the root of trust: the new process still prints its own
// session URL, and an operator who misses the handoff has lost nothing but a click.
//
// # It holds no session and grants no approval
//
// What crosses is "this browser was talking to the console a moment ago". The new
// process still mints its own session token, its own CSRF token and — for anything
// that reaches a live account — its own expiring nonce that a person types. Nothing
// about an approval survives a restart, by design: `Console.spent` resetting is the
// point of the restart, and the batch approval is asked for again from scratch.
package handoff

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileName is the token's name. It sits beside the console's other state — the
// evidence record and the soak pause marker — so an isolated --config-dir profile
// gets its own.
const FileName = "console-handoff.json"

// TTL is how long a minted token stays usable.
//
// Two minutes. The sequence it has to cover is a re-exec, a socket rebind and a
// browser following one redirect, which is a second or two on any machine that can
// run this at all; the rest is slack for a laptop that decided to swap. It is not
// long enough for the token to be worth finding, and the file is gone the moment it
// is used either way.
const TTL = 2 * time.Minute

// Errors a consumption can produce. They are distinguished because they mean
// different things to the operator: nothing to hand over is the ordinary case for a
// console that was started by hand, and a mismatch is not.
var (
	// ErrNoHandoff means there is no token waiting. A console started from a
	// terminal sees this on every request and it is not a problem.
	ErrNoHandoff = errors.New("handoff: no restart handoff is waiting")
	// ErrExpired means the window closed. Nothing was granted.
	ErrExpired = errors.New("handoff: the restart handoff expired")
	// ErrMismatch means the presented token is not the one that was minted.
	ErrMismatch = errors.New("handoff: the restart handoff token does not match")
)

// Store is one handoff file.
type Store struct {
	path string
	ttl  time.Duration
}

// New returns the store at path.
func New(path string) *Store { return &Store{path: path, ttl: TTL} }

// WithTTL returns a store with a different window. It exists for the tests; there
// is no flag and no option that reaches it.
func (s *Store) WithTTL(ttl time.Duration) *Store {
	return &Store{path: s.path, ttl: ttl}
}

// Path reports where the token lives.
func (s *Store) Path() string { return s.path }

// record is the file's contents. It carries no session token, no cookie and
// nothing about the account: only the one string and when it stops working.
type record struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	MintedAt  time.Time `json:"minted_at"`
}

// Mint writes a fresh single-use token and returns it.
//
// A token already waiting is overwritten rather than refused: two restarts in a row
// is an operator pressing a button twice, and the older token becoming worthless is
// the correct outcome.
func (s *Store) Mint(now time.Time) (string, error) {
	if strings.TrimSpace(s.path) == "" {
		return "", errors.New("handoff: no path was configured")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return "", fmt.Errorf("handoff: creating %s: %w", filepath.Dir(s.path), err)
	}

	token := newToken()
	body, err := json.Marshal(record{
		Token:     token,
		ExpiresAt: now.UTC().Add(s.ttl),
		MintedAt:  now.UTC(),
	})
	if err != nil {
		return "", fmt.Errorf("handoff: encoding the token: %w", err)
	}

	// O_TRUNC rather than an atomic rename: the file is tiny, it is written once,
	// and a torn write can only produce a token nothing will match — which fails
	// closed, into the ordinary "open the URL the terminal printed" path.
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("handoff: writing %s: %w", s.path, err)
	}
	if _, err := f.Write(append(body, '\n')); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("handoff: writing %s: %w", s.path, err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("handoff: syncing %s: %w", s.path, err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("handoff: closing %s: %w", s.path, err)
	}
	// The mode is re-asserted because an existing file keeps its own.
	if err := os.Chmod(s.path, 0o600); err != nil {
		return "", fmt.Errorf("handoff: securing %s: %w", s.path, err)
	}
	return token, nil
}

// Consume answers whether presented is the waiting token, and spends it.
//
// The file is removed before the comparison is answered, so every outcome — match,
// mismatch, expiry — leaves nothing behind to present a second time. That is
// stricter than "refuse a reuse" and it is the right direction: a token that a
// wrong guess did not invalidate is a token somebody can keep guessing at.
func (s *Store) Consume(presented string, now time.Time) error {
	if strings.TrimSpace(s.path) == "" {
		return ErrNoHandoff
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNoHandoff
		}
		return fmt.Errorf("handoff: reading %s: %w", s.path, err)
	}
	// Spend it whatever it turns out to be.
	if rmErr := os.Remove(s.path); rmErr != nil && !os.IsNotExist(rmErr) {
		// A token that cannot be removed is a token that could be replayed, so it
		// is refused rather than honoured.
		return fmt.Errorf("handoff: the token at %s could not be spent, so it is refused: %w", s.path, rmErr)
	}

	var rec record
	if err := json.Unmarshal(data, &rec); err != nil {
		return ErrMismatch
	}
	if strings.TrimSpace(rec.Token) == "" {
		return ErrMismatch
	}
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(presented)), []byte(rec.Token)) != 1 {
		return ErrMismatch
	}
	if !now.UTC().Before(rec.ExpiresAt) {
		return ErrExpired
	}
	return nil
}

// Discard removes a waiting token without judging it.
//
// A console that starts and finds a handoff nobody came back for should not leave
// it lying about for the rest of the day.
func (s *Store) Discard() {
	if strings.TrimSpace(s.path) == "" {
		return
	}
	_ = os.Remove(s.path)
}

// newToken returns a URL-safe random string. 20 bytes, base32 without padding: the
// same shape the console's own session token uses, and nothing types it by hand.
func newToken() string {
	var buf [20]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("handoff: crypto/rand is unavailable: " + err.Error())
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf[:])
}

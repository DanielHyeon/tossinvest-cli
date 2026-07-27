package main

// verify_account_test.go covers the account identity read (task 1.7 ③).
//
// It is one GET, and on 2026-07-26 it was the most expensive one in the run: the
// verification read /api/v1/accounts to name the account, then threw the sequence
// number away, and the official client resolved it again lazily on the first
// account-scoped GET of every step that needed it — into a 429 that cost three
// steps (measurements.md M4).
//
// Two properties, then: the sequence survives the first read, and a 429 on it is
// retried under the same policy the runner's reads use.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// accountsStub is the narrow surface resolveVerifyAccount reads.
type accountsStub struct {
	accounts []domain.Account
	// failures is how many calls answer with err before the accounts come back.
	failures int
	err      error
	calls    int
}

func (s *accountsStub) Accounts(context.Context) ([]domain.Account, error) {
	s.calls++
	if s.calls <= s.failures {
		return nil, s.err
	}
	return s.accounts, nil
}

// noSleep records the backoff without spending it.
func noSleep(waited *[]time.Duration) func(context.Context, time.Duration) error {
	return func(ctx context.Context, d time.Duration) error {
		*waited = append(*waited, d)
		return ctx.Err()
	}
}

// TestTheAccountSequenceComesBackWithTheReference, from the same entry, so the
// record cannot name one account while the header selects another.
func TestTheAccountSequenceComesBackWithTheReference(t *testing.T) {
	stub := &accountsStub{accounts: []domain.Account{{ID: "7", DisplayName: "123-45-678901"}}}

	ref, seq, err := resolveVerifyAccount(context.Background(), stub, noSleep(nil))
	if err != nil {
		t.Fatalf("resolveVerifyAccount: %v", err)
	}
	if ref != "123-45-678901" {
		t.Errorf("ref = %q, want the broker's account number", ref)
	}
	if seq != 7 {
		t.Errorf("seq = %d, want 7 — the lazy resolution exists to work this out and it is already here", seq)
	}
	if stub.calls != 1 {
		t.Errorf("the accounts endpoint was read %d time(s), want 1", stub.calls)
	}
}

// TestAnUnnumberedAccountFallsBackToTheLazyResolution. A zero sequence is not an
// error: the client resolves it the old way, and the verification proceeds.
func TestAnUnnumberedAccountFallsBackToTheLazyResolution(t *testing.T) {
	for _, id := range []string{"", "not-a-number", "0", "-3"} {
		stub := &accountsStub{accounts: []domain.Account{{ID: id, DisplayName: "123-45-678901"}}}
		ref, seq, err := resolveVerifyAccount(context.Background(), stub, noSleep(nil))
		if err != nil {
			t.Fatalf("id %q: resolveVerifyAccount: %v", id, err)
		}
		if ref == "" {
			t.Errorf("id %q: the account reference was lost", id)
		}
		if seq != 0 {
			t.Errorf("id %q: seq = %d, want 0 so the client resolves it itself", id, seq)
		}
	}
}

// TestTheIdentityReadIsRetriedOnA429 — this is the call that was rate limited,
// and failing it costs the whole run before a person has been asked anything.
func TestTheIdentityReadIsRetriedOnA429(t *testing.T) {
	var waited []time.Duration
	stub := &accountsStub{
		accounts: []domain.Account{{ID: "7", DisplayName: "123-45-678901"}},
		failures: 2,
		err:      official.ErrRateLimited,
	}
	sleep := func(_ context.Context, d time.Duration) error {
		waited = append(waited, d)
		return nil
	}

	ref, seq, err := resolveVerifyAccount(context.Background(), stub, sleep)
	if err != nil {
		t.Fatalf("resolveVerifyAccount: %v", err)
	}
	if ref == "" || seq != 7 {
		t.Errorf("ref = %q, seq = %d after the rate limit cleared", ref, seq)
	}
	if stub.calls != 3 {
		t.Errorf("the accounts endpoint was read %d time(s), want 3", stub.calls)
	}
	if len(waited) != 2 {
		t.Fatalf("the retry waited %d time(s), want 2", len(waited))
	}
	for i, d := range waited {
		if want := verifylive.ReadRetryBackoff(i); d != want {
			t.Errorf("wait %d was %s, want %s — the policy is verifylive's, not a second copy", i, d, want)
		}
	}
}

// TestTheIdentityReadGivesUpOnAPersistentRateLimit, and says what happened.
func TestTheIdentityReadGivesUpOnAPersistentRateLimit(t *testing.T) {
	stub := &accountsStub{failures: 99, err: official.ErrRateLimited}
	_, _, err := resolveVerifyAccount(context.Background(), stub,
		func(context.Context, time.Duration) error { return nil })
	if err == nil {
		t.Fatal("resolveVerifyAccount succeeded against a broker that never answered")
	}
	if !errors.Is(err, official.ErrRateLimited) {
		t.Errorf("err = %v, want the rate limit to survive the wrapping", err)
	}
	if want := 1 + verifylive.ReadRetryExtraAttempts; stub.calls != want {
		t.Errorf("the accounts endpoint was read %d time(s), want %d", stub.calls, want)
	}
}

// TestAnErrorThatIsNotARateLimitIsNotRetried.
func TestAnErrorThatIsNotARateLimitIsNotRetried(t *testing.T) {
	stub := &accountsStub{failures: 99, err: errors.New("official: auth")}
	if _, _, err := resolveVerifyAccount(context.Background(), stub,
		func(context.Context, time.Duration) error { return nil }); err == nil {
		t.Fatal("resolveVerifyAccount succeeded on an auth failure")
	}
	if stub.calls != 1 {
		t.Errorf("the accounts endpoint was read %d time(s), want 1", stub.calls)
	}
}

// TestAnAccountWithNoNumberIsAHardFailure. There is nothing to attest about.
func TestAnAccountWithNoNumberIsAHardFailure(t *testing.T) {
	stub := &accountsStub{accounts: []domain.Account{{ID: "7"}}}
	_, _, err := resolveVerifyAccount(context.Background(), stub, noSleep(nil))
	if err == nil {
		t.Fatal("resolveVerifyAccount accepted an account with no number")
	}
	if !strings.Contains(err.Error(), "no account number") {
		t.Errorf("err = %v, want it to say the broker named no account", err)
	}
}

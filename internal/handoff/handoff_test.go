package handoff

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func store(t *testing.T) *Store {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "state", FileName))
}

var now = time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)

// TestTheLifecycleIsMintConsumeRefuse is the contract in one test: a token works
// once and never again.
func TestTheLifecycleIsMintConsumeRefuse(t *testing.T) {
	s := store(t)

	token, err := s.Mint(now)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if token == "" {
		t.Fatal("Mint returned an empty token")
	}

	if err := s.Consume(token, now.Add(time.Second)); err != nil {
		t.Fatalf("the freshly minted token was refused: %v", err)
	}

	if err := s.Consume(token, now.Add(2*time.Second)); !errors.Is(err, ErrNoHandoff) {
		t.Fatalf("replaying the token gave %v, want ErrNoHandoff — a single-use token was used twice", err)
	}
	if _, err := os.Stat(s.Path()); !os.IsNotExist(err) {
		t.Errorf("the token file is still on disk after being consumed (%v)", err)
	}
}

// TestAnExpiredTokenIsRefusedAndStillSpent.
func TestAnExpiredTokenIsRefusedAndStillSpent(t *testing.T) {
	s := store(t)
	token, err := s.Mint(now)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}

	if err := s.Consume(token, now.Add(TTL)); !errors.Is(err, ErrExpired) {
		t.Fatalf("Consume at exactly the expiry gave %v, want ErrExpired", err)
	}
	if _, err := os.Stat(s.Path()); !os.IsNotExist(err) {
		t.Error("an expired token was left on disk; it must be spent whatever the verdict")
	}
}

// TestAWrongGuessSpendsTheToken.
//
// Stricter than the task's "refused on reuse", and deliberately so: a token a wrong
// guess leaves in place is a token somebody can keep guessing at, and the cost of
// the strict direction is one click on a console the operator is already sitting at.
func TestAWrongGuessSpendsTheToken(t *testing.T) {
	s := store(t)
	token, err := s.Mint(now)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}

	if err := s.Consume("NOT-THE-TOKEN", now.Add(time.Second)); !errors.Is(err, ErrMismatch) {
		t.Fatalf("a wrong token gave %v, want ErrMismatch", err)
	}
	if err := s.Consume(token, now.Add(2*time.Second)); !errors.Is(err, ErrNoHandoff) {
		t.Fatalf("the real token survived a wrong guess (%v); the guess must spend it", err)
	}
}

// TestNothingWaitingIsTheOrdinaryCase: a console started from a terminal sees this
// on every request and it must not look like a failure.
func TestNothingWaitingIsTheOrdinaryCase(t *testing.T) {
	s := store(t)
	if err := s.Consume("anything", now); !errors.Is(err, ErrNoHandoff) {
		t.Fatalf("Consume with no file gave %v, want ErrNoHandoff", err)
	}
}

// TestTheFileIsOwnerOnly. The token is the only thing standing between a stale
// browser tab and a session, for the two minutes it lives.
func TestTheFileIsOwnerOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes")
	}
	s := store(t)
	if _, err := s.Mint(now); err != nil {
		t.Fatalf("Mint: %v", err)
	}

	info, err := os.Stat(s.Path())
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("the handoff file is %04o, want 0600", perm)
	}
	dir, err := os.Stat(filepath.Dir(s.Path()))
	if err != nil {
		t.Fatalf("Stat on the directory: %v", err)
	}
	if perm := dir.Mode().Perm(); perm&0o077 != 0 {
		t.Errorf("the handoff directory is %04o; it must not be readable by anybody else", perm)
	}
}

// TestMintingAgainInvalidatesTheOlderToken — pressing restart twice.
func TestMintingAgainInvalidatesTheOlderToken(t *testing.T) {
	s := store(t)
	first, err := s.Mint(now)
	if err != nil {
		t.Fatalf("first Mint: %v", err)
	}
	second, err := s.Mint(now.Add(time.Second))
	if err != nil {
		t.Fatalf("second Mint: %v", err)
	}
	if first == second {
		t.Fatal("two mints produced the same token")
	}
	if err := s.Consume(first, now.Add(2*time.Second)); !errors.Is(err, ErrMismatch) {
		t.Fatalf("the superseded token gave %v, want ErrMismatch", err)
	}
}

// TestDiscardClearsAnUnclaimedToken.
func TestDiscardClearsAnUnclaimedToken(t *testing.T) {
	s := store(t)
	if _, err := s.Mint(now); err != nil {
		t.Fatalf("Mint: %v", err)
	}
	s.Discard()
	if _, err := os.Stat(s.Path()); !os.IsNotExist(err) {
		t.Error("Discard left the token on disk")
	}
	s.Discard() // twice is not an error
}

// TestGarbageInTheFileIsAMismatchAndNotAPanic.
func TestGarbageInTheFileIsAMismatchAndNotAPanic(t *testing.T) {
	s := store(t)
	if _, err := s.Mint(now); err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if err := os.WriteFile(s.Path(), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("corrupting the file: %v", err)
	}
	if err := s.Consume("anything", now); !errors.Is(err, ErrMismatch) {
		t.Fatalf("a torn file gave %v, want ErrMismatch", err)
	}
}

// TestTheWindowIsShort guards the constant itself: a handoff that lasted an hour
// would be a second, weaker session token with a long life.
func TestTheWindowIsShort(t *testing.T) {
	if TTL > 5*time.Minute {
		t.Errorf("TTL is %s; the handoff covers a redirect, not a working session", TTL)
	}
}

// TestAConfiguredWindowIsHonoured, so the console's tests can drive an expiry
// without sleeping.
func TestAConfiguredWindowIsHonoured(t *testing.T) {
	s := store(t).WithTTL(time.Second)
	token, err := s.Mint(now)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if err := s.Consume(token, now.Add(2*time.Second)); !errors.Is(err, ErrExpired) {
		t.Fatalf("Consume gave %v, want ErrExpired", err)
	}
}

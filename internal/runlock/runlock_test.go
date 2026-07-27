package runlock

// runlock_test.go pins the two behaviours the soak depends on and the one that
// stops this file from becoming a footgun: a held lock reads fresh, a released
// one is gone, and an abandoned one expires on its own.

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempLock(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), FileName)
}

func TestAHeldLockReadsFresh(t *testing.T) {
	path := tempLock(t)
	now := time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)

	if fresh, _ := Fresh(path, now, StaleAfter); fresh {
		t.Fatal("a lock that was never taken reads fresh")
	}

	lock, err := Acquire(path, now)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	fresh, at := Fresh(path, now, StaleAfter)
	if !fresh {
		t.Error("a lock taken this instant does not read fresh")
	}
	if !at.Equal(now) {
		t.Errorf("the lock's timestamp is %s, want %s — Fresh must read the clock it was written with", at, now)
	}

	lock.Release()
	if fresh, _ := Fresh(path, now, StaleAfter); fresh {
		t.Error("a released lock still reads fresh")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("the lock file survived Release: %v", err)
	}
}

// TestAnAbandonedLockGoesStale is the property that stops a crashed verification
// from wedging a multi-day soak.
func TestAnAbandonedLockGoesStale(t *testing.T) {
	path := tempLock(t)
	taken := time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)
	if _, err := Acquire(path, taken); err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	cases := []struct {
		after time.Duration
		fresh bool
	}{
		{0, true},
		{StaleAfter - time.Second, true},
		{StaleAfter, true},
		{StaleAfter + time.Second, false},
		{24 * time.Hour, false},
	}
	for _, tc := range cases {
		if got, _ := Fresh(path, taken.Add(tc.after), StaleAfter); got != tc.fresh {
			t.Errorf("%s after it was taken: fresh = %v, want %v", tc.after, got, tc.fresh)
		}
	}
}

// TestRefreshKeepsALongRunAlive: a verification is minutes of waiting on a
// person, and it must not go stale while somebody reads the batch summary.
func TestRefreshKeepsALongRunAlive(t *testing.T) {
	path := tempLock(t)
	taken := time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)
	lock, err := Acquire(path, taken)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	later := taken.Add(20 * time.Minute)
	if fresh, _ := Fresh(path, later, StaleAfter); fresh {
		t.Fatal("the lock did not go stale, so the refresh proves nothing")
	}
	if err := lock.Refresh(later); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if fresh, _ := Fresh(path, later, StaleAfter); !fresh {
		t.Error("a refreshed lock does not read fresh")
	}
}

// TestRefreshAfterReleaseDoesNotResurrectTheFile. The refresh loop and the
// release race by construction; the loser must not recreate a lock nobody holds.
func TestRefreshAfterReleaseDoesNotResurrectTheFile(t *testing.T) {
	path := tempLock(t)
	now := time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)
	lock, err := Acquire(path, now)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	lock.Release()
	if err := lock.Refresh(now); err != nil {
		t.Fatalf("Refresh after Release: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Refresh recreated a released lock")
	}
	lock.Release() // idempotent
}

// TestAFutureTimestampReadsFresh: a clock that disagrees with itself must make
// the soak yield, not charge ahead.
func TestAFutureTimestampReadsFresh(t *testing.T) {
	path := tempLock(t)
	now := time.Date(2026, 7, 27, 9, 0, 0, 0, time.UTC)
	if _, err := Acquire(path, now.Add(time.Hour)); err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	if fresh, _ := Fresh(path, now, StaleAfter); !fresh {
		t.Error("a lock timestamped in the future reads stale; the safe reading is fresh")
	}
}

func TestHoldTakesAndReleases(t *testing.T) {
	path := tempLock(t)
	release, err := Hold(context.Background(), path)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	if fresh, _ := Fresh(path, time.Now(), StaleAfter); !fresh {
		t.Error("Hold did not take the lock")
	}
	release()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Hold's release left the lock behind")
	}
	release() // idempotent
}

// TestHoldReportsAnUnwritableLocationWithoutRefusing. The caller is a live
// verification; a directory it cannot write to is worth a line of output and
// nothing more.
func TestHoldReportsAnUnwritableLocationWithoutRefusing(t *testing.T) {
	dir := t.TempDir()
	// A file where the lock's directory should be: MkdirAll cannot make it.
	blocker := filepath.Join(dir, "blocked")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("writing the blocker: %v", err)
	}
	release, err := Hold(context.Background(), filepath.Join(blocker, FileName))
	if err == nil {
		t.Error("Hold reported no error for an unwritable location")
	}
	if release == nil {
		t.Fatal("Hold returned a nil release; a caller must be able to defer it unconditionally")
	}
	release()
}

// TestFreshIgnoresADirectoryOrAMissingParent. Every failure direction is "carry
// on": the reader is a survey that must not be stopped by this package.
func TestFreshIgnoresAMissingPath(t *testing.T) {
	if fresh, _ := Fresh(filepath.Join(t.TempDir(), "nope", FileName), time.Now(), StaleAfter); fresh {
		t.Error("a missing lock reads fresh")
	}
}

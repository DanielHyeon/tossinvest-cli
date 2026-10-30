//go:build linux

package official

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
)

func TestA166TokenNilContextMatchesOff(t *testing.T) {
	var got []error
	for _, tr := range []http.RoundTripper{&http.Transport{}, &accountbudget.Transport{}} {
		m := newTokenManager(Credentials{}, "https://example.invalid", "", &http.Client{Transport: tr})
		_, err := m.exchange(context.Context(nil))
		if err == nil || !errors.Is(err, ErrTransport) {
			t.Fatalf("nil context err=%v", err)
		}
		got = append(got, err)
	}
	if got[0].Error() != got[1].Error() {
		t.Fatalf("OFF/required differ: %v / %v", got[0], got[1])
	}
}

func TestA166AccountsNilContextMatchesOff(t *testing.T) {
	cache := filepath.Join(t.TempDir(), "absent-token")
	off := New(Credentials{}, cache)
	guarded, err := NewWithAccountBudget(Credentials{}, cache, a166Budget(t), func(error) {})
	if err != nil {
		t.Fatal(err)
	}
	_, offErr := off.Accounts(nil)
	_, guardedErr := guarded.Accounts(nil)
	if offErr == nil || guardedErr == nil || !errors.Is(offErr, ErrTransport) || !errors.Is(guardedErr, ErrTransport) || offErr.Error() != guardedErr.Error() {
		t.Fatalf("OFF=%v guarded=%v", offErr, guardedErr)
	}
}

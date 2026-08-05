package engine_test

// a085: the alerts a person actually receives.
//
// The live outbox row this replaces:
//
//	042660 is held with no entry decision, so the exit policy will not manage it
//
// and the two contracts that must survive the translation: the payload stays
// machine-readable, and no alert leaks the account.

import (
	"strings"
	"testing"
	"unicode"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func TestTheUnmanagedAlertNamesTheStockInKorean(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = config.Adoption{} // adoption off: the holding is folded and reported
	})
	h.holds("042660", "2", "89500", 91800)
	h.holdings.items[0].Name = "한화오션"

	if cycle := h.cycle(); cycle.Err != nil {
		t.Fatalf("cycle: %v", cycle.Err)
	}

	alert, ok := h.alerts.first(obs.EventExitPositionUnmanaged)
	if !ok {
		t.Fatalf("no unmanaged alert; events = %v", eventTypes(h.alerts.events))
	}
	if !strings.Contains(alert.Title, "한화오션(042660)") {
		t.Errorf("title = %q, want 이름(코드); a person reading this on a phone has to know "+
			"which holding stopped being protected", alert.Title)
	}
	// The Hangul syllable range, not a 14-syllable sample: a reword that happens
	// to use none of the sampled syllables is plainly Korean and would fail a test
	// about nothing it changed.
	if !containsHangul(alert.Title + alert.Body) {
		t.Errorf("title/body is not Korean: %q / %q", alert.Title, alert.Body)
	}

	// The machine-readable half is untouched: the ledger is queried by code, and
	// what a display name was on the day an alert was written must not matter.
	if got := alert.Fields[obs.FieldSymbol]; got != "042660" {
		t.Errorf("Fields[symbol] = %v, want the raw code", got)
	}
	for key := range alert.Fields {
		for _, r := range key {
			if r > 127 {
				t.Errorf("field key %q is not ASCII; the payload is parsed, not read", key)
				break
			}
		}
	}

	// The prose a085 writes carries no account reference.
	//
	// This is deliberately NOT labelled as the §0.8 guarantee it originally
	// claimed. obs.Notifier appends the structured Fields to the *published*
	// notification body, and Fields[account] carries the broker account number, so
	// the account does reach the ntfy topic — through a path this assertion never
	// looks at. That exposure predates a085, lives in internal/obs, and is
	// recorded in issues.md B2 for its own change. A test that names a guarantee
	// it does not check is worse than one that names less.
	if strings.Contains(alert.Title+alert.Body, reconcileAccount) {
		t.Errorf("a085's own prose leaks the account reference: %q / %q", alert.Title, alert.Body)
	}
}

func TestAnUnnamedStockRendersAsItsCodeInTheAlert(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = config.Adoption{}
	})
	h.holds("042660", "2", "89500", 91800) // no Name from the broker

	if cycle := h.cycle(); cycle.Err != nil {
		t.Fatalf("cycle: %v", cycle.Err)
	}
	alert, ok := h.alerts.first(obs.EventExitPositionUnmanaged)
	if !ok {
		t.Fatalf("no unmanaged alert; events = %v", eventTypes(h.alerts.events))
	}
	if !strings.Contains(alert.Title, "042660") {
		t.Errorf("title = %q, want the code", alert.Title)
	}
	// A name, not a parenthesis: forbidding every "(" in the title would fail on a
	// market or quantity parenthetical that involves no name at all.
	var names engine.InstrumentNames
	if got := names.Label("042660"); strings.Contains(alert.Title, got+"(") {
		t.Errorf("title = %q, want no invented name", alert.Title)
	}
	if containsHangul(strings.SplitN(alert.Title, " ", 2)[0]) {
		t.Errorf("title = %q, want the bare code where no name is known", alert.Title)
	}
}

// containsHangul reports a Hangul syllable anywhere in s (U+AC00–U+D7A3).
func containsHangul(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Hangul, r) {
			return true
		}
	}
	return false
}

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
	if !strings.ContainsAny(alert.Title+alert.Body, "가나다라마바사아자차카타파하") {
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

	// §0.8: the account never rides along in an alert an operator's phone receives.
	if strings.Contains(alert.Title+alert.Body, reconcileAccount) {
		t.Errorf("the alert leaks the account reference: %q / %q", alert.Title, alert.Body)
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
	if strings.Contains(alert.Title, "(") {
		t.Errorf("title = %q, want no invented name", alert.Title)
	}
}

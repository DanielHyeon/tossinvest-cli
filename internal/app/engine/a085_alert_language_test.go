package engine_test

// a085: an alert says which stock, in Korean.
//
// The live outbox on 2026-08-05 read:
//
//	title: the exit policy could not judge 032820
//	body:  the stored protection state or the observed price is not usable, so this
//	       position is not being judged at all: …
//
// A person reading that on a phone cannot tell which holding stopped being
// protected. The console has called the same holding by name since a044; only
// the alerts were still speaking English in stock codes.

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
)

func TestALabelNamesTheStockAndKeepsTheCode(t *testing.T) {
	var names engine.InstrumentNames
	names.Learn("042660", "한화오션")

	if got := names.Label("042660"); got != "한화오션(042660)" {
		t.Errorf("Label = %q, want 한화오션(042660)", got)
	}
	// The code is never dropped: it is what an operator types into the app.
	if got := names.Label("042660"); !strings.Contains(got, "042660") {
		t.Errorf("Label = %q, want the code preserved", got)
	}
}

func TestAnUnknownSymbolRendersAsItsCode(t *testing.T) {
	var names engine.InstrumentNames
	if got := names.Label("PLTR"); got != "PLTR" {
		t.Errorf("Label = %q, want the bare code rather than a guess", got)
	}
	names.Learn("PLTR", "")
	if got := names.Label("PLTR"); got != "PLTR" {
		t.Errorf("Label = %q, want an empty name not to become a name", got)
	}
}

func TestANameSurvivesThePositionItNamed(t *testing.T) {
	var names engine.InstrumentNames
	names.Learn("042660", "한화오션")
	// The holding is gone from the account; the alert about it is not.
	if got := names.Label("042660"); got != "한화오션(042660)" {
		t.Errorf("Label = %q; the alerts that most need a name are about holdings that "+
			"have just stopped being holdings", got)
	}
}

func TestAnEmptyReadDoesNotEraseAKnownName(t *testing.T) {
	var names engine.InstrumentNames
	names.Learn("042660", "한화오션")
	names.Learn("042660", "   ")
	if got := names.Label("042660"); got != "한화오션(042660)" {
		t.Errorf("Label = %q, want a payload with no name to leave the known one alone", got)
	}
}

func TestTheZeroValueIsUsable(t *testing.T) {
	var names engine.InstrumentNames
	if got := names.Label("005930"); got != "005930" {
		t.Errorf("Label = %q; an engine with no reconciliation loop must still alert", got)
	}
	var nilNames *engine.InstrumentNames
	if got := nilNames.Name("005930"); got != "" {
		t.Errorf("Name on a nil registry = %q, want empty", got)
	}
}

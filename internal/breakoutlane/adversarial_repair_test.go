package breakoutlane

import "testing"

// These tests define the repair contract. They intentionally fail before the
// sealed-snapshot evaluator replaces the raw Event/Machine API.
func TestAdversarialSnapshotEvaluatorRejectsRawTransitionBypass(t *testing.T) {
	config, err := NewV1Config(V1ConfigInput{})
	if err == nil || config.Valid() {
		t.Fatal("empty v1 config was accepted")
	}
}

func TestAdversarialQuoteAndFXSealsRejectCallerControlledFreshness(t *testing.T) {
	_, err := NewQuoteSeal(QuoteSealInput{})
	if err == nil {
		t.Fatal("incomplete quote seal was accepted")
	}
	_, err = NewFXSeal(FXSealInput{})
	if err == nil {
		t.Fatal("incomplete FX seal was accepted")
	}
}

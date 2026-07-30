package main

import (
	"strings"
	"testing"
)

func TestEngineRunHelpDescribesUnwiredProtectionWithoutClaimingStartupIsImpossible(t *testing.T) {
	cmd := newEngineRunCmd(&rootOptions{})
	if strings.Contains(cmd.Long, "cannot start the engine") {
		t.Fatalf("engine run help still claims every start is impossible:\n%s", cmd.Long)
	}
	for _, want := range []string{"UNWIRED", "raises exposure", "Reduce-only exits"} {
		if !strings.Contains(cmd.Long, want) {
			t.Errorf("engine run help does not explain %q:\n%s", want, cmd.Long)
		}
	}
}

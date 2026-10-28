package engine

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

func TestA165EngineGatewayWiresOnlyTheDefinitiveCancelRefusalPolicy(t *testing.T) {
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}
	policy, ok := a098BuildGateway(t, context.Background(), j).gateway.CancelRefusalPolicy()
	if !ok || policy != (execgw.CancelRefusalPolicy{Definitive: true}) {
		t.Fatalf("cancel refusal policy: got %+v wired=%v", policy, ok)
	}
}

package strategyflow_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

func TestExternalCallerCannotPopulateOrResealExecutionTerms(t *testing.T) {
	typeOfTerms := reflect.TypeOf(strategyflow.ExecutionTerms{})
	for i := 0; i < typeOfTerms.NumField(); i++ {
		if typeOfTerms.Field(i).IsExported() {
			t.Fatalf("execution seal preimage leaked through exported field %q", typeOfTerms.Field(i).Name)
		}
	}
	var forged strategyflow.ExecutionTerms
	if err := json.Unmarshal([]byte(`{"accountRef":"acct","identity":"strategy-execution-terms:v1:sha256:forged"}`), &forged); err != nil {
		t.Fatal(err)
	}
	if forged.Valid() || forged.Identity() != "" {
		t.Fatalf("external JSON reconstructed a sealed result: %+v", forged)
	}
}

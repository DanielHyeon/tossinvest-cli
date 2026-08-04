package reversallane

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPurePackageHasNoBrokerJournalExitRegistryOrToggleAuthority(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range file.Imports {
			path := strings.Trim(spec.Path.Value, `"`)
			for _, forbidden := range []string{"/broker", "/journal", "/exit", "/gateway", "/operating", "/toggle", "/registry", "/strategyengine"} {
				if strings.Contains(path, forbidden) {
					t.Fatalf("%s imports forbidden authority %s", entry.Name(), path)
				}
			}
		}
	}
}

func TestProductionFilesContainNoMutationOrExitDecisionTypes(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Clean(entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"type OrderIntent struct", "PlaceOrder(", "CancelOrder(", "type JournalWriter", "type ExitDecision struct", "type ToggleWriter", "func MintRiskCap(", "func MintFrozenFX(", "A066RiskCapAuthority", "A066FXAuthority"} {
			if strings.Contains(string(body), forbidden) {
				t.Fatalf("%s contains forbidden authority symbol %s", entry.Name(), forbidden)
			}
		}
	}
}

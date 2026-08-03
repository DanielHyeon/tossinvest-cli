package protectionlifecycle

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionPackageHasNoLiveTransportRuntimeOrApprovalAuthority(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(".", entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"http://", "https://", "tossinvest.com", "toss.im", "PlaceOrder(", "CancelOrder(", "LiveApproval", "ToggleWriter", "LaneWriter"} {
			if strings.Contains(string(data), forbidden) {
				t.Fatalf("%s contains forbidden authority %q", path, forbidden)
			}
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, data, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			name := strings.Trim(spec.Path.Value, `"`)
			for _, forbidden := range []string{"net/http", "/internal/protection", "/internal/execgw", "/internal/app", "/internal/journal", "/internal/broker"} {
				if strings.Contains(name, forbidden) {
					t.Fatalf("%s imports forbidden package %q", path, name)
				}
			}
		}
	}
}

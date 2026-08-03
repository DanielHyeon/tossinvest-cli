package strategyrouter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageHasNoMutationAuthorityOrRuntimeDependency(t *testing.T) {
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
		source := string(data)
		for _, forbidden := range []string{"PlaceOrder(", "CancelOrder(", "JournalWriter", "BrokerClient", "ToggleWriter", "ActivationWriter", "CampaignWriter", "OwnerWriter"} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s contains mutation authority %q", entry.Name(), forbidden)
			}
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, data, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			path := strings.Trim(spec.Path.Value, `"`)
			if strings.Contains(path, "/internal/scheduler") || strings.Contains(path, "/internal/journal") || strings.Contains(path, "/internal/app") || strings.Contains(path, "/internal/httpapi") {
				t.Fatalf("%s imports runtime authority %q", entry.Name(), path)
			}
		}
		_ = ast.FileExports(parsed)
	}
}

func TestDescriptorsShipKRAndUSTogetherDefaultOFF(t *testing.T) {
	descriptors := Descriptors()
	if len(descriptors) != 2 || descriptors[0].Release != RouterRelease || descriptors[1].Release != RouterRelease {
		t.Fatalf("paired release missing=%+v", descriptors)
	}
	seen := map[Market]bool{}
	for _, descriptor := range descriptors {
		seen[descriptor.Market] = true
		if descriptor.Desired != StateOff || descriptor.Effective != StateOff || descriptor.Runtime != RuntimeUnobserved {
			t.Fatalf("descriptor not dormant=%+v", descriptor)
		}
	}
	if !seen[MarketKR] || !seen[MarketUS] || ValidateDescriptors(descriptors[:1]) == nil {
		t.Fatalf("one-market build passed paired conformance=%+v", descriptors)
	}
}

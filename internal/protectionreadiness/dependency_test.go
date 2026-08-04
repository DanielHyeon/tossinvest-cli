package protectionreadiness

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageHasNoMutationOrLiveTransportAuthority(t *testing.T) {
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
		for _, forbidden := range []string{"PlaceOrder(", "CancelOrder(", "ToggleWriter", "ActivationWriter", "LaneWriter", "LiveApproval", "tossinvest.com", "toss.im"} {
			if strings.Contains(string(data), forbidden) {
				t.Fatalf("%s contains authority/live dependency %q", entry.Name(), forbidden)
			}
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, data, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			importPath := strings.Trim(spec.Path.Value, `"`)
			for _, forbidden := range []string{"net/http", "/internal/protection", "/internal/execgw", "/internal/app", "/internal/journal"} {
				if strings.Contains(importPath, forbidden) {
					t.Fatalf("%s imports forbidden runtime package %q", entry.Name(), importPath)
				}
			}
		}
	}
}

func TestPairedDescriptorsAreSameReleaseDefaultUnwired(t *testing.T) {
	descriptors := Descriptors()
	if len(descriptors) != 2 || ValidateDescriptors(descriptors) != nil {
		t.Fatalf("paired descriptors=%+v", descriptors)
	}
	for _, descriptor := range descriptors {
		if descriptor.Release != ReadinessRelease || descriptor.State != Unwired || descriptor.SupervisorWired {
			t.Fatalf("descriptor created readiness=%+v", descriptor)
		}
	}
	if ValidateDescriptors(descriptors[:1]) == nil {
		t.Fatal("one-market release passed validation")
	}
}

package strategyruntime

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageHasNoBrokerLiveTransportToggleOrRuntimeWriter(t *testing.T) {
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
		for _, forbidden := range []string{"PlaceOrder(", "CancelOrder(", "BrokerClient", "JournalWriter", "GatewayClient", "ToggleWriter", "ActivationWriter", "LaneWriter", "tossinvest.com", "toss.im"} {
			if strings.Contains(string(data), forbidden) {
				t.Fatalf("%s contains mutation/live authority %q", entry.Name(), forbidden)
			}
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, data, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			importPath := strings.Trim(spec.Path.Value, `"`)
			for _, forbidden := range []string{"net/http", "/internal/execgw", "/internal/journal", "/internal/app", "/internal/scheduler"} {
				if strings.Contains(importPath, forbidden) {
					t.Fatalf("%s imports runtime authority %q", entry.Name(), importPath)
				}
			}
		}
	}
}

func TestDescriptorsShipPairedSameReleaseDefaultOFF(t *testing.T) {
	descriptors := Descriptors()
	if len(descriptors) != 2 || ValidateDescriptors(descriptors) != nil {
		t.Fatalf("paired descriptors=%+v", descriptors)
	}
	for _, descriptor := range descriptors {
		if descriptor.Release != RuntimeRelease || descriptor.Desired != EntryOff || descriptor.Effective != EntryOff || descriptor.Runtime != RuntimeUnobserved {
			t.Fatalf("descriptor enabled runtime=%+v", descriptor)
		}
	}
	if ValidateDescriptors(descriptors[:1]) == nil {
		t.Fatal("one-market runtime release accepted")
	}
}

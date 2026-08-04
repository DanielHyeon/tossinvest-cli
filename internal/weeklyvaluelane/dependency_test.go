package weeklyvaluelane

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPurePackageHasNoSourceAPIBrokerJournalExitOrToggleAuthority(t *testing.T) {
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
			for _, forbidden := range []string{"net/http", "/broker", "/journal", "/exit", "/gateway", "/operating", "/toggle", "/runtime", "/opendart", "/edgar"} {
				if strings.Contains(path, forbidden) {
					t.Fatalf("%s imports forbidden authority/source %s", entry.Name(), path)
				}
			}
		}
	}
}

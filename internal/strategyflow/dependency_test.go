package strategyflow

import (
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
)

func TestPureFlowAndLaneDependencyClosureHasNoMutationAuthority(t *testing.T) {
	command := exec.Command("go", "list", "-json", "-deps", ".")
	raw, err := command.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	seen := 0
	for {
		var entry struct{ ImportPath string }
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		seen++
		for _, forbidden := range []string{"/internal/journal", "/internal/execgw", "/internal/trading", "/internal/official", "/internal/config", "/internal/app/engine", "/internal/console", "/internal/httpapi"} {
			if strings.Contains(entry.ImportPath, forbidden) {
				t.Fatalf("pure strategyflow dependency reached mutation/operating authority: %s", entry.ImportPath)
			}
		}
	}
	if seen == 0 {
		t.Fatal("go list returned no dependency packages")
	}
}

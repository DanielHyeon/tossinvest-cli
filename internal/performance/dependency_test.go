package performance

import (
	"os/exec"
	"strings"
	"testing"
)

func TestPerformanceDependencyClosureHasNoTradingOrConfigurationAuthority(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, out)
	}
	for _, forbidden := range []string{
		"/internal/client", "/internal/config", "/internal/gateway", "/internal/strategy",
		"/internal/app/engine", "/internal/positionpolicy", "/internal/positionpolicyrpc",
	} {
		if strings.Contains(string(out), forbidden) {
			t.Errorf("performance dependency closure contains forbidden authority %s", forbidden)
		}
	}
}

package deployguard_test

import (
	"os"
	"strings"
	"testing"
)

func TestRepositoryComposeAndOperationsDeclareReadOnlyDeploymentGuardBoundary(t *testing.T) {
	composeBody, err := os.ReadFile("../../compose.yaml")
	if err != nil {
		t.Fatal(err)
	}
	compose := string(composeBody)
	for _, want := range []string{"Development/build tag only", "never a valid deployment preimage", "repository@sha256:<64-hex>"} {
		if !strings.Contains(compose, want) {
			t.Errorf("compose boundary lacks %q", want)
		}
	}
	if strings.Contains(compose, "/var/run/docker.sock") || strings.Contains(compose, "/run/docker.sock") {
		t.Fatal("Compose grants a service Docker replacement authority")
	}

	operationsBody, err := os.ReadFile("../../docs/operations.md")
	if err != nil {
		t.Fatal(err)
	}
	operations := string(operationsBody)
	for _, want := range []string{
		"Dormant digest-pinned deployment preflight", "`httpapi` → `tossos`", "최대 5분",
		"`NOT_APPLIED`와 `APPLIED`", "`ROLLBACK_INCOMPATIBLE`", "destructive rollback action을 내지 않고",
		"Docker/process 실행기", "어떤 recovery도 autostart", "journal 또는 volume을 수정",
	} {
		if !strings.Contains(operations, want) {
			t.Errorf("operations guard lacks %q", want)
		}
	}
}

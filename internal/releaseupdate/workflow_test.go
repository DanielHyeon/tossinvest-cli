package releaseupdate

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestReleaseWorkflowPinsActionsAttestsArchivesAndForbidsClobber(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "release.yml"))
	if err != nil {
		t.Fatal(err)
	}
	src := string(data)
	if strings.Contains(src, "--clobber") || strings.Contains(src, "gh release edit") {
		t.Fatal("attested release assets remain mutable")
	}
	for _, want := range []string{
		"attestations: write",
		"artifact-metadata: write",
		"id-token: write",
		"subject-path:",
		"actions/attest@",
		"tar --format=ustar",
		"contents: read",
		"contents: write",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("workflow missing %q", want)
		}
	}
	action := regexp.MustCompile(`uses:\s+[^@\s]+@([^\s#]+)`)
	for _, match := range action.FindAllStringSubmatch(src, -1) {
		if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(match[1]) {
			t.Errorf("action is not pinned to a full commit SHA: %s", match[0])
		}
	}
}

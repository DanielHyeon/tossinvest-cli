package execgw_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

// TestWriteReasonCodeGolden regenerates testdata/reason_codes.golden when run with
// -run TestWriteReasonCodeGolden -tags= and TOSSOS_UPDATE_GOLDEN=1. It is the only
// sanctioned way to change the fixture, so a rename is always a visible diff.
func TestWriteReasonCodeGolden(t *testing.T) {
	if os.Getenv("TOSSOS_UPDATE_GOLDEN") == "" {
		t.Skip("set TOSSOS_UPDATE_GOLDEN=1 to regenerate the reason-code fixture")
	}
	lines := make([]string, 0)
	for _, c := range execgw.AllReasonCodes() {
		lines = append(lines, string(c))
	}
	path := filepath.Join("testdata", "reason_codes.golden")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	t.Logf("wrote %s with %d codes", path, len(lines))
}

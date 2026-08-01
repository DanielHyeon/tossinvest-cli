//go:build !unix

package positionpolicyrpc

import (
	"os"
	"testing"
)

func TestOwnershipVerificationFailsClosedOnUnsupportedPlatforms(t *testing.T) {
	info, err := os.Stat(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := validateOwnerAndLinks(info, false); err == nil {
		t.Fatal("ownership verification silently succeeded on unsupported platform")
	}
}

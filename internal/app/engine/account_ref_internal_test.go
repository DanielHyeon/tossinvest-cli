package engine

import (
	"context"
	"strings"
	"testing"
)

func TestResolveAccountRefRejectsNilOfficialClient(t *testing.T) {
	_, err := resolveAccountRef(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "no official client") {
		t.Fatalf("resolveAccountRef(nil) error = %v, want no official client", err)
	}
}

package journal

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenReadOnlyRejectsV8BeforePolicyQuery(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	v8 := openJournalAtSchema(t, path, 8)
	if err := v8.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if !errors.Is(err, ErrSchemaTooOld) {
		t.Fatalf("OpenReadOnly against v8 = %v, want ErrSchemaTooOld", err)
	}
	if !strings.Contains(err.Error(), "exit_states.policy_id") {
		t.Fatalf("v8 refusal does not name the required column: %v", err)
	}
	if strings.Contains(err.Error(), "SQL logic error") {
		t.Fatalf("v8 escaped schema classification and failed as a later query: %v", err)
	}
}

package journal

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenReadOnlyRejectsV9BeforeSnapshotQueries(t *testing.T) {
	path := filepath.Join(t.TempDir(), DBFileName)
	v9 := openJournalAtSchema(t, path, 9)
	if err := v9.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := OpenReadOnly(context.Background(), ReadOnlyOptions{Path: path})
	if !errors.Is(err, ErrSchemaTooOld) {
		t.Fatalf("OpenReadOnly against v9 = %v, want ErrSchemaTooOld", err)
	}
	if !strings.Contains(err.Error(), "exit_states.snapshot_status") {
		t.Fatalf("v9 refusal does not name the required snapshot column: %v", err)
	}
	if strings.Contains(err.Error(), "SQL logic error") {
		t.Fatalf("v9 escaped schema classification and failed as a later query: %v", err)
	}
}

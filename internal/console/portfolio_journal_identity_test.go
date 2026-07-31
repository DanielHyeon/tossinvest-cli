package console

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestPositionsReadsOnlyTheSelectedJournal(t *testing.T) {
	dir := t.TempDir()
	defaultPath := filepath.Join(dir, "default", journal.DBFileName)
	profilePath := filepath.Join(dir, "profile", journal.DBFileName)
	seedEmptyJournal(t, defaultPath)
	seedJournal(t, profilePath)

	h := newDashboardHarness(t, func(o *Options) {
		o.JournalPath = profilePath
	})
	h.authenticate(t)
	page := h.page(t, "/positions")
	for _, want := range []string{"005930", "68000", "69500"} {
		if !strings.Contains(page, want) {
			t.Fatalf("positions page did not read selected profile marker %q", want)
		}
	}

	defaultReader, err := journal.OpenReadOnly(context.Background(), journal.ReadOnlyOptions{Path: defaultPath})
	if err != nil {
		t.Fatalf("open unrelated default journal: %v", err)
	}
	defer defaultReader.Close()
	refs, err := defaultReader.AccountRefs(context.Background())
	if err != nil {
		t.Fatalf("read unrelated default journal: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("unrelated default journal unexpectedly contains profile rows: %v", refs)
	}
}

// TestPositionsPreservesTheExactSelectedJournalPath closes the downstream half
// of the active-profile identity contract. A whitespace-only relative
// --config-dir is unusual but valid and the engine writes " /journal.db"; the
// console must not normalize that distinct path into "/journal.db".
func TestPositionsPreservesTheExactSelectedJournalPath(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.Mkdir(" ", 0o700); err != nil {
		t.Fatalf("create whitespace profile: %v", err)
	}
	selectedPath := filepath.Join(" ", journal.DBFileName)
	seedJournal(t, filepath.Join(root, selectedPath))

	h := newDashboardHarness(t, func(o *Options) {
		o.JournalPath = selectedPath
	})
	h.authenticate(t)
	page := h.page(t, "/positions")
	for _, want := range []string{"005930", "68000", "69500"} {
		if !strings.Contains(page, want) {
			t.Fatalf("positions page changed the selected journal path; marker %q is absent", want)
		}
	}
}

package position_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
)

// eligibility_drift_test.go keeps exit eligibility a single predicate (change
// adopt-external-positions task 1.3; position-ledger: 자격 판정은 단일 술어
// 함수로 모은다 SHALL).
//
// # What drifted before, and what it cost
//
// Until this change the answer to "may the exit policy manage this position"
// was spelled in four places: the observation loop's enumeration, the
// projection's method, this package's predicate, and reconcile's ingest — plus a
// fifth that was not a test at all but a hardcoded `false` on the alert the
// operator reads. They agreed, because there was one column and one answer.
//
// Adding a second justifying record is exactly the change that makes four copies
// disagree: any one of them left as `entry_decision_id != ""` reports an adopted
// position as unmanaged, and the two failure modes are both bad in the same
// direction — an operator told a protected position is bare, or a loop that
// silently stops protecting it.
//
// So the rule is that the columns are named only where they are *stored* or
// where a deliberate, documented exception exists, and every judgement goes
// through [position.ExitEligible].

// eligibilitySpellers are the files allowed to name the two columns, with the
// reason each one is not a drifted copy of the predicate.
var eligibilitySpellers = map[string]string{
	"internal/position/provenance.go": "the predicate itself",
	"internal/journal/adoption.go": "the adoption writer: it reads both columns inside its own " +
		"transaction to refuse an engine-entered position and to enforce set-once",
	"internal/journal/position_projection.go": "the projection: the select list and Adopted(), " +
		"which reports which record justifies the position rather than whether one does",
	"internal/journal/exit_state.go": "the opening transaction's re-read, which calls the predicate",
	"internal/journal/position_adjustments.go": "the INSERT that folds an external holding in with " +
		"entry_decision_id NULL",
	"internal/journal/core_domain.go":    "the DDL",
	"internal/journal/provenance.go":     "the lineage query's join path",
	"internal/journal/trade_outcomes.go": "the freeze's read of which record justifies the position",
	"internal/reconcile/external.go": "the fold guard, deliberately narrowed to an explicit " +
		"entry_decision_id comparison (adopt-external-positions design A1): a fold landing on an " +
		"adopted position is the ordinary re-reconciliation path and must not be refused",
	"internal/console/portfolio.go": "the dashboard's prose about what the labels mean",
	"internal/position/position.go": "a comment about what a re-entry mints",
	"internal/app/engine/exitloop.go": "the engine-entered arm of the opening path, which looks the " +
		"entry decision up. It is reached only after the predicate said yes and the position " +
		"reported which record justifies it — it is not a second eligibility test",
	"internal/app/engine/adoption.go": "the adoption alert's field list, which names the id of the " +
		"record it just wrote. The judgement above it calls the predicate",
}

// eligibilityTokens are the spellings a drifted copy would use.
var eligibilityTokens = []string{
	"EntryDecisionID",
	"entry_decision_id",
	"AdoptionID",
	"adoption_id",
}

// TestExitEligibilityHasOneSpelling walks every production source in the module
// and fails on a file that names an eligibility column without being on the
// list above.
func TestExitEligibilityHasOneSpelling(t *testing.T) {
	t.Parallel()

	root := moduleRoot(t)
	checked, spellers := 0, 0
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		checked++
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		for _, token := range eligibilityTokens {
			if !strings.Contains(string(source), token) {
				continue
			}
			if _, allowed := eligibilitySpellers[rel]; !allowed {
				t.Errorf("%s names %q. Exit eligibility is position.ExitEligible and nothing else — "+
					"a second copy of \"which columns count\" is how an adopted position comes to be "+
					"reported as unmanaged. If this file genuinely has to name the column, add it to "+
					"eligibilitySpellers with the reason.", rel, token)
			}
			spellers++
			break
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the module: %v", err)
	}
	// Positive controls: an empty walk, or a list that stopped matching anything,
	// would make the assertion above pass for the wrong reason.
	if checked == 0 {
		t.Fatal("no production files were inspected; the walk is broken")
	}
	if spellers == 0 {
		t.Fatal("no file names an eligibility column; the walk or the token list is broken")
	}
}

// TestExitEligibleIsTheWholeTruthTable pins the predicate itself. Two records
// justify a baseline and there is no third, so the table has exactly four rows
// and three of them are true.
func TestExitEligibleIsTheWholeTruthTable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		decision, adoption string
		want               bool
		why                string
	}{
		{"", "", false, "no record justifies the position: there is no stop to make a baseline of"},
		{"decision-1", "", true, "the engine opened it; the decision's stop is t0"},
		{"", "adopt-1", true, "it was adopted; the record's synthetic stop is t0"},
		{"decision-1", "adopt-1", true, "both, which the writers forbid but the predicate still answers"},
		{"  ", "  ", false, "whitespace is not a reference"},
	}
	for _, c := range cases {
		if got := position.ExitEligible(c.decision, c.adoption); got != c.want {
			t.Errorf("ExitEligible(%q, %q) = %v, want %v — %s", c.decision, c.adoption, got, c.want, c.why)
		}
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate go.mod above the test directory")
		}
		dir = parent
	}
}

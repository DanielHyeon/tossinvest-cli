package journal

import (
	"os"
	"strings"
	"testing"
)

// adoption_static_test.go is the layout half of "adoption_id is set once"
// (change adopt-external-positions task 1.2; position-ledger: 전용 tx API로만
// 기입하고 그 외 UPDATE의 언급은 정적 스캔이 거부한다 SHALL).
//
// # Why a static scan and not only a behavioural test
//
// A behavioural test can show that today's code path does not repoint a live
// position at a second adoption. It cannot show that tomorrow's will not: the
// column is one `UPDATE positions SET adoption_id = …` away from being
// re-writable, and that statement would compile, pass every existing test, and
// silently move the baseline a stop is measured from.
//
// So the rule is enforced the way apply_hook.go's guarded four are — by a scan
// that fails if any other production file in this package writes the column at
// all. A second writer cannot appear without somebody editing the list below and
// justifying it in review.
//
// # The two columns, and why they are guarded differently
//
//	adoption_id        may be *read* wherever a position is projected, but may be
//	                   written only by adoption.go, and only under a predicate
//	                   that makes the write a no-op on an already-adopted row.
//	entry_decision_id  may be written by the INSERT that opens an instance and by
//	                   nothing else, ever. Its immutability is a landed SHALL NOT
//	                   and this change explicitly does not become its first
//	                   mutator (design A1).

// adoptionIDReaders are the production files allowed to name `adoption_id` at
// all. Everything here reads it; adoption.go is additionally the sole writer.
var adoptionIDReaders = map[string]bool{
	"adoption.go":            true, // the DDL, the writer, the reads
	"position_projection.go": true, // the projection's select list
	"exit_state.go":          true, // the eligibility read inside the opening tx
	"trade_outcomes.go":      true, // the adopted-position freeze branch
	"provenance.go":          true, // the ADOPTION lineage arm
	"trade_analytics.go":     true, // the split-by-source join
}

// productionSources returns every non-test .go file in this package, with its
// text.
func productionSources(t *testing.T) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}
	out := map[string]string{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		source, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("reading %s: %v", name, err)
		}
		out[name] = string(source)
	}
	if len(out) == 0 {
		t.Fatal("no production files were inspected; the walk is broken")
	}
	return out
}

// updateStatements extracts every SQL UPDATE the file spells.
//
// Every statement in this package lives in a backquoted raw string, so a chunk
// that starts at "UPDATE" and ends at the closing backquote (or at the next
// statement separator) contains the whole of it. It is a lexical approximation
// on purpose: a scan that tried to parse SQL would have its own bugs, and the
// approximation errs towards *including* more text, which can only make the
// guard stricter.
func updateStatements(source string) []string {
	var out []string
	rest := source
	for {
		at := strings.Index(rest, "UPDATE ")
		if at < 0 {
			return out
		}
		rest = rest[at:]
		end := len(rest)
		if stop := strings.IndexAny(rest, "`;"); stop >= 0 {
			end = stop
		}
		out = append(out, rest[:end])
		rest = rest[end:]
	}
}

// TestAdoptionIDIsWrittenOnlyByTheAdoptionAPI is the scan.
func TestAdoptionIDIsWrittenOnlyByTheAdoptionAPI(t *testing.T) {
	t.Parallel()

	sources := productionSources(t)
	writer := false
	for name, text := range sources {
		mentions := strings.Contains(text, "adoption_id")
		if mentions && !adoptionIDReaders[name] {
			t.Errorf("%s names positions.adoption_id: that column is set once by "+
				"Journal.AdoptPosition (adoption.go), and a file that can spell it is a file "+
				"that can write it — add it to adoptionIDReaders in review if the read is genuinely "+
				"needed", name)
		}
		for _, stmt := range updateStatements(text) {
			if !strings.Contains(stmt, "adoption_id") {
				continue
			}
			if name != "adoption.go" {
				t.Errorf("%s contains an UPDATE naming adoption_id:\n\t%s\n"+
					"the reference is set once; a second writer would move the synthetic t0 a live "+
					"position's stop is measured from", name, strings.TrimSpace(stmt))
				continue
			}
			writer = true
			// The predicate is the guarantee. Without it the single writer is
			// still a writer that can repoint an adopted position.
			if !strings.Contains(stmt, "adoption_id IS NULL") {
				t.Errorf("the adoption_id write is not set-once:\n\t%s\n"+
					"it must carry `WHERE … adoption_id IS NULL` so a second write affects no row",
					strings.TrimSpace(stmt))
			}
		}
	}
	// Positive control: an assertion that passes because nothing writes the
	// column at all is protecting nothing.
	if !writer {
		t.Fatal("no production file writes adoption_id; the set-once rule is protecting nothing")
	}
}

// TestEntryDecisionIDIsNeverUpdated is the other half of design A1's promise.
//
// `positions.entry_decision_id` is written by the INSERT that opens an instance
// and is never updated (position-ledger: 설정 후 인스턴스 수명 동안 변경·NULL화
// 되지 않는다 SHALL NOT). The adoption change reads it in three new places and
// writes it in none, and this is what keeps that true as the file set grows.
func TestEntryDecisionIDIsNeverUpdated(t *testing.T) {
	t.Parallel()

	inserters := 0
	for name, text := range productionSources(t) {
		if strings.Contains(text, "INSERT INTO positions") &&
			strings.Contains(text, "entry_decision_id") {
			inserters++
		}
		for _, stmt := range updateStatements(text) {
			if strings.Contains(stmt, "entry_decision_id") {
				t.Errorf("%s contains an UPDATE naming entry_decision_id:\n\t%s\n"+
					"that column is set by the INSERT that opens the instance and never moved again",
					name, strings.TrimSpace(stmt))
			}
		}
	}
	if inserters == 0 {
		t.Fatal("no production file inserts entry_decision_id; the walk is broken")
	}
}

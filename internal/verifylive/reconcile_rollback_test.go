package verifylive_test

// reconcile_rollback_test.go — a121 tasks 2.4.1 의 「구 바이너리 롤백 방향」 시험.
//
// design G2 「구 바이너리」: FormatVersion 은 1 그대로(필드 추가)이고, 구 바이너리는 reconciled_absent 를 모르는 필드로
// 버리므로 대사 줄을 비-terminal 줄로 읽는다 → 그 artifact 는 여전히 outstanding 이고, holdGate 기본값과 그 줄의 위치
// 때문에 **다시 보유** 상태가 된다(틀리는 방향이 "다시 보인다" 이므로 안전 쪽).
//
// 구 판본은 이 파일 안의 **글자 그대로의 사본**으로 시뮬레이션한다(외부 시험 패키지라 원본 이름과 충돌하지 않는다).
// 사본이 구현 base 와 같음은 TestOldBinaryCopiesMatchTheImplementationBase 가 토큰열 sha256 으로 고정한다 — 기대값은
// base de147cc285c5274cad6d6ab7b208513027a70b40 의 같은 선언에서 같은 토큰화로 계산했다(base 는 불변이므로 한쪽 고정으로
// 충분하다; 학습 「옮겨 적은 코드는 양쪽을 다 못 박아야 한다」 의 원본 쪽은 커밋이 못 박는다).
// Entry 는 사본 함수가 읽는 필드만 가진 최소 형태다(아래 사본들이 읽는 필드: Kind·StepID·Verdict·Observations·
// Artifacts·M0Checkpoint).

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// --- 구 판본의 이름들(별칭) --------------------------------------------------------

type (
	StepID       = verifylive.StepID
	Verdict      = verifylive.Verdict
	Observation  = verifylive.Observation
	M0Checkpoint = verifylive.M0Checkpoint
)

const (
	KindOrder              = verifylive.KindOrder
	KindConditional        = verifylive.KindConditional
	KindM0Checkpoint       = verifylive.KindM0Checkpoint
	StepConditionalCancel  = verifylive.StepConditionalCancel
	StepConditionalTrigger = verifylive.StepConditionalTrigger
	VerdictFail            = verifylive.VerdictFail
)

// Entry 는 구 판본 사본이 읽는 필드만 가진 최소 줄이다(핀 대상 아님 — 위 머리말).
type Entry struct {
	Kind         string        `json:"kind"`
	StepID       StepID        `json:"step_id"`
	Verdict      Verdict       `json:"verdict"`
	Observations []Observation `json:"observations,omitempty"`
	Artifacts    []Artifact    `json:"artifacts,omitempty"`
	M0Checkpoint *M0Checkpoint `json:"m0_checkpoint,omitempty"`
}

// --- base de147cc2 의 글자 그대로 사본 (record.go · cleanup.go · m0_manual.go) --------

type Artifact struct {
	Kind        string    `json:"kind"`
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	CreatedAt   time.Time `json:"created_at"`
	CancelledAt time.Time `json:"cancelled_at,omitempty"`
	Cancelled   bool      `json:"cancelled"`
	Filled      bool      `json:"filled,omitempty"`
	FilledAt    time.Time `json:"filled_at,omitempty"`
	Deliberate  bool      `json:"deliberate,omitempty"`
	HeldUntil   StepID    `json:"held_until,omitempty"`
	ChainID     string    `json:"chain_id,omitempty"`
	Note        string    `json:"note,omitempty"`
}

func LastEntry(entries []Entry, id StepID) (Entry, bool) {
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].StepID == id {
			return entries[i], true
		}
	}
	return Entry{}, false
}

func Settled(entries []Entry, id StepID) bool {
	e, ok := LastEntry(entries, id)
	return ok && e.Verdict.Terminal()
}

func Outstanding(entries []Entry) []Artifact {
	var out []Artifact
	for _, l := range outstandingLines(entries) {
		out = append(out, l.Artifact)
	}
	return out
}

type outstandingLine struct {
	Artifact
	at int
}

func outstandingLines(entries []Entry) []outstandingLine {
	order := []string{}
	latest := map[string]outstandingLine{}
	for i, e := range entries {
		for _, a := range e.Artifacts {
			key := a.Kind + "\x00" + a.ID
			if _, seen := latest[key]; !seen {
				order = append(order, key)
			}
			prev, seen := latest[key]
			if seen && prev.terminal() && !a.terminal() {
				continue
			}
			latest[key] = outstandingLine{Artifact: a, at: i}
		}
	}
	var out []outstandingLine
	for _, key := range order {
		if l := latest[key]; !l.terminal() {
			out = append(out, l)
		}
	}
	return out
}

func (a Artifact) terminal() bool { return a.Cancelled || a.Filled }

func PendingCleanup(entries []Entry) []Artifact {
	return withoutM0ManualReconcile(entries, cleanupFrom(entries, func(id StepID) bool { return Settled(entries, id) }))
}

func cleanupFrom(entries []Entry, settled func(StepID) bool) []Artifact {
	var out []Artifact
	for _, l := range outstandingLines(entries) {
		gate := holdGate(l.Artifact)
		if gate == "" {
			out = append(out, l.Artifact)
			continue
		}
		if settled(gate) && heldAfter(entries, gate, l.at) {
			out = append(out, l.Artifact)
		}
	}
	return out
}

func holdGate(a Artifact) StepID {
	if a.HeldUntil != "" {
		return a.HeldUntil
	}
	if a.Kind == KindConditional {
		return StepConditionalCancel
	}
	return ""
}

func heldAfter(entries []Entry, gate StepID, at int) bool {
	decided := -1
	for i := range entries {
		if entries[i].StepID == gate {
			decided = i
		}
	}
	return decided > at
}

func m0ManualReconcileIDs(entries []Entry) map[string]bool {
	ids := map[string]bool{}
	for _, entry := range entries {
		if entry.StepID == StepConditionalTrigger && m0TriggeredUnresolved(entry) {
			for _, artifact := range entry.Artifacts {
				if artifact.Kind == KindConditional && !artifact.Cancelled && !artifact.Filled {
					ids[KindConditional+"\x00"+artifact.ID] = true
				}
			}
		}
		if entry.Kind != KindM0Checkpoint || entry.M0Checkpoint == nil {
			continue
		}
		if entry.M0Checkpoint.Kind != "parent-created" && entry.M0Checkpoint.Kind != "child-observed" {
			continue
		}
		if entry.M0Checkpoint.ParentConditionalID != "" {
			ids[KindConditional+"\x00"+entry.M0Checkpoint.ParentConditionalID] = true
		}
		if entry.M0Checkpoint.ChildOrderID != "" {
			ids[KindOrder+"\x00"+entry.M0Checkpoint.ChildOrderID] = true
		}
	}
	return ids
}

func m0TriggeredUnresolved(entry Entry) bool {
	if entry.Verdict != VerdictFail {
		return false
	}
	for _, observation := range entry.Observations {
		if (observation.Key == "conditional.trigger_observed" && observation.Value == "true") ||
			(observation.Key == "conditional.trigger.conditional_presumed_fired" && observation.Value == "true") {
			return true
		}
	}
	return false
}

func withoutM0ManualReconcile(entries []Entry, artifacts []Artifact) []Artifact {
	ids := m0ManualReconcileIDs(entries)
	out := make([]Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		if !ids[artifact.Kind+"\x00"+artifact.ID] {
			out = append(out, artifact)
		}
	}
	return out
}

// --- 사본 핀 ----------------------------------------------------------------------

// oldBinaryPins 는 base de147cc2 의 선언별 토큰열 sha256 이다(주석 제외, 자동 세미콜론 제외).
var oldBinaryPins = map[string]string{
	"Artifact":                 "7d1324de1e920c9ff7ee650afe4da9d2b253a14c4b6b09dd639f15020934705e",
	"terminal":                 "a9be08e24a37993c0aa4e9ff3fff5daa690b63a46a64bfcd86a083f4857c32ce",
	"outstandingLine":          "de8782a4a1aa17a8120383dd9a4bc5e2fb8ca595879b28d8c6b5bef1a3c9af82",
	"outstandingLines":         "e803f15783aa504c796622258492b432ab67a80b39523944adb179ce12305493",
	"Outstanding":              "2e82764c771ea1de8fd535efa515f0bc3da9e780dc44e06b6fd4ddbbdb39910f",
	"LastEntry":                "b4c23753b618f771bed037af20ee2656415c0889077e87267f2f06a9e8e29550",
	"Settled":                  "fef0972e8da46050cd96ecabf389f741927bb3177c67f1c4bf2a35144f53858b",
	"PendingCleanup":           "c8b2a0f734d3346aabb1ff1aacf14dcf9c49834a0892cfe245207e290a94a383",
	"cleanupFrom":              "1d5c763383b565bdaba8879b997e2a5eb81edd4928a6ef4eceacc3a3a4fcccf0",
	"holdGate":                 "ec14b564244f855e8eea60a9f83bde4a1a7ade1165f67b7969aaadcb1689d2f5",
	"heldAfter":                "a00224ee188f78902f664143edad47ce17dbf687aa4f7cd2f66852c6f6cbb874",
	"m0ManualReconcileIDs":     "10ef2865d3213d3016896c0d6fb2664dc5de41b9aa183d468bef670abafa80e3",
	"m0TriggeredUnresolved":    "ee168887e18f99a928f1c4178bd792c59f46e34defc83c2057d917f8f1a49e05",
	"withoutM0ManualReconcile": "85a47bc351ccbcbad185bf371abed034f5d40af665693b1bcf836e69f9690cbc",
}

// TestOldBinaryCopiesMatchTheImplementationBase 는 위 사본이 구현 base 의 선언과 토큰 단위로 같음을 고정한다.
func TestOldBinaryCopiesMatchTheImplementationBase(t *testing.T) {
	src, err := os.ReadFile("reconcile_rollback_test.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "reconcile_rollback_test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range oldBinaryPins {
		got, ok := oldDeclTokens(src, fset, file, name)
		if !ok {
			t.Fatalf("copy of %s is missing", name)
		}
		if got != want {
			t.Fatalf("copy of %s drifted from base de147cc2 (token sha %s, want %s)", name, got, want)
		}
	}
}

// oldDeclTokens 는 선언 하나의 토큰열 sha256 이다 — 기대값을 만든 도구와 같은 규칙.
func oldDeclTokens(src []byte, fset *token.FileSet, file *ast.File, name string) (string, bool) {
	for _, d := range file.Decls {
		var start, end token.Pos
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Name.Name != name {
				continue
			}
			start, end = d.Pos(), d.End()
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			found := false
			for _, s := range d.Specs {
				if ts := s.(*ast.TypeSpec); ts.Name.Name == name {
					found = true
					if d.Lparen.IsValid() {
						start, end = ts.Pos(), ts.End()
					} else {
						start, end = d.Pos(), d.End()
					}
				}
			}
			if !found {
				continue
			}
		default:
			continue
		}
		a, b := fset.Position(start).Offset, fset.Position(end).Offset
		var s scanner.Scanner
		fs := token.NewFileSet()
		f := fs.AddFile("", -1, b-a)
		s.Init(f, src[a:b], nil, 0)
		var sb strings.Builder
		for {
			_, tok, lit := s.Scan()
			if tok == token.EOF {
				break
			}
			if tok == token.SEMICOLON && lit == "\n" {
				continue
			}
			sb.WriteString(tok.String() + " " + lit + "\n")
		}
		sum := sha256.Sum256([]byte(sb.String()))
		return hex.EncodeToString(sum[:]), true
	}
	return "", false
}

// --- 롤백 방향 --------------------------------------------------------------------

func oldDecode(t *testing.T, raw []byte) []Entry {
	t.Helper()
	var out []Entry
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("old decoder: %v", err)
		}
		out = append(out, e)
	}
	return out
}

// TestAnOlderBinaryStillSeesTheReconciledArtifactAsOutstandingAndHeld — 구 바이너리에서 대사된 artifact 는 여전히
// outstanding 이고(살아 있다고 보인다), 재개 정리 대상이 아니다(다시 보유 — 다음 conditional-cancel 판정 전까지).
// 대조군: 대사 줄이 없으면 구 바이너리도 그것을 정리 대상으로 낸다 — 다시 보유로 만든 것은 대사 줄의 위치다.
func TestAnOlderBinaryStillSeesTheReconciledArtifactAsOutstandingAndHeld(t *testing.T) {
	record := verifylive.RCA063RecordForTest(t)
	without := oldDecode(t, record)
	if !oldHas(Outstanding(without), verifylive.RCTargetIDForTest) || !oldHas(PendingCleanup(without), verifylive.RCTargetIDForTest) {
		t.Fatal("control: the old binary must plan the released a063 artifact for cleanup")
	}
	with := oldDecode(t, append(record, []byte(verifylive.RCReconcileLineForTest())...))
	if !oldHas(Outstanding(with), verifylive.RCTargetIDForTest) {
		t.Fatal("an older binary hides a reconciled artifact — rollback must fail in the 'still visible' direction")
	}
	if oldHas(PendingCleanup(with), verifylive.RCTargetIDForTest) {
		t.Fatal("an older binary plans a DELETE for the reconciled artifact; it must be held again")
	}
}

func oldHas(arts []Artifact, id string) bool {
	for _, a := range arts {
		if a.ID == id {
			return true
		}
	}
	return false
}

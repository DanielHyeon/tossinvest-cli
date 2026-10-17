package verifylive

// reconcile_codex_test.go — a121 4.2 외부 모델(codex, clean) 발견 CG-1·CG-2·CG-3·CG-4(표시)·CG-5 의 시험.
// 각 거절은 그 가드의 코드·오류 표지로 단언한다.

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// rcAllowExpired 는 CLOSED 의 다른 id EXPIRED 행이 수락되도록 허용 목록을 시험 동안만 채운다(전사 전 생산 값은 빈 목록).
func rcAllowExpired(t *testing.T) {
	t.Helper()
	prev := reconcileClosedStatusAllowlist
	reconcileClosedStatusAllowlist = map[string]bool{"EXPIRED": true}
	t.Cleanup(func() { reconcileClosedStatusAllowlist = prev })
}

// --- CG-1 — 둘째 읽기에만 나타나는 OCO·타 심볼·타 시장·필드 결측 ---------------------------

// TestReconcileValidatesTheSecondSnapshotToo 는 codex CG-1 이다 — 비교 tuple(그룹, id, status, triggeredOrderId)에 없는
// second 다리·심볼·시장·필수 필드가 둘째 읽기에서만 어긋나도 multiset 비교는 같다고 본다. 행 검사가 두 스냅숏 각각에
// 돌아야 잡힌다(첫 스냅숏만 보면 OCO 다리를 본 채로 부재 사건을 쓴다 — F6 위반).
func TestReconcileValidatesTheSecondSnapshotToo(t *testing.T) {
	base := func() official.ReconcileConditionalRow { return rcCondRow("CO-OTHER-CG1", "EXPIRED") }
	cases := map[string]struct {
		mutate func(*official.ReconcileConditionalRow)
		want   ReconcileRefusalCode
	}{
		"second-leg":    {func(r *official.ReconcileConditionalRow) { r.HasSecond = true }, RefuseOCO},
		"other-symbol":  {func(r *official.ReconcileConditionalRow) { r.Symbol = "000660" }, RefuseRowSymbol},
		"other-market":  {func(r *official.ReconcileConditionalRow) { r.Market = MarketUS }, RefuseRowMarket},
		"missing-field": {func(r *official.ReconcileConditionalRow) { r.Market = "" }, RefuseRowIncomplete},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			rcAllowExpired(t)
			h := newRCHarness(t, rcA063Entries())
			second := base()
			c.mutate(&second)
			h.reader.set(0, ReconcileGroupConditionalClosed, rcCondPage("", false, base()))
			h.reader.set(1, ReconcileGroupConditionalClosed, rcCondPage("", false, second))
			rcRefuse(t, h, c.want)
		})
	}
}

// --- CG-2 — 기록 파일 자체의 배제 ------------------------------------------------------------

// TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile 는 codex CG-2 다 — 다른 프로필 경로가 같은 inode 를
// 가리켜도(하드링크) 기록 파일 자체의 flock 이 두 대사를 직렬화한다. 잡힌 동안은 거절하고 기록은 그대로.
func TestReconcileRefusesWhileAnotherReconciliationHoldsTheRecordFile(t *testing.T) {
	for _, alias := range []bool{false, true} {
		name := "same-path"
		if alias {
			name = "hardlink-alias"
		}
		t.Run(name, func(t *testing.T) {
			h := newRCHarness(t, rcA063Entries())
			other := h.path
			if alias {
				other = filepath.Join(t.TempDir(), "profile-q-record.jsonl")
				if err := os.Link(h.path, other); err != nil {
					t.Skipf("hard links unsupported here: %v", err)
				}
			}
			held, err := lockRecordForAppend(other)
			if err != nil {
				t.Fatal(err)
			}
			defer held.release()
			rcRefuse(t, h, RefuseRecordLocked)
		})
	}
}

// TestReconcileDetectsAWriterInterleavingAtTheAppend 는 codex CG-2 의 교차 쓰기 픽스처다 — 잠그지 않는 작성자(구 바이너리·
// --record 별칭의 verify run)가 판정 뒤·쓰기 순간에 끼어들면, 다시 읽은 바이트가 쓴 줄과 달라 침묵 없이 실패한다.
func TestReconcileDetectsAWriterInterleavingAtTheAppend(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	prev := recordWrite
	recordWrite = func(f *os.File, b []byte) (int, error) {
		rcAppendRaw(t, h.path, `{"format_version":1,"kind":"step","step_id":"costs","verdict":"skipped"}`+"\n")
		return f.Write(b)
	}
	t.Cleanup(func() { recordWrite = prev })
	_, err := h.run()
	if !errors.Is(err, errRecordAppendUnverified) {
		t.Fatalf("an interleaved writer must fail the read-back check loudly; got %v", err)
	}
}

// --- CG-3 — 자기 부분 쓰기는 침묵하지 않는다 ---------------------------------------------------

func TestReconcileAppendsInOneWriteEndingWithANewline(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	var writes [][]byte
	prev := recordWrite
	recordWrite = func(f *os.File, b []byte) (int, error) {
		writes = append(writes, append([]byte(nil), b...))
		return f.Write(b)
	}
	t.Cleanup(func() { recordWrite = prev })
	rcAccept(t, h)
	if len(writes) != 1 || !bytes.HasSuffix(writes[0], []byte("\n")) || bytes.Count(writes[0], []byte("\n")) != 1 {
		t.Fatalf("want exactly one write carrying one complete line with its newline; got %d write(s)", len(writes))
	}
}

func TestReconcileReportsAPartialAppendLoudly(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	prev := recordWrite
	recordWrite = func(f *os.File, b []byte) (int, error) {
		n, _ := f.Write(b[:len(b)/2])
		return n, errors.New("disk full")
	}
	t.Cleanup(func() { recordWrite = prev })
	_, err := h.run()
	if !errors.Is(err, errRecordAppendIncomplete) || !strings.Contains(err.Error(), "inspect its tail") {
		t.Fatalf("a partial append must be reported as incomplete with an inspection instruction; got %v", err)
	}
}

func TestReconcileReportsAFailedSyncLoudly(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	prev := recordSync
	recordSync = func(*os.File) error { return errors.New("EIO") }
	t.Cleanup(func() { recordSync = prev })
	_, err := h.run()
	if !errors.Is(err, errRecordAppendIncomplete) {
		t.Fatalf("a failed sync must be reported as an incomplete append; got %v", err)
	}
}

func TestReconcileVerifiesTheAppendByReadingItBack(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	prev := recordWrite
	recordWrite = func(f *os.File, b []byte) (int, error) {
		altered := append([]byte(nil), b...)
		altered[0] = '['
		return f.Write(altered)
	}
	t.Cleanup(func() { recordWrite = prev })
	_, err := h.run()
	if !errors.Is(err, errRecordAppendUnverified) {
		t.Fatalf("bytes that read back differently must fail the append; got %v", err)
	}
}

// --- CG-4 — 승인 표시 실패 ------------------------------------------------------------------

type rcFailingWriter struct{}

func (rcFailingWriter) Write([]byte) (int, error) { return 0, errors.New("terminal gone") }

// TestReconcileRefusesWhenTheApprovalCannotBeShown 는 codex CG-4 다 — 두 마스크·계좌 수를 보이지 못하면 승인을 묻지 않는다.
func TestReconcileRefusesWhenTheApprovalCannotBeShown(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	p := h.params()
	p.Out = rcFailingWriter{}
	before := h.bytes()
	_, err := Reconcile(t.Context(), p)
	rcRequireRefusal(t, err, RefuseApprovalDisplay)
	rcRequireUnchanged(t, h, before)
	if len(h.approvals) != 0 {
		t.Fatal("the operator was asked to approve what could not be shown")
	}
	rcRequireNoListRead(t, h)
}

// --- CG-5 — 최종 게이트는 쓰기 경계 직전 -------------------------------------------------------

// TestReconcileFinalGateSitsAtTheWriteBoundary 는 codex CG-5 의 구조 핀이다 — 기록 잠금·fd 준비와 줄 직렬화가 마지막
// 신선도·Q1 재검 **앞**에 있고, 재검 뒤에는 한 번의 쓰기(appendLine)만 온다. 일시 정지(suspend) 동안의 벽시계 공백은
// time.Now 의 단조 성분이 세지 않을 수 있다 — 측정 한계로 기록(analysis/green-lot/codex-green-repairs.md).
func TestReconcileFinalGateSitsAtTheWriteBoundary(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "reconcile.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	for _, d := range file.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name.Name != "Reconcile" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				switch f := c.Fun.(type) {
				case *ast.Ident:
					calls = append(calls, f.Name)
				case *ast.SelectorExpr:
					calls = append(calls, f.Sel.Name)
				}
			}
			return true
		})
	}
	last := func(name string) int {
		at := -1
		for i, c := range calls {
			if c == name {
				at = i
			}
		}
		return at
	}
	lock, encode, gate, write := last("lockRecordForAppend"), last("encodeReconcileLine"), last("checkReconcileWindow"), last("appendLine")
	if lock < 0 || encode < 0 || gate < 0 || write < 0 {
		t.Fatalf("Reconcile's write-boundary landmarks are missing: %v", calls)
	}
	if !(lock < encode && encode < gate && gate < write) {
		t.Fatalf("want lock → encode → final gate → write; calls %v", calls)
	}
	for _, c := range calls[gate+1 : write] {
		t.Fatalf("only the write may follow the final gate; found %s in between", c)
	}
	// 잠근 뒤의 판정 바이트는 잠근 fd 에서 읽는다(CG-2 — 판정한 바이트 = 쓸 파일). 경로로 다시 읽으면 잠근 파일과 다른
	// 파일을 판정할 수 있다.
	read := last("contents")
	if !(lock < read && read < encode) {
		t.Fatalf("the judged bytes must be read from the locked descriptor between the lock and the encode; calls %v", calls)
	}
	for _, c := range calls[lock:write] {
		if c == "readRecordRaw" || c == "readRecordStrictNoTail" || c == "readRecordStrict" {
			t.Fatalf("after the lock the record is re-read by path (%s), not through the locked descriptor", c)
		}
	}
}

// TestLockedAppendRefusesWhenTheRecordGrewAfterTheLockedRead 는 CG-2 의 쓰기 전 크기 검사를 층을 내려 잰다 — 잠근 fd 로
// 판정한 뒤 잠그지 않는 작성자가 덧붙이면, 쓰기 전에 거절하고 아무것도 쓰지 않는다(read-back 은 쓴 **뒤**에 잡으므로
// 이 검사가 없으면 대사 줄이 남의 줄 뒤에 남는다 — 변이 CG2-size-check).
func TestLockedAppendRefusesWhenTheRecordGrewAfterTheLockedRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), RecordFileName(MarketKR))
	rcWriteEntries(t, path, rcA063Entries())
	lock, err := lockRecordForAppend(path)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.release()
	judged, err := lock.contents()
	if err != nil {
		t.Fatal(err)
	}
	foreign := `{"format_version":1,"kind":"step","step_id":"costs","verdict":"skipped"}` + "\n"
	rcAppendRaw(t, path, foreign)
	line := []byte(`{"format_version":1,"kind":"reconcile"}` + "\n")
	if err := lock.appendLine(line, int64(len(judged))); !errors.Is(err, errRecordSizeMoved) {
		t.Fatalf("an append after the record grew must be refused before writing; got %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(judged)+foreign {
		t.Fatal("the refused append still wrote to the record")
	}
}

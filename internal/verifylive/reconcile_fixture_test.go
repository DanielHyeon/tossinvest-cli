package verifylive

// reconcile_fixture_test.go — a121 RED 시험의 공용 부품.
//
// 가짜 읽기(rcFakeReader)·고정 시계·a063 형 기록·설계 모양 대사 줄을 한 곳에 둔다. 거절 시험은 전부 같은 수락
// 픽스처(newRCHarness 기본값 = 모든 조건 참)에서 **하나만** 비틀어, 거절이 그 가드에서만 나올 수 있게 한다.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const (
	rcSymbol   = "005930"
	rcAccount  = "123-45-678901" // 시험용 원문 계좌 참조 — 기록에는 마스크만 남는다
	rcTargetID = "CO-A063-TARGET"
	rcChainID  = "chain-a063"
)

var (
	rcNow     = time.Date(2026, 10, 5, 1, 0, 0, 0, time.UTC)
	rcCreated = rcNow.Add(-2 * time.Hour)
)

// 기본 정책 seam — 수락 픽스처가 Q1·Q3 를 통과하도록 둔다(값 자체는 시험용, 제안값과 무관).
const (
	rcRetentionBound = 24 * time.Hour
	rcFreshness      = 30 * time.Second
)

func rcMask() string { return attest.Mask(rcAccount) }

// rcTarget 은 대사 대상 artifact(조건주문 등록 줄의 모양)다.
func rcTarget() Artifact {
	return Artifact{Kind: KindConditional, ID: rcTargetID, Symbol: rcSymbol, CreatedAt: rcCreated,
		Deliberate: true, ChainID: rcChainID}
}

// rcA063Entries 는 a063 형 기록이다: 조건주문 등록(보유) → 존속 확인 → 조건주문 취소 실패(DELETE 404, 보유 해제)
// → 재개 정리 prologue 실패(같은 DELETE 404). 대상은 PendingCleanup 이 내는 유일한 조건주문이다.
func rcA063Entries() []Entry {
	acct := rcMask()
	notFound := Call{Endpoint: EndpointCancelConditional, At: rcCreated.Add(30 * time.Minute), LatencyMS: 40,
		Error: "official: HTTP 404 conditional-order-not-found", ErrorCode: "conditional-order-not-found"}
	return []Entry{
		{Kind: KindStep, RunID: "run-1", StepID: StepReadFixtures, Verdict: VerdictPass, AccountRef: acct,
			StartedAt: rcCreated.Add(-time.Minute), FinishedAt: rcCreated.Add(-time.Minute),
			Calls: []Call{{Endpoint: "GET /api/v1/orders", At: rcCreated.Add(-time.Minute), LatencyMS: 30}}},
		{Kind: KindStep, RunID: "run-1", StepID: StepConditionalRegister, Verdict: VerdictPass, Mutating: true, AccountRef: acct,
			StartedAt: rcCreated, FinishedAt: rcCreated,
			Calls:     []Call{{Endpoint: EndpointCreateConditional, At: rcCreated, LatencyMS: 50}},
			Artifacts: []Artifact{rcTarget()}},
		{Kind: KindStep, RunID: "run-2", StepID: StepConditionalPersist, Verdict: VerdictPass, AccountRef: acct,
			StartedAt: rcCreated.Add(10 * time.Minute), FinishedAt: rcCreated.Add(10 * time.Minute)},
		{Kind: KindStep, RunID: "run-2", StepID: StepConditionalCancel, Verdict: VerdictFail, Mutating: true, AccountRef: acct,
			Reason: "conditional-order-not-found", StartedAt: notFound.At, FinishedAt: notFound.At, Calls: []Call{notFound}},
		{Kind: KindCleanup, RunID: "run-3", StepID: StepCleanup, Verdict: VerdictFail, Mutating: true, AccountRef: acct,
			Reason: "conditional-order-not-found", StartedAt: notFound.At.Add(time.Hour), FinishedAt: notFound.At.Add(time.Hour),
			Calls: []Call{{Endpoint: EndpointCancelConditional, At: notFound.At.Add(time.Hour), LatencyMS: 41,
				Error: notFound.Error, ErrorCode: notFound.ErrorCode}}},
	}
}

func rcWriteEntries(t *testing.T, path string, entries []Entry) {
	t.Helper()
	rec, err := OpenRecorder(path)
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Close()
	for _, e := range entries {
		if err := rec.Append(e); err != nil {
			t.Fatal(err)
		}
	}
}

func rcAppendRaw(t *testing.T, path string, raw string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(raw); err != nil {
		t.Fatal(err)
	}
}

// rcReconcileLineJSON 은 design G2 「결정(골격)」·freeze P2-6·재검 P2-c 가 정한 대사 줄의 모양을 손으로 만든 것이다.
// 투영 시험(2.1·2.3·2.4·S1)은 Reconcile 구현과 독립적으로 이 줄을 쓰고, TestReconcileLineMatchesTheDesignShape 가
// Reconcile 이 실제로 쓰는 줄이 이 모양과 같음을 묶는다.
func rcReconcileLineJSON(target Artifact, at time.Time, basis string) string {
	art := map[string]any{
		"kind": target.Kind, "id": target.ID, "symbol": target.Symbol,
		"created_at": target.CreatedAt, "cancelled": false,
		"reconciled_absent": true, "reconciled_at": at,
	}
	if target.ChainID != "" {
		art["chain_id"] = target.ChainID
	}
	if target.Deliberate {
		art["deliberate"] = true
	}
	line := map[string]any{
		"format_version": RecordFormatVersion, "kind": KindReconcile, "run_id": "reconcile-1",
		"process": map[string]any{"pid": 1, "instance_id": "proc-test", "started_at": at},
		"step_id": string(StepReconcile), "title": "reconciled absent",
		"started_at": at, "finished_at": at,
		"account_ref": rcMask(), "mutating": false, "verdict": "",
		"observations": []map[string]any{{"key": ObservationReconcileBasis, "value": basis}},
		"artifacts":    []map[string]any{art},
	}
	b, err := json.Marshal(line)
	if err != nil {
		panic(err)
	}
	return string(b) + "\n"
}

// rcEntriesWithReconcileLine 는 a063 기록 뒤에 설계 모양 대사 줄을 붙여 LoadEntries 로 다시 읽은 결과다.
func rcEntriesWithReconcileLine(t *testing.T, entries []Entry) []Entry {
	t.Helper()
	path := filepath.Join(t.TempDir(), RecordFileName(MarketKR))
	rcWriteEntries(t, path, entries)
	rcAppendRaw(t, path, rcReconcileLineJSON(rcTarget(), rcNow, ReconcileBasisDomain+":sha256:00"))
	out, err := LoadEntries(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(entries)+1 {
		t.Fatalf("hand-built reconcile line did not decode: %d entries", len(out))
	}
	return out
}

// rcLoad 는 entries 를 기록 파일로 쓰고 LoadEntries 로 다시 읽은 결과다(Append 가 채우는 기본값까지 같은 모양).
func rcLoad(t *testing.T, entries []Entry) []Entry {
	t.Helper()
	path := filepath.Join(t.TempDir(), RecordFileName(MarketKR))
	rcWriteEntries(t, path, entries)
	out, err := LoadEntries(path)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// --- 가짜 읽기 ---------------------------------------------------------------------

type rcPage struct {
	cond  official.ReconcileConditionalPage
	plain official.ReconcileOrderPage
	err   error
}

type rcCall struct {
	Kind   string // "list" | "instrument"
	Group  string
	Symbol string
	Cursor string
	Limit  int
	Pass   int
}

type rcFakeReader struct {
	mu     sync.Mutex
	script [2]map[string][]rcPage
	served [2]map[string]int
	pass   int
	calls  []rcCall
	// instrument 은 n 번째(0 부터) 종목 조회의 응답을 정한다. nil 이면 심볼을 그대로 되돌린다.
	instrument  func(n int, symbol string) (string, error)
	instrumentN int
	// hook 은 응답 직전에 불린다 — 시계를 움직이거나 기록 파일을 바꾸는 데 쓴다.
	hook func(c rcCall)
}

func newRCFakeReader() *rcFakeReader {
	r := &rcFakeReader{pass: -1}
	for p := 0; p < 2; p++ {
		r.script[p] = map[string][]rcPage{}
		r.served[p] = map[string]int{}
		for _, g := range []string{ReconcileGroupConditionalOpen, ReconcileGroupPlainOpen, ReconcileGroupConditionalClosed} {
			r.script[p][g] = []rcPage{{}}
		}
	}
	return r
}

// set 은 pass(0·1, -1 = 둘 다)의 group 페이지 열을 정한다.
func (r *rcFakeReader) set(pass int, group string, pages ...rcPage) {
	for p := 0; p < 2; p++ {
		if pass == -1 || pass == p {
			r.script[p][group] = pages
		}
	}
}

func (r *rcFakeReader) record(c rcCall) rcCall {
	r.mu.Lock()
	c.Pass = r.pass
	r.calls = append(r.calls, c)
	hook := r.hook
	r.mu.Unlock()
	if hook != nil {
		hook(c)
	}
	return c
}

func (r *rcFakeReader) next(group string) rcPage {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.pass < 0 || r.pass > 1 {
		return rcPage{err: fmt.Errorf("fake: list read outside the two read passes (pass %d)", r.pass)}
	}
	pages := r.script[r.pass][group]
	i := r.served[r.pass][group]
	r.served[r.pass][group] = i + 1
	if i >= len(pages) {
		return rcPage{err: fmt.Errorf("fake: unscripted page %d of %s in pass %d", i, group, r.pass)}
	}
	return pages[i]
}

func (r *rcFakeReader) ReconcileConditionalOrdersPage(ctx context.Context, status, symbol, cursor string, limit int) (official.ReconcileConditionalPage, error) {
	group := "conditional-" + strings.ToLower(status)
	switch status {
	case "OPEN":
		group = ReconcileGroupConditionalOpen
		if cursor == "" {
			r.mu.Lock()
			r.pass++
			r.mu.Unlock()
		}
	case "CLOSED":
		group = ReconcileGroupConditionalClosed
	}
	r.record(rcCall{Kind: "list", Group: group, Symbol: symbol, Cursor: cursor, Limit: limit})
	p := r.next(group)
	return p.cond, p.err
}

func (r *rcFakeReader) ReconcileOpenOrdersPage(ctx context.Context, symbol, cursor string, limit int) (official.ReconcileOrderPage, error) {
	r.record(rcCall{Kind: "list", Group: ReconcileGroupPlainOpen, Symbol: symbol, Cursor: cursor, Limit: limit})
	p := r.next(ReconcileGroupPlainOpen)
	return p.plain, p.err
}

func (r *rcFakeReader) ReconcileInstrument(ctx context.Context, symbol string) (string, error) {
	r.record(rcCall{Kind: "instrument", Symbol: symbol})
	r.mu.Lock()
	n := r.instrumentN
	r.instrumentN++
	fn := r.instrument
	r.mu.Unlock()
	if fn != nil {
		return fn(n, symbol)
	}
	return symbol, nil
}

func (r *rcFakeReader) listCalls() []rcCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []rcCall
	for _, c := range r.calls {
		if c.Kind == "list" {
			out = append(out, c)
		}
	}
	return out
}

func (r *rcFakeReader) allCalls() []rcCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]rcCall(nil), r.calls...)
}

func rcCondRow(id, status string) official.ReconcileConditionalRow {
	return official.ReconcileConditionalRow{ID: id, Symbol: rcSymbol, Market: MarketKR, Status: status}
}

func rcCondPage(next string, hasNext bool, rows ...official.ReconcileConditionalRow) rcPage {
	return rcPage{cond: official.ReconcileConditionalPage{Rows: rows, NextCursor: next, HasNext: hasNext}}
}

func rcPlainPage(next string, hasNext bool, rows ...official.ReconcileOrderRow) rcPage {
	return rcPage{plain: official.ReconcileOrderPage{Rows: rows, NextCursor: next, HasNext: hasNext}}
}

// --- 고정 시계 ---------------------------------------------------------------------

type rcClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *rcClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *rcClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// --- 하네스 ------------------------------------------------------------------------

type rcHarness struct {
	t       *testing.T
	path    string
	reader  *rcFakeReader
	clock   *rcClock
	account ReconcileAccount
	out     bytes.Buffer

	approvals           []ReconcileApproval
	outAtApproval       string
	listCallsAtApproval int
	approveErr          error
	onApprove           func()
}

// rcSetPolicies 는 Q1·Q3 seam 을 시험 동안만 바꾼다.
func rcSetPolicies(t *testing.T, retention *retentionMeasurement, freshness time.Duration) {
	t.Helper()
	prevR, prevF := reconcileRetention, reconcileFreshnessBound
	reconcileRetention, reconcileFreshnessBound = retention, freshness
	t.Cleanup(func() { reconcileRetention, reconcileFreshnessBound = prevR, prevF })
}

// newRCHarness 는 모든 조건이 참인 수락 픽스처를 만든다.
func newRCHarness(t *testing.T, entries []Entry) *rcHarness {
	t.Helper()
	rcSetPolicies(t, &retentionMeasurement{Bound: rcRetentionBound, Eviction: evictionDuration}, rcFreshness)
	h := &rcHarness{
		t:       t,
		path:    filepath.Join(t.TempDir(), RecordFileName(MarketKR)),
		reader:  newRCFakeReader(),
		clock:   &rcClock{now: rcNow},
		account: ReconcileAccount{Ref: rcAccount, Seq: 7, Count: 1},
	}
	rcWriteEntries(t, h.path, entries)
	return h
}

func (h *rcHarness) params() ReconcileParams {
	return ReconcileParams{
		RecordPath: h.path,
		Market:     MarketKR,
		Account:    h.account,
		Reader:     h.reader,
		Out:        &h.out,
		Now:        h.clock.Now,
		Approve: func(ctx context.Context, a ReconcileApproval) error {
			h.approvals = append(h.approvals, a)
			h.outAtApproval = h.out.String()
			h.listCallsAtApproval = len(h.reader.listCalls())
			if h.onApprove != nil {
				h.onApprove()
			}
			return h.approveErr
		},
	}
}

func (h *rcHarness) run() (ReconcileResult, error) {
	return Reconcile(context.Background(), h.params())
}

func (h *rcHarness) bytes() []byte {
	h.t.Helper()
	b, err := os.ReadFile(h.path)
	if err != nil {
		h.t.Fatal(err)
	}
	return b
}

// rcRequireRefusal 은 err 가 정확히 그 가드의 거절인지 단언한다(공유 문구가 아닌 코드).
func rcRequireRefusal(t *testing.T, err error, want ReconcileRefusalCode) *ReconcileRefusal {
	t.Helper()
	var refusal *ReconcileRefusal
	if !errors.As(err, &refusal) {
		t.Fatalf("want refusal %q, got %v", want, err)
	}
	if refusal.Code != want {
		t.Fatalf("want refusal %q, got %q (%s)", want, refusal.Code, refusal.Detail)
	}
	return refusal
}

// rcRequireUnchanged 는 거절이 기록에 한 바이트도 쓰지 않았음을 단언한다.
func rcRequireUnchanged(t *testing.T, h *rcHarness, before []byte) {
	t.Helper()
	if after := h.bytes(); !bytes.Equal(before, after) {
		t.Fatalf("a refusal changed the record:\nbefore %d bytes\nafter  %d bytes\n%s", len(before), len(after), after[len(before):])
	}
}

// rcRequireNoListRead 는 주문·조건주문 목록 읽기가 하나도 없었음을 단언한다(계좌·종목 조회는 허용).
func rcRequireNoListRead(t *testing.T, h *rcHarness) {
	t.Helper()
	if calls := h.reader.listCalls(); len(calls) != 0 {
		t.Fatalf("refusal must come before any order-list or conditional-order-list read; saw %d: %+v", len(calls), calls)
	}
}

// rcRefuse 는 "수락 픽스처에서 하나를 비틀면 그 가드가 거절하고 기록은 그대로" 를 한 번에 잰다.
func rcRefuse(t *testing.T, h *rcHarness, want ReconcileRefusalCode) *ReconcileRefusal {
	t.Helper()
	before := h.bytes()
	_, err := h.run()
	r := rcRequireRefusal(t, err, want)
	rcRequireUnchanged(t, h, before)
	return r
}

// rcAppendedLines 는 수락 뒤 기록에 새로 붙은 줄(원문)들이다.
func rcAppendedLines(t *testing.T, before, after []byte) []string {
	t.Helper()
	if !bytes.HasPrefix(after, before) {
		t.Fatalf("the record is append-only, but its original bytes changed")
	}
	tail := strings.TrimRight(string(after[len(before):]), "\n")
	if tail == "" {
		return nil
	}
	return strings.Split(tail, "\n")
}

// rcAccept 은 수락을 기대하고 붙은 줄 하나(해독한 map)를 돌려준다.
func rcAccept(t *testing.T, h *rcHarness) (ReconcileResult, map[string]any, string) {
	t.Helper()
	before := h.bytes()
	res, err := h.run()
	if err != nil {
		t.Fatalf("every condition holds, so reconciliation must append; got %v", err)
	}
	lines := rcAppendedLines(t, before, h.bytes())
	if len(lines) != 1 {
		t.Fatalf("want exactly one appended line, got %d", len(lines))
	}
	var line map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &line); err != nil {
		t.Fatalf("appended line is not JSON: %v", err)
	}
	return res, line, lines[0]
}

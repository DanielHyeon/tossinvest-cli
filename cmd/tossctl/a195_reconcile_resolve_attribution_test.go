package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a195 — 귀속 실패(ATTRIBUTION_FAILED) 종목 차단의 운영자 해제. 픽스처는 합성 값만 씀.

const (
	a195Operator = "ops-1"
	a195Note     = "fresh official snapshots agree"
	a195Refusal  = "requires cause-specific recovery evidence"
)

func a195Block(cause, symbol string) reconcile.Block {
	scope := reconcile.ScopeSymbol
	if symbol == "" {
		scope = reconcile.ScopeAccount
	}
	return reconcile.Block{Scope: scope, Account: "acct-a195", Symbol: symbol, Cause: cause}
}

// 원인 표를 헬퍼에 직접 고정함 — 허용 목록이 이 한 곳에 있으므로 표가 곧 계약임.
func TestA195RefusedCauseTable(t *testing.T) {
	qm := a195Block(journal.ReconcileCauseQuantityMismatch, "SYMA")
	af := a195Block(journal.ReconcileCauseAttributionFailed, "SYMB")
	ic := a195Block(journal.ReconcileCauseIdentifierConflict, "")
	accountAF := a195Block(journal.ReconcileCauseAttributionFailed, "")
	cases := []struct {
		name   string
		blocks []reconcile.Block
		refuse string // 빈 값이면 통과해야 함
	}{
		{"quantity mismatch only", []reconcile.Block{qm}, ""},
		{"symbol attribution failure only", []reconcile.Block{af}, ""},
		{"mismatch then attribution", []reconcile.Block{qm, af}, ""},
		{"attribution then mismatch", []reconcile.Block{af, qm}, ""},
		{"identifier conflict only", []reconcile.Block{ic}, journal.ReconcileCauseIdentifierConflict},
		{"attribution then identifier conflict", []reconcile.Block{af, ic}, journal.ReconcileCauseIdentifierConflict},
		{"identifier conflict then attribution", []reconcile.Block{ic, af}, journal.ReconcileCauseIdentifierConflict},
		{"account-wide attribution failure", []reconcile.Block{accountAF}, journal.ReconcileCauseAttributionFailed},
		{"empty list is left to the caller", nil, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := reconcileResolveRefusedCause(tc.blocks)
			if tc.refuse == "" {
				if err != nil {
					t.Fatalf("refused releasable causes: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.refuse) || !strings.Contains(err.Error(), a195Refusal) {
				t.Fatalf("error = %v, want refusal naming %s with %q", err, tc.refuse, a195Refusal)
			}
		})
	}
}

// a195Seed 는 장부에 넣을 활성 RECONCILE 행 하나임.
type a195Seed struct{ symbol, cause string }

// a195ProductionRuntime 은 production buildReconcileResolveRuntime 이 돌려준 런타임을 씀 —
// validate·resolve·localState·compare 는 production 그대로, 공식 스냅숏 수집만 가짜로 바꿈.
func a195ProductionRuntime(t *testing.T, seeds []a195Seed, holdings []reconcile.Holding) (reconcileResolveRuntime, string, string) {
	t.Helper()
	dir := isolate(t)
	writeConfigFile(t, dir, config.DefaultFile())
	writeCredentials(t, dir)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/accounts":
			fmt.Fprint(w, `{"result":[{"accountNo":"fake","accountSeq":1,"accountType":"BROKERAGE"}]}`)
		default:
			t.Errorf("unexpected local path %s", r.URL.Path)
			http.Error(w, "unexpected", http.StatusInternalServerError)
		}
	}))
	t.Cleanup(server.Close)
	transport := server.Client().Transport.(*http.Transport).Clone()
	transport.Proxy = nil
	root := &rootOptions{configDir: dir, officialOptions: []official.Option{
		official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: transport}),
	}}
	ctx := context.Background()

	// 계좌 신원은 production 해석 결과를 그대로 씀 — 첫 조립으로 읽고 닫음.
	probe, err := buildReconcileResolveRuntime(ctx, root)
	if err != nil {
		t.Fatalf("probe build: %v", err)
	}
	account := probe.accountRef
	if err := probe.close(); err != nil {
		t.Fatalf("probe close: %v", err)
	}

	path := filepath.Join(dir, journal.DBFileName)
	j, err := journal.Open(ctx, journal.Options{Path: path})
	if err != nil {
		t.Fatalf("open journal: %v", err)
	}
	for _, seed := range seeds {
		if _, entered, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
			AccountRef: account, Symbol: seed.symbol, Cause: seed.cause, Evidence: "synthetic a195 evidence",
		}); err != nil || !entered {
			t.Fatalf("seed %s/%s: entered=%v err=%v", seed.symbol, seed.cause, entered, err)
		}
	}
	if err := j.Close(); err != nil {
		t.Fatalf("close seeded journal: %v", err)
	}

	rt, err := buildReconcileResolveRuntime(ctx, root)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	first := time.Date(2026, 8, 3, 4, 0, 0, 0, time.UTC)
	next := 0
	rt.collect = func(context.Context) (reconcile.Snapshot, error) {
		at := first.Add(time.Duration(next) * reconcile.DefaultStabilisationInterval)
		next++
		return reconcile.Snapshot{AsOf: at, CompletedAt: at.Add(time.Millisecond), AccountRef: account,
			Holdings: append([]reconcile.Holding(nil), holdings...)}, nil
	}
	return rt, account, path
}

func a195Run(t *testing.T, rt reconcileResolveRuntime) (string, error) {
	t.Helper()
	released := false
	return executeReconcileResolve(t, reconcileResolveDeps{
		journalDir: func(*rootOptions) (string, error) { return "/journal", nil },
		acquire: func(string) (reconcileResolveLock, error) {
			return &reconcileResolveTestLock{path: "/journal/engine.lock", released: &released}, nil
		},
		build: func(context.Context, *rootOptions) (reconcileResolveRuntime, error) { return rt, nil },
		sleep: func(context.Context, time.Duration) error { return nil },
	}, "--confirm", "--operator", a195Operator, "--note", a195Note)
}

// a195States 는 장부의 그 계좌 RECONCILE 이력을 종목→행으로 돌려줌(계좌 전체는 빈 키).
func a195States(t *testing.T, path, account string) map[string]journal.ReconcileState {
	t.Helper()
	ctx := context.Background()
	j, err := journal.Open(ctx, journal.Options{Path: path})
	if err != nil {
		t.Fatalf("reopen journal: %v", err)
	}
	defer j.Close()
	history, err := j.ReconcileStateHistory(ctx, account)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	out := map[string]journal.ReconcileState{}
	for _, state := range history {
		out[state.Symbol+"|"+state.Cause] = state
	}
	tracker := &reconcile.Tracker{Journal: j, AccountRef: account}
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("restore: %v", err)
	}
	out["__restored_blocks__"] = journal.ReconcileState{Evidence: fmt.Sprint(len(tracker.Blocks()))}
	return out
}

func a195AssertOperatorReleased(t *testing.T, state journal.ReconcileState) {
	t.Helper()
	if state.Active() || state.ReleaseCause != journal.ReconcileReleaseOperator ||
		!strings.Contains(state.Evidence, "operator "+a195Operator+": "+a195Note) {
		t.Fatalf("state = %+v, want OPERATOR release carrying the operator and note", state)
	}
}

func a195AssertActive(t *testing.T, state journal.ReconcileState) {
	t.Helper()
	if !state.Active() || state.ReleaseCause != "" {
		t.Fatalf("state = %+v, want untouched active row", state)
	}
}

func TestA195OperatorReleasesASymbolAttributionFailure(t *testing.T) {
	rt, account, path := a195ProductionRuntime(t, []a195Seed{{"SYMB", journal.ReconcileCauseAttributionFailed}}, nil)
	if _, err := a195Run(t, rt); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	states := a195States(t, path, account)
	a195AssertOperatorReleased(t, states["SYMB|"+journal.ReconcileCauseAttributionFailed])
	if got := states["__restored_blocks__"].Evidence; got != "0" {
		t.Fatalf("restored blocks after release = %s, want 0", got)
	}
}

func TestA195BlockingDiffStillRefusesAnAttributionFailure(t *testing.T) {
	rt, account, path := a195ProductionRuntime(t, []a195Seed{{"SYMB", journal.ReconcileCauseAttributionFailed}}, nil)
	rt.compare = func(reconcile.Snapshot, reconcile.LocalState) reconcile.Diff {
		return reconcile.Diff{AccountRef: account, Quantities: []reconcile.QuantityMismatch{{Symbol: "SYMC", Local: "1", Broker: "2"}}}
	}
	_, err := a195Run(t, rt)
	if err == nil || !strings.Contains(err.Error(), "blocking reconciliation diff remains") {
		t.Fatalf("error = %v, want blocking-diff refusal", err)
	}
	_ = rt.close()
	a195AssertActive(t, a195States(t, path, account)["SYMB|"+journal.ReconcileCauseAttributionFailed])
}

func TestA195MismatchAndAttributionReleaseTogether(t *testing.T) {
	rt, account, path := a195ProductionRuntime(t, []a195Seed{
		{"SYMA", journal.ReconcileCauseQuantityMismatch},
		{"SYMB", journal.ReconcileCauseAttributionFailed},
	}, nil)
	if _, err := a195Run(t, rt); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	states := a195States(t, path, account)
	a195AssertOperatorReleased(t, states["SYMA|"+journal.ReconcileCauseQuantityMismatch])
	a195AssertOperatorReleased(t, states["SYMB|"+journal.ReconcileCauseAttributionFailed])
}

func TestA195IdentifierConflictKeepsEveryRowBlocked(t *testing.T) {
	rt, account, path := a195ProductionRuntime(t, []a195Seed{
		{"SYMB", journal.ReconcileCauseAttributionFailed},
		{"", journal.ReconcileCauseIdentifierConflict},
	}, nil)
	_, err := a195Run(t, rt)
	if err == nil || !strings.Contains(err.Error(), journal.ReconcileCauseIdentifierConflict) || !strings.Contains(err.Error(), a195Refusal) {
		t.Fatalf("error = %v, want refusal naming IDENTIFIER_CONFLICT", err)
	}
	_ = rt.close()
	states := a195States(t, path, account)
	a195AssertActive(t, states["SYMB|"+journal.ReconcileCauseAttributionFailed])
	a195AssertActive(t, states["|"+journal.ReconcileCauseIdentifierConflict])
}

func TestA195AccountWideAttributionFailureIsRefused(t *testing.T) {
	rt, account, path := a195ProductionRuntime(t, []a195Seed{{"", journal.ReconcileCauseAttributionFailed}}, nil)
	_, err := a195Run(t, rt)
	if err == nil || !strings.Contains(err.Error(), a195Refusal) {
		t.Fatalf("error = %v, want cause-specific refusal", err)
	}
	_ = rt.close()
	a195AssertActive(t, a195States(t, path, account)["|"+journal.ReconcileCauseAttributionFailed])
}

// 외부 보유가 있어도 진입을 막는 차이가 아니면 해제됨(design D2) — 실제 Comparer 로 입력을 만듦.
// 장부에 그 종목 인스턴스가 없는 것과 CLOSED 인스턴스만 있는 것은 HeldBySymbol 이 같은 입력으로 만듦(compare.go:155).
func TestA195ExternalHoldingDoesNotRefuseTheRelease(t *testing.T) {
	holdings := []reconcile.Holding{{Symbol: "SYMB", Quantity: "3"}}
	rt, account, path := a195ProductionRuntime(t, []a195Seed{{"SYMB", journal.ReconcileCauseAttributionFailed}}, holdings)
	local, err := rt.localState(context.Background(), account)
	if err != nil {
		t.Fatalf("local state: %v", err)
	}
	diff := rt.compare(reconcile.Snapshot{AccountRef: account, Holdings: holdings}, local)
	if diff.BlocksEntry() || len(diff.ExternalPos) != 1 || diff.ExternalPos[0].Symbol != "SYMB" {
		t.Fatalf("diff = %+v, want one non-blocking external holding on SYMB", diff)
	}
	if _, err := a195Run(t, rt); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	a195AssertOperatorReleased(t, a195States(t, path, account)["SYMB|"+journal.ReconcileCauseAttributionFailed])
}

// production 클로저의 빈 블록 검사(B12)를 그대로 지남.
func TestA195NoActiveBlockIsRefused(t *testing.T) {
	rt, _, _ := a195ProductionRuntime(t, nil, nil)
	_, err := a195Run(t, rt)
	if err == nil || !strings.Contains(err.Error(), "no active RECONCILE block exists") {
		t.Fatalf("error = %v, want no-active-block refusal", err)
	}
	_ = rt.close()
}

// 귀속 실패는 대사의 차이 없는 비교로 스스로 풀리지 않음 — 기존 동작을 고정함.
func TestA195ObserveNeverReleasesAnAttributionFailure(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), journal.DBFileName)
	j, err := journal.Open(ctx, journal.Options{Path: path})
	if err != nil {
		t.Fatalf("open journal: %v", err)
	}
	defer j.Close()
	if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
		AccountRef: "acct-a195", Symbol: "SYMB", Cause: journal.ReconcileCauseAttributionFailed, Evidence: "synthetic a195 evidence",
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	tracker := &reconcile.Tracker{Journal: j, AccountRef: "acct-a195"}
	if err := tracker.Restore(ctx); err != nil {
		t.Fatalf("restore: %v", err)
	}
	at := time.Date(2026, 8, 3, 4, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		clean := reconcile.Diff{AccountRef: "acct-a195", AsOf: at.Add(time.Duration(i) * time.Minute).Format(time.RFC3339)}
		if _, err := tracker.Observe(ctx, clean); err != nil {
			t.Fatalf("observe %d: %v", i, err)
		}
	}
	blocks := tracker.Blocks()
	if len(blocks) != 1 || blocks[0].Cause != journal.ReconcileCauseAttributionFailed {
		t.Fatalf("blocks after clean observations = %+v, want the attribution failure", blocks)
	}
	active, err := j.ActiveReconcileStates(ctx)
	if err != nil || len(active) != 1 {
		t.Fatalf("active rows = %+v err=%v, want the attribution failure row", active, err)
	}
}

// 도움말과 성공 출력은 원인 중립이어야 함(design D7) — 「released」는 기존 시험이 단언하므로 남김.
func TestA195HelpAndSuccessOutputAreCauseNeutral(t *testing.T) {
	cmd := newEngineReconcileResolveCmd(&rootOptions{})
	if strings.Contains(cmd.Long, "quantity-mismatch") || strings.Contains(cmd.Short, "quantity-mismatch") {
		t.Fatalf("help still names only quantity mismatch: %q", cmd.Long)
	}
	rt, _, _ := a195ProductionRuntime(t, []a195Seed{{"SYMB", journal.ReconcileCauseAttributionFailed}}, nil)
	output, err := a195Run(t, rt)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(output, "released") || strings.Contains(output, "quantity-mismatch") {
		t.Fatalf("output = %q, want a cause-neutral release acknowledgement", output)
	}
}

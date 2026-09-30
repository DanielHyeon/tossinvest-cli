package reconcile_test

// a094 4.N4 · 4.N4a · 4.N4b · 4.N4d · 4.N4f① — 기동은 ACKED 발주를 기록 번호의 바이트 일치로만 확정하고, 나머지는 상태를
// 바꾸지 않고 attempt 를 이름으로 알린다. 어느 실패도 복구를 실패시키지 않는다.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a094CountingOrders 는 주문 읽기 · 목록 조회를 셈.
type a094CountingOrders struct {
	single map[string]json.RawMessage
	fail   error
	reads  int
	lists  int
}

func (o *a094CountingOrders) OrderRaw(_ context.Context, id string) (json.RawMessage, error) {
	o.reads++
	if o.fail != nil {
		return nil, o.fail
	}
	if raw, ok := o.single[id]; ok {
		return raw, nil
	}
	return nil, errors.New("no such order")
}

func (o *a094CountingOrders) OrdersPageRaw(context.Context, execgw.OrderQuery, string) (execgw.OrderPage, error) {
	o.lists++
	return execgw.OrderPage{}, nil
}

type a094Alerts struct {
	events []obs.Event
	remind []time.Duration
}

func (a *a094Alerts) RecordCritical(_ context.Context, e obs.Event, r time.Duration) error {
	a.events = append(a.events, e)
	a.remind = append(a.remind, r)
	return nil
}

// a094CrashAfterAck 는 브로커가 접수(번호 기록)한 뒤 확정 전에 죽은 attempt 를 남김.
func a094CrashAfterAck(t *testing.T, path string, kind journal.MutationKind, orderID string) {
	t.Helper()
	ctx := context.Background()
	j := openJournalAt(t, path)
	req := journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-acked", Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-7",
			Symbol: "AAPL", Side: "SELL", OrderType: "LIMIT", Quantity: "10", Price: "200",
			Currency: "USD", Source: "engine", Fingerprint: "fp-acked",
		},
		Kind: kind, AttemptID: "attempt-acked",
	}
	if kind != journal.KindPlace {
		req.TargetOrderID = "O-target"
	}
	attempt, err := j.Prepare(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
}

func a094Recover(t *testing.T, path string, orders *a094CountingOrders, alerts *a094Alerts) (*journal.Journal, reconcile.Report) {
	t.Helper()
	j := openJournalAt(t, path)
	t.Cleanup(func() { _ = j.Close() })
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	opts := recoveryOptions(j, gate, recoveryCollector(nil, nil))
	opts.Resolver.Order = orders
	opts.Resolver.Orders = orders
	opts.Alerts = alerts
	r, err := reconcile.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v — no ACKED outcome may fail the recovery (every loop would stay down)", err)
	}
	return j, report
}

func a094Detail(id, symbol string) json.RawMessage {
	return json.RawMessage(`{"result":{"orderId":"` + id + `","symbol":"` + symbol + `","side":"SELL","status":"PENDING",` +
		`"quantity":"10","price":"200","currency":"USD","orderedAt":"2026-03-30T10:30:00-04:00","execution":{"filledQuantity":"0"}}}`)
}

// 4.N4 — 같은 번호 바이트 일치 + 같은 종목 → CONFIRMED, 알림 0, 읽기 정확히 1.
func TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	a094CrashAfterAck(t, path, journal.KindPlace, "O-acked")
	orders := &a094CountingOrders{single: map[string]json.RawMessage{"O-acked": a094Detail("O-acked", "AAPL")}}
	alerts := &a094Alerts{}
	j, report := a094Recover(t, path, orders, alerts)

	stored, err := j.LookupAttempt(context.Background(), "attempt-acked")
	if err != nil {
		t.Fatal(err)
	}
	if stored.State != journal.StateConfirmed || stored.BrokerOrderID != "O-acked" {
		t.Fatalf("state/order = %s/%s, want CONFIRMED/O-acked", stored.State, stored.BrokerOrderID)
	}
	if orders.reads != 1 || orders.lists != 0 {
		t.Errorf("reads/lists = %d/%d, want exactly one read and no list scan", orders.reads, orders.lists)
	}
	if len(alerts.events) != 0 {
		t.Errorf("alerts = %d, want 0 on a confirmation", len(alerts.events))
	}
	if len(report.ConfirmedAcked) != 1 || len(report.StillPending) != 0 {
		t.Errorf("report confirmed/pending = %v/%v", report.ConfirmedAcked, report.StillPending)
	}
}

// 4.N4a — 읽기 실패 · 다른 번호(대소문자 · 공백만 다른 번호 포함) · 다른 종목 · 해석 불가 → ACKED 그대로, 알림 1(attempt key),
// 해소기 · 목록 조회 0, 복구 성공.
func TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		orders *a094CountingOrders
	}{
		{"read fails", &a094CountingOrders{fail: errors.New("transport")}},
		{"other number", &a094CountingOrders{single: map[string]json.RawMessage{"O-acked": a094Detail("O-other", "AAPL")}}},
		{"number differs in case", &a094CountingOrders{single: map[string]json.RawMessage{"O-acked": a094Detail("o-acked", "AAPL")}}},
		{"number differs in space", &a094CountingOrders{single: map[string]json.RawMessage{"O-acked": a094Detail(" O-acked", "AAPL")}}},
		{"other symbol", &a094CountingOrders{single: map[string]json.RawMessage{"O-acked": a094Detail("O-acked", "MSFT")}}},
		{"empty symbol (4.N4e)", &a094CountingOrders{single: map[string]json.RawMessage{"O-acked": a094Detail("O-acked", "")}}},
		{"unreadable", &a094CountingOrders{single: map[string]json.RawMessage{"O-acked": json.RawMessage(`{"result":`)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "journal.db")
			a094CrashAfterAck(t, path, journal.KindPlace, "O-acked")
			alerts := &a094Alerts{}
			j, report := a094Recover(t, path, tc.orders, alerts)
			stored, _ := j.LookupAttempt(context.Background(), "attempt-acked")
			if stored.State != journal.StateAcked {
				t.Fatalf("state = %s, want ACKED untouched", stored.State)
			}
			if tc.orders.reads != 1 || tc.orders.lists != 0 {
				t.Errorf("reads/lists = %d/%d, want one read, no list", tc.orders.reads, tc.orders.lists)
			}
			if len(report.Resolutions) != 0 {
				t.Errorf("resolutions = %+v, want none — an ACKED attempt must not go to list matching", report.Resolutions)
			}
			a094AssertNamed(t, alerts, "attempt-acked")
		})
	}
}

// 4.N4b — CANCEL ACKED 와 번호 없는 PLACE ACKED: 읽기 0, 상태 무변경, 알림 1.
func TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed(t *testing.T) {
	for _, tc := range []struct {
		name  string
		kind  journal.MutationKind
		order string
	}{
		{"cancel", journal.KindCancel, "O-target"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "journal.db")
			a094CrashAfterAck(t, path, tc.kind, tc.order)
			orders := &a094CountingOrders{}
			alerts := &a094Alerts{}
			j, _ := a094Recover(t, path, orders, alerts)
			stored, _ := j.LookupAttempt(context.Background(), "attempt-acked")
			if stored.State != journal.StateAcked || orders.reads != 0 {
				t.Fatalf("state %s reads %d, want ACKED and no read", stored.State, orders.reads)
			}
			a094AssertNamed(t, alerts, "attempt-acked")
		})
	}
}

// 4.N4f① — 같은 attempt 의 알림은 재시작 뒤에도 같은 key(원장이 같은 행을 돌려줌), 창 0.
func TestA094TheAckedAlertKeyIsStableAcrossRestarts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	a094CrashAfterAck(t, path, journal.KindPlace, "O-acked")
	first := &a094Alerts{}
	j, _ := a094Recover(t, path, &a094CountingOrders{fail: errors.New("down")}, first)
	_ = j.Close()
	second := &a094Alerts{}
	a094Recover(t, path, &a094CountingOrders{fail: errors.New("down")}, second)
	if len(first.events) != 1 || len(second.events) != 1 || first.events[0].Key != second.events[0].Key {
		t.Fatalf("keys %v / %v, want the same single key across the restart", first.events, second.events)
	}
}

func a094AssertNamed(t *testing.T, alerts *a094Alerts, attemptID string) {
	t.Helper()
	if len(alerts.events) != 1 {
		t.Fatalf("alerts = %d, want exactly one named critical", len(alerts.events))
	}
	e := alerts.events[0]
	if obs.SeverityOf(e.Type) != obs.SeverityCritical || e.Key != string(e.Type)+"|acked:"+attemptID {
		t.Errorf("alert %s key %q, want critical keyed by the attempt", e.Type, e.Key)
	}
	if !strings.Contains(e.Body, attemptID) || alerts.remind[0] != 0 {
		t.Errorf("the alert must name %s and be recorded with window 0: %+v remind %s", attemptID, e, alerts.remind[0])
	}
	if _, ok := e.Fields[obs.FieldAccount]; ok {
		t.Error("the alert carries an account field")
	}
}

// 4.N4d — 번호가 일치했는데 확정 쓰기가 실패하면: ACKED 그대로, 명명 critical, 복구 성공.
func TestA094AFailedConfirmWriteLeavesTheAckAndNamesIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	a094CrashAfterAck(t, path, journal.KindPlace, "O-acked")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER a094_refuse_confirm BEFORE UPDATE OF state ON mutation_attempts
		WHEN NEW.state = 'CONFIRMED' BEGIN SELECT RAISE(ABORT, 'a094 fixture: the ledger refuses the write'); END`); err != nil {
		t.Fatalf("installing the fixture trigger: %v", err)
	}
	_ = db.Close()
	orders := &a094CountingOrders{single: map[string]json.RawMessage{"O-acked": a094Detail("O-acked", "AAPL")}}
	alerts := &a094Alerts{}
	j, report := a094Recover(t, path, orders, alerts)
	stored, _ := j.LookupAttempt(context.Background(), "attempt-acked")
	if stored.State != journal.StateAcked || len(report.ConfirmedAcked) != 0 {
		t.Fatalf("state %s confirmed %v, want ACKED and nothing confirmed", stored.State, report.ConfirmedAcked)
	}
	a094AssertNamed(t, alerts, "attempt-acked")
	if !strings.Contains(alerts.events[0].Body, "ledger write failed") {
		t.Errorf("the alert does not say the write failed: %q", alerts.events[0].Body)
	}
}

// 4.N4c — 구조: 기동 확정과 발주 직후 확인이 같은 판정 함수(execgw.ConfirmPlacedOrder)를 부른다.
func TestA094BootAndPostPlaceShareOneConfirmation(t *testing.T) {
	calls := func(path, fn string) map[string]bool {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]bool{}
		found := false
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Name.Name != fn {
				continue
			}
			found = true
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok {
					switch x := c.Fun.(type) {
					case *ast.SelectorExpr:
						out[x.Sel.Name] = true
					case *ast.Ident:
						out[x.Name] = true
					}
				}
				return true
			})
		}
		if !found {
			t.Fatalf("control: %s not in %s", fn, path)
		}
		return out
	}
	if !calls("acked_boot.go", "confirmAcked")["ConfirmPlacedOrder"] {
		t.Error("the boot confirmation does not call execgw.ConfirmPlacedOrder")
	}
	post := calls("../execgw/roundtrip.go", "confirmCreatedOrder")
	if !post["ConfirmPlacedOrder"] || post["parseOrderFacts"] || post["OrderRaw"] {
		t.Errorf("confirmCreatedOrder must delegate wholly to ConfirmPlacedOrder: %v", post)
	}
}

package verifylive

// a087_far_limit_only_pin_test.go — a087 tasks 4.2 「verifylive 무변화」 고정 시험.
//
// a087 은 엔진 보호 청산을 MARKET 으로 바꾸려고 실행 관문 ②(execgw.checkOrderShape)·③(trading.placeIntentSupported)
// 을 연다. 검증 도구의 안전 성질은 그 반대 — 「체결되면 안 되는 지정가」(pricing.go) — 이고, a087 design 「건드리지 않는 것」이
// 그 성질의 무변화를 약속함. 이 시험은 도구가 한 번의 전체 검증(등록 실행 + 재시작 뒤 재개 실행)에서 브로커로 보내는 일반 주문
// 생성·정정을 전부 가로채, 지금의 성질을 못 박음:
//
//   - 생성은 전부 LIMIT 이고 가격이 양수임(MARKET 이 섞이면 체결 가능한 주문이 생김).
//   - 매수는 최근 체결가보다 낮고 매도는 높음 — 즉 시장성 없는 가격임.
//   - 정정 가격도 같은 방향을 지킴(정정은 시장에서 한 틱 더 멀어지는 것이 설계).
//
// 조건주문(SINGLE+MARKET STOP)은 다른 endpoint 이고 발동 단계가 일부러 성질을 뒤집으므로(steps_trigger.go) 범위 밖임 —
// 그쪽은 a087 tasks 4.3.

import (
	"context"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// limitRecorder 는 fakeBroker 를 감싸 일반 주문 생성·정정 요청만 기록함. 나머지 메서드는 내장으로 그대로 위임됨.
type limitRecorder struct {
	*fakeBroker
	mu      sync.Mutex
	placed  []orderintent.PlaceIntent
	origin  map[string]orderintent.PlaceIntent // 생성 응답의 orderId → 생성 의도(정정 요청에는 방향·종목이 없어 생성 때 기억해 둠)
	amended []orderintent.AmendIntent
}

func (r *limitRecorder) PlaceOrder(ctx context.Context, intent orderintent.PlaceIntent) (trading.MutationResult, error) {
	res, err := r.fakeBroker.PlaceOrder(ctx, intent)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.placed = append(r.placed, intent)
	if err == nil && res.OrderID != "" {
		r.origin[res.OrderID] = intent
	}
	return res, err
}

func (r *limitRecorder) ModifyOrder(ctx context.Context, intent orderintent.AmendIntent) (trading.MutationResult, error) {
	res, err := r.fakeBroker.ModifyOrder(ctx, intent)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.amended = append(r.amended, intent)
	// 정정이 새 식별자를 내면(amendIssuesNewID) 그 식별자도 같은 방향으로 이어 기억함.
	if err == nil && res.OrderID != "" {
		if placed, ok := r.origin[intent.OrderID]; ok {
			r.origin[res.OrderID] = placed
		}
	}
	return res, err
}

// runRecorded 는 runToCompletion 의 비발동 경로와 같은 두 실행을 돌리되, 각 실행의 브로커를 기록기로 바꿔 끼움.
// h.run 과 같이 실행마다 기록 파일을 닫음 — 두 실행이 한 기록 파일을 동시에 열어 두지 않게 하려는 것임.
func runRecorded(t *testing.T, h *harness, rec *limitRecorder, opts Options) {
	t.Helper()
	runOnce := func(o Options) Summary {
		runner, closeRecord, err := h.build(o)
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		defer closeRecord()
		runner.broker = rec
		summary, err := runner.Run(context.Background())
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		return summary
	}
	if first := runOnce(opts); !first.Halted {
		t.Fatalf("the first run did not stop for the persistence check: %+v", first.Outcomes)
	}
	h.broker.restart()
	resumed := opts
	resumed.Prior = nil
	runOnce(resumed)
}

// TestA087VerifyPlacesOnlyFarLimitOrders 는 검증 도구가 내는 일반 주문이 전부 「체결되면 안 되는 지정가」임을 고정함.
func TestA087VerifyPlacesOnlyFarLimitOrders(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 3)
	h := newHarness(t, broker, alwaysConfirm())
	rec := &limitRecorder{fakeBroker: broker, origin: map[string]orderintent.PlaceIntent{}}

	runRecorded(t, h, rec, Options{HoldingSymbol: "005930"})

	// 표본이 비면 아래 전칭 단언이 공허하게 참이 됨 — 생성·정정이 실제로 있었는지 먼저 확인함.
	if len(rec.placed) == 0 || len(rec.amended) == 0 {
		t.Fatalf("placed %d, amended %d — the run must exercise both or this pin proves nothing",
			len(rec.placed), len(rec.amended))
	}
	t.Logf("표본: 생성 %d · 정정 %d", len(rec.placed), len(rec.amended))
	last := func(symbol string) float64 {
		broker.mu.Lock()
		defer broker.mu.Unlock()
		return broker.quotes[symbol]
	}
	for i, intent := range rec.placed {
		if intent.OrderType != "limit" {
			t.Errorf("placement %d %+v: order type %q, want limit — a verify order must not be able to fill", i, intent, intent.OrderType)
		}
		if intent.Price <= 0 {
			t.Errorf("placement %d %+v: price %v, want a positive limit", i, intent, intent.Price)
		}
		switch intent.Side {
		case "buy":
			if intent.Price >= last(intent.Symbol) {
				t.Errorf("placement %d: buy at %v is not below the last trade %v", i, intent.Price, last(intent.Symbol))
			}
		case "sell":
			if intent.Price <= last(intent.Symbol) {
				t.Errorf("placement %d: sell at %v is not above the last trade %v", i, intent.Price, last(intent.Symbol))
			}
		default:
			t.Errorf("placement %d: side %q", i, intent.Side)
		}
	}
	for i, intent := range rec.amended {
		if intent.Price == nil || *intent.Price <= 0 {
			t.Errorf("amendment %d %+v: no positive price", i, intent)
			continue
		}
		placed, ok := rec.origin[intent.OrderID]
		if !ok {
			t.Errorf("amendment %d targets %s, which this run did not place", i, intent.OrderID)
			continue
		}
		ref := last(placed.Symbol)
		if (placed.Side == "buy" && *intent.Price >= ref) || (placed.Side == "sell" && *intent.Price <= ref) {
			t.Errorf("amendment %d (%s %s) to %v crosses the last trade %v", i, placed.Side, placed.Symbol, *intent.Price, ref)
		}
	}
}

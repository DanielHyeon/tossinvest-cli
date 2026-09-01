//go:build tossos_testseams

package engine

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
)

// deliverForTest 는 시험이 dispatch 에 넘길 봉투를 **경계를 거쳐** 만든다.
//
// 봉투의 값 필드는 strategyhandoff 밖에서 채울 수 없으므로, 시험도 생산 코드와
// 같은 문(Admit → Deliver)을 지나야 한다. 시험만 쓰는 뒷문을 만들지 않는
// 것이 요점이다 — 뒷문을 만들면 그 뒷문이 곧 생산 코드가 쓸 수 있는 문이 된다.
func deliverForTest(t *testing.T, result strategyflow.Result) strategyhandoff.Delivered {
	t.Helper()
	var delivered strategyhandoff.Delivered
	err := strategyhandoff.Admit(true, []strategyflow.Result{result}).Deliver(
		func(value strategyhandoff.Delivered) error {
			delivered = value
			return nil
		})
	if err != nil {
		t.Fatalf("시험 봉투를 만들지 못했다: %v", err)
	}
	if delivered.Result() != result {
		t.Fatal("시험 봉투가 건네받은 값과 다른 것을 실었다")
	}
	return delivered
}

// 밖에서 만든 봉투는 Gateway 에 닿기 전에 걸린다.
//
// 봉투 타입이 지키는 성질의 **나머지 절반**이다. 타입은 "채워진 봉투는
// 경계에서만 나온다"를 보장하고, 이 시험은 "채워지지 않은 봉투는 아무것도
// 하지 못한다"를 값으로 확인한다. 둘 중 하나라도 가정으로 두면 fail-closed 는
// 우연이 된다 — 이 change 가 이미 두 번 틀린 자리다.
func TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall(t *testing.T) {
	cycle, _, _, spy := pairedStrategyDispatchCycleFixture(t)
	out, err := cycle.dispatch(context.Background(), strategyhandoff.Delivered{})
	if err == nil {
		t.Fatalf("위조한 봉투가 dispatch 를 통과했다: outcome=%+v", out)
	}
	spy.mu.Lock()
	places, observed := len(spy.calls), len(spy.observed)
	spy.mu.Unlock()
	if places != 0 {
		t.Fatalf("위조한 봉투가 Gateway 주문 %d 건을 냈다", places)
	}
	if observed != 0 {
		t.Fatalf("위조한 봉투가 Gateway 관측 %d 건을 냈다 — 첫 줄에서 걸려야 한다", observed)
	}
}

// 같은 봉투를 두 번 보내면 두 번째는 원장이 막는다.
//
// **이 시험은 가드가 못 하는 것을 대신 재는 것이다.** dispatch 호출 세기는
// 철자를 셀 뿐 실행 횟수를 세지 않아서, 적대 리뷰가 `for attempt := 0;
// attempt < 3; attempt++` 로 세 번 부르는 편집을 통과시켰다. 그래서 "한 주기에
// 한 번"은 소스 검사가 아니라 **원장의 lease** 가 지키는 성질이고, 그 사실을
// 여기서 값으로 확인한다. 확인하지 않으면 그 안전은 우연이다.
func TestTheSameEnvelopeCannotPlaceASecondOrder(t *testing.T) {
	cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
	delivered := deliverForTest(t, proposals.forMarket(StrategyMarketKR).entries[0].authority.Proposal())
	if _, err := cycle.dispatch(context.Background(), delivered); err != nil {
		t.Fatalf("첫 dispatch 가 실패했다: %v", err)
	}
	second, err := cycle.dispatch(context.Background(), delivered)
	if err == nil {
		t.Fatalf("같은 봉투가 두 번째 주문을 냈다: outcome=%+v", second)
	}
	spy.mu.Lock()
	places := len(spy.calls)
	spy.mu.Unlock()
	if places != 1 {
		t.Fatalf("Gateway 주문=%d 건, want 1 — 같은 봉투가 두 번 체결 경로에 닿았다", places)
	}
	// **막은 것이 무엇인지까지 단언한다.** 앞 판본은 이 자리에서 t.Logf 로
	// 이름을 찍기만 했다. 그래서 4차 적대 리뷰가 CAS 관문을 통째로 지워도 이
	// 시험은 그대로 통과했고 — 다른 것이 대신 막았다 — review.md 가 "원장 CAS
	// 가 막는 것을 쟀다"고 적은 문장은 재지 않은 귀속이었다.
	const wantBlocker = "production position campaign CAS changed"
	if !strings.Contains(err.Error(), wantBlocker) {
		t.Fatalf("두 번째를 막은 것=%v, want %q 를 담은 오류", err, wantBlocker)
	}
}

// **이 시험이 증명하지 않는 것.** 위 단언은 CAS 관문이 오늘 막는다는 것만
// 말한다. 4차 적대 리뷰가 그 관문을 지우고 재 보니 두 번째 주문은 여전히
// 막혔다 — 원장의 replay-identity 대조가 대신 막았고 Gateway 주문은 1건에서
// 멈췄다. 즉 at-most-once 는 과잉 결정되어 있다. 그 두 번째 방어선은 여기서
// 고정하지 않는다. 고정하려면 CAS 를 통과시키면서 replay identity 만 어긋나는
// 원장 상태를 만들어야 하고, 그것은 이 로트가 소유한 파일 밖이다.
// 침묵한 생략이 되지 않도록 여기 적어 둔다.

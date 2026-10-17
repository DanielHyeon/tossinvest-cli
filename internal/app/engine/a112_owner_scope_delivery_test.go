package engine

// a112 5.2.2.1 리뷰 수리(codex P1-1 · 보이스 B #2) — 주문 경로의 생산 몸통 `dispatchStrategyMarketHandoffs` 를 **그 자체로**
// 돌린다. 앞 판본의 반복 헬퍼 시험은 시험이 만든 몸통을 넘겼으므로, 생산 몸통 안에 「둘째 범위부터 버림」 카운터를 넣은
// 변이(보이스 B X09)를 아무 시험도 못 봤다. 여기서는 원장 읽기와 dispatch 두 문만 스파이로 바꾸고 몸통은 생산 것을 쓴다.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

type a112CampaignSpy struct {
	reads []string
	state map[string]journal.PositionCampaignCASRead
	fail  map[string]error
}

func (s *a112CampaignSpy) CurrentPositionCampaignCAS(_ context.Context, _, _, symbol string) (journal.PositionCampaignCASRead, error) {
	s.reads = append(s.reads, symbol)
	if err := s.fail[symbol]; err != nil {
		return journal.PositionCampaignCASRead{}, err
	}
	if read, ok := s.state[symbol]; ok {
		return read, nil
	}
	return journal.PositionCampaignCASRead{State: "FLAT"}, nil
}

type a112DispatchSpy struct {
	sent []string
	fail map[string]error
}

func (s *a112DispatchSpy) dispatch(_ context.Context, delivered strategyhandoff.Delivered) (execgw.Outcome, error) {
	symbol := delivered.Result().Lineage.Symbol
	s.sent = append(s.sent, symbol)
	return execgw.Outcome{}, s.fail[symbol]
}

func a112ThreeScopes() []strategyhandoff.Handoff {
	scope := func(symbol string) strategyflow.Result {
		return strategyflow.Result{Lineage: strategyflow.Lineage{Identity: symbol, AccountRef: "acct",
			Market: strategyrouter.MarketKR, Symbol: symbol, PositionGeneration: 1}}
	}
	return strategyhandoff.AdmitEachOwnerScope(true, []strategyflow.Result{scope("000660"), scope("005930"), scope("035420")})
}

func TestTheMarketDeliveryDispatchesEveryAdmittedScopeInOrder(t *testing.T) {
	campaigns, dispatcher := &a112CampaignSpy{}, &a112DispatchSpy{}
	if err := dispatchStrategyMarketHandoffs(context.Background(), campaigns, dispatcher, a112ThreeScopes()); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(dispatcher.sent, ","); got != "000660,005930,035420" {
		t.Fatalf("dispatched=%s, want every admitted owner scope in coordinator order", got)
	}
}

// 점유된(또는 FLAT/CLOSED 가 아닌) 범위는 건너뛰고 다음 범위로 간다 — 오류가 아니다. lease 소모도 오류가 아니다.
func TestASkippedScopeDoesNotStopTheNextOne(t *testing.T) {
	campaigns := &a112CampaignSpy{state: map[string]journal.PositionCampaignCASRead{
		"000660": {State: "FLAT", Claimed: true}, "005930": {State: "OPEN"}}}
	dispatcher := &a112DispatchSpy{fail: map[string]error{"035420": journal.ErrStrategyDispatchLeaseConsumed}}
	if err := dispatchStrategyMarketHandoffs(context.Background(), campaigns, dispatcher, a112ThreeScopes()); err != nil {
		t.Fatalf("err=%v, want nil (claimed · non-flat · lease-consumed are not faults)", err)
	}
	if got := strings.Join(campaigns.reads, ","); got != "000660,005930,035420" {
		t.Fatalf("CAS reads=%s, want all three scopes read", got)
	}
	if got := strings.Join(dispatcher.sent, ","); got != "035420" {
		t.Fatalf("dispatched=%s, want only the flat unclaimed scope", got)
	}
}

// 예상 밖 오류(원장 읽기 · dispatch 거절)는 같은 주기의 뒤 범위로 가지 않고 **그 값 그대로** 올라간다(보수 방향).
func TestAFaultInOneScopeStopsTheCycleBeforeTheNext(t *testing.T) {
	readFault, sendFault := errors.New("journal read failed"), errors.New("gateway refused")
	campaigns := &a112CampaignSpy{fail: map[string]error{"005930": readFault}}
	dispatcher := &a112DispatchSpy{}
	if err := dispatchStrategyMarketHandoffs(context.Background(), campaigns, dispatcher, a112ThreeScopes()); err != readFault {
		t.Fatalf("err=%v, want the read fault itself", err)
	}
	if got := strings.Join(dispatcher.sent, ","); got != "000660" {
		t.Fatalf("dispatched=%s, want only the scope before the fault", got)
	}
	campaigns, dispatcher = &a112CampaignSpy{}, &a112DispatchSpy{fail: map[string]error{"000660": sendFault}}
	if err := dispatchStrategyMarketHandoffs(context.Background(), campaigns, dispatcher, a112ThreeScopes()); err != sendFault {
		t.Fatalf("err=%v, want the dispatch fault itself", err)
	}
	if got := strings.Join(campaigns.reads, ","); got != "000660" {
		t.Fatalf("CAS reads=%s, want the cycle to stop at the first scope", got)
	}
}

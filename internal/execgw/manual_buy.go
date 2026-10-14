package execgw

// manual_buy.go 는 엔진 밖 수동 매수(토스 앱 등)를 하루 매수 총액에 넣는 발급자 쪽 절반임(a136 D9=b).
//
// # 나뉜 일
//
//   - 브로커 읽기: 엔진 보조 실행자(app/engine manual_buy_refresh.go)가 precheck 밖에서 공식 주문 조회를 읽어 시각이 찍힌
//     스냅숏(통화별 quote 금액)을 ManualBuyBoard 에 게시함. 첫 다리 precheck 의 「journal 변경·브로커 호출 없음」 계약과
//     PrecheckQFinalEntry 의 ctx 부재 때문에 읽기는 발급 경로 안에 있을 수 없음.
//   - 환산: 발급 문이 예약 요청을 만들 때(재수집 시도마다) 보드의 최신 스냅숏을 account base 로 바꿈. base 통화 금액은 그대로,
//     다른 통화 금액은 그 통화 시장의 봉인된 FX authority(요청 자기 것 또는 같은 파도의 짝 authority)로만 올림 환산 —
//     risk-management 「KR과 US는 하나의 account-base Guardian 권위를 공유한다」(사용자 승인 조항). 없으면 매수 거부.
//   - 판정: journal 예약 tx 가 스냅숏의 하루 · 신선도를 다시 보고 합계에 더함(daily_buy.go).

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// ErrManualBuyUnknown 은 수동 매수 합계를 정할 수 없어 매수를 거부했다는 뜻임(fail-closed).
var ErrManualBuyUnknown = errors.New("execgw: the day's buys outside the engine cannot be counted")

// ManualBuySnapshot 은 refresher 가 게시하는 한 번의 읽기 결과임.
type ManualBuySnapshot struct {
	// AsOf 는 브로커 읽기가 끝난 시각, Day 는 센 계좌 하루 라벨(clock.AccountTradingDay).
	AsOf time.Time
	Day  string
	// ByCurrency 는 통화별 quote 금액(decimal 문자열, 음수 불가) — 엔진 자신의 주문은 이미 빠짐.
	ByCurrency map[string]string
	// Err 가 있으면 이 읽기는 산정 불가였음(지정가 없는 미체결 · 해석 불가 행 · 조회 실패). 금액은 무시됨.
	Err error
}

// ManualBuyBoard 는 refresher 가 쓰고 발급 문이 읽는 스냅숏 한 칸(동시성 안전).
type ManualBuyBoard struct {
	mu       sync.Mutex
	snapshot ManualBuySnapshot
	set      bool
}

// NewManualBuyBoard 는 빈 보드를 만듦. 첫 게시 전에는 모든 매수가 거부됨(design D9 — 기동 직후 첫 refresh 전).
func NewManualBuyBoard() *ManualBuyBoard { return &ManualBuyBoard{} }

// Publish 는 새 스냅숏으로 바꿈. 지도는 복사해 둠 — 게시 뒤 호출자가 고쳐도 판정이 바뀌지 않게.
func (b *ManualBuyBoard) Publish(snapshot ManualBuySnapshot) {
	copied := make(map[string]string, len(snapshot.ByCurrency))
	for k, v := range snapshot.ByCurrency {
		copied[k] = v
	}
	snapshot.ByCurrency = copied
	b.mu.Lock()
	defer b.mu.Unlock()
	b.snapshot, b.set = snapshot, true
}

// Snapshot 은 최신 스냅숏과 게시 여부를 돌려줌.
func (b *ManualBuyBoard) Snapshot() (ManualBuySnapshot, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	copied := make(map[string]string, len(b.snapshot.ByCurrency))
	for k, v := range b.snapshot.ByCurrency {
		copied[k] = v
	}
	out := b.snapshot
	out.ByCurrency = copied
	return out, b.set
}

// ManualBuyBoard 는 이 Guardian 이 읽는 보드(생산 refresher 가 게시할 곳). nil 이면 모든 매수가 거부됨.
func (g *RiskGuardian) ManualBuyBoard() *ManualBuyBoard { return g.manualBuys }

// manualBuyFX 는 다른 통화 금액을 환산할 수 있는 봉인 authority 묶음 — 시장별 하나.
type manualBuyFX map[costs.Market]officialfx.Evidence

// manualBuyTotal 은 보드의 최신 스냅숏을 account base 금액으로 바꿔 예약 요청 칸을 만듦.
//
// 거부(ErrManualBuyUnknown): 보드 없음 · 게시 없음 · 산정 불가 스냅숏 · 모르는 통화 · base 가 아닌 통화 금액이 있는데 그 시장
// authority 가 없거나 봉인 검증(공식 원천 · 통화쌍 · 신선도 — risk.BindAccountBaseFX)에 실패. 금액 0 인 통화는 authority 를 요구하지 않음.
// 하루 · 신선도는 journal 이 같은 tx 에서 판정함(여기서 두 번 판정하지 않음).
func (g *RiskGuardian) manualBuyTotal(now time.Time, fx manualBuyFX) (*journal.ManualBuyTotal, error) {
	if g.manualBuys == nil {
		return nil, fmt.Errorf("%w: this Guardian has no manual buy board", ErrManualBuyUnknown)
	}
	snapshot, ok := g.manualBuys.Snapshot()
	if !ok {
		return nil, fmt.Errorf("%w: no manual buy snapshot has been published yet", ErrManualBuyUnknown)
	}
	if snapshot.Err != nil {
		return nil, fmt.Errorf("%w: %v", ErrManualBuyUnknown, snapshot.Err)
	}
	base := g.policy.LimitCurrency()
	total := "0"
	currencies := make([]string, 0, len(snapshot.ByCurrency))
	for currency := range snapshot.ByCurrency {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	for _, currency := range currencies {
		amount, err := riskcalc.CanonicalDecimal(snapshot.ByCurrency[currency])
		if err != nil {
			return nil, fmt.Errorf("%w: %s amount %q: %v", ErrManualBuyUnknown, currency, snapshot.ByCurrency[currency], err)
		}
		if negative, err := riskcalc.IsNegativeDecimal(amount); err != nil || negative {
			return nil, fmt.Errorf("%w: %s amount %s is negative", ErrManualBuyUnknown, currency, amount)
		}
		if zero, _ := riskcalc.CompareDecimal(amount, "0"); zero == 0 {
			continue
		}
		upper := strings.ToUpper(strings.TrimSpace(currency))
		if upper == base {
			if total, err = riskcalc.AddDecimal(total, amount); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrManualBuyUnknown, err)
			}
			continue
		}
		market, err := marketOfCurrency(upper)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrManualBuyUnknown, err)
		}
		authority, ok := fx[market]
		if !ok {
			return nil, fmt.Errorf("%w: %s %s of manual buys needs the %s authority of this request's paired FX collection, "+
				"and the request carries none", ErrManualBuyUnknown, amount, upper, market)
		}
		bound, err := risk.BindAccountBaseFX(now, market, g.policy, authority)
		if err != nil {
			return nil, fmt.Errorf("%w: the %s authority does not convert manual buys: %v", ErrManualBuyUnknown, market, err)
		}
		converted, err := risk.AccountBaseOrderNotional(now, market, amount, bound)
		if err != nil {
			return nil, fmt.Errorf("%w: converting %s %s: %v", ErrManualBuyUnknown, amount, upper, err)
		}
		if total, err = riskcalc.AddDecimal(total, converted.Amount); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrManualBuyUnknown, err)
		}
	}
	return &journal.ManualBuyTotal{Amount: total, AsOf: snapshot.AsOf, Day: snapshot.Day}, nil
}

// marketOfCurrency 는 주문 통화가 가리키는 시장임(/orders 응답에는 시장 칸이 없음 — official.marketFromCurrency 와 같은 표).
func marketOfCurrency(currency string) (costs.Market, error) {
	switch currency {
	case "KRW":
		return costs.MarketKR, nil
	case "USD":
		return costs.MarketUS, nil
	default:
		return "", fmt.Errorf("currency %q is not one this build prices", currency)
	}
}

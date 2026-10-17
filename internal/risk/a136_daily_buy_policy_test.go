package risk

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// a136: 하루 매수 총액 상한은 정책의 일부 — 기본값 10,000,000 KRW, 없음·0·다른 통화는 정책 검증이 거부(fail-closed).
func TestA136DefaultPolicyCarriesTheDailyBuyCeiling(t *testing.T) {
	p := DefaultPolicy()
	if p.MaxDailyBuyNotional != (riskcalc.Money{Amount: "10000000", Currency: "KRW"}) {
		t.Fatalf("default daily buy ceiling = %+v, want 10000000 KRW", p.MaxDailyBuyNotional)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("the default policy must validate: %v", err)
	}
}

func TestA136PolicyValidationRefusesAnUnusableDailyBuyCeiling(t *testing.T) {
	cases := map[string]riskcalc.Money{
		"absent":         {},
		"zero":           {Amount: "0", Currency: "KRW"},
		"negative":       {Amount: "-1", Currency: "KRW"},
		"not a decimal":  {Amount: "NaN", Currency: "KRW"},
		"other currency": {Amount: "10000000", Currency: "USD"},
	}
	for name, m := range cases {
		p := DefaultPolicy()
		p.MaxDailyBuyNotional = m
		err := p.Validate()
		if err == nil {
			t.Errorf("%s: a daily buy ceiling of %+v must be refused", name, m)
			continue
		}
		if !strings.Contains(err.Error(), "daily buy notional") {
			t.Errorf("%s: the refusal must name the daily buy ceiling, got %v", name, err)
		}
	}
}

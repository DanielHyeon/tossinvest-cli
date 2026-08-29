package strategyarbiter

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// 견줄 제안이 하나도 없으면 고를 것도 없다.
func TestNoProposalIsNotASelection(t *testing.T) {
	outcome := Arbitrate(Request{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930",
		PositionGeneration: 1, ObservedAt: time.Unix(1, 0).UTC()})
	if outcome.Refusal != RefusalSealMismatch || outcome.Detail != DetailNoProposal || outcome.Selected != -1 {
		t.Fatalf("outcome=%+v, want %q / %q with no index", outcome, RefusalSealMismatch, DetailNoProposal)
	}
}

// 신원이 빠진 요청은 중재를 시작조차 하지 않는다.
func TestARequestWithoutAnIdentityIsRefused(t *testing.T) {
	base := Request{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930",
		PositionGeneration: 1, ObservedAt: time.Unix(1, 0).UTC(), Proposals: []Proposal{{}}}
	cases := map[string]func(*Request){
		"account":    func(r *Request) { r.AccountRef = "" },
		"market":     func(r *Request) { r.Market = "" },
		"symbol":     func(r *Request) { r.Symbol = "" },
		"generation": func(r *Request) { r.PositionGeneration = 0 },
		"observedAt": func(r *Request) { r.ObservedAt = time.Time{} },
	}
	for name, mutate := range cases {
		request := base
		mutate(&request)
		if outcome := Arbitrate(request); outcome.Refusal != RefusalSealMismatch || outcome.Detail != DetailInvalidRequest {
			t.Fatalf("%s: refusal=%q detail=%q, want %q / %q", name, outcome.Refusal, outcome.Detail,
				RefusalSealMismatch, DetailInvalidRequest)
		}
	}
}

// 봉인되지 않은 빈 제안은 절대 선택될 수 없다.
func TestAZeroProposalIsNeverSelected(t *testing.T) {
	outcome := Arbitrate(Request{AccountRef: "acct", Market: strategyrouter.MarketKR, Symbol: "005930",
		PositionGeneration: 1, ObservedAt: time.Unix(1, 0).UTC(), Proposals: []Proposal{{}}})
	if outcome.Refusal == RefusalNone || outcome.Selected != -1 {
		t.Fatalf("outcome=%+v, want a refusal with no index", outcome)
	}
}

// 정확히 네 가족만 견줄 수 있다.
func TestOnlyTheFourApprovedFamiliesAreKnown(t *testing.T) {
	for _, family := range []strategyrouter.Family{strategyrouter.FamilyContinuation, strategyrouter.FamilyReversal,
		strategyrouter.FamilyWeeklyValue, strategyrouter.FamilyBreakoutRetest} {
		if !family.Known() {
			t.Fatalf("approved family %q is not known", family)
		}
	}
	for _, family := range []strategyrouter.Family{"", "MOMENTUM", "continuation", "CONTINUATION "} {
		if family.Known() {
			t.Fatalf("family %q must not be comparable", family)
		}
	}
}

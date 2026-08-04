//go:build tossos_testseams

package engine

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyaccount"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
)

func TestProductionFirstLegAuthorityLoaderPairedKRUS(t *testing.T) {
	riskFixture := newStrategyRiskLoaderFixture(t)
	now := riskFixture.results.observedAt
	riskPair := riskFixture.loader.collect(context.Background(), riskFixture.results, riskFixture.fx)
	proposals := strategyProposalAuthorityPair{observedAt: now}
	accounts := strategyAccountAuthorityPair{observedAt: now}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		result := riskFixture.results.forMarket(market).result
		batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:proposal-"+string(market), map[string]strategyflow.Result{result.Lineage.Symbol: result})
		authority, ok := batch.For(result.Lineage.Symbol)
		if !ok {
			t.Fatal("missing proposal test authority")
		}
		proposalMarket := strategyProposalMarketAuthority{market: market, entries: []strategyProposalEntryAuthority{{authority: authority}},
			snapshot: StrategyProposalMarketSnapshot{Market: market, Ready: true, Reason: StrategyProposalReady}}
		quote, cash := "KRW", "5000000"
		accountMarket := strategyaccount.MarketKR
		if market == StrategyMarketUS {
			quote, cash, accountMarket = "USD", "1000", strategyaccount.MarketUS
		}
		state := risk.AccountState{Mode: risk.ModeNormal, AllowedSymbols: []string{result.Lineage.Symbol}, HeldQuantity: "0",
			CashAvailable: riskcalc.Money{Amount: cash, Currency: quote}, OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
			DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: "KRW"}, AccountEquity: riskcalc.Money{Amount: "10000000", Currency: "KRW"}}
		accountAuthority := strategyaccount.AuthorityForTest(accountMarket, quote, state, now.Add(-time.Second), now.Add(time.Minute), 1,
			"sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
		accountValue := strategyAccountMarketAuthority{market: market, authority: accountAuthority,
			snapshot: StrategyAccountMarketSnapshot{Market: market, Ready: true, Reason: StrategyAccountReady}}
		if market == StrategyMarketKR {
			proposals.kr, accounts.kr = proposalMarket, accountValue
		} else {
			proposals.us, accounts.us = proposalMarket, accountValue
		}
	}
	j := openTestJournal(t)
	fakeClock := clock.NewFake(now)
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: j, Clock: fakeClock, AccountRef: "acct-risk-loader",
		Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "engine.automation_gate/risk-policy-v1"})
	if err != nil {
		t.Fatal(err)
	}
	loader := newProductionStrategyFirstLegAuthorityLoader(fakeClock, j, guardian, routeReadySchedulePair(now), proposals, riskPair, riskFixture.fx, accounts)
	markets := map[StrategyMarket]bool{}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		result := proposals.forMarket(market).entries[0].authority.Proposal()
		accepted, refusal := validateStrategyFirstLegResult(result)
		if refusal.Code != "" {
			t.Fatalf("%s result refused: %+v", market, refusal)
		}
		issuance, err := loader.collectStrategyFirstLegAuthority(context.Background(), accepted)
		if err != nil {
			t.Fatalf("%s collect: %v", market, err)
		}
		wantPrice := "100"
		if issuance.Entry.Market != string(market) || issuance.Entry.EntryPrice != wantPrice || issuance.Entry.QCandidate != result.Quantity ||
			len(issuance.Entry.Admission.Admission.Buckets) != 5 || len(issuance.Entry.Admission.Snapshots) != 5 ||
			issuance.Entry.ExpectedPolicyVersion != guardian.PolicyVersion() || issuance.Result.Lineage.Identity != result.Lineage.Identity ||
			issuance.Campaign.ExpectedPositionGeneration != 0 || issuance.Campaign.ExpectedPositionVersion != 0 || issuance.Entry.Collect == nil {
			t.Fatalf("%s issuance=%+v", market, issuance)
		}
		markets[market] = true
	}
	if !markets[StrategyMarketKR] || !markets[StrategyMarketUS] {
		t.Fatalf("unpaired first-leg production authorities: %+v", markets)
	}
}

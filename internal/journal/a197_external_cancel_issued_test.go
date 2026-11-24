package journal

// a197 design D7 · Manager 결정 M-1 — 「이미 냈다」 표시는 원장의 CANCEL attempt 행이고, 상태 · 거래일과 무관한 주문 생애당 한 번임.

import (
	"context"
	"testing"
)

func a197CancelRequest(id, day, market, symbol, target, source string) PrepareRequest {
	in := testIntent()
	in.ID = "intent-" + id
	in.TradingDay = day
	in.Market = market
	in.Symbol = symbol
	in.Side = "BUY"
	in.Source = source
	in.Fingerprint = "fp-" + id
	return PrepareRequest{Intent: in, Kind: KindCancel, AttemptID: "attempt-" + id, TargetOrderID: target}
}

func TestA197AnIssuedCancelIsFoundRegardlessOfStateAndDay(t *testing.T) {
	path := tempJournalPath(t)
	j := openTestJournalAt(t, path)
	ctx := context.Background()
	// RECORDED 에서 멈춘 취소(재시작 규칙으로 NOT_DISPATCHED)도 「냈다」.
	if _, err := j.Prepare(ctx, a197CancelRequest("c1", "2026-03-29", "kr", "005930", "EXT-1", "engine")); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if _, err := j.RecoverPending(ctx); err != nil {
		t.Fatalf("RecoverPending: %v", err)
	}
	cases := []struct {
		name                            string
		account, market, symbol, target string
		source                          string
		want                            bool
	}{
		{"same scope, another trading day", testIntent().AccountRef, "kr", "005930", "EXT-1", "engine", true},
		{"normalised market and symbol", testIntent().AccountRef, " KR ", "005930 ", "EXT-1", "engine", true},
		{"another number", testIntent().AccountRef, "kr", "005930", "EXT-2", "engine", false},
		{"bytes differ", testIntent().AccountRef, "kr", "005930", "ext-1", "engine", false},
		{"another symbol", testIntent().AccountRef, "kr", "000660", "EXT-1", "engine", false},
		{"another market", testIntent().AccountRef, "us", "005930", "EXT-1", "engine", false},
		{"another account", "acct-other", "kr", "005930", "EXT-1", "engine", false},
		{"flatten's cancel is not the engine's", testIntent().AccountRef, "kr", "005930", "EXT-1", "flatten-all", false},
	}
	for _, tc := range cases {
		got, err := j.ExternalCancelIssued(ctx, tc.account, tc.market, tc.symbol, tc.target, tc.source)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got != tc.want {
			t.Errorf("%s: issued = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestA197APlaceWithTheSameNumberIsNotAnIssuedCancel(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	req := testRequest()
	req.Intent.Market, req.Intent.Symbol, req.Intent.Source = "kr", "005930", "engine"
	a, err := j.Prepare(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.MarkAcked(ctx, "EXT-1"); err != nil {
		t.Fatal(err)
	}
	got, err := j.ExternalCancelIssued(ctx, req.Intent.AccountRef, "kr", "005930", "EXT-1", "engine")
	if err != nil || got {
		t.Fatalf("issued = %v err %v, want false — only CANCEL rows mark a number as issued", got, err)
	}
}

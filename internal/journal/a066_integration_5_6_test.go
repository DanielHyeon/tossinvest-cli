package journal

// a066 task 5.6.1 — 공유 bucket 정합(5.6 실측이 드러낸 선행 수리)의 계약 시험.
//
// 각색 기록: 이 파일은 2026-09-27 측정 전용(로그만 남기고 통과)으로 시작했음. 측정이 두 결함을 보였음 —
//   F2: risk_bucket_policies 는 (dimension, value, policy_version) 하나에 예약 가격 정책(가격·FX·fee·평가시각)
//       record 하나만 받으므로, 공유 bucket 에 **다른 가격**으로 들어오는 두 번째 진입이 늘 "immutable policy
//       collision" 으로 거절됨(가격 5→6: 거절).
//   F1: admission 은 호출자 snapshot 의 filled/held 를 원장과 대조하지 않으므로, 첫 진입 이전 snapshot(held 0)과
//       최신 예약 버전을 든 두 번째 진입이 공유 cap 을 넘겨 admit 됨(가격 5→5: q_final 10, 원장 사용량 100 > 한도 80).
//   F2 가 F1 을 우연히 가리고 있었음(다른 가격이면 collision 이 먼저 거절). 측정 로그를 계약 단언으로 바꾸어
//   수리의 RED 로 승격함(Manager 승인 2026-09-27).
//
// 생산 순서를 그대로 흉내 냄: 생산 첫 leg loader(internal/app/engine/strategy_account_first_leg_authority.go)는
// bucket snapshot 을 주기 앞에서 한 번 모으고, 예약 버전(ObservedVersion)만 발급 시점 collect 에서 새로 읽음.

import (
	"context"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// sharedBucketLimit 는 두 진입이 공유하는 bucket 한도. 한 진입의 예약은 가격×q = 5×10 = 50 이므로 하나는 들어가고
// 둘은 공유 bucket(horizon·strategy·sector, 같은 시장이면 market 까지)을 넘음.
const sharedBucketLimit = "80"

// sharedBucketPair 는 같은 계좌의 두 진입을 만듦. 두 번째는 종목(과 필요하면 시장)이 다르고 가격이 secondPrice 임.
func sharedBucketPair(t *testing.T, j *Journal, secondMarket riskbucket.Market, secondPrice string) (QFinalIssueRequest, QFinalIssueRequest) {
	t.Helper()
	first := qFinalIssueFixture(t, j, "shared-a")
	second := qFinalIssueFixture(t, j, "shared-b")
	for _, request := range []*QFinalIssueRequest{&first, &second} {
		for i := range request.Admission.Admission.Buckets {
			rebindRiskBucketLimit(t, &request.Admission, i, request.Admission.Admission.Buckets[i].Key, sharedBucketLimit)
		}
	}
	symbol, intentMarket := "000660", "kr"
	if secondMarket == riskbucket.MarketUS {
		symbol, intentMarket = "AAPL", "us"
		second.Admission.Owner.Key.Market = riskbucket.MarketUS
		second.Admission.Admission.Policy.QuoteCurrency = "USD"
		rebindRiskBucket(t, &second.Admission, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: string(riskbucket.MarketUS), PolicyVersion: "policy-v1"})
	}
	second.Admission.Owner.Key.Symbol = symbol
	rebindRiskBucket(t, &second.Admission, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: symbol, PolicyVersion: "policy-v1"})
	second.Admission.Admission.Policy.Price.WorstExecutableQuote = secondPrice
	intent := second.Issue.Decision.Preimage.(RiskIntent)
	intent.Market, intent.Symbol = intentMarket, symbol
	second.Issue.Decision.Preimage = intent
	return first, second
}

// withSnapshotUsage 는 공유되는 bucket(market 은 같은 시장일 때만)의 snapshot held 를 바꿔 다시 봉인함.
func withSnapshotUsage(t *testing.T, plan *RiskBucketAdmissionPlan, sameMarket bool, held string) {
	t.Helper()
	for i := range plan.Admission.Buckets {
		dimension := plan.Admission.Buckets[i].Key.Dimension
		if dimension == riskbucket.DimensionSymbol || (dimension == riskbucket.DimensionMarket && !sameMarket) {
			continue
		}
		plan.Admission.Buckets[i].HeldMinor = held
		rebindRiskBucket(t, plan, i, plan.Admission.Buckets[i].Key)
	}
}

// issueSecondWithFreshVersion 는 첫 진입 뒤 예약 버전만 새로 읽어 두 번째 진입을 발급함(생산 collect 순서).
func issueSecondWithFreshVersion(t *testing.T, j *Journal, second QFinalIssueRequest) (QFinalIssueResult, error) {
	t.Helper()
	version, err := j.ReservationVersion(context.Background(), "acct-1")
	if err != nil {
		t.Fatal(err)
	}
	second.Issue.Reserve.ObservedVersion = version
	return j.RecordQFinalDecisionAndReserve(context.Background(), second)
}

// TestA066SharedBucketSecondAdmissionAtAnotherPriceIsCappedNotCollided 는 F2 의 계약: 공유 bucket 에 다른 가격으로
// 들어오는 두 번째 진입은 원장과 맞는 snapshot 을 들고 오면 **admit 되고 남은 한도로 cap** 됨(80-50=30, 가격 6 → q 5).
// KR 두 종목과 KR→US(다른 통화 정책) 두 모양.
func TestA066SharedBucketSecondAdmissionAtAnotherPriceIsCappedNotCollided(t *testing.T) {
	for _, market := range []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			j := openTestJournal(t)
			first, second := sharedBucketPair(t, j, market, "6")
			if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
				t.Fatalf("first admission: %v", err)
			}
			withSnapshotUsage(t, &second.Admission, market == riskbucket.MarketKR, "50")
			// 결정 수량은 q_final 과 같아야 함(발급 계약) — 남은 30 을 가격 6 으로 나눈 5.
			intent := second.Issue.Decision.Preimage.(RiskIntent)
			intent.Quantity = "5"
			second.Issue.Decision.Preimage = intent
			result, err := issueSecondWithFreshVersion(t, j, second)
			if err != nil || result.Admission.QFinal != 5 {
				t.Fatalf("second admission at another price with a ledger-true snapshot: q_final=%d err=%v, want 5", result.Admission.QFinal, err)
			}
			if usage := sharedBucketUsage(t, j, riskbucket.DimensionStrategy, "strategy-alpha"); usage != "80" {
				t.Fatalf("strategy bucket journal usage %s, want 80 (50 + 30)", usage)
			}
		})
	}
}

// TestA066SharedBucketStaleSnapshotIsRefusedAtAnyPrice 는 F1 의 계약: 첫 진입 이전 snapshot(held 0)을 든 두 번째
// 진입은 가격이 같든 다르든 **stale** 로 거절되고(재수집 대상), 아무것도 쓰지 않으며 공유 cap 은 넘지 않음.
func TestA066SharedBucketStaleSnapshotIsRefusedAtAnyPrice(t *testing.T) {
	for _, tc := range []struct {
		name   string
		market riskbucket.Market
		price  string
	}{
		{"KR same price", riskbucket.MarketKR, "5"},
		{"KR other price", riskbucket.MarketKR, "6"},
		{"US other price", riskbucket.MarketUS, "6"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			first, second := sharedBucketPair(t, j, tc.market, tc.price)
			if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
				t.Fatalf("first admission: %v", err)
			}
			decisionsBefore := countRiskBucketRows(t, j, "decisions")
			result, err := issueSecondWithFreshVersion(t, j, second)
			usage := sharedBucketUsage(t, j, riskbucket.DimensionStrategy, "strategy-alpha")
			if err == nil {
				t.Fatalf("stale snapshot admitted: q_final=%d, strategy usage %s against limit %s", result.Admission.QFinal, usage, sharedBucketLimit)
			}
			if !errors.Is(err, ErrRiskBucketUsageStale) || !errors.Is(err, ErrSnapshotStale) || !riskbucket.IsRefusal(err, riskbucket.RefusalBucketUsageStale) {
				t.Fatalf("stale snapshot refused for another reason: %v", err)
			}
			if usage != "50" || countRiskBucketRows(t, j, "decisions") != decisionsBefore {
				t.Fatalf("refused admission wrote: strategy usage %s, decisions %d→%d", usage, decisionsBefore, countRiskBucketRows(t, j, "decisions"))
			}
		})
	}
}

// sharedBucketUsage 는 원장 사용량(해당 계좌·dimension·value 의 예약 held+filled 합)을 셈.
func sharedBucketUsage(t *testing.T, j *Journal, dimension riskbucket.Dimension, value string) string {
	t.Helper()
	var usage string
	if err := j.db.QueryRow(`SELECT CAST(COALESCE(SUM(CAST(held_minor AS INTEGER)+CAST(filled_minor AS INTEGER)),0) AS TEXT)
		FROM risk_bucket_reservations WHERE account_ref='acct-1' AND bucket_dimension=? AND bucket_value=?`, string(dimension), value).Scan(&usage); err != nil {
		t.Fatal(err)
	}
	return usage
}

// rebindRiskBucketLimit 는 bucket 한도만 바꾸고 provenance 를 다시 봉인함.
func rebindRiskBucketLimit(t *testing.T, plan *RiskBucketAdmissionPlan, index int, key riskbucket.BucketKey, limit string) {
	t.Helper()
	plan.Admission.Buckets[index].LimitMinor = limit
	rebindRiskBucket(t, plan, index, key)
}

// TestA066SharedBucketSecondEntryReadsItsOwnPolicyRecord 는 F2 수리의 소비 쪽: 공유 bucket 의 두 번째 진입(다른 가격·통화)이
// 주문 권위를 읽을 때 첫 진입의 record 가 아니라 **자기** record 를 읽음(주문 등록·체결 계상이 그 통화·digest 로 묶임).
func TestA066SharedBucketSecondEntryReadsItsOwnPolicyRecord(t *testing.T) {
	j := openTestJournal(t)
	first, second := sharedBucketPair(t, j, riskbucket.MarketUS, "6")
	if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	withSnapshotUsage(t, &second.Admission, false, "50")
	intent := second.Issue.Decision.Preimage.(RiskIntent)
	intent.Quantity = "5"
	second.Issue.Decision.Preimage = intent
	if _, err := issueSecondWithFreshVersion(t, j, second); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		decision, quote string
	}{{first.Issue.Decision.ID, "KRW"}, {second.Issue.Decision.ID, "USD"}} {
		tx, err := j.db.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		authority, err := loadRiskBucketOrderAuthority(context.Background(), tx, tc.decision)
		_ = tx.Rollback()
		if err != nil || authority.quoteCurrency != tc.quote || len(authority.bindings) != 5 {
			t.Fatalf("decision %s order authority: quote=%q bindings=%d err=%v, want %s", tc.decision, authority.quoteCurrency, len(authority.bindings), err, tc.quote)
		}
	}
	var records int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_policy_records WHERE bucket_dimension='strategy' AND bucket_value='strategy-alpha'`).Scan(&records); err != nil || records != 2 {
		t.Fatalf("strategy bucket policy records=%d err=%v, want 2 (one per entry)", records, err)
	}
}

// TestA066StaleUsageRecollectionTerminates 는 stale 거절의 재수집 경로가 끝남을 고정함(Manager 조건 2026-09-27).
// 재수집 루프는 recollectLoop(reservations.go)이고 상한은 둘임: policy.MaxAttempts 회("for attempt := 1; attempt <=
// policy.MaxAttempts") 와 policy.Budget 시간("now.After(deadline)"). ErrRiskBucketUsageStale 은 ErrSnapshotStale 을 감싸므로
// 그 루프의 재시도 갈래("errors.Is(err, ErrSnapshotStale)")를 탐.
func TestA066StaleUsageRecollectionTerminates(t *testing.T) {
	t.Run("stale then fresh snapshot is admitted within the cap", func(t *testing.T) {
		j := openTestJournal(t)
		first, second := sharedBucketPair(t, j, riskbucket.MarketKR, "5")
		if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
			t.Fatal(err)
		}
		calls := 0
		result, err := j.RecordQFinalDecisionAndReserveWithRecollection(context.Background(), func(ctx context.Context, attempt int) (QFinalIssueRequest, error) {
			calls++
			request := second
			request.Admission.Admission.Buckets = append([]riskbucket.BucketSnapshot(nil), second.Admission.Admission.Buckets...)
			request.Admission.Snapshots = append([]RiskBucketSnapshotReference(nil), second.Admission.Snapshots...)
			if attempt > 1 {
				// 재수집: 원장과 맞는 snapshot(공유 bucket held 50) → 남은 30 / 가격 5 = q 6.
				withSnapshotUsage(t, &request.Admission, true, "50")
				intent := request.Issue.Decision.Preimage.(RiskIntent)
				intent.Quantity = "6"
				request.Issue.Decision.Preimage = intent
			}
			version, err := j.ReservationVersion(ctx, "acct-1")
			request.Issue.Reserve.ObservedVersion = version
			return request, err
		}, RecollectPolicy{MaxAttempts: 3})
		if err != nil || calls != 2 || result.Admission.QFinal != 6 {
			t.Fatalf("recollection: calls=%d q_final=%d err=%v, want 2 calls and q_final 6", calls, result.Admission.QFinal, err)
		}
		if usage := sharedBucketUsage(t, j, riskbucket.DimensionStrategy, "strategy-alpha"); usage != "80" {
			t.Fatalf("strategy usage %s, want 80", usage)
		}
	})
	t.Run("a ledger that keeps moving ends at the attempt bound", func(t *testing.T) {
		j := openTestJournal(t)
		first, second := sharedBucketPair(t, j, riskbucket.MarketKR, "5")
		if _, err := j.RecordQFinalDecisionAndReserve(context.Background(), first); err != nil {
			t.Fatal(err)
		}
		calls := 0
		_, err := j.RecordQFinalDecisionAndReserveWithRecollection(context.Background(), func(ctx context.Context, attempt int) (QFinalIssueRequest, error) {
			calls++
			version, err := j.ReservationVersion(ctx, "acct-1")
			request := second
			request.Issue.Reserve.ObservedVersion = version
			return request, err // snapshot 은 늘 첫 진입 이전(held 0) — 원장이 계속 앞서 있는 모양
		}, RecollectPolicy{MaxAttempts: 4})
		if !errors.Is(err, ErrRecollectionExhausted) || !errors.Is(err, ErrRiskBucketUsageStale) || calls != 4 {
			t.Fatalf("moving ledger: calls=%d err=%v, want exactly 4 attempts then ErrRecollectionExhausted wrapping the stale refusal", calls, err)
		}
		if usage := sharedBucketUsage(t, j, riskbucket.DimensionStrategy, "strategy-alpha"); usage != "50" {
			t.Fatalf("exhausted recollection wrote: strategy usage %s", usage)
		}
	})
}

// refreshSnapshotUsageFromLedger 는 fixture 의 bucket snapshot 사용량을 **원장 그대로**(생산 reader 와 같은 함수) 채우고
// provenance 를 다시 봉인함. 5.6.1 F1 뒤로 admission 은 원장보다 적게 주장하는 snapshot 을 stale 로 거절하므로, 한 저널에
// 여러 진입을 쌓는 fixture 는 각 진입의 snapshot 을 생산처럼 원장에서 모아야 함.
func refreshSnapshotUsageFromLedger(t *testing.T, j *Journal, plan *RiskBucketAdmissionPlan) {
	t.Helper()
	for i := range plan.Admission.Buckets {
		key := plan.Admission.Buckets[i].Key
		usage, err := riskbucket.ReadJournalBucketUsage(context.Background(), j.db, plan.Owner.Key.AccountID, key.Dimension, key.Value)
		if err != nil {
			t.Fatalf("ledger usage for %s/%s: %v", key.Dimension, key.Value, err)
		}
		plan.Admission.Buckets[i].FilledMinor = usage.FilledMinor
		plan.Admission.Buckets[i].HeldMinor = usage.HeldMinor
		rebindRiskBucket(t, plan, i, key)
	}
}

// TestA066StaleUsageRuleEdgesAtBothAdmissionSites 는 F1 규칙의 경계를 두 admission 자리에서 각각 잼(변이 1회차가 보인
// 공백: CommitRiskBucketAdmission 자리는 행동 시험이 없었고, 부분적 과소 주장·filled 사용량·원장 읽기 실패는 안 재졌음).
func TestA066StaleUsageRuleEdgesAtBothAdmissionSites(t *testing.T) {
	type site struct {
		name  string
		admit func(t *testing.T, j *Journal, heldClaim string) error
	}
	sites := []site{
		{"CommitRiskBucketAdmission", func(t *testing.T, j *Journal, heldClaim string) error {
			seedExistingRiskReservation(t, j, "existing-edge-b", "acct-1")
			plan := riskBucketAdmissionFixture(t, "edge-b", "acct-1", "lane-b", "campaign-edge-b", "prospective-edge-b", sharedBucketLimit, heldClaim)
			plan.Owner.Key.Symbol = "000660"
			rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "000660", PolicyVersion: "policy-v1"})
			withSnapshotUsage(t, &plan, true, heldClaim)
			_, err := j.CommitRiskBucketAdmission(context.Background(), plan)
			return err
		}},
		{"RecordQFinalDecisionAndReserve", func(t *testing.T, j *Journal, heldClaim string) error {
			_, second := sharedBucketPair(t, j, riskbucket.MarketKR, "5")
			withSnapshotUsage(t, &second.Admission, true, heldClaim)
			// 발급 계약: 결정 수량 = snapshot 이 주장하는 여유로 계산한 q_final((80 - held) / 5, 최대 10).
			quantity := map[string]string{"0": "10", "40": "8", "50": "6"}[heldClaim]
			intent := second.Issue.Decision.Preimage.(RiskIntent)
			intent.Quantity = quantity
			second.Issue.Decision.Preimage = intent
			_, err := issueSecondWithFreshVersion(t, j, second)
			return err
		}},
	}
	firstAdmission := func(t *testing.T, j *Journal) {
		seedExistingRiskReservation(t, j, "existing-edge-a", "acct-1")
		plan := riskBucketAdmissionFixture(t, "edge-a", "acct-1", "lane-a", "campaign-edge-a", "prospective-edge-a", sharedBucketLimit, "0")
		if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); err != nil {
			t.Fatalf("first admission: %v", err)
		}
	}
	for _, s := range sites {
		t.Run(s.name+"/understated held 40 of 50 is stale", func(t *testing.T) {
			j := openTestJournal(t)
			firstAdmission(t, j)
			if err := s.admit(t, j, "40"); !errors.Is(err, ErrRiskBucketUsageStale) {
				t.Fatalf("held 40 against ledger 50 was not refused as stale: %v", err)
			}
		})
		t.Run(s.name+"/filled usage counts", func(t *testing.T) {
			j := openTestJournal(t)
			firstAdmission(t, j)
			// 첫 진입이 전부 체결된 모양: held 0, filled 50.
			if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET held_minor='0',filled_minor='50',state='FILLED'`); err != nil {
				t.Fatal(err)
			}
			if err := s.admit(t, j, "0"); !errors.Is(err, ErrRiskBucketUsageStale) {
				t.Fatalf("filled 50 was not counted: %v", err)
			}
		})
		t.Run(s.name+"/unreadable ledger refuses", func(t *testing.T) {
			j := openTestJournal(t)
			firstAdmission(t, j)
			if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET held_minor='not-an-amount' WHERE bucket_dimension='strategy'`); err != nil {
				t.Fatal(err)
			}
			if err := s.admit(t, j, "50"); err == nil || errors.Is(err, ErrRiskBucketUsageStale) {
				t.Fatalf("unreadable ledger usage must refuse (not as a retryable stale): %v", err)
			}
		})
		t.Run(s.name+"/ledger-true claim is admitted", func(t *testing.T) {
			j := openTestJournal(t)
			firstAdmission(t, j)
			if err := s.admit(t, j, "50"); err != nil && !riskbucket.IsRefusal(err, riskbucket.RefusalBucketCapExhausted) {
				t.Fatalf("ledger-true claim refused for a reason other than the cap: %v", err)
			}
		})
	}
}

package journal

// a066 task 5.6 — KR/US 동시 통합 시험. 계약 문장(tasks.md 5.6):
//
//	"independent market buckets, shared strategy caps, one symbol/one owner, monetary scale-in aggregation and
//	 failure isolation without requiring an operating toggle"
//
// 한 계좌에서 KR·US 진입을 한 흐름으로 쌓으며 문장의 다섯 성질을 원장 값으로 잼. 운영 토글·활성화는 쓰지 않음 —
// journal 권위(admission 트랜잭션)만으로 성립해야 하는 계약임. snapshot 은 생산처럼 wave 시작에 원장에서 모으고
// (riskbucket.ReadJournalBucketUsage), 예약 버전만 발급 시점에 새로 읽음.

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// contractLimits 는 dimension 순서(horizon, market, strategy, sector, symbol)의 한도. 한 진입의 예약은 가격 5 × q.
type contractLimits [5]string

var (
	// market 100 은 시장마다 따로, strategy 120 은 두 시장이 함께 씀(KR 50 + US 50 + 20 남음).
	flowLimits = contractLimits{"1000", "100", "120", "1000", "100"}
	// 격리 시험용: 공유 bucket 에 넉넉한 여유.
	isolationLimits = contractLimits{"1000", "100", "1000", "1000", "100"}
)

type contractEntry struct {
	suffix, market, symbol, lane, campaign, prospective string
	quantity                                            string // 기대 q_final — 발급 계약상 결정 수량과 같아야 함
	// sameHorizonSector 면 US 도 KR 과 같은 horizon(SHORT)·sector 를 씀 — 격리 시험에서 시장 범위만이 두 진입을 가르게 함.
	sameHorizonSector bool
}

// contractRequest 는 한 진입의 q_final 발급 요청을 만듦. snapshot 사용량은 이 호출 시점의 원장 그대로임.
func contractRequest(t *testing.T, j *Journal, e contractEntry, limits contractLimits) QFinalIssueRequest {
	t.Helper()
	request := qFinalIssueFixture(t, j, e.suffix)
	plan := &request.Admission
	plan.Owner.Key.Symbol = e.symbol
	plan.Owner.LaneID, plan.Owner.CampaignID, plan.Owner.Key.ProspectiveGeneration = e.lane, e.campaign, e.prospective
	intent := request.Issue.Decision.Preimage.(RiskIntent)
	intent.Symbol, intent.Quantity = e.symbol, e.quantity
	if e.market == "US" {
		plan.Owner.Key.Market = riskbucket.MarketUS
		plan.Admission.Policy.QuoteCurrency = "USD"
		intent.Market = "us"
	}
	request.Issue.Decision.Preimage = intent
	values := map[riskbucket.Dimension]string{riskbucket.DimensionMarket: e.market, riskbucket.DimensionSymbol: e.symbol}
	if e.market == "US" && !e.sameHorizonSector {
		// US 는 horizon·sector 도 KR 과 다르게 둠 — 두 시장을 잇는 공유 bucket 이 **strategy 하나뿐**이게 해서 "shared strategy
		// caps" 가 horizon·sector 공유에 기대지 않고 성립함을 잼.
		values[riskbucket.DimensionHorizon] = string(riskbucket.HorizonMedium)
		values[riskbucket.DimensionSector] = "sector-us"
	}
	for i := range plan.Admission.Buckets {
		key := plan.Admission.Buckets[i].Key
		if value, ok := values[key.Dimension]; ok {
			key.Value = value
		}
		plan.Admission.Buckets[i].LimitMinor = limits[i]
		rebindRiskBucket(t, plan, i, key)
	}
	refreshSnapshotUsageFromLedger(t, j, plan)
	return request
}

// issueWave 는 한 wave 의 진입들을 동시에 발급함 — 생산 collect 처럼 snapshot 은 고정, 예약 버전만 시도마다 새로 읽음.
func issueWave(t *testing.T, j *Journal, requests ...QFinalIssueRequest) []error {
	t.Helper()
	errs := make([]error, len(requests))
	var wait sync.WaitGroup
	for i := range requests {
		wait.Add(1)
		go func(i int) {
			defer wait.Done()
			_, errs[i] = j.RecordQFinalDecisionAndReserveWithRecollection(context.Background(), func(ctx context.Context, _ int) (QFinalIssueRequest, error) {
				request := requests[i]
				version, err := j.ReservationVersion(ctx, "acct-1")
				request.Issue.Reserve.ObservedVersion = version
				return request, err
			}, RecollectPolicy{MaxAttempts: 3})
		}(i)
	}
	wait.Wait()
	return errs
}

func contractUsage(t *testing.T, j *Journal, dimension riskbucket.Dimension, value string) string {
	t.Helper()
	usage, err := riskbucket.ReadJournalBucketUsage(context.Background(), j.db, "acct-1", dimension, value)
	if err != nil {
		t.Fatal(err)
	}
	total, ok := sumMinor(usage.FilledMinor, usage.HeldMinor)
	if !ok {
		t.Fatalf("usage %+v", usage)
	}
	return total.String()
}

// TestA066KRUSConcurrentContract 는 5.6 문장의 앞 네 성질을 한 흐름으로 잼.
func TestA066KRUSConcurrentContract(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	kr := contractEntry{"flow-kr", "KR", "005930", "lane-kr", "campaign-kr", "prospective-kr", "10", false}
	us := contractEntry{"flow-us", "US", "AAPL", "lane-us", "campaign-us", "prospective-us", "10", false}

	// 다른 계좌의 사용량은 이 계좌의 어떤 bucket 에도 들어가지 않음(같은 bucket 값 strategy-alpha·sector-tech·SHORT).
	seedExistingRiskReservation(t, j, "existing-other-account", "acct-2")
	other := riskBucketAdmissionFixture(t, "other-account", "acct-2", "lane-other", "campaign-other", "prospective-other", "1000", "0")
	other.ExistingReservationID = "existing-other-account"
	if _, err := j.CommitRiskBucketAdmission(ctx, other); err != nil {
		t.Fatalf("other account admission: %v", err)
	}

	// wave 1: KR·US 가 같은 snapshot(사용량 0)으로 동시에 들어옴 — 공유 strategy bucket 때문에 한 시장만 성립하고
	// 다른 하나는 BUCKET_USAGE_STALE(재시도 없음).
	errs := issueWave(t, j, contractRequest(t, j, kr, flowLimits), contractRequest(t, j, us, flowLimits))
	var loser contractEntry
	switch {
	case errs[0] == nil && errors.Is(errs[1], ErrRiskBucketUsageStale):
		loser = us
	case errs[1] == nil && errors.Is(errs[0], ErrRiskBucketUsageStale):
		loser = kr
	default:
		t.Fatalf("wave 1 must admit exactly one market and refuse the other as stale: KR=%v US=%v", errs[0], errs[1])
	}
	// wave 2: 패자는 원장에서 새로 모은 snapshot 으로 성립함.
	if errs := issueWave(t, j, contractRequest(t, j, loser, flowLimits)); errs[0] != nil {
		t.Fatalf("wave 2 %s: %v", loser.market, errs[0])
	}

	// (1) independent market buckets: 시장마다 자기 진입만 담음(다른 계좌의 50 도 섞이지 않음).
	if kr, us := contractUsage(t, j, riskbucket.DimensionMarket, "KR"), contractUsage(t, j, riskbucket.DimensionMarket, "US"); kr != "50" || us != "50" {
		t.Fatalf("market buckets KR=%s US=%s, want 50 each (independent)", kr, us)
	}
	// (2) shared strategy cap: strategy bucket 은 두 시장의 합.
	if got := contractUsage(t, j, riskbucket.DimensionStrategy, "strategy-alpha"); got != "100" {
		t.Fatalf("shared strategy usage %s, want 100 (KR 50 + US 50)", got)
	}

	// (3) one symbol / one owner: 같은 시장·종목에 다른 lane 이 들어오면 owner 충돌, 아무것도 쓰지 않음.
	decisionsBefore := countRiskBucketRows(t, j, "risk_bucket_final_decisions")
	for _, rival := range []contractEntry{
		{"rival-kr", "KR", "005930", "lane-rival", "campaign-rival-kr", "prospective-rival-kr", "4", false},
		{"rival-us", "US", "AAPL", "lane-rival", "campaign-rival-us", "prospective-rival-us", "4", false},
	} {
		if _, err := j.RecordQFinalDecisionAndReserve(ctx, contractRequest(t, j, rival, flowLimits)); !errors.Is(err, ErrRiskBucketOwnerConflict) {
			t.Fatalf("%s rival lane for %s: %v, want owner conflict", rival.market, rival.symbol, err)
		}
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_final_decisions"); got != decisionsBefore {
		t.Fatalf("owner conflicts wrote decisions: %d → %d", decisionsBefore, got)
	}
	var krOwners, usOwners int
	if err := j.db.QueryRow(`SELECT SUM(market='KR'),SUM(market='US') FROM risk_bucket_owners WHERE released_at IS NULL AND account_ref='acct-1'`).Scan(&krOwners, &usOwners); err != nil || krOwners != 1 || usOwners != 1 {
		t.Fatalf("active owners KR=%d US=%d err=%v, want one each", krOwners, usOwners, err)
	}

	// (4) monetary scale-in aggregation: 같은 owner(같은 lane·campaign·prospective)의 두 번째 결정은 기존 사용량 위에 쌓이고,
	// 두 시장이 함께 쓰는 strategy 여유(120-100=20)로 cap 됨 → q 4.
	scaleIn := kr
	scaleIn.suffix, scaleIn.quantity = "flow-kr-scale-in", "4"
	result, err := j.RecordQFinalDecisionAndReserve(ctx, contractRequest(t, j, scaleIn, flowLimits))
	if err != nil || result.Admission.QFinal != 4 || !result.Admission.OwnerReused {
		t.Fatalf("scale-in: q_final=%d ownerReused=%v err=%v, want 4 on the reused owner", result.Admission.QFinal, result.Admission.OwnerReused, err)
	}
	for _, check := range []struct {
		dimension riskbucket.Dimension
		value     string
		want      string
	}{
		{riskbucket.DimensionSymbol, "005930", "70"},
		{riskbucket.DimensionMarket, "KR", "70"},
		{riskbucket.DimensionMarket, "US", "50"},
		{riskbucket.DimensionStrategy, "strategy-alpha", "120"},
	} {
		if got := contractUsage(t, j, check.dimension, check.value); got != check.want {
			t.Fatalf("after scale-in %s/%s usage %s, want %s", check.dimension, check.value, got, check.want)
		}
	}
	// 공유 strategy 가 가득 찼으므로 어느 시장의 새 진입도 cap 거절(q_final 0).
	for _, next := range []contractEntry{
		{"full-kr", "KR", "000660", "lane-kr2", "campaign-kr2", "prospective-kr2", "1", false},
		{"full-us", "US", "MSFT", "lane-us2", "campaign-us2", "prospective-us2", "1", false},
	} {
		if _, err := j.RecordQFinalDecisionAndReserve(ctx, contractRequest(t, j, next, flowLimits)); !riskbucket.IsRefusal(err, riskbucket.RefusalBucketCapExhausted) {
			t.Fatalf("%s entry past the full shared strategy cap: %v", next.market, err)
		}
	}
}

// TestA066KRUSFailureIsolation 는 5.6 문장의 "failure isolation" 을 잼: 한 시장에만 걸린 실패(시장 범위 진입 손실 잠금,
// 시장 범위 대사, 시장 bucket 소진)는 그 시장의 진입만 막고 다른 시장의 진입은 막지 않음. 반대로 **공유 bucket 에 걸린**
// latch 는 설계 D5 대로 두 시장을 함께 막음 — 격리되지 않는 것이 계약임을 같이 고정함.
func TestA066KRUSFailureIsolation(t *testing.T) {
	ctx := context.Background()
	// 두 시장의 격리 진입은 **같은 종목 문자열**을 씀 — 격리가 종목 차이 덕분이 아니라 시장 범위 덕분임을 보이려고.
	const isolationSymbol = "SAME"
	failures := []struct {
		name   string
		inject func(t *testing.T, j *Journal)
		krWant func(error) bool
	}{
		{"KR entry loss lock", func(t *testing.T, j *Journal) {
			activateLossLockForTest(t, j, "acct-1", riskbucket.MarketKR, riskbucket.HorizonShort)
		}, func(err error) bool { return errors.Is(err, ErrRiskBucketEntryLossLocked) }},
		{"KR-scoped reconcile", func(t *testing.T, j *Journal) {
			if _, entered, err := j.EnterReconcile(ctx, EnterReconcileRequest{AccountRef: "acct-1", Symbol: isolationSymbol, ScopeMarket: "kr",
				Cause: ReconcileCauseQuantityMismatch, Evidence: "a066 5.6 isolation probe"}); err != nil || !entered {
				t.Fatalf("reconcile: entered=%v err=%v", entered, err)
			}
		}, func(err error) bool { return errors.Is(err, ErrRiskBucketEntryBlocked) }},
		{"KR market bucket exhausted", func(t *testing.T, j *Journal) {
			if _, err := j.RecordQFinalDecisionAndReserve(ctx, contractRequest(t, j, contractEntry{"fill-kr-market", "KR", "005490", "lane-fill", "campaign-fill", "prospective-fill", "10", false}, isolationLimits)); err != nil {
				t.Fatal(err)
			}
			if _, err := j.RecordQFinalDecisionAndReserve(ctx, contractRequest(t, j, contractEntry{"fill-kr-market-2", "KR", "035420", "lane-fill2", "campaign-fill2", "prospective-fill2", "10", false}, isolationLimits)); err != nil {
				t.Fatal(err)
			}
		}, func(err error) bool { return riskbucket.IsRefusal(err, riskbucket.RefusalBucketCapExhausted) }},
	}
	for _, failure := range failures {
		t.Run(failure.name, func(t *testing.T) {
			j := openTestJournal(t)
			failure.inject(t, j)
			krErr := func() error {
				_, err := j.RecordQFinalDecisionAndReserve(ctx, contractRequest(t, j, contractEntry{"iso-kr", "KR", isolationSymbol, "lane-kr", "campaign-iso-kr", "prospective-iso-kr", "10", false}, isolationLimits))
				return err
			}()
			if !failure.krWant(krErr) {
				t.Fatalf("KR entry under a KR-only failure: %v", krErr)
			}
			if _, err := j.RecordQFinalDecisionAndReserve(ctx, contractRequest(t, j, contractEntry{"iso-us", "US", isolationSymbol, "lane-us", "campaign-iso-us", "prospective-iso-us", "10", true}, isolationLimits)); err != nil {
				t.Fatalf("US entry was blocked by a KR-only failure (%s): %v", failure.name, err)
			}
		})
	}
	t.Run("a latch in a shared bucket blocks both markets (not isolated, by design D5)", func(t *testing.T) {
		j := openTestJournal(t)
		if _, err := j.RecordQFinalDecisionAndReserve(ctx, contractRequest(t, j, contractEntry{"latched-kr", "KR", "005930", "lane-kr", "campaign-latched", "prospective-latched", "10", false}, isolationLimits)); err != nil {
			t.Fatal(err)
		}
		if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET unknown_actual_latched=1`); err != nil {
			t.Fatal(err)
		}
		if _, err := j.RecordQFinalDecisionAndReserve(ctx, contractRequest(t, j, contractEntry{"latched-us", "US", "AAPL", "lane-us", "campaign-latched-us", "prospective-latched-us", "10", false}, isolationLimits)); !errors.Is(err, ErrRiskBucketEntryBlocked) {
			t.Fatalf("US entry into buckets that carry a KR latch was not blocked: %v", err)
		}
	})
}

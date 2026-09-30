package journal

// a126 — 체결 노출은 owner release receipt 와 함께 bucket 사용량을 떠난다(multi-horizon-risk-buckets 델타, design D1~D4).
//
// 시험은 전부 실 원장 위에서 잼: owner 수명주기(admission → 등록 주문 → 실제 체결 · actual 보완 → 종결 → 해제 영수증)를 생산 함수로
// 쌓고, 손상 · 불일치 원장만 SQL 로 직접 만듦(정상 작성자로는 생기지 않는 상태라서).

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

const a126Account = "acct-owner"

// a126Owner 는 실제 체결 금액을 쌓은 owner 하나임 — order 는 10주 전량 체결, actual 가격 5 · 환율 1 → 모든 dimension 에 filled 50.
type a126Owner struct {
	key      riskbucket.OwnerKey
	decision string
	orderID  string
}

func a126Decision(t *testing.T, j *Journal, key riskbucket.OwnerKey) string {
	t.Helper()
	var decision string
	if err := j.db.QueryRow(`SELECT decision_id FROM risk_bucket_final_decisions WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&decision); err != nil {
		t.Fatal(err)
	}
	return decision
}

// a126FilledOwner 는 owner 에 생산 작성자 경로로 실제 체결을 쌓음(filled 50, UNKNOWN latch 는 actual 보완으로 풀림 — 1.4 의 공용 fixture).
func a126FilledOwner(t *testing.T, j *Journal, key riskbucket.OwnerKey, suffix string) a126Owner {
	t.Helper()
	return a126Owner{key: key, decision: a126Decision(t, j, key), orderID: fillRiskBucketOwnerInFull(t, j, key, "a126-"+suffix)}
}

func a126Reserved(symbol, amount string) map[riskbucket.BucketKey]string {
	out := riskReservedMap(amount)
	delete(out, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	out[riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: symbol, PolicyVersion: "policy-v1"}] = amount
	return out
}

// a126ReleasedOwner 는 실제 체결 뒤 수명주기를 닫고 owner 를 해제함 — 영수증이 커밋된 owner.
func a126ReleasedOwner(t *testing.T, suffix string) (*Journal, a126Owner) {
	t.Helper()
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, suffix)
	o := a126FilledOwner(t, j, key, suffix)
	a126Release(t, j, key)
	return j, o
}

func a126Release(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
	t.Helper()
	ctx := context.Background()
	if _, err := j.bindRiskBucketOwnerActual(ctx, key); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, key, true)
	if result, err := j.releaseRiskBucketOwner(ctx, key); err != nil || !result.Released {
		var lifecycle *RiskBucketOwnerLifecycleError
		errors.As(err, &lifecycle)
		t.Fatalf("release result=%+v err=%v lifecycle=%+v", result, err, lifecycle)
	}
}

func a126Usage(t *testing.T, j *Journal, dimension riskbucket.Dimension, value string) riskbucket.JournalBucketUsage {
	t.Helper()
	usage, err := riskbucket.ReadJournalBucketUsage(context.Background(), j.db, a126Account, dimension, value)
	if err != nil {
		t.Fatalf("usage %s/%s: %v", dimension, value, err)
	}
	return usage
}

func a126StoredFilled(t *testing.T, j *Journal, key riskbucket.OwnerKey) []string {
	t.Helper()
	rows, err := j.db.Query(`SELECT filled_minor FROM risk_bucket_reservations WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=? ORDER BY bucket_dimension`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		out = append(out, v)
	}
	return out
}

// 양성 대조 — 해제 영수증이 커밋된 owner 의 filled 는 모든 적용 bucket 의 합에서 빠지고 저장값은 그대로임.
func TestA126AReleasedOwnersFilledLeavesEveryBucket(t *testing.T) {
	j, o := a126ReleasedOwner(t, "leaves")
	for _, stored := range a126StoredFilled(t, j, o.key) {
		if stored != "50" {
			t.Fatalf("stored filled_minor=%s, want the historical 50 kept (파생, 저장값 불변)", stored)
		}
	}
	for dim, value := range map[riskbucket.Dimension]string{
		riskbucket.DimensionHorizon: "SHORT", riskbucket.DimensionMarket: "US", riskbucket.DimensionStrategy: "strategy-alpha",
		riskbucket.DimensionSector: "sector-tech", riskbucket.DimensionSymbol: o.key.Symbol,
	} {
		if usage := a126Usage(t, j, dim, value); usage.FilledMinor != "0" || usage.HeldMinor != "0" {
			t.Errorf("%s/%s usage filled=%s held=%s, want 0 after the release receipt", dim, value, usage.FilledMinor, usage.HeldMinor)
		}
	}
}

var a126Shared = map[riskbucket.Dimension]string{riskbucket.DimensionHorizon: "SHORT", riskbucket.DimensionMarket: "US",
	riskbucket.DimensionStrategy: "strategy-alpha", riskbucket.DimensionSector: "sector-tech"}

func a126AssertShared(t *testing.T, j *Journal, filled string) {
	t.Helper()
	for dim, value := range a126Shared {
		if usage := a126Usage(t, j, dim, value); usage.FilledMinor != filled {
			t.Errorf("%s/%s filled=%s, want %s", dim, value, usage.FilledMinor, filled)
		}
	}
}

// M1 대조 — 해제되지 않은(영수증 없는) 체결 owner 의 filled 는 그대로 셈.
func TestA126AnActiveOwnersFilledStaysInEveryBucket(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "active")
	a126FilledOwner(t, j, key, "active")
	a126AssertShared(t, j, "50")
	// 종결까지 가도 영수증이 없으면 남음(델타 시나리오 「receipt 없는 종결은 사용량을 줄이지 않는다」).
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, key, false)
	a126AssertShared(t, j, "50")
}

// M2 — owner released_at 표식만 있고 영수증이 없으면 떠나지 않음.
func TestA126AReleaseMarkWithoutAReceiptDoesNotLeave(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "mark")
	a126FilledOwner(t, j, key, "mark")
	if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET released_at='2026-03-30T00:45:00Z' WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	a126AssertShared(t, j, "50")
}

// a126LateOrder 는 해제될 owner 에 두 번째 주문을 등록하고 체결 없이 취소 · 확정 실패로 닫음 — 해제 뒤 그 주문의 late BUY 가 올 자리.
func a126LateOrder(t *testing.T, j *Journal, o a126Owner, suffix string) string {
	t.Helper()
	order := "a126-late-" + suffix
	recordConfirmedFillOrderScope(t, j, "a126-late-intent-"+suffix, "a126-late-attempt-"+suffix, order, FillSnapshotScope{
		AccountRef: o.key.AccountID, Market: "us", TradingDay: "2026-03-30", Symbol: o.key.Symbol, Side: "BUY",
	})
	bindRiskOrderAttemptDecision(t, j, order, o.decision)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: order, DecisionID: o.decision, OrderQuantity: 10,
		ReservedMinor: a126Reserved(o.key.Symbol, "0"), CreatedAt: riskFillNow.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{Owner: o.key, DecisionID: o.decision, OrderID: order,
		Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(2 * time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='FAILED_CONFIRMED' WHERE id=?`, "a126-late-attempt-"+suffix); err != nil {
		t.Fatal(err)
	}
	return order
}

func a126LateFill(o a126Owner, order string) FillObservation {
	late := observation(order, "1")
	late.AccountRef, late.Symbol = o.key.AccountID, o.key.Symbol
	late.ObservedAt = "2026-03-30T00:50:00Z"
	return late
}

func a126OrphanLatched(t *testing.T, j *Journal, key riskbucket.OwnerKey) bool {
	t.Helper()
	var n int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND latch='ORPHAN_FILL'`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n > 0
}

// M3 (RecordFill 경로, Campaign hook 결선 전제) — 해제 뒤 등록 주문의 late BUY 가 ORPHAN_FILL 을 세우면 떠남이 되돌려짐.
func TestA126ALateBuyViaRecordFillWithCampaignHookRevertsTheDeparture(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "late-record")
	o := a126FilledOwner(t, j, key, "late-record")
	order := a126LateOrder(t, j, o, "record")
	a126Release(t, j, key)
	a126AssertShared(t, j, "0")
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='CONFIRMED' WHERE id='a126-late-attempt-record'`); err != nil {
		t.Fatal(err)
	}
	if err := j.SetApplyHooks(ApplyHooks{Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	if res, err := j.RecordFill(context.Background(), a126LateFill(o, order)); err != nil || !res.Changed {
		t.Fatalf("late fill=%+v err=%v", res, err)
	}
	if !a126OrphanLatched(t, j, key) {
		t.Fatal("arrangement: the late BUY did not set ORPHAN_FILL")
	}
	a126AssertShared(t, j, "50")
}

// M3 (전략 정산 경로, Campaign hook 미결선 전제) — backfillConfirmedStrategyFillTx 가 직접 binding 을 부르는 자리(strategy_dispatch_runtime.go
// 의 campaignBound=false 갈래)로 같은 late BUY 가 들어와도 되돌림.
func TestA126ALateBuyViaStrategySettlementWithoutCampaignHookRevertsTheDeparture(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "late-strategy")
	o := a126FilledOwner(t, j, key, "late-strategy")
	order := a126LateOrder(t, j, o, "strategy")
	a126Release(t, j, key)
	a126AssertShared(t, j, "0")
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='CONFIRMED' WHERE id='a126-late-attempt-strategy'`); err != nil {
		t.Fatal(err)
	}
	// hook 없음: RecordFill 은 스냅숏만 남기고 binding 을 부르지 않음(apply_hook 의 Campaign 조건).
	if res, err := j.RecordFill(context.Background(), a126LateFill(o, order)); err != nil || !res.Changed {
		t.Fatalf("late fill=%+v err=%v", res, err)
	}
	if a126OrphanLatched(t, j, key) {
		t.Fatal("arrangement: RecordFill without a Campaign hook already latched — the strategy path would not be the one measured")
	}
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	lease := StrategyDispatchLease{StrategyDispatchLeasePlan: StrategyDispatchLeasePlan{AccountRef: o.key.AccountID, Market: StrategyDispatchMarket(o.key.Market), Symbol: o.key.Symbol}}
	if err := j.backfillConfirmedStrategyFillTx(context.Background(), tx, lease, strategyDispatchAttemptAuthority{
		brokerOrderID: order, tradingDay: "2026-03-30", intentID: "a126-late-intent-strategy", settledAt: "2026-03-30T00:51:00Z"}); err != nil {
		t.Fatalf("strategy backfill: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if !a126OrphanLatched(t, j, key) {
		t.Fatal("the strategy settlement path did not set ORPHAN_FILL")
	}
	a126AssertShared(t, j, "50")
}

// M3b 잔여 핀 — 되돌림이 서지 않는 경로 (a) 소유 모호 · (b) 증분 판독 불가 는 `latchRiskBucketFillFailureForScope` 를 부르는데, 그 SQL 은
// 해제 owner 를 찾지 않음(`ow.released_at IS NULL`). 해제 owner 에 scope latch 가 **없음**을 단언 — 배선 로트(R3)가 바꾸면 깨지게.
func TestA126ResidualAmbiguousAndUnreadableLateFillsDoNotRevert(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "residual")
	o := a126FilledOwner(t, j, key, "residual")
	order := a126LateOrder(t, j, o, "residual")
	a126Release(t, j, key)
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	fill := AppliedFill{OrderID: order, AccountRef: o.key.AccountID, Market: "us", Symbol: o.key.Symbol, Side: "BUY", Delta: "1", CumulativeQuantity: "1"}
	for _, detail := range []string{"a126 residual (a) ownership ambiguous", "a126 residual (b) delta unreadable"} {
		if err := latchRiskBucketFillFailureForScope(context.Background(), tx, fill, detail); err != nil {
			t.Fatalf("%s: %v", detail, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if a126OrphanLatched(t, j, key) {
		t.Fatal("residual R3 (a)(b) closed: the released owner got a scope latch — update the residual and a126 tasks 3.1")
	}
	var any int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE prospective_generation=?`, key.ProspectiveGeneration).Scan(&any); err != nil || any != 0 {
		t.Fatalf("scope latches on the released owner=%d err=%v", any, err)
	}
	a126AssertShared(t, j, "0")
}

// a126Admit 은 다른 종목(MSFT) owner 의 fresh admission — 공유 bucket(horizon · market · strategy · sector)에 선언 한도 limit, 수량 q.
func a126Admit(t *testing.T, j *Journal, suffix, limit string, q int64) error {
	return a126AdmitSymbol(t, j, suffix, "MSFT", limit, "0", q)
}

func a126AdmitSymbol(t *testing.T, j *Journal, suffix, symbol, limit, held string, q int64) error {
	t.Helper()
	existing := "a126-existing-" + suffix
	seedExistingRiskReservation(t, j, existing, a126Account)
	plan := riskBucketAdmissionFixture(t, "a126-"+suffix, a126Account, "lane-short", "campaign-a126-"+suffix, "prospective-a126-"+suffix, limit, held)
	plan.ExistingReservationID = existing
	plan.Owner.Key.Market, plan.Owner.Key.Symbol = riskbucket.MarketUS, symbol
	plan.Admission.Policy.QuoteCurrency = "USD"
	rebindRiskBucket(t, &plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: "US", PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: symbol, PolicyVersion: "policy-v1"})
	plan.Admission.QCandidate, plan.Admission.QExistingGuardian = uint64(q), uint64(q)
	_, err := j.CommitRiskBucketAdmission(context.Background(), plan)
	return err
}

// M4 (D3) — 떠난 행이 기록한 한도(100)는 계속 cap 함: 더 큰 한도(1000)를 선언한 진입이 150 을 예약하려 하면 기록 한도로 거절.
func TestA126ADepartedRowsLimitStillCaps(t *testing.T) {
	j, _ := a126ReleasedOwner(t, "limit")
	if err := a126Admit(t, j, "big", "1000", 30); !riskbucket.IsRefusal(err, riskbucket.RefusalBucketCapExhausted) || !strings.Contains(err.Error(), "100") {
		t.Fatalf("err=%v, want the recorded 100 to cap a declared 1000", err)
	}
	// 대조: 떠남 덕에 원장 0 이라 50 은 들어감(떠남이 없으면 snapshot 주장 0 < 원장 50 으로 stale 거절 — 양성 대조).
	if err := a126Admit(t, j, "small", "1000", 10); err != nil {
		t.Fatalf("control within the recorded limit after the departure: %v", err)
	}
}

// 같은 대조의 반대편 — 해제 전에는 같은 진입이 stale 로 거절됨(떠남이 원인임을 보임).
func TestA126WithoutTheDepartureTheSameEntryIsStale(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "nodepart")
	a126FilledOwner(t, j, key, "nodepart")
	if err := a126Admit(t, j, "small", "1000", 10); !riskbucket.IsRefusal(err, riskbucket.RefusalBucketUsageStale) {
		t.Fatalf("err=%v, want stale while the filled 50 still counts", err)
	}
}

// M6 · M6b · codex #4 — 영수증 있는 owner 행의 손상 · 불일치는 판독 불가(scope latch 로 되돌려진 행 포함).
func TestA126CorruptReceiptedRowsAreUnreadable(t *testing.T) {
	cases := map[string]string{
		"HELD row":                  `UPDATE risk_bucket_reservations SET state='HELD',held_minor='5' WHERE owner_prospective_generation=? AND bucket_dimension='sector'`,
		"held remainder":            `UPDATE risk_bucket_reservations SET held_minor='5' WHERE owner_prospective_generation=? AND bucket_dimension='sector'`,
		"owner released_at cleared": `UPDATE risk_bucket_owners SET released_at=NULL WHERE prospective_generation=?`,
		"owner released_at moved":   `UPDATE risk_bucket_owners SET released_at='2026-03-30T23:59:00Z' WHERE prospective_generation=?`,
		"reservation owner key diverged from the decision": `UPDATE risk_bucket_reservations SET owner_prospective_generation='a126-other' WHERE owner_prospective_generation=? AND bucket_dimension='sector'`,
	}
	for name, corrupt := range cases {
		for _, reverted := range []bool{false, true} {
			t.Run(name+map[bool]string{false: "", true: " + scope latch"}[reverted], func(t *testing.T) {
				j, o := a126ReleasedOwner(t, "corrupt")
				if reverted {
					if _, err := j.db.Exec(`INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generation,latch,detail,first_seen_at,last_seen_at) VALUES(?,?,?,?,'ORPHAN_FILL','a126 probe','2026-03-30T00:50:00Z','2026-03-30T00:50:00Z')`,
						o.key.AccountID, string(o.key.Market), o.key.Symbol, o.key.ProspectiveGeneration); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := j.db.Exec(corrupt, o.key.ProspectiveGeneration); err != nil {
					t.Fatal(err)
				}
				if _, err := riskbucket.ReadJournalBucketUsage(context.Background(), j.db, a126Account, riskbucket.DimensionSector, "sector-tech"); !errors.Is(err, riskbucket.ErrJournalUsageInvalid) {
					t.Fatalf("err=%v, want ErrJournalUsageInvalid", err)
				}
			})
		}
	}
}

// a126ActiveMSFT 는 떠남 뒤 공유 bucket 에 새 활성 owner(MSFT)를 세우고 등록 주문에 4 주 체결을 쌓음.
func a126ActiveMSFT(t *testing.T, j *Journal) (riskbucket.OwnerKey, string) {
	t.Helper()
	if err := a126Admit(t, j, "msft", "1000", 10); err != nil {
		t.Fatal(err)
	}
	key := riskbucket.OwnerKey{AccountID: a126Account, Market: riskbucket.MarketUS, Symbol: "MSFT", ProspectiveGeneration: "prospective-a126-msft"}
	decision := a126Decision(t, j, key)
	order := "a126-msft-order"
	recordConfirmedFillOrderScope(t, j, "a126-msft-intent", "a126-msft-attempt", order, FillSnapshotScope{
		AccountRef: a126Account, Market: "us", TradingDay: "2026-03-30", Symbol: "MSFT", Side: "BUY",
	})
	bindRiskOrderAttemptDecision(t, j, order, decision)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: order, DecisionID: decision, OrderQuantity: 10,
		ReservedMinor: a126Reserved("MSFT", "50"), CreatedAt: riskFillNow.Add(20 * time.Minute)}); err != nil {
		t.Fatal(err)
	}
	fill := observation(order, "4")
	fill.AccountRef, fill.Symbol, fill.ObservedAt = a126Account, "MSFT", "2026-03-30T00:55:00Z"
	if res, err := j.RecordFill(context.Background(), fill); err != nil || !res.Changed {
		t.Fatalf("MSFT fill=%+v err=%v", res, err)
	}
	return key, order
}

// D2 파급 — 떠난 행이 손상되면 그 공유 bucket 에서 체결 중인 무관한 활성 owner 의 체결은 커밋되고 그 owner 에 scope latch 만 선다.
func TestA126ACorruptDepartedRowLatchesAnUnrelatedActiveOwnerButKeepsItsFill(t *testing.T) {
	j, gone := a126ReleasedOwner(t, "ripple")
	msft, order := a126ActiveMSFT(t, j)
	if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET released_at=NULL WHERE prospective_generation=?`, gone.key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	fill := observation(order, "6")
	fill.AccountRef, fill.Symbol, fill.ObservedAt = a126Account, "MSFT", "2026-03-30T00:56:00Z"
	if res, err := j.RecordFill(context.Background(), fill); err != nil || !res.Changed {
		t.Fatalf("the unrelated owner's fill must still commit: res=%+v err=%v", res, err)
	}
	var latched int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE prospective_generation=?`, msft.ProspectiveGeneration).Scan(&latched); err != nil || latched == 0 {
		t.Fatalf("unrelated active owner scope latches=%d err=%v, want a latch from the unreadable shared usage", latched, err)
	}
}

// M7 — 떠난 행의 latch 플래그는 latch 집계에 남아 진입을 막음(합에서만 빠짐, Q3 불변 조건 · 잔여 R6 의 보수적 막힘).
func TestA126ADepartedRowsLatchFlagStillBlocksEntry(t *testing.T) {
	j, o := a126ReleasedOwner(t, "latch")
	if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET risk_overage_latched=1 WHERE owner_prospective_generation=? AND bucket_dimension='sector'`, o.key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	usage := a126Usage(t, j, riskbucket.DimensionSector, "sector-tech")
	if usage.FilledMinor != "0" || !usage.OverageLatched {
		t.Fatalf("usage=%+v, want the sum to leave and the latch to stay", usage)
	}
	if err := a126Admit(t, j, "blocked", "1000", 10); !errors.Is(err, ErrRiskBucketEntryBlocked) || !strings.Contains(err.Error(), "RISK_OVERAGE") {
		t.Fatalf("err=%v, want entry blocked by the departed row's RISK_OVERAGE", err)
	}
}

// 델타 「떠남은 다른 owner 의 latch 를 풀지 않는다」 — B 의 RISK_OVERAGE 는 A 의 해제 뒤에도 남아 신규 진입을 막음.
func TestA126ADepartureDoesNotReleaseAnotherOwnersLatch(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "other-latch")
	a126FilledOwner(t, j, key, "other-latch")
	if err := a126AdmitSymbol(t, j, "b", "MSFT", "100", "50", 10); err != nil {
		t.Fatalf("B admission while A still counts: %v", err)
	}
	if _, err := j.db.Exec(`UPDATE risk_bucket_owners SET risk_overage_latched=1 WHERE prospective_generation='prospective-a126-b'`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET risk_overage_latched=1 WHERE owner_prospective_generation='prospective-a126-b'`); err != nil {
		t.Fatal(err)
	}
	a126Release(t, j, key)
	var ownerLatch int
	if err := j.db.QueryRow(`SELECT risk_overage_latched FROM risk_bucket_owners WHERE prospective_generation='prospective-a126-b'`).Scan(&ownerLatch); err != nil || ownerLatch != 1 {
		t.Fatalf("B owner latch=%d err=%v, want it kept after A's departure", ownerLatch, err)
	}
	usage := a126Usage(t, j, riskbucket.DimensionSector, "sector-tech")
	if usage.FilledMinor != "0" || usage.HeldMinor != "50" || !usage.OverageLatched {
		t.Fatalf("usage=%+v, want A's 50 gone, B's held 50 counted and B's latch kept", usage)
	}
	if err := a126Admit(t, j, "c", "1000", 1); !errors.Is(err, ErrRiskBucketEntryBlocked) {
		t.Fatalf("err=%v, want B's latch to keep blocking new exposure", err)
	}
}

// M8 (Q4) — 부분 매도는 사용량을 줄이지 않음.
func TestA126APartialSellLeavesUsageUnchanged(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "sell")
	a126FilledOwner(t, j, key, "sell")
	recordConfirmedFillOrderScope(t, j, "a126-sell-intent", "a126-sell-attempt", "a126-sell-order", FillSnapshotScope{
		AccountRef: a126Account, Market: "us", TradingDay: "2026-03-30", Symbol: key.Symbol, Side: "SELL",
	})
	sell := observation("a126-sell-order", "4")
	sell.AccountRef, sell.Symbol, sell.Side, sell.ObservedAt = a126Account, key.Symbol, "SELL", "2026-03-30T00:40:00Z"
	if res, err := j.RecordFill(context.Background(), sell); err != nil || !res.Changed {
		t.Fatalf("sell fill=%+v err=%v", res, err)
	}
	a126AssertShared(t, j, "50")
}

// a126Fingerprint 는 재시작 replay 비교용 — 공유 bucket 넷의 사용량 · RowDigest · latch.
func a126Fingerprint(t *testing.T, j *Journal) string {
	t.Helper()
	var parts []string
	for _, dim := range []riskbucket.Dimension{riskbucket.DimensionHorizon, riskbucket.DimensionMarket, riskbucket.DimensionStrategy, riskbucket.DimensionSector} {
		u := a126Usage(t, j, dim, a126Shared[dim])
		parts = append(parts, string(dim), u.FilledMinor, u.HeldMinor, u.RowDigest, map[bool]string{true: "L", false: "-"}[u.Latched])
	}
	return strings.Join(parts, "|")
}

// codex #5 — replay 결정성: 해제 전 · 해제 뒤 · 공유 owner 체결과 교차 · 되돌림 뒤 각 순간에 원장을 새로 열어 읽어도 같은 사용량 ·
// RowDigest · latch. 건강한 활성 owner 체결은 total < own(ReplayMismatch)을 만들지 않음.
func TestA126UsageIsReplayDeterministicAcrossReleaseRevertAndSharedFills(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "replay")
	o := a126FilledOwner(t, j, key, "replay")
	order := a126LateOrder(t, j, o, "replay")
	check := func(stage string) {
		t.Helper()
		want := a126Fingerprint(t, j)
		reopened := openTestJournalAt(t, j.path)
		if got := a126Fingerprint(t, reopened); got != want {
			t.Fatalf("%s: reopened fingerprint %s != %s", stage, got, want)
		}
	}
	check("before release")
	a126Release(t, j, key)
	check("after release")
	msft, msftOrder := a126ActiveMSFT(t, j)
	check("after a shared owner's fill")
	// 떠남으로 줄어든 공유 합 위에서 건강한 활성 owner 가 한 번 더 체결 — total 은 own 을 포함하므로 total < own(ReplayMismatch)이 서지 않음.
	more := observation(msftOrder, "6")
	more.AccountRef, more.Symbol, more.ObservedAt = a126Account, "MSFT", "2026-03-30T00:57:00Z"
	if res, err := j.RecordFill(context.Background(), more); err != nil || !res.Changed {
		t.Fatalf("MSFT fill after the departure=%+v err=%v", res, err)
	}
	var mismatch int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE prospective_generation=?`, msft.ProspectiveGeneration).Scan(&mismatch); err != nil || mismatch != 0 {
		var detail string
		_ = j.db.QueryRow(`SELECT detail FROM risk_bucket_scope_latches WHERE prospective_generation=?`, msft.ProspectiveGeneration).Scan(&detail)
		t.Fatalf("healthy active owner got a scope latch=%d err=%v: %s", mismatch, err, detail)
	}
	check("after the shared owner's second fill")
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='CONFIRMED' WHERE id='a126-late-attempt-replay'`); err != nil {
		t.Fatal(err)
	}
	if err := j.SetApplyHooks(ApplyHooks{Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	if res, err := j.RecordFill(context.Background(), a126LateFill(o, order)); err != nil || !res.Changed {
		t.Fatalf("late fill=%+v err=%v", res, err)
	}
	check("after the revert")
}

// 잔여 핀 (f) · codex #1 — 뒤 generation 이 같은 broker order id 를 쓰면 옛 generation 의 late 체결은 활성 owner 가 가로채고 옛 owner 는
// 떠난 채 남음(되돌림 없음). 배선 로트(R3)가 막으면 이 시험이 깨짐.
func TestA126ResidualReusedOrderIDAcrossGenerationsKeepsTheOldOwnerDeparted(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "reuse")
	o := a126FilledOwner(t, j, key, "reuse")
	a126Release(t, j, key)
	if err := a126AdmitSymbol(t, j, "reuse-next", key.Symbol, "1000", "0", 10); err != nil {
		t.Fatalf("next generation admission: %v", err)
	}
	next := riskbucket.OwnerKey{AccountID: a126Account, Market: riskbucket.MarketUS, Symbol: key.Symbol, ProspectiveGeneration: "prospective-a126-reuse-next"}
	nextDecision := a126Decision(t, j, next)
	recordConfirmedFillOrderScope(t, j, "a126-reuse-intent", "a126-reuse-attempt", o.orderID, FillSnapshotScope{
		AccountRef: a126Account, Market: "us", TradingDay: "2026-03-31", Symbol: key.Symbol, Side: "BUY",
	})
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET decision_id=(SELECT legacy.decision_id FROM risk_bucket_final_decisions d JOIN risk_reservations legacy ON legacy.id=d.existing_reservation_id WHERE d.decision_id=?) WHERE id='a126-reuse-attempt'`, nextDecision); err != nil {
		t.Fatal(err)
	}
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: o.orderID, DecisionID: nextDecision, OrderQuantity: 10,
		ReservedMinor: a126Reserved(key.Symbol, "50"), CreatedAt: riskFillNow.Add(30 * time.Minute)}); err != nil {
		t.Fatalf("reusing the broker order id in the next generation: %v", err)
	}
	if err := j.SetApplyHooks(ApplyHooks{Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	late := observation(o.orderID, "1")
	late.AccountRef, late.Symbol, late.TradingDay, late.ObservedAt = a126Account, key.Symbol, "2026-03-31", "2026-03-31T00:40:00Z"
	if res, err := j.RecordFill(context.Background(), late); err != nil || !res.Changed {
		t.Fatalf("fill on the reused order id=%+v err=%v", res, err)
	}
	if a126OrphanLatched(t, j, key) {
		t.Fatal("residual R3 (f) closed: the old generation got ORPHAN_FILL — update the residual and a126 tasks 3.1")
	}
	// 떠남의 결정 사실(영수증 ∧ scope latch 없음)이 그대로 — 옛 owner 에 어떤 scope latch 도 없음.
	var latches int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE prospective_generation=?`, key.ProspectiveGeneration).Scan(&latches); err != nil || latches != 0 {
		t.Fatalf("old owner scope latches=%d err=%v", latches, err)
	}
}

// 잔여 핀 codex #2 — 되돌림 뒤 공유 합이 기록 한도를 넘어도 이미 예약된 다른 owner 에는 bucket latch 도 scope latch 도 서지 않음(제출
// 재검증 `RevalidateQFinalAdmission` 은 latch 만 본다 — freeze census B27). 배선 로트(R3)가 이 구멍을 닫으면 이 시험이 깨짐.
func TestA126ResidualRevertAfterAnotherOwnersReservationLeavesItUnlatched(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "c2")
	o := a126FilledOwner(t, j, key, "c2")
	order := a126LateOrder(t, j, o, "c2")
	a126Release(t, j, key)
	if err := a126AdmitSymbol(t, j, "c2-b", "MSFT", "100", "0", 12); err != nil {
		t.Fatalf("B reserves 60 after A departed: %v", err)
	}
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='CONFIRMED' WHERE id='a126-late-attempt-c2'`); err != nil {
		t.Fatal(err)
	}
	if err := j.SetApplyHooks(ApplyHooks{Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	if res, err := j.RecordFill(context.Background(), a126LateFill(o, order)); err != nil || !res.Changed {
		t.Fatalf("late fill=%+v err=%v", res, err)
	}
	usage := a126Usage(t, j, riskbucket.DimensionSector, "sector-tech")
	if usage.FilledMinor != "50" || usage.HeldMinor != "60" {
		t.Fatalf("usage=%+v, want A's 50 back and B's 60 held (110 over the recorded 100)", usage)
	}
	var latches int
	if err := j.db.QueryRow(`SELECT (SELECT count(*) FROM risk_bucket_scope_latches WHERE prospective_generation='prospective-a126-c2-b')
		+ (SELECT count(*) FROM risk_bucket_reservations WHERE owner_prospective_generation='prospective-a126-c2-b' AND (risk_overage_latched=1 OR unknown_actual_latched=1))
		+ (SELECT count(*) FROM risk_bucket_owners WHERE prospective_generation='prospective-a126-c2-b' AND (risk_overage_latched=1 OR unknown_actual_latched=1))`).Scan(&latches); err != nil || latches != 0 {
		t.Fatalf("B latches=%d err=%v — residual C2 closed? update a126 R3 and tasks 3.1", latches, err)
	}
}

// a066 결함 수리의 결속 시험(Manager 조건 ②) — 해소 3상에서 해제 검사의 가부가 체결 재구성의 해소 판정과 같다. 두 자리는 같은 SQL 조각
// (riskBucketFillActualResolvedSQL)을 쓰고, 이 시험은 그 조각을 세 상태의 행에 직접 평가해 두 쪽이 쓰는 모양(해소 → 1, 해제 → NOT)을 잰다.
func TestA126ReleaseAndTransitionShareOneFillResolutionRule(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	if _, err := j.db.ExecContext(ctx, `CREATE TEMP TABLE a126_fills(fill_id TEXT, actual_known INTEGER)`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx, `INSERT INTO a126_fills VALUES('known-only',1),('evidence-only',0),('neither',0)`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx, `CREATE TEMP TABLE risk_bucket_fill_actual_evidence(fill_id TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx, `INSERT INTO temp.risk_bucket_fill_actual_evidence VALUES('evidence-only')`); err != nil {
		t.Fatal(err)
	}
	for id, want := range map[string]bool{"known-only": true, "evidence-only": true, "neither": false} {
		var resolved, blocking int
		if err := j.db.QueryRowContext(ctx, `SELECT CASE WHEN `+riskBucketFillActualResolvedSQL+` THEN 1 ELSE 0 END FROM a126_fills f WHERE f.fill_id=?`, id).Scan(&resolved); err != nil {
			t.Fatal(err)
		}
		if err := j.db.QueryRowContext(ctx, `SELECT count(*) FROM a126_fills f WHERE f.fill_id=? AND NOT `+riskBucketFillActualResolvedSQL, id).Scan(&blocking); err != nil {
			t.Fatal(err)
		}
		if (resolved == 1) != want || (blocking == 1) == want {
			t.Errorf("%s: resolved=%d release-blocking=%d, want resolved=%v and blocking the opposite", id, resolved, blocking, want)
		}
	}
}

// 결속 시험의 생산 짝 — 생산 경로로 만든 두 상태(actual 보완 뒤 = evidence 만 · 보완 전 = 둘 다 없음)에서 해제 검사가 재구성과 같은 답.
func TestA126ReleaseAcceptsAnActualCompletedFillAndRefusesAnUnresolvedOne(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "resolution")
	o := a126FilledOwner(t, j, key, "resolution") // evidence 만 — 해소
	a126Release(t, j, key)
	_ = o

	j2, key2 := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "unresolved")
	order := "a126-unresolved-order"
	decision := a126Decision(t, j2, key2)
	recordConfirmedFillOrderScope(t, j2, "a126-unresolved-intent", "a126-unresolved-attempt", order, FillSnapshotScope{
		AccountRef: a126Account, Market: "us", TradingDay: "2026-03-30", Symbol: key2.Symbol, Side: "BUY",
	})
	bindRiskOrderAttemptDecision(t, j2, order, decision)
	if err := j2.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: order, DecisionID: decision, OrderQuantity: 10,
		ReservedMinor: a126Reserved(key2.Symbol, "50"), CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	fill := observation(order, "10")
	fill.AccountRef, fill.Symbol, fill.State, fill.Terminal = a126Account, key2.Symbol, "FILLED", true
	if res, err := j2.RecordFill(context.Background(), fill); err != nil || !res.Changed {
		t.Fatalf("fill=%+v err=%v", res, err)
	}
	// owner UNKNOWN latch 를 치워 unresolved_fill 검사가 앞선 owner_latch 에 가려지지 않게 함(검사 순서 — 층을 내려서 잼).
	if _, err := j2.db.Exec(`UPDATE risk_bucket_owners SET unknown_actual_latched=0 WHERE prospective_generation=?`, key2.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	if _, err := j2.db.Exec(`UPDATE risk_bucket_reservations SET unknown_actual_latched=0 WHERE owner_prospective_generation=?`, key2.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	refreshRiskBucketOwnerSnapshot(t, j2, key2, "a126-unlatched")
	if _, err := j2.bindRiskBucketOwnerActual(context.Background(), key2); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j2, key2, true)
	_, err := j2.releaseRiskBucketOwner(context.Background(), key2)
	var lifecycle *RiskBucketOwnerLifecycleError
	if !errors.As(err, &lifecycle) || lifecycle.BlockingField != "unresolved_fill" {
		t.Fatalf("err=%v lifecycle=%+v, want the unresolved fill to block", err, lifecycle)
	}
}

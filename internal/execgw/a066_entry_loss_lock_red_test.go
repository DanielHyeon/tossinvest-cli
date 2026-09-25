//go:build a066_red_5_5

// [RED, 5.5 대기] a066 task 2.7 — horizon/market 진입 손실 잠금(loss lock)과 bucket 실패는
// 노출 증가(EXPOSURE_RAISING)만 막고 손절·비상 청산·대사·체결 감지를 막거나 지연시키지 못함을 고정함.
//
// 왜 빌드 태그 뒤에 있는가: loss lock 은 아직 없음(구현은 task 5.5). 이 파일은 5.5 가 채울 seam 하나
// (`activateEntryLossLock`)만 참조하고, 그 seam 이 nil 인 동안 시험은 반드시 빨갛게 멈춤. 기본 스위트
// (`make test`)와 `make test-seams`(태그 `tossos_testseams`)는 이 태그를 켜지 않으므로 스위트를 깨지 않음.
//
// 실행: go test -count=1 -tags a066_red_5_5 -run 'TestA066' ./internal/execgw
//
// 5.5 GREEN 조건: (1) seam 을 durable 잠금 활성화 API 로 연결, (2) 이 파일의 빌드 태그 제거,
// (3) 아래 두 시험이 태그 없이 통과. 시험의 기대값을 약하게 고쳐 통과시키는 것은 GREEN 이 아님.
package execgw_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// activateEntryLossLock 는 5.5 가 채울 유일한 seam 임.
// 요구하는 것은 보수 방향(잠금 활성화)의 durable 기록 하나뿐임 — 완화(해제)는 사람 승인·audit 경로라
// 이 시험의 범위가 아님. 잠금의 판정은 journal 에 영속된 상태를 exposure-raising 경로가 읽어서 내려야 함.
var activateEntryLossLock func(ctx context.Context, j *journal.Journal, account string,
	market riskbucket.Market, horizon riskbucket.Horizon, at time.Time) error

// requireEntryLossLockSeam 은 seam 부재를 RED 사유로 명시하고 시험을 멈춤.
func requireEntryLossLockSeam(t *testing.T) {
	t.Helper()
	if activateEntryLossLock == nil {
		t.Fatal("[RED, 5.5 대기] a066 entry loss lock 이 없음: horizon×market 잠금을 durable 로 활성화하는 " +
			"seam(activateEntryLossLock)이 연결되지 않았음 — task 5.5 가 구현해야 함")
	}
}

// lockScope 는 잠금 범위 하나(시장×horizon)를 나타냄.
type lockScope struct {
	market  riskbucket.Market
	horizon riskbucket.Horizon
}

// lossLockQFinalRequest 는 qFinalKRRequest 와 같은 KR/KRW q_final 진입 요청을 horizon 만 바꿔 만듦.
// 같은 모양의 요청이 잠금 없는 대조 rig 에서 통과해야 잠금 rig 의 거절이 잠금 때문임을 증명할 수 있음.
func lossLockQFinalRequest(t *testing.T, rig *guardianRig, suffix string, horizon riskbucket.Horizon) execgw.QFinalEntryIssuance {
	t.Helper()
	request := qFinalKRRequest(t, rig, suffix, 20)
	if horizon == riskbucket.HorizonShort {
		return request
	}
	// horizon bucket 만 교체함 — provenance 는 key 에 봉인되므로 key 를 바꾸면 봉인도 새로 만들어야 함.
	now := fixedNow
	for i, bucket := range request.Admission.Admission.Buckets {
		if bucket.Key.Dimension != riskbucket.DimensionHorizon {
			continue
		}
		key := riskbucket.BucketKey{Dimension: riskbucket.DimensionHorizon, Value: string(horizon), PolicyVersion: bucket.Key.PolicyVersion}
		policyEvidence := riskbucket.Evidence{Source: riskbucket.RiskPolicyAuthoritySource, Version: key.PolicyVersion, Digest: "policy-" + suffix + "-horizon-" + string(horizon), Official: true, Frozen: true, ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)}
		policyProvenance, err := riskbucket.NewPolicyProvenance(key, policyEvidence)
		if err != nil {
			t.Fatal(err)
		}
		binding := riskbucket.BucketSnapshotBinding{Key: key, LimitMinor: bucket.LimitMinor, FilledMinor: bucket.FilledMinor, HeldMinor: bucket.HeldMinor, SnapshotVersion: bucket.SnapshotVersion}
		snapshotEvidence := riskbucket.Evidence{Source: riskbucket.RiskSnapshotAuthoritySource, Version: binding.SnapshotVersion, Digest: "snapshot-" + suffix + "-horizon-" + string(horizon), Official: true, Frozen: true, ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)}
		snapshotProvenance, err := riskbucket.NewSnapshotProvenance(binding, snapshotEvidence)
		if err != nil {
			t.Fatal(err)
		}
		request.Admission.Admission.Buckets[i] = riskbucket.BucketSnapshot{Key: key, LimitMinor: binding.LimitMinor, FilledMinor: binding.FilledMinor, HeldMinor: binding.HeldMinor, SnapshotVersion: binding.SnapshotVersion, PolicyProvenance: policyProvenance, SnapshotProvenance: snapshotProvenance}
		request.Admission.Snapshots[i] = journal.RiskBucketSnapshotReference{Key: key, SnapshotID: "snapshot-" + suffix + "-horizon-" + string(horizon), SnapshotDigest: snapshotEvidence.Digest, SnapshotVersion: binding.SnapshotVersion, PolicyDigest: policyEvidence.Digest, ObservedAt: policyEvidence.ObservedAt, FreshUntil: policyEvidence.FreshUntil}
	}
	request.Admission.Owner.LaneID = "lane-" + string(horizon)
	return request
}

// TestA066EntryLossLockIsEntryOnlyPerHorizonAndMarket 는 잠금이 자기 범위(시장×horizon)의
// 신규 노출 증가 결정만 막고, 다른 horizon·다른 시장으로 전파되지 않음을 고정함
// (risk-management spec "Medium lock과 short entry").
func TestA066EntryLossLockIsEntryOnlyPerHorizonAndMarket(t *testing.T) {
	cases := []struct {
		name    string
		locked  []lockScope
		horizon riskbucket.Horizon
		refused bool
	}{
		{name: "control: no lock admits KR SHORT", horizon: riskbucket.HorizonShort},
		{name: "control: no lock admits KR MEDIUM", horizon: riskbucket.HorizonMedium},
		{name: "KR MEDIUM lock does not reach KR SHORT", locked: []lockScope{{riskbucket.MarketKR, riskbucket.HorizonMedium}}, horizon: riskbucket.HorizonShort},
		{name: "KR SHORT lock refuses KR SHORT", locked: []lockScope{{riskbucket.MarketKR, riskbucket.HorizonShort}}, horizon: riskbucket.HorizonShort, refused: true},
		{name: "KR MEDIUM lock refuses KR MEDIUM", locked: []lockScope{{riskbucket.MarketKR, riskbucket.HorizonMedium}}, horizon: riskbucket.HorizonMedium, refused: true},
		{name: "US locks do not reach KR", locked: []lockScope{{riskbucket.MarketUS, riskbucket.HorizonShort}, {riskbucket.MarketUS, riskbucket.HorizonMedium}}, horizon: riskbucket.HorizonShort},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 사례마다 새 journal — 앞 사례의 owner·예약이 판정에 섞이지 않게 함.
			rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
				options.NewID = fixedIDs("lossl-decision", "lossl-nonce")
			})
			ctx := context.Background()
			if len(tc.locked) > 0 {
				requireEntryLossLockSeam(t)
			}
			for _, scope := range tc.locked {
				if err := activateEntryLossLock(ctx, rig.journal, "acct-7", scope.market, scope.horizon, rig.clock.Now()); err != nil {
					t.Fatalf("activate %s/%s: %v", scope.market, scope.horizon, err)
				}
			}
			issued, err := rig.guardian.IssueQFinalEntry(ctx, lossLockQFinalRequest(t, rig, "lossl", tc.horizon))
			if !tc.refused {
				if err != nil || issued.RiskBucketReceipt.QFinal != 10 {
					t.Fatalf("entry must be admitted: q_final=%d err=%v", issued.RiskBucketReceipt.QFinal, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("locked %s entry was admitted: decision=%s q_final=%d", tc.horizon, issued.Decision.ID, issued.RiskBucketReceipt.QFinal)
			}
			// 거절은 기록 권한을 남기지 않아야 함 — 결정 행도, owner 도, 예약도 없음.
			if _, lookupErr := rig.journal.LookupDecision(ctx, "lossl-decision"); !errors.Is(lookupErr, journal.ErrDecisionNotFound) {
				t.Fatalf("refused entry left a decision row: lookup err=%v", lookupErr)
			}
			if held, heldErr := rig.journal.HeldReservations(ctx, "acct-7"); heldErr != nil || len(held) != 0 {
				t.Fatalf("refused entry left holds: %+v err=%v", held, heldErr)
			}
		})
	}
}

// TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths 는 두 시장·두 horizon 이 모두 잠기고
// q_final bucket 판정이 실패하는 동안에도 손절 매도·비상 청산·취소·대사 진입·체결 감지가 그대로
// 진행됨을 고정함(multi-horizon-risk-buckets spec "loss lock 중 stop", risk-management spec
// "Market lock 중 emergency exit"). 노출 증가 쪽 거절은 양성 대조군임.
func TestA066LossLockAndBucketFailureNeverBlockRiskReducingPaths(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-lock-exit"}}
	gw, j, clk := newGateway(t, broker)
	ctx := context.Background()

	// 양성 대조군(오늘도 통과): q_final 표식이 있지만 a066 admission 이 없는 노출 증가 결정은
	// broker 전에 bucket 불일치로 거절됨. 아래의 위험 감소 경로가 "bucket 판정을 안 거쳐서" 통과함을 보이려면
	// 같은 Gateway 에서 bucket 판정이 실제로 살아 있어야 함.
	policyVersion, err := journal.QFinalPolicyVersion("guardian-v1", "lossl-bucket-failure")
	if err != nil {
		t.Fatal(err)
	}
	limitsJSON, err := execgw.EncodeLimits(testLimits())
	if err != nil {
		t.Fatal(err)
	}
	buy, err := orderintent.NormalizePlace(orderintent.PlaceInput{Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit", Quantity: 2, Price: 70000, CurrencyMode: "KRW"})
	if err != nil {
		t.Fatal(err)
	}
	// 단계마다 새 결정을 씀 — 같은 결정을 다시 내면 nonce 재사용으로 거절되어 bucket 판정을 재지 못함.
	assertRaisingRefused := func(stage string) {
		t.Helper()
		version, err := j.ReservationVersion(ctx, "acct-7")
		if err != nil {
			t.Fatal(err)
		}
		raising, err := j.RecordDecisionAndReserve(ctx, journal.IssueRequest{
			Decision: journal.DecisionRequest{
				ID: "lossl-raising-" + stage, AccountRef: "acct-7", SafetyClass: journal.SafetyClassExposureRaising, Kind: journal.KindPlace,
				Preimage:   journal.RiskIntent{AccountRef: "acct-7", Market: "kr", Symbol: "005930", Side: "BUY", Quantity: "2", EntryPrice: "70000", StopPrice: "69000", TargetPrice: "72000", PolicyVersion: policyVersion},
				LimitsJSON: limitsJSON, Nonce: "lossl-raising-nonce-" + stage, IssuedAt: clk.Now(), ExpiresAt: clk.Now().Add(time.Minute),
			},
			Reserve: journal.ReserveRequest{
				SnapshotAsOf: clk.Now(), ObservedVersion: version,
				SnapshotUsage: []journal.AggregateAmount{{Kind: journal.ReservationKindOpenExposure, Amount: "0", Currency: "KRW"}},
				Limits:        []journal.AggregateAmount{{Kind: journal.ReservationKindOpenExposure, Amount: "5000000", Currency: "KRW"}},
				Reservations:  []journal.ReservationRequest{{ID: "lossl-raising-hold-" + stage, Kind: journal.ReservationKindOpenExposure, Amount: "140000", Currency: "KRW"}},
			},
		})
		if err != nil {
			t.Fatalf("%s: recording the exposure-raising probe decision: %v", stage, err)
		}
		before, _, _ := broker.totals()
		_, err = gw.Place(ctx, execgw.PlaceRequest{Intent: buy, Decision: execgw.GuardianDecision{ID: raising.Decision.ID, Generation: raising.Decision.Generation}})
		var rejected *execgw.RejectedError
		after, _, _ := broker.totals()
		if !errors.As(err, &rejected) || rejected.Reason == execgw.ReasonGuardianNonceReused || after != before {
			t.Fatalf("%s: exposure-raising entry reached the broker or was not refused by a risk gate: rejected=%+v places %d->%d err=%v", stage, rejected, before, after, err)
		}
	}
	assertRaisingRefused("before-lock")

	// 두 시장 × 두 horizon 전부 잠금 — 가장 넓은 잠금에서도 위험 감소가 열려 있어야 함.
	requireEntryLossLockSeam(t)
	for _, market := range []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS} {
		for _, horizon := range []riskbucket.Horizon{riskbucket.HorizonShort, riskbucket.HorizonMedium} {
			if err := activateEntryLossLock(ctx, j, "acct-7", market, horizon, clk.Now()); err != nil {
				t.Fatalf("activate %s/%s: %v", market, horizon, err)
			}
		}
	}
	assertRaisingRefused("under-lock")

	// 위험 감소 경로는 짧은 마감 안에 broker 에 정확히 한 번씩 닿아야 함 — 잠금 평가·bucket 계산·FX 수집을
	// 기다리는 경로가 끼어들면 이 마감이 먼저 끝남.
	within := func() (context.Context, context.CancelFunc) { return context.WithTimeout(ctx, 5*time.Second) }

	// 손절 매도(KR, limit).
	stop := orderintent.PlaceIntent{Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit", Quantity: 2, Price: 69000, CurrencyMode: "KRW"}
	stopCtx, cancelStop := within()
	defer cancelStop()
	if out, err := gw.Place(stopCtx, execgw.PlaceRequest{Intent: stop, Decision: exitDecision(t, j, clk, journal.KindPlace, stop.Market, stop.Symbol, stop.Side, stop.Quantity)}); err != nil || out.State != journal.StateConfirmed {
		t.Fatalf("KR stop exit under lock: state=%s err=%v", out.State, err)
	}
	// 비상 청산(US) — live place 가 받는 좁은 주문 모양(limit)만 씀. market 매도는 잠금과 무관하게
	// trading 정책이 NOT_DISPATCHED 로 거절하므로 이 시험의 판정 대상이 될 수 없음.
	emergency := orderintent.PlaceIntent{Symbol: "AAPL", Market: "us", Side: "sell", OrderType: "limit", Quantity: 5, Price: 150, CurrencyMode: "USD"}
	emergencyCtx, cancelEmergency := within()
	defer cancelEmergency()
	if out, err := gw.Place(emergencyCtx, execgw.PlaceRequest{Intent: emergency, Decision: exitDecision(t, j, clk, journal.KindPlace, emergency.Market, emergency.Symbol, emergency.Side, emergency.Quantity)}); err != nil || out.State != journal.StateConfirmed {
		t.Fatalf("US emergency exit under lock: state=%s err=%v", out.State, err)
	}
	// reduce-only 취소(KR).
	cancelCtx, cancelCancel := within()
	defer cancelCancel()
	if _, err := gw.Cancel(cancelCtx, execgw.CancelRequest{
		Intent:   orderintent.CancelIntent{OrderID: "O-open-buy", Symbol: "005930"},
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	}); err != nil {
		t.Fatalf("reduce-only cancel under lock: %v", err)
	}
	if places, cancels, _ := broker.totals(); places != 2 || cancels != 1 {
		t.Fatalf("risk-reducing broker calls under lock: places=%d cancels=%d, want 2 and 1", places, cancels)
	}

	// 대사 진입(시장 범위 KR) — 잠금이 대사 상태 기록을 막으면 안 됨.
	if _, entered, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest{
		AccountRef: "acct-7", Symbol: "005930", ScopeMarket: "kr",
		Cause: journal.ReconcileCauseQuantityMismatch, Evidence: "a066 2.7 loss-lock reconcile probe",
	}); err != nil || !entered {
		t.Fatalf("reconcile entry under lock: entered=%v err=%v", entered, err)
	}

	// 체결 감지 — 잠금 중에도 손절 매도의 체결이 거절·잘림 없이 원장에 적용되어야 함.
	fill, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: "O-lock-exit", AccountRef: "acct-7", Symbol: "005930", Market: "kr",
		TradingDay: "2026-03-30", Side: "SELL", State: "FILLED", Terminal: true,
		Quantity: "2", FilledQuantity: "2", AveragePrice: "69000",
		ObservedAt: clk.Now().UTC().Format(time.RFC3339),
	})
	if err != nil || fill.FailClosed || fill.Delta != "2" {
		t.Fatalf("fill detection under lock: delta=%q failClosed=%v reason=%s err=%v", fill.Delta, fill.FailClosed, fill.Reason, err)
	}

	// 위험 감소가 전부 지난 뒤에도 노출 증가는 여전히 막혀 있어야 함 — 잠금이 한 번 쓰이고 풀리면 안 됨.
	assertRaisingRefused("after-risk-reducing")
}

//go:build tossos_testseams

package execgw_test

// a157 첫 다리 결속 판정의 보고 순서 — 실제 Guardian · 실제 원장(시험 원장) · 실제 주간 예약으로 PrecheckQFinalCampaignFirstLeg 의 거절을 잼.
//
// 행은 셋으로 나뉨(tasks 0.4):
//   - 뒤집을 행: Guardian 상한 0 + 결속 결함 Bn(편집 전에는 상한 0 이 결속 결함을 가렸음), 입력 하나가 두 판정을 거는 (f) 두 행.
//   - 뒤에도 같아야 할 행: 결속 결함 단독 · 진입 거절 단독 · 일일 손실(한도 · 측정 불가) + 결속 결함 · 통과 봉인 값 · nil Guardian.
//   - 판정은 거절의 타입 · 코드 · field · 문구(그리고 측정 불가 표지 · 주간 충돌 원인)로만 비교함 — 문구는 거절 자신의 Error().

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// 거절 판정 문자열 — 행 기대값과 바로 비교할 수 있게 거절 하나를 「종류:코드/field: 문구 (+표지)」 한 줄로 만듦.
func a157Verdict(err error) string {
	if err == nil {
		return "ACCEPT"
	}
	var issue *execgw.IssueRefusal
	if errors.As(err, &issue) {
		out := "chain:" + string(issue.Stage) + "/" + issue.Reason
		var unmeasurable *execgw.DailyLossUnmeasurable
		if errors.As(err, &unmeasurable) {
			out += " +unmeasurable"
		}
		return out
	}
	var refusal *execgw.QFinalRefusal
	if errors.As(err, &refusal) {
		out := "q:" + string(refusal.Code) + "/" + refusal.Field + ": " + refusal.Detail
		if errors.Is(err, journal.ErrWeeklyReservationConflict) {
			out += " +weekly-conflict"
		}
		return out
	}
	if errors.Is(err, journal.ErrInvalidRequest) {
		return "journal:invalid-request"
	}
	return "other: " + err.Error()
}

// 기대 판정 상수 — 결속 일곱(B5 는 셋) · 진입 거절 · 일일 손실.
const (
	a157B2          = "journal:invalid-request"
	a157B3          = "q:OWNER_CONFLICT/prospective_generation: first-leg token is journal-owned"
	a157B4          = "q:RISK_CALCULATION_INVALID/first_leg_price_unit: sealed strategy execution price unit is invalid"
	a157B5          = "q:RISK_CALCULATION_INVALID/first_leg_lineage: sealed strategy result, q_candidate, owner and campaign are not exactly bound"
	a157B6          = "q:OWNER_CONFLICT/position_generation: strategy prospective generation must be exactly campaign current generation plus one"
	a157B7          = "q:RISK_CALCULATION_INVALID/first_leg_strategy: complete strategy attempt and activation binding required"
	a157B8          = "q:RISK_CALCULATION_INVALID/weekly_reservation: weekly first-leg reservation authority is missing or does not match the sealed lane +weekly-conflict"
	a157DailyLimit  = "chain:chain/DAILY_LOSS_LIMIT_REACHED"
	a157Unmeasured  = "chain:chain/INPUT_UNAVAILABLE +unmeasurable"
	a157Unmarked    = "chain:chain/INPUT_UNAVAILABLE"
	a157CapZeroWant = "q:EXISTING_GUARDIAN_CAP/q_existing_guardian: risk: strategy sizing permits fewer than one unit"
	a157BucketWant  = "q:BUCKET_CAP_EXHAUSTED/monetary_bucket: BUCKET_CAP_EXHAUSTED (monetary_bucket)"
)

// a157Binding 은 결속 결함 하나를 요청에 심는 변형과 그 단독 거절의 기대 판정.
type a157Binding struct {
	name   string
	want   string
	mutate func(*execgw.QFinalCampaignFirstLegIssuance)
}

// a157Bindings 는 결속 판정 일곱의 결함 — B5 는 입력 출처가 바뀌는 두 갈래(q_candidate · 소유자 LaneID/CampaignID)를 따로 둠(design D1 표).
// 모두 KR 주간 레인 요청에 심음(B8 은 그 레인의 예약 결속을 지움).
func a157Bindings() []a157Binding {
	return []a157Binding{
		{"B2-policy-version", a157B2, func(r *execgw.QFinalCampaignFirstLegIssuance) { r.Entry.Admission.TransactionID = "tx|bad" }},
		{"B3-prospective", a157B3, func(r *execgw.QFinalCampaignFirstLegIssuance) {
			r.Entry.Admission.Owner.Key.ProspectiveGeneration = "pg"
		}},
		{"B4-price-unit", a157B4, func(r *execgw.QFinalCampaignFirstLegIssuance) {
			r.Result.ExecutionTerms = strategyflow.ExecutionTerms{}
		}},
		{"B5-q-candidate", a157B5, func(r *execgw.QFinalCampaignFirstLegIssuance) { r.Entry.QCandidate-- }},
		{"B5-owner-lane", a157B5, func(r *execgw.QFinalCampaignFirstLegIssuance) { r.Entry.Admission.Owner.LaneID = "lane-other" }},
		// 캠페인 요청과 소유자를 함께 바꿔 계보 쪽 비교만 걸리게 함(소유자 입력을 계보 자기 비교로 바꾼 변이를 이 행이 잡음).
		{"B5-owner-campaign", a157B5, func(r *execgw.QFinalCampaignFirstLegIssuance) {
			r.Entry.Admission.Owner.CampaignID = "campaign-other"
			r.Campaign.CampaignID = "campaign-other"
		}},
		{"B6-position-generation", a157B6, func(r *execgw.QFinalCampaignFirstLegIssuance) { r.Campaign.ExpectedPositionGeneration = 3 }},
		{"B7-attempt", a157B7, func(r *execgw.QFinalCampaignFirstLegIssuance) { r.AttemptID = " a" }},
		{"B8-weekly", a157B8, func(r *execgw.QFinalCampaignFirstLegIssuance) { r.Weekly = nil }},
	}
}

func a157Descriptor(t *testing.T, match func(strategyflow.Descriptor) bool) strategyflow.Descriptor {
	t.Helper()
	for _, d := range strategyflow.Descriptors() {
		if match(d) {
			return d
		}
	}
	t.Fatal("arrangement: no matching strategy descriptor")
	return strategyflow.Descriptor{}
}

func a157KRWeekly(t *testing.T) strategyflow.Descriptor {
	return a157Descriptor(t, func(d strategyflow.Descriptor) bool { return d.LaneID == weeklyvaluelane.KRWeeklyLaneID })
}

func a157US(t *testing.T) strategyflow.Descriptor {
	return a157Descriptor(t, func(d strategyflow.Descriptor) bool {
		return string(d.Market) == "US" && d.LaneID != weeklyvaluelane.USWeeklyLaneID
	})
}

// a157Setup 은 요청의 진입 판정 쪽 상황(진입 거절 Y)을 만드는 준비 — 예산 · 일일 손실 원장 행 · 요청 변형.
type a157Setup struct {
	name   string
	budget string
	rows   func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error)
	entry  func(*testing.T, *execgw.QFinalCampaignFirstLegIssuance, string)
}

func a157DailyLossRows(rows ...journal.DailyLossRow) func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error) {
	return func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error) { return rows, nil }
}

// fixedNow = KST 2026-03-30 10:30 — 일 키 2026-03-30. 한도(기본 정책 100000 KRW)와 같은 손실 하나 = 한도 도달, 성과 행 없는 청산 = 측정 불가.
func a157LimitRows() func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error) {
	pnl := "-100000"
	return a157DailyLossRows(journal.DailyLossRow{PositionID: "p-a157", Market: "kr", Symbol: "005930", ClosedAt: "2026-03-30T00:30:00Z", Outcome: &pnl})
}

func a157UnmeasurableRows() func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error) {
	return a157DailyLossRows(journal.DailyLossRow{PositionID: "p-a157", Market: "kr", Symbol: "005930", ClosedAt: "2026-03-30T00:30:00Z"})
}

func a157CapZero() a157Setup { return a157Setup{name: "cap-zero", budget: "1"} }
func a157Limit() a157Setup   { return a157Setup{name: "daily-loss-limit", rows: a157LimitRows()} }
func a157Unmeasurable() a157Setup {
	return a157Setup{name: "daily-loss-unmeasurable", rows: a157UnmeasurableRows()}
}

// 측정과 무관한 입력 불가(호출자 일일 손실 값이 음수) — 측정 불가 표지가 없는 체인 거절이라 일일 손실 우선 대상이 아님(near miss).
func a157UnmarkedInput() a157Setup {
	return a157Setup{name: "unmarked-input-unavailable", entry: func(_ *testing.T, r *execgw.QFinalCampaignFirstLegIssuance, _ string) {
		r.Entry.Account.DailyRealizedLoss.Amount = "-1"
	}}
}

// 버킷 고갈 — strategy 차원 한도를 0 으로 다시 봉인한 버킷 스냅숏.
func a157BucketExhausted() a157Setup {
	return a157Setup{name: "bucket-exhausted", entry: func(t *testing.T, r *execgw.QFinalCampaignFirstLegIssuance, suffix string) {
		a157RebucketKR(t, &r.Entry, suffix, riskbucket.DimensionStrategy, "0")
	}}
}

// a157RebucketKR 는 KR 요청의 다섯 버킷 스냅숏을 다시 봉인하되 한 차원의 한도만 바꿈(qFinalKRRequest 와 같은 값 · 같은 증거 모양).
func a157RebucketKR(t *testing.T, r *execgw.QFinalEntryIssuance, suffix string, dimension riskbucket.Dimension, limit string) {
	t.Helper()
	now := fixedNow
	var buckets []riskbucket.BucketSnapshot
	var references []journal.RiskBucketSnapshotReference
	for _, old := range r.Admission.Admission.Buckets {
		key := old.Key
		policyEvidence := riskbucket.Evidence{Source: riskbucket.RiskPolicyAuthoritySource, Version: key.PolicyVersion,
			Digest: "policy-" + suffix + "-" + string(key.Dimension), Official: true, Frozen: true, ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)}
		policyProvenance, err := riskbucket.NewPolicyProvenance(key, policyEvidence)
		if err != nil {
			t.Fatal(err)
		}
		limitMinor := old.LimitMinor
		if key.Dimension == dimension {
			limitMinor = limit
		}
		binding := riskbucket.BucketSnapshotBinding{Key: key, LimitMinor: limitMinor, FilledMinor: "0", HeldMinor: "0", SnapshotVersion: "snapshot-v1"}
		snapshotEvidence := riskbucket.Evidence{Source: riskbucket.RiskSnapshotAuthoritySource, Version: binding.SnapshotVersion,
			Digest: "snapshot-" + suffix + "-" + string(key.Dimension), Official: true, Frozen: true, ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)}
		snapshotProvenance, err := riskbucket.NewSnapshotProvenance(binding, snapshotEvidence)
		if err != nil {
			t.Fatal(err)
		}
		buckets = append(buckets, riskbucket.BucketSnapshot{Key: key, LimitMinor: binding.LimitMinor, FilledMinor: binding.FilledMinor,
			HeldMinor: binding.HeldMinor, SnapshotVersion: binding.SnapshotVersion, PolicyProvenance: policyProvenance, SnapshotProvenance: snapshotProvenance})
		references = append(references, journal.RiskBucketSnapshotReference{Key: key, SnapshotID: "snapshot-" + suffix + "-" + string(key.Dimension),
			SnapshotDigest: snapshotEvidence.Digest, SnapshotVersion: binding.SnapshotVersion, PolicyDigest: policyEvidence.Digest,
			ObservedAt: policyEvidence.ObservedAt, FreshUntil: policyEvidence.FreshUntil})
	}
	r.Admission.Admission.Buckets = buckets
	r.Admission.Snapshots = references
}

// a157Precheck 는 준비(setup) 와 결속 결함(binding, nil 이면 없음)을 한 KR 주간 레인 요청에 함께 심고 precheck 거절을 돌려줌.
func a157Precheck(t *testing.T, descriptor strategyflow.Descriptor, setup a157Setup, binding *a157Binding) (*guardianRig, execgw.QFinalCampaignFirstLegIssuance, error) {
	t.Helper()
	suffix := "a157-" + setup.name
	if binding != nil {
		suffix += "-" + binding.name
	}
	rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
		o.NewID = fixedIDs("d-"+suffix, "n-"+suffix)
		if setup.budget != "" {
			o.Policy.RiskBudget = riskcalc.Money{Amount: setup.budget, Currency: "KRW"}
		}
	})
	if setup.rows != nil {
		rig.guardian.SetDailyLossRowsForTest(setup.rows)
	}
	request := qFinalCampaignFirstLegIssuanceFixture(t, rig, descriptor, suffix)
	if setup.entry != nil {
		setup.entry(t, &request, suffix)
	}
	if binding != nil {
		binding.mutate(&request)
	}
	// 측정 토큰은 준비가 끝난 뒤 다시 잼(생산의 첫 다리 승인도 precheck 바로 앞에서 잼).
	request.Entry.DailyLoss = rig.guardian.MeasureDailyLoss(context.Background())
	_, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request)
	return rig, request, err
}

func a157Expect(t *testing.T, row string, err error, want string) {
	t.Helper()
	if got := a157Verdict(err); got != want {
		t.Errorf("%s:\n  got  %s\n  want %s", row, got, want)
	}
}

// 결속 결함 단독 — 진입 판정을 통과하는 요청에서 결속 거절의 코드 · field · 문구 · 원인은 편집 전과 같음(design D3).
func TestA157ABindingDefectAloneIsReportedAsItself(t *testing.T) {
	krWeekly := a157KRWeekly(t)
	for _, binding := range a157Bindings() {
		_, _, err := a157Precheck(t, krWeekly, a157Setup{name: "alone"}, &binding)
		a157Expect(t, binding.name, err, binding.want)
	}
	// B8 의 다른 갈래: 주간이 아닌 레인에 예약 결속이 붙어 옴.
	us := a157US(t)
	_, _, err := a157Precheck(t, us, a157Setup{name: "alone-us", entry: func(t *testing.T, r *execgw.QFinalCampaignFirstLegIssuance, _ string) {
		r.Weekly = &journal.WeeklyFirstLegReservationBinding{ReservationID: "weekly-x"}
	}}, nil)
	a157Expect(t, "B8-binding-on-a-non-weekly-lane", err, a157B8)
}

// 진입 거절 단독 (a)~(e) — 결속 판정을 모두 통과하는 요청에서 진입 거절은 편집 전과 같음.
func TestA157AnEntryRefusalAloneIsReportedAsItself(t *testing.T) {
	krWeekly := a157KRWeekly(t)
	for _, row := range []struct {
		setup a157Setup
		want  string
	}{
		{a157CapZero(), a157CapZeroWant},
		{a157BucketExhausted(), a157BucketWant},
		{a157Limit(), a157DailyLimit},
		{a157Unmeasurable(), a157Unmeasured},
		{a157UnmarkedInput(), a157Unmarked},
	} {
		_, _, err := a157Precheck(t, krWeekly, row.setup, nil)
		a157Expect(t, row.setup.name, err, row.want)
	}
}

// 상한 0 + 결속 결함 — a157 로 뒤집힌 행: 결속 결함이 상한 0 보다 먼저 보고됨(편집 전 HEAD 에서는 상한 0 이 결함을 가렸음 —
// `analysis/red/characterization-head.txt`). 일곱 판정 각각(B5 는 셋)을 따로 둬서 일부만 앞으로 옮긴 구현도 잡음(design D6 M1~M8).
func TestA157CapZeroAndABindingDefect(t *testing.T) {
	krWeekly := a157KRWeekly(t)
	for _, binding := range a157Bindings() {
		_, _, err := a157Precheck(t, krWeekly, a157CapZero(), &binding)
		a157Expect(t, "cap-zero+"+binding.name, err, binding.want)
	}
}

// 버킷 고갈 · 표지 없는 입력 불가 + 결속 결함 — a157 로 뒤집힌 행(design D2 (b) · (e)). 표지 없는 입력 불가는 일일 손실 체인 거절이
// 아니므로(판별 함수가 측정 불가로 보지 않음) 결속 결함이 이김 — 판별의 표지 · 단계 조건을 지운 변이를 이 행이 잡음.
func TestA157OtherEntryRefusalsAndABindingDefect(t *testing.T) {
	krWeekly := a157KRWeekly(t)
	for _, binding := range a157Bindings() {
		if binding.name != "B2-policy-version" && binding.name != "B8-weekly" {
			continue
		}
		_, _, err := a157Precheck(t, krWeekly, a157BucketExhausted(), &binding)
		a157Expect(t, "bucket+"+binding.name, err, binding.want)
		_, _, err = a157Precheck(t, krWeekly, a157UnmarkedInput(), &binding)
		a157Expect(t, "unmarked+"+binding.name, err, binding.want)
	}
}

// 일일 손실(한도 · 측정 불가) + 결속 결함 — 일일 손실 체인 거절이 그대로 보고됨(사람 결정 1-1, 뒤에도 같아야 할 행).
func TestA157DailyLossIsReportedBeforeABindingDefect(t *testing.T) {
	krWeekly := a157KRWeekly(t)
	for _, binding := range a157Bindings() {
		_, _, err := a157Precheck(t, krWeekly, a157Limit(), &binding)
		a157Expect(t, "limit+"+binding.name, err, a157DailyLimit)
		_, _, err = a157Precheck(t, krWeekly, a157Unmeasurable(), &binding)
		a157Expect(t, "unmeasurable+"+binding.name, err, a157Unmeasured)
	}
}

// (f) 입력 하나가 결속 판정과 진입 판정을 함께 거는 경우 — a157 로 뒤집힌 행(freeze 리뷰 Opus P1-1 실측, design D2 (f)): 편집 전에는 진입
// 판정의 ZERO_QUANTITY · UNKNOWN_SYMBOL, 편집 뒤에는 결속 판정 B5 가 보고됨. 엔진 분류는 둘 다 AUTHORITY_MISMATCH 로 같음.
func TestA157OneInputThatTripsBothJudgements(t *testing.T) {
	us := a157US(t)
	_, _, err := a157Precheck(t, us, a157Setup{name: "f-qcandidate-zero", entry: func(_ *testing.T, r *execgw.QFinalCampaignFirstLegIssuance, _ string) {
		r.Entry.QCandidate = 0
	}}, nil)
	a157Expect(t, "f-qcandidate-zero", err, a157B5)
	_, _, err = a157Precheck(t, us, a157Setup{name: "f-symbol", entry: func(_ *testing.T, r *execgw.QFinalCampaignFirstLegIssuance, _ string) {
		r.Entry.Symbol = strings.ToLower(r.Entry.Symbol)
	}}, nil)
	a157Expect(t, "f-noncanonical-symbol", err, a157B5)
}

// 통과 봉인 값 — 결속 · 진입 판정을 모두 통과하는 요청은 고정 시계 아래 정확히 이 봉인을 냄: q_candidate · q_final, 전략 계획(계보 투영 ·
// attempt · 활성화 digest · revision · CreatedAt), 캠페인 요청, 주간 결속(값은 같고 요청과 다른 사본), 발급 시각. 그 뒤 원자 발급이 그 결정을 영속함.
// 계보 투영의 기대값은 명시한 RiskIntent(계좌 · 시장 · 종목 · 매수 · q_final · 가격 · 정책 버전 리터럴)로 따로 투영해 만듦 — 봉인이 쓰는 값을
// precheck 에서 읽지 않음(구현 리뷰 codex P3-2 · Opus P3-4).
func TestA157AnAcceptedFirstLegSealsTheSameValues(t *testing.T) {
	for _, descriptor := range []strategyflow.Descriptor{a157KRWeekly(t), a157US(t)} {
		market := string(descriptor.Market)
		want := a157AcceptedSeal[market]
		suffix := "a157-accept-" + strings.ToLower(market)
		rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) { o.NewID = fixedIDs("d-"+suffix, "n-"+suffix) })
		request := qFinalCampaignFirstLegIssuanceFixture(t, rig, descriptor, suffix)
		precheck, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request)
		if err != nil {
			t.Fatalf("%s: %v", descriptor.LaneID, err)
		}
		if precheck.QCandidate() != 20 || precheck.QFinal() != want.qFinal {
			t.Fatalf("%s: q_candidate=%d q_final=%d — want 20 / %d", descriptor.LaneID, precheck.QCandidate(), precheck.QFinal(), want.qFinal)
		}
		strategy, campaign, weekly, issuedAt := precheck.SealForTest()
		if !issuedAt.Equal(fixedNow) || issuedAt.Location() != time.UTC {
			t.Errorf("%s: issuedAt=%v — want the fixed clock %v in UTC", descriptor.LaneID, issuedAt, fixedNow)
		}
		lineage, err := journal.ProjectAcceptedStrategyflowLineage(journal.StrategyflowLineageProjectionRequest{
			Result: request.Result, ActivationManifestDigest: "sha256:" + strings.Repeat("a", 64), CreatedAt: fixedNow,
			RiskIntent: journal.RiskIntent{AccountRef: "acct-7", Market: market, Symbol: want.symbol, Side: "BUY",
				Quantity: strconv.FormatUint(want.qFinal, 10), EntryPrice: want.entry, StopPrice: want.stop, TargetPrice: want.target,
				PolicyVersion: want.policyVersion},
		})
		if err != nil {
			t.Fatalf("%s: arrangement — the expected lineage does not project: %v", descriptor.LaneID, err)
		}
		wantStrategy := journal.StrategyPlanRequest{Lineage: lineage, AttemptID: "attempt-" + suffix,
			ActivationManifestDigest: "sha256:" + strings.Repeat("a", 64), Revision: 1, CreatedAt: fixedNow}
		if !reflect.DeepEqual(strategy, wantStrategy) {
			t.Errorf("%s: strategy seal\n  got  %+v\n  want %+v", descriptor.LaneID, strategy, wantStrategy)
		}
		wantCampaign := journal.FirstLegCampaignRequest{CampaignID: "campaign-" + suffix, ExpectedPositionGeneration: 0, ExpectedPositionVersion: 0,
			CreateCommandKey: "create-" + suffix, FirstLegCommandKey: "leg-" + suffix, FirstLegPlanID: "plan-" + suffix}
		if campaign != wantCampaign {
			t.Errorf("%s: campaign seal %+v — want %+v", descriptor.LaneID, campaign, wantCampaign)
		}
		switch {
		case descriptor.LaneID == weeklyvaluelane.KRWeeklyLaneID:
			if weekly == nil || request.Weekly == nil || weekly == request.Weekly || *weekly != *request.Weekly ||
				weekly.ReservationID != "weekly-"+suffix || weekly.PlannedOrdinal != 1 || weekly.StableWeek != "KR-XKRX-2026-W14" {
				t.Errorf("%s: weekly seal %+v — want a copy of the reserved binding %+v", descriptor.LaneID, weekly, request.Weekly)
			}
		case weekly != nil:
			t.Errorf("%s: weekly seal %+v — want none on a non-weekly lane", descriptor.LaneID, weekly)
		}
		receipt, err := rig.guardian.IssuePrecheckedQFinalCampaignFirstLeg(context.Background(), precheck)
		if err != nil {
			t.Fatalf("%s: issue: %v", descriptor.LaneID, err)
		}
		if receipt.DecisionID != "d-"+suffix || !decisionOnDisk(t, rig.journal, receipt.DecisionID) {
			t.Fatalf("%s: receipt=%+v — want the sealed decision persisted", descriptor.LaneID, receipt)
		}
	}
}

// 편집 전 HEAD 에서 잰 봉인 값(KR 70000 · 손절 폭 1000 · 예산 기본 → q_final 10 / US 50 USD · FX 1400 → q_final 4) — 정책 버전은
// Guardian 정책 버전 + 거래 ID 결속(journal.QFinalPolicyVersion 의 결과를 리터럴로 고정).
var a157AcceptedSeal = map[string]struct {
	qFinal                                     uint64
	symbol, entry, stop, target, policyVersion string
}{
	"KR": {10, "005930", "70000", "69000", "72000", "add-core-domain/4.1|a066-q-final:v1|qfinal-tx-a157-accept-kr"},
	"US": {4, "AAPL", "50", "45", "60", "add-core-domain/4.1|a066-q-final:v1|qfinal-tx-a157-accept-us"},
}

// nil Guardian — 결속 판정 전에 PrecheckQFinalEntry 의 B1 과 같은 거절(panic 아님).
func TestA157ANilGuardianRefusesBeforeAnyBindingJudgement(t *testing.T) {
	var guardian *execgw.RiskGuardian
	_, err := guardian.PrecheckQFinalCampaignFirstLeg(execgw.QFinalCampaignFirstLegIssuance{})
	a157Expect(t, "nil-guardian", err, "q:RISK_CALCULATION_INVALID/guardian: Guardian is nil")
}

// 판별 함수가 실제 Guardian 의 precheck 거절을 어떻게 읽는지 — 일일 손실 한도 · 측정 불가만 일일 손실 체인 거절이고, 표지 없는 입력 불가 ·
// 상한 0 · 버킷 고갈 · 결속 결함은 아님. 표 시험(a157_daily_loss_kind_test.go)이 합성 모양을, 이 시험이 생산 모양을 맡음.
func TestA157DailyLossRefusalKindOfRealRefusals(t *testing.T) {
	krWeekly := a157KRWeekly(t)
	b2 := a157Bindings()[0]
	for _, row := range []struct {
		setup   a157Setup
		binding *a157Binding
		want    execgw.DailyLossRefusalKind
	}{
		{a157Limit(), nil, execgw.DailyLossRefusalKindLimitReached},
		{a157Unmeasurable(), nil, execgw.DailyLossRefusalKindUnmeasurable},
		{a157UnmarkedInput(), nil, execgw.DailyLossRefusalKindNone},
		{a157CapZero(), nil, execgw.DailyLossRefusalKindNone},
		{a157BucketExhausted(), nil, execgw.DailyLossRefusalKindNone},
		{a157Setup{name: "binding-only"}, &b2, execgw.DailyLossRefusalKindNone},
	} {
		_, _, err := a157Precheck(t, krWeekly, row.setup, row.binding)
		if err == nil {
			t.Fatalf("%s: arrangement — want a refusal", row.setup.name)
		}
		if got := execgw.DailyLossRefusalKindOf(err); got != row.want {
			t.Errorf("%s: DailyLossRefusalKindOf = %v, want %v (err %v)", row.setup.name, got, row.want, err)
		}
	}
}

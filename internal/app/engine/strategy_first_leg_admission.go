package engine

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// StrategyFirstLegAdmissionCode is a stable, non-authorizing outcome from the
// dormant strategy-to-first-leg seam.
type StrategyFirstLegAdmissionCode string

const (
	StrategyFirstLegAdmitted                  StrategyFirstLegAdmissionCode = "ADMITTED"
	StrategyFirstLegResultInvalid             StrategyFirstLegAdmissionCode = "STRATEGY_RESULT_INVALID"
	StrategyFirstLegAuthorityUnavailable      StrategyFirstLegAdmissionCode = "AUTHORITY_UNAVAILABLE"
	StrategyFirstLegAuthorityCollectionFailed StrategyFirstLegAdmissionCode = "AUTHORITY_COLLECTION_FAILED"
	StrategyFirstLegAuthorityMismatch         StrategyFirstLegAdmissionCode = "AUTHORITY_MISMATCH"
	StrategyFirstLegAtomicAdmissionFailed     StrategyFirstLegAdmissionCode = "ATOMIC_ADMISSION_FAILED"
	// StrategyFirstLegDailyLossLimit 은 일일 손실 한도로 계좌의 신규 진입이 막힌 것(a137) — 레인 결함이 아닌 정책 결과.
	StrategyFirstLegDailyLossLimit StrategyFirstLegAdmissionCode = "DAILY_LOSS_LIMIT"
	// StrategyFirstLegDailyLossUnmeasurable 은 그날의 손실을 잴 수 없어 신규 진입이 막힌 것(a137, Manager 결정 ①) — 매 주기 거절은
	// 유지하되 레인 결함이 아님.
	StrategyFirstLegDailyLossUnmeasurable StrategyFirstLegAdmissionCode = "DAILY_LOSS_UNMEASURABLE"
	// StrategyFirstLegGuardianCapZero 는 Guardian 상한이 한 단위도 허용하지 않아 그 진입 하나를 건너뛴 것(a146, 사용자 결정 2026-10-15) —
	// 레인 결함이 아닌 정책 결과이고 다음 주기에 다시 판정됨.
	StrategyFirstLegGuardianCapZero StrategyFirstLegAdmissionCode = "GUARDIAN_CAP_ZERO"
)

// StrategyFirstLegAdmissionResult exposes no reusable authority. Receipt is
// populated only after the existing atomic journal operation succeeds; that
// receipt contains neither the journal-owned prospective token nor a submit
// permit.
type StrategyFirstLegAdmissionResult struct {
	Code    StrategyFirstLegAdmissionCode
	Market  string
	Receipt journal.QFinalCampaignFirstLegReceipt
	Detail  string
	// cause 는 권한 수집이 낸 오류 그대로다(a112 5.2.2.2 J4 · 리뷰 수리). 주문 경로는 이 사슬의 **타입**으로만 「다음 범위로 가도 되는가」를
	// 가르고(범위 거절 타입), 결함이면 원인의 신원(원장 · ctx)을 잃지 않는다 — 문구로 가르지 않는다. Detail 은 같은 오류의 문구다.
	cause error
}

// strategyFirstLegAuthorityLoader has an unexported method deliberately. A
// caller in another package cannot implement it and turn a caller-populated DTO
// into production authority. The real implementation must live beside the
// engine's authoritative account, risk and FX readers.
type strategyFirstLegAuthorityLoader interface {
	collectStrategyFirstLegAuthority(context.Context, strategyFirstLegAccepted) (execgw.QFinalCampaignFirstLegIssuance, error)
}

// strategyFirstLegGuardian is the entire mutation capability available to the
// bridge: one read-only opaque precheck followed by one atomic issue. It cannot
// call raw Journal methods, mint a dispatch lease or reach the Gateway.
type strategyFirstLegGuardian interface {
	PrecheckQFinalCampaignFirstLeg(execgw.QFinalCampaignFirstLegIssuance) (execgw.QFinalCampaignFirstLegPrecheck, error)
	IssuePrecheckedQFinalCampaignFirstLeg(context.Context, execgw.QFinalCampaignFirstLegPrecheck) (journal.QFinalCampaignFirstLegReceipt, error)
	// a137: precheck 직전의 일일 손실 측정 토큰과, 한도 도달의 운영 모드 강화(원장 쓰기는 Guardian 안에서).
	MeasureDailyLoss(context.Context) execgw.DailyLossMeasurement
	EscalateDailyLossLimit(context.Context, error) error
}

var _ strategyFirstLegGuardian = (*execgw.RiskGuardian)(nil)

type strategyFirstLegAdmissionBridge struct {
	guardian strategyFirstLegGuardian
	loader   strategyFirstLegAuthorityLoader
}

func newStrategyFirstLegAdmissionBridge(guardian strategyFirstLegGuardian, loader strategyFirstLegAuthorityLoader) *strategyFirstLegAdmissionBridge {
	return &strategyFirstLegAdmissionBridge{guardian: guardian, loader: loader}
}

// admit remains package-private until the production authority loader exists.
// In particular, there is no exported request constructor that accepts account,
// price, bucket or FX facts from an arbitrary caller.
func (b *strategyFirstLegAdmissionBridge) admit(ctx context.Context, result strategyflow.Result) StrategyFirstLegAdmissionResult {
	accepted, refusal := validateStrategyFirstLegResult(result)
	if refusal.Code != "" {
		return refusal
	}
	if b == nil || b.loader == nil || b.guardian == nil {
		return strategyFirstLegRefusal(StrategyFirstLegAuthorityUnavailable, accepted.market, "production first-leg authority is not wired")
	}
	authority, err := b.loader.collectStrategyFirstLegAuthority(ctx, accepted)
	if err != nil {
		refusal := strategyFirstLegRefusal(StrategyFirstLegAuthorityCollectionFailed, accepted.market, err.Error())
		refusal.cause = err
		return refusal
	}
	if err := validateStrategyFirstLegAuthority(accepted, authority); err != nil {
		return strategyFirstLegRefusal(StrategyFirstLegAuthorityMismatch, accepted.market, err.Error())
	}
	// a137: 일일 손실 측정은 precheck **바로 앞**(토큰의 측정 순간이 precheck 에 가장 가깝게). 실패도 토큰에 봉인되어 daily_loss 단까지 감.
	authority.Entry.DailyLoss = b.guardian.MeasureDailyLoss(ctx)
	precheck, err := b.guardian.PrecheckQFinalCampaignFirstLeg(authority)
	if err != nil {
		// a137: 일일 손실 한도 체인 거절은 타입 · 단계로만 가름(문구 아님) — 강화에 성공하면 계좌 진입 차단(레인 결함 아님),
		// 실패하면 거절과 강화 오류를 함께 든 결함으로 남음(조용히 삼켜지지 않게).
		// 측정 불가(측정기 · 토큰에서만 생기는 타입 표지)는 강화 없이 계좌 단위 진입 차단 결과 — 그 밖의 입력 불가는 오늘처럼 결함.
		if unmeasurable := (*execgw.DailyLossUnmeasurable)(nil); errors.As(err, &unmeasurable) {
			if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&
				issue.Reason == string(risk.ReasonInputUnavailable) {
				refusal := strategyFirstLegRefusal(StrategyFirstLegDailyLossUnmeasurable, accepted.market, err.Error())
				refusal.cause = err
				return refusal
			}
		}
		if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&
			issue.Reason == string(risk.ReasonDailyLossLimitReached) {
			refusal := strategyFirstLegRefusal(StrategyFirstLegDailyLossLimit, accepted.market, err.Error())
			refusal.cause = b.guardian.EscalateDailyLossLimit(ctx, err)
			return refusal
		}
		// a146(사용자 결정 2026-10-15): Guardian 상한 0 — 한 단위도 못 사는 진입은 레인 결함이 아니라 그 진입 하나의 건너뜀. 거절 타입 · 코드 ·
		// 그 거절 **자신의** 원인 sentinel 셋이 모두 맞을 때만(문구 아님). 같은 코드의 해석 실패 · 원인 없는 0 · FX/정책 오류, 그리고 sentinel 이
		// 거절 곁에만 있는 모양은 아래 AUTHORITY_MISMATCH 그대로(결함). 흔적은 주기 ctx 의 기록자가 남김(계수 · 하루 한 줄).
		if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalExistingGuardianCap &&
			errors.Is(qFinal, risk.ErrStrategyQuantityZero) {
			// 가격은 validateStrategyFirstLegAuthority 가 이미 해석해 결속한 봉인 조건의 값(여기서는 표시용).
			entryMajor, _ := accepted.result.ExecutionTerms.Entry().MajorDecimal()
			stopMajor, _ := accepted.result.ExecutionTerms.EffectiveStop().MajorDecimal()
			skip := &strategyEntrySkip{market: accepted.market, symbol: accepted.result.Lineage.Symbol, entry: entryMajor, stop: stopMajor, cause: err}
			refusal := strategyFirstLegRefusal(StrategyFirstLegGuardianCapZero, accepted.market, err.Error())
			refusal.cause = skip
			recordStrategyEntrySkip(ctx, skip)
			return refusal
		}
		refusal := strategyFirstLegRefusal(StrategyFirstLegAuthorityMismatch, accepted.market, err.Error())
		// a137 F8: 원인을 실음 — 하류가 errors.As 로 거절 타입을 볼 수 있게(문구 · 분류는 그대로).
		refusal.cause = err
		// a112 6.2(Manager 판정 (A)): 그 범위의 버킷 고갈(q_final 0)은 결함이 아니라 그 범위의 정책 결과다 — **타입 · 코드**로만 범위 거절로
		// 싣는다(문구 아님). 그래야 한 family 버킷이 고갈돼도 같은 주기의 다른 범위가 계속 평가된다(스펙 「한 family risk bucket 고갈」).
		// 그 밖의 precheck 거절과, 발급 단계(원장 트랜잭션)의 거절 — 같은 파도 둘째 범위의 BUCKET_USAGE_STALE 등 CAS — 은 결함으로 남는다
		// (설계 ④ 보호). 진입 관문 관측 거절(R3)은 이 판정이 넓히지 않는다.
		if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted {
			if key, keyed := strategyOwnerKeyOf(authority.Result.Lineage); keyed {
				refusal.cause = &strategyScopeRefusal{scope: key, detail: "this owner scope's risk bucket is exhausted", cause: err}
			}
		}
		return refusal
	}
	receipt, err := b.guardian.IssuePrecheckedQFinalCampaignFirstLeg(ctx, precheck)
	if err != nil {
		return strategyFirstLegRefusal(StrategyFirstLegAtomicAdmissionFailed, accepted.market, err.Error())
	}
	return StrategyFirstLegAdmissionResult{Code: StrategyFirstLegAdmitted, Market: string(accepted.market), Receipt: receipt}
}

type strategyFirstLegAccepted struct {
	result   strategyflow.Result
	market   strategyrouter.Market
	currency string
	scale    int
}

func validateStrategyFirstLegResult(result strategyflow.Result) (strategyFirstLegAccepted, StrategyFirstLegAdmissionResult) {
	lineage := result.Lineage
	market := lineage.Market
	accepted := strategyFirstLegAccepted{result: result, market: market}
	switch market {
	case strategyrouter.MarketKR:
		accepted.currency, accepted.scale = "KRW", 0
	case strategyrouter.MarketUS:
		accepted.currency, accepted.scale = "USD", 2
	default:
		return strategyFirstLegAccepted{}, strategyFirstLegRefusal(StrategyFirstLegResultInvalid, market, "only KR and US are valid first-leg markets")
	}
	terms := result.ExecutionTerms
	invalid := result.Code != strategyflow.RefusalNone || result.Quantity == 0 || !lineage.Complete || !lineage.Valid() || !terms.Valid() ||
		lineage.AccountRef == "" || lineage.Symbol == "" || lineage.Symbol != strings.ToUpper(strings.TrimSpace(lineage.Symbol)) ||
		lineage.CandidateLifeID == "" || lineage.ThresholdVersion == "" || lineage.ThresholdSetDigest == "" ||
		lineage.CandidateEvidenceDigest == "" || lineage.RouterEvidenceDigest == "" || lineage.LaneEvidenceDigest == "" ||
		lineage.RouterID != strategyrouter.RouterID || lineage.RouterRelease != strategyrouter.RouterRelease ||
		lineage.LaneID == "" || lineage.LaneVersion == "" ||
		lineage.CampaignID == "" || lineage.LegOrdinal != 1 || lineage.PlannedCeiling < result.Quantity ||
		lineage.RiskBudgetDigest == "" || lineage.ExecutionPolicyDigest == "" || lineage.PositionGeneration > math.MaxInt64 ||
		terms.AccountRef() != lineage.AccountRef || terms.Market() != lineage.Market || terms.Symbol() != lineage.Symbol ||
		terms.CampaignID() != lineage.CampaignID || terms.LegOrdinal() != 1 || terms.Quantity() != result.Quantity ||
		terms.LineageIdentity() != lineage.Identity || terms.Policy().Identity() != lineage.ExecutionPolicyDigest ||
		result.GuardianCalls != 0 || result.BrokerCalls != 0 || result.Mutations != 0
	if invalid {
		return strategyFirstLegAccepted{}, strategyFirstLegRefusal(StrategyFirstLegResultInvalid, market, "strategy result is not an exact accepted first-leg value")
	}
	for _, price := range []strategyflow.PriceProvenance{terms.Entry(), terms.EffectiveStop(), terms.Target()} {
		if price.Currency() != accepted.currency || price.MinorScale() != accepted.scale || price.UnitVersion() != "minor-v1" {
			return strategyFirstLegAccepted{}, strategyFirstLegRefusal(StrategyFirstLegResultInvalid, market, "execution terms do not use the canonical market minor unit")
		}
	}
	return accepted, StrategyFirstLegAdmissionResult{}
}

func validateStrategyFirstLegAuthority(accepted strategyFirstLegAccepted, request execgw.QFinalCampaignFirstLegIssuance) error {
	authorityAccepted, refusal := validateStrategyFirstLegResult(request.Result)
	if refusal.Code != "" || authorityAccepted.market != accepted.market ||
		request.Result.Lineage.Identity != accepted.result.Lineage.Identity ||
		request.Result.ExecutionTerms.Identity() != accepted.result.ExecutionTerms.Identity() ||
		request.Result.Quantity != accepted.result.Quantity {
		return fmt.Errorf("sealed strategy result does not match the accepted result")
	}
	lineage, terms := accepted.result.Lineage, accepted.result.ExecutionTerms
	entryMajor, entryOK := terms.Entry().MajorDecimal()
	stopMajor, stopOK := terms.EffectiveStop().MajorDecimal()
	targetMajor, targetOK := terms.Target().MajorDecimal()
	if !entryOK || !stopOK || !targetOK {
		return fmt.Errorf("sealed strategy execution price unit is invalid")
	}
	entry, owner := request.Entry, request.Entry.Admission.Owner
	if entry.Market != string(lineage.Market) || entry.Currency != accepted.currency || entry.Symbol != lineage.Symbol ||
		entry.QCandidate != accepted.result.Quantity || entry.EntryPrice != entryMajor ||
		entry.StopPrice != stopMajor || entry.TargetPrice != targetMajor ||
		entry.Admission.TransactionID == "" || owner.Key.AccountID != lineage.AccountRef ||
		string(owner.Key.Market) != string(lineage.Market) || owner.Key.Symbol != lineage.Symbol ||
		owner.Key.ProspectiveGeneration != "" || owner.LaneID != lineage.LaneID || owner.CampaignID != lineage.CampaignID {
		return fmt.Errorf("Guardian entry authority does not exactly bind the accepted result")
	}
	if request.Campaign.CampaignID != lineage.CampaignID || request.Campaign.ExpectedPositionGeneration < 0 ||
		request.Campaign.ExpectedPositionGeneration >= math.MaxInt64 || request.Campaign.ExpectedPositionVersion < 0 ||
		lineage.PositionGeneration != uint64(request.Campaign.ExpectedPositionGeneration)+1 ||
		request.Campaign.CreateCommandKey == "" || request.Campaign.FirstLegCommandKey == "" || request.Campaign.FirstLegPlanID == "" {
		return fmt.Errorf("campaign current CAS and prospective strategy generation do not exactly bind")
	}
	if request.AttemptID == "" || request.AttemptID != strings.TrimSpace(request.AttemptID) || request.Revision < 1 ||
		request.ActivationManifestDigest == "" || request.ActivationManifestDigest != strings.TrimSpace(request.ActivationManifestDigest) {
		return fmt.Errorf("strategy activation and attempt metadata are incomplete")
	}
	return nil
}

func strategyFirstLegRefusal(code StrategyFirstLegAdmissionCode, market strategyrouter.Market, detail string) StrategyFirstLegAdmissionResult {
	return StrategyFirstLegAdmissionResult{Code: code, Market: string(market), Detail: detail}
}

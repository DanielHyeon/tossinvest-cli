package engine

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
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
		return strategyFirstLegRefusal(StrategyFirstLegAuthorityCollectionFailed, accepted.market, err.Error())
	}
	if err := validateStrategyFirstLegAuthority(accepted, authority); err != nil {
		return strategyFirstLegRefusal(StrategyFirstLegAuthorityMismatch, accepted.market, err.Error())
	}
	precheck, err := b.guardian.PrecheckQFinalCampaignFirstLeg(authority)
	if err != nil {
		return strategyFirstLegRefusal(StrategyFirstLegAuthorityMismatch, accepted.market, err.Error())
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

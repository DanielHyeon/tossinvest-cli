package continuationlane

import (
	"encoding/hex"
	"errors"
	"time"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
)

// CaptureOfflineEvidence는 호출자가 동시 변경을 차단한 스냅샷을 조회 전용으로 복사한다.
func CaptureOfflineEvidence(p CampaignPlan, r RiskState, at time.Time) (evidence.Evidence, error) {
	if p.PositionGeneration <= 0 || !validatePlan(p) || !validRiskStateScope(r, p) || at.IsZero() {
		return evidence.Evidence{}, errors.New("invalid plan or risk binding")
	}
	var fx *evidence.FX
	if p.FX != nil {
		f := p.FX
		official, frozen := f.official, f.frozen
		fx = &evidence.FX{QuoteID: f.QuoteID, AsOf: f.AsOf, Direction: f.Direction, Rate: f.RateQuoteToAccount, Haircut: f.Haircut, Digest: f.Digest, Official: &official, Frozen: &frozen, Seal: hex.EncodeToString(f.seal[:])}
	}
	plan := evidence.Plan{Scope: evidence.Scope{Account: p.AccountRef, Market: string(p.Market), Symbol: p.Symbol, Campaign: p.CampaignID, Lane: p.LaneID, Version: p.LaneVersion, Generation: uint64(p.PositionGeneration)}, Budget: p.RiskBudgetMinor, PerShareRisk: p.PerShareRiskMinor, Quantity: p.PlannedQuantity, Ceilings: append([]uint64(nil), p.LegCeilings[:]...), Weights: append([]uint64(nil), p.Weights[:]...), PolicyDigest: p.PolicyDigest, ConfigDigest: p.ConfigDigest, AccountCurrency: p.AccountCurrency, QuoteCurrency: p.QuoteCurrency, FX: fx, Digest: p.RiskBudgetDigest}
	risk := evidence.Risk{PlanDigest: r.RiskBudgetDigest, Budget: r.BudgetMinor, Filled: r.FilledMinor, Held: r.HeldMinor, CancelStatus: "captured", ReservationStatus: "not_applicable"}
	if r.Latches != nil {
		risk.Latches = make(map[string]bool, len(r.Latches))
		for k, v := range r.Latches {
			risk.Latches[string(k)] = v
		}
	}
	if r.Fills != nil {
		risk.ContinuationFills = make(map[string]evidence.ContinuationFill, len(r.Fills))
		for k, v := range r.Fills {
			risk.ContinuationFills[k] = evidence.ContinuationFill{Applied: v.Applied, Risk: v.RiskMinor, Known: v.Known, Fingerprint: v.Fingerprint}
		}
	}
	if r.Cancels != nil {
		risk.ContinuationCancels = make(map[string]evidence.ContinuationCancel, len(r.Cancels))
		for k, v := range r.Cancels {
			risk.ContinuationCancels[k] = evidence.ContinuationCancel{Applied: v.Applied, Fingerprint: v.Fingerprint}
		}
	}
	out := evidence.Evidence{Schema: "campaign-state-evidence:v1", Family: "continuation", CapturedAt: at.Format(time.RFC3339Nano), Purpose: "inspection-only", Origin: "caller-provided-snapshot", Currentness: "ledger-currentness-unverified", Plan: plan, Risk: risk}
	if _, err := evidence.Encode(out); err != nil {
		return evidence.Evidence{}, err
	}
	return out, nil
}

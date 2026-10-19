package reversallane

import (
	"errors"
	"time"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
)

// CaptureOfflineEvidence는 호출자가 동시 변경을 차단한 스냅샷을 조회 전용으로 복사한다.
func CaptureOfflineEvidence(p CampaignPlan, r RiskState, at time.Time) (evidence.Evidence, error) {
	if !p.valid() || r.PlanDigest != p.Digest() || at.IsZero() {
		return evidence.Evidence{}, errors.New("invalid plan or risk binding")
	}
	req := p.request
	var fx *evidence.FX
	if req.FX != nil {
		f := req.FX
		official, frozen := f.Official, f.Frozen
		fx = &evidence.FX{Authority: f.Authority, Version: f.Version, QuoteID: f.QuoteID, AsOf: f.AsOf, FreshUntil: f.FreshUntil, Direction: string(f.Direction), Rate: f.RateQuoteToAccount, Haircut: f.Haircut, Digest: f.Digest, Official: &official, Frozen: &frozen, Seal: f.seal}
	}
	plan := evidence.Plan{Scope: evidence.Scope{Account: req.AccountRef, Market: string(req.Market), Symbol: req.Symbol, Campaign: req.CampaignID, Lane: req.LaneID, Version: req.LaneVersion, Generation: req.PositionGeneration}, Budget: req.RiskBudgetMinor, PerShareRisk: req.PerShareRiskMinor, Quantity: req.PlannedQuantity, Ceilings: append([]uint64(nil), p.legCeilings[:]...), Weights: []uint64{2, 4, 8}, PolicyDigest: req.PolicyDigest, ConfigDigest: req.ConfigDigest, AccountCurrency: req.AccountCurrency, QuoteCurrency: req.QuoteCurrency, FX: fx, Digest: p.digest}
	risk := evidence.Risk{PlanDigest: r.PlanDigest, Filled: r.FilledMinor, Held: r.HeldMinor, CancelStatus: "captured", ReservationStatus: "not_applicable"}
	if r.Latches != nil {
		risk.Latches = make(map[string]bool, len(r.Latches))
		for k, v := range r.Latches {
			risk.Latches[string(k)] = v
		}
	}
	if r.Fills != nil {
		risk.ReversalFills = make(map[string]evidence.ReversalFill, len(r.Fills))
		for k, v := range r.Fills {
			risk.ReversalFills[k] = evidence.ReversalFill{Applied: v.Applied, Risk: v.RiskMinor, Fingerprint: v.Fingerprint}
		}
	}
	if r.Cancels != nil {
		risk.ReversalCancels = make(map[string]string, len(r.Cancels))
		for k, v := range r.Cancels {
			risk.ReversalCancels[k] = v
		}
	}
	out := evidence.Evidence{Schema: "campaign-state-evidence:v1", Family: "reversal", CapturedAt: at.Format(time.RFC3339Nano), Purpose: "inspection-only", Origin: "caller-provided-snapshot", Currentness: "ledger-currentness-unverified", Plan: plan, Risk: risk}
	if _, err := evidence.Encode(out); err != nil {
		return evidence.Evidence{}, err
	}
	return out, nil
}

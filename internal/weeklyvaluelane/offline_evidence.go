package weeklyvaluelane

import (
	"encoding/hex"
	"errors"
	"time"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
)

// CaptureOfflineEvidence는 호출자가 동시 변경을 차단한 스냅샷을 조회 전용으로 복사한다.
func CaptureOfflineEvidence(p CampaignPlan, r RiskState, at time.Time) (evidence.Evidence, error) {
	if !p.valid() || !validRiskState(r) || r.planDigest != p.digest || at.IsZero() {
		return evidence.Evidence{}, errors.New("invalid plan or risk binding")
	}
	var fx *evidence.FX
	if p.fx != nil {
		f := p.fx
		fx = &evidence.FX{Authority: f.authority, Version: f.version, QuoteID: f.quoteID, AsOf: f.asOf.Format(time.RFC3339Nano), FreshUntil: f.freshUntil.Format(time.RFC3339Nano), Direction: f.direction, Rate: f.rateQuoteToAccount, Haircut: f.haircut, Digest: f.digest, Seal: hex.EncodeToString(f.seal[:])}
	}
	plan := evidence.Plan{Scope: evidence.Scope{Account: p.accountRef, Market: string(p.market), Symbol: p.symbol, Campaign: p.campaignID, Lane: p.laneID, Version: p.laneVersion, Generation: p.positionGeneration}, Budget: p.riskBudgetMinor, PerShareRisk: p.perShareRiskMinor, Quantity: p.plannedQuantity, Ceilings: append([]uint64(nil), p.legCeilings[:]...), PolicyDigest: p.policyDigest, ConfigDigest: p.configDigest, AccountCurrency: p.accountCurrency, QuoteCurrency: p.quoteCurrency, FX: fx, Digest: p.digest, Seal: hex.EncodeToString(p.seal[:])}
	risk := evidence.Risk{PlanDigest: r.planDigest, Filled: r.filledMinor, Held: r.heldMinor, Seal: hex.EncodeToString(r.seal[:]), CancelStatus: "not_applicable", ReservationStatus: "not_captured"}
	if r.latches != nil {
		risk.Latches = make(map[string]bool, len(r.latches))
		for k, v := range r.latches {
			risk.Latches[string(k)] = v
		}
	}
	if r.fills != nil {
		risk.WeeklyFills = make(map[string]evidence.WeeklyFill, len(r.fills))
		for k, v := range r.fills {
			risk.WeeklyFills[k] = evidence.WeeklyFill{Fingerprint: v.Fingerprint, KnownIdentity: v.Known}
		}
	}
	out := evidence.Evidence{Schema: "campaign-state-evidence:v1", Family: "weekly", CapturedAt: at.Format(time.RFC3339Nano), Purpose: "inspection-only", Origin: "caller-provided-snapshot", Currentness: "ledger-currentness-unverified", Plan: plan, Risk: risk}
	if _, err := evidence.Encode(out); err != nil {
		return evidence.Evidence{}, err
	}
	return out, nil
}

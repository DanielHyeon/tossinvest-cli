package reversallane

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type PlanRequest struct {
	LaneID             string
	LaneVersion        string
	Market             Market
	AccountRef         string
	Symbol             string
	CampaignID         string
	PositionGeneration uint64
	RiskBudgetMinor    string
	PerShareRiskMinor  string
	PlannedQuantity    uint64
	PolicyDigest       string
	ConfigDigest       string
	AccountCurrency    string
	QuoteCurrency      string
	FX                 *FrozenFX
}

type CampaignPlan struct {
	request     PlanRequest
	legCeilings [3]uint64
	digest      string
}

func BuildCampaignPlan(request PlanRequest) (CampaignPlan, error) {
	request.LaneID = strings.TrimSpace(request.LaneID)
	request.LaneVersion = strings.TrimSpace(request.LaneVersion)
	request.AccountRef = strings.TrimSpace(request.AccountRef)
	request.Symbol = strings.ToUpper(strings.TrimSpace(request.Symbol))
	request.CampaignID = strings.TrimSpace(request.CampaignID)
	request.PolicyDigest = strings.TrimSpace(request.PolicyDigest)
	request.ConfigDigest = strings.TrimSpace(request.ConfigDigest)
	request.AccountCurrency = strings.ToUpper(strings.TrimSpace(request.AccountCurrency))
	request.QuoteCurrency = strings.ToUpper(strings.TrimSpace(request.QuoteCurrency))
	for _, value := range []string{request.LaneID, request.LaneVersion, request.AccountRef, request.Symbol, request.CampaignID, request.PolicyDigest, request.ConfigDigest, request.AccountCurrency, request.QuoteCurrency} {
		if len(value) > maxIdentityBytes {
			return CampaignPlan{}, fmt.Errorf("%s: identity exceeds bound", RefusalPlanInvalid)
		}
	}
	if request.LaneVersion != LaneVersionV1 || request.AccountRef == "" || request.Symbol == "" || request.CampaignID == "" || request.PositionGeneration == 0 || request.PolicyDigest == "" || request.ConfigDigest == "" || request.AccountCurrency == "" || request.QuoteCurrency == "" {
		return CampaignPlan{}, fmt.Errorf("%s: incomplete plan identity", RefusalPlanInvalid)
	}
	if (request.Market == MarketKR && request.LaneID != KRReversalLaneID) || (request.Market == MarketUS && request.LaneID != USReversalLaneID) || (request.Market != MarketKR && request.Market != MarketUS) {
		return CampaignPlan{}, fmt.Errorf("%s: market/lane mismatch", RefusalPlanInvalid)
	}
	if budget, ok := parseMinor(request.RiskBudgetMinor); !ok || budget.Sign() <= 0 || request.PlannedQuantity == 0 {
		return CampaignPlan{}, fmt.Errorf("%s: risk budget", RefusalPlanInvalid)
	}
	if perShare, ok := parseMinor(request.PerShareRiskMinor); !ok || perShare.Sign() <= 0 {
		return CampaignPlan{}, fmt.Errorf("%s: per-share risk", RefusalPlanInvalid)
	}
	if request.AccountCurrency != request.QuoteCurrency {
		if request.FX == nil || !validFrozenFX(*request.FX) {
			return CampaignPlan{}, fmt.Errorf("%s: frozen FX", RefusalPlanInvalid)
		}
		copyFX := *request.FX
		request.FX = &copyFX
	} else if request.FX != nil {
		return CampaignPlan{}, fmt.Errorf("%s: same-currency FX must be absent", RefusalPlanInvalid)
	}
	plan := CampaignPlan{request: request, legCeilings: AllocateTwoFourEight(request.PlannedQuantity)}
	plan.digest = planDigest(plan)
	return plan, nil
}

func AllocateTwoFourEight(quantity uint64) [3]uint64 {
	quotient, remainder := quantity/14, quantity%14
	first := quotient*2 + remainder*2/14
	second := quotient*4 + remainder*4/14
	return [3]uint64{first, second, quantity - first - second}
}

func (p CampaignPlan) valid() bool                { return p.digest != "" && p.digest == planDigest(p) }
func (p CampaignPlan) Digest() string             { return p.digest }
func (p CampaignPlan) Market() Market             { return p.request.Market }
func (p CampaignPlan) LaneID() string             { return p.request.LaneID }
func (p CampaignPlan) CampaignID() string         { return p.request.CampaignID }
func (p CampaignPlan) PositionGeneration() uint64 { return p.request.PositionGeneration }
func (p CampaignPlan) PlannedQuantity() uint64    { return p.request.PlannedQuantity }
func (p CampaignPlan) LegCeilings() [3]uint64     { return p.legCeilings }
func (p CampaignPlan) ConfigDigest() string       { return p.request.ConfigDigest }

func planDigest(plan CampaignPlan) string {
	r := plan.request
	fx := ""
	if r.FX != nil {
		fx = fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%t\x00%t\x00%s\x00%s", r.FX.Authority, r.FX.Version, r.FX.QuoteID, r.FX.AsOf,
			r.FX.Direction, r.FX.RateQuoteToAccount, r.FX.Haircut, r.FX.Digest, r.FX.Official, r.FX.Frozen, r.FX.FreshUntil, r.FX.seal)
	}
	preimage := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%d\x00%s\x00%s\x00%d\x00%s\x00%s\x00%s\x00%s\x00%d\x00%d\x00%d\x00%s",
		r.LaneID, r.LaneVersion, r.Market, r.AccountRef, r.Symbol, r.CampaignID, r.PositionGeneration, r.RiskBudgetMinor, r.PerShareRiskMinor, r.PlannedQuantity,
		r.PolicyDigest, r.ConfigDigest, r.AccountCurrency, r.QuoteCurrency, plan.legCeilings[0], plan.legCeilings[1], plan.legCeilings[2], fx)
	sum := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(sum[:])
}

func PlannedLegQuantity(plan CampaignPlan, progress LegProgress, cap RiskCap) uint64 {
	if !plan.valid() || progress.Ordinal < 1 || progress.Ordinal > 3 || progress.Cancelled || progress.Expired {
		return 0
	}
	ceiling := plan.legCeilings[progress.Ordinal-1]
	if progress.FilledQuantity >= ceiling {
		return 0
	}
	remaining := ceiling - progress.FilledQuantity
	if cap.QFinal < remaining {
		return cap.QFinal
	}
	return remaining
}

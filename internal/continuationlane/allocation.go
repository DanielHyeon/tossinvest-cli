package continuationlane

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const FXQuoteToAccount = "QUOTE_TO_ACCOUNT"

type FrozenFX struct {
	QuoteID            string
	AsOf               string
	Direction          string
	RateQuoteToAccount string
	Haircut            string
	Digest             string
	official           bool
	frozen             bool
	seal               [32]byte
}

type frozenFXInput struct {
	QuoteID            string
	AsOf               string
	Direction          string
	RateQuoteToAccount string
	Haircut            string
	Digest             string
}

func newFrozenFX(input frozenFXInput) (FrozenFX, error) {
	fx := FrozenFX{QuoteID: strings.TrimSpace(input.QuoteID), AsOf: input.AsOf, Direction: input.Direction,
		RateQuoteToAccount: input.RateQuoteToAccount, Haircut: input.Haircut, Digest: strings.TrimSpace(input.Digest), official: true, frozen: true}
	if !validFXFields(fx) {
		return FrozenFX{}, fmt.Errorf("continuation lanes: invalid official frozen FX")
	}
	fx.seal = frozenFXSeal(fx)
	return fx, nil
}

type PlanRequest struct {
	LaneID             string
	LaneVersion        string
	Market             Market
	AccountRef         string
	Symbol             string
	CampaignID         string
	PositionGeneration int64
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
	LaneID             string
	LaneVersion        string
	Market             Market
	AccountRef         string
	Symbol             string
	CampaignID         string
	PositionGeneration int64
	RiskBudgetMinor    string
	PerShareRiskMinor  string
	PlannedQuantity    uint64
	Weights            [3]uint64
	LegCeilings        [3]uint64
	PolicyDigest       string
	ConfigDigest       string
	AccountCurrency    string
	QuoteCurrency      string
	FX                 *FrozenFX
	RiskBudgetDigest   string
}

func AllocateEightFourTwo(quantity uint64) [3]uint64 {
	q := new(big.Int).SetUint64(quantity)
	first := new(big.Int).Quo(new(big.Int).Mul(new(big.Int).Set(q), big.NewInt(8)), big.NewInt(14)).Uint64()
	second := new(big.Int).Quo(new(big.Int).Mul(new(big.Int).Set(q), big.NewInt(4)), big.NewInt(14)).Uint64()
	return [3]uint64{first, second, quantity - first - second}
}

func BuildCampaignPlan(request PlanRequest) (CampaignPlan, error) {
	descriptor, ok := descriptorFor(request.Market)
	if !ok || request.LaneID != descriptor.ID || request.LaneVersion != descriptor.Version {
		return CampaignPlan{}, fmt.Errorf("continuation lanes: lane/market identity mismatch")
	}
	if strings.TrimSpace(request.AccountRef) == "" || strings.TrimSpace(request.Symbol) == "" || strings.TrimSpace(request.CampaignID) == "" ||
		request.PositionGeneration <= 0 || strings.TrimSpace(request.PolicyDigest) == "" || strings.TrimSpace(request.ConfigDigest) == "" ||
		request.PlannedQuantity == 0 || !canonicalCurrency(request.AccountCurrency) || !canonicalCurrency(request.QuoteCurrency) {
		return CampaignPlan{}, fmt.Errorf("continuation lanes: incomplete immutable plan identity")
	}
	budget, err := parseUnsigned(request.RiskBudgetMinor)
	if err != nil || budget.Sign() <= 0 {
		return CampaignPlan{}, fmt.Errorf("continuation lanes: invalid risk budget: %w", err)
	}
	perShare, err := parseUnsigned(request.PerShareRiskMinor)
	if err != nil || perShare.Sign() <= 0 {
		return CampaignPlan{}, fmt.Errorf("continuation lanes: invalid per-share risk: %w", err)
	}
	if request.AccountCurrency == request.QuoteCurrency && request.FX != nil {
		return CampaignPlan{}, fmt.Errorf("continuation lanes: same-currency plan must not carry FX")
	}
	if request.AccountCurrency != request.QuoteCurrency {
		if request.FX == nil || !validFX(*request.FX) {
			return CampaignPlan{}, fmt.Errorf("continuation lanes: frozen FX is required")
		}
	}
	plan := CampaignPlan{
		LaneID: request.LaneID, LaneVersion: request.LaneVersion, Market: request.Market,
		AccountRef: strings.TrimSpace(request.AccountRef), Symbol: strings.TrimSpace(request.Symbol), CampaignID: strings.TrimSpace(request.CampaignID),
		PositionGeneration: request.PositionGeneration, RiskBudgetMinor: budget.String(), PerShareRiskMinor: perShare.String(),
		PlannedQuantity: request.PlannedQuantity, Weights: [3]uint64{8, 4, 2}, LegCeilings: AllocateEightFourTwo(request.PlannedQuantity),
		PolicyDigest: request.PolicyDigest, ConfigDigest: request.ConfigDigest, AccountCurrency: request.AccountCurrency, QuoteCurrency: request.QuoteCurrency,
	}
	if request.FX != nil {
		copyFX := *request.FX
		plan.FX = &copyFX
	}
	plan.RiskBudgetDigest = planDigest(plan)
	return plan, nil
}

func validatePlan(plan CampaignPlan) bool {
	if plan.Weights != [3]uint64{8, 4, 2} || plan.LegCeilings != AllocateEightFourTwo(plan.PlannedQuantity) || plan.RiskBudgetDigest == "" {
		return false
	}
	copyPlan := plan
	copyPlan.RiskBudgetDigest = ""
	return plan.RiskBudgetDigest == planDigest(copyPlan)
}

func planDigest(plan CampaignPlan) string {
	h := sha256.New()
	parts := []string{plan.LaneID, plan.LaneVersion, string(plan.Market), plan.AccountRef, plan.Symbol, plan.CampaignID,
		strconv.FormatInt(plan.PositionGeneration, 10), plan.RiskBudgetMinor, plan.PerShareRiskMinor, canonicalUint(plan.PlannedQuantity),
		canonicalUint(plan.LegCeilings[0]), canonicalUint(plan.LegCeilings[1]), canonicalUint(plan.LegCeilings[2]),
		plan.PolicyDigest, plan.ConfigDigest, plan.AccountCurrency, plan.QuoteCurrency}
	if plan.FX != nil {
		parts = append(parts, plan.FX.QuoteID, plan.FX.AsOf, plan.FX.Direction, plan.FX.RateQuoteToAccount, plan.FX.Haircut, plan.FX.Digest,
			strconv.FormatBool(plan.FX.official), strconv.FormatBool(plan.FX.frozen), hex.EncodeToString(plan.FX.seal[:]))
	}
	for _, part := range parts {
		h.Write([]byte(strconv.Itoa(len(part))))
		h.Write([]byte{':'})
		h.Write([]byte(part))
	}
	return hex.EncodeToString(h.Sum(nil))
}

type LegProgress struct {
	Ordinal        int
	FilledQuantity uint64
	Cancelled      bool
	Expired        bool
}

func PlannedLegQuantity(plan CampaignPlan, leg LegProgress, qFinal uint64) uint64 {
	if leg.Ordinal < 1 || leg.Ordinal > 3 || leg.Cancelled || leg.Expired {
		return 0
	}
	ceiling := plan.LegCeilings[leg.Ordinal-1]
	if leg.FilledQuantity >= ceiling {
		return 0
	}
	remaining := ceiling - leg.FilledQuantity
	if qFinal < remaining {
		return qFinal
	}
	return remaining
}

func validFX(fx FrozenFX) bool {
	return validFXFields(fx) && fx.seal == frozenFXSeal(fx)
}

func validFXFields(fx FrozenFX) bool {
	if strings.TrimSpace(fx.QuoteID) == "" || strings.TrimSpace(fx.AsOf) == "" || fx.Direction != FXQuoteToAccount ||
		strings.TrimSpace(fx.Digest) == "" || !fx.official || !fx.frozen {
		return false
	}
	if _, err := time.Parse(time.RFC3339Nano, fx.AsOf); err != nil {
		return false
	}
	rate, err := parsePositiveDecimal(fx.RateQuoteToAccount)
	if err != nil || rate.Sign() <= 0 {
		return false
	}
	haircut, err := parsePositiveDecimal(fx.Haircut)
	return err == nil && haircut.Cmp(big.NewRat(1, 1)) >= 0
}

func frozenFXSeal(fx FrozenFX) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{fx.QuoteID, fx.AsOf, fx.Direction, fx.RateQuoteToAccount, fx.Haircut, fx.Digest,
		strconv.FormatBool(fx.official), strconv.FormatBool(fx.frozen)}, "\x00")))
}

func canonicalCurrency(currency string) bool {
	if len(currency) != 3 || currency != strings.ToUpper(currency) {
		return false
	}
	for _, r := range currency {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

package weeklyvaluelane

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const (
	a066FXAuthority      = "a066-frozen-fx-authority"
	a066RiskCapAuthority = "a066-risk-cap-authority"
	FXQuoteToAccount     = "QUOTE_TO_ACCOUNT"
)

type PlanRequest struct {
	LaneID, LaneVersion, AccountRef, Symbol, CampaignID string
	Market                                              Market
	PositionGeneration                                  uint64
	RiskBudgetMinor, PerShareRiskMinor                  string
	PlannedQuantity                                     uint64
	PolicyDigest, ConfigDigest                          string
	AccountCurrency, QuoteCurrency                      string
	FX                                                  *FrozenFX
}

type CampaignPlan struct {
	laneID, laneVersion, accountRef, symbol, campaignID string
	market                                              Market
	positionGeneration                                  uint64
	riskBudgetMinor, perShareRiskMinor                  string
	plannedQuantity                                     uint64
	legCeilings                                         [7]uint64
	policyDigest, configDigest                          string
	accountCurrency, quoteCurrency                      string
	fx                                                  *FrozenFX
	digest                                              string
	seal                                                [32]byte
}

func BuildCampaignPlan(request PlanRequest) (CampaignPlan, error) {
	wantLane := map[Market]string{MarketKR: KRWeeklyLaneID, MarketUS: USWeeklyLaneID}[request.Market]
	budget, budgetOK := parseUnsigned(request.RiskBudgetMinor)
	perShare, perShareOK := parseUnsigned(request.PerShareRiskMinor)
	if wantLane == "" || request.LaneID != wantLane || request.LaneVersion != LaneVersionV1 || request.PositionGeneration == 0 || request.PlannedQuantity == 0 ||
		budgetOK == false || perShareOK == false || budget.Sign() <= 0 || perShare.Sign() <= 0 {
		return CampaignPlan{}, errors.New("invalid immutable campaign plan")
	}
	for _, required := range []string{request.AccountRef, request.Symbol, request.CampaignID, request.PolicyDigest, request.ConfigDigest, request.AccountCurrency, request.QuoteCurrency} {
		if !validBoundedIdentity(required) {
			return CampaignPlan{}, errors.New("invalid immutable campaign identity")
		}
	}
	if request.AccountCurrency == request.QuoteCurrency {
		if request.FX != nil {
			return CampaignPlan{}, errors.New("same-currency plan must not carry FX")
		}
	} else if request.FX == nil || !request.FX.valid() {
		return CampaignPlan{}, errors.New("cross-currency plan requires sealed FX")
	}
	plan := CampaignPlan{laneID: request.LaneID, laneVersion: request.LaneVersion, accountRef: request.AccountRef, symbol: request.Symbol,
		campaignID: request.CampaignID, market: request.Market, positionGeneration: request.PositionGeneration, riskBudgetMinor: budget.String(),
		perShareRiskMinor: perShare.String(), plannedQuantity: request.PlannedQuantity, legCeilings: AllocateSeven(request.PlannedQuantity),
		policyDigest: request.PolicyDigest, configDigest: request.ConfigDigest, accountCurrency: request.AccountCurrency, quoteCurrency: request.QuoteCurrency,
		fx: cloneFX(request.FX)}
	plan.digest = planDigest(plan)
	plan.seal = sha256.Sum256([]byte(plan.digest))
	return plan, nil
}

func AllocateSeven(quantity uint64) [7]uint64 {
	var result [7]uint64
	base := quantity / 7
	for index := 0; index < 6; index++ {
		result[index] = base
	}
	result[6] = quantity - base*6
	return result
}

type LegProgress struct {
	Ordinal        int
	FilledQuantity uint64
	Cancelled      bool
	Expired        bool
}

func PlannedLegQuantity(plan CampaignPlan, progress LegProgress, qFinal uint64) uint64 {
	if !plan.valid() || progress.Ordinal < 1 || progress.Ordinal > 7 || progress.Cancelled || progress.Expired {
		return 0
	}
	ceiling := plan.legCeilings[progress.Ordinal-1]
	if progress.FilledQuantity >= ceiling {
		return 0
	}
	remaining := ceiling - progress.FilledQuantity
	if qFinal < remaining {
		return qFinal
	}
	return remaining
}

func (p CampaignPlan) Market() Market         { return p.market }
func (p CampaignPlan) AccountRef() string     { return p.accountRef }
func (p CampaignPlan) CampaignID() string     { return p.campaignID }
func (p CampaignPlan) LegCeilings() [7]uint64 { return p.legCeilings }
func (p CampaignPlan) FX() *FrozenFX          { return cloneFX(p.fx) }
func (p CampaignPlan) Digest() string         { return p.digest }
func (p CampaignPlan) LaneID() string         { return p.laneID }
func (p CampaignPlan) Symbol() string         { return p.symbol }

func (p CampaignPlan) valid() bool {
	return p.digest != "" && p.seal == sha256.Sum256([]byte(p.digest)) && p.digest == planDigest(p)
}

func planDigest(plan CampaignPlan) string {
	parts := []string{plan.laneID, plan.laneVersion, string(plan.market), plan.accountRef, plan.symbol, plan.campaignID,
		strconv.FormatUint(plan.positionGeneration, 10), plan.riskBudgetMinor, plan.perShareRiskMinor, strconv.FormatUint(plan.plannedQuantity, 10),
		plan.policyDigest, plan.configDigest, plan.accountCurrency, plan.quoteCurrency}
	for _, ceiling := range plan.legCeilings {
		parts = append(parts, strconv.FormatUint(ceiling, 10))
	}
	if plan.fx != nil {
		parts = append(parts, hex.EncodeToString(plan.fx.seal[:]))
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}

type frozenFXInput struct {
	Authority, Version, QuoteID, AsOf, FreshUntil, Direction string
	RateQuoteToAccount, Haircut, Digest                      string
}

type FrozenFX struct {
	authority, version, quoteID, direction string
	asOf, freshUntil                       time.Time
	rateQuoteToAccount, haircut, digest    string
	seal                                   [32]byte
}

func mintFrozenFX(input frozenFXInput) (FrozenFX, error) {
	asOf, firstErr := time.Parse(time.RFC3339Nano, input.AsOf)
	freshUntil, secondErr := time.Parse(time.RFC3339Nano, input.FreshUntil)
	_, rateOK := parsePositiveDecimal(input.RateQuoteToAccount)
	haircut, haircutOK := parsePositiveDecimal(input.Haircut)
	if firstErr != nil || secondErr != nil || input.Authority != a066FXAuthority || input.Version == "" || input.QuoteID == "" || input.Direction != FXQuoteToAccount ||
		input.Digest == "" || !rateOK || !haircutOK || haircut.Cmp(new(big.Rat).SetInt64(1)) < 0 || asOf.After(freshUntil) {
		return FrozenFX{}, errors.New("invalid frozen FX")
	}
	for _, identity := range []string{input.Authority, input.Version, input.QuoteID, input.Direction, input.Digest} {
		if !validBoundedIdentity(identity) {
			return FrozenFX{}, errors.New("invalid frozen FX identity")
		}
	}
	fx := FrozenFX{authority: input.Authority, version: input.Version, quoteID: input.QuoteID, asOf: asOf, freshUntil: freshUntil,
		direction: input.Direction, rateQuoteToAccount: input.RateQuoteToAccount, haircut: input.Haircut, digest: input.Digest}
	fx.seal = fxSeal(fx)
	return fx, nil
}

func (f FrozenFX) valid() bool { return f.seal != ([32]byte{}) && f.seal == fxSeal(f) }

func (f FrozenFX) validAt(evaluatedAt time.Time) bool {
	return f.valid() && !evaluatedAt.IsZero() && !evaluatedAt.Before(f.asOf) && !evaluatedAt.After(f.freshUntil)
}

func fxSeal(f FrozenFX) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{f.authority, f.version, f.quoteID, f.asOf.UTC().Format(time.RFC3339Nano),
		f.freshUntil.UTC().Format(time.RFC3339Nano), f.direction, f.rateQuoteToAccount, f.haircut, f.digest}, "\x00")))
}

func cloneFX(fx *FrozenFX) *FrozenFX {
	if fx == nil {
		return nil
	}
	copy := *fx
	return &copy
}

type riskCapInput struct {
	Authority, Version                        string
	QFinal, ReservationQuantity               uint64
	ReservationMinor, MaxStopDistanceMinor    string
	SnapshotID, PolicyDigest, BucketSetDigest string
	ObservedAt, FreshUntil                    string
	FX                                        *FrozenFX
}

type RiskCap struct {
	authority, version                        string
	qFinal, reservationQuantity               uint64
	reservationMinor, maxStopDistanceMinor    string
	snapshotID, policyDigest, bucketSetDigest string
	observedAt, freshUntil                    time.Time
	fx                                        *FrozenFX
	planDigest                                string
	seal                                      [32]byte
}

func mintRiskCap(plan CampaignPlan, input riskCapInput) (RiskCap, error) {
	reservation, reservationOK := parseUnsigned(input.ReservationMinor)
	stop, stopOK := parseUnsigned(input.MaxStopDistanceMinor)
	observed, observedErr := time.Parse(time.RFC3339Nano, input.ObservedAt)
	fresh, freshErr := time.Parse(time.RFC3339Nano, input.FreshUntil)
	if !plan.valid() || input.Authority != a066RiskCapAuthority || input.Version == "" || input.QFinal == 0 || input.ReservationQuantity == 0 ||
		!reservationOK || !stopOK || reservation.Sign() <= 0 || stop.Sign() <= 0 || input.SnapshotID == "" || input.PolicyDigest != plan.policyDigest ||
		input.BucketSetDigest == "" || observedErr != nil || freshErr != nil || observed.After(fresh) || !sameFX(input.FX, plan.fx) {
		return RiskCap{}, errors.New("invalid a066 risk cap")
	}
	for _, identity := range []string{input.Authority, input.Version, input.SnapshotID, input.PolicyDigest, input.BucketSetDigest} {
		if !validBoundedIdentity(identity) {
			return RiskCap{}, errors.New("invalid a066 cap identity")
		}
	}
	cap := RiskCap{authority: input.Authority, version: input.Version, qFinal: input.QFinal, reservationQuantity: input.ReservationQuantity,
		reservationMinor: reservation.String(), maxStopDistanceMinor: stop.String(), snapshotID: input.SnapshotID, policyDigest: input.PolicyDigest,
		bucketSetDigest: input.BucketSetDigest, observedAt: observed, freshUntil: fresh, fx: cloneFX(input.FX), planDigest: plan.digest}
	cap.seal = riskCapSeal(cap)
	return cap, nil
}

func (c RiskCap) validAt(plan CampaignPlan, evaluatedAt time.Time, quantity uint64) bool {
	return plan.valid() && c.seal != ([32]byte{}) && c.seal == riskCapSeal(c) && c.planDigest == plan.digest && c.policyDigest == plan.policyDigest &&
		!evaluatedAt.Before(c.observedAt) && !evaluatedAt.After(c.freshUntil) && quantity > 0 && quantity <= c.qFinal && quantity == c.reservationQuantity &&
		sameFX(c.fx, plan.fx) && (plan.fx == nil || plan.fx.validAt(evaluatedAt))
}

func riskCapSeal(cap RiskCap) [32]byte {
	parts := []string{cap.authority, cap.version, strconv.FormatUint(cap.qFinal, 10), strconv.FormatUint(cap.reservationQuantity, 10), cap.reservationMinor,
		cap.maxStopDistanceMinor, cap.snapshotID, cap.policyDigest, cap.bucketSetDigest, cap.observedAt.UTC().Format(time.RFC3339Nano), cap.freshUntil.UTC().Format(time.RFC3339Nano), cap.planDigest}
	if cap.fx != nil {
		parts = append(parts, hex.EncodeToString(cap.fx.seal[:]))
	}
	return sha256.Sum256([]byte(strings.Join(parts, "\x00")))
}

func sameFX(left, right *FrozenFX) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.valid() && right.valid() && left.seal == right.seal
}

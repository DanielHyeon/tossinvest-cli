package reversallane

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	maxRiskBits      = 256
	maxMinorDigits   = 78
	maxIdentityBytes = 256
)

const (
	a066RiskCapAuthority = "a066-risk-cap-authority"
	a066FXAuthority      = "a066-official-fx-authority"
)

func mintRiskCap(plan CampaignPlan, candidate RiskCap) (RiskCap, error) {
	if !plan.valid() || candidate.Authority != a066RiskCapAuthority || candidate.Version == "" || candidate.Market != plan.Market() ||
		candidate.QFinal == 0 || candidate.ReservationQuantity == 0 || candidate.SnapshotID == "" || candidate.PolicyDigest != plan.request.PolicyDigest || candidate.BucketSetDigest == "" || !candidate.Official || !candidate.Frozen ||
		candidate.ObservedAt.IsZero() || candidate.FreshUntil.IsZero() || candidate.ObservedAt.After(candidate.FreshUntil) {
		return RiskCap{}, fmt.Errorf("%s", RefusalCapInvalid)
	}
	if reservation, ok := parseMinor(candidate.ReservationMinor); !ok || reservation.Sign() <= 0 {
		return RiskCap{}, fmt.Errorf("%s", RefusalCapInvalid)
	}
	candidate.PlanDigest = plan.Digest()
	candidate.seal = riskCapSeal(candidate)
	return candidate, nil
}

func (c RiskCap) validAt(plan CampaignPlan, evaluatedAt time.Time) bool {
	return plan.valid() && c.seal != "" && c.seal == riskCapSeal(c) && c.Authority == a066RiskCapAuthority && c.Version != "" && c.PlanDigest == plan.Digest() &&
		c.Market == plan.Market() && c.QFinal > 0 && c.ReservationQuantity > 0 && c.PolicyDigest == plan.request.PolicyDigest && c.Official && c.Frozen &&
		!evaluatedAt.IsZero() && !c.ObservedAt.After(evaluatedAt) && !c.FreshUntil.Before(evaluatedAt)
}

func riskCapSeal(cap RiskCap) string {
	preimage := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%d\x00%d\x00%s\x00%s\x00%s\x00%s\x00%t\x00%t\x00%s\x00%s", cap.Authority, cap.Version, cap.PlanDigest, cap.Market, cap.QFinal,
		cap.ReservationQuantity, cap.ReservationMinor, cap.SnapshotID, cap.PolicyDigest, cap.BucketSetDigest, cap.Official, cap.Frozen, cap.ObservedAt.UTC().Format(time.RFC3339Nano), cap.FreshUntil.UTC().Format(time.RFC3339Nano))
	sum := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(sum[:])
}

func mintFrozenFX(candidate FrozenFX) (FrozenFX, error) {
	if candidate.Authority != a066FXAuthority || candidate.Version == "" || !candidate.Official || !candidate.Frozen || candidate.QuoteID == "" || candidate.Digest == "" || candidate.Direction != FXQuoteToAccount {
		return FrozenFX{}, fmt.Errorf("%s", RefusalPlanInvalid)
	}
	asOf, asOfErr := time.Parse(time.RFC3339Nano, candidate.AsOf)
	freshUntil, freshErr := time.Parse(time.RFC3339Nano, candidate.FreshUntil)
	if asOfErr != nil || freshErr != nil || asOf.After(freshUntil) {
		return FrozenFX{}, fmt.Errorf("%s", RefusalPlanInvalid)
	}
	rate, rateOK := parsePositiveRat(candidate.RateQuoteToAccount)
	haircut, haircutOK := parsePositiveRat(candidate.Haircut)
	if !rateOK || rate.Sign() <= 0 || !haircutOK || haircut.Cmp(big.NewRat(1, 1)) < 0 {
		return FrozenFX{}, fmt.Errorf("%s", RefusalPlanInvalid)
	}
	candidate.seal = frozenFXSeal(candidate)
	return candidate, nil
}

func (fx FrozenFX) validAt(evaluatedAt time.Time) bool {
	asOf, asOfErr := time.Parse(time.RFC3339Nano, fx.AsOf)
	freshUntil, freshErr := time.Parse(time.RFC3339Nano, fx.FreshUntil)
	return fx.seal != "" && fx.seal == frozenFXSeal(fx) && fx.Authority == a066FXAuthority && fx.Official && fx.Frozen && asOfErr == nil && freshErr == nil &&
		!asOf.After(evaluatedAt) && !freshUntil.Before(evaluatedAt)
}

func frozenFXSeal(fx FrozenFX) string {
	preimage := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%t\x00%t\x00%s", fx.Authority, fx.Version, fx.QuoteID, fx.AsOf,
		fx.Direction, fx.RateQuoteToAccount, fx.Haircut, fx.Digest, fx.Official, fx.Frozen, fx.FreshUntil)
	sum := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(sum[:])
}

func NewRiskState(plan CampaignPlan) RiskState {
	return RiskState{PlanDigest: plan.Digest(), FilledMinor: "0", HeldMinor: "0", Fills: make(map[string]AppliedFill), Cancels: make(map[string]string), Latches: make(map[Latch]bool)}
}

func AdmitRisk(plan CampaignPlan, state RiskState, cap RiskCap) RefusalCode {
	if !plan.valid() || state.PlanDigest != plan.Digest() {
		return RefusalPlanInvalid
	}
	if state.Latches[LatchCampaignRiskOverage] || state.Latches[LatchUnknownActualRisk] {
		return RefusalRiskLatched
	}
	if cap.seal == "" || cap.seal != riskCapSeal(cap) || cap.PlanDigest != plan.Digest() || !cap.Official || !cap.Frozen || cap.Market != plan.Market() || cap.QFinal == 0 || cap.ReservationQuantity == 0 || cap.PolicyDigest != plan.request.PolicyDigest {
		return RefusalCapInvalid
	}
	filled, ok := parseMinor(state.FilledMinor)
	if !ok {
		return RefusalArithmeticInvalid
	}
	held, ok := parseMinor(state.HeldMinor)
	if !ok {
		return RefusalArithmeticInvalid
	}
	proposed, ok := parseMinor(cap.ReservationMinor)
	if !ok {
		return RefusalArithmeticInvalid
	}
	budget, ok := parseMinor(plan.request.RiskBudgetMinor)
	if !ok {
		return RefusalPlanInvalid
	}
	total := new(big.Int).Add(filled, held)
	total.Add(total, proposed)
	if total.BitLen() > maxRiskBits {
		return RefusalArithmeticInvalid
	}
	if total.Cmp(budget) > 0 {
		return RefusalRiskBudgetExceeded
	}
	return ""
}

func CalculateActualRisk(plan CampaignPlan, event FillRiskEvent) (string, bool) {
	if !plan.valid() || strings.TrimSpace(event.FillID) == "" || event.CampaignID != plan.CampaignID() || event.LegOrdinal < 1 || event.LegOrdinal > 3 ||
		strings.TrimSpace(event.OrderRef) == "" || len(event.OrderRef) > maxIdentityBytes || event.Quantity == 0 || event.ObservedAt.IsZero() ||
		strings.TrimSpace(event.SourceDigest) == "" || len(event.SourceDigest) > maxIdentityBytes {
		return "", false
	}
	transferred, ok := parseMinor(event.TransferredReservationMinor)
	if !ok {
		return "", false
	}
	entry, ok := parseMinor(event.EntryPriceMinor)
	if !ok {
		return "", false
	}
	stop, ok := parseMinor(event.EffectiveStopMinor)
	if !ok {
		return "", false
	}
	entryFees, ok := parseMinor(event.EntryFeesMinor)
	if !ok {
		return "", false
	}
	exitFees, ok := parseMinor(event.EstimatedExitFeesLeviesMinor)
	if !ok {
		return "", false
	}
	distance := new(big.Int).Sub(entry, stop)
	if distance.Sign() < 0 {
		distance.SetInt64(0)
	}
	raw := new(big.Int).Mul(distance, new(big.Int).SetUint64(event.Quantity))
	raw.Add(raw, entryFees)
	raw.Add(raw, exitFees)
	if raw.BitLen() > maxRiskBits {
		return "", false
	}
	accountRisk := raw
	if plan.request.AccountCurrency != plan.request.QuoteCurrency {
		if plan.request.FX == nil || event.FX == nil || *event.FX != *plan.request.FX || !validFrozenFX(*event.FX) || !event.FX.validAt(event.ObservedAt) {
			return "", false
		}
		rate, ok := parsePositiveRat(event.FX.RateQuoteToAccount)
		if !ok {
			return "", false
		}
		haircut, ok := parsePositiveRat(event.FX.Haircut)
		if !ok || haircut.Cmp(big.NewRat(1, 1)) < 0 {
			return "", false
		}
		converted := new(big.Rat).SetInt(raw)
		converted.Mul(converted, rate)
		converted.Mul(converted, haircut)
		accountRisk = ceilRat(converted)
		if accountRisk.BitLen() > maxRiskBits {
			return "", false
		}
	} else if event.FX != nil {
		return "", false
	}
	if accountRisk.Cmp(transferred) < 0 {
		accountRisk = transferred
	}
	return accountRisk.String(), true
}

func ApplyFillRisk(state RiskState, plan CampaignPlan, event FillRiskEvent) (RiskState, RiskApplyResult) {
	next := cloneRiskState(state)
	fillID := strings.TrimSpace(event.FillID)
	fingerprint := fillRiskFingerprint(event)
	unidentified := fillID == "" || len(fillID) > maxIdentityBytes
	if unidentified {
		fillID = "missing:" + fingerprint
	}
	if previous, exists := next.Fills[fillID]; exists {
		if !plan.valid() || state.PlanDigest != plan.Digest() {
			next.Latches[LatchUnknownActualRisk] = true
		}
		if previous.Fingerprint != fingerprint {
			applied := event.Quantity > 0
			next.Fills["conflict:"+fillID+":"+fingerprint] = AppliedFill{Applied: applied, Fingerprint: fingerprint}
			next.Latches[LatchUnknownActualRisk] = true
			return next, RiskApplyResult{Applied: applied, Duplicate: true}
		}
		return next, RiskApplyResult{Duplicate: true}
	}
	if event.Quantity == 0 {
		next.Fills[fillID] = AppliedFill{Fingerprint: fingerprint}
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{}
	}
	if !plan.valid() || state.PlanDigest != plan.Digest() || unidentified {
		next.Fills[fillID] = AppliedFill{Applied: true, Fingerprint: fingerprint}
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{Applied: true}
	}
	risk, known := CalculateActualRisk(plan, event)
	next.Fills[fillID] = AppliedFill{Applied: true, RiskMinor: risk, Fingerprint: fingerprint}
	if !known {
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{Applied: true}
	}
	release, releaseOK := parseMinor(event.TransferredReservationMinor)
	held, heldOK := parseMinor(next.HeldMinor)
	filled, filledOK := parseMinor(next.FilledMinor)
	riskMinor, riskOK := parseMinor(risk)
	if !releaseOK || !heldOK || release.Cmp(held) > 0 || !filledOK || !riskOK {
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{Applied: true}
	}
	newFilled := new(big.Int).Add(filled, riskMinor)
	if newFilled.BitLen() > maxRiskBits {
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{Applied: true}
	}
	next.HeldMinor = new(big.Int).Sub(held, release).String()
	next.FilledMinor = newFilled.String()
	budget, budgetOK := parseMinor(plan.request.RiskBudgetMinor)
	if !budgetOK || newFilled.Cmp(budget) > 0 {
		next.Latches[LatchCampaignRiskOverage] = true
	}
	return next, RiskApplyResult{Applied: true}
}

func ApplyCancelRisk(state RiskState, event CancelRiskEvent) (RiskState, RiskApplyResult) {
	next := cloneRiskState(state)
	cancelID := strings.TrimSpace(event.CancelID)
	if cancelID == "" || len(cancelID) > maxIdentityBytes {
		cancelID = "missing:" + cancelRiskFingerprint(event)
		if _, exists := next.Cancels[cancelID]; exists {
			return next, RiskApplyResult{Duplicate: true}
		}
		next.Cancels[cancelID] = event.ReleaseHeldMinor
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{Applied: true}
	}
	if previous, exists := next.Cancels[cancelID]; exists {
		if previous != event.ReleaseHeldMinor {
			next.Cancels["conflict:"+cancelID+":"+cancelRiskFingerprint(event)] = event.ReleaseHeldMinor
			next.Latches[LatchUnknownActualRisk] = true
			return next, RiskApplyResult{Applied: true, Duplicate: true}
		}
		return next, RiskApplyResult{Duplicate: true}
	}
	release, releaseOK := parseMinor(event.ReleaseHeldMinor)
	held, heldOK := parseMinor(next.HeldMinor)
	if !releaseOK || !heldOK || release.Cmp(held) > 0 {
		next.Cancels[cancelID] = event.ReleaseHeldMinor
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{Applied: true}
	}
	next.HeldMinor = new(big.Int).Sub(held, release).String()
	next.Cancels[cancelID] = event.ReleaseHeldMinor
	return next, RiskApplyResult{Applied: true}
}

func cloneRiskState(state RiskState) RiskState {
	next := state
	next.Fills = make(map[string]AppliedFill, len(state.Fills))
	for key, value := range state.Fills {
		next.Fills[key] = value
	}
	next.Cancels = make(map[string]string, len(state.Cancels))
	for key, value := range state.Cancels {
		next.Cancels[key] = value
	}
	next.Latches = make(map[Latch]bool, len(state.Latches))
	for key, value := range state.Latches {
		next.Latches[key] = value
	}
	return next
}

func parseMinor(value string) (*big.Int, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxMinorDigits || strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-") || (len(value) > 1 && value[0] == '0') {
		return nil, false
	}
	result, ok := new(big.Int).SetString(value, 10)
	if !ok || result.Sign() < 0 || result.BitLen() > maxRiskBits {
		return nil, false
	}
	return result, true
}

func parsePositiveRat(value string) (*big.Rat, bool) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 || strings.ContainsAny(value, "/eE+-") || strings.Count(value, ".") > 1 {
		return nil, false
	}
	result, ok := new(big.Rat).SetString(value)
	if !ok || result.Sign() <= 0 || result.Num().BitLen() > maxRiskBits || result.Denom().BitLen() > maxRiskBits {
		return nil, false
	}
	return result, true
}

func ceilRat(value *big.Rat) *big.Int {
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(value.Num(), value.Denom(), remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

func validFrozenFX(fx FrozenFX) bool {
	if fx.seal == "" || fx.seal != frozenFXSeal(fx) || fx.Authority != a066FXAuthority || fx.Version == "" || strings.TrimSpace(fx.QuoteID) == "" || strings.TrimSpace(fx.AsOf) == "" || fx.Direction != FXQuoteToAccount || strings.TrimSpace(fx.Digest) == "" || !fx.Official || !fx.Frozen {
		return false
	}
	if _, err := time.Parse(time.RFC3339Nano, fx.AsOf); err != nil {
		return false
	}
	rate, rateOK := parsePositiveRat(fx.RateQuoteToAccount)
	haircut, haircutOK := parsePositiveRat(fx.Haircut)
	return rateOK && rate.Sign() > 0 && haircutOK && haircut.Cmp(big.NewRat(1, 1)) >= 0
}

func fillRiskFingerprint(event FillRiskEvent) string {
	fx := ""
	if event.FX != nil {
		fx = frozenFXSeal(*event.FX)
	}
	preimage := fmt.Sprintf("%s\x00%d\x00%s\x00%d\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s\x00%s", event.CampaignID, event.LegOrdinal,
		event.OrderRef, event.Quantity, event.TransferredReservationMinor, event.EntryPriceMinor, event.EffectiveStopMinor, event.EntryFeesMinor,
		event.EstimatedExitFeesLeviesMinor, event.ObservedAt.UTC().Format(time.RFC3339Nano), event.SourceDigest, fx)
	sum := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(sum[:])
}

func cancelRiskFingerprint(event CancelRiskEvent) string {
	preimage := fmt.Sprintf("%s\x00%d\x00%s\x00%s\x00%s\x00%s", event.CampaignID, event.LegOrdinal, event.OrderRef, event.ReleaseHeldMinor,
		event.ObservedAt.UTC().Format(time.RFC3339Nano), event.SourceDigest)
	sum := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(sum[:])
}

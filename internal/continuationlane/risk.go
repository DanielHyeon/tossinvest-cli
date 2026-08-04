package continuationlane

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"math/big"
	"strconv"
	"strings"
	"time"
)

const maxRiskIdentityBytes = 256

type FillRiskRecord struct {
	Applied     bool
	RiskMinor   string
	Known       bool
	Fingerprint string
}

type CancelRiskRecord struct {
	Applied     bool
	Fingerprint string
}

type RiskState struct {
	RiskBudgetDigest string
	BudgetMinor      string
	FilledMinor      string
	HeldMinor        string
	Latches          map[Latch]bool
	Fills            map[string]FillRiskRecord
	Cancels          map[string]CancelRiskRecord
}

type FillRiskEvent struct {
	FillID                       string
	CampaignID                   string
	LegOrdinal                   int
	OrderRef                     string
	Quantity                     uint64
	TransferredReservationMinor  string
	EntryPriceMinor              string
	EffectiveStopMinor           string
	EntryFeesMinor               string
	EstimatedExitFeesLeviesMinor string
	ObservedAt                   string
	SourceDigest                 string
	FX                           *FrozenFX
}

type CancelRiskEvent struct {
	CancelID         string
	CampaignID       string
	LegOrdinal       int
	OrderRef         string
	ReleaseHeldMinor string
	ObservedAt       string
	SourceDigest     string
}

type RiskApplyResult struct {
	Applied   bool
	Duplicate bool
}

func NewRiskState(plan CampaignPlan) RiskState {
	return RiskState{RiskBudgetDigest: plan.RiskBudgetDigest, BudgetMinor: plan.RiskBudgetMinor, FilledMinor: "0", HeldMinor: "0", Latches: make(map[Latch]bool), Fills: make(map[string]FillRiskRecord), Cancels: make(map[string]CancelRiskRecord)}
}

func CalculateActualRisk(plan CampaignPlan, event FillRiskEvent) (string, bool) {
	if event.Quantity == 0 || !validRiskEventScope(plan, event.CampaignID, event.LegOrdinal, event.OrderRef, event.ObservedAt, event.SourceDigest) {
		return "", false
	}
	transferred, err := parseUnsigned(event.TransferredReservationMinor)
	if err != nil {
		return "", false
	}
	entry, err := parseUnsigned(event.EntryPriceMinor)
	if err != nil {
		return "", false
	}
	stop, err := parseUnsigned(event.EffectiveStopMinor)
	if err != nil {
		return "", false
	}
	entryFees, err := parseUnsigned(event.EntryFeesMinor)
	if err != nil {
		return "", false
	}
	exitFees, err := parseUnsigned(event.EstimatedExitFeesLeviesMinor)
	if err != nil {
		return "", false
	}
	distance := new(big.Int).Sub(entry, stop)
	if distance.Sign() < 0 {
		distance.SetInt64(0)
	}
	quantityRisk, err := checkedMul(new(big.Int).SetUint64(event.Quantity), distance)
	if err != nil {
		return "", false
	}
	quoteRisk, err := checkedAdd(quantityRisk, entryFees, exitFees)
	if err != nil {
		return "", false
	}
	actual := quoteRisk
	if plan.AccountCurrency != plan.QuoteCurrency {
		if plan.FX == nil || event.FX == nil || *event.FX != *plan.FX || !validFX(*event.FX) {
			return "", false
		}
		rate, err := parsePositiveDecimal(event.FX.RateQuoteToAccount)
		if err != nil {
			return "", false
		}
		haircut, err := parsePositiveDecimal(event.FX.Haircut)
		if err != nil {
			return "", false
		}
		converted := new(big.Rat).SetInt(quoteRisk)
		converted.Mul(converted, rate)
		converted.Mul(converted, haircut)
		actual, err = ceilRat(converted)
		if err != nil {
			return "", false
		}
	}
	if transferred.Cmp(actual) > 0 {
		actual = transferred
	}
	return actual.String(), true
}

func ApplyFillRisk(state RiskState, plan CampaignPlan, event FillRiskEvent) (RiskState, RiskApplyResult) {
	next := cloneRiskState(state)
	fingerprint := fillRiskFingerprint(event)
	if !validRiskStateScope(state, plan) || !validRiskEventScope(plan, event.CampaignID, event.LegOrdinal, event.OrderRef, event.ObservedAt, event.SourceDigest) {
		key := "INVALID_FILL:" + fingerprint
		if _, exists := next.Fills[key]; exists {
			return next, RiskApplyResult{Duplicate: true}
		}
		next.Fills[key] = FillRiskRecord{Fingerprint: fingerprint}
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{}
	}
	fillID := event.FillID
	unidentified := fillID == ""
	if unidentified {
		fillID = "UNIDENTIFIED_FILL:" + fingerprint
	}
	if previous, exists := next.Fills[fillID]; exists {
		if previous.Fingerprint == fingerprint {
			return next, RiskApplyResult{Duplicate: true}
		}
		conflictID := "CONFLICT_FILL:" + fillID + ":" + fingerprint
		if _, conflictExists := next.Fills[conflictID]; conflictExists {
			return next, RiskApplyResult{Duplicate: true}
		}
		next.Fills[conflictID] = FillRiskRecord{Fingerprint: fingerprint}
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{}
	}
	if event.Quantity == 0 {
		next.Latches[LatchUnknownActualRisk] = true
		next.Fills[fillID] = FillRiskRecord{Fingerprint: fingerprint}
		return next, RiskApplyResult{}
	}
	if unidentified {
		next.Latches[LatchUnknownActualRisk] = true
	}
	risk, known := CalculateActualRisk(plan, event)
	transferred, transferredErr := parseUnboundedUnsigned(event.TransferredReservationMinor)
	if transferredErr != nil {
		known = false
	}
	if !known {
		next.Latches[LatchUnknownActualRisk] = true
		if transferredErr == nil {
			risk = transferred.String()
		} else {
			risk = ""
		}
	}
	held, heldErr := parseUnboundedUnsigned(next.HeldMinor)
	filled, filledErr := parseUnboundedUnsigned(next.FilledMinor)
	riskInt, riskErr := parseUnboundedUnsigned(risk)
	budget, budgetErr := parseUnboundedUnsigned(plan.RiskBudgetMinor)
	accountingValid := transferredErr == nil && heldErr == nil && filledErr == nil && riskErr == nil && budgetErr == nil && transferred.Cmp(held) <= 0
	var newHeld, newFilled *big.Int
	if accountingValid {
		newHeld = new(big.Int).Sub(new(big.Int).Set(held), transferred)
		var addErr error
		newFilled, addErr = checkedAdd(filled, riskInt)
		if addErr != nil {
			accountingValid = false
		}
	}
	if !accountingValid {
		next.Latches[LatchUnknownActualRisk] = true
		next.Fills[fillID] = FillRiskRecord{Applied: true, RiskMinor: risk, Known: known, Fingerprint: fingerprint}
		return next, RiskApplyResult{Applied: true}
	}
	// Commit both accounting values only after every parse, bound and overflow
	// check has succeeded. This prevents a fill from partially moving held risk.
	next.HeldMinor = newHeld.String()
	next.FilledMinor = newFilled.String()
	if newFilled.Cmp(budget) > 0 {
		next.Latches[LatchCampaignRiskOverage] = true
	}
	next.Fills[fillID] = FillRiskRecord{Applied: true, RiskMinor: risk, Known: known, Fingerprint: fingerprint}
	return next, RiskApplyResult{Applied: true}
}

func ApplyCancelRisk(state RiskState, plan CampaignPlan, event CancelRiskEvent) (RiskState, RiskApplyResult) {
	next := cloneRiskState(state)
	fingerprint := cancelRiskFingerprint(event)
	if !validRiskStateScope(state, plan) || !validRiskEventScope(plan, event.CampaignID, event.LegOrdinal, event.OrderRef, event.ObservedAt, event.SourceDigest) {
		key := "INVALID_CANCEL:" + fingerprint
		if _, exists := next.Cancels[key]; exists {
			return next, RiskApplyResult{Duplicate: true}
		}
		next.Cancels[key] = CancelRiskRecord{Fingerprint: fingerprint}
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{}
	}
	cancelID := event.CancelID
	unidentified := cancelID == ""
	if unidentified {
		cancelID = "UNIDENTIFIED_CANCEL:" + fingerprint
	}
	if previous, exists := next.Cancels[cancelID]; exists {
		if previous.Fingerprint == fingerprint {
			return next, RiskApplyResult{Duplicate: true}
		}
		conflictID := "CONFLICT_CANCEL:" + cancelID + ":" + fingerprint
		if _, conflictExists := next.Cancels[conflictID]; conflictExists {
			return next, RiskApplyResult{Duplicate: true}
		}
		next.Cancels[conflictID] = CancelRiskRecord{Fingerprint: fingerprint}
		next.Latches[LatchUnknownActualRisk] = true
		return next, RiskApplyResult{}
	}
	if unidentified {
		next.Latches[LatchUnknownActualRisk] = true
		next.Cancels[cancelID] = CancelRiskRecord{Applied: true, Fingerprint: fingerprint}
		return next, RiskApplyResult{Applied: true}
	}
	held, heldErr := parseUnboundedUnsigned(next.HeldMinor)
	release, releaseErr := parseUnboundedUnsigned(event.ReleaseHeldMinor)
	if heldErr != nil || releaseErr != nil {
		next.Latches[LatchUnknownActualRisk] = true
		next.Cancels[cancelID] = CancelRiskRecord{Applied: true, Fingerprint: fingerprint}
		return next, RiskApplyResult{Applied: true}
	}
	if release.Cmp(held) > 0 {
		next.Latches[LatchUnknownActualRisk] = true
		next.Cancels[cancelID] = CancelRiskRecord{Applied: true, Fingerprint: fingerprint}
		return next, RiskApplyResult{Applied: true}
	}
	next.HeldMinor = new(big.Int).Sub(held, release).String()
	next.Cancels[cancelID] = CancelRiskRecord{Applied: true, Fingerprint: fingerprint}
	return next, RiskApplyResult{Applied: true}
}

func parseUnboundedUnsigned(raw string) (*big.Int, error) {
	return parseUnsigned(raw)
}

func validRiskStateScope(state RiskState, plan CampaignPlan) bool {
	return validatePlan(plan) && state.RiskBudgetDigest == plan.RiskBudgetDigest && state.BudgetMinor == plan.RiskBudgetMinor
}

func validRiskEventScope(plan CampaignPlan, campaignID string, legOrdinal int, orderRef, observedAt, sourceDigest string) bool {
	if !validatePlan(plan) || campaignID != plan.CampaignID || len(campaignID) == 0 || len(campaignID) > maxRiskIdentityBytes || legOrdinal < 1 || legOrdinal > 3 ||
		strings.TrimSpace(orderRef) == "" || len(orderRef) > maxRiskIdentityBytes || strings.TrimSpace(sourceDigest) == "" || len(sourceDigest) > maxRiskIdentityBytes {
		return false
	}
	_, err := time.Parse(time.RFC3339Nano, observedAt)
	return err == nil
}

func fillRiskFingerprint(event FillRiskEvent) string {
	h := sha256.New()
	writeDigestPart(h, event.CampaignID)
	writeDigestPart(h, strconv.Itoa(event.LegOrdinal))
	writeDigestPart(h, event.OrderRef)
	writeDigestPart(h, strconv.FormatUint(event.Quantity, 10))
	writeDigestPart(h, event.TransferredReservationMinor)
	writeDigestPart(h, event.EntryPriceMinor)
	writeDigestPart(h, event.EffectiveStopMinor)
	writeDigestPart(h, event.EntryFeesMinor)
	writeDigestPart(h, event.EstimatedExitFeesLeviesMinor)
	writeDigestPart(h, event.ObservedAt)
	writeDigestPart(h, event.SourceDigest)
	if event.FX == nil {
		writeDigestPart(h, "NO_FX")
	} else {
		writeDigestPart(h, "FX")
		writeDigestPart(h, event.FX.QuoteID)
		writeDigestPart(h, event.FX.AsOf)
		writeDigestPart(h, event.FX.Direction)
		writeDigestPart(h, event.FX.RateQuoteToAccount)
		writeDigestPart(h, event.FX.Haircut)
		writeDigestPart(h, event.FX.Digest)
		writeDigestPart(h, strconv.FormatBool(event.FX.official))
		writeDigestPart(h, strconv.FormatBool(event.FX.frozen))
		writeDigestPart(h, hex.EncodeToString(event.FX.seal[:]))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func cancelRiskFingerprint(event CancelRiskEvent) string {
	h := sha256.New()
	writeDigestPart(h, event.CampaignID)
	writeDigestPart(h, strconv.Itoa(event.LegOrdinal))
	writeDigestPart(h, event.OrderRef)
	writeDigestPart(h, event.ReleaseHeldMinor)
	writeDigestPart(h, event.ObservedAt)
	writeDigestPart(h, event.SourceDigest)
	return hex.EncodeToString(h.Sum(nil))
}

func writeDigestPart(h hash.Hash, value string) {
	_, _ = h.Write([]byte(strconv.Itoa(len(value))))
	_, _ = h.Write([]byte{':'})
	_, _ = h.Write([]byte(value))
}

func cloneRiskState(state RiskState) RiskState {
	next := state
	next.Latches = make(map[Latch]bool, len(state.Latches))
	for key, value := range state.Latches {
		next.Latches[key] = value
	}
	next.Fills = make(map[string]FillRiskRecord, len(state.Fills))
	for key, value := range state.Fills {
		next.Fills[key] = value
	}
	next.Cancels = make(map[string]CancelRiskRecord, len(state.Cancels))
	for key, value := range state.Cancels {
		next.Cancels[key] = value
	}
	return next
}

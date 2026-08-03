package weeklyvaluelane

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
)

type AppliedFill struct {
	Fingerprint string
	Known       bool
}

type RiskState struct {
	planDigest             string
	filledMinor, heldMinor string
	latches                map[Latch]bool
	fills                  map[string]AppliedFill
	seal                   [32]byte
}

func NewRiskState(plan CampaignPlan) RiskState {
	state, _ := mintRiskState(plan, "0", "0")
	return state
}

func mintRiskState(plan CampaignPlan, filledMinor, heldMinor string) (RiskState, error) {
	if !plan.valid() {
		return RiskState{}, errors.New("invalid risk plan")
	}
	if _, ok := parseUnsigned(filledMinor); !ok {
		return RiskState{}, errors.New("invalid filled risk")
	}
	if _, ok := parseUnsigned(heldMinor); !ok {
		return RiskState{}, errors.New("invalid held risk")
	}
	state := RiskState{planDigest: plan.digest, filledMinor: filledMinor, heldMinor: heldMinor, latches: map[Latch]bool{}, fills: map[string]AppliedFill{}}
	state.seal = riskStateSeal(state)
	return state, nil
}

func (state RiskState) PlanDigest() string       { return state.planDigest }
func (state RiskState) FilledMinor() string      { return state.filledMinor }
func (state RiskState) HeldMinor() string        { return state.heldMinor }
func (state RiskState) Latched(latch Latch) bool { return state.latches[latch] }
func (state RiskState) FillCount() int           { return len(state.fills) }

func AdmitRisk(plan CampaignPlan, state RiskState, cap RiskCap) RefusalCode {
	if !plan.valid() || !validRiskState(state) || state.planDigest != plan.digest {
		return RefusalRiskLatched
	}
	for _, latched := range state.latches {
		if latched {
			return RefusalRiskLatched
		}
	}
	filled, filledOK := parseUnsigned(state.filledMinor)
	held, heldOK := parseUnsigned(state.heldMinor)
	proposed, proposedOK := parseUnsigned(cap.reservationMinor)
	budget, budgetOK := parseUnsigned(plan.riskBudgetMinor)
	if !filledOK || !heldOK || !proposedOK || !budgetOK {
		return RefusalRiskLatched
	}
	total, ok := checkedAdd(filled, held, proposed)
	if !ok {
		return RefusalRiskBudgetExceeded
	}
	if total.Cmp(budget) > 0 {
		return RefusalRiskBudgetExceeded
	}
	return ""
}

type FillRiskEvent struct {
	FillID, CampaignID, OrderRef                 string
	LegOrdinal                                   int
	Quantity                                     uint64
	TransferredReservationMinor                  string
	EntryPriceMinor, EffectiveStopMinor          string
	EntryFeesMinor, EstimatedExitFeesLeviesMinor string
	ObservedAt, SourceDigest                     string
	FX                                           *FrozenFX
}

type RiskApplyResult struct {
	Applied, Duplicate bool
	Code               RefusalCode
}

func ApplyFillRisk(state RiskState, plan CampaignPlan, event FillRiskEvent) (RiskState, RiskApplyResult) {
	if !validRiskState(state) {
		return state, RiskApplyResult{Code: RefusalRiskLatched}
	}
	fingerprint := fillFingerprint(event)
	key := event.FillID
	knownIdentity := key != ""
	if !knownIdentity {
		key = "missing:" + fingerprint
	}
	if prior, exists := state.fills[key]; exists {
		if prior.Fingerprint == fingerprint {
			return state, RiskApplyResult{Duplicate: true}
		}
		next := cloneRiskState(state)
		next.latches[LatchUnknownActualRisk] = true
		next.seal = riskStateSeal(next)
		return next, RiskApplyResult{Duplicate: true, Code: RefusalRiskLatched}
	}
	next := cloneRiskState(state)
	next.fills[key] = AppliedFill{Fingerprint: fingerprint, Known: knownIdentity}
	if !knownIdentity || event.Quantity == 0 || !plan.valid() || state.planDigest != plan.digest || event.CampaignID != plan.campaignID ||
		event.LegOrdinal < 1 || event.LegOrdinal > 7 || !validBoundedIdentity(event.OrderRef) || !validBoundedIdentity(event.SourceDigest) ||
		!validBoundedIdentity(event.CampaignID) || !validBoundedIdentity(event.FillID) {
		next.latches[LatchUnknownActualRisk] = true
		next.seal = riskStateSeal(next)
		return next, RiskApplyResult{Applied: true, Code: RefusalRiskLatched}
	}
	observed, err := time.Parse(time.RFC3339Nano, event.ObservedAt)
	if err != nil || (plan.fx == nil && event.FX != nil) || (plan.fx != nil && (!sameFX(event.FX, plan.fx) || observed.Before(plan.fx.asOf) || observed.After(plan.fx.freshUntil))) {
		next.latches[LatchUnknownActualRisk] = true
		next.seal = riskStateSeal(next)
		return next, RiskApplyResult{Applied: true, Code: RefusalRiskLatched}
	}
	held, heldOK := parseUnsigned(state.heldMinor)
	filled, filledOK := parseUnsigned(state.filledMinor)
	transfer, transferOK := parseUnsigned(event.TransferredReservationMinor)
	actual, actualOK := actualFillRisk(plan, event)
	if !heldOK || !filledOK || !transferOK || !actualOK || transfer.Cmp(held) > 0 {
		next.latches[LatchUnknownActualRisk] = true
		next.seal = riskStateSeal(next)
		return next, RiskApplyResult{Applied: true, Code: RefusalRiskLatched}
	}
	accounted := new(big.Int).Set(actual)
	if transfer.Cmp(accounted) > 0 {
		accounted.Set(transfer)
	}
	newFilled, addOK := checkedAdd(filled, accounted)
	if !addOK {
		next.latches[LatchUnknownActualRisk] = true
		next.seal = riskStateSeal(next)
		return next, RiskApplyResult{Applied: true, Code: RefusalRiskLatched}
	}
	next.heldMinor = new(big.Int).Sub(held, transfer).String()
	next.filledMinor = newFilled.String()
	if actual.Cmp(transfer) > 0 {
		next.latches[LatchCampaignRiskOverage] = true
	}
	next.seal = riskStateSeal(next)
	return next, RiskApplyResult{Applied: true}
}

type PositiveFillAtomicResult struct {
	Applied, Duplicate bool
	Code               RefusalCode
	Reservation        ReservationResult
	Risk               RiskApplyResult
}

type PositiveFillState struct {
	Reservations ReservationState
	Risk         RiskState
}

// ApplyPositiveFillAtomic is the only transition that may consume a weekly
// reservation for a positive fill. Reservation and risk accounting are one
// aggregate result; on validation failure the original aggregate is returned.
func ApplyPositiveFillAtomic(state PositiveFillState, plan CampaignPlan, command ReservationCommand, event FillRiskEvent) (PositiveFillState, PositiveFillAtomicResult) {
	if command.Action != ReservationPositiveFill || command.PositiveFillQuantity == 0 || command.PositiveFillQuantity != event.Quantity {
		return state, PositiveFillAtomicResult{Code: RefusalReservationConflict}
	}
	nextReservations, reservationResult := applyReservationTransition(state.Reservations, command, true)
	if reservationResult.Duplicate {
		nextRisk, riskResult := ApplyFillRisk(state.Risk, plan, event)
		if riskResult.Duplicate {
			return state, PositiveFillAtomicResult{Duplicate: true, Code: firstRefusal(reservationResult.Code, riskResult.Code), Reservation: reservationResult, Risk: riskResult}
		}
		_ = nextRisk
		return state, PositiveFillAtomicResult{Code: RefusalRiskLatched, Reservation: reservationResult, Risk: riskResult}
	}
	if !reservationResult.Applied {
		return state, PositiveFillAtomicResult{Code: reservationResult.Code, Reservation: reservationResult}
	}
	nextRisk, riskResult := ApplyFillRisk(state.Risk, plan, event)
	if !riskResult.Applied {
		return state, PositiveFillAtomicResult{Code: firstRefusal(riskResult.Code, RefusalRiskLatched), Reservation: reservationResult, Risk: riskResult}
	}
	return PositiveFillState{Reservations: nextReservations, Risk: nextRisk}, PositiveFillAtomicResult{Applied: true, Code: riskResult.Code, Reservation: reservationResult, Risk: riskResult}
}

func firstRefusal(codes ...RefusalCode) RefusalCode {
	for _, code := range codes {
		if code != "" {
			return code
		}
	}
	return ""
}

func actualFillRisk(plan CampaignPlan, event FillRiskEvent) (*big.Int, bool) {
	entry, entryOK := parseUnsigned(event.EntryPriceMinor)
	stop, stopOK := parseUnsigned(event.EffectiveStopMinor)
	entryFees, entryFeesOK := parseUnsigned(event.EntryFeesMinor)
	exitFees, exitFeesOK := parseUnsigned(event.EstimatedExitFeesLeviesMinor)
	if !entryOK || !stopOK || !entryFeesOK || !exitFeesOK || entry.Cmp(stop) <= 0 {
		return nil, false
	}
	distance := new(big.Int).Sub(entry, stop)
	gross, ok := checkedMul(distance, new(big.Int).SetUint64(event.Quantity))
	if !ok {
		return nil, false
	}
	total, ok := checkedAdd(gross, entryFees, exitFees)
	if !ok {
		return nil, false
	}
	if plan.fx == nil {
		return total, true
	}
	rate, rateOK := parsePositiveDecimal(plan.fx.rateQuoteToAccount)
	haircut, haircutOK := parsePositiveDecimal(plan.fx.haircut)
	if !rateOK || !haircutOK {
		return nil, false
	}
	converted := new(big.Rat).SetInt(total)
	converted.Mul(converted, rate)
	converted.Mul(converted, haircut)
	return ceilRat(converted)
}

func cloneRiskState(state RiskState) RiskState {
	next := RiskState{planDigest: state.planDigest, filledMinor: state.filledMinor, heldMinor: state.heldMinor, latches: make(map[Latch]bool, len(state.latches)), fills: make(map[string]AppliedFill, len(state.fills))}
	for key, value := range state.latches {
		next.latches[key] = value
	}
	for key, value := range state.fills {
		next.fills[key] = value
	}
	return next
}

func validRiskState(state RiskState) bool {
	if state.latches == nil || state.fills == nil || state.seal == ([32]byte{}) || state.seal != riskStateSeal(state) || !validBoundedIdentity(state.planDigest) {
		return false
	}
	_, filledOK := parseUnsigned(state.filledMinor)
	_, heldOK := parseUnsigned(state.heldMinor)
	return filledOK && heldOK
}

func riskStateSeal(state RiskState) [32]byte {
	parts := []string{"weekly-value-risk-state-v1", state.planDigest, state.filledMinor, state.heldMinor}
	latchKeys := make([]string, 0, len(state.latches))
	for latch := range state.latches {
		latchKeys = append(latchKeys, string(latch))
	}
	sort.Strings(latchKeys)
	for _, key := range latchKeys {
		parts = append(parts, key, strconv.FormatBool(state.latches[Latch(key)]))
	}
	fillKeys := make([]string, 0, len(state.fills))
	for key := range state.fills {
		fillKeys = append(fillKeys, key)
	}
	sort.Strings(fillKeys)
	for _, key := range fillKeys {
		fill := state.fills[key]
		parts = append(parts, key, fill.Fingerprint, strconv.FormatBool(fill.Known))
	}
	return sha256.Sum256([]byte(strings.Join(parts, "\x00")))
}

func fillFingerprint(event FillRiskEvent) string {
	parts := []string{event.FillID, event.CampaignID, strconv.Itoa(event.LegOrdinal), event.OrderRef, strconv.FormatUint(event.Quantity, 10), event.TransferredReservationMinor,
		event.EntryPriceMinor, event.EffectiveStopMinor, event.EntryFeesMinor, event.EstimatedExitFeesLeviesMinor, event.ObservedAt, event.SourceDigest}
	if event.FX != nil {
		parts = append(parts, hex.EncodeToString(event.FX.seal[:]))
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}

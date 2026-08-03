package riskbucket

import (
	"math/big"
	"strings"
	"time"
)

type Latch string

const (
	LatchRiskOverage       Latch = "RISK_OVERAGE"
	LatchUnknownActualRisk Latch = "UNKNOWN_ACTUAL_RISK"
)

type BucketUsage struct {
	LimitMinor   string
	HeldMinor    string
	FilledMinor  string
	OverageMinor string
	Latches      map[Latch]bool
}

type FillRecord struct {
	CumulativeFill uint64
	DeltaQuantity  uint64
	TransferMinor  map[BucketKey]string
	FilledMinor    map[BucketKey]string
	ActualKnown    bool
	ActualEvidence *ActualFillEvidence
}

type OrderFillState struct {
	OrderQuantity           uint64
	CumulativeFill          uint64
	QuoteCurrency           string
	BaseCurrency            string
	ReservedMinor           map[BucketKey]string
	TransferredMinor        map[BucketKey]string
	ReservationPolicyDigest string
	Fills                   map[string]FillRecord
}

type FillState struct {
	Buckets      map[BucketKey]BucketUsage
	Orders       map[string]OrderFillState
	OwnerLatches map[Latch]bool
}

func (s FillState) EntryBlocked() bool {
	if s.OwnerLatches[LatchRiskOverage] || s.OwnerLatches[LatchUnknownActualRisk] {
		return true
	}
	for _, usage := range s.Buckets {
		if usage.Latches[LatchRiskOverage] || usage.Latches[LatchUnknownActualRisk] {
			return true
		}
	}
	return false
}

type ActualFillEvidence struct {
	QuoteCurrency         string
	BaseCurrency          string
	PriceQuote            string
	FXRateQuoteToBase     string
	AllocatedFeeBaseMinor string
	Price                 Evidence
	FX                    Evidence
	EvaluatedAt           time.Time
	MaxDecimalBits        uint
}

type FillEvent struct {
	FillID                  string
	OrderID                 string
	OrderQuantity           uint64
	NewCumulativeFill       uint64
	ReservedMinor           map[BucketKey]string
	ReservationPolicyDigest string
	QuoteCurrency           string
	BaseCurrency            string
	Actual                  *ActualFillEvidence
}

type FillResult struct {
	DeltaQuantity           uint64
	Duplicate               bool
	ActualEvidenceCompleted bool
}

// ApplyFill returns a deep-copied next state. Any error returns an unchanged
// deep copy, so a caller can atomically commit all bucket changes or none.
func ApplyFill(state FillState, event FillEvent) (FillState, FillResult, error) {
	unchanged := cloneFillState(state)
	next := cloneFillState(state)
	result := FillResult{}
	if event.FillID == "" || event.OrderID == "" || event.OrderQuantity == 0 || event.NewCumulativeFill == 0 || event.NewCumulativeFill > event.OrderQuantity || event.ReservationPolicyDigest == "" ||
		!canonicalCurrency(event.QuoteCurrency) || !canonicalCurrency(event.BaseCurrency) {
		return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "fill_identity_or_quantity", nil)
	}
	if err := validateFillBuckets(next.Buckets, event.ReservedMinor); err != nil {
		return unchanged, result, err
	}
	order, exists := next.Orders[event.OrderID]
	if !exists {
		order = OrderFillState{
			OrderQuantity:           event.OrderQuantity,
			QuoteCurrency:           event.QuoteCurrency,
			BaseCurrency:            event.BaseCurrency,
			ReservedMinor:           cloneMinorMap(event.ReservedMinor),
			TransferredMinor:        make(map[BucketKey]string, len(event.ReservedMinor)),
			ReservationPolicyDigest: event.ReservationPolicyDigest,
			Fills:                   make(map[string]FillRecord),
		}
		for key := range event.ReservedMinor {
			order.TransferredMinor[key] = "0"
		}
	} else if order.OrderQuantity != event.OrderQuantity || order.QuoteCurrency != event.QuoteCurrency || order.BaseCurrency != event.BaseCurrency || order.ReservationPolicyDigest != event.ReservationPolicyDigest || !equalMinorMaps(order.ReservedMinor, event.ReservedMinor) {
		return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "order_reservation", nil)
	}

	if record, seen := order.Fills[event.FillID]; seen {
		if record.CumulativeFill != event.NewCumulativeFill {
			return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "fill_identity_watermark", nil)
		}
		if record.ActualKnown {
			result.Duplicate = true
			return next, result, nil
		}
		actualMinor, known := actualFillMinor(record.DeltaQuantity, event.Actual, order.QuoteCurrency, order.BaseCurrency)
		if !known {
			result.Duplicate = true
			return next, result, nil
		}
		for key, transferRaw := range record.TransferMinor {
			transfer, err := parseMinor(transferRaw, 0)
			if err != nil {
				return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "record_transfer", err)
			}
			previousFilled, err := parseMinor(record.FilledMinor[key], 0)
			if err != nil {
				return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "record_filled", err)
			}
			target := new(big.Int).Set(transfer)
			if actualMinor.Cmp(target) > 0 {
				target.Set(actualMinor)
			}
			if target.Cmp(previousFilled) > 0 {
				delta := new(big.Int).Sub(target, previousFilled)
				usage := next.Buckets[key]
				filled, err := parseMinor(usage.FilledMinor, 0)
				if err != nil {
					return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "filled_usage", err)
				}
				updatedFilled, err := addMinor(filled, delta, 0)
				if err != nil {
					return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "filled_usage_overflow", err)
				}
				usage.FilledMinor = updatedFilled.String()
				record.FilledMinor[key] = target.String()
				next.Buckets[key] = usage
			}
		}
		record.ActualKnown = true
		record.ActualEvidence = cloneActualFillEvidence(event.Actual)
		order.Fills[event.FillID] = record
		next.Orders[event.OrderID] = order
		if err := recomputeOverageLatches(&next); err != nil {
			return unchanged, FillResult{}, err
		}
		clearResolvedUnknownLatches(&next)
		result.ActualEvidenceCompleted = true
		return next, result, nil
	}

	if event.NewCumulativeFill <= order.CumulativeFill {
		return unchanged, result, refusal(RefusalFillEvidenceInconsistent, "cumulative_fill", nil)
	}
	deltaQuantity := event.NewCumulativeFill - order.CumulativeFill
	actualMinor, actualKnown := actualFillMinor(deltaQuantity, event.Actual, order.QuoteCurrency, order.BaseCurrency)
	record := FillRecord{
		CumulativeFill: event.NewCumulativeFill,
		DeltaQuantity:  deltaQuantity,
		TransferMinor:  make(map[BucketKey]string, len(event.ReservedMinor)),
		FilledMinor:    make(map[BucketKey]string, len(event.ReservedMinor)),
		ActualKnown:    actualKnown,
		ActualEvidence: nil,
	}
	if actualKnown {
		record.ActualEvidence = cloneActualFillEvidence(event.Actual)
	}
	for key, reservedRaw := range event.ReservedMinor {
		reserved, err := parseMinor(reservedRaw, 0)
		if err != nil {
			return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "reserved_minor", err)
		}
		allocated := proportionalAllocation(reserved, event.NewCumulativeFill, event.OrderQuantity)
		previousTransferred, err := parseMinor(order.TransferredMinor[key], 0)
		if err != nil || allocated.Cmp(previousTransferred) < 0 {
			return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "transferred_minor", err)
		}
		transfer := new(big.Int).Sub(allocated, previousTransferred)
		usage := next.Buckets[key]
		held, err := parseMinor(usage.HeldMinor, 0)
		if err != nil {
			return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "held_usage", err)
		}
		filled, err := parseMinor(usage.FilledMinor, 0)
		if err != nil {
			return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "filled_usage", err)
		}
		heldDeduction := new(big.Int).Set(transfer)
		if heldDeduction.Cmp(held) > 0 {
			heldDeduction.Set(held)
			latchUsage(&usage, LatchRiskOverage)
			next.OwnerLatches[LatchRiskOverage] = true
		}
		usage.HeldMinor = new(big.Int).Sub(held, heldDeduction).String()
		filledDelta := new(big.Int).Set(transfer)
		if actualKnown && actualMinor.Cmp(filledDelta) > 0 {
			filledDelta.Set(actualMinor)
		}
		updatedFilled, err := addMinor(filled, filledDelta, 0)
		if err != nil {
			return unchanged, FillResult{}, refusal(RefusalFillEvidenceInconsistent, "filled_usage_overflow", err)
		}
		usage.FilledMinor = updatedFilled.String()
		if !actualKnown {
			latchUsage(&usage, LatchUnknownActualRisk)
			next.OwnerLatches[LatchUnknownActualRisk] = true
		}
		next.Buckets[key] = usage
		order.TransferredMinor[key] = allocated.String()
		record.TransferMinor[key] = transfer.String()
		record.FilledMinor[key] = filledDelta.String()
	}
	order.CumulativeFill = event.NewCumulativeFill
	order.Fills[event.FillID] = record
	next.Orders[event.OrderID] = order
	if err := recomputeOverageLatches(&next); err != nil {
		return unchanged, FillResult{}, err
	}
	result.DeltaQuantity = deltaQuantity
	return next, result, nil
}

func validateFillBuckets(buckets map[BucketKey]BucketUsage, reserved map[BucketKey]string) error {
	if len(buckets) != len(requiredDimensions) || len(reserved) != len(requiredDimensions) {
		return refusal(RefusalFillEvidenceInconsistent, "bucket_count", nil)
	}
	seen := make(map[Dimension]bool, len(requiredDimensions))
	for key, raw := range reserved {
		usage, ok := buckets[key]
		if !ok {
			return refusal(RefusalFillEvidenceInconsistent, "reservation_bucket", nil)
		}
		if seen[key.Dimension] {
			return refusal(RefusalFillEvidenceInconsistent, "duplicate_dimension", nil)
		}
		seen[key.Dimension] = true
		if _, err := parseMinor(raw, 0); err != nil {
			return refusal(RefusalFillEvidenceInconsistent, "reserved_minor", err)
		}
		for field, amount := range map[string]string{
			"limit": usage.LimitMinor, "held": usage.HeldMinor, "filled": usage.FilledMinor, "overage": usage.OverageMinor,
		} {
			if _, err := parseMinor(amount, 0); err != nil {
				return refusal(RefusalFillEvidenceInconsistent, string(key.Dimension)+"_"+field, err)
			}
		}
	}
	for _, dimension := range requiredDimensions {
		if !seen[dimension] {
			return refusal(RefusalFillEvidenceInconsistent, string(dimension), nil)
		}
	}
	return nil
}

func actualFillMinor(quantity uint64, evidence *ActualFillEvidence, quoteCurrency, baseCurrency string) (*big.Int, bool) {
	if evidence == nil || evidence.QuoteCurrency != quoteCurrency || evidence.BaseCurrency != baseCurrency ||
		!canonicalCurrency(evidence.QuoteCurrency) || !canonicalCurrency(evidence.BaseCurrency) ||
		!evidence.Price.validAt(evidence.EvaluatedAt) || !evidence.FX.validAt(evidence.EvaluatedAt) {
		return nil, false
	}
	price, err := parseDecimal(evidence.PriceQuote, false, evidence.MaxDecimalBits)
	if err != nil {
		return nil, false
	}
	fx, err := parseDecimal(evidence.FXRateQuoteToBase, false, evidence.MaxDecimalBits)
	if err != nil {
		return nil, false
	}
	if quoteCurrency == baseCurrency && fx.Cmp(big.NewRat(1, 1)) != 0 {
		return nil, false
	}
	fee, err := parseDecimal(evidence.AllocatedFeeBaseMinor, true, evidence.MaxDecimalBits)
	if err != nil {
		return nil, false
	}
	actual := new(big.Rat).Mul(new(big.Rat).SetInt(new(big.Int).SetUint64(quantity)), price)
	actual.Mul(actual, fx)
	actual.Add(actual, fee)
	if checkRatBits(evidence.MaxDecimalBits, actual) != nil {
		return nil, false
	}
	minor, err := ceilRat(actual, evidence.MaxDecimalBits)
	return minor, err == nil
}

func proportionalAllocation(total *big.Int, cumulative, orderQuantity uint64) *big.Int {
	numerator := new(big.Int).Mul(total, new(big.Int).SetUint64(cumulative))
	denominator := new(big.Int).SetUint64(orderQuantity)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

func recomputeOverageLatches(state *FillState) error {
	anyOverage := state.OwnerLatches[LatchRiskOverage]
	for key, usage := range state.Buckets {
		limit, err := parseMinor(usage.LimitMinor, 0)
		if err != nil {
			return refusal(RefusalFillEvidenceInconsistent, "overage_limit", err)
		}
		held, err := parseMinor(usage.HeldMinor, 0)
		if err != nil {
			return refusal(RefusalFillEvidenceInconsistent, "overage_held", err)
		}
		filled, err := parseMinor(usage.FilledMinor, 0)
		if err != nil {
			return refusal(RefusalFillEvidenceInconsistent, "overage_filled", err)
		}
		used, err := addMinor(filled, held, 0)
		if err != nil {
			return refusal(RefusalFillEvidenceInconsistent, "overage_usage_overflow", err)
		}
		overage := new(big.Int).Sub(used, limit)
		if overage.Sign() > 0 {
			anyOverage = true
			previous, err := parseMinor(usage.OverageMinor, 0)
			if err != nil {
				return refusal(RefusalFillEvidenceInconsistent, "overage_previous", err)
			}
			if overage.Cmp(previous) > 0 {
				usage.OverageMinor = overage.String()
			}
		}
		state.Buckets[key] = usage
	}
	if !anyOverage {
		return nil
	}
	state.OwnerLatches[LatchRiskOverage] = true
	for key, usage := range state.Buckets {
		latchUsage(&usage, LatchRiskOverage)
		state.Buckets[key] = usage
	}
	return nil
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

func clearResolvedUnknownLatches(state *FillState) {
	for _, order := range state.Orders {
		for _, fill := range order.Fills {
			if !fill.ActualKnown {
				return
			}
		}
	}
	delete(state.OwnerLatches, LatchUnknownActualRisk)
	for key, usage := range state.Buckets {
		delete(usage.Latches, LatchUnknownActualRisk)
		state.Buckets[key] = usage
	}
}

func latchUsage(usage *BucketUsage, latch Latch) {
	if usage.Latches == nil {
		usage.Latches = make(map[Latch]bool)
	}
	usage.Latches[latch] = true
}

func cloneFillState(in FillState) FillState {
	out := FillState{Buckets: make(map[BucketKey]BucketUsage, len(in.Buckets)), Orders: make(map[string]OrderFillState, len(in.Orders)), OwnerLatches: make(map[Latch]bool, len(in.OwnerLatches))}
	for key, usage := range in.Buckets {
		usage.Latches = cloneLatchMap(usage.Latches)
		out.Buckets[key] = usage
	}
	for id, order := range in.Orders {
		order.ReservedMinor = cloneMinorMap(order.ReservedMinor)
		order.TransferredMinor = cloneMinorMap(order.TransferredMinor)
		fills := make(map[string]FillRecord, len(order.Fills))
		for fillID, record := range order.Fills {
			record.TransferMinor = cloneMinorMap(record.TransferMinor)
			record.FilledMinor = cloneMinorMap(record.FilledMinor)
			record.ActualEvidence = cloneActualFillEvidence(record.ActualEvidence)
			fills[fillID] = record
		}
		order.Fills = fills
		out.Orders[id] = order
	}
	for latch, set := range in.OwnerLatches {
		out.OwnerLatches[latch] = set
	}
	return out
}

func cloneActualFillEvidence(in *ActualFillEvidence) *ActualFillEvidence {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneMinorMap(in map[BucketKey]string) map[BucketKey]string {
	out := make(map[BucketKey]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneLatchMap(in map[Latch]bool) map[Latch]bool {
	out := make(map[Latch]bool, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func equalMinorMaps(a, b map[BucketKey]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}

package riskbucket

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

type PriceEvidence struct {
	WorstExecutableQuote string
	Evidence
}

type FXEvidence struct {
	RateQuoteToBase string
	Haircut         string
	Evidence
}

type FeePolicy struct {
	FixedBaseMinor   string
	PerUnitBaseMinor string
	MinimumBaseMinor string
	Version          string
	Digest           string
}

type ReservePolicy struct {
	AccountCurrency string
	QuoteCurrency   string
	EvaluatedAt     time.Time
	Price           PriceEvidence
	FX              FXEvidence
	Fee             FeePolicy
	MaxDecimalBits  uint
}

func ReservationMinor(quantity uint64, policy ReservePolicy) (string, error) {
	if quantity == 0 {
		return "0", nil
	}
	price, fx, haircut, fixedFee, perUnitFee, minimumFee, err := validateReservePolicy(policy)
	if err != nil {
		return "", err
	}
	maxBits := policy.MaxDecimalBits
	quantityRat := new(big.Rat).SetInt(new(big.Int).SetUint64(quantity))
	notional := new(big.Rat).Mul(quantityRat, price)
	notional.Mul(notional, fx)
	notional.Mul(notional, haircut)
	if err := checkRatBits(maxBits, notional); err != nil {
		return "", refusal(RefusalRiskCalculationInvalid, "notional", err)
	}
	fee := new(big.Rat).Mul(quantityRat, perUnitFee)
	fee.Add(fee, fixedFee)
	if fee.Cmp(minimumFee) < 0 {
		fee.Set(minimumFee)
	}
	if err := checkRatBits(maxBits, fee); err != nil {
		return "", refusal(RefusalRiskCalculationInvalid, "fee", err)
	}
	total := new(big.Rat).Add(notional, fee)
	if err := checkRatBits(maxBits, total); err != nil {
		return "", refusal(RefusalRiskCalculationInvalid, "total", err)
	}
	minor, err := ceilRat(total, maxBits)
	if err != nil {
		return "", refusal(RefusalRiskCalculationInvalid, "ceil_minor", err)
	}
	return minor.String(), nil
}

func validateReservePolicy(policy ReservePolicy) (price, fx, haircut, fixedFee, perUnitFee, minimumFee *big.Rat, err error) {
	maxBits := policy.MaxDecimalBits
	if policy.AccountCurrency == "" || policy.QuoteCurrency == "" {
		err = refusal(RefusalCurrencyUnresolved, "currency", nil)
		return
	}
	if !policy.Price.Evidence.validAt(policy.EvaluatedAt) {
		err = refusal(RefusalWorstPriceUnavailable, "price_evidence", nil)
		return
	}
	price, parseErr := parseDecimal(policy.Price.WorstExecutableQuote, false, maxBits)
	if parseErr != nil {
		if errors.Is(parseErr, errDecimalOverflow) {
			err = refusal(RefusalRiskCalculationInvalid, "worst_executable_price", parseErr)
			return
		}
		err = refusal(RefusalWorstPriceUnavailable, "worst_executable_price", parseErr)
		return
	}
	if !policy.FX.Evidence.validAt(policy.EvaluatedAt) {
		err = refusal(RefusalCurrencyUnresolved, "fx_evidence", nil)
		return
	}
	fx, parseErr = parseDecimal(policy.FX.RateQuoteToBase, false, maxBits)
	if parseErr != nil {
		if errors.Is(parseErr, errDecimalOverflow) {
			err = refusal(RefusalRiskCalculationInvalid, "fx_rate", parseErr)
			return
		}
		err = refusal(RefusalCurrencyUnresolved, "fx_rate", parseErr)
		return
	}
	haircut, parseErr = parseDecimal(policy.FX.Haircut, false, maxBits)
	if parseErr != nil {
		if errors.Is(parseErr, errDecimalOverflow) {
			err = refusal(RefusalRiskCalculationInvalid, "fx_haircut", parseErr)
			return
		}
		err = refusal(RefusalInvalidFXHaircut, "fx_haircut", parseErr)
		return
	}
	if haircut.Cmp(big.NewRat(1, 1)) < 0 {
		err = refusal(RefusalInvalidFXHaircut, "fx_haircut", nil)
		return
	}
	if policy.AccountCurrency == policy.QuoteCurrency && (fx.Cmp(big.NewRat(1, 1)) != 0 || haircut.Cmp(big.NewRat(1, 1)) != 0) {
		err = refusal(RefusalCurrencyUnresolved, "same_currency_identity", nil)
		return
	}
	if policy.Fee.Version == "" || policy.Fee.Digest == "" {
		err = refusal(RefusalFeePolicyUnavailable, "fee_evidence", nil)
		return
	}
	fixedFee, parseErr = parseDecimal(policy.Fee.FixedBaseMinor, true, maxBits)
	if parseErr != nil {
		if errors.Is(parseErr, errDecimalOverflow) {
			err = refusal(RefusalRiskCalculationInvalid, "fixed_fee", parseErr)
			return
		}
		err = refusal(RefusalFeePolicyUnavailable, "fixed_fee", parseErr)
		return
	}
	perUnitFee, parseErr = parseDecimal(policy.Fee.PerUnitBaseMinor, true, maxBits)
	if parseErr != nil {
		if errors.Is(parseErr, errDecimalOverflow) {
			err = refusal(RefusalRiskCalculationInvalid, "per_unit_fee", parseErr)
			return
		}
		err = refusal(RefusalFeePolicyUnavailable, "per_unit_fee", parseErr)
		return
	}
	minimumFee, parseErr = parseDecimal(policy.Fee.MinimumBaseMinor, true, maxBits)
	if parseErr != nil {
		if errors.Is(parseErr, errDecimalOverflow) {
			err = refusal(RefusalRiskCalculationInvalid, "minimum_fee", parseErr)
			return
		}
		err = refusal(RefusalFeePolicyUnavailable, "minimum_fee", parseErr)
		return
	}
	return
}

func MaximumQuantity(remainingMinor string, upperBound uint64, policy ReservePolicy) (uint64, error) {
	remaining, err := parseMinor(remainingMinor, policy.MaxDecimalBits)
	if err != nil {
		return 0, refusal(RefusalRiskCalculationInvalid, "remaining_minor", err)
	}
	if upperBound == 0 {
		return 0, nil
	}
	if _, _, _, _, _, _, err := validateReservePolicy(policy); err != nil {
		return 0, err
	}
	var low uint64
	high := upperBound
	for low < high {
		mid := low + (high-low)/2
		if mid < high {
			mid++
		}
		reserve, err := ReservationMinor(mid, policy)
		if err != nil {
			return 0, err
		}
		reserveMinor, ok := new(big.Int).SetString(reserve, 10)
		if !ok {
			return 0, refusal(RefusalRiskCalculationInvalid, "reserve", fmt.Errorf("invalid internal result"))
		}
		if reserveMinor.Cmp(remaining) <= 0 {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return low, nil
}

package console

import (
	"errors"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// NeedsStopPctCorrection reports an engine-valid legacy fraction that the
// stricter human percentage control cannot save unchanged.
func (p settingsPage) NeedsStopPctCorrection() bool {
	fraction := p.Block.DefaultStopPct
	if math.IsNaN(fraction) || math.IsInf(fraction, 0) || fraction < 0.02 || fraction >= 1 {
		return false
	}
	_, err := parseStopPercent(fractionPercentText(fraction))
	return err != nil
}

// fractionPercentText converts a finite decimal fraction without exposing the
// binary floating-point tail in the operator-facing percentage.
func fractionPercentText(fraction float64) string {
	raw := strconv.FormatFloat(fraction, 'f', -1, 64)
	places := 0
	if dot := strings.IndexByte(raw, '.'); dot >= 0 {
		places = len(raw) - dot - 1 - 2
		if places < 0 {
			places = 0
		}
	}
	value, ok := new(big.Rat).SetString(raw)
	if !ok {
		return "5"
	}
	value.Mul(value, big.NewRat(100, 1))
	text := value.FloatString(places)
	if strings.Contains(text, ".") {
		text = strings.TrimRight(strings.TrimRight(text, "0"), ".")
	}
	return text
}

func parseStopPercent(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errors.New("값이 비어 있다")
	}
	percent, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(percent) || math.IsInf(percent, 0) {
		return 0, errors.New("유한한 숫자가 아니다")
	}
	exact, ok := new(big.Rat).SetString(raw)
	if !ok {
		return 0, errors.New("십진수로 해석할 수 없다")
	}
	if exact.Cmp(big.NewRat(2, 1)) < 0 || exact.Cmp(big.NewRat(20, 1)) > 0 {
		return 0, errors.New("2%에서 20% 사이여야 한다")
	}
	ticks := new(big.Rat).Mul(exact, big.NewRat(2, 1))
	if ticks.Denom().Cmp(big.NewInt(1)) != 0 {
		return 0, errors.New("0.5% 단위여야 한다")
	}
	return percent / 100, nil
}

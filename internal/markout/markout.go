// Package markout measures post-decision prices from observations a caller
// already owns. It has no transport, clock, store, or polling dependency.
package markout

import (
	"math/big"
	"regexp"
	"sort"
	"strings"
	"time"
)

const Tolerance = time.Minute

var Horizons = [...]time.Duration{5 * time.Minute, 15 * time.Minute, 30 * time.Minute}

type Status string

const (
	StatusMeasured    Status = "measured"
	StatusNotMeasured Status = "not_measured"
)

type Reason string

const (
	ReasonDecisionInstantUnknown       Reason = "DECISION_INSTANT_UNKNOWN"
	ReasonDecisionPriceAbsent          Reason = "DECISION_PRICE_ABSENT"
	ReasonDecisionPriceUnreadable      Reason = "DECISION_PRICE_UNREADABLE"
	ReasonDecisionPriceNotPositive     Reason = "DECISION_PRICE_NOT_POSITIVE"
	ReasonNoObservationWithinTolerance Reason = "NO_OBSERVATION_WITHIN_TOLERANCE"
	ReasonObservationPriceAbsent       Reason = "OBSERVATION_PRICE_ABSENT"
	ReasonObservationPriceUnreadable   Reason = "OBSERVATION_PRICE_UNREADABLE"
	ReasonObservationPriceNotPositive  Reason = "OBSERVATION_PRICE_NOT_POSITIVE"
)

type Observation struct {
	At    time.Time `json:"at"`
	Price string    `json:"price"`
}

type Measurement struct {
	Minutes       int        `json:"minutes"`
	TargetAt      time.Time  `json:"target_at"`
	Status        Status     `json:"status"`
	Reason        Reason     `json:"reason,omitempty"`
	ObservedAt    *time.Time `json:"observed_at,omitempty"`
	ObservedPrice string     `json:"observed_price,omitempty"`
	ReturnPct     string     `json:"return_pct,omitempty"`
}

type Report struct {
	DecisionAt    time.Time     `json:"decision_at"`
	DecisionPrice string        `json:"decision_price"`
	Measurements  []Measurement `json:"measurements"`
}

func (r Report) ForMinutes(minutes int) (Measurement, bool) {
	for _, measurement := range r.Measurements {
		if measurement.Minutes == minutes {
			return measurement, true
		}
	}
	return Measurement{}, false
}

// Measure selects the first existing observation at or after each target, with
// the +60 second boundary included. Missing or invalid data remains not_measured;
// it is never converted to a zero return.
func Measure(decisionAt time.Time, decisionPrice string, observations []Observation) Report {
	report := Report{DecisionAt: decisionAt.UTC(), DecisionPrice: strings.TrimSpace(decisionPrice)}
	rows := append([]Observation(nil), observations...)
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].At.Before(rows[j].At) })

	base, baseReason := price(decisionPrice, true)
	for _, horizon := range Horizons {
		target := decisionAt.Add(horizon)
		measurement := Measurement{
			Minutes: int(horizon / time.Minute), TargetAt: target.UTC(), Status: StatusNotMeasured,
		}
		switch {
		case decisionAt.IsZero():
			measurement.Reason = ReasonDecisionInstantUnknown
		case baseReason != "":
			measurement.Reason = baseReason
		default:
			observation, found := firstWithin(rows, target, target.Add(Tolerance))
			if !found {
				measurement.Reason = ReasonNoObservationWithinTolerance
				break
			}
			observedAt := observation.At.UTC()
			measurement.ObservedAt = &observedAt
			measurement.ObservedPrice = strings.TrimSpace(observation.Price)
			observed, why := price(observation.Price, false)
			if why != "" {
				measurement.Reason = why
				break
			}
			gain := new(big.Rat).Sub(observed, base)
			gain.Quo(gain, base)
			gain.Mul(gain, big.NewRat(100, 1))
			measurement.Status = StatusMeasured
			measurement.ReturnPct = decimal(gain)
		}
		report.Measurements = append(report.Measurements, measurement)
	}
	return report
}

func firstWithin(rows []Observation, target, limit time.Time) (Observation, bool) {
	for _, row := range rows {
		if row.At.Before(target) {
			continue
		}
		if row.At.After(limit) {
			return Observation{}, false
		}
		return row, true
	}
	return Observation{}, false
}

var decimalPattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)$`)

func price(raw string, decision bool) (*big.Rat, Reason) {
	value := strings.TrimSpace(raw)
	if value == "" {
		if decision {
			return nil, ReasonDecisionPriceAbsent
		}
		return nil, ReasonObservationPriceAbsent
	}
	if !decimalPattern.MatchString(value) {
		if decision {
			return nil, ReasonDecisionPriceUnreadable
		}
		return nil, ReasonObservationPriceUnreadable
	}
	parsed, ok := new(big.Rat).SetString(value)
	if !ok {
		if decision {
			return nil, ReasonDecisionPriceUnreadable
		}
		return nil, ReasonObservationPriceUnreadable
	}
	if parsed.Sign() <= 0 {
		if decision {
			return nil, ReasonDecisionPriceNotPositive
		}
		return nil, ReasonObservationPriceNotPositive
	}
	return parsed, ""
}

func decimal(value *big.Rat) string {
	text := value.FloatString(12)
	text = strings.TrimRight(text, "0")
	text = strings.TrimRight(text, ".")
	if text == "-0" || text == "" {
		return "0"
	}
	return text
}

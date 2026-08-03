package positioncampaign

import (
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

type StopSelection string

const (
	StopFromSaved     StopSelection = "SAVED"
	StopFromCandidate StopSelection = "CANDIDATE"
)

type EffectiveStop struct {
	Price        string
	Source       string
	Policy       string
	ObservedAt   string
	SelectedFrom StopSelection
	Candidate    StopCandidate
}

type StopCandidate struct {
	Price      string
	Valid      bool
	Source     string
	Policy     string
	ObservedAt string
}

// ComposeLongStop selects max(saved,candidate). Missing or invalid candidate
// blocks additional exposure but never clears or lowers an existing stop.
func ComposeLongStop(saved *EffectiveStop, candidate StopCandidate) (EffectiveStop, bool, error) {
	var prior EffectiveStop
	if saved != nil {
		prior = *saved
		if strings.TrimSpace(prior.Price) != "" {
			price, err := nonNegative(prior.Price)
			if err != nil {
				return EffectiveStop{}, true, fmt.Errorf("saved stop: %w", err)
			}
			prior.Price = price
			prior.SelectedFrom = StopFromSaved
		}
	}
	if !candidate.Valid || strings.TrimSpace(candidate.Price) == "" ||
		strings.TrimSpace(candidate.Source) == "" || strings.TrimSpace(candidate.Policy) == "" ||
		strings.TrimSpace(candidate.ObservedAt) == "" {
		prior.Candidate = candidate
		return prior, true, nil
	}
	price, err := positiveDecimal(candidate.Price)
	if err != nil {
		prior.Candidate = candidate
		return prior, true, nil
	}
	if prior.Price != "" {
		cmp, err := riskcalc.CompareDecimal(price, prior.Price)
		if err != nil {
			return EffectiveStop{}, true, err
		}
		if cmp <= 0 {
			prior.Candidate = candidate
			return prior, false, nil
		}
	}
	return EffectiveStop{
		Price: price, Source: candidate.Source, Policy: candidate.Policy,
		ObservedAt: candidate.ObservedAt, SelectedFrom: StopFromCandidate,
		Candidate: candidate,
	}, false, nil
}

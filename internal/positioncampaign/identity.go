package positioncampaign

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidIdentity = errors.New("position campaign: invalid aggregate identity")

// PositionCampaign is the immutable aggregate identity. Mutable projections
// (state, quantities and watermarks) are intentionally not part of it.
type PositionCampaign struct {
	ID                         string
	AccountRef                 string
	Market                     string
	Symbol                     string
	LaneID                     string
	LaneVersion                string
	DecisionID                 string
	EvidenceDigest             string
	ProspectiveToken           string
	ExpectedPositionGeneration int64
	ActualPositionGeneration   int64
	Legs                       []CampaignLeg
}

// CampaignLeg carries the immutable ordered plan-to-execution lineage. Intent
// and attempt identifiers are set before a broker order identity is linked and
// cannot subsequently be replaced by another identity.
type CampaignLeg struct {
	CampaignID string
	Sequence   int64
	PlanID     string
	IntentID   string
	AttemptID  string
}

func (c PositionCampaign) Validate() error {
	for name, value := range map[string]string{
		"campaign id": c.ID, "account": c.AccountRef, "market": c.Market,
		"symbol": c.Symbol, "lane id": c.LaneID, "lane version": c.LaneVersion,
		"decision id": c.DecisionID, "evidence digest": c.EvidenceDigest,
		"prospective token": c.ProspectiveToken,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: %s is empty", ErrInvalidIdentity, name)
		}
	}
	if c.ExpectedPositionGeneration < 0 || c.ActualPositionGeneration < 0 {
		return fmt.Errorf("%w: negative position generation", ErrInvalidIdentity)
	}
	if c.ActualPositionGeneration != 0 && c.ActualPositionGeneration != c.ExpectedPositionGeneration+1 {
		return fmt.Errorf("%w: actual generation %d is not expected successor %d",
			ErrInvalidIdentity, c.ActualPositionGeneration, c.ExpectedPositionGeneration+1)
	}
	seenPlan := make(map[string]struct{}, len(c.Legs))
	for i, leg := range c.Legs {
		if err := leg.Validate(); err != nil {
			return err
		}
		if leg.CampaignID != c.ID || leg.Sequence != int64(i+1) {
			return fmt.Errorf("%w: leg %d belongs to %q at sequence %d",
				ErrInvalidIdentity, i+1, leg.CampaignID, leg.Sequence)
		}
		if _, duplicate := seenPlan[leg.PlanID]; duplicate {
			return fmt.Errorf("%w: duplicate plan id %q", ErrInvalidIdentity, leg.PlanID)
		}
		seenPlan[leg.PlanID] = struct{}{}
	}
	return nil
}

func (l CampaignLeg) Validate() error {
	for name, value := range map[string]string{
		"campaign id": l.CampaignID, "plan id": l.PlanID,
		"intent id": l.IntentID, "attempt id": l.AttemptID,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%w: leg %s is empty", ErrInvalidIdentity, name)
		}
	}
	if l.Sequence <= 0 {
		return fmt.Errorf("%w: leg sequence must be positive", ErrInvalidIdentity)
	}
	return nil
}

func (l CampaignLeg) SameIdentity(other CampaignLeg) bool {
	return l.CampaignID == other.CampaignID && l.Sequence == other.Sequence &&
		l.PlanID == other.PlanID && l.IntentID == other.IntentID && l.AttemptID == other.AttemptID
}

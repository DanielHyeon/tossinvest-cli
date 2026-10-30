package execgw

import (
	"context"
	"errors"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// replayOrigin는 원장에 기록된 decision과 정확한 intent 출처만 분류함.
func (g *Gateway) replayOrigin(ctx context.Context, rec journal.AttemptRecord, intent journal.Intent) (accountbudget.Priority, bool, error) {
	if rec.IntentID == "" || rec.IntentID != intent.ID || rec.AccountRef == "" || rec.AccountRef != intent.AccountRef || rec.AccountRef != g.accountRef || rec.DecisionID == "" || rec.ClientOrderID == "" || (intent.Market != "kr" && intent.Market != "us") || strings.TrimSpace(intent.Symbol) == "" {
		return accountbudget.Normal, false, nil
	}
	dec, err := g.journal.LookupDecision(ctx, rec.DecisionID)
	if errors.Is(err, journal.ErrDecisionNotFound) {
		return accountbudget.Normal, false, nil
	}
	if err != nil {
		return accountbudget.Normal, false, err
	}
	if dec.AccountRef != rec.AccountRef || dec.Generation != rec.Generation || dec.SafetyClass != rec.SafetyClass || dec.ClientOrderID != rec.ClientOrderID {
		return accountbudget.Normal, false, nil
	}
	if journal.HashPreimage(dec.RiskPreimage) != dec.RiskHash {
		return accountbudget.Normal, false, nil
	}
	preimage, err := journal.ParsePreimage(dec.PreimageKind, dec.RiskPreimage)
	if err != nil {
		return accountbudget.Normal, false, nil
	}
	account, market, symbol, side := journal.PreimageVenue(preimage)
	if account != intent.AccountRef || market != intent.Market || symbol != strings.ToUpper(strings.TrimSpace(intent.Symbol)) || side != intent.Side {
		return accountbudget.Normal, false, nil
	}
	switch {
	case intent.Source == "engine" && intent.Side == "BUY" && rec.SafetyClass == journal.SafetyClassExposureRaising && dec.PreimageKind == journal.PreimageKindRiskIntent:
		conflict, err := g.journal.HasExitEventForIntent(ctx, intent.ID)
		return accountbudget.Normal, !conflict && err == nil, err
	case intent.Source == "flatten-all" && intent.Side == "SELL" && rec.SafetyClass == journal.SafetyClassRiskReducing && dec.PreimageKind == journal.PreimageKindReductionIntent:
		conflict, err := g.journal.HasExitEventForIntent(ctx, intent.ID)
		return accountbudget.Critical, !conflict && err == nil, err
	case intent.Source == "engine" && intent.Side == "SELL" && rec.SafetyClass == journal.SafetyClassRiskReducing && dec.PreimageKind == journal.PreimageKindReductionIntent:
		action, ok, err := g.journal.ExitProposalForIntent(ctx, intent)
		if err != nil || !ok {
			return accountbudget.Normal, false, err
		}
		switch exitpolicy.Action(action) {
		case exitpolicy.ActionBaselineBreach, exitpolicy.ActionLadderStop:
			return accountbudget.Critical, true, nil
		case exitpolicy.ActionRatchetPartial, exitpolicy.ActionLadderPartial, exitpolicy.ActionLadderTakeProfit:
			return accountbudget.Normal, true, nil
		}
	}
	return accountbudget.Normal, false, nil
}

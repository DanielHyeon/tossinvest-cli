package engine

import (
	"context"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

func strategyProtectionMatchesExit(binding journal.StrategyProtection, p journal.Position, state journal.ExitState) bool {
	return binding.Policy != nil && !state.Completed && state.PositionGeneration == p.InstanceSeq &&
		state.PolicyKind == binding.Policy.Kind && state.PolicyIdentity == binding.Policy.Identity &&
		state.EntryPrice == binding.Risk.EntryPrice && state.InitialStop == binding.Risk.StopPrice
}

func (o *ExitObserver) queueStrategyProtectionAlert(p journal.Position, cause error, existing, blockFailed bool) {
	detail := "Admitted protection could not be reconciled; additional entry must remain blocked."
	if blockFailed {
		detail += " Durable entry block could not be recorded; retry and reconcile."
	}
	if existing {
		detail += " Existing exit remains under evaluation; admitted floor enforcement is unconfirmed until a compatible judgement commits."
	} else {
		detail += " The position may lack a protective exit."
	}
	o.logErr(obs.EventType("exit.strategy_protection_reconcile"), cause, detail)
	o.protectionAlerts = append(o.protectionAlerts, obs.Event{Type: obs.EventType("exit.strategy_protection_reconcile"), Key: "strategy-protection|" + p.ID, Title: "Strategy protection requires reconciliation", Body: detail, Fields: map[string]any{"position_id": p.ID, obs.FieldSymbol: p.Symbol}})
}

func (o *ExitObserver) flushStrategyProtectionAlerts(ctx context.Context) {
	if o.opts.Critical == nil {
		return
	}
	for _, event := range o.protectionAlerts {
		if err := o.opts.Critical.RecordCritical(ctx, event, 0); err != nil {
			o.logErr(event.Type, err, "strategy protection alarm could not be recorded")
		}
	}
}

package execgw

import (
	"context"
	"errors"

	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

type enginePollBinding struct {
	budget     *accountbudget.Budget
	client     *official.Client
	accountRef string
	seq        int
}

// NewEnginePollRetrier copies the existing retrier and binds an observed budget.
func NewEnginePollRetrier(base *Retrier, budget *accountbudget.Budget, client *official.Client, accountRef string, seq int) (*Retrier, error) {
	if base == nil || budget == nil || client == nil || seq <= 0 || accountRef == "" || base.AccountRef != accountRef || !client.UsesAccountBudget(budget) {
		return nil, accountbudget.ErrUnavailable
	}
	selected, ok := client.SelectedAccountSeq()
	if !ok || selected != seq {
		return nil, accountbudget.ErrUnavailable
	}
	if _, err := budget.EnginePollKey(accountRef, string(QueryPrice)); err != nil {
		return nil, err
	}
	copy := *base
	copy.poll = &enginePollBinding{budget: budget, client: client, accountRef: accountRef, seq: seq}
	return &copy, nil
}

// EnginePollReaderAllowed accepts arbitrary OFF readers; required uses the exact client.
func (r *Retrier) EnginePollReaderAllowed(reader any) bool {
	if r == nil || r.poll == nil {
		return true
	}
	client, ok := reader.(*official.Client)
	return ok && client != nil && client == r.poll.client
}

func (r *Retrier) pollContext(ctx context.Context, kind RequiredQuery) (context.Context, *accountbudget.ObservationCollector, error) {
	if r.poll == nil {
		return ctx, nil, nil
	}
	p := r.poll
	if r.AccountRef != p.accountRef || p.client == nil || !p.client.UsesAccountBudget(p.budget) {
		return nil, nil, accountbudget.ErrUnavailable
	}
	selected, ok := p.client.SelectedAccountSeq()
	if !ok || selected != p.seq {
		return nil, nil, accountbudget.ErrUnavailable
	}
	switch kind {
	case QueryOpenOrders, QueryBuyingPower, QueryHoldings, QueryPrice:
	default:
		return nil, nil, accountbudget.ErrUnavailable
	}
	key, err := p.budget.EnginePollKey(p.accountRef, string(kind))
	if err != nil {
		return nil, nil, err
	}
	remaining, err := p.budget.PollRemaining(ctx, key)
	if err != nil {
		return nil, nil, err
	}
	if remaining > 0 {
		return nil, nil, accountbudget.ErrObservationDeferred
	}
	collector := &accountbudget.ObservationCollector{}
	return accountbudget.WithObservation(ctx, accountbudget.ObservationPoll, key, collector), collector, nil
}

func pollError(c *accountbudget.ObservationCollector, err error) error {
	if c == nil {
		return nil
	}
	observed := c.Snapshot()
	if observed.RateLimited || ClassifyQueryError(err) == ClassRateLimited {
		return errors.Join(err, official.ErrRateLimited, observed.LocalError)
	}
	if observed.LocalError != nil {
		return errors.Join(err, observed.LocalError)
	}
	return nil
}

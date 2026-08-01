package protection

import "fmt"

type DiscrepancyKind string

const (
	DiscrepancyMissing          DiscrepancyKind = "MISSING"
	DiscrepancyDuplicate        DiscrepancyKind = "DUPLICATE"
	DiscrepancyOrphan           DiscrepancyKind = "ORPHAN"
	DiscrepancyQuantityMismatch DiscrepancyKind = "QUANTITY_MISMATCH"
	DiscrepancyTriggerMismatch  DiscrepancyKind = "TRIGGER_MISMATCH"
)

type Discrepancy struct {
	Kind     DiscrepancyKind
	Scope    Scope
	SagaID   string
	BrokerID string
}

// Compare is classification only. It never guesses ownership, cancels an
// orphan, replaces an order, or flattens a position.
func Compare(scope Scope, local []Saga, broker []BrokerProtection) ([]Discrepancy, error) {
	if err := scope.Validate(); err != nil {
		return nil, err
	}
	localByBroker := make(map[string]Saga, len(local))
	var relevantLocal []Saga
	for _, s := range local {
		if err := s.Validate(); err != nil {
			return nil, err
		}
		sagaScope := Scope{AccountRef: s.AccountRef, Profile: s.Profile, Market: s.Market, Symbol: s.Symbol}
		if !sagaScope.equal(scope) {
			return nil, fmt.Errorf("%w: local saga %s", ErrMixedScope, s.ID)
		}
		if s.State != StateActive && s.State != StateReplacing {
			continue
		}
		if s.BrokerID != "" {
			if _, exists := localByBroker[s.BrokerID]; exists {
				return nil, fmt.Errorf("%w: local broker id %s", ErrDuplicateBrokerID, s.BrokerID)
			}
			localByBroker[s.BrokerID] = s
		}
		relevantLocal = append(relevantLocal, s)
	}
	brokerByID := make(map[string]BrokerProtection, len(broker))
	seenBrokerIDs := make(map[string]bool, len(broker))
	for _, b := range broker {
		if !b.Scope.equal(scope) {
			return nil, fmt.Errorf("%w: broker protection %s", ErrMixedScope, b.ID)
		}
		if b.ID == "" {
			return nil, fmt.Errorf("%w: empty broker id", ErrDuplicateBrokerID)
		}
		if seenBrokerIDs[b.ID] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateBrokerID, b.ID)
		}
		seenBrokerIDs[b.ID] = true
		if b.Quantity < 1 || b.Trigger < 1 {
			return nil, fmt.Errorf("%w: broker protection %s has invalid quantity/trigger", ErrInvalidSaga, b.ID)
		}
		if b.Terminal {
			continue
		}
		brokerByID[b.ID] = b
	}

	var out []Discrepancy
	for id, s := range localByBroker {
		b, ok := brokerByID[id]
		if !ok {
			out = append(out, Discrepancy{Kind: DiscrepancyMissing, Scope: scope, SagaID: s.ID, BrokerID: id})
			continue
		}
		if b.Quantity != s.Quantity {
			out = append(out, Discrepancy{Kind: DiscrepancyQuantityMismatch, Scope: scope, SagaID: s.ID, BrokerID: id})
		}
		if b.Trigger != s.Trigger {
			out = append(out, Discrepancy{Kind: DiscrepancyTriggerMismatch, Scope: scope, SagaID: s.ID, BrokerID: id})
		}
	}
	for _, b := range broker {
		if b.Terminal {
			continue
		}
		if _, ok := localByBroker[b.ID]; !ok {
			out = append(out, Discrepancy{Kind: DiscrepancyOrphan, Scope: scope, BrokerID: b.ID})
		}
	}
	activeBrokerCount := len(brokerByID)
	if activeBrokerCount > 1 && len(relevantLocal) > 0 {
		out = append(out, Discrepancy{Kind: DiscrepancyDuplicate, Scope: scope, SagaID: relevantLocal[0].ID})
	}
	return out, nil
}

package protection

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
	SagaID   string
	BrokerID string
	Symbol   string
}

// Compare is classification only. It never guesses ownership, cancels an
// orphan, replaces an order, or flattens a position.
func Compare(local []Saga, broker []BrokerProtection) []Discrepancy {
	localByBroker := make(map[string]Saga, len(local))
	localBySymbol := make(map[string][]Saga, len(local))
	for _, s := range local {
		if s.State != StateActive && s.State != StateReplacing {
			continue
		}
		if s.BrokerID != "" {
			localByBroker[s.BrokerID] = s
		}
		localBySymbol[s.Symbol] = append(localBySymbol[s.Symbol], s)
	}
	brokerByID := make(map[string]BrokerProtection, len(broker))
	brokerBySymbol := make(map[string][]BrokerProtection, len(broker))
	for _, b := range broker {
		if b.Terminal {
			continue
		}
		brokerByID[b.ID] = b
		brokerBySymbol[b.Symbol] = append(brokerBySymbol[b.Symbol], b)
	}

	var out []Discrepancy
	for id, s := range localByBroker {
		b, ok := brokerByID[id]
		if !ok {
			out = append(out, Discrepancy{Kind: DiscrepancyMissing, SagaID: s.ID, BrokerID: id, Symbol: s.Symbol})
			continue
		}
		if b.Quantity != s.Quantity {
			out = append(out, Discrepancy{Kind: DiscrepancyQuantityMismatch, SagaID: s.ID, BrokerID: id, Symbol: s.Symbol})
		}
		if b.Trigger != s.Trigger {
			out = append(out, Discrepancy{Kind: DiscrepancyTriggerMismatch, SagaID: s.ID, BrokerID: id, Symbol: s.Symbol})
		}
	}
	for _, b := range broker {
		if b.Terminal {
			continue
		}
		if _, ok := localByBroker[b.ID]; !ok {
			out = append(out, Discrepancy{Kind: DiscrepancyOrphan, BrokerID: b.ID, Symbol: b.Symbol})
		}
	}
	for symbol, orders := range brokerBySymbol {
		if len(orders) > 1 && len(localBySymbol[symbol]) > 0 {
			out = append(out, Discrepancy{Kind: DiscrepancyDuplicate, SagaID: localBySymbol[symbol][0].ID, Symbol: symbol})
		}
	}
	return out
}

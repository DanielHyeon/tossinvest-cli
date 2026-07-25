package filldetect

// payload.go turns one raw broker order into a Snapshot.
//
// It is a separate reader from internal/brokerstate's and internal/execgw's
// because it needs a third slice of the same JSON: brokerstate reads the fields
// the state table judges, execgw reads the fields the IN_DOUBT matcher compares,
// and this reads the fields the fill ledger and the SLO need —
// execution.filledAt above all, which is the near end of the measurement and is
// carried by no other type in the codebase.
//
// Everything unreadable is an error, never a zero. A quantity that silently
// parses to 0 would show up downstream as "nothing filled", which is the one
// wrong answer that looks like good news.

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrMalformedOrder means a broker order payload could not be read.
var ErrMalformedOrder = errors.New("filldetect: malformed broker order payload")

// officialOrder mirrors the subset of the official Order schema this package
// reads. Field names follow internal/official.apiOrder.
type officialOrder struct {
	OrderID    string  `json:"orderId"`
	Symbol     string  `json:"symbol"`
	Status     string  `json:"status"`
	Quantity   *string `json:"quantity"`
	Currency   string  `json:"currency"`
	CanceledAt *string `json:"canceledAt"`
	Execution  *struct {
		FilledQuantity     *string `json:"filledQuantity"`
		AverageFilledPrice *string `json:"averageFilledPrice"`
		FilledAt           *string `json:"filledAt"`
	} `json:"execution"`
}

// timeLayouts are tried in order. RFC3339 is what the API documents; the rest
// exist so a format surprise degrades into "no broker timestamp" (and therefore
// the conservative fallback) rather than into a parse failure that stops the loop.
var timeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
}

// parseSnapshot reads one raw order, with or without the {"result": …} envelope.
//
// observedAt is when this cycle read it. fallbackVisible is the instant used as
// BrokerVisibleAt when the payload carries no execution.filledAt.
func parseSnapshot(raw []byte, observedAt, fallbackVisible time.Time) (Snapshot, error) {
	body := unwrapResult(raw)

	var payload officialOrder
	if err := json.Unmarshal(body, &payload); err != nil {
		return Snapshot{}, fmt.Errorf("%w: %v", ErrMalformedOrder, err)
	}

	quantity, err := parseDecimal(payload.Quantity)
	if err != nil {
		return Snapshot{}, fmt.Errorf("%w: quantity: %v", ErrMalformedOrder, err)
	}
	var (
		filled   float64
		avgPrice string
		filledAt *time.Time
	)
	if payload.Execution != nil {
		if filled, err = parseDecimal(payload.Execution.FilledQuantity); err != nil {
			return Snapshot{}, fmt.Errorf("%w: execution.filledQuantity: %v", ErrMalformedOrder, err)
		}
		avgPrice = trimPtr(payload.Execution.AverageFilledPrice)
		filledAt = parseTime(payload.Execution.FilledAt)
	}

	snap := Snapshot{
		OrderID:         strings.TrimSpace(payload.OrderID),
		Symbol:          strings.ToUpper(strings.TrimSpace(payload.Symbol)),
		RawStatus:       payload.Status,
		Quantity:        quantity,
		FilledQuantity:  filled,
		AveragePrice:    avgPrice,
		ObservedAt:      observedAt,
		BrokerVisibleAt: fallbackVisible,
	}
	if filledAt != nil {
		snap.BrokerVisibleAt = *filledAt
		snap.HasBrokerTime = true
	}
	if ts := trimPtr(payload.CanceledAt); ts != "" {
		// Set independently of a successful parse: an unreadable timestamp is
		// still evidence of a cancellation, and losing that evidence would turn a
		// cancel into a fill in the state table.
		snap.Canceled = true
		snap.CanceledAt = parseTime(payload.CanceledAt)
	}
	return snap, nil
}

func unwrapResult(data []byte) []byte {
	var envelope struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return data
	}
	if len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return data
	}
	return envelope.Result
}

// parseDecimal converts a nullable decimal string. Null and empty are 0 (that is
// what the API means for a pending order's execution); anything non-numeric is an
// error rather than a silent zero.
func parseDecimal(s *string) (float64, error) {
	trimmed := trimPtr(s)
	if trimmed == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not a decimal", trimmed)
	}
	return v, nil
}

func parseTime(s *string) *time.Time {
	trimmed := trimPtr(s)
	if trimmed == "" {
		return nil
	}
	for _, layout := range timeLayouts {
		if ts, err := time.Parse(layout, trimmed); err == nil {
			utc := ts.UTC()
			return &utc
		}
	}
	return nil
}

func trimPtr(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

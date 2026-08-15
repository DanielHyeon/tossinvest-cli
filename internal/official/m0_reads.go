package official

import (
	"context"
	"encoding/json"
)

// M0ReadSource binds raw trigger reads to the Client that performed them.
// Its fields are private, so callers cannot construct evidence for arbitrary
// bytes or inject an attempt independently of a request.
type M0ReadSource struct{ client *Client }

// M0ReadSourceFor accepts only the same concrete official client that performs
// M0 mutations. It intentionally rejects wrappers and split sources: accepting
// a separately supplied client would let a caller combine one account's POST
// with another account's authenticated read evidence.
func M0ReadSourceFor(v any) (*M0ReadSource, bool) {
	if client, ok := v.(*Client); ok && client != nil {
		return &M0ReadSource{client: client}, true
	}
	return nil, false
}

func (s *M0ReadSource) ConditionalOrderRaw(ctx context.Context, id string) (RawConditionalOrder, []AttemptTrace, error) {
	var attempts []AttemptTrace
	ctx = WithAttemptObserver(ctx, func(trace AttemptTrace) { attempts = append(attempts, trace) })
	value, err := s.client.ConditionalOrderRaw(ctx, id)
	return value, attempts, err
}

func (s *M0ReadSource) OrderRawByID(ctx context.Context, id string) (json.RawMessage, []AttemptTrace, error) {
	var attempts []AttemptTrace
	ctx = WithAttemptObserver(ctx, func(trace AttemptTrace) { attempts = append(attempts, trace) })
	value, err := s.client.OrderRawByID(ctx, id)
	return value, attempts, err
}

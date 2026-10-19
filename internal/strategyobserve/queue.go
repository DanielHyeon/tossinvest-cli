package strategyobserve

import "sync"

type Outcome string

const (
	Accepted Outcome = "accepted"
	Oversize Outcome = "oversize"
	Invalid  Outcome = "invalid"
	Full     Outcome = "queue_full"
	Closed   Outcome = "closed"
	Failed   Outcome = "writer_failed"
)

type Counters struct{ Accepted, Oversize, Invalid, Full, Closed, Failed, Durable, Duplicate, Pending uint64 }
type Queue struct {
	mu             sync.Mutex
	ch             chan []byte
	closed, failed bool
	counts         Counters
}

func NewQueue() *Queue { return &Queue{ch: make(chan []byte, 256)} }
func (q *Queue) Enqueue(b []byte) Outcome {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		q.counts.Closed++
		return Closed
	}
	if q.failed {
		q.counts.Failed++
		return Failed
	}
	if len(b) > MaxRecord {
		q.counts.Oversize++
		return Oversize
	}
	if _, err := ValidateLine(b); err != nil {
		q.counts.Invalid++
		return Invalid
	}
	copyBytes := append([]byte(nil), b...)
	select {
	case q.ch <- copyBytes:
		q.counts.Accepted++
		q.counts.Pending++
		return Accepted
	default:
		q.counts.Full++
		return Full
	}
}
func (q *Queue) CloseInput() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.closed {
		q.closed = true
		close(q.ch)
	}
}
func (q *Queue) Counts() Counters { q.mu.Lock(); defer q.mu.Unlock(); return q.counts }
func (q *Queue) Run(appendRecord func([]byte) (bool, error)) error {
	for b := range q.ch {
		duplicate, err := appendRecord(b)
		q.mu.Lock()
		if err != nil {
			q.failed = true
			q.counts.Failed++
			q.mu.Unlock()
			return err
		}
		q.counts.Pending--
		if duplicate {
			q.counts.Duplicate++
		} else {
			q.counts.Durable++
		}
		q.mu.Unlock()
	}
	return nil
}

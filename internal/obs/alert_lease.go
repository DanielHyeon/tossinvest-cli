package obs

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// DefaultPublishTimeout bounds one publish when a Publisher has no timeout of
// its own.
//
// It was a bare literal inside Ntfy.Publish until a099. The lease a sender holds
// on an alert row has to outlast that sender's whole cycle, and this value is a
// term in that sum — a number nobody can read from outside the function that
// uses it is a number the derivation has to copy, and a copied number is one
// that stops matching.
const DefaultPublishTimeout = 10 * time.Second

// AlertDeliveryBound is how long one sender can hold an alert before it is
// certainly finished with it.
//
// It bounds a *cycle*, not an attempt. A sender claims a row once and then
// spends its entire retry budget under that one claim, so the bound is every
// attempt, every wait between them, and every ledger write those attempts make:
//
//	attempts × (publish timeout + write wait)   every attempt times out, and each
//	                                            failure is recorded
//	+ (attempts−1) × retry delay                the waits in between
//	+ write wait                                the release at the end
//
// Every term is the worst case. That is the point: a lease shorter than this
// lets a sender lose its row while it is still working through a budget this
// code told it to spend, and then two senders publish while nothing anywhere has
// gone wrong.
//
// The four inputs are parameters rather than reads of the defaults so that a
// test can move each one and watch the result move. A derivation nobody can
// perturb is indistinguishable from a constant.
func AlertDeliveryBound(attempts int, publishTimeout, retryDelay, writeWait time.Duration) time.Duration {
	if attempts <= 0 {
		attempts = DefaultCriticalAttempts
	}
	if publishTimeout <= 0 {
		publishTimeout = DefaultPublishTimeout
	}
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}
	if writeWait <= 0 {
		writeWait = journal.DefaultBusyTimeout
	}
	return time.Duration(attempts)*(publishTimeout+writeWait) +
		time.Duration(attempts-1)*retryDelay +
		writeWait
}

// DefaultAlertDeliveryBound is AlertDeliveryBound for the configuration a
// Notifier runs with when nothing is set. With today's defaults it is 54s.
//
// It is exported because journal.DefaultAlertLease has to exceed it and journal
// cannot import this package — the dependency runs the other way. The only thing
// holding the two together is a test that reads both, which is why this has to
// be readable from outside.
func DefaultAlertDeliveryBound() time.Duration {
	return AlertDeliveryBound(DefaultCriticalAttempts, DefaultPublishTimeout,
		DefaultRetryDelay, journal.DefaultBusyTimeout)
}

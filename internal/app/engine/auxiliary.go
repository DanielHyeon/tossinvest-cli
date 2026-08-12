package engine

// auxiliary.go is a098's answer to "이 일을 감독 집합에 넣을 수는 없다".
//
// The runtime's supervision contract is deliberately blunt: a supervised loop
// that returns for any reason other than cancellation brings the whole engine
// down (runtime.go, 방어적 종료 계약). That is right for the loops it was written
// for — an engine reconciling without observing exits is worse than one that
// stopped.
//
// Alert delivery is not that kind of work. A delivery fault must not stop the
// loop that places a stop-loss (안전 불변식 4), so putting the delivery executor
// in Loops would turn an alert bug into a missing stop.
//
// # Why a second slice rather than a flag on SupervisedLoop
//
// Because a flag is a judgement, and a judgement can be wrong. Registering the
// executor as a supervised loop and then filtering it out by name at the moment
// it stops leaves a rule that has to be right every time it is read. Nothing
// started from Auxiliary sends to the stops channel, so there is no judgement to
// get wrong: it cannot become the first stop because it never speaks on that
// channel (engine-safety, 「배달 실행자의 정지가 다른 루프를 내려서는 안 된다」).

import "context"

// AuxiliaryExecutor is work the runtime starts and drains but does not supervise.
//
// "Not supervised" is a statement about two things the runtime will not do: it
// will not stop the other loops when this one stops, and it will not put this
// one on the degradation ladder (superviseHealth walks Loops only, and a health
// threshold on top of an unsupervised executor would be a ladder to nowhere).
//
// It is not a statement that the runtime forgets about it. The runtime waits for
// this executor before it returns, because the caller closes the journal the
// moment Run does and this executor writes to that journal.
type AuxiliaryExecutor struct {
	// Name identifies the executor in the record of its stop. Required, and it
	// has to be unique across the loops as well — a stop record that could mean
	// either of two things names neither.
	Name string
	// Run is the work. Required. Like a supervised loop it is expected to return
	// only when ctx ends, and unlike a supervised loop that expectation being
	// wrong is not the engine's problem to solve by dying.
	Run func(ctx context.Context) error
}

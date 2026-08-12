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

import (
	"context"
	"fmt"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// alertDeliveryName is what the delivery executor is called wherever its stop is
// recorded. It is not a supervised loop's name and must never appear in
// Runtime.LoopNames.
const alertDeliveryName = "alert-delivery"

// ErrAuxiliaryPanicked is an executor that panicked, turned into a return.
//
// A panic is a stop like any other here, and it has to become one *before* it
// reaches the goroutine's own stack unwinding: an unrecovered panic in any
// goroutine kills the process, and this process is the one holding the stop-loss
// loop (안전 불변식 4).
var ErrAuxiliaryPanicked = fmt.Errorf("engine: an auxiliary executor panicked")

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
	//
	// It has to report a cancellation *as* a cancellation — return ctx.Err(),
	// not nil. That is not a style preference: runAuxiliary judges the stop with
	// Runtime.gracefulStop, which requires the error to be the context's, so an
	// executor returning nil on cancellation has every ordinary shutdown read as
	// a death.
	Run func(ctx context.Context) error
	// OnStop is called when Run stops for a reason that is not the runtime being
	// cancelled — including a panic, which arrives here as ErrAuxiliaryPanicked.
	// Optional, and a nil one is the "nobody is watching" arrangement the
	// engine-safety delta forbids for the delivery executor.
	//
	// It is not called on a graceful stop. Blocking entry every time somebody
	// pressed Ctrl-C would make the next start demand an operator's approval for
	// nothing.
	OnStop func(ctx context.Context, err error)
}

// runAuxiliary runs one executor and turns whatever it does into a stop.
//
// It lives here rather than inline in Runtime.Run because Run's shape is the
// evidence that this change did not touch the stop policy: five branches, three
// returns, and a recover/judge/notify block inline would have made both grow.
func (r *Runtime) runAuxiliary(ctx context.Context, aux AuxiliaryExecutor) {
	err := runAuxiliaryBody(ctx, aux)
	if r.gracefulStop(ctx, err) {
		// The runtime was cancelled and this executor said so. Normal.
		//
		// ⛔ ctx here is the runtime's loopCtx, not its parent, and the two differ
		// in exactly the case that matters: when a *supervised* loop fails, Run
		// cancels loopCtx while the parent stays live. Judging against the parent
		// would read this executor's correct shutdown as a death and latch the
		// gate on the way out of an unrelated failure.
		return
	}
	// ⛔ Deliberately not EventEngineLoopFailed. That event means "the engine is
	// going down", and the runtime raises a critical alert with it; a filter on
	// it must not also pick up stops that stopped nothing. This one is the
	// alert-delivery event because alert delivery is the only auxiliary this
	// runtime has — a second one would have to bring its own event type rather
	// than borrow this, and internal/obs is closed to this change (결정 5-1) so
	// there was no third option to pick.
	r.log(obs.EventAlertUndelivered, true,
		"loop", aux.Name,
		obs.FieldReason, describe(err),
		obs.FieldDetail, "an auxiliary executor stopped and will not be restarted")
	if aux.OnStop == nil {
		return
	}
	aux.OnStop(ctx, err)
}

// runAuxiliaryBody is the recover boundary, and it is one place rather than
// every executor's own discipline.
//
// Without it a panic in the delivery executor kills the process, and with the
// process goes the loop that places a stop-loss — which makes "an executor's
// stop must not stop the other loops" false on exactly the path where it matters
// most (engine-safety, 「패닉도 정지로 다뤄야 한다」).
func runAuxiliaryBody(ctx context.Context, aux AuxiliaryExecutor) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("%w: %v", ErrAuxiliaryPanicked, rec)
		}
	}()
	return aux.Run(ctx)
}

// AuxiliaryNames lists the auxiliary executors, in the order they were supplied.
//
// It exists so "what this process started outside supervision" is answerable.
// LoopNames alone cannot say it, and an executor nobody can enumerate is one
// nobody notices the absence of — which is the defect this whole change is about,
// one level up.
func (r *Runtime) AuxiliaryNames() []string {
	out := make([]string, 0, len(r.opts.Auxiliary))
	for _, aux := range r.opts.Auxiliary {
		out = append(out, aux.Name)
	}
	return out
}

// AlertDeliverer builds the executor that drains the critical-alert outbox.
//
// It is constructed here, next to the journal and the gate it needs, for the
// reason Context.Recovery is: cmd/tossctl has neither, and a wiring assembled
// out there could be pointed at a journal the engine did not open.
//
// The publisher comes from the notifier because that is where the engine's one
// resolved transport lives (engine.go). A nil one is not an error — the row stays
// PENDING and the loop says so once per cycle, which is the existing behaviour
// for an engine with no notification transport configured.
func (c *Context) AlertDeliverer(clk clock.Clock) (AuxiliaryExecutor, error) {
	if c == nil || c.Journal == nil || c.Entry == nil {
		return AuxiliaryExecutor{}, fmt.Errorf(
			"%w: the alert delivery executor needs the journal and the entry gate", ErrRuntimeUnavailable)
	}
	if clk == nil {
		clk = c.clock()
	}
	var publisher obs.Publisher
	if c.Notifier != nil {
		publisher = c.Notifier.Publisher
	}
	deliverer := &alertDeliverer{
		Journal:   c.Journal,
		Publisher: publisher,
		Log:       c.Log,
		Clock:     clk,
	}
	gate, log := c.Entry, c.Log
	return AuxiliaryExecutor{
		Name: alertDeliveryName,
		Run:  deliverer.Run,
		OnStop: func(_ context.Context, err error) {
			blockEntryOnDeliveryStop(gate, log, err)
		},
	}, nil
}

// blockEntryOnDeliveryStop is what the delivery executor's death costs: new
// exposure is refused until an operator restarts the engine.
//
// # Why it blocks rather than escalates
//
// The operating mode is durable and its relaxation needs a human
// (risk-management), so escalating would keep entry shut after the reason for it
// had gone: a restarted process has a live delivery executor, which is precisely
// the recovery. A latch that a restart clears is the right shape here, and the
// case it would otherwise leave open — undelivered rows surviving the restart —
// is already closed by restoreAlertEntryLatch reading the ledger at startup
// (사용자 결정 11-1).
//
// # Why its own reason code
//
// Merging it with ReasonAlertUndelivered would let an operator acknowledging the
// backlog clear "nobody is sending" as well. The backlog would be empty, entry
// would be open, and there would still be nobody sending (사용자 결정 8-1).
func blockEntryOnDeliveryStop(gate *execgw.EntryGate, log *obs.Logger, err error) {
	if gate == nil {
		return
	}
	// The detail carries no part of err: an executor's failure can quote a
	// ledger row, and a gate detail is read wherever gate state is read
	// (불변식 8). What an operator needs from this string is what to do.
	gate.Block(execgw.ReasonAlertSenderDown,
		"the critical alert delivery executor stopped, so nothing is draining the outbox; "+
			"restart the engine to recover")
	if log == nil {
		return
	}
	log.Error(obs.EventAlertUndelivered, err,
		obs.FieldDetail, "the critical alert delivery executor stopped; new entry is blocked",
		obs.FieldReason, string(execgw.ReasonAlertSenderDown))
}

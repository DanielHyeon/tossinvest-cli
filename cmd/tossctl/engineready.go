package main

// engineready.go ties the boot survey to a signal instead of to a stopwatch
// (openspec change a102, D5·D5b·D6·D7).
//
// # What went wrong
//
// Measured 2026-08-13 02:03. The console's engine autostart returned after
// engineStartProbe (3s, engineproc.go), the survey started 2ms later, and the
// engine's restart recovery — which had barely begun, and which measured ~50s on
// a healthy morning — was refused by the broker with a rate limit and died. The
// engine was up for one second. Every stop-loss the engine watches went with it.
//
// a101 put the survey after the engine for exactly this reason, and the ordering
// bought three seconds. Twenty-six pages of open orders do not finish in three
// seconds, so this change binds the ordering to the engine's own statement that
// it is done rather than to a probe that only proves the process did not exit.
//
// # Why the judgements live here and not in runConsole
//
// runConsole is measured at 0.0% coverage (this change's
// analysis/function-logic/cmd-tossctl--runconsole/branch-test-map.md: 44 AST
// branches, 0 executed). A judgement written inside it is a judgement no test can
// reach. Everything below therefore takes its dependencies as arguments — the
// shape a101 arrived at for the same reason — and runConsole keeps a call and a
// print.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// --- D5: only a finished recovery may say the engine is ready -------------------

// recoverThenReady wraps the runtime's restart-recovery step so the ready signal
// belongs to a recovery that finished.
//
// The runtime calls Recover once, before any loop starts
// (internal/app/engine/runtime.go). ready is called on that path and on no other:
// a failed recovery leaves the entry gate latched shut and the account
// unreconciled, and a cancelled one belongs to a process that is on its way down.
// Publishing "ready" in either case would tell the survey that protection is
// standing when it is not.
//
// observe receives the report on every path — success, failure and cancellation —
// because the thing it reports is the wait, and the wait happened either way
// (D5b).
func recoverThenReady(
	run func(context.Context) (reconcile.Report, error),
	ready func(),
	observe func(reconcile.Report),
) func(context.Context) error {
	return func(ctx context.Context) error {
		if ctx == nil {
			// The runtime supplies one, but this closure is the boundary and a nil
			// here would panic on the cancellation check below — inside the engine's
			// boot, before any loop starts (a102 A2 F9).
			ctx = context.Background()
		}
		report, err := run(ctx)
		if observe != nil {
			observe(report)
		}
		if err != nil {
			return err
		}
		if cerr := ctx.Err(); cerr != nil {
			// 복구가 오류 없이 돌아왔지만 그 사이 정지 신호가 왔다. 곧 내려갈
			// 프로세스가 "준비됐다"를 남기면 다음 콘솔은 그것을 최대 StaleAfter 동안
			// 믿는다.
			return cerr
		}
		if ready != nil {
			ready()
		}
		return nil
	}
}

// engineRecoveryObserver turns a throttled restart recovery into one countable
// line (A1 F1).
//
// Before this the report was discarded at the call site (`_, rerr :=
// recovery.Run(ctx)`), so the rate-limit budget §1 introduced — up to five
// minutes of waiting with the entry gate shut — would have passed without a
// single line anywhere. The operator has to be able to count the time in which
// the engine was up and the account was not yet reconciled.
//
// Silence is the normal case: a recovery that was never refused writes nothing,
// so this line stays a marker of an abnormal morning rather than boot noise.
func engineRecoveryObserver(logger *obs.Logger) func(reconcile.Report) {
	return func(report reconcile.Report) {
		if report.RateLimitWaits <= 0 {
			return
		}
		logger.Warn(obs.EventRecoveryRateLimited,
			obs.FieldCount, report.RateLimitWaits,
			obs.FieldDurationMS, report.RateLimitWaited.Milliseconds(),
			obs.FieldReason, "브로커가 계좌 조회를 rate limit으로 거부해 재시작 복구가 기다렸다",
		)
	}
}

// --- D4b: the signal belongs to a process that is still alive --------------------

// engineLook folds one observation of the engine into the only three answers the
// wait needs.
//
// The three exist because a marker is a file and a file outlives the process that
// wrote it. A crash never calls Release, so a marker carrying ready_at stays
// readable for StaleAfter (5m) and freshness alone reads it as "an engine is
// protecting this account". A2 reproduced the next scene of the 02:03 incident
// from that: the console starts a replacement engine, and while the replacement
// is still assembling, the predecessor's ready_at is read as "ready" and the
// survey goes straight at the budget the new engine's recovery needs.
type engineLook int

const (
	// engineLookReady is a live engine process that has published ready_at.
	engineLookReady engineLook = iota
	// engineLookNotYet is an engine worth waiting for that has not said it is
	// ready — including the case where the only marker on disk belongs to a
	// process that is gone.
	engineLookNotYet
	// engineLookAbsent is no live engine at all. Waiting for one would delay the
	// survey for nothing, which is what today's behaviour already avoids.
	engineLookAbsent
)

// readEngineSignal decides what one look at the engine means.
//
// The process list is the authority on liveness and the marker is the authority
// on readiness; neither answers for the other. Rules, in order:
//
//   - an enumeration that failed proves nothing — not liveness, and not that the
//     marker on disk belongs to a corpse. It is "not yet", which costs at most
//     the cap and then starts anyway. bootSurvey draws the same line for the same
//     reason ("A failed enumeration is not 'nothing is running'").
//   - no engine process: absent, whatever the file says. A marker with ready_at
//     and no process behind it is exactly the ghost a container recreate leaves.
//   - ready_at present and the marker's pid is one of the live engines: ready.
//   - anything else: not yet. That covers the boot window (process up, marker not
//     written or not ready) and the corpse marker (process up, but a different
//     one) — in both, the live engine is about to write its own marker.
func readEngineSignal(status enginelock.Status, pids []int, findErr error) engineLook {
	if findErr != nil {
		return engineLookNotYet
	}
	if len(pids) == 0 {
		return engineLookAbsent
	}
	if !status.Ready() {
		return engineLookNotYet
	}
	for _, pid := range pids {
		if pid == status.Marker.PID {
			return engineLookReady
		}
	}
	return engineLookNotYet
}

// consoleEngineReadiness builds the console's observation seam.
//
// It is a named function and not a closure inside runConsole because runConsole
// is measured at 0.0%: a composition written there is a composition no test can
// reach, and A2's surviving mutation N2 was exactly that — blanking this
// observation left every unit test green.
func consoleEngineReadiness(dir, markerPath string, now func() time.Time) func() engineLook {
	return func() engineLook {
		if strings.TrimSpace(dir) == "" || strings.TrimSpace(markerPath) == "" {
			// The console could not resolve the engine directory. There is no
			// marker to read and no journal to scope a process search to.
			return engineLookAbsent
		}
		at := time.Now()
		if now != nil {
			at = now()
		}
		pids, findErr := engineFindProcesses(dir)
		return readEngineSignal(enginelock.Read(markerPath, at), pids, findErr)
	}
}

// --- D6: the bounded wait --------------------------------------------------------

// engineReadyVerdict is how a wait ended. All four are reported to the operator;
// none of them is silent (operator-console: 조용한 상한 초과는 금지다).
type engineReadyVerdict int

const (
	// engineReadyConfirmed saw the engine's ready signal.
	engineReadyConfirmed engineReadyVerdict = iota
	// engineReadyNoEngine found no live engine to wait for.
	engineReadyNoEngine
	// engineReadyCapExceeded gave up waiting and starts anyway.
	engineReadyCapExceeded
	// engineReadyAbandoned is the console going away mid-wait.
	engineReadyAbandoned
)

// engineReadyCap is how long the boot survey waits for the ready signal.
//
// 120s covers the measured restart recovery (~50s: the survey cycle either side
// of it ran 01:35:02 → 01:35:52 on 2026-08-13) with room for a rate-limit
// backoff or two on top. It is deliberately shorter than the recovery's own
// rate-limit budget (reconcile.DefaultMaxRateLimitWait, 5m): the cap is the
// limit on *quiet waiting*, not on the recovery, and a survey that starts after
// it is a contention the recovery's backoff now survives.
//
// It is a constant and not a setting. Nothing in the incident this fixes turns on
// tuning it, and a knob here would be a second place where the two halves of this
// change could be made to disagree.
const engineReadyCap = 2 * time.Minute

// engineReadyPoll is how often the marker is re-read while waiting.
//
// Finer than the marker's own refresh (enginelock.RefreshEvery, 1m) would be
// wasted work if the marker only changed on a refresh — but Ready rewrites it the
// moment recovery finishes, so polling is what turns that write into a start.
// Two seconds detects it within a cycle of the survey's own cadence and costs
// sixty stat calls across the whole cap.
const engineReadyPoll = 2 * time.Second

// awaitEngineReady waits for the engine to say it finished restart recovery.
//
// It ends in exactly four ways and each one is a verdict the caller reports:
//
//	confirmed     the marker carries ready_at — start now
//	no engine     nothing live to wait for — start now, which is today's behaviour
//	cap exceeded  waited long enough — start anyway, and say so
//	abandoned     the console is closing — start nothing
//
// "No engine" is checked every round rather than only at the start because the
// engine that caused this change was alive for one second: waiting the full cap
// for a marker that has already gone stale would delay the survey for no reason.
// The cap is named limit here only because `cap` is a builtin, and a parameter
// that shadows one in a file about waiting is a reading hazard for no gain.
func awaitEngineReady(
	ctx context.Context,
	observe func() engineLook,
	clk clock.Clock,
	limit, poll time.Duration,
) engineReadyVerdict {
	if observe == nil {
		// The console could not resolve the engine directory, so there is no
		// marker to read. That is "no engine", the same as an empty path
		// (enginelock.Read is lenient about both).
		return engineReadyNoEngine
	}
	if ctx != nil && ctx.Err() != nil {
		return engineReadyAbandoned
	}
	deadline := clk.Now().Add(limit)
	for {
		switch observe() {
		case engineLookAbsent:
			return engineReadyNoEngine
		case engineLookReady:
			return engineReadyConfirmed
		}
		if !clk.Now().Before(deadline) {
			return engineReadyCapExceeded
		}
		if err := clk.Sleep(ctx, poll); err != nil {
			return engineReadyAbandoned
		}
	}
}

// engineReadyNote is what the operator reads about the wait. Every verdict has a
// sentence: a survey that started 120 seconds late without saying why is the
// silent cap the spec forbids.
func engineReadyNote(verdict engineReadyVerdict) string {
	switch verdict {
	case engineReadyConfirmed:
		return "엔진 준비 확인 후"
	case engineReadyNoEngine:
		return "살아 있는 엔진이 없어 대기 없이"
	case engineReadyCapExceeded:
		return fmt.Sprintf("엔진 준비 신호가 %s 안에 오지 않아 상한 초과 후", engineReadyCap)
	case engineReadyAbandoned:
		// The only verdict that stands alone: no start note follows it, so it is a
		// finished sentence rather than the prefix the other three are.
		return "콘솔이 종료되어 서베이를 시작하지 않았다"
	}
	return "알 수 없는 대기 결과 후"
}

// errSoakBootNoStartWiring is the programming-error direction: a wait with
// nothing to start after it.
var errSoakBootNoStartWiring = errors.New(
	"이 빌드에는 부팅 서베이 기동 배선이 없다")

// --- D7: the wait wraps the start seam, not the console -------------------------

// soakStartAfterEngineReady puts the wait in front of a boot start seam.
//
// It wraps rather than edits: runConfiguredSoakAutostart and bootSurvey keep
// every branch they had (both measured at 100% — a101), and this adds one
// decision in front of the seam they already take as an argument. Autostart being
// off means start is never called, so the wait costs nothing there for free.
//
// The button is deliberately not wrapped. An operator who just pressed [soak
// 재시작] and waits two minutes cannot tell that from a button that does not work.
//
// A console that closed mid-wait comes back as a note and not as an error. The
// error shape made a normal shutdown print "soak 자동 시작 실패:", which reads as a
// fault an operator should chase (a102 A2 F8); the sentence itself already says
// nothing was started.
func soakStartAfterEngineReady(
	ctx context.Context,
	observe func() engineLook,
	clk clock.Clock,
	start func() (string, error),
) func() (string, error) {
	return func() (string, error) {
		verdict := awaitEngineReady(ctx, observe, clk, engineReadyCap, engineReadyPoll)
		if verdict == engineReadyAbandoned {
			return engineReadyNote(verdict), nil
		}
		if start == nil {
			// runConfiguredSoakAutostart already answers "this build has no soak
			// wiring" for its own argument. A nil arriving here is a wiring
			// mistake inside this package, and it comes back as one line rather
			// than as a panic in a goroutine that would take the console down —
			// which is the one thing the survey may never cost (a101).
			return "", errSoakBootNoStartWiring
		}
		note, err := start()
		if strings.TrimSpace(note) == "" {
			return engineReadyNote(verdict), err
		}
		return engineReadyNote(verdict) + " " + note, err
	}
}

// --- D7b: one console, one spawner at a time ------------------------------------

// soakSpawnGate serialises the two paths in this process that can spawn a survey.
//
// Before a102 the boot decision finished before the HTTP listener came up, so the
// button could not exist yet and the two could not overlap. Making the wait
// asynchronous opened a window of up to the cap in which both are live, and both
// are check-then-act over the same record: soakFindProcesses, then spawn. Without
// a lock the two enumerations can both answer "nothing is running" and both
// spawn, which soakproc itself calls worse than a survey that needs a person
// ("기록 하나에 두 서베이가 붙는 편이 더 나쁘다").
//
// It is package scope on purpose. Two gates serialise nothing, and a gate handed
// to each path as an argument is a wiring line inside runConsole, which is
// measured at 0.0% — the same hole that let A2's N2 survive.
var soakSpawnGate sync.Mutex

// spawnOneSurvey runs start with the console's spawn gate held.
func spawnOneSurvey(start func() (string, error)) (string, error) {
	if start == nil {
		return "", errSoakBootNoStartWiring
	}
	soakSpawnGate.Lock()
	defer soakSpawnGate.Unlock()
	return start()
}

// guardedSoakRestart is the dashboard button's spawn, serialised against the boot
// path and then recorded exactly as a101 recorded it.
func guardedSoakRestart(start func() (string, error), save func(bool) error) (string, error) {
	note, err := spawnOneSurvey(start)
	return rememberSoakApproval(save, note, err)
}

// --- D7c: the console never joins the wait --------------------------------------

// startSoakAutostartAsync is the whole boot-survey block, off the console's path.
//
// Everything the decision needs is built here rather than at the call site, so
// the observation seam, the clock and the two constants are covered by tests
// instead of living in runConsole where nothing reaches them. runConsole keeps one
// call.
//
// It returns before the survey does — always. A source-shape test cannot promise
// that: A2 wrote a version that kept `go func() {` and every other string the
// shape test looked for and still blocked the console for the full cap, by
// joining the goroutine before returning. So the promise is asserted by running
// it against a start that never returns.
func startSoakAutostartAsync(
	ctx context.Context,
	engineDir, markerPath string,
	load func() (bool, error),
	start func() (string, error),
	report func(string),
) {
	observe := consoleEngineReadiness(engineDir, markerPath, time.Now)
	guarded := soakStartAfterEngineReady(ctx, observe, clock.System(), func() (string, error) {
		return spawnOneSurvey(start)
	})
	go func() {
		note := runConfiguredSoakAutostart(load, guarded)
		if strings.TrimSpace(note) == "" || report == nil {
			return
		}
		report(note)
	}()
}

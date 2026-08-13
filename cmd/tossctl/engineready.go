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
	observe func() enginelock.Status,
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
		status := observe()
		if !status.Running {
			return engineReadyNoEngine
		}
		if status.Ready() {
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
		return "콘솔 종료로"
	}
	return "알 수 없는 대기 결과 후"
}

// errSoakBootConsoleClosed is the one verdict that does not start a survey.
//
// It comes back as an error rather than a note because that is how
// runConfiguredSoakAutostart already reports "nothing was started", and a console
// on its way down must not leave a survey nobody is watching behind it.
var errSoakBootConsoleClosed = errors.New(
	"콘솔이 종료되어 서베이를 시작하지 않았다")

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
func soakStartAfterEngineReady(
	ctx context.Context,
	observe func() enginelock.Status,
	clk clock.Clock,
	start func() (string, error),
) func() (string, error) {
	return func() (string, error) {
		verdict := awaitEngineReady(ctx, observe, clk, engineReadyCap, engineReadyPoll)
		if verdict == engineReadyAbandoned {
			return "", errSoakBootConsoleClosed
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

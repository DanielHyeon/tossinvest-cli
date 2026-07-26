package main

// console.go is `tossctl console`: the local operator console for the one-off
// live-account verification (openspec change verify-execution-capability,
// task 1.6).
//
// # Why the console exists at all
//
// `tossctl verify run` already does the measurement, and it will keep doing it.
// The console exists because the user decided (사용자 결정 2026-07-26) to drive
// the verification from a screen rather than a terminal stream: the batch summary
// is a dozen live requests with prices and reversals, and reading that in a
// scrollback while deciding whether to type a confirmation is worse than reading
// it on a page.
//
// It is a stopgap. internal/console's package doc says so and means it: single
// user, loopback only, deleted when the Phase 4 daemon lands.
//
// # What this file is responsible for
//
// Exactly one thing that matters: it is the only place in the binary that hands
// internal/console a way to reach a live account, and it hands it the same runner
// `verify run` builds, gated on the console's web confirmer instead of the
// terminal's. Everything else — the approval, the session, the rendering — is
// internal/console's.
//
// # The absences, which are the point
//
// There is no flag that presets the session token, answers the approval, or moves
// the console off the loopback interface. `verify run` gains nothing: its
// confirmers are still terminalConfirmer and terminalBatchConfirmer, and
// console_test.go asserts that in the source, because "the CLI must not gain a
// non-interactive approval path" is the condition under which task 1.6 permits the
// web form to exist at all.
//
// --confirm-each is not offered here either. The console is batch-only, and the
// per-mutation confirmer it wires refuses — so the finer gate fails closed rather
// than being silently satisfied if a future edit ever reaches it.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/handoff"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/spf13/cobra"
)

// consoleProbeSymbol is the KR symbol the buy-side probes are placed against.
//
// It is `verify run --symbol`'s default, restated rather than shared because
// verify.go is a live-order path and this task's scope is additive. Drift is not
// left to review: console_test.go asserts the two agree.
const consoleProbeSymbol = "005930"

type consoleOptions struct {
	port int
}

func newConsoleCmd(root *rootOptions) *cobra.Command {
	opts := &consoleOptions{}

	cmd := &cobra.Command{
		Use:   "console",
		Short: "Serve the local operator console for the live-account verification",
		Long: strings.TrimSpace(`
Serve a small web console on 127.0.0.1 for driving the one-off live-account
verification from a browser instead of a terminal.

It binds the loopback interface and nothing else. There is no flag to change
that. On start it prints a URL carrying a one-time session token: opening that
URL in this machine's browser is what authenticates you, so possession of this
terminal is the credential. Do not paste the link anywhere else.

  dashboard   soak progress, attestation state, verification progress
  verify      the step list, the batch summary, the typed approval, live progress
  report      the measured attributes and the ones still unverified, plus JSON

Everything is read-only except the verification approval. The console places no
order of its own, toggles no gate, edits no configuration and shows no
credential: the only requests that reach the account are the ones the verify
runner makes, under the same plan authorisation, the same exposure caps and the
same cancellation rails as ` + "`tossctl verify run`" + `.

Approving is a three-part act and all three are required: the session token, a
CSRF token the page carries, and the expiring confirmation string the page shows
you, typed back by hand. Anything missing or wrong sends nothing. There is no
flag, here or on ` + "`tossctl verify run`" + `, that answers it for you.

The conditional-order persistence check needs a NEW process, so this console runs
at most one verification per start: when it stops there, quit with Ctrl-C, start
the console again, and press 이어하기.

A step whose last verdict is fail or skipped can be attempted again from the
verify screen's 재측정 button — the market was closed, the broker throttled, the
account held nothing. The set is worked out from the evidence record, never from
the page: a step that passed is never re-measured, and a re-measurement asks for
its own batch approval with a new confirmation string like any other run.

While a verification is running this command marks the account busy so a
concurrent ` + "`tossctl soak run`" + ` delays its cycle rather than spending the same
rate limit.`),
		// official: the verification it drives reaches the Open API. mutating: it
		// can place live orders — through the verify runner, after a typed approval.
		Annotations:  map[string]string{"source": "official", "mutating": "true"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runConsole(cmd, root, opts)
		},
	}

	cmd.Flags().IntVar(&opts.port, "port", 0,
		"Loopback port to serve on; 0 lets the OS pick a free one")
	return cmd
}

func runConsole(cmd *cobra.Command, root *rootOptions, opts *consoleOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	// Ctrl-C has to let a verification in progress finish cancelling whatever it
	// has resting; internal/console waits for that before it shuts the socket.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	verifyRecord, err := resolveVerifyRecord(root, "")
	if err != nil {
		return err
	}
	soakRecord, err := resolveSoakRecord(root, "")
	if err != nil {
		return err
	}
	attestation, err := resolveSoakAttestationPath(root, "")
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	return console.ListenAndServe(ctx, console.Options{
		Port:              opts.port,
		StartVerify:       consoleVerifyStarter(root, verifyRecord),
		SoakRecord:        soakRecord,
		VerifyRecord:      verifyRecord,
		Attestation:       attestation,
		MinSoakDays:       soak.DefaultCriteria().MinConsecutiveDays,
		RequiredEndpoints: engine.RequiredEndpoints(),
		Out:               out,

		// The three seams task 1.8 puts behind the console's two restart buttons.
		// internal/console executes nothing: it decides whether the person asking
		// has cleared the session and CSRF gates, and then calls one of these.
		Relaunch:    consoleRelaunch(out),
		Handoff:     handoff.New(consoleHandoffPath(verifyRecord)),
		RestartSoak: func() (string, error) { return restartSoak(soakRecord) },
	})
}

// consoleHandoffPath is where the single-use restart token lives: beside the
// evidence record and the soak pause marker, so an isolated --config-dir profile
// gets its own and the default sits in the data directory with the rest of this
// change's state.
func consoleHandoffPath(recordPath string) string {
	return filepath.Join(filepath.Dir(recordPath), handoff.FileName)
}

// consoleRelaunch re-executes this binary so a NEW process instance starts.
//
// That is the whole product: internal/verifylive's conditional-persistence step
// refuses to certify a conditional from the process that registered it, and it
// judges that on process.instance_id — minted fresh at every startup — rather than
// on the PID, which an exec preserves. The restart is therefore a real process
// boundary for the measurement, and the record proves it.
//
// The port is pinned so the browser comes back to the address it is already on.
// Everything else about the command line is kept: the same subcommand, the same
// --config-dir, the same flags the operator typed.
func consoleRelaunch(out io.Writer) console.Relaunch {
	return func(port int) error {
		path, err := binstamp.SelfPath()
		if err != nil {
			return err
		}
		argv := argvWithPort(os.Args, port)
		fmt.Fprintf(out, "  %s\n\n", strings.Join(argv, " "))
		return reexecSelf(path, argv)
	}
}

// consolePortFlag is the flag argvWithPort rewrites. It is named once.
const consolePortFlag = "--port"

// argvWithPort preserves the command line and pins the loopback port.
//
// A console started without --port took whatever the OS offered, and the browser is
// sitting on it. Re-executing with the original argument list would take another
// free port and strand the tab, so the port this process ended up on is written in
// explicitly — which is the one thing about the restart the operator did not choose
// and should not have to.
func argvWithPort(args []string, port int) []string {
	out := make([]string, 0, len(args)+2)
	skipNext := false
	for i, a := range args {
		switch {
		case skipNext:
			skipNext = false
		case i == 0:
			out = append(out, a)
		case a == consolePortFlag:
			skipNext = true // and its value
		case strings.HasPrefix(a, consolePortFlag+"="):
		default:
			out = append(out, a)
		}
	}
	return append(out, consolePortFlag, strconv.Itoa(port))
}

// consoleVerifyStarter builds the runner the console drives.
//
// It is `runVerifyRun`'s wiring with two differences and no others: the batch
// confirmer is the console's instead of the terminal's, and the run is always a
// resumption. The second follows from the first — a console has no --resume flag
// to forget, and the runner already refuses to re-measure a settled step (it skips
// it, and the plan excludes it with the reason on the page), so continuing the
// record is both the safe default and the only sensible one.
//
// redo is `verify run --redo`'s field reached from a button instead of a flag
// (task 1.7). The console computes the set from the evidence record — never from
// the request — and it changes only which steps the runner will walk: the plan is
// rebuilt and a new expiring string still has to be typed before anything is sent.
func consoleVerifyStarter(root *rootOptions, recordPath string) console.StartVerify {
	return func(
		ctx context.Context,
		confirm verifylive.BatchConfirmer,
		out io.Writer,
		redo []verifylive.StepID,
	) (verifylive.Summary, []verifylive.Entry, error) {
		var empty verifylive.Summary

		prior, err := verifylive.LoadEntries(recordPath)
		if err != nil {
			return empty, nil, err
		}
		broker, accountRef, err := verifyBrokerFactory(root)
		if err != nil {
			return empty, nil, err
		}

		recorder, err := verifylive.OpenRecorder(recordPath)
		if err != nil {
			return empty, nil, err
		}
		defer recorder.Close()

		runner, err := verifylive.New(verifylive.Options{
			Broker:          broker,
			Recorder:        recorder,
			Confirm:         consoleMutationConfirmer(),
			ConfirmBatch:    confirm,
			Out:             out,
			AccountRef:      accountRef,
			Symbol:          consoleProbeSymbol,
			HoldingSymbol:   firstUsableHolding(ctx, broker),
			Offset:          verifylive.DefaultOffset,
			MaxSellQuantity: verifylive.DefaultMaxSellQuantity,
			TTLWait:         verifylive.DefaultTTLWait,
			Redo:            redo,
			Prior:           prior,
		})
		if err != nil {
			return empty, nil, err
		}

		fmt.Fprintf(out, "evidence record  %s\n", recordPath)
		releaseLock := holdVerifyRunLock(ctx, out, recordPath)
		defer releaseLock()

		summary, runErr := runner.Run(ctx)
		if runErr != nil && (errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded)) {
			// Ctrl-C on the console. Everything recorded stands, and the summary
			// already names whatever is still live.
			runErr = nil
		}
		return summary, runner.Entries(), runErr
	}
}

// consoleMutationConfirmer is the per-mutation gate the console does not offer.
//
// verifylive.Options requires a non-nil Confirmer and refuses a nil one rather
// than defaulting to something permissive; this is the value that satisfies it.
// The console never sets ConfirmEach, so nothing calls it — and if a later edit
// ever did, it refuses, which is the direction a mistake here has to fail in.
func consoleMutationConfirmer() verifylive.Confirmer {
	return func(verifylive.Mutation) error { return verifylive.ErrNotATerminal }
}

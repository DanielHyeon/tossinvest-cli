package main

// flatten.go is `tossctl flatten-all`: the emergency exit (harden-execution-base
// task 4.5, engine-safety "Flatten Saga").
//
// # Why this command uses the engine profile
//
// It builds internal/app/engine's wiring rather than the CLI's hybrid client, so
// every mutation goes through the official Open API and through
// internal/execgw.Gateway — which means every one of them is journalled, and a
// crash mid-flatten is resumable. The CLI's own order commands are unchanged;
// this is a new file and a new command, and nothing in the upstream order path is
// touched (design D1, §0.2).
//
// A consequence worth stating plainly: with no official credentials this command
// refuses to run. That is the engine profile's rule, and for an emergency exit it
// is the right one — a flatten that silently ran through a scraped web session
// would be a flatten with no durable record of what it did.
//
// # The two gates, and the one that is deliberately absent
//
//	--dry-run     prints the plan, submits nothing
//	confirmation  a terminal-typed, expiring nonce (internal/flatten/confirm.go)
//
// There is no --yes, no --force and no environment variable that stands in for
// the confirmation. §0.7 makes this a human decision, and the spec forbids an
// automation flag outright.
//
// The confirmation guards the *selling*. The cancels run first and are not
// separately confirmed: cancelling only reduces exposure, and §0.3 forbids
// putting anything in front of that.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/flatten"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/JungHoonGhae/tossinvest-cli/internal/tui"
	"github.com/spf13/cobra"
)

type flattenOptions struct {
	dryRun bool
	reason string
	// cancelOnly stops after phase 1. It is not a shortcut around the
	// confirmation: cancelling never needed one.
	cancelOnly bool
}

func newFlattenCmd(root *rootOptions) *cobra.Command {
	opts := &flattenOptions{}

	cmd := &cobra.Command{
		Use:   "flatten-all",
		Short: "Cancel every open order and sell every position (emergency exit)",
		Long: strings.TrimSpace(`
Cancel every open order on the account, then sell every position with aggressive
limit orders.

This is the emergency exit. It runs as a durable saga: if it is interrupted, run
it again and it resumes from where it stopped rather than re-submitting what it
already did.

Order of operations:

  1. new entries are blocked (and stay blocked until you clear it)
  2. every open order is cancelled and its result confirmed
  3. the account is re-read until it stops moving
  4. each position is sold at an aggressive limit, sized from the sellable
     quantity the account reports
  5. the account is re-read to confirm nothing is left

A symbol whose cancel could not be confirmed is held back from step 4: selling
the full holding while an order may still be live would oversell.

Requires official Open API credentials — the web session is never used to place
orders. Selling is confirmed by typing an expiring string at a terminal; there is
no flag that answers it for you.`),
		// official: orders are placed and cancelled through the Open API only —
		// the engine profile has no web-session path. mutating: this is the most
		// consequential trade action in the CLI.
		Annotations:  map[string]string{"source": "official", "mutating": "true"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runFlatten(cmd, root, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false,
		"Print what would be cancelled and sold, and submit nothing")
	cmd.Flags().StringVar(&opts.reason, "reason", "",
		"Why the account is being flattened, recorded in the journal")
	cmd.Flags().BoolVar(&opts.cancelOnly, "cancel-only", false,
		"Stop after cancelling open orders; do not sell positions")

	return cmd
}

// flattenWiring is everything the saga needs, assembled once.
type flattenWiring struct {
	orders  *engine.OrderPath
	journal *journal.Journal
	saga    *flatten.Saga
	account string
	close   func()
}

func runFlatten(cmd *cobra.Command, root *rootOptions, opts *flattenOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := output.ParseFormat(root.outputFormat)
	if err != nil {
		return err
	}

	wiring, err := buildFlattenWiring(ctx, root, opts)
	if err != nil {
		return err
	}
	defer wiring.close()

	out := cmd.OutOrStdout()

	// --- phase 1: cancel ----------------------------------------------------
	//
	// Unconfirmed on purpose. Cancelling reduces exposure and nothing else, and
	// §0.3 forbids putting a prompt in front of that.
	cancelReport, err := wiring.saga.CancelAll(ctx)
	if err != nil {
		return err
	}
	writeCancelReport(out, format, cancelReport)

	if opts.cancelOnly {
		return nil
	}

	// --- the human gate -----------------------------------------------------
	targets, err := wiring.saga.PlanLiquidation(ctx)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		fmt.Fprintln(out, "\nNo positions to liquidate; the account is flat.")
		return nil
	}

	if !opts.dryRun {
		confirmation := flatten.NewConfirmation(wiring.account, cancelReport.Found, targets, time.Now())
		interactive := tui.IsInteractive(os.Stdin, os.Stdout)
		if err := flatten.Confirm(cmd.InOrStdin(), out, confirmation, interactive, time.Now()); err != nil {
			return err
		}
	}

	// --- phase 2: liquidate -------------------------------------------------
	liquidateReport, err := wiring.saga.Liquidate(ctx)
	if err != nil {
		return err
	}
	writeLiquidateReport(out, format, liquidateReport)

	if !liquidateReport.Flat() && !opts.dryRun {
		return fmt.Errorf("flatten-all did not finish: %d position(s) remain, %d symbol(s) held — "+
			"see `tossctl flatten-all --dry-run` and the journal for detail",
			liquidateReport.Remaining, liquidateReport.Held)
	}
	return nil
}

// buildFlattenWiring assembles the order path, the journal and the saga.
//
// It builds its own order path rather than an engine.Context (task 7.4). The
// difference matters in both directions: this command is its own decision issuer
// and owns its journal, its entry gate and its gateway, so taking a
// trading.Service out of an engine it did not assemble was the exposure the seal
// removes (engine-safety: "기존 소비자인 flatten은 엔진 컨텍스트가 아니라 자체
// 배선으로 구성한다") — and an engine.Context now opens a journal of its own,
// which this command would then be holding a second handle to.
//
// Nothing about what the saga does changes. Same official client, same trading
// policy, same journal path, same gateway options, same refusal without
// credentials.
func buildFlattenWiring(ctx context.Context, root *rootOptions, opts *flattenOptions) (*flattenWiring, error) {
	orderPath, err := engine.NewOrderPath(engine.Options{ConfigDir: root.configDir})
	if err != nil {
		return nil, err
	}

	accounts, err := orderPath.Official.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("flatten-all: the account could not be identified: %w", err)
	}
	accountRef := ""
	for _, a := range accounts {
		if strings.TrimSpace(a.DisplayName) != "" {
			accountRef = strings.TrimSpace(a.DisplayName)
			break
		}
	}
	if accountRef == "" {
		return nil, errors.New("flatten-all: the broker returned no account number")
	}

	journalPath := ""
	if root.configDir != "" {
		// An explicit --config-dir means an isolated run; the journal follows it
		// so a test or a second profile cannot touch the real one.
		journalPath = filepath.Join(root.configDir, journal.DBFileName)
	}
	j, err := journal.Open(ctx, journal.Options{Path: journalPath})
	if err != nil {
		return nil, err
	}

	clk := clock.System()
	gate := execgw.NewEntryGate(clk, nil)
	gateway, err := execgw.New(execgw.Options{
		Journal:    j,
		Trading:    orderPath.Trading,
		Clock:      clk,
		AccountRef: accountRef,
		Source:     "flatten-all",
		Entry:      gate,
	})
	if err != nil {
		_ = j.Close()
		return nil, err
	}

	orders := execgw.OfficialOrders{Client: orderPath.Official}
	saga := &flatten.Saga{
		Journal:    j,
		Gateway:    gateway,
		Gate:       gate,
		Orders:     orders,
		Resolver:   &execgw.Resolver{Journal: j, Orders: orders, Order: orders, Clock: clk, Gate: gate},
		Log:        obs.NewLogger(obs.LogOptions{Writer: os.Stderr, Clock: clk}),
		Clock:      clk,
		AccountRef: accountRef,
		Operator:   os.Getenv("USER"),
		Reason:     opts.reason,
		DryRun:     opts.dryRun,
		Positions:  flattenPositions{client: orderPath.Official},
		Balance:    flattenBalance{client: orderPath.Official},
		Sellable:   orderPath.Official,
		Prices:     orderPath.Official,
	}

	return &flattenWiring{
		orders:  orderPath,
		journal: j,
		saga:    saga,
		account: accountRef,
		close:   func() { _ = j.Close() },
	}, nil
}

// flattenPositions adapts the official client to the snapshot's positions reader.
// An empty symbol filter is the whole-account sweep.
type flattenPositions struct{ client engine.OfficialReads }

func (p flattenPositions) Positions(ctx context.Context) ([]domain.Position, error) {
	return p.client.Holdings(ctx, "")
}

// flattenBalance adapts the official client to the snapshot's balance reader.
type flattenBalance struct{ client engine.OfficialReads }

func (b flattenBalance) BuyingPower(ctx context.Context, currency string) (float64, error) {
	bp, err := b.client.BuyingPower(ctx, currency)
	if err != nil {
		return 0, err
	}
	return bp.CashBuyingPower, nil
}

// --- output -----------------------------------------------------------------

func writeCancelReport(w io.Writer, format output.Format, report flatten.CancelReport) {
	if format == output.FormatJSON {
		_ = output.WriteJSON(w, report)
		return
	}
	prefix := ""
	if report.DryRun {
		prefix = "[dry-run] "
	}
	fmt.Fprintf(w, "%scancel phase — saga %s%s\n", prefix, report.SagaID, resumedNote(report.Resumed))
	fmt.Fprintf(w, "  open orders found  %d\n", report.Found)
	fmt.Fprintf(w, "  cancelled          %d\n", report.Cancelled)
	fmt.Fprintf(w, "  already gone       %d\n", report.Failed)
	if report.InDoubt+report.Held > 0 {
		fmt.Fprintf(w, "  unresolved         %d (their symbols are held back from selling)\n",
			report.InDoubt+report.Held)
	}
	for _, o := range report.Outcomes {
		fmt.Fprintf(w, "    %-12s %-10s %-8s %s\n", o.OrderID, o.Symbol, o.State, o.Detail)
	}
}

func writeLiquidateReport(w io.Writer, format output.Format, report flatten.LiquidateReport) {
	if format == output.FormatJSON {
		_ = output.WriteJSON(w, report)
		return
	}
	prefix := ""
	if report.DryRun {
		prefix = "[dry-run] "
	}
	fmt.Fprintf(w, "\n%sliquidation phase — saga %s (%d round(s))\n", prefix, report.SagaID, report.Rounds)
	for _, t := range report.Targets {
		fmt.Fprintf(w, "    %-10s %-6s qty %-10s @ %-10s %-8s %s\n",
			t.Symbol, t.Market, trimDecimal(t.Quantity), trimDecimal(t.Price), t.State, t.Detail)
	}
	fmt.Fprintf(w, "  submitted          %d\n", report.Submitted)
	fmt.Fprintf(w, "  held               %d\n", report.Held)
	fmt.Fprintf(w, "  positions left     %d\n", report.Remaining)
	fmt.Fprintf(w, "  phase              %s\n", report.Phase)
}

func resumedNote(resumed bool) string {
	if resumed {
		return " (resumed)"
	}
	return ""
}

func trimDecimal(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", v), "0"), ".")
}

package main

// engine_reconcile.go is the local, audited recovery door for active
// RECONCILE rows. It deliberately owns no broker mutation interface: the only
// broker handle retained by the wiring is engine.OfficialReads, and the only
// state change is Tracker.Resolve's durable journal release.

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/spf13/cobra"
)

type reconcileResolveOptions struct {
	confirm  bool
	operator string
	note     string
}

type reconcileResolveLock interface {
	Path() string
	Release()
}

// reconcileResolveRuntime is intentionally narrower than engine.Context. The
// command can read three account snapshots, reconstruct local state, compare, and
// invoke the existing audited release contract; it cannot name an order write.
type reconcileResolveRuntime struct {
	accountRef string
	collect    func(context.Context) (reconcile.Snapshot, error)
	localState func(context.Context, string) (reconcile.LocalState, error)
	compare    func(reconcile.Snapshot, reconcile.LocalState) reconcile.Diff
	validate   func() error
	resolve    func(context.Context, string, string) error
	close      func() error
}

type reconcileResolveDeps struct {
	journalDir func(*rootOptions) (string, error)
	acquire    func(string) (reconcileResolveLock, error)
	build      func(context.Context, *rootOptions) (reconcileResolveRuntime, error)
	sleep      func(context.Context, time.Duration) error
}

func productionReconcileResolveDeps() reconcileResolveDeps {
	clk := clock.System()
	return reconcileResolveDeps{
		journalDir: engineJournalDir,
		acquire: func(dir string) (reconcileResolveLock, error) {
			return enginelock.Acquire(dir)
		},
		build: buildReconcileResolveRuntime,
		sleep: clk.Sleep,
	}
}

func newEngineReconcileResolveCmd(root *rootOptions) *cobra.Command {
	return newEngineReconcileResolveCmdWithDeps(root, productionReconcileResolveDeps())
}

func newEngineReconcileResolveCmdWithDeps(root *rootOptions, deps reconcileResolveDeps) *cobra.Command {
	opts := &reconcileResolveOptions{}
	cmd := &cobra.Command{
		Use:   "reconcile-resolve",
		Short: "Release reconcile blocks after a fresh official comparison",
		Long: strings.TrimSpace(`
Release active quantity-mismatch RECONCILE blocks only after three identical
official account snapshots, each taken at least two seconds apart, agree with the corrected local
journal projection.

The automated engine must be stopped: this command takes the same exclusive lock
before it opens the journal or reads the account. It never places, previews,
amends, or cancels an order, and it does not change any engine setting or toggle.
Every successful release records the operator identity and note in the journal.`),
		Annotations:  map[string]string{"source": "official"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEngineReconcileResolve(cmd, root, opts, deps)
		},
	}
	cmd.Flags().BoolVar(&opts.confirm, "confirm", false,
		"Confirm that the engine is stopped and the audited reconcile release is intended")
	cmd.Flags().StringVar(&opts.operator, "operator", "", "Operator identity recorded with the release")
	cmd.Flags().StringVar(&opts.note, "note", "", "What was verified before releasing the reconcile block")
	_ = cmd.MarkFlagRequired("confirm")
	_ = cmd.MarkFlagRequired("operator")
	_ = cmd.MarkFlagRequired("note")
	return cmd
}

func runEngineReconcileResolve(cmd *cobra.Command, root *rootOptions, opts *reconcileResolveOptions,
	deps reconcileResolveDeps,
) error {
	if opts == nil || !opts.confirm {
		return errors.New("engine reconcile-resolve: --confirm must be explicitly true")
	}
	operator := strings.TrimSpace(opts.operator)
	if operator == "" {
		return errors.New("engine reconcile-resolve: --operator is required")
	}
	note := strings.TrimSpace(opts.note)
	if note == "" {
		return errors.New("engine reconcile-resolve: --note is required")
	}
	if deps.journalDir == nil || deps.acquire == nil || deps.build == nil || deps.sleep == nil {
		return errors.New("engine reconcile-resolve: recovery wiring is incomplete")
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	dir, err := deps.journalDir(root)
	if err != nil {
		return err
	}
	lock, err := deps.acquire(dir)
	if err != nil {
		return fmt.Errorf("engine reconcile-resolve: acquiring the engine lock: %w", err)
	}
	defer lock.Release()

	runtime, err := deps.build(ctx, root)
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if !closed && runtime.close != nil {
			_ = runtime.close()
		}
	}()
	if err := validateReconcileResolveRuntime(runtime); err != nil {
		return err
	}

	stabiliser := &reconcile.Stabiliser{
		MinInterval: reconcile.DefaultStabilisationInterval,
		Required:    reconcile.DefaultStabilisationCount,
	}
	first, err := runtime.collect(ctx)
	if err != nil {
		return fmt.Errorf("engine reconcile-resolve: collecting the first official snapshot: %w", err)
	}
	stabiliser.Offer(first)
	if err := deps.sleep(ctx, reconcile.DefaultStabilisationInterval); err != nil {
		return fmt.Errorf("engine reconcile-resolve: waiting between official snapshots: %w", err)
	}
	second, err := runtime.collect(ctx)
	if err != nil {
		return fmt.Errorf("engine reconcile-resolve: collecting the second official snapshot: %w", err)
	}
	stable := stabiliser.Offer(second)
	if !stable.Stable {
		why := strings.TrimSpace(stable.Why)
		if why == "" {
			why = "the two official snapshots did not agree"
		}
		return fmt.Errorf("engine reconcile-resolve: stable account evidence was not obtained: %s", why)
	}
	if err := deps.sleep(ctx, reconcile.DefaultStabilisationInterval); err != nil {
		return fmt.Errorf("engine reconcile-resolve: waiting for the final official snapshot: %w", err)
	}
	third, err := runtime.collect(ctx)
	if err != nil {
		return fmt.Errorf("engine reconcile-resolve: collecting the final official snapshot: %w", err)
	}
	finalStable := stabiliser.Offer(third)
	if !finalStable.Stable {
		why := strings.TrimSpace(finalStable.Why)
		if why == "" {
			why = "the final official snapshot changed"
		}
		return fmt.Errorf("engine reconcile-resolve: final bounded-fresh account evidence was not obtained: %s", why)
	}

	local, err := runtime.localState(ctx, runtime.accountRef)
	if err != nil {
		return fmt.Errorf("engine reconcile-resolve: reconstructing corrected local state: %w", err)
	}
	diff := runtime.compare(third, local)
	if diff.BlocksEntry() {
		return fmt.Errorf("engine reconcile-resolve: blocking reconciliation diff remains: %s", diff.Summary())
	}
	if err := runtime.validate(); err != nil {
		return err
	}
	if err := runtime.resolve(ctx, operator, note); err != nil {
		return err
	}
	if err := runtime.close(); err != nil {
		return fmt.Errorf("engine reconcile-resolve: closing the journal after release: %w", err)
	}
	closed = true

	fmt.Fprintf(cmd.OutOrStdout(), "quantity-mismatch reconcile blocks released after three stable official snapshots (lock %s)\n", lock.Path())
	return nil
}

func validateReconcileResolveRuntime(runtime reconcileResolveRuntime) error {
	switch {
	case strings.TrimSpace(runtime.accountRef) == "":
		return errors.New("engine reconcile-resolve: official account identity is empty")
	case runtime.collect == nil:
		return errors.New("engine reconcile-resolve: official snapshot collector is missing")
	case runtime.localState == nil:
		return errors.New("engine reconcile-resolve: local-state reader is missing")
	case runtime.compare == nil:
		return errors.New("engine reconcile-resolve: comparer is missing")
	case runtime.validate == nil:
		return errors.New("engine reconcile-resolve: active-cause validator is missing")
	case runtime.resolve == nil:
		return errors.New("engine reconcile-resolve: audited resolver is missing")
	case runtime.close == nil:
		return errors.New("engine reconcile-resolve: journal closer is missing")
	default:
		return nil
	}
}

func buildReconcileResolveRuntime(ctx context.Context, root *rootOptions) (reconcileResolveRuntime, error) {
	cfg, err := loadConfig(root)
	if err != nil {
		return reconcileResolveRuntime{}, err
	}
	credFile, tokenFile, err := resolveOpenAPIPaths(root)
	if err != nil {
		return reconcileResolveRuntime{}, err
	}
	creds, err := official.LoadCredentials(os.Getenv, credFile)
	if err != nil {
		return reconcileResolveRuntime{}, fmt.Errorf("engine reconcile-resolve: loading official credentials: %w", err)
	}
	if creds == nil || strings.TrimSpace(creds.APIKey) == "" || strings.TrimSpace(creds.SecretKey) == "" {
		return reconcileResolveRuntime{}, engine.ErrOfficialCredentialsRequired
	}

	client := official.New(*creds, tokenFile)
	accounts, err := client.Accounts(ctx)
	if err != nil {
		return reconcileResolveRuntime{}, fmt.Errorf("engine reconcile-resolve: resolving the official account: %w", err)
	}
	accountRef, err := reconcileResolveAccountRef(client, accounts)
	if err != nil {
		return reconcileResolveRuntime{}, err
	}
	var reads engine.OfficialReads = client

	journalPath := ""
	if root != nil && strings.TrimSpace(root.configDir) != "" {
		journalPath = filepath.Join(root.configDir, journal.DBFileName)
	}
	j, err := journal.Open(ctx, journal.Options{Path: journalPath, Clock: clock.System()})
	if err != nil {
		return reconcileResolveRuntime{}, fmt.Errorf("engine reconcile-resolve: opening the engine journal: %w", err)
	}

	tracker := &reconcile.Tracker{Journal: j, AccountRef: accountRef}
	if err := tracker.Restore(ctx); err != nil {
		_ = j.Close()
		return reconcileResolveRuntime{}, err
	}
	currency := strings.ToUpper(strings.TrimSpace(cfg.Engine.AutomationGate.LimitCurrency))
	if currency == "" {
		currency = engine.DefaultLimitCurrency
	}
	sweep := reconcileResolveAccountSweep{reads: reads}
	collector := &reconcile.Collector{
		Orders:     execgw.OfficialOrders{Client: reads},
		Positions:  sweep,
		Balance:    sweep,
		Currencies: []string{currency},
		AccountRef: accountRef,
		Clock:      clock.System(),
	}
	return reconcileResolveRuntime{
		accountRef: accountRef,
		collect:    collector.Collect,
		localState: func(ctx context.Context, account string) (reconcile.LocalState, error) {
			return reconcile.LocalStateFromJournal(ctx, j, account)
		},
		compare: reconcile.Comparer{}.Compare,
		validate: func() error {
			blocks := tracker.Blocks()
			if len(blocks) == 0 {
				return errors.New("engine reconcile-resolve: no active RECONCILE block exists")
			}
			for _, block := range blocks {
				if block.Cause != journal.ReconcileCauseQuantityMismatch {
					return fmt.Errorf(
						"engine reconcile-resolve: active cause %s on %s requires cause-specific recovery evidence",
						block.Cause, block.Symbol)
				}
			}
			return nil
		},
		resolve: tracker.Resolve,
		close:   j.Close,
	}, nil
}

func reconcileResolveAccountRef(client *official.Client, accounts []domain.Account) (string, error) {
	if len(accounts) == 0 {
		return "", errors.New("engine reconcile-resolve: the broker returned no accounts")
	}
	first := accounts[0]
	accountRef := strings.TrimSpace(first.DisplayName)
	if accountRef == "" {
		return "", errors.New("engine reconcile-resolve: the broker's first account record has no account number")
	}
	seq, err := strconv.Atoi(first.ID)
	if err != nil || seq <= 0 {
		return "", fmt.Errorf("engine reconcile-resolve: the broker's first account record has invalid account sequence %q", first.ID)
	}
	selected, ok := client.SelectedAccountSeq()
	if !ok || selected != seq {
		return "", fmt.Errorf("engine reconcile-resolve: selected official account sequence %d does not match first account %d", selected, seq)
	}
	return accountRef, nil
}

type reconcileResolveAccountSweep struct{ reads engine.OfficialReads }

func (s reconcileResolveAccountSweep) Positions(ctx context.Context) ([]domain.Position, error) {
	return s.reads.Holdings(ctx, "")
}

func (s reconcileResolveAccountSweep) PositionsRaw(ctx context.Context) ([]reconcile.RawHolding, error) {
	items, err := s.reads.HoldingsRaw(ctx, "")
	if err != nil {
		return nil, err
	}
	out := make([]reconcile.RawHolding, 0, len(items))
	for _, item := range items {
		out = append(out, reconcile.RawHolding{
			Symbol: item.Symbol, Market: item.MarketCountry,
			Quantity: item.Quantity, AveragePrice: item.AveragePurchasePrice,
		})
	}
	return out, nil
}

func (s reconcileResolveAccountSweep) BuyingPower(ctx context.Context, currency string) (float64, error) {
	bp, err := s.reads.BuyingPower(ctx, currency)
	if err != nil {
		return 0, err
	}
	return bp.CashBuyingPower, nil
}

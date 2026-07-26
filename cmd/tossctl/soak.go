package main

// soak.go is `tossctl soak`: the read-only capability survey whose output the
// engine's automation gate is interlocked on (openspec change
// verify-execution-capability, task 1.1).
//
// # Why this is a tossctl subcommand and not a separate binary
//
// The survey needs credentials, path resolution and an Open API client, and all
// three already exist here and are already tested. A second binary would either
// duplicate them or import them anyway, and it would be one more thing for the
// operator to build, find and keep up to date for a job they run once.
//
// # What makes it read-only, mechanically
//
// The survey itself lives in internal/soak and reaches the broker through
// soak.Reads — six methods, all GETs, no intent type anywhere in the signature.
// That package never imports internal/official, so it cannot call PlaceOrder even
// if somebody later tries; internal/soak/static_test.go asserts the import graph.
//
// This file is the one place that holds a real client, and it holds it as
// soakOfficialReads: an interface with the read methods and nothing else. The
// same static test greps this file for the mutating method names, and
// TestSoakIssuesNoMutatingRequest watches the actual HTTP verbs against a fake
// server.
//
// # No --base-url
//
// `soak attest` writes the file that decides whether unattended order placement
// is permitted. A flag that pointed the survey at an arbitrary server would let
// anyone produce a passing attestation from a server they wrote themselves, so
// there is no such flag. Tests replace soakReadsFactory instead, which is not
// reachable from the command line. The base URL that was actually surveyed is
// recorded in the attestation's notes.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
	"github.com/spf13/cobra"
)

type soakOptions struct {
	interval      time.Duration
	cycles        int
	symbols       []string
	currency      string
	maxOrderPages int
	record        string

	minDays  int
	validity time.Duration

	out        string
	verifiedBy string
	notes      string
}

func newSoakCmd(root *rootOptions) *cobra.Command {
	opts := &soakOptions{}

	cmd := &cobra.Command{
		Use:   "soak",
		Short: "Run and report the read-only capability soak behind the automation gate",
		Long: strings.TrimSpace(`
Survey the official Open API's read endpoints on a timer, for days, and record
what happened.

The automation gate refuses to start without a capability attestation: a local
record that the credentials renew themselves unattended, that the reads work, and
that the order list is complete. Nothing in a build can establish that, so this
command measures it.

It is read-only. It places no order, cancels nothing, and issues no request other
than GETs and the OAuth token exchange.

  tossctl soak run       start (or resume) the survey
  tossctl soak status    how far it has got, and what is still missing
  tossctl soak attest    write the attestation, if the record has earned it`),
	}

	cmd.AddCommand(
		newSoakRunCmd(root, opts),
		newSoakStatusCmd(root, opts),
		newSoakAttestCmd(root, opts),
	)
	return cmd
}

func newSoakRunCmd(root *rootOptions, opts *soakOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Survey the read endpoints on a timer and append the results to the soak record",
		Long: strings.TrimSpace(`
Walk the account, balance, holdings, order-list and quote endpoints once per
interval, and append what happened to the soak record.

Each cycle records three things:

  credentials    did the authenticated read work, and did the cached access
                 token's expiry move forward (an unattended refresh)
  endpoints      success, latency and throttling per endpoint
  completeness   does the order list paginate without loops or duplicates, does
                 it contain every order the broker reports as open, does a quote
                 request return a quote per symbol

Leave it running. Ctrl-C stops it and keeps everything already recorded; running
it again appends to the same record, so a reboot costs a cycle, not a day.

Three consecutive days is the bar (execution-verification). At the default
15-minute interval that is roughly 288 cycles.

While a live verification is running (` + "`tossctl verify run`" + ` or ` + "`tossctl console`" + `)
this command delays its cycle instead of competing for the same rate limit — the
two share one account and one budget, and a verification step lost to a 429 costs
a real order. The marker is a file beside the evidence record; a stale one (over
five minutes untouched) is ignored, so a crashed verification cannot wedge the
survey.`),
		// official: every read goes to the Open API. Not mutating: the survey
		// issues no request that can change an account.
		Annotations:  map[string]string{"source": "official"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSoakRun(cmd, root, opts)
		},
	}

	cmd.Flags().DurationVar(&opts.interval, "interval", 15*time.Minute,
		"Wait between cycles")
	cmd.Flags().IntVar(&opts.cycles, "cycles", 0,
		"Stop after this many cycles (0 runs until interrupted)")
	cmd.Flags().StringSliceVar(&opts.symbols, "symbols", []string{"005930"},
		"Symbols the quote read asks for")
	cmd.Flags().StringVar(&opts.currency, "currency", "KRW",
		"Currency for the buying-power read")
	cmd.Flags().IntVar(&opts.maxOrderPages, "max-order-pages", 25,
		"Stop the order-list walk after this many pages")
	cmd.Flags().StringVar(&opts.record, "record", "",
		"Override the soak record path")

	return cmd
}

func newSoakStatusCmd(root *rootOptions, opts *soakOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report how far the soak has got and what is still missing",
		Long: strings.TrimSpace(`
Read the soak record and report the consecutive-day streak, per-endpoint success
rates and latencies, throttling, completeness failures, and — when it is not yet
finished — every reason an attestation would be refused.

Reads only the local record. It makes no network call.`),
		Annotations:  map[string]string{"source": "local"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSoakStatus(cmd, root, opts)
		},
	}
	cmd.Flags().StringVar(&opts.record, "record", "", "Override the soak record path")
	cmd.Flags().IntVar(&opts.minDays, "min-days", 3, "Consecutive days of unattended refresh required")
	return cmd
}

func newSoakAttestCmd(root *rootOptions, opts *soakOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attest",
		Short: "Write the capability attestation, if the soak record has earned it",
		Long: strings.TrimSpace(`
Judge the soak record and, only if every criterion is met, write the capability
attestation the engine's automation gate reads.

The criteria:

  - the required consecutive days of unattended credential refresh, with no
    refused token on any of them
  - the access token observed renewing itself
  - every surveyed read endpoint succeeded at least once inside that window
  - no read-completeness failure inside that window
  - the record describes exactly one account

If any of them is unmet, nothing is written and each reason is printed.

The attestation covers reads only. The order-placement and cancel endpoints the
engine also requires cannot be proven by a read-only tool; they come from the
supervised one-off live check in the same change, so the gate will still refuse
to start until that has been done.`),
		Annotations:  map[string]string{"source": "local"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSoakAttest(cmd, root, opts)
		},
	}
	cmd.Flags().StringVar(&opts.record, "record", "", "Override the soak record path")
	cmd.Flags().IntVar(&opts.minDays, "min-days", 3, "Consecutive days of unattended refresh required")
	cmd.Flags().DurationVar(&opts.validity, "validity", 30*24*time.Hour,
		"How long the attestation is trusted for")
	cmd.Flags().StringVar(&opts.out, "out", "", "Override where the attestation is written")
	cmd.Flags().StringVar(&opts.verifiedBy, "verified-by", "", "Who ran the verification, for the audit trail")
	cmd.Flags().StringVar(&opts.notes, "notes", "", "Free-form operator context recorded in the attestation")
	return cmd
}

// --- run --------------------------------------------------------------------

func runSoakRun(cmd *cobra.Command, root *rootOptions, opts *soakOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	// A soak runs for days and is ended by hand. Ctrl-C has to close the record
	// cleanly rather than leave the last cycle half-written.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	wiring, err := soakReadsFactory(root)
	if err != nil {
		return err
	}

	recordPath, err := resolveSoakRecord(root, opts.record)
	if err != nil {
		return err
	}
	recorder, err := soak.OpenRecorder(recordPath)
	if err != nil {
		return err
	}
	defer recorder.Close()

	// The soak and the live verification share one account, one credential and one
	// rate limit; on 2026-07-26 they overlapped and the verification lost three
	// steps to a 429 (measurements.md M4). The survey yields — see
	// internal/runlock. Resolved from the verify record's location so an isolated
	// --config-dir profile gets its own marker.
	verifyRecord, err := resolveVerifyRecord(root, "")
	if err != nil {
		return err
	}
	lockPath := verifyRunLockPath(verifyRecord)

	out := cmd.OutOrStdout()
	runner, err := soak.New(soak.Options{
		Reads:         wiring.reads,
		Recorder:      recorder,
		Interval:      opts.interval,
		Cycles:        opts.cycles,
		Symbols:       opts.symbols,
		Currency:      opts.currency,
		MaxOrderPages: opts.maxOrderPages,
		Classify:      classifySoakError,
		TokenExpiry:   wiring.tokenExpiry,
		PauseWhile:    verifyRunLockPause(lockPath),
		// The survey runs for days across reinstalls. At each cycle boundary it
		// compares the executable at its own path against the one it was loaded
		// from and hands over when they differ (task 1.8 ②). The record is
		// append-only and synced per cycle, so the successor continues the same
		// file and the same streak.
		Binary:   binstamp.Self,
		ReExec:   soakReExec,
		Progress: out,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "capability soak — record %s\n", recordPath)
	fmt.Fprintf(out, "  interval %s, %s, symbols %s\n",
		opts.interval, cycleBudget(opts.cycles), strings.Join(opts.symbols, ","))
	fmt.Fprintf(out, "  read-only: no order is placed, amended or cancelled by this command\n")
	fmt.Fprintf(out, "  pauses while a live verification holds %s\n\n", lockPath)

	if err := runner.Run(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			// Stopping is how the soak normally ends.
			fmt.Fprintf(out, "\nstopped — everything recorded so far is in %s\n", recordPath)
			fmt.Fprintln(out, "Run `tossctl soak status` to see how far it got.")
			return nil
		}
		if errors.Is(err, soak.ErrUpgraded) {
			// A real re-exec never reaches this: the process is gone. It is here
			// so a future seam that hands over some other way is not reported as
			// a crash.
			fmt.Fprintf(out, "\nhanded over to the newly installed binary — %s continues\n", recordPath)
			return nil
		}
		return err
	}

	fmt.Fprintf(out, "\ndone — %s\n", recordPath)
	fmt.Fprintln(out, "Run `tossctl soak status` to see how far it got.")
	return nil
}

func cycleBudget(cycles int) string {
	if cycles <= 0 {
		return "running until interrupted"
	}
	return fmt.Sprintf("%d cycle(s)", cycles)
}

// --- status -----------------------------------------------------------------

// soakStatusReport is the JSON shape of `soak status --output json`.
type soakStatusReport struct {
	Record  string       `json:"record"`
	Ready   bool         `json:"ready"`
	Reasons []string     `json:"reasons,omitempty"`
	Summary soak.Summary `json:"summary"`
}

func runSoakStatus(cmd *cobra.Command, root *rootOptions, opts *soakOptions) error {
	format, err := output.ParseFormat(root.outputFormat)
	if err != nil {
		return err
	}
	summary, criteria, recordPath, err := loadSoakSummary(root, opts)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	out := cmd.OutOrStdout()
	if format == output.FormatJSON {
		ready, reasons := summary.Evaluate(now, criteria)
		return output.WriteJSON(out, soakStatusReport{
			Record:  recordPath,
			Ready:   ready,
			Reasons: reasons,
			Summary: maskSoakSummary(summary),
		})
	}

	fmt.Fprintf(out, "record             %s\n", recordPath)
	summary.WriteText(out, now, criteria)
	return nil
}

// maskSoakSummary replaces the account references with their masked forms. The
// record is a local file, but its summary ends up in agent transcripts and
// support threads, and an account number is not something to hand out there.
func maskSoakSummary(s soak.Summary) soak.Summary {
	s.AccountRef = attest.Mask(s.AccountRef)
	masked := make([]string, 0, len(s.AccountRefs))
	for _, ref := range s.AccountRefs {
		masked = append(masked, attest.Mask(ref))
	}
	s.AccountRefs = masked
	return s
}

// --- attest -----------------------------------------------------------------

func runSoakAttest(cmd *cobra.Command, root *rootOptions, opts *soakOptions) error {
	summary, criteria, recordPath, err := loadSoakSummary(root, opts)
	if err != nil {
		return err
	}
	if opts.validity > 0 {
		criteria.Validity = opts.validity
	}

	notes := opts.notes
	if base := soakSurveyedBase(root); base != "" {
		notes = strings.TrimSpace(notes + " surveyed base: " + base)
	}

	attestation, err := soak.BuildAttestation(summary, criteria, time.Now().UTC(), opts.verifiedBy, notes)
	if err != nil {
		// The reasons are already in the error, one per line. Nothing is written.
		return err
	}

	path, err := resolveSoakAttestationPath(root, opts.out)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("soak: creating %s: %w", filepath.Dir(path), err)
	}
	if err := attest.Save(path, attestation); err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "capability attestation written to %s\n", path)
	fmt.Fprintf(out, "  record       %s\n", recordPath)
	fmt.Fprintf(out, "  account      %s\n", attest.Mask(attestation.AccountRef))
	fmt.Fprintf(out, "  soak days    %d\n", attestation.SoakDays)
	fmt.Fprintf(out, "  expires      %s\n", attestation.ExpiresAt.Format(time.RFC3339))
	fmt.Fprintf(out, "  endpoints    %s\n", strings.Join(attestation.Endpoints, "\n               "))
	fmt.Fprintf(out, "  observed     %.2f req/s sustained without a 429\n", attestation.RateLimitPerSecond)
	fmt.Fprintln(out, "\nThe automation gate will still refuse to start: this attestation covers reads only.")
	fmt.Fprintln(out, "Still to be proven by the supervised live check (verify-execution-capability task 2.2):")
	for _, e := range soak.LiveOnlyEndpoints() {
		fmt.Fprintf(out, "  - %s\n", e)
	}
	return nil
}

// --- shared plumbing --------------------------------------------------------

func loadSoakSummary(root *rootOptions, opts *soakOptions) (soak.Summary, soak.Criteria, string, error) {
	recordPath, err := resolveSoakRecord(root, opts.record)
	if err != nil {
		return soak.Summary{}, soak.Criteria{}, "", err
	}
	cycles, err := soak.LoadCycles(recordPath)
	if err != nil {
		return soak.Summary{}, soak.Criteria{}, recordPath, err
	}
	criteria := soak.DefaultCriteria()
	if opts.minDays > 0 {
		criteria.MinConsecutiveDays = opts.minDays
	}
	return soak.Summarize(cycles), criteria, recordPath, nil
}

// resolveSoakRecord decides where the record lives.
//
// It follows the journal's rule: an explicit --config-dir means an isolated
// profile and the record follows it, and otherwise the record is durable state
// and belongs in the data directory, not among the configuration.
func resolveSoakRecord(root *rootOptions, override string) (string, error) {
	if trimmed := strings.TrimSpace(override); trimmed != "" {
		return trimmed, nil
	}
	if root != nil && strings.TrimSpace(root.configDir) != "" {
		return filepath.Join(root.configDir, soak.FileName), nil
	}
	dir, err := journal.DataDir()
	if err != nil {
		return "", fmt.Errorf("soak: resolving the record location: %w", err)
	}
	return filepath.Join(dir, soak.FileName), nil
}

// resolveSoakAttestationPath mirrors the engine's own resolution
// (internal/app/engine.attestationPath) so the file lands where the gate looks
// for it, including when the operator has moved it in config.
func resolveSoakAttestationPath(root *rootOptions, override string) (string, error) {
	if trimmed := strings.TrimSpace(override); trimmed != "" {
		return trimmed, nil
	}
	if cfg, err := loadConfig(root); err == nil {
		if p := strings.TrimSpace(cfg.Engine.AutomationGate.AttestationFile); p != "" {
			return p, nil
		}
	}
	if root != nil && strings.TrimSpace(root.configDir) != "" {
		return filepath.Join(root.configDir, attest.FileName), nil
	}
	paths, err := config.DefaultPaths()
	if err != nil {
		return "", fmt.Errorf("soak: resolving the attestation location: %w", err)
	}
	return filepath.Join(paths.ConfigDir, attest.FileName), nil
}

// classifySoakError maps the official client's sentinels onto the soak's classes.
//
// An IP that is not on the allow list counts as an auth failure rather than a
// transport one: it stops the unattended credential from working, which is
// exactly what the consecutive-day streak is supposed to notice.
func classifySoakError(err error) soak.Class {
	switch {
	case err == nil:
		return soak.ClassOK
	case errors.Is(err, official.ErrAuth), errors.Is(err, official.ErrIPNotAllowed):
		return soak.ClassAuth
	case errors.Is(err, official.ErrRateLimited):
		return soak.ClassRateLimited
	case errors.Is(err, official.ErrTransport):
		return soak.ClassTransport
	case errors.Is(err, official.ErrServer):
		return soak.ClassServer
	default:
		return soak.ClassOther
	}
}

// --- the broker adapter -----------------------------------------------------

// soakOfficialReads is the slice of the Open API client the survey is allowed to
// see.
//
// Declaring it means the adapter below holds a value with no mutating method on
// it: an edit that reached for PlaceOrder would not compile, rather than
// compiling and being caught by a review.
type soakOfficialReads interface {
	Accounts(ctx context.Context) ([]domain.Account, error)
	BuyingPower(ctx context.Context, currency string) (domain.BuyingPower, error)
	Holdings(ctx context.Context, symbol string) ([]domain.Position, error)
	OrdersPageRaw(ctx context.Context, filter official.OrdersFilter, cursor string) (official.RawOrderPage, error)
	OrderByID(ctx context.Context, orderID string) (domain.Order, error)
	Prices(ctx context.Context, symbols []string) ([]domain.Quote, error)
}

// soakWiring is what the survey needs from this package.
type soakWiring struct {
	reads       soak.Reads
	tokenExpiry func() (time.Time, bool)
	baseURL     string
}

// soakReadsFactory builds that wiring. It is a package variable so this
// package's own tests can point the survey at an httptest server; see the note
// at the top of this file about why there is no flag for it.
var soakReadsFactory = buildSoakWiring

func buildSoakWiring(root *rootOptions) (soakWiring, error) {
	credFile, tokenFile, err := resolveOpenAPIPaths(root)
	if err != nil {
		return soakWiring{}, err
	}
	creds, err := official.LoadCredentials(os.Getenv, credFile)
	if err != nil {
		return soakWiring{}, fmt.Errorf("soak: reading the Open API credentials: %w", err)
	}
	if creds == nil {
		return soakWiring{}, fmt.Errorf(
			"soak: no Open API credentials — the survey measures whether they renew themselves unattended, "+
				"so it cannot start without them. Run `tossctl openapi login` or set %s/%s",
			"TOSSCTL_OPENAPI_KEY", "TOSSCTL_OPENAPI_SECRET")
	}
	client := official.New(*creds, tokenFile)
	return soakWiring{
		reads:       soakReads{api: client},
		tokenExpiry: func() (time.Time, bool) { return soakTokenExpiry(tokenFile) },
		baseURL:     client.BaseURL(),
	}, nil
}

// soakSurveyedBase reports the base URL the survey would use, for the
// attestation's notes. It is best-effort: a missing credential here only costs
// the note.
func soakSurveyedBase(root *rootOptions) string {
	wiring, err := soakReadsFactory(root)
	if err != nil {
		return ""
	}
	return wiring.baseURL
}

// soakTokenExpiry reads the cached access token's expiry.
//
// Only the expiry. The token itself never leaves this function, and the soak
// record has no field that could hold it.
func soakTokenExpiry(tokenFile string) (time.Time, bool) {
	at := readTokenExpiry(tokenFile)
	if at == nil {
		return time.Time{}, false
	}
	return *at, true
}

// soakReads adapts the Open API's reads onto soak.Reads.
//
// Every method drops everything except what the survey records: a count, an
// identifier, an error. Balances and prices are read to prove the endpoint
// works, and then discarded — the record is evidence about an API, not a
// snapshot of a portfolio.
type soakReads struct{ api soakOfficialReads }

func (r soakReads) Accounts(ctx context.Context) ([]string, error) {
	accounts, err := r.api.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	refs := make([]string, 0, len(accounts))
	for _, a := range accounts {
		// DisplayName carries the official API's accountNo — the account number an
		// operator recognises, and the one the engine's interlock compares against.
		refs = append(refs, strings.TrimSpace(a.DisplayName))
	}
	return refs, nil
}

func (r soakReads) BuyingPower(ctx context.Context, currency string) error {
	_, err := r.api.BuyingPower(ctx, currency)
	return err
}

func (r soakReads) Holdings(ctx context.Context) (int, error) {
	positions, err := r.api.Holdings(ctx, "")
	if err != nil {
		return 0, err
	}
	return len(positions), nil
}

func (r soakReads) OrdersPage(ctx context.Context, status, cursor string) (soak.OrderPage, error) {
	page, err := r.api.OrdersPageRaw(ctx, official.OrdersFilter{Status: status}, cursor)
	if err != nil {
		return soak.OrderPage{}, err
	}
	out := soak.OrderPage{NextCursor: page.NextCursor, HasNext: page.HasNext}
	for _, raw := range page.Orders {
		var head struct {
			OrderID string `json:"orderId"`
		}
		if err := json.Unmarshal(raw, &head); err != nil {
			// A page entry that does not carry an identifier is itself a
			// completeness problem, and the walk should see it as one rather than
			// abort: an empty id can never match an open order, so the coverage
			// check reports it.
			out.IDs = append(out.IDs, "")
			continue
		}
		out.IDs = append(out.IDs, head.OrderID)
	}
	return out, nil
}

func (r soakReads) Order(ctx context.Context, id string) error {
	_, err := r.api.OrderByID(ctx, id)
	return err
}

func (r soakReads) Prices(ctx context.Context, symbols []string) (int, error) {
	quotes, err := r.api.Prices(ctx, symbols)
	if err != nil {
		return 0, err
	}
	return len(quotes), nil
}

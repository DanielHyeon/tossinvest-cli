package main

// candidate.go is the operator surface for discovery: one scan, and the loop.
//
// # Both commands are read-only and the annotation says so
//
// `mutating: false`. Discovery reads rankings and writes its own store, and the
// interface it is handed — internal/candidate.Source — has one method that takes a
// market and returns rows. There is no way to express an order through it, which is
// the point: internal/candidate's dependency closure is {internal/clock}, and the
// clients that could place something live behind internal/candidatesrc's adapters.
//
// # What the output is for
//
// Under D18 two of the three chase vetoes have no approved threshold anywhere in
// the repository, so `Chase.Passed()` is unreachable and the passed count is
// permanently zero. That is not an outage — it is what D18, D10 and the spec's
// "세 사유가 모두 측정되었고 모두 위험하지 않을 때만" produce together — and the one
// wrong repair is to count THRESHOLD_ABSENT as a pass.
//
// So the rendering is built around a single obligation: a reader must not be able to
// mistake "we could not check" for "we checked and it is fine". The unmeasured count
// leads, the passed count carries the sentence explaining why it is zero, and the
// shadow record is labelled as a record rather than as a result. spec Requirement 8:
// 화면의 기본 독해는 "대부분 안전"이 아니라 "대부분 미확인"이어야 한다.
//
// # And it never types a confirmation at anybody
//
// Read-only surfaces do not manufacture friction. A refusal says why and stops.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tossclient "github.com/JungHoonGhae/tossinvest-cli/internal/client"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/session"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/candidatesrc"
	"github.com/spf13/cobra"
)

// errVerificationHoldsTheBudget is the refusal task 5.3 asks for.
//
// spec Requirement 7 ranks engine > verification > discovery. The verification
// places real orders with a person watching and a step lost to a 429 costs another
// one; a discovery cycle costs nothing and there will be another in fifteen seconds.
var errVerificationHoldsTheBudget = errors.New(
	"candidate watch: a live account verification is running, and it has the account's rate budget")

type candidateOptions struct {
	market   string
	interval time.Duration
	cycles   int
}

// candidateStoreFactory opens the discovery store and hands back the release for
// what it opened. A package variable so this package's own tests can point the
// commands at a temporary one; there is no flag for it, because a path flag on a
// read-only command is one more way for an operator to end up with two stores and
// half a history in each.
//
// The release is part of the contract rather than left to the caller's judgement
// about ownership. The store this returns is a SQLite handle in WAL mode, and a
// handle nobody closes leaves the write-ahead log un-checkpointed beside the order
// ledger — on the filesystem D16 says discovery has to be the first to stop
// writing to. A fixture that owns its own store returns a release that does
// nothing, which is the only place the two differ.
//
// The context is the caller's: Open migrates, and a request the browser has
// already abandoned should not still be running a schema ladder (§5 review P2).
var candidateStoreFactory = openCandidateStore

// candidatePanelFactory builds a market's source panel, same seam and same reason.
var candidatePanelFactory = buildCandidatePanel

func newCandidateCmd(root *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "candidate",
		Short: "Observe what the market's ranking sources are saying, over time",
	}
	cmd.AddCommand(newCandidateScanCmd(root), newCandidateWatchCmd(root))
	return cmd
}

func newCandidateScanCmd(root *rootOptions) *cobra.Command {
	opts := &candidateOptions{market: candidate.MarketKR}
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Run one discovery scan and report what it saw and what it could not measure",
		Long: strings.TrimSpace(`
Runs a single pass over a market's ranking sources, records the readings, and
reports the chase-veto verdict for every live candidate.

Nothing is placed, amended or cancelled. The scan reads the official rankings (and
the WTS lists when a session exists), writes to the discovery store beside the
ledger, and prints.

Read the output from the bottom of the veto block up. Two of the three chase
vetoes — seen_late and extended — have no approved threshold anywhere in this
repository yet, so no candidate can be counted as having cleared all three and the
passed count is structurally zero. The number that means something today is how
many candidates could not be measured, and why. The shadow crossings and the
shadow bands are the record a threshold will eventually be derived from; they
decide nothing.`),
		// both: the official rankings are required and the WTS lists are additive.
		// Not mutating: see the file header.
		Annotations:  map[string]string{"source": "both", "mutating": "false"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCandidateScan(cmd, root, opts)
		},
	}
	cmd.Flags().StringVar(&opts.market, "market", candidate.MarketKR, "Market to scan: KR or US")
	return cmd
}

func newCandidateWatchCmd(root *rootOptions) *cobra.Command {
	opts := &candidateOptions{market: candidate.MarketKR}
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Repeat the discovery scan on an interval, enforcing retention as it goes",
		Long: strings.TrimSpace(fmt.Sprintf(`
Repeats the scan until interrupted. Nothing is placed, amended or cancelled.

The interval has a floor of %s and defaults to %s; a shorter one is raised, because
discovery shares one account and one rate budget with the engine and the live
verification, and being read-only does not make it free.

It is also never shorter than the schedule allows: the loop waits until the first
source may next be read rather than waking to find nothing due. Each source has
its own interval and the engine yield doubles them, so a tick that once matched
those numbers stops matching them the moment the engine starts.

Three things happen on every turn besides the scan:

  retention   raw observations past their window are pruned and the write-ahead
              log is reclaimed. This store sits on the same filesystem as the order
              ledger, and a cleanup nobody runs is unbounded growth on it.
  free space  below the floor, discovery stops writing and says so. When that
              filesystem fills the write that must not fail is the ledger's.
  the engine  while it is running, discovery lengthens its per-source intervals.
              The engine's delay reaches stop-loss immediacy through fill
              detection, so this yield is code rather than an operator's judgement.

The command refuses to start while a live account verification holds its run lock.`,
			candidate.MinWatchInterval, candidate.DefaultWatchInterval)),
		Annotations:  map[string]string{"source": "both", "mutating": "false"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCandidateWatch(cmd, root, opts)
		},
	}
	cmd.Flags().StringVar(&opts.market, "market", candidate.MarketKR, "Market to scan: KR or US")
	cmd.Flags().DurationVar(&opts.interval, "interval", candidate.DefaultWatchInterval,
		"How often to scan; raised to the floor if shorter, and to the instant a source may next be read")
	cmd.Flags().IntVar(&opts.cycles, "cycles", 0,
		"Stop after this many scans; 0 runs until interrupted")
	return cmd
}

// --- the two commands ---------------------------------------------------------------

func runCandidateScan(cmd *cobra.Command, root *rootOptions, opts *candidateOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	market, err := candidateMarket(opts.market)
	if err != nil {
		return err
	}
	store, sources, cleanup, err := candidateWiring(ctx, root, market)
	if err != nil {
		return err
	}
	defer cleanup()

	cycle, err := candidate.Cycle(ctx, store, candidateCycleOptions(root, market, sources,
		candidate.NewSchedule(candidateIntervals()), candidate.NewBackoff()))
	if err != nil {
		return err
	}
	return writeCandidateReport(cmd.OutOrStdout(), candidateFormat(root), cycle)
}

func runCandidateWatch(cmd *cobra.Command, root *rootOptions, opts *candidateOptions) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	market, err := candidateMarket(opts.market)
	if err != nil {
		return err
	}

	// The live verification outranks discovery and the run lock is the mechanism.
	// Checked before anything is opened: a refusal about the rate budget should not
	// leave a store file behind it.
	//
	// A path that cannot be resolved leaves the gate un-run, and that is reported
	// rather than passed over. Fail-open is the right direction here — a discovery
	// loop refusing to start because a data directory is misconfigured is worse than
	// one that starts beside a verification, and the loop is read-only — but a gate
	// that skips itself in silence is indistinguishable from a gate that ran and
	// found nothing, which is the one thing spec Requirement 7's ordering must not
	// be able to look like.
	lockPath, lerr := candidateVerifyLockPath(root)
	switch {
	case lerr != nil:
		fmt.Fprintf(cmd.ErrOrStderr(),
			"note: whether a live account verification is running could not be checked (%v). "+
				"Discovery is starting anyway because it is read-only, but the priority "+
				"engine > verification > discovery is not being enforced on this run\n", lerr)
	default:
		if fresh, at := runlock.Fresh(lockPath, time.Now(), runlock.StaleAfter); fresh {
			return fmt.Errorf("%w (%s, last touched %s). Discovery is read-only and late costs "+
				"nothing; a verification step lost to a rate limit costs a real order and a "+
				"person's attention",
				errVerificationHoldsTheBudget, lockPath, at.UTC().Format(time.RFC3339))
		}
	}

	store, sources, cleanup, err := candidateWiring(ctx, root, market)
	if err != nil {
		return err
	}
	defer cleanup()

	out := cmd.OutOrStdout()
	format := candidateFormat(root)
	interval := candidate.WatchInterval(opts.interval)
	if format == output.FormatTable {
		fmt.Fprintf(out, "candidate watch — %s every %s\n", market, interval)
		fmt.Fprintf(out, "  read-only: no order is placed, amended or cancelled by this command\n")
		fmt.Fprintf(out, "  retention and write-ahead log reclamation run inside every cycle\n\n")
	}

	var writeErr error
	err = candidate.Watch(ctx, store, candidate.WatchOptions{
		CycleOptions: candidateCycleOptions(root, market, sources,
			candidate.NewSchedule(candidateIntervals()), candidate.NewBackoff()),
		Interval: interval,
		Cycles:   opts.cycles,
		OnCycle: func(res candidate.CycleResult) {
			if writeErr == nil {
				writeErr = writeCandidateReport(out, format, res)
			}
		},
		// A market that could not be read for one tick is not a reason to end a
		// session; a store that cannot be written is.
		//
		// That was the intention and the wiring did not carry it. Cycle propagates
		// ErrNoSourceAnswered, so the one failure named here as survivable was the
		// one that ended the loop — and once the loop is gone nobody promotes, the
		// implicit cooling clock runs at last_seen_at + 10m and the expiry 30
		// minutes after that, so every first_seen_at in the store goes inside about
		// forty minutes. It is reachable without a market outage: a panel whose
		// sources are all inside their own 429 backoff answers exactly this, during
		// the rate limit the ladder exists to ride out.
		//
		// A turn on which the schedule had nothing due does not arrive here at all
		// any more — CycleResult.Quiet — so what is left is a genuine read failure,
		// and the message says the loop is continuing so it cannot be mistaken for
		// the kind that stops one.
		OnError: func(cycleErr error) bool {
			if errors.Is(cycleErr, candidate.ErrNoSourceAnswered) {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"cycle failed, still running: %v. The next turn is in %s; nothing was "+
						"cooled, because a source that did not answer does not vouch\n",
					cycleErr, interval)
				return true
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "cycle failed: %v\n", cycleErr)
			return false
		},
	})
	switch {
	case writeErr != nil:
		return writeErr
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		// Interrupting is how a watch normally ends.
		if format == output.FormatTable {
			fmt.Fprintln(out, "\nstopped — everything recorded so far is in the discovery store")
		}
		return nil
	}
	return err
}

// candidateWiring opens the store and builds the panel, returning a cleanup that is
// safe to defer.
//
// The cleanup is the store factory's own release, so the store is closed by whoever
// opened it and a fixture that owns its store is left alone. It used to be
// `func() {}` while this comment described the release — so `scan` and `watch`,
// the two commands whose whole argument is that they clean up after themselves on
// every turn, both leaked the handle. An interrupted `watch` left an
// un-checkpointed write-ahead log beside the ledger.
//
// A panel failure releases the store before returning: the caller is being handed
// an error, so nothing will defer the cleanup for it.
func candidateWiring(ctx context.Context, root *rootOptions, market string) (
	*candidate.Store, []candidate.Source, func(), error) {

	store, release, err := candidateStoreFactory(ctx, root)
	if err != nil {
		return nil, nil, func() {}, err
	}
	sources, err := candidatePanelFactory(root, market)
	if err != nil {
		release()
		return nil, nil, func() {}, err
	}
	if len(sources) == 0 {
		release()
		return nil, nil, func() {}, fmt.Errorf(
			"candidate: no source can serve %s. The official Open API rankings are the "+
				"required source (run `tossctl openapi login`); the WTS lists are additive",
			market)
	}
	return store, sources, release, nil
}

// candidateCycleOptions assembles one turn's configuration.
//
// The thresholds come from candidateVetoThresholds rather than from a literal here.
// There used to be a literal here and another in the /signals seam, and nothing held
// them together: editing one is the ordinary way to change a threshold, and the
// result is two screens answering differently about the same store at the same
// instant with nothing failing. vetothresholds.go carries the argument and the
// values.
func candidateCycleOptions(root *rootOptions, market string, sources []candidate.Source,
	sched *candidate.Schedule, backoff *candidate.Backoff) candidate.CycleOptions {

	return candidate.CycleOptions{
		Market:        market,
		Sources:       sources,
		Schedule:      sched,
		Backoff:       backoff,
		Thresholds:    candidateVetoThresholds(),
		EngineRunning: candidateEngineProbe(root),
	}
}

// candidateIntervals is the contract's per-source cadence (design.md, D13).
//
// Per source rather than per market, because the limits differ by orders of
// magnitude: the WTS lists have no limit anybody has observed, and the official
// RANKING group is not in the published limit table at all. One interval per market
// ties every source to the strictest one and throws away the earliness this whole
// change buys.
func candidateIntervals() map[candidate.SourceID]candidate.Interval {
	official := candidate.Interval{Every: 15 * time.Second, Floor: 5 * time.Second}
	wts := candidate.Interval{Every: 5 * time.Second, Floor: 3 * time.Second}
	return map[candidate.SourceID]candidate.Interval{
		candidate.SourceOfficialTradingValue:  official,
		candidate.SourceOfficialTradingVolume: official,
		candidate.SourceOfficialGainers:       official,
		candidate.SourceWTSPopular:            wts,
	}
}

// candidateEngineProbe answers "is the trading engine running" from its advisory
// marker.
//
// A function rather than a package hand-off: internal/enginelock is outside
// internal/candidate's dependency closure, and that closure is the actual lock on
// the isolation requirement (D7). Discovery is handed the answer, never the reader.
//
// A marker path that cannot be resolved yields nil, which Cycle records as "nobody
// was watching" rather than as "no engine" — the two are different, and only one of
// them means it is safe not to yield.
func candidateEngineProbe(root *rootOptions) func(time.Time) bool {
	dir, err := engineJournalDir(root)
	if err != nil {
		return nil
	}
	marker := enginelock.MarkerPath(dir)
	return func(at time.Time) bool { return enginelock.Read(marker, at).Running }
}

// candidateVerifyLockPath is verify.go's own resolution, reached through the same
// helpers so the two commands cannot disagree about which file they mean.
func candidateVerifyLockPath(root *rootOptions) (string, error) {
	record, err := resolveVerifyRecord(root, "")
	if err != nil {
		return "", err
	}
	return verifyRunLockPath(record), nil
}

func candidateMarket(market string) (string, error) {
	switch m := strings.ToUpper(strings.TrimSpace(market)); m {
	case candidate.MarketKR, candidate.MarketUS:
		return m, nil
	default:
		return "", fmt.Errorf("candidate: unknown market %q (want %s or %s)",
			market, candidate.MarketKR, candidate.MarketUS)
	}
}

func candidateFormat(root *rootOptions) output.Format {
	if root != nil && strings.EqualFold(strings.TrimSpace(root.outputFormat), string(output.FormatJSON)) {
		return output.FormatJSON
	}
	return output.FormatTable
}

// --- production wiring --------------------------------------------------------------

// openCandidateStore resolves the discovery store's location.
//
// It follows the verification record's rule so an isolated --config-dir profile gets
// its own store and the default one sits in the data directory beside the ledger —
// a separate file and a separate lock, which is D2's whole arrangement.
func openCandidateStore(ctx context.Context, root *rootOptions) (*candidate.Store, func(), error) {
	path := ""
	if root != nil && strings.TrimSpace(root.configDir) != "" {
		path = filepath.Join(root.configDir, candidate.DBFileName)
	}
	store, err := candidate.Open(ctx, candidate.Options{
		Path:  path,
		Clock: clock.System(),
	})
	if err != nil {
		return nil, func() {}, err
	}
	return store, func() { _ = store.Close() }, nil
}

// buildCandidatePanel assembles the market's sources from whatever credentials
// exist.
//
// The official rankings are required and the WTS popularity list is additive, which
// is spec Requirement 5's hard half: WTS sessions expire on ordinary days, so a
// configuration that stops without one stops routinely.
func buildCandidatePanel(root *rootOptions, market string) ([]candidate.Source, error) {
	credFile, tokenFile, err := resolveOpenAPIPaths(root)
	if err != nil {
		return nil, err
	}
	creds, err := official.LoadCredentials(os.Getenv, credFile)
	if err != nil {
		return nil, fmt.Errorf("candidate: reading the Open API credentials: %w", err)
	}
	if creds == nil {
		return nil, fmt.Errorf(
			"candidate: no Open API credentials. The official rankings are discovery's " +
				"required source — run `tossctl openapi login`")
	}
	client := official.New(*creds, tokenFile)

	// The WTS half is optional and its absence is ordinary. A missing or unreadable
	// session degrades the panel; it does not stop it.
	var wts *tossclient.Client
	if sess, serr := session.NewFileStore(resolveSessionFile(root)).Load(context.Background()); serr == nil && sess != nil {
		if cfg, cerr := loadConfig(root); cerr == nil {
			wts = tossclient.New(tossclient.Config{Session: sess, TradingPolicy: cfg.Trading})
		}
	}
	return candidatesrc.Panel(market, client, client, wtsPopularityReader(wts)), nil
}

// wtsPopularityReader converts a possibly-nil client into a possibly-nil reader.
//
// A typed nil inside a non-nil interface is the classic way for "no WTS session" to
// become a nil dereference three layers down, and Panel's contract is that a nil
// reader means the source is absent from the panel.
func wtsPopularityReader(c *tossclient.Client) candidatesrc.PopularityReader {
	if c == nil {
		return nil
	}
	return c
}

// --- rendering -----------------------------------------------------------------------

// candidateReport is the machine-readable shape of one cycle.
//
// It is written out rather than marshalled from candidate.CycleResult because the
// verdict types are deliberately opaque — VetoState's fields are unexported so that
// no literal outside its package can produce a state that reads as measured and safe,
// and encoding/json would render every one of them as `{}`. A JSON consumer reading
// `{}` for a veto would draw exactly the conclusion the type exists to prevent.
type candidateReport struct {
	Market     string `json:"market"`
	At         string `json:"at"`
	Halted     bool   `json:"halted"`
	HaltReason string `json:"halt_reason,omitempty"`
	// Quiet reports that the schedule had nothing due, so no source was read.
	//
	// It is a field of its own rather than a consumer's inference from
	// `sources.attempted == 0`, because those two facts used to be the same
	// rendering and one of them is an outage. See candidate.CycleResult.Quiet.
	Quiet bool `json:"quiet"`

	Sources struct {
		Attempted int      `json:"attempted"`
		Responded int      `json:"responded"`
		Degraded  bool     `json:"degraded"`
		NotDue    []string `json:"not_due,omitempty"`
		Missing   []struct {
			Source      string `json:"source"`
			Reason      string `json:"reason"`
			RateLimited bool   `json:"rate_limited"`
		} `json:"missing,omitempty"`
	} `json:"sources"`

	Backoff []struct {
		Source  string `json:"source"`
		Step    int    `json:"step"`
		Seconds int    `json:"seconds"`
		Until   string `json:"until"`
	} `json:"backoff,omitempty"`
	EngineYield bool `json:"engine_yield"`

	Recorded struct {
		Observations int `json:"observations"`
		Candidates   int `json:"candidates"`
		Cooled       int `json:"cooled"`
		FirstPrices  int `json:"first_prices"`
		FirstRanks   int `json:"first_ranks"`
		Rejected     int `json:"rejected"`
	} `json:"recorded"`

	Veto struct {
		Total      int    `json:"total"`
		Passed     int    `json:"passed"`
		PassedNote string `json:"passed_note,omitempty"`
		// Alarms is the tally's own arithmetic, failed. It is a list rather than a
		// boolean because each entry carries the numbers that produced it, and a
		// consumer told only that something is inconsistent has nothing to check.
		Alarms      []string       `json:"alarms,omitempty"`
		Vetoed      int            `json:"vetoed"`
		Unmeasured  int            `json:"unmeasured"`
		Raised      map[string]int `json:"raised"`
		NotMeasured map[string]int `json:"not_measured"`
		Reasons     map[string]int `json:"reasons"`
	} `json:"veto"`

	ShadowAcceleration struct {
		Note        string         `json:"note"`
		Total       int            `json:"total"`
		Computed    int            `json:"computed"`
		Crossed     map[string]int `json:"crossed"`
		NotComputed map[string]int `json:"not_computed"`
	} `json:"shadow_acceleration"`

	ShadowBands map[string]struct {
		Note        string         `json:"note"`
		Total       int            `json:"total"`
		Measured    int            `json:"measured"`
		Crossed     map[string]int `json:"crossed"`
		NotMeasured map[string]int `json:"not_measured"`
	} `json:"shadow_bands"`

	Retention struct {
		Ran    bool  `json:"ran"`
		Pruned int64 `json:"pruned"`
		// Summaries is the other half of D11's split (task 5.8). It is reported
		// separately from Pruned because the two mean different things: raw rows
		// going is a busy market, summaries going is lives ending — and the
		// summaries are where first_seen_at lived.
		Summaries    int64  `json:"summaries_pruned"`
		Window       string `json:"window"`
		Checkpointed bool   `json:"checkpointed"`
		Busy         bool   `json:"wal_busy"`
		Why          string `json:"why,omitempty"`
	} `json:"retention"`

	Space struct {
		Path    string `json:"path"`
		Checked bool   `json:"checked"`
		Free    int64  `json:"free_bytes"`
		Floor   int64  `json:"floor_bytes"`
		Below   bool   `json:"below_floor"`
		Why     string `json:"why,omitempty"`
	} `json:"space"`
}

// passedNote is the sentence that stops a structurally zero count from reading as a
// failing system.
//
// D18 wrote it as a consequence to be recorded rather than repaired: 둘이
// THRESHOLD_ABSENT인 동안 Chase.Passed()는 도달 불가이고 집계의 Passed는 항상 0이다.
// 이것은 고장이 아니다. A screen that showed "0 passed" alone would send somebody
// looking for a bug, and the most natural bug to "fix" is the one D18 names as the
// only wrong repair.
const passedNote = "structurally 0: seen_late and extended have no approved threshold " +
	"(design.md D18), so no candidate can have all three vetoes measured and clear. " +
	"An absent threshold is not a pass"

// passedUnexpected is what the same slot says when the count is NOT zero.
//
// The sentence above asserts a zero, so it cannot be printed beside a figure that
// contradicts it — the console reached the same conclusion in signals.go and wrote
// the reason down there: a note contradicted by the figure next to it is a note
// the next reader stops checking against. Printing it unconditionally would have
// produced `passed 7 (structurally 0: …)`.
//
// The slot is never empty, which is the other half. The way this count becomes
// non-zero without a human having approved a threshold is somebody counting
// THRESHOLD_ABSENT as a pass — the single repair D18 forbids — and that repair
// must not be able to make the output tidier than it was before.
const passedUnexpected = "not 0: either seen_late and extended have had thresholds approved " +
	"since design.md D18, or THRESHOLD_ABSENT is being counted as a pass. The second is the " +
	"one wrong repair D18 names, so check that first"

const shadowNote = "records and decides nothing; this is what a threshold will be derived from"

// tallyAlarm is one arithmetic contradiction as this surface words it.
//
// The judgement is candidate.VetoTally.Anomalies and it is not repeated here — the
// two screens must not be able to disagree about whether the numbers add up. What
// is here is the sentence, because who is reading differs: this one is read in a
// terminal beside the counts it is about, and it names the one repair D18 forbids
// because that is the repair that produces this state without anybody approving a
// threshold.
func tallyAlarm(a candidate.TallyAnomaly) string {
	switch a.Kind {
	case candidate.TallyPassedWithoutThreshold:
		return "ALARM " + a.Arithmetic() +
			". A veto with no threshold cannot be clear, so a pass counted beside one is " +
			"THRESHOLD_ABSENT being read as a pass — the single repair D18 forbids"
	default:
		return "ALARM " + a.Arithmetic() +
			". The three buckets are disjoint by construction, so a sum above the total " +
			"means a candidate was counted as passing while one of its vetoes was not " +
			"measured. Unmeasured is not a pass (D10)"
	}
}

// candidateTallyAlarms renders every anomaly in a tally.
func candidateTallyAlarms(t candidate.VetoTally) []string {
	var out []string
	for _, a := range t.Anomalies() {
		out = append(out, tallyAlarm(a))
	}
	return out
}

func buildCandidateReport(res candidate.CycleResult) candidateReport {
	var r candidateReport
	r.Market, r.At = res.Market, res.At.UTC().Format(time.RFC3339)
	r.Halted, r.HaltReason = res.Halted, res.HaltReason
	r.Quiet = res.Quiet

	r.Sources.Attempted = res.Scan.Attempted
	r.Sources.Responded = res.Scan.Responded
	r.Sources.Degraded = res.Scan.Degraded
	for _, id := range res.NotDue {
		r.Sources.NotDue = append(r.Sources.NotDue, string(id))
	}
	for _, m := range res.Scan.Missing {
		r.Sources.Missing = append(r.Sources.Missing, struct {
			Source      string `json:"source"`
			Reason      string `json:"reason"`
			RateLimited bool   `json:"rate_limited"`
		}{Source: string(m.Source), Reason: m.Reason, RateLimited: m.RateLimited})
	}
	for _, b := range res.Backoff {
		r.Backoff = append(r.Backoff, struct {
			Source  string `json:"source"`
			Step    int    `json:"step"`
			Seconds int    `json:"seconds"`
			Until   string `json:"until"`
		}{
			Source: string(b.Source), Step: b.Step,
			Seconds: int(b.For / time.Second), Until: b.Until.UTC().Format(time.RFC3339),
		})
	}
	r.EngineYield = res.EngineYield

	r.Recorded.Observations = res.Scan.Observations
	r.Recorded.Candidates = res.Scan.Candidates
	r.Recorded.Cooled = res.Scan.Cooled
	r.Recorded.FirstPrices = res.Scan.FirstPrices
	r.Recorded.FirstRanks = res.Scan.FirstRanks
	r.Recorded.Rejected = len(res.Scan.Rejected)

	r.Veto.Total = res.Vetoes.Total
	r.Veto.Passed = res.Vetoes.Passed
	r.Veto.Vetoed = res.Vetoes.Vetoed
	r.Veto.Unmeasured = res.Vetoes.Unmeasured
	r.Veto.PassedNote = passedNote
	if res.Vetoes.Passed != 0 {
		r.Veto.PassedNote = passedUnexpected
	}
	// Beside the note, not instead of it. The note is about a count that is
	// structurally zero while no threshold is approved; this is about the tally
	// contradicting itself, which is true or false whatever the thresholds are.
	r.Veto.Alarms = candidateTallyAlarms(res.Vetoes)
	r.Veto.Raised = map[string]int{}
	r.Veto.NotMeasured = map[string]int{}
	for code, n := range res.Vetoes.Raised {
		r.Veto.Raised[string(code)] = n
	}
	for code, n := range res.Vetoes.NotMeasured {
		r.Veto.NotMeasured[string(code)] = n
	}
	r.Veto.Reasons = map[string]int{}
	for why, n := range res.Vetoes.Reasons {
		r.Veto.Reasons[string(why)] = n
	}

	r.ShadowAcceleration.Note = shadowNote
	r.ShadowAcceleration.Total = res.Crossings.Total
	r.ShadowAcceleration.Computed = res.Crossings.Computed
	r.ShadowAcceleration.Crossed = res.Crossings.Crossed
	r.ShadowAcceleration.NotComputed = map[string]int{}
	for why, n := range res.Crossings.NotComputed {
		r.ShadowAcceleration.NotComputed[string(why)] = n
	}

	r.ShadowBands = map[string]struct {
		Note        string         `json:"note"`
		Total       int            `json:"total"`
		Measured    int            `json:"measured"`
		Crossed     map[string]int `json:"crossed"`
		NotMeasured map[string]int `json:"not_measured"`
	}{}
	for code, tally := range res.Bands {
		notMeasured := map[string]int{}
		for why, n := range tally.NotMeasured {
			notMeasured[string(why)] = n
		}
		r.ShadowBands[string(code)] = struct {
			Note        string         `json:"note"`
			Total       int            `json:"total"`
			Measured    int            `json:"measured"`
			Crossed     map[string]int `json:"crossed"`
			NotMeasured map[string]int `json:"not_measured"`
		}{
			Note: shadowNote, Total: tally.Total, Measured: tally.Measured,
			Crossed: tally.Crossed, NotMeasured: notMeasured,
		}
	}

	r.Retention.Ran = res.Cleanup.Ran
	r.Retention.Pruned = res.Cleanup.Pruned
	r.Retention.Summaries = res.Cleanup.Summaries
	r.Retention.Window = res.Cleanup.Retention.String()
	r.Retention.Checkpointed = res.Cleanup.Checkpointed
	r.Retention.Busy = res.Cleanup.Busy
	r.Retention.Why = res.Cleanup.Why

	r.Space.Path = res.Space.Path
	r.Space.Checked = res.Space.Checked
	r.Space.Free = res.Space.Free
	r.Space.Floor = res.Space.Floor
	r.Space.Below = res.Space.Below()
	r.Space.Why = res.Space.Why
	return r
}

func writeCandidateReport(w io.Writer, format output.Format, res candidate.CycleResult) error {
	report := buildCandidateReport(res)
	if format == output.FormatJSON {
		body, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("candidate: rendering the scan report: %w", err)
		}
		_, err = fmt.Fprintln(w, string(body))
		return err
	}
	return writeCandidateTable(w, res, report)
}

// writeCandidateTable is the human rendering.
//
// The order is the argument. Degradation and retreat first, because they say how
// much of the market was actually looked at; the veto block next with the unmeasured
// count above the passed count; the shadow records last and labelled as records.
func writeCandidateTable(w io.Writer, res candidate.CycleResult, report candidateReport) error {
	fmt.Fprintf(w, "candidate scan — %s at %s\n", report.Market, report.At)
	fmt.Fprintf(w, "  read-only: no order is placed, amended or cancelled by this command\n")

	if res.Halted {
		fmt.Fprintf(w, "\nHALTED  %s\n", res.HaltReason)
	}

	if report.Quiet {
		// 0 attempted with no sentence beside it is the shape an operator reads as
		// an outage, and the shape the console renders as "every source answered".
		// Nobody was asked; nothing was lost.
		fmt.Fprintf(w, "\nsources        nothing was due — no source was read this turn, and "+
			"nothing was lost\n")
		fmt.Fprintf(w, "               the schedule had not come round to any of them "+
			"(the engine yield doubles the intervals; a tick shorter than every source's "+
			"interval and a clock that steps backwards do the same)\n")
	} else {
		fmt.Fprintf(w, "\nsources        %d attempted, %d responded%s\n",
			report.Sources.Attempted, report.Sources.Responded,
			map[bool]string{true: "  (degraded)"}[report.Sources.Degraded])
	}
	for _, m := range report.Sources.Missing {
		limited := ""
		if m.RateLimited {
			limited = "  [rate limited]"
		}
		fmt.Fprintf(w, "  missing      %s — %s%s\n", m.Source, m.Reason, limited)
	}
	if len(report.Sources.NotDue) > 0 {
		fmt.Fprintf(w, "  not due      %s\n", strings.Join(report.Sources.NotDue, ", "))
	}
	for _, b := range report.Backoff {
		fmt.Fprintf(w, "  backoff      %s — %ds (rung %d of %d), until %s\n",
			b.Source, b.Seconds, b.Step, len(candidate.BackoffLadder), b.Until)
	}
	if report.EngineYield {
		fmt.Fprintf(w, "  engine yield the engine is running, so the source intervals are longer\n")
	}

	fmt.Fprintf(w, "\nrecorded       %d observations, %d candidates, %d cooled\n",
		report.Recorded.Observations, report.Recorded.Candidates, report.Recorded.Cooled)
	fmt.Fprintf(w, "  firsts       %d first prices, %d first ranks newly stored\n",
		report.Recorded.FirstPrices, report.Recorded.FirstRanks)
	if report.Recorded.Rejected > 0 {
		fmt.Fprintf(w, "  rejected     %d symbol(s) the store declined\n", report.Recorded.Rejected)
		for _, r := range res.Scan.Rejected {
			fmt.Fprintf(w, "               %s — %s\n", r.Symbol, r.Reason)
		}
	}

	fmt.Fprintf(w, "\nchase veto     %d candidate(s) assessed\n", report.Veto.Total)
	fmt.Fprintf(w, "  unmeasured   %d  ← the default reading is \"mostly unchecked\", not \"mostly safe\"\n",
		report.Veto.Unmeasured)
	fmt.Fprintf(w, "  vetoed       %d\n", report.Veto.Vetoed)
	fmt.Fprintf(w, "  passed       %d  (%s)\n", report.Veto.Passed, report.Veto.PassedNote)
	for _, code := range candidate.VetoCodes {
		fmt.Fprintf(w, "  %-12s raised %d, unmeasured %d\n", string(code),
			report.Veto.Raised[string(code)], report.Veto.NotMeasured[string(code)])
	}
	for _, line := range sortedCounts(report.Veto.Reasons) {
		fmt.Fprintf(w, "  reason       %s\n", line)
	}
	for _, line := range report.Veto.Alarms {
		fmt.Fprintf(w, "  %s\n", line)
	}

	fmt.Fprintf(w, "\nshadow crossings — acceleration (%s)\n", shadowNote)
	fmt.Fprintf(w, "  measured     %d of %d series\n",
		report.ShadowAcceleration.Computed, report.ShadowAcceleration.Total)
	fmt.Fprintf(w, "  crossed      %s\n", orderedCounts(candidate.ShadowThresholds,
		report.ShadowAcceleration.Crossed))
	for _, line := range sortedCounts(report.ShadowAcceleration.NotComputed) {
		fmt.Fprintf(w, "  not computed %s\n", line)
	}

	for _, code := range []candidate.VetoCode{candidate.VetoSeenLate, candidate.VetoExtended} {
		band, ok := report.ShadowBands[string(code)]
		if !ok {
			continue
		}
		fmt.Fprintf(w, "\nshadow bands — %s (%s; not a veto)\n", string(code), shadowNote)
		fmt.Fprintf(w, "  measured     %d of %d\n", band.Measured, band.Total)
		fmt.Fprintf(w, "  crossed      %s\n", orderedCounts(candidate.BandsFor(code), band.Crossed))
		for _, line := range sortedCounts(band.NotMeasured) {
			fmt.Fprintf(w, "  not measured %s\n", line)
		}
	}

	fmt.Fprintf(w, "\nretention      %d raw row(s) and %d expired summary/summaries pruned past %s",
		report.Retention.Pruned, report.Retention.Summaries, report.Retention.Window)
	switch {
	case report.Retention.Why != "":
		fmt.Fprintf(w, "; the sweep could not finish: %s\n", report.Retention.Why)
	case report.Retention.Busy:
		fmt.Fprintf(w, "; the write-ahead log is held by a reader and was not reclaimed\n")
	default:
		fmt.Fprintf(w, "; write-ahead log reclaimed\n")
	}

	if report.Space.Checked {
		fmt.Fprintf(w, "free space     %d bytes on %s, floor %d\n",
			report.Space.Free, report.Space.Path, report.Space.Floor)
	} else {
		fmt.Fprintf(w, "free space     unmeasured on %s (%s)\n", report.Space.Path, report.Space.Why)
	}
	fmt.Fprintln(w)
	return nil
}

// orderedCounts renders a count map in the order its keys were declared, so a band
// that nobody crossed still occupies its column.
func orderedCounts(keys []string, counts map[string]int) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", k, counts[k]))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " · ")
}

// sortedCounts renders a reason map deterministically, biggest first then by name.
func sortedCounts(counts map[string]int) []string {
	type entry struct {
		name string
		n    int
	}
	var entries []entry
	for name, n := range counts {
		if n == 0 {
			continue
		}
		entries = append(entries, entry{name, n})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].n != entries[j].n {
			return entries[i].n > entries[j].n
		}
		return entries[i].name < entries[j].name
	})
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, fmt.Sprintf("%s %d", e.name, e.n))
	}
	return out
}

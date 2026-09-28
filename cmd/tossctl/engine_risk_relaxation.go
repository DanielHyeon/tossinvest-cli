package main

// engine_risk_relaxation.go 는 a066 5.5 완화(해제) 메커니즘의 운영자 표면이다(사용자 결정 2026-09-28, design D8).
//
// 원칙: 자동 경로는 조이기만 한다. 완화·해제는 OPERATOR 가 사람 승인 참조와 함께 요청하고, audit 줄이 원장 commit 앞에
// 기록되며(audit 가 안 열리거나 실패하면 아무것도 안 바뀜), 승인자가 본 상태에 결속된다 — 그 뒤 상태가 바뀌었으면
// stale 로 거절된다(동시 조이기는 보수 쪽 승리). 콘솔 버튼은 없다.
//
// 두 해제 명령은 `mutating: true` 다 — 대화형 에이전트는 자동 실행하지 않는다(AGENTS.md). 읽기 전용 `risk-latch-show`
// 가 해제가 결속할 값(잠금 번호·마지막 사건 번호·owner 상태 digest)을 보여 준다.
//
// # engine lock 을 잡지 않는다
//
// 원장 쓰기는 BEGIN IMMEDIATE 로 직렬화되고 결속·판정은 트랜잭션 안에서 다시 한다. lock 을 잡으려면 엔진을 멈춰야 하는데,
// 엔진이 멈추면 (보호가 UNWIRED 인 동안) 손절이 없다 — 해제가 손절 연속성을 깨는 조건이 되면 안 된다. 그래서
// `engine reconcile-resolve` 와 달리 이 명령들은 엔진이 도는 채로 쓴다(짧은 트랜잭션 하나).

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/spf13/cobra"
)

// riskRelaxationWriter 는 두 해제 명령이 쓰는 원장 쓰기 면이다.
type riskRelaxationWriter interface {
	ReleaseEntryLossLock(context.Context, journal.EntryLossLockReleaseRequest) (journal.EntryLossLockReleaseRecord, error)
	ReleaseRiskOverageLatch(context.Context, journal.RiskOverageLatchReleaseRequest) (journal.RiskOverageLatchReleaseRecord, error)
	Close() error
}

// riskRelaxationReader 는 읽기 전용 표면이다.
type riskRelaxationReader interface {
	ReadEntryLossLocks(context.Context, string) ([]journal.EntryLossLockView, error)
	ReadRiskOwnerLatches(context.Context, string, riskbucket.Market) ([]journal.RiskOwnerLatchView, error)
	Close() error
}

type riskRelaxationDeps struct {
	journalPath func(*rootOptions) (string, error)
	openWriter  func(context.Context, string) (riskRelaxationWriter, error)
	openReader  func(context.Context, string) (riskRelaxationReader, error)
	openAuditor func(operator string) (journal.RiskRelaxationAuditor, error)
	now         func() time.Time
}

func productionRiskRelaxationDeps() riskRelaxationDeps {
	return riskRelaxationDeps{
		journalPath: func(root *rootOptions) (string, error) {
			dir, err := engineJournalDir(root)
			if err != nil {
				return "", err
			}
			return filepath.Join(dir, journal.DBFileName), nil
		},
		openWriter: func(ctx context.Context, path string) (riskRelaxationWriter, error) {
			return journal.Open(ctx, journal.Options{Path: path, Clock: clock.System()})
		},
		openReader: func(ctx context.Context, path string) (riskRelaxationReader, error) {
			return journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: path})
		},
		// 엔진과 같은 자리(데이터 디렉터리)의 audit 로그 — interlock.go openAuditLog 와 같은 해석.
		openAuditor: func(operator string) (journal.RiskRelaxationAuditor, error) {
			dir, err := journal.DataDir()
			if err != nil {
				return nil, fmt.Errorf("resolving the audit log location: %w", err)
			}
			log, err := audit.Open(audit.Options{Path: filepath.Join(dir, audit.FileName), Subject: operator})
			if err != nil {
				return nil, err
			}
			return log, nil
		},
		now: func() time.Time { return clock.System().Now() },
	}
}

func newEngineRiskRelaxationCmds(root *rootOptions) []*cobra.Command {
	deps := productionRiskRelaxationDeps()
	return []*cobra.Command{
		newEngineEntryLockReleaseCmd(root, deps),
		newEngineRiskLatchReleaseCmd(root, deps),
		newEngineRiskLatchShowCmd(root, deps),
	}
}

type relaxationCommon struct {
	account, market, approval, operator, reason string
}

func (c *relaxationCommon) bind(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.account, "account", "", "Account the scope belongs to, exactly as the journal records it")
	cmd.Flags().StringVar(&c.market, "market", "", "KR or US")
	cmd.Flags().StringVar(&c.approval, "approval", "", "Human approval reference this relaxation rests on (ticket, runbook step, approver)")
	cmd.Flags().StringVar(&c.operator, "operator", "", "Operator identity recorded in the audit log")
	cmd.Flags().StringVar(&c.reason, "reason", "", "Why the relaxation is justified")
	for _, name := range []string{"account", "market", "approval", "operator", "reason"} {
		_ = cmd.MarkFlagRequired(name)
	}
}

// check 는 원장이나 audit 로그를 열기 **전에** 빈 값을 거절한다 — 승인 없는 해제는 시작조차 하지 않는다.
func (c *relaxationCommon) check(name string) (riskbucket.Market, error) {
	for flag, value := range map[string]string{"--account": c.account, "--approval": c.approval, "--operator": c.operator, "--reason": c.reason} {
		if strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("engine %s: %s is required", name, flag)
		}
	}
	market := riskbucket.Market(strings.ToUpper(strings.TrimSpace(c.market)))
	if market != riskbucket.MarketKR && market != riskbucket.MarketUS {
		return "", fmt.Errorf("engine %s: --market must be KR or US", name)
	}
	return market, nil
}

func (c *relaxationCommon) reasonWithOperator() string {
	return strings.TrimSpace(c.reason) + " (operator " + strings.TrimSpace(c.operator) + ")"
}

func newEngineEntryLockReleaseCmd(root *rootOptions, deps riskRelaxationDeps) *cobra.Command {
	common := &relaxationCommon{}
	var horizon string
	var lockSeq, expectEvent int64
	cmd := &cobra.Command{
		Use:   "entry-lock-release",
		Short: "Release one open entry loss lock with a human approval (audited)",
		Long: strings.TrimSpace(`
Release the open horizon×market entry loss lock that "engine risk-latch-show" printed.

The release is bound to what the approver saw: --lock-seq and --expect-event are the
lock number and its last REAFFIRM event from risk-latch-show. When a new tightening
arrived after that (a REAFFIRM, or another lock), the release is refused as stale and
must be re-approved on the current state.

--approval, --operator and --reason are required. The audit line is written before
the journal commit; when the audit log cannot be opened or written, nothing changes.
Automatic paths never call this. The command places no order and does not need the
engine to be stopped.`),
		Annotations:  map[string]string{"source": "local", "mutating": "true"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			market, err := common.check("entry-lock-release")
			if err != nil {
				return err
			}
			h := riskbucket.Horizon(strings.ToUpper(strings.TrimSpace(horizon)))
			if h != riskbucket.HorizonShort && h != riskbucket.HorizonMedium {
				return errors.New("engine entry-lock-release: --horizon must be SHORT or MEDIUM")
			}
			if lockSeq <= 0 || expectEvent < 0 {
				return errors.New("engine entry-lock-release: --lock-seq and --expect-event must come from engine risk-latch-show")
			}
			return withRelaxationJournal(cmd, root, deps, common.operator, func(ctx context.Context, w riskRelaxationWriter, auditor journal.RiskRelaxationAuditor) error {
				record, err := w.ReleaseEntryLossLock(ctx, journal.EntryLossLockReleaseRequest{
					AccountRef: strings.TrimSpace(common.account), Market: market, Horizon: h, LockSeq: lockSeq, ExpectedLastEvent: expectEvent,
					Actor: journal.RelaxationActorOperator, Approval: common.approval, Reason: common.reasonWithOperator(),
					ReleasedAt: deps.now(), Auditor: auditor,
				})
				if err != nil {
					return fmt.Errorf("engine entry-lock-release: %w", err)
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "released entry loss lock %d (%s/%s/%s) as release %d\n",
					record.LockSeq, strings.TrimSpace(common.account), market, h, record.ReleaseSeq)
				return err
			})
		},
	}
	common.bind(cmd)
	cmd.Flags().StringVar(&horizon, "horizon", "", "SHORT or MEDIUM")
	cmd.Flags().Int64Var(&lockSeq, "lock-seq", 0, "Lock number shown by engine risk-latch-show")
	cmd.Flags().Int64Var(&expectEvent, "expect-event", -1, "Last event number shown by engine risk-latch-show (0 when none)")
	for _, name := range []string{"horizon", "lock-seq", "expect-event"} {
		_ = cmd.MarkFlagRequired(name)
	}
	return cmd
}

func newEngineRiskLatchReleaseCmd(root *rootOptions, deps riskRelaxationDeps) *cobra.Command {
	common := &relaxationCommon{}
	var symbol, generation, expectState string
	cmd := &cobra.Command{
		Use:   "risk-latch-release",
		Short: "Release one owner's RISK_OVERAGE latch with a human approval (audited)",
		Long: strings.TrimSpace(`
Release the RISK_OVERAGE latch of one active risk owner generation.

The release is bound to the owner state digest that "engine risk-latch-show" printed
(--expect-state). Any fill or latch recorded after that reseals the state and makes
the release stale. Only RISK_OVERAGE is released: UNKNOWN_ACTUAL_RISK and the
recorded overage amounts stay, and a later fill that is still over the limit latches
again. Releasing the latch is what lets the owner release (broker zero and the other
checks) proceed afterwards.

--approval, --operator and --reason are required. The audit line is written before
the journal commit; when the audit log cannot be opened or written, nothing changes.
The command places no order and does not need the engine to be stopped.`),
		Annotations:  map[string]string{"source": "local", "mutating": "true"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			market, err := common.check("risk-latch-release")
			if err != nil {
				return err
			}
			if strings.TrimSpace(symbol) == "" || strings.TrimSpace(generation) == "" || strings.TrimSpace(expectState) == "" {
				return errors.New("engine risk-latch-release: --symbol, --generation and --expect-state must come from engine risk-latch-show")
			}
			owner := riskbucket.OwnerKey{AccountID: strings.TrimSpace(common.account), Market: market,
				Symbol: strings.TrimSpace(symbol), ProspectiveGeneration: strings.TrimSpace(generation)}
			return withRelaxationJournal(cmd, root, deps, common.operator, func(ctx context.Context, w riskRelaxationWriter, auditor journal.RiskRelaxationAuditor) error {
				record, err := w.ReleaseRiskOverageLatch(ctx, journal.RiskOverageLatchReleaseRequest{
					Owner: owner, ExpectedStateDigest: strings.TrimSpace(expectState), Actor: journal.RelaxationActorOperator,
					Approval: common.approval, Reason: common.reasonWithOperator(), ReleasedAt: deps.now(), Auditor: auditor,
				})
				if err != nil {
					return fmt.Errorf("engine risk-latch-release: %w", err)
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "released RISK_OVERAGE of %s/%s/%s/%s as release %d\n",
					owner.AccountID, owner.Market, owner.Symbol, owner.ProspectiveGeneration, record.ReleaseSeq)
				return err
			})
		},
	}
	common.bind(cmd)
	cmd.Flags().StringVar(&symbol, "symbol", "", "Owner symbol shown by engine risk-latch-show")
	cmd.Flags().StringVar(&generation, "generation", "", "Owner prospective generation shown by engine risk-latch-show")
	cmd.Flags().StringVar(&expectState, "expect-state", "", "Owner state digest shown by engine risk-latch-show")
	for _, name := range []string{"symbol", "generation", "expect-state"} {
		_ = cmd.MarkFlagRequired(name)
	}
	return cmd
}

// withRelaxationJournal 은 audit 로그를 **먼저** 열고(안 열리면 원장을 열지도 않음) 원장을 연 뒤 해제를 실행한다.
func withRelaxationJournal(cmd *cobra.Command, root *rootOptions, deps riskRelaxationDeps, operator string,
	run func(context.Context, riskRelaxationWriter, journal.RiskRelaxationAuditor) error) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	auditor, err := deps.openAuditor(strings.TrimSpace(operator))
	if err != nil || auditor == nil {
		return fmt.Errorf("engine: the audit log is unavailable, so no relaxation is recorded (nothing was changed): %v", err)
	}
	path, err := deps.journalPath(root)
	if err != nil {
		return err
	}
	w, err := deps.openWriter(ctx, path)
	if err != nil {
		return fmt.Errorf("engine: opening the engine journal: %w", err)
	}
	defer w.Close()
	return run(ctx, w, auditor)
}

func newEngineRiskLatchShowCmd(root *rootOptions, deps riskRelaxationDeps) *cobra.Command {
	var account, market string
	cmd := &cobra.Command{
		Use:   "risk-latch-show",
		Short: "Show open entry loss locks and owner latches with the values a release binds to",
		Long: strings.TrimSpace(`
Print, for one account, the open entry loss locks (lock number, last REAFFIRM event,
cause) and the active risk owners with their RISK_OVERAGE / UNKNOWN_ACTUAL_RISK
latches and state digest. These are the values "engine entry-lock-release" and
"engine risk-latch-release" bind to. Read-only: the journal is opened read-only and
is never migrated or written.`),
		Annotations:  map[string]string{"source": "local", "mutating": "false"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEngineRiskLatchShow(cmd, root, deps, account, market)
		},
	}
	cmd.Flags().StringVar(&account, "account", "", "Account to show")
	cmd.Flags().StringVar(&market, "market", "", "KR or US (both when empty)")
	_ = cmd.MarkFlagRequired("account")
	return cmd
}

type riskLatchShow struct {
	Locks  []journal.EntryLossLockView  `json:"locks"`
	Owners []journal.RiskOwnerLatchView `json:"owners"`
}

func runEngineRiskLatchShow(cmd *cobra.Command, root *rootOptions, deps riskRelaxationDeps, account, market string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := output.ParseFormat(root.outputFormat)
	if err != nil {
		return err
	}
	if strings.TrimSpace(account) == "" {
		return errors.New("engine risk-latch-show: --account is required")
	}
	markets := []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS}
	if m := riskbucket.Market(strings.ToUpper(strings.TrimSpace(market))); m != "" {
		if m != riskbucket.MarketKR && m != riskbucket.MarketUS {
			return errors.New("engine risk-latch-show: --market must be KR or US")
		}
		markets = []riskbucket.Market{m}
	}
	path, err := deps.journalPath(root)
	if err != nil {
		return err
	}
	r, err := deps.openReader(ctx, path)
	if err != nil {
		return fmt.Errorf("engine risk-latch-show: opening the engine journal read-only: %w", err)
	}
	defer r.Close()
	show := riskLatchShow{Locks: []journal.EntryLossLockView{}, Owners: []journal.RiskOwnerLatchView{}}
	locks, err := r.ReadEntryLossLocks(ctx, strings.TrimSpace(account))
	if err != nil {
		return err
	}
	for _, lock := range locks {
		for _, m := range markets {
			if lock.Lock.Market == m {
				show.Locks = append(show.Locks, lock)
			}
		}
	}
	for _, m := range markets {
		owners, err := r.ReadRiskOwnerLatches(ctx, strings.TrimSpace(account), m)
		if err != nil {
			return err
		}
		show.Owners = append(show.Owners, owners...)
	}
	return writeRiskLatchShow(cmd.OutOrStdout(), format, show)
}

func writeRiskLatchShow(w io.Writer, format output.Format, show riskLatchShow) error {
	if format == output.FormatJSON {
		return output.WriteJSON(w, show)
	}
	table := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "ENTRY LOSS LOCKS")
	fmt.Fprintln(table, "MARKET\tHORIZON\tLOCK-SEQ\tLAST-EVENT\tACTIVATED\tCAUSE")
	for _, lock := range show.Locks {
		fmt.Fprintf(table, "%s\t%s\t%d\t%d\t%s\t%s\n", lock.Lock.Market, lock.Lock.Horizon, lock.Lock.Seq, lock.LastEvent,
			lock.Lock.ActivatedAt.Format(time.RFC3339), lock.Lock.Cause)
	}
	fmt.Fprintln(table, "")
	fmt.Fprintln(table, "RISK OWNERS")
	fmt.Fprintln(table, "MARKET\tSYMBOL\tGENERATION\tRISK_OVERAGE\tUNKNOWN_ACTUAL_RISK\tSTATE-DIGEST")
	for _, owner := range show.Owners {
		fmt.Fprintf(table, "%s\t%s\t%s\t%t\t%t\t%s\n", owner.Owner.Market, owner.Owner.Symbol, owner.Owner.ProspectiveGeneration,
			owner.OverageLatched, owner.UnknownLatched, owner.StateDigest)
	}
	return table.Flush()
}

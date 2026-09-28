package main

// engine_risk_relaxation.go 는 a066 5.5 완화(해제) 메커니즘의 운영자 표면이다(사용자 결정 2026-09-28, design D8).
//
// 원칙: 자동 경로는 조이기만 한다. 완화·해제는 OPERATOR 가 사람 승인 참조와 함께 요청하고, audit 줄이 원장 commit 앞에
// 기록되며(audit 가 없거나 실패하면 아무것도 안 바뀜), 승인자가 본 상태에 결속된다 — 그 뒤 상태가 바뀌었으면
// stale 로 거절된다(동시 조이기는 보수 쪽 승리). 콘솔 버튼은 없다.
//
// 두 해제 명령은 `mutating: true` 다 — 대화형 에이전트는 자동 실행하지 않는다(AGENTS.md). 읽기 전용 `risk-latch-show`
// 가 해제가 결속할 값(잠금 번호·마지막 사건 번호·owner 상태 digest)을 보여 준다.
//
// # 경로: 엔진 제어 endpoint (Manager 판정 2026-09-29, a092 완화 명령 가족 계약)
//
// 해제 명령은 원장을 직접 열지 않는다. 원장은 단일 writer(엔진)이고 journal.Open 은 마이그레이션을 하므로, CLI 가 직접
// 쓰면 새 바이너리가 도는 엔진 밑에서 스키마를 올릴 수 있다. 그래서 엔진이 발행한 position-policy 제어 endpoint 에
// 요청하고, 엔진이 자기 journal 핸들과 자기 audit 로그로 해제한다(internal/app/engine/risk_relaxation_command.go).
// 엔진이 없으면 거절한다 — 엔진이 없으면 진입도 없으므로 잠금·latch 는 그동안 아무것도 막지 않는다. 엔진은 latch 가
// 걸린 채로 안전하게 기동하므로 "엔진 기동 → 해제" 순서가 성립한다.
//
// engine lock 을 잡지 않는다: 엔진을 멈춰야 하는데, 엔진이 멈추면 (보호가 UNWIRED 인 동안) 손절이 없다.
//
// 커밋 뒤 엔진이 원장 alert 를 enqueue 한다(통지). 그것이 실패하면 해제는 유효하고 명령은 「완화됨·통지 실패」로
// 0 이 아닌 코드로 끝난다. `risk-latch-show` 는 읽기 전용 연결(mode=ro)로 직접 읽는다 — writer 가 아니다.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/output"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskrelaxation"
	"github.com/spf13/cobra"
)

// riskRelaxationClient 는 엔진 제어 endpoint 의 해제 면이다(positionpolicyrpc.Client 가 구현).
type riskRelaxationClient interface {
	ReleaseEntryLossLock(context.Context, riskrelaxation.EntryLockReleaseRequest) (riskrelaxation.Result, error)
	ReleaseRiskOverageLatch(context.Context, riskrelaxation.LatchReleaseRequest) (riskrelaxation.Result, error)
}

// riskRelaxationReader 는 읽기 전용 표면이다.
type riskRelaxationReader interface {
	ReadEntryLossLocks(context.Context, string) ([]journal.EntryLossLockView, error)
	ReadRiskOwnerLatches(context.Context, string, riskbucket.Market) ([]journal.RiskOwnerLatchView, error)
	Close() error
}

type riskRelaxationDeps struct {
	engineDir  func(*rootOptions) (string, error)
	dial       func(ctx context.Context, engineDir string) (riskRelaxationClient, error)
	openReader func(context.Context, string) (riskRelaxationReader, error)
}

func productionRiskRelaxationDeps() riskRelaxationDeps {
	return riskRelaxationDeps{
		engineDir: engineJournalDir,
		dial: func(ctx context.Context, engineDir string) (riskRelaxationClient, error) {
			// 부재를 먼저 봄 — descriptor 가 없다는 것은 "엔진이 돌지 않는다"(또는 endpoint 가 강등됨)이다.
			descriptor := positionpolicyrpc.DescriptorPath(engineDir)
			if _, err := os.Stat(descriptor); err != nil {
				return nil, err
			}
			client, err := positionpolicyrpc.Dial(ctx, descriptor)
			if err != nil {
				// 타입 있는 nil 포인터를 인터페이스에 담아 돌려주지 않는다.
				return nil, err
			}
			return client, nil
		},
		openReader: func(ctx context.Context, path string) (riskRelaxationReader, error) {
			return journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: path})
		},
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
	cmd.Flags().StringVar(&c.operator, "operator", "", "Operator identity recorded with the release")
	cmd.Flags().StringVar(&c.reason, "reason", "", "Why the relaxation is justified")
	for _, name := range []string{"account", "market", "approval", "operator", "reason"} {
		_ = cmd.MarkFlagRequired(name)
	}
}

// check 는 엔진에 닿기 **전에** 빈 값을 거절한다 — 승인 없는 해제는 시작조차 하지 않는다.
func (c *relaxationCommon) check(name string) (riskbucket.Market, error) {
	for _, field := range []struct{ flag, value string }{
		{"--account", c.account}, {"--approval", c.approval}, {"--operator", c.operator}, {"--reason", c.reason},
	} {
		if strings.TrimSpace(field.value) == "" {
			return "", fmt.Errorf("engine %s: %s is required", name, field.flag)
		}
	}
	market := riskbucket.Market(strings.ToUpper(strings.TrimSpace(c.market)))
	if market != riskbucket.MarketKR && market != riskbucket.MarketUS {
		return "", fmt.Errorf("engine %s: --market must be KR or US", name)
	}
	return market, nil
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

The running engine performs the release through its control endpoint, with its own
journal and audit log; the audit line is written before the journal commit. When the
engine is not running the command refuses — nothing is released, and entries cannot
happen without the engine either. Start the engine and retry. --approval, --operator
and --reason are required. The command places no order and does not stop the engine.`),
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
			return runRiskRelaxation(cmd, root, deps, "entry-lock-release", func(ctx context.Context, c riskRelaxationClient) (riskrelaxation.Result, error) {
				return c.ReleaseEntryLossLock(ctx, riskrelaxation.EntryLockReleaseRequest{
					AccountRef: strings.TrimSpace(common.account), Market: string(market), Horizon: string(h),
					LockSeq: lockSeq, ExpectedLastEvent: expectEvent,
					Operator: strings.TrimSpace(common.operator), Approval: strings.TrimSpace(common.approval),
					Reason: strings.TrimSpace(common.reason),
				})
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

The running engine performs the release through its control endpoint, with its own
journal and audit log; the audit line is written before the journal commit. When the
engine is not running the command refuses; start the engine and retry. --approval,
--operator and --reason are required. The command places no order.`),
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
			return runRiskRelaxation(cmd, root, deps, "risk-latch-release", func(ctx context.Context, c riskRelaxationClient) (riskrelaxation.Result, error) {
				return c.ReleaseRiskOverageLatch(ctx, riskrelaxation.LatchReleaseRequest{
					AccountRef: strings.TrimSpace(common.account), Market: string(market),
					Symbol: strings.TrimSpace(symbol), Generation: strings.TrimSpace(generation),
					ExpectedState: strings.TrimSpace(expectState), Operator: strings.TrimSpace(common.operator),
					Approval: strings.TrimSpace(common.approval), Reason: strings.TrimSpace(common.reason),
				})
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

// runRiskRelaxation 은 엔진 제어 endpoint 에 붙어 해제를 요청하고 결과를 말한다.
//
// 세 결과를 가른다: (1) 엔진에 닿지 못함 → 아무것도 안 바뀜, 엔진 기동 후 재시도. (2) 엔진이 거절(stale·invalid·audit
// 없음) → 아무것도 안 바뀜. (3) 커밋됨 — 통지까지 됐으면 0, 통지가 실패했으면 「완화됨·통지 실패」로 0 이 아닌 종료
// (해제는 유효). 요청은 보냈는데 응답을 못 읽은 경우는 결과 불명이라고 말한다 — 재시도는 안전하다(이미 해제됐으면
// 결속이 stale 로 거절한다).
func runRiskRelaxation(cmd *cobra.Command, root *rootOptions, deps riskRelaxationDeps, name string,
	call func(context.Context, riskRelaxationClient) (riskrelaxation.Result, error)) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	dir, err := deps.engineDir(root)
	if err != nil {
		return err
	}
	client, err := deps.dial(ctx, dir)
	if err != nil || client == nil {
		return fmt.Errorf("engine %s: the engine is not running or its control endpoint is unavailable (%v); "+
			"nothing was released — a stopped engine takes no entries, so start it (tossctl engine run) and retry", name, err)
	}
	result, err := call(ctx, client)
	if err != nil {
		switch {
		case errors.Is(err, riskrelaxation.ErrStale), errors.Is(err, riskrelaxation.ErrInvalidRequest),
			errors.Is(err, riskrelaxation.ErrAuditUnavailable), errors.Is(err, riskrelaxation.ErrUnwired):
			return fmt.Errorf("engine %s: refused, nothing was released: %w", name, err)
		}
		return fmt.Errorf("engine %s: the outcome is unknown (%w); run engine risk-latch-show before retrying — "+
			"a repeated release of an already released target is refused as stale", name, err)
	}
	out := cmd.OutOrStdout()
	if _, err := fmt.Fprintf(out, "released %s as release %d at %s\n", result.Target, result.ReleaseSeq, result.ReleasedAt); err != nil {
		return err
	}
	if !result.Notified {
		return fmt.Errorf("engine %s: 완화됨·통지 실패 — release %d of %s is in effect, but the operator notice could not be recorded: %s",
			name, result.ReleaseSeq, result.Target, result.NotifyError)
	}
	return nil
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
	dir, err := deps.engineDir(root)
	if err != nil {
		return err
	}
	r, err := deps.openReader(ctx, filepath.Join(dir, journal.DBFileName))
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

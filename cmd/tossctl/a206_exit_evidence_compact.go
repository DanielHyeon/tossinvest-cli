package main

// a206_exit_evidence_compact.go 는 닫힌 포지션의 exit_events 판정 증거 정비 명령의 운영자 표면임
// (`tossctl engine exit-evidence-compact`, a206 design D3 · D4).
//
// 완화 명령 가족 계약(a066 · a092)을 따름: `mutating: true`(대화형 에이전트는 자동 실행하지 않음) · 운영자 · 승인 참조 · note 필수
// (--apply 일 때 — dry-run 은 아무것도 기록하지 않으므로 받지 않음) · 콘솔 버튼 없음. 브로커를 호출하지 않음.
// 다른 가족 명령과 달리 엔진 endpoint 가 아니라 **엔진 잠금**을 잡음: VACUUM 은 저널 파일 전체를 다시 쓰므로 엔진이 저널을 연 채로는
// 할 수 없음. 엔진 정지 = 손절 없음이라 실행은 두 시장이 닫힌 창에서 사람이 함(사람 결정 3).

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/spf13/cobra"
)

// auditActionExitEvidenceCompact 는 이 정비의 audit action 임 — 운영자가 grep 하는 안정 문자열.
const auditActionExitEvidenceCompact = "journal.exit_evidence_compact"

type exitEvidenceCompactDeps struct {
	journalDir func(*rootOptions) (string, error)
	acquire    func(string) (reconcileResolveLock, error)
	clock      clock.Clock
}

func productionExitEvidenceCompactDeps() exitEvidenceCompactDeps {
	return exitEvidenceCompactDeps{
		journalDir: engineJournalDir,
		acquire: func(dir string) (reconcileResolveLock, error) {
			return enginelock.Acquire(dir)
		},
		clock: clock.System(),
	}
}

type exitEvidenceCompactOptions struct {
	apply                    bool
	operator, approval, note string
}

func newEngineExitEvidenceCompactCmd(root *rootOptions) *cobra.Command {
	return newEngineExitEvidenceCompactCmdWithDeps(root, productionExitEvidenceCompactDeps())
}

func newEngineExitEvidenceCompactCmdWithDeps(root *rootOptions, deps exitEvidenceCompactDeps) *cobra.Command {
	opts := &exitEvidenceCompactOptions{}
	cmd := &cobra.Command{
		Use:   "exit-evidence-compact",
		Short: "Empty the exit judgement snapshots of closed positions and shrink the journal (audited)",
		Long: strings.TrimSpace(`
Empty the twenty judgement-evidence columns of exit_events rows that belong to closed
positions (CLOSED, exit state completed) and are not linked to a broker order, then
VACUUM the journal. The rows themselves, their action, intent, prices and times stay;
the emptied rows read as legacy events. Rows of open positions, rows whose intent has an
attempt with a broker order number, and repeat-refusal compact rows with their anchors
are left untouched. A second run changes nothing.

Without --apply the command only prints what it would change and writes nothing.
With --apply it backs the journal up next to itself (<journal>.pre-a206-<UTC>.bak),
empties the evidence in one transaction checked by PRAGMA integrity_check, runs VACUUM,
and records the counts and sizes in the audit log. --operator, --approval and --note
are required with --apply.

The engine must be stopped: the command takes the engine lock before it opens the
journal. Stopping the engine stops the stop-loss, so run it only while both markets
are closed. It places no order and calls no broker.`),
		Annotations:  map[string]string{"source": "local", "mutating": "true"},
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEngineExitEvidenceCompact(cmd, root, deps, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.apply, "apply", false, "Write the change (default: print the counts only)")
	cmd.Flags().StringVar(&opts.operator, "operator", "", "Operator identity recorded with the compaction (required with --apply)")
	cmd.Flags().StringVar(&opts.approval, "approval", "", "Human approval reference (ticket, runbook step, approver; required with --apply)")
	cmd.Flags().StringVar(&opts.note, "note", "", "What was checked before the compaction (required with --apply)")
	return cmd
}

func runEngineExitEvidenceCompact(cmd *cobra.Command, root *rootOptions, deps exitEvidenceCompactDeps,
	opts *exitEvidenceCompactOptions,
) error {
	// 승인 없는 적용은 잠금 · 저널에 닿기 전에 거절함.
	if opts.apply {
		for _, f := range []struct{ flag, value string }{
			{"--operator", opts.operator}, {"--approval", opts.approval}, {"--note", opts.note},
		} {
			if strings.TrimSpace(f.value) == "" {
				return fmt.Errorf("engine exit-evidence-compact: %s is required with --apply", f.flag)
			}
		}
	}
	if deps.journalDir == nil || deps.acquire == nil || deps.clock == nil {
		return errors.New("engine exit-evidence-compact: wiring is incomplete")
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	dir, err := deps.journalDir(root)
	if err != nil {
		return err
	}
	// dry-run 도 잠금을 잡음 — 집계가 적용과 같은 정지 상태의 값이어야 하고, 엔진이 돌 때는 어느 쪽도 시작하지 않음.
	lock, err := deps.acquire(dir)
	if err != nil {
		return fmt.Errorf("engine exit-evidence-compact: refused, nothing was changed — stop the engine first: %w", err)
	}
	defer lock.Release()

	path := filepath.Join(dir, journal.DBFileName)
	out := cmd.OutOrStdout()
	if !opts.apply {
		// 읽기 전용 연결이라 마이그레이션 · 파일 생성 · 쓰기가 없음.
		ro, err := journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: path})
		if err != nil {
			return fmt.Errorf("engine exit-evidence-compact: opening the journal read-only: %w", err)
		}
		plan, err := ro.ExitEvidenceCompactionPlan(ctx)
		if closeErr := ro.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "dry-run: rows=%d targets=%d kept_repeat_rows=%d snapshot_bytes=%d file_bytes=%d — nothing was written; add --apply to compact\n",
			plan.Rows, plan.Targets, plan.Kept, plan.SnapshotBytes, plan.FileBytes)
		return err
	}

	// journal.Open 은 없는 파일을 새로 만들므로, 잘못 지정한 디렉터리에서 빈 원장을 「정비 성공」으로 만들지 않게 먼저 확인함.
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("engine exit-evidence-compact: refused, nothing was changed — no journal at %s: %w", path, err)
	}
	j, err := journal.Open(ctx, journal.Options{Path: path, Clock: deps.clock})
	if err != nil {
		return fmt.Errorf("engine exit-evidence-compact: opening the journal: %w", err)
	}
	report, compactErr := j.CompactClosedExitEvidence(ctx)
	if closeErr := j.Close(); compactErr == nil && closeErr != nil {
		compactErr = fmt.Errorf("engine exit-evidence-compact: closing the journal: %w", closeErr)
	}

	// 백업이 생긴 뒤의 결과는 성공이든 실패든 audit 에 남김 — 백업 파일이 왜 있는지 audit 로 설명할 수 있어야 함.
	summary := fmt.Sprintf("rows=%d rows_after=%d targets=%d emptied=%d remaining=%d kept_repeat_rows=%d snapshot_bytes=%d "+
		"file_bytes=%d file_bytes_after=%d",
		report.Rows, report.RowsAfter, report.Targets, report.Emptied, report.Remaining, report.Kept, report.SnapshotBytes,
		report.FileBytes, report.FileBytesAfter)
	if report.Backup != "" {
		outcome := "ok"
		if compactErr != nil {
			outcome = "failed: " + compactErr.Error()
		}
		log, err := audit.Open(audit.Options{Path: filepath.Join(dir, audit.FileName), Clock: deps.clock,
			Subject: strings.TrimSpace(opts.operator)})
		if err == nil {
			err = log.RecordAction(auditActionExitEvidenceCompact, "journal.exit_events.evidence", summary,
				fmt.Sprintf("approval=%s; note=%s; backup=%s; outcome=%s",
					strings.TrimSpace(opts.approval), strings.TrimSpace(opts.note), report.Backup, outcome))
		}
		if err != nil {
			return errors.Join(compactErr, fmt.Errorf("engine exit-evidence-compact: recording the audit line: %w "+
				"(the journal result above stands; backup %s)", err, report.Backup))
		}
	}
	if compactErr != nil {
		return compactErr
	}
	_, err = fmt.Fprintf(out, "compacted: %s — backup %s (delete it after you have checked the journal)\n", summary, report.Backup)
	return err
}

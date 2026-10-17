package main

// scrub_account_records.go — a131 일회성 정리 명령(design D5).
//
// a131 이전 판본이 쓴 soak·검증 기록과 감사 기록에는 계좌가 평문으로 있음. 이 명령은 그 파일들을 a131 의 **같은 함수**로
// 다시 씀 — 기록 줄은 redact.SealRecordLine(쓰기 경로와 같은 함수), 감사 줄은 audit.RedactEntry(audit.Log.Record 가 부르는 함수).
// 규칙 사본 없음.
//
// 두 단계: 기본은 미리보기(파일별 줄 수·바뀔 줄 수만 출력, 쓰기 0 — 봉인 키도 만들지 않음), `--apply` 는 사람이 실행.
// 적용은 같은 디렉터리 임시 파일 → fsync → 원본이 읽은 때와 같은지 다시 봄 → rename(원자 교체) → 디렉터리 fsync.
// 백업은 만들지 않음(사용자 결정 2026-10-12: 계좌 정보는 남기지 않음).

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// scrubDrainInterval 은 교체 뒤 옛 inode 꼬리를 다시 보는 간격, scrubDrainRounds 는 그 상한(약 10초)임.
var (
	scrubDrainInterval = 200 * time.Millisecond
	scrubDrainRounds   = 50
)

// scrubFileOwner 는 디렉터리 소유자 uid 를 읽는 자리임(시험이 다른 소유자를 흉내 냄).
var scrubFileOwner = fileOwner

// scrubBeforeReplace · scrubAfterReplace 는 시험이 「읽은 뒤 교체 전」·「교체 직후」 순간에 파일에 써 보는 자리임. 생산에서는 아무것도 안 함.
var (
	scrubBeforeReplace = func(string) {}
	scrubAfterReplace  = func(string) {}
)

type scrubTarget struct {
	path  string
	audit bool // 감사 기록(줄 = audit.Entry)인지, soak·검증 기록(줄 = JSON 객체, account_ref)인지
}

type scrubPlan struct {
	target  scrubTarget
	file    *os.File // 읽은 원본 fd — rename 뒤에도 쥐고 교체 창에 들어온 꼬리를 옮김
	info    os.FileInfo
	key     *redact.SealKey
	lines   int
	rewrite int
	broken  int // 해독할 수 없어 그대로 두는 줄(찢긴 마지막 줄 등)
	output  []byte
}

func newScrubAccountRecordsCmd(root *rootOptions) *cobra.Command {
	var (
		apply     bool
		files     []string
		auditFile string
	)
	cmd := &cobra.Command{
		Use:   "scrub-account-records",
		Short: "Remove account information from existing soak, verification and audit records (one-time cleanup)",
		Long: strings.TrimSpace(`
Rewrite the records a131's predecessors wrote with the account in the clear:
the capability soak record, the supervised verification records (KR, US) and
the audit log. Record lines get their account sealed with the local key the
record writers now use; audit lines get the account text replaced. Everything
else on each line is kept, so the attestation's account binding and the audit
trail's meaning (action, setting, old/new value, time, subject) are unchanged.

Without --apply nothing is written: each file's line count and the number of
lines that would change are printed. With --apply each file is replaced
atomically, without a backup. Stop the soak survey first — it holds the record
open — and do not run this while a verification is running.`),
		// mutating 은 브로커 변경(거래 동작)에만 씀(TestMutatingAnnotationOnTradeCommands). 이 명령의 쓰기는 --apply 일 때뿐이고
		// 적용은 사람이 함(design D5).
		Annotations:  map[string]string{"source": "local"},
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runScrubAccountRecords(cmd.OutOrStdout(), root, files, auditFile, apply)
		},
	}
	cmd.Flags().BoolVar(&apply, "apply", false, "Rewrite the files (default: preview only, nothing is written)")
	cmd.Flags().StringSliceVar(&files, "file", nil,
		"Record file to scrub instead of the defaults; repeatable. A file named audit.log is treated as an audit log")
	cmd.Flags().StringVar(&auditFile, "audit-file", "", "Override the audit log path")
	return cmd
}

// scrubTargets 는 기본 대상을 엔진·soak·verify 와 같은 해석으로 정함(규칙을 복사하지 않고 같은 해석 함수를 부름).
func scrubTargets(root *rootOptions, files []string, auditFile string) ([]scrubTarget, error) {
	var out []scrubTarget
	if len(files) > 0 {
		for _, f := range files {
			if f = strings.TrimSpace(f); f != "" {
				out = append(out, scrubTarget{path: f, audit: filepath.Base(f) == audit.FileName})
			}
		}
	} else {
		soakPath, err := resolveSoakRecord(root, "")
		if err != nil {
			return nil, err
		}
		out = append(out, scrubTarget{path: soakPath})
		for _, market := range []string{verifylive.MarketKR, verifylive.MarketUS} {
			p, err := resolveVerifyRecordFor(root, "", market)
			if err != nil {
				return nil, err
			}
			out = append(out, scrubTarget{path: p})
		}
	}
	if auditFile = strings.TrimSpace(auditFile); auditFile != "" {
		out = append(out, scrubTarget{path: auditFile, audit: true})
	} else if len(files) == 0 {
		dir, err := journal.DataDir()
		if err != nil {
			return nil, fmt.Errorf("scrub: resolving the audit log location: %w", err)
		}
		out = append(out, scrubTarget{path: filepath.Join(dir, audit.FileName), audit: true})
	}
	return out, nil
}

func runScrubAccountRecords(out io.Writer, root *rootOptions, files []string, auditFile string, apply bool) error {
	targets, err := scrubTargets(root, files, auditFile)
	if err != nil {
		return err
	}
	// 감사 줄의 원문 치환에는 등록 값이 필요함 — 기록의 계좌(평문·봉인)와 attestation 의 계좌를 읽어 등록(읽기만).
	if registerKnownAccounts(root, targets) == 0 {
		fmt.Fprintln(out, "note: no full account reference was found in the records or the attestation — audit lines are scrubbed "+
			"by the mask and derived-id rules only; raw account numbers in them would be left")
	}

	if apply {
		if err := refuseForeignOwner(targets); err != nil {
			return err
		}
		if err := refuseWhileRecordsAreHeld(targets); err != nil {
			return err
		}
	}
	mode := "preview — nothing written; re-run with --apply to rewrite"
	if apply {
		mode = "applied"
	}
	for _, t := range targets {
		plan, err := planScrub(t, apply)
		if err != nil {
			return err
		}
		if plan == nil {
			fmt.Fprintf(out, "%s: absent, skipped\n", t.path)
			continue
		}
		if apply && plan.rewrite > 0 {
			err = replaceScrubbed(plan)
		}
		_ = plan.file.Close()
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s: %d line(s), %d to rewrite", t.path, plan.lines, plan.rewrite)
		if plan.broken > 0 {
			fmt.Fprintf(out, ", %d undecodable left as is", plan.broken)
		}
		fmt.Fprintln(out)
	}
	fmt.Fprintf(out, "%s\n", mode)
	return nil
}

// registerKnownAccounts 는 평문 계좌가 남아 있는 곳(기록의 account_ref · attestation)에서 계좌를 읽어 치환 규칙에 등록함.
func registerKnownAccounts(root *rootOptions, targets []scrubTarget) int {
	known := map[string]bool{}
	note := func(ref string) {
		if ref = strings.TrimSpace(ref); ref != "" && !strings.Contains(ref, "*") {
			known[ref] = true
			redact.Register(ref)
		}
	}
	for _, t := range targets {
		if t.audit {
			continue
		}
		data, err := os.ReadFile(t.path)
		if err != nil {
			continue
		}
		// 이미 봉인된 줄도 엶 — 기록이 먼저 정리됐어도 감사 줄의 원문을 치환할 수 있게(구현 리뷰 I-12).
		opener := redact.NewOpener(filepath.Dir(t.path))
		for _, line := range bytes.Split(data, []byte{'\n'}) {
			var probe struct {
				AccountRef       string `json:"account_ref"`
				AccountRefSealed string `json:"account_ref_sealed"`
			}
			if json.Unmarshal(line, &probe) != nil {
				continue
			}
			note(probe.AccountRef)
			if probe.AccountRefSealed != "" {
				if ref, ok := opener.Open(probe.AccountRefSealed); ok {
					note(ref)
				}
			}
		}
	}
	if path, err := resolveAttestationPath(root, ""); err == nil {
		if data, err := os.ReadFile(path); err == nil {
			var a struct {
				AccountRef string `json:"account_ref"`
			}
			if json.Unmarshal(data, &a) == nil {
				note(a.AccountRef)
			}
		}
	}
	return len(known)
}

// refuseForeignOwner 는 실행 uid 가 대상 디렉터리 소유자와 다르면 거절함 — root 로 돌려 root 소유 봉인 키를 만들면 같은 디렉터리의
// 컨테이너 soak(소유자 uid)이 키를 못 읽어 기록을 못 열고, attestation 갱신이 멈춤.
func refuseForeignOwner(targets []scrubTarget) error {
	uid := os.Geteuid()
	for _, t := range targets {
		owner, ok := scrubFileOwner(filepath.Dir(t.path))
		if ok && owner != uid {
			return fmt.Errorf("scrub: %s belongs to uid %d but this runs as uid %d — run it as the directory's owner",
				filepath.Dir(t.path), owner, uid)
		}
	}
	return nil
}

// refuseWhileRecordsAreHeld 는 기록을 쥔 프로세스가 있으면 적용을 거절함 — soak run 은 기록 fd 를 쥐고 있어 교체 뒤의
// 쓰기가 사라진 inode 로 가고, 검증 실행은 같은 기록에 줄을 더함.
func refuseWhileRecordsAreHeld(targets []scrubTarget) error {
	now := time.Now()
	for _, t := range targets {
		if t.audit {
			continue
		}
		lock := filepath.Join(filepath.Dir(t.path), runlock.FileName)
		if fresh, _ := runlock.Fresh(lock, now, runlock.StaleAfter); fresh {
			return fmt.Errorf("scrub: a verification is running (%s) — run the cleanup after it finishes", lock)
		}
		if pids, err := soakFindProcesses(t.path); err == nil && len(pids) > 0 {
			return fmt.Errorf("scrub: a soak survey is appending to %s (pid %s) — stop it first, then restart it after the cleanup",
				t.path, joinPIDs(pids))
		}
	}
	return nil
}

// planScrub 는 파일을 한 번 읽어 줄마다 판정함. apply 면 다시 쓸 바이트까지 만듦(봉인 — 키가 없으면 만듦),
// 아니면 판정만 함(키를 읽지도 만들지도 않음). 없는 파일은 nil.
func planScrub(t scrubTarget, apply bool) (*scrubPlan, error) {
	f, err := os.Open(t.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scrub: %s: %w", t.path, err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("scrub: %s: %w", t.path, err)
	}
	data, err := io.ReadAll(io.NewSectionReader(f, 0, info.Size()))
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("scrub: reading %s: %w", t.path, err)
	}
	plan := &scrubPlan{target: t, file: f, info: info}
	if apply && !t.audit {
		// 쓰는 쪽과 같은 키 준비 — 봉인 줄이 있는데 키가 없으면 새 키를 만들지 않고 거절.
		if plan.key, err = redact.PrepareSealKey(t.path); err != nil {
			_ = f.Close()
			return nil, err
		}
	}
	plan.output, plan.lines, plan.rewrite, plan.broken = scrubLines(t, data, plan.key, apply)
	return plan, nil
}

// scrubLines 는 바이트 덩어리를 줄마다 같은 규칙에 통과시킴(원본 읽기와 교체 창 꼬리 옮기기가 같이 씀).
func scrubLines(t scrubTarget, data []byte, key *redact.SealKey, apply bool) (out []byte, lines, rewrite, broken int) {
	var buf bytes.Buffer
	for len(data) > 0 {
		i := bytes.IndexByte(data, '\n')
		var line []byte
		if i < 0 {
			line, data = data, nil
		} else {
			line, data = data[:i+1], data[i+1:]
		}
		if len(bytes.TrimSpace(line)) == 0 {
			buf.Write(line)
			continue
		}
		lines++
		next, changed, err := scrubLine(t, line, key, apply)
		if err != nil {
			broken++
			buf.Write(line)
			continue
		}
		if changed {
			rewrite++
		}
		buf.Write(next)
	}
	return buf.Bytes(), lines, rewrite, broken
}

func scrubLine(t scrubTarget, line []byte, key *redact.SealKey, apply bool) ([]byte, bool, error) {
	if t.audit {
		body := bytes.TrimRight(line, "\r\n")
		var e audit.Entry
		if err := json.Unmarshal(body, &e); err != nil {
			return line, false, err
		}
		cleaned := audit.RedactEntry(e)
		if cleaned == e {
			return line, false, nil
		}
		encoded, err := json.Marshal(cleaned)
		if err != nil {
			return line, false, err
		}
		return append(encoded, line[len(body):]...), true, nil
	}
	if !apply {
		changed, err := redact.RecordLineNeedsRewrite(line)
		return line, changed, err
	}
	next, err := redact.SealRecordLine(line, key)
	if err != nil {
		return line, false, err
	}
	return next, !bytes.Equal(next, line), nil
}

// replaceScrubbed 는 원자 교체임. 원본의 크기·수정 시각이 읽은 때와 다르면(그 사이 누가 줄을 더함) 원본을 두고 거절함.
func replaceScrubbed(plan *scrubPlan) error {
	path := plan.target.path
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".scrub-a131-*.tmp")
	if err != nil {
		return fmt.Errorf("scrub: creating a temporary file beside %s: %w", path, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("scrub: %w", err)
	}
	if _, err := tmp.Write(plan.output); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("scrub: writing %s: %w", tmpName, err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("scrub: syncing %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("scrub: closing %s: %w", tmpName, err)
	}
	scrubBeforeReplace(path)
	now, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("scrub: %s: %w", path, err)
	}
	if now.Size() != plan.info.Size() || !now.ModTime().Equal(plan.info.ModTime()) {
		return fmt.Errorf("scrub: %s changed while it was being scrubbed — nothing was replaced; run again", path)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("scrub: replacing %s: %w", path, err)
	}
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	scrubAfterReplace(path)
	return drainReplacedTail(plan)
}

// drainReplacedTail 는 교체 창에 옛 inode 로 들어온 줄을 같은 규칙으로 새 파일에 옮김 — 교체 전에 경로를 연 쓰기(감사 줄 · 재열기
// 직전의 기록 줄)가 사라진 inode 에 남지 않게. 옛 inode 크기가 두 번 연속 같을 때까지(200ms 간격).
func drainReplacedTail(plan *scrubPlan) error {
	offset := plan.info.Size()
	stable := 0
	// 상한: 옛 inode 가 계속 자라면(재열기 없는 옛 바이너리가 fd 를 쥠) 끝없이 돌지 않고 멈춰 알림(구현 리뷰 I-11).
	for rounds := 0; stable < 2; rounds++ {
		if rounds >= scrubDrainRounds {
			return fmt.Errorf("scrub: %s kept growing on the replaced file — a writer without the a131 reopen is holding it; "+
				"stop it, then run the cleanup again", plan.target.path)
		}
		time.Sleep(scrubDrainInterval)
		st, err := plan.file.Stat()
		if err != nil {
			return fmt.Errorf("scrub: watching the replaced %s: %w", plan.target.path, err)
		}
		if st.Size() == offset {
			stable++
			continue
		}
		stable = 0
		tail, err := io.ReadAll(io.NewSectionReader(plan.file, offset, st.Size()-offset))
		if err != nil {
			return fmt.Errorf("scrub: reading the replaced tail of %s: %w", plan.target.path, err)
		}
		// 완성된 줄만 옮김 — 쓰는 중인 줄의 앞부분을 옮기면 나머지가 다음 판에 따로 붙어 치환되지 않은 조각이 됨.
		end := bytes.LastIndexByte(tail, '\n')
		if end < 0 {
			stable = 0
			continue
		}
		tail = tail[:end+1]
		offset += int64(len(tail))
		moved, _, _, _ := scrubLines(plan.target, tail, plan.key, true)
		f, err := os.OpenFile(plan.target.path, os.O_WRONLY|os.O_APPEND, 0)
		if err != nil {
			return fmt.Errorf("scrub: appending the replaced tail to %s: %w", plan.target.path, err)
		}
		_, werr := f.Write(moved)
		serr := f.Sync()
		_ = f.Close()
		if werr != nil || serr != nil {
			return fmt.Errorf("scrub: appending the replaced tail to %s: %v %v", plan.target.path, werr, serr)
		}
	}
	return nil
}

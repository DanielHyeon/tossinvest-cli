package main

// a066 5.5(design D8) 운영자 해제 명령의 CLI 절반.
//
// 원장 쪽 계약(OPERATOR·승인·결속·audit-before-commit)은 internal/journal/a066_relaxation_test.go 가 잰다.
// 여기서 재는 것은 **인자를 만드는 쪽**임: 명령이 mutating 으로 선언되는가, 승인·운영자 없이 시작조차 안 하는가,
// audit 로그가 안 열리면 원장을 열지도 않는가, 사람이 본 값(show)이 그대로 해제의 결속으로 건너가는가.
// 모든 시험은 임시 저널만 씀 — 운영 원장에는 닿지 않음(사용자 결정 2026-09-28).

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/spf13/cobra"
)

var relaxationCLINow = time.Date(2026, 3, 30, 3, 0, 0, 0, time.UTC)

// cliRelaxationAuditor 는 audit 줄을 모으고 fail 이면 쓰기를 실패시킴.
type cliRelaxationAuditor struct {
	lines []string
	fail  bool
}

func (a *cliRelaxationAuditor) RecordAction(action, setting, value, detail string) error {
	if a.fail {
		return errors.New("audit disk full")
	}
	a.lines = append(a.lines, strings.Join([]string{action, setting, value, detail}, " | "))
	return nil
}

// relaxationTestJournal 은 KR/SHORT 잠금 하나가 열린 임시 저널 경로를 돌려줌.
func relaxationTestJournal(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), journal.DBFileName)
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     path,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	defer j.Close()
	if _, _, err := j.ActivateEntryLossLock(context.Background(), journal.EntryLossLock{AccountRef: "acct-7",
		Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonShort, Cause: "SHORT_LOSS_LIMIT", ActivatedAt: relaxationCLINow}); err != nil {
		t.Fatalf("ActivateEntryLossLock: %v", err)
	}
	return path
}

func openLocks(t *testing.T, path string) []journal.EntryLossLockView {
	t.Helper()
	r, err := journal.OpenReadOnly(context.Background(), journal.ReadOnlyOptions{Path: path})
	if err != nil {
		t.Fatalf("OpenReadOnly: %v", err)
	}
	defer r.Close()
	views, err := r.ReadEntryLossLocks(context.Background(), "acct-7")
	if err != nil {
		t.Fatalf("ReadEntryLossLocks: %v", err)
	}
	return views
}

// testRelaxationDeps 는 실제 journal 열기를 쓰고 audit 만 주입함.
func testRelaxationDeps(path string, auditor journal.RiskRelaxationAuditor) riskRelaxationDeps {
	deps := productionRiskRelaxationDeps()
	deps.journalPath = func(*rootOptions) (string, error) { return path, nil }
	deps.openAuditor = func(string) (journal.RiskRelaxationAuditor, error) { return auditor, nil }
	deps.now = func() time.Time { return relaxationCLINow }
	return deps
}

func runRelaxationCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

// TestEngineRiskRelaxationCommandsDeclareWhatTheyAre 는 표면 모양을 고정함.
//
// 두 해제 명령은 mutating=true(대화형 에이전트가 자동 실행하지 않음 — 사용자 결정), show 는 false.
// 승인·운영자·사유는 필수이고 기본값이 비어 있어야 함 — 기본값이 있으면 아무도 안 고른 승인이 원장에 들어감.
func TestEngineRiskRelaxationCommandsDeclareWhatTheyAre(t *testing.T) {
	want := map[string]string{
		"tossctl engine entry-lock-release": "true",
		"tossctl engine risk-latch-release": "true",
		"tossctl engine risk-latch-show":    "false",
	}
	found := map[string]*cobra.Command{}
	for _, c := range leafCommands(newRootCmd()) {
		if _, ok := want[c.CommandPath()]; ok {
			found[c.CommandPath()] = c
		}
	}
	for path, mutating := range want {
		c := found[path]
		if c == nil {
			t.Fatalf("%s is not registered under engine", path)
		}
		if c.Annotations["mutating"] != mutating || c.Annotations["source"] != "local" {
			t.Errorf("%s annotations = %v, want mutating=%s source=local", path, c.Annotations, mutating)
		}
		if mutating != "true" {
			continue
		}
		for _, name := range []string{"account", "market", "approval", "operator", "reason"} {
			flag := c.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("%s has no --%s", path, name)
			}
			if flag.DefValue != "" {
				t.Errorf("%s --%s defaults to %q", path, name, flag.DefValue)
			}
			if ann := flag.Annotations[cobra.BashCompOneRequiredFlag]; len(ann) != 1 || ann[0] != "true" {
				t.Errorf("%s --%s is not required", path, name)
			}
		}
	}
}

// TestEngineEntryLockReleaseBindsWhatShowPrinted 는 show → release 한 바퀴를 임시 저널에서 돔.
func TestEngineEntryLockReleaseBindsWhatShowPrinted(t *testing.T) {
	path := relaxationTestJournal(t)
	auditor := &cliRelaxationAuditor{}
	deps := testRelaxationDeps(path, auditor)

	root := &rootOptions{outputFormat: "json"}
	out, err := runRelaxationCmd(t, newEngineRiskLatchShowCmd(root, deps), "--account", "acct-7", "--market", "kr")
	if err != nil {
		t.Fatalf("risk-latch-show: %v (%s)", err, out)
	}
	var show riskLatchShow
	if err := json.Unmarshal([]byte(out), &show); err != nil {
		t.Fatalf("show output is not JSON: %v (%s)", err, out)
	}
	if len(show.Locks) != 1 {
		t.Fatalf("show locks = %+v, want the one open lock", show.Locks)
	}
	shown := show.Locks[0]

	out, err = runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{}, deps),
		"--account", "acct-7", "--market", "KR", "--horizon", "short",
		"--lock-seq", strconv.FormatInt(shown.Lock.Seq, 10), "--expect-event", strconv.FormatInt(shown.LastEvent, 10),
		"--approval", "OPS-2026-0330-1", "--operator", "박지훈", "--reason", "loss verified as a data error")
	if err != nil {
		t.Fatalf("entry-lock-release: %v (%s)", err, out)
	}
	if !strings.Contains(out, "released entry loss lock") {
		t.Errorf("output = %q", out)
	}
	if locks := openLocks(t, path); len(locks) != 0 {
		t.Errorf("open locks after release = %+v", locks)
	}
	if len(auditor.lines) != 1 || !strings.Contains(auditor.lines[0], journal.AuditActionEntryLockRelease) ||
		!strings.Contains(auditor.lines[0], "OPS-2026-0330-1") || !strings.Contains(auditor.lines[0], "박지훈") {
		t.Errorf("audit lines = %q, want one entry_lock_release line naming the approval and the operator", auditor.lines)
	}
}

// TestEngineRelaxationRefusesBeforeOpeningAnything 는 입력 거절이 audit·원장 어느 것도 열기 전에 일어남을 잼.
func TestEngineRelaxationRefusesBeforeOpeningAnything(t *testing.T) {
	deps := riskRelaxationDeps{
		journalPath: func(*rootOptions) (string, error) { t.Error("journal path resolved"); return "", nil },
		openWriter: func(context.Context, string) (riskRelaxationWriter, error) {
			t.Error("journal opened")
			return nil, errors.New("unreachable")
		},
		openAuditor: func(string) (journal.RiskRelaxationAuditor, error) {
			t.Error("audit log opened")
			return nil, errors.New("unreachable")
		},
		now: func() time.Time { return relaxationCLINow },
	}
	base := []string{"--account", "acct-7", "--market", "KR", "--approval", "OPS-1", "--operator", "ops", "--reason", "why"}
	with := func(flag, value string, extra ...string) []string {
		args := append([]string{}, base...)
		for i := 0; i < len(args); i += 2 {
			if args[i] == flag {
				args[i+1] = value
			}
		}
		return append(args, extra...)
	}
	lock := []string{"--horizon", "SHORT", "--lock-seq", "1", "--expect-event", "0"}
	latch := []string{"--symbol", "005930", "--generation", "g1", "--expect-state", "abc"}
	cases := map[string]struct {
		cmd  func() *cobra.Command
		args []string
	}{
		"blank approval":   {func() *cobra.Command { return newEngineEntryLockReleaseCmd(&rootOptions{}, deps) }, with("--approval", "  ", lock...)},
		"blank operator":   {func() *cobra.Command { return newEngineEntryLockReleaseCmd(&rootOptions{}, deps) }, with("--operator", " ", lock...)},
		"blank reason":     {func() *cobra.Command { return newEngineEntryLockReleaseCmd(&rootOptions{}, deps) }, with("--reason", "", lock...)},
		"bad market":       {func() *cobra.Command { return newEngineEntryLockReleaseCmd(&rootOptions{}, deps) }, with("--market", "JP", lock...)},
		"bad horizon":      {func() *cobra.Command { return newEngineEntryLockReleaseCmd(&rootOptions{}, deps) }, with("--market", "KR", "--horizon", "LONG", "--lock-seq", "1", "--expect-event", "0")},
		"unset event":      {func() *cobra.Command { return newEngineEntryLockReleaseCmd(&rootOptions{}, deps) }, with("--market", "KR", "--horizon", "SHORT", "--lock-seq", "1", "--expect-event", "-1")},
		"latch no state":   {func() *cobra.Command { return newEngineRiskLatchReleaseCmd(&rootOptions{}, deps) }, with("--market", "KR", "--symbol", "005930", "--generation", "g1", "--expect-state", " ")},
		"latch no approve": {func() *cobra.Command { return newEngineRiskLatchReleaseCmd(&rootOptions{}, deps) }, with("--approval", "", latch...)},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := runRelaxationCmd(t, tc.cmd(), tc.args...); err == nil {
				t.Fatal("accepted")
			}
		})
	}
}

// TestEngineRelaxationDoesNotOpenTheJournalWithoutAnAuditLog 는 audit 로그가 없으면 원장을 열지도 않음을 잼.
//
// (*audit.Log)(nil).RecordAction 은 nil 을 돌려주므로 "열기 실패 → nil 로그로 진행" 구현은 audit 없이 해제를
// 기록함. 그래서 오류와 nil 둘 다 원장 열기 전에 멈춰야 함.
func TestEngineRelaxationDoesNotOpenTheJournalWithoutAnAuditLog(t *testing.T) {
	for name, open := range map[string]func(string) (journal.RiskRelaxationAuditor, error){
		"open error": func(string) (journal.RiskRelaxationAuditor, error) { return nil, errors.New("permission denied") },
		"nil log":    func(string) (journal.RiskRelaxationAuditor, error) { return nil, nil },
	} {
		t.Run(name, func(t *testing.T) {
			path := relaxationTestJournal(t)
			deps := testRelaxationDeps(path, nil)
			deps.openAuditor = open
			opened := false
			deps.openWriter = func(context.Context, string) (riskRelaxationWriter, error) {
				opened = true
				return nil, errors.New("must not open")
			}
			_, err := runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{}, deps),
				"--account", "acct-7", "--market", "KR", "--horizon", "SHORT", "--lock-seq", "1", "--expect-event", "0",
				"--approval", "OPS-1", "--operator", "ops", "--reason", "why")
			if err == nil || !strings.Contains(err.Error(), "nothing was changed") {
				t.Fatalf("err = %v, want the audit refusal", err)
			}
			if opened {
				t.Error("the journal was opened without an audit log")
			}
			if locks := openLocks(t, path); len(locks) != 1 {
				t.Errorf("open locks = %+v, want the lock still in force", locks)
			}
		})
	}
}

// TestEngineRelaxationFailingAuditChangesNothing 는 audit 쓰기 실패가 원장 commit 을 막음을 실제 저널에서 잼.
func TestEngineRelaxationFailingAuditChangesNothing(t *testing.T) {
	path := relaxationTestJournal(t)
	deps := testRelaxationDeps(path, &cliRelaxationAuditor{fail: true})
	locks := openLocks(t, path)
	_, err := runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{}, deps),
		"--account", "acct-7", "--market", "KR", "--horizon", "SHORT",
		"--lock-seq", strconv.FormatInt(locks[0].Lock.Seq, 10), "--expect-event", strconv.FormatInt(locks[0].LastEvent, 10),
		"--approval", "OPS-1", "--operator", "ops", "--reason", "why")
	if err == nil {
		t.Fatal("release succeeded although the audit line was not written")
	}
	if after := openLocks(t, path); len(after) != 1 {
		t.Errorf("open locks = %+v, want the lock still in force", after)
	}
}

// fakeRelaxationWriter 는 latch 해제 요청을 받아 적음(배관 시험용).
type fakeRelaxationWriter struct {
	latch  journal.RiskOverageLatchReleaseRequest
	closed bool
}

func (f *fakeRelaxationWriter) ReleaseEntryLossLock(context.Context, journal.EntryLossLockReleaseRequest) (journal.EntryLossLockReleaseRecord, error) {
	return journal.EntryLossLockReleaseRecord{}, errors.New("unexpected")
}

func (f *fakeRelaxationWriter) ReleaseRiskOverageLatch(_ context.Context, req journal.RiskOverageLatchReleaseRequest) (journal.RiskOverageLatchReleaseRecord, error) {
	f.latch = req
	return journal.RiskOverageLatchReleaseRecord{ReleaseSeq: 9}, nil
}

func (f *fakeRelaxationWriter) Close() error { f.closed = true; return nil }

// TestEngineRiskLatchReleaseCarriesTheBindingToTheJournal 은 사람이 적은 값이 그대로 원장 요청이 됨을 잼.
func TestEngineRiskLatchReleaseCarriesTheBindingToTheJournal(t *testing.T) {
	auditor := &cliRelaxationAuditor{}
	writer := &fakeRelaxationWriter{}
	deps := testRelaxationDeps("/unused", auditor)
	deps.openWriter = func(context.Context, string) (riskRelaxationWriter, error) { return writer, nil }
	out, err := runRelaxationCmd(t, newEngineRiskLatchReleaseCmd(&rootOptions{}, deps),
		"--account", "acct-7", "--market", "us", "--symbol", "AAPL", "--generation", "gen-3", "--expect-state", "digest-1",
		"--approval", "OPS-9", "--operator", "야간당직", "--reason", "limit raised by the risk desk")
	if err != nil {
		t.Fatalf("risk-latch-release: %v (%s)", err, out)
	}
	got := writer.latch
	wantOwner := riskbucket.OwnerKey{AccountID: "acct-7", Market: riskbucket.MarketUS, Symbol: "AAPL", ProspectiveGeneration: "gen-3"}
	if got.Owner != wantOwner || got.ExpectedStateDigest != "digest-1" || got.Actor != journal.RelaxationActorOperator ||
		got.Approval != "OPS-9" || !strings.Contains(got.Reason, "야간당직") || got.Auditor != journal.RiskRelaxationAuditor(auditor) ||
		!got.ReleasedAt.Equal(relaxationCLINow) {
		t.Errorf("request = %+v", got)
	}
	if !writer.closed {
		t.Error("journal writer not closed")
	}
	if !strings.Contains(out, "release 9") {
		t.Errorf("output = %q", out)
	}
}

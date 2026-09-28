package main

// a066 5.5(design D8) 운영자 해제 명령의 CLI 절반.
//
// 원장 쪽 계약은 internal/journal/a066_relaxation_test.go, 엔진 endpoint 쪽은 internal/app/engine/a066_risk_relaxation_test.go
// 가 잰다. 여기서 재는 것은 **명령**임: mutating 선언, 승인·운영자 없이 엔진에 닿지도 않음, 해제가 원장을 직접 열지 않고
// 엔진 endpoint 를 거침(생산 deps 로 실제 엔진 endpoint 에), 엔진이 없으면 거절, 결과 세 갈래(거절 · 결과 불명 ·
// 「완화됨·통지 실패」)를 다르게 말함. 모든 시험은 임시 디렉터리만 씀 — 운영 원장에는 닿지 않음(사용자 결정 2026-09-28).

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskrelaxation"
	"github.com/spf13/cobra"
)

var relaxationCLINow = time.Date(2026, 3, 30, 3, 0, 0, 0, time.UTC)

func runRelaxationCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

// relaxationEngine 은 임시 엔진 디렉터리에 원장(KR/SHORT 잠금 하나)·audit 로그·명령 서비스·제어 endpoint 를 세움 —
// `engine run` 이 세우는 것과 같은 조합이고, 명령은 생산 deps 로 여기에 붙음.
func relaxationEngine(t *testing.T) (dir, auditPath string, j *journal.Journal) {
	t.Helper()
	dir, err := os.MkdirTemp("", "a066-cli-relax-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	j, err = journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(dir, journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	if _, _, err := j.ActivateEntryLossLock(context.Background(), journal.EntryLossLock{AccountRef: "acct-7",
		Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonShort, Cause: "SHORT_LOSS_LIMIT", ActivatedAt: relaxationCLINow}); err != nil {
		t.Fatalf("ActivateEntryLossLock: %v", err)
	}
	auditPath = filepath.Join(dir, "audit.log")
	log, err := audit.Open(audit.Options{Path: auditPath, Subject: "engine"})
	if err != nil {
		t.Fatal(err)
	}
	service, err := engine.NewPositionPolicyCommandService(&engine.Context{Journal: j, Audit: log}, clock.NewFake(relaxationCLINow))
	if err != nil {
		t.Fatal(err)
	}
	server, err := engine.StartPositionPolicyCommandServer(dir, service)
	if err != nil {
		t.Fatalf("StartPositionPolicyCommandServer: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return dir, auditPath, j
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

// TestEngineEntryLockReleaseGoesThroughTheRunningEngine 은 생산 deps 로 show → release 한 바퀴를 돔. 해제는 엔진
// endpoint 를 거쳐 엔진의 원장·audit 로그·통지에 닿고, CLI 는 원장을 쓰기로 열지 않음(show 는 읽기 전용).
func TestEngineEntryLockReleaseGoesThroughTheRunningEngine(t *testing.T) {
	dir, auditPath, j := relaxationEngine(t)
	deps := productionRiskRelaxationDeps()

	out, err := runRelaxationCmd(t, newEngineRiskLatchShowCmd(&rootOptions{outputFormat: "json", configDir: dir}, deps),
		"--account", "acct-7", "--market", "kr")
	if err != nil {
		t.Fatalf("risk-latch-show: %v (%s)", err, out)
	}
	var show riskLatchShow
	if err := json.Unmarshal([]byte(out), &show); err != nil || len(show.Locks) != 1 {
		t.Fatalf("show = %s (%v)", out, err)
	}
	shown := show.Locks[0]

	out, err = runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{configDir: dir}, deps),
		"--account", "acct-7", "--market", "KR", "--horizon", "short",
		"--lock-seq", strconv.FormatInt(shown.Lock.Seq, 10), "--expect-event", strconv.FormatInt(shown.LastEvent, 10),
		"--approval", "OPS-2026-0330-1", "--operator", "박지훈", "--reason", "loss verified as a data error")
	if err != nil {
		t.Fatalf("entry-lock-release: %v (%s)", err, out)
	}
	if !strings.Contains(out, "released entry_loss_lock:acct-7/KR/SHORT") {
		t.Errorf("output = %q", out)
	}
	if locks, err := j.ReadEntryLossLocks(context.Background(), "acct-7"); err != nil || len(locks) != 0 {
		t.Errorf("open locks after release = %+v (%v)", locks, err)
	}
	body, err := os.ReadFile(auditPath)
	if err != nil || !strings.Contains(string(body), journal.AuditActionEntryLockRelease) || !strings.Contains(string(body), "박지훈") {
		t.Errorf("engine audit log = %s (%v)", body, err)
	}

	// 같은 값으로 다시: 이미 풀렸으므로 stale — "아무것도 안 바뀜"으로 말함.
	_, err = runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{configDir: dir}, deps),
		"--account", "acct-7", "--market", "KR", "--horizon", "short",
		"--lock-seq", strconv.FormatInt(shown.Lock.Seq, 10), "--expect-event", strconv.FormatInt(shown.LastEvent, 10),
		"--approval", "OPS-2026-0330-1", "--operator", "박지훈", "--reason", "loss verified as a data error")
	if err == nil || !errors.Is(err, riskrelaxation.ErrStale) || !strings.Contains(err.Error(), "nothing was released") {
		t.Fatalf("repeated release: %v", err)
	}
}

// TestEngineRelaxationRefusesWhenTheEngineIsNotRunning 은 엔진 endpoint 가 없으면 거절하고, 원장을 직접 열어 대신
// 쓰지 않음을 잼(Manager 판정 Q2(a)). 원장 파일이 있어도 잠금은 그대로여야 함.
func TestEngineRelaxationRefusesWhenTheEngineIsNotRunning(t *testing.T) {
	dir, err := os.MkdirTemp("", "a066-cli-stopped-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	j, err := journal.Open(context.Background(), journal.Options{Path: filepath.Join(dir, journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	lock, _, err := j.ActivateEntryLossLock(context.Background(), journal.EntryLossLock{AccountRef: "acct-7",
		Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonShort, Cause: "SHORT_LOSS_LIMIT", ActivatedAt: relaxationCLINow})
	if err != nil {
		t.Fatal(err)
	}
	_, err = runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{configDir: dir}, productionRiskRelaxationDeps()),
		"--account", "acct-7", "--market", "KR", "--horizon", "SHORT", "--lock-seq", strconv.FormatInt(lock.Seq, 10), "--expect-event", "0",
		"--approval", "OPS-1", "--operator", "ops", "--reason", "why")
	if err == nil || !strings.Contains(err.Error(), "not running") || !strings.Contains(err.Error(), "nothing was released") ||
		!strings.Contains(err.Error(), "retry") {
		t.Fatalf("err = %v, want the engine-not-running refusal naming the retry", err)
	}
	if locks, err := j.ReadEntryLossLocks(context.Background(), "acct-7"); err != nil || len(locks) != 1 {
		t.Fatalf("locks = %+v (%v), want the lock still in force", locks, err)
	}
}

// fakeRelaxationClient 는 엔진 endpoint 대역임(결과 갈래 시험용).
type fakeRelaxationClient struct {
	result riskrelaxation.Result
	err    error
	lock   riskrelaxation.EntryLockReleaseRequest
	latch  riskrelaxation.LatchReleaseRequest
	calls  int
}

func (f *fakeRelaxationClient) ReleaseEntryLossLock(_ context.Context, req riskrelaxation.EntryLockReleaseRequest) (riskrelaxation.Result, error) {
	f.calls++
	f.lock = req
	return f.result, f.err
}

func (f *fakeRelaxationClient) ReleaseRiskOverageLatch(_ context.Context, req riskrelaxation.LatchReleaseRequest) (riskrelaxation.Result, error) {
	f.calls++
	f.latch = req
	return f.result, f.err
}

func fakeRelaxationDeps(client *fakeRelaxationClient) riskRelaxationDeps {
	return riskRelaxationDeps{
		engineDir: func(*rootOptions) (string, error) { return "/engine", nil },
		dial:      func(context.Context, string) (riskRelaxationClient, error) { return client, nil },
	}
}

var lockArgs = []string{"--account", "acct-7", "--market", "KR", "--horizon", "SHORT", "--lock-seq", "3", "--expect-event", "2",
	"--approval", "OPS-1", "--operator", "ops", "--reason", "why"}

// TestEngineRelaxationRefusesBeforeReachingTheEngine 은 입력 거절이 엔진에 닿기 전에 일어남을 잼.
func TestEngineRelaxationRefusesBeforeReachingTheEngine(t *testing.T) {
	with := func(flag, value string) []string {
		args := append([]string{}, lockArgs...)
		for i := 0; i < len(args); i += 2 {
			if args[i] == flag {
				args[i+1] = value
			}
		}
		return args
	}
	latch := []string{"--account", "acct-7", "--market", "KR", "--symbol", "005930", "--generation", "g1", "--expect-state", " ",
		"--approval", "OPS-1", "--operator", "ops", "--reason", "why"}
	cases := map[string]struct {
		latch bool
		args  []string
	}{
		"blank approval": {false, with("--approval", "  ")},
		"blank operator": {false, with("--operator", " ")},
		"blank reason":   {false, with("--reason", "")},
		"bad market":     {false, with("--market", "JP")},
		"bad horizon":    {false, with("--horizon", "LONG")},
		"unset event":    {false, with("--expect-event", "-1")},
		"no lock number": {false, with("--lock-seq", "0")},
		"latch no state": {true, latch},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			deps := riskRelaxationDeps{
				engineDir: func(*rootOptions) (string, error) { t.Error("engine directory resolved"); return "", nil },
				dial: func(context.Context, string) (riskRelaxationClient, error) {
					t.Error("engine dialled")
					return nil, errors.New("unreachable")
				},
			}
			cmd := newEngineEntryLockReleaseCmd(&rootOptions{}, deps)
			if tc.latch {
				cmd = newEngineRiskLatchReleaseCmd(&rootOptions{}, deps)
			}
			if _, err := runRelaxationCmd(t, cmd, tc.args...); err == nil {
				t.Fatal("accepted")
			}
		})
	}
}

// TestEngineRelaxationOutcomesAreToldApart 는 세 결과 갈래가 다르게 말해짐을 잼.
func TestEngineRelaxationOutcomesAreToldApart(t *testing.T) {
	t.Run("refused by the engine", func(t *testing.T) {
		client := &fakeRelaxationClient{err: riskrelaxation.ErrStale}
		_, err := runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{}, fakeRelaxationDeps(client)), lockArgs...)
		if err == nil || !strings.Contains(err.Error(), "refused, nothing was released") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("outcome unknown", func(t *testing.T) {
		client := &fakeRelaxationClient{err: errors.New("connection reset")}
		_, err := runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{}, fakeRelaxationDeps(client)), lockArgs...)
		if err == nil || !strings.Contains(err.Error(), "outcome is unknown") || strings.Contains(err.Error(), "nothing was released") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("released but not notified", func(t *testing.T) {
		client := &fakeRelaxationClient{result: riskrelaxation.Result{ReleaseSeq: 5, Target: "entry_loss_lock:acct-7/KR/SHORT",
			ReleasedAt: "2026-03-30T03:00:00Z", NotifyError: "outbox disk full"}}
		out, err := runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{}, fakeRelaxationDeps(client)), lockArgs...)
		if err == nil || !strings.Contains(err.Error(), "완화됨·통지 실패") || !strings.Contains(err.Error(), "outbox disk full") {
			t.Fatalf("err = %v", err)
		}
		if !strings.Contains(out, "released entry_loss_lock:acct-7/KR/SHORT as release 5") {
			t.Fatalf("output = %q — the release is in effect and must be said", out)
		}
	})
	t.Run("released and notified", func(t *testing.T) {
		client := &fakeRelaxationClient{result: riskrelaxation.Result{ReleaseSeq: 6, Target: "t", Notified: true}}
		if _, err := runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{}, fakeRelaxationDeps(client)), lockArgs...); err != nil {
			t.Fatalf("err = %v", err)
		}
		want := riskrelaxation.EntryLockReleaseRequest{AccountRef: "acct-7", Market: "KR", Horizon: "SHORT", LockSeq: 3,
			ExpectedLastEvent: 2, Operator: "ops", Approval: "OPS-1", Reason: "why"}
		if client.lock != want || client.calls != 1 {
			t.Fatalf("request = %+v calls=%d", client.lock, client.calls)
		}
	})
}

// TestEngineRiskLatchReleaseCarriesTheBinding 은 사람이 적은 값이 그대로 엔진 요청이 됨을 잼.
func TestEngineRiskLatchReleaseCarriesTheBinding(t *testing.T) {
	client := &fakeRelaxationClient{result: riskrelaxation.Result{ReleaseSeq: 9, Target: "risk_owner:acct-7/US/AAPL/gen-3", Notified: true}}
	out, err := runRelaxationCmd(t, newEngineRiskLatchReleaseCmd(&rootOptions{}, fakeRelaxationDeps(client)),
		"--account", "acct-7", "--market", "us", "--symbol", "AAPL", "--generation", "gen-3", "--expect-state", "digest-1",
		"--approval", "OPS-9", "--operator", "야간당직", "--reason", "limit raised by the risk desk")
	if err != nil {
		t.Fatalf("risk-latch-release: %v (%s)", err, out)
	}
	want := riskrelaxation.LatchReleaseRequest{AccountRef: "acct-7", Market: "US", Symbol: "AAPL", Generation: "gen-3",
		ExpectedState: "digest-1", Operator: "야간당직", Approval: "OPS-9", Reason: "limit raised by the risk desk"}
	if client.latch != want {
		t.Errorf("request = %+v", client.latch)
	}
	if !strings.Contains(out, "release 9") {
		t.Errorf("output = %q", out)
	}
}

// TestEngineRelaxationTreatsADialErrorAsNotReached 는 dial 오류가 클라이언트와 함께 와도 요청을 보내지 않음을 잼
// (뮤테이션 C04 생존 → 추가). 오류가 난 연결로 해제를 보내면 결과를 말할 수 없음.
func TestEngineRelaxationTreatsADialErrorAsNotReached(t *testing.T) {
	client := &fakeRelaxationClient{result: riskrelaxation.Result{Notified: true}}
	deps := fakeRelaxationDeps(client)
	deps.dial = func(context.Context, string) (riskRelaxationClient, error) {
		return client, errors.New("health check failed")
	}
	_, err := runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{}, deps), lockArgs...)
	if err == nil || !strings.Contains(err.Error(), "nothing was released") || client.calls != 0 {
		t.Fatalf("err = %v calls = %d", err, client.calls)
	}
}

// TestEngineRiskLatchShowNeverCreatesOrMigratesAJournal 은 show 가 읽기 전용 연결만 씀을 잼(리뷰 R3: openReader 를
// journal.Open 으로 바꾼 변이 생존 — 90e5170d 가 고친 단일 writer 위험 그 자체). journal.Open 은 없는 원장을 만들고
// 이주하므로, 빈 디렉터리에서 show 뒤에 원장 파일이 생기면 writer 로 연 것임.
func TestEngineRiskLatchShowNeverCreatesOrMigratesAJournal(t *testing.T) {
	dir := t.TempDir()
	_, err := runRelaxationCmd(t, newEngineRiskLatchShowCmd(&rootOptions{outputFormat: "json", configDir: dir}, productionRiskRelaxationDeps()),
		"--account", "acct-7")
	if err == nil || !errors.Is(err, journal.ErrJournalMissing) {
		t.Fatalf("show on an empty directory: %v, want ErrJournalMissing", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, journal.DBFileName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("show created a journal (stat err %v)", statErr)
	}
}

// TestEngineRelaxationNamedRefusalsSayNothingWasReleased 는 엔진의 이름 있는 거절(audit 불가 · 봉인 불일치)이 "결과
// 불명"이 아니라 "거절, 아무것도 안 풀림"으로 말해짐을 잼(리뷰 R2 P3 · R1 P2 · R3 CX19).
func TestEngineRelaxationNamedRefusalsSayNothingWasReleased(t *testing.T) {
	for _, refusal := range []error{riskrelaxation.ErrAuditUnavailable, riskrelaxation.ErrStateMismatch, riskrelaxation.ErrInvalidRequest, riskrelaxation.ErrUnwired} {
		client := &fakeRelaxationClient{err: refusal}
		_, err := runRelaxationCmd(t, newEngineEntryLockReleaseCmd(&rootOptions{}, fakeRelaxationDeps(client)), lockArgs...)
		if err == nil || !strings.Contains(err.Error(), "refused, nothing was released") || strings.Contains(err.Error(), "outcome is unknown") {
			t.Errorf("%v: err = %v", refusal, err)
		}
	}
}

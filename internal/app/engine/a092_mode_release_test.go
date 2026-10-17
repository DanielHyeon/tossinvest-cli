package engine

// a092 착지 단위 ④ — 사람의 모드 완화(`tossctl engine mode-release`)의 엔진 쪽. 델타 ADDED 완화 경로 목록 · 22.3 C12~C17 · 23.3 K16 · 24.3 M12.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

type a092RecordingAuditor struct {
	calls []string
	fail  bool
}

func (a *a092RecordingAuditor) RecordAction(action, setting, value, detail string) error {
	a.calls = append(a.calls, action+"|"+setting+"|"+value+"|"+detail)
	if a.fail {
		return errors.New("audit disk full")
	}
	return nil
}

type a092ReleaseFixture struct {
	j       *journal.Journal
	gate    *execgw.EntryGate
	n       *obs.Notifier
	auditor *a092RecordingAuditor
	ops     *ModeOperations
}

// a092ReleaseEngine 은 생산 조립의 모드 배선(bindOperatingModeProjection)으로 게이트를 묶고 ENTRY_BLOCKED 로 강화된 계좌를 만듦.
func a092ReleaseEngine(t *testing.T) *a092ReleaseFixture {
	t.Helper()
	ctx := context.Background()
	j := openTestJournal(t)
	gate := execgw.NewEntryGate(clock.System(), nil)
	if err := bindOperatingModeProjection(ctx, j, gate, a092Account, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := j.EscalateOperatingMode(ctx, a092Account, journal.ModeTriggerCredentialRejected, nil); err != nil {
		t.Fatal(err)
	}
	n := &obs.Notifier{Journal: j, Gate: gate, AccountRef: a092Account}
	auditor := &a092RecordingAuditor{}
	return &a092ReleaseFixture{j: j, gate: gate, n: n, auditor: auditor,
		ops: newModeOperations(j, gate, n, auditor, a092Account)}
}

func a092ReleaseRequest() ModeReleaseRequest {
	return ModeReleaseRequest{To: "normal", Operator: "박지훈", Approval: "OPS-2026-0930-1", Reason: "credential rotated and verified"}
}

func a092ModeLatched(g *execgw.EntryGate) bool {
	_, ok := g.Blocks()[execgw.ReasonOperatingModeBlocked]
	return ok
}

// 사람의 완화: audit 가 commit 앞에 남고 · OPERATOR 행이 승인 참조와 함께 · 산 게이트의 모드 사유가 풀리고 · 다른 사유는 남고 보임 ·
// 통지 행이 기록 전용으로 PENDING.
func TestA092AHumanReleaseOpensTheLiveGateAndSaysWhatRemains(t *testing.T) {
	fx := a092ReleaseEngine(t)
	fx.gate.Block(execgw.ReasonAlertUndelivered, "backlog")
	res, err := fx.ops.Release(context.Background(), a092ReleaseRequest())
	if err != nil {
		t.Fatalf("Release: %v", err)
	}
	if !res.Changed || res.Mode != journal.ModeNormal || res.Seq <= 0 || res.TransitionID == "" {
		t.Fatalf("result = %+v", res)
	}
	if a092ModeLatched(fx.gate) {
		t.Error("the live mode latch survived a human release")
	}
	if len(res.EntryBlocks) != 1 || res.EntryBlocks[0] != string(execgw.ReasonAlertUndelivered) {
		t.Errorf("remaining reasons = %v, want the alert latch only", res.EntryBlocks)
	}
	if len(fx.auditor.calls) != 1 || !strings.Contains(fx.auditor.calls[0], "OPS-2026-0930-1") ||
		!strings.Contains(fx.auditor.calls[0], "박지훈") {
		t.Errorf("audit calls = %v", fx.auditor.calls)
	}
	cur, _ := fx.j.CurrentOperatingMode(context.Background(), a092Account)
	if cur.Actor != journal.ModeActorOperator || !strings.Contains(cur.Cause, "approved-by: OPS-2026-0930-1") {
		t.Errorf("ledger row = %+v", cur)
	}
	if !res.Notified || !res.NoticePending || res.NotifyError != "" {
		t.Errorf("notice: notified=%v pending=%v err=%q, want a recorded pending notice", res.Notified, res.NoticePending, res.NotifyError)
	}
}

// 필수 칸 넷 — 하나라도 비면 거절하고 아무것도 안 바뀜. 허용 밖 목적 모드도.
func TestA092AReleaseNeedsEveryField(t *testing.T) {
	for name, mutate := range map[string]func(*ModeReleaseRequest){
		"no operator": func(r *ModeReleaseRequest) { r.Operator = " " },
		"no approval": func(r *ModeReleaseRequest) { r.Approval = "" },
		"no reason":   func(r *ModeReleaseRequest) { r.Reason = "" },
		"halt target": func(r *ModeReleaseRequest) { r.To = journal.ModeHaltAll },
		"no target":   func(r *ModeReleaseRequest) { r.To = "" },
	} {
		t.Run(name, func(t *testing.T) {
			fx := a092ReleaseEngine(t)
			req := a092ReleaseRequest()
			mutate(&req)
			if _, err := fx.ops.Release(context.Background(), req); !errors.Is(err, ErrModeReleaseInvalid) {
				t.Fatalf("err = %v, want invalid", err)
			}
			if !a092ModeLatched(fx.gate) || len(fx.auditor.calls) != 0 {
				t.Errorf("a refused release changed something: latched=%v audit=%v", a092ModeLatched(fx.gate), fx.auditor.calls)
			}
		})
	}
}

// audit 로그가 없는 엔진은 거절 — 인터페이스 안의 타입 있는 nil 포함.
func TestA092AReleaseWithoutAnAuditLogIsRefused(t *testing.T) {
	fx := a092ReleaseEngine(t)
	fx.ops.auditor = nil
	if _, err := fx.ops.Release(context.Background(), a092ReleaseRequest()); !errors.Is(err, ErrModeReleaseUnavailable) {
		t.Fatalf("err = %v, want unavailable", err)
	}
	if !a092ModeLatched(fx.gate) {
		t.Error("the mode latch was released without an audit log")
	}
}

// audit 쓰기가 실패하면 원장 행 0 · 게이트 그대로(원장 규칙 핀).
func TestA092AReleaseWhoseAuditFailsChangesNothing(t *testing.T) {
	fx := a092ReleaseEngine(t)
	fx.auditor.fail = true
	if _, err := fx.ops.Release(context.Background(), a092ReleaseRequest()); err == nil {
		t.Fatal("a release whose audit write failed reported success")
	}
	if cur, _ := fx.j.CurrentOperatingMode(context.Background(), a092Account); cur.Mode != journal.ModeEntryBlocked {
		t.Errorf("mode = %s, want ENTRY_BLOCKED — nothing may change without the audit line", cur.Mode)
	}
	if !a092ModeLatched(fx.gate) {
		t.Error("the live gate opened without an audit line")
	}
}

// 커밋 뒤 통지 기록만 실패 — 완화는 성공으로, 통지 실패를 함께(M12). 입구는 전달 실패 사유로 잠금.
func TestA092AReleaseWhoseNoticeCannotBeRecordedStillReportsTheRelease(t *testing.T) {
	fx := a092ReleaseEngine(t)
	broken, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = broken.Close()
	fx.ops.announcer = obs.RecordOnly{N: &obs.Notifier{Journal: broken, Gate: fx.gate}}
	res, err := fx.ops.Release(context.Background(), a092ReleaseRequest())
	if err != nil {
		t.Fatalf("Release returned an error for a committed release: %v", err)
	}
	if !res.Changed || res.Notified || res.NotifyError == "" || res.Mode != journal.ModeNormal {
		t.Fatalf("result = %+v, want released and notice failure reported", res)
	}
	found := false
	for _, r := range res.EntryBlocks {
		found = found || r == string(execgw.ReasonAlertUndelivered)
	}
	if !found {
		t.Errorf("remaining reasons = %v, want the undelivered latch the failed record raised", res.EntryBlocks)
	}
}

// K16: 결과는 다시 읽은 값 — 같은 호출 안의 다른 경로가 방금 푼 모드를 다시 조이면 결과가 그것을 말함.
type a092RetighteningAnnouncer struct{ j *journal.Journal }

func (a a092RetighteningAnnouncer) AnnounceOperatingMode(ctx context.Context, _ string, _ journal.OperatingModeRecord) error {
	_, _, err := a.j.EscalateOperatingMode(ctx, a092Account, journal.ModeTriggerCriticalAlertUndelivered, nil)
	return err
}

func TestA092TheReleaseResultIsReReadNotAssumed(t *testing.T) {
	fx := a092ReleaseEngine(t)
	fx.ops.announcer = a092RetighteningAnnouncer{j: fx.j}
	res, err := fx.ops.Release(context.Background(), a092ReleaseRequest())
	if err != nil {
		t.Fatal(err)
	}
	if res.Mode != journal.ModeEntryBlocked {
		t.Errorf("result mode = %s, want ENTRY_BLOCKED — the mode was re-tightened inside the call", res.Mode)
	}
	found := false
	for _, r := range res.EntryBlocks {
		found = found || r == string(execgw.ReasonOperatingModeBlocked)
	}
	if !found {
		t.Errorf("remaining reasons = %v, want the re-tightened mode latch", res.EntryBlocks)
	}
}

// 변화 없음 — 이미 NORMAL 이면 행 · 통지 없음.
func TestA092AReleaseToTheCurrentModeChangesNothing(t *testing.T) {
	fx := a092ReleaseEngine(t)
	if _, err := fx.ops.Release(context.Background(), a092ReleaseRequest()); err != nil {
		t.Fatal(err)
	}
	res, err := fx.ops.Release(context.Background(), a092ReleaseRequest())
	if err != nil {
		t.Fatal(err)
	}
	if res.Changed || res.Notified || res.NoticePending {
		t.Errorf("result = %+v, want no change and no notice", res)
	}
}

// 원장만 고친 완화는 산 게이트를 풀지 않음 — 그래서 명령은 엔진을 거침.
func TestA092ALedgerOnlyReleaseDoesNotOpenTheLiveGate(t *testing.T) {
	fx := a092ReleaseEngine(t)
	other, err := journal.Open(context.Background(), journal.Options{
		Path: fx.j.Path(), FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if _, _, err := other.TransitionOperatingMode(context.Background(), journal.TransitionModeRequest{
		AccountRef: a092Account, Mode: journal.ModeNormal, Cause: "outside", Actor: journal.ModeActorOperator,
		Approval: "OPS-X", Auditor: &a092RecordingAuditor{},
	}); err != nil {
		t.Fatal(err)
	}
	if !a092ModeLatched(fx.gate) {
		t.Error("a relaxation written from outside the engine opened the live gate")
	}
}

// 조립: Context 에서 만들 때 핸들이 하나라도 없으면 거절.
// 게이트 준비 gstack 리뷰(testing): 핸들마다 하나씩 빼서 잼 — 빈 Context 하나로는 첫 검사(Journal)만 잰다. 양성 대조: 전부 있으면 섬.
func TestA092TheReleaseSurfaceNeedsEveryHandle(t *testing.T) {
	if _, err := (&Context{}).ModeOperations(); !errors.Is(err, ErrRuntimeUnavailable) {
		t.Errorf("err = %v, want runtime unavailable", err)
	}
	fx := a092ReleaseEngine(t)
	full := func() *Context {
		return &Context{Journal: fx.j, Notifier: fx.n, Entry: fx.gate, AccountRef: a092Account}
	}
	if ops, err := full().ModeOperations(); err != nil || ops == nil {
		t.Fatalf("control: a context with every handle built no surface: %v", err)
	}
	for name, strip := range map[string]func(*Context){
		"journal":       func(c *Context) { c.Journal = nil },
		"notifier":      func(c *Context) { c.Notifier = nil },
		"entry-gate":    func(c *Context) { c.Entry = nil },
		"account":       func(c *Context) { c.AccountRef = "" },
		"blank-account": func(c *Context) { c.AccountRef = "   " },
	} {
		c := full()
		strip(c)
		if _, err := c.ModeOperations(); !errors.Is(err, ErrRuntimeUnavailable) {
			t.Errorf("without %s: err = %v, want runtime unavailable", name, err)
		}
	}
}

// 26라운드 codex #5: 커밋 뒤 재읽기가 실패해도 완화 사실 · 통지 결과는 보존되고, 읽지 못한 상태는 추정하지 않음.
func TestA092AReReadFailureKeepsTheCommittedRelease(t *testing.T) {
	fx := a092ReleaseEngine(t)
	fx.ops.current = func(context.Context, string) (journal.ModeSnapshot, error) {
		return journal.ModeSnapshot{}, errors.New("ledger read failed")
	}
	res, err := fx.ops.Release(context.Background(), a092ReleaseRequest())
	if err != nil {
		t.Fatalf("Release returned an error after a committed release: %v", err)
	}
	if !res.Changed || res.TransitionID == "" || !res.Notified || res.ReReadError == "" {
		t.Fatalf("result = %+v, want the committed release kept and the re-read failure named", res)
	}
	if res.Mode != "" || res.EntryBlocks != nil {
		t.Errorf("unread state was guessed: mode=%q blocks=%v", res.Mode, res.EntryBlocks)
	}
	if a092ModeLatched(fx.gate) {
		t.Error("the live gate did not open — the release itself must stand")
	}
}

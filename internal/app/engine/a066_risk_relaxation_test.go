package engine

// a066 5.5(design D8) 운영자 해제의 엔진 절반 — 제어 endpoint 를 거친 해제.
//
// journal API 의 판정(OPERATOR·승인·결속·audit-before-commit)은 internal/journal/a066_relaxation_test.go 가 잰다.
// 여기서 재는 것: (1) 해제가 엔진의 journal 핸들과 엔진 audit 로그로 실제 원장에 닿는가(HTTP 를 거쳐), (2) audit 로그가
// 없으면 아무것도 안 바뀌는가, (3) 커밋 뒤 통지(alert enqueue)와 그 실패의 「완화됨·통지 실패」, (4) 오류 어휘가 전송을
// 건너 살아남는가, (5) capability 없는 엔진은 "제공하지 않음"인가. 모두 임시 저널이다.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskrelaxation"
)

var a066RelaxNow = time.Date(2026, 3, 30, 3, 0, 0, 0, time.UTC)

type a066RelaxFixture struct {
	dir      string
	j        *journal.Journal
	auditLog string
	service  *PositionPolicyCommandService
	client   *positionpolicyrpc.Client
}

// a066RelaxEngine 은 임시 엔진 디렉터리에 원장·audit 로그·명령 서비스·제어 endpoint 를 세우고 KR/SHORT 잠금 하나를 엶.
// withAudit=false 는 audit 로그 없는 엔진임.
func a066RelaxEngine(t *testing.T, withAudit bool, wrap func(*journal.Journal) positionPolicyRepository) *a066RelaxFixture {
	t.Helper()
	// Unix 경로 길이와 무관한 loopback TCP 이지만 제어 디렉터리 검사는 0700 을 요구함.
	dir, err := os.MkdirTemp("", "a066-relax-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(dir, journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	if _, _, err := j.ActivateEntryLossLock(context.Background(), journal.EntryLossLock{AccountRef: "acct-7",
		Market: riskbucket.MarketKR, Horizon: riskbucket.HorizonShort, Cause: "SHORT_LOSS_LIMIT", ActivatedAt: a066RelaxNow}); err != nil {
		t.Fatal(err)
	}
	ectx := &Context{Journal: j}
	fx := &a066RelaxFixture{dir: dir, j: j, auditLog: filepath.Join(dir, "audit.log")}
	if withAudit {
		log, err := audit.Open(audit.Options{Path: fx.auditLog, Subject: "engine"})
		if err != nil {
			t.Fatal(err)
		}
		ectx.Audit = log
	}
	service, err := NewPositionPolicyCommandService(ectx, clock.NewFake(a066RelaxNow))
	if err != nil {
		t.Fatal(err)
	}
	if wrap != nil {
		service.j = wrap(j)
	}
	fx.service = service
	server, err := StartPositionPolicyCommandServer(dir, service)
	if err != nil {
		t.Fatalf("StartPositionPolicyCommandServer: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client, err := positionpolicyrpc.Dial(context.Background(), positionpolicyrpc.DescriptorPath(dir))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	fx.client = client
	return fx
}

func (fx *a066RelaxFixture) openLocks(t *testing.T) []journal.EntryLossLockView {
	t.Helper()
	views, err := fx.j.ReadEntryLossLocks(context.Background(), "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	return views
}

func (fx *a066RelaxFixture) request(t *testing.T) riskrelaxation.EntryLockReleaseRequest {
	t.Helper()
	locks := fx.openLocks(t)
	if len(locks) != 1 {
		t.Fatalf("fixture locks = %+v", locks)
	}
	return riskrelaxation.EntryLockReleaseRequest{AccountRef: "acct-7", Market: "kr", Horizon: "short",
		LockSeq: locks[0].Lock.Seq, ExpectedLastEvent: locks[0].LastEvent,
		Operator: "박지훈", Approval: "OPS-2026-0330-1", Reason: "loss verified as a data error"}
}

func (fx *a066RelaxFixture) alerts(t *testing.T) []journal.Alert {
	t.Helper()
	pending, err := fx.j.PendingAlerts(context.Background(), 100)
	if err != nil {
		t.Fatal(err)
	}
	var out []journal.Alert
	for _, a := range pending {
		if a.Type == EventRiskRelaxation {
			out = append(out, a)
		}
	}
	return out
}

// TestA066EntryLockReleaseThroughTheEngineEndpoint 는 CLI 가 쓰는 바로 그 클라이언트로 엔진 endpoint 를 거친 해제가
// 엔진 원장·엔진 audit 로그·alert outbox 에 닿음을 잼.
func TestA066EntryLockReleaseThroughTheEngineEndpoint(t *testing.T) {
	fx := a066RelaxEngine(t, true, nil)
	req := fx.request(t)
	result, err := fx.client.ReleaseEntryLossLock(context.Background(), req)
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	if !result.Notified || result.ReleaseSeq <= 0 || result.Target != "entry_loss_lock:acct-7/KR/SHORT" {
		t.Fatalf("result = %+v", result)
	}
	if locks := fx.openLocks(t); len(locks) != 0 {
		t.Fatalf("open locks after release = %+v", locks)
	}
	body, err := os.ReadFile(fx.auditLog)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{journal.AuditActionEntryLockRelease, "OPS-2026-0330-1", "operator 박지훈"} {
		if !strings.Contains(string(body), want) {
			t.Errorf("audit log lacks %q:\n%s", want, body)
		}
	}
	alerts := fx.alerts(t)
	if len(alerts) != 1 || !strings.Contains(alerts[0].Body, "박지훈") || !strings.Contains(alerts[0].Body, "OPS-2026-0330-1") ||
		alerts[0].Severity != "critical" {
		t.Fatalf("relaxation notices = %+v", alerts)
	}
	// 제목·본문은 외부 전송(notifier)으로 나감 — 계좌 식별자를 싣지 않음(안전 불변식 8, 리뷰 R2 P2). 대상은 시장·horizon 으로 말함.
	if strings.Contains(alerts[0].Title, "acct-7") || strings.Contains(alerts[0].Body, "acct-7") ||
		!strings.Contains(alerts[0].Title, "entry_loss_lock:KR/SHORT") {
		t.Fatalf("published notice text carries the account or loses the target: %q / %q", alerts[0].Title, alerts[0].Body)
	}
	// 같은 요청을 다시 보내면 stale — 전송을 건너 타입이 살아남아야 CLI 가 "아무것도 안 바뀜"이라고 말할 수 있음.
	if _, err := fx.client.ReleaseEntryLossLock(context.Background(), req); !errors.Is(err, riskrelaxation.ErrStale) {
		t.Fatalf("repeated release: err=%v, want stale", err)
	}
	if got := len(fx.alerts(t)); got != 1 {
		t.Fatalf("a refused release enqueued a notice (%d)", got)
	}
}

// TestA066RelaxationRefusedWithoutAnEngineAuditLog 는 audit 로그 없는 엔진이 해제를 거절하고 아무것도 안 바꿈을 잼.
func TestA066RelaxationRefusedWithoutAnEngineAuditLog(t *testing.T) {
	fx := a066RelaxEngine(t, false, nil)
	_, err := fx.client.ReleaseEntryLossLock(context.Background(), fx.request(t))
	if !errors.Is(err, riskrelaxation.ErrAuditUnavailable) {
		t.Fatalf("err = %v, want audit unavailable", err)
	}
	if locks := fx.openLocks(t); len(locks) != 1 {
		t.Fatalf("locks = %+v, want the lock still in force", locks)
	}
	if got := len(fx.alerts(t)); got != 0 {
		t.Fatalf("notices = %d", got)
	}
}

// TestA066RelaxationRequestRefusals 는 모양이 틀린 요청이 invalid 로 건너오고 원장을 안 바꿈을 잼.
func TestA066RelaxationRequestRefusals(t *testing.T) {
	for name, mutate := range map[string]func(*riskrelaxation.EntryLockReleaseRequest){
		"blank operator": func(r *riskrelaxation.EntryLockReleaseRequest) { r.Operator = " " },
		"blank approval": func(r *riskrelaxation.EntryLockReleaseRequest) { r.Approval = "" },
		"blank reason":   func(r *riskrelaxation.EntryLockReleaseRequest) { r.Reason = "" },
		"bad market":     func(r *riskrelaxation.EntryLockReleaseRequest) { r.Market = "JP" },
	} {
		t.Run(name, func(t *testing.T) {
			fx := a066RelaxEngine(t, true, nil)
			req := fx.request(t)
			mutate(&req)
			if _, err := fx.client.ReleaseEntryLossLock(context.Background(), req); !errors.Is(err, riskrelaxation.ErrInvalidRequest) {
				t.Fatalf("err = %v, want invalid", err)
			}
			if locks := fx.openLocks(t); len(locks) != 1 {
				t.Fatalf("a refused release changed the lock: %+v", locks)
			}
		})
	}
}

// a066FailingNotices 는 alert enqueue 만 실패시키는 원장임(나머지는 실제 원장).
type a066FailingNotices struct{ *journal.Journal }

func (a066FailingNotices) EnqueueAlert(context.Context, journal.Alert) (int64, error) {
	return 0, errors.New("outbox disk full")
}

// TestA066ReleaseStandsWhenTheNoticeFails 는 커밋 뒤 통지 실패가 해제를 되돌리지 않고 결과가 그것을 말함을 잼.
func TestA066ReleaseStandsWhenTheNoticeFails(t *testing.T) {
	fx := a066RelaxEngine(t, true, func(j *journal.Journal) positionPolicyRepository { return a066FailingNotices{j} })
	result, err := fx.client.ReleaseEntryLossLock(context.Background(), fx.request(t))
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	if result.Notified || !strings.Contains(result.NotifyError, "outbox disk full") || result.ReleaseSeq <= 0 {
		t.Fatalf("result = %+v, want released but not notified", result)
	}
	if locks := fx.openLocks(t); len(locks) != 0 {
		t.Fatalf("the release did not stand: %+v", locks)
	}
}

// a066LatchRecorder 는 latch 해제 요청을 받아 적는 원장임(배관 시험 — latch 의 원장 판정은 journal 시험 몫).
type a066LatchRecorder struct {
	*journal.Journal
	got journal.RiskOverageLatchReleaseRequest
}

func (r *a066LatchRecorder) ReleaseRiskOverageLatch(_ context.Context, req journal.RiskOverageLatchReleaseRequest) (journal.RiskOverageLatchReleaseRecord, error) {
	r.got = req
	return journal.RiskOverageLatchReleaseRecord{ReleaseSeq: 4, Owner: req.Owner, ReleasedAt: req.ReleasedAt}, nil
}

// TestA066LatchReleaseCarriesTheBindingIntoTheEngine 은 요청 값이 그대로 엔진의 journal 요청이 되고, Auditor 가 엔진
// audit 로그이며 actor 가 OPERATOR 임을 잼.
func TestA066LatchReleaseCarriesTheBindingIntoTheEngine(t *testing.T) {
	recorder := &a066LatchRecorder{}
	fx := a066RelaxEngine(t, true, func(j *journal.Journal) positionPolicyRepository { recorder.Journal = j; return recorder })
	result, err := fx.client.ReleaseRiskOverageLatch(context.Background(), riskrelaxation.LatchReleaseRequest{
		AccountRef: "acct-7", Market: "us", Symbol: "AAPL", Generation: "gen-3", ExpectedState: "digest-1",
		Operator: "야간당직", Approval: "OPS-9", Reason: "limit raised by the risk desk"})
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	got := recorder.got
	want := riskbucket.OwnerKey{AccountID: "acct-7", Market: riskbucket.MarketUS, Symbol: "AAPL", ProspectiveGeneration: "gen-3"}
	if got.Owner != want || got.ExpectedStateDigest != "digest-1" || got.Actor != journal.RelaxationActorOperator ||
		got.Approval != "OPS-9" || got.Reason != "limit raised by the risk desk (operator 야간당직)" ||
		got.Auditor != journal.RiskRelaxationAuditor(fx.service.audit) || !got.ReleasedAt.Equal(a066RelaxNow) {
		t.Fatalf("journal request = %+v", got)
	}
	if !result.Notified || result.Target != "risk_owner:acct-7/US/AAPL/gen-3" || len(fx.alerts(t)) != 1 {
		t.Fatalf("result = %+v notices = %d", result, len(fx.alerts(t)))
	}
	if notice := fx.alerts(t)[0]; strings.Contains(notice.Title+notice.Body, "acct-7") || !strings.Contains(notice.Title, "risk_owner:US/AAPL/gen-3") {
		t.Fatalf("published notice text = %q / %q", notice.Title, notice.Body)
	}
}

// TestA066AnEngineWithoutTheCapabilityOffersNoRelease 는 capability 없는 명령 서비스의 endpoint 가 해제 route 를 내지
// 않고, 클라이언트가 그것을 "제공하지 않음"으로 말함을 잼(a079 와 같은 발견 방식의 대조군).
func TestA066AnEngineWithoutTheCapabilityOffersNoRelease(t *testing.T) {
	dir, err := os.MkdirTemp("", "a066-unwired-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	server, err := StartPositionPolicyCommandServer(dir, a066PolicyOnly{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client, err := positionpolicyrpc.Dial(context.Background(), positionpolicyrpc.DescriptorPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ReleaseEntryLossLock(context.Background(), riskrelaxation.EntryLockReleaseRequest{}); !errors.Is(err, riskrelaxation.ErrUnwired) {
		t.Fatalf("err = %v, want unwired", err)
	}
}

// TestA066RelaxationRoutesAreTheClientsRoutes 는 경로 문자열 두 벌(엔진·클라이언트)이 같음을 고정함.
func TestA066RelaxationRoutesAreTheClientsRoutes(t *testing.T) {
	if RiskRelaxationEntryLockPath != positionpolicyrpc.RiskRelaxationEntryLockPath ||
		RiskRelaxationLatchPath != positionpolicyrpc.RiskRelaxationLatchPath {
		t.Fatal("engine and client disagree on the relaxation routes")
	}
}

// a066PolicyOnly 는 해제 capability 가 없는 명령 서비스임(5.5 이전 엔진의 모양).
type a066PolicyOnly struct{}

func (a066PolicyOnly) List(context.Context) ([]positionpolicy.State, error) { return nil, nil }
func (a066PolicyOnly) Preview(context.Context, positionpolicy.Request) (positionpolicy.Preview, error) {
	return positionpolicy.Preview{}, positionpolicy.ErrInvalidRequest
}
func (a066PolicyOnly) Apply(context.Context, positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	return positionpolicy.State{}, positionpolicy.ErrInvalidRequest
}

// TestA066NoticeSurvivesTheCallerHangingUp 은 요청 문맥이 커밋 뒤 끊겨도 통지가 기록됨을 잼(뮤테이션 E10 생존 → 추가).
// 해제는 이미 커밋됐으므로, 클라이언트가 끊었다고 통지를 버리면 사람이 모르는 완화가 남음.
func TestA066NoticeSurvivesTheCallerHangingUp(t *testing.T) {
	fx := a066RelaxEngine(t, true, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := notifyRelaxation(ctx, fx.j, "entry_lock", 1, "entry_loss_lock:acct-7/KR/SHORT", "entry_loss_lock:KR/SHORT", "ops", "OPS-1", a066RelaxNow)
	if !result.Notified || len(fx.alerts(t)) != 1 {
		t.Fatalf("result = %+v notices = %d, want the notice recorded after a hang-up", result, len(fx.alerts(t)))
	}
}

// a066RefusingRepo 는 journal 판정 하나를 흉내 내는 원장임(오류 어휘 시험용).
type a066RefusingRepo struct {
	*journal.Journal
	err error
}

func (r a066RefusingRepo) ReleaseEntryLossLock(context.Context, journal.EntryLossLockReleaseRequest) (journal.EntryLossLockReleaseRecord, error) {
	return journal.EntryLossLockReleaseRecord{}, r.err
}

// TestA066JournalRefusalsCrossTheWireAsRefusals 는 아무것도 바꾸지 않은 journal 거절이 "결과 불명"(internal)이 아니라
// 이름 있는 거절로 건너옴을 잼(리뷰 R1 P2: 봉인과 어긋난 상태 · R2 P3: audit 쓰기 실패).
func TestA066JournalRefusalsCrossTheWireAsRefusals(t *testing.T) {
	for name, tc := range map[string]struct {
		journalErr error
		want       error
	}{
		"state does not match its seal":   {fmt.Errorf("%w: state digest drift", journal.ErrRiskBucketReplayMismatch), riskrelaxation.ErrStateMismatch},
		"audit line could not be written": {fmt.Errorf("%w: disk full", journal.ErrRiskRelaxationAuditFailed), riskrelaxation.ErrAuditUnavailable},
	} {
		t.Run(name, func(t *testing.T) {
			fx := a066RelaxEngine(t, true, func(j *journal.Journal) positionPolicyRepository { return a066RefusingRepo{j, tc.journalErr} })
			if _, err := fx.client.ReleaseEntryLossLock(context.Background(), fx.request(t)); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

// TestA066RelaxationRoutesRequireTheBearerToken 은 두 해제 route 가 토큰 없는·틀린 토큰 요청을 거절함을 잼(리뷰 R3:
// server.auth 를 뺀 등록이 생존).
func TestA066RelaxationRoutesRequireTheBearerToken(t *testing.T) {
	fx := a066RelaxEngine(t, true, nil)
	raw, err := os.ReadFile(positionpolicyrpc.DescriptorPath(fx.dir))
	if err != nil {
		t.Fatal(err)
	}
	var descriptor positionpolicyrpc.Descriptor
	if err := json.Unmarshal(raw, &descriptor); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{RiskRelaxationEntryLockPath, RiskRelaxationLatchPath} {
		for name, token := range map[string]string{"no token": "", "wrong token": strings.Repeat("x", len(descriptor.Token))} {
			req, err := http.NewRequest(http.MethodPost, "http://"+descriptor.Address+path, strings.NewReader(`{}`))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("%s %s: status %d, want 401", path, name, resp.StatusCode)
			}
		}
	}
	if locks := fx.openLocks(t); len(locks) != 1 {
		t.Fatalf("an unauthenticated request changed the lock: %+v", locks)
	}
}

// TestA066LatchReleaseRefusedWithoutAnEngineAuditLog 는 latch 경로도 audit 로그 없는 엔진에서 거절됨을 잼(R3: latch
// 경로가 relaxationAuditor 를 건너뛰는 변이 생존 — nil *audit.Log 가 audit 없이 커밋될 자리).
func TestA066LatchReleaseRefusedWithoutAnEngineAuditLog(t *testing.T) {
	recorder := &a066LatchRecorder{}
	fx := a066RelaxEngine(t, false, func(j *journal.Journal) positionPolicyRepository { recorder.Journal = j; return recorder })
	_, err := fx.client.ReleaseRiskOverageLatch(context.Background(), riskrelaxation.LatchReleaseRequest{
		AccountRef: "acct-7", Market: "US", Symbol: "AAPL", Generation: "gen-3", ExpectedState: "digest-1",
		Operator: "ops", Approval: "OPS-9", Reason: "why"})
	if !errors.Is(err, riskrelaxation.ErrAuditUnavailable) {
		t.Fatalf("err = %v, want audit unavailable", err)
	}
	if recorder.got.Auditor != nil || recorder.got.Owner.Symbol != "" {
		t.Fatalf("the journal was asked to release without an audit log: %+v", recorder.got)
	}
}

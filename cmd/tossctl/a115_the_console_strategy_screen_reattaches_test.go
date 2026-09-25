//go:build unix

package main

// a115 — 콘솔 전략 화면도 재부착한다 (design D1·D2, spec operator-console a115 delta).
//
// 병: 콘솔 부팅이 전략 projection 을 **한 번** dial 하고, 실패를 nil 로 접었다(편집 전 console.go
// B33–B37). 화면은 dormant 와 도달 불가를 구분할 줄 알지만 부팅이 접으므로 「descriptor 는 있는데
// 엔진이 안 받는다」가 NOT_CONFIGURED 로 보였고, 엔진이 늦게 뜨거나 재시작하면 콘솔 재시작 전까지
// 회복하지 않았다.
//
// 원형은 a109 D4 의 httpapi 재부착 시험과 a114 의 콘솔 lifecycle 시험이다. 여기서는 **결과**(진짜
// `strategyprojectionrpc.Start` 로 늦은 기동·재시작·잔재 → 화면 값)와 **기전**(렌더 없는 펌프 · 부팅
// dial 금지 · 판정 동치 · 재시도 경고 침묵 · 간격 가드 · 종료)을 둘 다 잰다.
//
// 화면 값은 진짜 콘솔(`console.New` → `/strategy-runtime`)이 그린 본문으로 판정한다 — reader 를 직접
// Read 하면 「부재」와 「못 읽음」이 같은 오류로 접혀 구분이 사라진다.

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// ---- 세우기 ------------------------------------------------------------------------

// a115Interval 은 이 시험 동안의 콘솔 재부착 간격이다. 운영값(30s)은 따로 핀한다.
func a115Interval(t *testing.T, interval time.Duration) {
	t.Helper()
	previous := consoleStrategyRuntimeRedialInterval
	consoleStrategyRuntimeRedialInterval = interval
	t.Cleanup(func() { consoleStrategyRuntimeRedialInterval = previous })
}

// a115Boot 는 콘솔 부팅의 전략 reader 를 세운다. 정리는 cancel 뒤 **진행 중 시도가 끝날 때까지**
// 기다린다 — 시도 goroutine 이 지워지는 임시 디렉터리를 만지거나 다음 시험으로 새지 않게 한다.
func a115Boot(t *testing.T, dir string, errOut io.Writer) (*strategyRuntimeAttachment, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	attachment := consoleStrategyRuntimeReaderFor(ctx, dir, errOut)
	if attachment == nil {
		t.Fatal("engineDir 가 있는데 전략 reader 가 nil 이다 — 화면이 「미배선」으로 굳는다")
	}
	t.Cleanup(func() {
		cancel()
		a109WaitFor(t, "시도 goroutine 정리", func() bool { return !attachment.inFlight() })
	})
	return attachment, cancel
}

// a115Projection 은 엔진이 내보내는 전략 projection 이다. KR 에 뚜렷한 판정을 실어 두면 어느 엔진에
// 붙었는지를 화면 값으로 구별할 수 있다.
func a115Projection(code strategyprojection.RefusalCode) a109Projection {
	at := time.Date(2026, 9, 26, 6, 0, 0, 0, time.UTC)
	return a109Projection{snapshot: strategyprojection.WithMarketFailure(
		strategyprojection.DormantSnapshot(at), strategyprojection.MarketKR, code, at)}
}

func a115StartEngine(t *testing.T, dir string, code strategyprojection.RefusalCode) *strategyprojectionrpc.Server {
	t.Helper()
	server, err := strategyprojectionrpc.Start(dir, a115Projection(code))
	if err != nil {
		t.Fatalf("엔진 projection endpoint 기동: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return server
}

// a115Screen 은 진짜 콘솔 하나다. reader 를 한 번 꽂고 화면을 여러 번 그린다.
type a115Screen struct {
	srv    *httptest.Server
	client *http.Client
}

func a115Console(t *testing.T, reader console.MultiMarketStrategyRuntimeReader) *a115Screen {
	t.Helper()
	app, err := console.New(console.Options{
		StartVerify: func(context.Context, verifylive.BatchConfirmer, io.Writer, string,
			[]verifylive.StepID) (verifylive.Summary, []verifylive.Entry, error) {
			return verifylive.Summary{}, nil, nil
		},
		StrategyRuntime: reader,
	})
	if err != nil {
		t.Fatalf("console.New: %v", err)
	}
	srv := httptest.NewServer(app.Handler())
	t.Cleanup(srv.Close)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	screen := &a115Screen{srv: srv, client: &http.Client{Jar: jar}}
	if page := screen.get(t, "/?session="+app.SessionToken()); page == "" {
		t.Fatal("세션 인증 응답이 비었다")
	}
	return screen
}

func (s *a115Screen) get(t *testing.T, path string) string {
	t.Helper()
	response, err := s.client.Get(s.srv.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d", path, response.StatusCode)
	}
	return string(body)
}

// verdict 는 지금 전략 화면이 운영자에게 말하는 것이다.
//
// dormant·도달 불가는 **안내 줄**로 가른다(templates_optimization.go 의 두 notice) — 시장 코드만 보면 live
// 스냅샷의 US(NOT_CONFIGURED)가 dormant 로 읽힌다. live 는 안내가 없고 KR 이 엔진의 판정을 싣는 것이다.
func (s *a115Screen) verdict(t *testing.T) string {
	t.Helper()
	page := s.get(t, "/strategy-runtime")
	dormant := strings.Contains(page, "runtime endpoint 미기동")
	unavailable := strings.Contains(page, "runtime projection을 읽지 못했다")
	switch {
	case dormant && !unavailable:
		return "dormant"
	case unavailable && !dormant:
		return "unavailable"
	case dormant || unavailable:
		return "both-notices"
	}
	for _, code := range []strategyprojection.RefusalCode{
		strategyprojection.RefusalEvidenceStale, strategyprojection.RefusalSchedulerBlocked,
	} {
		if strings.Contains(page, string(code)) {
			return "live:" + string(code)
		}
	}
	return "unknown"
}

// waitFor 는 화면이 want 가 되기를 기다린다. 시도는 백그라운드이므로 폴링이다.
func (s *a115Screen) waitFor(t *testing.T, want string) string {
	t.Helper()
	deadline := time.Now().Add(a109ReattachWindow)
	got := s.verdict(t)
	for got != want && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
		got = s.verdict(t)
	}
	return got
}

// a115Seat 는 자리의 reader 와 세대, 부착 보고 상태다 — 렌더 없이 자리를 보려고 잠금 아래서 읽는다.
func a115Seat(a *strategyRuntimeAttachment) (httpapi.StrategyRuntimeReader, uint64, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.reader, a.seat, a.attached
}

func a115LastTry(a *strategyRuntimeAttachment) time.Time {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lastTry
}

// a115SeatReads 는 **자리의 reader 를 직접** 읽는다 — wrapper 의 Read 를 거치지 않으므로 아무것도 깨우지
// 않는다(렌더 없는 관측). 붙은 엔진의 KR 판정을 돌려준다("" = 못 읽음).
func a115SeatReads(a *strategyRuntimeAttachment) strategyprojection.RefusalCode {
	reader, _, _ := a115Seat(a)
	if reader == nil {
		return ""
	}
	snapshot, err := reader.Read(context.Background())
	if err != nil {
		return ""
	}
	kr, ok := snapshot.Markets[strategyprojection.MarketKR]
	if !ok || kr.Error == nil {
		return ""
	}
	return kr.Error.Code
}

// ---- 결과: spec 시나리오 넷 ------------------------------------------------------------

// TestTheConsoleStrategyScreenShowsADeadDescriptorAsUnreachable 는 시나리오 「descriptor 가 남은 다운」이다.
//
// descriptor 와 socket 파일이 남았고 주인은 죽었다(전원 단절의 기본 모양, a108 S3). 편집 전 부팅은 dial
// 실패를 nil 로 접어 NOT_CONFIGURED 를 그렸다. 이제는 도달 불가를 그리고, 엔진이 뜨면 콘솔 재시작 없이
// live 로 돌아온다.
func TestTheConsoleStrategyScreenShowsADeadDescriptorAsUnreachable(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a108DeadSocketLeftover(t, dir)
	a115Interval(t, 20*time.Millisecond)
	errOut := &a109SyncWriter{}
	attachment, _ := a115Boot(t, dir, errOut)
	screen := a115Console(t, attachment)

	if got := screen.verdict(t); got != "unavailable" {
		t.Fatalf("죽은 descriptor 의 화면 = %q, want unavailable — 도달 불가가 미구성으로 접혔다\n%s",
			got, errOut.String())
	}
	a115StartEngine(t, dir, strategyprojection.RefusalEvidenceStale)
	if got := screen.waitFor(t, "live:EVIDENCE_STALE"); got != "live:EVIDENCE_STALE" {
		t.Fatalf("엔진 기동 뒤 %s 안에 화면이 회복되지 않았다 (%q)\n%s", a109ReattachWindow, got, errOut.String())
	}
}

// TestTheConsoleStrategyScreenRecoversWhenTheEngineStartsLater 는 시나리오 「미구성은 그대로 미구성이다」와
// 「descriptor 부재는 미기동 안내이고 재부착으로 회복한다」다.
func TestTheConsoleStrategyScreenRecoversWhenTheEngineStartsLater(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a115Interval(t, 20*time.Millisecond)
	errOut := &a109SyncWriter{}
	attachment, _ := a115Boot(t, dir, errOut)
	screen := a115Console(t, attachment)

	if got := screen.verdict(t); got != "dormant" {
		t.Fatalf("descriptor 없는 기동의 화면 = %q, want dormant — 미구성이 도달 불가로 오귀속됐다", got)
	}
	if strings.TrimSpace(errOut.String()) != "" {
		t.Fatalf("descriptor 부재는 부팅 경고가 아니다(오늘처럼 조용히): %q", errOut.String())
	}
	a115StartEngine(t, dir, strategyprojection.RefusalEvidenceStale)
	if got := screen.waitFor(t, "live:EVIDENCE_STALE"); got != "live:EVIDENCE_STALE" {
		t.Fatalf("엔진이 뒤늦게 뜬 뒤 %s 안에 화면이 회복되지 않았다 (%q)", a109ReattachWindow, got)
	}
}

// TestTheConsoleStrategyRuntimeAttachesWithoutAnyRender 는 펌프다(design D1 — 콘솔엔 publisher 가 없다).
// 화면이 하나도 열려 있지 않아도 엔진이 뜨면 자리가 live 가 된다.
func TestTheConsoleStrategyRuntimeAttachesWithoutAnyRender(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a115Interval(t, 20*time.Millisecond)
	attachment, _ := a115Boot(t, dir, io.Discard)
	a115StartEngine(t, dir, strategyprojection.RefusalEvidenceStale)
	a109WaitFor(t, "렌더 없이 붙기", func() bool {
		return a115SeatReads(attachment) == strategyprojection.RefusalEvidenceStale
	})
}

// TestTheConsoleStrategyScreenReattachesAfterTheEngineRestarts 는 시나리오 「가동 중 엔진 재시작」이다 —
// **화면을 열지 않아도** 새 endpoint 에 다시 붙는다. 부팅 때 잡은 client 는 새 토큰을 모른다.
//
// 펌프의 무조건 wake(design D1, freeze 리뷰 P1-2)가 여기서 필요하다: 렌더가 없으면 아무도 옛 client 를
// Read 하지 않으므로 `failed` 게이트로는 죽은 자리가 드러나지 않는다.
func TestTheConsoleStrategyScreenReattachesAfterTheEngineRestarts(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a115Interval(t, 20*time.Millisecond)
	first := a115StartEngine(t, dir, strategyprojection.RefusalEvidenceStale)
	attachment, _ := a115Boot(t, dir, io.Discard)
	if got := a115SeatReads(attachment); got != strategyprojection.RefusalEvidenceStale {
		t.Fatalf("부팅 때 떠 있는 엔진에 즉시 붙지 않았다: %q", got)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	a115StartEngine(t, dir, strategyprojection.RefusalSchedulerBlocked)
	a109WaitFor(t, "렌더 없이 재시작한 엔진에 다시 붙기", func() bool {
		return a115SeatReads(attachment) == strategyprojection.RefusalSchedulerBlocked
	})
	if got := a115Console(t, attachment).verdict(t); got != "live:SCHEDULER_BLOCKED" {
		t.Fatalf("재부착 뒤 화면 = %q", got)
	}
}

// ---- 기전 -------------------------------------------------------------------------

// TestRunConsoleNeverDialsTheStrategyProjectionItself 는 부팅 1회 dial 재도입의 구조 핀이다(base 에서 RED).
func TestRunConsoleNeverDialsTheStrategyProjectionItself(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "console.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var run *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "runConsole" {
			run = fn
		}
	}
	if run == nil {
		t.Fatal("runConsole 을 찾지 못했다")
	}
	dials, wires := 0, 0
	ast.Inspect(run.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "strategyprojectionrpc" && fun.Sel.Name == "Dial" {
				dials++
			}
		case *ast.Ident:
			if fun.Name == "consoleStrategyRuntimeReaderFor" {
				wires++
			}
		}
		return true
	})
	if dials != 0 {
		t.Errorf("runConsole 이 strategyprojectionrpc.Dial 을 %d번 직접 부른다 — 부팅 1회 client 는 굳고 실패는 nil 로 접힌다", dials)
	}
	if wires != 1 {
		t.Errorf("runConsole 이 consoleStrategyRuntimeReaderFor 를 %d번 부른다, want 1", wires)
	}
}

// a115Shape 은 해석 결과의 모양이다 — 판정 동치는 값이 아니라 모양(부재·sentinel·client)과 live 로 본다.
func a115Shape(reader httpapi.StrategyRuntimeReader, live bool) string {
	var kind string
	switch reader.(type) {
	case nil:
		kind = "absent"
	case unavailableStrategyRuntime:
		kind = "sentinel"
	case *strategyprojectionrpc.Client:
		kind = "client"
	default:
		kind = "other"
	}
	if closer, ok := reader.(io.Closer); ok {
		_ = closer.Close()
	}
	if live {
		return kind + "/live"
	}
	return kind
}

// TestTheConsoleResolutionMatchesTheDaemons 는 판정 동치다(design D1 — 옮겨 적은 판정은 시험으로 고정한다).
// 같은 디스크 상태에서 콘솔 해석과 httpapi `resolveStrategyRuntimeReader` 가 같은 모양을 내야 한다.
func TestTheConsoleResolutionMatchesTheDaemons(t *testing.T) {
	for name, test := range map[string]struct {
		setup func(t *testing.T, dir string)
		want  string
	}{
		"descriptor 부재": {setup: func(*testing.T, string) {}, want: "absent"},
		"조사 불가(ENOTDIR)": {setup: func(t *testing.T, dir string) {
			if err := os.WriteFile(strategyprojectionrpc.ControlDirectory(dir), []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
		}, want: "absent"},
		"반쪽 잔재(socket 없음)":     {setup: a108HalfLeftover, want: "sentinel"},
		"죽은 descriptor+socket": {setup: a108DeadSocketLeftover, want: "sentinel"},
		"live 엔진": {setup: func(t *testing.T, dir string) {
			a115StartEngine(t, dir, strategyprojection.RefusalEvidenceStale)
		}, want: "client/live"},
	} {
		t.Run(name, func(t *testing.T) {
			dir := a108HTTPAPIDir(t)
			test.setup(t, dir)
			daemon := a115Shape(resolveStrategyRuntimeReader(context.Background(),
				&rootOptions{configDir: dir}, io.Discard))
			consoleShape := a115Shape(resolveConsoleStrategyRuntime(context.Background(), dir, io.Discard))
			if daemon != test.want || consoleShape != test.want {
				t.Errorf("데몬 %q · 콘솔 %q, want 둘 다 %q", daemon, consoleShape, test.want)
			}
		})
	}
}

// TestTheAbsenceJudgementIsOneForBothPackages 는 판정 한 벌이다(design D2, freeze 리뷰 P1-1).
// presence 인터페이스는 한 타입이고, httpapi 판정은 strategyprojection 판정의 위임이다 — 세 자리 상태와
// 신호 없는 reader·nil 에서 둘이 같은 답을 한다.
func TestTheAbsenceJudgementIsOneForBothPackages(t *testing.T) {
	if reflect.TypeOf((*httpapi.StrategyRuntimePresence)(nil)).Elem() !=
		reflect.TypeOf((*strategyprojection.StrategyRuntimePresence)(nil)).Elem() {
		t.Error("presence 인터페이스가 두 벌이다 — httpapi 쪽은 alias 여야 한다")
	}
	clock := &a109Clock{now: time.Unix(1_760_000_000, 0).UTC()}
	attachment := a109Attachment(t, clock, io.Discard,
		func(context.Context) (httpapi.StrategyRuntimeReader, bool) { return nil, false })
	check := func(what string, reader httpapi.StrategyRuntimeReader, want bool) {
		t.Helper()
		leaf := strategyprojection.StrategyRuntimeAbsent(reader)
		if leaf != want || httpapi.StrategyRuntimeAbsent(reader) != want {
			t.Errorf("%s: strategyprojection=%v httpapi=%v, want %v", what, leaf,
				httpapi.StrategyRuntimeAbsent(reader), want)
		}
	}
	attachment.attach(nil, false)
	check("부재 자리", attachment, true)
	attachment.attach(unavailableStrategyRuntime{cause: errors.New("dial")}, false)
	check("sentinel 자리", attachment, false)
	attachment.attach(&a109FakeReader{}, true)
	check("live 자리", attachment, false)
	check("신호 없는 reader", &a109FakeReader{}, false)
	check("nil", nil, true)
}

// TestTheConsoleRetryResolutionIsSilent 는 재시도 경고 침묵이다(design D1, freeze 리뷰 P2-3).
// 엔진이 하루 내려가 있으면 간격마다 찍는 경고는 2880줄이다. 부팅 경고 한 줄 뒤에는 조용해야 한다.
func TestTheConsoleRetryResolutionIsSilent(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a108DeadSocketLeftover(t, dir)
	a115Interval(t, 10*time.Millisecond)
	errOut := &a109SyncWriter{}
	attachment, _ := a115Boot(t, dir, errOut)
	seen := map[time.Time]bool{}
	a109WaitFor(t, "재시도 세 번", func() bool {
		if at := a115LastTry(attachment); !at.IsZero() {
			seen[at] = true
		}
		return len(seen) >= 3
	})
	a109WaitFor(t, "시도 종료", func() bool { return !attachment.inFlight() })
	if got := strings.Count(errOut.String(), "연결할 수 없다"); got != 1 {
		t.Fatalf("연결 실패 경고 %d줄, want 1(부팅만)\n%s", got, errOut.String())
	}
}

// TestTheConsoleBootWarningDoesNotPromiseDormant 는 부팅 문구다(freeze 리뷰 P2-3). a115 뒤에 죽은
// descriptor 는 도달 불가로 뜬다 — 「dormant로 뜬다」는 거짓이다.
func TestTheConsoleBootWarningDoesNotPromiseDormant(t *testing.T) {
	for name, setup := range map[string]func(*testing.T, string){
		"죽은 descriptor": a108DeadSocketLeftover,
		"조사 불가": func(t *testing.T, dir string) {
			if err := os.WriteFile(strategyprojectionrpc.ControlDirectory(dir), []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir := a108HTTPAPIDir(t)
			setup(t, dir)
			var warning strings.Builder
			_, _ = resolveConsoleStrategyRuntime(context.Background(), dir, &warning)
			text := warning.String()
			if strings.TrimSpace(text) == "" {
				t.Fatal("부팅 경고가 없다 — 강등이 사유를 말하지 않는다")
			}
			if strings.Contains(text, "dormant로 뜬다") {
				t.Errorf("부팅 경고가 「dormant로 뜬다」를 말한다: %q", text)
			}
			if !strings.Contains(text, "재시작 없이") {
				t.Errorf("부팅 경고가 회복이 저절로 온다는 사실을 말하지 않는다: %q", text)
			}
		})
	}
}

// TestTheConsoleStrategyPumpNeedsAPositiveInterval 은 간격 가드다 — 0 주기 ticker 는 패닉이다(a114 선례).
// 간격이 0 이하면 펌프를 띄우지 않는다: 엔진이 떠도 렌더 없이는 시도가 없다.
func TestTheConsoleStrategyPumpNeedsAPositiveInterval(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a115Interval(t, 0)
	attachment, _ := a115Boot(t, dir, io.Discard)
	a115StartEngine(t, dir, strategyprojection.RefusalEvidenceStale)
	time.Sleep(150 * time.Millisecond)
	if at := a115LastTry(attachment); !at.IsZero() {
		t.Fatalf("간격 0 에서 펌프가 돌았다(lastTry=%s)", at)
	}
}

// TestTheConsoleStrategyPumpStopsWithTheConsole 은 종료다: 콘솔 ctx 가 끝나면 펌프도 시도도 멈춘다.
func TestTheConsoleStrategyPumpStopsWithTheConsole(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a115Interval(t, 10*time.Millisecond)
	attachment, cancel := a115Boot(t, dir, io.Discard)
	cancel()
	a109WaitFor(t, "시도 종료", func() bool { return !attachment.inFlight() })
	stopped := a115LastTry(attachment)
	a115StartEngine(t, dir, strategyprojection.RefusalEvidenceStale)
	time.Sleep(100 * time.Millisecond)
	if at := a115LastTry(attachment); !at.Equal(stopped) {
		t.Fatalf("콘솔이 끝난 뒤에도 시도가 돌았다 (%s → %s)", stopped, at)
	}
	if reader, _, _ := a115Seat(attachment); reader != nil {
		t.Fatalf("콘솔이 끝난 뒤에 자리가 붙었다: %T", reader)
	}
}

// TestTheProductionConsoleStrategyRedialIntervalIsThirtySeconds 는 배포가 쓰는 값이다 — 주입 가능성을
// 「기본값을 안 재도 되는 이유」로 쓰면 그것이 구멍이다.
func TestTheProductionConsoleStrategyRedialIntervalIsThirtySeconds(t *testing.T) {
	if consoleStrategyRuntimeRedialInterval != 30*time.Second {
		t.Errorf("콘솔 전략 재부착 간격 = %s, want 30s", consoleStrategyRuntimeRedialInterval)
	}
}

// ---- 뮤테이션 1판 생존자 (mutation-ledger.md) --------------------------------------------

// TestTheConsoleBootKeepsAnUnreachableEndpointUnreachable 은 부팅 nil 접힘의 **결정적** 핀이다(K1).
//
// 결과 시험(`…ShowsADeadDescriptorAsUnreachable`)도 K1 을 잡았지만 그것은 경합에 기댄다: 펌프의 첫 틱이나
// 첫 렌더의 wake 가 빈 자리를 sentinel 로 **승격**(a109 G1)하기 전에 화면을 그려야 빨개진다. 여기서는 간격을
// 1시간으로 두어 틱이 없게 하고, 렌더 없이 부팅 직후의 자리를 본다 — 부팅 해석이 접으면 자리는 nil 이다.
func TestTheConsoleBootKeepsAnUnreachableEndpointUnreachable(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a108DeadSocketLeftover(t, dir)
	a115Interval(t, time.Hour)
	attachment, _ := a115Boot(t, dir, io.Discard)
	reader, _, attached := a115Seat(attachment)
	if _, ok := reader.(unavailableStrategyRuntime); !ok || attached {
		t.Fatalf("죽은 descriptor 부팅의 자리 = %T(attached=%v), want sentinel — 부팅이 도달 불가를 부재로 접었다",
			reader, attached)
	}
}

// TestTheConsoleStrategyPumpReturnsWhenTheConsoleEnds 는 펌프 goroutine 의 종료다(K12).
//
// ctx 종료를 무시하는 펌프는 wake 가 ctx 를 보므로 **시도는** 멈춘다 — 그래서 시도 수로는 누수가 안 보였다.
// 콘솔이 내려가도 goroutine 과 ticker 가 남는다. 펌프를 직접 돌려 돌아오는지 본다.
func TestTheConsoleStrategyPumpReturnsWhenTheConsoleEnds(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	attachment := &strategyRuntimeAttachment{
		resolve: func(context.Context) (httpapi.StrategyRuntimeReader, bool) { return nil, false },
		ctx:     ctx, now: time.Now, interval: 10 * time.Millisecond, log: io.Discard,
	}
	done := make(chan struct{})
	go func() {
		pumpConsoleStrategyRuntime(attachment)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("콘솔 ctx 가 끝났는데 펌프가 돌아오지 않는다 — goroutine·ticker 누수")
	}
}

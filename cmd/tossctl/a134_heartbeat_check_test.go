package main

// a134_heartbeat_check_test.go — `tossctl engine heartbeat-check` 를 실제 명령 트리로 몰아 봄(design D4 · D6).
// 전송은 httptest ntfy 서버(루프백)로만 — 실 ntfy 로 나가지 않음. 저널 디렉터리는 t.TempDir().

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginewatch"
)

// a134Ntfy 는 받은 요청의 우선순위 · 태그 헤더와 본문을 기록하는 httptest ntfy.
type a134Ntfy struct {
	mu       sync.Mutex
	requests []a134Request
}

type a134Request struct {
	path, priority, tags, body string
}

func (n *a134Ntfy) handler(w http.ResponseWriter, r *http.Request) {
	var body bytes.Buffer
	_, _ = body.ReadFrom(r.Body)
	n.mu.Lock()
	n.requests = append(n.requests, a134Request{
		path: r.URL.Path, priority: r.Header.Get("Priority"), tags: r.Header.Get("Tags"), body: body.String(),
	})
	n.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (n *a134Ntfy) count() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.requests)
}

// a134Config 는 알림 블록과 autostart 를 담은 config.json 을 씀(엔진이 읽는 바로 그 파일 모양).
func a134Config(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func a134RunCheck(t *testing.T, dir string, args ...string) (string, string, error) {
	t.Helper()
	// 호스트 환경의 토픽 · 토큰이 시험 결과를 바꾸지 않게 비움(엔진과 같은 해석 규칙이 env 를 먼저 봄).
	t.Setenv("TOSSCTL_NTFY_TOPIC", "")
	t.Setenv("TOSSCTL_NTFY_TOKEN", "")
	root := &rootOptions{outputFormat: "table", configDir: dir}
	cmd := newEngineCmd(root)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(append([]string{"heartbeat-check"}, args...))
	err := cmd.ExecuteContext(context.Background())
	return out.String(), errOut.String(), err
}

// TestA134HeartbeatCheckIsALocalReadOnlyLeafUnderEngine: 등록 위치와 표지(source=local · mutating=false).
func TestA134HeartbeatCheckIsALocalReadOnlyLeafUnderEngine(t *testing.T) {
	var found bool
	for _, c := range leafCommands(newRootCmd()) {
		if c.CommandPath() != "tossctl engine heartbeat-check" {
			continue
		}
		found = true
		if c.Annotations["source"] != "local" || c.Annotations["mutating"] != "false" {
			t.Fatalf("annotations = %v, want source=local mutating=false", c.Annotations)
		}
	}
	if !found {
		t.Fatal("tossctl engine heartbeat-check is not registered")
	}
}

// TestA134HeartbeatCheckFailsWhenItCannotAlert: 알림 블록 없음(= OFF) → 이유를 말하는 오류 · 상태 파일 없음. 엔진이 살아 있어도 마찬가지.
func TestA134HeartbeatCheckFailsWhenItCannotAlert(t *testing.T) {
	dir := t.TempDir()
	at := time.Now()
	if err := os.WriteFile(enginelock.MarkerPath(dir), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(enginelock.MarkerPath(dir), at, at)
	a134Config(t, dir, `{"engine":{"autostart":true}}`)
	_, _, err := a134RunCheck(t, dir)
	if err == nil {
		t.Fatal("a watcher that cannot alert reported success")
	}
	if !strings.Contains(err.Error(), "알림") {
		t.Fatalf("the refusal does not say why: %v", err)
	}
	if _, serr := os.Stat(filepath.Join(dir, enginewatch.StateFileName)); serr == nil {
		t.Fatal("a watcher that cannot alert wrote its state")
	}
}

// TestA134HeartbeatCheckAlertsOnceThroughNtfy: autostart ON + 마커 없음 → ntfy 에 critical(우선순위 5) 1건, 두 번째 실행은 0건,
// 마커가 살아나면 복구 1건. 요청 경로는 설정한 토픽뿐.
func TestA134HeartbeatCheckAlertsOnceThroughNtfy(t *testing.T) {
	ntfy := &a134Ntfy{}
	server := httptest.NewServer(http.HandlerFunc(ntfy.handler))
	defer server.Close()

	dir := t.TempDir()
	a134Config(t, dir, `{"engine":{"autostart":true,"notifications":{"enabled":true,"base_url":"`+
		server.URL+`","topic":"a134-watch-topic"}}}`)

	out, _, err := a134RunCheck(t, dir)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	if ntfy.count() != 1 {
		t.Fatalf("first run sent %d requests, want 1", ntfy.count())
	}
	first := ntfy.requests[0]
	if first.path != "/a134-watch-topic" || first.priority != "5" ||
		!strings.HasPrefix(first.tags, string(enginewatch.EventLivenessLost)) {
		t.Fatalf("lost alert = %+v, want topic path, priority 5, %s tag", first, enginewatch.EventLivenessLost)
	}
	if !strings.Contains(out, "down") {
		t.Fatalf("the command did not report the verdict: %q", out)
	}

	if _, _, err := a134RunCheck(t, dir); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if ntfy.count() != 1 {
		t.Fatalf("the same outage was alerted again: %d requests", ntfy.count())
	}

	now := time.Now()
	if err := os.WriteFile(enginelock.MarkerPath(dir), []byte(`{"ready_at":"`+now.UTC().Format(time.RFC3339)+`"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(enginelock.MarkerPath(dir), now, now)
	out, _, err = a134RunCheck(t, dir)
	if err != nil {
		t.Fatalf("recovery run: %v", err)
	}
	if ntfy.count() != 2 || !strings.HasPrefix(ntfy.requests[1].tags, string(enginewatch.EventLivenessRestored)) {
		t.Fatalf("recovery sent %d requests (%+v), want the restored alert", ntfy.count(), ntfy.requests)
	}
	if !strings.Contains(out, "up") {
		t.Fatalf("the command did not report the recovery: %q", out)
	}
}

// TestA134HeartbeatCheckRefusesATooShortThreshold: --stale-after ≤ 마커 갱신 주기는 거절(D2).
func TestA134HeartbeatCheckRefusesATooShortThreshold(t *testing.T) {
	ntfy := &a134Ntfy{}
	server := httptest.NewServer(http.HandlerFunc(ntfy.handler))
	defer server.Close()
	dir := t.TempDir()
	a134Config(t, dir, `{"engine":{"autostart":true,"notifications":{"enabled":true,"base_url":"`+
		server.URL+`","topic":"a134-watch-topic"}}}`)
	if _, _, err := a134RunCheck(t, dir, "--stale-after", "30s"); err == nil {
		t.Fatal("a threshold shorter than the marker refresh was accepted")
	}
	if ntfy.count() != 0 {
		t.Fatalf("a refused run sent %d requests", ntfy.count())
	}
}

// TestA134HeartbeatCheckWithAutostartOffLeavesAStoppedEngineAlone: autostart OFF + 마커 없음 → idle · 요청 0 (cmd 가 설정의 autostart 를 넘김).
func TestA134HeartbeatCheckWithAutostartOffLeavesAStoppedEngineAlone(t *testing.T) {
	ntfy := &a134Ntfy{}
	server := httptest.NewServer(http.HandlerFunc(ntfy.handler))
	defer server.Close()
	dir := t.TempDir()
	a134Config(t, dir, `{"engine":{"autostart":false,"notifications":{"enabled":true,"base_url":"`+
		server.URL+`","topic":"a134-watch-topic"}}}`)
	out, _, err := a134RunCheck(t, dir)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if ntfy.count() != 0 || !strings.Contains(out, "idle") {
		t.Fatalf("autostart off sent %d requests, output %q; want idle and none", ntfy.count(), out)
	}
}

// TestA134HeartbeatCheckSendTestReachesTheChannelAndNamesFailure: --send-test 는 판정 없이 한 건 보내고, 서버가 거절하면 실패로 끝남 · 상태 파일 무변경.
func TestA134HeartbeatCheckSendTestReachesTheChannelAndNamesFailure(t *testing.T) {
	ntfy := &a134Ntfy{}
	server := httptest.NewServer(http.HandlerFunc(ntfy.handler))
	defer server.Close()
	dir := t.TempDir()
	a134Config(t, dir, `{"engine":{"autostart":true,"notifications":{"enabled":true,"base_url":"`+
		server.URL+`","topic":"a134-watch-topic"}}}`)
	if _, _, err := a134RunCheck(t, dir, "--send-test"); err != nil {
		t.Fatalf("send-test: %v", err)
	}
	if ntfy.count() != 1 || ntfy.requests[0].priority != "3" ||
		!strings.HasPrefix(ntfy.requests[0].tags, string(enginewatch.EventWatchTest)) {
		t.Fatalf("send-test sent %+v, want one normal test notification", ntfy.requests)
	}
	if _, serr := os.Stat(filepath.Join(dir, enginewatch.StateFileName)); serr == nil {
		t.Fatal("send-test judged the engine and wrote state")
	}

	refusing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer refusing.Close()
	a134Config(t, dir, `{"engine":{"autostart":true,"notifications":{"enabled":true,"base_url":"`+
		refusing.URL+`","topic":"a134-watch-topic"}}}`)
	if _, _, err := a134RunCheck(t, dir, "--send-test"); err == nil {
		t.Fatal("an undelivered send-test reported success")
	}
}

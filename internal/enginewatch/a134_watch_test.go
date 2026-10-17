package enginewatch_test

// a134_watch_test.go — 엔진 밖 감시자의 판정표(design D2) · 한 번만 알리기(D3) · 전송(D4) · 계좌 칸 없음(D5)을 잼.
// 실 ntfy 로 보내지 않음: Publisher 는 시험이 주입한 기록기. 저널 디렉터리는 t.TempDir() 뿐.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginewatch"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

var a134Now = time.Date(2026, 10, 14, 3, 0, 0, 0, time.UTC)

// a134Publisher 는 받은 알림을 기록하고, fail 이 있으면 그 오류로 전송 실패를 흉내 냄.
type a134Publisher struct {
	sent []obs.Notification
	fail error
}

func (p *a134Publisher) Publish(_ context.Context, n obs.Notification) error {
	if p.fail != nil {
		return p.fail
	}
	p.sent = append(p.sent, n)
	return nil
}

// a134Marker 는 엔진이 쓰는 그 경로에 복구를 마친(ready_at) 마커를 두고 수정 시각을 at 으로 맞춤(엔진 write 와 같은 사실 = mtime).
func a134Marker(t *testing.T, dir string, at time.Time) string {
	t.Helper()
	return a134MarkerBody(t, dir, at, `{"pid":1,"ready_at":"`+at.UTC().Format(time.RFC3339)+`"}`)
}

// a134MarkerBody 는 본문을 지정해 마커를 둠(복구 전 마커 = ready_at 없음).
func a134MarkerBody(t *testing.T, dir string, at time.Time, body string) string {
	t.Helper()
	path := enginelock.MarkerPath(dir)
	if err := os.WriteFile(path, []byte(body+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatal(err)
	}
	return path
}

func a134Check(dir string, pub obs.Publisher, expected bool, now time.Time) enginewatch.Check {
	return enginewatch.Check{
		Dir:       dir,
		Expected:  expected,
		Publisher: pub,
		Clock:     clock.NewFake(now),
	}
}

func a134State(t *testing.T, dir string) (enginewatch.State, bool) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(dir, enginewatch.StateFileName))
	if errors.Is(err, os.ErrNotExist) {
		return enginewatch.State{}, false
	}
	if err != nil {
		t.Fatal(err)
	}
	var s enginewatch.State
	if err := json.Unmarshal(body, &s); err != nil {
		t.Fatalf("state file is not JSON: %v (%q)", err, body)
	}
	return s, true
}

// TestA134TheVerdictTable 는 design D2 의 표를 한 행씩 잼.
func TestA134TheVerdictTable(t *testing.T) {
	for _, tc := range []struct {
		name     string
		beatAge  *time.Duration // nil = 마커 없음
		expected bool
		want     enginewatch.Verdict
		reason   string
	}{
		{"fresh", a134Dur(30 * time.Second), false, enginewatch.VerdictUp, ""},
		{"fresh at the threshold", a134Dur(enginewatch.DefaultStaleAfter), true, enginewatch.VerdictUp, ""},
		{"stale with autostart off", a134Dur(4 * time.Minute), false, enginewatch.VerdictDown, enginewatch.ReasonStale},
		{"stale with autostart on", a134Dur(4 * time.Minute), true, enginewatch.VerdictDown, enginewatch.ReasonStale},
		{"future timestamp", a134Dur(-2 * time.Minute), true, enginewatch.VerdictUp, ""},
		{"missing while expected", nil, true, enginewatch.VerdictDown, enginewatch.ReasonMissing},
		{"missing while not expected", nil, false, enginewatch.VerdictIdle, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.beatAge != nil {
				a134Marker(t, dir, a134Now.Add(-*tc.beatAge))
			}
			got, err := enginewatch.Observe(enginelock.MarkerPath(dir), a134Now,
				enginewatch.DefaultStaleAfter, tc.expected)
			if err != nil {
				t.Fatalf("Observe: %v", err)
			}
			if got.Verdict != tc.want || got.Reason != tc.reason {
				t.Fatalf("verdict = %q/%q, want %q/%q", got.Verdict, got.Reason, tc.want, tc.reason)
			}
		})
	}
}

func a134Dur(d time.Duration) *time.Duration { return &d }

// TestA134AnUnreadableMarkerIsNotAbsence: 부재 외의 stat 실패는 「없음」으로 접히지 않고 오류가 됨(D2 마지막 행).
// 부모 경로가 일반 파일이면 stat 은 ENOTDIR — ErrNotExist 가 아님.
func TestA134AnUnreadableMarkerIsNotAbsence(t *testing.T) {
	dir := t.TempDir()
	notADir := filepath.Join(dir, "file")
	if err := os.WriteFile(notADir, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := enginewatch.Observe(filepath.Join(notADir, enginelock.MarkerFileName), a134Now,
		enginewatch.DefaultStaleAfter, true)
	if err == nil {
		t.Fatal("an unreadable marker was judged; it must be an error, not an absence")
	}
}

// TestA134AThresholdNotLongerThanTheRefreshIsRefused: 기준 ≤ 마커 갱신 주기면 살아 있는 엔진을 죽었다고 읽음(D2).
func TestA134AThresholdNotLongerThanTheRefreshIsRefused(t *testing.T) {
	dir := t.TempDir()
	a134Marker(t, dir, a134Now)
	pub := &a134Publisher{}
	check := a134Check(dir, pub, true, a134Now)
	check.StaleAfter = enginelock.RefreshEvery
	if _, err := check.Run(context.Background()); !errors.Is(err, enginewatch.ErrStaleAfterTooShort) {
		t.Fatalf("err = %v, want ErrStaleAfterTooShort", err)
	}
	if len(pub.sent) != 0 {
		t.Fatalf("a refused threshold sent %d alerts", len(pub.sent))
	}
}

// TestA134OneAlertPerOutageAndOnePerRecovery 는 D3 의 상태 이동 전부를 한 줄기로 잼.
func TestA134OneAlertPerOutageAndOnePerRecovery(t *testing.T) {
	dir := t.TempDir()
	pub := &a134Publisher{}
	ctx := context.Background()

	// 1) 엔진이 죽어 마커가 4분 묵음 → critical 1건, 상태 down.
	a134Marker(t, dir, a134Now.Add(-4*time.Minute))
	out, err := a134Check(dir, pub, true, a134Now).Run(ctx)
	if err != nil {
		t.Fatalf("run 1: %v", err)
	}
	if len(pub.sent) != 1 || out.Sent != enginewatch.EventLivenessLost {
		t.Fatalf("run 1 sent %d (%q), want one lost alert", len(pub.sent), out.Sent)
	}
	lost := pub.sent[0]
	if lost.Type != enginewatch.EventLivenessLost || lost.Severity != obs.SeverityCritical {
		t.Fatalf("lost alert = %q/%q, want %q/critical", lost.Type, lost.Severity, enginewatch.EventLivenessLost)
	}
	if strings.TrimSpace(lost.Title) == "" || strings.TrimSpace(lost.Body) == "" {
		t.Fatalf("lost alert has no title or body: %+v", lost)
	}
	if s, ok := a134State(t, dir); !ok || !s.Down || s.Reason != enginewatch.ReasonStale {
		t.Fatalf("state after run 1 = %+v (exists %v), want down/stale", s, ok)
	}

	// 2) 1분 뒤에도 그대로 → 알림 없음.
	if _, err := a134Check(dir, pub, true, a134Now.Add(time.Minute)).Run(ctx); err != nil {
		t.Fatalf("run 2: %v", err)
	}
	if len(pub.sent) != 1 {
		t.Fatalf("the same outage alerted again: %d alerts", len(pub.sent))
	}

	// 3) 엔진이 다시 살아 마커 갱신 → 복구 1건(normal), 상태 해제.
	later := a134Now.Add(10 * time.Minute)
	a134Marker(t, dir, later.Add(-20*time.Second))
	out, err = a134Check(dir, pub, true, later).Run(ctx)
	if err != nil {
		t.Fatalf("run 3: %v", err)
	}
	if len(pub.sent) != 2 || out.Sent != enginewatch.EventLivenessRestored {
		t.Fatalf("run 3 sent %d (%q), want the restored alert", len(pub.sent), out.Sent)
	}
	restored := pub.sent[1]
	if restored.Type != enginewatch.EventLivenessRestored || restored.Severity != obs.SeverityNormal {
		t.Fatalf("restored alert = %q/%q, want %q/normal", restored.Type, restored.Severity,
			enginewatch.EventLivenessRestored)
	}
	if s, _ := a134State(t, dir); s.Down {
		t.Fatalf("state after recovery is still down: %+v", s)
	}

	// 4) 계속 살아 있음 → 알림 없음.
	if _, err := a134Check(dir, pub, true, later.Add(30*time.Second)).Run(ctx); err != nil {
		t.Fatalf("run 4: %v", err)
	}
	if len(pub.sent) != 2 {
		t.Fatalf("an up engine alerted: %d alerts", len(pub.sent))
	}
}

// TestA134AMissingEngineThatShouldRunAlerts: autostart ON + 마커 없음 → critical 1건(reason missing).
func TestA134AMissingEngineThatShouldRunAlerts(t *testing.T) {
	dir := t.TempDir()
	pub := &a134Publisher{}
	if _, err := a134Check(dir, pub, true, a134Now).Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(pub.sent) != 1 || pub.sent[0].Type != enginewatch.EventLivenessLost {
		t.Fatalf("sent %+v, want one lost alert", pub.sent)
	}
	if s, _ := a134State(t, dir); !s.Down || s.Reason != enginewatch.ReasonMissing {
		t.Fatalf("state = %+v, want down/missing", s)
	}
}

// TestA134AStoppedEngineWithoutAutostartIsQuiet: 정상 정지 · 수동 운용(autostart OFF, 마커 없음) → 알림 0, 상태 파일 0.
func TestA134AStoppedEngineWithoutAutostartIsQuiet(t *testing.T) {
	dir := t.TempDir()
	pub := &a134Publisher{}
	out, err := a134Check(dir, pub, false, a134Now).Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.Observation.Verdict != enginewatch.VerdictIdle || len(pub.sent) != 0 {
		t.Fatalf("verdict %q with %d alerts, want idle and none", out.Observation.Verdict, len(pub.sent))
	}
	if _, ok := a134State(t, dir); ok {
		t.Fatal("an idle check wrote a state file")
	}
}

// TestA134AFailedSendDoesNotAdvanceTheState: 전송 실패 → 오류 · 상태 무변경 → 다음 실행이 다시 보냄(D3 첫 줄).
func TestA134AFailedSendDoesNotAdvanceTheState(t *testing.T) {
	dir := t.TempDir()
	a134Marker(t, dir, a134Now.Add(-10*time.Minute))
	broken := &a134Publisher{fail: errors.New("ntfy unreachable")}
	if _, err := a134Check(dir, broken, true, a134Now).Run(context.Background()); err == nil {
		t.Fatal("a failed send reported success")
	}
	if _, ok := a134State(t, dir); ok {
		t.Fatal("a failed send recorded the outage as alerted")
	}
	pub := &a134Publisher{}
	if _, err := a134Check(dir, pub, true, a134Now.Add(time.Minute)).Run(context.Background()); err != nil {
		t.Fatalf("retry: %v", err)
	}
	if len(pub.sent) != 1 || pub.sent[0].Type != enginewatch.EventLivenessLost {
		t.Fatalf("the retry sent %+v, want the lost alert", pub.sent)
	}

	// 복구 전송이 실패해도 같은 방향: down 이 유지되어 다음 실행이 복구를 다시 보냄.
	a134Marker(t, dir, a134Now.Add(2*time.Minute))
	if _, err := a134Check(dir, broken, true, a134Now.Add(2*time.Minute)).Run(context.Background()); err == nil {
		t.Fatal("a failed recovery send reported success")
	}
	if s, _ := a134State(t, dir); !s.Down {
		t.Fatal("a failed recovery send cleared the outage")
	}
}

// TestA134NoPublisherFailsLoudly: 알릴 수단이 없으면 판정과 무관하게 오류(엔진이 살아 있어도) — 조용한 성공 금지(D4).
func TestA134NoPublisherFailsLoudly(t *testing.T) {
	dir := t.TempDir()
	a134Marker(t, dir, a134Now)
	_, err := a134Check(dir, nil, true, a134Now).Run(context.Background())
	if !errors.Is(err, enginewatch.ErrNoPublisher) {
		t.Fatalf("err = %v, want ErrNoPublisher", err)
	}
	if _, ok := a134State(t, dir); ok {
		t.Fatal("a check that cannot alert wrote a state file")
	}
}

// TestA134ACorruptStateDoesNotSuppress: 읽을 수 없는 상태는 「down 아님」 — 억제 방향으로 읽히지 않음(D3).
func TestA134ACorruptStateDoesNotSuppress(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, enginewatch.StateFileName), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	a134Marker(t, dir, a134Now.Add(-10*time.Minute))
	pub := &a134Publisher{}
	if _, err := a134Check(dir, pub, true, a134Now).Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(pub.sent) != 1 {
		t.Fatalf("a corrupt state suppressed the alert: %d sent", len(pub.sent))
	}
}

// TestA134TheStateIsWrittenAtomicallyAndPrivately: 상태 파일 0600 · 임시 파일 잔재 0.
func TestA134TheStateIsWrittenAtomicallyAndPrivately(t *testing.T) {
	dir := t.TempDir()
	pub := &a134Publisher{}
	if _, err := a134Check(dir, pub, true, a134Now).Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, enginewatch.StateFileName))
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("state mode = %o, want 600", mode)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != enginewatch.StateFileName {
			t.Errorf("the watcher left %q in the journal directory", e.Name())
		}
	}
}

// TestA134TheWatcherNeedsNoEngine: 엔진 · 저널 없이 판정하고, 마커 · 상태 파일 외에는 아무것도 만들지 않음(엔진 잠금 · 저널 포함).
func TestA134TheWatcherNeedsNoEngine(t *testing.T) {
	dir := t.TempDir()
	a134Marker(t, dir, a134Now.Add(-5*time.Minute))
	pub := &a134Publisher{}
	if _, err := a134Check(dir, pub, true, a134Now).Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{enginelock.MarkerFileName: true, enginewatch.StateFileName: true}
	for _, e := range entries {
		if !allowed[e.Name()] {
			t.Errorf("the watcher created %q; it must not touch the engine's directory beyond its own state", e.Name())
		}
	}
	if _, err := os.Stat(filepath.Join(dir, enginelock.LockFileName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("the watcher took the engine lock: %v", err)
	}
}

// TestA134TheLivenessFileCarriesNoAccountField: 엔진 활성 마커는 이제 엔진 밖이 읽는 생존 신호 — 계좌 키 칸이 생기면 실패(D5).
func TestA134TheLivenessFileCarriesNoAccountField(t *testing.T) {
	keys := a134JSONKeys(reflect.TypeOf(enginelock.Marker{}))
	if len(keys) < 5 {
		t.Fatalf("only %d marker fields were read; the census is blind", len(keys))
	}
	for _, k := range keys {
		if redact.IsAccountKey(k) || strings.Contains(strings.ToLower(k), "account") {
			t.Errorf("the engine liveness marker carries an account field %q", k)
		}
	}
	for _, k := range a134JSONKeys(reflect.TypeOf(enginewatch.State{})) {
		if redact.IsAccountKey(k) || strings.Contains(strings.ToLower(k), "account") {
			t.Errorf("the watcher state carries an account field %q", k)
		}
	}
}

func a134JSONKeys(typ reflect.Type) []string {
	var out []string
	for i := 0; i < typ.NumField(); i++ {
		tag := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		if tag == "" || tag == "-" {
			tag = typ.Field(i).Name
		}
		out = append(out, tag)
	}
	return out
}

// TestA134TheAlertsSayWhatToDoAndCarryNoAccount: 두 알림 모두 판정 사실(시각 · 기준)과 할 일을 말하고 계좌 문구가 없음.
func TestA134TheAlertsSayWhatToDoAndCarryNoAccount(t *testing.T) {
	for _, reason := range []string{enginewatch.ReasonStale, enginewatch.ReasonMissing} {
		dir := t.TempDir()
		if reason == enginewatch.ReasonStale {
			a134Marker(t, dir, a134Now.Add(-7*time.Minute))
		}
		pub := &a134Publisher{}
		if _, err := a134Check(dir, pub, true, a134Now).Run(context.Background()); err != nil {
			t.Fatalf("%s: %v", reason, err)
		}
		a134Marker(t, dir, a134Now.Add(time.Minute))
		if _, err := a134Check(dir, pub, true, a134Now.Add(time.Minute)).Run(context.Background()); err != nil {
			t.Fatalf("%s recovery: %v", reason, err)
		}
		if len(pub.sent) != 2 {
			t.Fatalf("%s: %d alerts, want 2", reason, len(pub.sent))
		}
		if !strings.Contains(pub.sent[0].Body, enginelock.MarkerFileName) {
			t.Errorf("%s: the lost alert does not name the liveness file: %q", reason, pub.sent[0].Body)
		}
		if !strings.Contains(pub.sent[0].Body, "손절") {
			t.Errorf("%s: the lost alert does not say what is at stake: %q", reason, pub.sent[0].Body)
		}
		for _, n := range pub.sent {
			text := strings.ToLower(n.Title + "\n" + n.Body + "\n" + strings.Join(n.Tags, ","))
			if strings.Contains(text, "account") {
				t.Errorf("%s: an alert mentions an account: %q", reason, text)
			}
		}
	}
}

// TestA134TheEventTypesAreTheirOwn: 새 이벤트 값의 문자 고정 + obs/event.go 의 어떤 값과도 겹치지 않음(a102 유일성과 같은 방법 — 원문을 읽음).
func TestA134TheEventTypesAreTheirOwn(t *testing.T) {
	if enginewatch.EventLivenessLost != "engine.liveness_lost" ||
		enginewatch.EventLivenessRestored != "engine.liveness_restored" {
		t.Fatalf("event values moved: %q / %q", enginewatch.EventLivenessLost, enginewatch.EventLivenessRestored)
	}
	body, err := os.ReadFile(filepath.Join("..", "obs", "event.go"))
	if err != nil {
		t.Fatal(err)
	}
	decl := regexp.MustCompile(`(?m)^\s*(Event[A-Za-z0-9_]+)\s+EventType\s*=\s*"([^"]+)"`)
	matches := decl.FindAllStringSubmatch(string(body), -1)
	if len(matches) < 20 {
		t.Fatalf("only %d obs event constants were read; the census is blind", len(matches))
	}
	if enginewatch.EventWatchTest != "engine.liveness_watch_test" {
		t.Fatalf("test event value moved: %q", enginewatch.EventWatchTest)
	}
	for _, m := range matches {
		for _, mine := range []obs.EventType{enginewatch.EventLivenessLost, enginewatch.EventLivenessRestored,
			enginewatch.EventWatchTest} {
			if m[2] == string(mine) {
				t.Errorf("%s already publishes %q; the watcher's event would be indistinguishable", m[1], mine)
			}
		}
	}
	if obs.SeverityOf(enginewatch.EventLivenessLost) == obs.SeverityCritical {
		// 감시자는 엔진 Notifier 를 거치지 않으므로 등급표에 오를 이유가 없음 — 오르면 엔진 안 누군가가 이 값으로 outbox 를 쓸 수 있게 됨.
		t.Error("the watcher's event is in obs' critical table; it must stay outside the engine's outbox")
	}
}

// TestA134RecoveryNeedsAnEngineThatFinishedItsRecovery: 끊김 뒤 신선하지만 ready_at 없는 마커(재기동 루프 · Recover 중)는 복구가 아님 —
// 끊김을 유지하고 알림 0, ready_at 이 생긴 뒤에만 복구 1건(리뷰 P2-1).
func TestA134RecoveryNeedsAnEngineThatFinishedItsRecovery(t *testing.T) {
	dir := t.TempDir()
	pub := &a134Publisher{}
	ctx := context.Background()
	if _, err := a134Check(dir, pub, true, a134Now).Run(ctx); err != nil { // 마커 없음 → 끊김
		t.Fatalf("run 1: %v", err)
	}
	booting := a134Now.Add(2 * time.Minute)
	a134MarkerBody(t, dir, booting, `{"pid":7}`)
	out, err := a134Check(dir, pub, true, booting).Run(ctx)
	if err != nil {
		t.Fatalf("run 2: %v", err)
	}
	if len(pub.sent) != 1 || out.Observation.Verdict != enginewatch.VerdictUp || out.Observation.Ready {
		t.Fatalf("a booting engine (no ready_at) produced %d alerts / %+v; want the outage kept", len(pub.sent), out.Observation)
	}
	if s, _ := a134State(t, dir); !s.Down {
		t.Fatal("a booting engine cleared the outage")
	}
	// 해석할 수 없는 본문도 준비 안 됨.
	a134MarkerBody(t, dir, booting, `{not json`)
	if _, err := a134Check(dir, pub, true, booting).Run(ctx); err != nil || len(pub.sent) != 1 {
		t.Fatalf("an unparseable marker was read as recovered: err=%v sent=%d", err, len(pub.sent))
	}
	a134Marker(t, dir, booting.Add(time.Minute))
	if _, err := a134Check(dir, pub, true, booting.Add(time.Minute)).Run(ctx); err != nil {
		t.Fatalf("run 3: %v", err)
	}
	if len(pub.sent) != 2 || pub.sent[1].Type != enginewatch.EventLivenessRestored {
		t.Fatalf("a recovered engine sent %+v, want the restored alert", pub.sent)
	}
}

// TestA134RecoveryToIdleSaysTheWatchEnded: 끊김 뒤 autostart OFF · 마커 없음 → 복구 1건(감시 대상에서 벗어남) · 상태 해제.
func TestA134RecoveryToIdleSaysTheWatchEnded(t *testing.T) {
	dir := t.TempDir()
	pub := &a134Publisher{}
	ctx := context.Background()
	a134Marker(t, dir, a134Now.Add(-10*time.Minute))
	if _, err := a134Check(dir, pub, false, a134Now).Run(ctx); err != nil {
		t.Fatalf("run 1: %v", err)
	}
	if err := os.Remove(enginelock.MarkerPath(dir)); err != nil {
		t.Fatal(err)
	}
	out, err := a134Check(dir, pub, false, a134Now.Add(time.Minute)).Run(ctx)
	if err != nil {
		t.Fatalf("run 2: %v", err)
	}
	if out.Observation.Verdict != enginewatch.VerdictIdle || len(pub.sent) != 2 ||
		pub.sent[1].Type != enginewatch.EventLivenessRestored || !strings.Contains(pub.sent[1].Body, "감시 대상에서 벗어남") {
		t.Fatalf("recovery to idle = %q, %+v", out.Observation.Verdict, pub.sent)
	}
	if s, _ := a134State(t, dir); s.Down {
		t.Fatal("recovery to idle kept the outage")
	}
}

// TestA134TheRealMarkerLifecycleIsWhatTheWatcherReads: 엔진이 실제로 쓰는 enginelock.Hold/Ready/Release 를 그대로 읽음 —
// Hold 직후 = Up · 복구 전, Ready 뒤 = Up · 복구 끝, Release(정상 정지) 뒤 = 마커 없음(autostart ON 이면 끊김).
func TestA134TheRealMarkerLifecycleIsWhatTheWatcherReads(t *testing.T) {
	dir := t.TempDir()
	path := enginelock.MarkerPath(dir)
	now := time.Now()
	held, err := enginelock.Hold(context.Background(), path, now)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer held.Release()
	o, err := enginewatch.Observe(path, now, enginewatch.DefaultStaleAfter, true)
	if err != nil || o.Verdict != enginewatch.VerdictUp || o.Ready {
		t.Fatalf("after Hold: %+v err=%v, want up and not ready", o, err)
	}
	held.Ready(now)
	o, err = enginewatch.Observe(path, now, enginewatch.DefaultStaleAfter, true)
	if err != nil || o.Verdict != enginewatch.VerdictUp || !o.Ready {
		t.Fatalf("after Ready: %+v err=%v, want up and ready", o, err)
	}
	held.Release()
	o, err = enginewatch.Observe(path, now, enginewatch.DefaultStaleAfter, true)
	if err != nil || o.Verdict != enginewatch.VerdictDown || o.Reason != enginewatch.ReasonMissing {
		t.Fatalf("after Release: %+v err=%v, want down/missing", o, err)
	}
}

// TestA134TheSendTestProvesDelivery: --send-test 의 몸통 — 전달 실패는 오류, 수단 없음은 ErrNoPublisher, 성공은 normal 한 건.
func TestA134TheSendTestProvesDelivery(t *testing.T) {
	if err := enginewatch.SendTest(context.Background(), nil); !errors.Is(err, enginewatch.ErrNoPublisher) {
		t.Fatalf("nil publisher: %v", err)
	}
	if err := enginewatch.SendTest(context.Background(), &a134Publisher{fail: errors.New("401")}); err == nil {
		t.Fatal("an undelivered test reported success")
	}
	pub := &a134Publisher{}
	if err := enginewatch.SendTest(context.Background(), pub); err != nil {
		t.Fatal(err)
	}
	if len(pub.sent) != 1 || pub.sent[0].Type != enginewatch.EventWatchTest || pub.sent[0].Severity != obs.SeverityNormal {
		t.Fatalf("test alert = %+v", pub.sent)
	}
}

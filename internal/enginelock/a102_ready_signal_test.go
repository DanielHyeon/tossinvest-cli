package enginelock

// a102_ready_signal_test.go is written before the code it tests (openspec change
// a102, task 3.2).
//
// # 왜 내부 패키지 테스트인가
//
// 나머지 marker 테스트는 `enginelock_test`(외부 패키지)에 있다. 이 파일만 안쪽인 이유는
// **갱신 주기가 1분이기 때문**이다. `RefreshEvery`는 spec이 못박은 수치이고(runlock 선례),
// 테스트가 1분을 기다릴 수는 없다. 그래서 goroutine의 ticker 갈래가 부르는 쓰기를
// `(*Held).refresh(at)`라는 이름 있는 seam으로 두고, 테스트가 그것을 직접 부른다 —
// **ticker가 부르는 것과 정확히 같은 함수**다.
//
// 대안은 주기를 주입 가능한 노브로 만드는 것이었고, 그것은 spec이 고정한 수치를 테스트가
// 흔들 수 있게 만든다. 이름 있는 seam이 더 좁다.

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

var a102Now = time.Date(2026, 8, 13, 1, 35, 0, 0, time.UTC)

// TestReadyPublishesTheSignalOnlyWhenItIsCalled — 부재가 "준비 안 됨"이다.
// 마커가 생기는 시점(engine.go 6단계)은 복구(7단계)보다 **먼저**이므로, 마커의 존재
// 자체는 신호가 될 수 없다. spec: "그 신호는 엔진이 복구를 완료한 뒤에만 만들어져야 한다".
func TestReadyPublishesTheSignalOnlyWhenItIsCalled(t *testing.T) {
	path := MarkerPath(t.TempDir())

	held, err := Hold(context.Background(), path, a102Now)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer held.Release()

	if got := readMarkerFile(t, path); got.ReadyAt != nil {
		t.Fatalf("a fresh marker already carries ready_at (%s); the signal would fire before recovery",
			got.ReadyAt)
	}
	if status := Read(path, a102Now.Add(time.Second)); status.Ready() {
		t.Fatal("Read reported an engine as ready before Ready was ever called")
	}

	readyAt := a102Now.Add(50 * time.Second) // 실측 복구 소요와 같은 자리
	held.Ready(readyAt)

	got := readMarkerFile(t, path)
	if got.ReadyAt == nil {
		t.Fatal("Ready did not publish ready_at into the marker file")
	}
	if !got.ReadyAt.Equal(readyAt) {
		t.Errorf("ready_at = %s, want %s", got.ReadyAt, readyAt)
	}
	if status := Read(path, readyAt.Add(time.Second)); !status.Ready() {
		t.Error("Read does not expose the ready signal a held marker just published")
	}
}

// TestTheJSONFieldIsReadyAt. 콘솔과 서베이가 같은 파일을 읽는다 — 이름이 흔들리면
// 관대한 reader가 조용히 nil을 돌려주고 대기가 영원히 상한까지 간다.
func TestTheJSONFieldIsReadyAt(t *testing.T) {
	path := MarkerPath(t.TempDir())
	held, err := Hold(context.Background(), path, a102Now)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer held.Release()
	held.Ready(a102Now.Add(time.Minute))

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the marker: %v", err)
	}
	if !strings.Contains(string(body), `"ready_at"`) {
		t.Fatalf("the marker does not carry a ready_at field:\n%s", body)
	}
}

// TestRefreshPreservesTheReadySignal is design 검증계약 (d).
//
// 갱신은 1분마다 마커를 통째로 다시 쓴다. 캡처한 값을 다시 쓰면 그 값에는 ready_at이
// 없으므로, **준비를 알린 엔진이 1분 뒤 준비 안 한 엔진으로 돌아간다.** 콘솔은 그때부터
// 상한까지 기다린다.
func TestRefreshPreservesTheReadySignal(t *testing.T) {
	path := MarkerPath(t.TempDir())
	held, err := Hold(context.Background(), path, a102Now)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer held.Release()

	readyAt := a102Now.Add(50 * time.Second)
	held.Ready(readyAt)

	// ticker 갈래가 부르는 것과 같은 함수다.
	for i := 1; i <= 3; i++ {
		at := a102Now.Add(time.Duration(i) * RefreshEvery)
		if err := held.refresh(at); err != nil {
			t.Fatalf("refresh %d: %v", i, err)
		}
		got := readMarkerFile(t, path)
		if got.ReadyAt == nil {
			t.Fatalf("refresh %d erased ready_at; a ready engine became unready after %s",
				i, time.Duration(i)*RefreshEvery)
		}
		if !got.ReadyAt.Equal(readyAt) {
			t.Errorf("refresh %d moved ready_at to %s, want the original %s", i, got.ReadyAt, readyAt)
		}
		if !got.RefreshedAt.Equal(at) {
			t.Errorf("refresh %d wrote refreshed_at %s, want %s", i, got.RefreshedAt, at)
		}
		if status := Read(path, at.Add(time.Second)); !status.Ready() {
			t.Fatalf("after refresh %d the reader no longer sees a ready engine", i)
		}
	}
}

// TestReadyIsIdempotentAndKeepsTheFirstInstant. 복구가 한 번 끝난 순간이 신호의 의미다.
// 두 번째 호출이 그것을 뒤로 밀면 "언제 보호가 섰는가"가 마지막 호출 시각이 된다.
func TestReadyIsIdempotentAndKeepsTheFirstInstant(t *testing.T) {
	path := MarkerPath(t.TempDir())
	held, err := Hold(context.Background(), path, a102Now)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer held.Release()

	first := a102Now.Add(50 * time.Second)
	held.Ready(first)
	held.Ready(first.Add(10 * time.Minute))
	held.Ready(first.Add(time.Hour))

	got := readMarkerFile(t, path)
	if got.ReadyAt == nil || !got.ReadyAt.Equal(first) {
		t.Fatalf("ready_at = %v, want the first instant %s", got.ReadyAt, first)
	}
}

// TestAMarkerWithoutReadyAtStaysReadable is the lenient-reader rule the package
// comment already states: 없는 필드는 nil이고, 그것이 "준비 안 됨"이다. 예전 빌드가 쓴
// 마커를 읽고 오류를 내면 대시보드가 그리지 못한다.
func TestAMarkerWithoutReadyAtStaysReadable(t *testing.T) {
	path := MarkerPath(t.TempDir())
	body := `{"pid":4242,"started_at":"2026-08-13T01:35:00Z","refreshed_at":"2026-08-13T01:35:00Z"}`
	if err := os.WriteFile(path, []byte(body+"\n"), 0o600); err != nil {
		t.Fatalf("writing the marker: %v", err)
	}
	if err := os.Chtimes(path, a102Now, a102Now); err != nil {
		t.Fatalf("timestamping: %v", err)
	}

	status := Read(path, a102Now.Add(time.Second))
	if !status.Running {
		t.Fatal("a marker from a build that predates ready_at read as a stopped engine")
	}
	if status.Marker.ReadyAt != nil {
		t.Errorf("ready_at = %s, want nil", status.Marker.ReadyAt)
	}
	if status.Ready() {
		t.Error("an engine with no ready signal read as ready")
	}
	if status.Marker.PID != 4242 {
		t.Errorf("pid = %d, want 4242 — the rest of the marker must still parse", status.Marker.PID)
	}
}

// TestReadyRequiresBothALiveMarkerAndASignal pins the predicate itself, over the
// four combinations, and it exists because a mutation survived the file-driven
// tests.
//
// `Read` returns before it parses the body when the file is stale
// (enginelock.go: `if !fresh { return s }`), so a Status that came from a stale
// file has a zero Marker and its ready_at is nil anyway — deleting the Running
// half of Ready changes nothing on that path. But Status is a value, and
// `awaitEngineReady` takes `func() enginelock.Status`: anything can hand one over.
// The claim is about the predicate, so the test is about the predicate.
func TestReadyRequiresBothALiveMarkerAndASignal(t *testing.T) {
	stamp := a102Now
	for _, tc := range []struct {
		name   string
		status Status
		want   bool
	}{
		{"살아 있고 신호 있음", Status{Running: true, Marker: Marker{ReadyAt: &stamp}}, true},
		{"살아 있으나 신호 없음", Status{Running: true}, false},
		{"신호는 있으나 죽음", Status{Marker: Marker{ReadyAt: &stamp}}, false},
		{"둘 다 없음", Status{}, false},
	} {
		if got := tc.status.Ready(); got != tc.want {
			t.Errorf("%s: Ready() = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestAStaleMarkerIsNeverReady. 신선하지 않은 마커의 ready_at은 죽은 엔진의 것이다.
// 그것을 준비로 읽으면 서베이가 "엔진이 살아 있고 준비됐다"고 판단해 바로 시작한다 —
// 그 판단 자체는 오늘의 동작(엔진 없음)과 결과가 같지만, 이유가 거짓이 된다.
func TestAStaleMarkerIsNeverReady(t *testing.T) {
	path := MarkerPath(t.TempDir())
	held, err := Hold(context.Background(), path, a102Now)
	if err != nil {
		t.Fatalf("Hold: %v", err)
	}
	defer held.Release()
	held.Ready(a102Now)

	if status := Read(path, a102Now.Add(StaleAfter+time.Second)); status.Ready() {
		t.Fatal("a stale marker still reported a ready engine")
	}
}

// TestAFailedHoldPublishesNothing. Hold가 쓰기에 실패하면 release는 no-op이었다
// (편집 전 `return func(){}, werr`). 핸들이 그 의미를 잃으면 실패한 Hold가 **남의 마커를
// 지우거나** 준비 신호를 쓸 수 있다. engine.go의 B10은 이 실패에서 return하지 않는다.
func TestAFailedHoldPublishesNothing(t *testing.T) {
	dir := t.TempDir()
	// 디렉터리 자리에 파일을 두면 marker 경로의 MkdirAll이 실패한다.
	blocked := dir + "/blocked"
	if err := os.WriteFile(blocked, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("preparing the blocked path: %v", err)
	}
	path := MarkerPath(blocked)

	held, err := Hold(context.Background(), path, a102Now)
	if err == nil {
		t.Fatal("Hold reported success on a path it cannot write")
	}
	if held == nil {
		t.Fatal("a failed Hold returned a nil handle; the caller defers Release before reading err")
	}
	held.Ready(a102Now) // must not panic and must not create anything
	held.Release()
	held.Release()

	if _, statErr := os.Stat(path); statErr == nil {
		t.Fatal("a failed Hold created a marker file after all")
	}
}

func readMarkerFile(t *testing.T, path string) Marker {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the marker: %v", err)
	}
	var m Marker
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("parsing the marker: %v\n%s", err, body)
	}
	return m
}

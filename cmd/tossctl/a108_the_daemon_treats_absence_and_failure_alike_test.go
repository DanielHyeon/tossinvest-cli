//go:build unix

package main

// a108_the_daemon_treats_absence_and_failure_alike_test.go — 겹3 (tasks 3.3).
//
// httpapi 는 이미 강등을 안다: descriptor 가 **없으면** 전략 화면 없이 뜬다. 그런데
// descriptor 가 **있는데 dial 이 실패하면** fatal 이다 (httpapi.go B9). 2026-08-13 의
// 반쪽 잔재 — endpoint.json 만 남고 runtime.sock 은 사라진 상태 — 는 정확히 후자를
// 때렸고, 컨테이너는 `Restarting (1)` 로 돌았다.
//
// 이 비대칭이 결함이다. 설계된 강등과 사고 사이의 차이가 「descriptor 파일이 하나
// 남아 있는가」일 이유가 없다. fatal 로 남을 것은 descriptor 를 **조사할 수조차 없는
// 경우**뿐이다 (권한 등 비-NotExist stat 오류) — 그것은 잔재가 아니라 환경 이상이다.
//
// 여기서는 seam 을 쓰지 않는다. 겹2 와 달리 실패를 만드는 것이 T1 이 편집 중인
// 회수 규칙(`reclaimStaleControlDirectory`)이 아니라 `Dial` 자신의 socket 검사이고,
// 그 검사는 이 change 가 건드리지 않는다.

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojectionrpc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

// a108ServeWindow 는 강등 기동이 「실제로 떴다」를 보이기 위해 서비스를 살려 두는 시간이다.
//
// 강등하면 `runHTTPAPI` 는 listener 를 열고 ctx 가 끝날 때까지 돌아오지 않는다. 그래서
// 이 창이 종료 조건이고, 창이 만료돼 nil 로 돌아온 것 자체가 「떴다」의 증거다. 실패
// 경로는 listener 앞에서 돌아오므로 이 창을 쓰지 않는다.
const a108ServeWindow = 3 * time.Second

// a108FreePort 는 커널이 고르게 한다. 고정 포트는 개발 호스트에서 돌고 있는 진짜
// httpapi 데몬(37086)과 부딪힌다.
func a108FreePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("빈 포트를 찾지 못했다: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("probe listener close: %v", err)
	}
	return port
}

// a108HTTPAPIDir 는 daemon 이 읽을 엔진 디렉터리다.
func a108HTTPAPIDir(t *testing.T) string {
	t.Helper()
	testenv.Isolate(t)
	dir, err := os.MkdirTemp("/tmp", "a108-httpapi-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

// a108HalfLeftover 는 사고 당일의 디스크 상태다: control 디렉터리와 descriptor 는
// 있고 socket 은 없다.
//
// graceful shutdown 이 socket 을 unlink 하고(Go unix listener 기본값) `Close` 의 제거
// 루프가 끝나기 전에 호스트 종료가 프로세스를 죽이면 정확히 이 모양이 남는다.
func a108HalfLeftover(t *testing.T, dir string) {
	t.Helper()
	control := strategyprojectionrpc.ControlDirectory(dir)
	if err := os.Mkdir(control, 0o700); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(strategyprojectionrpc.Descriptor{
		SchemaVersion: "tossos.strategy-runtime-unix/v1",
		Socket:        strategyprojectionrpc.SocketFileName,
		Token:         strings.Repeat("a", 64),
		PID:           16,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(strategyprojectionrpc.DescriptorPath(dir), body, 0o600); err != nil {
		t.Fatal(err)
	}
	// socket 은 만들지 않는다 — 그것이 이 fixture 의 전부다.
	if _, err := os.Lstat(strategyprojectionrpc.SocketPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("fixture 가 socket 을 남겼다 (stat err = %v)", err)
	}
}

// a108RunHTTPAPI 은 데몬을 한 번 돌린다.
//
// TLS 는 forwarded 모드다 — 직접 종단을 쓰면 PEM 두 개를 만들어야 하고, 이 테스트가
// 재는 것은 boundary 가 아니라 **descriptor 분기**다.
func a108RunHTTPAPI(t *testing.T, dir string) (string, string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), a108ServeWindow)
	defer cancel()
	port := a108FreePort(t)
	cmd := newHTTPAPICmd(&rootOptions{configDir: dir})
	var out, errOut strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{
		"--port", strconv.Itoa(port),
		"--bind", "127.0.0.1",
		"--allowed-cidr", "127.0.0.0/8",
		"--public-url", fmt.Sprintf("https://127.0.0.1:%d", port),
		"--trusted-proxy", "127.0.0.2",
		"--tls-forwarded",
	})
	err := cmd.ExecuteContext(ctx)
	return out.String(), errOut.String(), err
}

// TestADeadDescriptorDoesNotStopTheDaemon 은 crash loop 그 자체다.
func TestADeadDescriptorDoesNotStopTheDaemon(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	a108HalfLeftover(t, dir)

	out, errOut, err := a108RunHTTPAPI(t, dir)
	if err != nil {
		t.Fatalf("httpapi = %v — 반쪽 잔재 하나가 조회 전용 데몬을 재시작 루프에 넣었다\n%s", err, errOut)
	}
	if !strings.Contains(out, "httpapi private endpoint") {
		t.Errorf("데몬이 endpoint 를 알리지 않았다 — 뜬 것이 맞는지 알 수 없다:\n%s", out)
	}
	if !strings.Contains(errOut, "strategy runtime projection") {
		t.Errorf("강등이 사유를 말하지 않았다; 전략 화면이 왜 비었는지 알 수 없다:\n%s", errOut)
	}
}

// TestAnAbsentDescriptorAndADeadOneBootTheSame 은 3.3 의 대칭 요구다.
//
// 설계된 강등(부재)과 사고(죽은 descriptor)가 **관측적으로 같은 기동**이어야 한다.
// 부재 쪽을 대조군으로 같이 돌리지 않으면 「둘 다 실패한다」도 통과한다.
func TestAnAbsentDescriptorAndADeadOneBootTheSame(t *testing.T) {
	absentOut, _, absentErr := a108RunHTTPAPI(t, a108HTTPAPIDir(t))
	if absentErr != nil {
		t.Fatalf("descriptor 부재 기동이 실패했다: %v", absentErr)
	}

	dead := a108HTTPAPIDir(t)
	a108HalfLeftover(t, dead)
	deadOut, _, deadErr := a108RunHTTPAPI(t, dead)
	if deadErr != nil {
		t.Fatalf("죽은 descriptor 기동이 실패했다: %v", deadErr)
	}
	if strings.Contains(absentOut, "httpapi private endpoint") !=
		strings.Contains(deadOut, "httpapi private endpoint") {
		t.Errorf("두 경로의 기동 결과가 다르다:\n부재: %s\n죽음: %s", absentOut, deadOut)
	}
	// descriptor 가 있던 쪽은 그 파일을 지우지 않는다. 잔재 정리는 엔진(겹1)의 일이고,
	// 조회 전용 데몬이 남의 control 디렉터리를 치우기 시작하면 그것은 새 사고 유형이다.
	if _, err := os.Stat(strategyprojectionrpc.DescriptorPath(dead)); err != nil {
		t.Errorf("데몬이 descriptor 를 치웠다 (stat err = %v) — 조회 전용이 아니다", err)
	}
}

// TestAnUninspectableDescriptorIsStillFatal 은 강등의 상한이다.
//
// descriptor 가 없는 것과 **조사할 수 없는 것**은 다르다. 후자는 잔재가 아니라 환경
// 이상이고, 강등으로 삼키면 잘못 배치된 데몬이 조용히 반쪽으로 뜬다.
//
// control 디렉터리 자리에 **일반 파일**을 두면 그 아래 경로의 `os.Stat` 은 ENOTDIR 로
// 실패한다 — NotExist 가 아니다. root 로 돌려도 결과가 같은 형태라 권한 fixture 보다
// 안정적이다.
func TestAnUninspectableDescriptorIsStillFatal(t *testing.T) {
	dir := a108HTTPAPIDir(t)
	if err := os.WriteFile(strategyprojectionrpc.ControlDirectory(dir), []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	// fixture 가 실제로 비-NotExist 오류를 만드는지 먼저 확인한다. 이것을 안 보면
	// 「그냥 없어서 통과」한 테스트를 fatal 의 증거로 쓰게 된다.
	if _, statErr := os.Stat(strategyprojectionrpc.DescriptorPath(dir)); statErr == nil || os.IsNotExist(statErr) {
		t.Fatalf("fixture 가 조사 불가능 상태를 못 만들었다: %v", statErr)
	}

	_, _, err := a108RunHTTPAPI(t, dir)
	if err == nil {
		t.Fatal("조사할 수 없는 descriptor 로 데몬이 떴다 — 강등이 환경 이상까지 삼켰다")
	}
	if !strings.Contains(err.Error(), "inspect strategy runtime projection") {
		t.Errorf("거절이 사유를 안 말한다: %v", err)
	}
}

// TestTheDaemonReadsTheDescriptorBesideItsOwnJournal 은 경로 계약이다.
//
// `--config-dir` 격리 프로파일이 진짜 엔진의 소켓을 읽으면 두 프로파일이 섞인다.
func TestTheDaemonReadsTheDescriptorBesideItsOwnJournal(t *testing.T) {
	dir := t.TempDir()
	got, err := engineJournalDir(&rootOptions{configDir: dir})
	if err != nil {
		t.Fatalf("engineJournalDir: %v", err)
	}
	if want := filepath.Join(dir, strategyprojectionrpc.ControlDirectoryName,
		strategyprojectionrpc.DescriptorFileName); strategyprojectionrpc.DescriptorPath(got) != want {
		t.Errorf("descriptor 경로 = %s, want %s", strategyprojectionrpc.DescriptorPath(got), want)
	}
}

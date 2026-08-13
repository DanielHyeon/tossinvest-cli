//go:build unix

package strategyprojectionrpc

// a108 Fix 라운드 — **발행도 회수만큼 전체적이어야 한다**(design D1-2).
//
// 겹1이 회수의 커버리지를 넓혔지만, A1 적대 리뷰가 남은 영구 거부 2종을 실측했다.
// 둘 다 **생산자**가 만든다 — 회수가 못 치우는 것이 아니라 발행이 원자적이지 않아서
// 생기는 모양이다.
//
//	pre-chmod socket   net.Listen이 만드는 권한은 umask가 정한다(컨테이너 실측 077 →
//	                   0700). listen과 chmod 사이에서 죽으면 최종 이름 자리에 0600이
//	                   아닌 socket이 남고, 회수의 정확-0600 검사가 영구 거부한다.
//	반쯤 쓰인 descriptor  O_EXCL 생성 → write → sync는 원자적이지 않다. 그 사이에서
//	                   죽으면 0바이트·잘린 JSON이 남고, 내용 파싱 실패는 영구 거부다.
//
// **umask를 테스트에서 바꾸지 않는다.** syscall.Umask는 프로세스 전역이라 병렬로 도는
// 다른 테스트가 만드는 파일 권한까지 오염시킨다. 원하는 권한의 파일을 직접 만드는
// 것으로 같은 디스크 상태를 재현한다(A1 절차).

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// a108DeadSocketWithMode는 죽은 socket을 **원하는 권한 그대로** 남긴다.
// listen이 만든 파일을 chmod로 바꾸므로 프로세스 umask를 건드리지 않는다.
func a108DeadSocketWithMode(t *testing.T, dir string, mode os.FileMode) {
	t.Helper()
	listener, err := net.Listen("unix", SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	listener.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := os.Chmod(SocketPath(dir), mode); err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestStartRecoversFromPreChmodSocketLeftover는 A1 F1이다.
// 구버전 바이너리가 listen과 chmod 사이에서 죽으면 이 모양이 남는다 — 0700 socket.
// 아무도 수락하지 않고, 우리 uid 소유이고, 0700 디렉터리 안이고, hard link도 없다.
// group/other 비트가 하나도 없으므로 회수해야 한다.
func TestStartRecoversFromPreChmodSocketLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o700)
	a108WriteLeftoverDescriptor(t, dir, 1<<30)
	a108StartServes(t, dir)
}

// TestStartRecoversFromEmptyDescriptorLeftover는 A1 F2의 0바이트 모양이다.
// **이번 사고와 같은 S1**(socket 없음 = listener 없음 = 주인 사망)인데 descriptor가
// 반쯤 쓰였다는 이유만으로 영구 거부였다. D2 이후 descriptor 내용은 어떤 판정에도
// 쓰이지 않으므로 그 게이트는 거부만 생산한다.
func TestStartRecoversFromEmptyDescriptorLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	if err := os.WriteFile(DescriptorPath(dir), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	a108StartServes(t, dir)
}

// TestStartRecoversFromTruncatedDescriptorLeftover는 같은 결함의 S3 모양이다.
// socket도 남아 있고(전원 단절) descriptor는 JSON 중간에서 끊겼다.
// 주인의 사망은 socket probe가 증명하고, 그 뒤에 내용 실패는 회수다.
func TestStartRecoversFromTruncatedDescriptorLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o600)
	if err := os.WriteFile(DescriptorPath(dir), []byte(`{"schemaVersion":"tossos.st`), 0o600); err != nil {
		t.Fatal(err)
	}
	a108StartServes(t, dir)
}

// TestStartRecoversFromUnpublishedStagingLeftover는 D1-2가 새로 만드는 잔재다.
// stage+rename 발행은 최종 이름을 갖기 전의 임시 이름을 만든다. 그 이름으로 죽으면
// 남는 것은 **낯선 엔트리가 아니라 자기 잔재**다 — 최종 이름을 가진 적이 없으므로
// 아무도 읽지 않았고, 검증할 것도 없다.
//
// 임시 이름을 코드 상수가 아니라 문자열 그대로 쓴다: 이 테스트가 고정하는 것은
// "회수가 접두사를 안다"가 아니라 **디스크에 남는 모양**이다.
func TestStartRecoversFromUnpublishedStagingLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	for _, name := range []string{".staging-endpoint-2716057", ".staging-sock-9f3ac180"} {
		if err := os.WriteFile(filepath.Join(ControlDirectory(dir), name), []byte("half"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	a108StartServes(t, dir)
	entries, err := os.ReadDir(ControlDirectory(dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".staging") {
			t.Fatalf("발행 전 임시 이름이 살아남았다: %s", entry.Name())
		}
	}
}

// TestDialRefusesSocketWithNoListener는 D4-2의 T1 몫이다(A2 F3 실측).
// 전원이 끊기면 unlink할 기회가 없으므로 socket 파일이 그대로 남는다(S3). 그 위에서
// Dial이 lazy client를 돌려주면 실패는 **첫 Read**에서야 나타나고, 그때는 소비자가
// 이미 "붙었다"고 판단한 뒤라 강등할 기회를 놓친다 — 집계 스냅샷 전체가 함께 죽었다.
func TestDialRefusesSocketWithNoListener(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocketWithMode(t, dir, 0o600)
	a108WriteLeftoverDescriptor(t, dir, 1<<30)
	client, err := Dial(context.Background(), DescriptorPath(dir))
	if err == nil {
		t.Fatalf("아무도 수락하지 않는 socket에 Dial이 성공했다 (client=%v)", client != nil)
	}
}

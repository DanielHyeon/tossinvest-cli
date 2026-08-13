//go:build unix

package strategyprojectionrpc

// a108 — 부팅은 어떤 잔재에서도 회복한다.
//
// 2026-08-13 23:35 KST 호스트 재부팅 뒤 엔진이 영구 기동 루프에 빠졌다. 디스크에는
// `.strategy-runtime-read/`와 `endpoint.json`만 남아 있었고(socket은 graceful shutdown이
// unlink했다), 회수 함수는 "descriptor와 socket이 **둘 다** 있을 때만" 회수했기 때문에
// 어떤 재시도로도 벗어날 수 없었다. 그동안 손절 감시가 US 장중에 사라졌다.
//
// 여기 있는 테스트는 두 묶음이다.
//
//	회수해야 하는 상태   S0(빈 디렉터리)·S1(descriptor만·이번 사고)·S2(socket만·주인 사망)·
//	                     S3(둘 다·주인 사망) — 사람 손 없이 기동이 이어져야 한다
//	거부해야 하는 상태   S2/S3-생존(누가 아직 수락한다)·S4(낯선 엔트리)·
//	                     소유권/권한/symlink 검사 실패 — 우리가 만들지 않은 것,
//	                     아직 남의 것인 것은 치우지 않는다
//
// 생존 판정은 PID(kill-0)가 아니라 **socket connect probe**다. 컨테이너 재생성 후 PID
// 재배정은 근사-결정적이라(a102 D4b-2 실측) descriptor의 PID 자리에 무관한 프로세스가
// 앉는다. 그 오판을 두 방향으로 고정한다:
//
//	죽은 endpoint + 살아 있는 PID → 회수해야 한다 (kill-0은 영구 거부를 만든다)
//	살아 있는 socket + 죽은 PID   → 거부해야 한다 (kill-0은 남의 socket을 지운다)

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

// a108Reader는 기동이 성공했는지만 보면 되는 테스트를 위한 최소 reader다.
func a108Reader() projectionReaderStub {
	return projectionReaderStub{snapshot: strategyprojection.DormantSnapshot(time.Now().UTC())}
}

// a108MakeControlDir는 잔재 디렉터리를 사고 당시와 같은 0700으로 만든다.
func a108MakeControlDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.Mkdir(ControlDirectory(dir), 0o700); err != nil {
		t.Fatal(err)
	}
}

// a108WriteLeftoverDescriptor는 사고 당일 남아 있던 endpoint.json과 같은 모양
// (schemaVersion v1 · socket 이름 · 토큰 · pid)을 쓴다.
func a108WriteLeftoverDescriptor(t *testing.T, dir string, pid int) {
	t.Helper()
	if err := writeDescriptor(DescriptorPath(dir), Descriptor{
		SchemaVersion: descriptorSchema,
		Socket:        SocketFileName,
		Token:         strings.Repeat("x", 43),
		PID:           pid,
	}); err != nil {
		t.Fatal(err)
	}
}

// a108DeadSocket은 socket 파일만 남기고 listener를 없앤다. 전원이 끊겨 unlink를 못 한
// 상태와 같다 — 파일은 있는데 아무도 수락하지 않는다(connect는 ECONNREFUSED).
func a108DeadSocket(t *testing.T, dir string) {
	t.Helper()
	listener, err := net.Listen("unix", SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	listener.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := os.Chmod(SocketPath(dir), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
}

// a108LiveSocket은 실제로 수락하는 listener를 연다. Accept를 부르지 않아도 커널이
// 연결을 받아 주므로 "주인이 있다"는 사실 자체를 만든다.
func a108LiveSocket(t *testing.T, dir string) net.Listener {
	t.Helper()
	listener, err := net.Listen("unix", SocketPath(dir))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	if err := os.Chmod(SocketPath(dir), 0o600); err != nil {
		t.Fatal(err)
	}
	return listener
}

// a108StartServes는 기동이 성공했고 그 endpoint가 실제로 읽힌다는 것까지 본다.
// "거부는 안 했다"만 보면 잔재만 지우고 아무것도 세우지 않은 구현이 통과한다.
func a108StartServes(t *testing.T, dir string) {
	t.Helper()
	server, err := Start(dir, a108Reader())
	if err != nil {
		t.Fatalf("잔재에서 기동이 거부됐다: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client, err := Dial(context.Background(), DescriptorPath(dir))
	if err != nil {
		t.Fatalf("회수 후 descriptor를 못 읽는다: %v", err)
	}
	if _, err := client.Read(context.Background()); err != nil {
		t.Fatalf("회수 후 endpoint가 응답하지 않는다: %v", err)
	}
}

// a108StartRefusesAndKeeps는 기동이 거부되고 **잔재가 그대로 남아 있는지**를 본다.
// 거부하면서 남의 파일을 지우는 구현은 거부만 보는 테스트를 통과한다.
func a108StartRefusesAndKeeps(t *testing.T, dir string, keep ...string) {
	t.Helper()
	server, err := Start(dir, a108Reader())
	if err == nil {
		_ = server.Close()
		t.Fatal("치우면 안 되는 상태에서 기동이 받아들여졌다")
	}
	for _, path := range keep {
		if _, statErr := os.Lstat(path); statErr != nil {
			t.Fatalf("거부하면서 %s를 치웠다: %v", path, statErr)
		}
	}
}

// TestStartRecoversFromDescriptorOnlyLeftover는 **이번 사고 그 상태**다(D1 S1).
// socket이 없다 = listener가 없다 = endpoint는 죽었다.
func TestStartRecoversFromDescriptorOnlyLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108WriteLeftoverDescriptor(t, dir, 16) // 사고 당일 descriptor의 PID
	a108StartServes(t, dir)
}

// TestStartRecoversFromEmptyControlDirectoryLeftover는 D1 S0이다.
// Mkdir 뒤 listen 전에 죽었거나, 회수가 두 파일을 지우고 rmdir 전에 죽은 상태.
// 검증할 것이 없으므로 디렉터리만 치우고 진행한다.
func TestStartRecoversFromEmptyControlDirectoryLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108StartServes(t, dir)
}

// TestStartRecoversFromDeadSocketOnlyLeftover는 D1 S2-사망이다.
// 제거 순서가 descriptor→socket이므로 descriptor만 지우고 죽으면 이 모양이 된다.
func TestStartRecoversFromDeadSocketOnlyLeftover(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocket(t, dir)
	a108StartServes(t, dir)
}

// TestStartRecoversFromDeadEndpointWhoseDescriptorPIDIsAlive는 D2의 핀이다.
// descriptor의 PID 자리에 **살아 있는 남의 프로세스**(여기서는 테스트 자신)가 앉은
// 상태 — kill-0 판정은 "주인이 살아 있다"며 영구 거부하지만, socket은 아무도 수락하지
// 않으므로 죽은 것이다. 이 테스트가 processAlive 되돌리기 뮤테이션을 죽인다.
func TestStartRecoversFromDeadEndpointWhoseDescriptorPIDIsAlive(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocket(t, dir)
	a108WriteLeftoverDescriptor(t, dir, os.Getpid()) // 확실히 살아 있는 PID
	a108StartServes(t, dir)
}

// TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead는 D2 핀의 반대 방향이다.
// PID는 죽었는데 socket은 수락 중 — kill-0 판정은 살아 있는 주인의 socket을 지운다.
// connect probe는 수락하는 자를 보고 거부한다.
func TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108LiveSocket(t, dir)
	a108WriteLeftoverDescriptor(t, dir, 1<<30) // 존재할 수 없는 PID
	a108StartRefusesAndKeeps(t, dir, SocketPath(dir), DescriptorPath(dir))
}

// TestStartRefusesLiveSocketWithoutDescriptor는 D1 S2-생존이다.
// descriptor가 없어도 수락하는 자가 있으면 그 경로의 주인이 있는 것이다.
func TestStartRefusesLiveSocketWithoutDescriptor(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108LiveSocket(t, dir)
	a108StartRefusesAndKeeps(t, dir, SocketPath(dir))
}

// TestStartRefusesControlDirectoryWithUnknownEntry는 D1 S4다.
// 우리 수명주기 밖의 엔트리가 하나라도 있으면 디렉터리 전체를 그대로 둔다.
func TestStartRefusesControlDirectoryWithUnknownEntry(t *testing.T) {
	dir := shortRuntimeDir(t)
	a108MakeControlDir(t, dir)
	a108DeadSocket(t, dir)
	a108WriteLeftoverDescriptor(t, dir, 1<<30)
	stray := filepath.Join(ControlDirectory(dir), "stranger.json")
	if err := os.WriteFile(stray, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	a108StartRefusesAndKeeps(t, dir, stray, SocketPath(dir), DescriptorPath(dir))
}

// TestStartRefusesUnsafeLeftoverShapes는 회수 가능한 상태에서도 유지되는 보안 속성이다.
// 커버리지를 넓히는 것이지 검사를 지우는 것이 아니다(design D1 마지막 문단).
func TestStartRefusesUnsafeLeftoverShapes(t *testing.T) {
	for _, test := range []struct {
		name  string
		build func(t *testing.T, dir string) []string
	}{
		{"디렉터리 권한이 0700이 아니다", func(t *testing.T, dir string) []string {
			a108MakeControlDir(t, dir)
			a108DeadSocket(t, dir)
			a108WriteLeftoverDescriptor(t, dir, 1<<30)
			if err := os.Chmod(ControlDirectory(dir), 0o750); err != nil {
				t.Fatal(err)
			}
			return []string{ControlDirectory(dir), SocketPath(dir)}
		}},
		{"디렉터리가 symlink다", func(t *testing.T, dir string) []string {
			target := filepath.Join(dir, "elsewhere")
			if err := os.Mkdir(target, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, ControlDirectory(dir)); err != nil {
				t.Fatal(err)
			}
			return []string{ControlDirectory(dir), target}
		}},
		{"socket 권한이 0600이 아니다", func(t *testing.T, dir string) []string {
			a108MakeControlDir(t, dir)
			a108DeadSocket(t, dir)
			a108WriteLeftoverDescriptor(t, dir, 1<<30)
			if err := os.Chmod(SocketPath(dir), 0o644); err != nil {
				t.Fatal(err)
			}
			return []string{SocketPath(dir), DescriptorPath(dir)}
		}},
		{"socket 자리에 일반 파일이 있다", func(t *testing.T, dir string) []string {
			a108MakeControlDir(t, dir)
			if err := os.WriteFile(SocketPath(dir), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			return []string{SocketPath(dir)}
		}},
		{"socket에 hard link가 걸려 있다", func(t *testing.T, dir string) []string {
			a108MakeControlDir(t, dir)
			a108DeadSocket(t, dir)
			twin := filepath.Join(dir, "twin.sock")
			if err := os.Link(SocketPath(dir), twin); err != nil {
				t.Skipf("이 파일시스템은 socket hard link를 허용하지 않는다: %v", err)
			}
			return []string{SocketPath(dir), twin}
		}},
		{"descriptor 권한이 0600이 아니다", func(t *testing.T, dir string) []string {
			a108MakeControlDir(t, dir)
			a108DeadSocket(t, dir)
			a108WriteLeftoverDescriptor(t, dir, 1<<30)
			if err := os.Chmod(DescriptorPath(dir), 0o644); err != nil {
				t.Fatal(err)
			}
			return []string{DescriptorPath(dir), SocketPath(dir)}
		}},
		// 이 행은 a108이 **닫지 못한 잔재 상태**를 드러낸 채로 고정한다.
		// writeDescriptor는 O_EXCL 생성 → write → sync 순이라 그 사이에서 죽으면
		// 0바이트 endpoint.json이 남고, 검증 실패는 design D1대로 거부다.
		// 즉 이 상태는 여전히 사람 손을 요구한다 — 보고서에 잔여 결함으로 적었다.
		{"descriptor가 반쯤 쓰인 채 남았다", func(t *testing.T, dir string) []string {
			a108MakeControlDir(t, dir)
			a108DeadSocket(t, dir)
			if err := os.WriteFile(DescriptorPath(dir), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			return []string{DescriptorPath(dir), SocketPath(dir)}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := shortRuntimeDir(t)
			keep := test.build(t, dir)
			a108StartRefusesAndKeeps(t, dir, keep...)
		})
	}
}

// TestCloseToleratesLeftoverAlreadyRemoved는 D2 경합 창의 근거다.
// S1에서 주인이 아직 Close 도중일 수 있고, 그러면 양쪽이 같은 경로를 지운다.
// Close는 ErrNotExist를 용인하므로 그 경합은 양성이다.
func TestCloseToleratesLeftoverAlreadyRemoved(t *testing.T) {
	dir := shortRuntimeDir(t)
	server, err := Start(dir, a108Reader())
	if err != nil {
		t.Fatal(err)
	}
	// 다른 부팅의 회수가 먼저 지나갔다.
	for _, path := range []string{DescriptorPath(dir), SocketPath(dir), ControlDirectory(dir)} {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	if err := server.Close(); err != nil {
		t.Fatalf("이미 치워진 잔재를 Close가 실패로 읽었다: %v", err)
	}
	// 그리고 그 뒤의 기동은 빈 자리를 본다 — 경합이 상태를 남기지 않는다.
	a108StartServes(t, dir)
}

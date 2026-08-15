//go:build unix

package engine

// a109 §1 — **형제 endpoint들도 회복한다.**
//
// a108은 strategy projection 하나를 고쳤고, 그 change의 A1 적대 리뷰가 나머지 세 엔진
// endpoint를 실측해 **같은 병이 표면만 다르게 존재한다**는 것을 보였다. 여기 깔리는 핀은
// a108 2.5의 관용 핀(`a108_every_endpoint_survives_a_leftover_test.go`)과 다르다: 그쪽은
// 현재 코드가 이미 관용하는 모양만 만들어 **실패할 수 없었다**(a108 review §1 F4).
//
// 이 파일이 만드는 것은 **사고급 모양**이다.
//
//	pre-chmod 0700 socket   listen과 chmod 사이의 죽음이 최종 이름에 남긴 잔재.
//	                        회수의 정확-0600 검사가 이것을 매 부팅 영구 거부한다.
//	산 주인의 socket        수락 중인 socket 위에 두 번째 서버가 올라선다(탈취).
//	staging 잔재            발행 전 임시 이름. 아무도 치우지 않아 조용히 쌓인다.
//	낯선 엔트리             우리가 만들 수 없는 이름. 건드리면 남의 것을 지우는 것이다.
//
// ⛔ 잔재 socket의 권한은 **명시적 chmod**로 만든다. umask 조작은 프로세스 전역이라
// 병렬 테스트를 오염시킨다(freeze QA①). 재려는 것은 "umask가 그 권한을 만든다"가 아니라
// "그 권한의 잔재에서 기동이 이어진다"이므로 만드는 방법은 자유다.

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// a109Endpoint는 socket을 발행하는 두 형제를 한 표로 묶는다. 병이 공통이므로 핀도 공통이다.
type a109Endpoint struct {
	name       string
	controlDir func(engineDir string) string
	socket     func(engineDir string) string
	descriptor func(engineDir string) string
	// stagingLeftover는 이 endpoint의 **현행** descriptor 발행이 crash 시 남기는
	// os.CreateTemp 이름 하나다. legacy가 아니라 현행이다 — descriptor 발행 3벌의
	// fold는 a109의 선언된 생략이므로 a109 바이너리도 같은 이름을 계속 만든다.
	stagingLeftover string
	start           func(t *testing.T, engineDir string) (interface{ Close() error }, error)
}

func a109Endpoints() []a109Endpoint {
	return []a109Endpoint{
		{
			name:            "position policy runtime",
			controlDir:      positionpolicyrpc.RuntimeControlDirectory,
			socket:          positionpolicyrpc.RuntimeSocketPath,
			descriptor:      positionpolicyrpc.RuntimeDescriptorPath,
			stagingLeftover: ".position-policy-runtime-2118745398",
			start: func(_ *testing.T, engineDir string) (interface{ Close() error }, error) {
				return StartPositionPolicyRuntimeServer(engineDir, &fakePolicyCommands{
					runtime: positionpolicy.ManagementRuntime{EffectiveKnown: true},
				})
			},
		},
		{
			name:            "alert control",
			controlDir:      AlertControlDirectory,
			socket:          AlertControlSocketPath,
			descriptor:      AlertControlDescriptorPath,
			stagingLeftover: ".endpoint-3641220975",
			start: func(_ *testing.T, engineDir string) (interface{ Close() error }, error) {
				return StartAlertControlServer(engineDir, &AlertOperations{})
			},
		},
	}
}

// a109ControlDir는 지난 기동이 남긴 control 디렉터리를 만든다(0700 — 그때도 0700이었다).
func a109ControlDir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
}

// a109DeadSocket은 아무도 수락하지 않는 socket 파일을 주어진 권한으로 남긴다.
//
// SetUnlinkOnClose(false)가 있어야 Close가 경로를 지우지 않는다 — 지워지면 만들려던
// 잔재가 사라진다.
func a109DeadSocket(t *testing.T, path string, perm os.FileMode) {
	t.Helper()
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	listener.(*net.UnixListener).SetUnlinkOnClose(false)
	if err := os.Chmod(path, perm); err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode().Perm() != perm {
		t.Fatalf("잔재를 만들지 못했다: info=%v err=%v want perm=%04o", info, err, perm)
	}
}

// a109Accepts는 endpoint가 말로만 기동한 것이 아니라 실제로 수락하는지 본다.
func a109Accepts(t *testing.T, path string) {
	t.Helper()
	conn, err := net.DialTimeout("unix", path, time.Second)
	if err != nil {
		t.Fatalf("endpoint가 수락하지 않는다(%s): %v", path, err)
	}
	_ = conn.Close()
}

// TestTheSiblingEndpointsStartOverAPreChmodSocket — a109 §1.1.
//
// 사고의 모양 그대로다. 구버전(=오늘의) 발행은 **최종 이름에 바로 bind한 뒤 chmod**
// 하므로, 그 두 줄 사이에서 죽으면 최종 이름 자리에 umask가 정한 권한(컨테이너 실측
// 077 → 0700)의 socket이 남는다. 그리고 회수의 정확-0600 검사가 그것을 "우리 것이
// 아니다"로 읽어 **매 부팅 거부**한다. 재시도로는 벗어날 수 없다 — 거부가 결정적이기
// 때문이다.
//
// 그 거부는 `cmd/tossctl/engine.go`에서 fatal이라 엔진 기동 루프 전체가 죽는다. 즉 이
// 핀이 재는 것은 "조회 화면이 뜨는가"가 아니라 **"장중에 손절이 서는가"**다.
func TestTheSiblingEndpointsStartOverAPreChmodSocket(t *testing.T) {
	for _, endpoint := range a109Endpoints() {
		t.Run(endpoint.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			a109ControlDir(t, endpoint.controlDir(dir))
			a109DeadSocket(t, endpoint.socket(dir), 0o700)

			server, err := endpoint.start(t, dir)
			if err != nil {
				t.Fatalf("pre-chmod 0700 socket 잔재에서 기동이 거부됐다: %v", err)
			}
			t.Cleanup(func() { _ = server.Close() })

			a109Accepts(t, endpoint.socket(dir))
			// 회수가 관용한 권한이 **발행**으로 새면 안 된다. 새 socket은 0600이다.
			info, err := os.Lstat(endpoint.socket(dir))
			if err != nil || info.Mode().Perm() != 0o600 {
				t.Fatalf("재발행된 socket info=%v err=%v, want perm=0600", info, err)
			}
		})
	}
}

// TestTheSiblingEndpointsRefuseToTakeALiveOwnersSocket — a109 §1.2.
//
// a108이 strategy projection에 세운 규칙은 **"수락 중인 socket은 건드리지 않는다"**이다.
// 두 형제는 정반대로 한다: 유효한 0600 socket이면 주인의 생사를 묻지 않고 unlink하고
// 그 자리에 자기 listener를 올린다. 그래서 두 번째 서버가 첫 번째의 운영자 표면을
// 가져간다 — 첫 번째의 클라이언트는 자기 토큰이 통하지 않는 socket을 보게 된다.
//
// 실제 방어가 journal flock뿐이라는 것이 문제의 절반이다. flock은 **엔진**을 하나로
// 강제하지만, 이 경로를 점유한 것이 엔진이 아니거나 다른 journal 디렉터리의 엔진이면
// 막지 못한다. 그 심층 방어가 없다는 사실을 코드도 문서도 말하지 않았다.
//
// 이 핀이 재는 것은 두 가지다: ① 두 번째 기동이 **거부**되는가, ② 거부하면서 주인의
// socket을 **그대로 두는가**. 둘째가 없으면 "거부했지만 이미 지웠다"가 통과한다.
func TestTheSiblingEndpointsRefuseToTakeALiveOwnersSocket(t *testing.T) {
	for _, endpoint := range a109Endpoints() {
		t.Run(endpoint.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			a109ControlDir(t, endpoint.controlDir(dir))
			socketPath := endpoint.socket(dir)

			// 산 주인: 실제로 수락하는 listener. 커널이 backlog에 받아 주므로
			// Accept 루프가 없어도 connect는 성공한다 — 그것이 "살아 있다"의 정의다.
			owner, err := net.Listen("unix", socketPath)
			if err != nil {
				t.Fatal(err)
			}
			owner.(*net.UnixListener).SetUnlinkOnClose(false)
			t.Cleanup(func() { _ = owner.Close(); _ = os.Remove(socketPath) })
			if err := os.Chmod(socketPath, 0o600); err != nil {
				t.Fatal(err)
			}
			before, err := os.Lstat(socketPath)
			if err != nil {
				t.Fatal(err)
			}
			a109Accepts(t, socketPath)

			server, startErr := endpoint.start(t, dir)
			if startErr == nil {
				_ = server.Close()
				t.Fatal("주인이 수락 중인 socket 위에서 두 번째 기동이 성공했다 — " +
					"산 endpoint의 탈취다")
			}
			after, err := os.Lstat(socketPath)
			if err != nil || !os.SameFile(before, after) {
				t.Fatalf("거부하면서 주인의 socket을 지웠다: err=%v", err)
			}
			a109Accepts(t, socketPath)
		})
	}
}

// TestTheSiblingEndpointsRejectAPermissiveSocketLeftover는 완화의 **폭**을 잰다.
//
// §1.1이 통과시키는 것은 "group/other 비트가 없는" 잔재뿐이다. 0640처럼 남이 읽을 수
// 있는 socket은 우리가 만들 수 있는 모양이 아니므로 여전히 거부다. 이 핀이 없으면
// "0700을 받아들인다"가 "아무 권한이나 받아들인다"로 넓어져도 아무도 모른다.
func TestTheSiblingEndpointsRejectAPermissiveSocketLeftover(t *testing.T) {
	for _, endpoint := range a109Endpoints() {
		t.Run(endpoint.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			a109ControlDir(t, endpoint.controlDir(dir))
			a109DeadSocket(t, endpoint.socket(dir), 0o640)

			server, err := endpoint.start(t, dir)
			if err == nil {
				_ = server.Close()
				t.Fatal("group 읽기가 열린 socket 잔재 위에서 기동이 성공했다 — " +
					"완화의 폭이 우리가 만들 수 있는 모양을 넘었다")
			}
			if _, err := os.Lstat(endpoint.socket(dir)); err != nil {
				t.Fatalf("거부하면서 잔재를 지웠다: %v", err)
			}
		})
	}
}

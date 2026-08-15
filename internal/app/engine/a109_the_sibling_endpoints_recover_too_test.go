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
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
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

// TestTheSiblingEndpointsRefuseALiveOwnerWithoutTheWriteBit — a109 §1-fix F1.
//
// §1.2와 디스크 상태가 하나만 다르다: 주인의 socket에서 **owner 쓰기 비트가 깎여 있다**.
// 예전 판의 probe는 그 비트 하나로 주인을 죽었다고 **추정**했고, 그래서 수락 중인
// socket을 지우고 두 번째 서버가 그 자리에 섰다. 권한 비트는 주인의 생사를 말해 주지
// 않는다 — 물어봐야 안다(probe 전에 0600으로 chmod하고 연결한다).
func TestTheSiblingEndpointsRefuseALiveOwnerWithoutTheWriteBit(t *testing.T) {
	for _, endpoint := range a109Endpoints() {
		t.Run(endpoint.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			a109ControlDir(t, endpoint.controlDir(dir))
			socketPath := endpoint.socket(dir)

			owner, err := net.Listen("unix", socketPath)
			if err != nil {
				t.Fatal(err)
			}
			owner.(*net.UnixListener).SetUnlinkOnClose(false)
			t.Cleanup(func() { _ = owner.Close(); _ = os.Remove(socketPath) })
			if err := os.Chmod(socketPath, 0o400); err != nil {
				t.Fatal(err)
			}
			before, err := os.Lstat(socketPath)
			if err != nil {
				t.Fatal(err)
			}

			server, startErr := endpoint.start(t, dir)
			if startErr == nil {
				_ = server.Close()
				t.Fatal("owner 쓰기 비트가 깎였다는 이유로 산 주인의 socket을 " +
					"지우고 두 번째 기동이 성공했다 — 추정이 만든 탈취다")
			}
			after, err := os.Lstat(socketPath)
			if err != nil || !os.SameFile(before, after) {
				t.Fatalf("거부하면서 주인의 socket을 지웠다: err=%v", err)
			}
			// 권한과 무관하게 주인은 계속 수락한다. (probe가 0600을 복원했든 아니든
			// 이 단정이 재는 것은 inode가 그대로 살아 있다는 사실이다.)
			if err := os.Chmod(socketPath, 0o600); err != nil {
				t.Fatal(err)
			}
			a109Accepts(t, socketPath)
		})
	}
}

// TestTheSiblingEndpointsRefuseALiveStagingSocket — a109 §1-fix F5.
//
// 회수는 최종 이름에만 probe를 걸었다. staging 이름의 socket은 "우리 잔재"라는 이름표만
// 보고 probe 없이 지웠다. 그런데 발행의 **첫 걸음**이 staging 이름에 bind 하는 것이다 —
// 그 창에 있는 후계자의 socket이 정확히 이 모양이다.
//
// 규칙은 이름이 아니라 사실이어야 한다: 수락 중인 socket은 이름과 무관하게 unlink하지
// 않는다.
func TestTheSiblingEndpointsRefuseALiveStagingSocket(t *testing.T) {
	for _, endpoint := range a109Endpoints() {
		t.Run(endpoint.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			controlDir := endpoint.controlDir(dir)
			a109ControlDir(t, controlDir)
			live := filepath.Join(controlDir, positionpolicyrpc.StagingPrefix+"sfeedfac")
			owner, err := net.Listen("unix", live)
			if err != nil {
				t.Fatal(err)
			}
			owner.(*net.UnixListener).SetUnlinkOnClose(false)
			t.Cleanup(func() { _ = owner.Close(); _ = os.Remove(live) })
			if err := os.Chmod(live, 0o600); err != nil {
				t.Fatal(err)
			}
			a109Accepts(t, live)

			server, startErr := endpoint.start(t, dir)
			if startErr == nil {
				_ = server.Close()
				t.Fatal("수락 중인 staging socket 위에서 기동이 성공했다 — " +
					"발행 중인 후계자의 socket을 지운 것이다")
			}
			if _, err := os.Lstat(live); err != nil {
				t.Fatalf("거부하면서 수락 중인 staging socket을 지웠다: %v", err)
			}
			a109Accepts(t, live)
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

// TestTheStagedSocketNameFitsInsideEverySiblingsFinalName — a109 §1.5.
//
// bind 하는 이름은 최종 이름이 아니라 **임시 이름**이고, unix socket 경로에는 sun_path
// 상한이 있다. 임시 이름이 최종 이름보다 길면 최종 경로는 상한 안인데 발행만 죽는
// 배포가 생기고, 운영자에게 보이는 것은 「매 부팅 그 표면이 없다」뿐이라 원인이 보이지
// 않는다.
//
// ⛔ 재는 것은 **basename 길이**이지 절대 경로 길이가 아니다. 두 이름은 같은 디렉터리
// 안에 있으므로 basename이 길지 않으면 경로도 길지 않다 — 디렉터리가 얼마나 깊든
// 성립한다. 절대 경로 상한(Linux 실측 107)을 테스트에 넣으면 그것은 호스트의 tmp 경로
// 길이를 재는 테스트가 된다.
//
// 형제 중 `alerts.sock`이 11자로 가장 짧다. a108(strategy projection)의 12자 staging을
// 그대로 쓰면 여기서 계약이 깨진다 — 그래서 형제 공용 staging은 11자다.
func TestTheStagedSocketNameFitsInsideEverySiblingsFinalName(t *testing.T) {
	staged, err := positionpolicyrpc.StagedSocketName()
	if err != nil {
		t.Fatal(err)
	}
	for _, final := range []string{
		AlertControlSocketFileName(),            // alerts.sock — 11자
		positionpolicyrpc.RuntimeSocketFileName, // runtime.sock — 12자
	} {
		if len(staged) > len(final) {
			t.Errorf("임시 이름 %q(%d자)이 최종 이름 %q(%d자)보다 길다 — "+
				"최종 경로가 sun_path 상한 직전인 배포에서 bind만 죽는다",
				staged, len(staged), final, len(final))
		}
	}
}

// TestEveryNameThePublishingPathMakesIsKnownToItsReclaim — freeze P2-5.
//
// 아는-이름 집합에서 하나라도 빠지면 낯선-엔트리 거부가 **우리 자신의 잔재**를 거부한다.
// 그 거부는 결정적이라 매 부팅 같은 자리에서 멈추고, 운영자가 볼 수 있는 것은 그 표면이
// 사라진 것뿐이다.
//
// 측정하는 것과 구조로 묶은 것을 나눠 적는다.
//
//	측정   socket staging 은 발행이 쓰는 **그 생성기**(StagedSocketName)를 실제로 돌리고,
//	       descriptor staging 은 **같은 상수로** os.CreateTemp 를 돌려 나온 이름을 집합에
//	       물어본다. os.CreateTemp 가 접두 뒤에 붙이는 임의 숫자의 길이는 판마다 다르므로
//	       "접두로 시작한다"는 가정도 여기서 깨질 수 있다.
//	정적   ⛔ 이 테스트는 **발행 경로가 그 상수를 쓴다는 것을 재지 않는다** — 상수를
//	       가져다 자기가 CreateTemp 를 돌릴 뿐이다. 발행 쪽이 리터럴로 갈라져도 여기는
//	       통과한다(A1 뮤테이션 A1-N1 이 그렇게 살아남았다). 그 절반은
//	       TestEveryPublishedStagingPrefixIsTheSharedConstant 가 소스를 파서 잰다.
func TestEveryNameThePublishingPathMakesIsKnownToItsReclaim(t *testing.T) {
	for _, test := range []struct {
		name             string
		names            positionpolicyrpc.PrivateEndpointNames
		descriptorPrefix string
		final            []string
	}{
		{
			name:             "position policy runtime",
			names:            positionPolicyRuntimeEndpointNames(),
			descriptorPrefix: positionPolicyRuntimeStagingPrefix,
			final: []string{positionpolicyrpc.RuntimeDescriptorFileName,
				positionpolicyrpc.RuntimeSocketFileName},
		},
		{
			name:             "alert control",
			names:            alertControlEndpointNames(),
			descriptorPrefix: privateDescriptorStagingPrefix,
			final:            []string{alertControlDescriptorFileName, alertControlSocketFileName},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			made := append([]string{}, test.final...)

			staged, err := positionpolicyrpc.StagedSocketName()
			if err != nil {
				t.Fatal(err)
			}
			made = append(made, staged)

			temporary, err := os.CreateTemp(t.TempDir(), test.descriptorPrefix+"*")
			if err != nil {
				t.Fatal(err)
			}
			if err := temporary.Close(); err != nil {
				t.Fatal(err)
			}
			made = append(made, filepath.Base(temporary.Name()))

			for _, name := range made {
				if !test.names.Knows(name) {
					t.Errorf("발행이 만드는 이름 %q를 회수가 모른다 — "+
						"낯선-엔트리 거부가 우리 잔재를 매 부팅 거부한다", name)
				}
			}
			// 그리고 아무 이름이나 아는 것이어서는 안 된다. 빈 접두 하나가 그렇게 만든다.
			if test.names.Knows("somebody-elses.json") {
				t.Error("회수가 낯선 이름을 우리 것으로 안다 — 남의 파일을 치우게 된다")
			}
		})
	}
}

// a109StagingPrefixSources는 세 transport의 (파일, 공유 상수) 짝이다.
//
// 셋 다 "descriptor를 임시 이름에 만들고 rename한다"는 같은 의례를 각자 한 벌씩 가지고
// 있다(fold는 a109의 선언된 생략). 그래서 접두도 셋이고, 각각 발행과 회수 **두 자리**에서
// 읽힌다.
func a109StagingPrefixSources() []struct{ file, constant string } {
	return []struct{ file, constant string }{
		{"alert_control_transport_unix.go", "privateDescriptorStagingPrefix"},
		{"position_policy_runtime_transport_unix.go", "positionPolicyRuntimeStagingPrefix"},
		{"position_policy_transport.go", "positionPolicyControlStagingPrefix"},
	}
}

// a109FileNameArgument는 파일을 만드는 호출에서 **이름을 정하는 인자**를 돌려준다.
func a109FileNameArgument(call *ast.CallExpr) (ast.Expr, string, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, "", false
	}
	pkg, ok := selector.X.(*ast.Ident)
	if !ok || pkg.Name != "os" {
		return nil, "", false
	}
	switch {
	case selector.Sel.Name == "CreateTemp" && len(call.Args) == 2:
		return call.Args[1], "os.CreateTemp", true // (dir, pattern)
	case selector.Sel.Name == "OpenFile" && len(call.Args) == 3:
		return call.Args[0], "os.OpenFile", true // (name, flag, perm)
	}
	return nil, "", false
}

// TestEveryPublishedStagingPrefixIsTheSharedConstant — a109 §1-fix F2.
//
// 발행과 회수가 **한 소스**를 읽어야 한다. 두 곳에 따로 적으면 회수가 자기 잔재를 낯선
// 것으로 거부하고, 그 거부는 결정적이라 매 부팅 같은 자리에서 멈춘다.
//
// 그런데 "상수를 쓴다"는 동적 테스트로 잴 수 없다. 이름-집합 완전성 테스트는 **그 상수로**
// os.CreateTemp를 돌려 보므로, 발행 쪽이 리터럴로 갈라져도 그 테스트는 통과한다 — A1
// 적대 리뷰의 뮤테이션 A1-N1(발행 접두만 `.endpoint2-`로)이 그렇게 살아남았다.
//
// 그래서 **소스를 판다.** 파일을 만드는 호출(os.CreateTemp·os.OpenFile)에서 이름을 정하는
// 인자에 staging 접두 모양의 문자열 리터럴이 있으면 실패한다. 접두는 공유 상수 참조여야
// 하고, 그 상수는 같은 파일의 회수·위생 쪽에서도 읽혀야 한다.
func TestEveryPublishedStagingPrefixIsTheSharedConstant(t *testing.T) {
	for _, source := range a109StagingPrefixSources() {
		t.Run(source.file, func(t *testing.T) {
			fileSet := token.NewFileSet()
			parsed, err := parser.ParseFile(fileSet, source.file, nil, 0)
			if err != nil {
				t.Fatal(err)
			}

			publications := 0
			ast.Inspect(parsed, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				argument, callee, ok := a109FileNameArgument(call)
				if !ok {
					return true
				}
				publications++
				usesConstant := false
				ast.Inspect(argument, func(inner ast.Node) bool {
					switch value := inner.(type) {
					case *ast.BasicLit:
						if value.Kind == token.STRING && value.Value != `"*"` &&
							strings.HasPrefix(value.Value, `".`) {
							t.Errorf("%s:%d — %s의 이름 인자에 접두 리터럴 %s가 있다. "+
								"발행이 회수와 갈라지면 회수가 자기 잔재를 낯선 것으로 "+
								"거부하고, 그 거부는 결정적이다 — %s를 쓸 것",
								source.file, fileSet.Position(value.Pos()).Line, callee,
								value.Value, source.constant)
						}
					case *ast.Ident:
						if value.Name == source.constant {
							usesConstant = true
						}
					}
					return true
				})
				if callee == "os.CreateTemp" && !usesConstant {
					t.Errorf("%s:%d — os.CreateTemp의 접두가 %s를 참조하지 않는다",
						source.file, fileSet.Position(call.Pos()).Line, source.constant)
				}
				return true
			})
			if publications == 0 {
				t.Fatalf("%s에서 파일을 만드는 호출을 하나도 찾지 못했다 — "+
					"이 핀은 발행 경로를 잰다고 주장하면서 아무것도 재지 않고 있다",
					source.file)
			}

			// 그리고 그 상수는 **두 자리 이상**에서 읽혀야 한다: 선언 · 발행 · 회수(또는
			// 위생). 셋 미만이면 발행과 회수 중 한쪽이 그 상수를 안 쓰고 있다는 뜻이다.
			references := 0
			ast.Inspect(parsed, func(node ast.Node) bool {
				if identifier, ok := node.(*ast.Ident); ok && identifier.Name == source.constant {
					references++
				}
				return true
			})
			if references < 3 {
				t.Errorf("%s에서 %s 참조가 %d개다(선언 1 + 발행 1 + 회수 1 이상이어야 한다) — "+
					"한쪽이 그 상수를 읽지 않는다", source.file, source.constant, references)
			}
		})
	}
}

// a109ControlEntries는 control 디렉터리에 남은 이름을 정렬해 돌려준다.
func a109ControlEntries(t *testing.T, controlDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(controlDir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names
}

// TestTheSiblingEndpointsReclaimTheirStagingLeftovers — a109 §1.3(잔재 쪽).
//
// 발행 전 임시 이름으로 죽은 산출물이다. 두 종류가 있고 **둘 다 현행이다**:
//
//	os.CreateTemp 이름   descriptor 발행 3벌이 오늘도 만든다(`.position-policy-runtime-*`,
//	                     `.endpoint-*`). fold는 a109의 선언된 생략이므로 계속 만들어진다.
//	`.s-` 이름           a109가 socket·descriptor 발행에 들이는 staging 접두.
//
// 오늘은 어떤 검증도 디렉터리를 **열거하지 않으므로** 이것들이 조용히 쌓인다. 기동은
// 계속되니 아무도 모르고, 부팅마다 하나씩 늘어난다.
//
// 재는 것은 "기동이 되는가"가 아니라 **"회수 후 잔재가 0인가"**다. 전자는 오늘도 참이다.
func TestTheSiblingEndpointsReclaimTheirStagingLeftovers(t *testing.T) {
	for _, endpoint := range a109Endpoints() {
		t.Run(endpoint.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			controlDir := endpoint.controlDir(dir)
			a109ControlDir(t, controlDir)

			// 현행 CreateTemp 잔재(반쯤 쓰인 descriptor)
			if err := os.WriteFile(filepath.Join(controlDir, endpoint.stagingLeftover),
				[]byte(`{"socket":"`), 0o600); err != nil {
				t.Fatal(err)
			}
			// 신규 `.s-` 잔재 두 모양: 정규 파일(descriptor가 되려다 만 것)과
			// socket(bind는 됐는데 rename 전에 죽은 것)
			if err := os.WriteFile(filepath.Join(controlDir, ".s-e1a2b3c"),
				[]byte(""), 0o600); err != nil {
				t.Fatal(err)
			}
			a109DeadSocket(t, filepath.Join(controlDir, ".s-s7f6e5d4"), 0o600)

			server, err := endpoint.start(t, dir)
			if err != nil {
				t.Fatalf("staging 잔재에서 기동이 거부됐다: %v", err)
			}
			t.Cleanup(func() { _ = server.Close() })

			a109Accepts(t, endpoint.socket(dir))
			got := a109ControlEntries(t, controlDir)
			want := []string{filepath.Base(endpoint.descriptor(dir)),
				filepath.Base(endpoint.socket(dir))}
			sort.Strings(want)
			if !slices.Equal(got, want) {
				t.Fatalf("회수 후 control 디렉터리 = %v, want %v — "+
					"치우지 않은 staging 잔재는 부팅마다 하나씩 쌓인다", got, want)
			}
		})
	}
}

// TestTheSiblingEndpointsRefuseAForeignEntry — a109 §1.3(낯섦 쪽).
//
// 우리가 만들 수 없는 이름이 있으면 이 디렉터리는 우리 잔재가 아니다. 그러면 **아무것도
// 건드리지 않고 거부한다** — 치우면 남의 것을 치우는 것이고, 남의 것을 치우지 않으려면
// 우리 것도 못 치운다.
//
// 이 규칙은 socket을 발행하는 endpoint에만 적용된다(design D2a). command endpoint에
// 같은 거부를 넣으면 이물 하나가 격리 해제 표면을 매 부팅 지운다 — 아래 command 핀이
// 그 반대 방향을 잰다.
func TestTheSiblingEndpointsRefuseAForeignEntry(t *testing.T) {
	for _, endpoint := range a109Endpoints() {
		t.Run(endpoint.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			controlDir := endpoint.controlDir(dir)
			a109ControlDir(t, controlDir)
			foreign := filepath.Join(controlDir, "somebody-elses.json")
			if err := os.WriteFile(foreign, []byte("{}"), 0o600); err != nil {
				t.Fatal(err)
			}

			server, err := endpoint.start(t, dir)
			if err == nil {
				_ = server.Close()
				t.Fatal("낯선 엔트리가 있는 control 디렉터리에서 기동이 성공했다 — " +
					"이 디렉터리는 우리 잔재가 아니다")
			}
			if _, err := os.Lstat(foreign); err != nil {
				t.Fatalf("거부하면서 낯선 엔트리를 지웠다: %v", err)
			}
		})
	}
}

// TestTheSiblingEndpointsRefuseAForeignEntryAndKeepTheirOwnLeftovers — a109 §1-fix F4.
//
// 낯선 엔트리 하나만 있는 §1.3 핀은 "거부하면서 남의 것을 지우지 않는가"만 잰다. 그런데
// 거부의 진짜 계약은 **아무것도 건드리지 않는 것**이다: 우리 잔재까지 그대로 있어야 한다.
//
// 왜 그것이 중요한가. 거부하면서 우리 socket 잔재를 지우면, 운영자가 이물을 치운 뒤의
// 재시작은 **다른 디스크 상태**에서 출발한다 — 재현이 안 되는 회복이다. 그리고 그 삭제는
// probe가 "죽었다"고 읽은 결과이므로, 같은 경로가 산 주인의 socket에도 열려 있다.
//
// A1 적대 리뷰가 이 조합(pre-chmod 잔재 + 이물)을 만들었고, 뮤테이션 M1a(낯선 엔트리를
// 거부 대신 무시)가 정확히 이 핀에서 죽는다.
func TestTheSiblingEndpointsRefuseAForeignEntryAndKeepTheirOwnLeftovers(t *testing.T) {
	for _, endpoint := range a109Endpoints() {
		t.Run(endpoint.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			controlDir := endpoint.controlDir(dir)
			a109ControlDir(t, controlDir)
			a109DeadSocket(t, endpoint.socket(dir), 0o700)
			foreign := filepath.Join(controlDir, "somebody-elses.json")
			if err := os.WriteFile(foreign, []byte("{}"), 0o600); err != nil {
				t.Fatal(err)
			}

			server, err := endpoint.start(t, dir)
			if err == nil {
				_ = server.Close()
				t.Fatal("낯선 엔트리가 있는데 기동이 성공했다 — " +
					"이 디렉터리는 우리 잔재가 아니다")
			}
			if _, err := os.Lstat(foreign); err != nil {
				t.Fatalf("거부하면서 낯선 엔트리를 지웠다: %v", err)
			}
			if _, err := os.Lstat(endpoint.socket(dir)); err != nil {
				t.Fatalf("거부하면서 우리 socket 잔재를 지웠다: %v — "+
					"거부는 아무것도 건드리지 않는 것이다", err)
			}
		})
	}
}

// TestTheSiblingEndpointsNameTheEntryTheyRefused — a109 §1-fix F6.
//
// D3 강등의 회복 수단은 "엔진 재시작"이 아니라 **"보고된 원인을 제거한 뒤 재시작"**이다
// (freeze P0-4) — 원인이 결정적이라 재시작만으로는 같은 강등이 재현되기 때문이다. 그
// 회복 경로는 거부가 **무엇을** 거부했는지 말할 때만 실행 가능하다.
//
// 재는 것은 운영자가 실제로 보는 문자열이다: Start의 오류는 회수의 오류를 감싸므로,
// 여기서 이름이 보이면 엔진 로그의 강등 보고에도 보인다.
func TestTheSiblingEndpointsNameTheEntryTheyRefused(t *testing.T) {
	for _, endpoint := range a109Endpoints() {
		t.Run(endpoint.name, func(t *testing.T) {
			dir := privateEngineTestDir(t)
			controlDir := endpoint.controlDir(dir)
			a109ControlDir(t, controlDir)
			// 우리 잔재와 이물이 함께 있다 — 이름이 없으면 셋 중 무엇이 원인인지 모른다.
			a109DeadSocket(t, endpoint.socket(dir), 0o700)
			if err := os.WriteFile(endpoint.descriptor(dir), []byte("{}"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(controlDir, "somebody-elses.json"),
				[]byte("{}"), 0o600); err != nil {
				t.Fatal(err)
			}

			server, err := endpoint.start(t, dir)
			if err == nil {
				_ = server.Close()
				t.Fatal("낯선 엔트리가 있는데 기동이 성공했다")
			}
			if !strings.Contains(err.Error(), "somebody-elses.json") {
				t.Fatalf("기동 거부 오류 %q가 위반 엔트리를 지목하지 않는다 — "+
					"운영자는 무엇을 치워야 할지 모른 채 재시작만 반복한다", err)
			}
		})
	}
}

// TestTheCommandEndpointSweepsItsStagingAndLeavesStrangers — a109 §1.3(command 쪽).
//
// loopback TCP endpoint에는 socket이 없으므로 pre-chmod 병도 탈취도 없다. 남는 병은
// 하나다: `os.CreateTemp(".position-policy-control-*")`가 crash 때 남기는 잔재를
// **아무도 치우지 않는다.**
//
// 그래서 여기 더하는 것은 **위생뿐**이고, 낯선 엔트리는 오늘처럼 무시한다. 이 endpoint가
// 잃는 것은 격리 해제 — 격리된 포지션의 손절 포함 미판정 상태를 푸는 유일한 장중
// 경로다. 이물 하나가 그 표면을 매 부팅 지우는 새 실패 경로를 만들지 않는다(D2a).
func TestTheCommandEndpointSweepsItsStagingAndLeavesStrangers(t *testing.T) {
	dir := privateEngineTestDir(t)
	controlDir := positionpolicyrpc.ControlDirectory(dir)
	a109ControlDir(t, controlDir)
	staging := filepath.Join(controlDir, ".position-policy-control-1902847563")
	if err := os.WriteFile(staging, []byte(`{"address":`), 0o600); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(controlDir, "somebody-elses.json")
	if err := os.WriteFile(foreign, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	server, err := StartPositionPolicyCommandServer(dir, &fakePolicyCommands{})
	if err != nil {
		t.Fatalf("staging 잔재와 이물이 있는 디렉터리에서 command 기동이 거부됐다: %v", err)
	}
	t.Cleanup(func() { _ = server.Close() })

	if _, err := os.Lstat(staging); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("자기 staging 잔재가 남았다(err=%v) — 부팅마다 하나씩 쌓인다", err)
	}
	if _, err := os.Lstat(foreign); err != nil {
		t.Fatalf("낯선 엔트리를 지웠다: %v — command endpoint의 회수는 위생뿐이다", err)
	}
}

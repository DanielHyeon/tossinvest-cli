//go:build unix

package positionpolicyrpc

// a109 §1.4·§1.5 — 이름-독립 발행·회수 기계의 경계.
//
// 여기서 재는 것은 세 가지다.
//
//	이름의 길이   임시 basename 이 최종 basename 보다 길면, 최종 경로는 sun_path 상한
//	              안인데 발행만 죽는 배포가 생긴다. 길이는 **직접 센다**.
//	완화의 위치   회수의 좁은 완화(perm&0o077==0)가 클라이언트 검증으로 새면 안 된다.
//	              두 클라이언트 검증이 0700 socket 을 거부하는지 잰다.
//	판정의 갈래   probe 의 **두** 사망 판정과 생존 판정을 각각 만든다(a109 §1-fix F1로
//	              owner-write 추정 절이 사라져 셋에서 둘이 됐다).
//	거부의 지목    거부 오류가 위반 엔트리를 지목하는지(§1-fix F6). D3 회복 경로가 그것을
//	              전제한다 — 원인이 결정적이라 "재시작"만으로는 같은 강등이 재현된다.

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// a109TestSocket은 아무도 수락하지 않는 socket 파일을 주어진 권한으로 만든다.
func a109TestSocket(t *testing.T, path string, perm os.FileMode) {
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
}

// a109TestControlDir은 이름-독립 검증을 통과하는 control 디렉터리를 만든다.
func a109TestControlDir(t *testing.T, engineDir, name string) string {
	t.Helper()
	controlDir := filepath.Join(engineDir, name)
	if err := os.Mkdir(controlDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return controlDir
}

// TestTheStagedSocketNameIsElevenCharacters — a109 §1.5.
//
// 길이를 상수에서 계산하지 않고 **생성기의 산출물을 센다.** 계산으로 재면 재는 것은
// 계산이지 이름이 아니다. hex 절단이 홀수 길이를 내는지도 여기서 확인된다 —
// `hex.EncodeToString`은 짝수 길이만 내므로 절단이 빠지면 12자가 된다.
func TestTheStagedSocketNameIsElevenCharacters(t *testing.T) {
	for round := 0; round < 64; round++ {
		name, err := StagedSocketName()
		if err != nil {
			t.Fatal(err)
		}
		if got := len(name); got != 11 {
			t.Fatalf("staging basename %q = %d자, want 11 — "+
				"형제 중 가장 짧은 최종 socket 이름(alerts.sock)이 11자다", name, got)
		}
		if !strings.HasPrefix(name, StagingPrefix) {
			t.Fatalf("staging basename %q에 회수가 아는 접두가 없다", name)
		}
		// 난수가 실제로 섞이는지: 두 번 뽑아 같으면 이름 충돌 방어가 없는 것이다.
		other, err := StagedSocketName()
		if err != nil {
			t.Fatal(err)
		}
		if name == other {
			t.Fatalf("연속 두 staging 이름이 같다(%q) — 난수가 붙지 않았다", name)
		}
	}
}

// TestTheStagedListenBindsAStagingNameNotTheFinalName — a109 §1-fix F3 (원장 M23).
//
// 원장은 M23(발행을 최종 경로 bind로 되돌린다)을 "crash 창이라 죽일 수 없다"고 선언했다.
// **프로세스를 죽일 필요가 없다** — bind한 이름이 무엇인지는 listener가 그대로 들고 있다.
// A1 적대 리뷰가 이 반증을 만들었다.
//
// 재는 것은 D1의 의례 그 자체다: bind는 staging 이름에 하고, 최종 이름은 완성된 0600
// socket으로만 나타난다. 최종 이름에 직접 bind하면 pre-chmod 상태가 최종 이름을 가지는
// 창이 다시 열린다 — 사고의 발생원이다.
func TestTheStagedListenBindsAStagingNameNotTheFinalName(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	final := filepath.Join(controlDir, "endpoint.sock")

	listener, err := ListenStagedPrivateSocket(controlDir, final)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close(); _ = os.Remove(final) })

	bound := listener.Addr().String()
	if bound == final {
		t.Fatal("최종 이름에 직접 bind했다 — pre-chmod 상태가 최종 이름을 가지는 창이 " +
			"다시 열렸다(design D1)")
	}
	if !strings.HasPrefix(filepath.Base(bound), StagingPrefix) {
		t.Fatalf("bind한 이름 %q에 회수가 아는 staging 접두가 없다 — "+
			"회수가 자기 잔재를 낯선 것으로 거부한다", bound)
	}
	info, err := os.Lstat(final)
	if err != nil || info.Mode()&os.ModeSocket == 0 || info.Mode().Perm() != 0o600 {
		t.Fatalf("최종 이름 info=%v err=%v, want 0600 socket", info, err)
	}
}

// TestTheStagedListenDoesNotUnlinkTheNameItRemembers — a109 §1-fix F3 (원장 M22).
//
// 원장은 M22(`SetUnlinkOnClose(false)` 제거)를 "경합이라 관측 가능한 피해가 없다"고
// 선언했다. **경합을 재현할 필요가 없다** — listener가 기억하는 이름에 파일을 다시
// 만들어 두고 Close하면, unlink가 켜져 있으면 그 파일이 사라지고 꺼져 있으면 남는다.
// A1 적대 리뷰가 이 반증을 만들었다.
//
// 재는 것은 D2b의 전제다: 경로를 지울 권한은 Close의 제거 루프 하나뿐이다. listener가
// 자기 이름을 늦게 unlink하면 그 정리가 후계자의 파일에 도착한다(a108 A1 F5).
func TestTheStagedListenDoesNotUnlinkTheNameItRemembers(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	final := filepath.Join(controlDir, "endpoint.sock")

	listener, err := ListenStagedPrivateSocket(controlDir, final)
	if err != nil {
		t.Fatal(err)
	}
	remembered := listener.Addr().String()
	t.Cleanup(func() { _ = os.Remove(final); _ = os.Remove(remembered) })

	// 후계자가 같은 임시 이름을 쓴 상황을 만든다(rename으로 비었던 자리다).
	if err := os.WriteFile(remembered, []byte("successor"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(remembered); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Close가 자기가 기억하는 이름 %q를 unlink했다 — "+
			"늦은 정리가 후계자의 파일을 지운다(design D2b)", remembered)
	}
}

// TestTheClientSocketChecksStayExactlyZeroSixHundred — freeze P1-3.
//
// 회수는 pre-chmod 0700 잔재를 관용하지만, 그 완화는 **회수 전용**이다. 클라이언트·발행
// 확인 경로가 같이 넓어지면 "발행이 0600이다"라는 계약을 아무도 확인하지 않게 된다.
//
// 공유 helper(statPrivateSocket)를 완화하는 손쉬운 구현이 정확히 그 누출을 만든다 —
// 이 테스트가 그것을 죽인다.
func TestTheClientSocketChecksStayExactlyZeroSixHundred(t *testing.T) {
	t.Run("ValidatePrivateSocket", func(t *testing.T) {
		engineDir := privateTestDir(t)
		controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
		path := filepath.Join(controlDir, "endpoint.sock")
		a109TestSocket(t, path, 0o700)
		if err := ValidatePrivateSocket(path); err == nil {
			t.Fatal("클라이언트 검증이 0700 socket을 받아들였다 — 회수 전용 완화가 샜다")
		}
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := ValidatePrivateSocket(path); err != nil {
			t.Fatalf("0600 socket을 거부했다: %v", err)
		}
	})

	t.Run("ValidateRuntimeSocket", func(t *testing.T) {
		engineDir := privateTestDir(t)
		a109TestControlDir(t, engineDir, RuntimeControlDirectoryName)
		path := RuntimeSocketPath(engineDir)
		a109TestSocket(t, path, 0o700)
		if err := ValidateRuntimeSocket(path); err == nil {
			t.Fatal("클라이언트 검증이 0700 socket을 받아들였다 — 회수 전용 완화가 샜다")
		}
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := ValidateRuntimeSocket(path); err != nil {
			t.Fatalf("0600 socket을 거부했다: %v", err)
		}
	})
}

// TestThePrivateSocketProbeReadsOnlyTwoThingsAsDead — design D2의 사망 판정.
//
// 잘못 살아 있다고 보면 이번 기동만 실패한다. 잘못 죽었다고 보면 **남의 socket을
// 지운다.** 그래서 사망으로 읽는 것은 **둘뿐**이다 — 연결 거부와 파일 부재.
//
// 셋이 아니라 둘인 것이 a109 §1-fix F1이다. 예전 판은 "owner 쓰기 비트가 없다"를 세 번째
// 사망 판정으로 썼는데, 그것은 **추정**이었다: 쓰기 비트가 깎인 socket이 수락 중일 수도
// 있고, 그때 그 추정은 산 주인의 socket을 지운다. 지금은 추정하지 않고 **묻는다** —
// probe 전에 0600으로 chmod 하고 연결해 본다.
func TestThePrivateSocketProbeReadsOnlyTwoThingsAsDead(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")

	t.Run("수락 중이면 살아 있다", func(t *testing.T) {
		path := filepath.Join(controlDir, "alive.sock")
		listener, err := net.Listen("unix", path)
		if err != nil {
			t.Fatal(err)
		}
		listener.(*net.UnixListener).SetUnlinkOnClose(false)
		t.Cleanup(func() { _ = listener.Close(); _ = os.Remove(path) })
		if !privateSocketAccepts(path) {
			t.Fatal("수락 중인 socket을 죽었다고 읽었다 — 이 오판이 남의 socket을 지운다")
		}
	})

	t.Run("연결 거부는 사망", func(t *testing.T) {
		path := filepath.Join(controlDir, "refused.sock")
		a109TestSocket(t, path, 0o600)
		t.Cleanup(func() { _ = os.Remove(path) })
		if privateSocketAccepts(path) {
			t.Fatal("아무도 듣지 않는 socket을 살아 있다고 읽었다 — 영구 거부가 된다")
		}
	})

	t.Run("파일 부재는 사망", func(t *testing.T) {
		if privateSocketAccepts(filepath.Join(controlDir, "absent.sock")) {
			t.Fatal("없는 socket을 살아 있다고 읽었다")
		}
	})

	t.Run("쓰기 비트가 깎인 죽은 socket도 사망", func(t *testing.T) {
		// connect 에는 쓰기 권한이 필요하고, 없으면 커널은 EACCES 를 준다. 그 오류를
		// "답이 안 왔다 = 살아 있다"로 읽으면 아무도 없는 socket 하나가 영구 거부를
		// 만든다 — 이 change 가 지우려는 모양 그대로다. 그래서 EACCES 를 만나지 않는다:
		// probe 가 먼저 0600 으로 chmod 하므로 이 잔재는 **연결 거부**로 갈린다.
		path := filepath.Join(controlDir, "unwritable.sock")
		a109TestSocket(t, path, 0o400)
		t.Cleanup(func() { _ = os.Remove(path) })
		if privateSocketAccepts(path) {
			t.Fatal("아무도 듣지 않는 socket을 살아 있다고 읽었다 — 영구 거부가 된다")
		}
	})

	// ⛔ a109 §1-fix F1의 핵심. 위 잔재와 **디스크 상태가 같고**(0400 socket) 결과만
	// 반대다. 권한 비트는 주인의 생사를 말해 주지 않는다 — 물어봐야 안다.
	t.Run("쓰기 비트가 깎여도 수락 중이면 생존", func(t *testing.T) {
		path := filepath.Join(controlDir, "unwritable-alive.sock")
		listener, err := net.Listen("unix", path)
		if err != nil {
			t.Fatal(err)
		}
		listener.(*net.UnixListener).SetUnlinkOnClose(false)
		t.Cleanup(func() { _ = listener.Close(); _ = os.Remove(path) })
		if err := os.Chmod(path, 0o400); err != nil {
			t.Fatal(err)
		}
		if !privateSocketAccepts(path) {
			t.Fatal("수락 중인 socket을 권한 비트만 보고 죽었다고 읽었다 — " +
				"그 추정이 산 주인의 socket을 지운다")
		}
	})
}

// TestReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped — a109 §1-fix F1.
//
// A1 재현: 주인이 **수락 중인** socket 의 owner 쓰기 비트만 깎여 있으면, owner-write
// 사망 추정이 그것을 죽었다고 읽고 회수가 지운다. 그리고 그 자리에 두 번째 서버가 선다.
//
// 재는 것은 둘이다 — ① 회수가 **거부**하는가, ② 거부하면서 주인의 socket 을 **그대로
// 두는가**. 둘째가 없으면 "거부했지만 이미 지웠다"가 통과한다.
func TestReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	socketPath := filepath.Join(controlDir, "endpoint.sock")
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

	names := PrivateEndpointNames{
		Descriptor: "endpoint.json", Socket: "endpoint.sock",
		StagingPrefixes: []string{StagingPrefix},
	}
	if err := ReclaimStalePrivateEndpoint(controlDir, names); err == nil {
		t.Fatal("수락 중인 socket을 쓰기 비트만 보고 회수했다 — 산 주인의 탈취다")
	}
	after, err := os.Lstat(socketPath)
	if err != nil || !os.SameFile(before, after) {
		t.Fatalf("거부하면서 주인의 socket을 지웠다: err=%v", err)
	}
	if !privateSocketAccepts(socketPath) {
		t.Fatal("주인의 socket이 더 이상 수락하지 않는다")
	}
}

// TestReclaimRefusesALiveStagingSocket — a109 §1-fix F5.
//
// A1 재현: 회수는 **최종 이름에만** probe 를 걸었다. staging 이름의 socket 은 "우리
// 잔재"라는 이름표만 보고 probe 없이 지웠다. 그런데 발행의 첫 걸음이 바로 staging 이름에
// bind 하는 것이다 — 그 창에 있는 후계자의 socket 이 정확히 이 모양이다.
//
// 규칙은 이름이 아니라 사실이어야 한다: **수락 중인 socket 은 이름과 무관하게 절대
// unlink 하지 않는다.**
func TestReclaimRefusesALiveStagingSocket(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	live := filepath.Join(controlDir, StagingPrefix+"sfeedfac")
	owner, err := net.Listen("unix", live)
	if err != nil {
		t.Fatal(err)
	}
	owner.(*net.UnixListener).SetUnlinkOnClose(false)
	t.Cleanup(func() { _ = owner.Close(); _ = os.Remove(live) })
	if err := os.Chmod(live, 0o600); err != nil {
		t.Fatal(err)
	}
	if !privateSocketAccepts(live) {
		t.Fatal("준비 실패: staging socket이 수락 중이 아니다")
	}

	names := PrivateEndpointNames{
		Descriptor: "endpoint.json", Socket: "endpoint.sock",
		StagingPrefixes: []string{StagingPrefix},
	}
	if err := ReclaimStalePrivateEndpoint(controlDir, names); err == nil {
		t.Fatal("수락 중인 staging socket을 probe 없이 회수했다")
	}
	if _, err := os.Lstat(live); err != nil {
		t.Fatalf("거부하면서 수락 중인 staging socket을 지웠다: %v", err)
	}
	if !privateSocketAccepts(live) {
		t.Fatal("staging socket이 더 이상 수락하지 않는다")
	}
}

// TestReclaimRefusesNamesItCannotTrust — 이름 집합 자체의 방어.
//
// 빈 접두는 **모든 이름을 staging으로** 만든다. 즉 낯선 엔트리 거부가 통째로 사라지고
// 회수가 남의 파일을 치우게 된다. 호출자의 실수 하나가 그 결과를 낳지 않게 막는다.
func TestReclaimRefusesNamesItCannotTrust(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	stranger := filepath.Join(controlDir, "somebody-elses.json")
	if err := os.WriteFile(stranger, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name  string
		names PrivateEndpointNames
	}{
		{"descriptor 이름이 없다", PrivateEndpointNames{Socket: "x.sock"}},
		{"빈 staging 접두", PrivateEndpointNames{
			Descriptor: "endpoint.json", StagingPrefixes: []string{StagingPrefix, ""}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := ReclaimStalePrivateEndpoint(controlDir, test.names); err == nil {
				t.Fatal("회수가 믿을 수 없는 이름 집합을 받아들였다")
			}
			if _, err := os.Lstat(stranger); err != nil {
				t.Fatalf("거부하면서 남의 파일을 지웠다: %v", err)
			}
		})
	}
}

// TestReclaimRefusesAStagingEntryOfTheWrongShape — 모양 검증.
//
// 이름만 보고 지우면 우리가 임시 이름으로 만들 수 없는 모양까지 지우려 든다: 비어 있으면
// 남의 디렉터리를 지우고, 비어 있지 않으면 ENOTEMPTY 로 회수 전체가 매 부팅 멈춘다.
func TestReclaimRefusesAStagingEntryOfTheWrongShape(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	intruder := filepath.Join(controlDir, StagingPrefix+"sdeadbee")
	if err := os.Mkdir(intruder, 0o700); err != nil {
		t.Fatal(err)
	}
	names := PrivateEndpointNames{
		Descriptor: "endpoint.json", Socket: "endpoint.sock",
		StagingPrefixes: []string{StagingPrefix},
	}
	if err := ReclaimStalePrivateEndpoint(controlDir, names); err == nil {
		t.Fatal("임시 이름을 가진 디렉터리를 회수 대상으로 받아들였다")
	}
	if _, err := os.Lstat(intruder); err != nil {
		t.Fatalf("거부하면서 그 디렉터리를 지웠다: %v", err)
	}
}

// TestReclaimRefusesASocketWithASecondHardLink — 검증을 완화하지 않는다.
//
// 잔재 socket의 perm 검사만 완화했고 나머지는 그대로 요구한다. hard link 가 하나 더
// 걸린 socket 은 우리가 만든 것이 아니거나 누군가 이름을 하나 더 붙인 것이므로,
// 지우면 그 이름 쪽 주인의 것을 지우는 셈이다.
//
// (이 핀은 뮤테이션 M10a — `validateOwnerAndLinks(info, true)` → `false` — 가 살아남아
// 추가됐다. 소유 uid 절은 비root 테스트로 죽일 수 없어 원장에 생존으로 남는다.)
func TestReclaimRefusesASocketWithASecondHardLink(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	socketPath := filepath.Join(controlDir, "endpoint.sock")
	a109TestSocket(t, socketPath, 0o600)
	if err := os.Link(socketPath, filepath.Join(engineDir, "second-name.sock")); err != nil {
		t.Skipf("이 파일시스템은 socket 의 hard link 를 허용하지 않는다: %v", err)
	}
	names := PrivateEndpointNames{
		Descriptor: "endpoint.json", Socket: "endpoint.sock",
		StagingPrefixes: []string{StagingPrefix},
	}
	if err := ReclaimStalePrivateEndpoint(controlDir, names); err == nil {
		t.Fatal("hard link 가 걸린 socket 을 회수 대상으로 받아들였다")
	}
	if _, err := os.Lstat(socketPath); err != nil {
		t.Fatalf("거부하면서 그 socket 을 지웠다: %v", err)
	}
}

// TestReclaimRefusesAnUnsafeControlDirectory — 디렉터리 위생은 완화하지 않는다.
//
// owner 비트를 깎거나 group 에 열어 준 control 디렉터리는 **환경 이상**이지 우리 잔재가
// 아니다. 회수가 그것을 검사 없이 비우면, "우리 것인지 모르는 디렉터리를 통째로 지우는"
// 동작이 된다(design D2, P1-7①).
func TestReclaimRefusesAnUnsafeControlDirectory(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	leftover := filepath.Join(controlDir, "endpoint.json")
	if err := os.WriteFile(leftover, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(controlDir, 0o750); err != nil {
		t.Fatal(err)
	}
	names := PrivateEndpointNames{
		Descriptor: "endpoint.json", StagingPrefixes: []string{StagingPrefix},
	}
	if err := ReclaimStalePrivateEndpoint(controlDir, names); err == nil {
		t.Fatal("0700이 아닌 control 디렉터리를 회수했다")
	}
	if _, err := os.Lstat(leftover); err != nil {
		t.Fatalf("거부하면서 그 안의 파일을 지웠다: %v", err)
	}
}

// TestReclaimRefusesADescriptorOfTheWrongShape — 관용하는 것은 **내용**뿐이다.
//
// 0바이트·잘린 JSON 은 우리 자신의 반쯤 쓴 파일이므로 관용한다(그 관용은
// TestReclaimEmptiesTheDirectoryItRecovers 가 고정한다). 하지만 **모양**은 여전히 본다 —
// descriptor 자리의 디렉터리나 남에게 열린 파일은 우리가 만들 수 있는 것이 아니다.
func TestReclaimRefusesADescriptorOfTheWrongShape(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	intruder := filepath.Join(controlDir, "endpoint.json")
	if err := os.Mkdir(intruder, 0o700); err != nil {
		t.Fatal(err)
	}
	names := PrivateEndpointNames{
		Descriptor: "endpoint.json", StagingPrefixes: []string{StagingPrefix},
	}
	if err := ReclaimStalePrivateEndpoint(controlDir, names); err == nil {
		t.Fatal("descriptor 자리의 디렉터리를 회수 대상으로 받아들였다")
	}
	if _, err := os.Lstat(intruder); err != nil {
		t.Fatalf("거부하면서 그것을 지웠다: %v", err)
	}
}

// TestReclaimSaysWhichEntryItRefused — a109 §1-fix F6.
//
// D3 강등의 회복 수단은 "엔진 재시작"이 아니라 **"보고된 원인을 제거한 뒤 재시작"**이다
// (freeze P0-4). 원인이 결정적이므로 재시작만으로는 같은 강등이 재현되기 때문이다.
//
// 그 회복 경로에는 전제가 있다: 거부가 **무엇을** 거부했는지 말해야 한다. control
// 디렉터리에 파일이 여럿이면 "unexpected entries"만으로는 운영자가 어느 것을 치워야
// 할지 모르고, 지울 대상을 고르는 일은 그 자체로 위험하다.
//
// 경로가 아니라 basename 이면 충분하다 — 디렉터리는 이미 오류의 문맥에 있다.
func TestReclaimSaysWhichEntryItRefused(t *testing.T) {
	names := PrivateEndpointNames{
		Descriptor: "endpoint.json", Socket: "endpoint.sock",
		StagingPrefixes: []string{StagingPrefix},
	}

	t.Run("낯선 엔트리", func(t *testing.T) {
		engineDir := privateTestDir(t)
		controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
		// 우리 잔재도 함께 둔다 — 이름이 없으면 운영자는 셋 중 무엇이 원인인지 모른다.
		if err := os.WriteFile(filepath.Join(controlDir, "endpoint.json"),
			[]byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
		a109TestSocket(t, filepath.Join(controlDir, StagingPrefix+"s1234567"), 0o600)
		if err := os.WriteFile(filepath.Join(controlDir, "somebody-elses.json"),
			[]byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}

		err := ReclaimStalePrivateEndpoint(controlDir, names)
		if err == nil {
			t.Fatal("낯선 엔트리를 받아들였다")
		}
		if !strings.Contains(err.Error(), "somebody-elses.json") {
			t.Fatalf("거부 오류 %q가 위반 엔트리를 지목하지 않는다 — "+
				"운영자는 무엇을 치워야 할지 모른 채 재시작만 반복한다", err)
		}
	})

	t.Run("모양이 다른 staging 엔트리", func(t *testing.T) {
		engineDir := privateTestDir(t)
		controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
		if err := os.Mkdir(filepath.Join(controlDir, StagingPrefix+"sdeadbee"), 0o700); err != nil {
			t.Fatal(err)
		}

		err := ReclaimStalePrivateEndpoint(controlDir, names)
		if err == nil {
			t.Fatal("모양이 다른 staging 엔트리를 받아들였다")
		}
		if !strings.Contains(err.Error(), StagingPrefix+"sdeadbee") {
			t.Fatalf("거부 오류 %q가 위반 엔트리를 지목하지 않는다", err)
		}
	})

	t.Run("수락 중인 staging socket", func(t *testing.T) {
		engineDir := privateTestDir(t)
		controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
		live := filepath.Join(controlDir, StagingPrefix+"sfeedfac")
		listener, err := net.Listen("unix", live)
		if err != nil {
			t.Fatal(err)
		}
		listener.(*net.UnixListener).SetUnlinkOnClose(false)
		t.Cleanup(func() { _ = listener.Close(); _ = os.Remove(live) })
		if err := os.Chmod(live, 0o600); err != nil {
			t.Fatal(err)
		}

		err = ReclaimStalePrivateEndpoint(controlDir, names)
		if err == nil {
			t.Fatal("수락 중인 staging socket을 받아들였다")
		}
		if !strings.Contains(err.Error(), StagingPrefix+"sfeedfac") {
			t.Fatalf("거부 오류 %q가 수락 중인 엔트리를 지목하지 않는다", err)
		}
	})
}

// TestReclaimEmptiesTheDirectoryItRecovers — 회수의 전체성.
//
// 회수가 끝나면 디렉터리 자체가 없어야 한다. 남으면 다음 줄의 Mkdir 이 ErrExist 로
// 실패하고, 그 실패는 결정적이라 매 부팅 같은 자리에서 멈춘다.
func TestReclaimEmptiesTheDirectoryItRecovers(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	if err := os.WriteFile(filepath.Join(controlDir, "endpoint.json"),
		[]byte(`{"token":`), 0o600); err != nil {
		t.Fatal(err)
	}
	a109TestSocket(t, filepath.Join(controlDir, "endpoint.sock"), 0o700)
	a109TestSocket(t, filepath.Join(controlDir, StagingPrefix+"s1234567"), 0o600)
	if err := os.WriteFile(filepath.Join(controlDir, StagingPrefix+"e7654321"),
		[]byte(""), 0o600); err != nil {
		t.Fatal(err)
	}

	names := PrivateEndpointNames{
		Descriptor: "endpoint.json", Socket: "endpoint.sock",
		StagingPrefixes: []string{StagingPrefix},
	}
	if err := ReclaimStalePrivateEndpoint(controlDir, names); err != nil {
		t.Fatalf("회수가 자기 잔재를 거부했다: %v", err)
	}
	if _, err := os.Lstat(controlDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("회수 후 control 디렉터리가 남았다: %v", err)
	}
}

// TestTheStagingSweepLeavesEverythingElseAlone — command endpoint의 위생(design D2a).
//
// 자기 접두의 잔재만 치우고, 낯선 엔트리도 모양이 다른 것도 건드리지 않는다. 그리고
// 무엇을 못 치우든 **오류를 내지 않는다** — 이 위생이 격리 해제 표면을 없애서는 안 된다.
func TestTheStagingSweepLeavesEverythingElseAlone(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	ours := filepath.Join(controlDir, ".mine-1234567890")
	if err := os.WriteFile(ours, []byte("half written"), 0o600); err != nil {
		t.Fatal(err)
	}
	stranger := filepath.Join(controlDir, "somebody-elses.json")
	if err := os.WriteFile(stranger, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	wrongShape := filepath.Join(controlDir, ".mine-0987654321")
	if err := os.Mkdir(wrongShape, 0o700); err != nil {
		t.Fatal(err)
	}

	SweepPrivateStagingLeftovers(controlDir, []string{".mine-"})

	if _, err := os.Lstat(ours); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("자기 staging 잔재가 남았다: %v", err)
	}
	if _, err := os.Lstat(stranger); err != nil {
		t.Fatalf("낯선 엔트리를 지웠다: %v", err)
	}
	if _, err := os.Lstat(wrongShape); err != nil {
		t.Fatalf("우리가 만들 수 없는 모양을 지웠다: %v", err)
	}
	// 없는 디렉터리에 위생을 돌려도 조용해야 한다 — 새 실패 경로를 만들지 않는다.
	SweepPrivateStagingLeftovers(filepath.Join(engineDir, "no-such-dir"), []string{".mine-"})
}

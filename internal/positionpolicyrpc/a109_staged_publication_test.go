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
//	판정의 갈래   probe 의 세 사망 판정과 생존 판정을 각각 만든다.

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

// TestThePrivateSocketProbeReadsOnlyThreeThingsAsDead — design D2의 사망 판정.
//
// 잘못 살아 있다고 보면 이번 기동만 실패한다. 잘못 죽었다고 보면 **남의 socket을
// 지운다.** 그래서 사망으로 읽는 것은 셋뿐이고, 그 셋을 각각 만든다.
func TestThePrivateSocketProbeReadsOnlyThreeThingsAsDead(t *testing.T) {
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

	t.Run("owner 쓰기 비트가 없으면 사망", func(t *testing.T) {
		// connect 에는 쓰기 권한이 필요하고, 없으면 커널은 EACCES 를 준다. 그 오류를
		// "답이 안 왔다 = 살아 있다"로 읽으면 아무도 없는 socket 하나가 영구 거부를
		// 만든다. 우리가 발행한 산 endpoint 는 반드시 0600 이므로 그렇게 읽어도 된다.
		path := filepath.Join(controlDir, "unwritable.sock")
		a109TestSocket(t, path, 0o400)
		t.Cleanup(func() { _ = os.Remove(path) })
		if privateSocketAccepts(path) {
			t.Fatal("owner가 쓸 수 없는 socket을 살아 있다고 읽었다")
		}
	})
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

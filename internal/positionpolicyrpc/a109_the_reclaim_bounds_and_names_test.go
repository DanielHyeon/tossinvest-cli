//go:build unix

package positionpolicyrpc

// a109_the_reclaim_bounds_and_names_test.go — a109 §2b.3 G6·G7·G8.
//
// 회수는 부팅 1단계에서 journal flock 을 쥔 채로 돈다. 그래서 이 함수의 두 성질이
// 운영 성질이 된다: **얼마나 오래 도는가**와 **거부할 때 무엇을 말하는가.**
//
//	G6  최종 socket 거부가 이름도 이유도 말하지 않는다("stale socket is unsafe" 한 줄).
//	    D3 의 회복 수단은 "보고된 원인을 제거한 뒤 재시작"이므로 이름 없는 거부는
//	    실행 가능한 안내가 아니다 — §1-fix F6 이 다른 두 경로에만 적용됐다.
//	G7  staging 엔트리에 nlink 요구가 없다. 하드 링크로 inode 를 공유한 파일이 우리
//	    잔재로 통과하고, probe 전 chmod 0600 이 **그 inode 에** 걸린다.
//	G8  probe 횟수에 상한이 없다. 잔재 socket N 개면 순차 200ms × N 을 flock 을 쥔 채
//	    문다 — 엔진 기동이 그만큼 늦어지고, 그 시간에 손절은 없다.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// a109BoundedNames 는 이 파일이 쓰는 이름 집합이다.
func a109BoundedNames() PrivateEndpointNames {
	return PrivateEndpointNames{
		Descriptor: "endpoint.json", Socket: "endpoint.sock",
		StagingPrefixes: []string{StagingPrefix},
	}
}

// TestReclaimNamesWhichClauseRefusedTheSocket 는 G6 다.
//
// 세 갈래(모양·symlink·권한)가 하나의 문장으로 접혀 있으면 운영자는 socket 을 지워야
// 하는지, 권한을 고쳐야 하는지, symlink 를 걷어야 하는지 알 수 없다. 이름도 없다 —
// 형제 endpoint 셋이 같은 문장을 쓰므로 어느 디렉터리인지조차 오류만으로는 모른다.
func TestReclaimNamesWhichClauseRefusedTheSocket(t *testing.T) {
	cases := []struct {
		name    string
		make    func(t *testing.T, controlDir string)
		clue    string
		refused string
	}{
		{
			name: "socket 이 아니다",
			make: func(t *testing.T, controlDir string) {
				if err := os.WriteFile(filepath.Join(controlDir, "endpoint.sock"),
					[]byte("not a socket"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			clue:    "socket",
			refused: "endpoint.sock",
		},
		{
			name: "group·other 에 열려 있다",
			make: func(t *testing.T, controlDir string) {
				a109TestSocket(t, filepath.Join(controlDir, "endpoint.sock"), 0o640)
			},
			clue:    "0640",
			refused: "endpoint.sock",
		},
		{
			name: "symlink 다",
			make: func(t *testing.T, controlDir string) {
				target := filepath.Join(controlDir, StagingPrefix+"s7777777")
				a109TestSocket(t, target, 0o600)
				if err := os.Symlink(target, filepath.Join(controlDir, "endpoint.sock")); err != nil {
					t.Fatal(err)
				}
			},
			clue:    "symlink",
			refused: "endpoint.sock",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			engineDir := privateTestDir(t)
			controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
			testCase.make(t, controlDir)

			err := ReclaimStalePrivateEndpoint(controlDir, a109BoundedNames())
			if err == nil {
				t.Fatal("우리가 만들 수 없는 모양의 socket 을 회수 대상으로 받아들였다")
			}
			if !strings.Contains(err.Error(), testCase.refused) {
				t.Errorf("거부 오류 %q 가 엔트리 이름을 지목하지 않는다 — 운영자는 형제 셋 중 "+
					"어느 디렉터리의 무엇을 치워야 할지 모른다", err)
			}
			if !strings.Contains(err.Error(), testCase.clue) {
				t.Errorf("거부 오류 %q 가 **어느 절**이 거부했는지 말하지 않는다 (want %q) — "+
					"지우라는 것인지 권한을 고치라는 것인지 알 수 없다", err, testCase.clue)
			}
		})
	}
}

// TestReclaimRefusesAStagingEntryWithASecondHardLink 는 G7 의 앞 절이다.
//
// 우리 staging 은 언제나 링크 하나다: `os.CreateTemp` 도 `net.Listen` 도 새 이름을
// 만들고, 발행은 rename 으로 이름을 **옮긴다**(링크를 더하지 않는다). 그러므로 링크가
// 둘이면 그것은 우리 잔재가 아니거나 누군가 이름을 하나 더 붙인 것이고, 회수의 chmod
// 0600 과 unlink 는 그 두 번째 이름 쪽 주인에게도 일어난다.
//
// 최종 이름 쪽은 이미 nlink 1 을 요구한다(`TestReclaimRefusesASocketWithASecondHardLink`).
// 여기서 닫는 것은 staging 쪽의 비대칭이다.
func TestReclaimRefusesAStagingEntryWithASecondHardLink(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	staged := filepath.Join(controlDir, StagingPrefix+"e1234567")
	if err := os.WriteFile(staged, []byte("half written descriptor"), 0o600); err != nil {
		t.Fatal(err)
	}
	second := filepath.Join(engineDir, "second-name.json")
	if err := os.Link(staged, second); err != nil {
		t.Skipf("이 파일시스템은 hard link 를 허용하지 않는다: %v", err)
	}

	err := ReclaimStalePrivateEndpoint(controlDir, a109BoundedNames())
	if err == nil {
		t.Fatal("링크가 둘인 staging 엔트리를 우리 잔재로 받아들였다 — 회수가 남의 이름 " +
			"쪽 내용을 지운다")
	}
	if _, statErr := os.Lstat(staged); statErr != nil {
		t.Fatalf("거부하면서 그 엔트리를 지웠다: %v", statErr)
	}
}

// TestTheProbeRefusesASocketThatChangedUnderIt 는 G7 의 뒷 절이다.
//
// probe 는 EACCES 를 없애려고 **chmod 0600 을 먼저** 한다(§1-fix F1). 그 chmod 는
// 검증할 때 본 inode 가 아니라 **이름**에 걸린다 — 같은 uid 의 다른 프로세스가 그 사이에
// 이름을 갈아끼우면 우리는 다른 파일의 권한을 바꾸고 그 파일에 연결한다. 우리 신뢰
// 경계(같은 uid)가 그것을 "불가능"으로 만들지는 않는다. 그러니 **확인한다**: chmod 뒤
// 다시 Lstat 해서 같은 파일이 아니면 답을 얻지 못한 것이고, 답을 얻지 못한 것은
// 살아 있는 것으로 읽는다(보수 방향 — 기동을 거부시킨다).
func TestTheProbeRefusesASocketThatChangedUnderIt(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	socketPath := filepath.Join(controlDir, "endpoint.sock")
	a109TestSocket(t, socketPath, 0o600)
	before, err := os.Lstat(socketPath)
	if err != nil {
		t.Fatal(err)
	}
	// 대조군: 그 자리의 그 파일이면 죽었다고 답한다.
	if privateSocketAccepts(socketPath, before) {
		t.Fatal("아무도 수락하지 않는 socket 을 살아 있다고 읽었다 — 대조군이 틀렸다")
	}

	// 같은 이름, 다른 파일. 검증한 것과 만지는 것이 갈라진 순간이다.
	if err := os.Remove(socketPath); err != nil {
		t.Fatal(err)
	}
	a109TestSocket(t, socketPath, 0o600)
	if !privateSocketAccepts(socketPath, before) {
		t.Error("검증한 파일과 다른 파일을 **죽었다**고 읽었다 — 회수는 그것을 지운다. " +
			"물어보지 못한 것을 죽었다고 읽지 않는다")
	}
}

// TestReclaimRefusesToProbeAnUnboundedNumberOfSockets 는 G8 이다.
//
// probe 하나는 최대 `privateProbeTimeout`(200ms)이고 순차다. 그리고 이 함수는 부팅
// 1단계에서 journal flock 을 쥔 채로 돈다 — 잔재 socket 이 수십 개면 엔진 기동이 그만큼
// 늦고, 그 시간에 손절 루프는 아직 없다. 상한을 넘으면 **치우지 않고 거부한다**:
// 지우는 쪽으로 서두르면 그것은 probe 없는 unlink 이고, 그 방향은 산 주인의 socket 을
// 지운다.
func TestReclaimRefusesToProbeAnUnboundedNumberOfSockets(t *testing.T) {
	engineDir := privateTestDir(t)
	controlDir := a109TestControlDir(t, engineDir, ".private-endpoint-under-test")
	over := maxPrivateSocketProbes + 1
	for index := range over {
		a109TestSocket(t, filepath.Join(controlDir, fmt.Sprintf("%ss%07d", StagingPrefix, index)), 0o600)
	}

	err := ReclaimStalePrivateEndpoint(controlDir, a109BoundedNames())
	if err == nil {
		t.Fatalf("잔재 socket %d개를 flock 을 쥔 채 전부 probe 했다 — 순차 200ms × %d 가 "+
			"엔진 기동 앞에 붙는다", over, over)
	}
	overflow := fmt.Sprintf("%ss%07d", StagingPrefix, maxPrivateSocketProbes)
	if !strings.Contains(err.Error(), overflow) {
		t.Errorf("거부 오류 %q 가 상한을 넘긴 엔트리(%s)를 지목하지 않는다", err, overflow)
	}
	if _, statErr := os.Lstat(filepath.Join(controlDir, overflow)); statErr != nil {
		t.Fatalf("거부하면서 엔트리를 지웠다: %v", statErr)
	}
}

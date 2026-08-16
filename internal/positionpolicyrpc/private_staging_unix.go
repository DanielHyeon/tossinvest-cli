//go:build unix

package positionpolicyrpc

// private_staging_unix.go는 socket을 발행하는 endpoint의 발행·회수 의례다. 이름-독립이고,
// 이식 원형은 `internal/strategyprojectionrpc/transport_unix.go`의 listenPrivateSocket ·
// reclaimStaleControlDirectory · projectionSocketAccepts다(a108 확정 코드 — 원형은
// 수정하지 않는다).

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// privateProbeTimeout은 잔재 socket에 "누가 받아 주나"를 묻는 시간이다. 같은 호스트의
// Unix socket 연결은 즉시 끝나므로, 이 시간을 다 쓰면 답이 안 온 것이고 그것은 죽었다는
// 증거가 아니다 — 그 경우는 살아 있다고 보고 기동을 거부한다.
const privateProbeTimeout = 200 * time.Millisecond

// maxPrivateSocketProbes는 한 번의 회수가 물어보는 socket 수의 상한이다 — a109 §2b.3 G8.
//
// # 왜 상한이 필요한가
//
// probe는 **순차**이고 하나가 최대 privateProbeTimeout이다. 그리고 회수는 부팅 1단계에서
// journal flock을 쥔 채로 돈다 — N개면 N × 200ms가 통째로 엔진 기동 앞에 붙고, 그 시간에
// 보호 루프는 아직 서지 않았다. 상한 없는 순차 대기를 flock 안에 두지 않는다.
//
// # 왜 10인가
//
// 우리 수명주기가 한 control 디렉터리에 남길 수 있는 socket은 기동 하나당 **최대 둘**이다
// (최종 이름 + 발행 중이던 staging 이름). 회수는 매 기동에 돌고 디렉터리째 비우므로,
// 정상 상태에서 이 수는 0~2다. 10은 그 정상 범위의 다섯 배이고, 넘었다는 것은 우리
// 수명주기가 만든 모양이 아니라는 뜻이다. 상한을 크게 잡을수록 그 이상 상태에서 flock을
// 쥔 채 기다리는 시간이 길어질 뿐이다.
const maxPrivateSocketProbes = 10

// privateSocketProbe는 "물어볼 socket 하나"다: 경로와, **chmod 전에 검증한 그 파일**.
//
// 정보를 함께 들고 다니는 이유는 probe가 chmod 뒤에 같은 파일인지 확인하기 때문이다
// (a109 §2b.3 G7 — privateSocketAccepts).
type privateSocketProbe struct {
	path string
	info os.FileInfo
}

// checkPrivateProbeBudget은 물어볼 socket 수가 상한 안인지 본다.
//
// 넘으면 **아무것도 건드리지 않고** 거부하면서 상한을 넘긴 엔트리의 이름을 지목한다.
// 지목이 필요한 이유는 D3의 회복 수단이 "보고된 원인을 제거한 뒤 재시작"이기 때문이다 —
// 이 거부는 결정적이라 재시작만으로는 벗어날 수 없다.
func checkPrivateProbeBudget(final *privateSocketProbe, staged []privateSocketProbe) error {
	total := len(staged)
	if final != nil {
		total++
	}
	if total <= maxPrivateSocketProbes {
		return nil
	}
	ordered := make([]privateSocketProbe, 0, total)
	if final != nil {
		ordered = append(ordered, *final)
	}
	ordered = append(ordered, staged...)
	return fmt.Errorf(
		"private endpoint: too many stale sockets to probe (%d > %d), starting at: %s",
		total, maxPrivateSocketProbes, filepath.Base(ordered[maxPrivateSocketProbes].path))
}

// ListenStagedPrivateSocket은 socket을 **임시 이름에 bind → 0600 → rename** 순으로
// 발행한다.
//
// 왜 이 의례인가(design D1). `net.Listen`이 만드는 socket 파일의 권한은 umask가 정한다 —
// 컨테이너 실측 umask 077이면 0700이다. 예전 코드는 최종 이름에 바로 bind한 뒤 chmod
// 했으므로, 그 두 줄 사이에서 죽으면 **최종 이름 자리에 0600이 아닌 socket**이 남았다.
// 그리고 회수의 정확-0600 검사가 그것을 "우리 것이 아니다"로 영구 거부했다. 버그가 보안
// 핀으로 봉인돼 있었던 것이다(a108 A1 F3, 두 형제에서 각 3/3 재현).
//
// rename은 같은 디렉터리 안이라 원자적이고, unix socket의 rename은 유효하다: bind는
// 경로가 아니라 inode에 붙으므로 이름이 바뀌어도 그 listener는 계속 수락한다. 그래서
// 최종 이름은 "완성된 socket이 나타나는 순간"에만 존재한다.
//
// `SetUnlinkOnClose(false)`가 design D2b다. Go의 unix listener는 닫힐 때 **자기가 기억하는
// 경로**를 unlink한다. `Shutdown`이 `Serve`의 listener 등록을 앞지르면 그 정리가 Serve
// goroutine의 defer로 밀리고, 늦게 도착한 unlink가 그 경로에 이미 앉아 있는 **후계자의
// socket**을 지운다(a108 A1 F5는 300라운드 중 3회 재현). 여기서 두 겹으로 막는다 —
// listener가 기억하는 이름은 임시 이름이고, unlink 자체를 끈다.
func ListenStagedPrivateSocket(controlDir, socketPath string) (net.Listener, error) {
	name, err := StagedSocketName()
	if err != nil {
		return nil, err
	}
	staged := filepath.Join(strings.TrimSpace(controlDir), name)
	listener, err := net.Listen("unix", staged)
	if err != nil {
		return nil, err
	}
	if unixListener, ok := listener.(*net.UnixListener); ok {
		unixListener.SetUnlinkOnClose(false)
	}
	discard := func() {
		_ = listener.Close()
		_ = os.Remove(staged)
	}
	if err := os.Chmod(staged, 0o600); err != nil {
		discard()
		return nil, err
	}
	if err := os.Rename(staged, socketPath); err != nil {
		discard()
		return nil, err
	}
	return listener, nil
}

// ReclaimStalePrivateEndpoint는 지난 기동이 남긴 잔재를 치우고 디렉터리째 비운다.
//
// 이 함수가 다루는 상태는 endpoint의 기동·종료·회수가 만들 수 있는 **전부**다. 어느
// 지점에서 프로세스가 죽어도 디스크에는 아래 중 하나만 남는다.
//
//	빈 디렉터리    Mkdir 뒤 발행 전에 죽었다 — 검증할 것이 없다
//	descriptor만   종료가 socket을 unlink한 뒤 죽었다(2026-08-13 사고의 모양)
//	socket만       제거 순서(descriptor→socket)의 중간에서 죽었다
//	둘 다          SIGKILL·전원 단절
//	staging만      발행 전 임시 이름으로 죽었다 — 아무도 읽지 않은 우리 잔재다
//
// 빠진 파일은 "덜 만들어진 것"이지 "낯선 것"이 아니다.
//
// 주인이 살아 있는지는 PID가 아니라 socket에 실제로 연결해서 묻는다. 컨테이너를 다시
// 만들면 PID가 거의 같은 자리에 재배정되므로 descriptor에 적힌 PID는 남의 프로세스를
// 가리킬 수 있다(a102 D4b-2) — 그 오판은 살아 있다고 잘못 읽으면 영구 거부를, 죽었다고
// 잘못 읽으면 남의 socket 삭제를 만든다.
//
// # 1차 방어는 이 함수가 아니라 journal flock이다
//
// `cmd/tossctl`의 `runEngineRun`은 endpoint 기동 **전에** journal 디렉터리의 flock을
// 쥔다(부팅 1단계). 그래서 같은 journal을 쓰는 두 번째 엔진은 존재할 수 없고, probe와
// 제거 사이에 새 주인이 들어오는 경합은 그 잠금이 막는다. 여기 있는 probe는 flock이
// 막지 못하는 것 — 엔진이 아닌 프로세스의 경로 점유, 다른 journal 디렉터리의 엔진 —
// 에 대한 **심층 방어**다.
//
// 보안 검사(0700 디렉터리·소유 uid·symlink 금지·hard link·descriptor 형식)는 회수
// 가능한 상태에서도 전부 한다. 넓히는 것은 회수의 범위이지 검증이 아니다.
func ReclaimStalePrivateEndpoint(controlDir string, names PrivateEndpointNames) error {
	if err := names.validate(); err != nil {
		return err
	}
	clean := strings.TrimSpace(controlDir)
	if err := ValidatePrivateControlDirectory(clean); err != nil {
		return fmt.Errorf("private endpoint: existing control directory is unsafe: %w", err)
	}
	entries, err := os.ReadDir(clean)
	if err != nil {
		return err
	}
	// 우리가 만들 수 있는 이름은 최종 둘과 발행 전 임시 이름뿐이다. 하나라도 낯선 것이
	// 있으면 이 디렉터리는 우리 잔재가 아니므로 **아무것도 건드리지 않는다.**
	seen := make(map[string]bool, len(entries))
	staged := make([]string, 0, len(entries))
	stagedSockets := make([]privateSocketProbe, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		switch {
		case names.isFinal(name):
			seen[name] = true
		case names.isStaging(name):
			path := filepath.Join(clean, name)
			info, err := verifyPrivateStagingEntry(path)
			if err != nil {
				return err
			}
			staged = append(staged, path)
			if info.Mode()&os.ModeSocket != 0 {
				stagedSockets = append(stagedSockets, privateSocketProbe{path: path, info: info})
			}
		default:
			// 오류는 **어느 엔트리**인지 말한다(a109 §1-fix F6). 이 거부는 결정적이라
			// 재시작으로는 벗어날 수 없고, 회복은 "보고된 원인을 제거한 뒤 재시작"뿐이다
			// (freeze P0-4). 이름 없는 거부는 운영자에게 "디렉터리 안 무언가"만 말한다.
			return fmt.Errorf(
				"private endpoint: stale control directory has an unexpected entry: %s", name)
		}
	}
	// 주인의 생사를 먼저 판정한다. 사망이 증명되지 않은 채로 다음 줄에 가면 살아 있는
	// 주인의 파일을 지우게 된다.
	socketPath := ""
	var finalSocket *privateSocketProbe
	if names.Socket != "" {
		socketPath = filepath.Join(clean, names.Socket)
		if seen[names.Socket] {
			info, err := verifyStalePrivateSocket(socketPath)
			if err != nil {
				return err
			}
			finalSocket = &privateSocketProbe{path: socketPath, info: info}
		}
	}
	// probe에 상한을 둔다(a109 §2b.3 G8). probe 하나는 최대 privateProbeTimeout이고
	// **순차**이며, 이 함수는 부팅 1단계에서 journal flock을 쥔 채로 돈다 — 잔재 socket이
	// 수십 개면 그 시간이 통째로 엔진 기동 앞에 붙고, 그 동안 보호 루프는 아직 없다.
	//
	// 상한을 넘을 때 하는 일이 **거부**인 이유: 서둘러 지우는 쪽은 probe 없는 unlink이고,
	// 그 방향의 오판은 산 주인의 socket을 지운다. 이 정도로 쌓인 디렉터리는 우리 수명주기가
	// 만들 수 있는 모양이 아니므로(발행은 매 기동 하나씩만 만들고 회수가 그때 비운다),
	// 사람이 보는 것이 맞다 — 그래서 넘긴 엔트리의 이름을 지목한다.
	if err := checkPrivateProbeBudget(finalSocket, stagedSockets); err != nil {
		return err
	}
	if finalSocket != nil && privateSocketAccepts(finalSocket.path, finalSocket.info) {
		return errors.New("private endpoint: the endpoint owner is still alive")
	}
	// staging 이름의 socket도 똑같이 묻는다(a109 §1-fix F5). 이름표가 "우리 잔재"라는
	// 것은 그것이 **죽었다**는 뜻이 아니다 — 발행의 첫 걸음이 바로 staging 이름에
	// bind하는 것이고(ListenStagedPrivateSocket), 그 창에 있는 후계자의 socket이 정확히
	// 이 모양이다. 규칙은 이름이 아니라 사실이다: **수락 중인 socket은 이름과 무관하게
	// 절대 unlink하지 않는다.**
	for _, probe := range stagedSockets {
		if privateSocketAccepts(probe.path, probe.info) {
			return fmt.Errorf(
				"private endpoint: a staging socket is still accepting: %s", filepath.Base(probe.path))
		}
	}
	// 여기 왔다면 주인은 죽었다: socket이 없으면 listener도 없고(발행은 listen이 성공한
	// 뒤에만 descriptor를 쓴다), 있어도 아무도 수락하지 않았다.
	//
	// 그래서 descriptor는 **형식**만 본다. 0바이트·잘린 JSON은 우리 자신의 반쯤 쓴
	// 파일이고 그 내용은 어떤 판정에도 쓰이지 않는다 — 기동은 이전 descriptor를 읽지
	// 않고 rename이 덮는다. 내용 파싱 실패를 거부로 읽으면 그 거부는 순수한 손실이다.
	if seen[names.Descriptor] {
		if err := ValidatePrivateFile(filepath.Join(clean, names.Descriptor)); err != nil {
			return fmt.Errorf("private endpoint: stale descriptor is unsafe: %w", err)
		}
	}
	// 없는 파일을 지우라는 요청은 성공과 같다. 주인이 아직 자기 Close 도중이면 양쪽이
	// 같은 경로를 지우는데, 그 경합은 결과가 같으므로 양성이다. 디렉터리도 같은 규칙을
	// 쓴다 — 파일에만 ErrNotExist를 용인하면 같은 경합이 파일에서는 양성이고
	// 디렉터리에서는 기동 실패가 되는데, 그 비대칭에 근거가 없다.
	//
	// # probe와 이 제거 사이의 창 (a109 §2b.3 G10 — codex 적대 패스의 TOCTOU 지적)
	//
	// 위에서 "죽었다"를 확인한 뒤 여기서 지우기까지의 창에 **새 주인이 그 이름에 bind
	// 하면** 우리는 산 socket을 지운다. 그 창을 닫는 것은 이 함수가 아니라 **journal
	// flock**이다: `runEngineRun`은 endpoint 기동 전에 그것을 쥐고(부팅 1단계) endpoint
	// Close까지 쥔 채로 있으므로(defer LIFO — `lock.Release`가 가장 먼저 등록된다),
	// 같은 journal의 두 엔진이 이 창에서 만날 수 없다. 같은 근거를 a108이 자기 회수에
	// 명문화했다(strategyprojectionrpc/transport_unix.go의 reclaimStaleControlDirectory
	// "1차 방어" 절). 엔진이 아닌 same-uid 프로세스의 점유는 freeze부터 수용된 위협
	// 경계이고, probe는 그 경계 **안에서의** 심층 방어다.
	removals := append(staged, filepath.Join(clean, names.Descriptor))
	if socketPath != "" {
		removals = append(removals, socketPath)
	}
	for _, path := range removals {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("private endpoint: remove stale entry: %w", err)
		}
	}
	if err := os.Remove(clean); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("private endpoint: remove stale control directory: %w", err)
	}
	return nil
}

// verifyStalePrivateSocket은 잔재 socket이 **우리가 만든 모양**인지 보고, 통과하면 그
// 엔트리의 Lstat 정보를 함께 돌려준다.
//
// 정보를 돌려주는 이유는 probe가 chmod **전에 본 파일**을 알아야 하기 때문이다
// (a109 §2b.3 G7 — privateSocketAccepts의 SameFile 재확인).
//
// 권한 검사가 정확-0600이 아니라 "group/other 비트 없음"인 것이 design D2의 좁은
// 완화다. 구버전 바이너리가 listen과 chmod 사이에서 죽으면 umask가 정한 권한(실측
// 077 → 0700)의 socket이 남고, 정확-0600 검사는 **자기 자신이 만든 그 잔재**를 영구
// 거부했다. 완화의 폭은 딱 거기까지다 — group이나 other 비트가 하나라도 있으면 여전히
// 거부다(0640 socket은 지금도 unsafe).
//
// ⛔ 이 완화는 **회수 전용**이다. 클라이언트·발행 확인 경로(ValidatePrivateSocket ·
// ValidateRuntimeSocket)는 정확-0600을 유지한다. 공유 helper(statPrivateSocket)를
// 완화하는 손쉬운 구현은 클라이언트 경계를 함께 넓힌다(freeze P1-3).
//
// 이 완화가 안전한 것은 나머지 조건이 전부 성립할 때뿐이고, 그 전부를 여기와 호출부에서
// 확인한다: 우리 uid 소유 · 0700 디렉터리 안 · symlink 아님 · hard link 없음 · 그리고
// 아무도 수락하지 않음(주인의 사망 입증).
// 거부는 **엔트리 이름과 거부한 절**을 말한다(a109 §2b.3 G6 — §1-fix F6이 다른 두
// 경로에만 적용됐다). 형제 endpoint 셋이 같은 문장을 쓰므로, 이름 없는 한 줄은
// 운영자에게 "어느 디렉터리의 무언가가 이상하다"까지만 말한다. 그리고 D3의 회복 수단은
// "보고된 원인을 제거한 뒤 재시작"이므로(freeze P0-4), 어느 절이 거부했는지 — 지우라는
// 것인지 권한을 고치라는 것인지 — 가 그 안내의 내용이다.
func verifyStalePrivateSocket(socketPath string) (os.FileInfo, error) {
	name := filepath.Base(socketPath)
	info, err := os.Lstat(socketPath)
	if err != nil {
		return nil, fmt.Errorf("private endpoint: stale socket cannot be inspected: %s: %w", name, err)
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return nil, fmt.Errorf("private endpoint: stale socket is a symlink: %s", name)
	case info.Mode()&os.ModeSocket == 0:
		return nil, fmt.Errorf("private endpoint: stale socket is not a socket: %s", name)
	case info.Mode().Perm()&0o077 != 0:
		return nil, fmt.Errorf(
			"private endpoint: stale socket is open to group or other (%#o): %s",
			info.Mode().Perm(), name)
	}
	if err := validateOwnerAndLinks(info, true); err != nil {
		return nil, fmt.Errorf("%w: %s", err, name)
	}
	return info, nil
}

// privateSocketAccepts는 이 socket 경로에서 연결을 받아 주는 자가 있는지 본다.
//
// 발행은 listen이 성공한 뒤에야 descriptor를 쓰므로, 수락하지 않는 socket 파일은 죽은
// 주인의 것이다. 반대로 수락한다면 그게 누구든 이 디렉터리는 남의 것이다.
//
// 죽었다고 읽는 것은 **두 가지뿐**이다: 연결 거부(listener 없음)와 파일 부재. 그 밖의
// 오류(타임아웃 등)는 죽었다는 증거가 아니므로 살아 있다고 본다 — 잘못 살아 있다고 보면
// 이번 기동만 실패하지만, 잘못 죽었다고 보면 남의 socket을 지운다.
//
// # 왜 probe 전에 chmod 하는가 (a109 §1-fix F1)
//
// unix socket에 connect 하려면 그 파일에 쓰기 권한이 있어야 하고, 없으면 커널은 EACCES를
// 준다. 그 오류를 "답이 안 왔다 = 살아 있다"로 읽으면 아무도 없는 socket 하나가 **영구
// 거부**를 만든다 — 이 change가 지우려는 모양 그대로다.
//
// 예전 판은 그래서 "owner 쓰기 비트가 없으면 죽었다"를 세 번째 사망 판정으로 썼다.
// 그것은 묻는 대신 **추정하는** 것이었고, 추정은 틀릴 수 있다: 쓰기 비트가 깎인 socket이
// **수락 중일 수도 있다.** 그때 그 추정은 산 주인의 socket을 지우고 두 번째 서버를
// 그 자리에 세운다(A1 적대 리뷰 P1-A 재현, 두 형제 모두).
//
// 지금은 추정하지 않고 **묻는다**. probe 전에 0600으로 chmod하면 EACCES 자체가 사라지고,
// 산 socket은 수락(→거부·보존)으로 죽은 socket은 ECONNREFUSED(→회수)로 **결정적으로**
// 갈린다.
//
// # 왜 그 chmod가 안전한가 — 그리고 그 안전이 기대는 경계는 무엇인가
//
// ① 두 호출부 모두 chmod 전에 **우리 uid 소유 · symlink 아님 · 0700 디렉터리 안**을
// 확인한다(ValidatePrivateControlDirectory + verifyStalePrivateSocket 또는
// verifyPrivateStagingEntry — 둘 다 Lstat 이므로 symlink는 socket으로 분류되지 않는다).
// 최종 이름 쪽과 staging 쪽 모두 nlink 1을 요구하므로(a109 §2b.3 G7), 검증한 inode에
// **다른 이름**이 걸려 있는 경우도 여기 오지 않는다.
// ② 결과 권한이 **0600뿐**이라 접근이 넓어질 수 없다. 어떤 입력 권한에서 출발하든
// chmod 0600은 owner 전용으로 좁히거나 그대로 둔다. 그리고 그 값은 발행 계약 그 자체다 —
// 최종 socket은 ListenStagedPrivateSocket이 chmod 0600을 지난 뒤에만 rename하므로,
// 이 chmod가 하는 일은 우리 자신의 계약을 되돌리는 것이다.
//
// # 검사한 것과 만지는 것 (a109 §2b.3 G7)
//
// 검사는 **inode**를 보고 chmod·connect는 **이름**을 쓴다. 그 사이에 같은 uid의 다른
// 프로세스가 이름을 갈아끼우면 우리는 다른 파일의 권한을 바꾸고 그 파일에 연결한다.
// 우리 신뢰 경계는 "같은 uid는 신뢰한다"이지 "그런 일이 **불가능**하다"가 아니다 —
// 같은 uid는 이 디렉터리에 이미 쓸 수 있고, 그것이 freeze부터 수용된 위협 경계다.
// 그러니 단정하는 대신 **확인한다**: chmod 뒤 다시 Lstat 해서 검사한 그 파일이 아니면
// 우리는 답을 얻지 못한 것이고, 답을 얻지 못한 것은 **살아 있다**로 읽는다(회수 거부 —
// 보수 방향). before가 없으면(호출자가 못 준 경우) 같은 이유로 살아 있다고 읽는다.
//
// chmod가 실패하면 보수적으로 읽는다: 파일이 없어졌으면 사망(그 자체가 사망 판정 중
// 하나다), 그 밖의 실패는 생존으로 보아 기동을 거부시킨다 — 물어보지 못한 것을 죽었다고
// 읽지 않는다.
func privateSocketAccepts(socketPath string, before os.FileInfo) bool {
	if err := os.Chmod(socketPath, 0o600); err != nil {
		return !errors.Is(err, os.ErrNotExist)
	}
	after, err := os.Lstat(socketPath)
	if err != nil || before == nil || !os.SameFile(before, after) {
		return true
	}
	conn, err := net.DialTimeout("unix", socketPath, privateProbeTimeout)
	if err == nil {
		_ = conn.Close()
		return true
	}
	if errors.Is(err, unix.ECONNREFUSED) || errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

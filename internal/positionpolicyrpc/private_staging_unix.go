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
	stagedSockets := make([]string, 0, len(entries))
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
				stagedSockets = append(stagedSockets, path)
			}
		default:
			return errors.New("private endpoint: stale control directory has unexpected entries")
		}
	}
	// 주인의 생사를 먼저 판정한다. 사망이 증명되지 않은 채로 다음 줄에 가면 살아 있는
	// 주인의 파일을 지우게 된다.
	socketPath := ""
	if names.Socket != "" {
		socketPath = filepath.Join(clean, names.Socket)
		if seen[names.Socket] {
			if err := verifyStalePrivateSocket(socketPath); err != nil {
				return err
			}
			if privateSocketAccepts(socketPath) {
				return errors.New("private endpoint: the endpoint owner is still alive")
			}
		}
	}
	// staging 이름의 socket도 똑같이 묻는다(a109 §1-fix F5). 이름표가 "우리 잔재"라는
	// 것은 그것이 **죽었다**는 뜻이 아니다 — 발행의 첫 걸음이 바로 staging 이름에
	// bind하는 것이고(ListenStagedPrivateSocket), 그 창에 있는 후계자의 socket이 정확히
	// 이 모양이다. 규칙은 이름이 아니라 사실이다: **수락 중인 socket은 이름과 무관하게
	// 절대 unlink하지 않는다.**
	for _, path := range stagedSockets {
		if privateSocketAccepts(path) {
			return errors.New("private endpoint: a staging socket is still accepting")
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

// verifyStalePrivateSocket은 잔재 socket이 **우리가 만든 모양**인지 본다.
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
func verifyStalePrivateSocket(socketPath string) error {
	info, err := os.Lstat(socketPath)
	if err != nil || info.Mode()&os.ModeSocket == 0 || info.Mode()&os.ModeSymlink != 0 ||
		info.Mode().Perm()&0o077 != 0 {
		return errors.New("private endpoint: stale socket is unsafe")
	}
	return validateOwnerAndLinks(info, true)
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
// # 왜 그 chmod가 안전한가
//
// ① 이 지점은 이미 검증을 통과했다 — 우리 uid 소유, group/other 비트 없음, symlink 아님,
// 0700 디렉터리 안(호출부의 ValidatePrivateControlDirectory · verifyStalePrivateSocket ·
// verifyPrivateStagingEntry). 즉 남의 파일에 chmod하는 경로가 없다.
// ② 발행 계약상 최종 socket은 **원래 0600이다**(ListenStagedPrivateSocket이 chmod 0600
// 뒤에만 rename한다). 그러니 이 chmod가 하는 일은 우리 자신의 발행 계약을 되돌리는
// 것이고, 결과 권한은 어떤 경우에도 **owner 전용**이라 접근이 넓어지지 않는다.
//
// chmod가 실패하면 보수적으로 읽는다: 파일이 없어졌으면 사망(그 자체가 사망 판정 중
// 하나다), 그 밖의 실패는 생존으로 보아 기동을 거부시킨다 — 물어보지 못한 것을 죽었다고
// 읽지 않는다.
func privateSocketAccepts(socketPath string) bool {
	if err := os.Chmod(socketPath, 0o600); err != nil {
		return !errors.Is(err, os.ErrNotExist)
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

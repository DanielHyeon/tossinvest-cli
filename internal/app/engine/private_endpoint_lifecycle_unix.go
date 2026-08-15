//go:build unix

package engine

// private_endpoint_lifecycle_unix.go — socket을 발행하는 두 형제 endpoint가 **글자 그대로
// 같게** 하던 두 블록이다(a109 §2b.3 G9).
//
// 왜 모으는가: 이 저장소의 규율은 "복사한 검사는 어긋나기 시작한 검사다"(a098 design
// D7.1)이고, a109 자신이 그 근거로 회수 기계를 positionpolicyrpc 한 곳에 두었다. 그런데
// 그 기계를 **부르는 의례**는 두 벌로 복사돼 있었다 — 회수를 한 곳에 두는 이유가 그
// 부름에는 적용되지 않을 근거가 없다.
//
// 두 블록 다 기계적으로 옮겼다. 오류 문자열까지 그대로 나오도록 label 하나만 받는다 —
// 문구가 바뀌면 그것은 리팩터링이 아니라 동작 변경이고, 운영자가 읽는 것은 그 문구다.

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// openReclaimedControlDirectory는 control 디렉터리를 **이번 기동의 것으로** 만든다.
//
// 디렉터리가 이미 있다는 것은 지난 기동의 잔재가 있다는 뜻이다. 회수는 전체성이다 —
// 자기 수명주기가 만들 수 있는 부분 상태를 전부 치우고 디렉터리째 비운 뒤 새로 만든다.
// **수락 중인 socket은 최종 이름이든 staging 이름이든 지우지 않는다**(design D2 +
// a109 §1-fix F5). 생사는 권한 비트로 추정하지 않고 연결해서 묻는다.
//
// 성공하면 호출자는 "이 디렉터리는 이번 기동이 만든 것"을 전제로 실패 정리를 조건 없이
// 할 수 있다.
func openReclaimedControlDirectory(controlDir, label string,
	names positionpolicyrpc.PrivateEndpointNames) error {
	if err := os.Mkdir(controlDir, 0o700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("engine: creating %s: %w", label, err)
		}
		if err := positionpolicyrpc.ReclaimStalePrivateEndpoint(controlDir, names); err != nil {
			return fmt.Errorf("engine: reclaiming %s: %w", label, err)
		}
		if err := os.Mkdir(controlDir, 0o700); err != nil {
			return fmt.Errorf("engine: recreating %s: %w", label, err)
		}
	}
	return nil
}

// closePrivateEndpointFiles는 listener를 닫고 이 endpoint의 경로를 지운다.
//
// listener는 **여기서 우리가** 닫는다. Shutdown이 Serve의 listener 등록을 앞지르면 Go는
// 정리를 Serve goroutine의 defer로 미루는데, 그 늦은 정리가 도착할 때 경로에는 이미
// 후계자의 socket이 앉아 있을 수 있다(a108 A1 F5는 300라운드 중 3회 재현). 닫는 시점을
// 예정된 goroutine에 맡기지 않는다. 이미 Shutdown이 닫았으면 net.ErrClosed가 오는데,
// 그것은 성공과 같다.
//
// 경로를 지울 권한은 이 루프 하나뿐이다 — listener는 자기 이름을 unlink하지 않고
// (SetUnlinkOnClose(false)), 기억하는 이름도 이미 사라진 임시 이름이다.
//
// 먼저 난 오류를 유지한다(result가 nil일 때만 채운다): Shutdown의 실패가 첫 소식이면
// 그것이 보고돼야 하고, 뒤따르는 정리 실패는 그 결과다.
func closePrivateEndpointFiles(result error, listener net.Listener, paths ...string) error {
	if listener != nil {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) &&
			result == nil {
			result = err
		}
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) && result == nil {
			result = err
		}
	}
	return result
}

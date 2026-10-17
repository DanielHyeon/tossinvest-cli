//go:build unix

package strategyprojectionrpc

// transport_probe_unix.go — a113: 회수는 주인의 사망을 **추정하지 않고 증명한다**.
//
// # 왜 새 파일인가
//
// 이 함수를 transport_unix.go 에 두면 logic-map 게이트가 삽입 지점의 이웃 함수까지
// 「수정됨」으로 보고 아무도 편집하지 않은 함수의 증거를 요구한다(a102·a108·a109 가 배운 것).
//
// # 왜 `projectionSocketAccepts` 에 chmod 를 넣지 않는가 (design D1)
//
// 형제(a109 `privateSocketAccepts`)는 probe 를 회수만 부르므로 그 함수 안에 chmod 를 넣었다.
// 이 패키지의 probe 는 조회 클라이언트 `Dial` 도 부른다. 거기에 chmod 를 넣으면 콘솔·조회
// 데몬이 **엔진 socket 의 권한을 바꾸는** 부작용이 생긴다. 그래서 순수 질문은 그대로 두고
// (추정 절만 지웠다), 권한 복원은 회수가 부르는 이 함수 하나에만 둔다.

import (
	"errors"
	"os"
)

// chmodStaleSocket 은 회수 probe 가 권한을 되돌리는 호출이다. 운영 값은 `os.Chmod` 하나다.
//
// package 변수인 이유는 **chmod 와 connect 사이**를 테스트가 밟을 수 있어야 하기 때문이다
// (a113 post-review P1-1). 그 사이에 이름이 바뀌는 경합, chmod 가 부재·권한 오류로 실패하는
// 경우는 디스크만으로 결정적으로 만들 수 없고, 못 만드는 분기는 지워도 초록이다(원장 R1·R3).
var chmodStaleSocket = os.Chmod

// staleProjectionSocketAccepts 는 **회수 전용** 생존 질문이다 — a109 §1-fix F1·§2b.3 G7 원형판.
//
// 사망으로 읽는 것은 둘뿐이다: 연결 거부와 파일 부재. 예전 판정은 「owner 쓰기 비트가 없으면
// 죽었다」를 셋째로 썼는데, 그것은 묻는 대신 추정한 것이었고 틀릴 수 있다 — 쓰기 비트가
// 외부 chmod 로 깎인 socket 도 **수락 중일 수 있다.** 그 추정은 산 주인의 socket 을 지우고
// 그 자리에 두 번째 서버를 세운다(a109 A1 P1-A 재현, a109 issues I1).
//
// 지금은 묻는다. probe 전에 0600 으로 chmod 하면 EACCES 자체가 사라지고, 산 socket 은
// 수락(→생존·보존)으로, 죽은 socket 은 ECONNREFUSED(→회수)로 **결정적으로** 갈린다.
//
// # 왜 그 chmod 가 안전한가
//
// ① 회수가 이 함수에 닿기 전에 control 디렉터리(정확 0700·우리 uid·비symlink)와 socket
// (우리 uid·비symlink·group/other 비트 없음·nlink 1)을 검증했다(`reclaimStaleControlDirectory`
// B1·B2, `verifyStaleSocketShape`). ② 결과 권한이 0600 뿐이라 접근이 넓어질 수 없다. 그리고
// 0600 은 발행 계약 그 자체다 — 최종 socket 은 `listenPrivateSocket` 이 chmod 0600 을 지난
// 뒤에만 rename 된다.
//
// # 검사한 것과 만지는 것 (G7)
//
// 검사는 inode 를 보고 chmod·connect 는 **이름**을 쓴다. 그 사이 같은 uid 의 다른 프로세스가
// 이름을 갈아끼우면 우리는 다른 파일을 만진다. 그래서 chmod **앞뒤로** 다시 Lstat 해서 검증한
// 그 파일(`before`)인지 확인한다. 앞의 확인은 검증하지 않은 inode 를 chmod 하는 창을 좁히고
// (a113 freeze P1-2), 뒤의 확인은 chmod 한 그 파일에 연결한다는 것을 보증한다. 다른 파일이면
// 답을 얻지 못한 것이고, 답을 얻지 못한 것은 **생존**으로 읽는다(회수 거부 — 보수 방향).
// `before` 가 없으면 검증하지 않은 이름이므로 아무것도 만지지 않고 생존으로 읽는다.
//
// 부재는 어느 자리에서 드러나든 사망이다(주인이 자기 Close 로 지운 순간과 같다). 그 밖의
// 실패는 생존 — 물어보지 못한 것을 죽었다고 읽지 않는다.
//
// `os.Chmod` 는 symlink 를 따라간다. 앞의 확인 뒤에 이름이 symlink 로 바뀌면 뒤의 확인 전에
// 대상이 chmod 될 수 있다 — 같은 uid 신뢰 경계 안이며 원형이 수용한 것과 같다(issues R2).
func staleProjectionSocketAccepts(socketPath string, before os.FileInfo) bool {
	if before == nil {
		return true
	}
	if dead, answered := sameVerifiedSocket(socketPath, before); !answered {
		return !dead
	}
	if err := chmodStaleSocket(socketPath, 0o600); err != nil {
		return !errors.Is(err, os.ErrNotExist)
	}
	if dead, answered := sameVerifiedSocket(socketPath, before); !answered {
		return !dead
	}
	return projectionSocketAccepts(socketPath)
}

// sameVerifiedSocket 은 그 이름이 아직 검증한 파일인지 다시 본다.
// answered 가 false 면 더 묻지 않고 끝낸다: dead 는 이름이 사라진 경우(사망)만 참이다.
func sameVerifiedSocket(socketPath string, before os.FileInfo) (dead, answered bool) {
	now, err := os.Lstat(socketPath)
	if err != nil {
		return errors.Is(err, os.ErrNotExist), false
	}
	return false, os.SameFile(before, now)
}

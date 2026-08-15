package positionpolicyrpc

// private_staging.go는 **아직 발행되지 않은** 산출물의 이름과, 그 이름으로 죽은 잔재의
// 위생을 다룬다. 이름-독립이다 — 어느 endpoint의 것인지는 호출자가 안다.
//
// # 왜 이 패키지인가
//
// 여기 있는 검사(소유 uid·hard link·모양)는 private_endpoint.go의 자기 문서가 말하는
// 보안 원시를 그대로 재사용한다. 두 번째 패키지로 복사하면 그것은 어긋나기 시작한
// 검사다(a098 design D7.1). 그래서 기계를 옮기지 않고 이름을 받는다.

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// StagingPrefix는 발행 전 임시 이름의 앞머리다.
//
// socket도 descriptor도 임시 이름에 먼저 만들고 rename으로 최종 이름을 준다. 그래서
// 최종 이름이 보이면 그것은 항상 완성된 산출물이다 — 반쯤 만들어진 것이 최종 이름을
// 가지는 순간이 없다. 대신 임시 이름으로 죽는 새 잔재가 생기는데, 그것은 낯선 것이
// 아니라 **우리 잔재**이므로 회수 목록에 들어간다.
const StagingPrefix = ".s-"

// 임시 이름의 종류 한 글자. 사람이 잔재를 볼 때 무엇이 되려다 만 것인지 알려 준다.
const stagingSocketKind = "s"

// stagingHexDigits는 임시 이름 뒤에 붙는 난수 hex의 길이다.
//
// # 왜 홀수인가
//
// bind 하는 이름은 최종 이름이 아니라 **이 임시 이름**이고, unix socket 경로에는
// sun_path 상한이 있다(실측 107자까지 bind 된다). 임시 이름이 최종 이름보다 길면
// 최종 경로는 상한 안인데 발행만 죽는 배포가 생기고, 운영자에게 보이는 것은 「매 부팅
// 그 표면이 없다」뿐이라 원인이 보이지 않는다.
//
// 형제 중 가장 짧은 최종 socket 이름은 `alerts.sock` 11자다. `.s-` 3 + 종류 1 + hex 7 =
// **11자**로 그 이하를 맞춘다. a108(strategy projection)의 12자는 그 endpoint의 최종
// 이름이 12자라 적법하고, 여기서 통일하지 않는다 — 그것은 a108 소유의 확정 코드다.
//
// hex 7자는 `hex.EncodeToString`이 바로 내지 못한다(짝수 길이만 낸다). 4바이트 난수의
// hex 8자를 7자로 **절단한다**. 엔트로피 28비트이고, 겹치면 bind가 EADDRINUSE로
// 실패하지만 회수가 방금 비운 디렉터리라 겹칠 일이 없다 — 겹치지 않는 것에 기대지
// 않으려고 난수를 쓸 뿐이다.
const stagingHexDigits = 7

// StagedSocketName은 socket 발행에 쓸 임시 basename 하나를 만든다.
//
// 길이를 계약으로 잡는 테스트가 **이 함수의 산출물을 직접 센다** — 상수를 읽어서
// 계산하면 재는 것은 계산이지 이름이 아니다.
func StagedSocketName() (string, error) {
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}
	return StagingPrefix + stagingSocketKind + hex.EncodeToString(suffix)[:stagingHexDigits], nil
}

// PrivateEndpointNames는 한 endpoint가 자기 control 디렉터리에 만들 수 있는 이름의
// **전부**다. 회수는 이 집합 밖의 이름을 낯선 것으로 본다.
//
// 하나라도 빠지면 낯선-엔트리 거부가 우리 자신의 잔재를 거부하고, 그 거부는 결정적이라
// 매 부팅 같은 자리에서 멈춘다. 그래서 집합의 완전성을 호출자 쪽 테스트가 고정한다
// (각 발행 경로가 실제로 만드는 이름이 그 endpoint가 넘기는 집합에 속하는가).
type PrivateEndpointNames struct {
	// Descriptor는 최종 descriptor 파일 이름이다(예: endpoint.json).
	Descriptor string
	// Socket은 최종 socket 파일 이름이다. socket을 발행하지 않는 endpoint는 빈 문자열.
	Socket string
	// StagingPrefixes는 발행 전 임시 이름의 앞머리 전부다 — StagingPrefix와,
	// descriptor 발행이 os.CreateTemp에 넘기는 현행 접두를 포함한다.
	StagingPrefixes []string
}

func (n PrivateEndpointNames) validate() error {
	if strings.TrimSpace(n.Descriptor) == "" {
		return errors.New("private endpoint: the descriptor name is required")
	}
	for _, prefix := range n.StagingPrefixes {
		if prefix == "" {
			// 빈 접두는 모든 이름을 staging으로 만든다. 즉 낯선 엔트리 거부가 통째로
			// 사라지고 회수가 남의 파일을 치우게 된다.
			return errors.New("private endpoint: a staging prefix cannot be empty")
		}
	}
	return nil
}

func (n PrivateEndpointNames) isFinal(name string) bool {
	return name == n.Descriptor || (n.Socket != "" && name == n.Socket)
}

func (n PrivateEndpointNames) isStaging(name string) bool {
	return hasAnyPrefix(name, n.StagingPrefixes)
}

// Knows는 이 이름이 우리가 만들 수 있는 이름인지 답한다.
//
// 회수의 분류와 **같은 두 술어**로 이루어진다(isFinal · isStaging). 완전성 테스트가
// 자기 판정을 따로 쓰면 재는 것은 테스트의 판정이지 회수의 판정이 아니다 — 그래서
// 술어를 노출하고 회수가 그것을 쓴다.
func (n PrivateEndpointNames) Knows(name string) bool {
	return n.isFinal(name) || n.isStaging(name)
}

func hasAnyPrefix(name string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if prefix != "" && strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// verifyPrivateStagingEntry는 임시 이름의 잔재가 **우리가 만들 수 있는 모양**인지 본다.
//
// 최종 이름을 가진 적이 없는 산출물이므로 아무도 그것을 읽거나 연결하지 않았다 —
// **내용**은 검증할 것이 없다. 다만 **모양**은 본다: 이름만 보고 지우면 우리가 임시
// 이름으로 만들 수 없는 모양(디렉터리 등)까지 지우려 든다. 비어 있으면 남의 디렉터리를
// 지우고, 비어 있지 않으면 ENOTEMPTY로 회수 전체가 매 부팅 멈춘다.
//
// 소유 uid까지 보는 이유(freeze P1-7③): a108은 자기 접두 하나만 다뤘지만 여기는
// `.endpoint-` 같은 일반적 접두까지 넓힌다. 회수의 범위를 넓히면서 검사 부재의 폭까지
// 함께 넓히지 않는다.
// 통과하면 그 엔트리의 Lstat 정보를 함께 돌려준다. 호출부가 **socket 모양인지** 알아야
// 하기 때문이다 — 수락 중인 socket은 이름이 무엇이든 지우지 않는다(a109 §1-fix F5).
func verifyPrivateStagingEntry(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil || !(info.Mode().IsRegular() || info.Mode()&os.ModeSocket != 0) {
		return nil, errors.New("private endpoint: stale staging entry has an unexpected shape")
	}
	if err := validateOwnerAndLinks(info, false); err != nil {
		return nil, err
	}
	return info, nil
}

// SweepPrivateStagingLeftovers는 socket을 발행하지 **않는** endpoint의 위생이다.
//
// 자기 staging 잔재만 치우고 그 밖의 것은 보지도 않는다. 오류를 돌려주지 않는 것이
// 계약이다(design D2a): 이 endpoint에 열거+거부를 넣으면 이물 하나가 격리 해제 표면을
// 매 부팅 지우는 **새 실패 경로**가 생긴다. 격리 해제는 격리된 포지션의 손절 포함
// 미판정 상태를 푸는 유일한 장중 경로이므로, 위생이 그 표면을 없애서는 안 된다.
//
// 디렉터리를 읽지 못하거나 잔재 하나를 지우지 못하면 그냥 남긴다 — 다음 부팅이 다시
// 시도한다.
//
// ⛔ **범위의 정직한 서술(a109 §1-fix F5)**: 회수 쪽(ReclaimStalePrivateEndpoint)은
// staging 이름의 socket에도 connect probe를 걸어 수락 중이면 지우지 않는다. 이 위생에는
// 그 probe가 없다 — 이 함수를 쓰는 유일한 endpoint(policy command)는 loopback TCP라
// socket을 **발행하지 않고**, 그래서 자기 접두의 socket을 만들 수 있는 경로가 없기
// 때문이다. socket을 발행하는 endpoint에서 이 함수를 쓰려면 probe를 먼저 들여야 한다.
func SweepPrivateStagingLeftovers(controlDir string, prefixes []string) {
	entries, err := os.ReadDir(strings.TrimSpace(controlDir))
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if !hasAnyPrefix(name, prefixes) {
			continue
		}
		path := filepath.Join(strings.TrimSpace(controlDir), name)
		if _, err := verifyPrivateStagingEntry(path); err != nil {
			continue
		}
		_ = os.Remove(path)
	}
}

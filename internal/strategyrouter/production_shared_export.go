package strategyrouter

import (
	"os"
	"sort"
	"time"
)

// a112 7.3.1 SHADOW(브리프 v3.3 §3) — shadow 적재기(internal/strategyshadow)가 쓰는 **읽기 전용 중립 wrapper**.
//
// 왜 wrapper 인가: shadow 매니페스트는 활성화 매니페스트와 같은 파일 규칙(0400 · 소유자 · 크기 상한 · 정규 시각 · 식별자 · digest)과 같은
// 서술자 표를 써야 한다. 그 규칙을 옮겨 적으면 두 사본이 갈릴 수 있고, 기존 함수를 고치면 활성화 적재기(주문 경로의 판정)가 흔들린다.
// 그래서 기존 함수는 그대로 두고(무편집 — 편집 전후 본문 digest 대조), 그 함수를 부르기만 하는 한 줄 wrapper 를 이 새 파일에만 둔다.
// 쓰기 wrapper 는 없다 — 이 파일의 export 목록과 각 본문의 호출은 a112_shared_export_test.go 가 이름으로 고정한다.
//
// 읽기 함수의 오류는 ErrProductionRouteUnavailable 을 그대로 돌려준다. 받는 쪽(shadow 적재기)은 그 오류를 `%v` 로 접어 자기 sentinel 하나로만
// 감싼다 — 둘째 `%w` 는 오류가 두 신원을 갖게 한다(8.5 보이스 2 P2-1 과 같은 규칙).

// SharedLaneDescriptor 는 서술자 표 한 행의 값 사본이다(표 자체는 productionRouteDescriptors 하나).
type SharedLaneDescriptor struct {
	Family      Family
	Horizon     Horizon
	LaneID      string
	LaneVersion string
}

// SharedProductionRouteOwnerUID 는 매니페스트 파일이 가져야 할 소유자(현재 UID)다.
func SharedProductionRouteOwnerUID() (uint32, bool) { return productionRouteOwnerUID() }

// SharedReadProductionRouteFile 은 정규 파일 · 모드 · 소유자 · 크기 상한 · 열기 전후 동일성을 지킨 읽기다.
func SharedReadProductionRouteFile(path string, owner uint32, mode os.FileMode, maximum int64) ([]byte, error) {
	return readProductionRouteFile(path, owner, mode, maximum)
}

// SharedProductionRouteDigest 는 바이트의 `sha256:<hex>` digest 다.
func SharedProductionRouteDigest(data []byte) string { return productionRouteDigest(data) }

// SharedProductionRouteTime 은 UTC RFC3339Nano 정규 시각만 받는다.
func SharedProductionRouteTime(raw string) (time.Time, bool) { return productionRouteTime(raw) }

// SharedProductionRouteIdentity 는 식별자 규칙(비지 않음 · 앞뒤 공백 없음 · 256 바이트 · 제어 문자 없음)이다.
func SharedProductionRouteIdentity(value string) bool { return productionRouteIdentity(value) }

// SharedProductionRouteDigestValid 는 `sha256:` + 소문자 hex 64 자리 규칙이다.
func SharedProductionRouteDigestValid(value string) bool { return productionRouteDigestValid(value) }

// SharedProductionRouteDescriptors 는 그 시장의 서술자 표를 가족 순으로 돌려준다. 표를 복사하지 않고 같은 함수에서 읽어 행을 옮길 뿐이다.
func SharedProductionRouteDescriptors(market Market) []SharedLaneDescriptor {
	table := productionRouteDescriptors(market)
	if len(table) == 0 {
		return nil
	}
	lanes := make([]SharedLaneDescriptor, 0, len(table))
	for laneID, descriptor := range table {
		lanes = append(lanes, SharedLaneDescriptor{Family: descriptor.Family, Horizon: descriptor.Horizon, LaneID: laneID,
			LaneVersion: descriptor.LaneVersion})
	}
	// 가족 순 — map 순회 순서에 기대면 같은 표가 실행마다 다른 순서로 나간다.
	sort.Slice(lanes, func(i, j int) bool { return lanes[i].Family < lanes[j].Family })
	return lanes
}

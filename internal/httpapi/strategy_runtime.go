package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

type StrategyRuntimeReader interface {
	Read(context.Context) (strategyprojection.Snapshot, error)
}

// StrategyRuntimePresence 는 reader 가 「이 배포에 전략 화면이 설정돼 있는가」를 스스로
// 말하는 선택적 능력이다 — a109 D4.
//
// # 왜 필요한가
//
// 이 패키지의 소비자들은 그 질문에 **nil 검사**로 답해 왔다: reader 자리가 비었으면
// dormant(NOT_CONFIGURED, 「이 배포는 전략 화면을 안 쓴다」), 있는데 못 읽으면
// unavailable(RUNTIME_UNAVAILABLE, 「엔진이 없다」). 두 사실을 한 값으로 접으면
// 운영자는 화면에서 둘을 구별할 수 없다(a108 D4-2).
//
// a109 는 엔진이 돌아오면 다시 붙는 wrapper 를 그 자리에 꽂는데, wrapper 는 **정의상
// non-nil** 이라 nil 검사가 영원히 거짓이 된다. 그러면 화면이 dormant → unavailable 로
// 회귀한다 — a108 이 금지한 접힘의 같은 모양이다. 그래서 부재를 값의 nil 이 아니라
// **상태 신호**로 묻는다.
//
// # ⚠ 이 술어에는 부작용이 있다 — a109 §2-fix F8 (A2 P2-7)
//
// 구현이 이 질문을 받고 **백그라운드 재부착 시도를 깨울 수 있다**
// (`cmd/tossctl` 의 `strategyRuntimeAttachment`). 그것이 의도다: 부재 상태에서
// 요청 경로가 endpoint 를 만지는 지점이 여기뿐이라, 깨우기를 `Read` 에만 두면
// 냉부팅 순서 — 엔진이 나중에 뜨는 경우 — 가 영원히 회복하지 못한다.
//
// 따라서 **잠금 안이나 hot loop 에서 부르지 마라.** 구현은 막히지 않도록 쓰여
// 있지만(깨우기는 goroutine 하나를 띄우고 즉시 돌아온다), 이 계약을 모르는 호출자가
// 화면 한 번에 이 질문을 수십 번 하면 그만큼 시도 창을 두드리는 것이 된다.
type StrategyRuntimePresence interface {
	StrategyRuntimeConfigured() bool
}

// StrategyRuntimeAbsent 는 「reader 자리가 비었는가」의 **유일한** 판정이다.
//
// ⛔ 이 검사를 소비자마다 다시 쓰지 마라. 오늘 이 질문을 하는 곳은 네 곳이다 —
// router 의 REST 경로, 이 파일의 SSE helper, cmd/tossctl 의 집계 스냅샷, 그리고
// 같은 패키지의 publisher 루프(재부착을 깨우려고 값 없이 부른다).
// 여러 곳에 여러 벌을 두면 갈라지고, 갈라지면 같은 디스크 상태가 화면마다 다른 값이
// 된다 (복사한 검사는 어긋나기 시작한 검사다 — a098 D7.1).
//
// ⚠ **부작용 있는 술어다**: `StrategyRuntimePresence` 를 구현한 reader 는 이 호출로
// 백그라운드 재부착 시도를 깨울 수 있다(위 인터페이스 주석). 잠금 안·hot loop 에서
// 부르지 마라.
func StrategyRuntimeAbsent(reader StrategyRuntimeReader) bool {
	if reader == nil {
		return true
	}
	if presence, ok := reader.(StrategyRuntimePresence); ok {
		return !presence.StrategyRuntimeConfigured()
	}
	// 상태를 말하지 않는 reader 는 「있다」다. 오늘의 스텁·client 가 전부 그쪽이고,
	// 그것이 nil 검사 시절의 뜻과 같다.
	return false
}

func StrategyRuntimeSnapshotFunc(reader StrategyRuntimeReader, now func() time.Time) SnapshotFunc {
	return func(ctx context.Context) ([]byte, error) {
		if StrategyRuntimeAbsent(reader) || now == nil {
			return nil, errors.New("httpapi: strategy runtime stream reader unavailable")
		}
		snapshot, err := reader.Read(ctx)
		if err != nil {
			return nil, err
		}
		if err := strategyprojection.Validate(snapshot); err != nil {
			return nil, err
		}
		return json.Marshal(Envelope{SchemaVersion: SchemaVersion, Resource: "strategy-runtime",
			GeneratedAt: now().UTC(), Data: strategyprojection.Clone(snapshot)})
	}
}

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
type StrategyRuntimePresence interface {
	StrategyRuntimeConfigured() bool
}

// StrategyRuntimeAbsent 는 「reader 자리가 비었는가」의 **유일한** 판정이다.
//
// ⛔ 이 검사를 소비자마다 다시 쓰지 마라. 오늘 이 질문을 하는 곳은 세 곳이다 —
// router 의 REST 경로, 이 파일의 SSE helper, 그리고 cmd/tossctl 의 집계 스냅샷.
// 세 곳에 세 벌을 두면 갈라지고, 갈라지면 같은 디스크 상태가 화면마다 다른 값이 된다
// (복사한 검사는 어긋나기 시작한 검사다 — a098 D7.1).
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

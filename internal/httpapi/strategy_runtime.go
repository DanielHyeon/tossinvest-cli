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

// StrategyRuntimePresence 는 reader 가 「이 배포에 전략 화면이 설정돼 있는가」를 스스로 말하는
// 선택적 능력이다 — a109 D4. a115 로 정의는 `strategyprojection` 으로 옮겼고 여기는 **alias** 다:
// 두 이름이 한 타입이므로 기존 구현·`var _` 결속이 그대로 선다. 부작용 경고는 원본 주석에 있다.
type StrategyRuntimePresence = strategyprojection.StrategyRuntimePresence

// StrategyRuntimeAbsent 는 「reader 자리가 비었는가」다 — 판정은 `strategyprojection.StrategyRuntimeAbsent`
// **한 벌**이고 여기는 위임이다(a115 design D2). 이름을 남기는 이유는 기존 네 소비처(router REST·SSE
// helper·집계 스냅샷·publisher)와 시험을 무편집으로 세우기 위해서다.
//
// ⛔ 여기에 판정을 다시 쓰지 마라 — 콘솔도 같은 원본을 쓴다. 갈라지면 같은 디스크 상태가 화면마다
// 다른 값이 된다. ⚠ 부작용 있는 술어다(재부착 시도를 깨운다) — 잠금 안·hot loop 에서 부르지 마라.
func StrategyRuntimeAbsent(reader StrategyRuntimeReader) bool {
	return strategyprojection.StrategyRuntimeAbsent(reader)
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

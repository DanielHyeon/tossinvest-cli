package obs

// heartbeat.go is the liveness publish (harden-execution-base task 4.3,
// engine-safety: "죽은 프로세스는 스스로 통지할 수 없으므로 heartbeat … 방식을
// 사용한다").
//
// # 생산 배선 없음 — 생존 판정은 다른 경로가 함 (a134)
//
// 이 타입은 생산 생성자가 0 임(비시험 코드에 이 타입의 복합 리터럴 없음). 예전 이 자리의 설명 —
// 엔진이 ntfy 로 박동을 보내면 「ntfy 가 박동 부재에 스스로 경보한다」— 은 근거가 없었음: ntfy 는
// 받은 메시지를 전달할 뿐 오지 않은 메시지를 판정하지 않음(a134 design D0). 수신 측이 아무것도
// 판정하지 않는 박동은 소음일 뿐임.
//
// 실제 생존 경로(a134): 엔진은 저널 디렉터리의 활성 마커(engine-run.json, internal/enginelock)를
// 1분마다 원자적으로 다시 쓰고, 엔진 밖 호스트 타이머가 매 분 부르는 `tossctl engine heartbeat-check`
// (internal/enginewatch)가 그 마커의 침묵을 판정해 ntfy 로 한 번 알림. 침묵을 알아채는 책임은
// 여전히 살아 있는 쪽(엔진 밖)에 있음 — 아래 단락의 원칙은 그 감시자에도 그대로 적용됨.
//
// # Why a heartbeat is best-effort and its absence is not
//
// A missed publish is not an incident: the next one lands and the receiver's
// window is longer than one interval. Escalating a single failed heartbeat into a
// critical alert would make the liveness signal the noisiest thing in the system,
// which is how liveness signals get muted.

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// DefaultHeartbeatInterval is how often the engine says it is alive.
//
// A minute is short enough that an operator learns about a dead engine inside the
// window where a position can still be managed, and long enough to be nothing
// against the query budget (§0.4 — it is not a broker call at all).
const DefaultHeartbeatInterval = time.Minute

// Heartbeat publishes a periodic liveness beat.
//
// 생산 배선 없음(파일 주석) — 엔진 생존 판정은 enginelock 마커 + internal/enginewatch 가 함(a134).
// 삭제는 후속 정리 후보.
type Heartbeat struct {
	// Publisher sends the beat. Required; nil makes Run a no-op.
	Publisher Publisher
	// Log records each beat and each failure. Optional.
	Log *Logger
	// Clock drives the interval. Defaults to clock.System().
	Clock clock.Clock
	// Interval is the publish period. Zero uses DefaultHeartbeatInterval.
	Interval time.Duration
	// Status is called for each beat to describe the engine's condition. Optional;
	// nil sends an empty body.
	Status func() string
}

// Beat publishes one heartbeat.
//
// It reports the publish error for a caller that wants it (the tests do), but
// nothing in the engine treats it as an incident — see the file comment.
func (h *Heartbeat) Beat(ctx context.Context) error {
	if h.Publisher == nil {
		return nil
	}
	body := ""
	if h.Status != nil {
		body = h.Status()
	}
	err := h.Publisher.Publish(ctx, Notification{
		Type:     EventEngineHeartbeat,
		Severity: SeverityNormal,
		Title:    "tossos engine alive",
		Body:     body,
		// Low priority: this one is meant to arrive silently and be noticed only
		// by its absence.
		Priority: 1,
	})
	if h.Log != nil {
		if err != nil {
			h.Log.Warn(EventEngineHeartbeat, FieldError, err.Error())
		} else {
			h.Log.Event(EventEngineHeartbeat)
		}
	}
	return err
}

// Run beats until the context ends.
//
// The first beat is immediate. An engine that starts, publishes nothing for a
// minute and then dies would otherwise be indistinguishable from one that never
// started — and the receiver's window would not have opened yet.
func (h *Heartbeat) Run(ctx context.Context) error {
	if h.Publisher == nil {
		<-ctx.Done()
		return ctx.Err()
	}
	clk := h.Clock
	if clk == nil {
		clk = clock.System()
	}
	interval := h.Interval
	if interval <= 0 {
		interval = DefaultHeartbeatInterval
	}

	for {
		_ = h.Beat(ctx)
		if err := clk.Sleep(ctx, interval); err != nil {
			return err
		}
	}
}

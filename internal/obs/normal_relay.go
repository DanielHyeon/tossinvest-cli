package obs

// normal_relay.go 는 a092 착지 단위 ⑤(C8) — exit 관측 goroutine 의 일반 등급 알림을 그 goroutine 밖에서 보내는 유계 이관임.
//
// 델타: 「이 요구는 등급으로 좁혀지지 않는다」 — 일반 등급의 동기 발행도 같은 네트워크 왕복을 손절 goroutine 에서 기다림.
// 일반 등급은 durable 기록이 없으므로 이관은 유실을 허용하되, 무엇을 버렸는지는 기록해야 함.
//
// 비우는 쪽은 별도 보조 실행자(엔진 런타임 Auxiliary)임 — critical 배달 실행자의 사이클에 얹지 않음(얹으면 그 사이클 길이에
// 일반 등급 발행 시간이 더해져 critical 래치가 늦어짐, design D0.3g 6).

import (
	"context"
)

// DefaultNormalRelayCapacity 는 버퍼 크기임. exit 루프의 일반 등급은 capped · unmanaged 둘이고 포지션마다 한 사이클에 하나 —
// 이 수를 넘게 밀리면 전송이 오래 멈춘 것이고, 그때 버리는 것이 이 등급의 정의임.
const DefaultNormalRelayCapacity = 64

// NormalRelay 는 일반 등급 알림의 유계 버퍼와 그것을 비우는 실행자임.
type NormalRelay struct {
	n  *Notifier
	ch chan Event
}

// NewNormalRelay 는 capacity(0 이하면 기본값) 크기의 이관을 만듦.
func NewNormalRelay(n *Notifier, capacity int) *NormalRelay {
	if capacity <= 0 {
		capacity = DefaultNormalRelayCapacity
	}
	return &NormalRelay{n: n, ch: make(chan Event, capacity)}
}

// Offer 는 알림을 넘기고 곧바로 반환함 — 막히지 않음. 버퍼가 차 있으면 버리고 기록함.
func (r *NormalRelay) Offer(e Event) {
	if r == nil {
		return
	}
	select {
	case r.ch <- e:
	default:
		r.n.logNormalDrop(e, "the normal-grade relay buffer is full")
	}
}

// Run 은 버퍼를 비우며 최선 발송함. ctx 가 끝나면 남은 알림을 버림으로 기록하고 ctx.Err() 를 돌려줌(런타임은 취소를 취소로
// 보고받아야 정상 종료로 판정함 — auxiliary.go).
func (r *NormalRelay) Run(ctx context.Context) error {
	for {
		// 종료가 먼저 — select 는 준비된 갈래를 무작위로 고르므로, 취소와 대기 알림이 같이 준비되면 종료 뒤에 발송을 시작할 수 있음.
		if ctx.Err() != nil {
			return r.drainOnShutdown(ctx)
		}
		select {
		case <-ctx.Done():
			return r.drainOnShutdown(ctx)
		case e := <-r.ch:
			if r.n != nil {
				r.n.publishBestEffort(ctx, e, SeverityOf(e.Type))
			}
		}
	}
}

// drainOnShutdown 은 종료 배수 — 남은 알림을 하나씩 버림으로 기록하고 ctx.Err() 를 돌려줌(K13).
func (r *NormalRelay) drainOnShutdown(ctx context.Context) error {
	for {
		select {
		case e := <-r.ch:
			r.n.logNormalDrop(e, "the engine stopped before this normal-grade alert was sent")
		default:
			return ctx.Err()
		}
	}
}

// logNormalDrop 은 버린 일반 등급 알림 한 줄 — 유형과 키(필드는 싣지 않음: 계좌를 담을 수 있음).
func (n *Notifier) logNormalDrop(e Event, why string) {
	if n == nil || n.Log == nil {
		return
	}
	n.Log.Warn(EventNormalAlertDropped,
		FieldEvent, string(e.Type),
		"alert_key", n.eventKey(e),
		FieldDetail, why)
}

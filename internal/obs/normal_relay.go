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
	"sync"
)

// DefaultNormalRelayCapacity 는 버퍼 크기임. exit 루프의 일반 등급은 capped · unmanaged 둘이고 포지션마다 한 사이클에 하나 —
// 이 수를 넘게 밀리면 전송이 오래 멈춘 것이고, 그때 버리는 것이 이 등급의 정의임.
const DefaultNormalRelayCapacity = 64

// NormalRelay 는 일반 등급 알림의 유계 버퍼와 그것을 비우는 실행자임.
//
// 모든 유실 경로가 기록됨(델타 「무엇을 버렸는지는 기록되어야 한다」, 26라운드 codex #2 · #3): 가득 참 · 이관 없음(RecordOnly) ·
// 종료 배수 · 실행자가 멈춘 뒤의 넘김 · 발행 중 패닉(그 알림과 남은 큐) · 전송기 없음 · 발행 실패.
type NormalRelay struct {
	n  *Notifier
	ch chan Event

	// mu 는 「넘김」과 「멈춤 표시」를 직렬화하는 짧은 로컬 임계 구역임 — 멈춘 뒤 들어온 알림이 아무도 안 비우는 큐에
	// 조용히 쌓이지 않게. 전송은 이 잠금 밖.
	mu      sync.Mutex
	stopped bool
}

// NewNormalRelay 는 capacity(0 이하면 기본값) 크기의 이관을 만듦.
func NewNormalRelay(n *Notifier, capacity int) *NormalRelay {
	if capacity <= 0 {
		capacity = DefaultNormalRelayCapacity
	}
	return &NormalRelay{n: n, ch: make(chan Event, capacity)}
}

// Offer 는 알림을 넘기고 곧바로 반환함 — 막히지 않음. 가득 찼거나 실행자가 멈췄으면 버리고 기록함.
func (r *NormalRelay) Offer(e Event) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopped {
		r.n.logNormalDrop(e, "the normal-grade relay has stopped")
		return
	}
	select {
	case r.ch <- e:
	default:
		r.n.logNormalDrop(e, "the normal-grade relay buffer is full")
	}
}

// Run 은 버퍼를 비우며 최선 발송함. ctx 가 끝나면 남은 알림을 버림으로 기록하고 ctx.Err() 를 돌려줌(런타임은 취소를 취소로
// 보고받아야 정상 종료로 판정함 — auxiliary.go). 어떻게 끝나든(패닉 포함) 멈춤을 표시하고 처리 중이던 알림과 남은 큐를
// 버림으로 기록함 — 패닉은 그대로 위로 올라가 런타임이 실행자 정지로 기록함.
func (r *NormalRelay) Run(ctx context.Context) (err error) {
	var inFlight *Event
	defer func() {
		if inFlight != nil {
			r.n.logNormalDrop(*inFlight, "the normal-grade relay stopped while sending this alert")
		}
		r.stop("the normal-grade relay stopped before this alert was sent")
	}()
	for {
		// 종료가 먼저 — select 는 준비된 갈래를 무작위로 고르므로, 취소와 대기 알림이 같이 준비되면 종료 뒤에 발송을 시작할 수 있음.
		if ctx.Err() != nil {
			return ctx.Err()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-r.ch:
			inFlight = &e
			r.publish(ctx, e)
			inFlight = nil
		}
	}
}

// stop 은 멈춤을 표시하고 남은 큐를 하나씩 버림으로 기록함(K13 종료 배수 포함).
func (r *NormalRelay) stop(why string) {
	r.mu.Lock()
	r.stopped = true
	r.mu.Unlock()
	for {
		select {
		case e := <-r.ch:
			r.n.logNormalDrop(e, why)
		default:
			return
		}
	}
}

// publish 는 한 알림을 최선 발송하고, 전송기가 없거나 발행이 실패하면 유형 + 키로 버림을 기록함.
func (r *NormalRelay) publish(ctx context.Context, e Event) {
	if r.n == nil || r.n.Publisher == nil {
		r.n.logNormalDrop(e, "no notification publisher is configured")
		return
	}
	if err := r.n.Publisher.Publish(ctx, notificationFor(e, SeverityOf(e.Type))); err != nil {
		r.n.logNormalDrop(e, "publishing the normal-grade alert failed: "+err.Error())
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

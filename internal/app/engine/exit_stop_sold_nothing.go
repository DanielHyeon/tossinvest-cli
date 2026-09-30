package engine

// exit_stop_sold_nothing.go 는 a091 의 보고 — 확정 하한이 청산을 0주로 만든 사실을 누가 · 어떤 등급으로 아는가.
//
// # 왜 있나
//
// 2026-08-02 042660: 엔진이 편입해 보호하던 포지션의 손절이 3분간 13회 연속 0주로 깎였고(한정 항 매도가능 0), 그 13회에 대해 durable
// outbox 에 남은 행은 0 이었음 — 0주가 부분 캡과 같은 normal 종류(exit.proposal_capped)로만 보고됐기 때문임. 게다가 하한을 계산하지
// 못한 경우(B2)는 알림조차 없었음.
//
// # 무엇을 지키나 (engine-safety 「등급화된 알림」 a091 문단 · design D1 · D3 · D5 · D7 · D8)
//
//   - 제출 수량 · 제출 여부 · 발의 해제는 여기서 정하지 않음 — applyFloor 의 반환값은 이 파일이 생기기 전과 같음(§0.3 · §0.9).
//   - 보호 청산 · 알림 켜짐 · 보고 대상 원인(계산 실패 · 매도가능/로컬/스냅숏)만 critical(EventExitStopSoldNothing). 알림 꺼짐 ·
//     계좌 보유 0 · 종료 취소 · 익절은 옛 종류(normal) — 꺼진 엔진에서 critical 은 보낼 수 없는 행이 되어 진입을 멈춤(a095 규칙).
//   - 사건 키는 포지션당 하나(에피소드). 행 본문 = 첫 원인 + 관측 시각, 관측마다의 원인 = 로그 줄(RecordOnly 가 본문을 detail 로 씀).
//   - 원문 오류 · 계좌는 제목 · 본문 · payload 에 없음. 로그 줄은 계좌 필드 없이, 오류는 MaskAccount.
//   - 기록은 context.WithoutCancel — 판정 뒤 종료가 끼어도 끝난 ctx 로 기록이 실패해 가짜 래치 · 승격이 나지 않음. 대가는 종료가
//     그 기록만큼(기한 없이) 늦는 것(「종료 중 보고 대기」).

import (
	"context"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// zeroCause 는 0주의 원인 범주임(design D3).
type zeroCause int

const (
	zeroFloorUnknown zeroCause = iota // ① 하한 계산 실패(B2)
	zeroFloorAllows                   // ② 하한이 0 — 매도가능 · 로컬 매도 · 스냅숏
	zeroNoHolding                     // ③ 계좌 보유 0 — 엔진 밖 종결 진행 중
	zeroShutdown                      // ④ 하한 조회가 관측을 끝내는 취소뿐으로 실패
)

// zeroFloor 는 applyFloor 가 0주를 돌려주기 직전의 사실 하나임.
type zeroFloor struct {
	cause      zeroCause
	proposed   string
	protective bool
	floor      riskcalc.ConfirmedFloor // ②③ 에서만
	err        error                   // ①④ 에서만
}

// classifyZero 는 0주의 원인을 가름. err 가 있으면 B2(①·④), 없으면 끝 경로(②·③).
//
// ③ 판정은 한 방향만 정확함(design D3 ③): Bound == Holdings ∧ "0" 이면 신선한 보유가 0 임. 보유 0 을 읽었는데 매도가능이 없거나
// 낡아 다른 한정 항이 붙는 칸은 ②(critical)로 남음 — 과보고 방향(이름 붙은 잔여).
func classifyZero(ctx context.Context, floor riskcalc.ConfirmedFloor, err error) zeroCause {
	switch {
	case err != nil && ctx.Err() != nil && cancellationOnly(err):
		return zeroShutdown
	case err != nil:
		return zeroFloorUnknown
	case floor.Bound == riskcalc.FloorBoundHoldings:
		return zeroNoHolding
	default:
		return zeroFloorAllows
	}
}

// cancellationOnly 는 오류 나무의 잎이 전부 context.Canceled 인지 봄(design D5 ④ 6판).
//
// errors.Is(err, context.Canceled) 하나로는 모자람 — 생산 Retrier.Query 는 인증 거절과 승격 실패를 errors.Join 으로 합쳐 돌려주므로,
// 종료가 승격의 원장 연산을 실패시키면 「인증 거절 + 취소」가 되고, 그것을 억제하면 진짜 브로커 실패를 숨김. 잎이 문자열로 지워진
// 오류(공식 클라이언트의 ErrTransport 감쌈)는 취소가 아닌 것으로 읽힘 — 과보고 방향(이름 붙은 잔여).
func cancellationOnly(err error) bool {
	if err == nil {
		return false
	}
	switch u := err.(type) {
	case interface{ Unwrap() []error }:
		leaves := u.Unwrap()
		seen := false
		for _, e := range leaves {
			if e == nil {
				continue
			}
			if !cancellationOnly(e) {
				return false
			}
			seen = true
		}
		return seen
	case interface{ Unwrap() error }:
		if inner := u.Unwrap(); inner != nil {
			return cancellationOnly(inner)
		}
		return false
	}
	return err == context.Canceled
}

// critical 은 이 사실이 새 종류(critical)로 보고되는지임 — 보호 청산 · 알림 켜짐 · 보고 대상 원인(①②).
func (o *ExitObserver) zeroIsCritical(z zeroFloor) bool {
	return z.protective && o.opts.NotificationsEnabled && (z.cause == zeroFloorUnknown || z.cause == zeroFloorAllows)
}

// reportZeroFloor 는 0주 사실 하나를 보고함. 반환값에는 영향이 없음.
//
// B2(err 있음)는 오늘처럼 로그 한 줄을 남기고, critical 일 때만 알림을 더함(익절 · 알림 꺼짐 · 종료 취소는 알림 0 — 오늘과 같음).
// 끝 경로는 오늘처럼 알림 하나 — 종류만 critical 여부로 고름.
func (o *ExitObserver) reportZeroFloor(ctx context.Context, m managed, z zeroFloor) {
	kind := obs.EventExitProposalCapped
	if o.zeroIsCritical(z) {
		kind = obs.EventExitStopSoldNothing
	}
	if z.err != nil {
		o.logZeroFloor(kind, z.err, "the confirmed floor of "+m.position.Symbol+" could not be computed; nothing is submitted")
		if kind != obs.EventExitStopSoldNothing {
			return
		}
	}
	o.notifyZeroFloor(ctx, kind, o.zeroFloorEvent(m, kind, z))
}

// zeroFloorEvent 는 0주 알림을 만듦. 제목 · 본문은 한국어 · 이름(코드) · 계좌 없음(정본 a085). 본문은 원인 범주와 관측 시각을 담음 —
// 미전달 행은 첫 기록의 본문을 유지하므로 시각이 없으면 늦게 읽힌 본문이 거짓이 됨(design D7).
func (o *ExitObserver) zeroFloorEvent(m managed, kind obs.EventType, z zeroFloor) obs.Event {
	label := o.label(m.position.Symbol)
	title := label + " 청산이 확정 하한에 걸려 한 주도 나가지 않았다"
	if kind == obs.EventExitStopSoldNothing {
		title = label + " 손절이 한 주도 나가지 않았다"
	}
	fields := map[string]any{
		obs.FieldSymbol:   m.position.Symbol,
		obs.FieldQuantity: "0",
		"proposed":        z.proposed,
		"remainder":       z.proposed,
		"position_id":     m.position.ID,
		"cause":           zeroCauseCode(z),
	}
	if z.err == nil {
		fields["floor_bound"] = z.floor.Bound
	}
	return obs.Event{
		Type:  kind,
		Key:   string(kind) + "|" + m.position.ID,
		Title: title,
		Body: fmt.Sprintf("%s 관측: %s. 제안 %s주 중 0주를 제출했다 — 손절이 나가지 않았다. "+
			"불일치가 해소되면 같은 단계를 다시 제안한다.",
			o.clk.Now().UTC().Format(time.RFC3339), zeroCauseText(z), z.proposed),
		Fields: fields,
	}
}

func zeroCauseCode(z zeroFloor) string {
	switch z.cause {
	case zeroFloorUnknown:
		return "floor_unknown"
	case zeroNoHolding:
		return "no_holding"
	case zeroShutdown:
		return "shutdown"
	}
	return "floor_zero"
}

// zeroCauseText 는 원인의 한국어 범주임. 원문 오류는 싣지 않음(엔진 로그에만, 가려서).
func zeroCauseText(z zeroFloor) string {
	switch z.cause {
	case zeroFloorUnknown, zeroShutdown:
		return "확정 하한을 계산하지 못했다(계좌 조회 실패)"
	case zeroNoHolding:
		return "계좌에 보유가 없다 — 엔진 밖에서 종결되는 중일 수 있다"
	}
	switch z.floor.Bound {
	case riskcalc.FloorBoundSellable:
		return "매도가능 수량이 0이다 — 다른 미체결 매도가 주식을 잡고 있을 수 있다"
	case riskcalc.FloorBoundLocalSells:
		return "엔진의 미체결 매도가 남은 수량을 모두 쓰고 있다"
	case riskcalc.FloorBoundNoSnapshot:
		return "계좌 스냅숏을 읽지 못했다"
	case riskcalc.FloorBoundStaleSnapshot:
		return "계좌 스냅숏이 낡았다"
	}
	return "확정 하한이 0이다"
}

// notifyZeroFloor 는 알림기에 직접 기록함 — o.alert 를 거치지 않음. o.alert 의 실패 로그(logErr)는 계좌 원문을 싣기 때문임(D8).
// 재알림 창은 알림기(RecordOnly)의 것이 그대로 쓰임.
func (o *ExitObserver) notifyZeroFloor(ctx context.Context, kind obs.EventType, e obs.Event) {
	if o.opts.Alerts == nil {
		return
	}
	if err := o.opts.Alerts.Notify(context.WithoutCancel(ctx), e); err != nil {
		o.logZeroFloor(kind, err, "the report that the stop sold nothing could not be made durable")
	}
}

// logZeroFloor 는 a091 의 로그 줄임 — 계좌 필드 없이, 오류는 계좌를 가려서.
func (o *ExitObserver) logZeroFloor(kind obs.EventType, err error, detail string) {
	if o.opts.Log == nil {
		return
	}
	o.opts.Log.Error(kind, obs.MaskAccount(err, o.opts.AccountRef), obs.FieldDetail, detail)
}

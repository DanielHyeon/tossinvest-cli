package reconcile

// acked_boot.go 는 기동이 ACKED 로 남은 attempt 를 다루는 한 자리임(a094 D−4.2 · D−5.1 · D−5.2).
//
// # 왜 있나
//
// 재시작 규칙은 ACKED 를 그대로 둠(RecoverPending) — 브로커가 번호를 줬고 이전 프로세스가 확정하기 전에 죽었음.
// 그 attempt 는 PendingAttempts 에 남아 종목의 모든 mutation 을 막고, 그것이 무장한 발의는 재시작마다 얼어 있었음.
//
// # 무엇을 하나 — 바이트 일치 확정만, 나머지는 알림만
//
//   - ACKED PLACE + 기록 번호: 그 번호로 **한 번** 읽어(execgw.ConfirmPlacedOrder — 발주 직후 확인과 같은 판정) 번호 바이트
//     일치 · 종목 일치면 CONFIRMED 로 종결함.
//   - 그 외 전부(읽기 실패 · 번호/종목 불일치 · 해석 불가 · 번호 없음 · CANCEL/AMEND · 확정 쓰기 실패): 상태를 바꾸지 않고
//     attempt 를 에피소드로 한 critical 을 창 0 으로 기록함. 목록 대조 해소로 보내지 않음(해소기 matcher 에 번호 판별자가
//     없어 지문이 같은 다른 주문의 번호로 덮을 수 있음 — indoubt.go 의 단일 일치 덮어쓰기). IN_DOUBT 로 올리지도 않음.
//   - 이 경로의 실패는 복구를 실패시키지 않음 — ErrRecoveryIncomplete 는 관측 루프를 하나도 시작시키지 않음(모든 손절 정지).
//
// # 예산(§0.4)
//
// ACKED PLACE 행마다 주문 GET 1(401 재발급 재시도 포함 ≤3) + 토큰 POST ≤2, 전부 roundTripTimeout(3초) 안. 기동에서만, ready 앞.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// CriticalRecorder 는 알림기의 critical 기록 단일 입구임(a092 K6). *obs.Notifier 가 구현함.
type CriticalRecorder interface {
	RecordCritical(ctx context.Context, e obs.Event, remindAfter time.Duration) error
}

// ReasonBootAckedConfirmed 는 기동이 ACKED 발주를 기록 번호의 바이트 일치로 확정했음을 뜻하는 전이 사유임.
const ReasonBootAckedConfirmed = "boot_acked_readback_confirmed"

// confirmAcked 는 ACKED attempt 하나를 다룸. true 면 CONFIRMED 로 종결됐음(더는 미종결이 아님).
func (r *Recovery) confirmAcked(ctx context.Context, rec journal.AttemptRecord) bool {
	intent, err := r.opts.Journal.LookupIntent(ctx, rec.IntentID)
	if err != nil {
		r.alertAcked(ctx, rec, "", fmt.Sprintf("the intent could not be read: %v", err))
		return false
	}
	if rec.Kind != journal.KindPlace {
		r.alertAcked(ctx, rec, intent.Symbol, "a CANCEL/AMEND acknowledgement has no read-back rule; it stays ACKED")
		return false
	}
	if strings.TrimSpace(rec.BrokerOrderID) == "" {
		r.alertAcked(ctx, rec, intent.Symbol, "the acknowledgement recorded no broker order number to read back")
		return false
	}
	if err := execgw.ConfirmPlacedOrder(ctx, r.opts.Resolver.Order, rec.BrokerOrderID, intent.Symbol); err != nil {
		r.alertAcked(ctx, rec, intent.Symbol, "the read-back did not confirm it: "+withoutBrokerBody(err))
		return false
	}
	attempt, err := r.opts.Journal.Resume(ctx, rec.ID)
	if err == nil {
		err = attempt.ResolveConfirmed(ctx, rec.BrokerOrderID, ReasonBootAckedConfirmed,
			"restart recovery read the recorded order number back byte-for-byte on the same symbol")
	}
	if err != nil {
		r.alertAcked(ctx, rec, intent.Symbol, fmt.Sprintf("the read-back matched but the ledger write failed: %v", err))
		return false
	}
	return true
}

// alertAcked 는 확정하지 못한 ACKED attempt 를 명명 critical 로 기록함 — key 는 attempt(원장 행이라 재시작에도 같은 행).
func (r *Recovery) alertAcked(ctx context.Context, rec journal.AttemptRecord, symbol, reason string) {
	key := string(obs.EventOrderInDoubt) + "|acked:" + rec.ID
	e := obs.Event{
		Type:  obs.EventOrderInDoubt,
		Key:   key,
		Title: fmt.Sprintf("재시작 뒤 ACKED 로 남은 %s attempt — 확정하지 못했다", rec.Kind),
		Body: fmt.Sprintf("attempt %s(%s · 종목 %s · intent %s)는 브로커가 접수했으나 이전 프로세스가 확정하기 전에 멈췄다. "+
			"사유: %s. 상태를 바꾸지 않았다 — 그 종목의 mutation 은 막힌 채이고 그것이 무장한 발의는 무장된 채 남는다.",
			rec.ID, rec.Kind, orUnknown(symbol), rec.IntentID, reason),
		Fields: map[string]any{
			obs.FieldAttemptID: rec.ID,
			obs.FieldOrderID:   rec.BrokerOrderID,
			obs.FieldSymbol:    symbol,
			"attempt_kind":     string(rec.Kind),
			"intent_id":        rec.IntentID,
			obs.FieldReason:    reason,
		},
	}
	if r.opts.Alerts == nil {
		return
	}
	// 기록 실패는 입구가 진입을 잠금(a092). 복구는 계속함.
	_ = r.opts.Alerts.RecordCritical(ctx, e, 0)
}

// CatchUpExitProposals 는 복구 뒤 · 준비 신호 앞의 기동 따라잡기를 부름(a094 D−2.5 · D−2.6-2). 조립부가 넣은 함수가
// 없으면 아무것도 하지 않음 — 생산 조립(engine.Context.Recovery)은 늘 넣음.
func (r *Recovery) CatchUpExitProposals(ctx context.Context) {
	if r.opts.CatchUp != nil {
		r.opts.CatchUp(ctx)
	}
}

// CatchUpWired 는 조립부가 기동 따라잡기를 넣었는지임 — 배선 시험이 역할을 재는 자리(생산 조립은 늘 넣음).
func (r *Recovery) CatchUpWired() bool { return r.opts.CatchUp != nil }

// withoutBrokerBody 는 알림 본문에 넣을 오류 문구임 — 브로커 응답 본문은 계좌 식별자를 담을 수 있으므로(official/errors.go
// 의 경고) 공식 API 오류는 상태 코드만 남기고 본문을 뺌(불변식 8). 그 밖의 오류(번호 · 종목 불일치 · 전송)는 그대로.
func withoutBrokerBody(err error) string {
	var apiErr *official.APIError
	if errors.As(err, &apiErr) {
		return fmt.Sprintf("the broker answered HTTP %d to the order read (body withheld)", apiErr.Code)
	}
	return err.Error()
}

func orUnknown(s string) string {
	if strings.TrimSpace(s) == "" {
		return "?"
	}
	return s
}

package obs

// record_only.go 는 a092 의 알림기 기록 전용 입구임.
//
// # 왜 있나
//
// exit 관측 goroutine 은 손절 판정을 도는 루프임. 그 goroutine 이 critical 알림이나 모드 전이 통지를 동기로 보내면
// 원격 전송이 멈춘 동안 다음 손절 판정이 그 전송을 기다림(정본 engine-safety 「등급화된 알림」). 그래서 그 goroutine 에
// 주입되는 알림 부품은 **기록까지만** 하고 반환함. 발송은 배달 실행자(a124 alertDeliverer)가 원장의 PENDING 행을
// 집어서 함.
//
// # 무엇을 지키나
//
//   - 발송 임차를 잡지 않음(Journal.RecordAlert) — 잡으면 배달 실행자가 그 임차가 끝날 때까지 행을 못 집음.
//   - 알림기의 배제 잠금(n.mu) 아래에서 기록함 — 운영자 승인(Acknowledge)의 「미전달 셈 ~ 전달 실패 사유 해제」 사이에
//     이 경로로 PENDING 행이 끼어들지 못함. 이 잠금 아래 기록하는 경로(이 입구 + 동기 claim)가 정본의 critical 기록 부류임.
//   - 기록 실패는 그 자리에서 전달 실패 사유로 잠그고 durable 차단(승격)을 시도함 — 아무것도 안 적힌 채 진입이 열리지 않게.
//   - 원격 전송(Publisher)을 부르지 않음 — 원장이 없을 때도(동기 publish 갈래로 떨어지지 않음, C16).

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// RecordOnly 는 알림기를 기록 전용으로 감싼 값임. exit 관측 루프의 알림 인터페이스(Notify)와
// journal.ModeAnnouncer 를 구현함. 값 타입이라 주입 지점마다 복사해도 같은 알림기를 가리킴.
//
// N 이 nil 이면 무동작 — Notifier.AnnounceOperatingMode 의 n == nil 과 같은 계약(알림 없는 조립도 전이는 함).
type RecordOnly struct {
	N *Notifier
	// Relay 는 일반 등급 알림을 넘길 유계 버퍼임(a092 C8). nil 이면 일반 등급은 버림으로 기록됨 — 동기 발행으로 떨어지지 않음.
	Relay *NormalRelay
}

// Notify 는 사건을 로그로 남기고 critical 이면 원장에 기록만 함.
//
// 일반 등급은 유계 버퍼(Relay)에 넣고 반환함(a092 C8) — 보조 실행자가 비움. 원격 전송을 여기서 부르지 않음: capped 알림은
// 축소 청산 제출 **앞**에서 나오므로(25라운드 보이스 A #1) 여기서 기다리면 그 손절 주문이 전송만큼 늦음.
func (r RecordOnly) Notify(ctx context.Context, e Event) error {
	n := r.N
	if n == nil {
		return nil
	}
	severity := SeverityOf(e.Type)
	n.logEvent(withoutFields(e), severity)
	if severity != SeverityCritical {
		if r.Relay == nil {
			n.logNormalDrop(e, "no normal-grade relay is wired")
			return nil
		}
		r.Relay.Offer(e)
		return nil
	}
	return n.recordCritical(ctx, e, n.remindAfter())
}

// AnnounceOperatingMode 는 모드 전이 통지를 기록만 함. 사건 모양은 동기 통지(Notifier.AnnounceOperatingMode)와
// 같은 함수(operatingModeEvent)로 만듦 — 두 통지가 같은 전이를 다르게 적지 않게.
func (r RecordOnly) AnnounceOperatingMode(ctx context.Context, previous string, rec journal.OperatingModeRecord) error {
	n := r.N
	if n == nil {
		return nil
	}
	e := operatingModeEvent(previous, rec)
	n.logEvent(withoutFields(e), SeverityOf(e.Type))
	return n.recordCritical(ctx, e, n.remindAfter())
}

// ErrAlertNotDurable 는 기록 전용 입구가 critical 사건을 원장에 남길 수 없음(알림기 없음 · 원장 없음)을 뜻함.
// obs 밖 기록자가 「통지됨」을 거짓으로 보고하지 않게 오류로 돌려줌.
var ErrAlertNotDurable = errors.New("obs: no journal is wired, so the critical record is not durable")

// RecordCritical 은 알림기의 기록 전용 입구를 obs 밖 기록자에게 여는 공개 메서드임(a092 25.6 — a066 완화 통지).
//
// 세울 자기 진입 차단 사유가 없는 기록자는 원장에 직접 쓰지 말고 이 입구를 써야 함(정본 「critical 기록 부류」):
// 알림기의 배제 잠금 아래에서 기록하므로 운영자 승인의 셈~해제 사이에 끼어들지 못함. 등급표와 무관하게 critical 로
// 기록함 — 호출자가 durable 통지를 원해서 부르는 입구이기 때문임. 원격 전송은 하지 않음(배달 실행자 몫).
//
// remindAfter 는 기록자별(0 이면 정착 행을 재무장하지 않음). n 이 nil 이거나 원장이 없으면 ErrAlertNotDurable.
func (n *Notifier) RecordCritical(ctx context.Context, e Event, remindAfter time.Duration) error {
	if n == nil || n.Journal == nil {
		if n != nil && n.Log != nil {
			n.Log.Warn(EventAlertUndelivered,
				FieldTriggerEvent, string(e.Type),
				FieldDetail, "no journal is wired, so this critical record is not durable")
		}
		return ErrAlertNotDurable
	}
	// 구조화 로그 줄에는 필드를 싣지 않음 — 필드는 원장 payload 로만 감(Manager 판정 2026-09-30, 불변식 8 잠정 규칙:
	// 브로커 계좌번호 원문은 어떤 형태로도 로그 금지, 원장 내부 키 · 마스킹 형식만 허용. 이 입구의 기록자(a066 완화 통지)는
	// 필드에 계좌를 담은 대상을 싣므로 로그에서 뺌. 사람의 계좌 가림 설계가 확정되면 그쪽이 우선함).
	// 유형 · 키 · 제목 · 본문은 남음 — 제목 · 본문은 기록자가 외부 전송용으로 계좌를 뺀 문구임.
	n.logEvent(withoutFields(e), SeverityCritical)
	return n.recordCritical(ctx, e, remindAfter)
}

// withoutFields 는 기록 전용 입구의 로그 줄에서 필드를 뺀 사본임 — 입구의 세 메서드가 모두 씀(a092 26라운드 보이스 B #1:
// f48e7865 가 RecordCritical 하나에만 적용했음). 필드는 원장 payload 로만 감. 계좌를 담는 필드(exit 두절 · 모드 통지의
// FieldAccount)가 이 입구로 들어오므로 로그 줄에는 유형 · 키 · 본문만 남김(불변식 8, Manager 판정 2b 의 연장).
func withoutFields(e Event) Event {
	e.Fields = nil
	return e
}

// recordCritical 는 critical 사건 하나를 발송 임차 없이 원장에 기록함 — 알림기의 기록 전용 입구.
//
// remindAfter 는 기록자별임: exit 기록은 알림기의 재알림 창을 넘겨 정착 행을 재무장하고(정본 SHALL), 다른 기록자는
// 0(재무장 안 함)을 넘길 수 있음.
func (n *Notifier) recordCritical(ctx context.Context, e Event, remindAfter time.Duration) error {
	if n.Journal == nil {
		// 원장이 없으면 durable 기록이 불가함. 동기 publish 로 떨어지면 exit goroutine 이 다시 전송을 기다리므로
		// 크게 경고만 하고 반환함.
		if n.Log != nil {
			n.Log.Warn(EventAlertUndelivered,
				FieldTriggerEvent, string(e.Type),
				FieldDetail, "no journal is wired, so this critical alert is neither durable nor sent")
		}
		return nil
	}
	record := journal.Alert{
		EventKey: n.eventKey(e),
		Type:     string(e.Type),
		Severity: string(SeverityCritical),
		Title:    e.Title,
		Body:     e.Body,
		Payload:  encodeFields(e.Fields),
	}

	// 배제 잠금 아래 기록 — 잠금 안은 로컬 원장 쓰기 하나뿐이고 원격 전송은 없음.
	n.mu.Lock()
	_, _, err := n.Journal.RecordAlert(ctx, record, remindAfter)
	if err != nil {
		// 기록 실패 = 행도 발송도 없음. 호출자가 오류를 버려도 결과가 같도록 여기서 진입을 잠금(청산은 무관).
		if n.Log != nil {
			// 원문 오류는 사건 키(모드 통지 · exit 두절 — 계좌 원문 포함)를 품으므로 계좌를 가려서 남김(26라운드 보이스 B
			// 재확인 N1 — Manager 판정 (a) 「원문은 계좌를 가려 엔진 로그에만」).
			n.Log.Error(EventAlertUndelivered, MaskAccount(err, n.AccountRef), FieldTriggerEvent, string(e.Type))
		}
		if n.Gate != nil {
			// 설명은 고정 문구 — 원문 오류는 사건 키(모드 통지 · exit 두절은 계좌를 담음)를 품고, 게이트 설명은 상태 출력이
			// 어디서나 읽는 칸임(a092 26라운드 보이스 B #2). 원문은 위 로그 줄에만.
			n.Gate.Block(execgw.ReasonAlertUndelivered, fmt.Sprintf(
				"a critical %s alert could not be recorded in the outbox (details are in the engine log)", e.Type))
		}
	}
	n.mu.Unlock()

	if err != nil {
		// 게이트 래치는 메모리에만 있으므로 재시작을 견디는 절반(모드 승격)을 잠금 밖에서 시도함 — 승격의 통지자는
		// nil 이라 이 입구로 되돌아오지 않지만, 잠금 밖에 두는 것은 notifyCritical 과 같은 이유(재진입 교착 방지).
		n.escalate(ctx, e)
		return fmt.Errorf("obs: recording a critical alert: %w", err)
	}
	return nil
}

// OperatingModeEventKey 는 모드 전이 통지의 중복 제거 신원임(a092 K1 — 전이 행 하나). 완화 명령이 통지 행을 다시 읽을 때도 씀.
func OperatingModeEventKey(rec journal.OperatingModeRecord) string {
	return "operating_mode:" + rec.AccountRef + ":" + rec.Mode + ":" + rec.ID
}

// operatingModeEvent 는 모드 전이 하나의 통지 사건을 만듦 — 동기 통지와 기록 전용 통지가 같이 씀.
func operatingModeEvent(previous string, rec journal.OperatingModeRecord) Event {
	direction := "tightened"
	if journal.MoreConservativeMode(previous, rec.Mode) == previous && previous != rec.Mode {
		direction = "relaxed"
	}
	return Event{
		Type: EventOperatingMode,
		// 중복 제거 신원은 전이 하나(전이 행 id)임(a092 K1). 계정+목적 모드만으로 두면 「강화 → 완화 → 재알림 창 안의
		// 재강화」의 통지가 옛 정착 행에 흡수되어 나가지 않음. 같은 전이의 이중 통지는 원장이 막음 — 변화 없는 전이는
		// 통지 전에 반환하므로(TransitionOperatingMode 무변화 규칙) 이 신원이 통지 수를 늘리지 않음.
		Key:   OperatingModeEventKey(rec),
		Title: fmt.Sprintf("operating mode %s: %s → %s", direction, modeLabel(previous), rec.Mode),
		Body: fmt.Sprintf("%s by %s — %s. Exposure-raising mutations are %s; risk-reducing ones are unaffected.",
			rec.Mode, rec.Actor, rec.Cause, permission(rec)),
		Fields: map[string]any{
			FieldAccount:   rec.AccountRef,
			FieldFromState: modeLabel(previous),
			FieldToState:   rec.Mode,
			FieldActor:     rec.Actor,
			FieldReason:    rec.Cause,
		},
	}
}

// MaskAccount 는 오류 문구 속 계좌 원문을 "[account]" 로 가린 오류를 돌려줌. 계좌가 비었거나 오류가 nil 이면 그대로.
// 가리는 것은 넘겨받은 계좌 하나뿐임 — 일반 마스커가 아님(base 의 계좌 로그 관행은 사람 결정 큐). 기록 입구와
// mode-release(ModeOperations.logFailure)가 이 한 함수를 씀.
func MaskAccount(err error, account string) error {
	acct := strings.TrimSpace(account)
	if err == nil || acct == "" || !strings.Contains(err.Error(), acct) {
		return err
	}
	return errors.New(strings.ReplaceAll(err.Error(), acct, "[account]"))
}

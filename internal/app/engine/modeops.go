package engine

// modeops.go 는 사람의 운영 모드 완화를 엔진 프로세스 안에서 실행함(a092 착지 단위 ④, 델타 ADDED 「운영 모드는 산 프로세스의
// 진입 게이트에 닿는다」의 완화 경로).
//
// # 왜 엔진 안인가
//
// 모드 사유 래치는 이 프로세스의 EntryGate 메모리에 삶. 다른 프로세스가 원장만 고치면 그 프로세스에는 투영기가 없으므로
// 산 게이트는 재시작 전까지 안 풀림(정본 「운영자가 밀린 알림을 읽고 승인하는 경로」와 같은 이유).
//
// # 판정은 원장이 함
//
// 자동은 조이기만 · 완화는 OPERATOR + 승인 참조 · audit 가 commit 앞 — 전부 `Journal.TransitionOperatingMode` 가 이미 강제함
// (design D0.3f 1). 여기서 다시 판정하지 않음(판정이 둘이면 서로의 시험을 통과시킴). 이 파일이 더하는 것은 (1) 입력 네 칸의
// 필수 검사, (2) audit 로그 부재 거절, (3) 운영자 이름을 사유에 붙임, (4) 통지를 기록 전용 입구로, (5) 전이 뒤 재읽기.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

var (
	// ErrModeReleaseInvalid 는 요청 모양이 틀렸음(빈 칸 · 허용 밖 목적 모드). 아무것도 바뀌지 않음.
	ErrModeReleaseInvalid = errors.New("engine: invalid mode release request")
	// ErrModeReleaseUnavailable 은 이 엔진에 audit 로그가 없음. 기록 없는 완화가 이 경로가 막으려는 상태라 거절함.
	ErrModeReleaseUnavailable = errors.New("engine: the mode release surface is unavailable")
)

// ModeOperations 는 모드 완화 명령의 엔진 쪽 구현임.
type ModeOperations struct {
	journal    *journal.Journal
	gate       *execgw.EntryGate
	notifier   *obs.Notifier
	auditor    journal.ModeAuditor
	accountRef string
	// announcer 는 통지 기록자 — 기본은 알림기의 기록 전용 입구. 시험이 재강화 픽스처를 꽂을 수 있게 필드로 둠.
	announcer journal.ModeAnnouncer
	// current 는 전이 뒤 재읽기 — 기본은 원장. 시험이 재읽기 실패를 꽂을 수 있게 필드로 둠.
	current func(ctx context.Context, accountRef string) (journal.ModeSnapshot, error)
}

// ModeOperations 는 이 엔진의 핸들로 완화 표면을 만듦. 하나라도 없으면 만들지 않음 — nil 게이트 위의 표면은 원장만 고치고
// 진입은 막힌 채로 남기는, 이 표면이 없애려는 상태 그 자체임.
func (c *Context) ModeOperations() (*ModeOperations, error) {
	if c == nil || c.Journal == nil || c.Notifier == nil || c.Entry == nil || strings.TrimSpace(c.AccountRef) == "" {
		return nil, fmt.Errorf("%w: the mode release surface needs the journal, the notifier, the entry gate "+
			"and the account", ErrRuntimeUnavailable)
	}
	var auditor journal.ModeAuditor
	if c.Audit != nil {
		auditor = c.Audit
	}
	return newModeOperations(c.Journal, c.Entry, c.Notifier, auditor, c.AccountRef), nil
}

func newModeOperations(j *journal.Journal, gate *execgw.EntryGate, notifier *obs.Notifier,
	auditor journal.ModeAuditor, accountRef string) *ModeOperations {
	return &ModeOperations{
		journal: j, gate: gate, notifier: notifier, auditor: auditor, accountRef: accountRef,
		announcer: obs.RecordOnly{N: notifier},
		current:   j.CurrentOperatingMode,
	}
}

// Release 는 운영자의 모드 완화 하나를 실행하고 다시 읽은 상태를 돌려줌.
func (o *ModeOperations) Release(ctx context.Context, req ModeReleaseRequest) (ModeReleaseResult, error) {
	to := strings.ToUpper(strings.TrimSpace(req.To))
	operator := strings.TrimSpace(req.Operator)
	approval := strings.TrimSpace(req.Approval)
	reason := strings.TrimSpace(req.Reason)
	switch {
	case to != journal.ModeNormal && to != journal.ModeEntryBlocked:
		// HALT_ALL 로의 조이기는 이 명령의 일이 아님(design D0.3f 2).
		return ModeReleaseResult{}, fmt.Errorf("%w: --to must be %s or %s", ErrModeReleaseInvalid,
			journal.ModeNormal, journal.ModeEntryBlocked)
	case operator == "", approval == "", reason == "":
		return ModeReleaseResult{}, fmt.Errorf("%w: the operator, the approval reference and the reason are all required",
			ErrModeReleaseInvalid)
	}
	// (*audit.Log)(nil).RecordAction 은 nil 을 돌려주므로 nil 로그를 Auditor 로 넘기면 audit 줄 없이 커밋됨 — 인터페이스
	// 안의 타입 있는 nil 까지 거절함.
	if o.auditor == nil {
		return ModeReleaseResult{}, fmt.Errorf("%w: the engine has no audit log", ErrModeReleaseUnavailable)
	}
	if log, ok := o.auditor.(*audit.Log); ok && log == nil {
		return ModeReleaseResult{}, fmt.Errorf("%w: the engine has no audit log", ErrModeReleaseUnavailable)
	}

	rec, changed, err := o.journal.TransitionOperatingMode(ctx, journal.TransitionModeRequest{
		AccountRef: o.accountRef,
		Mode:       to,
		Cause:      reason + " | operator: " + operator,
		Actor:      journal.ModeActorOperator,
		Approval:   approval,
		Auditor:    o.auditor,
		Announcer:  o.announcer,
	})
	result := ModeReleaseResult{Changed: changed}
	switch {
	case err != nil && errors.Is(err, journal.ErrModeAnnouncementFailed):
		// 커밋은 됐고 통지 기록만 실패함 — 성공으로 보고하고 통지 실패를 함께 보임(델타). 실패한 기록은 입구가 이미 래치 · 승격함.
		result.NotifyError = err.Error()
	case err != nil:
		return ModeReleaseResult{}, err
	case changed:
		result.Notified = true
	}
	if changed {
		result.TransitionID = rec.ID
	}
	// 재읽기 — 같은 호출 안의 다른 경로(통지 기록 실패의 승격 등)가 방금 푼 모드를 다시 조일 수 있음(K16).
	// 재읽기가 실패해도 커밋 사실은 버리지 않음(26라운드 codex #5) — 오류 대신 결과에 실패를 싣고, 읽지 못한 상태는 비워 둠.
	current, err := o.current(ctx, o.accountRef)
	if err != nil {
		result.ReReadError = "re-reading the operating mode after the release failed: " + err.Error()
		return result, nil
	}
	result.Mode, result.Seq = current.Mode, current.Seq
	result.EntryBlocks = entryBlockReasons(o.gate)
	if changed {
		pending, err := o.journal.PendingAlerts(ctx, 0)
		if err != nil {
			result.ReReadError = "re-reading the release notice failed: " + err.Error()
			return result, nil
		}
		key := obs.OperatingModeEventKey(rec)
		for _, row := range pending {
			if row.EventKey == key {
				result.NoticePending = true
				break
			}
		}
	}
	return result, nil
}

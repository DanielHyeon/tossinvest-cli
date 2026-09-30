package engine

// attempt_thaw_command.go 는 a094 D−4.5 park 해동 명령을 엔진 프로세스 안에서 실행함.
//
// # 왜 있나
//
// park(UNRESOLVED_IN_DOUBT)의 유일한 출구는 Journal.OperatorResolve 이고, 그 비시험 호출자가 0 이었음 — 자동 경로가 조이는
// 상태를 사람이 풀 문이 코드에 없었음(「해제 경로 부재」 셋째 사례). park 된 발의 위에서는 손절이 나가지 않으므로(D−3.2) 이
// 문이 없으면 그 포지션의 손절은 영구히 멈춤.
//
// # 규칙(사용자 승인 해제 원칙 2026-09-28 · a066/a092 완화 명령 가족)
//
//   - 운영자 신원 · 승인 참조 · note · 목표 종결 필수, 접수 확정이면 브로커 주문 번호 필수.
//   - audit 줄이 원장 전이보다 **먼저** — audit 가 실패하면 아무것도 바꾸지 않음. 전이가 뒤에서 실패하면 보상 줄을 남김.
//   - 명령 시점에 park 가 아니면 거절(stale). OperatorResolve 자신도 그 전이만 받음(from = UNRESOLVED_IN_DOUBT).
//   - 비수용으로 닫으면 **같은 명령 안에서** 해제 판정 함수(ReleaseUnacceptedExitProposal)로 발의를 풂 — 판정은 한 곳.
//     두 쓰기 사이의 충돌은 다음 기동 따라잡기가 닫음.
//   - 엔진 lock 을 잡지 않음 · 엔진을 멈추게 하지 않음 · 콘솔 버튼 없음. tossctl 명령은 mutating: true.
//
// 브로커 호출 0 — 운영자가 확인한 값을 받음.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attemptthaw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// AuditActionAttemptThaw 는 해동의 audit 동작 이름임.
const AuditActionAttemptThaw = "attempt_thaw"

// attemptThawRepository 는 엔진 journal 이 가진 해동 면임 — 없는 빌드는 ErrUnwired.
type attemptThawRepository interface {
	LookupAttempt(context.Context, string) (journal.AttemptRecord, error)
	LookupIntent(context.Context, string) (journal.Intent, error)
	OperatorResolve(ctx context.Context, attemptID string, to journal.AttemptState, operator, brokerOrderID, note string) error
	ArmedExitProposals(context.Context, string) ([]journal.ArmedExitProposal, error)
	ReleaseUnacceptedExitProposal(ctx context.Context, positionID, intentID string,
		resolution journal.ProposalResolution) (journal.ExitIntentVerdict, bool, error)
}

// ResolveParkedAttempt 는 운영자가 확인한 park 된 attempt 하나를 종결하고, 비수용이면 그것이 무장한 발의를 풂.
func (s *PositionPolicyCommandService) ResolveParkedAttempt(ctx context.Context, req attemptthaw.Request) (attemptthaw.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	repo, ok := s.j.(attemptThawRepository)
	if !ok {
		return attemptthaw.Result{}, attemptthaw.ErrUnwired
	}
	if s.audit == nil {
		return attemptthaw.Result{}, attemptthaw.ErrAuditUnavailable
	}
	to, err := thawTarget(req)
	if err != nil {
		return attemptthaw.Result{}, err
	}
	attemptID := strings.TrimSpace(req.AttemptID)
	rec, err := repo.LookupAttempt(ctx, attemptID)
	if err != nil {
		if errors.Is(err, journal.ErrAttemptNotFound) {
			return attemptthaw.Result{}, fmt.Errorf("%w: %s", attemptthaw.ErrNotFound, attemptID)
		}
		return attemptthaw.Result{}, err
	}
	if rec.State != journal.StateUnresolvedInDoubt {
		return attemptthaw.Result{}, fmt.Errorf("%w: attempt %s is %s", attemptthaw.ErrStale, attemptID, rec.State)
	}

	operator, approval, note := strings.TrimSpace(req.Operator), strings.TrimSpace(req.Approval), strings.TrimSpace(req.Note)
	setting := "attempt:" + attemptID
	detail := fmt.Sprintf("operator %s closes parked attempt %s as %s (approval: %s; broker order: %s): %s",
		operator, attemptID, to, approval, strings.TrimSpace(req.BrokerOrderID), note)
	// audit 줄이 먼저 — 실패하면 원장을 건드리지 않음.
	if err := s.audit.RecordAction(AuditActionAttemptThaw, setting, "attempt", detail); err != nil {
		return attemptthaw.Result{}, fmt.Errorf("%w: %w", attemptthaw.ErrAuditUnavailable, err)
	}
	journalNote := fmt.Sprintf("%s (approval: %s)", note, approval)
	if err := repo.OperatorResolve(ctx, attemptID, to, operator, strings.TrimSpace(req.BrokerOrderID), journalNote); err != nil {
		_ = s.audit.RecordAction(AuditActionAttemptThaw, setting, "not_committed", "the ledger transition failed after the attempt line: "+err.Error())
		var stateErr *journal.StateError
		if errors.As(err, &stateErr) || errors.Is(err, journal.ErrIllegalTransition) {
			return attemptthaw.Result{}, fmt.Errorf("%w: %w", attemptthaw.ErrStale, err)
		}
		if errors.Is(err, journal.ErrInvalidRequest) {
			return attemptthaw.Result{}, fmt.Errorf("%w: %w", attemptthaw.ErrInvalidRequest, err)
		}
		return attemptthaw.Result{}, err
	}
	result := attemptthaw.Result{AttemptID: attemptID, State: string(to)}
	if to != journal.StateFailedConfirmed {
		// 접수 확정 — 발의를 끝내는 것은 체결 경로의 일임(spec: CONFIRMED 면 해제하지 않음).
		return result, nil
	}
	s.releaseThawedProposal(ctx, repo, rec, &result)
	return result, nil
}

// releaseThawedProposal 은 그 attempt 의 intent 를 무장한 발의를 해제 판정 함수로 풂. 실패는 결과에만 남김 — 해소는
// 이미 커밋됐고, 다음 기동 따라잡기가 같은 판정으로 푼다(D−2.5).
func (s *PositionPolicyCommandService) releaseThawedProposal(ctx context.Context, repo attemptThawRepository,
	rec journal.AttemptRecord, result *attemptthaw.Result) {
	intent, err := repo.LookupIntent(ctx, rec.IntentID)
	if err != nil {
		result.ReleaseError = err.Error()
		return
	}
	armed, err := repo.ArmedExitProposals(ctx, intent.AccountRef)
	if err != nil {
		result.ReleaseError = err.Error()
		return
	}
	for _, a := range armed {
		if a.IntentID != rec.IntentID {
			continue
		}
		result.PositionID = a.PositionID
		_, released, err := repo.ReleaseUnacceptedExitProposal(ctx, a.PositionID, a.IntentID, journal.ProposalRefused)
		if err != nil {
			result.ReleaseError = err.Error()
			return
		}
		result.ProposalReleased = released
		return
	}
}

func thawTarget(req attemptthaw.Request) (journal.AttemptState, error) {
	for _, f := range []struct{ name, value string }{
		{"attempt", req.AttemptID}, {"operator", req.Operator}, {"approval", req.Approval}, {"note", req.Note},
	} {
		if strings.TrimSpace(f.value) == "" {
			return "", fmt.Errorf("%w: the %s is required", attemptthaw.ErrInvalidRequest, f.name)
		}
	}
	switch strings.ToUpper(strings.TrimSpace(req.Target)) {
	case attemptthaw.TargetFailedConfirmed:
		return journal.StateFailedConfirmed, nil
	case attemptthaw.TargetConfirmed:
		if strings.TrimSpace(req.BrokerOrderID) == "" {
			return "", fmt.Errorf("%w: confirming needs the broker order number the operator found", attemptthaw.ErrInvalidRequest)
		}
		return journal.StateConfirmed, nil
	default:
		return "", fmt.Errorf("%w: the target must be %s or %s", attemptthaw.ErrInvalidRequest,
			attemptthaw.TargetFailedConfirmed, attemptthaw.TargetConfirmed)
	}
}

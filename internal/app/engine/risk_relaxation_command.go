package engine

// risk_relaxation_command.go 는 a066 5.5(design D8) 운영자 해제를 엔진 프로세스 안에서 실행함.
//
// 왜 엔진 안인가(Manager 판정 2026-09-29 Q1, a092 완화 명령 가족 계약): 원장은 단일 writer(엔진)이고, CLI 가 원장을
// 직접 열면 새 바이너리가 도는 엔진 밑에서 마이그레이션할 수 있음. 그래서 tossctl 은 엔진 제어 endpoint 에 요청하고,
// 여기서 엔진이 이미 flock 아래 연 journal 핸들과 자기 audit 로그로 해제함. engine lock 을 잡지 않음 — 엔진을 멈추면
// (보호가 UNWIRED 인 동안) 손절이 없음(design D8).
//
// 해제 판정(OPERATOR · 승인 · 결속 · audit 가 commit 앞)은 전부 journal API 가 함 — 여기서 다시 판정하지 않음(판정이
// 둘이면 서로의 시험을 통과시킴). 이 파일이 더하는 것은 (1) audit 로그 부재 거절, (2) 운영자 이름을 사유에 붙임,
// (3) 커밋 **뒤** 원장 alert enqueue(통지)와 그 실패의 「완화됨·통지 실패」 보고, (4) 오류 어휘 변환뿐임.
//
// a079 의 격리 해제와 같이 PositionPolicyCommandService 의 선택 capability 로 붙음 — 생성자·배선 지점은 audit 필드 한
// 줄뿐임.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskrelaxation"
)

// EventRiskRelaxation 은 운영자 해제 통지의 alert 유형임. alert outbox 는 critical 전용이라(alertdelivery.go) 전달
// 실패는 진입을 잠금 — 통지가 끝내 안 가면 조이는 쪽으로 끝남.
const EventRiskRelaxation = "engine.risk_relaxation"

// riskRelaxationRepository 는 엔진 journal 이 가진 해제·통지 면임. PositionPolicyCommandService.j 에서 단언으로 찾음
// (quarantineRepo 와 같은 방식) — 없는 빌드는 ErrUnwired.
type riskRelaxationRepository interface {
	ReleaseEntryLossLock(context.Context, journal.EntryLossLockReleaseRequest) (journal.EntryLossLockReleaseRecord, error)
	ReleaseRiskOverageLatch(context.Context, journal.RiskOverageLatchReleaseRequest) (journal.RiskOverageLatchReleaseRecord, error)
	EnqueueAlert(context.Context, journal.Alert) (int64, error)
}

func (s *PositionPolicyCommandService) relaxationRepo() (riskRelaxationRepository, error) {
	repo, ok := s.j.(riskRelaxationRepository)
	if !ok {
		return nil, riskrelaxation.ErrUnwired
	}
	return repo, nil
}

// relaxationAuditor 는 엔진 audit 로그를 해제의 Auditor 로 돌려줌. nil 로그는 거절함 — (*audit.Log)(nil).RecordAction
// 은 nil 을 돌려주므로, nil 로그를 넘기면 audit 줄 없이 해제가 커밋됨.
func (s *PositionPolicyCommandService) relaxationAuditor() (journal.RiskRelaxationAuditor, error) {
	if s.audit == nil {
		return nil, riskrelaxation.ErrAuditUnavailable
	}
	return s.audit, nil
}

// ReleaseEntryLossLock 은 운영자가 본 열린 진입 손실 잠금 하나를 해제함.
func (s *PositionPolicyCommandService) ReleaseEntryLossLock(ctx context.Context,
	req riskrelaxation.EntryLockReleaseRequest) (riskrelaxation.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	repo, err := s.relaxationRepo()
	if err != nil {
		return riskrelaxation.Result{}, err
	}
	auditor, err := s.relaxationAuditor()
	if err != nil {
		return riskrelaxation.Result{}, err
	}
	operator, reason, err := relaxationOperator(req.Operator, req.Reason)
	if err != nil {
		return riskrelaxation.Result{}, err
	}
	at := s.clk.Now().UTC()
	record, err := repo.ReleaseEntryLossLock(ctx, journal.EntryLossLockReleaseRequest{
		AccountRef: strings.TrimSpace(req.AccountRef), Market: riskbucket.Market(strings.ToUpper(strings.TrimSpace(req.Market))),
		Horizon: riskbucket.Horizon(strings.ToUpper(strings.TrimSpace(req.Horizon))), LockSeq: req.LockSeq,
		ExpectedLastEvent: req.ExpectedLastEvent, Actor: journal.RelaxationActorOperator, Approval: req.Approval,
		Reason: reason, ReleasedAt: at, Auditor: auditor,
	})
	if err != nil {
		return riskrelaxation.Result{}, relaxationError(err)
	}
	scope := strings.ToUpper(strings.TrimSpace(req.Market)) + "/" + strings.ToUpper(strings.TrimSpace(req.Horizon))
	target := "entry_loss_lock:" + strings.TrimSpace(req.AccountRef) + "/" + scope
	return notifyRelaxation(ctx, repo, "entry_lock", record.ReleaseSeq, target, "entry_loss_lock:"+scope, operator,
		req.Approval, record.ReleasedAt), nil
}

// ReleaseRiskOverageLatch 는 운영자가 본 상태의 owner generation 에서 RISK_OVERAGE latch 만 해제함.
func (s *PositionPolicyCommandService) ReleaseRiskOverageLatch(ctx context.Context,
	req riskrelaxation.LatchReleaseRequest) (riskrelaxation.Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	repo, err := s.relaxationRepo()
	if err != nil {
		return riskrelaxation.Result{}, err
	}
	auditor, err := s.relaxationAuditor()
	if err != nil {
		return riskrelaxation.Result{}, err
	}
	operator, reason, err := relaxationOperator(req.Operator, req.Reason)
	if err != nil {
		return riskrelaxation.Result{}, err
	}
	owner := riskbucket.OwnerKey{AccountID: strings.TrimSpace(req.AccountRef),
		Market: riskbucket.Market(strings.ToUpper(strings.TrimSpace(req.Market))),
		Symbol: strings.TrimSpace(req.Symbol), ProspectiveGeneration: strings.TrimSpace(req.Generation)}
	record, err := repo.ReleaseRiskOverageLatch(ctx, journal.RiskOverageLatchReleaseRequest{
		Owner: owner, ExpectedStateDigest: req.ExpectedState, Actor: journal.RelaxationActorOperator,
		Approval: req.Approval, Reason: reason, ReleasedAt: s.clk.Now().UTC(), Auditor: auditor,
	})
	if err != nil {
		return riskrelaxation.Result{}, relaxationError(err)
	}
	scope := fmt.Sprintf("%s/%s/%s", owner.Market, owner.Symbol, owner.ProspectiveGeneration)
	target := "risk_owner:" + owner.AccountID + "/" + scope
	return notifyRelaxation(ctx, repo, "overage_latch", record.ReleaseSeq, target, "risk_owner:"+scope, operator,
		req.Approval, record.ReleasedAt), nil
}

// relaxationOperator 는 운영자 이름을 요구하고 사유에 붙임 — journal 의 해제 기록에는 운영자 칸이 없고(actor 는
// OPERATOR 역할), audit 줄의 detail 이 사유를 담으므로 사람이 누구인지는 사유로 남음.
func relaxationOperator(operator, reason string) (string, string, error) {
	operator = strings.TrimSpace(operator)
	if operator == "" {
		return "", "", fmt.Errorf("%w: the operator is named", riskrelaxation.ErrInvalidRequest)
	}
	if strings.TrimSpace(reason) == "" {
		// 빈 사유는 journal 이 거절함 — 이름을 붙여 비지 않은 사유로 만들지 않음.
		return operator, "", nil
	}
	return operator, strings.TrimSpace(reason) + " (operator " + operator + ")", nil
}

// relaxationError 는 journal 판정을 전송 어휘로 옮김. 원래 오류는 감싸 둠(문구 보존).
func relaxationError(err error) error {
	switch {
	case errors.Is(err, journal.ErrRiskRelaxationStale):
		return fmt.Errorf("%w: %w", riskrelaxation.ErrStale, err)
	case errors.Is(err, journal.ErrRiskRelaxationAuditFailed):
		return fmt.Errorf("%w: %w", riskrelaxation.ErrAuditUnavailable, err)
	case errors.Is(err, journal.ErrRiskBucketReplayMismatch):
		return fmt.Errorf("%w: %w", riskrelaxation.ErrStateMismatch, err)
	case errors.Is(err, journal.ErrRiskRelaxationRequiresOperator), errors.Is(err, journal.ErrRiskRelaxationApprovalRequired),
		errors.Is(err, journal.ErrInvalidRequest):
		return fmt.Errorf("%w: %w", riskrelaxation.ErrInvalidRequest, err)
	}
	return err
}

// notifyRelaxation 은 커밋된 해제를 원장 alert 로 enqueue 함. 실패해도 해제는 유효함 — 결과가 「완화됨·통지 실패」를
// 말함(Notified=false). 요청 문맥이 끊겨도 통지는 기록되도록 취소를 떼어 냄.
//
// target 은 계좌를 포함한 전체 대상(CLI 결과·원장 payload)이고, published 는 외부 전송으로 나가는 제목·본문용 대상임 —
// 계좌 식별자를 빼고 시장·범위만 말함(안전 불변식 8; 다른 엔진 alert 제목도 종목만 실음).
func notifyRelaxation(ctx context.Context, repo riskRelaxationRepository, kind string, seq int64, target, published,
	operator, approval string, at time.Time) riskrelaxation.Result {
	result := riskrelaxation.Result{ReleaseSeq: seq, Target: target, ReleasedAt: journal.RFC3339(at)}
	payload, _ := json.Marshal(map[string]any{
		"kind": kind, "target": target, "release_seq": seq, "operator": operator,
		"approval": strings.TrimSpace(approval), "released_at": result.ReleasedAt,
	})
	_, err := repo.EnqueueAlert(context.WithoutCancel(ctx), journal.Alert{
		EventKey: EventRiskRelaxation + "|" + kind + "|" + strconv.FormatInt(seq, 10),
		Type:     EventRiskRelaxation,
		Severity: "critical",
		Title:    "RISK RELAXATION: " + published,
		Body: fmt.Sprintf("operator %s released %s (release %d, approval: %s)", operator, published, seq,
			strings.TrimSpace(approval)),
		Payload: string(payload),
	})
	if err != nil {
		result.NotifyError = err.Error()
		return result
	}
	result.Notified = true
	return result
}

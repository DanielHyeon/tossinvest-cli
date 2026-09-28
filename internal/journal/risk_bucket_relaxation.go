package journal

// a066 task 5.5 — 완화(해제) 메커니즘(사용자 결정 2026-09-28, design D8).
//
// 원칙: 자동 경로는 조이기만 함. 완화·해제는 OPERATOR 가 승인 참조 문자열과 함께 요청하고, audit 줄이 commit **앞**에
// 기록되며(audit 실패 = 아무것도 안 바뀜), 동시 조이기는 보수 쪽이 이김 — 해제는 승인자가 본 상태에 결속되고 그 뒤
// 상태가 바뀌었으면 stale 로 거절됨. 진입점은 이 journal API 와 tossctl mutating 명령뿐이고 콘솔 버튼은 없음.
// 형태는 TransitionOperatingMode(operating_mode.go:346–470)를 따름: 방향·결속은 트랜잭션 안에서 현재 상태로 판정함.
//
// engine lock 을 잡지 않음: BEGIN IMMEDIATE 와 트랜잭션 내 재판정이 직렬성을 주고, lock 을 잡으려면 엔진을 멈춰야
// 해서(손절 부재 — 보호가 UNWIRED 인 동안 손절은 엔진이 살아 있을 때뿐) 해제가 손절 연속성을 깨는 조건이 됨.

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "embed"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

// schemaV35 는 완화 기록 표와 "열린 잠금 1건" 트리거임(a066 5.5, D8).
//
//go:embed risk_bucket_relaxation_v35.sql
var schemaV35 string

// RelaxationActorOperator 는 완화를 요청할 수 있는 유일한 actor 임.
const RelaxationActorOperator = "OPERATOR"

// Audit action 문자열 — journal 이 사건을 소유하므로 문자열도 journal 이 소유함(operating_mode.transition 선례).
const (
	AuditActionEntryLockRelease    = "risk_bucket.entry_lock_release"
	AuditActionOverageLatchRelease = "risk_bucket.overage_latch_release"
)

var (
	// ErrRiskRelaxationRequiresOperator 는 OPERATOR 가 아닌 actor 의 완화 요청임 — 자동 경로는 조이기만 함.
	ErrRiskRelaxationRequiresOperator = errors.New("journal: a risk relaxation is an operator act; automatic paths only tighten")
	// ErrRiskRelaxationApprovalRequired 는 사람 승인 참조가 없는 완화 요청임.
	ErrRiskRelaxationApprovalRequired = errors.New("journal: a risk relaxation names the human approval it rests on")
	// ErrRiskRelaxationStale 는 승인자가 본 상태 뒤에 상태가 바뀐(더 조여졌을 수 있는) 해제임 — 보수 쪽이 이김.
	ErrRiskRelaxationStale = errors.New("journal: the state changed after the approver read it; the release is refused and must be re-approved on the current state")
)

// RiskRelaxationAuditor 는 완화의 audit 줄을 받음. commit 앞에 불리고, 실패하면 완화는 되돌려짐.
type RiskRelaxationAuditor interface {
	RecordAction(action, setting, value, detail string) error
}

// validRelaxation 은 완화 요청의 공통 모양(actor · 승인 · 사유 · audit · 시각)을 판정함.
func validRelaxation(actor, approval, reason string, auditor RiskRelaxationAuditor, at time.Time) error {
	switch {
	case strings.TrimSpace(actor) != RelaxationActorOperator:
		return fmt.Errorf("%w: actor %q", ErrRiskRelaxationRequiresOperator, actor)
	case strings.TrimSpace(approval) == "":
		return ErrRiskRelaxationApprovalRequired
	case strings.TrimSpace(reason) == "":
		return fmt.Errorf("%w: a relaxation names why; a change nobody can explain afterwards is not auditable", ErrInvalidRequest)
	case auditor == nil:
		return fmt.Errorf("%w: a relaxation requires an audit log (§0.5)", ErrInvalidRequest)
	case at.IsZero():
		return fmt.Errorf("%w: a relaxation needs its time", ErrInvalidRequest)
	}
	return nil
}

// EntryLossLockView 는 열린 잠금 하나와 그 잠금의 마지막 REAFFIRM 사건 번호(없으면 0)임 — 해제가 결속할 값.
type EntryLossLockView struct {
	Lock      EntryLossLock
	LastEvent int64
}

// ReadEntryLossLocks 는 계좌의 열린 잠금을 읽음.
func (j *Journal) ReadEntryLossLocks(ctx context.Context, account string) ([]EntryLossLockView, error) {
	if j == nil || j.db == nil {
		return nil, fmt.Errorf("%w: journal unavailable", ErrInvalidRequest)
	}
	return readEntryLossLocks(ctx, j.db, account)
}

// ReadEntryLossLocks 는 읽기 전용 연결로 계좌의 열린 잠금을 읽음(tossctl risk-latch-show — 원장을 쓰거나 이주하지 않음).
func (r *ReadOnly) ReadEntryLossLocks(ctx context.Context, account string) ([]EntryLossLockView, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("%w: journal unavailable", ErrInvalidRequest)
	}
	return readEntryLossLocks(ctx, r.db, account)
}

func readEntryLossLocks(ctx context.Context, q riskBucketQueryer, account string) ([]EntryLossLockView, error) {
	rows, err := q.QueryContext(ctx, `SELECT l.lock_seq,l.account_ref,l.market,l.horizon,l.cause,l.activated_at,
		COALESCE((SELECT MAX(e.event_seq) FROM risk_bucket_entry_loss_lock_events e WHERE e.lock_seq=l.lock_seq),0)
		FROM risk_bucket_entry_loss_locks l WHERE l.account_ref=?
		AND NOT EXISTS (SELECT 1 FROM risk_bucket_entry_loss_lock_releases r WHERE r.lock_seq=l.lock_seq)
		ORDER BY l.market,l.horizon`, strings.TrimSpace(account))
	if err != nil {
		return nil, fmt.Errorf("journal: reading entry loss locks: %w", err)
	}
	defer rows.Close()
	var views []EntryLossLockView
	for rows.Next() {
		var view EntryLossLockView
		var market, horizon, activatedAt string
		if err := rows.Scan(&view.Lock.Seq, &view.Lock.AccountRef, &market, &horizon, &view.Lock.Cause, &activatedAt, &view.LastEvent); err != nil {
			return nil, fmt.Errorf("journal: reading entry loss locks: %w", err)
		}
		parsed, err := time.Parse(time.RFC3339Nano, activatedAt)
		if err != nil {
			return nil, fmt.Errorf("journal: entry loss lock activated_at: %w", err)
		}
		view.Lock.Market, view.Lock.Horizon, view.Lock.ActivatedAt = riskbucket.Market(market), riskbucket.Horizon(horizon), parsed.UTC()
		views = append(views, view)
	}
	return views, rows.Err()
}

// EntryLossLockReleaseRequest 는 열린 잠금 하나의 해제 요청임. LockSeq 와 ExpectedLastEvent 는 승인자가 본 상태임.
type EntryLossLockReleaseRequest struct {
	AccountRef        string
	Market            riskbucket.Market
	Horizon           riskbucket.Horizon
	LockSeq           int64
	ExpectedLastEvent int64
	Actor             string
	Approval          string
	Reason            string
	ReleasedAt        time.Time
	Auditor           RiskRelaxationAuditor
}

// EntryLossLockReleaseRecord 는 기록된 해제임.
type EntryLossLockReleaseRecord struct {
	ReleaseSeq, LockSeq int64
	ReleasedAt          time.Time
}

// ReleaseEntryLossLock 은 승인자가 본 열린 잠금을 해제함. 트랜잭션 안에서 그 잠금이 여전히 그 범위의 열린 잠금이고
// 마지막 REAFFIRM 이 승인자가 본 것과 같을 때만 해제하고, 아니면 ErrRiskRelaxationStale(보수 쪽 승리).
func (j *Journal) ReleaseEntryLossLock(ctx context.Context, req EntryLossLockReleaseRequest) (EntryLossLockReleaseRecord, error) {
	if j == nil || j.db == nil {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("%w: journal unavailable", ErrInvalidRequest)
	}
	if err := validRelaxation(req.Actor, req.Approval, req.Reason, req.Auditor, req.ReleasedAt); err != nil {
		return EntryLossLockReleaseRecord{}, err
	}
	if !validEntryLossLockScope(req.AccountRef, req.Market, req.Horizon) || req.LockSeq <= 0 || req.ExpectedLastEvent < 0 {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("%w: entry loss lock release needs its scope and the lock it saw", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("journal: begin entry loss lock release: %w", err)
	}
	defer tx.Rollback()
	open, found, err := activeEntryLossLock(ctx, tx, req.AccountRef, req.Market, req.Horizon)
	if err != nil {
		return EntryLossLockReleaseRecord{}, err
	}
	if !found || open.Seq != req.LockSeq {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("%w: lock %d is not the open lock of %s/%s/%s", ErrRiskRelaxationStale,
			req.LockSeq, req.AccountRef, req.Market, req.Horizon)
	}
	var lastEvent int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_seq),0) FROM risk_bucket_entry_loss_lock_events WHERE lock_seq=?`, req.LockSeq).Scan(&lastEvent); err != nil {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("journal: reading entry loss lock events: %w", err)
	}
	if lastEvent != req.ExpectedLastEvent {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("%w: lock %d was reaffirmed after the approval (last event %d, approved at %d)",
			ErrRiskRelaxationStale, req.LockSeq, lastEvent, req.ExpectedLastEvent)
	}
	releasedAt := formatJournalTime(req.ReleasedAt)
	result, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_entry_loss_lock_releases(lock_seq,actor,approval,reason,expected_last_event,released_at)
		VALUES(?,?,?,?,?,?)`, req.LockSeq, RelaxationActorOperator, strings.TrimSpace(req.Approval), strings.TrimSpace(req.Reason), req.ExpectedLastEvent, releasedAt)
	if err != nil {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("journal: recording entry loss lock release: %w", err)
	}
	seq, err := result.LastInsertId()
	if err != nil {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("journal: reading entry loss lock release identity: %w", err)
	}
	// audit 는 commit 앞 — 기록 없는 완화가 이 경로가 불가능하게 만들려는 상태임(operating_mode 선례).
	setting := fmt.Sprintf("entry_loss_lock:%s/%s/%s", req.AccountRef, req.Market, req.Horizon)
	detail := fmt.Sprintf("%s — %s | approved-by: %s (lock %d: %s; last event %d)", RelaxationActorOperator,
		strings.TrimSpace(req.Reason), strings.TrimSpace(req.Approval), req.LockSeq, open.Cause, lastEvent)
	if err := req.Auditor.RecordAction(AuditActionEntryLockRelease, setting, "released", detail); err != nil {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("journal: recording the entry loss lock release in the audit log (nothing was changed): %w", err)
	}
	if err := tx.Commit(); err != nil {
		return EntryLossLockReleaseRecord{}, fmt.Errorf("journal: commit entry loss lock release: %w", err)
	}
	return EntryLossLockReleaseRecord{ReleaseSeq: seq, LockSeq: req.LockSeq, ReleasedAt: req.ReleasedAt.UTC().Truncate(time.Second)}, nil
}

// RiskOwnerLatchView 는 활성 owner 하나의 latch 와 마지막 상태 봉인 digest 임 — latch 해제가 결속할 값.
type RiskOwnerLatchView struct {
	Owner                          riskbucket.OwnerKey
	OverageLatched, UnknownLatched bool
	StateDigest                    string
}

// ReadRiskOwnerLatches 는 계좌·시장의 활성 owner 들의 latch 와 봉인 digest 를 읽음.
func (j *Journal) ReadRiskOwnerLatches(ctx context.Context, account string, market riskbucket.Market) ([]RiskOwnerLatchView, error) {
	if j == nil || j.db == nil {
		return nil, fmt.Errorf("%w: journal unavailable", ErrInvalidRequest)
	}
	return readRiskOwnerLatches(ctx, j.db, account, market)
}

// ReadRiskOwnerLatches 는 읽기 전용 연결로 owner latch 와 봉인 digest 를 읽음(tossctl risk-latch-show).
func (r *ReadOnly) ReadRiskOwnerLatches(ctx context.Context, account string, market riskbucket.Market) ([]RiskOwnerLatchView, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("%w: journal unavailable", ErrInvalidRequest)
	}
	return readRiskOwnerLatches(ctx, r.db, account, market)
}

func readRiskOwnerLatches(ctx context.Context, q riskBucketQueryer, account string, market riskbucket.Market) ([]RiskOwnerLatchView, error) {
	rows, err := q.QueryContext(ctx, `SELECT o.account_ref,o.market,o.symbol,o.prospective_generation,o.risk_overage_latched,o.unknown_actual_latched,
		COALESCE((SELECT s.state_digest FROM risk_bucket_state_snapshots s WHERE s.account_ref=o.account_ref AND s.market=o.market AND s.symbol=o.symbol
			AND s.prospective_generation=o.prospective_generation ORDER BY s.event_sequence DESC LIMIT 1),'')
		FROM risk_bucket_owners o WHERE o.account_ref=? AND o.market=? AND o.released_at IS NULL ORDER BY o.symbol,o.prospective_generation`,
		strings.TrimSpace(account), string(market))
	if err != nil {
		return nil, fmt.Errorf("journal: reading risk owner latches: %w", err)
	}
	defer rows.Close()
	var views []RiskOwnerLatchView
	for rows.Next() {
		var view RiskOwnerLatchView
		var overage, unknown int
		var marketText string
		if err := rows.Scan(&view.Owner.AccountID, &marketText, &view.Owner.Symbol, &view.Owner.ProspectiveGeneration, &overage, &unknown, &view.StateDigest); err != nil {
			return nil, fmt.Errorf("journal: reading risk owner latches: %w", err)
		}
		view.Owner.Market = riskbucket.Market(marketText)
		view.OverageLatched, view.UnknownLatched = overage != 0, unknown != 0
		views = append(views, view)
	}
	return views, rows.Err()
}

// RiskOverageLatchReleaseRequest 는 owner generation 하나의 RISK_OVERAGE latch 해제 요청임. ExpectedStateDigest 는
// 승인자가 본 마지막 상태 봉인임.
type RiskOverageLatchReleaseRequest struct {
	Owner               riskbucket.OwnerKey
	ExpectedStateDigest string
	Actor               string
	Approval            string
	Reason              string
	ReleasedAt          time.Time
	Auditor             RiskRelaxationAuditor
}

// RiskOverageLatchReleaseRecord 는 기록된 latch 해제임.
type RiskOverageLatchReleaseRecord struct {
	ReleaseSeq int64
	Owner      riskbucket.OwnerKey
	ReleasedAt time.Time
}

// ReleaseRiskOverageLatch 는 승인자가 본 상태의 owner 에서 RISK_OVERAGE 만 해제함(UNKNOWN_ACTUAL_RISK 와 overage 수치는
// 남음). 그 사이 체결·latch 가 끼어 봉인이 바뀌었으면 ErrRiskRelaxationStale. 해제 뒤 상태를 다시 봉인하고, 다음
// 체결이 여전히 한도 위면 recomputeOverageLatches 가 다시 latch 함. 이 해제가 먼저이고 owner 해제(releaseRiskBucketOwner
// 의 owner_latch 검사)는 그 뒤에야 열림.
func (j *Journal) ReleaseRiskOverageLatch(ctx context.Context, req RiskOverageLatchReleaseRequest) (RiskOverageLatchReleaseRecord, error) {
	if j == nil || j.db == nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("%w: journal unavailable", ErrInvalidRequest)
	}
	if err := validRelaxation(req.Actor, req.Approval, req.Reason, req.Auditor, req.ReleasedAt); err != nil {
		return RiskOverageLatchReleaseRecord{}, err
	}
	key := req.Owner
	if strings.TrimSpace(key.AccountID) == "" || key.AccountID != strings.TrimSpace(key.AccountID) || strings.TrimSpace(key.Symbol) == "" ||
		strings.TrimSpace(key.ProspectiveGeneration) == "" || (key.Market != riskbucket.MarketKR && key.Market != riskbucket.MarketUS) ||
		strings.TrimSpace(req.ExpectedStateDigest) == "" {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("%w: overage latch release needs its owner and the state digest it saw", ErrInvalidRequest)
	}
	tx, err := j.db.BeginTx(ctx, nil) // BEGIN IMMEDIATE
	if err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: begin overage latch release: %w", err)
	}
	defer tx.Rollback()
	var overage int
	err = tx.QueryRowContext(ctx, `SELECT risk_overage_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND released_at IS NULL`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&overage)
	if errors.Is(err, sql.ErrNoRows) {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("%w: no active owner %s/%s/%s/%s", ErrRiskRelaxationStale, key.AccountID, key.Market, key.Symbol, key.ProspectiveGeneration)
	}
	if err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: reading owner latch: %w", err)
	}
	if overage == 0 {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("%w: owner carries no RISK_OVERAGE latch", ErrRiskRelaxationStale)
	}
	// 승인자가 본 봉인과 현재 봉인이 같아야 하고, 현재 원장이 그 봉인과 맞아야 함(판정한 바이트에 결속).
	var persisted string
	if err := tx.QueryRowContext(ctx, `SELECT state_digest FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? ORDER BY event_sequence DESC LIMIT 1`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&persisted); err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: reading owner state seal: %w", err)
	}
	if persisted != strings.TrimSpace(req.ExpectedStateDigest) {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("%w: owner state was resealed after the approval", ErrRiskRelaxationStale)
	}
	if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {
		return RiskOverageLatchReleaseRecord{}, err
	}
	releasedAt := formatJournalTime(req.ReleasedAt)
	if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_owners SET risk_overage_latched=0 WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: clearing owner RISK_OVERAGE: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations SET risk_overage_latched=0 WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`,
		key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: clearing reservation RISK_OVERAGE: %w", err)
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_latch_releases(account_ref,market,symbol,prospective_generation,latch,expected_state_digest,actor,approval,reason,released_at)
		VALUES(?,?,?,?,'RISK_OVERAGE',?,?,?,?,?)`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration, persisted,
		RelaxationActorOperator, strings.TrimSpace(req.Approval), strings.TrimSpace(req.Reason), releasedAt)
	if err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: recording overage latch release: %w", err)
	}
	seq, err := result.LastInsertId()
	if err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: reading overage latch release identity: %w", err)
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("overage-latch-release\x00%d\x00%s", seq, persisted)))
	eventDigest := hex.EncodeToString(sum[:])
	if err := j.recordRiskBucketStateTx(ctx, tx, key, "OVERAGE_LATCH_RELEASED", "overage-latch-release-"+eventDigest[:24], eventDigest, releasedAt); err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: resealing owner state after the latch release: %w", err)
	}
	setting := fmt.Sprintf("risk_owner:%s/%s/%s/%s", key.AccountID, key.Market, key.Symbol, key.ProspectiveGeneration)
	detail := fmt.Sprintf("%s — %s | approved-by: %s (RISK_OVERAGE; bound state %s)", RelaxationActorOperator,
		strings.TrimSpace(req.Reason), strings.TrimSpace(req.Approval), persisted)
	if err := req.Auditor.RecordAction(AuditActionOverageLatchRelease, setting, "released", detail); err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: recording the overage latch release in the audit log (nothing was changed): %w", err)
	}
	if err := tx.Commit(); err != nil {
		return RiskOverageLatchReleaseRecord{}, fmt.Errorf("journal: commit overage latch release: %w", err)
	}
	return RiskOverageLatchReleaseRecord{ReleaseSeq: seq, Owner: key, ReleasedAt: req.ReleasedAt.UTC().Truncate(time.Second)}, nil
}

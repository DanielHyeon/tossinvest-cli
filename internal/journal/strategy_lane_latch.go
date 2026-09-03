package journal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// 이 파일은 전략 레인 하나의 **신규 진입 잠금**을 프로세스보다 오래 살게 한다.
//
// 왜 원장인가. 이 기록의 고장 방식은 이미 분류되어 있다 — 저널 고장은 신규
// 진입을 막지만 엔진을 세우지는 않는다는 것을 a112 의
// `a112_fault_classification_test.go` 가 값으로 확인해 두었다. 별도 저장소를
// 만들면 그 분류를 처음부터 다시 정해야 하고, 기계(트랜잭션·백업·불변 트리거)도
// 다시 만들어야 한다.
//
// 두 테이블이고 둘 다 append-only 다. 가변 플래그 하나면 "이 레인이 잠긴 적이
// 있었다"가 복구하는 순간 사라지는데, 운영자의 질문은 대개 "왜 멈췄고 누가
// 열었나"이다.

// ErrStrategyLaneLatchRecoveryEvidence 는 복구 증거가 모자랄 때의 답이다.
//
// 조용한 no-op 이 아니라 오류인 이유: 복구를 요청한 쪽은 자기가 무엇을 열었다고
// 믿는다. 아무 일도 안 일어난 것과 열린 것을 같은 답으로 돌려주면 그 믿음을
// 아무도 정정하지 못한다.
var ErrStrategyLaneLatchRecoveryEvidence = errors.New(
	"journal: strategy lane latch recovery needs a strictly newer signed activation generation")

// ErrStrategyLaneLatchInvalid 는 잠금 기록이 계약을 벗어났다는 답이다.
var ErrStrategyLaneLatchInvalid = errors.New("journal: invalid strategy lane latch record")

// strategyLaneFamilies 는 이 테이블이 받는 가족 이름이다. 골든 산문이 아니라
// `strategyrouter/production.go:57-60` 에서 읽었고, 스키마의 CHECK 제약과 같은
// 넷이어야 한다 — 갈라지면 Go 쪽은 통과시키고 SQLite 가 거절한다.
var strategyLaneFamilies = map[string]struct{}{
	"CONTINUATION": {}, "REVERSAL": {}, "WEEKLY_VALUE": {}, "BREAKOUT_RETEST": {},
}

// StrategyLaneLatch 는 레인 하나가 신규 진입을 잠근 기록이다.
//
// 이것은 복구 영수증이 아니다. 이 값에는 진입을 다시 여는 방법이 없고, 여는
// 방법은 `RecoverStrategyLaneLatch` 하나이며 그것은 더 큰 서명 활성화 세대를
// 요구한다.
type StrategyLaneLatch struct {
	// Seq 는 원장이 정하는 이 잠금의 신원이다. 레인이 만든 LatchID 를 신원으로
	// 쓰지 않는 이유: 복구된 레인이 다시 잠기면 그 문자열이 되풀이된다.
	Seq                  int64
	AccountRef           string
	Market               string
	Family               string
	LaneID               string
	LaneVersion          string
	LatchID              string
	LatchRevision        uint64
	Reason               string
	Abnormal             bool
	ActivationGeneration uint64
	ObservedAt           time.Time
}

func (latch StrategyLaneLatch) valid() bool {
	_, knownFamily := strategyLaneFamilies[latch.Family]
	return strings.TrimSpace(latch.AccountRef) != "" &&
		(latch.Market == "KR" || latch.Market == "US") && knownFamily &&
		strings.TrimSpace(latch.LaneID) != "" && strings.TrimSpace(latch.LaneVersion) != "" &&
		strings.TrimSpace(latch.LatchID) != "" && latch.LatchRevision > 0 &&
		strings.TrimSpace(latch.Reason) != "" && !latch.ObservedAt.IsZero()
}

// RecordStrategyLaneLatch 는 이 레인의 **열려 있는** 잠금을 돌려준다. 없으면
// 건네받은 것을 기록하고 그것을 돌려준다.
//
// 답이 하나인 것이 요점이다. "새로 넣었다/이미 있었다"를 bool 로 함께 돌려주면
// 호출자가 그것을 버릴 수 있고, 버리면 두 번째 실패가 첫 원인을 덮어썼는지
// 아닌지를 아무도 모른다. 여기서는 언제나 **지금 열려 있는 잠금**이 답이고,
// 그것이 첫 원인이다.
//
// 그래서 이 함수는 주기마다 불러도 안전하다. 잠긴 레인을 매 사이클 다시
// 기록하려 해도 첫 기록만 남는다 — 쓰기가 한 번 실패해도 다음 주기가 낫는다.
func (j *Journal) RecordStrategyLaneLatch(ctx context.Context, latch StrategyLaneLatch) (StrategyLaneLatch, error) {
	if j == nil || j.db == nil {
		return StrategyLaneLatch{}, fmt.Errorf("%w: journal unavailable", ErrStrategyLaneLatchInvalid)
	}
	if !latch.valid() {
		return StrategyLaneLatch{}, ErrStrategyLaneLatchInvalid
	}
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return StrategyLaneLatch{}, err
	}
	defer tx.Rollback()
	existing, found, err := openStrategyLaneLatchTx(ctx, tx, latch)
	if err != nil {
		return StrategyLaneLatch{}, err
	}
	if found {
		return existing, nil
	}
	abnormal := 0
	if latch.Abnormal {
		abnormal = 1
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO strategy_lane_latches
		(account_ref,market,family,lane_id,lane_version,latch_id,latch_revision,reason,abnormal,activation_generation,observed_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		latch.AccountRef, latch.Market, latch.Family, latch.LaneID, latch.LaneVersion, latch.LatchID,
		latch.LatchRevision, latch.Reason, abnormal, latch.ActivationGeneration,
		formatJournalTime(latch.ObservedAt.UTC()))
	if err != nil {
		return StrategyLaneLatch{}, fmt.Errorf("journal: recording strategy lane latch: %w", err)
	}
	seq, err := result.LastInsertId()
	if err != nil {
		return StrategyLaneLatch{}, fmt.Errorf("journal: reading strategy lane latch identity: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return StrategyLaneLatch{}, err
	}
	latch.Seq = seq
	latch.ObservedAt = latch.ObservedAt.UTC()
	return latch, nil
}

// OpenStrategyLaneLatches 는 이 계좌에서 **아직 복구되지 않은** 잠금 전부다.
//
// 레인은 이 목록에서 태어난다. 목록에 없으면 열린 채로 태어나고, 있으면 잠긴
// 채로 태어난다 — 재시작이 잠금을 지우지 않는다는 말의 뜻이 이것이다.
func (j *Journal) OpenStrategyLaneLatches(ctx context.Context, accountRef string) ([]StrategyLaneLatch, error) {
	if j == nil || j.db == nil {
		return nil, fmt.Errorf("%w: journal unavailable", ErrStrategyLaneLatchInvalid)
	}
	if strings.TrimSpace(accountRef) == "" {
		return nil, ErrStrategyLaneLatchInvalid
	}
	rows, err := j.db.QueryContext(ctx, `SELECT latched.latch_seq,latched.account_ref,latched.market,latched.family,
		latched.lane_id,latched.lane_version,latched.latch_id,latched.latch_revision,latched.reason,latched.abnormal,
		latched.activation_generation,latched.observed_at
		FROM strategy_lane_latches latched
		LEFT JOIN strategy_lane_latch_recoveries recovered ON recovered.latch_seq=latched.latch_seq
		WHERE latched.account_ref=? AND recovered.latch_seq IS NULL
		ORDER BY latched.latch_seq`, accountRef)
	if err != nil {
		return nil, fmt.Errorf("journal: reading open strategy lane latches: %w", err)
	}
	defer rows.Close()
	latches := []StrategyLaneLatch{}
	for rows.Next() {
		latch, err := scanStrategyLaneLatch(rows)
		if err != nil {
			return nil, err
		}
		latches = append(latches, latch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("journal: reading open strategy lane latches: %w", err)
	}
	return latches, nil
}

// RecoverStrategyLaneLatch 는 잠금 하나를 닫는 **유일한** 방법이다.
//
// 성공한 사이클로는 열 수 없다. 잠근 이유에 대해 아무것도 증명하지 못하기
// 때문이다. 여기서 받는 것은 서명된 활성화 매니페스트의 세대이고, 그것이
// 기록된 세대보다 **엄격히 커야** 한다는 판정은 Go 가 아니라 SQLite 트리거가
// 한다 — 판정을 코드에 두면 다른 호출자가 다른 판정을 쓸 수 있다.
func (j *Journal) RecoverStrategyLaneLatch(ctx context.Context, seq int64, activationGeneration uint64) error {
	if j == nil || j.db == nil {
		return fmt.Errorf("%w: journal unavailable", ErrStrategyLaneLatchInvalid)
	}
	if seq <= 0 || activationGeneration == 0 {
		return ErrStrategyLaneLatchInvalid
	}
	_, err := j.db.ExecContext(ctx, `INSERT INTO strategy_lane_latch_recoveries(latch_seq,activation_generation,observed_at)
		VALUES(?,?,?)`, seq, activationGeneration, formatJournalTime(j.clk.Now().UTC()))
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "strictly newer signed activation") {
		return fmt.Errorf("%w (latch %d)", ErrStrategyLaneLatchRecoveryEvidence, seq)
	}
	return fmt.Errorf("journal: recovering strategy lane latch: %w", err)
}

func openStrategyLaneLatchTx(ctx context.Context, tx *sql.Tx, latch StrategyLaneLatch) (StrategyLaneLatch, bool, error) {
	row := tx.QueryRowContext(ctx, `SELECT latched.latch_seq,latched.account_ref,latched.market,latched.family,
		latched.lane_id,latched.lane_version,latched.latch_id,latched.latch_revision,latched.reason,latched.abnormal,
		latched.activation_generation,latched.observed_at
		FROM strategy_lane_latches latched
		LEFT JOIN strategy_lane_latch_recoveries recovered ON recovered.latch_seq=latched.latch_seq
		WHERE latched.account_ref=? AND latched.market=? AND latched.family=? AND latched.lane_id=?
		  AND latched.lane_version=? AND recovered.latch_seq IS NULL
		ORDER BY latched.latch_seq LIMIT 1`,
		latch.AccountRef, latch.Market, latch.Family, latch.LaneID, latch.LaneVersion)
	open, err := scanStrategyLaneLatch(row)
	if errors.Is(err, sql.ErrNoRows) {
		return StrategyLaneLatch{}, false, nil
	}
	if err != nil {
		return StrategyLaneLatch{}, false, err
	}
	return open, true, nil
}

type strategyLaneLatchScanner interface {
	Scan(dest ...any) error
}

func scanStrategyLaneLatch(row strategyLaneLatchScanner) (StrategyLaneLatch, error) {
	var latch StrategyLaneLatch
	var abnormal int
	var observedAt string
	if err := row.Scan(&latch.Seq, &latch.AccountRef, &latch.Market, &latch.Family, &latch.LaneID,
		&latch.LaneVersion, &latch.LatchID, &latch.LatchRevision, &latch.Reason, &abnormal,
		&latch.ActivationGeneration, &observedAt); err != nil {
		return StrategyLaneLatch{}, err
	}
	parsed, err := time.Parse(time.RFC3339Nano, observedAt)
	if err != nil {
		return StrategyLaneLatch{}, fmt.Errorf("journal: strategy lane latch observed_at: %w", err)
	}
	latch.Abnormal = abnormal == 1
	latch.ObservedAt = parsed.UTC()
	return latch, nil
}

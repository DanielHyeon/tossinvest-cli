package journal

// a199_order_path_evidence.go 는 「이 계좌로 그 생산자가 낸 주문이 확정된 적이 있다」를 원장에서 읽는 자리임(a199 G2).
//
// # 왜 있나
//
// 낡은(만료 · 주문 endpoint 누락) capability attestation 으로 엔진을 청산 전용으로 띄우는 근거는 「이 계좌에서 자동 주문 경로가 한때
// 실제로 동작했고 그 증명이 낡았다」임. 그 사실은 attestation 파일에 남지 않음 — 감독 검증을 한 번도 안 한 GET 만의 첫 파일과
// 감독 증거가 기한을 넘겨 빠진 파일이 같은 모양임. 원장의 확정된 엔진 PLACE 는 그 경로가 이 계좌에서 실제로 동작했다는 가장 직접적인
// 기록이고 계좌에 묶여 있음(감사 로그의 수락 줄은 계좌를 `[account]` 로 치환해 묶이지 않음 — a131).
//
// # 지키는 것
//
//   - PLACE 만 셈 — a198 이전의 운영 원장에는 CANCEL attempt 가 없음(취소 발행이 늘 거절됐음).
//   - CONFIRMED 만 셈 — IN_DOUBT · 거절 · 미송신은 「동작했다」가 아님.
//   - source 는 호출자가 정확히 넘김(엔진 게이트웨이의 값) — flatten 같은 다른 생산자의 주문은 엔진 자동 경로의 기록이 아님.
//   - 계좌 대조는 다른 계좌 범위 질의와 같은 TRIM 정규화(가림 꼴 대조가 아님).
//   - 읽기 전용 · 존재 질의(LIMIT 1). 쓰기 0.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// confirmedPlaceQuery 는 HasConfirmedPlace 의 질의 하나임 — 질의 계획 시험이 사본이 아니라 이 상수를 잼(구현 리뷰 r2 처분 2).
// 인자 순서: 상태 · kind · source · 계좌.
const confirmedPlaceQuery = `SELECT 1 FROM mutation_attempts a JOIN intents i ON i.id = a.intent_id
		  WHERE a.state = ? AND a.kind = ? AND i.source = ? AND TRIM(i.account_ref) = ? LIMIT 1`

// HasConfirmedPlace 는 accountRef 계좌로 source 가 낸 PLACE attempt 중 CONFIRMED 인 것이 하나라도 있는지임.
func (j *Journal) HasConfirmedPlace(ctx context.Context, accountRef, source string) (bool, error) {
	var one int
	err := j.db.QueryRowContext(ctx, confirmedPlaceQuery,
		string(StateConfirmed), string(KindPlace), source, strings.TrimSpace(accountRef)).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("journal: looking for a confirmed %q place on the account: %w", source, err)
	}
	return true, nil
}

# R6 판정 — 시장 범위(`scope_market`) QM 행 위의 credit 해제(구현 적대 리뷰 r1)

판정: **도달 불가(생산)**. 기제는 base 부터 있고 a201 은 credit 근거를 하나 더할 뿐이지만, 그 행을 만드는 생산 경로가 없음.

## 1. 생산 호출자(원문 `callers.txt` — CodeGraph 1.6.0 `callers` + grep)

- 시장 범위 QM 행을 쓰는 생산 자리는 `latchReleasedOwnerLateFillInTx`(risk_bucket_owner.go:595) 하나. `INSERT INTO reconcile_states` 생산 자리 셋 가운데 나머지 둘(position_projection.go:414 · reconcile_states.go:215 `EnterReconcile`)은 생산 호출자가 `ScopeMarket` 을 주지 않음(lineage.go:450 · fills.go:1808 · mismatch.go:914 — 전부 무지정).
- 그 함수는 생산 경로(`RecordFill` → `runApplyHooks` → `applyRiskBucketOwnerBindingInTx`, 전략 backfill 도)에서 불리지만, 행을 쓰는 조건이 **해제된 위험 버킷 owner**(`risk_bucket_owners.released_at IS NOT NULL`)임.
- `released_at` 을 쓰는 자리는 `releaseRiskBucketOwner`(risk_bucket_owner.go:1017) 하나이고, 그 함수의 호출자 11 은 전부 `internal/journal/*_test.go`. 그 해제의 전제인 broker-zero 관측 기록(`recordRiskBucketBrokerZeroObservation`)도 호출자가 시험뿐(함수 주석: "deliberately no production constructor/caller").
- 그래서 운영 장부에는 그 행이 생길 수 없음. (마이그레이션 · 운영자 SQL 로 넣는 경우는 범위 밖.)

## 2. 탐침(원문 `head-run.txt` · `base-run.txt`, 소스 `*.go.txt` — 커밋하지 않은 시험을 돌리고 지움)

입력: `EnterReconcile(ScopeMarket "KR")` 로 시장 범위 QM 행 활성 → Refresh → 같은 종목 수량 불일치 주기 → 깨끗한 주기 3.

| 판 | 결과 |
|---|---|
| (가) 조정 credit, HEAD `b72182eb6` | Refresh 블록 Key 가 종목 범위(시장 무시)라 NULL 범위 행은 안 섬(added 0). 깨끗한 주기마다 `releasing … the journal did not confirm the exact-cause release` 오류 3/3, credit 은 남음(blocks 1), 행 활성 |
| (가) 조정 credit, base `35d28a535` | 같음 — a201 이전부터의 기제 |
| (나) 체결 credit, HEAD | 같음 — a201 은 같은 기제에 credit 근거를 하나 더함 |

엔진에서는 그 오류가 Observe 오류라 그 주기의 편입 판정(`judgeHoldings`)을 건너뜀 — 생긴다면 P3 가 아님. 다만 1 에 따라 생산에서는 생기지 않음.

## 3. 결론

- 「도달 불가」. a201 처분 대상 아님 — 수리하지 않음.
- 위험 버킷 owner 해제를 생산에 배선하는 change 가 생기면 이 기제(시장 범위 행 + `ScopeMarket` 없는 해제 요청)를 그 change 의 전제로 다시 볼 것.

# Function Logic Map: `ValidReconcileReleaseCause` (journal) — post-edit(v3)

- Source: `internal/journal/reconcile_states.go` (111-119), GREEN `ffd9dc6d6`.
- AST evidence: `ast.json`(sha 563d41a68363 — gstack r1 G4 로 같은 파일 상수 주석만 바뀜, 줄 수 · 이 함수 무변, 분기 3 — 편집 전 3, 추가 0, return@115 @117). 재번호 `../renumber.py` — B1~B3 equal(좌표도 같음).
- Risk scan: `risk-pattern-report.md`(일치 0).
- 커버: `../../coverage/post-internal-journal--validreconcilereleasecause.txt`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cause` | 문자열 | 호출자(`ReleaseReconcile` · `ReleaseReconciles` · a192 종결 해제) | 목록 밖 → false → 쓰기 거절 |
| 불변식 | `release_cause` 에 DDL CHECK 가 없으므로 이 목록이 쓰기 제약 그 자체 | 함수 주석 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch@112) | 원인 분기 | — | — | 기존 `TestUnknownReleaseCauseIsRefused` |
| B2 (case@113) | 이 판본이 쓰는 원인 — **a201: `ReconcileReleaseFillApplied`(`FILL_APPLIED`) 추가** | — | return@115 true | `TestA201FillAppliedIsAReleaseCauseThisBuildWrites` · `TestA201TheFillReleaseCauseIsFillApplied` |
| B3 (case@116) | 그 밖 | — | return@117 false | 기존 `TestUnknownReleaseCauseIsRefused` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | AST calls 0 |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: case 값 1. 다른 원인 판정 불변.
- High-risk impact: yes(원장 쓰기 제약) — 새 원인 하나를 허용 목록에 더함. broker-zero 경로의 허용 목록(risk_bucket_owner.go:280 · :677)은 무편집이라 새 원인을 거절(`TestA201BrokerZeroRefusesAFillAppliedRelease`).

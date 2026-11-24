# Function Logic Map: `ValidReconcileReleaseCause` (journal) — pre-edit(v3)

- Source: `internal/journal/reconcile_states.go` (111-119), HEAD `6e36fd2c0`.
- AST evidence: `ast.json`(sha 1591e0153868, 분기 3, return@115 @117). a192 아카이브 번들의 편집 뒤 ast 와 sha 동일.
- Risk scan: `risk-pattern-report.md`(일치 0).
- v3 편집: B2 case@113 의 목록에 새 상수 `ReconcileReleaseFillApplied`(`FILL_APPLIED`) 1 추가. 분기 수 무변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cause` | 문자열 | 호출자(`ReleaseReconcile` :269 · `ReleaseReconciles` :351 · a192 종결 해제 :106 [손]) | 목록 밖 → false → 쓰기 거절 |
| 불변식 | `release_cause` 에 DDL CHECK 가 없으므로 이 목록이 쓰기 제약 그 자체(execution_contract.go:210) | 함수 주석 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (switch@112) | 원인 분기 | — | — | 기존 `TestUnknownReleaseCauseIsRefused` |
| B2 (case@113) | 이 판본이 쓰는 원인 — **v3 편집: `FILL_APPLIED` 추가** | — | return@115 true | 기존 + v3 「FILL_APPLIED 는 유효」 |
| B3 (case@116) | 그 밖 | — | return@117 false | 기존 `TestUnknownReleaseCauseIsRefused` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | AST calls 0 |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: case 값 1. 다른 원인 판정 불변.
- High-risk impact: yes(원장 쓰기 제약) — 새 원인 하나를 허용 목록에 더함. broker-zero 경로의 허용 목록(risk_bucket_owner.go:280 · :677)은 무편집이라 새 원인을 거절(V6 의도).

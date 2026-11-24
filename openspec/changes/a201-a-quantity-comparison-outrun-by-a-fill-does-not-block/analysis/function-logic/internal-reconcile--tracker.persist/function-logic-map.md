# Function Logic Map: `Tracker.persist` (reconcile) — pre-edit(v3)

- Source: `internal/reconcile/mismatch.go` (878-954), HEAD `6e36fd2c0`.
- AST evidence: `ast.json`(sha 9895f31184e5, 분기 11, return@880 @909 @915 @927 @943 @947 @953). a192 아카이브 번들의 편집 뒤 ast 와 sha 동일.
- Risk scan: `risk-pattern-report.md`(일치 0).
- v3 편집: B7 if@922 거짓 쪽(크레딧 해제)에서 `ReleaseReconcile` call@935 의 `Cause` · `Evidence` 를 블록의 원인 표지로 고름(`FILL_APPLIED` 또는 기존 `ADJUSTMENT_APPLIED`). 분기 1 추가 예상(표지 if).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `out.Added` | pending 블록 전부 | `pendingBlocksForPersistence` | 첫 오류에서 반환 |
| `out.Cleared` | 크레딧 해제(조정 · v3 체결) + a192 종결 증명 후보 | Observe | 장부 오류 반환 — 블록 · 게이트 유지 |
| 불변식(v3) | 삽입 갈래(B2~B5) · a192 갈래(B7~B9) 무변. 크레딧 해제의 요청은 원인 · 근거 문구만 다름(범위 · `ExpectCause` 같음) | design V6 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@879) | `t.Journal == nil` | — | return@880 | 기존 |
| B2 (range@890) | pending 블록마다 | `enterReconcileWithBinding` call@902 | — | 기존 |
| B3 (if@891) | 빈 종목 수량 블록(표현 불가) | — | continue | 기존 |
| B4 (if@908) | 삽입 쓰기 오류 | — | return@909 | 기존 a192 |
| B5 (if@912) | 다른 원인이 그 범위를 쥠 | — | return@915 | 기존 |
| B6 (range@921) | Cleared 블록마다 | — | — | 기존 · v3 |
| B7 (if@922) | a192 `terminalRelease` 표시 | `releaseOnOrderTerminal` call@925 | B8/B9 뒤 continue | a192 |
| B8 (if@926) | 종결 증명 해제 오류 | — | return@927 | a192 |
| B9 (if@930) | 재판정이 해제 | `released` append | — | a192 |
| (B7 거짓 → call@935) | 크레딧 해제 — **v3 편집 지점: 원인 · 근거를 표지로 고름** | `ReleaseReconcile(Cause, Evidence, ExpectCause=QM)` | — | 기존 a083 · v3 `FILL_APPLIED` 행 단언 |
| B10 (if@942) | 해제 쓰기 오류 | — | return@943 | 기존 |
| B11 (if@946) | 활성 행 없음(`!ok`) | — | return@947 | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.enterReconcileWithBinding`(call@902) | 삽입 + a192 결속 | 오류 반환 | AST |
| `t.releaseOnOrderTerminal`(call@925) | a192 해제 | 오류 반환 | AST |
| `t.Journal.ReleaseReconcile`(call@935) | 크레딧 해제 — 원인은 `ValidReconcileReleaseCause` 가 검사(reconcile_states.go:269 [손]) | 오류 · 활성 행 없음이면 반환 | AST |

## State mutations and fallbacks

- 장부 행 삽입 · 해제(원인 · 근거 문구 기록)만. v3 는 해제 행의 `release_cause` · `evidence` 값만 바꿀 수 있음.

## Safety conclusion

- Safe edit boundary: call@935 요청의 `Cause` · `Evidence` 두 칸. 해제 여부는 Observe 가 정함 — persist 는 무엇을 풀지 바꾸지 않음.
- High-risk impact: yes(원장 기록). 새 원인이 유효 집합에 없으면 해제 쓰기가 오류(fail-closed) — `ValidReconcileReleaseCause` 편집과 짝.

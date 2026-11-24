# Function Logic Map: `Tracker.persist` (reconcile) — post-edit(v3)

- Source: `internal/reconcile/mismatch.go` (890-973), GREEN `ffd9dc6d6`.
- AST evidence: `ast.json`(sha 0facba7489a3, 분기 12 — 편집 전 11 + 1, return@892 @921 @927 @939 @962 @966 @972). 명령 `go run ./tools/logic-map --file internal/reconcile/mismatch.go --func Tracker.persist`.
- 재번호: 편집 전 번들(sha 9895f31184e5, 878-954)과 `../renumber.py` 정렬 — B1~B9 equal(같은 번호), **새 분기 B10 if@950**(`block.releasedByFill`), 옛 B10 if@942 → 새 B11 if@961, 옛 B11 if@946 → 새 B12 if@965.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 커버: `../../coverage/post-internal-reconcile--tracker.persist.txt`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `out.Added` | pending 블록 전부 | `pendingBlocksForPersistence` | 첫 오류에서 반환 |
| `out.Cleared` | credit 해제(조정 · 체결) + a192 종결 증명 후보 | Observe | 장부 오류 반환 — 블록 · 게이트 유지 |
| `block.releasedByFill` | Observe 의 원인 표지(V5b 등식) | Observe @712 | 거짓이면 오늘과 같은 `ADJUSTMENT_APPLIED` |
| 불변식(v3) | 삽입 갈래(B2~B5) · a192 갈래(B7~B9) 무변. credit 해제 요청은 원인 · 근거 문구만 다름(범위 · `ExpectCause` 같음) | design V6 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@891) | `t.Journal == nil` | — | return@892 | 기존 |
| B2 (range@902) | pending 블록마다 | `enterReconcileWithBinding` call@914 | — | 기존 |
| B3 (if@903) | 빈 종목 수량 블록(표현 불가) | — | continue | 기존 |
| B4 (if@920) | 삽입 쓰기 오류 | — | return@921 | `TestA201AnObserveErrorCycleGrantsNoFillCredit` · 기존 a192 |
| B5 (if@924) | 다른 원인이 그 범위를 쥠 | — | return@927 | 기존 |
| B6 (range@933) | Cleared 블록마다 | — | — | a201 해제 시험 |
| B7 (if@934) | a192 `terminalRelease` 표시 | `releaseOnOrderTerminal` call@937 | B8/B9 뒤 continue | `TestA201AnA192ReleasableRowIsNotInterceptedByAFillCredit` · a192 |
| B8 (if@938) | 종결 증명 해제 오류 | — | return@939 | a192 |
| B9 (if@942) | 재판정이 해제 | `released` append | — | a192 |
| **B10 (if@950)** | **a201: `block.releasedByFill` — 체결 credit 해제** | 원인 · 근거를 `FILL_APPLIED` · 체결 문구로 바꿈(거짓이면 `ADJUSTMENT_APPLIED` 그대로) | — | 참: `TestA201AFillCreditIsSpentOnlyByAStrictlyLaterCleanComparison` · W2/W3 해제 시험. 거짓: `TestA201ALaterAdjustmentNamesTheRelease` · `TestA201ARealDifferenceReleasesOnlyThroughTheAdjustment` |
| (call@954) | credit 해제 | `ReleaseReconcile(Cause, Evidence, ExpectCause=QM)` | — | 같은 시험 |
| B11 (if@961) | 해제 쓰기 오류(유효 집합에 없는 원인 포함) | — | return@962 | 기존 `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` 계열 · 변이 「ValidReconcileReleaseCause 에서 새 원인 삭제」(2.5) |
| B12 (if@965) | 활성 행 없음(`!ok`) | — | return@966 | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.enterReconcileWithBinding`(call@914) | 삽입 + a192 결속 | 오류 반환 | AST |
| `t.releaseOnOrderTerminal`(call@937) | a192 해제 | 오류 반환 | AST |
| `t.Journal.ReleaseReconcile`(call@954) | credit 해제 — 원인은 `ValidReconcileReleaseCause` 가 쓰기 시점에 검사 | 오류 · 활성 행 없음이면 반환 | AST |

## State mutations and fallbacks

- 장부 행 삽입 · 해제(원인 · 근거 문구 기록)만. a201 은 credit 해제 행의 `release_cause` · `evidence` 값만 표지로 가름.

## Safety conclusion

- Safe edit boundary: B10 의 두 값 선택 + call@954 의 `Cause` · `Evidence` 두 칸. 해제 여부는 Observe 가 정함 — persist 는 무엇을 풀지 바꾸지 않음.
- High-risk impact: yes(원장 기록). 새 원인이 유효 집합에 없으면 해제 쓰기가 오류(B11 — fail-closed) — `ValidReconcileReleaseCause` 편집과 짝.

# Function Logic Map: `ReconcileDriver.RunOnce` (engine) — post-edit(v3 + V4c)

- Source: `internal/app/engine/reconcileloop.go` (430-492), `5ac67068f`(gstack r1 G1 — V4c).
- AST evidence: `ast.json`(sha a47a346fb26f, 분기 8 — 편집 전 8, 추가 0, return@435 @444 @471 @483 @491, defer@431). 명령 `go run ./tools/logic-map --file internal/app/engine/reconcileloop.go --func ReconcileDriver.RunOnce`. 재번호는 `../renumber.py` — 2.4 판(`ffd9dc6d6`) 대비 `../renumber-output.txt`, V4c 판(`3bfbed9` 대비) `../renumber-output-gstack-r1.txt`: B1~B8 모두 같은 번호, 좌표만 이동(파일 머리 단계 목록 +5 줄 · V4c 2 줄).
- Risk scan: `risk-pattern-report.md`(일치 0).
- 커버: `../../coverage/post-gstack-r1-internal-app-engine--reconciledriver.runonce.txt`(mode set, engine 선택 시험 — `../../verify/filter-engine.txt`). 2.4 판 측정은 `../../coverage/post-internal-app-engine--reconciledriver.runonce.txt`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot` | 안정화된 두 수집 중 둘째 | `d.stabilise`(call@433) | 불안정 · 수집 오류면 B1 → return@435 |
| `fillMark` | 로컬 읽기 **직전** 전역 최대 체결 id(또는 「모름」) | `reconcile.MarkBeforeLocalRead`(call@440, 새 파일) — 로컬 읽기(call@441) 바로 앞 문장(구조 시험) | 읽기 실패는 「모름」 — 주기를 실패시키지 않고 그 주기 체결 credit 만 없음 |
| `local` | 투영의 비-CLOSED 인스턴스 합 + 추적 주문 | `reconcile.LocalStateFromJournal`(call@441) | 오류면 B2 → `cycle.Err` · return@444 |
| `diff` | `Compare(snapshot, local)`(call@446) | 순수 | — |
| `converged` | `ConvergeReport` | `ConvergeQuantities`(call@458) | 오류는 `cycle.Err` 가 비었을 때만(B4) 담고 계속 |
| 불변식 1 | `Observe`(call@473)는 수렴 · Refresh 뒤, 같은 `diff`. 편집 전과 입력 · 순서 · 오류 처분 같음 | AST | — |
| 불변식 2 | Observe 오류면 부여 · 편입 판정 안 감 | B7 → return@483 | fail-closed |
| 불변식 3(a201) | 체결 credit 부여(call@487)는 B7 통과 뒤 · `judgeHoldings`(call@490) 앞의 최상위 문장 1, 분기 0 | 구조 시험 `TestA201TheFillCreditIsGrantedAtOneSiteAfterASuccessfulObserve` | — |
| 불변식 4(a201 V4c) | Refresh 오류로 Observe 앞에서 끝나도 그 비교가 본 분쟁 종목의 엄격히 나중 credit 은 반박(call@470 — B5 몸 안 return@471 직전) | 구조 시험 `TestA201ARefreshFailureRefutesOnlyOnItsOwnReturnPath` | 오류 없는 경로 무변 |
| 불변식 5 | 실패한 주기도 `d.note` 가 셈 | defer@431 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@434) | `!ok` — 불안정 또는 수집 오류 | 없음 | return@435 | 기존 `TestAnUnstableAccountDefersSilently` |
| B2 (if@442) | 로컬 읽기 오류(표식 읽기 실패는 그 앞에서 「모름」으로 흡수) | `cycle.Err` | return@444 | `TestA201AnUnreadableJournalFailsTheCycleAtTheLocalReadAsBefore` |
| B3 (if@453) | 접기 오류 ∧ `cycle.Err == nil` | `cycle.Err` | 계속 | 선택 시험 0(편집 무관, body=0) |
| B4 (if@459) | 수렴 오류 ∧ `cycle.Err == nil` | `cycle.Err` | 계속 — 수렴 오류에도 같은 diff 로 Observe, 그 뒤 부여 | `TestA201W3aFillAfterConvergencePositionsIsCreditedLikeAnAdjustment` 외 |
| B5 (if@465) | Refresh 오류 | **a201 V4c: `d.opts.Tracker.RefuteCreditsDisputedBy(diff)`(call@470)** | return@471(Observe · 부여 안 함) | `TestA201ARefreshFailureCycleStillRefutesADisputedFillCredit` · `TestA201ARefreshFailureCycleStillRefutesADisputedAdjustmentCredit` · `TestA201ARefreshFailureCycleKeepsACreditItsComparisonDoesNotDispute` · `TestA201ARefreshFailureDoesNotRefuteTheCreditItsOwnConvergenceGave` |
| B6 (if@466) | B5 안, `cycle.Err == nil` | `cycle.Err` | — | 같은 시험(`refreshFails` 가 `c.Err` 단언) |
| B7 (if@479) | Observe 오류 | — | return@483(부여 · 편입 안 감) | `TestA201AnObserveErrorCycleGrantsNoFillCredit` · 기존 `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` |
| B8 (if@480) | B7 안, `cycle.Err == nil` | `cycle.Err` | — | 같은 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.stabilise`(call@433) | 수집 2회 · 안정화 | 오류 · 불안정 → B1 | AST |
| `reconcile.MarkBeforeLocalRead`(call@440) | 로컬 읽기 직전 전역 체결 표식(V2 (b)) | 실패 → 「모름」(오류 없음) | AST · `internal/reconcile/a201_fill_credit.go` |
| `reconcile.LocalStateFromJournal`(call@441) | 로컬 믿음 | 오류 → B2 | AST |
| `reconcile.Comparer{}.Compare`(call@446) | 비교 | 순수 | AST |
| `d.ingest.IngestExternalPositions`(call@452) | 외부 포지션 접기 | 오류 기록만 | AST |
| `d.opts.Converge.ConvergeQuantities`(call@458) | 수량 수렴 | 오류 기록만 | AST |
| `d.opts.Tracker.Refresh`(call@465) | 차단 투영 재적재 | 오류 → B5 | AST |
| `d.opts.Tracker.RefuteCreditsDisputedBy`(call@470) | V4c — 그 비교의 분쟁 종목 credit 반박(잠금 아래 기존 `refuteUsableAdjustmentCredits` · `symbolsInDispute` 만) | 오류 없음(메모리) | AST · 새 파일 |
| `d.opts.Tracker.Observe`(call@473) | 차단 · 해제 판정 · 장부 쓰기 | 오류 → B7 | AST(편집 전과 같은 호출 · 인자) |
| `d.opts.Tracker.FillsOutranAdjustment`(call@487) · `reconcile.FillOutranSymbols`(call@488) | 체결 credit 부여(V2 · V3 · V4) | 오류 없음 · 판독 실패 → 빈 목록 | AST · 새 파일 |
| `d.judgeHoldings`(call@490) | 편입 판정 | — | AST |

## State mutations and fallbacks

- 이 함수 자신의 변이는 `cycle` 필드뿐. 추적기 메모리에 대한 것은 Observe 뒤 credit 부여 1회(성공 경로)와 Refresh 오류 경로의 credit 반박 1회 — 장부 쓰기 없음.

## Safety conclusion

- Safe edit boundary: call@440 · call@470 · call@487 세 문장. 분기 · 반환 · Observe 호출 무변(재번호 equal 8).
- High-risk impact: yes — 차단의 해제 근거를 하나 더하고(체결 credit), Refresh 실패 주기의 반박을 더함(V4c — 오늘의 조정 credit 경로도 반박이 늘어나는 보수 방향; base 측정 `../../review/g1-base/base-run.txt` 에서는 그 입력이 ADJUSTMENT_APPLIED 로 풀렸음).

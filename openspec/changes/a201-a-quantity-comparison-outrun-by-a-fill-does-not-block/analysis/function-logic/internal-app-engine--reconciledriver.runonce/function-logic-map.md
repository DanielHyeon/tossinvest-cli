# Function Logic Map: `ReconcileDriver.RunOnce` (engine) — post-edit(v3)

- Source: `internal/app/engine/reconcileloop.go` (425-485), GREEN `ffd9dc6d6`.
- AST evidence: `ast.json`(sha 12b7acc07498, 분기 8 — 편집 전 8, 추가 0). 명령 `go run ./tools/logic-map --file internal/app/engine/reconcileloop.go --func ReconcileDriver.RunOnce`. 편집 전 번들(sha 52a738a58082, 425-479)과의 분기 대응은 `../renumber.py` 출력(B1→B1 … B8→B8, 좌표만 +2: 435→437 · 446→448 · 452→454 · 458→460 · 459→461 · 470→472 · 471→473).
- Risk scan: `risk-pattern-report.md`(일치 0).
- 커버: `../../coverage/post-internal-app-engine--reconciledriver.runonce.txt`(mode set, 선택 시험 engine · reconcile · journal 세 프로파일 합집합).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot` | 안정화된 두 수집 중 둘째 | `d.stabilise`(call@428) | 불안정 · 수집 오류면 B1 → return@430 |
| `fillMark` | 로컬 읽기 **전** 전역 최대 체결 id(또는 「모름」) | `reconcile.MarkBeforeLocalRead`(call@435, 새 파일) | 읽기 실패는 「모름」 — 주기를 실패시키지 않고 그 주기 체결 credit 만 없음(분기는 새 함수 안) |
| `local` | 투영의 비-CLOSED 인스턴스 합 + 추적 주문 | `reconcile.LocalStateFromJournal`(call@436) | 오류면 B2 → `cycle.Err` · return@439 |
| `diff` | `Compare(snapshot, local)`(call@441) | 순수 | — |
| `converged` | `ConvergeReport` | `ConvergeQuantities`(call@453) | 오류는 `cycle.Err` 가 비었을 때만(B4) 담고 계속 |
| 불변식 1 | `Observe`(call@466)는 수렴 · Refresh 뒤, 같은 `diff`. 편집 전과 입력 · 순서 · 오류 처분 같음 | AST | — |
| 불변식 2 | Observe 오류면 부여 · 편입 판정 안 감 | B7 → return@476 | fail-closed |
| 불변식 3(a201) | 체결 credit 부여(call@480)는 B7 통과 뒤 · `judgeHoldings`(call@483) 앞의 최상위 문장 1, 분기 0 | AST · 구조 시험 `TestA201TheFillCreditIsGrantedAtOneSiteAfterASuccessfulObserve` | — |
| 불변식 4 | 실패한 주기도 `d.note` 가 셈 | defer@426 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@429) | `!ok` — 불안정 또는 수집 오류 | 없음 | return@430 | 기존 `TestAnUnstableAccountDefersSilently` |
| B2 (if@437) | 로컬 읽기 오류(표식 읽기는 그 앞에서 「모름」으로 흡수) | `cycle.Err` | return@439 | `TestA201AnUnreadableJournalFailsTheCycleAtTheLocalReadAsBefore` |
| B3 (if@448) | 접기 오류 ∧ `cycle.Err == nil` | `cycle.Err` | 계속 | 선택 시험 0(편집 전후 body=0, 편집 무관) |
| B4 (if@454) | 수렴 오류 ∧ `cycle.Err == nil` | `cycle.Err` | 계속 — 수렴 오류에도 같은 diff 로 Observe, 그 뒤 부여 | `TestA201W3aFillAfterConvergencePositionsIsCreditedLikeAnAdjustment` 외 |
| B5 (if@460) | Refresh 오류 | — | return@464(Observe · 부여 안 함) | 선택 시험 0(편집 전후 body=0, 편집 무관) |
| B6 (if@461) | B5 안, `cycle.Err == nil` | `cycle.Err` | — | 선택 시험 0 |
| B7 (if@472) | Observe 오류 | — | return@476(부여 · 편입 안 감) | `TestA201AnObserveErrorCycleGrantsNoFillCredit` · 기존 `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` |
| B8 (if@473) | B7 안, `cycle.Err == nil` | `cycle.Err` | — | 같은 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.stabilise`(call@428) | 수집 2회 · 안정화 | 오류 · 불안정 → B1 | AST |
| `reconcile.MarkBeforeLocalRead`(call@435) | 로컬 읽기 전 전역 체결 표식(V2 (b)) | 실패 → 「모름」(오류 없음) | AST · `internal/reconcile/a201_fill_credit.go` |
| `reconcile.LocalStateFromJournal`(call@436) | 로컬 믿음 | 오류 → B2 | AST |
| `reconcile.Comparer{}.Compare`(call@441) | 비교 | 순수 | AST |
| `d.ingest.IngestExternalPositions`(call@447) | 외부 포지션 접기 | 오류 기록만 | AST |
| `d.opts.Converge.ConvergeQuantities`(call@453) | 수량 수렴 | 오류 기록만 | AST |
| `d.opts.Tracker.Refresh`(call@460) | 차단 투영 재적재 | 오류 → B5 | AST |
| `d.opts.Tracker.Observe`(call@466) | 차단 · 해제 판정 · 장부 쓰기 | 오류 → B7 | AST(편집 전과 같은 호출 · 인자) |
| `reconcile.FillOutranSymbols`(call@481) | 부여 대상(V2 (a)(b)(c)) | 판독 실패 → 빈 목록 | AST · 새 파일 |
| `d.opts.Tracker.FillsOutranAdjustment`(call@480) | 체결 credit 부여(V2 (d) · V3 · V4) | 오류 없음(메모리) | AST · 새 파일 |
| `d.judgeHoldings`(call@483) | 편입 판정 | — | AST |

## State mutations and fallbacks

- 이 함수 자신의 변이는 `cycle` 필드뿐. 새로 더한 것은 추적기 메모리의 credit 부여 1회(Observe 뒤) — 장부 쓰기 없음.

## Safety conclusion

- Safe edit boundary: call@435 1문장 + call@480 1문장. 분기 · 반환 · Observe 호출 무변.
- High-risk impact: yes — 차단의 **해제** 근거를 하나 더함. 그 주기의 차단 판정은 무변(부여는 Observe 뒤).

# Function Logic Map: `ReconcileDriver.RunOnce` (engine) — pre-edit

- Source: `internal/app/engine/reconcileloop.go` (425-479), HEAD `86a131a51`(base `f9b818a0e` 와 이 파일 동일).
- AST evidence: `ast.json`(sha 4e21e03de561, 분기 8, return@430 @437 @462 @474 @478, defer@426). 명령 `go run ./tools/logic-map --file internal/app/engine/reconcileloop.go --func ReconcileDriver.RunOnce`. `../../cited-ast/RunOnce.ast.json` 과 바이트 동일, 사전 증거(7fec501f1)와도 동일.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 이 번들은 **편집 전** 지도다. 편집 뒤 재추출·재번호는 2.4.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot` | 안정화된 두 수집 중 둘째 | `d.stabilise`(call@428) | 불안정 · 수집 오류면 B1 → return@430(비교 · 관측 없음) |
| `local` | 투영의 비-CLOSED 인스턴스 합 + 추적 주문 | `reconcile.LocalStateFromJournal`(call@434) — 한 tx 아님 | 오류면 B2 → `cycle.Err` · return@437 |
| `diff` | `Compare(snapshot, local)`(call@439) | 순수 함수 | — |
| `converged` | `ConvergeReport{Converged, Refused, Credited, Closed}` | `ConvergeQuantities`(call@451) | 오류는 `cycle.Err` 가 비었을 때만 담고(B4) 계속 |
| 불변식 1 | `Observe` 는 수렴 · Refresh **뒤**, 같은 `diff` 를 받음(call@464) | AST 순서 | — |
| 불변식 2 | Observe 오류면 편입 판정(call@477) 안 감 | B7 → return@474 | fail-closed |
| 불변식 3 | 실패한 주기도 `d.note` 가 셈 | defer@426 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@429) | `!ok` — 불안정 또는 수집 오류 | 없음(`cycle.Err` 는 stabilise 가 씀) | return@430 | 기존 `TestAnUnstableAccountDefersSilently` |
| B2 (if@435) | 로컬 읽기 오류 | `cycle.Err` | return@437 | 기존 선택 시험 0(편집 전 측정 body=0) — 편집이 이 앞에 워터마크 읽기를 넣으므로 2.2 에서 오류 갈래 시험 필요 |
| B3 (if@446) | 접기 오류 ∧ `cycle.Err == nil` | `cycle.Err` | 계속 | 선택 시험 0(body=0) |
| B4 (if@452) | 수렴 오류 ∧ `cycle.Err == nil` | `cycle.Err` | 계속 — **수렴 오류에도 같은 diff 로 Observe** | `TestA201CurrentBehaviourW3a…` 등(body=1) |
| B5 (if@458) | Refresh 오류 | — | return@462(Observe 안 함) | 선택 시험 0(body=0) |
| B6 (if@459) | B5 안, `cycle.Err == nil` | `cycle.Err` | — | 선택 시험 0 |
| B7 (if@470) | Observe 오류 | — | return@474(편입 안 감) | 기존 `TestTrackerReleaseFailureStopsBeforePriceAndAdoption`(body=1) |
| B8 (if@471) | B7 안, `cycle.Err == nil` | `cycle.Err` | — | 같은 시험(body=1) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `d.stabilise`(call@428) | 수집 2회 · 안정화 2초 | 오류 · 불안정 → B1 | AST |
| `reconcile.LocalStateFromJournal`(call@434) | 로컬 믿음 | 오류 → B2. **워터마크를 읽지 않음** — 편집 예정 자리(D2: 이 호출 **앞**) | AST, compare.go:193 |
| `reconcile.Comparer{}.Compare`(call@439) | 비교 | 순수 | AST |
| `d.ingest.IngestExternalPositions`(call@445) | `ExternalPos` 접기 | 오류 기록만 | AST |
| `d.opts.Converge.ConvergeQuantities`(call@451) | `Quantities` 수렴 | 오류 기록만; stale 은 판 전체 중단(converge.go:228-229), 보고서는 커밋분을 담음(defer@172) | AST |
| `d.opts.Tracker.Refresh`(call@458) | 장부의 차단 투영 재적재 | 오류 → B5 return | AST |
| `d.opts.Tracker.Observe`(call@464) | 차단 · 해제 · 승격 · 크레딧 판정, 장부 쓰기 | 오류 → B7 | AST; **편집 예정 자리(D1: 이 호출 직전 재판정 · 인자를 사본으로)** |
| `d.logOrderTerminalReleases`(call@466) | a192 로그 | 오류 없음 | AST |
| `d.judgeHoldings`(call@477) | 편입 판정 | — | AST |

## State mutations and fallbacks

- 이 함수 자신의 변이는 `cycle` 필드뿐. 장부 쓰기는 Ingest · 수렴 · Observe 안.
- 수렴의 `Refused` 는 버려지고(`cycle` 에 안 담김) 재수집은 없음 — `Observe` 는 수렴 전 비교를 그대로 받음(call@464 인자 `diff`).

## 편집 계획(v3 design V2 · V3 — 확정은 Manager freeze 뒤; HEAD `6e36fd2c0` 에서 이 함수 sha 무변)

(v1 · v2 의 편집 계획은 기각 — design 「기각한 대안」.)
1. call@434 앞: 로컬 읽기 전 전역 최대 체결 id 를 한 번 읽어 표식으로 쥠(새 파일의 함수). 읽기 실패는 「모름」 표식 — 주기를 실패시키지 않고 그 주기 체결 credit 만 없음(B2 와 다른 갈래, 분기는 새 함수 안).
2. B7 if@470(Observe 오류) 통과 뒤 · call@477(`judgeHoldings`) 앞: 부여 대상 종목 계산(새 함수 — 표식 · `diff` · `converged`) → `d.opts.Tracker` 의 새 부여 메서드 1회(`diff.AsOf`). 그 주기 `Observe` 호출(call@464) · 인자 · 오류 처분 무변.
3. 분기 추가 예상: 0(두 줄 모두 새 함수 호출). 기존 B1~B8 의 조건 · 순서 · 반환 무변.

## Safety conclusion

- Safe edit boundary: call@434 앞 1문장 + call@477 앞 1자리. Observe 호출 · 오류 처분(B7) · 편입 판정 순서 무변.
- High-risk impact: **yes** — 차단의 **해제** 근거를 하나 더함(안전 부트스트랩 5). 그 주기의 차단 판정은 무변(Observe 는 credit 부여 전에 끝남).

# Function Logic Map: `StrategyRuntimeAbsent`

- Source: `internal/httpapi/strategy_runtime.go`
- AST evidence: `ast.json` — **구현 후 재생성**(revision current, :27–29, 분기 0 · 반환 1 · 호출 1).
  편집 전 base `8688f74f` 는 :56–66, 분기 2 · 반환 3.
- Risk scan: `risk-pattern-report.md`

## 편집 전후 대조 (옛/새 ast.json difflib 정렬)

편집 전 B1(`if reader == nil {`, :57)·B2(`if presence, ok := reader.(StrategyRuntimePresence); ok {`, :60)는 **delete**
— 두 갈래와 종단(신호 없는 reader → false)은 새 잎 함수 `strategyprojection.StrategyRuntimeAbsent`
(`internal/strategyprojection/presence.go`, 새 파일 — 새 leaf 라 FLM not-applicable)로 텍스트 그대로 옮겨 갔다.
이 함수는 위임 한 줄 `return strategyprojection.StrategyRuntimeAbsent(reader)`(:28)이 됐다. `StrategyRuntimePresence`
인터페이스도 같은 곳으로 옮기고 여기는 type alias(분기 아님)다. 판정 한 벌은 `TestTheAbsenceIsJudgedInOnePlace`
(모듈 전수 구조 세기)와 `TestTheAbsenceJudgementIsOneForBothPackages`(세 자리 상태 + 신호 없는 reader + nil 동치,
reflect 타입 동일성)가 고정한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader` | nil · presence 를 말하는 reader(wrapper) · 신호 없는 reader | 호출자(router REST·SSE helper·집계 스냅샷·publisher) | 없음 — bool 판정(위임) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (분기 없음) | 위임 한 줄 (:28) | 판정 함수의 presence 질문 부작용(wake) 그대로 | `strategyprojection.StrategyRuntimeAbsent(reader)` | `TestAbsenceIsAskedAsAStateNotANil` · `TestTheAbsenceJudgementIsOneForBothPackages` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strategyprojection.StrategyRuntimeAbsent` | 판정 한 벌 | 부작용: presence 를 말하는 reader 의 재부착 시도 깨우기(비차단) | AST call :28 |

## State mutations and fallbacks

- 자체 상태 없음. 판정은 옮겨 간 곳에서 편집 전과 같은 세 갈래다(nil → true · presence → !Configured · 그 밖 → false).

## Safety conclusion

- Safe edit boundary: 본문 → 위임. 시그니처·이름 유지(기존 네 소비처·시험 무편집).
- High-risk impact: no — 조회 전용 projection 의 부재 판정.

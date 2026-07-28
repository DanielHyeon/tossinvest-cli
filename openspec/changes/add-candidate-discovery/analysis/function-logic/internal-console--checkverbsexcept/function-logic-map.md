# Function Logic Map: `checkVerbsExcept`

- Source: `internal/console/static_test.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

`checkVerbs`에 한 seam의 근거가 적힌 철자만 통과시킨다.

**예외의 키는 동사가 아니라 식별자 전체다.** `Orders`를 면제해도 `PlaceOrder`·`CancelOrders`·`OrdersPlacer`는 같은 seam에서 계속 실패한다 — 그 중 어느 것도 그 문자열이 아니기 때문이다. 이것이 해치를 **옆에 적힌 근거의 크기**로 묶어 둔다.

존재 이유는 리뷰 P0-3이다 — `OrdersReader{Orders}`가 금지 동사 그 자체이고, 자연스러운 해제는 목록에서 `order`를 지우는 것인데 그러면 미래의 `/order/place`와 미래의 `PlaceOrder`가 **같은 순간에** 통과한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `name` | 식별자 | 호출자 | 예외에 있으면 즉시 통과 |
| `exempt` | 철자 전체 → 근거 문장 | `consoleCapabilities[...].VerbExemptions` | nil 허용(= `checkVerbs`) |
| `mutationVerbs` | 15개 철자 | static_test.go:813 | 목록을 줄이면 `TestTheOrdersSeamIsTheOnlyCapabilityWithVerbExemptionsAndTheyAreEnumerated`가 직접 실패한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `_, ok := exempt[name]; ok` | 없음 | return — 통과 | `Orders`·`OrdersReader`·`OrdersReading`·`OrderRecord`·`ConditionalRecord` 다섯 |
| B2 | `for _, verb := range mutationVerbs` | 없음 | 없음 | 그 밖의 모든 이름 |
| B3 | `strings.Contains(lowered, verb)` | `t.Errorf` | 없음 | `PlaceOrder` 추가 변이 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.ToLower` / `strings.Contains` | 부분 문자열 검사 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음(판정 전용).

## Safety conclusion

- Safe edit boundary: 신설. 금지 목록은 손대지 않고 필드별·철자별 해치를 열었다.
- High-risk impact: yes (주문 능력 주입 차단 — 해치가 넓어지면 금지 목록이 조용히 짧아진 것과 같다)

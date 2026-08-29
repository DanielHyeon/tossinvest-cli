# Function Logic Map: `checkNoEmbedding`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

인터페이스가 다른 인터페이스를 embed하면 실패한다. embed된 쪽이 얻는 것은 이쪽도 **조용히** 얻는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `iface` | 인터페이스 선언 | `capabilityClosure`가 모은 것 또는 패키지 순회 | 해당 없음 |
| `subject` | 실패 메시지의 주어 | 호출자 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, method := range iface.Methods.List` | 없음 | 없음 | 인터페이스 seam 6 + 패키지의 나머지 인터페이스 |
| B2 | `len(method.Names) == 0` | `t.Errorf` | 없음 | embed 삽입 변이 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | AST 필드 검사뿐 | 순수 | ast.json calls=null 아님 — `t.Helper`와 `t.Errorf`만 |

## State mutations and fallbacks

- 없음(판정 전용).

## Safety conclusion

- Safe edit boundary: 신설(추출). 이전에는 `HoldingsReader` 한 인터페이스에 대해 인라인으로 같은 검사를 했다.
- High-risk impact: yes (주문 능력 주입 차단 — embed 한 줄이 메서드 집합 열거를 무의미하게 만든다)

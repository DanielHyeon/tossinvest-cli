# Function Logic Map: `carriesCapability`

- Source: `internal/console/static_test.go`
- Change: `add-candidate-discovery`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

메서드 집합이나 호출 가능한 것을 담을 수 있는 타입을 보고한다. **보수적**이다 — `default: return true`이고, 다른 패키지의 타입(`*ast.SelectorExpr`)은 여기서 읽을 수 없으므로 **능력이 있다고 가정한다**.

**경계**: 패키지가 선언하지 않은 무한정 식별자(`any`, 제네릭 파라미터, builtin)는 false다. 그 경우의 안전망은 `methodless`의 긍정형이지 이 함수가 아니다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `expr` | 구조체 필드의 타입 | `capabilityClosure` B25 | 인식 밖은 참(보수적) |
| `declaredTypes` | 패키지 선언 타입 색인 | `packageTypes` | 미선언 식별자는 false |
| `seen` | 순환 방지 | 호출자가 새 맵을 준다 | 순환이면 false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch v := expr.(type)` | 없음 | 아래 | 구조체를 가진 필드 전부 |
| B2 | `case *ast.FuncType, *ast.InterfaceType, *ast.IndexExpr, *ast.IndexListExpr` | 없음 | `true` | `PlaceOrder func(...)` 필드 변이 |
| B3 | `case *ast.SelectorExpr` | 없음 | `true`(다른 패키지 = 읽을 수 없음 = 능력 가정) | `domain.Position` 필드 |
| B4 | `case *ast.Ident` | 없음 | 아래 둘 | 선언된 이름 |
| B5 | `seen[v.Name]` | 없음 | `false` | 순환 타입 변이 |
| B6 | `decl, ok := declaredTypes[v.Name]` | 없음 | 재귀 | 패키지 선언 타입 필드 |
| B7 | `case *ast.StarExpr` | 없음 | 재귀 | 포인터 필드 |
| B8 | `case *ast.ParenExpr` | 없음 | 재귀 | 현재 표면에 없음(방어) |
| B9 | `case *ast.ArrayType` | 없음 | 재귀(Elt) | `[]string` 필드 |
| B10 | `case *ast.Ellipsis` | 없음 | 재귀(Elt) | 현재 표면에 없음(방어) |
| B11 | `case *ast.MapType` | 없음 | Key 또는 Value | 현재 표면에 없음(방어) |
| B12 | `case *ast.ChanType` | 없음 | 재귀(Value) | 현재 표면에 없음(방어) |
| B13 | `case *ast.StructType` | 없음 | 필드 중 하나라도 참이면 참 | 중첩 구조체 필드 |
| B14 | `for _, f := range v.Fields.List` | 없음 | 없음 | 같은 위 |
| B15 | `carriesCapability(f.Type, …)` | 없음 | `true` | 같은 위 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (재귀) | 타입을 따라 내려간다 | `seen`이 순환을 끊는다 | ast.json calls |

## State mutations and fallbacks

- 없음(순수 함수).

## Safety conclusion

- Safe edit boundary: 신설. 구조체 필드 **이름**의 동사 검사를 켤지 끌지 결정하는 유일한 지점이다.
- High-risk impact: yes (주문 능력 주입 차단 — 여기서 false를 답하면 그 필드 이름은 동사 검사를 받지 않는다)

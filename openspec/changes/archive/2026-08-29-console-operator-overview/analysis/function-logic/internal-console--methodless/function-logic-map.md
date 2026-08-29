# Function Logic Map: `methodless`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

가드가 **적극적으로 빈 메서드 집합을 읽어낼 수 있는** 모양을 보고한다: func 타입, 평문 데이터, 근거가 적힌 다른 패키지 타입.

**긍정형인 이유**: '가드가 알아보지 못했다'가 '괜찮다'로 읽혀서는 안 된다. `consoleCapabilities`의 빈 항목은 그 필드가 메서드 집합을 갖지 않는다는 **주장**이고, 해석 불가능한 이름은 그 주장의 근거가 아니다. 그래서 `any`·제네릭 파라미터·미해석 이름은 전부 false이고 `checkCapability` B5에서 소리 내어 실패한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `expr` | 해석된 필드 타입 | `resolveDeclared` | 인식 밖은 false |
| `goBuiltinTypes` | 메서드 집합이 확실히 없는 철자 19개 | static_test.go:1010 | 그 밖의 식별자는 선언을 따라간다 |
| `externalOptionTypes` | `{io.Writer}` | static_test.go:1007 | 다른 패키지 타입은 여기 적힌 것만 통과 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch v := expr.(type)` | 없음 | 아래 | func 타입 seam 8 + 평문 10 |
| B2 | `case *ast.FuncType` | 없음 | `true` | func 타입 seam 8 |
| B3 | `case *ast.StructType` | 없음 | `true` — 데이터. 필드 타입은 closure가 동사 검사한다 | 구조체 필드 |
| B4 | `case *ast.Ident` | 없음 | 아래 셋 | 평문 10 + 선언 이름 |
| B5 | `goBuiltinTypes[v.Name]` | 없음 | `true` | `Port int`, `SoakRecord string`, `MinSoakDays int` |
| B6 | `seen[v.Name]` | 없음 | `false` | 순환 별칭 변이 |
| B7 | `decl, ok := declaredTypes[v.Name]` | 없음 | 재귀 | 패키지 선언 이름 |
| B8 | `case *ast.SelectorExpr` | 없음 | `externalOptionTypes`에 있으면 true | `Out io.Writer` |
| B9 | `case *ast.StarExpr` | 없음 | 재귀 | 포인터 — 현재 `Options`에 없음 |
| B10 | `case *ast.ParenExpr` | 없음 | 재귀 | 현재 표면에 없음(방어) |
| B11 | `case *ast.ArrayType` | 없음 | 재귀(Elt) | `RequiredEndpoints []string` |
| B12 | `case *ast.MapType` | 없음 | Key **와** Value 둘 다 | 현재 표면에 없음(방어) |
| B13 | `case *ast.ChanType` | 없음 | 재귀(Value) | 현재 표면에 없음(방어) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (재귀) | 선언 사슬을 따라간다 | `seen`이 순환을 끊는다 | ast.json calls |

## State mutations and fallbacks

- 없음(순수 함수).

## Safety conclusion

- Safe edit boundary: 신설. 이전 가드에는 이 판정이 아예 없었고, nil로 열거된 필드는 인터페이스 오류를 건너뛰었다.
- High-risk impact: yes (주문 능력 주입 차단 — '메서드 집합 없음' 주장의 유일한 근거 판정)

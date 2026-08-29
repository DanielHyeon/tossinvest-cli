# Function Logic Map: `opaqueHandler`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

등록의 핸들러 인자에서 게이트를 읽어낼 수 없는 모양을 보고한다.

**잡는 것**: 식별자(다른 곳에서 대입된 값 — `*http.ServeMux`도 합법적인 값이다), 복합 리터럴, 주소 연산, 그 밖에 아래 넷이 아닌 모든 노드.
**통과시키는 것**: 게이트 래퍼의 연쇄(CallExpr, 인자 중 BasicLit은 래퍼 자신의 인자로 보고 건너뛴다), 메서드 값(SelectorExpr), 함수 리터럴(FuncLit), 괄호.

**측정된 경계**: `SelectorExpr`를 메서드 값으로 보고 통과시키므로 `mux.Handle("/x", c.subMux)` 같은 **필드 셀렉터**는 불투명으로 잡히지 않는다. 그 모양은 다른 가드가 잡는다 — 게이트 체인이 없으므로 `Session=false`가 되고 `TestEveryRouteGoesThroughTheSessionGate`가 실패한다. 이 함수 혼자가 그 구멍을 막는다고 읽으면 안 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `expr` | 등록의 두 번째 이후 인자 | `registeredRoutes` | 인식하지 못하는 노드는 **불투명으로 판정**한다 — 모르는 것을 괜찮다고 읽지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch v := expr.(type)` | 없음 | 아래 넷 | 표 형태의 직접 테스트는 없고 실제 표 등록 19건이 커버한다 |
| B2 | `case *ast.CallExpr` | 없음 | 인자 재귀 | 모든 게이트 체인 등록 |
| B3 | `for _, arg := range v.Args` | 없음 | 없음 | 같은 위 |
| B4 | `if _, ok := arg.(*ast.BasicLit); ok` | 없음 | continue — 래퍼 자신의 리터럴 인자 | 현재 표에는 리터럴 인자를 받는 래퍼가 없다 |
| B5 | `opaqueHandler(arg)` 재귀 참 | 없음 | `true` | 변이 대상 — 식별자 핸들러를 등록하면 참이 된다 |
| B6 | `case *ast.SelectorExpr, *ast.FuncLit` | 없음 | `false` | `c.handleX` 형태 전부 |
| B7 | `case *ast.ParenExpr` | 없음 | 내부 재귀 | 현재 표에 없음 |
| B8 | `default` | 없음 | `true` | 위 경계 설명의 ④ 부류 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (재귀) | 인자를 따라 내려간다 | 순수 AST 판정 | ast.json calls |

## State mutations and fallbacks

- 없음(순수 함수).

## Safety conclusion

- Safe edit boundary: 신설. 이전에는 핸들러 인자를 아예 검사하지 않았다.
- High-risk impact: yes (라우트 게이트 감시 — 게이트를 등록에서 읽을 수 없는 라우트는 게이트가 없는 라우트와 구별되지 않는다)

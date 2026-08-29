# Function Logic Map: `receiverIsConsole`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

메서드가 `*Console`(또는 `Console`) 위의 것인지 보고한다. `TestNoCapabilityReachesTheConsoleAroundOptions`의 exported-메서드 걷기가 이 판정 위에 서 있다.

**기록된 경계**: 이 함수가 항상 false를 답해도 그 걷기는 조용히 비고 **테스트는 통과한다** — 그 테스트의 positive control은 `len(seams)`뿐이고 exported-메서드 걷기 건수를 세는 단언은 없다. 가드의 가드가 없는 자리다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `recv` | 메서드 수신자 필드 목록 | `*ast.FuncDecl.Recv` | nil 또는 빈 목록이면 false |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `recv == nil || len(recv.List) == 0` | 없음 | `false` | 함수 선언(수신자 없음) |
| B2 | `star, ok := expr.(*ast.StarExpr); ok` | `expr = star.X` | 없음 | `*Console` 수신자 전부 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | 타입 단언뿐 | 순수 | ast.json calls=null |

## State mutations and fallbacks

- 없음(순수 함수).

## Safety conclusion

- Safe edit boundary: 신설. `Options` 밖 걷기의 대상 선별 기준이다.
- High-risk impact: yes (주문 능력 주입 차단의 걷기 범위 — 위에 적은 대로 이 판정이 조용히 좁아지면 걷기가 공허해지고 아무 테스트도 실패하지 않는다)

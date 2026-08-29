# Function Logic Map: `capabilityClosure`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

한 선언에서 도달 가능한 모든 것을 **고정점까지** 걷는다 — 타입 자신, 시그니처의 타입들, 그 선언들이 부르는 타입들 — 그리고 동사 검사할 모든 식별자와 도중에 만난 모든 인터페이스를 돌려준다.

**구조체 필드 이름은 그 필드의 타입이 능력을 가질 수 있을 때만 검사한다**. `GateLimits.MaxOrderNotional`은 주문에 대한 **상한**이지 주문이 아니고, 그것에 실패하면 다음 사람은 상한의 이름을 바꾸게 된다. `PlaceOrder func(...)`라는 필드는 여전히 실패한다 — func 타입은 능력을 가질 수 있기 때문이다.

**측정된 경계**: 이 워커가 돌려주는 것은 **이름**이고, 이름 위에서 도는 것은 철자 필터다. `interface{ Do(context.Context, domain.Position) error }`는 어떤 금지 동사도 쓰지 않으므로 이 워커를 그대로 통과한다. 그것을 잡는 것은 `Options` 필드에 대한 **메서드 집합 대조**이고, `Options` 밖에서는 그 대조가 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `expr` | 시작 선언 | 필드 타입 / 메서드 시그니처 / 패키지 var 타입 | 인식하지 못하는 노드는 이름을 내지 않는다 — 그 경우의 안전망은 `methodless`의 긍정형이다 |
| `declaredTypes` | 패키지 선언 타입 색인 | `packageTypes` | 없는 이름은 더 내려가지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `push`: `e == nil || seenNode[e]` | 없음 | return | 고정점 종료 조건 |
| B2 | `addName`: `seenName[n]` | 없음 | return | 같은 위 |
| B3 | `addName`: `decl, ok := declaredTypes[n]` | 선언을 큐에 넣는다 | 없음 | 선언된 타입 이름 전부 |
| B4 | `pushFieldTypes`: `fl == nil` | 없음 | return | 인자·결과 없는 시그니처 |
| B5 | `for _, f := range fl.List` | 필드 타입 push | 없음 | func 타입 seam 8 |
| B6 | `for len(queue) > 0` | 큐 소진 | 없음 | 전 필드 |
| B7 | `switch v := e.(type)` | 없음 | 아래 분기 | 전 필드 |
| B8 | `case *ast.Ident` | `addName` | 없음 | 이름 있는 모든 타입 |
| B9 | `case *ast.SelectorExpr` | `addName(Sel)` | 없음 | `io.Writer`, `context.Context`, `domain.Position`, `time.Time` 등 |
| B10 | `case *ast.StarExpr` | `push(X)` | 없음 | 포인터 타입 |
| B11 | `case *ast.ParenExpr` | `push(X)` | 없음 | 현재 표면에 없음 — 방어 |
| B12 | `case *ast.ArrayType` | `push(Elt)` | 없음 | `[]string`(RequiredEndpoints), `[]domain.Position` |
| B13 | `case *ast.Ellipsis` | `push(Elt)` | 없음 | 가변 인자 — 현재 표면에 없음 |
| B14 | `case *ast.MapType` | `push(Key)`, `push(Value)` | 없음 | 현재 표면에 없음 — 방어 |
| B15 | `case *ast.ChanType` | `push(Value)` | 없음 | 현재 표면에 없음 — 방어 |
| B16 | `case *ast.IndexExpr` | `push(X)`, `push(Index)` | 없음 | 제네릭 seam 변이 `Seam[OrderPlacer]` — 이 케이스가 없어서 이름이 하나도 안 나왔다 |
| B17 | `case *ast.IndexListExpr` | `push(X)` + 인덱스 전부 | 없음 | `Seam[A, B]` 변이 |
| B18 | `for _, idx := range v.Indices` | 인덱스 push | 없음 | 같은 위 |
| B19 | `case *ast.FuncType` | TypeParams·Params·Results push | 없음 | func 타입 seam 8 |
| B20 | `case *ast.InterfaceType` | 인터페이스 수집 + 메서드 이름 `addName` + 메서드 타입 push | 없음 | 인터페이스 seam 6 |
| B21 | `for _, m := range v.Methods.List` | 같은 위 | 없음 | 같은 위 |
| B22 | `for _, n := range m.Names` | 메서드 이름 수집 | 없음 | 같은 위 — `Orders`·`Holdings`·`GateLimits`·`Load`·`Save`·`Mint`·`Consume`·`Signals` |
| B23 | `case *ast.StructType` | 필드 순회 | 없음 | `GateLimits` 값 타입, `domain.Position` 등 |
| B24 | `for _, f := range v.Fields.List` | 같은 위 | 없음 | 같은 위 |
| B25 | `carriesCapability(f.Type, …)` | 능력 가능 필드의 **이름**을 수집 | 없음 | `GateLimits.MaxOrderNotional`이 실패하지 않는 이유이자 `PlaceOrder func(...)`가 실패하는 이유 |
| B26 | `for _, n := range f.Names` | 필드 이름 수집 | 없음 | 같은 위 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `carriesCapability` | 구조체 필드 이름을 검사할지 | 보수적 — 모르는 것은 참 | static_test.go:1247 |
| (재귀 없음, BFS 큐) | 고정점 | `seenNode`/`seenName`이 종료를 보장 | ast.json branches |

## State mutations and fallbacks

- 없음(순수 함수). 큐와 두 방문 집합만 쓴다.

## Safety conclusion

- Safe edit boundary: 신설. 이전에는 필드 타입을 한 홉 따라가고 지나온 이름의 철자만 봤다.
- High-risk impact: yes (주문 능력 주입 차단의 도달성 계산 — 여기서 놓친 타입은 어느 검사도 열어 보지 않는다)

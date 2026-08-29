# Function Logic Map: `TestNoCapabilityReachesTheConsoleAroundOptions`

- Source: `internal/console/static_test.go`
- Change: `console-orders-screen`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit에 이미 존재했고, 이번 작업에서 본문이 수정됐다(positive control 셋 — 세 걷기가 각자 공회전하지 않는다는 대조). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

`Options` 걷기는 '**그 구조체를 통해** 건네지는 모든 능력이 열거돼 있는가'를 답한다. `consoleCapabilities`가 실제로 하는 말은 '이 콘솔이 받는 모든 능력이 여기 열거돼 있다'이고, 그 차이는 **패키지 수준 변수 하나만큼** 넓다:

```go
var packageDesk Desk
func (c *Console) SetDesk(d Desk) { packageDesk = d }
```

`Desk`에 `PlaceOrder`·`CancelOrder`를 달아도 **패키지 전 스위트가 통과했다**. `Options`는 그것을 언급하지 않고, 금지 import도 필요 없으며(cmd/tossctl가 채운다), 이 파일의 다른 어떤 것도 그 구조체 말고는 보고 있지 않았다.

**잡는 것**: `*Console`의 모든 **exported** 메서드(이름 + 시그니처 closure), 모든 패키지 수준 var(이름 + **명시 타입이 있으면** closure), 패키지 어디든 선언된 인터페이스 중 열거된 seam이 아닌 것(embed + 메서드 이름) — 타입 단언의 인라인 `interface{ PlaceOrder(…) }`를 포함한다.

**잡지 못하는 것(측정된 경계, 셋)**: ① `Options` 밖에서는 **메서드 집합 대조가 없다**. 여기의 검사는 전부 이름 필터이므로 `interface{ Do(context.Context, domain.Position) error }`는 통과한다. ② `vs.Type == nil`이면 closure를 걷지 않는다 — `var desk = newDesk()`처럼 타입을 추론으로 받은 패키지 var는 **이름만** 검사된다. ③ unexported 메서드는 walk 대상이 아니다(`d.Name.IsExported()`). ①②③ 모두 이 가드가 열거된 seam 규율의 **보완재**이지 대체재가 아님을 뜻한다.

**positive control 셋(B21·B22·B25·B26)**: 이 함수의 단언은 전부 '아무것도 못 찾았다' 형태이므로, 세 걷기 중 어느 것이든 **아무것도 보지 않으면 가장 조용히 통과한다**. 그래서 걷기마다 자기 대조가 붙는다 — seam 집합이 비지 않았는가(B21), 패키지 전역 걷기가 인터페이스 선언에 **닿기는 했는가**(B22), 그 걷기가 `Options` 걷기와 **같은 패키지를 보고 있는가**(B23–B25), `*Console`의 exported 메서드를 하나라도 검사했는가(B26).

의도적으로 단언하지 **않는** 것: '열거된 seam 밖의 인터페이스가 하나 이상 검사됐다'. 현재 이 패키지가 선언하는 인터페이스 6개는 **전부** `Options` seam이고(측정: `optionsSeamInterfaces` 6개 = 패키지 전역 걷기가 만난 `*ast.InterfaceType` 6개), 그것은 걷기가 고장난 상태가 아니라 건강한 상태다. 우발적 `interface{ PlaceOrder(…) }`의 등장은 이 가드가 **막으려는 사건**이지 가드가 성립하기 위한 전제가 아니다. 그래서 그 걷기의 대조는 '비-seam을 셌다'가 아니라 '선언에 닿았고, `Options` 걷기가 자기 경로로 찾아낸 것과 같은 노드에 닿았다'로 세운다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `parsedPackage(t)` | 비테스트 소스 전부 | 디스크 | 0개면 Fatal |
| `packageTypes(files)` | 선언 타입 색인 | 같은 곳 | 비면 해석이 멈추고 seam 집합이 비어 B21이 실패한다 |
| `optionsSeamInterfaces(...)` | 열거된 seam인 인터페이스 집합 | `Options` 필드 해석 | 0개면 `t.Error`(B21) — positive control |
| `visitedInterfaces` | 패키지 전역 걷기가 만난 `*ast.InterfaceType` 노드 집합 | `ast.Inspect` | 비면 `t.Error`(B22); seam을 빠뜨리면 `t.Errorf`(B25) |
| `checkedConsoleMethods` | `*Console`의 exported 메서드 검사 횟수 | `receiverIsConsole` + `IsExported` | 0이면 `t.Error`(B26) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for name := range files` | 이름 수집 | 없음 | 전 파일 |
| B2 | `for _, name := range names`(정렬) | 없음 | 없음 | 같은 위 |
| B3 | `for _, decl := range file.Decls` | 없음 | 없음 | 최상위 선언 전부 |
| B4 | `switch d := decl.(type)` | 없음 | 아래 | 같은 위 |
| B5 | `case *ast.FuncDecl` | 없음 | 아래 | 함수·메서드 선언 |
| B6 | `d.Recv == nil \|\| !d.Name.IsExported() \|\| !receiverIsConsole(d.Recv)` | 아니면 `checkedConsoleMethods++` | continue | **경계 ③** — unexported 메서드는 여기서 빠진다; 전부 빠지면 B26이 잡는다 |
| B7 | `for _, n := range closureNames` | 동사 검사 | 없음 | `SessionToken`·`Handler`·`Addr`·`URL`·`Serve` 등 |
| B8 | `for _, iface := range ifaces` | embed 검사 | 없음 | 같은 위 |
| B9 | `case *ast.GenDecl` | 없음 | 아래 | var 선언 |
| B10 | `d.Tok != token.VAR` | 없음 | continue | const·type·import |
| B11 | `for _, spec := range d.Specs` | 없음 | 없음 | var 선언 |
| B12 | `vs, ok := spec.(*ast.ValueSpec); !ok` | 없음 | continue | 구조 분기 |
| B13 | `for _, ident := range vs.Names` | 이름 동사 검사 | 없음 | 패키지 var 전부 |
| B14 | `vs.Type == nil` | 없음 | continue | **경계 ②** — 추론 타입 var는 이름만 검사된다 |
| B15 | `for _, n := range closureNames` | 동사 검사 | 없음 | 명시 타입 var |
| B16 | `for _, iface := range ifaces` | embed 검사 | 없음 | 같은 위 |
| B17 | `iface, ok := n.(*ast.InterfaceType); !ok` | 없음 | 순회 계속 | 인터페이스가 아닌 노드 전부 |
| B18 | `seams[iface]`(B17 통과 후, `visitedInterfaces[iface] = true` 뒤) | 방문 기록 | 순회 계속 | 열거된 seam 6개는 여기서 제외된다 — 단, **방문 기록은 남는다**(B25가 그것을 읽는다) |
| B19 | `for _, m := range iface.Methods.List` | 없음 | 없음 | seam이 아닌 인터페이스 |
| B20 | `for _, mn := range m.Names` | 메서드 이름 동사 검사 | 없음 | 같은 위 — 인라인 `interface{ PlaceOrder(…) }` 변이 |
| B21 | `len(seams) == 0` | `t.Error` | 없음 | positive control — `Options` 걷기 |
| B22 | `len(visitedInterfaces) == 0` | `t.Error` | 없음 | positive control — 패키지 전역 인터페이스 걷기 |
| B23 | `for iface := range seams` | 미방문 seam 계수 | 없음 | 두 걷기의 교차 대조 |
| B24 | `!visitedInterfaces[iface]` | `missedSeams++` | 없음 | 같은 위 |
| B25 | `missedSeams > 0` | `t.Errorf` | 없음 | positive control — 두 걷기가 같은 패키지를 보는가 |
| B26 | `checkedConsoleMethods == 0` | `t.Error` | 없음 | positive control — `*Console` 메서드 걷기 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles(t)` | 패키지의 **비테스트** `.go` 파일 전부 | 0개면 `t.Fatal` | static_test.go:36 |
| (런타임 무접촉) | 가드는 **선언**을 읽는다 — 존재하지만 호출되지 않는 메서드가 정확히 잡아야 할 대상이다 | reflect를 쓰지 않는다 | static_test.go:936 |
| `receiverIsConsole` | `*Console` 수신자 판정 | 전부 false면 메서드 걷기가 통째로 빈다 — B26이 그것을 잡는다 | static_test.go:1546 |
| `capabilityClosure` / `checkVerbs` / `checkNoEmbedding` | 이름 필터와 embed 검사 | **메서드 집합 대조는 없다** | static_test.go:1155, 1329, 1114 |
| `optionsSeamInterfaces` | seam 식별 | 0개면 `t.Error`(B21) | static_test.go:1561 |
| `ast.Inspect` | 패키지 전역 인터페이스 걷기 | 아무 노드에도 닿지 않으면 B22·B25가 잡는다 | static_test.go:1487 |

## State mutations and fallbacks

- 판정 전용. 소스도 디스크도 건드리지 않는다.
- 지역 상태 둘만 쓴다. `visitedInterfaces`는 패키지 전역 걷기가 **만난** 인터페이스 노드 집합이고, `checkedConsoleMethods`는 `*Console` 메서드 걷기가 **검사한** 횟수다. 둘 다 사후에 재구성할 수 없어서 걷는 자리에서 센다 — 아무것도 보지 않은 걷기는 다른 흔적을 남기지 않기 때문이다.
- 두 값은 모두 **단언에 쓰인다**(B22·B25·B26). 이전 판본은 인터페이스 쪽 계수를 세기만 하고 `_ = checkedInterfaces`로 버렸고, 바로 위 주석은 '어느 쪽이든 0이면 걷기가 멈춘 것'이라고 적으면서 실제로는 `seams` 한쪽만 단언했다 — 즉 P0-3을 막으려고 만든 가드 안에 같은 모양의 과장이 들어 있었다. 지금은 세 걷기가 각자 자기 대조를 갖는다.
- 대조는 '비-seam 인터페이스를 하나 이상 셌다'가 아니다. 그 수는 현재 정당하게 0이며(패키지의 인터페이스 6개가 전부 seam), 0을 실패로 만들면 건강한 트리가 빨개진다. 대신 '선언에 닿았다'(B22)와 '`Options` 걷기가 찾은 seam 노드에 전부 닿았다'(B25)를 단언한다 — 후자는 두 걷기가 서로 다른 경로로 같은 AST 노드에 도달했음을 요구하므로, 파일 하나가 조용히 걷기에서 빠지는 변이도 잡는다.

## Safety conclusion

- Safe edit boundary: `Options` 걷기의 사각(패키지 수준 능력)을 범위에 넣은 신설 가드에, 그 가드 자신이 공회전하지 않는다는 대조를 더했다. 범위 안 검사는 여전히 이름 필터다(경계 ①②③ 불변).
- 변이 확인: `receiverIsConsole`이 항상 false를 답하면 B26이 실패한다(수정 전에는 통과했다). 패키지 전역 걷기가 첫 노드에서 멈추면 B22·B25가 실패한다(수정 전에는 통과했다). 파일 하나만 걷기에서 빠지면 B25가 `6개 중 1개`로 실패한다.
- High-risk impact: yes (주문 능력 주입 차단의 `Options` 밖 절반 — 시연된 `SetDesk` 우회가 이 가드 이전에는 전 스위트를 통과했고, 대조가 없던 동안에는 가드 자신이 빈 채로도 전 스위트가 통과했다)

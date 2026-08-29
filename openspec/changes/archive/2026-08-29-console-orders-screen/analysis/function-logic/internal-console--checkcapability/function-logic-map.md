# Function Logic Map: `checkCapability`

- Source: `internal/console/static_test.go`
- Change: `console-orders-screen`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 change의 diff가 이 함수의 본문을 바꿨다. 아래 분석은 현재 HEAD 본문에 대한 것이다.

`Options` 필드 하나의 선언 타입을 **고정점까지** 해석하고 열거된 메서드 집합에 붙든다.

**한 홉이 아니라 고정점인 이유 — 리뷰가 변이로 통과시킨 네 모양**: ① 두 번째 인터페이스 — `HoldingsReader.Holdings`가 `PlaceOrder`·`CancelOrder`·`Flatten`을 선언한 `AccountHandle`을 추가로 돌려주는 형태. 이전 가드는 이름 `AccountHandle`을 동사 검사하고 아무것도 못 찾고 **열어 보지 않았다**. 같은 타입을 `OrderHandle`로 개명하면 실패했다 — 검사되던 것이 메서드 집합이 아니라 **단어**였다는 증명이다. ② 제네릭 seam `type Desk Seam[OrderPlacer]` — 워커에 `*ast.IndexExpr` 케이스가 없어 이름을 하나도 돌려주지 않았고 동사 검사가 빈 목록 위에서 돌았다. ③ 별칭 사슬 `type Ticker = Wide` — 한 홉은 `*ast.Ident`에 내려앉고 인터페이스가 아니며, nil로 열거된 필드는 그 오류를 건너뛰었다. ④ 빈 정적 타입 `Feed any` — 사용 지점에서 `interface{ PlaceOrder(...) }`로 타입 단언한다. 무엇이든 담을 수 있는 필드는 열거가 기술할 수 없는 필드다.

**잡지 못하는 것**: 다른 패키지의 메서드 집합. 이 가드는 이 패키지의 소스를 읽으므로 `io.Writer` 같은 한정 타입은 읽을 수 없고, 그래서 `externalOptionTypes`에 **하나씩 근거와 함께** 적어야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `expr` | 필드의 선언 타입 | `optionsFields` | 해석 불가능한 모양이면 소리 내어 실패한다 |
| `cap.Methods` | 허용 메서드 집합(빈 목록 = '메서드 집합 없음'이라는 주장) | `consoleCapabilities` | 양방향 대조 — 없는 것도 남는 것도 실패 |
| `cap.VerbExemptions` | 철자 전체 키의 예외 | 같은 곳 | `Orders` seam에만 있고 다섯뿐 |
| `declaredTypes` | 패키지 선언 타입 색인 | `packageTypes` | 비어 있으면 `methodless`가 false를 답해 실패한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, name := range names`(closure의 모든 이름) | 동사 검사 | 없음 | `Orders` seam의 다섯 철자 예외 + 나머지 전 필드 |
| B2 | `for _, iface := range ifaces` | embed 검사 | 없음 | 인터페이스 seam 6개 |
| B3 | `!isInterface` | 없음 | 아래 둘 뒤 return | func 타입 seam 8개 + 평문 데이터 10개 |
| B4 | `len(allowed) > 0`(인터페이스가 아닌데 메서드가 열거됨) | `t.Errorf` | 없음 | 별칭 사슬 변이(④가 아니라 ③) |
| B5 | `!methodless(resolved, …)` | `t.Errorf` | 없음 | `Feed any` 변이 |
| B6 | `len(seam.Methods.List) == 0` | `t.Errorf` — 빈 인터페이스 | return | 빈 인터페이스 seam 변이 |
| B7 | `for _, method := range seam.Methods.List` | 선언 수집 | 없음 | 인터페이스 seam 6개 |
| B8 | `len(method.Names) == 0` | 없음(embed는 `checkNoEmbedding`이 보고) | continue | embed 삽입 변이 |
| B9 | `for _, name := range method.Names` | 선언 수집 | 없음 | 같은 위 |
| B10 | `for _, name := range allowed` | `want` 구성 | 없음 | 같은 위 |
| B11 | `for name := range declared` | 없음 | 없음 | 같은 위 |
| B12 | `!want[name]` | `t.Errorf` — 허용 밖 메서드 | 없음 | 두 번째 메서드 추가 변이 |
| B13 | `for name := range want` | 없음 | 없음 | 같은 위 |
| B14 | `!declared[name]` | `t.Errorf` — 약속한 메서드 상실 | 없음 | 메서드 제거 변이 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `capabilityClosure` | 필드에서 도달 가능한 모든 이름·인터페이스 | 고정점 | static_test.go:1155 |
| `checkVerbsExcept` | 이름 필터 | 예외는 철자 전체 키 | static_test.go:1341 |
| `checkNoEmbedding` | embed 금지 | 도달 가능한 모든 인터페이스에 | static_test.go:1114 |
| `resolveDeclared` | 이름 사슬을 끝까지 | 순환 방지 | static_test.go:1130 |
| `methodless` | **긍정형** — 메서드 집합이 없다고 적극적으로 읽을 수 있는 모양 | '가드가 알아보지 못했다'가 '괜찮다'로 읽히면 안 된다 | static_test.go:1293 |

## State mutations and fallbacks

- 없음(판정 전용).

## Safety conclusion

- Safe edit boundary: 신설. 이전 가드는 필드 타입을 한 홉 따라가고 지나온 이름의 **철자**를 검사했다.
- High-risk impact: yes (주문 능력 주입 차단의 실질 검사 — 메서드 집합 대조가 여기 있고, 동사 검사는 그 위의 보조다)

# Function Logic Map: `TestTheConsoleBrokerInterfaceDeclaresNothingButReads`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**HEAD에는 없는 함수다** — base에 존재했고 이 change가 삭제했다. `ast.json`은 base revision의 것이고 아래 분석은 base 본문에 대한 것이다.

**HEAD에는 없는 함수다.** base revision에 존재했고 console-operator-overview가 `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads`로 대체했다. AST·hash는 base revision의 것이고, 아래 분석은 base 본문에 대한 것이다.

**이 가드가 잡던 것**: `holdings.go`에 선언된 `HoldingsReader`의 메서드 집합이 `{Holdings}`인가, 그 인터페이스가 다른 인터페이스를 embed하는가, 그리고 패키지 어디에도 `verifylive.Broker`가 이름으로 등장하지 않는가.

**대체된 이유 — 리뷰가 변이로 확인한 구멍 셋**: ① `packageFiles(t)["holdings.go"]` 한 파일만 읽었다. 다른 파일에 선언된 광폭 인터페이스는 아무것도 실패시키지 않았다. ② 주입 seam 일곱 중 **다섯이 func 타입**(`StartVerify`·`StartEngine`·`StopEngine`·`Relaunch`·`RestartSoak`)이고 `*ast.InterfaceType` 스캔은 그 중 하나도 보지 못한다 — `type PlaceOrderFunc func(...)`는 인터페이스도 없고 금지 import도 없이(cmd/tossctl가 채우므로) 그대로 통과했다. ③ allowlist가 하나뿐이라 `Handoff{Mint,Consume}`·`AdoptionSettings{Load,Save}`(spec이 명시 요구)를 담으려면 넓혀야 했고, 넓히면 `interface{ Holdings(...); PlaceOrder(...) }`가 통과한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `packageFiles(t)["holdings.go"]` | 존재해야 한다 | 디스크 | 없으면 `t.Fatal` — 파일 이름 고정이 구멍 ①이었다 |
| `allowed` | `{Holdings}` | 테스트 본문 | 그 밖의 메서드는 발견 |
| `banned` | 14개 동사 | 테스트 본문 | 부분 문자열 일치 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `src == ""` | 없음 | `t.Fatal` | 파일 삭제 변이 |
| B2 | TypeSpec이 `HoldingsReader`가 아님 | 없음 | 순회 계속 | 구조 분기 |
| B3 | `HoldingsReader`가 인터페이스가 아님 | 없음 | `t.Fatal` | 타입 교체 변이 |
| B4 | `for _, method := range iface.Methods.List` | 메서드 수집 | 없음 | 선언 1개 |
| B5 | `for _, name := range method.Names` | 이름 수집 | 없음 | 같은 위 |
| B6 | `len(method.Names) == 0` | `t.Error` — embed | 없음 | embed 삽입 변이 |
| B7 | `len(declared) == 0` | 없음 | `t.Fatal` | positive control — 인터페이스를 못 읽은 상태 |
| B8 | `for _, name := range declared` | 없음 | 없음 | 선언 1개 |
| B9 | `!allowed[name]` | `t.Errorf` | 없음 | 두 번째 메서드 추가 변이 |
| B10 | `for _, verb := range banned` | 없음 | 없음 | 같은 위 |
| B11 | `strings.Contains(lowered, verb)` | `t.Errorf` | 없음 | `PlaceOrder` 추가 변이 |
| B12 | `for name, fileSrc := range packageFiles(t)` | 없음 | 없음 | 패키지 전 파일 |
| B13 | `strings.Contains(code, "verifylive.Broker")` | `t.Errorf` | 없음 | 광폭 브로커 이름 사용 변이 — 이 절반만 패키지 전체를 봤다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles(t)` | 패키지의 **비테스트** `.go` 파일 전부 | 0개면 `t.Fatal` | static_test.go:36 |
| (런타임 무접촉) | 가드는 **선언**을 읽는다 — 존재하지만 호출되지 않는 메서드가 정확히 잡아야 할 대상이다 | reflect를 쓰지 않는다 | static_test.go:936 |
| `parseFile` | `holdings.go` 파싱 | 실패 시 Fatal | static_test.go:60 |
| `nonCommentLines` | 주석을 뺀 코드만 문자열 검사 | 주석에 적힌 이름이 발견을 만들지 않는다 | static_test.go:1680 |

## State mutations and fallbacks

- 없음(판정 전용).

## Safety conclusion

- Safe edit boundary: HEAD에서 삭제됨. 대체 가드는 검사 단위를 '인터페이스'에서 '`Options`가 받는 능력(필드)'으로 옮겼고, `verifylive.Broker` 문자열 검사(B12·B13)는 그대로 옮겨 갔다.
- High-risk impact: yes (주문 능력 주입 차단 — 이 가드가 덮던 표면이 '콘솔은 주문을 낼 수 없다'를 핸들러가 무엇을 부르는가가 아니라 **무엇을 부를 수 있는가**의 사실로 만든다)

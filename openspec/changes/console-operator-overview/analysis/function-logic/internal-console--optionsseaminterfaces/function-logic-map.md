# Function Logic Map: `optionsSeamInterfaces`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

**열거된 seam인** 인터페이스 선언의 집합. 패키지 전체 걷기가 나머지 인터페이스를 더 엄격한 '금지 동사 하나도 없음' 규칙에 붙들면서 seam을 다시 보고하지 않게 한다 — seam은 근거가 적힌 예외를 갖고 있기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `files` | 파싱된 전 파일 | `parsedPackage` | 해당 없음 |
| `declaredTypes` | 선언 타입 색인 | `packageTypes` | 해석 실패면 seam이 집합에 안 들어가고 더 엄격한 규칙을 받는다 — 보수적 방향 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, field := range optionsFields(t, files)` | 없음 | 없음 | `Options` 필드 26개 |
| B2 | `iface, ok := resolveDeclared(...).(*ast.InterfaceType); ok` | 집합에 추가 | 없음 | 인터페이스 seam 6개 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `optionsFields` | 필드 목록 | 선언이 1이 아니면 Fatal | static_test.go:1355 |
| `resolveDeclared` | 이름 사슬 해석 | 고정점 | static_test.go:1130 |

## State mutations and fallbacks

- 없음(집합 생성).

## Safety conclusion

- Safe edit boundary: 신설. 패키지 전체 걷기가 seam과 나머지를 가르는 유일한 기준이다.
- High-risk impact: yes (주문 능력 주입 차단 — 집합이 너무 넓으면 seam 아닌 인터페이스가 예외를 얻고, 비면 `TestNoCapabilityReachesTheConsoleAroundOptions` B20이 소리를 낸다)

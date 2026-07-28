# Function Logic Map: `optionsFields`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

`Options` 구조체를 패키지 **어디에 선언돼 있든** 찾아 필드 목록을 돌려준다. 선언이 정확히 하나가 아니면 `t.Fatalf` — 주입 표면은 구조체 하나다. 그것이 이 함수의 positive control이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `files` | 파싱된 패키지 전 파일 | `parsedPackage` | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for name, file := range files` | 없음 | 없음 | 전 파일 |
| B2 | `spec.Name.Name != "Options"` | 없음 | 순회 계속 | 구조 분기 |
| B3 | `Options`가 구조체가 아님 | 없음 | `t.Fatalf` | 타입 교체 변이 |
| B4 | `found != 1` | 없음 | `t.Fatalf` | positive control — 0이면 검사가 표면을 못 봤고 2 이상이면 표면이 둘이다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ast.Inspect` | TypeSpec 탐색 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음(조회 전용).

## Safety conclusion

- Safe edit boundary: 신설. 이전 가드는 `holdings.go`라는 파일 이름에 고정돼 있었다.
- High-risk impact: yes (주입 표면 검사의 입력 — 비면 능력 열거가 공허해진다. B4가 그것을 Fatal로 막는다)

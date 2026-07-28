# Function Logic Map: `parsedPackage`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

비테스트 소스 전부를 한 번 파싱한다. `packageFiles`의 `len(out) == 0` Fatal이 이 함수의 positive control을 대신한다 — 파일을 하나도 못 읽으면 그쪽에서 멈춘다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `packageFiles(t)` | 비테스트 소스 | 디스크 | 0개면 `t.Fatal` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for name, src := range packageFiles(t)` | `out[name] = parseFile(...)` | 없음 | 전 파일 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles` | 파일 읽기 | 0개면 Fatal | static_test.go:36 |
| `parseFile` | 파싱 | 실패 시 Fatal | static_test.go:60 |

## State mutations and fallbacks

- 없음(조회 전용).

## Safety conclusion

- Safe edit boundary: 신설. 이전에는 각 검사가 필요한 파일을 그때그때 파싱했다.
- High-risk impact: yes (주입 표면 검사의 입력)

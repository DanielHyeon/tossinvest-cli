# Function Logic Map: `packageTypes`

- Source: `internal/console/static_test.go`
- Change: `console-operator-overview`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

이 함수는 이 change의 base commit 이후에 **신설**됐다(base에 같은 이름의 선언이 없다). gate가 evidence를 요구하는 것은 diff hunk가 현재 본문과 교차하기 때문이며, 아래 분석은 현재 HEAD 본문에 대한 것이다.

패키지가 선언한 모든 타입을 색인한다. 필드가 부르는 이름을 실제 모양으로 해석할 수 있게 하는 입력.

**공허 실패 모드는 소리를 낸다**: 이 맵이 비면 `resolveDeclared`가 이름을 그대로 돌려주고 `methodless`는 미해석 이름에 false를 답하므로 `checkCapability` B5가 실패한다. 이 함수에 별도 positive control이 없는 것은 그 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `files` | 파싱된 패키지 전 파일 | `parsedPackage` | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, file := range files` | 없음 | 없음 | 전 파일 |
| B2 | `if spec, ok := n.(*ast.TypeSpec); ok` | `out[spec.Name.Name] = spec.Type` | 없음 | 같은 위 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ast.Inspect` | TypeSpec 수집 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음(색인 생성). 같은 이름이 두 번 선언되면 뒤가 이긴다 — Go에서는 컴파일되지 않는 상태다.

## Safety conclusion

- Safe edit boundary: 신설. 이전 가드에는 타입 해석 자체가 없었다.
- High-risk impact: yes (주입 표면 검사의 입력 — 다만 공허해지면 `methodless`의 긍정형이 소리를 낸다)

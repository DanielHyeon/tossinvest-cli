# Function Logic Map: `routeFindings`

- Source: `internal/console/static_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `consoleStateChanging` | 논증된 상태변경 라우트 | 스펙 본문 | 여기 없는 행위 라우트는 발견으로 보고된다 |
| `actVerbs` | 행위를 뜻하는 어휘 | 이 함수 | 어휘에 없는 동사는 **검사 자체가 보지 못한다** |
| `accountVerbs` | 계좌를 건드리는 어휘 | 공유 목록 | 예외 없이 금지 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range consoleStateChanging` | `allowed` 집합 구성 | — | `TestNoRouteNamesAnAccountMutation` |
| B2 | `range accountVerbs` | 없음 | 계좌 어휘 발견 | 같은 테스트 |
| B3 | 경로가 계좌 동사를 담고 읽기가 아님 | 없음 | finding | 같은 테스트 |
| B4 | `allowed[path] || reads` | 없음 | 조기 반환 | 같은 테스트 |
| B5 | `range actVerbs` | 없음 | — | 같은 테스트 |
| B6 | 경로가 행위 동사를 담음 | 없음 | finding — 논증되지 않은 행위 | 같은 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `routeReadsTheAccountRecord` | 읽기 전용 wrapper 적용 여부 | CSRF 부재를 읽기의 증거로 삼지 않는다 | CodeGraph + AST |

## State mutations and fallbacks

- 이 change의 변경은 `actVerbs`에 `"exclude"`를 더한 것이다.
- 근거: `"exclude"`는 `"include"`의 부분문자열이 **아니다**. 추가 전에는 논증되지 않은 `/settings/exclude`가 B5·B6를 통째로 비켜가 조용히 통과했다 — 이 가드의 주석이 경계하던 바로 그 경우다.
- 이 change의 라우트는 `consoleStateChanging`에 올라 있어 B4에서 조기 반환하므로, 새 어휘가 막는 것은 다음번이다.

## Safety conclusion

- Safe edit boundary: 어휘 목록 1개 원소. 어휘를 넓히면 거짓 양성이 늘 수 있으나, 상태변경 목록에 논증된 경로는 B4에서 면제된다.
- High-risk impact: no — 테스트다. 다만 이 어휘의 구멍은 미논증 쓰기 라우트를 통과시킨다.

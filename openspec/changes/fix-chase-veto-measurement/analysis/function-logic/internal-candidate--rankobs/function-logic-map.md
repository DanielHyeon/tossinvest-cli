# Function Logic Map: `rankObs`

- Source: `internal/candidate/metrics_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L52–62, 분기 1개)
- Risk scan: `risk-pattern-report.md`

순위 관측 fixture. 이 change가 `newly bool` 인자를 **유지하되** 본문에서 3-상태로
변환하도록 고쳤다.

`bool`을 남긴 것이 판단이다. 이 헬퍼의 모든 호출자는 **측정된 두 답 중 하나**를 단언하고
있고 그것이 그 테스트들의 주제다. 세 번째 상태는 이 헬퍼로 **일부러 철자할 수 없다** —
직전 읽기가 없는 소스의 읽기는 다른 fixture이고, 그것이 필요한 테스트는 직접 만든다.
그래서 "unknown"이 인자를 빠뜨려서 도착하는 일이 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `newly` | bool — 측정된 두 답만 | 호출자 | 세 번째 상태는 이 헬퍼로 만들 수 없다 |
| `rank`/`total` | 양수 | 호출자 | `validate`가 경계에서 검사 |
| `source` | `SourceID` | 호출자 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `newly` | `answer`를 `NewlyListedYes()`로 | 아니면 `NewlyListedNo()` | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewlyListedYes()` / `NewlyListedNo()` | 3-상태 생성자 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — `Observation` 값 하나를 만든다.
- fallback 없음. 미상은 이 헬퍼로 도달할 수 없고, 그것이 설계다.

## Safety conclusion

- Safe edit boundary: 테스트 헬퍼. 본문 3줄 변환 가산.
- High-risk impact: no (테스트 전용). 이 헬퍼가 `unknown`을 만들 수 있게 되면 `metrics_test.go`의 단언들이 인자를 빠뜨려 조용히 미측정을 재게 된다.

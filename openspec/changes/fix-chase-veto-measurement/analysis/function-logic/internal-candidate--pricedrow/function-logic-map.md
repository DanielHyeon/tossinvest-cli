# Function Logic Map: `pricedRow`

- Source: `internal/candidate/wiring_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L33–39, 분기 0개)
- Risk scan: `risk-pattern-report.md`

`wiring_test.go`의 가격 있는 순위 행 fixture. 이 change가 **두 자격 사실**을 더했다 —
`RankRequested: total`과 `NewlyListed: NewlyListedNo()`.

실제 소스는 둘 다 보고하고, 그것이 없는 행은 그 위치가 `seen_late`에 전혀 답할 수 없는
행이다. zero value로 뒀다면 이 패키지의 **모든 스캔 수준 테스트가 거부 쪽만** 시험하게
된다 — 맞추기 쉬운 절반이다.

다른 상태는 그것이 주제인 곳에서 명시적으로 만든다(`firstsighting_source_test.go`,
`truncation_test.go`) — "미상"이 인자를 빠뜨려 도달하는 상태가 되지 않도록.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rank`/`total` | 위치와 도착 행 수 | 호출자 | — |
| `RankRequested` | **`total`** — 온전한 읽기 | 이 헬퍼 | 0이면 스캔 테스트가 전부 `REQUEST_UNRECORDED` |
| `NewlyListed` | **`NewlyListedNo()`** | 이 헬퍼 | zero면 전부 `NEW_ENTRANT_UNKNOWN` |
| `price` | 가격 문자열 | 호출자 | 첫 가격 경로용 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (무분기) `Row` 하나를 만든다 | 없음 | 자격 둘이 채워진 행 | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne`(B9) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 없음 — 값 하나를 만든다.
- fallback 없음. 두 자격을 **명시적으로** 적는 것이 요점이다.

## Safety conclusion

- Safe edit boundary: 테스트 헬퍼. 필드 2개 가산.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk — 이 두 필드를 비우면 스캔 수준 테스트가 거부 경로만 시험하면서 통과한다.

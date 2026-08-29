# Function Logic Map: `storedFirstRank`

- Source: `internal/candidate/veto_test.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L32–37, 분기 0개)
- Risk scan: `risk-pattern-report.md`

`veto_test.go`의 최초 순위 fixture. 이 change가 **두 자격 사실**을 더했다.

둘 다 **평범한 측정된 경우**로 적는다 — 소스가 직전 읽기를 갖고 있었고 이 심볼이 거기
있었으며, 읽기는 온전했다(`Requested: total`). 그것이 최초 관측이 이 테스트들이 말하는
의미를 갖는 상태이기 때문이다.

zero value로 뒀다면 이 파일의 **모든** sighting이 미측정이 되고, 테스트들은 **아무것도 재지
않으면서 통과**했을 것이다.

다른 두 상태는 그것이 주제인 곳에서 직접 만든다 —
`TestAPositionFromASourcesFirstReadingCannotAnswerSeenLate`와
`TestATruncatedReadingsPositionIsNotAPercentile`이 자기 `FirstRank`를 만든다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rank`/`total` | 양수 | 호출자 | — |
| `at`/`source` | instant와 소스 | 호출자 | — |
| `NewlyListed` | **`NewlyListedNo()`** — 측정된 부정 | 이 헬퍼 | zero면 파일 전체가 미측정 |
| `Requested` | **`total`** — 온전한 읽기 | 이 헬퍼 | 0이면 `REQUEST_UNRECORDED` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (무분기) `FirstRank` 하나를 만든다 | 없음 | 자격 둘이 채워진 값 | `veto_test.go` 전반 · `store_test.go` 8곳 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 없음 — 값 하나를 만든다.
- fallback 없음. 이 헬퍼가 두 자격을 **명시적으로** 적는 것이 요점이다 — 누락으로 도달하는 상태가 없어야 한다.

## Safety conclusion

- Safe edit boundary: 테스트 헬퍼. 필드 2개 가산.
- High-risk impact: no (테스트 전용). 재는 성질은 High-risk — 이 두 필드를 비우면 `veto_test.go`와 `store_test.go`의 sighting 단언이 전부 미측정 경로로 빠져 조용히 통과한다.

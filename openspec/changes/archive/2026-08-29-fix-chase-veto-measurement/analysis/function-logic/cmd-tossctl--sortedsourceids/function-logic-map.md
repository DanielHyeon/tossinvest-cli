# Function Logic Map: `sortedSourceIDs`

- Source: `cmd/tossctl/candidate.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L955–962, 분기 1개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. 소스별 맵을 id 순으로 정렬한다.

같은 스캔을 두 번 돌리면 같은 리포트가 나와야 하고, 두 리포트의 diff는 숫자에 대한
것이어야 한다. Go 맵 순회는 무작위이므로 정렬 없이는 매번 줄 순서가 바뀐다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in` | `map[SourceID]ReadingFacts` | `ScanResult.Readings` | 빈 맵이면 빈 슬라이스 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for id := range in` | `out` append | — | `TestTheJSONReportCarriesBothBlocks`(순서 단언) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `sort.Slice(out, ...)` | id 사전순 | 순수 | ast.json calls |

## State mutations and fallbacks

- 없음 — 새 슬라이스. 입력 맵을 건드리지 않는다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (정렬).

# Function Logic Map: `Store.Summaries`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1691–1759, 분기 10개)
- Risk scan: `risk-pattern-report.md`

모든 후보를 상태·baseline·최초 순위와 함께 읽는다. 이 change가 SELECT에 두 칼럼,
스캔 변수 두 개, `decodeFirstRank` 인자 두 개를 더했다.

이 함수가 `recordFirsts`의 "아직 없는 것"을 판정하는 유일한 입력이라는 것이 중요하다 —
`needRank[s.Symbol] = !s.FirstRank.Recorded()`. 여기서 자격 칼럼을 떨어뜨리면
`recordFirsts`가 자격 판단을 못 하는 것이 아니라(그것은 관측 행에서 온다) 저장된 위치의
자격이 조용히 미상이 되어 모든 후보가 `NEW_ENTRANT_UNKNOWN`이 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `at` | 상태 파생 기준 instant | 호출자 | 쓸기(sweep)가 아니라 읽기 시점 파생 — 재시작 후에도 같은 답 |
| SELECT 18칼럼 | `rows.Scan`의 순서와 일치 | 이 함수 | 불일치는 Scan 오류 |
| `rankNew`/`rankReq` | NULL 또는 유효값 | 칼럼 CHECK | `decodeFirstRank`가 접는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `QueryContext` 오류 | 없음 | 래핑된 오류 | 커버 없음 — I/O |
| B2 | `for rows.Next()` | `out` append | — | `recordFirsts`를 지나는 모든 스캔 테스트 |
| B3 | `rows.Scan` 오류 | 없음 | 오류 | 칼럼 불일치 시 전 스캔 테스트가 실패 |
| B4 | `parseStamp(first)` 실패 | 없음 | 오류 | 커버 없음 — 손상된 행 |
| B5 | `parseStamp(last)` 실패 | 없음 | 오류 | 커버 없음 — 손상된 행 |
| B6 | `cooled.Valid` | `c.CooledAt` 설정 | — | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` |
| B7 | `parseStamp(cooled)` 실패 | 없음 | 오류 | 커버 없음 — 손상된 행 |
| B8 | `decodeBaseline` 오류 | 없음 | 오류 | 커버 없음 — 손상된 행 |
| B9 | `decodeFirstRank` 오류 | 없음 | 오류 | 커버 없음 — 손상된 행 |
| B10 | `rows.Err()` | 없음 | 오류 | 커버 없음 — I/O |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.QueryContext` | 전체 후보 | 오류 래핑 | ast.json calls |
| `stateAt(c, s.cooling, s.staleness, at)` | 읽기 시점 상태 파생 | 순수 | ast.json calls |
| `decodeBaseline` / `decodeFirstRank` | 두 stored first | 파싱 실패는 오류 | ast.json calls |
| `rows.Close` (defer) | 커서 반환 | — | ast.json defers |

## State mutations and fallbacks

- 없음 — 읽기 전용.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: SELECT 칼럼 2개 + 스캔 변수 2개 + `decodeFirstRank` 인자 2개.
- High-risk impact: no (조회 전용). `recordFirsts`의 needRank 판정 입력이므로, `FirstRank.Recorded()`가 잘못되면 write-once 칼럼이 다시 쓰인다 — SQL의 `WHERE first_rank IS NULL`이 그 아래에서 막는다.

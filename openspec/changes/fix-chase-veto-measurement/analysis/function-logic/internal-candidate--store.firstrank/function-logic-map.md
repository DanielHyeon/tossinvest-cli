# Function Logic Map: `Store.FirstRank`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1569–1590, 분기 3개)
- Risk scan: `risk-pattern-report.md`

후보의 저장된 최초 관측 위치를 읽는다. 이 change가 **SELECT에 두 칼럼**과 스캔 변수 두 개를
더했다 — `first_rank_newly_listed`, `first_rank_requested`.

이 두 사실을 관측 테이블에서 조인해 오지 않는 이유가 issues.md I1이다. `Assess`는 최근
`DefaultAssessHistory`(10분)만 읽으므로, `seen_late` 질문이 실제로 대상으로 하는 **오래된**
후보의 최초 관측 행은 그 슬라이스 밖이다. 슬라이스에서 읽으면 답이 하루 종일 `unknown`이 되고
design D3의 거부가 정직한 측정이 아니라 **구조적 측정 불가**가 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market`/`symbol` | 정규화된 키 | 호출자 | 행이 없으면 `(FirstRank{}, false, nil)` |
| SELECT 6칼럼 | `decodeFirstRank`의 인자 순서와 일치 | 이 함수 | 불일치는 Scan 오류 |
| `newly sql.NullString` | NULL / 'yes' / 'no' | 칼럼 CHECK | 그 밖은 `unknown` |
| `requested sql.NullInt64` | NULL 또는 양수 | 칼럼 CHECK | NULL·비양수는 0 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `row.Scan` switch | — | — | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` |
| B2 | `sql.ErrNoRows` | 없음 | `FirstRank{}, false, nil` | `TestARankOfZeroIsNotAFirstSighting`(999999) |
| B3 | 그 밖의 Scan 오류 | 없음 | 래핑된 오류 | 커버 없음 — I/O |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.QueryRowContext` | 한 행 | 오류는 Scan으로 | ast.json calls |
| `decodeFirstRank(rank, total, at, source, newly, requested, ...)` | 6칼럼 → `FirstRank` | stamp 파싱 실패는 오류 | ast.json calls |

## State mutations and fallbacks

- 없음 — 읽기 전용.
- fallback 없음. 없는 행은 `found=false`이고 비어 있는 칼럼은 `Recorded()==false`다.

## Safety conclusion

- Safe edit boundary: SELECT 칼럼 2개 + 스캔 변수 2개. 반환 형태 무변경.
- High-risk impact: no (조회 전용). 두 칼럼을 읽지 않으면 `MeasureFirstSighting`의 두 거부가 항상 발화해 `seen_late`가 영구 미측정이 된다 — 안전 방향이지만 측정이 죽는다.

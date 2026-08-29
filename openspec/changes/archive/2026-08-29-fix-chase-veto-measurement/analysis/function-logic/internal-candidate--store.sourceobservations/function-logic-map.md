# Function Logic Map: `Store.SourceObservations`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L826–869, 분기 6개)
- Risk scan: `risk-pattern-report.md`

읽기 자체는 이 change의 대상이 아니다 — 바뀐 것은 **SELECT 칼럼 목록**이다.
`newly_listed`(schema-1 boolean)를 `newly_listed_state, rank_requested` 둘로 바꿨고,
`scanObservation`의 `Scan` 순서가 그것과 정확히 맞아야 한다(14칼럼).

칼럼 수나 순서가 어긋나면 `Scan`이 오류를 내므로 이 파일의 모든 관측 읽기 테스트가 즉시
실패한다 — 조용히 틀릴 수 있는 형태가 아니다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market`/`symbol`/`source` | 정규화된 키 | `normaliseSource` | 정규화 실패는 오류 |
| `since` | zero면 전부 | 호출자 | zero면 절이 붙지 않는다 |
| SELECT 칼럼 | 14개, `scanObservation`과 일치 | 이 함수 | 불일치는 Scan 오류 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `normaliseSource` 실패 | 없음 | 래핑된 오류 | **커버 없음** |
| B2 | `!since.IsZero()` | WHERE 절 추가 | — | `TestOneSourcesReadingsCanBeReadWithoutTheOthers` |
| B3 | `QueryContext` 오류 | 없음 | 오류 | 커버 없음 — I/O |
| B4 | `for rows.Next()` | append | — | `TestOneSourcesReadingsCanBeReadWithoutTheOthers` |
| B5 | `scanObservation` 오류 | 없음 | 오류 | **커버 없음** |
| B6 | `rows.Err()` | 없음 | 오류 | 커버 없음 — I/O |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.QueryContext` | 행 조회 | 오류는 래핑 | ast.json calls |
| `scanObservation(rows)` | 행 → `Observation` | instant 파싱 실패는 오류 | ast.json calls |
| `rows.Close` (defer) | 커서 반환 | — | ast.json defers |
| `rows.Err` | 반복 중 오류 | 래핑 | ast.json calls |

## State mutations and fallbacks

- 없음 — 읽기 전용.
- fallback 없음. 행이 없으면 nil 슬라이스이고 그것이 '아직 보지 않았다'다.

## Safety conclusion

- Safe edit boundary: SELECT 칼럼 목록 1줄. 본문 로직 무변경.
- High-risk impact: no (조회 전용).

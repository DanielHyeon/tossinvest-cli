# Function Logic Map: `positive`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L2012–2017, 분기 1개)
- Risk scan: `risk-pattern-report.md`

**신규 함수**. 개수의 부재를 SQL NULL로 보낸다.

`nullable`이 아니라 별도 함수인 이유는 두 타입의 부재가 다르기 때문이다. 요청 행 수는 이
저장소에서 **0이 더 작은 수량이 아닌** 유일한 정수다 — "0행을 요청했다"는 아무것도 만들 수
없는 읽기이고, 0을 저장하면 `truncationOf`가 실제 도착 수와 비교해 그런 읽기를 전부
**절단**이라고 부른다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n` | 0 또는 양수(음수는 상위 경계가 거부) | `Reported.RankRequested` / `FirstRank.Requested` | 비양수는 NULL |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `n <= 0` | 없음 | `nil` (SQL NULL) / 아니면 `n` | `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext`(0) · `TestATruncatedReadingReachesTheVerdictAsTruncated`(양수) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls=null |

## State mutations and fallbacks

- 없음 — 순수 함수.
- 칼럼 CHECK(`rank_requested IS NULL OR rank_requested > 0`)와 짝이다. 이 함수가 0을 통과시키면 CHECK가 INSERT를 거부한다 — 두 겹이지 한 겹이 아니다.

## Safety conclusion

- Safe edit boundary: 신규 함수. 기존 동작 없음.
- High-risk impact: no (값 정규화). 음수는 여기서 NULL이 되어 **숨는다** — 그래서 상위 두 경계(`Observation.validate`, `NoteFirstRank`)가 음수를 오류로 거부한다.

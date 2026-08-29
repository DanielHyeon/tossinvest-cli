# Function Logic Map: `decodeFirstRank`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1592–1614, 분기 4개)
- Risk scan: `risk-pattern-report.md`

`candidates`의 여섯 칼럼을 `FirstRank`로 옮긴다. 이 change가 인자 두 개(`newly`, `requested`)와
분기 하나(B2)를 더했다.

`requested`는 **valid이고 양수일 때만** 실린다 — 그 밖은 0이고 0은 `truncationOf`에서
`unknown`이다. `NewlyListed`는 `newlyListedFromStore`가 NULL과 미인식 문자열을 모두
`unknown`으로 접는다. 두 경우 모두 스키마 4 이전에 쓰인 행이 **정확히 그대로** 미상으로
읽힌다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rank`/`total` | 둘 다 valid하고 양수여야 기록으로 친다 | 칼럼 | 아니면 zero `FirstRank` |
| `newly` | NULL / 'yes' / 'no' | 칼럼 CHECK | 그 밖은 `unknown` |
| `requested` | NULL 또는 양수 | 칼럼 CHECK | 그 밖은 0 |
| `at` | NULL 또는 stamp | 칼럼 | 파싱 실패는 오류 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `!rank.Valid || !total.Valid || rank <= 0 || total <= 0` | 없음 | `FirstRank{}, nil` — 기록 없음 | `TestARankFromOutsideTheIdentityWindowIsNotStored` |
| B2 | `requested.Valid && requested.Int64 > 0` | `f.Requested` 설정 | — | `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext`(양수·미기록 양쪽) |
| B3 | `at.Valid` | `f.At` 설정 | — | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` |
| B4 | `parseStamp` 실패 | 없음 | 심볼을 실은 오류 | 커버 없음 — 손상된 행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newlyListedFromStore(newly.String, newly.Valid)` | 3-상태 복원 | 미인식은 `unknown` | ast.json calls |
| `parseStamp(at.String)` | instant | 오류 래핑 | ast.json calls |

## State mutations and fallbacks

- 없음 — 값 하나를 만든다.
- fallback: 미인식 상태 문자열 → `unknown`. 모르는 것을 모른다고 말하는 것이며 값을 지어내지 않는다.

## Safety conclusion

- Safe edit boundary: 인자 2개 + 필드 2개 + 분기 1개 가산.
- High-risk impact: no (디코딩). 두 필드를 잘못 복원하면 write-once 위치가 잘못된 자격으로 측정된다.

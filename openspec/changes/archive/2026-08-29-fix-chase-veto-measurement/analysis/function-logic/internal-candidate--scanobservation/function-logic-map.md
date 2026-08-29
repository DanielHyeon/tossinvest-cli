# Function Logic Map: `scanObservation`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1915–1951, 분기 3개)
- Risk scan: `risk-pattern-report.md`

한 행을 `Observation`으로 옮긴다. 이 change가 **두 칼럼의 해석**을 바꿨다.

`newly_listed`(schema-1 boolean)는 **일부러 읽지 않는다.** 그 칼럼은 지금까지 쓰인 모든
행에서 0이고 — 어떤 생산 소스도 그 필드를 대입하지 않았기 때문이다 — 0은 "아니오"와
"아무도 보지 않았다"를 구분할 수 없다. 읽으면 이 change가 되돌리려는 접기가 그대로
되살아난다. 그래서 `newlyListedFromStore(newly.String, newly.Valid)`가 `newly_listed_state`
를 읽고, NULL과 알 수 없는 문자열은 모두 `unknown`이다.

`rank_requested`는 **valid이고 양수일 때만** 실린다. NULL과 비양수는 0으로 남고, 0은
`truncationOf`에서 `unknown`이 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Scan 순서 | 14칼럼 — 세 SELECT와 정확히 일치 | `Observations`/`SourceObservations`/`ObservationsSince` | 불일치는 Scan 오류 |
| `newly sql.NullString` | NULL / 'yes' / 'no' | `newly_listed_state` 칼럼 + CHECK | 그 밖은 `unknown` |
| `requested sql.NullInt64` | NULL 또는 양수 | `rank_requested` 칼럼 + CHECK | NULL·비양수는 0(미기록) |
| `observed` | 고정폭 stamp | `stamp()` | 파싱 실패는 오류 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `row.Scan` 오류 | 없음 | 래핑된 오류 | 칼럼 불일치 시 전 관측 테스트가 실패 |
| B2 | `parseStamp(observed)` 실패 | 없음 | 심볼을 실은 오류 | `store_test.go`의 unreadable instant 케이스 |
| B3 | `requested.Valid && requested.Int64 > 0` | `o.Reported.RankRequested` 설정 | — | `TestATruncatedReadingReachesTheVerdictAsTruncated` · `TestAWholeReadingOfTheSameLengthIsMeasured` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `row.Scan` | 14칼럼 | 오류 래핑 | ast.json calls |
| `parseStamp` | instant | 오류 래핑 | ast.json calls |
| `newlyListedFromStore(newly.String, newly.Valid)` | 3-상태 복원 | 미인식 값은 `unknown` | ast.json calls |

## State mutations and fallbacks

- 없음 — 값 하나를 만들어 돌려준다.
- fallback: 인식하지 못한 `newly_listed_state` 문자열은 `unknown`이다. 이것은 미측정을 값으로 접는 것이 아니라 **모르는 것을 모른다고 말하는 것**이며, `Chase.State`가 미인식 veto code에 적용하는 규칙과 같다.

## Safety conclusion

- Safe edit boundary: `newly`의 타입이 `int`에서 `sql.NullString`으로, `requested` 스캔 3줄 가산.
- High-risk impact: no (디코딩). `newly_listed`를 다시 읽으면 이 change 전체가 무효가 된다 — 그 칼럼은 전 행 0이다.

# Function Logic Map: `positionRow.HasDetail`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

change refresh-positions-screen에서 신설된 무상태 leaf 함수다(diff 교차로 evidence 요구). 행의 둘째 줄(tr) 렌더 여부 — exit 라인·원장 정보·행별 사유 중 하나라도 있을 때만.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.HasExit`·`r.InJournal`·`r.Reason()` | bool·bool·string | joinPositions/Reason | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기 — 단락 평가) | `HasExit \|\| InJournal \|\| Reason() != ""` | 없음 | bool | `TestTheJournalAbsenceNoticeAppearsOncePerPage`(전역 사유 행의 둘째 줄 부재), 기존 exit 라인·원장 정보 케이스 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.Reason()` | 행별 사유 유무 판정 | 순수 | ast.json calls + HEAD |

## State mutations and fallbacks

- 없음(순수 판독).

## Safety conclusion

- Safe edit boundary: 신규 leaf — 렌더 게이트만
- High-risk impact: no (콘솔 read-only 렌더 경로 — 주문·위험·원장 코드 무접촉)

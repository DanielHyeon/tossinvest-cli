# Function Logic Map: `positionsView.AnyJournalAbsent`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

change refresh-positions-screen에서 신설된 무상태 leaf 함수다(diff 교차로 evidence 요구). 원장 부재 보유(수동 매수)의 페이지 공지 1회 렌더 여부를 결정한다(design D1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `v.Rows` | []positionRow | `joinPositions` 산출 | 해당 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 range | `v.Rows` 순회 | 없음 | — | `TestTheJournalAbsenceNoticeAppearsOncePerPage` |
| B2 if | `r.JournalReadable && !r.InJournal` early return true | 없음 | `true`/순회 종료 후 `false` | 같은 테스트 (공지 정확히 1회) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (필드 판독만) | — | — | ast.json + HEAD |

## State mutations and fallbacks

- 없음(순수 판독). 미판독 상태에서는 JournalReadable=false라 AnyUnknown과 상호 배타다.

## Safety conclusion

- Safe edit boundary: 신규 leaf — 라벨·자격 판정 무접촉
- High-risk impact: no (콘솔 read-only 렌더 경로 — 주문·위험·원장 코드 무접촉)

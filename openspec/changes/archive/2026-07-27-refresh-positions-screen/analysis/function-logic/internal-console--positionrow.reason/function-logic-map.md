# Function Logic Map: `positionRow.Reason`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json` (구현 후 재추출 — B1 switch + case 4개, return 4개)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.JournalReadable`(`Unknown()` 경유) | bool | `joinPositions` 인자 ← `positions()`의 `v.Journal.Readable()` | 해당 없음 — 순수 판독 |
| `r.InJournal` | bool | `joinPositions`의 journal 측 병합 | 해당 없음 |
| `r.Eligible` | bool | `journal.Position.ExitEligible()`(자격 술어 단일화) | 해당 없음 |
| `r.HasExit` | bool | `journal.PositionExit.HasExit` | 해당 없음 |
| `r.Basis()` | "편입 기록"/"진입 결정"/"—" | 같은 파일의 순수 함수 | 해당 없음 |

불변식: 렌더 전용 순수 함수 — mutation·I/O·live binding 없음. 반환 ""는 "이 행에 사유
문장이 없다"를 뜻하며, 페이지 전역 사유는 `positionsView.AnyUnknown`/`AnyJournalAbsent`가
공지 1회로 소유한다(change refresh-positions-screen design D1).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 switch | — | 없음 | — | — |
| B2 case | `r.Unknown() \|\| !r.InJournal` (전역 스코프 사유) | 없음 | `""` — 행 문장 제거, 페이지 공지가 대체 | `TestTheUnknownStateNoticeAppearsOncePerPage`, `TestTheJournalAbsenceNoticeAppearsOncePerPage` |
| B3 case | `!r.Eligible` (자격 기록 없는 원장 포지션) | 없음 | 고정 문장(진입 결정·편입 기록 모두 없음) | `TestAnUnmanagedHoldingIsLabelledExactlyOnce` |
| B4 case | `!r.HasExit` (자격 있음·exit 미개설) | 없음 | `Basis()` 포함 문장 | 자격 표시 기존 케이스 + `TestTheJournalAbsenceNoticeAppearsOncePerPage`의 행별 사유 유지 단언 |
| B5 default | exit 라인 존재 | 없음 | `""` (exit 라인이 렌더) | `TestThePositionsScreenShowsTheExitLineOfAManagedPosition` |

변경 전 대비: 구 Unknown 분기(문장 반환)와 구 `!InJournal` 분기(문장 반환)가 B2의 `""`
반환으로 통합되고 문장은 템플릿 공지로 이동했다. 나머지 분기 순서·내용은 동일하다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.Unknown()` | 전역 사유 판정 | 순수, 오류 없음 | ast.json calls + HEAD |
| `r.Basis()` | B4 문장의 자격 근거 표기 | 순수, 오류 없음 | ast.json calls + HEAD |

## State mutations and fallbacks

- 없음(순수 함수). fallback 없음 — 빈 문자열이 유일한 "없음" 표현이고, 소비처
  (`HasDetail`·템플릿 `{{with .Reason}}`… 실제로는 `{{if .HasDetail}}` 가드와 기존
  `{{.Reason}}` 렌더)가 빈 값을 렌더 생략으로 처리한다.

## Safety conclusion

- Safe edit boundary: 표시 문자열과 분기 재배치만 — journal 판독·자격 판정 로직 무접촉
- High-risk impact: no (콘솔 read-only 렌더 경로; 주문·위험·원장 코드와 무관)

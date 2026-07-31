# Function Logic Map: `positionRow.Label`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change context: 사용자 UX 결정 2026-07-27 — 미편입 행의 판정 라벨이 편입 지정
  체크박스 상태를 반영한다(포지션 가시성 delta). 이번 수정은 B3 한 분기 추가다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.JournalReadable`/`r.InJournal` (`Unknown()`) | bool | openJournal 판독 결과 | 원장 미판독이면 지정 여부와 무관하게 "관리 여부 불명" — 관측하지 않은 보호를 단정하지 않는다 |
| `r.InJournal && r.Eligible` (`Managed()`) | bool | journal.Position.ExitEligible (자격 정의는 원장이 소유) | 자격 없으면 미편입 라벨 계열로 떨어진다 |
| `r.Designated` | bool | 설정 seam의 include 목록 — handlePositions가 표시 전용으로 stamp | seam 미배선이면 항상 false = 기존(관리 외) 라벨 그대로 (§0.3 zero-value) |
| `r.HasExit`, `r.Exit.Completed` | bool | journal exit_states | exit 유무·완결로 관리 라벨 3종 분기 |

불변식: 라벨 문자열의 정의는 이 함수뿐이다(TestTheUnmanagedLabelIsSpelledOnce).
Designated는 표시 전용이며 판정(Eligible)을 승격시키지 않는다 — Managed 분기보다
뒤에서만 읽힌다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | switch 진입 | 없음(순수 함수) | — | 전 분기 공통 |
| B2 | `r.Unknown()` | 없음 | "관리 여부 불명" | TestAHoldingIsNotCalledUnmanagedWhenTheJournalCouldNotBeRead |
| B3 | `!r.Managed() && r.Designated` (신규) | 없음 | "관리 편입" | TestAnUnmanagedRowsLabelFollowsItsCheckbox |
| B4 | `!r.Managed()` | 없음 | "관리 외(미편입)" | TestAnUnmanagedHoldingIsLabelledExactlyOnce |
| B5 | `r.HasExit && r.Exit.Completed` | 없음 | "관리 종료" | 직접 pin 없음 — 기존 분기, 이번 변경 무접촉(잔여: review.md P4) |
| B6 | `r.HasExit` | 없음 | "엔진 관리" | TestAnAdoptedHoldingRendersAsManagedWithItsBasis |
| B7 | default(자격 있으나 exit 미개설) | 없음 | "엔진 관리(대기)" | 직접 pin 없음 — 기존 분기, 이번 변경 무접촉(잔여: review.md P4) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.Unknown()` | 원장 미판독 구분 | 순수 bool, 오류 없음 | ast.json calls[0] |
| `r.Managed()` (×2) | exit 정책 소유 여부 | 순수 bool, 오류 없음 | ast.json calls[1..2] |

live binding 없음 — 렌더 전용 순수 함수이며 config·브로커·원장에 닿지 않는다.

## State mutations and fallbacks

- 없음. 수신자 값 복사에 대한 읽기만 있고 mutation·side effect·fallback이 없다.

## Safety conclusion

- Safe edit boundary: 표시 라벨 선택만 바꾼다. 판정(Eligible)·편입 실행·주문
  경로와 무관하고, B3는 B2(불명) 뒤·B4(관리 외) 앞에만 삽입되어 원장 미판독
  보수성이 유지된다.
- High-risk impact: no — 콘솔 표시 전용, 주문·손절·사이징·Guardian 무접촉.

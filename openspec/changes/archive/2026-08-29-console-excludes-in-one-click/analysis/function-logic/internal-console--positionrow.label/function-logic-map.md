# Function Logic Map: `positionRow.Label`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `Unknown()` | bool — 원장 미판독 | `JournalReadable`·`InJournal` | 판정 최우선. 관측하지 않은 것을 단정하지 않는다 |
| `Excluded` | bool — `exclude_symbols` 등재 | `handlePositions`가 seam에서 스탬프 | zero value = 이 change 이전 동작 |
| `Designated` | bool — `include_symbols` 등재 | 같은 스탬프 | zero value = 변경 전 동작 |
| `Managed()` | bool — exit 자격 | `InJournal && Eligible` | 관리 중이면 두 목록을 보지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` 진입 | 없음(순수 함수) | 아래 case 중 하나 | 전 라벨 테스트 |
| B2 | `Unknown()` | 없음 | `"관리 여부 불명"` | `TestAnUnreadableJournalStaysUnknownEvenWhenExcluded` |
| B3 | `!Managed() && Excluded` — **이 change가 추가** | 없음 | `"관리 제외"` | `TestAnExcludedRowIsLabelledAndOffersRelease` |
| B4 | `!Managed() && Designated` | 없음 | `"관리 편입"` | `TestAnUnmanagedRowsLabelFollowsItsCheckbox` |
| B5 | `!Managed()` | 없음 | `"관리 외(미편입)"` | `TestAnUnmanagedRowsLabelFollowsItsCheckbox` |
| B6 | `HasExit && Exit.Completed` | 없음 | `"관리 종료"` | 기존 포지션 화면 테스트 |
| B7 | `HasExit` | 없음 | `"엔진 관리"` | 기존 포지션 화면 테스트 |
| B8 | default | 없음 | `"엔진 관리(대기)"` | 기존 포지션 화면 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionRow.Unknown` | 원장 판독 여부 | 없음 | CodeGraph + AST |
| `positionRow.Managed` | exit 자격 | 없음 | CodeGraph + AST |

## State mutations and fallbacks

- 상태 변경 없음 — 값 수신자의 순수 함수다.
- B3이 B4보다 **앞**인 것이 계약이다: 엔진(`adoption.go`)이 제외를 편입보다 먼저 판정하므로, 순서가 뒤집히면 라벨이 엔진의 행동을 잘못 예고한다.
- fallback: 세 불리언이 전부 false면 기존 세 갈래(B6·B7·B8)가 변경 전과 동일하게 답한다.

## Safety conclusion

- Safe edit boundary: case 하나 추가 + 순서 고정. 라벨 철자는 이 함수가 단독 정의하며 템플릿에 두 번째 철자를 두지 않는다.
- High-risk impact: no — 표시 전용이다. 다만 잘못된 라벨은 운영자가 보호 상태를 오해하게 만들므로 우선순위를 엔진과 일치시켰다.

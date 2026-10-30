# Function Logic Map: `ExitObserver.sellIntent`

- Source: `internal/app/engine/exitloop.go:1805-1831`
- AST evidence: `ast.json` (Pre-Edit HEAD `a284a2b0a`, 구현 뒤 재생성)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| 입력/상태 | 정본 | 실패 동작 |
|---|---|---|
| `observed` 인자 | `submit`의 단일 `exitLimitPrice` 선택값 | 빈 값이면 baseline으로 폴백 |
| baseline | `exit_states` | 둘 다 비면 주문 의도 거부 |
| quantity | `applyFloor` 뒤 수량 | 파싱 실패 거부 |
| market | 원장 position | 통화 모드만 결정 |

## Branches and early returns

| 분기 | 조건 | 효과 | 시험 |
|---|---|---|---|
| B1 | 선택값이 빈 문자열 | baseline으로 대체 | a087 empty fallback |
| B2 | baseline도 빈 문자열 | 주문 의도 오류 | 기존 직접 시험 |
| B3 | 수량 파싱 실패 | 주문 의도 오류 | 기존 직접 시험 |
| B4 | 가격 파싱 실패 | 주문 의도 오류 | 기존 직접 시험 |
| 정상 | — | SELL LIMIT 의도, 선택값을 Price로 전달 | a087 PlaceIntent/wire golden |

## Calls and live bindings

- broker·원장·설정 호출 없음.
- a087은 본문을 바꾸지 않고 `submit` 인자만 선택값으로 바꿈.
- 주문 유형과 수량은 기존대로이며, 가격이 빈 경우의 기존 거부 경로도 남음.

## State mutations and fallbacks

- 상태 변경 없음. 선택값이 비었을 때만 기존 baseline 폴백을 수행함.

## Safety conclusion

- `submit`의 단일 선택값만 받아 기존 LIMIT 생성·거부 경로를 보존함.

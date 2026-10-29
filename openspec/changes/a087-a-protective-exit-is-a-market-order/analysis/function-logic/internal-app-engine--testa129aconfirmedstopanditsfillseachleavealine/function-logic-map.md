# Function Logic Map: `TestA129AConfirmedStopAndItsFillsEachLeaveALine`

- Source: `internal/app/engine/a129_exit_order_log_test.go:217-303`
- Pre-Edit HEAD: `20ef25ef5e64fa5ff179c62fbeed7c3d5f176052`
- 대상: 휴면 admission에서 청산 주문·로그 가격의 골든을 고정하는 기존 시험

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 하한가 admission | nil | 생산 `NewContext` | 관측가 LIMIT 유지 |
| 제출가와 로그가 | 같은 decimal | `submit`의 단일 choice | 불일치 시 실패함 |
| 가격 출처 | observed | 휴면 판 | lower_limit이면 실패함 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 손절 확정 | 주문·확정 줄 생성 | 계속 | 이 함수 |
| B2 | 부분·전량 체결 | 체결 줄 생성 | 계속 | 기존 a129 행 |
| B3 | 동일 누적 체결 | 새 줄 없음 | 계속 | 이 함수 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `a129Setup` | 실제 Guardian·journal 하네스 조립 | setup 실패 시 `Fatal` | `ast.json` |
| `breach` | 보호 청산 제출 | 기존 LIMIT 제출 | B1 |
| `a129Want` | JSON golden 필드 비교 | 불일치 시 오류 | B1-B3 |

## State mutations and fallbacks

- 시험 하네스만 임시 journal에 intent·fill을 기록함.
- admission nil에서 `lower_limit`은 금지하고 관측가/기준선 출처만 허용함.

## Safety conclusion

- admission nil 상태의 `PlaceIntent` 가격·LIMIT 유형·주문 순서는 기존 골든과 같아야 함.
- 이 change가 허용하는 로그 변화는 `price_source=observed|baseline` 키 추가뿐임.

# Function Logic Map: `exitOrderLine.args`

- Source: `internal/app/engine/exit_order_log.go:138-168`
- AST evidence: `ast.json` (Pre-Edit HEAD `e1eaf720c`, source_sha256 `419352e9…` = 파일 해시 — a087 의 `exitorderline.args`·`exitorderlinefor` 번들과 같은 값, 분기 1)
- Risk scan: `risk-pattern-report.md` (적중 0)
- 판: **pre-edit**. 편집 범위(design D4, Manager 결정 2026-11-23): 줄이 취소일 때만 종류·주문 방향을 내보냄. 취소 아닌 줄의 출력은 바이트 동일.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `l`(exitOrderLine) | 화이트리스트 필드만(계좌 필드 없음) | 만드는 자리(exitOrderLineFor·placed·filled·because) | — |
| 불변식 | 줄에는 구조체 필드만 나감; 상태·사유는 `exitOrderToken` 거름(영문자·밑줄) | a129 design D3 | 거름 실패 → `unrecognised` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L160 | `l.fill` | 체결 4 필드 덧붙임 | args | 기존 a129 체결 줄 시험 |
| (편집 예정) | 줄 종류가 취소 | `side` = 주문 방향, 종류 = 취소, 수량·가격 빈 값 | args | a198 §4-4·4-5, 변이 「취소 줄이 side SELL」 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exitOrderToken` | 상태·사유 거름 | 순수 | AST calls |
| live config | 없음 | — | — |

## State mutations and fallbacks

- 없음(값 수신자, 슬라이스만 만듦). 고정 리터럴 `side "SELL"`(L143)·`order_type "LIMIT"`(L144)가 지금은 모든 줄에 붙음.

## Safety conclusion

- Safe edit boundary: L143-144 의 두 리터럴을 「취소 줄이면 다른 값」으로 가르는 것뿐. 키 순서·다른 필드 무편집.
- High-risk impact: no(관측 줄 — 판정·주문 무영향). 단 계좌 비노출 불변식이 걸린 출력 함수이므로 새 필드는 화이트리스트 안에서만.

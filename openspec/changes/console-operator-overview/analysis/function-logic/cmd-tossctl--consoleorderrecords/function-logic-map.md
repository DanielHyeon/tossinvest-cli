# Function Logic Map: `consoleOrderRecords`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (revision=current, L594–606, 분기 1개)
- Risk scan: `risk-pattern-report.md`
- 이 change의 base: `47672c6f` — base에는 이 함수가 없었다. 현재 본문을 diff hunk가 덮으므로 evidence가 요구된다 (revision=current)

raw 주문 한 페이지를 콘솔 경계 너머로 옮긴다. **값 변환이 전혀 없다는 것**이 이 함수의 전부다.

- **넘어가는 것**: `[]console.OrderRecord` — 브로커가 보낸 문자열 그대로.
- **넘어가지 않는 것**: `official.RawOrder` 타입, 그리고 파싱된 숫자. 이 경로 어디에서든 `parseDecimal`을 부르면 raw 읽기가 막으려던 0이 돌아온다: 시장가 주문의 `price`는 null이고 살아있는 주문의 `execution` 전체가 null이므로, 변환된 판독은 **모든 미체결 주문이 0원에 체결됐다**고 말한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `orders []official.RawOrder` | nil 허용 | `Client.OrdersRaw`의 원문 디코딩 | nil이면 길이 0 슬라이스 — nil이 아니다(템플릿이 구분하지 않도록) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `range orders` | `out` append | `[]console.OrderRecord` (항상 non-nil) | `TestTheOrdersSeamCarriesEachListsOutcomeSeparately`의 null price/execution 검사 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `make` / `append` / `len` | 정확한 용량으로 한 번에 할당 | — | ast.json calls |

## State mutations and fallbacks

- 순수 매핑. 부재(`null`)는 빈 문자열로 도착하고 빈 문자열로 나간다 — 0이 되지 않는다.

## Safety conclusion

- Safe edit boundary: 필드 대응 관계. 숫자 변환을 넣는 것은 이 함수의 존재 이유를 지우는 편집이다.
- High-risk impact: yes (주문 경로 — 데이터 정합) — 주문을 내지는 않지만, 여기서 문자열을 숫자로 바꾸면 살아있는 주문의 체결가가 0으로 렌더되어 운영자가 잔여물을 오독한다.

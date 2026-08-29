# Function Logic Map: `TestTheRawOrderReadKeepsAnAbsentValueApartFromAZeroOne`

- Source: `internal/official/orders_reads_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

**신규 테스트**다(HEAD L511-546). 화면의 주장 전체 — "브로커가 보내지 않은 것은 0이 아니다" —
를 `OrdersRaw`에 대해 잰다. API는 시장가 주문의 `price`를 null로, 살아 있는 주문의
`execution` 객체 **전체**를 null로 보낸다. 그래서 모든 라이브 주문이 평균 체결가 "없음"으로
도착하고, 그것을 0으로 렌더하면 **체결됐다**는 말이 된다.

기대표가 부재와 진짜 0을 나란히 놓는다: 미체결 행의 `price`/`filledQuantity`/
`averageFilledPrice`는 `""`, 체결 행의 같은 필드는 `"0"`/`"5"`/`"0"`이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 서버 | `ordersRawServer(t, false, "")` | 같은 파일 헬퍼 | `defer srv.Close()` |
| 호출 | `OrdersRaw(ctx, OrdersFilter{Status: "OPEN"})` | 이 테스트 | 그룹 미지정이면 거부(별도 테스트) |
| 기대 | 부재 → `""`, 측정된 0 → `"0"`, 상태는 브로커 어휘 그대로 | 이 테스트 | `t.Errorf` |

불변식: `status`는 수정하지 않는다 — 상태를 유도하는 것은 `internal/brokerstate`의 일이고
그쪽은 브로커의 단어(이 빌드가 본 적 없는 값 포함)를 필요로 한다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if, L516) | 읽기 오류 | 없음 | `t.Fatal` | 자체 실행 |
| B2 (if, L519) | 건수 ≠ 2 | 없음 | `t.Fatalf` | 자체 실행 |
| B3 (range, L524) | 필드 기대표 7건 순회 | 없음 | — | 자체 실행 |
| B4 (if, L533) | 필드 불일치 | 없음 | `t.Errorf` — 부재가 숫자로 도착하면 화면에서 측정된 0과 구분할 수 없다 | 자체 실행 |
| B5 (if, L538) | `status`가 `PENDING`이 아님(가공됐다) | 없음 | `t.Errorf` | 자체 실행 |
| B6 (if, L542) | 주문 id 불일치 | 없음 | `t.Errorf` — 주문번호는 화면의 열이다 | 자체 실행 |

## State mutations and fallbacks

- `httptest` 서버와 `t.TempDir()`의 토큰 캐시만 쓴다. 실계좌·실브로커·네트워크 외부 접촉 0.
- 주문을 내지 않는다(GET 경로만). LIVE side effect 없음.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ordersRawServer` / `ordersRawClient` | 픽스처 | `defer srv.Close()` | ast.json calls/defers |
| `Client.OrdersRaw` | 측정 대상 | 오류 그대로 단언 | ast.json calls |

## Safety conclusion

- Safe edit boundary: 신규 테스트 가산.
- High-risk impact: **no** (테스트 전용, 실계좌 무접촉) — 재는 대상은 High-risk다.
  이 테스트가 약해지면 라이브 주문이 "0에 전량 체결"로 렌더되고, 그것은 운영자가
  잔여물이 없다고 결론 내리는 화면이다.

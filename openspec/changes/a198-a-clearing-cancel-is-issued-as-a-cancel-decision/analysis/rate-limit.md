# a198 rate limit 계상 (tasks 1.5)

retry matrix 정본: `openspec/changes/archive/2026-07-26-harden-execution-base/retry-matrix.md`(HEAD `e1eaf720c`).

| 이 수리로 실제로 나가기 시작하는 호출 | matrix 행 | 계상 |
|---|---|---|
| `POST /api/v1/orders/{id}/cancel` — 엔진 귀속 미체결 주문당 1 | **31** 「엔진 청산 전 취소 · 409 already-processing · 0회 · FAILED_CONFIRMED; 같은 주기 재전송·대기 없음」, **33** 「그 외 취소 · 전부 · 0회」, §0.1(엔진은 재전송 없이 FAILED_CONFIRMED) · 103줄 `EngineCancelRefusal()` | **이미 계상** — 길 치우기의 취소를 명시한 행이 있다. 코드도 일치: 엔진 게이트웨이 `CancelRefusal: execgw.EngineCancelRefusal()`(internal/app/engine/gateway.go:329) → withCancelRefusal B3/B4 |
| `GET /api/v1/orders/{id}` — 취소마다 앞서는 사전 조회(trading.Service.Cancel B3 → broker.go:57 → precheck.go:102 `OrderRawByID`, 2s 시한, 오류여도 진행) | **37** 「단건 주문 상태 · transient · 최대 2회 재시도」(일반 행) | 일반 행으로만 계상 — 길 치우기 전용 표기 없음. 재시도 여부는 Retrier 가 아니라 official.Client 직접 호출이라 행 37 의 「최대 2회」가 이 자리에 적용되는지 미확인 |
| 재취소(위 claim-check §2-1·(c): 접수된 매수 취소 뒤 종결 기록 전 다음 주기) | 31/33 | 주기당 주문당 POST 1 + GET 1 상한(같은 주기 재전송 0) — 빈도는 관측 주기(Interval)가 정함 |

판정: 새 endpoint·새 호출 종류 없음(proposal Impact 와 일치). 정본 exit-policy 「이 요구는 손절 경로에 브로커 조회를 더하지 않는다(대상 목록은 원장에서 온다)」 — 대상 목록은 원장(LiveOrdersForSymbol)이 맞지만, **취소마다 사전 GET 1 이 따라 나간다**. 그 조회가 이 문장 범위 안인지는 정본 해석 문제 — Manager 판단 대기(문서 미편집).

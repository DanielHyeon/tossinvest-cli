# retry matrix · 폴링 SLO 수치 — task 1.3 (확정)

> **확정 — 사용자 결정 2026-10-22.** 같은 날 초안(`retry-matrix-slo-draft-2026-10-22.md`, 이 파일의 이전 이름)을
> 사용자가 세 곳을 고쳐 승인했다(§2 의 ✱ 행). 나머지 행은 초안 그대로 승인됐다. 결정 기록은
> `issues.md` 「사용자 결정 (2026-10-22, 1.3 확정)」 절이다.
>
> **이 change 는 코드를 바꾸지 않는다.** 코드 기본값(`execgw.DefaultRetryPolicy()`·`DefaultStaleness()`,
> `filldetect.DefaultConfig()`·`DefaultSLO()`)과 정본 표(`openspec/changes/archive/2026-07-26-harden-execution-base/retry-matrix.md`)는
> 그대로다. 바뀌어야 하는 행은 후속 change 두 개가 가져간다 — 정본 표 머리말 규칙대로 **표와 코드 기본값은 그 change 의
> 같은 커밋에서** 함께 바꾼다.
>
> - `a165-safe-order-placement-retry` — §2 의 주문 접수 재시도 ✱, 취소·정정 `409 already-processing` 재시도 ✱
> - `a166-account-shared-rate-limit` — §2 의 계좌 공유 요청 속도 상한 ✱, GET 429 처리(엔진 폴링 경로)
>
> 「유지」 행은 바꿀 코드가 없다.

## 1. 입력 — 이미 있는 측정만

| # | 값 | 출처 |
|---|---|---|
| R1 | 300ms 안 3회 시도가 전부 429 — 짧은 간격 재시도는 회복하지 못함 | M4 |
| R2 | 간격 없는 페이지 연발 7요청/535ms(≈13 req/s) 뒤 429, 직후 단건 1요청도 429(패널티 창) | M8 |
| R3 | 15s 백오프(추가 2회 15s/30s)가 429 를 흡수하고 walk 를 완주시킴 | M10, `internal/verifylive/retry.go` `ReadRetryBackoff` |
| R4 | attestation `rate_limit_per_second` = 3.5636(관측된 지속 속도, 한도 아님) | `~/.config/tossctl/capability-attestation.json` 의 해당 필드만 읽음(2026-10-22, `issued_at 2026-10-21T21:18:31Z`) |
| R5 | 최근 3일 soak 272 사이클(2026-10-19T01:12Z–10-22T01:02Z): 9개 endpoint 최종 class 전부 `ok`, `rate_limited` 0 | `~/.config/tossctl/capability-soak.jsonl` 집계(로컬 기록만, 계좌 API 호출 없음) |
| R6 | 같은 3일, `GET /api/v1/orders` walk latency p50 62.4s · max 78.3s, 272 중 267 사이클이 ≥15s, `requests − order_pages` 최빈값 4 — **최종 class 는 ok 지만 walk 는 거의 매번 429 를 만나 백오프로 흡수한다**(latency 에 백오프 대기 포함 — issues.md 1.11 잔여 위험 ②). `orders/{id}` 도 57/272 사이클이 ≥15s | R5 와 같은 집계 |
| R7 | 단건 GET latency 3일 p99: accounts 569ms(max 4,171ms), conditional-orders 539ms, holdings 282ms, prices 252ms, buying-power 241ms, sellable 233ms | R5 와 같은 집계 |
| R8 | 주문 왕복 최대 169ms(US)·182ms(KR), KR MARKET 매도 약 149ms | M14, M35, M50 |
| R9 | 취소·정정 `409 already-processing`, `retryAfterSeconds: 1`, 1회 재시도로 수락 | M16, M24, M25, M27 |
| R10 | 계정 CLOSED 이력 ≥2,500건, `limit` 최대 100 → 전체 walk ≈ 26페이지 | M10 |
| R11 | 같은 `clientOrderId`+같은 본문 재요청은 같은 `orderId` 를 돌려주고 두 번째 주문을 만들지 않음(open delta 1), 본문이 다르면 `422 idempotency-key-conflict`. 유효 창(문서상 10분)의 경계는 **측정 거부로 미검증**(2.7 사용자 결정) | M14, M35 |

`0 throttled` 는 R5(최종 class)의 뜻이다. R6 이 보여 주듯 요청 단위로는 429 가 계속 난다 — 두 문장을 섞지 않는다.

## 2. 확정 수치

✱ = 사용자가 초안을 고친 행. 나머지는 초안 그대로 승인.

| 항목 | 현재 provisional(코드·정본 표) | 확정 | 근거 | 반영 |
|---|---|---|---|---|
| ✱ 주문 접수(place) 자동 재시도 | 0회 | **최대 2회, 안전 방식** — (a) 브로커가 **확실히 접수하지 않은** 경우만 즉시 재시도: HTTP 429, 또는 요청을 보내기 **전**의 연결 실패. 같은 `clientOrderId`, 최초 전송부터 8s 예산 안, 최대 2회. (b) 결과 불명(타임아웃·5xx·전송 **후** 연결 끊김) → 맹목 재전송 금지: 먼저 `clientOrderId`/미체결 목록으로 접수 여부를 조회하고 **없을 때만** 재전송. (c) 확정 거절(4xx 검증 오류) → 재시도 없음 | 사용자 결정 1, R11, R8 | a165 |
| ✱ 취소·정정 `409 already-processing` | 0회(mutation 전부) | **허용** — `retryAfterSeconds` 를 따르되 총 대기 5s 이내, 최대 2회 재시도. HTTP 상태와 본문 code 가 모두 일치할 때만(정본 요구 「취소의 일시적 거절 처리」와 같은 판정) | 사용자 결정 2, R9, `verifylive.TransientRetryMaxWait`=5s | a165 |
| ✱ 요청 속도 상한 | 없음 | **3 req/s, 계좌 전체 공유** — 엔진·soak·verify·프로브·CLI 등 같은 계좌를 치는 모든 프로세스의 합. 프로세스 간 조정이 필요하므로 구현 change 가 따로 있다 | 사용자 결정 3, R2 의 13 req/s 에서 429, R4 의 3.56 req/s 는 지속 관측 — 그 아래로 둠 | a166 |
| GET 429 처리(엔진 폴링 경로) | `min(Retry-After, 30s)` 1회 대기 | **같은 주기 안 즉시 재시도 금지**. `Retry-After` 가 있으면 ≤30s 일 때만 존중, 없으면 이번 조회 포기 → staleness 규칙에 맡김 | R1(짧은 재시도 무효), `Retry-After` 헤더 존재는 [미측정] | a166 |
| GET 429 처리(soak·verify·IN_DOUBT 스캔 같은 비지연 walk) | — | 추가 2회, 15s → 30s(현행 `ReadRetryBackoff` 그대로) | R3, R6 | 유지(코드 변경 없음) |
| GET transient(5xx·전송) | 3회·400ms→3s·총 8s | **유지** | R7 p99 ≤ 0.6s 라 8s 예산 안에서 3회 가능 | 유지 |
| 체결 감지 폴링 주기 | 3s | **유지**(OPEN 목록 1요청 ≈ 0.33 req/s) | OPEN 은 cursor/limit 무시·전량 1요청(M7) | 유지 |
| 체결 감지 SLO | target 10s · p95 · 5m 창 · 최소 3표본 · grace 30s | **유지** | 3s 주기 + 단건 p99 0.6s 면 정상 최악 ≈ 4s, 10s 는 두 배 여유 | 유지 |
| staleness 임계 | open_orders 20s · buying_power 45s · holdings 60s · price 15s | **유지**. 단 open_orders 20s 는 429 패널티 창 1회(≤15s 로 관측 — R3 에서 15s 대기 후 회복)를 흡수할 여유일 뿐이다 | R3 은 상한만 준다, 창 크기 자체는 [미측정] | 유지 |
| IN_DOUBT 해소의 CLOSED walk | OPEN+CLOSED 각 1페이지 × 안정화 3회 | **1페이지(최신 100건) 유지, 전체 walk 금지**. 전체 walk 가 필요하면 3 req/s 로 ≈9s + 429 시 15s 백오프 | R6, R10 | 유지 |

### 2.1 주문 접수 재시도와 task 2.7(재생 비활성)의 관계

- **2.7 결정은 그대로다.** IN_DOUBT 해소의 재생 경로(`execgw.Gateway.ReplayInDoubt`, 유일한 재전송 자리 `internal/execgw/replay.go:323`)는
  생산 배선의 `Attested: nil` 로 꺼진 채 남는다(issues.md 2026-10-22 결정 1 확인 결과). a165 는 그 스위치를 켜지 않는다.
- a165 의 재전송은 **최초 전송 직후의 짧은 창 안에서만** 같은 `clientOrderId` 로 일어나는 별개 경로다. 안전 근거는 R11 —
  같은 키·같은 본문 재요청이 두 번째 주문을 만들지 않았다(M14·M35). 그 측정은 첫 응답 직후의 재요청이었고, 창 경계(문서상 10분)는
  측정하지 않았다.
- **8s 예산은 측정된 안전 범위 깊숙이 안에 있다**: 문서상 유효 창 10분(600s)의 1/75, 보수 마진 TTL/2(300s)의 1/37.5,
  현행 `DefaultReplayMargin` 60s(`replay.go:78`)의 1/7.5. 창 경계가 미검증이어도 8s 는 그 경계에 닿지 않는다.
- 남는 위험(측정이 없어서): 키의 계좌 스코프(단일 계좌라 원리적 미검증, M14·M35), 브로커가 429 를 돌려준 POST 를 실제로 접수하지 않았다는
  것은 **사용자 결정의 전제**이지 측정이 아니다 — a165 design 은 이 전제를 이름 붙여 적고, 같은 키 재전송이 그 전제가 틀렸을 때도
  두 번째 주문을 만들지 않는다(R11)는 것을 안전망으로 쓴다.
- (b) 의 「조회 후 없으면 재전송」을 8s 창 밖에서도 할지는 이 결정이 정하지 않았다. a165 design 의 사람 확인 항목으로 넘긴다.

## 3. 측정이 없는 것 — 확정 뒤에도 남는다

- 429 의 실제 한도 값·창 크기·집계 단위(계좌/키/IP), `Retry-After` 헤더 존재 여부(M4·M8 모두 [미측정]).
- 조회 실패가 연속될 때의 패널티 연장 여부.
- 여러 프로세스(엔진·soak·console·CLI)가 같은 계좌로 동시에 칠 때의 합산 한도. 사용자 결정 3 이 상한을 **계좌 공유**로 정했으므로
  이 질문은 이제 a166 의 입력이다. soak 사이클(약 16분 간격, R5)이 엔진 폴링과 겹치는 순간의 합산은 재지 않았다.
- 429 를 받은 주문 POST 가 정말 미접수인지(§2.1).

## 4. 결정 이력

1. 2026-10-22 오전 — 초안 작성(Manager 권고 5, 「권장대로」). 기존 측정만 사용.
2. 2026-10-22 — 사용자 결정: 주문 접수 재시도 「안전한 방식」 최대 2회 · 409 재시도 허용(5s · 2회) · 3 req/s 계좌 공유.
   나머지 행 초안대로 승인. 코드·정본 표 반영은 a165 · a166.

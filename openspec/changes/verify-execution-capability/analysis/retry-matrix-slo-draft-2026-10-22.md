# retry matrix · 폴링 SLO 수치 제안 — task 1.3

> **초안 — 사용자 승인 대기.** 이 문서는 수치를 제안할 뿐 확정하지 않는다. 승인 전까지 코드 기본값
> (`execgw.DefaultRetryPolicy()`·`DefaultStaleness()`, `filldetect.DefaultConfig()`·`DefaultSLO()`)과
> 정본 표(`openspec/changes/archive/2026-07-26-harden-execution-base/retry-matrix.md`)는 바꾸지 않는다.
> 승인되면 그 표와 코드 기본값을 같은 커밋에서 함께 바꾼다(정본 표 머리말 규칙).

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

`0 throttled` 는 R5(최종 class)의 뜻이다. R6 이 보여 주듯 요청 단위로는 429 가 계속 난다 — 두 문장을 섞지 않는다.

## 2. 제안 수치

| 항목 | 현재 provisional | 제안 | 근거 |
|---|---|---|---|
| 주문 mutation 자동 재시도(place) | 0회 | **0회 유지** — 결과 불명은 IN_DOUBT 조회 해소, 재생(같은 키 재전송)은 비활성(2.7 결정) | §0.1, 2.7 |
| 취소·정정 `409 already-processing` | 0회(mutation 전부) | **결정 필요**: 브로커가 명시적으로 거절한 이 코드에 한해 `max(retryAfterSeconds, 1s)` 대기, 상한 5s, 최대 2회 재시도(노출 축소 방향 mutation). 거부하면 0회 유지 | R9, `verifylive.TransientRetryMaxWait`=5s |
| 계좌당 클라이언트 요청 속도 상한 | 없음 | **3 req/s**(페이지 간 최소 간격 350ms) | R2 의 13 req/s 에서 429, R4 의 3.56 req/s 는 지속 관측 — 그 아래로 둠 |
| GET 429 처리(엔진 폴링 경로) | `min(Retry-After, 30s)` 1회 대기 | **같은 주기 안 즉시 재시도 금지**. `Retry-After` 가 있으면 ≤30s 일 때만 존중, 없으면 이번 조회 포기 → staleness 규칙에 맡김 | R1(짧은 재시도 무효), `Retry-After` 헤더 존재는 [미측정] |
| GET 429 처리(soak·verify·IN_DOUBT 스캔 같은 비지연 walk) | — | 추가 2회, 15s → 30s(현행 `ReadRetryBackoff` 그대로) | R3, R6 |
| GET transient(5xx·전송) | 3회·400ms→3s·총 8s | **유지** | R7 p99 ≤ 0.6s 라 8s 예산 안에서 3회 가능 |
| 체결 감지 폴링 주기 | 3s | **유지**(OPEN 목록 1요청 ≈ 0.33 req/s) | OPEN 은 cursor/limit 무시·전량 1요청(M7) |
| 체결 감지 SLO | target 10s · p95 · 5m 창 · 최소 3표본 · grace 30s | **유지** | 3s 주기 + 단건 p99 0.6s 면 정상 최악 ≈ 4s, 10s 는 두 배 여유 |
| staleness 임계 | open_orders 20s · buying_power 45s · holdings 60s · price 15s | **유지**. 단 open_orders 20s 는 429 패널티 창 1회(≤15s 로 관측 — R3 에서 15s 대기 후 회복)를 흡수할 여유일 뿐이다 | R3 은 상한만 준다, 창 크기 자체는 [미측정] |
| IN_DOUBT 해소의 CLOSED walk | OPEN+CLOSED 각 1페이지 × 안정화 3회 | **1페이지(최신 100건) 유지, 전체 walk 금지**. 전체 walk 가 필요하면 3 req/s 로 ≈9s + 429 시 15s 백오프 | R6, R10 |

## 3. 측정이 없는 것 — 승인해도 남는다

- 429 의 실제 한도 값·창 크기·집계 단위(계좌/키/IP), `Retry-After` 헤더 존재 여부(M4·M8 모두 [미측정]).
- 조회 실패가 연속될 때의 패널티 연장 여부.
- 여러 프로세스(엔진·soak·console)가 같은 계좌로 동시에 칠 때의 합산 한도 — 위 3 req/s 상한은 **프로세스당**이다.
  soak 사이클(약 16분 간격, R5)이 엔진 폴링과 겹치는 순간의 합산은 재지 않았다.

## 4. 사람이 고를 것

1. 위 표의 제안 수치를 그대로 승인할지.
2. 취소·정정 `409 already-processing` 재시도를 허용할지(허용 시 정본 §0.1 에 예외 한 줄을 추가해야 한다).
3. 3 req/s 상한을 프로세스당으로 둘지, 계좌 전체 공유 상한으로 만들지(후자는 프로세스 간 조정이 필요하다).

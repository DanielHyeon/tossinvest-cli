# 실계좌 측정 기록 — verify-execution-capability

> 근거: `~/.local/share/tossos/capability-verify.jsonl` (계좌 `*******5921`, digest는 해당 파일의 step 레코드 참조).
> 규율: 브로커 주장은 실측 근거 또는 [미측정] 태그를 동반한다.

## 1차 실행 (2026-07-26, 일요일 — 콘솔 경유)

runs: `run-FOVHBDARTNFK3RKD`(21:31 KST) → `run-WE3EF3ZOHDCNBWGT`(22:46, 승인 거부) → `run-GTXLBOTXQIIMWORW`(22:48, 승인 통과·9단계 기록).

### 실측으로 확정된 사실

| # | 사실 | 근거 |
|---|---|---|
| M1 | **`order-hours-closed`**: 휴장일(일요일) POST /api/v1/orders는 HTTP 422, `code: "order-hours-closed"`, `message: "주문가능일이 아닙니다."` — 주문 오류 분류표에 추가할 실측 코드 (2.1 입력) | order-cancel·order-amend 단계, requestId `qlpibU2Ib7XBFyIX`·`qH2In6Yf5e1k57gi` |
| M2 | **오승인 거부 실전 동작**: 틀린 확인 문자열 → `verdict: refused`, "nothing was sent" — 전송 0건. 배치 승인 레일이 실계좌 조건에서 검증됨 | run-WE3EF3ZOHDCNBWGT approval 레코드 |
| M3 | **프로세스 재시작 간 resume**: 별개 pid 3개에 걸쳐 동일 plan digest `sha256:fac7f233…`로 재개, `approval.resumed=true` | run 2·3 approval 레코드 |
| M4 | **account-seq 해석 429**: GET orders/holdings의 lazy account-seq 해석이 브로커 429(ErrRateLimited는 HTTP 429 매핑 — internal/official/errors.go) 3연속(300ms 내 3회 시도 전부). 직전 ~4분 전 soak cycle 1이 동일 계열 endpoint 호출. **한도 창 크기·집계 단위는 미상** — 1.3 retry matrix의 입력 | run 1 read-fixtures·sellable-baseline·idempotency 단계 |
| M5 | **계좌 보유 KR 종목 0** — sell-boundary·conditional-* 단계가 설계대로 skip("the tool never buys") | run 3 skip 레코드 |
| M6 | costs 단계: 이 실행이 만든 주문 0건 → 수집 대상 없음(설계 일치 — 비용은 실제 체결에서만) | costs 단계 observations |

### soak 기록에서 확정된 추가 사실 (2026-07-27, `tossctl soak status`)

| # | 사실 | 근거 |
|---|---|---|
| M7 | **`GET /api/v1/orders`의 `status`는 필수 파라미터**: status 없는 요청은 HTTP 400 `code: "invalid-request"`, `data.field: "status"`. openapi 스펙과 일치(enum `OPEN`/`CLOSED`, "전체" 그룹 없음; `OPEN`은 cursor/limit 무시·전량 반환, `CLOSED`만 cursor 페이지네이션). soak full walk가 status 없이 호출해 **46/46 사이클 전패** → `GET /api/v1/orders` 미성공 → attestation 영구 차단. → task 1.9 발주 | `~/.local/share/tossos/capability-soak.jsonl` 46 cycle, requestId 예 `rkqNEGbfWwUq0jPx`; `docs/migration/openapi.latest.json` paths./api/v1/orders.get.parameters |

### 1.9/1.10 수정 바이너리 설치 후 실측 (2026-07-27 09:20 KST 설치, soak 사이클)

| # | 사실 | 근거 |
|---|---|---|
| M8 | **orders walk 429 버스트**: 1.9 수정으로 400은 사라졌으나(walk가 5페이지/100건 수집 시작), CLOSED walk가 기본 `limit`(20)으로 페이지를 간격 없이 연발 — 클린 사이클(직전 사이클과 15분 간격, 타 엔드포인트 전부 OK)에서도 **7요청/535ms(≈13 req/s) 후 429**, 직후 `GET /api/v1/orders/{id}` 1요청도 429(패널티 창). 계정 CLOSED 이력 ≥100건이라 기본 limit으로는 매 사이클 재현 → orders 미성공 지속 → attestation 차단. openapi `limit`은 최대 100(기본 20). 기존 실측과 정합: 관측 지속 한도 9.67 req/s(soak status), M4 한도 창 미상. → task 1.11 발주 | capability-soak.jsonl 사이클 `2026-07-27T00:40:42Z`(requests 7·latency 535ms·rate_limited), `00:25:27Z`·`00:25:41Z` 동일 패턴; openapi paths./api/v1/orders.get.parameters.limit |
| M9 | **매수 1주 holdings 반영**: 사용자 토스 앱 KR 1주 매수(2026-07-27 개장 직후) 후 holdings positions 2 → 3 관측 — 2.5·sell-boundary·2.8의 선행 조건 충족 | capability-soak.jsonl positions 추이(41사이클 연속 2 → 3), `00:40:42Z` 사이클 positions 3 |

### 1.11 수정 바이너리 실브로커 확인 (2026-07-27 10:12 KST 설치)

| # | 사실 | 근거 |
|---|---|---|
| M10 | **`GET /api/v1/orders`·`orders/{id}` 최초 성공** (56사이클 만): 설치 후 첫 완주 사이클(10:19:14 KST)에서 6/6 엔드포인트 OK·completeness evaluated·ok. walk는 26페이지/2,320건(limit=100), 도중 429 두 번을 15s 백오프로 흡수(soak.log 한글 안내 2건, 요청 28건 계상, latency 35.2s에 대기 포함). **계정 CLOSED 이력 ≥2,500건**(25페이지 상한 도달 — detail에 명시) — 구 limit 20으로는 125+페이지가 필요해 페이지 상한만으로도 원리적 완주 불가였음을 확증. attest 차단은 streak 3일차(7/28 충족 예정)·token refresh 관측(당일 21:27 KST 만료 예정) 2건만 잔존. 재시작 직후 버스트 사이클 2건은 연속 재시작 SIGINT로 중단된 것(2건 중 1건은 945건/10페이지까지 진행) — account-seq 해석 429는 M4 기지 사항 유지 | capability-soak.jsonl 사이클 `2026-07-27T01:19:14Z`(전건 OK), `01:18:50Z`·`01:19:02Z`(중단 사이클); soak.log 01:19:19·01:19:35 백오프 안내; `tossctl soak status` NOT READY 잔여 2건 |

### 2026-07-27 장중 창 (09:00–15:30 KST) — 측정 0건

| # | 사실 | 근거 |
|---|---|---|
| M11 | **장중 창이 무동작 버튼으로 소모됨**: 이 날 콘솔에서 검증이 두 번 시작됐고(`YW7XSICTO4` 14:09 이전, `ORZEYQU6SI` 18:39 이후) 둘 다 `0 step(s) recorded`(에러 없음)로 끝났다 — 증거 기록 파일의 마지막 쓰기는 여전히 2026-07-26 22:48이고 단계 판정도 1차 실행 그대로다. 원인은 승인 이전이다: 시작 화면의 기본 버튼 [이어하기](`mode=resume`)는 판정이 terminal인 단계를 건너뛰는데(`Runner.settled`) 현재 기록은 전 단계가 terminal이라 **구조적으로 무동작**이며, 실제로 측정하는 [재측정](`mode=redo`)은 아래쪽 별도 섹션에 있었다. 승인 화면·nonce 타이핑에는 도달하지도 않았다. → change `console-click-approval`(무동작 기본값 제거 + 승인 클릭화) 발주 | `~/.local/share/tossos/console-launch.log` 두 줄(`verification … finished — 0 step(s) recorded`), `capability-verify.jsonl` mtime 2026-07-26T22:48:58+0900, `internal/console/pages.go` handleStart·`templates.go` 시작 화면 |

### 미측정으로 남은 항목 (이번 실행 기준)

- 2.1 status enum fixture / 2.7 멱등키 재생·conflict / 2.2 place-cancel-amend / 2.8 sellable 의미: **휴장 + 429로 전부 미측정** — 장중 재실행 필요.
- 2.5 조건주문 전체 / sell-boundary / sellable-reserved: **보유 0으로 원리적 측정 불가** — SELL측 검증(SINGLE+MARKET 손절 가설, 2.6의 임계 입력)은 KR 종목 최소 1주 보유가 선행 조건.
- `conditional.trigger*`: 설계상 deferred(시장 조건 필요) — 변동 없음.

### 발견된 도구 갭 (→ task 1.7)

`fail`/`skipped`는 terminal verdict라 resume이 재시도하지 않는다(runner.go `settled`, `Verdict.Terminal()`).
CLI는 `--redo`로 넘지만 **콘솔은 always-resume + redo 부재**라, 환경 원인(휴장·429·보유 0)으로
실패·생략된 측정을 웹 화면에서 다시 수행할 방법이 없다 — 콘솔 단독 운용(사용자 결정)에서는
검증이 "완료·측정 0건"으로 고착된다. 재측정 경로·장시간 사전 경고·429 강건성을 1.7로 발주.

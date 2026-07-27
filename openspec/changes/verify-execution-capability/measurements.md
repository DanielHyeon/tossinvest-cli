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

### 미측정으로 남은 항목 (이번 실행 기준)

- 2.1 status enum fixture / 2.7 멱등키 재생·conflict / 2.2 place-cancel-amend / 2.8 sellable 의미: **휴장 + 429로 전부 미측정** — 장중 재실행 필요.
- 2.5 조건주문 전체 / sell-boundary / sellable-reserved: **보유 0으로 원리적 측정 불가** — SELL측 검증(SINGLE+MARKET 손절 가설, 2.6의 임계 입력)은 KR 종목 최소 1주 보유가 선행 조건.
- `conditional.trigger*`: 설계상 deferred(시장 조건 필요) — 변동 없음.

### 발견된 도구 갭 (→ task 1.7)

`fail`/`skipped`는 terminal verdict라 resume이 재시도하지 않는다(runner.go `settled`, `Verdict.Terminal()`).
CLI는 `--redo`로 넘지만 **콘솔은 always-resume + redo 부재**라, 환경 원인(휴장·429·보유 0)으로
실패·생략된 측정을 웹 화면에서 다시 수행할 방법이 없다 — 콘솔 단독 운용(사용자 결정)에서는
검증이 "완료·측정 0건"으로 고착된다. 재측정 경로·장시간 사전 경고·429 강건성을 1.7로 발주.

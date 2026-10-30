# a166 — 계좌 공유 요청 속도 상한(3 req/s)과 엔진 GET 429 처리

> **3판 초안 · NOT READY (2026-10-30).** 사용자는 계좌 전체 3건/초 상한 우선과 보호 요청 대기 허용 방향을 승인했다. a166 재설계에 한정된 보호 대기 허용은 승인 완료이며, 구체 지연·장애 정책이나 구현·활성화 승인은 아니다. 독립 리뷰 `analysis/review-v2-20261030.md`는 FAIL이다.

## Why

`verify-execution-capability` task 1.3 에서 사용자가 「요청 속도 상한 3 req/s 는 **계좌 전체 공유**」로 정했다(2026-10-22,
`openspec/changes/verify-execution-capability/issues.md` 「사용자 결정 (2026-10-22, 1.3 확정)」 결정 3, 확정 표
`openspec/changes/verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md` §2). 같은 표의 「GET 429 처리(엔진 폴링
경로)」 행 — 같은 주기 안 즉시 재시도 금지, `Retry-After` 가 있으면 ≤30s 일 때만 존중, 없으면 이번 조회 포기 → staleness —
도 코드와 다르다(현행 `min(Retry-After, 30s)` 1회 대기). 지금 TossOS 에는 클라이언트 쪽 속도 상한이 **없다**(정본 retry matrix
§2·확정 표의 「현재 provisional」 열).

근거 측정:

| # | 관측 | 출처 |
|---|---|---|
| M4 | lazy account-seq 해석이 300ms 안 3회 시도 전부 429. 직전 약 4분 전 soak 사이클이 같은 계열 endpoint 를 불렀다. 한도 창 크기·집계 단위 미상 | `verify-execution-capability/measurements.md` M4 |
| M8 | 간격 없는 페이지 연발 7요청/535ms(≈13 req/s) 뒤 429, 직후 단건 `GET /api/v1/orders/{id}` 1요청도 429(패널티 창) | measurements.md M8 |
| M10 | 26페이지 walk 가 도중 429 두 번을 15s 백오프로 흡수하고 완주(요청 28건, latency 35.2s 에 대기 포함) | measurements.md M10 |
| R4·R6 | attestation 지속 관측 3.5636 req/s(한도 아님). 3일 soak 272 사이클 중 267 사이클의 orders walk 가 ≥15s — 최종 class 는 ok 지만 요청 단위 429 는 계속 난다 | 확정 표 §1 |
| 2026-10-22 | 엔진과 사람의 수동 조회가 같은 시각에 함께 429 를 맞았다 — 한도가 프로세스가 아니라 계좌(또는 키·IP) 단위로 묶인다는 관측 | Manager 가 이 change 정의 지시로 전달. **이 저장소에 로그 영수증이 없다 — 0.2 에서 확보** |

한 프로세스 안에서만 속도를 지키면 엔진·soak·verify·프로브·CLI·콘솔이 같은 계좌를 동시에 칠 때 합이 한도를 넘는다. 그 합이 넘으면
엔진의 필수 조회가 429 로 막히고 staleness 가 신규 진입을 차단한다 — 안전 쪽으로 실패하지만 보호 청산 경로의 조회도 같은 예산을
쓴다는 것이 문제다.

## What Changes

- 같은 계좌를 치는 모든 TossOS 프로세스가 **하나의** 3 req/s 예산을 나눠 쓴다(간격 350ms 는 3 req/s 에서 유도한 미승인 후보이며 확정 표에는 없다 — 1판의 「확정 표의 값」은 오기, design 2판 §0에서 정정).
- 엔진 폴링 경로의 GET 429 처리를 확정 표대로 바꾼다(같은 주기 안 즉시 재시도 금지).
- 비지연 walk(soak·verify·IN_DOUBT 스캔)의 15s→30s 백오프(`ReadRetryBackoff`)는 그대로다.
- 정본 retry matrix §1·§2·§3(예산 계상)을 **코드와 같은 커밋에서** 고친다.
- 보호 요청도 상한 안에서 기다린다. 보호 우선은 순서 정책이며 상한 면제가 아니다. 공유 장애 시에도 우회하지 않는다. 구체 지연과 실패 결과는 미결이다.

## 지금 코드 — 상한이 들어갈 자리 후보

근거: CodeGraph 1.6.0 `codegraph callers`(`analysis/code-context/codegraph-callers.txt`, 2026-10-22) + `grep -rn`. 이 proposal 은
함수 내부 분기를 근거로 쓰지 않는다 — 아래는 호출 자리와 파일 수준 사실뿐이다.

- `internal/official/client.go:191` `Client.doRequest` 가 `official.Client` 안의 유일한 `c.hc.Do`(:193) 자리다. CodeGraph 상 호출자는
  `send`(:320) 하나이고, `send` 의 호출자는 `deleteAcct`(:369)·`getWithHeaders`(:392)·`postWithHeaders`(:427)다.
- 이 자리를 **거치지 않고** 공식 API 를 치는 전송: `internal/official/token.go:144`(토큰 교환), `internal/official/a112_mbus_read.go:153`
  (일회성 측정 GET), `tools/a087-market-sell-probe/send.go:85`(프로브 POST). 토큰 교환이 같은 한도에 세어지는지는 [미측정].
- 이미 있는 것(search-first, 저장소만 봄):
  - `internal/official/ratebudget.go` — 응답의 `X-RateLimit-Limit/-Remaining/-Reset` 를 경로별로 기록(`doRequest` :199
    `c.rates.record`). **관측만** 하고 막지 않으며 프로세스 안에만 있다.
  - `internal/verifylive/retry.go:61`·`internal/soak/soak.go:239` `ReadRetryBackoff`, `reconcile.DefaultRateLimitBackoff` — 429 뒤
    백오프. 사전 속도 제한이 아니다.
  - 프로세스 간 조정의 선례: 토큰 캐시 파일을 콘솔·엔진·API 데몬이 공유(a082, `internal/official/token.go`, `token_shared_cache_test.go`),
    커널 flock(`internal/enginelock/flock_unix.go`, `internal/config/adoption_flock_unix.go`).
  - 외부 의존성: `golang.org/x/time/rate` 는 `go.mod` 에 없고, 있어도 프로세스 안 상한이다. 외부 레지스트리·MCP 는 이 정의 단계에서
    **보지 않았다** — design 의 search-first 에서 본다(들이는 것은 사람 승인 + `dependency-auditor`).

## 선택지 스케치 (고르지 않음)

1. **공유 상태 파일 + flock 토큰 버킷** — 계좌별 파일 하나에 마지막 발급 시각/토큰 수를 두고 각 요청 전에 flock 아래서 갱신.
   새 프로세스·데몬 없음, 선례 있음(a082·enginelock). 비용: 요청마다 파일 잠금, `/mnt/D` 같은 파일시스템의 flock 의미, 크래시 시
   상태 회복.
2. **단일 브로커 프록시(엔진 소유)** — 엔진이 전송을 독점하고 다른 프로세스는 로컬 소켓으로 요청. 예산·우선순위를 한 곳에서 정확히
   다룬다. 비용: 엔진이 꺼져 있을 때 soak·CLI 경로, 새 표면, upstream CLI 동작(토글 OFF = upstream) 보존.
3. **정적 분할** — 프로세스 종류별 고정 몫(예: 엔진 2, 그 밖 합 1). 조정 없음. 비용: 몫을 지키는지 아무도 모름, 유휴 몫 낭비, 같은
   종류 프로세스 둘이면 깨짐.
4. **헤더 피드백 공유** — `ratebudget.go` 가 이미 읽는 `X-RateLimit-Remaining` 을 공유 파일로 내보내 모두가 따른다. 비용: 헤더의 의미·
   집계 단위가 [미측정], 사후 신호라 버스트를 못 막음.

어느 안이든 **키는 계좌 식별**이어야 한다 — config 디렉터리가 아니다. 다른 `--config-dir` 로 같은 계좌를 치는 프로세스가 예산을
따로 가지면 공유가 아니다(토큰 캐시가 config 디렉터리별로 갈려 서로 무효화한 2026-10-19 사례가 같은 모양이다 — 파일 기억 기준, 이 change 에서 재검증하지 않음).

## 위험도와 의무 (High-risk)

retry matrix·rate limit 은 WORKFLOW 「위험도 분류」의 High-risk 경로이고, 상한은 엔진의 필수 조회와 주문·청산 전송 모두에 걸린다.

- **FLM/BTM 필수**: 고른 삽입 자리의 기존 함수(후보 `official.Client.doRequest`·`send`)와 엔진 폴링의 429 처리 함수. 분기 주장은 AST 로만.
- **Pre-Edit 선언** 필수.
- **적대적 리뷰**: 적대적 Eng 보이스 포함 다중 리뷰 — 손절·flatten 지연, 교착(잠금 쥔 채 크래시), 시계 역행, 공정성(한 프로세스가
  예산 독점), 상한을 우회하는 전송 자리 누락을 반례로 찾는다.
- **변이 원장**: 공유 키·간격 계산·잠금·우선순위·429 갈래 가드에 변이, 양성 대조군 포함. 동시성 주장은 `-race` 와 다중 프로세스
  시험으로(통과 ≠ 증거 — 실패시켜 본 뒤 근거로).
- 실계좌 부하 시험 금지. 합산 한도 실측이 필요하면 사람 승인 절차로 따로 한다.

## Capabilities

### Modified Capabilities

- `order-execution`: 「Retry Matrix 산출물」(429 처리·클라이언트 속도 상한·계좌 공유 예산). spec delta 는 design 1판에서 「Retry Matrix 산출물」 MODIFIED + 「계좌 공유 요청 admission」 ADDED 로 다시 썼다.

## Impact

- 대상 코드(예상): `internal/official/`(전송 자리), 엔진 폴링 조회 경로, soak·verify·console·CLI 가 같은 클라이언트를 쓰는지에 따라
  그 진입점들, 정본 `openspec/changes/archive/2026-07-26-harden-execution-base/retry-matrix.md`.
- 운영: 공유 상태 파일·소켓을 고르면 컨테이너 배포(볼륨·권한)와 config 경로 유도가 바뀐다 — design 에 rollback 과 함께 적는다.
- a165(주문 접수 재시도)의 재시도 요청도 이 예산에 계상한다. 착지 순서는 design 에서 정한다.
- 위험도: **High**. NOT READY.

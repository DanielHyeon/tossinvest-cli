# Retry Matrix — harden-execution-base (task 2.6)

> order-execution "Retry Matrix 산출물"의 산출물. **표 없이 구현하지 않는다(SHALL NOT)** 규정에 따라
> 구현(`internal/execgw/retry.go`)보다 이 표가 먼저다.
>
> 수치는 **보수 기본값(provisional)** 이다. 실측 확정은 별도 change
> `verify-execution-capability`의 rate-limit·지연 측정 결과로 갱신한다. 갱신 시 이 파일과
> `execgw.DefaultRetryPolicy()`·`DefaultStaleness()`를 같은 커밋에서 함께 바꾼다.

> **a166 정합화(2026-10-31):** 아래의 required engine-poll은 검증된 v2 공통 Budget과 poll binding이 연결된 경우다. OFF/legacy는 기존 동작을 보존한다. 운영자가 required로 전환했거나 모든 참여자 검증이 끝났다는 뜻이 아니다. transient/default staleness 수치는 이 로트에서 변경하지 않는다.

## 0. 최상위 규칙

1. **주문 mutation의 맹목 재전송은 금지하며 기본 자동 재시도는 0회다.** 결과 불명은
   IN_DOUBT 해소 절차(task 2.7/2.8)로 처리하고, 같은 키·본문을 쓰는 별도 replay도 그 절차의
   원장·권한·멱등성 창 가드를 충족해야 한다. `409 already-processing`이 HTTP 상태와
   본문 code 양쪽에서 일치한 **취소**는 a165의 경로 한정 예외다: 엔진은 재전송 없이
   FAILED_CONFIRMED로 끝내고, CLI flatten만 최초 전송을 포함한 8초 창 안에서 추가 2회까지
   재전송한다. 주문 제출·정정·조건주문에는 이 예외가 없다.
2. 조회(GET)만 재시도한다. 재시도는 **횟수 예산 + 총 시간 예산 + bounded jitter** 세 가지로 동시에 제한한다.
3. required engine-poll은 429를 관측한 같은 논리 Query를 재시도하지 않는다. 유효한 `Retry-After` ≤30s는 같은 계좌·역할·RequiredQuery의 다음 호출을 새 wire 없이 보류한다. 누락·invalid·초과 값은 새 hint를 만들거나 기존 유효 hint를 지우지 않는다. 보호·일반 price가 같은 slot을 공유하며 주문 본체·flatten·별도 walk를 계좌 전역 cooldown으로 막지 않는다. OFF/legacy는 기존 RetryAfterSource와 backoff를 유지한다.
4. Retrier까지 반환된 401/403은 재시도 0회 + **즉시 신규 진입 차단(latch)** + durable escalation/알림. 자동 해제 없음 — 자격증명 갱신 후 운영자가 해제. official.Client 내부의 기존 token refresh·401 재전송은 이 최종 오류 처리와 별개이며 required에서는 각각 공통 admission을 거친다. 중간401을 새 AuthFatal로 승격하지 않는다.
5. 필수 조회의 staleness가 임계를 넘으면 신규 진입 차단. 조회가 성공하면 **자동 해제**된다.
6. EntryGate 차단은 "신규 진입"에만 적용된다. 취소·청산 mutation에 이 gate를 적용하지 않는다. 다만 사용자가 승인한 a166 required 공통 admission에서는 보호 요청도 간격·소유권을 기다리며, 보호 무지연을 보장하지 않는다.

## 1. endpoint × method × 오류 클래스

| endpoint | method | 용도 | 오류 클래스 | 재시도 | 대기 | 부수 효과 |
|---|---|---|---|---|---|---|
| `/api/v1/orders` | POST | 주문 제출 | 전부 | **0회** | — | 2xx=CONFIRMED, 400/401/403/404/405/415/422=FAILED_CONFIRMED, 그 외·429·5xx·전송중단=IN_DOUBT |
| `/api/v1/orders/{id}/cancel` | POST | 엔진 청산 전 취소 | `409 already-processing`(HTTP·본문 code 일치) | **0회** | — | FAILED_CONFIRMED; 같은 주기 재전송·대기 없음, 대상 주문이 살아 있으면 신규 진입만 latch |
| `/api/v1/orders/{id}/cancel` | POST | CLI flatten 취소 | `409 already-processing`(HTTP·본문 code 일치) | **추가 최대 2회** | `retryAfterSeconds`(기본 1s), 누적 5s 이내·최초 전송 전부터 8s 창 | 성공은 재전송 횟수 꼬리표를 남김; 소진·중단·전송 전 실패는 IN_DOUBT로 정착하고 해당 종목 매도 보류 |
| `/api/v1/orders/{id}/cancel` | POST | 그 외 취소 | 전부 | **0회** | — | 위와 동일 |
| `/api/v1/orders/{id}/modify` | POST | 정정 | 전부 | **0회** | — | 위와 동일 |
| `/api/v1/conditional-orders*` | POST/DELETE | 조건주문 | 전부 | **0회** | — | 위와 동일 |
| `/api/v1/orders` | GET | 미체결·체결 목록(pagination) | transient(전송·5xx) | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| `/api/v1/orders/{id}` | GET | 단건 주문 상태 | transient | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| `/api/v1/buying-power` | GET | 매수가능금액 | transient | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| `/api/v1/holdings` | GET | 보유수량 | transient | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| 가격·시세 GET | GET | 진입 판단 | transient | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| required engine-poll GET | GET | 같은 논리 Query | **429** | **0회** | 현재 Query는 종료; 유효 ≤30s hint 동안 후속 동종 poll은 wire0 | freshness 성공 0; token429도 Query 종료, poll hint는 생성하지 않음 |
| OFF/legacy Retrier GET | GET | 기존 동작 유지 | **429** | 기존 MaxAttempts/8s budget 안 | RetryAfterSource가 있으면 상한 검사; 없으면 기존 backoff | 현재 생산 OFF 엔진은 source nil; 과거 '1회 대기' 문구를 실제 보장으로 쓰지 않음 |
| 별도 비지연 walk GET | GET | soak·verify·IN_DOUBT 스캔 | **429** | 기존 추가 2회 | 기존 15s → 30s | a166에서 호출자 재시도 정책 변경 없음; required wire별 admission/기록 적용 |
| Retrier에 최종 반환된 오류 | GET | — | **401/403** | **0회** | — | **즉시 신규 진입 차단(latch)** + durable escalation/알림 |
| 모든 GET | GET | — | 그 외 4xx(permanent) | **0회** | — | 호출자에게 오류 반환, freshness 미갱신 |

오류 클래스 판정(`execgw.ClassifyQueryError`):

| 관측 | 클래스 |
|---|---|
| `official.ErrTransport`(연결·타임아웃), 5xx(`official.ErrServer`) | `ClassTransient` |
| 429(`official.ErrRateLimited` 또는 `official.APIError.Code == 429`) | `ClassRateLimited` |
| 401/403(`official.ErrAuth`, `official.ErrIPNotAllowed`) | `ClassAuthFatal` |
| 그 외 `*official.APIError`(400/404/422 …) | `ClassPermanent` |
| `context.Canceled`/`DeadlineExceeded` | `ClassCanceled`(재시도 없음) |

## 2. 보수 기본값 (provisional)

| 파라미터 | 값 | 근거 |
|---|---|---|
| 조회 최대 시도 횟수 | **3회**(최초 1 + 재시도 2) | 4회 이상은 예산만 태우고 staleness 임계를 먼저 넘긴다 |
| 조회 총 시간 예산 | **8s** | 진입 판단 1회가 8초를 넘으면 그 판단은 이미 낡았다 |
| 최초 backoff | **400ms** | 일시적 네트워크 흔들림 회복에 충분한 최소값 |
| backoff 배수 | **2×** | 400ms → 800ms → 1.6s |
| backoff 상한 | **3s** | 총 예산 8s 안에서 3회 시도가 가능한 최대값 |
| jitter | **±25% bounded** | 동시 재시도 동기화(thundering herd) 방지. 상·하한이 있는 곱셈 jitter |
| `Retry-After` 상한 | **30s** | required는 응답별 hint로 후속 동종 poll에 적용; OFF/legacy 기존 정책은 유지 |
| 재시도 대상 | GET만 | §0.1 |

### staleness 임계 (신규 진입 차단 기준)

| 필수 조회 | 임계 | 근거 |
|---|---|---|
| 미체결 주문 목록(`open_orders`) | **20s** | 체결 감지 폴링 주기(3s 목표)의 여유 배수. 이보다 낡으면 in-flight 주문을 모른 채 진입하게 된다 |
| 매수가능금액(`buying_power`) | **45s** | 잔고는 주문·체결·환전으로만 변한다 |
| 보유수량(`holdings`) | **60s** | 체결 감지가 별도로 delta를 반영한다 |
| 가격(`price`) | **15s** | 진입 가격 판단의 유효 수명 |

- **한 번도 성공하지 않은 필수 조회는 "무한히 낡음"으로 취급해 진입을 차단한다**(fail-closed).
- 임계 초과로 차단된 뒤 해당 조회가 성공하면 **자동 해제**된다(order-execution "조회 복구 후 자동 해제").
- 401/403 latch는 자동 해제되지 않는다.

## 3. rate limit 예산 계상 (§0.4)

본 change가 추가하는 정상 상태 호출량(계좌 1개 기준, 초당):

| 호출 | 빈도 | 근거 task |
|---|---|---|
| 미체결 목록 GET (pagination 1페이지 가정) | 3s마다 1회 ≈ 0.33 req/s | 3.1 체결 감지 폴링 |
| 잔고 GET | 15s마다 1회 ≈ 0.07 req/s | 3.4 reconcile 스냅샷 |
| 보유 GET | 15s마다 1회 ≈ 0.07 req/s | 3.4 |
| 주문 단건 GET | in-flight 주문당 3s마다 1회 (심볼당 in-flight 1건 제한) | 2.7/3.1 |
| IN_DOUBT 해소 스캔 | 발생 시에만, OPEN+CLOSED 각 페이지 1회 × 안정화 3회 | 2.7 |

위 빈도표는 초기 설계의 평균 추정이며 현재 전체 호출자·token·재전송을 포함한 상한 증거가 아니다. 총 시간 예산 8s만으로 순간 버스트 1.5 req/s 이하를 증명할 수 없다. a166 required의 로컬 요청 시작은 같은 공통 group wire 소유권과 350ms 최소 간격으로 제한한다. 기준은 TLS 연결의 첫 HTTP plaintext Write이며 브로커 도착률·처리 완료·429 미발생을 보장하지 않는다. OFF와 required 미참여 프로세스까지 포함한 운영 총량은 별도 참여자 검증 없이는 주장하지 않는다.

## 4. 구현 대응표

| 규정 | 구현 |
|---|---|
| mutation 무재시도 | `execgw.Gateway.submit`이 `journal.Attempt.Dispatch`를 1회만 호출. `Retrier`에 mutation 진입점 자체가 없음. a165 예외는 `CancelRefusalPolicy`: 엔진 `EngineCancelRefusal()`은 409 already-processing을 재전송 없이 확정 거절로 분류하고, flatten `FlattenCancelRefusal()`만 제한된 취소 재전송을 허용함 |
| 조회 예산·jitter | `execgw.Retrier.Query` + `execgw.RetryPolicy` |
| required 429·후속 poll hint | `accountbudget.Transport` 응답별 event/hint + token collector 전파 + `Retrier`의 검증된 poll binding; 같은 Query collector 429/기록 오류 뒤 후속 wire·freshness 성공 금지 |
| OFF/legacy 429 | `Retrier.rateLimitWait`와 선택적 `RetryAfterSource`; `RetryAfterTransport`의 존재만으로 생산 배선을 주장하지 않음 |
| 401/403 즉시 차단 | `execgw.Retrier.Query` → `EntryGate.Block(ReasonBrokerAuthRejected)` (latch) |
| staleness 진입 차단 | `execgw.EntryGate.CheckEntry()` — 게이트웨이가 노출 증가 mutation 직전에 조회 |

required hint는 boot source와 동일 slot의 영속 state로 복원하며 반복 reboot의 wall-time 상한은 주장하지 않는다. 기록 오류·orphan poison은 성공으로 처리하지 않는다. 응답 수용을 이미 관측한 뒤 기록 실패했다고 원래 HTTP/원장 결과를 NOT_SENT로 바꾸지 않는다. 상세 계약과 고정 시험 영수증은 a166의 engine-poll-observation-plan 및 해당 구현 로트를 참조한다.

# Retry Matrix — harden-execution-base (task 2.6)

> order-execution "Retry Matrix 산출물"의 산출물. **표 없이 구현하지 않는다(SHALL NOT)** 규정에 따라
> 구현(`internal/execgw/retry.go`)보다 이 표가 먼저다.
>
> 수치는 **보수 기본값(provisional)** 이다. 실측 확정은 별도 change
> `verify-execution-capability`의 rate-limit·지연 측정 결과로 갱신한다. 갱신 시 이 파일과
> `execgw.DefaultRetryPolicy()`·`DefaultStaleness()`를 같은 커밋에서 함께 바꾼다.

## 0. 최상위 규칙

1. **주문 mutation(POST place/cancel/modify, 조건주문 create/modify/cancel)은 어떤 오류에도 자동 재시도하지 않는다.**
   토스 공식 API에 멱등성 키가 없다 — 재시도는 곧 주문 중복이다. 결과 불명은 재시도가 아니라
   IN_DOUBT 해소 절차(task 2.7/2.8)로 처리한다.
2. 조회(GET)만 재시도한다. 재시도는 **횟수 예산 + 총 시간 예산 + bounded jitter** 세 가지로 동시에 제한한다.
3. 429는 `Retry-After`를 존중하되 **상한**을 둔다. 상한 초과 값이 오면 대기하지 않고 그 조회를 포기한다
   (진입은 staleness 규칙으로 차단된다).
4. 401/403은 재시도 0회 + **즉시 신규 진입 차단(latch)** + 알림. 자동 해제 없음 — 자격증명 갱신 후 운영자가 해제.
5. 필수 조회의 staleness가 임계를 넘으면 신규 진입 차단. 조회가 성공하면 **자동 해제**된다.
6. 차단은 항상 "신규 진입"에만 적용된다. 취소·청산 등 노출을 줄이는 mutation은 차단하지 않는다(§0.3).

## 1. endpoint × method × 오류 클래스

| endpoint | method | 용도 | 오류 클래스 | 재시도 | 대기 | 부수 효과 |
|---|---|---|---|---|---|---|
| `/api/v1/orders` | POST | 주문 제출 | 전부 | **0회** | — | 2xx=CONFIRMED, 400/401/403/404/405/415/422=FAILED_CONFIRMED, 그 외·429·5xx·전송중단=IN_DOUBT |
| `/api/v1/orders/{id}/cancel` | POST | 취소 | 전부 | **0회** | — | 위와 동일 |
| `/api/v1/orders/{id}/modify` | POST | 정정 | 전부 | **0회** | — | 위와 동일 |
| `/api/v1/conditional-orders*` | POST/DELETE | 조건주문 | 전부 | **0회** | — | 위와 동일 |
| `/api/v1/orders` | GET | 미체결·체결 목록(pagination) | transient(전송·5xx) | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| `/api/v1/orders/{id}` | GET | 단건 주문 상태 | transient | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| `/api/v1/buying-power` | GET | 매수가능금액 | transient | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| `/api/v1/holdings` | GET | 보유수량 | transient | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| 가격·시세 GET | GET | 진입 판단 | transient | 최대 2회 재시도 | backoff+jitter | 성공 시 freshness 갱신 |
| 모든 GET | GET | — | **429** | Retry-After ≤ 상한이면 1회 대기 후 재시도 | `min(Retry-After, 30s)` + jitter | 상한 초과면 즉시 포기 |
| 모든 GET | GET | — | **401/403** | **0회** | — | **즉시 신규 진입 차단(latch)** + critical 알림 |
| 모든 GET | GET | — | 그 외 4xx(permanent) | **0회** | — | 호출자에게 오류 반환, freshness 미갱신 |

오류 클래스 판정(`execgw.ClassifyQueryError`):

| 관측 | 클래스 |
|---|---|
| `official.ErrTransport`(연결·타임아웃), 5xx(`official.ErrServer`) | `ClassTransient` |
| 429(`official.ErrRateLimited`) | `ClassRateLimited` |
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
| `Retry-After` 상한 | **30s** | 그 이상 기다리느니 진입을 막는 편이 안전하다 |
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

정상 상태 합계 ≈ **0.5 req/s 미만**. 재시도 최악값은 조회당 3배이나 총 시간 예산 8s가 상한이므로
버스트도 1.5 req/s를 넘지 않는다. 실제 rate limit 값은 verify-execution-capability에서 측정한다.

## 4. 구현 대응표

| 규정 | 구현 |
|---|---|
| mutation 무재시도 | `execgw.Gateway.submit`이 `journal.Attempt.Dispatch`를 1회만 호출. `Retrier`에 mutation 진입점 자체가 없음 |
| 조회 예산·jitter | `execgw.Retrier.Query` + `execgw.RetryPolicy` |
| 429 Retry-After 상한 | `execgw.RetryAfterTransport`(헤더 포착) + `RetryPolicy.MaxRetryAfter` |
| 401/403 즉시 차단 | `execgw.Retrier.Query` → `EntryGate.Block(ReasonBrokerAuthRejected)` (latch) |
| staleness 진입 차단 | `execgw.EntryGate.CheckEntry()` — 게이트웨이가 노출 증가 mutation 직전에 조회 |

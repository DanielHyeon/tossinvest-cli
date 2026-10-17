# Function Logic Map: `Ntfy.Publish`

- Source: `internal/obs/ntfy.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **a099가 이 함수에서 바꾼 것은 리터럴 하나다.** `10 * time.Second` →
> `DefaultPublishTimeout` (`:97`).
>
> **그 리터럴이 임차 길이를 정하는 항이었다.** `AlertDeliveryBound`는
> `attempts × (publish timeout + write wait) + …`이고, publish timeout이 여기
> 함수 안에 숨어 있으면 **유도가 그 숫자를 복사해야 한다.** 복사한 숫자는
> 언젠가 안 맞고, 안 맞는 순간 발송자가 자기 예산을 쓰는 도중에 임차를 잃는다.
>
> **이 번들은 「숫자 하나를 이름으로 바꾼 것이 왜 안전 편집인가」의 증거다.**
> 값도, 분기도, 이탈도, 헤더도 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Topic` | 공백 아님 | B1 `:87` | 이탈 `:88` `ErrNtfyNotConfigured` |
| `n.BaseURL` | 공백이면 `DefaultNtfyBaseURL` | B2 `:91` | — |
| **`n.Timeout`** | **`<= 0`이면 `DefaultPublishTimeout`** | B3 `:96` | **a099가 바꾼 유일한 줄** |
| `msg.Body` | 임의 | 호출자 | 본문이 된다 |
| `n.HTTPClient` | nil이면 새로 만든다 | B7 `:122` | 같은 timeout을 쓴다 |
| 응답 상태 | `2xx` | B9 `:134` | 그 밖은 오류 |

**불변식**: *"한 번의 publish는 `timeout`을 못 넘긴다."*
`context.WithTimeout` `:99`와 `http.Client{Timeout: timeout}` `:123`이 둘 다 건다.
**`HTTPClient`가 주입되면 두 번째 보장이 사라진다** — ctx만 남는다.

## Branches and early returns

AST 열거 — 분기 9 · 이탈 5 · 호출 29 · defer 2.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:87` | topic이 공백 | 없음 | 이탈 `:88` `ErrNtfyNotConfigured` | 기존 |
| B2 `:91` | base가 공백 → 기본 URL | 값 채우기 | — | 기존 |
| **B3 `:96`** | **`timeout <= 0` → `DefaultPublishTimeout`** | 값 채우기 | — | **기존 — 값이 안 바뀌었다** |
| B4 `:104` | 요청 생성 실패 | 없음 | 이탈 `:105` 포장 오류 | 없음 |
| B5 `:110` | title이 비어 있지 않다 | `Title` 헤더 | — | 기존 |
| B6 `:117` | token이 비어 있지 않다 | `Authorization` 헤더 | — | 기존 |
| B7 `:122` | `HTTPClient == nil` | 새 클라이언트 (`timeout` 포함) | — | 기존 |
| B8 `:126` | `client.Do` 실패 | 없음 | 이탈 `:127` 포장 오류 | 기존 |
| B9 `:134` | 상태가 `2xx`가 아니다 | 없음 | 이탈 `:135` 오류 (본문 포함) | 기존 |
| — 이탈 `:138` | 정상 | **네트워크 발송이 끝났다** | `nil` | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` `:86`, `:90`, `:110`, `:117`, `:136` | 입력 정리 | — | `ast.json` calls |
| **`context.WithTimeout` `:99`** | **한 번의 publish의 기한** | `defer cancel()` `:100` | 같음 · defers |
| `http.NewRequestWithContext` `:102` | 요청 | 실패면 B4 | 같음 |
| `client.Do` `:125` | **네트워크** | 실패면 B8. 재시도는 **이 함수 밖**(`deliver`) | 같음 |
| `io.ReadAll(io.LimitReader(…, 4096))` `:132` | 응답 본문 | 오류를 버린다 (`_`) — 연결 재사용이 목적 | 같음 |
| `resp.Body.Close` `:129` (defer) | 정리 | — | `ast.json` defers |

**live binding — 유일한 프로덕션 호출자 경로**: `Notifier.deliver` `:435`와
`Notifier.Flush` `:631`이 `n.Publisher.Publish`를 부르고, 배선된 구현이 이것이다.

**`DefaultPublishTimeout`을 읽는 두 번째 자리**: `obs.AlertDeliveryBound`
(`alert_lease.go:44`). **그 두 자리가 같은 상수를 보는 것이 a099의 편집이다.**

## State mutations and fallbacks

- **원장을 안 건드린다.** 이 함수는 네트워크만 한다.
- **재시도가 없다.** 한 번 보내고 결과를 돌려준다. 예산은 `deliver`가 진다.
- **폴백**: base URL과 timeout에 기본값이 있다. topic에는 없다 — 없으면 거절이다.

## Safety conclusion

- **Safe edit boundary**: **리터럴 하나를 상수 이름으로 바꿨다.** 값(10초)도
  분기도 이탈도 그대로다. §5.6 실측: 분기 9 · 이탈 5 · defer 2로 base와 같다.
- **High-risk impact**: **yes (간접)** — 이 함수의 기한이 배달 상한의 항이고,
  배달 상한이 임차 길이를 정한다. **값을 바꾸면 임차가 따라 움직여야 한다.**
  `TestTheDefaultLeaseOutlastsTheDeliveryBound` `a099_lease_events_test.go:242`가
  그 관계를 고정한다 — 두 값을 **읽어서** 비교하므로 상수를 바꾸면 그 테스트가
  같이 움직인다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **B4에 테스트가 없다.** `http.NewRequestWithContext` 실패 경로다.
  - **`HTTPClient`가 주입되면 `:123`의 두 번째 기한이 안 걸린다.**
    ctx의 기한만 남는다. **테스트가 주입하는 클라이언트에는 timeout이 없다** —
    실측(§5.7)이 그 조합을 잰다면 프로덕션과 다른 것을 재게 된다.
    a099가 만든 문제가 아니지만 **이름을 적어 둔다.**
  - **`AlertDeliveryBound`가 이 함수의 기한을 인자로 받는다.** 다른 `Publisher`가
    배선되면 유도가 깨진다 — **기본 배선 밖의 조합은 아무도 검사하지 않는다.**

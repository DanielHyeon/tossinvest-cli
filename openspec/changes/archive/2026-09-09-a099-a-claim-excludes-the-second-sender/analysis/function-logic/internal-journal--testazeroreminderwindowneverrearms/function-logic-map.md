# Function Logic Map: `TestAZeroReminderWindowNeverReArms`

- Source: `internal/journal/a096_claim_for_delivery_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> `remindAfter = 0`은 시간 기반 재무장을 끈다. 기록만 하는 호출자가 쓰는 값이고,
`EnqueueAlert`가 그것을 넘긴다.

a099는 반환 타입을 옮겼다. **0이 임차까지 끄는 것으로 확장되지 않았다** —
`EnqueueAlert`가 임차를 안 잡는 것은 값이 아니라 **호출의 부재**다(D13).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 저널 | 임시 디렉터리의 새 파일 | `outboxJournal` / 파일 상단의 헬퍼 | `Open` 실패면 `t.Fatalf` |
| 시계 | **주입된 가짜 시계** | `clock.NewFake` | 창·만료 판정이 실시간에 안 매인다 |
| 청구자 이름 | `testClaimant` | `a096_claim_for_delivery_test.go:33` | 배제는 이름이 아니라 토큰이 진다 |
| 원장 상태 | 이 함수가 직접 만든다 | 아래 「State mutations」 | 배치가 실패하면 단언에 도달하지 않는다 |

**불변식**: 이 함수의 모든 `t.Fatalf`는 **배치 실패**이고, `t.Errorf`는 **단언 실패**다.
둘을 섞으면 배치 오류가 단언 실패로 보고된다.

## Branches and early returns

AST 열거 — 분기 6 · 이탈 0 · 호출 12.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:285` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:288` | `if _, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:293` | `if got, err := j.ClaimAlertForDelivery(ctx, alert, 0, testClaimant); err != ni…` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:295` | `} else if got.Disposition != ClaimSettled {` | **창이 0이면 정산된 행은 `ClaimSettled`로 남는다** — 재무장이 없다 |
| B5 `:295` | `} else if got.Disposition != ClaimSettled {` | 같은 조건의 `else if` 짝 |
| B6 `:298` | `if got := alertState(t, j, claim.ID); got != AlertDelivered {` | **행이 `DELIVERED` 그대로다** — 0이 상태를 안 건드린다 |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`
- `j.MarkAlertDelivered`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **옮김뿐이다.**
- **D13이 이 값에 안 얹혔다**는 것이 a099의 결정이다 — 재무장 정책과 임차 정책을 한 값에 묶지 않았다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **음수 `remindAfter`를 안 본다.** `claimOwed`가 `<= 0`으로 같이 다루지만 이 테스트는 0만 준다.
  - **0이 상태 복구까지 끄지 않는다**는 것은 여기서 안 본다 — `TestRecoveringAnUnknownStateIgnoresTheReminderWindow` `a097_rearm_is_a_new_episode_test.go:214`가 본다.

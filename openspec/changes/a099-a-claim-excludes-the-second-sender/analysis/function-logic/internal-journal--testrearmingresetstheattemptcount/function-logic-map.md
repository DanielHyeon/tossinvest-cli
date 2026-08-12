# Function Logic Map: `TestReArmingResetsTheAttemptCount`

- Source: `internal/journal/a097_rearm_is_a_new_episode_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> 새 episode는 재시도 예산도 새로 받는다. 이전 episode에서 두 번 실패한
행이 재무장 뒤에 한 번만 남은 예산으로 시작하면 안 된다.

a097이 만든 테스트다. a099는 반환 타입과 **두 정산 호출의 토큰**을 옮겼다 —
B3의 실패 기록과 B4의 배달 기록이 **같은 claim의 토큰**을 쓴다.

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

AST 열거 — 분기 12 · 이탈 0 · 호출 19.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:96` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:103` | `for i := 0; i < 2; i++ {` | 두 번 실패시킨다 — 배치 |
| B3 `:104` | `if _, err := j.MarkAlertAttemptFailed(ctx, id, claim.Token, "transport down");…` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:108` | `if _, err := j.MarkAlertDelivered(ctx, id, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B5 `:111` | `if before, err := j.LookupAlert(ctx, id); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B6 `:113` | `} else if before.Attempts == 0 {` | **배치가 성립했다** — `attempts`가 0이 아니다 |
| B7 `:113` | `} else if before.Attempts == 0 {` | 같은 조건의 `else if` 짝 |
| B8 `:118` | `if again, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant…` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B9 `:120` | `} else if again.Disposition != ClaimAcquired {` | **재무장 뒤 `ClaimAcquired`다** |
| B10 `:120` | `} else if again.Disposition != ClaimAcquired {` | 같은 조건의 `else if` 짝 |
| B11 `:125` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B12 `:128` | `if got.Attempts != 0 {` | **`attempts`가 0으로 돌아왔다** |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`
- `j.LookupAlert`
- `j.MarkAlertAttemptFailed`
- `j.MarkAlertDelivered`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **이 테스트가 「한 claim 아래 여러 시도」를 간접적으로 고정한다.** B3의 반복이 같은 토큰을 쓰기 때문이다.
- **직접 고정하는 것은 `TestRecordingAFailedAttemptKeepsTheLease` `a099_regression_pins_test.go:28`이다.**
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **세 번째 시도(예산 소진)를 안 본다.** 예산 소진의 해제는 `TestASenderThatSpendsItsBudgetReleasesTheRow` `a099_lease_events_test.go:61`다.
  - **`last_error`가 재무장으로 지워지는지 안 본다.** 같은 UPDATE가 지운다.

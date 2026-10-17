# Function Logic Map: `TestClaimingADeliveredAlertPastTheWindowIsReArmed`

- Source: `internal/journal/a096_claim_for_delivery_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> 창이 지나면 행이 PENDING으로 돌아가고 알림이 첫 배달 경로를 다시 걷는다.
a099는 반환 타입을 옮기고 **단언 하나를 더했다**(B8): 재무장된 행의
`MarkAlertAttemptFailed`가 `SettleApplied`를 돌려줘야 한다.

**그 단언이 D13의 짝이다.** 재무장이 이전 episode의 임차를 안 지우면
새 청구의 토큰으로 정산이 안 되고 B8이 실패한다.

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

AST 열거 — 분기 8 · 이탈 0 · 호출 16.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:113` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:116` | `if _, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:123` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:126` | `if again.ID != claim.ID {` | 재알림은 같은 행이다 — 새 행이 아니다 |
| B5 `:129` | `if again.Disposition != ClaimAcquired {` | **창이 지나면 `ClaimAcquired`다** — 재알림은 반드시 나간다 |
| B6 `:132` | `if got := alertState(t, j, claim.ID); got != AlertPending {` | **상태가 PENDING으로 돌아왔다** — 첫 배달 경로를 걷는다 |
| B7 `:138` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B8 `:141` | `if res.Outcome != SettleApplied {` | **재무장된 행의 정산이 `SettleApplied`다** — 새 episode의 임차로 정산된다 (a099가 더한 단언) |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`
- `j.MarkAlertAttemptFailed`
- `j.MarkAlertDelivered`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **B8이 이 테스트에서 a099가 실제로 더한 계약이다.** 나머지는 옮김이다.
- 재무장이 임차를 지운다는 것을 **간접적으로** 확인한다 — 지우지 않으면 B8이 깨진다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **이전 episode의 토큰으로 정산을 시도하지 않는다.** 그것을 보는 것은 `TestReArmingClearsTheLeaseOfThePreviousEpisode` `a099_lease_lifecycle_test.go:340`다.
  - **창 경계 정확히(`== claimRemind`)를 본다.** 경계 바로 앞은 앞 테스트가 본다. 그 둘 사이에 다른 값은 안 본다.

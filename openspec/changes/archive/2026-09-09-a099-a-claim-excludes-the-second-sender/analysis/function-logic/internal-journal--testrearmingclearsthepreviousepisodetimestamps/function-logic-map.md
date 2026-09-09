# Function Logic Map: `TestReArmingClearsThePreviousEpisodeTimestamps`

- Source: `internal/journal/a097_rearm_is_a_new_episode_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> a097 초안은 `delivered_at`을 남겼다 — 「이전 episode가 정말 배달됐다」는 기록이라며.
그것이 틀렸다. body가 덮이는 순간, 살아남은 `delivered_at`은
**이 내용이 그때 배달됐다**고 주장하고, 실제로 배달된 내용은 아무 데도 없다.

행은 모든 열이 한 사건을 가리킬 때만 증거다. a099는 **그 목록에 임차 열 넷을 더했고**,
이유는 정확히 같다.

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

AST 열거 — 분기 12 · 이탈 0 · 호출 20.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:156` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:160` | `if _, err := j.MarkAlertDelivered(ctx, id, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:163` | `if before, err := j.LookupAlert(ctx, id); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:165` | `} else if before.DeliveredAt == nil \|\| before.LastAttemptAt == nil {` | **배치가 성립했다** — 두 도장이 찍혀 있다 |
| B5 `:165` | `} else if before.DeliveredAt == nil \|\| before.LastAttemptAt == nil {` | 같은 조건의 `else if` 짝 |
| B6 `:172` | `if again, err := j.ClaimAlertForDelivery(ctx, a097Alert("insufficient-quantity…` | 두 번째 episode를 **다른 이유로** 청구한다 — 오류 가드가 여러 줄에 걸쳐 있다 |
| B7 `:175` | `} else if again.Disposition != ClaimAcquired {` | **재무장 뒤 `ClaimAcquired`다** |
| B8 `:175` | `} else if again.Disposition != ClaimAcquired {` | 같은 조건의 `else if` 짝 |
| B9 `:180` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B10 `:183` | `if got.State != AlertPending {` | **상태가 PENDING이다** |
| B11 `:186` | `if got.DeliveredAt != nil {` | **`delivered_at`이 nil이다** |
| B12 `:191` | `if got.LastAttemptAt != nil {` | **`last_attempt_at`이 nil이다** |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`
- `j.LookupAlert`
- `j.MarkAlertDelivered`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **옮김뿐이다.**
- **a099의 임차 초기화가 이 테스트의 원칙에서 나왔다** — 「모든 열이 한 사건을 가리킨다」에 임차가 예외라고 말하는 규칙이 없다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **임차 열 넷은 이 표에 없다.** 같은 원칙이 같은 UPDATE에서 그것들도 지우지만, 그것을 보는 것은 `TestReArmingClearsTheLeaseOfThePreviousEpisode` `a099_lease_lifecycle_test.go:340`다. **원칙은 하나인데 테스트가 둘이다.**
  - **`created_at`은 안 지운다** — 조건이 처음 생긴 시각이라 episode에 속하지 않는다. 이 테스트가 그것을 **확인하지도 않는다.**

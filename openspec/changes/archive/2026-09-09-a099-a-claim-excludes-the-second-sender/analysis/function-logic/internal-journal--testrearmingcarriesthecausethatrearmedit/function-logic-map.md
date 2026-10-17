# Function Logic Map: `TestReArmingCarriesTheCauseThatReArmedIt`

- Source: `internal/journal/a097_rearm_is_a_new_episode_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> event key는 조건을 담고 원인은 안 담는다. 그래서 한 key의 두 번째 episode는
흔히 다른 이유다. `Flush`가 보내는 내용을 **행에서** 만들므로,
재무장은 title·body·payload를 새 것으로 덮어야 한다.

a097이 만든 테스트다. a099는 반환 타입과 정산 토큰을 옮겼다.

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

AST 열거 — 분기 9 · 이탈 0 · 호출 17.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:50` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:54` | `if _, err := j.MarkAlertDelivered(ctx, id, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:60` | `if again, err := j.ClaimAlertForDelivery(ctx, second, claimRemind, testClaiman…` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:62` | `} else if again.Disposition != ClaimAcquired {` | **두 번째 청구가 `ClaimAcquired`다** — 재무장됐다 |
| B5 `:62` | `} else if again.Disposition != ClaimAcquired {` | 같은 조건의 `else if` 짝 |
| B6 `:67` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B7 `:70` | `if got.Title != second.Title {` | **title이 두 번째 것이다** |
| B8 `:74` | `if got.Body != second.Body {` | **body가 두 번째 것이다** |
| B9 `:77` | `if got.Payload != second.Payload {` | **payload가 두 번째 것이다** |

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
- 이 테스트가 지키는 것은 **내용의 일치**이고, a099가 더한 것은 **소유의 일치**다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **`payload`를 보는 유일한 테스트가 이것이다.** 위의 `TestReArmingClearsThePreviousAcknowledgement`가 안 보는 열을 여기가 본다.
  - **임차가 새 episode의 것인지는 안 본다** — `a099_lease_lifecycle_test.go:340`.

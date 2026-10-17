# Function Logic Map: `TestClaimingAnAcknowledgedAlertFollowsTheSameWindow`

- Source: `internal/journal/a096_claim_for_delivery_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> 운영자의 확인은 그 episode의 빚을 끝내되 영원히 끝내지는 않는다.
a099는 반환 타입을 옮겼고, **주석을 더했다**: 확인이 **임차를 든 채로** 일어난다.

그것은 우연이 아니라 설계다 — 사람이 봤다는 사실이
기계가 말하는 중이라는 사실보다 우선한다. `AcknowledgeAlert`는 임차를 안 본다.

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

AST 열거 — 분기 8 · 이탈 0 · 호출 14.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:157` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:164` | `if err := j.AcknowledgeAlert(ctx, claim.ID, "operator"); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:169` | `if got, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant);…` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:171` | `} else if got.Disposition != ClaimSettled {` | **창 안에서는 `ClaimSettled`다** — 확인 뒤 재발생은 아직 뉴스가 아니다 |
| B5 `:171` | `} else if got.Disposition != ClaimSettled {` | 같은 조건의 `else if` 짝 — AST가 `else`와 `if`를 따로 센다 |
| B6 `:176` | `if got, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant);…` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B7 `:178` | `} else if got.Disposition != ClaimAcquired {` | **창이 지나면 `ClaimAcquired`다** — 확인 뒤의 재발생은 뉴스다 |
| B8 `:178` | `} else if got.Disposition != ClaimAcquired {` | 같은 조건의 `else if` 짝 |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.AcknowledgeAlert`
- `j.ClaimAlertForDelivery`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **옮김뿐이다.** a099가 이 테스트에 계약을 안 더했다.
- **주석은 더했다** — 확인이 임차 위에서 일어나는 것이 설계된 경로임을 적는다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **확인이 임차를 지우는지 안 지우는지를 이 테스트가 안 본다.** 그것은 `TestAcknowledgementIgnoresTheLease` `a099_regression_pins_test.go:101`다.
  - **확인한 사람의 이름을 안 본다.** `outbox_test.go`의 `TestDeliveryAndAcknowledgementAreDistinctStates`가 본다.

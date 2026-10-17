# Function Logic Map: `TestDeliveryAndAcknowledgementAreDistinctStates`

- Source: `internal/journal/outbox_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> 배달과 확인은 다른 상태다. 기계가 보낸 것과 사람이 본 것은 같지 않다.

a099는 정산 호출에 토큰을 넘기게 했다(B2). **단언은 그대로다.**

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

AST 열거 — 분기 7 · 이탈 0 · 호출 16.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:123` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:129` | `if _, err := j.MarkAlertDelivered(ctx, delivered, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:134` | `if err := j.AcknowledgeAlert(ctx, acked, "operator"); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:139` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B5 `:142` | `if count != 0 {` | **미전달 수가 0이다** — 둘 다 PENDING을 떠났다 |
| B6 `:147` | `if a.State != AlertDelivered \|\| a.DeliveredAt == nil {` | **배달된 행은 `DELIVERED`이고 `delivered_at`이 있다** |
| B7 `:151` | `if b.State != AlertAcknowledged \|\| b.AcknowledgedBy != "operator" {` | **확인된 행은 `ACKNOWLEDGED`이고 확인자 이름이 있다** |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.AcknowledgeAlert`
- `j.ClaimAlertForDelivery`
- `j.EnqueueAlert`
- `j.LookupAlert`
- `j.MarkAlertDelivered`
- `j.UndeliveredCount`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **옮김뿐이다.**
- **`AcknowledgeAlert`의 시그니처가 안 바뀐 것이 a099의 결정이다** — 사람의 확인은 임차 위에 있다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **확인이 임차를 무시하는 것을 안 본다** — `TestAcknowledgementIgnoresTheLease` `a099_regression_pins_test.go:101`.
  - **`AcknowledgeAlert`는 토큰을 안 받는다.** 그것이 설계이고, 이 테스트가 그 사실을 **암묵적으로만** 보인다 — 인자가 없으니까.

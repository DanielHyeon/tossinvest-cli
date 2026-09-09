# Function Logic Map: `TestClaimingADeliveredAlertInsideTheWindowOwesNothing`

- Source: `internal/journal/a096_claim_for_delivery_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> a096의 폭풍 대책을 고정하는 테스트다. a099는 반환 타입만 옮겼다 —
`owed=false`가 `ClaimSettled`가 됐다. **뜻은 그대로다**: 창 안의 재발생은
운영자에게 두 번 말하지 않는다.

정산 호출도 토큰을 제시하게 바뀌었다(B3).

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

AST 열거 — 분기 7 · 이탈 0 · 호출 14.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:71` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:74` | `if claim.Disposition != ClaimAcquired {` | 첫 청구가 `ClaimAcquired`다 — 배치가 성립했다 |
| B3 `:77` | `if _, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:84` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B5 `:87` | `if again.ID != claim.ID {` | 같은 조건은 같은 행이다 — id가 안 바뀐다 |
| B6 `:90` | `if again.Disposition != ClaimSettled {` | **창 안에서는 `ClaimSettled`다** — 옛 `owed=false`. 운영자는 방금 들었다 |
| B7 `:94` | `if got := alertState(t, j, claim.ID); got != AlertDelivered {` | **억제된 청구가 행을 안 건드린다** — 상태가 `DELIVERED` 그대로 |

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

- **옮김의 뜻이 보존됐다**: `owed=false` → `ClaimSettled`.
- 이 테스트가 지키는 것은 **억제**이고, a099가 더한 것은 **배제**다. 둘은 다른 축이다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **임차 열을 안 본다.** 정산된 행에 임차가 안 남는지는 `TestClaimingADeliveredAlertInsideTheWindowOwesNothing`이 아니라 `MarkAlertDelivered`의 UPDATE가 진다.
  - **`B5`가 실패해도 `B6`가 돈다** (`t.Errorf`). 두 단언이 독립이라 출력이 둘 다 뜰 수 있다 — 의도된 것이다.

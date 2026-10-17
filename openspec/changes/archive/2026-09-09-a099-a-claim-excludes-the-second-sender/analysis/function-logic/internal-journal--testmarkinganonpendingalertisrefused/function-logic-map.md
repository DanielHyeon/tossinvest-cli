# Function Logic Map: `TestMarkingANonPendingAlertIsRefused`

- Source: `internal/journal/outbox_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> PENDING이 아닌 행의 정산은 거절된다. a096이 만든 CAS이고,
**a099가 그 거절을 오류에서 타입 있는 결과로 바꿨다.**

`ErrAlertNotFound` 하나였던 답이 `SettleAlreadySettled`와 `SettleNotFound`로 갈렸다.
호출자가 둘을 다르게 다뤄야 하기 때문이다 — 전자는 정상, 후자는 이상.

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

AST 열거 — 분기 8 · 이탈 0 · 호출 13.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:179` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:182` | `if res, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:184` | `} else if res.Outcome != SettleApplied {` | 첫 정산이 `SettleApplied`다 — 배치 |
| B4 `:184` | `} else if res.Outcome != SettleApplied {` | 같은 조건의 `else if` 짝 |
| B5 `:189` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B6 `:192` | `if res.Outcome != SettleAlreadySettled {` | **두 번째 정산이 `SettleAlreadySettled`다** — 옛 `ErrAlertNotFound`의 절반 |
| B7 `:197` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B8 `:200` | `if missing.Outcome != SettleNotFound {` | **없는 id는 `SettleNotFound`다** — 나머지 절반 (a099가 가른 것) |

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

- **a099가 이 테스트에서 옛 오류 하나를 결과 둘로 갈랐다.** 가른 이유가 호출자에 있다: `deliver`가 `SettleNotFound`를 게이트 래치로 보낸다.
- **네 결과 중 셋을 여기가, 넷째를 lifecycle 테스트가 본다.**
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **`SettleLeaseLost`를 이 테스트가 안 본다.** 네 결과 중 셋만 본다 — 임차 상실은 `TestTheSameSenderNameCannotSettleAnEarlierLease` `a099_lease_lifecycle_test.go:266`가 본다.
  - **토큰이 맞는데 상태가 틀린 경우와 상태가 맞는데 토큰이 틀린 경우를 따로 안 본다.** 여기는 앞의 것만이다.

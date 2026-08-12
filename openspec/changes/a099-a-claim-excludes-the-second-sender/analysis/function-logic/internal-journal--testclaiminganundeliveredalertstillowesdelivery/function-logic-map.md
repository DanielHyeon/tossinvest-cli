# Function Logic Map: `TestClaimingAnUndeliveredAlertStillOwesDelivery`

- Source: `internal/journal/a096_claim_for_delivery_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> **a099가 옛 bool을 실제로 둘로 가른 유일한 자리다.**

임차 전에는 두 번째 청구가 `owed=true`로 돌아왔고 답은 거기서 끝났다.
지금 그 행은 **아직 미정산이면서 동시에 남이 쥐고 있다**:
실패 기록은 임차를 일부러 유지한다(발송자에게 재시도가 남았다).
두 시도 사이에 도착한 두 번째 발송자는 보내면 안 되고,
그것이 `ClaimHeldElsewhere`가 말하는 바다.

**안 바뀐 것**은 이 테스트가 존재하는 이유다: 실패한 발송이 성공한 발송으로
오인되지 않는다.

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

AST 열거 — 분기 11 · 이탈 0 · 호출 18.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:204` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:208` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:211` | `if res.Outcome != SettleApplied {` | 실패 기록이 `SettleApplied`다 — **임차를 유지한 채로** 기록됐다 |
| B4 `:214` | `if got := alertState(t, j, claim.ID); got != AlertPending {` | 행이 여전히 PENDING이다 — 실패는 배달이 아니다 |
| B5 `:219` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B6 `:222` | `if again.Disposition == ClaimSettled {` | **두 번째 청구가 `ClaimSettled`가 아니다** — 빚은 그대로다 (옛 `owed=true`의 절반) |
| B7 `:225` | `if again.Disposition != ClaimHeldElsewhere {` | **두 번째 청구가 `ClaimHeldElsewhere`다** — 남이 쥐고 있다 (a099가 가른 나머지 절반) |
| B8 `:232` | `if _, err := j.ReleaseAlertClaim(ctx, claim.ID, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B9 `:235` | `if got, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant);…` | 해제 뒤 다시 청구하면 잡힌다 |
| B10 `:237` | `} else if got.Disposition != ClaimAcquired {` | **해제 뒤 `ClaimAcquired`다** — 임차를 놓으면 빚이 다시 잡힌다 |
| B11 `:237` | `} else if got.Disposition != ClaimAcquired {` | 같은 조건의 `else if` 짝 |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`
- `j.MarkAlertAttemptFailed`
- `j.ReleaseAlertClaim`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **이 테스트에서만 옛 단언의 뜻이 갈렸다.** 그 사실이 파일의 doc comment와 이 함수 바로 위 주석에 적혀 있다 — 조용히 바꾸지 않았다.
- **B6과 B7이 짝이다**: 「빚은 남았다」와 「내 것이 아니다」를 따로 단언한다. 하나로 합치면 갈라진 뜻이 다시 붙는다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **만료로 풀리는 경로를 안 본다.** 여기서는 `ReleaseAlertClaim`으로 명시 해제한다. 만료는 `TestAnExpiredClaimIsPickedUpByAnotherSender` `a099_lease_lifecycle_test.go:70`다.
  - **두 발송자가 정말 동시에 들어오는 것을 안 본다** — 순차다. 동시성은 `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend` `a099_claim_excludes_the_second_sender_test.go:41`가 진다.

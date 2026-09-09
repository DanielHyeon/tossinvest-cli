# Function Logic Map: `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend`

- Source: `internal/journal/a099_claim_excludes_the_second_sender_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> **a099가 존재하는 이유를 관측하는 테스트다.**

한 PENDING 행에 발송자 둘이 동시에 닿으면 발송권은 정확히 하나여야 한다.
구현 전 관측: `senders granted the right to send = 2, want 1`.

배치는 `EnqueueAlert`로 한다(B1) — **기록은 임차를 안 잡으므로**
두 경합자 중 누구도 배치 때문에 지지 않는다. 그것 자체가 D13의 확인이다.

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

AST 열거 — 분기 8 · 이탈 0 · 호출 20.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:50` | `if _, err := j.EnqueueAlert(ctx, a099Alert()); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:53` | `if got := alertState(t, j, 1); got != AlertPending {` | 배치된 행이 PENDING이고 id가 1이다 — 경합 전 상태 |
| B3 `:65` | `for i := 0; i < senders; i++ {` | 발송자 여럿을 동시에 띄운다 |
| B4 `:73` | `switch {` | 각 결과를 세 갈래로 분류하는 `switch` |
| B5 `:74` | `case err != nil:` | 오류는 따로 모은다 — 배제가 아니라 고장이다 |
| B6 `:76` | `case claim.Disposition == ClaimAcquired:` | **`ClaimAcquired`를 센다** — 발송권의 수 |
| B7 `:84` | `for _, err := range failed {` | 모은 오류를 보고한다 |
| B8 `:87` | `if granted != 1 {` | **발송권이 정확히 하나다** — a099의 전부 |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`
- `j.EnqueueAlert`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **이 테스트가 a099의 RED다.** 나머지 열넷은 그 주변을 고정한다.
- **배치를 `EnqueueAlert`로 하는 것이 설계의 일부다.** 청구로 배치하면 그 청구가 임차를 잡아 **두 경합자 모두 배치에게 진다** — 아무것도 관측 못 한다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **`-race`로 안 돌렸다.** 이 테스트가 관측하는 배타성은 SQLite의 잠금 모형에 기댄다. §5.2가 관측하고, **관측 전에는 성립한다고 적지 않는다.**
  - **발송자 수를 하나만 쓴다.** 셋 이상으로 늘리면 다른 것이 보일 수 있지만 안 봤다.
  - **어느 발송자가 이기는지는 안 본다** — 정해져 있지 않고 정해질 필요도 없다.

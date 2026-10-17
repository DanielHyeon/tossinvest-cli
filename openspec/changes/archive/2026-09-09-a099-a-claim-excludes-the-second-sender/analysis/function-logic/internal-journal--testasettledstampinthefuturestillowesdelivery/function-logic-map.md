# Function Logic Map: `TestASettledStampInTheFutureStillOwesDelivery`

- Source: `internal/journal/a096b_round2_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> 시계가 뒤로 갔거나 미래의 도장이 찍힌 행도 **결국은 보내야 한다**.
a096 2라운드가 만든 테스트이고, a099는 반환 타입을 옮겼다.

**임차 쪽에도 같은 문제의 짝이 있다** — `claimed_at`이 미래면 술어가 행을 다시 연다
(`alertClaimSkew`). **두 판정은 서로 다른 자리이고 값도 다르다.**

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

AST 열거 — 분기 6 · 이탈 0 · 호출 16.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:37` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:41` | `if _, err := j.MarkAlertDelivered(ctx, id, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:48` | `if _, err := j.db.ExecContext(ctx,` | 테스트가 직접 SQL로 미래의 `delivered_at`을 심는다 — 배치 |
| B4 `:54` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B5 `:57` | `if skewed.Disposition != ClaimAcquired {` | **미래 도장이어도 `ClaimAcquired`다** — 보낸 적 없는 것으로 다룬다 |
| B6 `:61` | `if got := alertState(t, j, id); got != AlertPending {` | **상태가 PENDING으로 돌아왔다** |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`
- `j.MarkAlertDelivered`
- `j.clk.Now`
- `j.db.ExecContext`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **옮김뿐이다.**
- **a099가 같은 종류의 판정을 하나 더 만들었다**는 것이 이 번들의 소득이다 — 두 자리가 갈라질 수 있다는 사실을 이름으로 적어 둔다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **임차의 미래 도장과 이 테스트는 다른 판정이다.** 임차 쪽은 `TestAClaimIssuedInTheFutureIsReopened` `a099_lease_lifecycle_test.go:119`이고, **`alertClaimSkew`(2초)라는 여유가 거기에만 있다.** 이 테스트의 판정에는 여유가 없다.
  - **두 판정이 같은 방향인지 아무도 검사하지 않는다.** 둘 다 「열어 둔다」이지만 그 일치는 우연이 아니라 설계이고, **설계를 고정하는 테스트가 없다.**

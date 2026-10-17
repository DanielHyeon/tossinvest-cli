# Function Logic Map: `TestAnUnrecognisedAlertStateOwesDelivery`

- Source: `internal/journal/a096_claim_for_delivery_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> 스키마에 `CHECK`가 없으므로 `state`에 알 수 없는 값이 들어갈 수 있다.
그때 원장은 **열린 쪽으로 실패한다** — 보내야 한다고 답한다.

a099는 반환 타입을 옮겼고, 정산 호출에 토큰을 넘기게 했다(B6).
**폴백의 방향은 안 바꿨다.**

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

AST 열거 — 분기 6 · 이탈 0 · 호출 13.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:252` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:255` | `if _, err := j.db.ExecContext(ctx,` | 테스트가 직접 SQL로 알 수 없는 상태를 심는다 — 배치 |
| B3 `:261` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:264` | `if again.Disposition != ClaimAcquired {` | **알 수 없는 상태는 `ClaimAcquired`다** — 열린 쪽 실패 |
| B5 `:268` | `if got := alertState(t, j, claim.ID); got != AlertPending {` | **상태가 PENDING으로 복구됐다** — 다음 관측이 정상 경로를 걷는다 |
| B6 `:271` | `if _, err := j.MarkAlertDelivered(ctx, claim.ID, again.Token); err != nil {` | 복구된 행이 새 토큰으로 정산된다 — 임차가 복구와 함께 갱신됐다 |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`
- `j.MarkAlertDelivered`
- `j.db.ExecContext`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **폴백 방향이 임차 술어와 같다** — 모르면 열어 둔다.
- **옮김뿐이다.** a099가 이 테스트의 계약을 안 바꿨다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **어떤 알 수 없는 값인지는 하나만 본다.** 값의 종류에 따라 갈리는 것은 없다 — `claimOwed`의 `default`가 전부를 같이 다룬다.
  - **`remindAfter`가 0일 때 이 복구가 여전히 도는지**는 여기가 아니라 `TestRecoveringAnUnknownStateIgnoresTheReminderWindow` `a097_rearm_is_a_new_episode_test.go:214`가 본다.

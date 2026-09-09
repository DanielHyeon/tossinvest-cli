# Function Logic Map: `TestReArmingClearsThePreviousAcknowledgement`

- Source: `internal/journal/a096b_round2_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> 재무장은 새 episode다. 이전 episode의 확인이 남으면
행이 「미전달」과 「daniel이 확인함」을 동시에 말한다 —
사고 뒤 백로그를 훑는 운영자가 **자기 이름을 믿고 살아 있는 critical 알림을 건너뛴다.**

a096 2라운드가 만든 테스트다. a099는 반환 타입을 옮겼다.

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

AST 열거 — 분기 9 · 이탈 0 · 호출 15.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:83` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:89` | `if err := j.AcknowledgeAlert(ctx, id, "daniel"); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:94` | `if again, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant…` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:96` | `} else if again.Disposition != ClaimAcquired {` | **창이 지난 뒤 `ClaimAcquired`다** — 확인된 행도 재무장된다 |
| B5 `:96` | `} else if again.Disposition != ClaimAcquired {` | 같은 조건의 `else if` 짝 |
| B6 `:101` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B7 `:104` | `if got.State != AlertPending {` | **상태가 PENDING이다** |
| B8 `:108` | `if got.AcknowledgedBy != "" {` | **`acknowledged_by`가 비었다** — 이전 확인자가 안 남는다 |
| B9 `:113` | `if got.AcknowledgedAt != nil {` | **`acknowledged_at`이 nil이다** |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.AcknowledgeAlert`
- `j.ClaimAlertForDelivery`
- `j.LookupAlert`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **옮김뿐이다.**
- **a099가 같은 UPDATE에 임차 열 넷을 더했다** — 이 테스트의 세 단언과 정확히 같은 이유로.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **임차 열 넷이 지워지는지는 이 테스트가 안 본다.** 같은 UPDATE가 지우지만 그것을 보는 것은 `TestReArmingClearsTheLeaseOfThePreviousEpisode` `a099_lease_lifecycle_test.go:340`다.
  - **재무장 UPDATE가 지우는 열 열둘 중 셋만 본다.** 나머지는 a097의 테스트들이 나눠 진다 — **`payload`는 아무도 안 본다.**

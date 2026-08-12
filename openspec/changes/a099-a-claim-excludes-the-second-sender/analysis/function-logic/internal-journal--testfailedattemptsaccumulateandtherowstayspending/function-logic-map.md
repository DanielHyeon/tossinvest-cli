# Function Logic Map: `TestFailedAttemptsAccumulateAndTheRowStaysPending`

- Source: `internal/journal/outbox_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> critical 알림은 배달에 실패해도 버려지지 않는다. 시도가 쌓이고 행은 PENDING으로 남는다.

**a099가 이 테스트에 계약을 하나 더했다**: 세 시도가 **한 claim 아래**서 일어난다.
시도 사이에 임차를 잃는 발송자는 남은 재시도를 남의 발송에 넘기는 것이므로,
임차는 예산 내내 일부러 유지된다(B4).

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

AST 열거 — 분기 9 · 이탈 0 · 호출 14.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:80` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:87` | `for i := 0; i < 3; i++ {` | 세 번 실패시킨다 — 한 claim 아래서 |
| B3 `:90` | `if ferr != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:93` | `if res.Outcome != SettleApplied {` | **매 시도의 정산이 `SettleApplied`다** — 임차가 예산 내내 유지된다 (a099가 더한 단언) |
| B5 `:100` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B6 `:103` | `if alert.State != AlertPending {` | **상태가 PENDING이다** — critical 알림은 보존된다 |
| B7 `:106` | `if alert.Attempts != 3 {` | **`attempts`가 3이다** |
| B8 `:109` | `if alert.LastAttemptAt == nil {` | **`last_attempt_at`이 기록됐다** |
| B9 `:112` | `if alert.LastError != "transport is down" {` | **`last_error`가 마지막 이유다** |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`
- `j.LookupAlert`
- `j.MarkAlertAttemptFailed`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **B4가 a099가 더한 계약이다.** 나머지는 옮김이다.
- **주석이 그 이유를 코드 안에 적는다** — 되돌리려는 사람이 문서를 안 봐도 읽는다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **예산이 다 떨어졌을 때의 해제를 안 본다.** 이 테스트는 세 번 기록만 하고 `deliver`의 루프를 안 탄다.
  - **시도 사이에 다른 발송자가 끼어드는 것을 안 본다.** `TestClaimingAnUndeliveredAlertStillOwesDelivery` `a096_claim_for_delivery_test.go:197`이 그 자리를 본다.

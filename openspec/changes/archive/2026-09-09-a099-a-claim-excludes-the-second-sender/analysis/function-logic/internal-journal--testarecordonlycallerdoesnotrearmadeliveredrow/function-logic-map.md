# Function Logic Map: `TestARecordOnlyCallerDoesNotReArmADeliveredRow`

- Source: `internal/journal/a097_rearm_is_a_new_episode_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> 기록만 하는 호출자는 남의 몫으로 정산된 행을 재무장하면 안 된다.
`EnqueueAlert`가 `remindAfter = 0`을 넘기는 이유다.

**a099가 이 경로를 통째로 바꿨다**: `EnqueueAlert`는 더 이상
`ClaimAlertForDelivery`에 위임하지 않고 자기 트랜잭션에서 `recordAlertTx`를 부른다.
**단언은 그대로다** — 정산된 행은 정산된 채로 남는다.

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

AST 열거 — 분기 5 · 이탈 0 · 호출 13.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:255` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:259` | `if _, err := j.MarkAlertDelivered(ctx, id, claim.Token); err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B3 `:267` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B4 `:270` | `if again.Disposition != ClaimSettled {` | **`EnqueueAlert` 뒤에도 `ClaimSettled`다** — 기록이 재무장을 안 만들었다 |
| B5 `:274` | `if got := alertState(t, j, id); got != AlertDelivered {` | **상태가 `DELIVERED` 그대로다** |

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

- **이 테스트의 뜻은 그대로고 그 아래 구현은 통째로 바뀌었다.** 그 조합이 a099에서 가장 좋은 신호다 — 계약이 구현과 독립이라는 증거.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **`EnqueueAlert`가 임차를 안 잡는 것을 직접 단언하지 않는다.** 여기서는 결과(정산 유지)로만 본다. 직접 보는 것은 `TestARecordedAlertCanBeClaimedImmediately` `a099_lease_lifecycle_test.go:305`다.

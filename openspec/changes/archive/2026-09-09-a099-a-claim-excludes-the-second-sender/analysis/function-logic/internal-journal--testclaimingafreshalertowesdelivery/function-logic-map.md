# Function Logic Map: `TestClaimingAFreshAlertOwesDelivery`

- Source: `internal/journal/a096_claim_for_delivery_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **이것은 테스트 함수의 Function Logic Map이다.** `check_analysis.py`가
> 요구하는 이유는 a099의 diff가 이 함수의 본문을 바꿨기 때문이다 —
> 테스트 함수도 「수정된 기존 함수」다.
>
> a099가 `ClaimAlertForDelivery`의 반환을 `(id, owed, error)`에서
`(ClaimResult, error)`로 바꿨다. 이 테스트의 본문이 그것을 따라 옮겨졌고,
**단언 하나가 늘었다** — 취득한 claim은 토큰을 들고 있어야 한다(B4).
`owed=true`가 `ClaimAcquired`가 된 것은 이름만 바뀐 것이지만,
**토큰 단언은 a099가 새로 만든 계약이다.**

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

AST 열거 — 분기 4 · 이탈 0 · 호출 7.

| Branch | 소스 | 뜻 |
|---|---|---|
| B1 `:46` | `if err != nil {` | 호출이 오류를 내면 즉시 중단 — 배치 실패는 단언이 아니다 |
| B2 `:49` | `if claim.ID == 0 {` | 삽입된 행의 id가 0이 아니다 — 기록이 실제로 일어났다 |
| B3 `:52` | `if claim.Disposition != ClaimAcquired {` | **처분이 `ClaimAcquired`다** — 방금 삽입된 행은 보낸 적이 없다 (옛 `owed=true`) |
| B4 `:56` | `if claim.Token == "" {` | **토큰이 비어 있지 않다** — 없으면 이 발송을 정산할 수단이 없다 (a099가 더한 단언) |

## Calls and live bindings

이 테스트가 부르는 원장 API:

- `j.ClaimAlertForDelivery`

**live binding**: 테스트 함수이므로 프로덕션 호출자가 없다.
이 함수를 부르는 것은 `go test`뿐이다.

## State mutations and fallbacks

- **임시 저널 파일 하나.** `t.TempDir()` 아래이고 `t.Cleanup`이 닫는다.
- **프로세스 밖 side effect 없음.** 네트워크도 실계좌도 안 건드린다.
- **폴백 없음.** 배치가 실패하면 `t.Fatalf`로 끝난다.

## Safety conclusion

- **옮김의 뜻이 보존됐다**: `owed=true` → `ClaimAcquired`. 같은 상황에서 같은 판정을 요구한다.
- **늘어난 단언은 새 계약이다** — a099 이전에는 정산할 토큰이라는 것이 없었다.
- **High-risk impact**: **no (직접)** — 테스트 함수다. 다만 **이 함수가 지키는
  계약은 High-risk다**: critical 알림의 배달과 정산이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **토큰의 내용은 안 본다.** 비어 있지 않다는 것만 본다. 엔트로피·유일성은 `mintAlertClaimToken`의 몫이고 이 테스트가 안 진다.
  - **두 번째 청구를 안 한다.** 이 테스트는 배제를 안 본다 — 그것은 `TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend`다.

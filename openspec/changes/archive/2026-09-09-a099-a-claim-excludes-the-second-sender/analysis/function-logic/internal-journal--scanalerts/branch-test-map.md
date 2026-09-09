# Branch Test Map: `scanAlerts`

`ast.json`의 열거가 정본이다: 분기 3 · 이탈 3.
**a099가 이 함수에 더한 것은 분기가 아니라 대입 셋이다** (`:576`~`:578`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:552` 행마다 돈다 | `TestEnqueueAlertIsIdempotentOnTheEventKey` `outbox_test.go:30` · `TestPendingAlertsProjectTheLeaseButNeverTheToken` `a099_lease_lifecycle_test.go:391` | no | **yes** |
| B2 | `:565` `rows.Scan` 실패 | 없음 — **컴파일러가 안 잡는 짝이다** | no | **no** |
| B3 | `:581` `rows.Err()` | 없음 | no | **no** |

이탈 `:584`(정상)은 분기가 아니다.

## a099가 더한 대입 셋

| 대입 | 열 | Test |
|---|---|---|
| `:576` `a.ClaimedBy = claimedBy` | `claimed_by` | `TestPendingAlertsProjectTheLeaseButNeverTheToken` `a099_lease_lifecycle_test.go:391` |
| `:577` `a.ClaimedAt = parseOptionalStamp(claimedAt)` | `claimed_at` | 같음 |
| `:578` `a.ClaimExpiresAt = parseOptionalStamp(claimExpiry)` | `claim_expires_at` | 같음 |
| **없음** | **`claim_token`** | 같은 테스트가 **부재**를 본다 |

**RED observed**: **yes — §4.2 되돌림 관측.** `alertSelect`에 세 열을 안 더한
상태로 되돌리면 `Alert.ClaimedBy`가 비어 있고 그 테스트가 실패한다.
`claim_token`을 **더하면** 같은 테스트가 다른 이유로 실패한다 —
그것이 「부재」를 고정하는 방식이다.

## B2가 왜 위험한 공백인가

`alertSelect`(`outbox.go:497-501`)의 열 목록과 `rows.Scan`(`:565-567`)의
인자 목록은 **한 쌍이고 컴파일러가 짝을 안 본다.** 한쪽만 바꾸면
런타임에 B2가 오류를 낸다.

**a099는 그 쌍을 동시에 바꿨고**, `TestPendingAlertsProjectTheLeaseButNeverTheToken`이
결과를 읽어서 확인한다. **하지만 B2 자체(오류 경로)는 여전히 안 덮여 있다** —
쌍이 어긋난 상태를 만드는 테스트가 없다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **B2 · B3** — 드라이버·커서 오류 경로.
  **`not-applicable`: 이 change는 둘을 근거로 아무것도 주장하지 않는다.**

## 덮이지 않은 것을 이름으로 적는다

- **`Alert` 구조체가 커진 것의 하류 영향을 이 change에서 확인 안 했다.**
  이 구조체를 JSON으로 내보내는 경로가 있으면 표면이 셋 늘었다.
  `claim_token`이 없으므로 **유출 위험은 없다.**
- **빈 결과가 `nil, nil`이다.** `out`이 nil로 시작하고 append가 없으면 nil이다.
  호출자가 `len() == 0`으로 다루므로 문제가 아니지만, **a099가 만든 것이 아니고
  이 표에도 행이 없다** — 분기가 아니기 때문이다.

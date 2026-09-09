# Branch Test Map: `Ntfy.Publish`

`ast.json`의 열거가 정본이다: 분기 9 · 이탈 5.
**a099가 이 함수에서 바꾼 것은 리터럴 하나다** (`10 * time.Second` →
`DefaultPublishTimeout`, `:97`). **아래 표는 전부 인용이고 RED는 하나도 없다.**

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:87` topic이 공백 → `ErrNtfyNotConfigured` | 기존 | no | **yes (기존)** |
| B2 | `:91` base URL이 공백 → 기본값 | 기존 | no | **yes (기존)** |
| B3 | `:96` **timeout이 0 → `DefaultPublishTimeout`** | 기존 (주입 쪽) | no | **부분 — 기본값 쪽 미검증** |
| B4 | `:104` 요청 생성 실패 | 없음 | no | **no** |
| B5 | `:110` title이 있으면 헤더 | 기존 | no | **yes (기존)** |
| B6 | `:117` token이 있으면 `Authorization` | 기존 | no | **yes (기존)** |
| B7 | `:122` `HTTPClient == nil`이면 새로 만든다 | 기존 | no | **yes (기존)** |
| B8 | `:126` `client.Do` 실패 | 기존 | no | **yes (기존)** |
| B9 | `:134` 상태가 `2xx`가 아니다 | 기존 | no | **yes (기존)** |

이탈 `:138`(정상)은 분기가 아니다.

## a099가 이 함수에서 실제로 한 일

| 무엇 | 값 | 분기 | 이탈 | defer |
|---|---|---|---|---|
| base | `10 * time.Second` | 9 | 5 | 2 |
| 지금 | `DefaultPublishTimeout` = 10초 | 9 | 5 | 2 |

**넷 다 같다.** 바뀐 것은 **그 숫자를 패키지 밖에서 읽을 수 있게 된 것**뿐이다.

## 왜 그것이 필요했나

`obs.AlertDeliveryBound`(`alert_lease.go:39`)가 배달 상한을 계산하고,
`journal.DefaultAlertLease`(81초)가 그 상한(54초)을 넘어야 한다.
publish timeout이 이 함수 안의 리터럴로 남아 있으면 **유도가 그 숫자를 복사한다.**

| 복사하면 | 결과 |
|---|---|
| 이 함수의 timeout이 15초로 바뀐다 | 유도는 여전히 10초를 쓴다 |
| 상한은 54초라고 계산된다 | **실제는 69초다** |
| 임차는 81초 | **여유가 12초로 줄어든다 — 아무도 모른 채로** |

`TestTheDefaultLeaseOutlastsTheDeliveryBound` `a099_lease_events_test.go:242`가
**두 값을 다 읽어서** 비교하는 이유가 그것이다. 숫자를 테스트에 베끼면
그 테스트는 아무것도 안 지킨다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

- **아홉 전부.** a099는 이 함수의 어떤 판정도 안 바꿨다.
  **`not-applicable`: 이 change는 이 함수의 분기를 근거로 아무것도 주장하지 않는다.**

## 덮이지 않은 것을 이름으로 적는다

- **B4에 테스트가 없다.** `http.NewRequestWithContext` 실패 경로.
- **B3의 「기본값이 쓰인다」쪽이 직접 단언되지 않는다.** 테스트는 timeout을 주입한다.
  `DefaultPublishTimeout`이 `AlertDeliveryBound`의 기본 인자로 쓰이므로
  **간접적으로는** `TestTheDefaultLeaseOutlastsTheDeliveryBound`가 그 값을 읽지만,
  **이 함수가 그 상수를 실제로 쓰는지는 그 테스트가 안 본다.**
- **주입된 `HTTPClient`에는 timeout이 없다.** `:123`의 두 번째 기한이 그때 사라지고
  ctx의 기한만 남는다. §5.7 실측이 그 조합을 재면 **프로덕션과 다른 것을 잰다.**

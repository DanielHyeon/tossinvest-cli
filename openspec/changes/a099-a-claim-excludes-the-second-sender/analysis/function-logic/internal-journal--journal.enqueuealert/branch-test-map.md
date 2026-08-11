# Branch Test Map: `Journal.EnqueueAlert`

**GREEN 칸은 실측해서 채운다.**
`ast.json`의 열거가 정본이다: **분기 0** · 이탈 1.

**a099는 이 함수의 본문에 분기를 만들지 않는다.** 아래 항목은 위임 계약을 고정한다.

**분기가 없으므로 행은 유일한 이탈 `:121`(happy path) 하나다.**
아래 표의 `B1`은 AST의 분기가 아니라 **그 happy path를 가리키는 이름**이다
(`check_analysis`가 분기 없는 함수에 요구하는 형식).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기록만 하는 호출이 **발송 권한을 잡지 않는다** | **a099 R16** | **planned RED — 미관측** | 미관측 |
| 〃 이탈 `:121` | 같은 `event_key`를 두 번 기록하면 같은 id | `internal/journal/outbox_test.go` (기존) | — | **yes (기존)** |

## R16의 RED 조건 — 무엇이 실패해야 하는가

R16은 **오늘 실행하면 통과한다**(임차 자체가 없으므로). 그러므로 R16을 *"오늘 빨간불"*로
적으면 거짓이다. **R16이 RED가 되는 시점은 §4.1(schemaV31)과 §4.3(claim)이 들어간 직후,
D13을 아직 구현하지 않은 상태다.**

| 시점 | R16 |
|---|---|
| 오늘 (a099 전) | **통과 — born-GREEN** |
| §4.1 + §4.3 직후, D13 전 | **실패해야 한다** ← 이것이 진짜 RED 순간이다 |
| §4 전체 후 | 통과 |

**tasks §3에 이 순서를 적지 않으면 R16은 관측 불가능한 RED다.**
이것을 여기 적는 이유는 1라운드 A-P12·B-P8이 잡은 것과 같은 실패를 반복하지 않기 위해서다.

## 이 함수에 테스트를 새로 쓰지 않는 것

- **`replay.go:551`의 오류 무시.** a099 밖이다.
  **`not-applicable`: 이 change는 그 경로의 오류 처리를 근거로 쓰지 않는다.**

## 덮이지 않은 것을 이름으로 적는다

- 이 함수의 **유일한 프로덕션 호출자가 반환 둘을 다 버린다**(`_, _ =`).
  그러므로 어떤 프로덕션 테스트도 이 함수의 반환을 통해 회귀를 잡을 수 없다.
  **회귀를 잡는 자리는 `ClaimAlertForDelivery`의 계약 테스트다.**
- 위 표의 Test 칸이 지목하는 기존 테스트는 실측했다:
  `internal/journal/outbox_test.go` · `a096_claim_for_delivery_test.go` ·
  `a097_rearm_is_a_new_episode_test.go`가 `EnqueueAlert`를 호출한다.
  **1라운드에서 없는 파일을 인용한 실패를 반복하지 않기 위해 `grep`으로 확인했다.**

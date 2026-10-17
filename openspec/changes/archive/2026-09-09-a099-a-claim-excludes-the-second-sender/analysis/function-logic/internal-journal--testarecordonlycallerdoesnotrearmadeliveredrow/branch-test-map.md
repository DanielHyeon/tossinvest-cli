# Branch Test Map: `TestARecordOnlyCallerDoesNotReArmADeliveredRow`

`ast.json`의 열거가 정본이다: 분기 5 · 이탈 0.

**이 표의 「Test」 열은 전부 이 함수 자신이다.** 테스트 함수의 분기는 단언과
배치 가드이고, 그것을 실행하는 것은 이 함수 하나뿐이다.
**GREEN은 실측이다**: `go test ./...`가 exit 0이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:255` 배치 오류 가드 — 단언이 아니다 | `TestARecordOnlyCallerDoesNotReArmADeliveredRow` | (아래) | **yes** |
| B2 | `:259` 배치 오류 가드 — 단언이 아니다 | `TestARecordOnlyCallerDoesNotReArmADeliveredRow` | (아래) | **yes** |
| B3 | `:267` 배치 오류 가드 — 단언이 아니다 | `TestARecordOnlyCallerDoesNotReArmADeliveredRow` | (아래) | **yes** |
| B4 | `:270` **`EnqueueAlert` 뒤에도 `ClaimSettled`다** — 기록이 재무장을 안 만들었다 | `TestARecordOnlyCallerDoesNotReArmADeliveredRow` | (아래) | **yes** |
| B5 | `:274` **상태가 `DELIVERED` 그대로다** | `TestARecordOnlyCallerDoesNotReArmADeliveredRow` | (아래) | **yes** |

## RED observed — 이 함수 전체에 대해

**yes — 컴파일 실패**(타입). 그리고 **§4.7 되돌림 관측**: `EnqueueAlert`를 `ClaimAlertForDelivery` 위임으로 되돌리면 **그 호출이 임차를 잡고**, 이어지는 청구가 `ClaimHeldElsewhere`를 받는다.

**RED 칸을 분기마다 따로 적지 않는 이유**: 이 함수의 분기는 한 시나리오를 이루는
단언들이고, 컴파일이 안 되면 **전부 동시에** 못 돈다. 분기별로 다른 값을 적으면
관측하지 않은 것을 관측한 것처럼 보인다.

## RED — 시점 관측의 이탈

§3.2에 적은 대로 §4를 **한 번에 구현한 뒤** 각 task가 더하는 것 하나만 되돌려
그 자리에서 실패를 재현했다. **되돌릴 자리를 내가 골랐으므로 「그 자리에서 본 것」과
같지 않다.** 그 판정은 리뷰의 몫이다.

## 덮이지 않은 것을 이름으로 적는다

- **`EnqueueAlert`가 임차를 안 잡는 것을 직접 단언하지 않는다.** 여기서는 결과(정산 유지)로만 본다. 직접 보는 것은 `TestARecordedAlertCanBeClaimedImmediately` `a099_lease_lifecycle_test.go:305`다.
- **배치 가드(`if err != nil`)에는 테스트가 없다.** 배치가 실패하는 상황을
  일부러 만들지 않는다. **`not-applicable`: 그 가드들은 단언이 아니라
  「단언에 도달했는지」의 표시다.**

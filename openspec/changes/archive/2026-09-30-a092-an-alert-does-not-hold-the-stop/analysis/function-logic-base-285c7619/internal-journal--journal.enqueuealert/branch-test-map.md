# Branch Test Map: `Journal.EnqueueAlert`

Source: `internal/journal/outbox.go` (115-122). AST 기준 분기 0 / 이탈 1 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없음 — `ClaimAlertForDelivery(ctx, a, 0)`에 위임하고 `id, err`만 돌려준다 `:120-121` | `TestEnqueueAlertIsIdempotentOnTheEventKey` (`outbox_test.go:31`) · `TestEnqueueAlertRequiresAKeyAndAType` (`:63`) | no | yes |

> **18라운드 B-P1이 지운 여덟 행.** 이 표는 B1~B9를 나열하고 B1·B2·B5에 GREEN을
> 달고 있었다. `ast.json`은 `"branches": null`이다 — 그 아홉은 a097 이전 구현의
> 분기이고, 지금은 **호출자가 아니라 피호출자인 `ClaimAlertForDelivery`의 것**이다.
> 그 함수의 분기 커버리지는
> `internal-journal--journal.claimalertfordelivery/branch-test-map.md`가 진다.
>
> 남은 한 행은 위임 자체를 단언한다: `TestEnqueueAlertIsIdempotentOnTheEventKey`는
> 같은 `event_key`를 두 번 넣어 `first == second`와 `PendingAlerts` 1행을 확인하고
> (`outbox_test.go:31-52`), `TestEnqueueAlertRequiresAKeyAndAType`는 빈 키·빈 종류가
> 오류가 되는 것을 확인한다(`:63-73`). **둘 다 이 래퍼를 통과해서** 피호출자의 성질을
> 관측하므로, 위임이 끊기면 둘 다 깨진다.

## a092가 이 함수에 대해 지는 것은 없다

편집하지 않으므로 새 RED가 없다.

미테스트로 남는 것은 **`owed`를 버리는 성질**이다 — 그 값을 버려도 안전한 호출자만
이 함수를 써야 한다는 계약(`outbox.go:112-114`)을 강제하는 테스트가 없다.
`execgw.parkAlert`(`replay.go:551`)가 그 계약을 지키는지는 이 함수가 아니라
`internal-execgw--gateway.parkalert`의 표가 다룬다.

이것을 여기서 RED로 만들지 않는 이유는 범위가 아니어서가 아니라 **대상이 이 함수가
아니어서다**: 강제해야 할 것은 "래퍼가 `owed`를 버린다"가 아니라 "버려진 `owed`를
누군가 대신 본다"이고, 그 누군가가 배달 실행자다(R18-1).

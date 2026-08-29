# Branch Test Map: `Notifier.notifyCritical`

측정: `go test -covermode=set ./internal/obs/...` — RED 84.8%(base `ec29dc72`), GREEN 85.4%(3판, a096b 반영).
줄 번호는 **GREEN 기준**이다. 이 함수는 RED에서도 GREEN에서도 4분기이며, 억제 판정과 전송은
2판에서 `claimAndDeliver`로 옮겼다. 바뀐 것은 B4의 **조건**이다.

판정 규칙: 분기 줄에서 **시작하는** 커버리지 블록의 count > 0 이면 `진입`,
자기 블록이 없으면 `—`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | journal 미배선 `:171` | `TestCriticalWithoutAJournalIsLoudRatherThanSilent` | 진입 | 진입 |
| B2 | 강등을 로그로 알린다 `:175` | 위와 같음 | 진입 | 진입 |
| B3 | claim/전송이 오류 `:195` | 없음 — 원장 실패 주입 하니스 없음 | 미진입 | 미진입 |
| B4 | **`owed && !sent` — 보내야 했는데 못 보냈을 때만 escalate** `:199` | `TestPersistentDeliveryFailureBlocksEntries`, `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` | 진입(`!n.deliver`) | 진입 |

## B4의 조건이 바뀐 것이 요점이다

RED에서 B4는 `!n.deliver(...)`였다 — 전송이 실패하면 escalate. 억제를 도입하면 그 조건은
**틀린다**: 보낼 필요가 없어서 안 보낸 것도 `!sent`이므로, 억제될 때마다 gate가 잠긴다.
`owed &&`가 그것을 막는다.

리뷰가 지적한 항목이 아니다. 억제를 넣으면서 스스로 확인해야 했던 반대 방향이다.

## RED이 무엇이었나

RED에서 `EnqueueAlert`는 무엇을 돌려주든 `deliver`가 불렸다. 새 테스트가 그것을 측정했다:

```text
--- FAIL: TestOneConditionIsOneSend
    sends = 3, want 1 — the operator was told once and the condition did not change
--- FAIL: TestSuppressingTheSendKeepsTheRecord
    sends = 5, want 1
    log lines = 9, want 5 — suppressing the push must not suppress the record
```

`log lines = 9`가 운영 증상의 재현이다: 관측 5줄 + `alert_undelivered` 오류 4줄.
그 4줄이 `no such alert`이고, 각각이 이미 나간 push 하나를 뜻한다.

`TestAnUndeliveredConditionIsStillRetried`는 **RED에서도 통과했다** — 과잉 억제
반증이므로 GREEN에서도 통과해야 하고, 통과한다.

## B4의 진입만으로는 왜 부족했나

RED 시점에도 그 자리의 분기(당시 B4@182)는 이미 진입했다. 진입한다는 사실은 **몇 번
진입하는지**를 말하지 않는다. [obs_test.go:620](../../../../../internal/obs/obs_test.go#L620)
`TestNotifierIsConcurrencySafe`는 전송이 성공하는 publisher로 key당 20회씩 critical
`Notify`를 돌리고 **단언이 없다**. 그 20회 중 19회가 재전송이며, 커버리지는 그것을
한 칸으로 접었다.

그래서 이 change의 완료 조건은 분기 진입이 아니라 **전송 횟수**다(spec: "중복 제거 계약은
전송 횟수로 검증한다").

## 미커버 분기에 대한 판단

B3은 `EnqueueAlert`가 오류를 돌려주는 경우이고, 그 자체가 DB 실패 주입을 요구한다
(`EnqueueAlert` BTM의 B3·B6·B7·B8과 같은 사유). a096은 B3의 조건도 반환도 바꾸지 않는다.
`not-applicable`: 실패 주입 하니스 부재, a096 범위 밖.

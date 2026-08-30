# Branch Test Map: `cleanupFrom`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range Outstanding(entries)` — 잔여물이 없으면 계획에 정리 줄이 없다 | `TestNoLeftoversMeansNoCleanupLines` | no — 기존 동작 | yes |
| B2 | `switch a.Kind` — 주문과 조건주문이 다른 규칙을 받는다 | `TestALeftoverOrderIsNotSubjectToTheOrderingRule` | no — 기존 동작 | yes |
| B3 | `case KindOrder` — 잔여 주문은 **항상** 대상이다 | `TestALeftoverOrderIsCancelledOnTheNextRun`, `TestALeftoverOrderIsNotSubjectToTheOrderingRule` | no — 이 change가 바꾸지 않은 경로. 회귀 방지용 | yes |
| B4 | `case KindConditional` — 조건주문은 B5의 두 조건을 지나야 한다 | `TestTheConditionalLeftForPersistenceIsNotCleanedUp` | no — 기존 동작 | yes |
| B5 | `settled(...) && decidedAfter(...)` — **이 change가 우변을 추가한 분기**. 세 하위 경로를 아래에 적는다 | `TestAVerdictOlderThanTheConditionalDoesNotCleanItUp` | **yes** | yes |

B5 하위 경로 (분기 id는 하나이고 조건이 AND 두 개다):

| 경로 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B5 좌변 false | 취소 단계 판정이 없거나 redo 중이면 정리하지 않는다 | `TestAConditionalWithNoCancelVerdictIsNotCleanedUp`, `TestTheConditionalLeftForPersistenceIsNotCleanedUp` | no — fail-closed 고정용 | yes |
| B5 우변 false | **조건주문보다 오래된 판정은 그것을 정리하지 못한다** | `TestAVerdictOlderThanTheConditionalDoesNotCleanItUp` | **yes** | yes |
| B5 둘 다 true | 등록 이후의 취소 실패는 정리를 연다 | `TestAVerdictNewerThanTheConditionalOpensCleanup` | no — 잔여물 영구 잔존 회귀 방지용 | yes |

RED 실행 기록 (구현 전, `go test ./internal/verifylive/`):

```
--- FAIL: TestAVerdictOlderThanTheConditionalDoesNotCleanItUp (0.00s)
    cleanup_test.go:299: the prologue would cancel the conditional the persistence step has to
    read, on the authority of a verdict recorded before it existed:
    [{Kind:conditional-order ID:grLKqiGuCVS7mj Symbol:333430 CreatedAt:0001-01-01 00:00:00
     +0000 UTC CancelledAt:0001-01-01 00:00:00 +0000 UTC Cancelled:false Deliberate:true Note:}]
```

GREEN 실행 기록 (구현 후): `go test ./internal/verifylive/ -count=1` → 195 passed.

B5-c·B5-a·B3의 "RED no"는 의도된 것이다. 이 change는 조건주문이 대상이 되는 경우를 **줄이는**
방향이므로, 줄어들지 말아야 할 경로는 수정 전에도 통과해야 정상이다. 그 세 테스트는 수정이
너무 넓게 닫히지 않았음을 증명한다.

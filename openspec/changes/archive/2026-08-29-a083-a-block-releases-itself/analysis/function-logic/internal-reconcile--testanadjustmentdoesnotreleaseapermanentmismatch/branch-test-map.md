# Branch Test Map: `TestAnAdjustmentDoesNotReleaseAPermanentMismatch`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (248) `for` — for i := 0; i < 3; i++ | `TestAnAdjustmentDoesNotReleaseAPermanentMismatch` 자신 | n/a | yes |
| B2 | (252) `if` — if !tracker.Permanent() | `TestAnAdjustmentDoesNotReleaseAPermanentMismatch` 자신 | n/a | yes |
| B3 | (262) `if` — if !tracker.Permanent() | `TestAnAdjustmentDoesNotReleaseAPermanentMismatch` 자신 | n/a | yes |
| B4 | (266) `if` — if rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent | `TestAnAdjustmentDoesNotReleaseAPermanentMismatch` 자신 | n/a | yes |

이 함수는 테스트 자체이므로 검증 주체와 대상이 같다. a083이 바꾼 것은 credit·diff의
as-of와 관측 사이의 시계 진행뿐이고 단언은 그대로다.

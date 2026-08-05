# Branch Test Map: `TestAnAdjustmentOnAnotherSymbolDoesNotRelease`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (231) `if` — if len(out.Cleared) != 0 | `TestAnAdjustmentOnAnotherSymbolDoesNotRelease` 자신 | n/a | yes |
| B2 | (234) `if` — if tracker.EntryAllowed("us", "AAPL") == nil | `TestAnAdjustmentOnAnotherSymbolDoesNotRelease` 자신 | n/a | yes |

이 함수는 테스트 자체이므로 검증 주체와 대상이 같다. a083이 바꾼 것은 credit·diff의
as-of와 관측 사이의 시계 진행뿐이고 단언은 그대로다.

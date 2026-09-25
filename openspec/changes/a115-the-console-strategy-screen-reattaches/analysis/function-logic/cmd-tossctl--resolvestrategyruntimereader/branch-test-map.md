# Branch Test Map: `resolveStrategyRuntimeReader`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 디렉터리 해석 실패 | not-applicable: 콘솔 판은 engineDir 해석 뒤에만 불린다 | no | no |
| B2 | 부재 → nil | `TestTheConsoleResolutionMatchesTheDaemons` | no | no |
| B3 | 비부재 stat 오류 → 경고+nil | `TestTheConsoleResolutionMatchesTheDaemons` | no | no |
| B4 | dial 실패 → sentinel | `TestTheConsoleResolutionMatchesTheDaemons` | no | no |

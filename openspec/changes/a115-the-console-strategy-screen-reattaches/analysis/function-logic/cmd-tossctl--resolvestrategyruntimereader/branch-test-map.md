# Branch Test Map: `resolveStrategyRuntimeReader`

구현 후 재확인: HEAD 재추출이 base ast.json 과 바이트 동일(무편집). 동치는 다섯 디스크 상태(부재·ENOTDIR·반쪽 잔재·죽은 socket·live)에서 잰다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 디렉터리 해석 실패 | not-applicable: 콘솔 판은 engineDir 해석 뒤에만 불린다 | no | no |
| B2 | 부재 → nil | `TestTheConsoleResolutionMatchesTheDaemons` | no — 무편집(동치 대조) | yes |
| B3 | 비부재 stat 오류 → 경고+nil | `TestTheConsoleResolutionMatchesTheDaemons` · `TestTheConsoleBootWarningDoesNotPromiseDormant` | no — 무편집(동치 대조) | yes |
| B4 | dial 실패 → sentinel | `TestTheConsoleResolutionMatchesTheDaemons` | yes — 콘솔판에 변이 K2(sentinel 대신 nil) 시 동치 시험 FAIL | yes |

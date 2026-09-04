# Branch Test Map: `NewPairedStrategyEntryProductionAssembly`

- Source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`; AST branch locations are authoritative.
- L5 5.3.3 이 이 함수에서 바꾼 것은 반환 리터럴 한 줄이다(assembly 가 서명된 일정
  **권위**를 함께 싣는다). **분기는 하나도 바뀌지 않았고 개수도 그대로 7 이다** —
  아래 좌표는 그 한 줄이 만든 이동만 반영한다.
- 그 한 줄이 하는 일을 재는 것은 이 함수의 분기가 아니라
  `TestTheRecoveryGenerationComesFromTheSignedActivationAndNothingElse`(복구 세대를
  읽는 식의 패키지 전체 열거)와 `TestALatchOnlyReopensForAStrictlyNewerSignedActivation`
  (그 값으로 실제로 잠금이 열리고 닫히는지)이다.
- 그 밖에는 L0 와 같다: 이 함수의 분기를 실행하는 시험은 없다.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 267:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B2 | if at 273:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B3 | if at 284:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B4 | if at 294:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B5 | if at 313:3 | planned targeted RED before any edit; not run by L0 | no | no |
| B6 | range at 330:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B7 | if at 335:2 | planned targeted RED before any edit; not run by L0 | no | no |
| B8 | if at 342:2 | planned targeted RED before any edit; not run by L0 | no | no |

A lot may replace a planned row only after recording its exact test name and actual RED/GREEN command result.

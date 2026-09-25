# Branch Test Map: `runConsole`

구현 후 AST 기준(분기 33). 편집 전 B33–B37 은 삭제됐다 — 그 삭제의 핀은 B32 행(구조 시험). 편집 전 B38 → B33.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B2 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B3 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B4 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B5 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B6 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B7 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B8 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B9 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B10 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B11 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B12 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B13 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B14 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B15 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B16 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B17 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B18 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B19 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B20 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B21 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B22 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B23 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B24 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B25 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B26 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B27 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B28 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B29 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B30 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B31 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |
| B32 | engineDir 해석 → 전략 wrapper(부재·sentinel·live 로 출발, 펌프 재부착), runConsole 안 직접 `strategyprojectionrpc.Dial` 0 | `TestRunConsoleNeverDialsTheStrategyProjectionItself` · `TestTheConsoleStrategyScreenRecoversWhenTheEngineStartsLater` · `TestTheConsoleStrategyScreenShowsADeadDescriptorAsUnreachable` | yes — base 컴파일 RED(새 심볼) · 변이 K7(부팅 dial 재도입) CAUGHT | yes |
| B33 | 편집 밖 부팅 배관 | not-applicable: a115 편집 밖 | no | no |

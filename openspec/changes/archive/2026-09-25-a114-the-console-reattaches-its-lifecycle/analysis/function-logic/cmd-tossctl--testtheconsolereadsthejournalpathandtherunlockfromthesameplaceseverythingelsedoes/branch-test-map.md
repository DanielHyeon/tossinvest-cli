# Branch Test Map: `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 필수 문자열 5개 순회 | `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes` | yes — a114 GREEN 직후 console.go 만 읽는 판본이 FAIL(`positionpolicyrpc.Dial(ctx, descriptorPath)` 없음) | yes |
| B2 | 필수 문자열 없음 | `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes` | no | yes |
| B3 | `journal.Open(` 금지 | `TestTheConsoleReadsTheJournalPathAndTheRunLockFromTheSamePlacesEverythingElseDoes` | no | yes |

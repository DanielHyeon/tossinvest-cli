# Branch Test Map: `TestMigrationV12CommitAndUserVersionSurviveSIGKILL`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | crash child takes only explicit migration mode | `TestMigrationV12CommitAndUserVersionSurviveSIGKILL` | yes | yes |
| B2 | child observes v12 before SIGKILL | same | yes | yes |
| B3 | parent v11 fixture closes | same | yes | yes |
| B4 | raw DB opens after crash | same | yes | yes |
| B5 | raw user_version is already 12 | same | yes | yes |
| B6 | v11/v12 artifacts are enumerated | same | yes | yes |
| B7 | raw artifact query succeeds | same | yes | yes |
| B8 | every artifact exists exactly once | same | yes | yes |
| B9 | prior rows survive normal reopen | same | yes | yes |

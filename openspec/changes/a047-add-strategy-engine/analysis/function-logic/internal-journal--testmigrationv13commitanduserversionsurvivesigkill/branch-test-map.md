# Branch Test Map: `TestMigrationV13CommitAndUserVersionSurviveSIGKILL`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | child reaches current schema before SIGKILL | same test child branch | baseline v13 | v14 pass |
| B2 | raw DB version/artifacts and legacy counts survive | same test parent branch | hardcoded v13 stale | `SchemaVersion` pass |
| B3 | closing the seeded v12 journal succeeds | same test | baseline | pass |
| B4 | opening the raw post-crash database succeeds | same test | baseline | pass |
| B5 | raw `user_version` equals current schema | same test | hardcoded v13 stale | pass |
| B6 | each required v13 protection artifact exists exactly once | same test artifact loop | baseline | pass |
| B7 | artifact query/count fails closed | same test assertion | baseline | pass |
| B8 | reopened legacy row counts are byte-preserving | same test | baseline | pass |

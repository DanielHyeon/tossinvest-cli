# Branch Test Map: a112 트립와이어 옛 판본(revision: base — a127 에서 지움)

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:459` `if len(scoped) != 2 {` | 대체 양성 시험 `TestTheRiskLoaderReadsTheRealJournal` — review 1.1~1.4 · 변이 S1 | n/a | n/a |
| B2 | `:467` `if err == nil \|\| !strings.HasSuffix(err.Error(), "risk bucket: exact journal schema unavailable") {` | 대체 양성 시험 `TestTheRiskLoaderReadsTheRealJournal` — review 1.1~1.4 · 변이 S1 | n/a | n/a |
| B3 | `:473` `for _, scope := range collected.kr.scopes {` | 대체 양성 시험 `TestTheRiskLoaderReadsTheRealJournal` — review 1.1~1.4 · 변이 S1 | n/a | n/a |
| B4 | `:474` `if scope.ready {` | 대체 양성 시험 `TestTheRiskLoaderReadsTheRealJournal` — review 1.1~1.4 · 변이 S1 | n/a | n/a |

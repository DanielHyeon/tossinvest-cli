# Branch Test Map: `TestA126CorruptReceiptedRowsAreUnreadable`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:349` `for name, corrupt := range cases {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B2 | `:350` `for _, reverted := range []bool{false, true} {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B3 | `:353` `if reverted {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B4 | `:354` `if _, err := j.db.Exec(`INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generat…` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B5 | `:359` `if _, err := j.db.Exec(corrupt, o.key.ProspectiveGeneration); err != nil {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B6 | `:362` `if _, err := riskbucket.ReadJournalBucketUsage(context.Background(), j.db, a126Account, riskbucket.Dimensio…` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B7 | `:368` `if err := a126AdmitSymbol(t, j, "after-corrupt", "MSFT", "1000", "50", 1); !errors.Is(err, ErrRiskBucketSna…` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |

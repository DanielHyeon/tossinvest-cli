# Branch Test Map: `TestA126ResidualRevertAfterAnotherOwnersReservationLeavesItUnlatched`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:556` `if err := a126AdmitSymbol(t, j, "c2-b", "MSFT", "100", "0", 12); err != nil {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B2 | `:559` `if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='CONFIRMED' WHERE id='a126-late-attempt-c2'`); e…` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B3 | `:562` `if err := j.SetApplyHooks(ApplyHooks{Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return …` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B4 | `:565` `if res, err := j.RecordFill(context.Background(), a126LateFill(o, order)); err != nil \|\| !res.Changed {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B5 | `:569` `if usage.FilledMinor != "50" \|\| usage.HeldMinor != "60" {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B6 | `:573` `if err := j.db.QueryRow(`SELECT (SELECT count(*) FROM risk_bucket_scope_latches WHERE prospective_generatio…` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |

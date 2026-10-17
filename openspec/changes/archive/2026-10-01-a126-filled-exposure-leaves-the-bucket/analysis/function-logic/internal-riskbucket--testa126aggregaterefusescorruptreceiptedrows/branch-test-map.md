# Branch Test Map: `TestA126AggregateRefusesCorruptReceiptedRows`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:76` `for name, corrupt := range cases {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B2 | `:77` `for _, latched := range []int{0, 1} {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B3 | `:80` `if _, err := aggregateProductionRiskUsage([]productionRiskUsageRow{row}); !errors.Is(err, ErrJournalUsageIn…` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |

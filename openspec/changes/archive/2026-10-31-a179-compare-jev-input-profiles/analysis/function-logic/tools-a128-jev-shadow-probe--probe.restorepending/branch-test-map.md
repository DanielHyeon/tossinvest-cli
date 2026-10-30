# Branch Test Map: `probe.restorePending`

Relevant package test surfaces: TestCompareCycleSharesCollectionAndLabels, TestRestorePendingAfterRestart. These tests cover named scenarios, not every AST branch; rows below identify unproven individual branch coverage rather than asserting it. Existing socket-backed tests could not execute in this sandbox.

| Branch | Source branch | Individual branch coverage |
|---|---|---|
| B1 | range at 802:2: `for _, row := range contents.Judgments {` | Not individually established |
| B2 | if at 803:3: `if _, labeled := contents.Labels[row.ID]; labeled {` | Not individually established |
| B3 | if at 807:3: `if err != nil {` | Not individually established |
| B4 | if at 811:3: `if row.Cohort != "" {` | Not individually established |
| B5 | if at 814:3: `if groups[key] == nil {` | Not individually established |
| B6 | range at 820:2: `for _, key := range order {` | Not individually established |

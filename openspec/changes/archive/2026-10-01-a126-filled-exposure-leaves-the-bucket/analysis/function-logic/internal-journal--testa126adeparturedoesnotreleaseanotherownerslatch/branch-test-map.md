# Branch Test Map: `TestA126ADepartureDoesNotReleaseAnotherOwnersLatch`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:500` `if err := a126AdmitSymbol(t, j, "b", "MSFT", "100", "50", 10); err != nil {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B2 | `:507` `if overage, _, _ := ownerFlags(t, j, b); overage != 1 {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B3 | `:511` `if overage, _, _ := ownerFlags(t, j, b); overage != 1 {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B4 | `:515` `if usage.FilledMinor != "24" \|\| usage.HeldMinor != "30" \|\| !usage.OverageLatched {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B5 | `:518` `if err := a126Admit(t, j, "c", "1000", 1); !errors.Is(err, ErrRiskBucketEntryBlocked) {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B6 | `:522` `if _, err := j.ReleaseRiskOverageLatch(ctx, latchRelease(b, ownerLatchView(t, j, b).StateDigest, &relaxatio…` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B7 | `:526` `if overage, unknown, rows := ownerFlags(t, j, b); overage != 0 \|\| unknown != 0 \|\| rows != 0 {` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |
| B8 | `:529` `if usage := a126Usage(t, j, riskbucket.DimensionSector, "sector-tech"); usage.FilledMinor != "30" \|\| usag…` | 이 시험 자체 — review 1.5.2 · 변이 표 | n/a | n/a |

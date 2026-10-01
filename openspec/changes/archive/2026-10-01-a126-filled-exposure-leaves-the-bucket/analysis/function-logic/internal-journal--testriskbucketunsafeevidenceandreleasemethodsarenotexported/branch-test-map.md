# Branch Test Map: `TestRiskBucketUnsafeEvidenceAndReleaseMethodsAreNotExported`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:326` `for _, name := range []string{"CompleteRiskBucketFillActual", "ReleaseRiskBucketOrder"} {` | 이 시험 자체(census) | n/a | n/a |
| B2 | `:327` `if _, exists := typeOfJournal.MethodByName(name); exists {` | 이 시험 자체(census) | n/a | n/a |
| B3 | `:332` `if err != nil {` | 이 시험 자체(census) | n/a | n/a |
| B4 | `:335` `for _, path := range files {` | 이 시험 자체(census) | n/a | n/a |
| B5 | `:336` `if strings.HasSuffix(path, "_test.go") {` | 이 시험 자체(census) | n/a | n/a |
| B6 | `:340` `if err != nil {` | 이 시험 자체(census) | n/a | n/a |
| B7 | `:343` `for _, call := range []string{".completeRiskBucketFillActual(", ".releaseRiskBucketOrder(", ".rel…` | 이 시험 — 변이 T1(비시험 파일에 `.releaseRiskBucketOwner(` 한 줄) CAUGHT | yes | yes |
| B8 | `:344` `if strings.Contains(string(raw), call) {` | 이 시험 — 변이 T1(비시험 파일에 `.releaseRiskBucketOwner(` 한 줄) CAUGHT | yes | yes |

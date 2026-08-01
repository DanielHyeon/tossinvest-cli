# Branch Test Map: `missingEvidence`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero complete rows reports missing complete lineage | `TestProviderFailsClosedForEveryMissingEvidenceClass/no_complete_lineage` | adapter absent at `948e721` | PASS |
| B2 | undersampled/insufficient/empty aggregate evidence reports minimum sample | `TestProviderFailsClosedForEveryMissingEvidenceClass/insufficient_sample` | adapter absent at `948e721` | PASS |
| B3 | link-missing count remains observable because recommendation query is not complete-only | `TestProviderFailsClosedForEveryMissingEvidenceClass/link_missing`, fixed-query test | adapter absent at `948e721` | PASS |
| B4 | not-measured count reports explicit state | `TestProviderFailsClosedForEveryMissingEvidenceClass/not_measured` | adapter absent at `948e721` | PASS |
| B5 | every aggregate is inspected | `TestProviderFailsClosedForLineageSemanticsAndRequiredMetricShape` table cases | adapter absent at `948e721` | PASS |
| B6 | blank market/lane/version/policy lineage is insufficient | `TestProviderFailsClosedForLineageSemanticsAndRequiredMetricShape/blank_lineage` | adapter absent at `948e721` | PASS |
| B7 | incomplete or undersampled aggregate is insufficient | `TestProviderFailsClosedForEveryMissingEvidenceClass/insufficient_sample` | adapter absent at `948e721` | PASS |
| B8 | semantics version drift is insufficient | `TestProviderFailsClosedForLineageSemanticsAndRequiredMetricShape/semantics_version` | adapter absent at `948e721` | PASS |
| B9 | every metric is indexed for coverage and duplicate detection | required-metric-shape table cases | adapter absent at `948e721` | PASS |
| B10 | duplicate metric key is rejected | `TestProviderFailsClosedForLineageSemanticsAndRequiredMetricShape/duplicate_metric` | adapter absent at `948e721` | PASS |
| B11 | every frozen required key is checked | required-metric-shape table cases | adapter absent at `948e721` | PASS |
| B12 | malformed required metric is named explicitly | missing/incomplete/duplicate/undersampled/blank metric cases | adapter absent at `948e721` | PASS |
| B13 | unique reasons are returned in stable lexical order | `TestProviderFailsClosedForLineageSemanticsAndRequiredMetricShape` sorted reason assertion | adapter absent at `948e721` | PASS |

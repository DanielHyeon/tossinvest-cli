# Branch Test Map: `TestOptimizationCandidateFiltersRenderDormantReadOnlyEvidenceState`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | iterate required evidence labels | self | category routing initially absent | yes |
| B2 | fail on missing label | self | category routing initially absent | yes |
| B3 | ignore non-elements | self | DOM contract preserved | yes |
| B4 | classify forbidden controls | self | DOM contract preserved | yes |
| B5 | form/textarea/input is forbidden | self | DOM contract preserved | yes |
| B6 | contenteditable is forbidden | self | DOM contract preserved | yes |

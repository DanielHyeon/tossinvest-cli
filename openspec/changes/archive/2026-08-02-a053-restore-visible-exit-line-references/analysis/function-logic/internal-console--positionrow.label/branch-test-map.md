# Branch Test Map: `positionRow.Label`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | typed management projection wins | `TestPendingDesignationTruthTable/typed_projection` | no | yes |
| B2 | ordered status switch selects exactly one fallback | `TestPendingDesignationTruthTable` | yes | yes |
| B3 | unreadable journal stays unknown | `TestPendingDesignationTruthTable/journal_unknown` | no | yes |
| B4 | known released lifecycle wins over desired include | `TestPendingDesignationTruthTable/operator_released` | yes | yes |
| B5 | explicit exclusion wins over desired include | `TestPendingDesignationTruthTable/excluded` | no | yes |
| B6 | true desired-only broker holding is pending | `TestPendingDesignationTruthTable/desired_only` | no | yes |
| B7 | remaining unmanaged row is not pending | `TestPendingDesignationTruthTable/broker_absent` | yes | yes |
| B8 | managed completed exit | `TestPendingDesignationTruthTable/managed_completed_exit` | no | yes |
| B9 | managed active exit | `TestPendingDesignationTruthTable/managed_active_exit` | no | yes |
| B10 | managed row awaiting exit state | `TestPendingDesignationTruthTable/managed` | no | yes |

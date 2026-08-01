# Branch Test Map: `Journal.recordExitJudgementTx`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B2 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B3 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B4 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B5 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B6 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B7 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B8 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B9 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B10 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B11 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B12 | input/provenance/proposal/transaction/completed validation fails before mutation | `exit judgement validation suite` | yes | yes |
| B13 | legacy caller resolves zero to current lifecycle | `exit judgement compatibility suite` | yes | yes |
| B14 | old judgement generation is refused | `TestLateOldGenerationJudgementIsQuarantined` | yes | yes |
| B15 | pre-lifecycle row defaults to generation 1 managed | `migration/legacy exit suite` | yes | yes |
| B16 | lifecycle query error aborts | `exit judgement persistence suite` | yes | yes |
| B17 | non-not-found lifecycle query error aborts | `exit judgement persistence suite` | yes | yes |
| B18 | released or non-current lifecycle is refused | `TestLateOldGenerationJudgementIsQuarantined` | yes | yes |
| B19 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B20 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B21 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B22 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B23 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B24 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B25 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B26 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B27 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B28 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B29 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B30 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B31 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B32 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B33 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B34 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B35 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B36 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B37 | position generation, duplicate identity, monotonicity and snapshot recovery remain fail-closed | `exit snapshot recovery/corruption suite` | yes | yes |
| B38 | state update failure aborts | `exit judgement persistence suite` | yes | yes |
| B39 | after-state crash hook rolls back | `exit atomicity suite` | yes | yes |
| B40 | proposal path arms only after state write | `exit proposal suite` | yes | yes |
| B41 | arm failure rolls back | `exit proposal suite` | yes | yes |
| B42 | after-arm crash hook rolls back | `exit atomicity suite` | yes | yes |
| B43 | complete evaluation is attached to event | `exit snapshot persistence suite` | yes | yes |
| B44 | event failure rolls back | `exit atomicity suite` | yes | yes |
| B45 | after-event crash hook rolls back | `exit atomicity suite` | yes | yes |
| B46 | commit failure returns error | `exit atomicity suite` | yes | yes |
| B47 | result classification is exhaustive | `exit judgement result suite` | yes | yes |
| B48 | saved-monotone recovery result | `exit snapshot recovery suite` | yes | yes |
| B49 | new proposal result | `exit proposal suite` | yes | yes |
| B50 | working-order suppression result | `exit working-order suite` | yes | yes |

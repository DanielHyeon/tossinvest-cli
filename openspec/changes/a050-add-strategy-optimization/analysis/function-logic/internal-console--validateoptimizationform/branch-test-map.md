# Branch Test Map: `validateOptimizationForm`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | duplicate action | strict request tests | silently selected | yes |
| B2 | optional action extraction | lifecycle tests | existing path | yes |
| B3 | action present | lifecycle tests | existing path | yes |
| B4 | action dispatch | lifecycle tests | existing path | yes |
| B5 | preview allowlist | preview tests | existing path | yes |
| B6 | preview fields | preview tests | existing path | yes |
| B7 | apply allowlist | apply tests | existing path | yes |
| B8 | apply fields | apply tests | existing path | yes |
| B9 | rollback allowlist | rollback tests | existing path | yes |
| B10 | rollback fields | rollback tests | existing path | yes |
| B11 | unknown action | strict request tests | fell into preview | yes |
| B12 | unexpected field | strict request tests | ignored | yes |
| B13 | duplicate field | strict request tests | selected one | yes |
| B14 | a present allowed field still must have exactly one value | `TestOptimizationRejectsUnexpectedAndDuplicateFields` | duplicate values were selected implicitly | PASS |
| B15 | every action-specific required field is checked even when absent | strict missing-field request cases | missing fields reached downstream parsing | PASS |
| B16 | missing or duplicated required field is rejected | strict exact-set request cases | incomplete form reached downstream parsing | PASS |

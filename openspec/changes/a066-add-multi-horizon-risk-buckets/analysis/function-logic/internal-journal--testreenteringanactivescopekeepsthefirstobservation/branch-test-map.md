# Branch Test Map: `TestReEnteringAnActiveScopeKeepsTheFirstObservation`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | first enter error/state | this test | existing | yes |
| B2 | second enter error | this test | existing | yes |
| B3 | second enter unexpectedly new | this test | existing | yes |
| B4 | first observation changed | this test | existing | yes |
| B5 | active read error | this test | existing | yes |
| B6 | active count mismatch | this test | existing | yes |

2026-09-29 (base re-pin to the landing `e9355a82`): the `ast.json` was `revision: base` at the old base `23794f86`. It is re-extracted at the landing content, which is also HEAD. Branch positions are identical (B1 67 … B6 90, 59–93); only the file hash changed (`9bf1a3a6` edited elsewhere in the file). Rows unchanged.

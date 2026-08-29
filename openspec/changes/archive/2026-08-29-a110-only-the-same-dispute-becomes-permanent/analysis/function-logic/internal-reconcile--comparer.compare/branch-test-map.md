# Branch Test Map: `Comparer.Compare`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | snapshot timestamp is present | comparer timestamp baseline | preserve | yes |
| B2 | broker holdings indexed | comparer baseline | preserve | yes |
| B3 | broker symbols enumerated | comparer baseline | preserve | yes |
| B4 | broker symbol first seen | comparer baseline | preserve | yes |
| B5 | local symbols enumerated | comparer baseline | preserve | yes |
| B6 | local symbol first seen | comparer baseline | preserve | yes |
| B7 | normalized symbol comparison loop | quantity suite | preserve | yes |
| B8 | either present side is unreadable | raw comparer→tracker tests | yes | yes |
| B9 | quantity classification switch | quantity suite | preserve | yes |
| B10 | both sides exact zero | zero tests | preserve | yes |
| B11 | absent local plus positive broker is nonblocking external | `TestExternalPositionIsClassified` | preserve | yes |
| B12 | absent local plus negative broker is ordinary no-promotion mismatch | `TestA110NegativeBrokerOnlyHoldingFailsClosedWithoutPromotion` | yes | yes |
| B13 | known local quantity disagrees | tolerance-zero tests | yes | yes |
| B14 | valid equality/finite-ULP artifact increments match | agreement/0.3 tests | preserve | yes |
| B15 | broker orders enumerated | order comparer suite | preserve | yes |
| B16 | local orders enumerated for candidates | order comparer suite | preserve | yes |
| B17 | incompatible identity skipped | A110 missing-order scope | preserve | yes |
| B18 | broker candidate sets enumerated | order comparer suite | preserve | yes |
| B19 | ambiguous broker candidate skipped | order comparer suite | preserve | yes |
| B20 | ambiguous local ownership skipped | order comparer suite | preserve | yes |
| B21 | broker orders classified | order comparer suite | preserve | yes |
| B22 | unmatched broker order emitted external | external-order suite | preserve | yes |
| B23 | local orders classified | A110 missing-order suite | yes | yes |
| B24 | unmatched local order appended | A110 missing-order identity | yes | yes |
| B25 | non-empty missing list assigned | A110 missing-order identity | preserve | yes |

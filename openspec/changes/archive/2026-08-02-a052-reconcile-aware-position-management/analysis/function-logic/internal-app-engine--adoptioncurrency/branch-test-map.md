# Branch Test Map: `adoptionCurrency`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `switch` line 300 | `switch strings.ToLower(strings.TrimSpace(market)) {` true/entered and complementary path | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency; TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B2 | `case` line 301 | `case "kr":` true/entered and complementary path | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency; TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B3 | `case` line 303 | `case "us":` true/entered and complementary path | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency; TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 | covered by a052 contract RED or pre-existing regression | verified by focused package suite |
| B4 | `case` line 305 | `default:` true/entered and complementary path | TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency; TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0 | covered by a052 contract RED or pre-existing regression | verified by focused package suite |

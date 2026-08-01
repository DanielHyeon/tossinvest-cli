# Branch Test Map: `TestTradingViewsCarryWholePageResponsiveAndFocusContracts`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | inspect every responsive/focus contract | `TestTradingViewsCarryWholePageResponsiveAndFocusContracts` | removing a required CSS/ARIA token fails | yes |
| B2 | one contract is absent | same test | missing token reported | yes |

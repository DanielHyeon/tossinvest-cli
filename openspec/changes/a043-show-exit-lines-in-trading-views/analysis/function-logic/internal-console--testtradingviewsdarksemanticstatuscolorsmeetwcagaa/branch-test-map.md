# Branch Test Map: `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | primary dark media missing | `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA` | malformed stylesheet boundary fails | yes |
| B2 | dark/mobile closing boundary missing | same test | malformed stylesheet boundary fails | yes |
| B3 | inspect both semantic states | same test | old stylesheet had no dark overrides | yes |
| B4 | token is outside/missing from dark scope | same test | old `.ok/.bad` selectors used light-only colours | yes |
| B5 | dark token misses WCAG AA | same test | old ratios were 2.54:1 and 2.11:1 | new ratios are at least 4.5:1 |

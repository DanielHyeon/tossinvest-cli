# Branch Test Map: `relativeLuminance`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | malformed shape rejected | fixed literals make this defensive branch unreachable in production test data | not applicable | reviewed defense |
| B2 | parse all three channels | `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA` | no measured contrast existed | yes |
| B3 | malformed channel rejected | fixed valid literals make this defensive branch unreachable | not applicable | reviewed defense |
| B4 | linear low-sRGB branch | WCAG helper formula review; current approved tokens use gamma branch | not applicable to approved token fixture | reviewed formula |
| B5 | gamma sRGB branch | `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA` | no measured contrast existed | yes |

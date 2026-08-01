# Branch Test Map: `TestRiskPreviewWaitsThreeSecondsAndAppliesSameCapabilityExactlyOnce`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | preview returns 200 | this test | existing | pass |
| B2 | CSP exactly authorizes static countdown | this test | existing | pass |
| B3 | inspect countdown/checkbox markers | this test | expanded | pass |
| B4 | each countdown marker exists | this test | expanded | pass |
| B5 | preview HTML parses | this test | expanded | pass |
| B6 | one aria-readonly review section | this test | new | pass |
| B7 | one sticky risk approval block | this test | new | pass |
| B8 | DOM walk ignores text nodes | this test | new | pass |
| B9 | textarea/select absent | this test | new | pass |
| B10 | contenteditable absent | this test | new | pass |
| B11 | inspect input controls | this test | new | pass |
| B12 | checkbox counted | this test | new | pass |
| B13 | non-checkbox input branch | this test | new | pass |
| B14 | only hidden non-checkbox inputs | this test | new | pass |
| B15 | exactly one checkbox and button | this test | new | pass |
| B16 | exactly one inline script with text | this test | existing | pass |
| B17 | CSP contains digest of exact script bytes | this test | existing | pass |
| B18 | script cannot read/mutate capability | this test | existing | pass |
| B19 | early apply returns 425 | this test | existing | pass |
| B20 | early apply leaves version/audit unchanged | this test | existing | pass |
| B21 | t+3 apply returns 200 | this test | existing | pass |
| B22 | t+3 apply advances once | this test | existing | pass |
| B23 | replay leaves version/audit unchanged | this test | existing | pass |

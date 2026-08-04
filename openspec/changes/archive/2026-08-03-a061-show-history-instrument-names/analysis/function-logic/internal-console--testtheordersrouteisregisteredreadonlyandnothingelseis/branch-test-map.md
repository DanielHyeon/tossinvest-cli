# Branch Test Map: `TestTheOrdersRouteIsRegisteredReadOnlyAndNothingElseIs`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | every extracted route is judged | self | yes | yes |
| B2 | required history view is recognized | self | yes | yes |
| B3 | missing readOnly is rejected | route extractor positive controls | existing green | existing green |
| B4 | CSRF on a read route is rejected | self | existing green | existing green |
| B5 | missing session gate is rejected | self | existing green | existing green |
| B6 | unreviewed readOnly route is rejected | self | existing green | existing green |
| B7 | every required path was seen | self | yes | yes |
| B8 | missing required route is rejected | self | existing green | existing green |

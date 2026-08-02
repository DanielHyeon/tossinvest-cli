# Branch Test Map: `StartPositionPolicyCommandServer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | service/directory missing | existing construction tests | no | no |
| B2 | private path unsafe | existing private FS tests | no | no |
| B3 | missing bearer | runtime endpoint returns unauthorized | no | no |
| B4 | runtime endpoint POST | method-not-allowed, no call | no | no |
| B5 | runtime read success/error | exact JSON / typed error | no | no |
| B6 | normal start/dial/close | descriptor lifecycle remains valid | no | no |
| B7 | control directory created/existing | cleanup ownership preserved | no | no |
| B8 | listener/token construction | failure closes acquired resource | no | no |
| B9 | health method gate | GET only | no | no |
| B10 | positions method gate | GET only | no | no |
| B11 | positions service error | typed failure | no | no |
| B12 | runtime method gate | GET only | no | no |
| B13 | runtime service error | typed failure | no | no |
| B14 | descriptor write error | listener/control cleanup | no | no |
| B15 | serving success | bounded server starts | no | no |

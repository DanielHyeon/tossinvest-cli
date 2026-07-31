# Branch Test Map: `TestTheExcludeControlAsksForNoTyping`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | choose first violated CSP/one-click condition | TestTheExcludeControlAsksForNoTyping | yes | yes |
| B2 | inline confirm exists | same test | yes — prior template required it | yes |
| B3 | explicit exclusion action missing | same test | yes before button implementation | yes |
| B4 | prompt asks for typing | same test | n/a preserved guard | yes |
| B5 | text input asks for typing | same test | n/a preserved guard | yes |

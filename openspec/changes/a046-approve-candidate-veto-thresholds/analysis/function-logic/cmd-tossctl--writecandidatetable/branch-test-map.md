# Branch Test Map: `writeCandidateTable`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | halted report | candidate CLI table suite | existing coverage | halt rendered |
| B2 | quiet turn | candidate CLI table suite | existing coverage | no-source explanation |
| B3 | attempted sources | candidate CLI table suite | existing coverage | attempted/responded rendered |
| B4 | missing sources | candidate CLI table suite | existing coverage | each rendered |
| B5 | source rate limited | candidate CLI table suite | existing coverage | marker rendered |
| B6 | not-due sources exist | candidate CLI table suite | existing coverage | list rendered |
| B7 | source backoff entries | candidate CLI table suite | existing coverage | each rendered |
| B8 | engine yield | candidate CLI table suite | existing coverage | yield rendered |
| B9 | source readings | candidate CLI table suite | existing coverage | each rendered |
| B10 | whole reading | candidate CLI table suite | existing coverage | whole label |
| B11 | held first ranks | candidate CLI table suite | existing coverage | held explanation |
| B12 | rejected candidates | candidate CLI table suite | existing coverage | rejected summary |
| B13 | rejected rows | candidate CLI table suite | existing coverage | each row rendered |
| B14 | D3 veto counts | candidate CLI veto output tests | exported array writable | accessor copy preserves order/counts |
| B15 | veto reasons | candidate CLI veto output tests | existing coverage | each reason rendered |
| B16 | veto alarms | candidate CLI veto output tests | existing coverage | each alarm rendered |
| B17 | sightings exist | candidate CLI table suite | existing coverage | heading rendered |
| B18 | sighting sources | candidate CLI table suite | existing coverage | each source rendered |
| B19 | sighting reasons | candidate CLI table suite | existing coverage | each refusal rendered |
| B20 | acceleration crossing lines | candidate CLI shadow tests | existing coverage | wrapped lines rendered |
| B21 | acceleration missing reasons | candidate CLI shadow tests | existing coverage | each rendered |
| B22 | shadow band codes | candidate CLI shadow tests | existing coverage | each known band considered |
| B23 | band absent | candidate CLI shadow tests | existing coverage | skipped |
| B24 | band alarm exists | candidate CLI shadow tests | existing coverage | alarm rendered first |
| B25 | band crossing lines | candidate CLI shadow tests | existing coverage | rendered |
| B26 | band quantile lines | candidate CLI shadow tests | existing coverage | rendered |
| B27 | quantile base differs | candidate CLI shadow tests | existing coverage | discrepancy rendered |
| B28 | band missing reasons | candidate CLI shadow tests | existing coverage | each rendered |
| B29 | retention state switch | retention CLI tests | existing coverage | one case selected |
| B30 | retention error | retention CLI tests | existing coverage | error rendered |
| B31 | WAL busy | retention CLI tests | existing coverage | busy rendered |
| B32 | retention success | retention CLI tests | existing coverage | reclaimed rendered |
| B33 | disk checked | free-space CLI tests | existing coverage | bytes/floor rendered |
| B34 | disk unmeasured | free-space CLI tests | existing coverage | reason rendered |

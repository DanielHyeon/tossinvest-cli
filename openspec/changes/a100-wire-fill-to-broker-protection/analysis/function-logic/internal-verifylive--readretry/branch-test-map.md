# Branch Test Map: `readRetry`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B2-B3 | success/non-rate error | existing retry tests | no | yes |
| B4 | 429 then success | M0 irreversible 429 gap | yes | HOLD retained |
| B5 | sleep/read failure | M0 attempt receipt | yes | HOLD retained in window |
| B1 | retry iteration | existing retry test | no | yes |
| B2 | successful closure | existing retry test | no | yes |
| B3 | non-rate failure | existing retry test | no | yes |

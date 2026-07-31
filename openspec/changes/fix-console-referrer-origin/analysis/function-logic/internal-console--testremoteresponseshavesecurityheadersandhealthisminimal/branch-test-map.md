# Branch Test Map: `TestRemoteResponsesHaveSecurityHeadersAndHealthIsMinimal`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | inspect every required security header | this function | unchanged | passed |
| B2 | every required header remains non-empty | this function | unchanged | passed |
| B3 | effective remote referrer policy is exact `same-origin` | this function | failed: `no-referrer` | passed |
| B4 | loopback GET health remains minimal | this function | unchanged | passed |
| B5 | POST health remains method-rejected | this function | unchanged | passed |

# Branch Test Map: `sealOfficial`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | malformed rate/window/pair or policy seal/canonicality refuses | TestSealOfficialRejectsInvalidResponseOrPolicy | yes (sealed policy API absent) | yes |
| B2 | response freshness window does not intersect policy window | TestSealOfficialRejectsInvalidResponseOrPolicy | yes (caller haircut had no policy window) | yes |
| B3 | canonical digest construction fails | TestSealOfficialRejectsInvalidResponseOrPolicy | yes (caller supplied digest) | yes |
| B4 | sealed evidence fails self-validation | TestSealOfficialRejectsInvalidResponseOrPolicy | yes (public seal was discarded) | yes |

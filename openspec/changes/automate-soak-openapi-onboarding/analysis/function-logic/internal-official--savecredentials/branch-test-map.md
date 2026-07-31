# Branch Test Map: `SaveCredentials`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | parent creation failure returns | filesystem error characterization | existing | passed |
| B2 | credential marshals without secret output | existing save/load tests | existing | passed |
| B3 | temp creation failure returns | filesystem error characterization | existing | passed |
| B4 | temporary artifact is cleaned | save tests inspect target directory | existing | passed |
| B5 | temporary chmod failure returns | filesystem seam characterization | existing | passed |
| B6 | temporary write failure returns | filesystem seam characterization | existing | passed |
| B7 | temporary fsync failure returns | filesystem seam characterization | existing | passed |
| B8 | temporary close failure returns | filesystem seam characterization | existing | passed |
| B9 | rename publishes replacement | `TestSaveCredentialsTightensExistingFileTo0600` | yes | passed |
| B10 | final target can be inspected | `TestSaveCredentialsIs0600` | existing | passed |
| B11 | final target is regular 0600 | `TestSaveCredentialsTightensExistingFileTo0600` | yes | passed |
| B12 | parent sync completes | existing save/load tests | existing | passed |

# Branch Test Map: `protectionVerifierPolicy.validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | bad version/generation | policy malformed table | policy did not exist | rejected |
| B2 | noncanonical attestation path/basename | policy path table | caller path was accepted | rejected |
| B3 | noncanonical root path/basename | policy path table | caller root path was accepted | rejected |
| B4 | malformed/unpinned digest | policy digest table | caller digest was accepted | rejected |
| B5 | three distinct parents | separated non-root layout test | no canonical policy | accepted |
| B6 | shared parent permits ownership coupling | same-parent test | colocated artifacts accepted | rejected |

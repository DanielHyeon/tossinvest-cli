# Branch Test Map: `verifyCandidatePayloadMAC`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | malformed/mismatched MAC fails | `TestApplyRejectsSelfConsistentCandidatePayloadTamperWithoutMAC` | yes | yes |
| B2 | constant-time comparison determines result | `TestApplyRejectsSelfConsistentCandidatePayloadTamperWithoutMAC` | yes | yes |

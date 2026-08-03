# Branch Test Map: `Evidence.EvidenceAt`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | copied/tampered evidence refuses | TestReserveAtFailsClosedOutsideWindowAndAfterTamper | yes (opaque reserve API absent) | yes |
| B2 | mismatched quote currency refuses | TestReserveAtFailsClosedOutsideWindowAndAfterTamper | yes (opaque reserve API absent) | yes |
| B3 | evidence source is neither official nor identity | TestReserveAtFailsClosedOutsideWindowAndAfterTamper | yes (opaque reserve API absent) | yes |
| B4 | official-evidence branch selected | TestSealOfficialPreservesExactValuesAndCanonicalPolicy | yes (opaque reserve API absent) | yes |
| B5 | official policy binding is invalid | TestReserveAtFailsClosedOutsideWindowAndAfterTamper | yes (public seal was discarded) | yes |
| B6 | identity-evidence branch selected | TestIdentityRequiresSealedBoundedSnapshot | yes (caller freshness was accepted) | yes |
| B7 | unsupported source falls through | TestReserveAtFailsClosedOutsideWindowAndAfterTamper | yes (opaque reserve API absent) | yes |
| B8 | identity snapshot binding is invalid | TestIdentityRequiresSealedBoundedSnapshot | yes (caller freshness was accepted) | yes |

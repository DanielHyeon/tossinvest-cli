# Branch Test Map: `TestAdapterKeepsCredentialOutOfRequestMetadata`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | a scripted 200 makes the fetch succeed | `TestAdapterKeepsCredentialOutOfRequestMetadata` | existing coverage | PASS |
| B2 | the metadata serialises | `TestAdapterKeepsCredentialOutOfRequestMetadata` | existing coverage | PASS |
| B3 | neither the secret nor the request identity is in it | `TestAdapterKeepsCredentialOutOfRequestMetadata` | existing coverage | PASS |
| B4 | the transport did receive the secret | `TestAdapterKeepsCredentialOutOfRequestMetadata` | existing coverage | PASS |

# Branch Test Map: `TestExplicitOpaqueOriginCannotReachMutationHandler`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | explicit opaque origin receives origin-specific 403 | this function | safety behavior passed during RED | passed |
| B2 | opaque origin never reaches wrapped handler | this function | safety behavior passed during RED | passed |

# Branch Test Map: `snapshotDigest`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | same normalized query yields a deterministic preimage | existing deterministic snapshot tests | existing coverage | PASS |
| B2 | item ordering is deterministic and does not mutate the caller slice | existing deterministic snapshot tests | existing coverage | PASS |
| B3 | immutable Header scope/timestamp provenance changes alter digest and replay is refused | `TestDormantSnapshotReadRejectsTamperedHeaderScopeAndCutoffs` | replay incorrectly succeeded for all six tamper cases | PASS |

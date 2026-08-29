# Branch Test Map: `Snapshot.Digest`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | order rendering is deterministic | digest order-independence tests | preserve | yes |
| B2 | blank and exact-zero holding quantities have different digests and cannot corroborate; equal blank/equal zero can | `TestA110SnapshotDigestAndStabiliserDistinguishBlankHoldingFromExactZero` | yes (M26) | yes |
| B3 | balance rendering is deterministic | digest order-independence tests | preserve | yes |

# Branch Test Map: `snapshotDigest`

Rows corrected 2026-09-07. All three previously claimed coverage that measurement disproved; the RED
column now records what each mutation actually did. The scope half of the pair this function shares with
`snapshotItemMatchesQuery` is bound separately by `TestSnapshotItemScopeIsCheckedIndependentlyOfTheDigest`,
which holds the digest correct on purpose so only the scope check can refuse.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the six query fields are each bound: varying one with the item set held constant moves the digest | `TestSnapshotDigestBindsEveryQueryField`, `TestSnapshotDigestMatchesFrozenGoldenVector` | dropping issuer, mapping version and both cutoffs from the preimage left every test green; the old row cited "deterministic snapshot tests", which assert stability only — the trivially-passing direction | PASS |
| B2 | two items produce the ascending preimage whichever order they arrive in, and the caller's slice is untouched | `TestSnapshotDigestOrdersItemsAscendingWithoutMutatingCaller` | no test anywhere sealed a snapshot with two items, so the comparator never ran and flipping `<` to `>` was undetectable | PASS |
| B3 | every immutable `Header` field and the payload digest changes the snapshot identity | `TestSnapshotDigestBindsEveryHeaderFieldAndPayloadDigest`, `TestSnapshotDigestMatchesFrozenGoldenVector` | reducing the item preimage to `EvidenceID + PayloadDigest` — the pre-fix shape — left the package green. The old row cited `TestDormantSnapshotReadRejectsTamperedHeaderScopeAndCutoffs`, which cannot reach this branch: `snapshotItemMatchesQuery` refuses the same six tamper cases earlier, in the row loop, before the digest is recomputed | PASS |

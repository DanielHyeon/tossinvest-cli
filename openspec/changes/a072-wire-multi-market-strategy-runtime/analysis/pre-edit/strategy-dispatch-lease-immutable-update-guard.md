# Strategy dispatch lease immutable update guard — pre-edit map

## Function / migration logic

1. Released v25 transition validation permits only the state graph but does not freeze authority columns.
2. Add an append-only v28 migration; do not rewrite v25-v27 rows.
3. On every lease UPDATE, compare every identity, authority, lineage, timing, owner, reservation, and digest column byte-for-byte.
4. Permit changes only to lifecycle state/disposition/revision and the explicitly defined transport/outcome fields.
5. Abort the whole transaction on any rebind attempt.

## Branch Test Map

| Branch | Expected |
|---|---|
| Valid ISSUED→CLAIMED lifecycle-only update | accepted by v25 + v28 triggers |
| Campaign/decision/reservation/owner/authority rebind during transition | aborted |
| Lease digest or immutable timestamp rewrite | aborted |
| Migration from v27 | rows preserved, v28 trigger installed |

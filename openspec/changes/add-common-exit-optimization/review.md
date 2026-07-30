## Adversarial engineering proposal-freeze review

Status: accepted with the external-RUNNER correction below.

### Findings

1. Arbitrary rung editing would turn malformed UI input into live exit geometry. Rejected; only three immutable registry IDs are accepted.
2. Applying a changed common value to existing states would reinterpret an active rung. Rejected; policy ID is snapshotted per state.
3. Reading current config during adoption crash recovery creates a time-of-check/time-of-use policy change. Corrected; adoption transaction stores the selected ID and recovery reads that record.
4. Treating external RUNNER like self-entered RUNNER would create an unrequested 15% partial sell. Corrected; adopted RUNNER keeps floors but zeroes all partial ratios, matching StockOS A168.
5. Quiet fallback from an unknown non-empty ID would hide an operator/config error. Rejected; observer construction/evaluation refuses the unknown ID.
6. Reusing operating/adoption settings capability could accidentally expand the POST authority. Rejected; optimization receives a dedicated load/save-only seam.
7. Migration backfill would rewrite historical meaning and enlarge rollback risk. Rejected; v9 columns are nullable and legacy LADDER NULL is interpreted as `default_v1`.

### Approval

The proposal is frozen for implementation with no remote exposure, engine start, LIVE gate mutation, or account mutation in scope.

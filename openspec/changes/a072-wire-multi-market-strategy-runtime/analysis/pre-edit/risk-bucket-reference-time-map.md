# Function Logic Map: risk-bucket authority reference time binding

## `journal.validateRiskBucketAdmission`

Current flow validates ordered five-bucket identity, then compares both policy provenance and
snapshot provenance to the same `RiskBucketSnapshotReference.ObservedAt/FreshUntil` pair. Production
authority intentionally has an earlier intersected policy observation and a point-in-time snapshot
observation, so this shared pair rejects authentic production input at the horizon branch.

Change: retain the legacy pair as a compatibility fallback, add explicit policy and snapshot windows,
and compare each sealed provenance to its corresponding effective window. Market/symbol binding and
all other branches stay unchanged.

## `journal.insertFreshRiskBucketReservations` and q_final equivalent

Current flow writes policy time from the sealed bucket but snapshot time from the shared reference.
Change: write the explicit effective snapshot window. No released schema or historical row is edited.

# Branch Test Map

| Branch | Expected |
|---|---|
| legacy reference with only shared window | unchanged acceptance |
| explicit equal policy/snapshot windows | accepted |
| explicit distinct sealed policy/snapshot windows | accepted for KR and US |
| either explicit window differs from sealed provenance | `ErrRiskBucketSnapshotMismatch` |
| missing/invalid bucket identity, order, market or symbol | existing refusal unchanged |

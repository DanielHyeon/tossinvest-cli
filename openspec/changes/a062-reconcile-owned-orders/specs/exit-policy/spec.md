## ADDED Requirements

### Requirement: Automatic adoption obeys every authoritative RECONCILE state

Before pricing or adopting a holding, the engine SHALL use active journal RECONCILE states for the account as the authority and SHALL block every candidate covered by an account- or symbol-scoped state, regardless of which producer created the state. The runtime management projection SHALL expose the same covering states. Failure to read or update that authority MUST stop the cycle before adoption.

#### Scenario: Non-quantity reconcile cause covers a candidate
- **WHEN** an included or globally enabled holding is covered by `SNAPSHOT_UNAVAILABLE`, `SNAPSHOT_STALE`, `IDENTIFIER_CONFLICT`, or `ATTRIBUTION_FAILED`
- **THEN** the engine performs no adoption and no price read, while `/positions` reports the authoritative reconcile block rather than ordinary adoption pending

#### Scenario: Tracker persistence fails before adoption
- **WHEN** the reconciliation comparison cannot durably enter or release its block state
- **THEN** the cycle returns an error before candidate pricing and the holding remains unadopted

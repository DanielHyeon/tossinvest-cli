# Pre-Edit Snapshot Digest Integrity Gate

- Function: `internal/strategyevidence/store.go:snapshotDigest`
- Pre-edit source SHA-256: `5134fd3a8da3e06828b7cfe427f8cb7f5e494cf8b27fe9f4a8a307a5ec5df9a1`
- Existing callers: `Store.SealSnapshot` and `DormantSnapshotReadPort.Replay`.
- Risk: HIGH. The pre-edit item preimage binds only EvidenceID and payload digest, so direct evidence.db
  corruption of valid Header scope or timestamp fields can preserve the sealed snapshot digest.
- Planned change: canonicalize and hash the complete normalized immutable Header; independently enforce query
  scope, dual cutoffs and market-effective date during dormant replay.
- Explicit non-goals: no runtime activation, Guardian, dispatch, broker, apply-hook or toggle integration.

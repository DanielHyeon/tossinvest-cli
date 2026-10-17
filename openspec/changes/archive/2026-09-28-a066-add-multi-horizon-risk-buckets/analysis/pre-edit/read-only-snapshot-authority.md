# Read-only risk snapshot authority bundle

## Scope

This wave adds a pure read contract in `internal/riskbucket`. It does not add a
journal writer, broker transport, activation path, operating toggle, or live
order authority.

The contract accepts one exact KR or US scope and returns one sealed bundle
only when it can prove all of the following:

1. exactly one bucket exists for each dimension in canonical order:
   horizon, market, strategy, sector, symbol;
2. every bucket value matches the requested account/market/strategy scope;
3. the strategy bucket carries the requested strategy risk version;
4. policy and snapshot provenance are official, frozen, bound and fresh at the
   common `AsOf` boundary;
5. the reserve policy uses the exact scope currencies (KR/KRW or US/USD quote)
   and has fresh price/FX plus a complete fee policy;
6. every immutable journal reference repeats the exact bucket key, policy
   digest/times and snapshot ID/version/digest/times consumed by the bundle;
7. snapshot IDs and dimensions are unique; and
8. a canonical SHA-256 seal still matches the complete ordered preimage.

## Authority boundary

The source interface, raw material, material-entry constructor and per-entry
seal are package-private. Code outside `riskbucket` cannot implement the source
or promote an arbitrary `source` string into an authority. The public production
constructor returns `ErrRiskSnapshotAuthorityUnavailable` until a package-owned
loader is implemented.

The bundle stores all fields privately. Its getters return value copies, and
`Validate` repeats scope/provenance/reference checks before verifying the
canonical bundle seal.

## Failure behavior

- missing, duplicate or unexpected dimensions: typed fail-closed refusal;
- stale/invalid policy or snapshot provenance: typed provenance/freshness
  refusal;
- scope, market or currency mismatch: `ErrRiskSnapshotScopeMismatch`;
- mismatched or tampered journal reference: `ErrRiskSnapshotReferenceMismatch`;
- changed sealed bundle preimage: `ErrRiskSnapshotBundleTampered`;
- absent production loader: `ErrRiskSnapshotAuthorityUnavailable`.

KR material cannot satisfy a US request or validate as a US bundle, and the
reverse boundary is identical.

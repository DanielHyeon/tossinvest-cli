# KR/US read-only snapshot bridge checkpoint

## What landed

`internal/riskbucket` now defines the downstream read-only shape required by a
future strategyflow-to-q_final bridge: one market-scoped bundle with exactly
five ordered authoritative policy/snapshot bindings and matching immutable
journal references.

The same contract is tested for KR/KRW and US/USD, including cross-market
reuse refusal. The result is immutable-by-copy and sealed over the complete
scope, reserve policy, bucket amounts, provenance and journal-reference
preimage.

## What remains deliberately unavailable

- no package-owned production journal/snapshot loader;
- no strategyflow adapter or candidate promotion;
- no converted US Guardian-cap mint;
- no owner, lease, Gateway, broker or activation mutation; and
- no change to paired KR/US `OFF/UNOBSERVED` defaults.

Accordingly, the public service constructor fails closed. This checkpoint
defines and verifies the authority boundary needed for concurrent KR/US work;
it does not connect or activate either market lane.

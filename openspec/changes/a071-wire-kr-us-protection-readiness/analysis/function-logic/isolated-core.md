# Function Logic and Branch Test Map — isolated protection-readiness core

This wave adds `internal/protectionreadiness` only. Existing protection, gateway, engine and journal functions
are intentionally unchanged and their pre/post-edit maps remain pending until the integration wave.

## `Assess`

1. Start from a sealed paired KR/US `UNWIRED` snapshot and a copy of durable state.
2. Validate the pinned policy, durable state and sealed trusted-time observation.
3. Reject trusted-time rollback against the durable global floor. Otherwise advance the floor for every valid
   trusted-time observation, including when evidence is absent or invalid.
4. Evaluate only markets with evidence; missing peers remain independently `UNWIRED`.
5. For each market, call strict attestation verification, then publish `WIRED` only on success.
6. Advance `(account, profile, market)` serial only for accepted evidence; return a new sealed state without
   mutating the input. If the input state is corrupt, preserve its exact preimage and mark state commit forbidden.
7. Count the atomic pure durable-state transition as `Mutations=1` when state changed, while
   `ExternalMutations=0` always preserves existing protection and reduce-only exit intent.

Branches: missing/invalid evidence with floor advancement, invalid/corrupt state preservation, unavailable time,
time rollback, KR-only/US-only success, one-market key failure with valid peer, and exact sealed supervisor
binding success/failure.

## `verifyAttestation`

1. Validate sealed file bytes and exact resolved path, owner, `0600` mode, size, regular/non-symlink metadata.
2. Reject duplicate JSON keys before strict decode; reject unknown fields and non-canonical encoding separately.
3. Require schema v1 and explicit Ed25519 allowlist.
4. Resolve the pinned key ID, reject revoked/out-of-window keys and enforce bounded overlap.
5. Parse canonical UTC issue/expiry, reject invalid lifetime, future issue and expiry.
6. Require serial greater than the durable account/profile/market value, including across key rotation.
7. Verify the Ed25519 signature over the canonical body.
8. Require every broker client-key/echo/lookup/uniqueness/pending/terminal/cancel/dedup capability.
9. Match exact runtime account/profile/market/order/session/quantity/trigger/replace/tool/build/evidence/broker
   scope and the sealed supervisor binding.

Branches: schema, algorithm, key ID, revocation, overlap end, bad signature, maximum lifetime, future/expired,
serial replay, all exact scope fields, each broker capability field and supervisor forgery.

## Canonical JSON and anti-rollback helpers

The duplicate-key scanner recursively checks objects and arrays before `DisallowUnknownFields` decoding. The
entire envelope must equal its deterministic re-encoding. Durable maps are sealed in sorted scope order;
policy roots and paths are likewise canonicalized before sealing.

Fuzz properties prove arbitrary bytes never produce `WIRED` and a candidate serial is accepted exactly when
it is greater than the durable market serial. External-package tests prove public code cannot mint pinned
trust, observed evidence, supervisor binding, trusted time or durable state; it can only obtain the paired
`UNWIRED` default snapshot.

# Review — a071-wire-kr-us-protection-readiness

- Date: 2026-08-04
- Stage: isolated readiness core complete; runtime/controller/gateway integration pending
- Voices: Manager safety review, independent operations/security review, authority-boundary self-review

## Findings and disposition

- Readiness is market-scoped `WIRED|UNWIRED` plus typed refusal; no combined KR+US state exists.
- Signed attestation binds pinned trust roots, key ID/algorithm allowlist, rotation/revocation, monotonic
  serial, maximum lifetime and durable trusted-time floor.
- Scope includes exact broker client-key echo, lookup/uniqueness, pending/terminal/cancel query,
  dedup/idempotency and replace semantics. Missing capability is `UNWIRED`, not guessed support.
- Submit/cancel unknown and orphan orders use exact identity reconciliation only; no symbol/time inference,
  blind resubmit or inferred cancellation is allowed.
- The isolated core verifies a canonical JSON envelope signed with Ed25519. The pinned policy has an explicit
  algorithm allowlist, key lifecycle windows, revocation timestamps and bounded rotation overlap.
- Monotonic serial is durable at `(account, profile, market)` scope, so rotation to a different key ID cannot
  reset the counter. Trusted time has a sealed durable floor and assessment is a pure old-state/new-state step.
- **Resolved independent-review HIGH:** every valid non-rollback trusted-time observation advances the durable
  floor even when evidence is missing or invalid. A later rollback cannot hide inside an unattested interval.
- **Resolved independent-review HIGH:** corrupt durable state is returned as the exact non-committable preimage;
  assessment never clones and re-seals missing or modified serials into a repaired state.
- File bytes, resolved path, owner, exact `0600` mode, regular/symlink status and size are sealed modeled inputs;
  duplicate and unknown JSON fields are distinct typed refusals.
- Runtime scope binds exact account/profile/market/order/session/quantity/trigger/replace/tool/build/evidence and
  the complete broker client-key/lookup/uniqueness/query/dedup contract.
- A market becomes `WIRED` only when both attestation and an exact sealed supervisor binding validate. Snapshot
  fields and all authority-producing constructors are private; the only public constructor returns paired
  `UNWIRED` defaults.

## Verification

- Strict OpenSpec validation: PASS.
- `go test ./internal/protectionreadiness -count=1`: PASS.
- `go test -race ./internal/protectionreadiness -count=1`: PASS.
- `go vet ./internal/protectionreadiness`: PASS.
- `FuzzArbitraryAttestationNeverWires` and `FuzzSerialMustStrictlyIncrease` (3s each): PASS.
- Statement coverage: 87.6%.
- Static dependency/API tests exclude live transport, runtime mutation packages and exported trust/evidence/
  supervisor/state minting constructors.
- Existing protection controller, gateway, engine and journal integration tests remain pending by design.
- `Mutations` counts one atomic pure durable-state transition; `ExternalMutations` remains zero. A separate
  `StateCommitAllowed` bit prevents persistence of corrupt/untrusted-time results.

## Verdict

Isolated core approved for integration review. KR and US ship in the same release and default independently
to `UNWIRED`. Actual signed evidence and production supervisor wiring still require the remaining integration
work and a separate human-approved workflow; this core creates no lane, activation, automation or LIVE authority.

# Review — a071-wire-kr-us-protection-readiness

- Date: 2026-08-04
- Stage: decision-boundary integration complete; production controller/supervisor transport wiring still pending
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
- **Resolved independent-review HIGH:** the public `Options.ProtectionOverrideForTest` scalar forge and exported
  WIRED/UNWIRED test pointers were removed. `ProfileProtection` remains `UNWIRED` compatibility reporting only;
  engine status overrides are never forwarded to Gateway execution authority. Static and reflection tests reject
  any non-test declaration or exported scalar readiness field.
- Gateway admission uses a market-scoped sealed adapter, and the exact checkpoint generation/identity is
  revalidated immediately inside `DispatchVerified` before broker transport. Missing provider, attestation,
  supervisor evidence, scope mismatch, expiry, corruption or drift refuses only the exposure-raising mutation in
  that market; a peer-market seal remains independently usable.
- Reduce-only SELL, CANCEL and exposure-reducing AMEND succeed without reading the readiness provider. Static
  source isolation keeps stop, emergency liquidation, reconciliation and fill paths free of this dependency.
- The dormant lifecycle core derives stable operation identity from exact account, position, market, generation,
  revision and operation kind. Submit recovery uses that exact key; cancel and replacement recovery use the exact
  broker order ID. A scoped `NOT_FOUND` result must repeat every identity field before a same-key retry is even
  considered.
- Entry is closed while unprotected, pending, unknown, reconciling or terminal. It opens only for an unlatched
  ACTIVE order whose broker claim plus other sell claims exactly equals holdings. Registration and replacement
  therefore require full available coverage; partial fills reduce holdings and broker claim in the same pure
  transition and re-evaluate the equality.
- Unknown replacement and cancellation retain the pre-existing ACTIVE observation. Replacement is a single
  continuous-coverage command, refuses trigger retreat and never models cancel-then-place. Unknown submit without
  attested idempotency becomes no-resubmit reconciliation.
- Duplicate fills are stable-ID no-ops; conflicting duplicate fills latch reconciliation. Unowned orphan orders
  are never adopted, canceled or guessed, and a conflicting re-observation preserves the first evidence.
- The KR and US position maps share only a canonical container seal. A KR recovery latch does not alter US ACTIVE
  state, and exact fill/exit handling remains immediate in both markets.

## Verification

- Strict OpenSpec validation: PASS.
- `go test ./internal/protectionreadiness -count=1`: PASS.
- `go test -race ./internal/protectionreadiness -count=1`: PASS.
- `go vet ./internal/protectionreadiness`: PASS.
- Affected packages (`protectionreadiness`, `protection`, `execgw`, `reconcile`, `app/engine`) targeted and full
  tests: PASS; `cmd/tossctl`: PASS.
- Affected-package `go test -race`: PASS (`execgw` 264.135s, `reconcile` 101.866s, `app/engine` 288.622s).
- Affected packages plus `cmd/tossctl` `go vet`: PASS; strict OpenSpec: PASS; `git diff --check`: PASS.
- Post-edit Function Logic Maps are complete for every a071-owned changed function. The repository-wide checker
  reports only concurrent a066/a073 worktree functions outside this wave.
- `make sdd-sync` refreshed the CodeGraph fingerprint (11 changed files). Its advisory CodeGraphContext phase
  stalled for more than two minutes and was interrupted as permitted; no runtime or broker state was touched.
- `FuzzArbitraryAttestationNeverWires` and `FuzzSerialMustStrictlyIncrease` (3s each): PASS.
- Statement coverage: 87.6%.
- Static dependency/API tests exclude live transport, runtime mutation packages and exported trust/evidence/
  supervisor/state minting constructors.
- Existing protection controller, gateway, engine and journal integration tests remain pending by design.
- Isolated lifecycle unit/crash-matrix tests, race and vet: PASS.
- Lifecycle fuzz (`FuzzOperationIdentitySeparation`, `FuzzDuplicateFillNeverDoubleDecrements`, 3s each): PASS.
- Lifecycle statement coverage: 83.4%.
- Post-edit Go AST evidence: `lifecycle.go` SHA-256
  `de50441bc89c79ec5cfeb8308a837db1cffede7e3ab52c661eccb9515d1688e5`; `state.go` SHA-256
  `df5c5459c6d2add80bcfdcadb03f4af1dbcb19cce1424687c2d855a40fd66cbe`.
- Test-only fake official broker has no socket or hostname. Static dependency/API tests reject live transport,
  runtime protection/execgw/engine/journal/broker imports, toggle/lane/LIVE approval authority and exported
  authority-minting functions.
- `Mutations` counts one atomic pure durable-state transition; `ExternalMutations` remains zero. A separate
  `StateCommitAllowed` bit prevents persistence of corrupt/untrusted-time results.

## Verdict

The isolated cores and Gateway decision boundary are approved for the next integration review. KR and US ship in
the same release and default independently to `UNWIRED`. Actual signed evidence and production supervisor/official
gateway wiring remain task 3.5 and require a separate human-approved workflow; this wave creates no lane,
activation, automation or LIVE authority and makes no official order call.

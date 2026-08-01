# Dormant implementation evidence

Date: 2026-07-31

## Approved boundary

This implementation now contains the complete testable protection core: signed capability
verification, durable create/replace/cancel attempts, response-loss recovery, reconciliation,
the typed a041 protection-line bridge, cancel-first flatten authorization, and an official-only
adapter. It still has no production activation capability: `Activation` has no exported minter,
`cmd/` and `internal/app` do not construct the controller or adapter, and the shipped
`execgw.ProfileProtection = ProtectionUnwired` constant is unchanged. Consequently no code in
the shipped application can dispatch these broker mutations.

## CodeGraph hard evidence

The worktree index was refreshed with `make sdd-sync` before editing.

- `Attestation.Verify` impact reaches the legacy execution attestation writer/reader,
  `internal/soak.BuildAttestation`, console display, and engine interlock tests (23 symbols in
  the pre-edit CodeGraph impact). To preserve all of those callers, the protection matrix is
  a separate file and type; no existing `attest.Load`, `Save`, `Verify`, or `Attestation`
  function/type was edited.
- `ProfileProtection` resolves to the constant in `internal/execgw/protection.go`, is forwarded
  by the engine interlock, and gates `Gateway.checkProtection` before exposure-raising broker
  dispatch. It was not edited. The new dormant static test also requires it to remain
  `UNWIRED` and rejects any `cmd/` or `internal/app/` import of `internal/protection`.
- Existing conditional mutations remain the concrete official methods in
  `internal/official/conditional_writes.go`. The new `internal/protectionofficial` adapter calls
  only that official client, requires exact integer/capability/readback identity, scans bounded
  OPEN+CLOSED pages for recovery, and treats cancel disappearance or unknown lifecycle as
  `IN_DOUBT`. It imports neither WTS/hybrid transports nor `internal/trading`.
- The new protection attempt journal is additive v13 and is written before dispatch. Protection
  quantity converges from holding minus the single open-sell/local claim and refuses oversell.
  Cancel-first flatten issues only an exact-scope, exact-broker, exact-quantity, two-second,
  one-shot authorization after terminal non-triggered cancel and authoritative sellable reads.
  The authorization itself has no sell method and the dormant application has no consumer.

## Go AST / Function Logic Map decision

Function Logic Map: not-applicable — the diff adds new Go files and changes no existing Go
function body. Consequently there is no pre-existing branch, early return, mutation, fallback,
or side effect whose map could be updated. New functions are covered directly by table-driven,
adversarial, repository, static-wiring, and race tests.

The future tasks that edit official conditional writes, engine gateway/reservations,
reconciliation, or flatten remain unchecked and will require their own per-symbol AST,
Function Logic Map, Branch Test Map, risk report, and human-attested evidence before editing.

### Security follow-up mapping

The independent review required changes to functions introduced by the first dormant commit.
Before those edits, AST, Function Logic Map, Branch Test Map, and risk reports were created under
`analysis/function-logic/` for attestation parsing/validation/scope checks, saga validation and
transition, sell-claim and flatten decisions, reconciliation comparison, and repository update.
The maps record the RED cases for external evidence recomputation, file/parent integrity, strict
account grammar, state-specific saga fields, immutable persisted identity, typed exact scope,
duplicate broker IDs, ordered two-second flatten observations, and int64 boundaries. They replace
the original new-functions-only exemption for this follow-up diff; the initial commit's statement
remains an accurate record of its own pre-edit gate.

## Dormant TDD coverage

- strict version, unknown/trailing JSON, issue/expiry window, exact account/profile/market/
  session/type/order/trigger/quantity scope, both verifier tool builds and SHA-256 evidence
  digests;
- symlink, non-regular file, mode, owner, open-file identity, and size checks;
- PLANNED/REGISTERING/ACTIVE/REPLACING/IN_DOUBT transitions, response-loss quarantine,
  monotonic trigger replacement, and old/new broker identity retention;
- one-share protection and aggregate sell-claim oversell refusal;
- first-fill 1-second arm and 2-second ACTIVE boundary behavior;
- missing/duplicate/orphan/quantity/trigger discrepancy classification;
- 2-second flatten decision with terminal cancel and authoritative sellable quantity, returning
  `IN_DOUBT` for response loss, trigger race, lateness, or insufficient quantity;
- additive SQLite repository round-trip and optimistic-revision conflict;
- static absence of engine/command wiring, production HTTP/official/trading adapters, and
  paper/shadow/canary protection paths.

The security follow-up additionally covers external evidence byte absence/tampering/extras,
canonical capability-digest binding, exact evidence basename, parent/file link and permission
checks, strict protection-only account normalization, cross-account/profile/market/symbol input,
duplicate broker identity, invalid saga state-field combinations, persisted identity/state jumps,
observation ordering and identity, and overflow-safe maximum-int64 sell claims.

## Intentionally incomplete

Production policy/trust-root/envelope provisioning, an exported activation minter, app/cmd
construction, an execution-gateway readiness flip, a runtime toggle, and all LIVE handoff work
remain intentionally absent. Signed readiness can be evaluated but cannot activate the controller.
The console therefore defaults to `OFF / 지원 확인 전 사용 불가`; it accepts no free symbol,
trigger, quantity, reason, or typed confirmation, and weakening is limited to a server-issued
opaque capability plus checkbox and three-second delay.

After the `4679fe2` independent review, restart trust was tightened further: durable rows never open
entry by themselves. A current controller instance must exactly re-confirm every ACTIVE broker row;
all other states and all recovery/I/O errors keep the latch closed. Broker calls now carry bounded
one-/two-second operation budgets, and cancel/recovery compare a typed durable target rather than an
identifier alone. These runtime contracts remain dormant because the shipped application still has
no activation construction path.

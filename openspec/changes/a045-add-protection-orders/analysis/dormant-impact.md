# Dormant implementation evidence

Date: 2026-07-31

## Approved boundary

This implementation stops before external capability attestation. It adds a strict parser,
pure state/reconciliation decisions, an additive repository schema, and a gateway interface
whose only implementation is test code. It does not import the package from `cmd/` or
`internal/app`, does not add an official client adapter, and does not change the shipped
`execgw.ProfileProtection = ProtectionUnwired` constant.

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
- Existing conditional mutations are the concrete methods in
  `internal/official/conditional_writes.go` and the gated service in
  `internal/trading/conditional.go`. Neither package was edited or imported by the dormant
  protection package.
- Existing reservations are the entry-exposure ledger in `internal/journal/reservations.go`,
  checked at `internal/execgw/gateway.go:checkReservation`. Existing flatten and reconciliation
  paths are also outside this diff. The new sell-claim, discrepancy, and flatten-decision code
  is pure classification only and cannot dispatch a mutation.

## Go AST / Function Logic Map decision

Function Logic Map: not-applicable — the diff adds new Go files and changes no existing Go
function body. Consequently there is no pre-existing branch, early return, mutation, fallback,
or side effect whose map could be updated. New functions are covered directly by table-driven,
adversarial, repository, static-wiring, and race tests.

The future tasks that edit official conditional writes, engine gateway/reservations,
reconciliation, or flatten remain unchecked and will require their own per-symbol AST,
Function Logic Map, Branch Test Map, risk report, and human-attested evidence before editing.

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

## Intentionally incomplete

External evidence production, real official gateway calls, sell-reservation integration,
engine/a041 wiring, operational reconciliation/flatten, `ProtectionReady=WIRED`, console
activation, and all LIVE handoff work remain blocked by the review's human-attestation gate.

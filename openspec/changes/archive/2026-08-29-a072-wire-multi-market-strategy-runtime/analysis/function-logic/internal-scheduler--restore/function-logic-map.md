# Function Logic Map: `Restore`

- Source: `internal/scheduler/desired.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Restore is fail-closed unless desired/current bindings match and a verifier returns signed non-zero generation with a future expiry.

## Branches and early returns

B1–B13 cover disabled desired state, invalid desired/current bindings, revision/config/calendar/build mismatches, missing verifier, verifier failure, zero/expired evidence and exact success.

## Calls and live bindings

Calls package-private `ActivationVerifier.verifyActivation`; production uses owner-only Ed25519 manifest bytes pinned by digest and environment trust key.

## State mutations and fallbacks

No desired state is written. Success mints only opaque read authority bound to the exact activation binding and evidence lifetime.

## Safety conclusion

The added generation/expiry evidence narrows dispatch authority and cannot activate a lane.

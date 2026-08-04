# Function Logic Map: `newRootCmd`

- Source: `cmd/tossctl/root.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

Root construction registers commands but executes none. The new performance projector is a separate derived-data command annotated non-account-mutating.

## Branches and early returns

B1–B7 cover persistent output/session/update/config/onboarding setup and optional first-run hints; command registration itself is branchless data assembly.

## Calls and live bindings

`newPerformanceCmd` exposes only journal read-only to performance derived-store projection; it receives no application/broker context.

## State mutations and fallbacks

No command runs during construction. Existing cache notifications remain the only pre/post-run local writes.

## Safety conclusion

Registration cannot issue an order or activate KR/US lanes.

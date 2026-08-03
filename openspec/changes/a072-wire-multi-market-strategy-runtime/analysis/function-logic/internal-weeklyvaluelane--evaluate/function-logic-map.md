# Function Logic Map: `evaluate`

- Source: `internal/weeklyvaluelane/evaluate.go`
- AST evidence: `ast.json`

## Inputs and invariants

Consumes sealed weekly plan/evidence/calendar/reservation/cap, explicit entry/staged target, and a package-private saved-stop authority when the saved stop wins. `CalculateRR` already computes exact capped target.

## Branches and early returns

B1-B22 cover lineage, authorization, schema, scope, invalidation, evidence, calendar, leg, reservation, cap and stop selection. B23 rejects a selected saved stop without a plan/evidence/price-bound private seal. B24-B31 cover canonical terms, stop distance, RR, risk and scale refusals. B32 replaces candidate provenance with sealed saved-stop provenance exactly when the saved stop wins.

## Calls and live bindings

Pure evidence evaluator, market-week validator, sealed stop-provenance selection, capped RR and risk admission only.

## State mutations and fallbacks

None. Accepted output must copy entry, effective stop and `RRResult.TargetMinor`; it must not copy uncapped staged target or candidate provenance for a different saved-stop price.

## Safety conclusion

Safe edit boundary: preserve validated RR terms and exact selected-stop authority in Outcome. High-risk impact: all execution prices and provenance identities are used by later dispatch.

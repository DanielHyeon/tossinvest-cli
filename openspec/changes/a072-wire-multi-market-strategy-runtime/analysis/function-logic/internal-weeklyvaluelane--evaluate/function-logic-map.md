# Function Logic Map: `evaluate`

- Source: `internal/weeklyvaluelane/evaluate.go`
- AST evidence: `ast.json`

## Inputs and invariants

Consumes sealed weekly plan/evidence/calendar/reservation/cap, explicit entry/staged target, and a package-private saved-stop authority. The public saved-stop scalar is compatibility input only and cannot select or weaken the saved price. `CalculateRR` computes the exact capped target.

## Branches and early returns

B1-B21 cover lineage, authorization, schema, scope, invalidation, evidence, calendar, leg, reservation and cap. B22 rejects a public saved scalar without a valid private authority. B23-B24 validate stop selection and its sealed authority. B25-B32 cover canonical terms, stop distance, RR, risk and scale refusals. B33 replaces candidate provenance with sealed saved-stop provenance exactly when the saved stop wins.

## Calls and live bindings

Pure evidence evaluator, market-week validator, private saved-stop price/provenance selection, capped RR and risk admission only.

## State mutations and fallbacks

None. Accepted output must copy entry, private-authority-derived effective stop and `RRResult.TargetMinor`; it must not copy uncapped staged target or candidate provenance for a different saved-stop price.

## Safety conclusion

Safe edit boundary: preserve validated RR terms and exact selected-stop authority in Outcome. High-risk impact: all execution prices and provenance identities are used by later dispatch.

# Function Logic Map: `evaluate`

- Source: `internal/weeklyvaluelane/evaluate.go`
- AST evidence: `ast.json`

## Inputs and invariants

Consumes sealed weekly plan/evidence/calendar/durable reservation, explicit entry/staged target, and a
package-private saved-stop authority. Admitted evaluation additionally consumes a sealed a066 cap;
proposal evaluation has no cap and uses current planned remaining as `q_candidate`. The public
saved-stop scalar is compatibility input only and cannot select or weaken the saved price.
`CalculateRR` computes the exact capped target.

## Branches and early returns

B1-B19 cover lineage, authorization, schema, scope, invalidation, evidence, calendar, leg and durable
reservation. B20 computes exact planned remaining. Proposal keeps it as `q_candidate`; admitted
evaluation validates the cap and applies q_final. B22 rejects a public saved scalar without a valid
private authority. B23-B24 validate stop selection and its sealed authority. B25-B30 cover canonical
terms, stop distance and RR. Admitted evaluation alone performs proposed-risk admission; both modes
validate currency scale and B33 replaces candidate provenance with sealed saved-stop provenance exactly
when the saved stop wins.

## Calls and live bindings

Pure evidence evaluator, market-week validator, private saved-stop price/provenance selection, capped RR and risk admission only.

## State mutations and fallbacks

None. Proposal carries no cap snapshot/reservation lineage but still requires the durable weekly
reservation. Accepted output must copy entry, private-authority-derived effective stop and
`RRResult.TargetMinor`; it must not copy uncapped staged target or candidate provenance for a different
saved-stop price.

## Safety conclusion

Safe edit boundary: share all evidence/calendar/reservation/stop/RR validation and branch only around
cap-specific quantity, max-stop-distance authority and proposed-risk admission. Proposal needs an
equivalent non-cap stop-distance policy authority; it may not silently skip the structural stop bound.
High-risk impact: all execution prices and provenance identities are used by later dispatch.

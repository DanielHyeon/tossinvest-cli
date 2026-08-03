# Function Logic Map: `evaluate`

- Source: `internal/weeklyvaluelane/evaluate.go`
- AST evidence: `ast.json`

## Inputs and invariants

Consumes sealed weekly plan/evidence/calendar/reservation/cap and explicit entry/staged target. `CalculateRR` already computes exact capped target.

## Branches and early returns

B1-B25 cover lineage, authorization, schema, scope, invalidation, evidence, calendar, leg, reservation, cap, stop, distance, RR and risk refusals.

## Calls and live bindings

Pure evidence evaluator, market-week validator, stop selection, capped RR and risk admission only.

## State mutations and fallbacks

None. Accepted output must copy entry, effective stop and `RRResult.TargetMinor`; it must not copy uncapped staged target.

## Safety conclusion

Safe edit boundary: preserve already validated RR terms in Outcome. High-risk impact: target identity used by later dispatch.

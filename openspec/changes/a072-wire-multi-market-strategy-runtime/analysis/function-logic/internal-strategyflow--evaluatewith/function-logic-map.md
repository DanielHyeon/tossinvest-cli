# Function Logic Map: `evaluateWith`

- Source: `internal/strategyflow/flow.go`
- AST evidence: `ast.json`

## Inputs and invariants

`evaluateWith` accepts one sealed approved candidate, exact router request and tagged lane input. Accepted output requires matching candidate/router/lane/campaign/leg/risk lineage and will additionally require exact canonical entry/stop/target terms.

## Branches and early returns

B1-B7 reject invalid candidate/scope/router/descriptor/input/registry before lane acceptance. B8-B11 reject native lane refusal, lineage mismatch, owner mismatch and incomplete campaign/leg/risk data. The planned branch rejects missing or invalid execution terms before sealing a complete result.

## Calls and live bindings

Calls the fixed router and exactly one registry evaluator. It must not call Guardian, journal, broker, configuration writer or activation APIs.

## State mutations and fallbacks

Pure value composition only. Refusals preserve zero Guardian calls, broker calls and mutations. No price or target is inferred.

## Safety conclusion

Safe edit boundary: require lane-provided validated terms and seal them against exact accepted lineage and quantity. High-risk impact: accepted decisions feed later dispatch, so malformed or absent terms fail closed.

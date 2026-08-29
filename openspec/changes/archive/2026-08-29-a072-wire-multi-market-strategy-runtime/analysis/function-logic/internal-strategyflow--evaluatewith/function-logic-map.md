# Function Logic Map: `evaluateWith`

- Source: `internal/strategyflow/flow.go`
- AST evidence: `ast.json`

## Inputs and invariants

`evaluateWith` accepts one sealed approved candidate, exact router request and tagged lane input. Accepted output requires matching candidate/router/lane/campaign/leg/risk lineage, exact canonical `strategyrouter.RouterID` plus `RouterRelease`, and validated entry/stop/target terms. Both router fields participate in lineage sealing.

## Branches and early returns

B1-B7 reject invalid candidate/scope/router/descriptor/input/registry before lane acceptance. Canonical descriptor validation rejects a missing/forged fixed router identity or release before this function can accept it. B8-B11 reject native lane refusal, lineage mismatch, owner mismatch and incomplete campaign/leg/risk data. The planned branch rejects missing or invalid execution terms before sealing a complete result; success writes both canonical router fields into the sealed lineage.

## Calls and live bindings

Calls the fixed router and exactly one registry evaluator. It must not call Guardian, journal, broker, configuration writer or activation APIs.

## State mutations and fallbacks

Pure value composition only. Refusals and accepted results preserve zero Guardian calls, broker calls and mutations. No price, target, router identity or router release is inferred from market/horizon.

## Safety conclusion

Safe edit boundary: require lane-provided validated terms and the fixed router registry identity/release, then seal them against exact accepted lineage and quantity. High-risk impact: accepted decisions feed later dispatch, so malformed or absent terms and forged/missing router descriptors fail closed.

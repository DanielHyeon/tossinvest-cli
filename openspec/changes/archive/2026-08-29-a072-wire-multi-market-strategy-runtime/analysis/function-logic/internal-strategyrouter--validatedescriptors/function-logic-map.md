# Function Logic Map: `ValidateDescriptors`

- Source: `internal/strategyrouter/registry.go`
- Qualified function: `ValidateDescriptors`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

The descriptor slice must contain exactly one KR and one US row, exact router/release constants, OFF/OFF
desired/effective state and UNOBSERVED runtime state.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | row count differs from two | none | paired-shipping error | missing/extra tests |
| B2 | iterate each descriptor | local seen map only | continue/error | paired table tests |
| B3 | invalid/duplicate market, identity/release drift, non-OFF or observed runtime | none | invalid dormant descriptor | substitution matrix |
| B4 | final seen set lacks KR or US | none | paired market missing | peer omission tests |

## Calls and live bindings

Only `validMarket` and package constants; no runtime setting or activation source is consulted.

## State mutations and fallbacks

Mutates a local seen map only. It cannot repair rows, default a missing market or activate either descriptor.

## Safety conclusion

- Safe edit boundary: reject any one-market, duplicate, version-drifted or non-dormant descriptor set.
- High-risk impact: yes for release truth and KR/US paired-delivery enforcement.

# Function Logic Map: `DormantRuntimeDescriptor`

- Source: `internal/strategyengine/runtime_descriptor.go`
- CodeGraph callers/callees: authenticated read-only `/strategy-runtime`
- AST: generated after implementation

## Inputs and invariants

| Input/state | Range | Source of truth | Failure behavior |
|---|---|---|---|
| parameter cards | every frozen lane constant, server-owned and read-only | StockOS commit/source digest | malformed descriptor makes console fail closed |
| blockers | nine closed-vocabulary authority states | dormant runtime contract | effective entry OFF |
| actions | section ownership only; no action/control/input | a050 handoff | GET/HEAD projection only |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Test |
|---|---|---|---|---|
| Success | fixed descriptor | none | immutable descriptor | descriptor/DOM tests |

## Calls and live bindings

| Callee | Contract | Failure path | Evidence |
|---|---|---|---|
| `field` | complete label/help/default/desired/effective/unit/range/provenance/timing | console rejects incomplete/reordered data | tests |

## State mutations and fallbacks

- Returns fixed read-only metadata. Effective values remain unconfigured/OFF; there is no action or fallback.

## Safety conclusion

- Read-only metadata only; no arbitrary input, write route, order, or toggle.

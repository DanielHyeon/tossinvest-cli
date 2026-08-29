# Function Logic Map: `Descriptors`

- Source: `internal/strategyrouter/registry.go`
- Qualified function: `Descriptors`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

No inputs. The immutable release contract is exactly two descriptors: KR and US, same router ID/release,
both desired/effective OFF and runtime UNOBSERVED.

## Branches and early returns

Branchless literal construction; no environment/config value can turn either market ON.

## Calls and live bindings

None. Constants in the package are the only bindings.

## State mutations and fallbacks

Returns a fresh slice and mutates no global/runtime state. No single-market fallback exists.

## Safety conclusion

- Safe edit boundary: KR and US must ship as one dormant pair.
- High-risk impact: yes for defaults — accidental ON/effective state would synthesize authority.

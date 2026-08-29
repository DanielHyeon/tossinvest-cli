# Function Logic Map: `canonicalProtectionQuantity`

- Source: `internal/execgw/protection.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| order intent quantity | positive canonical integral float64 up to 2^53-1 | signed order intent | return `(0,false)` before provider/broker |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | decimal text empty or signed/fractional/exponent form | none | reject | canonical quantity unit matrix |
| B2 | parse fails, zero, above safe integer, or round-trip differs | none | reject | canonical quantity unit matrix |
| B3 | exact positive safe integer | none | uint64 plus true | canonical quantity unit matrix |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decimalString`, `strconv.ParseUint` | canonicalize without float rounding authority | any failure rejects | CodeGraph + AST |

## State mutations and fallbacks

- Pure conversion; never calls readiness provider or broker.

## Safety conclusion

- Safe edit boundary: exact integral quantity validation before readiness
- High-risk impact: yes; fail closed before transport

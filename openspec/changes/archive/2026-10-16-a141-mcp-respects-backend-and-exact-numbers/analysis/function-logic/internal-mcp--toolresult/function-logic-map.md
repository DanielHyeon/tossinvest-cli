# Function Logic Map: `toolResult`

- Source: `internal/mcp/server.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| payload | JSON-marshallable value | operation handler result | marshal failure returns RPC invalid-params error |
| encoded payload | <= or > `maxResultBytes` | first JSON encoding | small payload passes through unchanged |
| oversized JSON | valid JSON with arbitrary number lexemes | decoder used for trimming | retained numbers must remain JSON numbers with exact value |
| truncation result | arrays may be shortened | `shrinkToCap` / `trim` | valid JSON and omission marker required |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | initial `json.MarshalIndent` fails | no output construction | RPC invalid-params error | Existing error behavior retained |
| B2 | encoded bytes exceed cap | enter decode/trim/re-encode path | fall through if decoding cannot proceed | RED: oversized nested exact-number payload |
| B3 | `Decoder.UseNumber` decode succeeds | feeds exact-number tree to `shrinkToCap` | retained tokens remain json.Number | exact-token regression |
| B4 | trimmed exact-number value re-encodes | replaces text | cap and marker behavior retained | cap and exact-number regressions |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.MarshalIndent` | Initial and final JSON text | errors must retain current RPC behavior | AST B1/B4 |
| exact-number decoder | Build trim tree without float64 coercion | malformed JSON leaves original text behavior intact | edit target at AST B3 |
| `shrinkToCap` | Remove array tail elements only | no number conversion | AST B3/B4 |

## State mutations and fallbacks

- No external side effect. It changes only the JSON text returned to the MCP host.
- The size cap and `_omitted_items` compatibility marker are invariant.

## Safety conclusion

- Safe edit boundary: decode path entered only after the existing size-cap decision; do not alter catalog dispatch or response schema.
- High-risk impact: yes. Rounded IDs and decimal values can cause a financial agent to act on a different value.

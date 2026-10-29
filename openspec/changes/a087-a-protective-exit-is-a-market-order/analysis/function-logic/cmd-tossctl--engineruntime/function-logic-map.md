# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go:655-752`
- AST evidence: `ast.json` (Pre-Edit HEAD `a284a2b0a`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Context engine | validated owned engine | `NewContext` | existing runtime assembly error |
| logger | optional structured sink | CLI runtime | exit order log preserves whitelist |
| lower-limit refresher | Context auxiliary, dormant when nil admission | Context board | runtime never enables admission |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B5 | detector/recovery/loop assembly failures | existing cleanup/error wrapping | error return | existing runtime tests |
| B6 | runtime construction failure | none | error return | existing runtime tests |
| success | auxiliary list construction | add dormant refresher | runtime | a087 auxiliary structural test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ectx.ExitObserver` | build protection loop | existing error return | AST setup |
| `ectx.ManualBuyRefresher` | current auxiliary | unchanged | final literal |
| `ectx.ExitLowerLimitRefresher` | start dormant consumer | no broker call without admission | a087 insertion |
| `engine.NewRuntime` | own loops and auxiliaries | B6 | AST B6 |

## State mutations and fallbacks

- only append an Auxiliary executor; supervised loops and order builders stay unchanged.
- executor with nil admission blocks on context and makes zero broker calls.

## Safety conclusion

- Safe edit boundary: additive `Auxiliary` literal item.
- High-risk impact: yes — test must prove observer begins quiet and engine wiring remains nil-admission.

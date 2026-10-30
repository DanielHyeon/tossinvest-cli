# Function Logic Map: `reportCommand`

Source: `tools/a128-jev-shadow-probe/main.go`. Current AST and source digest: `ast.json`. Pre-edit snapshot: `../../pre-edit-ast/tools-a128-jev-shadow-probe--reportcommand.json`.

## Inputs and invariants

ledger path, report flags and output writer. Reads local ledger and states; explicit legacy path keeps original output, default emits comparison after calibration. Public-only data and isolated experimental output remain invariants.

## Branches and early returns

| Branch | Source condition |
|---|---|
| B1 | if at 237:2: `if err := flags.Parse(args); err != nil {` |
| B2 | if at 242:3: `if f.Name == "ledger" {` |
| B3 | if at 246:2: `if !explicitLedger {` |
| B4 | if at 250:2: `if err != nil {` |
| B5 | if at 253:2: `if *compare {` |
| B6 | if at 260:2: `if err == nil && !explicitLedger {` |

## Calls and live bindings

The complete current call enumeration is in `ast.json`. The current source is authoritative; observed call context includes the local ledger, profile helper, and existing probe/report operations appropriate to this function. CodeGraph source context located probe.cycle and assembleState; precise isolated checkout graph freshness was unavailable.

## State mutations and fallbacks

Reads local ledger and states; explicit legacy path keeps original output, default emits comparison after calibration. Failures remain explicit; no production order or runtime configuration changes are introduced.

## Safety conclusion

Shadow measurement only. Existing broker/source calls are read-only; validation uses fake transports. Full production high-risk mutation review is not applicable.

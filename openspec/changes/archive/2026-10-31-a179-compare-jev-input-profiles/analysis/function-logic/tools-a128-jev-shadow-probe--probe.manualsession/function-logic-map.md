# Function Logic Map: `probe.manualSession`

Source: `tools/a128-jev-shadow-probe/collect.go`. Current AST and source digest: `ast.json`. Pre-edit snapshot: `../../pre-edit-ast/tools-a128-jev-shadow-probe--probe.manualsession.json`.

## Inputs and invariants

manual window, market and injected clock. Builds market-local open and close and appends one session gap; experimental rows carry cohort. Public-only data and isolated experimental output remain invariants.

## Branches and early returns

| Branch | Source condition |
|---|---|
| B1 | if at 234:2: `if err != nil {` |
| B2 | if at 244:2: `if p.cfg.InputProfile != "" && p.cfg.InputProfile != "current" {` |
| B3 | if at 247:2: `if err := p.ledger.gap(row); err != nil {` |

## Calls and live bindings

The complete current call enumeration is in `ast.json`. The current source is authoritative; observed call context includes the local ledger, profile helper, and existing probe/report operations appropriate to this function. CodeGraph source context located probe.cycle and assembleState; precise isolated checkout graph freshness was unavailable.

## State mutations and fallbacks

Builds market-local open and close and appends one session gap; experimental rows carry cohort. Failures remain explicit; no production order or runtime configuration changes are introduced.

## Safety conclusion

Shadow measurement only. Existing broker/source calls are read-only; validation uses fake transports. Full production high-risk mutation review is not applicable.

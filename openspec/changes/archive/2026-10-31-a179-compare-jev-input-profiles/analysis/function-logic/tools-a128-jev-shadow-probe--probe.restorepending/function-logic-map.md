# Function Logic Map: `probe.restorePending`

Source: `tools/a128-jev-shadow-probe/collect.go`. Current AST and source digest: `ast.json`. Pre-edit snapshot: `../../pre-edit-ast/tools-a128-jev-shadow-probe--probe.restorepending.json`.

## Inputs and invariants

read ledger judgment and label rows. Skips labeled or invalid due times; groups experimental rows by cohort/source/base and preserves existing legacy grouping. Public-only data and isolated experimental output remain invariants.

## Branches and early returns

| Branch | Source condition |
|---|---|
| B1 | range at 802:2: `for _, row := range contents.Judgments {` |
| B2 | if at 803:3: `if _, labeled := contents.Labels[row.ID]; labeled {` |
| B3 | if at 807:3: `if err != nil {` |
| B4 | if at 811:3: `if row.Cohort != "" {` |
| B5 | if at 814:3: `if groups[key] == nil {` |
| B6 | range at 820:2: `for _, key := range order {` |

## Calls and live bindings

The complete current call enumeration is in `ast.json`. The current source is authoritative; observed call context includes the local ledger, profile helper, and existing probe/report operations appropriate to this function. CodeGraph source context located probe.cycle and assembleState; precise isolated checkout graph freshness was unavailable.

## State mutations and fallbacks

Skips labeled or invalid due times; groups experimental rows by cohort/source/base and preserves existing legacy grouping. Failures remain explicit; no production order or runtime configuration changes are introduced.

## Safety conclusion

Shadow measurement only. Existing broker/source calls are read-only; validation uses fake transports. Full production high-risk mutation review is not applicable.

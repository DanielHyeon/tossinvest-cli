# Function Logic Map: `main`

- Source: `tools/logic-map/check_analysis.py` (Python pre-edit evidence)
- AST evidence: `ast.json`

## Inputs and invariants

Parses `--change` and `--root`, delegates all validation to `check`, and keeps
the CLI exit contract: diagnostics are nonzero and a clean ordinary check exits
zero.

## Branches and early returns

| Branch | Source | Condition | Result |
|---|---:|---|---|
| B1 | 625 | `check` returned errors | Print each diagnostic and return 1. |
| B2 | 626 | Iterate diagnostics | Print every checker diagnostic before returning 1. |
| success | 629 | No errors | Print the generic completion line and return 0. |

## Calls and live bindings

- `argparse.ArgumentParser`, `check`, and `print` only.
- No checkout, write, runtime, account, order, or network operation occurs.

## State mutations and fallbacks

No filesystem state changes. The success text must not turn record presence
alone into an adoption claim; any adoption label must follow successful
validator-derived context.

## Safety conclusion

This is a completion-gate CLI only. Preserve the ordinary success API while
making a successful, validated adoption distinguishable.

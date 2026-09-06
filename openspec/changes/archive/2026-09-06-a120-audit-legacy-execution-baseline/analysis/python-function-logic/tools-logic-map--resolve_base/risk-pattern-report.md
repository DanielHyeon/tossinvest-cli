# Risk Pattern Report: `resolve_base`

- Source: `tools/logic-map/check_analysis.py`
- Language: Python; this pre-edit artifact uses the Python `ast` inventory in `ast.json`. The repository risk-pattern helper targets Go input and is therefore not applicable to this source.

## Findings

| Pattern | Classification | Evidence |
|---|---|---|
| Shell/Git subprocess boundary | reviewed-safe before edit | AST call inventory and current source; all non-zero returns are fail-closed where this function owns validation. |
| Repository write/checkout mutation | reviewed-safe before edit | `resolve_base` and `check` do not write; `changed_existing_functions` only creates an OS temporary base file and removes it. |
| Live runtime/account/order side effect | not-applicable | No imports or calls reach TossOS trading runtime. |

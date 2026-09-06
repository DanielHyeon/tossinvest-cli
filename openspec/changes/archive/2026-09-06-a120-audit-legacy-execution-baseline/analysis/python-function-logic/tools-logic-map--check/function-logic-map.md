# Function Logic Map: `check`

- Source: `tools/logic-map/check_analysis.py` (Python; current AST range is recorded in `ast.json`)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

Pre-edit Python AST evidence is intentionally stored here because this change changes SDD tooling only. The Go-only `check_analysis.py` completion validator does not consume these Python artifacts. The a120 review must retain `Function Logic Map: not-applicable` for a120's Go source set; these maps do not waive the Python review obligation.

## Inputs and invariants

Change directory, persisted base, optional reference file, review text, and all analysis bundles under the current repository root.

## Branches and early returns

| Branch | Source | Condition | Mutation/side effect | Return/error | Baseline test |
|---|---:|---|---|---|---|
| B1 | 516 | Base/function derivation raises | Return a diagnostic list; no successful empty-change interpretation. | `CheckAnalysisTests.test_explicit_exemption_is_accepted` |
| B2 | 521 | Function-logic reference exists | Reject coexistence, invalid/recursive ref, invalid ref base, or base mismatch before using reference evidence. | `no focused current test; adoption/reference fixture required` |
| B3 | 535 | No local/reference analysis directory | Permit only no required functions plus explicit not-applicable marker; otherwise fail. | `CheckAnalysisTests.test_explicit_exemption_is_accepted` |
| B4 | 542 | Analysis directory is empty | Fail because evidence targets are missing. | `no focused current test; existing `check` negative-path fixture required` |
| B5 | 556 | Each bundle target | Validate bundle then reject duplicate bindings. | `no focused current test; existing `check` negative-path fixture required` |
| B6 | 563 | Each required changed-existing function | Require binding, parse AST, compare expected source hash and revision. | `CheckAnalysisTests.test_modified_function_cannot_use_exemption` |

## Calls and live bindings

- `resolve_base`, `changed_existing_functions`, `test_index`, `call_enumeration_in_use`, `_bundle_text`, `_ast_value`, `validate_target`.
- No live trading, account, runtime process, network call, or configuration toggle is involved.

## State mutations and fallbacks

No source mutation; produces only a list of validation errors.

## Safety conclusion

- Safe edit boundary: After adoption envelope validation, select E only for fixed a063 and retain every ordinary bundle/hash/revision requirement across the full E..H set.
- High-risk impact: no product/trading path; high assurance applies because this checker is a completion gate.

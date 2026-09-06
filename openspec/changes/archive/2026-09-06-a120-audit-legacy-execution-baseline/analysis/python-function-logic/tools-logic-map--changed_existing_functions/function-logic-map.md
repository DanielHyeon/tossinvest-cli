# Function Logic Map: `changed_existing_functions`

- Source: `tools/logic-map/check_analysis.py` (Python; current AST range is recorded in `ast.json`)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

Pre-edit Python AST evidence is intentionally stored here because this change changes SDD tooling only. The Go-only `check_analysis.py` completion validator does not consume these Python artifacts. The a120 review must retain `Function Logic Map: not-applicable` for a120's Go source set; these maps do not waive the Python review obligation.

## Inputs and invariants

`root` and explicit `base` define the diff and immutable base files. Parsed Go functions are derived through the existing Go AST helper.

## Branches and early returns

| Branch | Source | Condition | Mutation/side effect | Return/error | Baseline test |
|---|---:|---|---|---|---|
| B1 | 90 | No explicit base | Raise `ValueError`; comparison cannot be implicit. | `no focused current test; new immutable-target fixture required` |
| B2 | 109 | `git diff` fails | Raise `RuntimeError`; never interpret a failed diff as empty. | `CheckAnalysisTests.test_git_diff_failure_is_not_treated_as_empty_change` |
| B3 | 118 | No file/hunks pending in nested `flush` | Reset hunk state and return from that flush. | `CheckAnalysisTests.test_new_function_in_existing_file_is_not_reported_as_modified_existing` |
| B4 | 122 | Base file cannot be loaded | Raise `RuntimeError`; deleted/base lookup is required evidence. | `CheckAnalysisTests.test_base_file_load_failure_is_not_treated_as_new_file` |
| B5 | 140 | Current path exists | Compare current AST only against functions that existed at base. | `CheckAnalysisTests.test_new_function_in_existing_file_is_not_reported_as_modified_existing` |
| B6 | 148 | Current function was absent at base | Skip only that new function; it is not a changed-existing function. | `CheckAnalysisTests.test_new_function_in_existing_file_is_not_reported_as_modified_existing` |
| B7 | 152 | Current function intersects a new-side hunk | Record current source hash. | `CheckAnalysisTests.test_modified_function_cannot_use_exemption` |
| B8 | 167-179 | Diff parser sees file header, old/new source, or hunk | Accumulate exact zero-context diff coordinates. | `CheckAnalysisTests.test_new_function_in_existing_file_is_not_reported_as_modified_existing` |

## Calls and live bindings

- `subprocess.run(git diff)`, `base_file`, `go_functions`, `qualified`, `intersects`, nested `flush`.
- No live trading, account, runtime process, network call, or configuration toggle is involved.

## State mutations and fallbacks

Only temporary base-source file creation/removal; no repository content or checkout switch.

## Safety conclusion

- Safe edit boundary: Add immutable-target extraction without switching the caller checkout, and preserve the complete P..E/E..S function derivation semantics.
- High-risk impact: no product/trading path; high assurance applies because this checker is a completion gate.

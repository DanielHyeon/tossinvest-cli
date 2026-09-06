# Function Logic Map: `resolve_base`

- Source: `tools/logic-map/check_analysis.py` (Python; current AST range is recorded in `ast.json`)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

Pre-edit Python AST evidence is intentionally stored here because this change changes SDD tooling only. The Go-only `check_analysis.py` completion validator does not consume these Python artifacts. The a120 review must retain `Function Logic Map: not-applicable` for a120's Go source set; these maps do not waive the Python review obligation.

## Inputs and invariants

`change_dir` supplies the persisted `base-commit.txt`; `root` is the Git repository passed to `git rev-parse`. `SDD_BASE_REF` is only a consistency assertion.

## Branches and early returns

| Branch | Source | Condition | Mutation/side effect | Return/error | Baseline test |
|---|---:|---|---|---|---|
| B1 | 194 | Read failure for `base-commit.txt` | Raise `ValueError`; no default or HEAD fallback. | `CheckAnalysisTests.test_invalid_base_fails_closed` |
| B2 | 212 | `git rev-parse --verify` fails for either persisted candidate or override | Raise `ValueError` identifying invalid base. | `CheckAnalysisTests.test_environment_base_cannot_override_persisted_change_base` |
| B3 | 220 | An environment override resolves differently from persisted base | Raise mismatch error; override cannot select the base. | `CheckAnalysisTests.test_environment_base_cannot_override_persisted_change_base` |

## Calls and live bindings

- `Path.read_text`, nested `resolve`, `subprocess.run`, `os.environ.get`.
- No live trading, account, runtime process, network call, or configuration toggle is involved.

## State mutations and fallbacks

No repository mutation. It selects and validates an immutable commit ID for the ordinary checker.

## Safety conclusion

- Safe edit boundary: Keep ordinary persisted-P behavior unless the frozen a063 record satisfies every adoption predicate; an invalid record must fail closed.
- High-risk impact: no product/trading path; high assurance applies because this checker is a completion gate.

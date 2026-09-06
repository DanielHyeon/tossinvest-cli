# CodeGraph Evidence: `tools/sdd/sdd_doctor.py::report` (pre-edit)

Captured against current HEAD `19021ada90135532adfaa70712fb199e42e1c95b` with
CodeGraph `0.9.8`; `codegraph status .` reported the index up to date.

| Query | Exact doctor evidence |
|---|---|
| `codegraph query "report"` | `function report` at `tools/sdd/sdd_doctor.py:91`, signature `(root: Path = ROOT) -> dict`. The query is basename-ambiguous: it also returned unrelated Go `Report`/`report` symbols. |
| `codegraph callers "report"` | Local `main` at `tools/sdd/sdd_doctor.py:121`; the generic query also returned an unrelated Go caller. Direct source/AST confirms the doctor `main` call is `report(Path(args.root))` at line 126. |
| `codegraph callees "report"` | Doctor-local callees: `command_status` line 36, `skill_status` line 51, `service_status` line 67. Other returned Go symbols are basename noise and are out of scope. |
| `codegraph impact "report"` | Includes the doctor file and its local `main`, plus unrelated basename matches. No product execution path is an evidence-backed dependency of the Python function. |
| `codegraph context report` | Rendered the exact pre-edit body and the three local helper definitions; AST is the authoritative function-internal evidence in [ast.json](ast.json). |

This evidence supports a localized doctor/test/doc change only. It does not
authorize changes to adoption source validation, Go product code, services,
runtime processes, locks, global packages, or trading paths.

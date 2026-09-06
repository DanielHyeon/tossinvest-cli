# Branch Test Map: `tools/sdd/sdd_doctor.py::report` (pre-edit)

This is a pre-edit map, not a coverage claim. The focused baseline executed five
unit tests; it did **not** measure coverage.

| Branch / contract | AST evidence | Existing baseline test | Observed baseline | Required later regression |
|---|---|---|---|---|
| Local venv exists | N005 true → N016 | None directly patches `report`; CLI baseline exercised it with the existing local venv. | `typedb-driver 3.11.5` reported OK. | Absent `SDD_PYTHON` preserves this exact selection/probe behavior. |
| Local venv absent | N005 false → N017 | None. | Not measured in this run. | Absent variable emits the current setup hint. |
| Required tool command fails | N003 via `command_status` | `test_required_cli_missing_has_install_hint`, `test_command_timeout_fails_closed` | Both passed. | Keep command-status failure behavior unchanged. |
| Advisory service lookup fails | N086/N087 via `service_status` | `test_missing_service_is_advisory`, `test_service_timeout_is_advisory` | Both passed. | Keep services advisory. |
| Explicit external interpreter | No pre-edit AST node; this is the new contract. | None. | Not present. | Valid absolute external interpreter outside lexical and resolved root; pinned `typedb-driver==3.11.5` succeeds and is reported. |
| Invalid explicit interpreter | No pre-edit AST node; this is the new contract. | None. | Not present. | Empty, relative, missing, directory, non-executable, lexical-in-root, resolved-in-root, and mismatched/missing driver all fail without N004 fallback. |
| External interpreter symlink | No pre-edit AST node; this is the new contract. | None. | Not present. | External symlink to external executable is accepted if its resolved target is outside root and the pinned probe succeeds. |

Focused baseline command and actual result:

```text
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tools/sdd -p 'test_sdd_doctor.py' -v
Ran 5 tests in 0.002s
OK
```

# External-resolution helper pre-fix evidence

Captured 2026-09-06 immediately after the real symlink-loop RED run and before
the loop correction. This supplements, and does not replace or rewrite, the
existing frozen `python-function-logic/tools-sdd-sdd_doctor--report/` pre-edit
artifact.

Source: `tools/sdd/sdd_doctor.py` SHA-256
`98e9bb7bde9b7e1aa26494f6964a71bebb7df9b83b0ab136889d24ae57437da4`.
The direct `ast.parse` inventory was:

| Helper | Lines | Branch nodes |
| --- | --- | --- |
| `_resolved_detail` | 97–101 | `Try:98` |
| `_external_python` | 128–145 | `If:129`, `If:132`, `If:135`, `Try:137`, `If:141`, `If:143` |
| `_driver_status` | 148–188 | `If:149`, `If:152`, `If:172`, `If:175`, `If:186` |

RED proof: `/tmp/a120-external-loop-red.log` and its numeric exit file
`/tmp/a120-external-loop-red.exit` record exit `1`. A real `external-a ->
external-b -> external-a` loop reached `Path.resolve(strict=True)` in
`_external_python` and raised `RuntimeError`; no typedb-driver command was
issued. The same uncaught exception was reachable in `_resolved_detail` for
the local diagnostic path.

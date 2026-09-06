# External interpreter lexical-boundary amendment — pre-edit evidence

Captured 2026-09-06 before the boundary amendment. This is additive evidence;
the prior external-interpreter and symlink-loop records remain unchanged.

`tools/sdd/sdd_doctor.py` SHA-256:
`9e09b3b781b377221f3cc66a41634c871098d6338eda9cc0d2bf48221c346b04`.

Direct `ast.parse` inventory:

| Helper | Lines | Branches |
| --- | --- | --- |
| `_resolved_detail` | 97–101 | `Try:98` |
| `_typedb_pin` | 104–125 | `Try:106`, `For:112`, `If:114,117,119,121,123` |
| `_external_python` | 128–145 | `If:129,132,135,141,143`, `Try:137` |
| `_driver_status` | 148–188 | `If:149,152,172,175,186` |

The pre-edit lexical guard used `root.absolute()` and the raw `Path(raw)`.
It therefore did not collapse `..` segments and did not compare an aliased
supplied root to its canonical target before resolving the candidate link.
The amendment must reject lexical candidates inside either normalized supplied
root or canonical root before candidate resolution, while retaining the
separate resolved-target check.

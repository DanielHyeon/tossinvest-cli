# Final post-edit Python branch map

Source: `tools/logic-map/check_analysis.py`, SHA-256
`f7d95a10cee85773189b4ee446a282eed7fb78021e361868029bda4ae5f1ebf2`.
`final-function-ast.json` is a direct `ast.parse` inventory of every `If`,
`For`, and `Try` in the changed checker functions.

| Function | AST locations (all `If`/`For`/`Try`) | Evidence / limit |
|---|---|---|
| `changed_existing_functions` | `if` 116, 139, 148, 152, 163, 171, 179, 183, 195; `try` 157; `for` 160, 172, 199; parser `if` 200, 204, 207, 212 | Newline-path and immutable-target fixtures exercise the new preflight/target paths. The remaining unchanged parser/error routes are inventoried but not individually instrumented by a new test. |
| `resolve_base` | `try` 229, 254; `if` 247, 259, 263 | Ordinary P/HEAD/other and real adoption-E fixtures cover the selected-base and override paths. The base-read and rev-parse error routes remain existing fail-closed behavior, not newly branch-instrumented proof. |
| `check` | `try` 561, 573, 613; `if` 566, 567, 570, 577, 580, 581, 587, 604, 605, 610, 618, 621; `for` 601, 608 | Real ordinary dirty-Go and reference-only fixtures cover context retention, reference base equality/mismatch, and required evidence. Other existing bundle-validation paths are inventoried only. |
| `main` | `if` 633, 637; `for` 634 | `test_main_distinguishes_ordinary_adoption_and_invalid_results` covers ordinary/adoption/invalid exit text; `test_main_real_adoption_prints_exception_label_only_after_validation` supplies a real validated-adoption and malformed-record pair. |

These are Python-only review artifacts. They neither claim Go Function Logic Map
coverage nor replace final independent review and command gates.

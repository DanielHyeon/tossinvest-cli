# a120 task 1.1 and 1.2 pre-edit evidence

## 1.1 Contract baseline

- Reserved `STORY-TOS-a120` under `FEAT-TOS-001` in the isolated worktree.
- `python3 tools/sdd/capture_change_base.py --change a120-audit-legacy-execution-baseline` captured `e65e394bf84b3c6e4559a219e816af96d341d75d` once; no recapture was attempted.
- `openspec validate a120-audit-legacy-execution-baseline --strict --no-interactive` returned exit 0.
- a119's spec-only repair is present solely as a validation dependency; it is not an a120 implementation input.

## 1.2 Hard evidence and pre-edit maps

- CodeGraph definition: `resolve_base` at `tools/logic-map/check_analysis.py:192`; its sole direct caller is `check` at line 510.
- CodeGraph definition: `changed_existing_functions` at line 86; its sole direct caller is `check`.
- The direct `check` dependencies observed by CodeGraph are `resolve_base`, `changed_existing_functions`, `test_index`, `call_enumeration_in_use`, `_bundle_text`, `_ast_value`, and `validate_target`.
- Exact Python AST snapshots and branch maps are in `analysis/python-function-logic/` for `resolve_base`, `changed_existing_functions`, and `check`. `changed_existing_functions` is included because immutable-target inventories may require a no-checkout derivation.
- `make sdd-sync` reached a CodeGraph `Already up to date` report, then began CodeGraphContext indexing; this Stage 1B execution did not produce a final success exit, so it is not recorded as a passed gate.
- a120 changes no Go product files. Its future `review.md` must carry the exact Go gate marker `Function Logic Map: not-applicable`; Python pre-edit maps are stored outside `analysis/function-logic` so the Go-only completion checker never parses a fabricated Python AST bundle.

Tasks 1.1 and 1.2 remain unchecked for Manager confirmation. Task 1.3 remains pending independent adversarial and gstack review/freeze.

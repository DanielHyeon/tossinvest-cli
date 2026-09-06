# Post-edit Python evidence

Date: 2026-09-06

The changed Python functions were re-read after implementation. Current source
SHA-256 values are `check_analysis.py` `f7d95a10cee85773189b4ee446a282eed7fb78021e361868029bda4ae5f1ebf2`
and `execution_baseline.py` `e9d8a55c718fe3568d09facb355e507c06c565f5bed915ef1b0f780d0104e1c3`.

The post-edit function paths are: `changed_existing_functions` performs a
NUL-delimited safety preflight before unified-diff parsing; `resolve_base`
delegates exception validation and only accepts the selected effective base;
`check` derives required maps through that path. `validate` verifies the strict
record/envelope, immutable inventories, evidence and S/H/worktree source lock.

The hashes and AST inventory were refreshed against the final implementation
snapshot on 2026-09-06. `analysis/post-edit-python/final-function-ast.json`
and `final-branch-test-map.md` bind the current `changed_existing_functions`,
`resolve_base`, `check`, and `main` branches to executable proof.

This is Python tooling evidence, stored outside Go Function Logic Map inputs.
Focused executable proof is recorded in `implementation-evidence.md`; it does
not replace independent review or the final SDD/gate sequence.

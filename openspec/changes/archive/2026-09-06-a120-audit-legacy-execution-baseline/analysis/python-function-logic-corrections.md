# Historical Python branch-map correction

Date: 2026-09-06. Author: independent post-adversarial gstack reviewer.

The original `analysis/python-function-logic/` artifacts are retained unchanged.
This correction was written during implementation review, after implementation;
it does not establish original pre-edit compliance or new executed coverage.
The original three AST files bind `tools/logic-map/check_analysis.py` SHA-256
`88aff065cdad4aa2c155451d23aef783b71519c4c333a2c8ad9fc47e8d091fe5`.
Their branch IDs follow the recorded AST enumeration, not source-line order.

## Discrepancy and effect

- `resolve_base`: historical prose and test-map B2 describe Git failure, while
  AST B2 is the environment mismatch at line 220. Prose B3 describes the
  environment mismatch, while AST B3 is Git failure at line 212.
- `changed_existing_functions`: historical prose has eight grouped rows but
  its AST has sixteen branches. For example, prose B3 is the flush early return
  at line 118, which is AST B4; AST B3 is the diff-line loop at line 166.
  The historical table cannot be treated as exact AST branch coverage.
- `check`: the six historical prose rows group behavior from seventeen AST
  entries. The grouping is useful explanation but does not establish per-ID
  coverage for B7 through B17.

The tables below correct IDs and meanings against the retained AST and original
source. No historical test citation is promoted to executed per-branch proof.
Final source hashes, AST enumeration and explicit coverage limitations must be
recorded separately in post-edit evidence before evidence completion is claimed.

## resolve_base: exact historical AST mapping

| ID | Kind | Original line | Meaning |
| --- | --- | ---: | --- |
| B1 | try | 194 | Read persisted base text; read errors are handled. |
| B2 | if | 220 | Reject an environment override resolving differently from persisted base. |
| B3 | if | 212 | Reject a failing Git commit-resolution command. |

## changed_existing_functions: exact historical AST mapping

| ID | Kind | Original line | Meaning |
| --- | --- | ---: | --- |
| B1 | if | 90 | Reject a missing explicit base. |
| B2 | if | 109 | Reject a failed Git diff. |
| B3 | for | 166 | Iterate unified-diff lines. |
| B4 | if | 118 | Flush returns when source or hunks are absent. |
| B5 | if | 122 | Reject an unavailable immutable base file. |
| B6 | try | 126 | Analyze base/current functions with temporary-file cleanup. |
| B7 | if | 167 | Start a new diff file by flushing and clearing source names. |
| B8 | for | 129 | Iterate base functions. |
| B9 | if | 140 | Analyze current functions only when a current file exists. |
| B10 | if | 171 | Parse the old-source header. |
| B11 | if | 132 | Record a base function intersecting old-side hunks. |
| B12 | for | 141 | Iterate current functions. |
| B13 | if | 174 | Parse the current-source header. |
| B14 | if | 148 | Skip a qualified function absent from the base file. |
| B15 | if | 152 | Record an existing function intersecting current-side hunks. |
| B16 | if | 179 | Append parsed hunk coordinates. |

## check: exact historical AST mapping

| ID | Kind | Original line | Meaning |
| --- | --- | ---: | --- |
| B1 | try | 516 | Resolve the base and derive changed-function obligations. |
| B2 | if | 521 | Enter reference resolution when a reference file exists. |
| B3 | if | 535 | Handle a missing analysis directory. |
| B4 | if | 542 | Reject an empty analysis target set. |
| B5 | for | 556 | Validate each evidence target. |
| B6 | for | 563 | Check every required function against collected evidence. |
| B7 | if | 522 | Reject reference/local-evidence coexistence. |
| B8 | if | 525 | Reject invalid or self-referencing change names. |
| B9 | try | 528 | Resolve the referenced change's base with error handling. |
| B10 | if | 532 | Reject differing reference and local comparison bases. |
| B11 | if | 536 | Reject absent analysis when changed functions require maps. |
| B12 | if | 559 | Record a valid bundle binding. |
| B13 | if | 565 | Report a missing required binding. |
| B14 | try | 568 | Read the bound AST with error handling. |
| B15 | if | 573 | Report a source-hash mismatch against the required revision. |
| B16 | if | 576 | Report current/base revision mismatch. |
| B17 | if | 560 | Report duplicate evidence for one binding. |

## Later main artifact correction

The `main` pre-edit artifact belongs to the separate CLI-label correction and
binds the preceding source hash
`2f769d5fe8e8cfa69c68acefa22ebac7c4c4dee90371a904360bea7980924417`.
Its first reviewed version listed only the `if errors` entry at line 625.
During post-edit review, the omitted `for error in errors` entry at line 626
was identified and Terra added it as B2 to the AST and prose maps. That addition
occurred after the CLI implementation; it is a disclosed correction, not a
claim that the original pre-edit inventory was complete. The initial three
historical function artifacts above remain unchanged.

The final checker inventory is separately bound to the current source hash in
`post-edit-python/final-function-ast.json`. Its accompanying map enumerates all
recorded If/For/Try locations and distinguishes changed-path fixture evidence
from unchanged routes that were not individually instrumented. No suite count
is used as proof of exhaustive branch execution.

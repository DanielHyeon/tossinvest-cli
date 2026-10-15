# Evidence reconciliation

## Baseline status

`a141` is the single writer and has not landed. Per Manager instruction, this a142 preparation does not capture `base-commit.txt`, refresh indexes, edit production/test files, commit, or run tests. a142 remains implementation-blocked until the Manager's GO after a141 landing.

## Evidence agreement

- Intake AST hashes for `ListPendingOrders`, `ListCompletedOrdersRange`, and `ListWatchlist` match current source exactly.
- CodeGraph 1.6.0 definitions, direct source, and intake agree on current branch boundaries and direct dependencies.
- a142 spec decisions resolve earlier watchlist uncertainty: bulk GET only; required `watchlists`/`items` arrays; `Symbol=item.code`; stable numeric ordering; optional currency empty and nullable prices zero.

## Evidence limitations and resolution

- CodeGraph index was not refreshed to avoid a141's reserved single-writer operation. Current target-file hashes and matching definition lines establish the queried source is unchanged from intake.
- CodeGraphContext is not indexed for this `/tmp` worktree and has unrelated name-match context. It contributes only broad consumer awareness and does not override CodeGraph/direct source.
- Planned collection guards and shared mapper branches do not exist in pre-edit AST. They are explicitly marked `planned` in branch maps and require fresh AST/maps after implementation.

## Pre-edit decision

Implementation may be prepared only within the owned read decoder/test surface after Manager GO. No policy change is authorized for order absence, retries, reconciliation, mutations, domain identity, tokens, settings, or live activity.

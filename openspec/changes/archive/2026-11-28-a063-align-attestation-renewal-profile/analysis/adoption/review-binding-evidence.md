# a063 review-binding evidence

## Scope and source lock

- Isolated worktree: `/tmp/tossos-a063-execution-adoption`.
- Fixed source snapshot `S`: `c727ad12a42dcd15c494c1997e92816edaf17b6b`.
- `S` tree: `2d608c945d4b2a4a2fd87824bf06a42815d256a1`.
- The complete worktree diff from `S` is limited to `openspec/` (no `docs/pm/` path is changed); there are no tracked or untracked paths outside `openspec/` or `docs/pm/`.
- Every tracked `*.go` path has the same blob object and mode as `S`; no Go source or test path is changed.
- `git diff --check S --` passed.

## Bound review fields

`execution-baseline.json` is the only file edited by this binding step before this evidence report. Its review fields bind the actual review artifacts with independently recomputed SHA-256 values:

| Field | Value |
| --- | --- |
| `adversarial_review_path` | `openspec/changes/a063-align-attestation-renewal-profile/analysis/adoption/adversarial-review.md` |
| `adversarial_review_sha256` | `5de872ba232e5bbd806d6e9d9024416b3cca30ea7c68bddabeaaa977e3abd763` |
| `gstack_review_path` | `openspec/changes/a063-align-attestation-renewal-profile/analysis/adoption/gstack-review.md` |
| `gstack_review_sha256` | `0ea873742f47e95c9251656d9b9267a780645e37e51800c6d06d7554a152f209` |

## Detached H and staged metadata boundary

- Intended detached commit message: `docs(openspec): bind a063 execution-baseline adoption reviews`.
- The commit identity (`H` SHA and tree) cannot be embedded in its own committed tree: including it would change that tree and therefore H. The actual H SHA/tree and the exact committed-path list are consequently recorded in the post-commit handoff, derived from `git show --name-only H`.
- The staged set is restricted to the complete current `openspec/changes/a063-align-attestation-renewal-profile` metadata set. No report, Function Logic Map, source, test, ledger content, task/design, runtime/service/timer, or `docs/pm/` file is edited by this binding step.
- Therefore `S -> H` is metadata-only: all committed paths are under `openspec/`, and the source lock above remains true.

No final analysis checker, SDD synchronization/check, make gate, tests, runtime action, or archive action is performed by this binding step.

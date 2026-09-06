# a120 review and Manager disposition

## Scope and authority — 2026-09-06

The user authorized a separate common SDD repair and a119 validation repair
while a063 remains held. This change introduces an explicit, exact-tuple a063
execution-baseline adoption exception. It does not overwrite a063's planning
base or establish original pre-edit compliance. Operational approval and final
a063 acceptance remain separate.

Current implementation owner: Terra `a120_verify`, succeeding Terra
`a120_impl` and `a120_finish` in the same preserved isolated detached worktree
`/tmp/tossos-a120-audit-legacy-execution-baseline`.
Dedicated independent adversary: Terra `sdd_adversary`.
Post-adversarial gstack reviewer: a separate reviewer context.
Manager authors OpenSpec, decomposes work and independently verifies acceptance.

## Evidence boundary

The baseline is `e65e394bf84b3c6e4559a219e816af96d341d75d`, captured once.
`analysis/pre-edit-evidence.md` records CodeGraph definitions/callers and
Python AST/maps for the three existing checker functions before edits.
Gstack later identified mismatches between historical AST branch IDs and the
prose maps. Originals remain intact; `analysis/python-function-logic-corrections.md`
records the exact corrections and the limits of that historical evidence.
Python maps are in `analysis/python-function-logic/`; they are not Go AST files.

Function Logic Map: not-applicable

This Go-only completion marker applies because the isolated a120 change modifies
no Go source. It does not waive Python function review, the future adopted a063
Go maps, or any test/gate. The root's uncommitted a063 implementation is not copied
into this tooling worktree. a119 document repair is a declared validation dependency.

## Proposal review

The independent adversary reviewed the actual proposal/design/spec/tasks and
required four corrections: strict source-before-evidence ancestry; a source guard
that covers hidden non-Go build inputs without rejecting legitimate SDD caches;
deterministic complete commit/merge-parent history; and binding the actual
unchanged planning-base file. All four are incorporated and independently
re-reviewed clear in `analysis/adversarial-proposal-review.md`.

Gstack autoplan reviewed CEO, Design applicability, Engineering and developer
experience in a separate reviewer context. Its actual bounded result is
CLEAR FOR PROPOSAL FREEZE in `analysis/gstack/proposal-review.md`. No external
voices, runtime proof or implementation success are inferred from that review.

Manager freezes the current proposal/design/spec/tasks and releases Terra
`a120_impl` to implement tasks 2 and 3 within the isolated worktree. Preserve the
fixed P/E tuple, ordinary behavior, pre-edit snapshots and all acceptance checks.
Any blocking contract change returns to Manager and the reviewers before coding.

## Completion boundary

Implementation success requires targeted and whole-repository tests, separate
adversarial review followed by gstack, current SDD evidence, the actual a120 gate,
Manager acceptance and official spec/PM archive. Valid proposal artifacts or
checked prerequisite tasks alone do not establish completion.

## Implementation review continuation

Independent adversarial review cleared the helper/checker implementation and
focused fixture matrix. The subsequent gstack review found a default-branch
assumption in fixtures and missing explicit CLI exception labeling. Terra owns
the fixes and executable verification. These findings and the historical
branch-map correction remain part of the review record, not erased history.

Gstack directly executed two reproductions before the Manager restated the
Terra-only test-execution boundary. That role deviation is disclosed; subsequent
fix verification is delegated to Terra. No external-model review is inferred.
Final gstack disposition is CLEAR for reviewed implementation with disclosed
historical/process concerns, zero unresolved findings; see
`analysis/gstack/implementation-review.md`. Manager directly compared all four
recorded source/test SHA-256 values to the current files and found exact matches.
Final SDD tests, strict OpenSpec (61/61) and PM checks have actual exit 0.
Untagged and seam-tagged whole-repository Go tests, repository race targets,
ordinary vet and seam-tagged vet all have actual exit 0. Manager read the
durable exit records directly. SDD freshness/check, gate and archive remain
pending.

## External interpreter amendment freeze

Normal-mode SDD check passed after local venv setup, but Manager identified an
integration conflict: the doctor requires that local venv while actual a063
adoption rejects its untracked source. This is a real completion blocker, not
a reason to relax the source guard. Task 2.5 and final verification/review tasks
were reopened; prior test results remain scoped to their recorded version.

Dedicated Terra `a120_env_adversary` independently confirmed the conflict and
cleared design decision 5 plus the exact doctor pre-edit AST/maps. Separate
gstack planning review also cleared the amendment in
`analysis/gstack/external-interpreter-plan-review.md`. Manager freezes that
contract, including absent versus empty environment values, independent lexical
and resolved path boundaries, exact driver-pin validation and no local fallback.

Because the host rejected further native implementation threads, implementation
is delegated to local Codex CLI with explicit `--model gpt-5.6-terra`, preserved
event/exit logs and the same isolated worktree. Its preparation completed with
exit 0 before doctor edits. Hooks, global configuration and memory writes are
excluded. The separately assigned adversary and gstack reviewer remain the
review owners. Manager releases only doctor/test/docs amendment implementation;
adoption guards, Go product sources, runtime, services and ownership locks remain
outside this edit scope. Actual external-mode compatibility must be verified
before final gate readiness.

## External interpreter amendment review closure

Terra repaired the independently reproduced loop, alias/dot-boundary and
malformed UTF-8 diagnostic failures. The dedicated adversary cleared the final
implementation; the subsequent gstack pass requested two missing regression
cases. A bounded Terra context added successful-process/wrong-version coverage
and the same real adoption fixture's forbidden-local-input rejection. The
adversary independently verified both, then gstack closed both findings and
returned CLEAR. Reports are `analysis/external-interpreter-adversarial-review.md`
and `analysis/gstack/external-interpreter-implementation-review.md`.

Manager read the final code and actual exit records and matched the reviewed
source/test digests. The final external-mode SDD suite has 242 Python tests plus
Go logic-map tests, exit 0. Earlier Go/seams/race/vet results remain applicable
because this amendment changes no Go code. Final strict validation, freshness,
SDD check and actual gate remain separate pending steps. The unplanned nested
review process described in implementation evidence is excluded from acceptance.

## Manager final acceptance — 2026-09-06

Manager independently read the actual completion gate's eleven step results
and preserved numeric exit 0. The command was
`SDD_PYTHON=/tmp/tossos-a120-external-sdd-venv/bin/python make gate CHANGE=a120-audit-legacy-execution-baseline`.
Tasks, independent review record, Function Logic Map policy, SDD check, full Go
tests, seams tests, race targets, vet and strict OpenSpec validation all passed.
The final log ends `GATE PASS: a120-audit-legacy-execution-baseline`; strict
validation reports 61 passed and zero failed.

Evidence: `/tmp/a120-completion-gate.log`, actual command exit preserved as
`/tmp/a120-completion-gate-command.exit`, log SHA-256
`8000ee11cb33e25b694ce32d719779b3e8cbb78862ca3587bb8b7e1003d2cb3d`.
The six source/test hashes remain bound in the Manager verification checklist.
All planning artifacts are done, all tasks are checked, and no implementation
review finding remains open. The historical/process disclosures above remain.

Manager accepts a120 implementation and authorizes the already-requested
official spec-syncing archive, dated PM synchronization and strict post-archive
validation. Shared-checkout integration is separately verified afterward. This
acceptance does not adopt a063 automatically or complete its operational proof.

## Official archive verified — 2026-09-06

The official `openspec archive a120-audit-legacy-execution-baseline --yes`
exited 0, synchronized one modified and four added `sdd-workflow` requirements,
and created this dated archive. Manager inspected the actual canonical diff,
dated Story path and recorded exits: canonical strict validation, all strict
validation, PM check, diff check and source-drift check each exited 0.
Evidence logs are `/tmp/a120-official-archive.*` and `/tmp/a120-archive-*`.
A119 remains active as the authorized documentation dependency.

Manager authorizes integration of the complete reviewed change from original E,
including this archive and canonical/PM synchronization. The incomplete WIP
commit alone must not be integrated. All unrelated shared-checkout changes and
its HEAD/index must remain unchanged, with root PM views regenerated for their
actual local state. Integration verification is still required.

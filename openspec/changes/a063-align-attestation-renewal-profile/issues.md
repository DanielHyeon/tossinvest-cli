# a063 continuation issues — 2026-09-05

## I1 — immutable historical base blocks final Function Logic Map gate

- Status: unresolved; final acceptance and archive blocked.
- Persisted base: `da80ce31b6a1ab5d443016768f970a82bab102db`, captured by the historical planning
  commit `47a7f90a8738baaafb79a8a33fe5c618d432a9ce`.
- Current source has hundreds of subsequent commits. The implementation teammate reports 846 Go paths
  in the base-to-current diff and missing unrelated function maps from `check_analysis.py`.
- `tools/sdd/capture_change_base.py` treats the persisted base as immutable and refuses recapture.
  `check_analysis.py` compares the whole worktree to that base; a private comparison against today's
  HEAD cannot replace this required gate.
- Independent reviewer measurement during implementation: 332 modified-existing functions across
  140 files (280 current-revision bindings and 52 base-revision bindings; 217 production and 115 test
  functions). This is a time-specific measurement, not an immutable total while edits continue.
- The supported `function-logic-reference.txt` mechanism requires one active change with the exact
  same base and forbids combining that reference with local maps. It is not per-function or archived
  evidence reuse. No valid same-base complete reference was established; other active change contents
  were not opened to evade the one-change-at-a-time boundary.
- Manager and independent adversarial reviewer reject overwriting the base, reducing checker scope,
  fabricated historical maps, or passing a different `SDD_BASE_REF` as a workaround.
- Continue gathering genuine current-function evidence and isolated engineering results within a063.
  These results do not establish final-gate success. A successor change or an explicitly approved,
  separately specified baseline-migration mechanism would require a concrete scope decision before use.
- 2026-09-06 continuation: direct validation passed eleven scoped bundles (five production and six
  checker-derived test functions). Three base-revision test ASTs were extracted retrospectively,
  not captured before implementation. The global checker still exits 1 with 327 missing-evidence
  rows outside these scoped bundles; see `analysis/coverage-status.md` for the current measurement.

## I2 — operational acceptance remains separate from engineering proof

- Status: bounded installation proposal independently reviewed; not executed. Explicit approval,
  same-profile evidence and operational verification pending.
- The existing survey/timer predate this continuation. Read-only observations are in
  `analysis/sanitized-operational-evidence-2026-09-05.md`; no service or survey was activated by this work.
- Current qualifying-day and attestation metadata can inform verification, but cannot establish that
  the new repository service was installed, or that a reviewed installation was approved.
- Preserve original records and all operational settings. Do not backdate the expired 2026-08-29
  deadline, simulate live evidence, or check deployment tasks from isolated tests.
- Prepare exact unit template digest, target profile, backup/reload/verification commands and rollback
  for review before requesting operational approval.
- The continuation adversary found two high and one medium defects in the proposed activation
  procedure, then verified all three corrections: validated backups before timer stop, immediate
  activation-precondition rechecks, and fail-closed service-state inspection. See
  `analysis/adversarial-continuation-review.md` and `analysis/deployment-plan.md`.
  This bounded procedural review does not establish operational approval or acceptance.

## I3 — unrelated repository-wide verification failures

- Status: observed during engineering verification; not changed within a063.
- 2026-09-06 expanded-scope repair: the user authorized resolving common SDD and a119 validation
  blockers. Manager completed a119's missing two declared specification deltas with future
  implementation tasks left pending. Separate adversarial validation passed a119 and all 60 current
  OpenSpec items (exit 0); post-adversarial gstack found no additional contract defect. This closes
  the observed OpenSpec parsing blocker only, not a119 implementation or the platform issue below.
- `make validate` returned exit 2 because the unrelated `a119-codex-session-handoff-and-gbrain-startup`
  change failed validation; a063 itself validated. No a119 content was folded into this implementation.
- Whole-CLI Windows/macOS cross-build attempts reported the existing
  `internal/strategyaccount/production_owner_unix.go:17` undefined `productionFileUID` reference.
  Targeted Windows soak-package compilation passed after a063 platform separation. A successful
  package compile is not a successful release binary build.
- Record final rerun outcomes in `analysis/verification.md`; do not present these commands as passed
  or modify unrelated source merely to obtain a green a063 completion report.

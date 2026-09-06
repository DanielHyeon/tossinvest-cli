## Why

Final integration review also identified a tooling conflict: the doctor requires
a local Python venv that the adopted-worktree source guard rejects. The bounded
amendment in design decision 5 adds explicit external `SDD_PYTHON` selection to
the doctor, with fail-closed path/dependency checks and unchanged default local
behavior. It preserves the adoption source guard and requires its own pre-edit
evidence, independent reviews and actual external-mode verification.

a063 was planned at an immutable August baseline and resumed after substantial
committed development. Its completion checker now requires 338 existing-function
bundles, including 327 missing rows outside the resumed implementation. Resetting
the original baseline would discard provenance; exact archived evidence reuse
also leaves substantial gaps. The user authorized a separate common SDD repair.

## What Changes

- Introduce one explicit, audited execution-baseline adoption exception for
  `a063-align-attestation-renewal-profile` and one fixed planning/execution commit
  pair. Preserve the original `base-commit.txt` without rewriting it.
- Require complete deterministic accounting of intervening committed history,
  independent adoption review, a committed source snapshot and an isolated clean
  verification worktree before the exception can select its execution baseline.
- Continue requiring real Function Logic Maps for every modified existing Go
  function after that fixed execution baseline. No path filter or environment
  variable may omit an implementation change.
- Keep ordinary changes on the existing immutable-base policy. Document this
  exceptional result explicitly; it does not establish original pre-edit
  compliance or complete the inherited history's missing analysis.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `sdd-workflow`: Define a narrowly allowlisted legacy execution-baseline
  adoption, its historical accounting, committed-source checks and honest
  completion terminology.

## Impact

- `tools/logic-map/check_analysis.py`, a dedicated adoption validation/generation
  helper, and isolated Git-fixture regression tests.
- `docs/WORKFLOW.md`, paired PM Story `STORY-TOS-a120`, and a120 review evidence.
- a119 specification repair is an explicit validation dependency copied into the
  isolated worktree; it does not authorize or complete the a119 feature.
- No Go product implementation, live process, account, operational control or
  original a063 baseline is changed by this tooling change. Actual a063 adoption
  follows tool acceptance and remains distinct from operational acceptance.

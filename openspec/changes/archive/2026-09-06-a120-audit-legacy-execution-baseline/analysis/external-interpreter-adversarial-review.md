# External interpreter amendment: independent adversarial review

Date: 2026-09-06. Reviewer: dedicated Terra `a120_env_adversary`, separate
from the Manager and the planned CLI Terra implementation context.

## Scope and verdict

**Freeze verdict: CLEAR. Implementation verdict: pending.**

This review covers the a120 integration amendment only: design decision 5, the
new `sdd-workflow` requirement, `SDD_PYTHON` contract, and pre-edit evidence
for `tools/sdd/sdd_doctor.py::report`. It does not review the pre-existing
a120 execution-baseline helper, alter adoption source validation, inspect a063
implementation, or authorize a gate, archive, runtime operation, service,
lock, global package, or configuration change.

## Independent reproduction of the blocker

The current doctor constructs only
`<root>/.sdd/.venv/bin/python` in `report()` and treats its TypeDB-driver probe
as required. `make sdd-check` invokes `sdd-doctor` first.

In this isolated worktree, the local `.sdd/.venv` was ignored by
`.gitignore` and present for the normal doctor probe. Direct source-guard
enumeration found 152 untracked entries below that environment; all 152 were
outside `execution_baseline._allowed_metadata`, including:

```text
.sdd/.venv/bin/python
.sdd/.venv/lib/python3.13/site-packages/typedb/driver.py
```

The normal doctor did report `typedb-driver 3.11.5` as OK, but that only shows
the local environment lets an ordinary worktree pass. A valid a063 adoption
must reject those untracked paths. Thus the two original requirements cannot
both be satisfied by a repository-local environment. The source guard is
correctly restrictive and must remain unchanged.

Focused existing regressions passed before the amendment:

```text
tools/logic-map: filesystem metadata/source mutation, ignored build inputs,
and dual Go-enumeration guard tests: 3 passed
tools/sdd: existing doctor tests: 5 passed
```

## Contract review

The chosen repair is appropriately bounded: doctor-only optional
`SDD_PYTHON`, with normal local behavior retained only when the variable is
absent. It cannot select E, alter the adoption helper, add a source exception,
or create an environment.

The following implementation constraints are frozen and must be tested:

1. Presence distinguishes absence from an explicit empty value. Empty is a
   required failure and cannot fall back to the local venv.
2. An explicit value is an absolute executable regular file outside the
   repository both before resolution and after resolution. Check containment
   by path components, not string prefix. Lexical-internal to external,
   external to internal, broken/looped links, and alias/dot-spelled internal
   paths fail. A sibling-prefix path remains valid when both its normalized
   lexical path and resolved target are outside the repository; external to
   external links may pass.
3. Explicit mode obtains the sole expected `typedb-driver==...` pin from
   `tools/sdd/requirements.txt`. Missing, unreadable, malformed, or ambiguous
   pins, a failed probe, malformed result, timeout, missing dependency, and a
   version mismatch all fail closed. A valid local interpreter must not be
   consulted after any explicit failure.
4. The report must identify selected mode/interpreter and dependency result,
   including an invalid explicit selection. Command execution must remain an
   argument vector so a path containing spaces is not shell-evaluated.
5. A real Git adoption fixture must prove that an external pinned interpreter
   probe can succeed with no repository `.sdd/.venv` while unchanged adoption
   validation succeeds; the same fixture with a local forbidden input must
   fail source validation.

## Pre-edit evidence binding

`analysis/python-function-logic/tools-sdd-sdd_doctor--report/ast.json` binds
the edited existing function to current HEAD blob
`c777d10b5216c4c830817992532de9fa58d52842` and source SHA-256
`eb6c4ecb4d1f80c17a1514bd93ba7683b37f40d5c8e508aec5c89e95a6a7c26d`.
Its N004 local interpreter assignment and N005 present/absent branch match the
current source. The companion map correctly labels local-present as observed,
local-absent as unmeasured, and the external cases as new required regressions.
CodeGraph evidence scopes the only direct caller to doctor `main` and identifies
the three local helper callees. This is sufficient pre-edit evidence for the
limited Python edit.

The gstack amendment plan independently reaches the same absence/empty,
two-boundary, exact-pin/no-fallback, and real-fixture requirements. Its
duplicate `Real use` journey row is editorial only and does not affect the
freeze verdict.

## Required post-implementation review

Re-review the actual diff and execute the focused tests after the separate
implementation context finishes. Block acceptance if any constraint above is
missing, if this amendment changes `execution_baseline.py`, if an explicit
failure can use the local interpreter, or if the claimed integration proof only
mocks the source guard or the external dependency probe. Final SDD/check/gate
results remain separate evidence and are not claimed here.

## Final implementation review

**Implementation verdict: CLEAR for the external-interpreter amendment.**

Reviewed immutable source identities:

| File | Git blob | SHA-256 |
| --- | --- | --- |
| `tools/sdd/sdd_doctor.py` | `987f9c8e70b3241c9ce59871c5b1c5f3d0ba843c` | `e038dce83e10cec6df49977a20c6bcb8dc5d873edae6bf81a95e08f2f7f04268` |
| `tools/sdd/test_sdd_doctor.py` | `a5875c23972bf2d1166d1c8c5c6ef4a029387372` | `43fca70980b7bd24c8cc9931eff04bea2c2d78788dd301f84dcd22b274c2369a` |
| `tools/logic-map/test_check_analysis.py` | `d487eda07bf8da43eca2d9fe1d9b2d1c46f8d432` | `85e0de6a38ebbc5e68ea00781c314bc01db416fe44742a30d5df74c312b73430` |

The final implementation distinguishes environment membership from truthiness;
parses one exact TypeDB pin only in explicit mode; emits mode/raw/resolved
diagnostics; retains the absent-variable local probe; and does not add a local
fallback. `_normalized_path` removes lexical dot segments without resolving
links, then checks both the supplied-root spelling and canonical root before
the final resolved-target boundary check. Resolution now handles `OSError`,
`RuntimeError` (symlink loops), and `ValueError`; malformed UTF-8 requirements
are a structured dependency failure. The NUL regression is correctly marked
as direct-helper defense because an OS environment cannot carry NUL.

Independent commands and actual results:

```text
cd tools/sdd && python3 -m unittest -v test_sdd_doctor
15 tests, exit 0

SDD_PYTHON=/tmp/tossos-a120-external-sdd-venv/bin/python \
  python3 tools/sdd/sdd_doctor.py --json
exit 0; mode=external and typedb-driver 3.11.5

cd tools/logic-map && SDD_PYTHON=/tmp/tossos-a120-external-sdd-venv/bin/python \
  python3 -m unittest -v \
  test_check_analysis.CheckAnalysisTests.test_real_adoption_without_local_venv_accepts_external_doctor_probe
1 real Git adoption/doctor integration test, exit 0

cd tools/logic-map && SDD_PYTHON=/tmp/tossos-a120-external-sdd-venv/bin/python \
  python3 -m unittest -v \
  test_execution_baseline.ExecutionBaselineUnitTests.test_filesystem_scan_metadata_and_source_mutations \
  test_execution_baseline.ExecutionBaselineUnitTests.test_ignored_build_inputs_are_not_hidden_by_git_ignore_rules \
  test_execution_baseline.ExecutionBaselineUnitTests.test_validate_unions_default_and_seams_inputs_and_blocks_each_profile_failure
3 unchanged source-guard tests, exit 0

git diff --check
exit 0
```

I also directly reproduced and verified the repaired boundaries: looped
external links now return an invalid selection; alias-root, root-dot and
candidate-dot attempts to name a local venv are rejected before probing; an
external sibling-prefix path remains accepted; malformed UTF-8 requirements
return a dependency error. These checks found no remaining boundary bypass.

Post-review coverage closure changed tests only; the doctor source hash above
is unchanged. I independently ran the two added cases: an external probe that
returns process success with `typedb-driver 0.0.0` is rejected and records only
the external command, and the same real external-doctor/adoption fixture first
accepts its clean source then rejects an added
`.sdd/.venv/forbidden.py`. Each command exited 0. This delta preserves the
implementation verdict **CLEAR**.

This clears only the independent amendment review. The required post-review
gstack pass, final `make sdd-sync`, external-mode `make sdd-check`, actual gate,
Manager acceptance, and archive remain separate completion conditions.

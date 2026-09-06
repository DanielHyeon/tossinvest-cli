# External interpreter amendment: post-adversarial gstack delta review

Date: 2026-09-06. Reviewer: assigned native gstack reviewer, separate from
the Terra implementation context and dedicated Terra amendment adversary.

## Scope and method

Reviewed frozen design decision 5, the external-interpreter delta requirement,
workflow setup documentation, complete final doctor and doctor tests, the
checker fixture integration, pre/post-edit Python evidence and the dedicated
adversary's final implementation report. The original execution-baseline helper
review remains historical; this is the subsequent amendment implementation pass.

Applied gstack `/review` scope/completion, critical and informational passes and
testing, maintainability, security and performance checklists in this single
context. The installed skill/checklists were read; no nested specialist or
external-model verdict is claimed. User scope replaces remote branch discovery
with the isolated worktree's actual amendment diff including uncommitted files.
No fetch, PR/Greptile interaction, global state, telemetry, memory update,
autofix, test execution, installation or runtime action was performed here.
The implementation context's stopped unintended nested review is excluded.

## Scope check and substantive review

Scope check: CLEAN. The doctor alone consumes `SDD_PYTHON`. Its existing main
still treats the dependency status as required; the five report groups and
advisory service semantics remain intact. The workflow documents external setup
and explicitly excludes `refresh_indexes.py` interpreter routing. No adoption
allowlist, baseline selector or product behavior was changed by the amendment.

Environment membership separates absent and empty selection. Absent mode keeps
the same local command and setup hint, adding mode/path diagnostics. Explicit
mode validates normalized lexical containment against supplied and canonical
roots, then separately checks the resolved target. Component containment avoids
the sibling-prefix false rejection. External-to-external symlinks work while
internal lexical paths to external targets and external paths to internal
targets fail. Symlink loops and malformed UTF-8 requirements are structured
failures. NUL tests accurately identify direct-helper defense, not real OS
environment input.

The explicit probe requires one exact requirements pin, successful command
status and matching returned version. Invalid selection/pin branches return
before probing; failed probes never fall back locally. Execution uses an
argument tuple and the existing eight-second timeout. There is no shell
interpolation, source write, environment creation, database operation,
authentication change, or asynchronous context in this delta. Helpers have
direct callers; pin authority remains the repository requirements. One bounded
dependency probe per report adds no cache or repeated adoption validation.

This is a local tooling readiness check, not attestation against a malicious
operator who can replace their chosen executable concurrently. No stronger
environment immutability or full operating-system portability claim is made.

## Regression findings

Two narrow test gaps were sent to Manager for Terra implementation, without
claiming a demonstrated production-code defect:

1. [P2] (confidence: 10/10) `tools/sdd/test_sdd_doctor.py`,
   `test_explicit_mode_rejects_missing_and_mismatched_driver`: the stub returns
   `{"ok": False, "detail": detail}` for both cases. Thus wrong-version input
   short-circuits the `not status["ok"]` branch and does not test a successful
   probe whose version mismatches. Require an `ok=True` wrong-version case.
2. [P2] (confidence: 10/10) `tools/logic-map/test_check_analysis.py`,
   `test_real_adoption_without_local_venv_accepts_external_doctor_probe`: the
   fixture stops after successful `adoption.validate(...)`. The frozen
   adversarial contract also calls for the same fixture with a forbidden local
   input to fail source validation. Require that negative continuation while
   retaining the valid external interpreter.

Status: BOTH RESOLVED. Terra changed tests only. The mismatch case now supplies
`ok=True` with `typedb-driver 0.0.0`, checks expected/observed diagnostics and
asserts exactly the external invocation even with a local executable present.
The real adoption fixture now creates `.sdd/.venv/forbidden.py` after its clean
success and requires the exact untracked/ignored-input `AdoptionError`. Read
both final bodies and the dedicated adversary's subsequent report: it executed
each strengthened case independently with exit 0 and retained CLEAR. No RED
production defect is claimed; these close missing regression assertions.

## Evidence actually inspected

The test files contain isolated environment patches (`clear=True`) for local
and explicit doctor cases. The real Git fixture patches only its fixed P/E
identities, supplies a committed requirements pin, and runs the actual doctor
dependency probe and actual adoption validator. It is explicitly skipped when
no external interpreter is supplied; ordinary discovery alone is not its proof.

Read durable logs and numeric exit files:

| Evidence | Observed result | Limit |
| --- | --- | --- |
| `/tmp/a120-external-doctor-tests` | 11 tests, exit 0 | Earlier intermediate doctor source, not the final 15-test inventory. |
| `/tmp/a120-external-loop-red` / `loop-green` | 12 tests: exit 1 / 0 | Historical symlink-loop repair proof. |
| `/tmp/a120-external-boundary-red` / `boundary-green` | 2 tests: exit 1 / 0 | Alias/dot and malformed UTF-8 repair proof. |
| `/tmp/a120-external-adoption-integration` | 40 tests, exit 0 | Logged suite proof at that execution point; not a gate. |
| `/tmp/a120-external-typedb-version` | `3.11.5`, exit 0 | External environment metadata. |
| `/tmp/a120-external-sdd-doctor` | external mode, driver 3.11.5, exit 0 | Resolved external CPython 3.14.5; not full SDD check. |
| `/tmp/a120-external-sdd-test` | scripts 15, logic-map 115, SDD 57, history 22, PM 15, deploy 18; Go logic-map passed; exit 0 | Suite log; final-source rerun/gate evidence belongs to Manager. |
| `/tmp/a120-external-diff-check` | exit 0, no diagnostics | Execution attributed to Terra. |

The dedicated adversary separately records final-source 15 doctor tests, actual
external doctor success, one real adoption fixture and three unchanged source
guard tests, all exit 0. That is attributed evidence, not a reviewer rerun.
Final AST inventory binds doctor hash below; branch map gives relevant source
locations and named tests, not measured exhaustive branch coverage.

## Reviewed identities

SHA-256 values read directly from current files:

| File | SHA-256 |
| --- | --- |
| `tools/sdd/sdd_doctor.py` | `e038dce83e10cec6df49977a20c6bcb8dc5d873edae6bf81a95e08f2f7f04268` |
| `tools/sdd/test_sdd_doctor.py` | `43fca70980b7bd24c8cc9931eff04bea2c2d78788dd301f84dcd22b274c2369a` |
| `tools/logic-map/test_check_analysis.py` | `85e0de6a38ebbc5e68ea00781c314bc01db416fe44742a30d5df74c312b73430` |
| Unchanged helper `tools/logic-map/execution_baseline.py` | `e9d8a55c718fe3568d09facb355e507c06c565f5bed915ef1b0f780d0104e1c3` |

Final SDD synchronization/check, actual gate, Manager acceptance, archive,
adoption of a063 and operating-profile/runtime proof remain separate conditions.

## Final verdict

Pre-Landing Review: No unresolved issues. Two P2 regression gaps were fixed
by Terra and independently verified before this subsequent gstack signoff.
**CLEAR for the external-interpreter implementation amendment.**

This is code/review clearance at the hashes above. Final full SDD verification
was still running at signoff; no gate, archive or runtime completion is claimed.
No tests were executed by this reviewer, and no nested-model result was used.

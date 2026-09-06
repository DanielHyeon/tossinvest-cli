# Post-adversarial gstack implementation review

Date: 2026-09-06. Reviewer: separate Codex review context, independent of the
implementation teammate and preceding adversarial reviewer.

Status: **DONE_WITH_CONCERNS — CLEAR for the reviewed implementation**.
This is an implementation review, not a final gate, a063 adoption, runtime
acceptance, archive, or authorization to integrate the worktree.

## Scope and method

Reviewed the installed `/home/daniel/.agents/skills/gstack/review/SKILL.md`,
`review/checklist.md`, Greptile instructions and testing, maintainability,
security, performance and red-team checklists. Applied their substantive
checks in this single reviewer context after reading the final CLEAR section
of `../adversarial-implementation-review.md`.

The comparison is frozen E `e65e394bf84b3c6e4559a219e816af96d341d75d` to the
actual isolated worktree, including uncommitted files. Detached WIP HEAD
`19021ada90135532adfaa70712fb199e42e1c95b` is not the reviewed final content.
Read current full helper and checker implementation, all changed test bodies,
workflow/tool guidance, a120 proposal/design/spec/tasks/reviews and evidence,
PM changes and the a119 documentation dependency. Fixtures were read as
defensive regression data, not merely summarized.

Scope check: **CLEAN**. The implementation is one fixed a063/P/E policy helper,
immutable endpoint inventory support, checker integration, isolated fixtures
and documentation. The a119 files repair the expressly declared validation
dependency; no a119 feature or Go product code is implemented here. The ordinary
checker remains the full map validator. The live sell/cancel/amend items in
`TODOS.md` are unrelated and remain open.

## Findings and disposition

### R1: fixture branch depends on global Git configuration — fixed

**P2 / INFORMATIONAL, confidence 10/10, testing.** The original helper fixture
ran `subprocess.run(["git", "init", "-q"], ...)`, but two cases used
`subprocess.run(["git", "checkout", "-q", "master"], ...)`. With
`init.defaultBranch=main`, both failed before reaching the intended assertions.

The reviewer executed this concrete reproduction before Manager restated the
Terra-only execution boundary:

```text
GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=init.defaultBranch GIT_CONFIG_VALUE_0=main \
python3 -m unittest \
  test_execution_baseline.ExecutionBaselineUnitTests.test_real_git_history_keeps_merge_parent_and_reverted_path \
  test_execution_baseline.ExecutionBaselineUnitTests.test_valid_fixture_refuses_symlink_base_attached_head_and_dirty_tracked_source
Result: exit 1, 2 errors, pathspec 'master' did not match.
```

Terra changed `_fixture()` to explicit `git init -q -b master`. The reviewer
read that fix and `/tmp/a120-gstack-p2-cli.log` plus `.exit`: actual exit 0,
three tests, including both named cases under the override and the CLI unit
test. This finding is closed; no global Git configuration was changed.

### R2: successful adoption lacked its required exception label — fixed

**P2 / INFORMATIONAL, confidence 10/10, completeness/DX.** The frozen spec
requires `execution-baseline adoption exception`, but original `main()` always
printed `evidence complete or diff-proven exempt` after a successful check.
The reviewer invoked the real detached adoption fixture and `main()` with only
P/E and argv patches: actual return 0 and that generic output confirmed the gap.
This was the second reproduction already underway when the role reminder arrived.
All subsequent execution was delegated to Terra.

Terra retained `check()`'s error-list API and added optional validator-derived
context through `resolve_base` and `check`. `main()` checks errors before
selecting the success label. Successful adoption now prints the literal
exception label; ordinary success retains its previous text. The mismatch
diagnostic now names the selected effective comparison base, covering P and E.
There is no second expensive adoption-validation pass to obtain the label.

Read `test_main_distinguishes_ordinary_adoption_and_invalid_results` and the
real `test_main_real_adoption_prints_exception_label_only_after_validation`.
The latter runs actual helper/checker validation, then commits a malformed
record and checks nonzero status with no success label. Terra's
`/tmp/a120-gstack-cli-adoption.log` and `.exit` record one test, exit 0.
This finding is closed on code inspection and supplied executable evidence.

### R3: historical branch IDs and current evidence status — correction recorded

**INFORMATIONAL, confidence 10/10, evidence/documentation.** Historical
`resolve_base` prose reverses AST B2/B3. Historical
`changed_existing_functions` prose groups eight rows against sixteen AST
entries; historical `check` groups six rows against seventeen AST entries.
These are not exact per-ID coverage maps. The initial post-edit `functions.json`
contained only function names/ranges and did not fix that limitation.

Manager authorized the reviewer to write
`../python-function-logic-corrections.md`: it retains the original artifacts,
records every historical AST ID/location/meaning, and expressly rejects a
retrospective pre-edit-compliance or executed-coverage claim. Current source
received separately refreshed AST branch evidence. The new `main`
pre-edit artifact's omitted error-loop entry was also corrected, with the
post-edit timing disclosed in `../pre-edit-main-evidence-correction.md` and the
reviewer's historical-corrections document.

The implementation-evidence status table was reconciled: the earlier 31-checker
result is explicitly historical and a separate current matrix cites the later
fixture evidence. The reviewer inspected `final-function-ast.json`, its exact
current source hash and all 42 If/For/Try locations across the four changed
checker functions. `final-branch-test-map.md` enumerates those same locations,
maps changed policy paths to fixtures and explicitly leaves other existing
routes uninstrumented. This is sufficient current change-path evidence; it is
not exhaustive per-edge test coverage. R3 is resolved by correction and truthful
limits, without restoring a claim of perfect original process compliance.

## Contract and checklist assessment

| Area | Reviewed evidence and conclusion |
| --- | --- |
| Fixed policy and ordinary fallback | `validate` only returns no adoption for an absent record; fixed change/P/E, committed full P, exact integer schema and provenance checks gate E selection. Ordinary dirty-Go, P override and same-base-reference fixtures exercise real Git/checker paths. |
| Source lock | `go_tree_entries` plus `verify_source_go_lock` compare S/H Go paths, mode/blob and actual regular worktree bytes/executable mode. The explicit mode fixture covers `core.filemode=false`; symlink and source substitution cases are present. |
| Hidden build inputs | Filesystem enumeration covers ordinary and ignored entries without following directory links. Metadata restrictions and both `go_inputs` profiles remain mandatory. Tests exercise allowed metadata roots, forbidden Go/C/assembly/header inputs, ignored embed data, profile union and errors. Input-overlap fault injection is mocked and is not claimed as an actual embedded production asset. |
| Evidence trust and history | Strict JSON rejects duplicate keys; committed regular evidence paths and SHA-256 bind record/ledger/reviews. `range_payload` shares history and function serialization; exact recomputation compares both endpoint inventories. Merge-parent/reverted-path and deleted-function fixtures were read. Review hashes bind documents; they do not mechanically prove a human or model really performed a review. That remains the recorded review workflow. |
| Full map obligations | `check` still validates every ordinary required binding, source hash and revision after E selection. Existing branch/call/test-citation validators are unchanged. Real adoption fixtures reject missing maps, stale hashes, incorrect deleted-function revision and local/reference coexistence. |
| Generator writes | Fixed output paths and parent checks precede writes; exclusive creation prevents overwrite; the failure fixture verifies cleanup of its own ledger after record creation fails. No baseline rewrite, checkout mutation or approval generation is added. |
| Shell/SQL/LLM safety | Changed subprocess calls use argument arrays; no `shell=True`, eval, SQL, LLM output execution, HTTP endpoint or live trading call was added. Strict JSON/path handling is the relevant trust boundary. No new applicable enum consumer was omitted. |
| Performance/maintainability | One recomputation per immutable range and both Go enumerations are deliberate one-time audit costs. Context propagation avoids duplicate validation for reporting. No async server, query loop, frontend bundle, migration, public API or release pipeline is changed. |
| Documentation and scope | WORKFLOW and logic-map README state exact P/E, ordinary behavior, exception/debt limits and draft command. PM links match a120; proposal identifies a119 dependency. No new root README feature promise is required for this internal workflow exception. |

No additional concrete security, data-loss or map-waiver defect was found in
the final helper/checker implementation. This does not assert exhaustive
branch execution, concurrent-writer safety, hermetic external toolchains or
runtime correctness. Acceptance is explicitly for an isolated controlled
worktree; the design excludes concurrent writers and broader environment claims.

## Execution provenance and skill limitations

- The preceding independent adversary records actual helper 27/27, checker
  37/37 and final ordinary/reference matrix 2/2 before the CLI follow-up. Those
  are that reviewer's runs, not new full-suite execution by this reviewer.
  Manager attempted to resume that reviewer for the later delta, but the host
  rejected the follow-up with `agent thread limit reached`. No later adversary
  rerun is claimed. This gstack reviewer and Manager read the final CLI/context
  and fixture changes; Terra supplied the subsequent execution evidence.
- The two direct reproductions above are disclosed role deviations. After
  Manager's reminder, this reviewer only read code/logs and wrote the two
  authorized review documents; Terra executed fixes and regressions.
- Skill steps 0–3 used the user-fixed local E baseline and detached worktree.
  No remote fetch or remote PR state was needed. Greptile, version queue and
  unavailable slop-scan integration were not invoked.
- Critical/informational and specialist checklists were applied in one context.
  No independent specialist army, nested Claude agent, external Codex CLI,
  cross-model consensus or structured external-model P1 gate is claimed.
  UI/design, API contract and database migration specialists are not applicable.
- Setup, upgrades, hooks/configuration, telemetry, global review/learnings logs,
  artifact sync, external messages, commits and memory writes were excluded by
  the authorized review scope. This repository report is the review record.
- Pre-edit CodeGraph/AST evidence was read; current index freshness, full SDD
  tests, broad Go/seams/race/vet runs, strict validation, `make sdd-sync`,
  `make sdd-check` and the final gate remain separately owned acceptance proof.

## Reviewed source fingerprint

| File | SHA-256 |
| --- | --- |
| tools/logic-map/check_analysis.py | f7d95a10cee85773189b4ee446a282eed7fb78021e361868029bda4ae5f1ebf2 |
| tools/logic-map/execution_baseline.py | e9d8a55c718fe3568d09facb355e507c06c565f5bed915ef1b0f780d0104e1c3 |
| tools/logic-map/test_check_analysis.py | 7b4046c5573920945461c1562b185696a14b0f2527119fa1e2f176ef07eaf4b1 |
| tools/logic-map/test_execution_baseline.py | 95bbd64dd0acc61a69ee1b18a809f3a75e80e867fc1ad7b4a6d526d99b8182ea |

## Final disposition

Pre-Landing Review: **no unresolved implementation findings**. R1/R2 are fixed
and supported by actual Terra regression results; R3 has a source-bound refresh
and explicit historical corrections. The concern retained in DONE_WITH_CONCERNS
is the disclosed pre-edit evidence/process defect and reviewer execution-role
deviation, not an unreported code blocker. No whole-change completion is asserted.

Recommendation: proceed to the separately required acceptance checks because
the concrete Git-branch and exception-label defects are fixed and the remaining
historical evidence limits are now explicit. Manager must still require final
SDD freshness/check, actual gate success, acceptance and archive validation.

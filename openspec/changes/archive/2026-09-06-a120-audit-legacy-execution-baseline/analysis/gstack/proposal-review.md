# a120 proposal-freeze autoplan review

Date: 2026-09-06. Reviewed HEAD: `e65e394bf84b3c6e4559a219e816af96d341d75d`.
Target: the current a120 proposal, design, tasks and `sdd-workflow` delta in the
isolated worktree, including the final B2 amendment and independent adversarial
proposal review. This review is a separate context from the author and adversary.

Method: installed gstack `autoplan`, `plan-ceo-review`, `plan-design-review`,
`plan-eng-review`, and `plan-devex-review` review methodologies, including their
referenced review sections. CEO, Design applicability, Eng and DX were assessed
sequentially in this one context. External CLI voices, nested agents, telemetry,
global review logs, skill setup and files outside this report were excluded by
the assigned review scope. Four perspectives do not mean four independent models.
The test plan and decision trail are embedded here instead of global artifacts.
The original plan is unchanged, so no external restore-point copy is needed.

## Phase 1 — CEO: problem, scope and alternatives

The premise reviewed is that a063's old planning base should remain visible while
resumed implementation is checked against a fixed, separately reviewed execution
base. A blanket baseline reset would lose the distinction the user needs to audit.
Doing nothing leaves the reported inherited-history blocker in place; reporting
historical maps as originally complete would misstate the evidence. The review
accepts the already-authorized exceptional historical adoption premise, without
reconfirming it or claiming to independently rerun the reported 338/327 counts.

| Approach | Effort / risk | Coverage and reuse | Disposition |
| --- | --- | --- | --- |
| Rewrite a063 base or add an environment override | Small / high assurance risk | Smallest diff, but removes provenance and permits hiding implementation | Reject: contradicts the approved boundary |
| Reconstruct all inherited maps under P | Large / historical truth remains limited | Reuses ordinary checker; retrospective documents still cannot establish original pre-edit timing | Outside this authorized migration |
| Fixed a063/P/E exception with ledger, source lock and complete E maps | Medium / bounded policy risk | Reuses ordinary function derivation and bundle validation while displaying inherited debt | Retain current plan |

The ideal long-term state is immutable ordinary baselines with separately visible
historical debt. This plan moves toward that state by naming the exception and
limiting it to one tuple; a reusable baseline-selection platform would move away
from it. Expansion candidates considered were a general migration registry,
automatic cleanup, review auto-approval, a status dashboard, and an inventory
cache. None is necessary: the first three widen authority, the dashboard adds UI
scope, and a cache risks bypassing fresh source validation. No expansion is added.

### Existing components and architecture

`tools/logic-map/check_analysis.py` already provides `resolve_base`,
`changed_existing_functions`, `go_functions` and the full `check` pipeline.
The supplied Python AST/maps identify their pre-edit boundaries before this
review's code reading. Existing `test_check_analysis.py` uses unittest and covers
invalid bases, environment mismatch, failed diff, absent base file, new functions
and failed exemptions. `Makefile` already exposes SDD, untagged, seams, race, vet
and gate entry points. Reusing them avoids creating a second acceptance engine.

```text
BEFORE: base-commit.txt P -> ordinary inventory(P, worktree) -> full FLM checks

AFTER:  base-commit.txt P + optional fixed-path record
                         |
             +-----------+-------------------+
             | absent                        | present
             v                               v
       ordinary P                    validate fixed tuple and P bytes
                                             |
                           committed ledger + reviews + S/H envelope
                                             |
                           exact history/inventory + source/input lock
                                             |
                                   E selected only if valid
                                             |
                        ordinary inventory(E, worktree) == recorded E..S
                                             |
                              unchanged full FLM checks -> existing gate

draft generator -> same immutable history/inventory serializer -> draft files
                   (no baseline reset, approval, commit or checkout switch)
```

### CEO sections 1–11

1. **Architecture:** the new helper is a policy validator around the existing
   checker, not an alternative checker. Git objects establish P/E/S history;
   detached H establishes the reviewed evidence state. The fixed tuple makes
   adding another migration an explicit future code/spec decision.
2. **Error/rescue:** the plan rejects malformed records, failed Git/AST derivation,
   failed package enumeration and mismatched evidence rather than falling back.
   Missing optional records retain ordinary behavior, whereas present empty records
   fail. The registry below makes these different outcomes visible to implementation.
3. **Security:** this introduces local file and Git inputs, not credentials or a
   service endpoint. Committed regular-file checks, exact paths, digest binding,
   duplicate-key rejection and immutable recomputation constrain substitution.
   The final B2 rule closes source-looking and embedded-input cache exemptions.
4. **Data flow:** draft creation and adoption acceptance are distinct transitions.
   Empty or reordered inventories cannot substitute for recomputation, and a
   same-count ledger cannot hide omitted rows. Source S precedes evidence H,
   avoiding a self-referential commit reference.
5. **Code quality:** the design requires one serializer shared by generator and
   validator and the same function derivation as ordinary checks. Splitting a new
   helper from integration keeps the policy locally reviewable. No generalized
   policy framework or duplicate Go parser is needed.
6. **Tests:** tasks 2.1/2.1a specify real Git fixtures, positive controls and
   negatives for each new acceptance boundary. Existing regression tests supply
   ordinary-behavior anchors, not proof that adoption already works. The full
   mapping below must be implemented before acceptance.
7. **Performance:** history traversal, endpoint AST extraction and two package
   enumerations dominate cost. This is one fixed migration at gate time, so doing
   each required range once is appropriate. No latency or memory measurement is
   claimed and no cache is permitted to replace validation.
8. **Observability:** the record binds commit identities, inventories and actual
   review documents, allowing a later reviewer to reconstruct the decision.
   Failures should identify the offending path, record field or command under
   task 2.4's recovery guidance. Telemetry and an operational dashboard add no
   necessary evidence to this repository-local gate.
9. **Rollout:** implement and accept a120 in its isolated worktree before applying
   the exception to a063. Then commit S, produce/review evidence, commit H and
   run acceptance. Reverting tooling restores P-based blocking; no runtime or
   database migration is involved.
10. **Trajectory:** historical missing FLMs remain named debt, even if the exception
    succeeds. The record does not normalize recapture for future changes. Tooling
    rollback is reversible, while rewriting history or deleting evidence is
    explicitly excluded.
11. **Design/UX:** there is no a120 screen or rendering change. Console paths in
    the allowlist constrain later a063 source; they are not a120 UI deliverables.
    The developer-facing status, draft and error experience is assessed in DX.

### Error and rescue / failure modes registry

These are required implementation behaviors, not claims that new handlers or
tests already exist. Exception names refer to likely Python/Git interfaces;
the observable requirement is a failure exit and useful diagnostic.

| Path | Failure | Required outcome / recovery | Planned proof |
| --- | --- | --- | --- |
| Optional record lookup | Record absent | Ordinary P; no exception label | Ordinary positive control |
| Record parsing | Empty, duplicate keys, wrong fields/types/digest | Value/JSON error; fail, correct record, never choose P silently | 2.1 strict schema negatives |
| Tuple/base validation | Wrong change, P bytes, E, S tree or ancestry | Reject exact mismatch; retain original P | 2.1 fixed tuple / S=H negatives |
| Committed evidence read | OSError, symlink, escape, altered H bytes | Reject path; regenerate/review and commit legitimate evidence | 2.1 evidence substitutions |
| Git/AST recomputation | Command error, lost object, malformed AST | Fail derivation with command context; do not treat as zero rows | 2.1 + ordinary failure tests |
| Ledger equality | Omitted/duplicate/reordered rows, merge-parent omission | Reject exact disagreement; regenerate complete draft | 2.1 merge/revert/row negatives |
| Snapshot scan | Staged, unstaged, missing/extra or symlinked source | Reject drift; use correct committed isolated worktree | 2.1 source lock fixtures |
| Metadata and Go inputs | Source/executable in cache, embed overlap, either go list fails | Reject offending input or failed enumeration; no blanket cache escape | 2.1a positive and negative fixtures |
| E-based bundle check | Missing current/base map, bad hash, reference conflict | Existing FLM rejection; supply real complete maps | 2.3 + existing checker regressions |
| Draft output | Destination already exists or write fails | Refuse overwrite; preserve existing evidence | 2.2 generator tests |

There is no unhandled silent-success path in the reviewed contract. Actual error
handling, including timeout diagnostics, remains an implementation review item.
CEO completion: scope retained, zero new blockers, no added/deferred product tasks.

## Phase 2 — Design applicability and trust presentation

The plan adds a local checker/helper and documentation. It adds no screen,
component, layout, typography, responsive behavior or browser interaction.
The installed Design skill's no-UI exit applies; generating mockups would invent
scope. All seven visual dimensions are therefore not-applicable with this reason.

| Design dimension | Assessment |
| --- | --- |
| Information architecture | CLI status must distinguish draft, exception accepted and final gate; DX below |
| Interaction states | Missing/invalid/valid record states mapped above; no visual states |
| User journey | Developer evidence sequence only; mapped in DX |
| AI visual slop | No generated UI |
| Design system | No component/style changes |
| Responsive layout | No viewport surface |
| Accessibility | No new graphical interaction; text diagnostics belong to DX |

Design completion: no visual score assigned and no visual deficiency invented.
Trust presentation remains material: `execution-baseline adoption exception`
must be visible, and `retrospective-exception` must never become a claim of
original pre-edit compliance.

## Phase 3 — Engineering review and test plan

### Architecture and code quality

The existing `resolve_base` at line 192, `changed_existing_functions` at line 86
and `check` at line 510 are the planned integration boundary. Their pre-edit
Python maps record the ordinary mismatch, derivation-failure, reference and
bundle-validation paths. No Go application function is being edited by a120.
Function Logic Map: not-applicable — for a120's Go source set only; supplied
Python pre-edit artifacts and later Python review remain required.

The design's fixed tuple, regular committed P, strict S-before-H, committed
evidence and exact range equality address B1/B3/B4. B2 now enumerates static
generated-metadata locations, uses lstat without symlink traversal, rejects
executable/source/object/library candidates and checks both real Go enumeration
modes for any input overlap. The updated separate adversarial report is explicitly
clear for proposal freeze, and this review found no remaining B2 contract gap.

The ordinary checker and its validation errors should be reused. In particular,
passing the adoption envelope must not bypass `check`'s reference coexistence,
duplicate-binding or source/revision checks. If implementation changes additional
existing Python functions, including the CLI status printer at `main`, their
pre-edit analysis must be captured before editing; the current three maps are
not a blanket exemption for any checker change.

### State and test coverage diagram

```text
generator inputs(P,E,S)
  + valid immutable commits -> exact ledger + non-approving draft [2.2 integration]
  + existing output/invalid input -> fail without overwrite       [2.1/2.2 negative]

check(change)
  + no record -> ordinary P -> existing full bundle path           [ordinary regressions]
  + present record
      + empty/invalid/another tuple -> FAIL                        [2.1 schema/tuple]
      + valid tuple
          + P or S/H ancestry/tree wrong -> FAIL                   [2.1 ancestry/base]
          + record/review/ledger substitution -> FAIL               [2.1 file/hash/path]
          + incomplete/reordered history -> FAIL                   [2.1 merge/revert]
          + S/H tracked drift -> FAIL                              [2.1 clean worktree]
          + extra ignored/untracked/symlink input -> FAIL           [2.1a source inputs]
          + allowlisted plain metadata, not Go input -> continue    [2.1a positive]
          + either go list or JSON decode fails -> FAIL             [2.1a enumeration]
          + exact E..S inventory -> E ordinary FLM validation
              + absent deleted-function base map -> FAIL           [2.1/2.3 deletion]
              + absent/altered current map -> FAIL                  [2.3 full bundle]
              + reference conflict/duplicates -> FAIL              [2.3 preservation]
              + all checks pass -> labeled exception               [2.2/2.3 positive]
                  -> ordinary tests + sdd + gate still required     [3.1-3.4 / gate]
```

All adoption nodes are planned, not tested by this review. No LLM invocation or
prompt behavior is changed, so model evals are not applicable. No background job
or browser flow is added. Real Git and real Go-package-enumeration integration
tests are needed because mocks alone would hide graph and embed-input errors.

### Embedded implementation test plan

| Area | Positive control | Negative/boundary cases | Expected evidence |
| --- | --- | --- | --- |
| Ordinary compatibility | Existing no-record change on normal worktree | Bad persisted P and mismatching env | Existing tests plus no-record integration |
| Exception identity | Exact a063/P/E and full S/tree | Other IDs, abbreviated refs, changed/symlink P, E moved later | Real Git fixture assertions |
| Snapshot timeline | P <= E <= S < detached H | S=H, wrong ancestry/tree, attached/dirty H, evidence outside H | Clean detached fixtures |
| Immutable source | S paths/modes/blobs unchanged | Rename/delete, staged additions, ignored Go, tracked/untracked symlink | Rejection after one independent mutation |
| Generated metadata | Actual allowed index/state/GBrain/pyc shapes | Executable/source/object suffixes, arbitrary location, symlink component | Positive controls and one-guard negatives |
| Actual inputs | Both Go modes enumerate valid packages | C/assembly/header/embed input, cache-file embed overlap, either command/JSON error | Real input-selection fixtures; failure stubs only for command errors |
| History | Merge plus reverted path and base deletion | Omitted parent/commit/path/function, duplicate/reordered/altered rows | Exact computed row comparison |
| Evidence trust | Two distinct actual review files and valid ledger hash | Same/escaped/symlinked/uncommitted/changed files, duplicate JSON keys | Binding and strict parser fixtures |
| Full FLM obligation | Every E-derived current/base function covered | One missing deletion bundle, bad hash, reference conflict, duplicate binding | Ordinary checker path is exercised |
| Generator | Drafts can be reviewed independently | Existing output, invalid source reference | No overwrite, baseline or checkout mutation |

Use unittest discovery in `tools/logic-map` for focused tests; `make sdd-test`
already includes that suite. Follow with the plan's `make test`, `make test-seams`,
`make test-race`, `make vet`, strict OpenSpec/PM checks, final SDD sync/check and
`make gate CHANGE=a120-audit-legacy-execution-baseline`. This report executes none
of these commands and does not replace their actual exit records.

### Performance and sequencing

The three expensive operations are full history/path reconstruction, endpoint
AST extraction and package dependency enumeration. The plan fixes the relevant
range and requests one derivation per range; that is sufficient for this one-time
acceptance path. The external toolchain/module cache is explicitly outside the
repository source lock, so no hermetic environment claim follows. Actual timing
and resource behavior must be observed during implementation verification.

Sequential implementation is appropriate: helper schema/serializer, checker
integration and fixtures share the same policy and module. Terra implements;
the separate adversary reviews; gstack follows; Manager independently verifies.
No concurrent writer should modify S/H acceptance inputs during their checks.
The supplied pre-edit report does not claim its interrupted sdd-sync passed;
final hard-evidence refresh remains task 3.3.

Engineering completion: zero new architecture/code-quality/performance blockers;
the planned test matrix covers the contract and remains wholly unexecuted here.
No new implementation task beyond tasks 2.1–3.4 is introduced by this review.

## Phase 4 — Developer experience

Persona: the TossOS maintainer or implementation agent diagnosing an FLM gate
failure, with Git/Python/Go already installed through the repository workflow.
The product here is a repository-local audit tool, not a public onboarding API.
DX posture is polish within the fixed policy; making arbitrary baseline selection
easy would undermine the purpose of this change.

### Developer perspective

I open `tools/logic-map/README.md` because the gate tells me maps are missing.
It shows scaffold, Go AST, risk report and checker commands, and it says my CI
base must match `base-commit.txt`. I need the new workflow documentation to explain
why a063 alone can select E while that normal rule continues for other changes.
I do not want to guess whether an error means bad JSON, a dirty checkout, or an
unreviewed source file. I need the offending field or path and a repair direction
that preserves the original evidence. The proposed draft generator is useful
because it can enumerate history without changing my checkout, but a draft must
not look like an approval. I expect to inspect the ledger, obtain two actual
reviews, commit evidence after S, and run the checker at detached H. If the
checker accepts that exception, I still need every E-based function map and the
ordinary tests. The literal exception label tells me what passed. It must not
tell me that historical pre-edit work or a063's operational acceptance is complete.

### Journey and staged feedback

| Stage | Developer action | Reviewed handling |
| --- | --- | --- |
| 1 Discover | Read WORKFLOW and logic-map README after FLM failure | Task 2.4 documents the narrow exception |
| 2 Install | Use existing Git/Python/Go toolchain | No new service/package distribution |
| 3 First feedback | Run existing checker or future helper help | Existing `--change`/`--root`; exact new help belongs to implementation |
| 4 Prepare | Commit reviewed source S after fixed E | Full commits/tree and explicit ancestry |
| 5 Generate | Produce ledger/record drafts | No overwrite, approval or checkout mutation |
| 6 Review | Obtain adversary then gstack adoption reviews | Distinct actual reviews bound by hashes |
| 7 Accept | Commit evidence H and run detached check | Source/input closure, complete E maps, labeled exception |
| 8 Debug | Read failure and repair legitimate source/evidence | Error registry above; no clean/reset workaround |
| 9 Upgrade/rollback | Integrate accepted tooling or revert it | Ordinary P behavior preserved/restored; a063 gate remains separate |

Time to first useful feedback is not measured because the helper does not exist.
Target: locate the relevant help and receive a field/path-specific result within
one terminal session; first-feedback documentation target under five minutes.
That target excludes the deliberately mandatory human reviews, commit creation
and full gate. Public SaaS onboarding comparisons are not a meaningful benchmark
for this one-time audit exception and no external performance figures are claimed.

### DX scorecard — plan readiness only

| Dimension | Score | Evidence and what remains to verify |
| --- | --- | --- |
| Getting started | 8/10 | Existing checker commands plus 2.4 guidance; first-feedback timing not measured |
| API/CLI design | 8/10 | Fixed change/record path, read-only validation and draft-only generation; new help pending |
| Errors/debugging | 8/10 | Fail-closed cases are explicit; actual problem/cause/repair messages pending |
| Documentation | 8/10 | WORKFLOW + tool guidance required in 2.4; exact examples must match delivered CLI |
| Upgrade/migration | 9/10 | Absent record preserves P; invalid record fails; rollback restores ordinary policy |
| Environment/tooling | 8/10 | Real Git fixtures, both Go modes, safe generated metadata; full checks pending |
| Community/ecosystem | N/A | Internal single-tuple tool; no public API, plugin market or support channel added |
| Measurement/feedback | 8/10 | Actual exit retention and separate reviews specified; no timing baseline yet |

Overall applicable plan score: 8.1/10, not an implementation usability rating.
Remaining score limits are delivery/measurement work already in tasks 2.4/3.1,
not unresolved proposal decisions. No artificial 10/10 runtime score is assigned.

Three concrete error experiences were traced: a wrong environment base currently
reports a persisted-base mismatch; the new path must identify selected E only
after valid adoption. A missing bundle currently names the modified function;
that diagnostic must survive E selection. A package-enumeration failure must say
which mode failed and must not advise ignoring source inputs. These expectations
fit the existing task 2.4 recovery documentation and negative test obligations.

DX implementation checklist: document normal versus exceptional paths; show
literal exception/provenance labels; provide accurate generator/checker examples;
refuse overwrite clearly; name offending fields/paths/modes; retain nonzero exits;
keep final gate and operational acceptance distinct. These are implementation
verification details of the accepted contract, not new feature scope.

## Decision trail, exclusions and cross-phase limits

| Decision | Classification / principle | Rationale |
| --- | --- | --- |
| Retain only fixed a063/P/E adoption | Mechanical / explicit | Matches authorized scope, blocks arbitrary recapture |
| Share ordinary inventory and one canonical serializer | Mechanical / DRY | Prevents generator/checker disagreement |
| Require both metadata constraints and actual Go input closure | Mechanical / completeness | Cache paths alone cannot exclude embedded input |
| Keep all E-derived FLMs and ordinary gate | Mechanical / completeness | Historical migration does not validate resumed code |
| Keep draft creation separate from approval | Mechanical / explicit | File production cannot fabricate review |
| No general migration platform/cache/dashboard | Mechanical / pragmatic | Unnecessary authority and scope expansion |
| No runtime success claim at proposal freeze | Mechanical / explicit | New implementation and tests do not yet exist |

NOT in scope: recapturing P, arbitrary future tuples, deleting evidence, performing
actual a063 adoption, fixing a119 runtime behavior, live orders/process changes,
operational a063 acceptance, and fabricating historical pre-edit compliance.
TODOS.md's live sell and KR cancel/amend checks are unrelated and remain unchanged.
No new TODO or taste decision requires user input.

Cross-phase themes are evidence truthfulness, preserving ordinary checks and
making failure repair understandable. They recur within this one context and are
not independent cross-model consensus. External Codex/Claude voices and their
consensus tables are N/A in every phase; no model identities or votes are invented.
The separate adversarial proposal report is a real upstream review artifact,
whose final B1–B4 clearance was read before this verdict.

## GSTACK REVIEW REPORT

| Review | Runs | Status | Findings |
| --- | --- | --- | --- |
| CEO | 1 single-context pass | CLEAR | Fixed exceptional policy appropriate; no scope expansion |
| Design | 1 applicability pass | N/A visual scope | No UI; trust wording assessed in DX |
| Engineering | 1 single-context pass | CLEAR | B1–B4 resolved; complete implementation test matrix retained |
| DX | 1 single-context pass | CLEAR for plan | Documentation/error delivery and measurement remain implementation work |
| External CLI voices | 0 | Not invoked by scope | No cross-model consensus claimed |

VERDICT: CLEAR FOR PROPOSAL FREEZE. Manager may release Terra to implement this
exact contract after recording the freeze. This verdict does not mark a120
implemented, validate an a063 adoption record, waive E-based maps, pass a final
gate, or grant operational acceptance. All final checks and independent
implementation review remain mandatory.

NO UNRESOLVED DECISIONS

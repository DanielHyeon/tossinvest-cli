# External SDD interpreter amendment: gstack planning review

Date: 2026-09-06. Scope: design decision 5, its new `sdd-workflow`
requirement, proposal explanation and task 2.5. This is a new planning review;
the earlier implementation CLEAR applies only to its recorded source hashes.

Method: bounded single-context application of the installed gstack `autoplan`,
`plan-eng-review` and `plan-devex-review` planning methodology and their review
sections. Strategy/scope was assessed first, visual-design applicability next,
then architecture, code quality, tests, performance and eight DX dimensions.
No external model, nested specialist, runtime probe, test, code edit, global
state update or user prompt was invoked. This is not a claim that the full
interactive, cross-model autoplan orchestration ran. The assigned scope already
fixes the review target and forbids unrelated setup/telemetry/approval prompts.

## Strategy and scope

The premise is supported by the actual integration surface: the doctor's
`report()` names `.sdd/.venv/bin/python`, while `Makefile` makes `sdd-doctor` a
prerequisite of `sdd-check`. The adoption contract excludes the local
environment's untracked source files and symlinks. Earlier ordinary-worktree
green checks cannot establish that the adopted-worktree gate is feasible.

The proposed repair changes dependency-probe selection rather than broadening
the source allowlist. That preserves the purpose of the original audit and is
within the common SDD integration problem. It creates no new service, package,
product feature or approval mechanism. The minimum implementation surface is
doctor selection/probing, focused tests and usage documentation; no new class
hierarchy or global Python installation is justified.

| Alternative | Benefit | Cost / decision |
| --- | --- | --- |
| Explicit external interpreter in doctor | Retains source guard and default behavior; permits isolated tooling dependencies | Selected. Requires path validation, exact dependency probe and documented environment propagation. |
| Exempt repository-local venv from source guard | Reuses current setup command | Rejected: admits large untracked executable/source trees and changes the approved source boundary. |
| Require TypeDB driver in global Python | Avoids a local venv | Rejected: changes global dependencies and interpreter assumptions beyond the bounded repair. |

Current state → proposed state → longer-term desired state:

```text
doctor local-venv requirement conflicts with adopted clean source
  -> explicit external doctor probe, guard unchanged
  -> repeatable documented adoption check using separately managed tools
```

Visual UI work is not applicable: no screen, component, styling or interaction
state is added. The terminal's success/failure wording is assessed under DX.
No market or public-onboarding benchmark is relevant to this internal repair.

## Engineering: architecture and existing components

```text
make gate -> make sdd-check -> sdd_doctor
                               |
                  SDD_PYTHON absent / explicit
                       /                 \
              existing local probe     validate external path
                                           |
                              probe requirements-pinned driver
                                           |
                                report required dependency status

adoption source validation remains independent and unchanged
```

Existing reuse: `command_status` provides argument-array execution, bounded
timeout and structured status; `report` assembles dependency results;
`main` already turns required dependency failure into nonzero status;
`tools/sdd/requirements.txt` is the existing version authority and currently
pins `typedb-driver==3.11.5`. Reuse these boundaries instead of adding a second
doctor or hardcoding another version authority.

`tools/sdd-history/refresh_indexes.py` also names a local interpreter for its
separate optional TypeDB ingestion. The amendment deliberately changes only
the doctor's dependency probe. Documentation must not describe `SDD_PYTHON`
as selecting every SDD command's Python or enabling TypeDB synchronization.
No change to that advisory worker is needed to satisfy this bounded plan.

Doctor pre-edit evidence arrived before this review closed. Read
`analysis/python-function-logic/tools-sdd-sdd_doctor--report/ast.json`, its
function/branch maps, CodeGraph evidence and external-environment proposal.
Its source hash matches the reviewed unchanged doctor. The existing `If` at
94:5 is N005; true reaches N016 and false N017. The map explicitly distinguishes
the five reported baseline unit tests from unmeasured report branches and new
interpreter cases. CodeGraph basename noise is separated from the actual local
`main` caller and three helpers. This is sufficient pre-edit evidence for
`report`; any additional existing function edit needs its own prior map.
The AST metadata calls its `ast.walk` ordinal "preorder"; these are
evidence-local traversal ordinals, not a language-stable preorder guarantee.
This wording does not affect the source-location binding or branch mapping.
The proposed external venv commands remain unexecuted in that artifact; this
review ran no baseline or environment setup commands.

## Engineering: code quality and failure handling

The optional variable must be distinguished by presence, so an explicitly
empty value cannot act like absence. Perform lexical containment and resolved
target checks separately; resolving first would lose the lexical check.
Use path-component containment rather than string-prefix comparison, including
root equality and sibling paths with similar names. Treat filesystem and
symlink-resolution errors as structured failures. An executable directory is
not an interpreter file. An external venv link to an external interpreter is
allowed, while either lexical or resolved containment inside the checkout fails.

Keep the invocation as arguments, including paths containing spaces. The exact
pin must come from the requirements file; a missing/unreadable/ambiguous pin,
missing package, version mismatch, process failure, malformed probe result or
timeout must not become success. These are implementation details of the
existing fail-closed requirement, not permission to expand the source guard.
No local-environment fallback or setup command should execute on an invalid
explicit value.

| Failure | Required observable outcome / recovery |
| --- | --- |
| Explicit path absent, relative, empty, directory or non-executable | Required doctor failure naming the invalid selection; choose a valid absolute external interpreter. |
| Lexical internal path or external link resolving internally | Required failure before dependency probe; use an environment outside the checkout. |
| Missing/mismatched driver or invalid pin | Required failure showing expected dependency and observed failure/version; install reviewed requirements into the external environment. |
| Process failure or timeout | Nonzero required status with bounded diagnostic; no local fallback. |
| Variable absent | Preserve current local compatibility and setup hint; no new global requirement. |
| Valid doctor probe but source contamination | Adoption still fails; doctor success cannot exempt a local venv or another build input. |

## Engineering: test plan

The existing `tools/sdd/test_sdd_doctor.py` uses unittest/mock and covers generic
missing CLI, command timeout, advisory service behavior and missing indexers.
It does not yet cover the new interpreter-selection contract. All new cases
below are planned, not executed by this review.

```text
selection
  absent -> existing local present/missing behavior           [unit regression]
  explicit
    empty/relative/missing/dir/nonexec -> fail, no fallback    [unit/filesystem]
    lexical-inside or resolved-inside -> fail before probe    [filesystem]
    valid outside regular file/link -> dependency probe      [filesystem]
      exact pinned version -> required status succeeds       [probe integration]
      missing/wrong version/error/timeout -> required fail    [unit fault injection]
combined adoption
  external probe succeeds, no local venv, valid S/H -> pass   [real Git integration]
  same fixture + forbidden local input -> source guard fails [real Git integration]
```

Tests for absent behavior must explicitly remove inherited `SDD_PYTHON`, since
the final SDD suite itself will run with that variable set. Explicit-selection
tests must set their own temporary value and prove a usable local venv is not
consulted on failure. Include external-to-external links, external-to-internal
links, internal-to-external links, symlink loops/broken links, a root whose path
is resolved differently, and a sibling directory with a shared textual prefix.
These filesystem cases make the two containment checks reviewable.

At least one real adopted-worktree scenario must exercise the successful
external dependency probe together with the unchanged adoption validation.
Mocked path checks alone cannot prove the integration that prompted this
amendment. Final SDD/check/gate commands must retain actual exits under the
documented external mode. Earlier Go/ordinary-SDD results remain scoped to
their actual source/mode; no extra runtime or account action is part of testing.

## Engineering: performance

This is a single local dependency probe, not a request-serving path. Retain a
bounded subprocess timeout and avoid recreating environments or installing
packages during each doctor invocation. Caching the result is unnecessary and
could obscure a changed interpreter or dependency. No load test, database
index, frontend bundle work or new observability service is warranted.

## Developer experience

Persona: the TossOS maintainer or authorized implementation agent preparing a
clean detached adoption worktree. They already have Git, Go, Python and SDD
tools; they need an actionable local diagnostic and reproducible commands.

Developer perspective: I reach the SDD doctor while trying to validate an
adopted source snapshot. The usual `make sdd-infra` creates the very local
environment that the source guard refuses. I need the documentation to explain
that conflict before suggesting setup. I should be able to create one external
environment from the repository's pinned requirements, pass its interpreter
explicitly, and see the selected path and driver result in the doctor output.
If the path is invalid, I need a path-specific error rather than a silent
switch back to the local environment. If the driver version is wrong, I need
the expected pin and the observed version or failure. After doctor succeeds,
I still expect the source guard, full maps and final gate to run. I should not
have to edit a baseline, delete evidence or change a service to make this work.
Unset mode should continue to serve an ordinary checkout exactly as before.

| Journey | Planned experience and acceptance |
| --- | --- |
| Discover | WORKFLOW/tool guidance identify the adopted-worktree conflict and doctor-only override. |
| Install | Create an external environment using existing pinned requirements; no global package mutation. |
| First feedback | Run doctor with explicit `SDD_PYTHON`; see selected interpreter and driver result. |
| Real use | Pass the same environment variable to SDD/check/gate commands. |
| Debug | Path error, dependency mismatch and timeout are distinct failures without local fallback. |
| Upgrade | Reinstall changed pinned requirements externally when intentionally updating the tool environment. |
| Rollback | Unset mode retains normal behavior; reverting this repair restores the known adoption blocker. |

| DX dimension | Plan readiness | Assessment |
| --- | ---: | --- |
| Getting started | 8/10 | Bounded external setup and first probe are specified; commands/output still require implementation verification. |
| CLI design | 9/10 | One optional variable, explicit invalid-value semantics, compatible absence. Doctor-only scope must be prominent. |
| Errors/debugging | 8/10 | Path, version and timeout cases are specified; actual messages remain to be reviewed. |
| Documentation | 8/10 | Existing guidance is the entry point; add copyable quoted commands and no-local-venv warning for adoption. |
| Upgrade/migration | 9/10 | No automatic migration or baseline change; normal setup is preserved. |
| Environment/tooling | 8/10 | External link/path and exact-pin tests plus real adoption proof are required; portability is not yet measured. |
| Community/ecosystem | N/A | No public package, pricing, extension system or community channel is added. |
| Measurement/feedback | 8/10 | Actual doctor and gate exits provide useful feedback; no invented onboarding timings or telemetry needed. |

The readiness scores evaluate the plan, not a tested product. Target first
feedback is one terminal session after dependencies are available; no numeric
time-to-success is claimed. No new scope or unresolved product taste decision
was identified. Existing live-trading TODOs remain unrelated.

## Decision trail and limits

Select external doctor probing, retain local absence behavior, preserve the
source guard and reuse the pinned requirements: explicit, minimal and reversible.
Keep invalid explicit selection fail-closed rather than guessing: necessary for
auditable environment identity. Require real adopted-worktree integration:
necessary because the earlier isolated successes missed this exact boundary.

No source/function behavior is declared proven by this plan review. CodeGraph,
pre-edit maps, Terra implementation/tests, independent amendment implementation
review, post-implementation gstack delta and final actual gates remain required.
The concurrently assigned independent adversary's verdict is not invented or
substituted by this single-context review; Manager must receive both clear
planning verdicts before authorizing implementation.

Reviewed design SHA-256:
`14481f764779b9567a62d6377166c0a5c8ea6a84b6abc6bbd23eda5619a24603`.
Reviewed delta spec SHA-256:
`4f2fb97b84937853aa933f641336cd8775636b270c0a9da4315c47d4e1bcd00c`.
Doctor source at review:
`eb6c4ecb4d1f80c17a1514bd93ba7683b37f40d5c8e508aec5c89e95a6a7c26d`.

## GSTACK REVIEW REPORT

| Review | Runs | Status | Findings |
| --- | --- | --- | --- |
| Strategy/scope | 1 bounded pass | CLEAR | Existing integration conflict warrants doctor-only repair. |
| Visual design | 1 applicability pass | N/A | Terminal/tooling amendment; DX assessed separately. |
| Engineering | 1 single-context pass | CLEAR FOR PLAN | No unresolved architecture blocker; explicit test matrix and pre-edit prerequisites retained. |
| DX | 1 single-context pass | CLEAR FOR PLAN | Doctor-only meaning, path/version diagnostics and external setup are implementation acceptance details. |
| External/nested voices | 0 | Not invoked by scope | No cross-model consensus claimed. |

VERDICT: CLEAR FOR AMENDMENT FREEZE, subject to Manager obtaining the separately
assigned adversary's clear verdict. This is not implementation or gate clearance.

NO UNRESOLVED DECISIONS

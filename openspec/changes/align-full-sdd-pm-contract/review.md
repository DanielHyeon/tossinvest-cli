# Proposal-freeze review: align-full-sdd-pm-contract

Date: 2026-07-31
Base commit: `1dbef864038c81f0a2982a03e7d9549369d21669`
Verdict: **ACCEPT WITH REQUIRED CORRECTIONS**

## Review composition

The proposal, design, delta spec, PM validator, PM tests, StockOS Full SDD reference,
and TossOS workflow were reviewed sequentially through Manager/CEO, adversarial
engineering, and developer-experience lenses. Repository policy for this session
prohibits spawning separate agents, so no independent-agent result is represented as
having occurred. Strict OpenSpec validation passed after correcting requirement
sentences so the normative keyword is present in the parser's first line.

## Findings and decisions

| ID | Severity | Finding | Decision |
|---|---|---|---|
| R1 | P0 | Written 1:1 policy is false in enforcement: 32 pre-existing active changes bypass Story mapping through the registry allowlist | Accept. Backfill exactly one Story for every one, cover this change with STORY-TOS-002, then delete the bypass |
| R2 | P1 | This change makes the active total 33; proposal wording could be read as covering only 32 | Clarify that 32 means pre-existing changes and STORY-TOS-002 covers the current change |
| R3 | P1 | Manual Story `status` can disagree with active/archive evidence | Replace with `intent` and derive `designed`, `in_progress`, `implemented`, or `archived` deterministically |
| R4 | P1 | A Story path must remain verifiable after archive adds a date prefix | Validate the declared active path exactly or the actual date-prefixed archive directory for the same change ID |
| R5 | P1 | Removing the allowlist before backfill would break all SDD gates | Backfill hierarchy and Stories first; switch validator and registry atomically in the same logical unit |
| R6 | P1 | Copying StockOS text mechanically would corrupt TossOS safety and tooling facts | Preserve the explicit TossOS list in design and verify it in the WORKFLOW diff |
| R7 | P2 | UI/design review is not applicable because this change has no operator UI or runtime route | No design phase artifacts required |
| R8 | P2 | Generated trackers can conceal stale manual state | Render only derived status and fail `--check` when generated files drift |

## Architecture and failure modes

```text
portfolio source (intent + hierarchy + openspec mapping)
              |
              v
PM validator -----> active OpenSpec directories
      |             archived OpenSpec directories
      |             tasks.md checkbox evidence
      v
derived Story state
      |
      v
generated trackers (read-only views)
```

| Failure mode | Prevention/test |
|---|---|
| Active change has no Story | Exact set coverage test |
| Two Stories claim one change | Duplicate reverse-map test |
| Story declares wrong path | Active/archive path validation test |
| Registry reintroduces bootstrap bypass | Registry-key rejection test |
| Manual lifecycle claim returns | Manual `status` rejection test |
| Archive transition leaves stale tracker status | Derived archive-state test and generated drift check |
| TossOS project rules disappear during alignment | Preservation checklist in design, tasks, and final diff review |

## Scope decision

In scope: PM hierarchy, Story mappings, PM validator/tests/generated views,
`docs/WORKFLOW.md`, and this OpenSpec contract.

Not in scope: production trading code, operator runtime toggles, broker behavior,
Guardian values, journal writes, VPN exposure, containers, and retroactive Stories
for already archived historical changes.

There are no unresolved taste decisions or user challenges. The user's stated
premises already settle the two important choices: StockOS is the methodology
reference, and TossOS-specific project behavior must remain unchanged.

## RED/GREEN evidence

- RED: `python3 -m unittest test_generate_master_tracker.py` produced 5 assertion
  failures and 1 error against the old flat `change_id`/manual `status`/allowlist
  implementation. The failures covered current-repository coverage, duplicate mapping,
  invalid path, manual status, stale rendering, and derived lifecycle.
- GREEN: the same command runs 9 tests successfully after the PM migration and
  validator update.
- Coverage audit: 33 active OpenSpec changes, 33 active Story mappings, 0 missing,
  0 extra, 0 duplicate.
- Function Logic Map: `not-applicable`. The implementation surface is the Python PM
  validator and Markdown/JSON portfolio data; TossOS Function Logic Map tooling is a
  Go AST extractor. Branch behavior is pinned by the 9 PM unit tests instead.

## Post-implementation review

Date: 2026-07-31
Verdict: **IMPLEMENTATION APPROVED; LANDING BLOCKED**

The final review checked the PM source registry and every generated view, active and
archive path resolution, Story lifecycle transitions, the WORKFLOW preservation list,
strict OpenSpec validation, and whitespace/diff integrity.

The generic Python quality analyzer reported the pre-existing monolithic shape of
`validate` and many false-positive “magic numbers” from PM IDs in tests. No security,
resource, exception, concurrency, or correctness finding was produced. A validator
refactor is not required to enforce the contract and would widen this governance
change; the explicit helper functions and branch tests are retained as the smaller
implementation.

## SDD fingerprint investigation

Symptom: an early `make sdd-check` reported a stale CodeGraph fingerprint immediately
after `make sdd-sync`.

Root cause: the long-running sync command had yielded a live process session while
CodeGraphContext/GBrain work continued; the check was started before that session was
polled to completion, so it correctly read the previous fingerprint.

Resolution: no repository code change. The live sync sessions were polled to exit 0,
which recorded the current worktree fingerprint. A subsequent `make sdd-check`
confirmed the CodeGraph hard-evidence match. GBrain was owned by another live project
process and remained an advisory busy warning, exactly as `docs/WORKFLOW.md` specifies.

## Landing gate blocker

`make gate CHANGE=align-full-sdd-pm-contract` passed tasks-file, unchecked-task, and
review-file checks, then stopped at Function Logic Map discovery. The persisted base
commit predates unrelated uncommitted Go work already present in the shared worktree,
so the gate attributed about 80 modified existing Go functions in console, engine,
config, exit-policy, and journal packages to this documentation/PM change.

Those Go functions are not in this Story scope and many already belong to other active
OpenSpec changes. Stashing or committing the user's work, changing the base to conceal
it, or copying another change's Function Logic Maps into this change would violate the
single-writer and evidence-ownership rules. Task 5.3 therefore remains unchecked.
This change must be gated in a clean worktree containing only its own diff, or after
the owning changes have landed and this change's base has been legitimately rebased.
Until then it must remain active and must not be archived or reported as Full SDD
complete.

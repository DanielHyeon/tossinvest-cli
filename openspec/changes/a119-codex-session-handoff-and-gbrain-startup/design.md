## Context

This change previously contained only a proposal. On 2026-09-06 the user
authorized repairing its repository-wide validation blocker while a063 is held.
This continuation completes the proposal's two declared specification deltas;
it does not implement or activate session capture or MCP configuration changes.

## Goals / Non-Goals

- Define testable host-event coverage and one effective Codex GBrain registration.
- Preserve Codex/Claude storage isolation and the project wrapper's existing lock.
- Keep document validity separate from implementation and runtime acceptance.
- No host configuration, session store, running process, or trading mutation is
  authorized by this documentation repair.

## Decisions

Accepted additional event names must come from sanitized supported-host fixtures.
The proposal does not establish those names or prove the current host dispatches
the hook. Matcher tests and actual host delivery are distinct evidence.

Select one effective registration only after checking the host's configuration
loading behavior. Do not infer that every registration found on disk is loaded
by Codex or delete shared configuration based only on duplicate text.

The existing canonical isolation, redaction, atomic writer and lock contracts
remain authoritative. Completing these documents does not change those contracts.

## Risks / Trade-offs

Host event delivery and configuration precedence remain unverified. These facts
must be established before an implementation plan chooses event names or an
authoritative registration location. No historical baseline is invented for
unperformed implementation work.

## Migration Plan

This continuation has no runtime migration. Future implementation must capture
its baseline, establish hard evidence and pre-edit maps, obtain proposal review,
then implement and test through a separate teammate. Archive requires all tasks
and real acceptance evidence, not merely strict document validation.

## Evidence-backed implementation plan (2026-09-25, task 2.3 — NOT FROZEN)

Status: the proposal-freeze review of 2026-09-25 rejected this plan (review.md). It stays a draft
until the scope decision in issues.md I-1 is made. Evidence: `analysis/host-evidence.md` (sanitized).

"Supported host" in this change means the hosts actually measured: Codex CLI `exec` 0.154.0 for
hook delivery (inferred), and the Codex Desktop bundled CLI 0.155.0-alpha.16.3 plus CLI 0.154.0 for
configuration loading. The interactive VS Code/Desktop host where the stale handoff was reported was
not observed.

- PostToolUse names: `Bash` has delivery inferred on `codex exec` only, and `apply_patch` exists only
  as a binary constant. No additional name is established. The matcher would stay unchanged. The claim
  that a matcher edit invalidates the stored hook trust is an inference, not a measurement.
- GBrain: the CLI configuration loader reports one `gbrain` wrapper registration for the workspace.
  Its source layer, `.codex/config.toml`, is inferred by elimination, and app-server capability
  discovery of `.mcp.json` is unverified. The per-thread cause of the double start is an inference.

Draft regressions (kept on branch `wip/a119-3.1` until the freeze is granted):

1. `tools/sdd-history/fixtures/codex_post_tool_use_names.json`: names with honest evidence kinds
   (`delivery-inferred`, `binary-constant`). The joint fixture×matcher test is required by the spec
   delta. Today it only restates the existing test at `test_codex_session_save.py:280-305`, and it
   starts to bite when a name is added.
2. `tools/sdd-history/test_codex_host_event_coverage.py`:
   - matcher anchoring;
   - a fixture schema with no free text, paths or session ids;
   - saver stdout stays empty on both the success and the failure path. This is new: a PostToolUse
     hook's stdout is read by the host as a decision.
3. `tools/sdd/fixtures/codex_gbrain_registration.json` and `tools/sdd/test_codex_gbrain_registration.py`:
   - exactly one enabled wrapper entry;
   - no Codex server launches raw `gbrain` (flock bypass);
   - a duplicate is counted.

   The count against the fixture is a drift pin, not host verification.

RED is defined by mutating copies of the real files
(`analysis/harness/mutate_codex_config.py`; the no-mutation control must stay green):
- drop `apply_patch`;
- unanchor the matcher;
- add a second wrapper;
- replace the wrapper with a raw `gbrain`;
- make the saver print to stdout.

Task 3.2 evidence map:

| Requirement | Evidence |
| --- | --- |
| Lock-owner preservation | existing `tools/sdd/test_gbrain_project.py:62,93,118,154` |
| Isolation, redaction and atomic persistence | existing `tools/sdd-history/test_codex_session_save.py:76,134,217,226,256` |
| Unchanged tool result | the new stdout tests |
| Claude-owned and trading files unchanged | `git diff --stat 54004f44 -- .mcp.json .claude save-session.sh tools/sdd/gbrain_project.py .codex/hooks.json .codex/config.toml` must be empty |

Runtime delivery and single startup stay pending (task 3.3). One observation route needs a human to
approve it: a temporary probe hook that records only `tool_name` under `.codex-context/`.

# a119 — Codex session handoff and GBrain startup: pin what was measured

> **Scope decision (2026-09-25, user: option (a)).** The 2026-09-25 proposal-freeze review rejected the
> plan because it no longer matched this proposal's original Why. The user chose to rewrite the proposal
> to "evidence and regression pins only" and hand both symptoms to named follow-ups. The 2026-08-29 text
> is superseded; its two cause hypotheses are recorded in `analysis/host-evidence.md` as neither
> established nor refuted.

## Why

The 2026-08-29 proposal named two causes for a stale Codex handoff and a doubled GBrain start: a
PostToolUse matcher that misses the host's event names, and a duplicate MCP registration loaded
twice. The 2026-09-25 evidence (`analysis/host-evidence.md`, sanitized) establishes neither and
refutes neither:

- Delivery of `Bash` is inferred only on `codex exec` 0.154.0. The interactive VS Code/Desktop host
  where the stale handoff was reported was not observed, and the 2026-08-29 session with ten
  `FileChange` events and no refresh remains unexplained. No additional event name is established by
  a sanitized fixture.
- The CLI configuration loader reports one `gbrain` wrapper registration for this workspace. The
  observed double start is one MCP launch per concurrent Codex thread; the later one exits 75 by the
  wrapper's single-writer contract.

Editing configuration on an unestablished cause risks the one thing that works today: the hook's
stored trust hash (an inference, not a measurement — still a reason not to touch it). This change
therefore pins the measured facts as regressions and changes no configuration.

## What Changes

- Add regression tests that pin today's facts:
  1. every sanitized fixture event name is matched by the configured PostToolUse matcher, and the
     fixture and matcher are tested together;
  2. the saver writes nothing to stdout on any of its three exits — success, exception warning and
     lock contention. This is a precaution: a PostToolUse hook's stdout may be read by the host as a
     decision, which has not been observed for Codex `async` hooks;
  3. the Codex-effective configuration has exactly one enabled TossOS GBrain wrapper registration
     and no raw `gbrain serve`.
  Landed in task 3.1 (`c202b804`) from branch `wip/a119-3.1`, and hardened after the task 3.4 review
  (12 tests; mutation harness M1–M8 all caught; control green and non-empty).
- Change no configuration file. `.codex/hooks.json`, `.codex/config.toml`, `.mcp.json`,
  `save-session.sh` and Claude-owned stores stay byte-identical to the implementation baseline
  `54004f44`.
- Record what stays unobserved in this change's evidence, and name the follow-ups below.

## Follow-ups (named here; numbered when opened)

- **Interactive-host handoff refresh.** Observe PostToolUse delivery on the interactive host with a
  human-approved probe hook that records only `tool_name` under `.codex-context/`; establish or
  refute the matcher cause (issues I-1).
- **Per-thread MCP startup warning.** `MCP startup incomplete (failed: gbrain)` on a later concurrent
  thread is the wrapper's busy exit. Removing it needs a decision the canonical spec forbids here:
  (a) change the wrapper's busy behavior, (b) disable the Codex registration, (c) an HTTP broker or
  shared backend (issues I-1). "Exactly once" in the 2026-08-29 text is read as an observation,
  "at most once per thread", not a promise of this change. It also carries the observations left open
  here: how many times the wrapper starts per interactive-host thread, and whether the workspace
  `.mcp.json` is loaded through executor capability discovery (`analysis/host-evidence.md` §4 U5, U6).
- **SDD agent-save handler on Codex.** Its matcher `Write|Edit|MultiEdit|NotebookEdit` never matches
  the names Codex is known to emit (issues I-2).

## Capabilities

### New Capabilities

- `gbrain-codex-mcp-startup`: the Codex-effective configuration has exactly one enabled TossOS GBrain
  wrapper registration, pinned by a static regression; the wrapper's lock and busy behavior are
  preserved.

### Modified Capabilities

- `codex-session-save`: the fixture of accepted event names and the PostToolUse matcher are tested
  together; today no name beyond `Bash` and `apply_patch` is established.

## Non-goals

- Adding an event name not established by a sanitized fixture.
- Deleting or editing either GBrain registration, or the wrapper's busy behavior or lock.
- Claiming that an ordinary host tool call refreshed the handoff, or that the workspace starts the
  wrapper once across threads.

## Impact

- `tools/sdd-history/test_codex_host_event_coverage.py` and its fixture,
  `tools/sdd/test_codex_gbrain_registration.py` and its fixture, `analysis/harness/mutate_codex_config.py`.
- No trading runtime, account state, order path, Claude hook, or `save-session.sh` change.

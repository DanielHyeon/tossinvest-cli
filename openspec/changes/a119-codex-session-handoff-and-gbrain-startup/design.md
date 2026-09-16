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

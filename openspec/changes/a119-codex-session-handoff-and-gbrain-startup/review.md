# Independent adversarial review — a119-codex-session-handoff-and-gbrain-startup

- Date: 2026-09-06
- Reviewer: separate adversarial review context
- Stage: specification-completeness repair only; no implementation or runtime acceptance
- Scope reviewed: original proposal, added design/tasks, both proposed spec deltas, canonical
  `codex-session-save` and `sdd-workflow` specifications, and the current repository hook,
  MCP, saver, and GBrain-wrapper contracts.

## Original blocker and disposition

The original a119 commit (`80d3931d`) contained a proposal but no `specs/` deltas; its
commit record identifies that absence as the reason strict/all-change OpenSpec validation
failed. The added deltas remove that document-completeness failure. This review does not
reinterpret the original operational report as new host or runtime evidence.

## Adversarial findings

1. **Accepted — modified block is complete.** The `codex-session-save` delta modifies the
   entire existing `Codex PostToolUse session capture` requirement: it retains `Bash`,
   `apply_patch`, coexistence with the SDD saver, asynchronous execution, and unchanged tool
   results; it retains both existing scenarios and adds bounded additional-event and
   unobserved-host scenarios. The separate canonical requirements for storage isolation,
   bounded/redacted handoff, atomic failure-safe persistence, and Claude preservation remain
   authoritative and are explicitly preserved by the modified requirement. No canonical
   requirement is weakened or silently replaced.
2. **Accepted — no unsupported host claim in the new deltas.** Additional event names require
   sanitized supported-host fixtures; matcher coverage alone is explicitly insufficient to
   claim event delivery or a refreshed handoff. The inherited proposal's historical symptom is
   not treated as proof of current host dispatch.
3. **Accepted — registration contract preserves the safety boundary.** The new capability
   distinguishes static repository configuration from actual Codex loading, requires one
   *effective* Codex-owned wrapper registration, preserves other agents' registrations, and
   forbids lock-owner/process/database deletion. It retains the canonical wrapper's project
   home, singleton lock, and exit-75 busy behavior.
4. **Required before implementation — do not choose matcher names or remove either current
   configuration entry from text inspection alone.** Current `.codex/hooks.json` matches only
   `Bash|apply_patch`; both `.codex/config.toml` and `.mcp.json` contain a GBrain wrapper
   registration. The required supported-host fixture and configuration-loading evidence are
   still absent, so neither observation proves which entry the host loads. Task 2.2 remains
   open.
5. **Required before any completion claim — implementation and runtime proof remain open.**
   No source/configuration/test implementation was reviewed or changed in this pass; there is
   no observed host event delivery, single host startup observation, implementation baseline,
   proposal-freeze review, teammate implementation, SDD/gstack/final-gate evidence, or Manager
   acceptance. This review does not mark a119 complete and does not authorize archive, PM
   synchronization, a runtime mutation, or a next `a0xx` change.

## Verification actually run

| Command | Actual result |
| --- | --- |
| `openspec validate a119-codex-session-handoff-and-gbrain-startup --strict --no-interactive` | exit 0 — valid |
| `openspec validate --all --strict --no-interactive` | exit 0 — 60 passed, 0 failed |

`make validate --all` is not an all-change OpenSpec command in this repository; GNU make
rejects that option (exit 2). It is not presented as validation evidence.

## Task ownership and status

Tasks 1.1 and 1.2 are Manager-owned status decisions. This reviewer neither changes their
checkboxes nor treats the successful document validation as implementation completion. All
code-feature tasks (2.1 through 3.4) remain unchecked and require the evidence and separate
implementation/review sequence stated in the change.

## Verdict

The added OpenSpec deltas truthfully repair the original missing-delta validation blocker and
preserve the canonical isolation and lock contracts. They are clear for this specification-only
stage. Host/runtime assertions and all implementation acceptance remain explicitly pending.

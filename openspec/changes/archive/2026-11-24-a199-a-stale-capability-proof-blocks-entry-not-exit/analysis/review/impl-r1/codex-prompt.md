You are an adversarial reviewer of a Go change in a real-money automated trading engine. Read-only: do not modify any file.

Inputs (read them):
- The implementation diff: ../a199.diff (relative to this directory: /tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/6ba8fcd4-b024-437e-a4c2-ebe621826fe5/scratchpad/rev-a199/a199.diff)
- The design: openspec/changes/a199-a-stale-capability-proof-blocks-entry-not-exit/design.md
- Source tree at the current directory (HEAD of the change). Key files: internal/app/engine/exit_only_start.go, interlock.go, engine.go, attestation_watch.go, exitwiring.go, reconcileloop.go, strategy_entry_supervisor.go, internal/execgw/retry.go, gateway.go, symbolgate.go, strategy_entry_gate_authority.go, internal/attest/attest.go, account_match.go, cmd/tossctl/engine.go.

Intent: when the automation gate is on and the capability attestation is only STALE (expired, or missing required endpoints) and every other interlock clause holds, the engine starts "exit-only": Verified=false, ExitOnly=true, Guardian kept, an EntryGate latch capability_attestation_stale blocks exposure-raising mutations, and exit/reconcile loops run. Any other attestation defect or any other failed clause still refuses startup.

Answer these four, each with file:line evidence from the source (not from the diff alone). If you find nothing, say what you checked.
1. Can an exposure-raising mutation (buy, or an amend raising quantity/price) reach the broker from an exit-only engine? Look for submit paths that do not read the latch, a window before the latch is set where a loop or auxiliary executor could run, code that clears this latch (Clear, ClearSymbolReason, RebuildReconcileProjection, operating-mode projection, alert ack), the gate being nil or a different object from the gateway's, and consumers that open entry without reading Verified (note Guardian is now non-nil in exit-only).
2. Which attestation inputs reach the exit-only verdict that should be refused (e.g. expired + different account, masked account, blank endpoint items, format_version 0/negative/too new, zero expiry, issued after expiry), and which normal issuer-shaped files are refused but should go exit-only?
3. Combination with other clauses (limits, Guardian limit source, daily buy 7b, open positions 7c, gateway wiring, keyless transport, trading policy, no Guardian): is each still refused with only that clause's sentinel? Does verifyGate's order now skip anything or change messages/audit for non-stale inputs?
4. Restart and races: stale restart, repaired restart, attestation file rewritten between interlock read and watch read, and what happens when the exit-only audit record fails.

Output: a numbered list of findings, each with severity (P0 = entry opens / stop-loss blocked / wrong account; P1 = contract violation; P2 = other), file:line, and a one-line reproduction idea. Keep it under 60 lines.

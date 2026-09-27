You are an independent ADVERSARIAL senior engineer doing the FIRST IMPLEMENTATION review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. The Go tree here is HEAD of the working branch WITH the change's implementation
applied (the working-tree versions of the files below). The exact implementation diff is also stored as
openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/analysis/impl-review/a124-impl.patch.

Contract (frozen after 14 proposal-freeze rounds; read ALL of it first):
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/design.md (revision 8: D1 judgement tables, D2 limit 3,
  D3 missing publisher counts, D4 selection, D6, D7 principle E + judgement→action table + executor ordering, D8
  record-failure counter transitions (AA2 table), D9 sanitization, D10 enforcement boundary)
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/specs/engine-safety/spec.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/tasks.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/review.md §1 (implementation lot record, measurements,
  two recorded deviations) and analysis/harness/mutation-ledger.tsv (30 mutants, all CAUGHT)

Implementation files: internal/app/engine/alertdelivery.go, internal/app/engine/auxiliary.go,
internal/journal/alert_claim.go, internal/journal/outbox.go, internal/execgw/retry.go,
internal/execgw/a124_entry_gate_testseam.go (build tag), and tests internal/app/engine/a124_*_test.go,
internal/journal/a124_*_test.go, internal/execgw/a124_*_test.go. Relevant unchanged code: internal/obs/notifier.go
(Acknowledge, deliver, notifyCritical, escalate), internal/journal/operating_mode.go, internal/execgw/modegate.go,
internal/execgw/symbolgate.go, internal/app/engine/gateway.go (restoreAlertEntryLatch), cmd/tossctl/engine.go.

Safety invariants that override everything: no weakening/delaying of stop-loss or emergency flatten; the executor holds
no lock across a remote publish or a ledger transaction, takes only the entry gate lock and does only map operations
inside it, and never takes the Notifier mutex; toggle OFF equals prior behaviour; only conservative direction changes to
entry blocking; operating-mode relaxation needs a human; no secrets/account refs/raw errors/row content/lease tokens in
new log lines or gate details.

Produce (concise, evidence with file:line):
1. Conformance: for every row of design D1 (both tables), the D7 judgement→action table and ordering, the D8 AA2 transition
   table and pruning rules, D3, D4 and D9 — CONFORMS / DEVIATES / MISSING, one line why. Judge the two deviations recorded in
   review.md §1 (delivery-settle error under a cancelled context is not judged; empty AccountRef/nil Gate skip) — are they
   conservative?
2. New findings table: id | severity (P0 = unsafe or wrong in production, P1 = must fix before landing, P2 = should record,
   P3 = editorial) | finding | evidence | suggested fix. Attack in particular: fail-open paths (an error/outcome that yields
   neither latch nor count), interleavings where a clear after the evidence leaves a latch or a clear before it drops one
   (beyond the permitted conservative window), races, the same-transaction attempts read and its rollback, test vacuity
   (hooks that never fire, structural scans over the wrong root, measurements whose margins cannot fail), and whether the
   mutation ledger's mutants are real (not build failures, not unreachable).
3. Whether the tests cover what tasks 2.1–2.14 promise.
4. Final line: `VERDICT: PASS` (no P0/P1 left) or `VERDICT: REJECT`, with one sentence why.

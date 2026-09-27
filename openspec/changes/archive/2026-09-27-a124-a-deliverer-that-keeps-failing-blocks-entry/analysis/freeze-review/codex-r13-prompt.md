You are an independent ADVERSARIAL senior engineer doing the THIRTEENTH proposal-freeze review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. Base commit: 4798d3992c95f8dcd0fdf7faf812203910bbf9e3 (the Go code in this tree is
identical to it; only the change directory below differs).

Change under review (read ALL of it first; the design is revision 7 (7판)):
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/proposal.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/design.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/tasks.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/specs/engine-safety/spec.md
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/analysis/function-logic/*/function-logic-map.md
  (AST-derived branch maps; branch IDs such as B8/B9/B10 refer to these)
- openspec/changes/a124-a-deliverer-that-keeps-failing-blocks-entry/review.md sections §0 … §0.15 (findings F1–F13, N1–N8, R1–R8, Q1–Q7,
  V1–V6, W1–W5, X1–X5, Y1–Y6, Z1–Z5, AA1–AA4, AB1–AB2, AC1–AC2, and the decisions M1 = B′ (per-reason clear epoch on EntryGate; the
  executor never takes the Notifier mutex), M2 = (ii), principle E, and the §0.15 manager verdict on AC1). Treat all earlier findings,
  dispositions and verdicts as claims to verify, not as truth.

Revision 7 changes no judgement rule. It responds to round-12 finding AC1 (production never binds the operating-mode projector
nor restores it at startup, so a durable ENTRY_BLOCKED row blocks nothing) by: a new design section D10 "enforcement boundary"
(what actually blocks entry today versus only after projector wiring), restating the W1 / Y1 / cases ㉩ ㉪ arguments on that
boundary, naming projector wiring + startup restoration as an OPERATIONAL-EFFECT precondition owned by a092 (explicitly not a
code-landing condition of a124), recording AC2 as a latent pre-existing defect, and adding tasks 2.14 (boundary pins).

Verify every claim against the code with file:line. Relevant code: internal/app/engine/alertdelivery.go,
internal/app/engine/auxiliary.go, internal/app/engine/alertops.go, internal/app/engine/gateway.go,
internal/app/engine/exitwiring.go, internal/app/engine/tracer.go, internal/journal/outbox.go, internal/journal/alert_claim.go,
internal/journal/operating_mode.go, internal/journal/journal.go, internal/execgw/retry.go, internal/execgw/modegate.go,
internal/execgw/gateway.go, internal/execgw/replay.go, internal/obs/notifier.go, internal/obs/alert_lease.go,
internal/risk/chain.go, openspec/specs/engine-safety/spec.md,
openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md,
openspec/changes/a092-an-alert-does-not-hold-the-stop/proposal.md.

Safety invariants that override everything: no weakening/delaying of stop-loss or emergency flatten; toggle OFF must
equal prior behaviour; only conservative direction changes to entry blocking; operating-mode relaxation needs a human;
no secrets/account data in logs or gate details; no lock may be held across a remote publish. A safety argument must be
true of the shipped production wiring, not only of a test fixture.

Produce (concise, evidence with file:line):
1. For each earlier finding F1–F13, N1–N8, R1–R8, Q1–Q7, V1–V6, W1–W5, X1–X5, Y1–Y6, Z1–Z5, AA1–AA4, AB1–AB2 and AC1–AC2:
   RESOLVED / PARTIAL / NOT RESOLVED, one line why. Judge W1 in two parts as §0.15 does: (a) "a124's delayed application makes
   a judgement be lost" (equivalence leg) and (b) "row B's entry protection holds without B's acknowledgement" (protection leg);
   say whether revision 7 states (b) honestly as unresolved and whether its closers (a092 count–clear fix, or projector wiring)
   are correct and sufficient.
2. New findings table: id | severity (P0 = unsafe or unimplementable as written, P1 = must fix before freeze, P2 = should
   record, P3 = editorial) | finding | evidence | suggested fix. Attack in particular: D10 (is the enforcement table true —
   every place that can block entry today, every production reader of operating_modes, the startup restore conditions; is
   anything still claimed to be protected by the mode row anywhere in proposal/design/spec/tasks), whether any remaining
   sentence relies on mode enforcement, whether the spec delta's new SHALL/SHALL NOT paragraph is testable and consistent
   with its scenarios, whether tasks 2.14 pins the boundary without fabricating enforcement in fixtures, whether naming the
   precondition "operational effect, not code landing" is sound (can a124 land and be safer than HEAD without it?), and
   whether AC2's coordinates and manifestation condition are right. Re-check that revision 7 changed no judgement rule
   (D1 table, D7 judgement→action table and ordering, D8 transitions) versus revision 6 fifth printing.
3. Whether design D2 (limit 3) and D3 (count missing publisher) remain adequately justified.
4. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.

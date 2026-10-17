You are an independent ADVERSARIAL senior engineer doing the THIRD proposal-freeze review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. Base commit: 3937e341f31dfc643da563dbe5dcdbba53e8121d (working tree Go code is
identical to it; only the change directory below differs from that commit).

Change under review — `openspec/changes/a094-a-stop-clears-what-blocks-it/`. The documents are the THIRD draft (3판).
Their `file.go:line` citations were written against an older base (ec29dc72); the current positions are listed in the
errata. Read ALL of these first:
- proposal.md, design.md, tasks.md, issues.md
- specs/order-execution/spec.md, specs/exit-policy/spec.md
- analysis/function-logic/*/ (15 bundles: ast.json + function-logic-map.md + branch-test-map.md; refreshed to the current
  source — branch IDs like B5/B8/B10 refer to these ast.json files, NOT to the older numbers in the prose)
- analysis/third-round-errata.md (citation errata: 91 citations, 32 moved, 2 content changed; `record` renumbered
  B1..B14 → B3..B16; neighbour changes; a094 R1 vs a089 R2 sentence comparison)
- analysis/third-round-review-materials.md (what this round receives, incl. section G on a124's finding AC1)
- review.md: §1 (round 1, 8 blockers), §2 (round 2, 8 blockers) and §2.11 "3라운드가 받는 것", and the final section
  「증거 재생성」. Treat every earlier finding and every claim in the draft as a claim to verify, not as truth.

Relevant code (verify against it with file:line): internal/app/engine/exitloop.go (record, submit, clearTheSymbol,
release), internal/execgw/classify.go, internal/execgw/failclosed.go, internal/execgw/gateway.go (checkSymbolFree),
internal/execgw/replay.go, internal/execgw/indoubt.go, internal/journal/dispatch.go, internal/journal/apply_hook.go
(armExitProposalTx, ResolveExitProposal), internal/journal/fills.go, internal/journal/recovery.go,
internal/reconcile/recovery.go (Recovery.Run), internal/filldetect/detect.go, internal/brokerstate/derive.go,
internal/obs/notifier.go, internal/journal/operating_mode.go, cmd/tossctl/engine.go,
openspec/specs/order-execution/spec.md, openspec/specs/exit-policy/spec.md,
openspec/changes/a089-an-unserved-stop-is-counted/specs/engine-safety/spec.md and its design.md §D4.

Safety invariants that override everything: no weakening/delaying of stop-loss or emergency flatten; toggle OFF must equal
prior behaviour; only conservative-direction changes to stop/take-profit/sizing; no order may be cancelled or placed that
this engine cannot attribute; operating-mode relaxation needs a human; no secrets/account data in logs.

Produce (concise, evidence with file:line against the CURRENT tree):
1. For each of round 2's eight blockers in review.md §2.11 (lock re-identification / B8 release; R2 self-direction
   absence check; spec contradiction pair; a087 ordering and floatOf; R3 Context.Resolver and resolveCancel r.Order;
   R2↔R3 interaction via park; §0.4 call frequency and detector OPEN-snapshot reuse; coordinates + FLM/AST regeneration)
   and the "split recommendation" (R3 as a separate change): RESOLVED / PARTIAL / NOT RESOLVED, one line why, pointing at
   where the third draft answers it.
2. New findings table: id | severity (P0 = unsafe or unimplementable as written, P1 = must fix before freeze, P2 = should
   record, P3 = editorial) | finding | evidence | suggested fix. Attack in particular:
   - R1 `classifyRefusalCode`: reading top-level `code` and `error.code`; placing it before `classifyRefusalBody`; whether
     `opposite-pending-order-exists` really means "not accepted" in every case; the replay boundary SHALL NOT; the boot-time
     retro-reclassification of stored IN_DOUBT attempts (can it terminate an attempt whose order exists?).
   - The a094 R1 SHALL (classify by code → terminate) versus a089 R2 SHALL NOT (never branch behaviour on the broker
     reason code) — errata §5. Is this a real contradiction in the main spec if both land? Which clause should yield?
   - R2 (clearTheSymbol sees the broker's non-journal open orders via the detector snapshot): can it cancel an order this
     engine cannot attribute, a take-profit, or another position's order? PENDING_CANCEL exclusion. Staleness of the
     snapshot. The a111 early return now in front of `record` (errata §4).
   - R3 (a terminated attempt releases the armed proposal; "park also releases"): can a release allow a second sell
     while an order may still be live (submit B8's safety argument)? Interaction with UNRESOLVED_IN_DOUBT.
   - The two changed citations (errata §2/§4: restart recovery now runs via `engineRecoverySequence`,
     cmd/tossctl/engine.go:604-606) — does tasks 4.4 still pin anything real?
   - AC1 (materials §G): `SetModeProjector` / `RestoreOperatingModeProjection` have zero non-test callers, so a durable
     ENTRY_BLOCKED is not projected back after restart. The draft cites "critical delivery failure goes to ENTRY_BLOCKED"
     (design.md:441, tasks.md:169, specs/exit-policy/spec.md:79) as the cost that justifies excluding PENDING_CANCEL.
     Is that premise true, partially true (in-process latch only), or false, and does the conclusion survive?
   - Whether the spec deltas and tasks cover what the design promises, and whether the order-execution MODIFIED delta
     still reproduces the canonical requirement text of openspec/specs/order-execution/spec.md verbatim.
3. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.

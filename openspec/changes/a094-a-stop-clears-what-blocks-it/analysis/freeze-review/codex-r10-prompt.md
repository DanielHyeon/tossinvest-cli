You are an independent ADVERSARIAL senior engineer doing the TENTH (NARROW CONFIRMATION) proposal-freeze review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. It is an export of commit bc53211f (all code and documents are exactly that
commit). Every Go file the change cites is byte-identical to the change's base `3937e341` except four (unchanged since the round-7 tree 9fa0bb90; none of the cited files changed up to this commit), all additive and outside this
change: `internal/journal/schema.go` (a066 schema v34/v35 entries, +7 lines at :165-171 — schema.go lines the change cites after :165 are
+7 here), `internal/journal/outbox.go` (a124 `PendingAlertsForDelivery`, +22 lines at :512-533 — lines after :512 are +22; `EnqueueAlert`
at :131 is unchanged), `cmd/tossctl/engine.go` (+2 lines at :121-122 — lines after :121 are +2), and `internal/execgw/retry.go` (a124
`clearEpochs`, `ClearEpoch`, `BlockUnlessClearedSince`, +39 lines from :475 — the 7th/8th draft cite these new symbols at their
positions in this tree).

Change under review — `openspec/changes/a094-a-stop-clears-what-blocks-it/`. The documents are the ELEVENTH draft (11판, bc53211f).
**design.md §D−9 overrides §D−8, which overrides §D−7 (D−7.2 retracted), and so on.**

THIS IS A NARROW CONFIRMATION ROUND, scoped to: (1) the new named critical for "terminal evidence pending while detection is healthy" (D−9.3) and the
named follow-up candidate (D−9.4); (2) the corrected proof wording (D−9.1) and the conditional alert table (D−9.2). Do not re-open settled items.
Read: design.md §D−9 (D−9.1 … D−9.4) and the 11판 banners in §D−8; specs/exit-policy/spec.md (the new sentence and scenario); tasks.md (0.5t–0.5u,
3.R9, 3.R9a); review.md 「9라운드」 and 「11판」; analysis/freeze-review/codex-r9-output.md. Cross-reference only if needed:
openspec/changes/a090-an-unobserved-position-is-counted/design.md D1 (the unobserved-position alert D−9.3 relies on for unjudged positions).

Relevant code: internal/app/engine/exitloop.go (judge, clearTheSymbol, noteDelay, DefaultExitLiquidationDelayBound), internal/obs/event.go
(EventExitLiquidationDelayed), internal/journal/schema.go (mutation_attempts.settled_at), internal/journal/fills.go (RecordFill refusal path,
TrackedFillOrders), internal/filldetect/detect.go (cycle success path, blockSymbol), internal/filldetect/ledger.go, internal/brokerstate/derive.go,
internal/app/engine/runtime.go, cmd/tossctl (to confirm no UNKNOWN_BROKER_STATE hand-resolution command exists).

Produce (concise, evidence with file:line against this tree):
1. For each round-9 finding R9-1 … R9-3: RESOLVED / PARTIAL / NOT RESOLVED, one line why.
2. Attack D−9.3: is the alert condition implementable from the journal alone; is the episode key finite and stable across restart; does it fire in
   the healthy-detection/non-terminal-evidence case the round-9 finding described (including the refused CLOSED example); is the union with
   a090's unobserved alert a real cover for unjudged positions; can it spam, double-send, or be silently absorbed? Is the wall-clock residual
   honestly bounded?
3. Is D−9.1's corrected reasoning now accurate, and does it still justify keeping the retraction?
4. New findings from the 11th draft only (table: id | severity P0/P1/P2/P3 | finding | evidence | fix).
5. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.

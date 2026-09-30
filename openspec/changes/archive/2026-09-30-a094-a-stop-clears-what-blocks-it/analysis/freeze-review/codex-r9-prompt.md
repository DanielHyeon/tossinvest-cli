You are an independent ADVERSARIAL senior engineer doing the NINTH (NARROW CONFIRMATION) proposal-freeze review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. It is an export of commit e25d9b36 (all code and documents are exactly that
commit). Every Go file the change cites is byte-identical to the change's base `3937e341` except four (unchanged since the round-7 tree 9fa0bb90; none of the cited files changed up to this commit), all additive and outside this
change: `internal/journal/schema.go` (a066 schema v34/v35 entries, +7 lines at :165-171 — schema.go lines the change cites after :165 are
+7 here), `internal/journal/outbox.go` (a124 `PendingAlertsForDelivery`, +22 lines at :512-533 — lines after :512 are +22; `EnqueueAlert`
at :131 is unchanged), `cmd/tossctl/engine.go` (+2 lines at :121-122 — lines after :121 are +2), and `internal/execgw/retry.go` (a124
`clearEpochs`, `ClearEpoch`, `BlockUnlessClearedSince`, +39 lines from :475 — the 7th/8th draft cite these new symbols at their
positions in this tree).

Change under review — `openspec/changes/a094-a-stop-clears-what-blocks-it/`. The documents are the TENTH draft (10판, e25d9b36).
**design.md §D−8 overrides §D−7 (D−7.2 is retracted), which overrides §D−6, and so on.**

THIS IS A NARROW CONFIRMATION ROUND, scoped to exactly two things: (1) the retraction of the operator terminal-assertion (D−8.1) and everything it
removed, and (2) the invariant-4 argument (D−8.2) with its human path (D−8.3). Also confirm R8-4/R8-5 (D−8.4). Do not re-open other settled items.
Read: design.md §D−8 (D−8.1 … D−8.4) and the D−7.2 retraction banner; specs/order-execution/spec.md and specs/exit-policy/spec.md (check the
assertion requirement and scenario are gone and the terminal-evidence sentence now reads "종결 체결 기록뿐"); tasks.md (0.5r–0.5s, 3.R7d–3.R7f,
the retracted 4.Tb/4.Tc/4.Td/8.1a lines, 8.2); review.md 「8라운드」 and 「10판」; analysis/freeze-review/codex-r8-output.md.

Relevant code: internal/filldetect/detect.go (PollOnce, collect, Health, poll mutex), internal/filldetect/hints.go (Refresh and its comment),
internal/filldetect/ledger.go (TrackedOrders), internal/journal/fills.go (TrackedFillOrders, RecordFill, LiveOrdersForSymbol),
internal/app/engine/exitloop.go (clearTheSymbol, noteDelay, submit quantity), internal/app/engine/exitwiring.go, internal/app/engine/runtime.go
(supervised loop degradation), cmd/tossctl/engine.go (supervised loops), internal/journal/backup.go.

Produce (concise, evidence with file:line against this tree):
1. For each round-8 finding R8-1 … R8-5: RESOLVED / MOOT (withdrawn with the assertion) / PARTIAL / NOT RESOLVED, one line why.
2. Verify the retraction argument itself: is it true that no single-order fill catch-up exists and that the full cycle is all-or-nothing? Is X
   (engine CONFIRMED sell, cancel CONFIRMED, no terminal snapshot) really a tracked order that a successful cycle would settle? Is there any
   case where the retracted assertion would have changed the outcome safely?
3. Attack D−8.2: is "the hold is enforcement of the invariant under uncertainty, not a violation" sound? Is the human path (repair fill
   detection) real — do the named alerts actually fire during the hold (delay alert, the D−2.7 counter, the filldetect degradation alert)?
   Is there any state where the hold is silent, or where detection is healthy but the hold still persists forever?
4. New findings from the 10th draft only (table: id | severity P0/P1/P2/P3 | finding | evidence | fix).
5. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.

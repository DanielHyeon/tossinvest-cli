You are an independent ADVERSARIAL senior engineer doing the EIGHTH (NARROW CONFIRMATION) proposal-freeze review of an OpenSpec change in a Go
repository that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file,
do NOT run git commands that write, do NOT run network calls, do NOT run the engine or any command that could place an
order. Reading files and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. It is an export of commit 0b5ea307 (all code and documents are exactly that
commit). Every Go file the change cites is byte-identical to the change's base `3937e341` except four (unchanged since the round-7 tree 9fa0bb90), all additive and outside this
change: `internal/journal/schema.go` (a066 schema v34/v35 entries, +7 lines at :165-171 — schema.go lines the change cites after :165 are
+7 here), `internal/journal/outbox.go` (a124 `PendingAlertsForDelivery`, +22 lines at :512-533 — lines after :512 are +22; `EnqueueAlert`
at :131 is unchanged), `cmd/tossctl/engine.go` (+2 lines at :121-122 — lines after :121 are +2), and `internal/execgw/retry.go` (a124
`clearEpochs`, `ClearEpoch`, `BlockUnlessClearedSince`, +39 lines from :475 — the 7th/8th draft cite these new symbols at their
positions in this tree).

Change under review — `openspec/changes/a094-a-stop-clears-what-blocks-it/`. The documents are the NINTH draft (9판, 60a8898d) plus the
Q9-1 approval record (0b5ea307). **design.md §D−7 overrides §D−6, which overrides §D−5, and so on.**

THIS IS A NARROW CONFIRMATION ROUND. Scope: only whether the round-7 findings are resolved by the 9th draft and whether the 9th-draft
changes themselves introduce a new P0/P1. Do not re-litigate decisions recorded as Manager decisions unless they create a safety defect.
Read: design.md §D−7 (D−7.1 … D−7.5) and the parts of §D−5/§D−6 it amends; specs/order-execution/spec.md (the new 「운영자 종결 단언」
requirement and scenario); specs/exit-policy/spec.md (terminal evidence sentence); tasks.md (0.5p–0.5q, 3.R4, 3.R7d, 4.N4f, 4.Tb–4.Td, 4.7,
8.1a); review.md 「7라운드」, 「9판」, 「Manager 판정 — Q9-1」; analysis/freeze-review/codex-r7-output.md; and the a092 text D−7.3 cites
(openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md:44 and design.md §D0.3h item 4).

Relevant code: internal/app/engine/exitloop.go (ObserveOnce, clearTheSymbol, noteDelay/delayAlerted, NewExitObserver/NewID),
internal/journal/fills.go (RecordFill, LiveOrdersForSymbol, terminal snapshots), internal/journal/apply_hook.go, internal/journal/outbox.go,
internal/journal/resolution.go, internal/journal/schema.go (migrations), internal/filldetect/detect.go, cmd/tossctl/engine_reconcile.go.

Produce (concise, evidence with file:line against this tree):
1. For each round-7 finding R7-1 … R7-8: RESOLVED / PARTIAL / NOT RESOLVED, one line why.
2. New findings introduced by the 9th draft only (table: id | severity P0/P1/P2/P3 | finding | evidence | fix). Attack in particular the
   operator terminal-assertion command (D−7.2): preconditions, audit-before-write, no synthesized fills, the assertion as terminal evidence
   for both the clearing verdict and proposal release — can it release a proposal while the original sell is actually live? can it be
   applied to the wrong order or a non-engine order? — and the additive schema/rollback note (8.1a).
3. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.

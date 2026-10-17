DECLARATION: I did not read or write anything under ~/.codex and made no file changes.

Reviewed `951b3ec3` statically. Below, `a127/` means `openspec/changes/a127-strategy-authorities-read-the-current-ledger/`. No P0 or P1 established.

1. **P2 — D6’s approval narrative omits necessary entry conditions.**  
   `a127/design.md:114–117`; `a127/proposal.md:58–59`.  
   The no-four-family-activation path exists, but schedule activation plus signed route/risk/account manifests is insufficient. `internal/app/engine/strategy_proposal_authority.go:323–350` additionally requires proposal signing configuration, manifest and evidence binding. `strategy_account_first_leg_authority.go:161–163` requires **exactly one valid proposal** without family activation; two proposals are rejected, not reduced to one handoff selection. Candidate/FX readiness and downstream protection, entry-gate and admission checks also remain.

   State these prerequisites explicitly before “곧 실진입 경로를 연다.” Human approval remains appropriate. Also distinguish worker promotion from order enforcement: `strategy_entry_supervisor.go:500–502` explicitly says dispatch does not consult the worker’s `Effective` descriptor; actual schedule/FX checks occur at `strategy_dispatch_cycle.go:86–88`.

2. **P2 — D7’s snapshot guarantee lacks a matching spec requirement and complete falsifier.**  
   `a127/design.md:124–127,147`; `a127/specs/strategy-runtime/spec.md:7–10`.  
   S13 checks only that usage reads receive a transaction. A mutant leaving `PRAGMA user_version` or the scope-latch read on `db`, then beginning a transaction for usage, satisfies that assertion while violating D7’s same-snapshot requirement.

   Specify that version, latch and all five dimension reads use **the same read-only transaction**; assert their receiver identity and transaction lifetime. The change is feasible: `UsageQueryer` accepts `*sql.Tx` (`internal/riskbucket/production_snapshot_authority.go:443–445`), already used by journal admission. It improves consistency without changing usage arithmetic or replacing admission revalidation.

3. **P2 — “Before reading” exceeds the risk design’s guarantee.**  
   `a127/specs/strategy-runtime/spec.md:34–35`; `a127/design.md:124–129`.  
   The delta requires missing-column failure **before reading**, but D7 only pre-prepares route queries. Risk still reads the scope-latch count before preparing the usage query (`internal/riskbucket/production_snapshot_authority.go:380–399`); a missing usage column therefore surfaces after a ledger read. A nonzero latch can return even earlier with the scope-refusal identity.

   Either define the guarantee as rejection **before authority creation**, with explicit error precedence, or require risk read-set validation before those reads. The route conditional-query clause itself now matches D7 and is testable.

4. **P3 — Round-1 artifact cleanup remains incomplete.**  
   `a127/design.md:160` still contains `</content>`. The embedded command and duplicate proposal specification were removed, but F1 should not claim complete cleanup until this remaining tag is removed.

Round-1 closure otherwise holds at the proposal level:

- **Conditional campaign query:** D7’s unconditional preparation of both shared SQL constants closes the original P1. The pinned SQLite driver prepares immediately (`modernc.org/sqlite@v1.54.0/stmt.go:30–36`).
- **Route diagnostics:** D3 explicitly changes the error-erasing wrapper at `internal/strategyrouter/production.go:350–352`, preserving sentinel identity. The single loader forwards Batch errors unchanged at `:303–304`.
- **Falsification:** S1–S12 have distinguishing planned observations. S6 requires otherwise-valid inputs so the nonexistent ledger path isolates guard ordering. S10 correctly needs structural checking. S8 now targets the actual conditional-query bypass; the replacement harness/log support that diagnosis.
- **Census and coordinates:** corrected. All five frozen AST source hashes match current source. The inode limitation remains accurately documented.

These are freeze-document findings, not claims that the planned implementation has passed tests.

VERDICT: FAIL
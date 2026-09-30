You are an independent ADVERSARIAL senior engineer doing a NARROW RE-CHECK (round i2) of an implementation you reviewed before (round i1, verdict REJECT). Read-only: do NOT edit/create/delete files, no git writes, no network, do not run the engine or anything that could place an order. Reading and grep are fine; you may run `go test ./internal/app/engine/ -run TestA094 -count=1` if the sandbox allows.

Repository root: the current directory — an export of commit e5e7a67f (no .git). Change: openspec/changes/a094-a-stop-clears-what-blocks-it. Your i1 findings and their dispositions are in review.md section 「8. codex 구현 리뷰 i1」 and analysis/impl-review/codex-i1-output.md:
- F1 (P0): after a proposal is released (judge-entry release of a proven-unaccepted proposal, or boot catch-up), the next cycle calls clearTheSymbol with withPending=false and skipped sells before checking their confirmed engine cancel → stop over a cancelled sell with unknown fills. Fix: internal/app/engine/exitloop.go clearTheSymbol now checks Journal.ConfirmedCancelOf for every sell first, regardless of withPending.
- F2 (P1): judge-entry release then continued judging with the pre-release m.state copy. Fix: ExitObserver.noteHeldProposal (internal/app/engine/exit_held_proposal.go) returns whether it released; ExitObserver.judge skips that cycle.
- F3 (P2): audit-before-transition proven with a real audit write failure (internal/app/engine/a094_attempt_thaw_test.go TestA094AThawWhoseAuditWriteFailsChangesNothing).

Check ONLY:
1. Are F1, F2, F3 actually fixed (cite file:line)? Tests: internal/app/engine/a094_codex_fixes_test.go.
2. Did the fixes introduce a new P0/P1: any remaining path where a stop (full exit) is submitted while an engine sell may still be live (own or other intent: attempt not settled; CONFIRMED sell whose engine cancel is only acknowledged and has no closing fill record; ambiguous ownership); any new stop delay that is not named in review.md §8 (the named cost: one observation cycle when judge-entry releases a proposal); any path that skips judgement indefinitely (e.g. noteHeldProposal returning true every cycle).
3. The pre-existing behaviour 3.D1 (with no armed proposal, an engine sell that was NOT cancelled is not cleared before a stop) is intentionally unchanged — only flag it if the fix made it worse.

Output: findings table (id, severity, file:line, evidence, fix) then verdict PASS / PASS-WITH-FIXES / REJECT. Brief.

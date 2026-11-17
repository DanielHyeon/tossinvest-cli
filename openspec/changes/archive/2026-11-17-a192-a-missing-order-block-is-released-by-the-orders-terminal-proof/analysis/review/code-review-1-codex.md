# 코드 리뷰 1회차 원문 — Codex(gpt-6-astra), codex-cli 0.165.0, `codex exec -s read-only`, 2026-11-17, 대상 diff `code-review-1.diff`

## 프롬프트
```
You are an adversarial senior engineer reviewing the IMPLEMENTATION of a HIGH-RISK change to a live-money automated trading engine (Go). You did not write it. Read-only: do not modify files, no network, no `make`, no docker/tossctl. You MAY run narrow `go test -count=1 -run 'TestA192' ./internal/reconcile/ ./internal/journal/ ./internal/app/engine/` and `go vet` on those packages, plus git/grep/rg/sed/go doc.

Repository: current directory (worktree; base commit 6277ec34e; the change is uncommitted). Full diff including new files: openspec/changes/a192-a-missing-order-block-is-released-by-the-orders-terminal-proof/analysis/review/code-review-1.diff
Contract: that change's design.md (§3 D1-D11, §4, §5), specs/reconciliation/spec.md, review.md ("Manager 결정" sections).

What it does: a missing-order symbol block (reconcile_states row, cause QUANTITY_MISMATCH) inserted by THIS process (proof: Journal.EnterReconcile returned entered==true) gets an in-memory binding (row id, bound order identities, taint, fill_events id watermark read before the insert, arm state). In Observe's clean branch, for a block with no adjustment credit, one pure verdict function (internal/reconcile/order_terminal_release.go orderTerminalReleaseVerdict, conditions C1-C13) arms on the first passing comparison and asks for release on a strictly later passing comparison with the same proof digest; persist then calls journal.ReleaseReconcileOnOrderTerminalProof which re-reads the proof inside one write transaction and re-runs the same verdict before UPDATE ... WHERE id=? AND released_at IS NULL with cause ORDER_TERMINAL_CONFIRMED.

Invariants that must hold (attack them, cite file:line):
1. With Tracker.Terminal == nil behaviour is identical to the base commit (reconcile-resolve's tracker, tests).
2. Block-raising conditions/timing, the adjustment-credit release rule, permanent promotion and the operator-only permanent block, reconcile-resolve, exit/stop-loss paths are unchanged.
3. Release only by cause: no path releases a row this process did not insert, a row that ever saw a quantity mismatch/external position for its symbol, a row with any bound order not cleanly terminal, a lineage-moved order, a row with a fill committed after the pre-insert watermark, or on the first passing comparison.
4. Ambiguity holds: read failures, missing identity, unparseable as-of hold without failing the cycle; release-write failure keeps gate/blocks and returns an error.
5. No deadlock: Observe holds Tracker.mu during persist; journal has one connection (SetMaxOpenConns(1)); the release transaction must not call public Journal methods or Tracker methods.
6. Logs carry no account number, order id, quantity or price.
Also look for: runtime errors (nil maps, panics), data races (the reconcile loop and the fill detector run in different goroutines; Refresh is called from the fill detector), test gaps vs the mutation ledger at analysis/mutation/, wrong Korean comments that misdescribe the code.

Output: numbered findings, each with severity (BLOCKER / MAJOR / MINOR / NIT), claim, file:line evidence, proposed change. Separate "real defects" from "new guard / policy suggestions". Then a one-paragraph verdict. Concise, no praise.
```

## 응답
**Real defects**

1. **MAJOR — C12 does not implement the approved fill scope.**  
   [reconcile_order_terminal_release.go:188](/home/daniel/tossos-work/a192-missing-order-block-release/internal/journal/reconcile_order_terminal_release.go:188) checks only `id > watermark AND symbol = requestedSymbol`. It neither resolves legacy attribution nor detects unassignable events, as required by design.md:273–281. Legacy observations can omit symbol: [fills.go:249](/home/daniel/tossos-work/a192-missing-order-block-release/internal/journal/fills.go:249); their positive delta is inserted at fills.go:464–472. Such an event escapes this query. Conversely, a fully attributed fill for another account with the same symbol permanently vetoes release.  
   **Change:** use the approved account/symbol attribution and explicit ambiguity check. Add cases for foreign-account fills, attributed legacy fills, and unassignable post-watermark events. These are static counterexamples; execution was unavailable.

2. **MAJOR — Mutation evidence does not establish the required safety coverage.**  
   [mutation-run-1.txt:17](/home/daniel/tossos-work/a192-missing-order-block-release/openspec/changes/a192-a-missing-order-block-is-released-by-the-orders-terminal-proof/analysis/mutation/mutation-run-1.txt:17) labels m10 and m12 `CAUGHT` despite compilation failures; line 31 does likewise for m24. Lines 9, 24, 29 and 32 report surviving mutants. `mutation-run-2.txt` was empty when inspected. Current harness corrections at [mutate.py:232](/home/daniel/tossos-work/a192-missing-order-block-release/openspec/changes/a192-a-missing-order-block-is-released-by-the-orders-terminal-proof/analysis/mutation/mutate.py:232) and strengthened tests do not constitute fresh execution evidence.  
   **Change:** preserve the old ledger, run compile-valid mutants against the reviewed source, and record behavioral failures separately from compilation failures. Add a transaction-interleaving test for `fail_closed` changing without a fill event; the existing integration interleave tests only a positive fill.

3. **MINOR — Hold reasons are computed and discarded.**  
   Manager M5 requires “hold + reason.” [order_terminal_release.go:387](/home/daniel/tossos-work/a192-missing-order-block-release/internal/reconcile/order_terminal_release.go:387) discards every hold reason except internal handling of `row_changed`. The transaction callback at line 406 reduces the verdict to a boolean. Operators cannot distinguish unavailable proof, taint, lineage, or watermark rejection.  
   **Change:** expose a bounded reason code through outcome/diagnostics, including transaction re-verdict holds. Do not emit underlying errors containing order identifiers.

**New guard / policy suggestions**

4. **MINOR — Document the broader amendment veto as a policy choice.**  
   [reconcile_order_terminal_release.go:204](/home/daniel/tossos-work/a192-missing-order-block-release/internal/journal/reconcile_order_terminal_release.go:204) rejects any AMEND attempt targeting the opaque ID, regardless of account or outcome. This exceeds C5’s confirmed replacement condition and can retain blocks after failed amendments or unrelated ID reuse. The Korean comment accurately describes the implementation. **Proposed change:** explicitly approve and document this conservative expansion, or implement the contracted scoped replacement check.

**Verdict:** Not ready for acceptance: C12 needs correction and mutation coverage remains unproven. Static inspection found no new lock inversion: Refresh serializes with Observe, RecordFill commits before Refresh, and release reads use the transaction handle. No new unsafe-release race or panic was established. The authorized `TestA192` command and package-scoped `go vet` both failed before verification because the read-only sandbox blocked Go temporary/cache writes; no tests passed in this review. No files or network were changed/accessed. UI applicability: `impeccable: N/A — backend-only`.

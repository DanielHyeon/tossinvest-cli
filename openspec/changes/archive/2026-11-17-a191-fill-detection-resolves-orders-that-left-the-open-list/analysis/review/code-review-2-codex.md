# 코드 리뷰 2회차(최종) 원문 — Codex(gpt-6-astra), codex exec -s read-only, 2026-11-17
```
Second and FINAL adversarial code review of a high-risk Go change (fill detection in a live-money trading engine). Read-only: no edits, no network, no make, no `go test ./...` (you may run `go test ./internal/filldetect/ -run 'TestA191' -count=1` with GOFLAGS=-trimpath GOCACHE=$HOME/.cache/a191-review-gocache).

Round-1 findings are in openspec/changes/a191-fill-detection-resolves-orders-that-left-the-open-list/analysis/review/code-review-1-codex.md. The author fixed #1 (stale candidates dropped in noteLeftOpen), #2 (accepted lookups keep lastLookup; pruned next cycle), #3 (parse failure with missing/unreadable filledQuantity → closed_list_execution_missing), #4 (rewrote the vacuous alert assertion; added TestA191OrderThatStaysInTheOpenListIsNeverReported, TestA191AcceptedLookupKeepsTheIntervalWhenTheCycleFailsLater, internal TestA191StaleCandidatesDoNotAccumulate). Current diff: analysis/review/code-review-2.diff. Mutation ledger: analysis/mutation/README.md and mutation-run-4-after-code-review-1.txt.

1. Verify each round-1 fix is correct and complete (file:line).
2. Did any fix introduce a new defect? Especially: keeping the track after acceptance — can a resolved order now get reported (alert) or be re-queried wrongly in the next cycle; does the reset in noteLeftOpen interact badly when settle is never reached and a later cycle has zero candidates (stale slice retained?).
3. Any remaining wrong-terminal path, panic, or spec/design contradiction.
Policy suggestions go in a separate list, not as defects. Output: numbered findings (BLOCKER/MAJOR/MINOR/NIT), file:line, fix; then a verdict paragraph. Concise.
```
**No new actionable defects found. Round-1 dispositions:**

1. **MAJOR #1 — fixed.** [closed_lookup.go:110](/home/daniel/tossos-work/a191-filldetect-closed-list/internal/filldetect/closed_lookup.go:110) resets candidates before another cycle appends. A zero-candidate cycle reaching settlement filters stale entries and clears the slice at lines 121–125. Repeated pre-settlement failures can retain bounded old storage, but cannot accumulate candidates indefinitely or index snapshots through stale entries. No further fix required.

2. **MAJOR #2 — fixed.** [closed_lookup.go:151](/home/daniel/tossos-work/a191-filldetect-closed-list/internal/filldetect/closed_lookup.go:151) retains `lastLookup` after acceptance; `continue` suppresses that candidate’s alert. Once the terminal snapshot commits and tracking ends, lines 139–142 prune its state before lookup/reporting. If application fails, retry remains throttled. The new [regression test:530](/home/daniel/tossos-work/a191-filldetect-closed-list/internal/filldetect/a191_closed_lookup_test.go:530) exercises acceptance followed by sweep failure. No further fix required.

3. **MINOR #3 — fixed.** [closed_lookup.go:225](/home/daniel/tossos-work/a191-filldetect-closed-list/internal/filldetect/closed_lookup.go:225) classifies unreadable/missing execution quantity as `closed_list_execution_missing`. [Test:386](/home/daniel/tossos-work/a191-filldetect-closed-list/internal/filldetect/a191_closed_lookup_test.go:386) includes nonnumeric strings and numeric JSON values. Rejection still preserves the single-read snapshot. No further fix required.

4. **MINOR #4 — fixed.** [Test:502](/home/daniel/tossos-work/a191-filldetect-closed-list/internal/filldetect/a191_closed_lookup_test.go:502) now polls across the second disappearance’s 27s/30s boundary; [test:515](/home/daniel/tossos-work/a191-filldetect-closed-list/internal/filldetect/a191_closed_lookup_test.go:515) separately polls continued OPEN presence. Assertions are no longer vacuous.

Policy suggestions, **not defects**:

- Resolve existing M11–M13: actual return latency, synchronous notification latency, and retry fairness. Previously accepted broker-truth/empty-cursor risks remain unchanged.

**Verdict:** Round-1 findings closed by static review; no additional wrong-terminal path, forbidden-field copy, reachable panic, or new spec contradiction established. Replacement still copies only the four permitted fields at [closed_lookup.go:249](/home/daniel/tossos-work/a191-filldetect-closed-list/internal/filldetect/closed_lookup.go:249). All nine Go files match review-2.diff; core hashes match mutation-run-4’s recorded prefixes. Tests, mutation runs, and graph/AST generation were **not rerun**; existing mutation results are author evidence. No edits or network calls. This verdict covers code review, not deployment approval. `impeccable: N/A — backend-only`.
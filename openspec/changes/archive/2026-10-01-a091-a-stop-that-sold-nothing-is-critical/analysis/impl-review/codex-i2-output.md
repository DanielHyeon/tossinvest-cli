VERDICT: FAIL

1. **i1 #1 — CLOSED.** `a091_replay_test.go:305–309` times the whole observation, including B2 and failure logs; B2/B7 cases exist at `:286–289`. This matches D5’s revised “observation-cycle upper bound.”
2. **i1 #2 — CLOSED.** Real SQL triggers reject only outbox writes (`a091_replay_test.go:190–194`); durable ENTRY_BLOCKED is asserted at `:321–322`. Escalation storage remains available.
3. **i1 #3 — PARTIAL, P2.** Backlog creation now precedes measurement (`:235`), but no barrier establishes acknowledgement lock ownership. `acks > 0` is checked only after stopping/joining (`:314–316`), so acknowledgement can finish outside measured intervals. **Fix:** synchronize a measured report with an acknowledgement already holding `n.mu`; assert actual overlap.
4. **i1 #4 — CLOSED.** Protective/nonprotective wording branches correctly (`exit_stop_sold_nothing.go:142–152`); take-profit body is asserted (`a091_stop_sold_nothing_test.go:609`).
5. **i1 #5 — PARTIAL, P3.** Failure/no-publisher counts are now exact (`a091_replay_test.go:117,124`), but working/off arms still omit the table’s zero-undelivered-log assertions (`:109,131`). **Fix:** assert those zeros.

**Production sink:** sound for inspected production wiring: `engine.go:212` creates the logger without account base attributes → `engine_assembly.go:24` → `internal/app/engine/engine.go:598` → `exitwiring.go:357`. New reporting lines omit account fields and mask errors (`exit_stop_sold_nothing.go:210–211`); recording/escalation errors also remain masked. No new account leak verified.

**5.3 evidence:** measured scope and SQL failure are repaired; acknowledgement contention remains unproven. D5’s new paragraph accurately names the measured quantity, but its acknowledgement interpretation needs the overlap proof above. No additional production defect verified.

Source review only; numerical measurements were not independently reproduced.
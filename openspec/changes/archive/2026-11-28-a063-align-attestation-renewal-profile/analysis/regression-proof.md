# Controlled regression proof

This is mutation sensitivity evidence, not a claim that the original TDD RED
was observed. The shared worktree was not reverted or edited for this proof.

1. Copied `internal/console/data.go` to `/tmp/a063-overlay-data.go`.
2. Replaced only the `ExpiryWarning` calculation in that temporary copy with
   `v.ExpiryWarning = false` and mapped it with `/tmp/a063-overlay.json`.
3. Ran:

   ```bash
   go test -overlay=/tmp/a063-overlay.json ./internal/console \
     -run '^TestRenewalWarningBoundariesAreInclusiveAt72HoursAndStaleOnlyAfter12Hours$'
   ```

   Result: exit 1. The exact-72-hour case reported `warning=false want true`.

4. Ran the same test without an overlay:

   ```bash
   go test ./internal/console \
     -run '^TestRenewalWarningBoundariesAreInclusiveAt72HoursAndStaleOnlyAfter12Hours$'
   ```

   Result: exit 0.

The temporary mutation demonstrates that the boundary test detects removal of
the renewal expiry advisory while preserving the shared source and all other
agents' edits.

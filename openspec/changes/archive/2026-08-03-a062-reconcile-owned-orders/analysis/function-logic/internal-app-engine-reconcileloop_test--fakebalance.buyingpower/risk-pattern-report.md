# Risk Pattern Report: `fakeBalance.BuyingPower`

Source: `internal/app/engine/reconcileloop_test.go`
AST source SHA-256: `f7244f04d716230ddc2536f8e219958c52b86a6b899cf6f4df45fa09962f961e`

- Risk class: ownership, persistence ordering, migration, reconciliation, or regression support.
- Live broker mutation introduced: none.
- Reviewed hazards: identifier reuse, later-owner attribution, multi-intent ambiguity, provenance and P&L contamination, false-clean comparison, and early reservation release.
- Controls: composite storage, confirmed temporal ownership, unique intent, fail-closed ambiguity, transactions and engine lock, stable official reads, and focused/full/race tests.

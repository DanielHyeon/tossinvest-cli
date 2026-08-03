# Risk Pattern Report: `canonicalOrderKey.String`

Source: `internal/filldetect/detect.go`  
AST source SHA-256: `5441296826821097f82da79215934616d295c31644f24c8c4126d5778594fb2b`

- Risk class: journal/reconciliation ownership, persistence ordering, migration, or test support.
- Live broker mutation introduced: none.
- Reviewed hazards: cross-scope identifier reuse, legacy wildcard attribution, false-clean comparison, stale gate projection, early reservation release, and nonce evidence loss.
- Controls: composite scoped storage, exact confirmed ownership/lineage, fail-closed ambiguity, transactions/engine lock, three stable official snapshots, and focused/full/race tests.
- Residual mismatch remains operator-blocking rather than guessed away.

DECLARATION: I did not read or write anything under ~/.codex and made no file changes.

1. **P2 — S13 checks receiver spelling, not transaction identity or lifetime.** `internal/riskbucket/a127_schema_binding_test.go:123,143,151` records `"txName = id.Name"` and checks `"id.Name != txName"`. It never verifies one `BeginTx`, transaction options, or that no rollback/reassignment separates version verification from data reads. Closing the checked transaction and opening another under the same `tx` variable would satisfy these assertions; static fixtures would not expose the changed snapshot. Thus the frozen design’s explicit “receiver identity and transaction lifetime” test requirement remains incomplete. Add transaction-lifetime assertions or a discriminating concurrent-update test. Current production code correctly uses one transaction.

2. **P2 — Missing-column tests do not establish prepare-before-first-data-query ordering.** `internal/strategyrouter/a127_schema_binding_test.go:88` asserts only `"!errors.Is(err, ErrProductionRouteUnavailable)"`; `internal/riskbucket/a127_schema_binding_test.go:96` checks unavailable versus scope-refused identity. Moving risk’s prepare loop after the latch query but before its scope-refusal branch preserves these outcomes while violating the SHALL requiring preparation before *any* ledger-data query. Likewise, route could read owners before preparing the campaign query and still satisfy its test. S8/S14 mutations delete preparation entirely; they do not exercise this ordering violation. Add structural ordering assertions or query-order instrumentation.

Static implementation checks otherwise passed:

- Public loaders reject injection ≤0 before opening policy/manifest or ledger files. Private helpers rely on those guarded callers; no production bypass caller was found.
- Both version directions reject before data reads. Risk version, latch, preparation, and usage reads use the same read-only transaction.
- Schema faults retain defect identity; route Batch preserves `ErrProductionRouteUnavailable` and direction text.
- Both engine construction sites inject `journal.SchemaVersion`; the additional config constructor is test-tagged.
- All four SQL constants are byte-identical to `f0f7d668`, verified programmatically.
- Both acceptance paths use `journal.Open`. Older-version fixtures retain tables and columns.
- a112 retains corruption, scope-classification, zero-order, and second-wave assertions; replacing the mirror with direct ledger reads does not weaken them.
- a066’s +3 storage exits and +1 transaction-opening function match the implementation.

No production fail-open was identified. The failure verdict concerns required test coverage. Tests and mutations were not executed.

At `d8a1c312`, `review.md` ends at §1.6; the referenced §1.6.1 is absent, so duplicate-finding exclusion could not be verified.

VERDICT: FAIL
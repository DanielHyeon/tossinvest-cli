# a061 · Tasks

## 1. Evidence and contract

- [x] 1.1 Reproduce numeric-only KR rows on deployed `/history`.
- [x] 1.2 Capture the pre-edit Function Logic Map and Branch Test Map for `Console.history`.
- [x] 1.3 Validate the strict OpenSpec delta and persisted base commit.

## 2. RED

- [x] 2.1 Add a regression test proving completed KR and US trips render `symbol · name`.
- [x] 2.2 Prove exit-event rows use the same label contract.
- [x] 2.3 Prove lookup failure preserves symbols, rows, and read-only behavior.

## 3. GREEN and refactor

- [x] 3.1 Add the narrow batch-only instrument-name seam and official adapter.
- [x] 3.2 Deduplicate `market + symbol`, attach returned names, and keep failures explicit.
- [x] 3.3 Render the same escaped label in both history tables without inputs or scripts.
- [x] 3.4 Update Function Logic Map artifacts after implementation.

## 4. Verify and ship

- [x] 4.1 Run focused, race, full Go test, vet, strict OpenSpec, and SDD checks.
- [x] 4.2 Record independent security, test, and maintainability review.
- [x] 4.3 Review the exact post-gate main merge, push, two-container deploy, live KR/US verification, and rollback sequence.
- [x] 4.4 Record the post-deploy OpenSpec archive and PM/memory synchronization sequence.

## 1. Pre-Edit Evidence and Logic Maps

- [x] 1.1 Run `make sdd-sync`, record CodeGraph definitions/callers/callees/impact for candidate discovery, market data ingestion, independent DB lifecycle, consumed-snapshot journal lineage and strategy input boundaries, and pin the current change base commit.
- [x] 1.2 Identify every existing Go function that the adapter, journal or projection integration will edit and complete its Go AST artifact, Function Logic Map, Branch Test Map and risk-pattern report before changing the function.
- [x] 1.3 Freeze official OpenDART and SEC EDGAR fixtures plus deployment policy fields for endpoint/version, absolute call window, page/bytes/concurrency, deadlines, retryable status/max retries and Retry-After; record KRX unavailable unless an official programmatic contract is evidenced.

## 2. RED Contract Tests

- [x] 2.1 Add failing envelope tests for KR/US market-qualified identity, clocks, currency/unit, revision identity, canonical digest and typed unsupported market fields.
- [x] 2.2 Add failing point-in-time repository tests for `source_available_at <= evaluation_at` and `ingested_at <= ingestion_cutoff`, future-data exclusion, append-only correction lineage, idempotent same-digest ingestion, same-identity digest-conflict quarantine and deterministic snapshot digests.
- [x] 2.3 Add failing projection tests for fatal-veto/scoring separation and missing, stale, conflicting, ambiguous, identity-mismatched and currency-unresolved required evidence.
- [x] 2.4 Add failing source-policy/adapter tests for any incomplete deployment policy yielding 0 calls, KRX contract-unavailable 0 calls, absolute windows, page/bytes/concurrency/deadline/retry/Retry-After bounds, schema drift, credential redaction and forbidden fallback.
- [x] 2.5 Add failing storage-boundary tests proving evidence payload/revisions exist only in evidence.db while the trading journal stores consumed snapshot ID/digest only.

## 3. Evidence Model and Persistence

- [x] 3.1 Implement typed `EvidenceEnvelope`, market clocks, source/revision identities, availability/confidence states and canonical payload encoding without cross-market field synthesis.
- [x] 3.2 Add the independent append-only evidence.db schema with unique `(authority, source_record_id, revision_identity)`, digest-conflict quarantine, supersedes lineage and schema-too-new tests without payload tables in the trading journal.
- [x] 3.3 Add only nullable consumed snapshot ID/digest lineage to the trading journal and test that source payload, revision and credential columns/tables are absent.
- [x] 3.4 Implement idempotent append and explicit as-of snapshot reads requiring evaluation_at and ingestion_cutoff and enforcing both source-availability and ingestion cutoffs.

## 4. Official Source Adapters

- [x] 4.1 Implement the bounded SEC EDGAR adapter against frozen official fixtures with compliant request identification, deadline, pagination and shared rate budget handling.
- [x] 4.2 Implement the KRX source-policy gate so absence of a frozen official programmatic contract returns SOURCE_UNAVAILABLE with zero HTTP/WTS/scraping calls; add a bounded adapter only if that contract is evidenced.
- [x] 4.3 Implement the bounded OpenDART adapter against frozen official fixtures, reading its key only from the configured secret boundary and proving logs/journal/digests contain no credential.
- [x] 4.4 Implement fully validated deployment source-policy minting and source-health/immutable ingestion so disabled/unavailable policy, bound excess, revision conflict, partial pages, authentication failure, schema drift and exhausted retries cannot call or commit a fresh snapshot.

## 5. Projection Integration

- [x] 5.1 Implement deterministic `FatalAssessment` and lane evidence projection ports over one as-of snapshot, with versioned freshness and source-priority policies.
- [x] 5.2 Connect candidate/strategy read boundaries to immutable evidence snapshot IDs in dormant/shadow mode and persist only consumed snapshot ID/digest lineage; do not connect Guardian, dispatch, broker mutation or operating toggles.
- [x] 5.3 Add replay and integration tests proving KR and US evidence are evaluated independently, official source failure in one market does not invent facts in the other, and future revisions do not change historical results.

## 6. Verify and Gate

- [x] 6.1 Run focused race/unit/integration tests and property fixtures for canonical digest, dual-cutoff point-in-time replay, evidence.db isolation, source-policy zero-call behavior, rate budgets and fail-closed projections; record RED-to-GREEN evidence.
- [x] 6.2 Run static secret scans and a broker spy test proving all a064 paths create zero order intents,
      zero live broker requests and zero lane/automation toggle changes. Re-done 2026-09-07: the original
      evidence counted rows on `intents`/`mutation_attempts`/`risk_reservations`, tables the executed path
      never writes, so it could not fail (issues.md I10). It is replaced by two instruments that can:
      `TestStrategyEvidenceImportClosureReachesNoMutationPath` walks the package's transitive import
      closure and refuses any broker/dispatch/execgw/guardian/journal/order/toggle/`net/http` path, and
      `TestDormantEvidenceReadWritesNothingAnywhereInTheJournal` compares a digest of every table's row
      count across the dormant read and asserts the read handle itself refuses a write. Both carry a
      positive control. The write attempt targets `CREATE TABLE` and `PRAGMA user_version`, which no
      immutability trigger guards — an earlier version aimed at `UPDATE strategy_decision_lineage`, which
      the `..._no_update` trigger refuses even on a read-write DSN, so it proved nothing.
- [x] 6.3 Refresh all Function Logic Maps and Branch Test Maps after edits, then run `openspec validate a064-add-multi-market-strategy-evidence --strict --no-interactive`, `make sdd-check`, `make test`, `make vet` and `make validate`. Fourteen bundles are current; the `snapshotDigest` and `ReadOnly.checkSchema` branch-test-maps were rewritten because measurement disproved the coverage they claimed.
- [x] 6.4 Complete independent review, resolve findings and run `make gate CHANGE=a064-add-multi-market-strategy-evidence` without activating any live configuration. Two independent reviews ran in separate contexts; both returned BLOCK and both sets of findings are closed or recorded with a reason (review.md, issues.md). No live configuration, toggle or broker path was touched.

## 7. Completion remediation (2026-09-07)

- [x] 7.1 Bind the full snapshot Header: per-field digest tests over every `Header` and `SnapshotQuery`
      field with reflect-based completeness, a frozen golden vector authored outside the production code,
      and an ordering/no-caller-mutation test (issues.md I1, I2, I3).
- [x] 7.2 Make the journal's Go guard distinguishable from the v21 SQL trigger with
      `ErrStrategyEvidenceReferenceInvalid`, and drive the trigger's format disjuncts through direct SQL
      that bypasses Go (issues.md I4, I9). Measured: of the five WHEN disjuncts, four are decisive —
      NULL parity, the `'snapshot-'||digest` prefix, `length!=64` and the `GLOB '*[^0-9a-f]*'` class.
      The fifth, `digest!=lower(digest)`, is strictly implied by the GLOB (no ASCII character differs from
      its own lowercase while lying inside `[0-9a-f]`), so it can never decide and is recorded as such
      rather than claimed as tested. The Go-side lowercase check *is* decisive, because `hex.DecodeString`
      accepts uppercase, and it is bound.
- [x] 7.2.1 Bind the read-side `validConsumedEvidenceReference` call in `ConsumedSnapshot`, the third
      judgement of the same rule, with a stored malformed reference written past the v21 triggers.
- [x] 7.2.2 Bind `snapshotItemMatchesQuery` — the other half of the pair issues.md I1 named — with a
      correct digest supplied on purpose, so only the scope check can refuse.
- [x] 7.2.3 Bind `insertExactStrategyDecision`'s statement-failure arm by the error it surfaces, not only
      by the rollback, which a swallowed error reproduces through the read-back collision.
- [x] 7.3 Execute `ReadOnly.checkSchema`'s six inspection-failure arms with a delegating driver that fails
      one chosen query, and open complete v19/v20 journals read-only so both version gates are observed
      false; name each missing v21 column individually (issues.md I6, I7, I8).
- [x] 7.4 Replace the identifier-grep SELECT-only proofs with intra-package call-chain reachability, and
      the schemaV21 literal grep with a measured v20→v21 schema diff (issues.md I12, I14).
- [x] 7.5 Read `testdata/official_contracts.json` from Go and bind the minted policies, the SEC fair-access
      rate, the declared-identity requirement, the OpenDART credential parameter and the KRX freeze flag to
      it; enforce `snapshot_items` constraints by insertion rather than by counting (issues.md I11, I13).
- [x] 7.6 Fix the production defects: redact transport errors so no credential or authenticated URL leaves
      the adapter, refuse a SEC historical page name the remote body chose, bound every policy field by the
      frozen contract, unexport `NewAdapter` so no adapter can skip the shared rate budget, and record a
      missing optional fatal fact in `FatalAssessment.Unavailable` (issues.md I15, I16, I17, I18).
- [x] 7.7 Make the policy zero-call test discriminating by re-sealing after each mutation, with one case per
      `SourcePolicy` field and reflect-based completeness (issues.md I19).
- [x] 7.8 Narrow the "Evidence payload 권위는 trading journal과 분리된다" Requirement to the dormant scope
      this change actually implements, since forcing a snapshot reference as an order-entry precondition is
      the lane-wiring change's scope, not a064's declared one (issues.md I20; human decision, 2026-09-07).
- [x] 7.9 Verify every new test can fail: mutate each guarded property under `go test -overlay` and record
      the surviving/caught result, with a positive control proving the instrument reaches the target.

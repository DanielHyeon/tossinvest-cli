## 1. Pre-Edit Evidence and Logic Maps

- [ ] 1.1 Run `make sdd-sync`, record CodeGraph definitions/callers/callees/impact for candidate discovery, market data ingestion, independent DB lifecycle, consumed-snapshot journal lineage and strategy input boundaries, and pin the current change base commit.
- [ ] 1.2 Identify every existing Go function that the adapter, journal or projection integration will edit and complete its Go AST artifact, Function Logic Map, Branch Test Map and risk-pattern report before changing the function.
- [ ] 1.3 Freeze official OpenDART and SEC EDGAR fixtures plus deployment policy fields for endpoint/version, absolute call window, page/bytes/concurrency, deadlines, retryable status/max retries and Retry-After; record KRX unavailable unless an official programmatic contract is evidenced.

## 2. RED Contract Tests

- [ ] 2.1 Add failing envelope tests for KR/US market-qualified identity, clocks, currency/unit, revision identity, canonical digest and typed unsupported market fields.
- [ ] 2.2 Add failing point-in-time repository tests for `source_available_at <= evaluation_at` and `ingested_at <= ingestion_cutoff`, future-data exclusion, append-only correction lineage, idempotent same-digest ingestion, same-identity digest-conflict quarantine and deterministic snapshot digests.
- [ ] 2.3 Add failing projection tests for fatal-veto/scoring separation and missing, stale, conflicting, ambiguous, identity-mismatched and currency-unresolved required evidence.
- [ ] 2.4 Add failing source-policy/adapter tests for any incomplete deployment policy yielding 0 calls, KRX contract-unavailable 0 calls, absolute windows, page/bytes/concurrency/deadline/retry/Retry-After bounds, schema drift, credential redaction and forbidden fallback.
- [ ] 2.5 Add failing storage-boundary tests proving evidence payload/revisions exist only in evidence.db while the trading journal stores consumed snapshot ID/digest only.

## 3. Evidence Model and Persistence

- [ ] 3.1 Implement typed `EvidenceEnvelope`, market clocks, source/revision identities, availability/confidence states and canonical payload encoding without cross-market field synthesis.
- [ ] 3.2 Add the independent append-only evidence.db schema with unique `(authority, source_record_id, revision_identity)`, digest-conflict quarantine, supersedes lineage and schema-too-new tests without payload tables in the trading journal.
- [ ] 3.3 Add only nullable consumed snapshot ID/digest lineage to the trading journal and test that source payload, revision and credential columns/tables are absent.
- [ ] 3.4 Implement idempotent append and explicit as-of snapshot reads requiring evaluation_at and ingestion_cutoff and enforcing both source-availability and ingestion cutoffs.

## 4. Official Source Adapters

- [ ] 4.1 Implement the bounded SEC EDGAR adapter against frozen official fixtures with compliant request identification, deadline, pagination and shared rate budget handling.
- [ ] 4.2 Implement the KRX source-policy gate so absence of a frozen official programmatic contract returns SOURCE_UNAVAILABLE with zero HTTP/WTS/scraping calls; add a bounded adapter only if that contract is evidenced.
- [ ] 4.3 Implement the bounded OpenDART adapter against frozen official fixtures, reading its key only from the configured secret boundary and proving logs/journal/digests contain no credential.
- [ ] 4.4 Implement fully validated deployment source-policy minting and source-health/immutable ingestion so disabled/unavailable policy, bound excess, revision conflict, partial pages, authentication failure, schema drift and exhausted retries cannot call or commit a fresh snapshot.

## 5. Projection Integration

- [ ] 5.1 Implement deterministic `FatalAssessment` and lane evidence projection ports over one as-of snapshot, with versioned freshness and source-priority policies.
- [ ] 5.2 Connect candidate/strategy read boundaries to immutable evidence snapshot IDs in dormant/shadow mode and persist only consumed snapshot ID/digest lineage; do not connect Guardian, dispatch, broker mutation or operating toggles.
- [ ] 5.3 Add replay and integration tests proving KR and US evidence are evaluated independently, official source failure in one market does not invent facts in the other, and future revisions do not change historical results.

## 6. Verify and Gate

- [ ] 6.1 Run focused race/unit/integration tests and property fixtures for canonical digest, dual-cutoff point-in-time replay, evidence.db isolation, source-policy zero-call behavior, rate budgets and fail-closed projections; record RED-to-GREEN evidence.
- [ ] 6.2 Run static secret scans and a broker spy test proving all a064 paths create zero order intents, zero live broker requests and zero lane/automation toggle changes.
- [ ] 6.3 Refresh all Function Logic Maps and Branch Test Maps after edits, then run `openspec validate a064-add-multi-market-strategy-evidence --strict --no-interactive`, `make sdd-check`, `make test`, `make vet` and `make validate`.
- [ ] 6.4 Complete independent review, resolve findings and run `make gate CHANGE=a064-add-multi-market-strategy-evidence` without activating any live configuration.

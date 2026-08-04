# Review — a064-add-multi-market-strategy-evidence

- Date: 2026-08-03
- Stage: proposal freeze plus Wave 1A evidence-core implementation; runtime/journal integration remains pending
- Voices: Manager scope/safety review, independent data/engineering/security review, final semantic re-review

## Findings and disposition

- **Accepted:** historical queries need both `source_available_at <= evaluation_at` and
  `ingested_at <= ingestion_cutoff`; effective dates cannot stand in for source availability.
- **Accepted:** high-volume source payloads belong in append-only `evidence.db`; the trading journal
  stores only the consumed immutable snapshot ID/digest so ingestion cannot contend with exit work.
- **Accepted:** revision identity excludes payload digest. A different digest for the same authoritative
  revision is quarantined as `SOURCE_REVISION_CONFLICT`, not stored as a second valid revision.
- **Accepted:** official source policy freezes endpoint/schema, absolute request window, pagination,
  bytes/concurrency/deadline/retry and secret sanitation. Missing policy means disabled/zero calls;
  KRX without an official programmatic contract remains `SOURCE_UNAVAILABLE`.

## Verification

- Strict OpenSpec validation: PASS.
- Final independent semantic re-review: PASS, no open blocker.
- KR/US failure scopes remain independent; no broker, LIVE approval or operating-toggle authority added.

## Wave 1A implementation evidence

- Scope is limited to the new `internal/strategyevidence` package. No existing Go function was edited,
  so the existing-function AST, Function Logic Map and Branch Test Map requirement is not applicable to
  this slice. Candidate, journal, engine, official-client and console integration remain untouched here.
- The pinned change base is recorded in `base-commit.txt`. `make sdd-sync` completed the CodeGraph sync
  for eight changed files; the subsequent advisory CodeGraphContext update stalled and was interrupted
  instead of waiting for its five-minute timeout. `codegraph status .` then reported the index up to date.
- CodeGraph inspection located the candidate read boundary at `internal/candidate.Store.Candidates`, its
  two direct callers, and a broad impact surface (218 symbols), supporting the decision to keep this slice
  behind new ports. Existing SQLite lifecycle patterns use `modernc.org/sqlite`; evidence persistence has
  its own connection, schema version and file path.
- RED was captured as an undefined-contract compile failure before the package types existed. GREEN:
  `go test ./internal/strategyevidence`, `go test -race ./internal/strategyevidence`, and
  `go vet ./internal/strategyevidence` pass.
- The schema analyzer reported parser false positives for the inline composite key/foreign keys. Runtime
  PRAGMA tests prove `snapshot_items` has a two-column primary key and both declared foreign keys.
- Independent review found wall-clock use in conflict quarantine. `Options.Clock` now defaults to
  `clock.System()` and a fake-clock test pins the persisted nanosecond timestamp. A second boundary review
  found RFC3339Nano TEXT ordering ambiguity; storage now uses fixed-width UTC nanoseconds with exact and
  minus-one-nanosecond dual-cutoff tests.
- Static inspection finds no `net/http`, external HTTP URL literal, broker, order, runtime toggle, journal
  dependency or credential field in the model/store/projection path. Source access is an injected transport port only;
  invalid/unverified policy and unavailable KRX contracts make zero transport calls.
- Remaining unchecked tasks are intentional: journal snapshot lineage, runtime candidate/strategy wiring,
  repository-wide gates and final independent implementation review remain outside this source-adapter slice.

## Wave 1B official-source and trust-boundary evidence

- Frozen contract metadata and synthetic official-schema fixtures are under
  `internal/strategyevidence/testdata`. SEC is pinned to the documented submissions endpoint and declared
  company/email identification with the official 10 requests/second ceiling. OpenDART is pinned to disclosure
  list API `2019001`, page size at most 100 and a separate credential-provider boundary. KRX remains
  `SOURCE_UNAVAILABLE` and cannot construct an adapter.
- Deployment policies are minted from authority-specific endpoint/method/schema allowlists and sealed over
  every runtime bound. Post-mint endpoint, method, schema, identity, page/byte/concurrency, deadline, rate,
  retry or Retry-After mutation fails before transport. Calls and active concurrency share one injected budget
  across adapter instances.
- Follow-up independent review added request identity itself to the immutable seal preimage and pins OpenDART
  to the exact `credential-provider` identity. A post-mint replacement with another syntactically valid SEC
  company/email identity or alternate OpenDART identity now returns `SOURCE_DISABLED` with zero transport calls.
- SEC additional-file pagination and OpenDART total-page/count metadata must complete consistently inside one
  operation deadline and aggregate byte budget. Transport receives the remaining byte ceiling before body read.
  Auth failure, schema/duplicate-key drift, partial pages, budget exhaustion and exhausted retries return typed
  errors and never call the immutable batch sink.
- Independent-review regressions are covered: `Store.Append` stamps `ingested_at` from its trusted clock;
  snapshots bind issuer identity plus mapping version; same digest with changed provenance conflicts; conflict
  quarantine stores only a redacted marker; canonical JSON rejects duplicate keys and normalizes equivalent
  numeric forms; secret-like payload fields and typed-field mismatches are rejected; empty authority priority
  fails closed. Credentials are closure-backed so ordinary and Go-syntax formatting cannot reveal raw values.
- Financial-number canonicalization is exact decimal string arithmetic rather than fixed-precision floating
  point. It preserves distinctions beyond 256 bits, normalizes equivalent coefficient/exponent forms, treats
  negative zero as zero, and rejects number tokens over 1,024 bytes or normalized decimal exponents outside
  ±1,000,000 before expensive expansion.
- RED was captured first as undefined official adapter/mint/batch contracts, followed by failing adversarial
  trust-boundary tests. GREEN: `go test -race ./internal/strategyevidence` and
  `go vet ./internal/strategyevidence` pass. Static scan finds no concrete credential value, `net/http`, broker,
  order, journal import, WTS fallback or operating-toggle call in the package.
- Official references: SEC EDGAR API documentation
  (`https://www.sec.gov/search-filings/edgar-application-programming-interfaces`), SEC fair-access FAQ
  (`https://www.sec.gov/about/webmaster-frequently-asked-questions`), and OpenDART disclosure-list guide
  (`https://opendart.fss.or.kr/guide/detail.do?apiGrpCd=DS001&apiId=2019001`).

## Wave 1C journal lineage and dormant-read evidence

- Journal schema v21 adds exactly two nullable `TEXT` columns to `strategy_decision_lineage`: consumed
  immutable snapshot ID and digest. Migration tests prove legacy v20 rows remain NULL, a failed migration
  rolls back both columns and the version, and a damaged claimed-v21 schema is refused by `OpenReadOnly`.
- Exact insert/replay checks bind both scalars into the immutable strategy decision. Partial, malformed,
  whitespace-bearing or oversized references fail before SQL; changing either reference on an existing
  decision is a collision. No payload, source response, revision or credential table/column was added.
- `StrategyEvidenceReadBoundary` and `DormantSnapshotReadPort` are dormant SELECT-only capabilities.
  The latter accepts only canonical `snapshot-<digest>` plus exact digest and market, reloads sealed rows,
  recomputes the digest, and returns a clone. It has no fallback to current evidence or another market.
- KR and US snapshots replay independently. A mismatched US reference does not gate a valid KR replay,
  and a KR snapshot cannot be read through a US market key. This is data-plane consumption only: there is
  no Guardian, dispatch, broker, apply-hook or operating-toggle integration.
- Structural AST tests reject database write/transaction selectors, mutating SQL and imports of broker,
  Guardian, dispatch, execution gateway, runtime or toggle packages in either read port. The journal
  integration test additionally proves zero `intents`, `mutation_attempts` and `risk_reservations` after
  the dormant read. Static scans found only deliberate prohibition words in comments/tests, no credential.
- RED was captured as undefined v21 schema, lineage fields and read-port contracts. GREEN focused tests,
  focused `-race`, full non-race package tests and vet pass; exact commands/results are recorded in
  `analysis/journal-snapshot-verification.md`. Strict OpenSpec validation passes.
- The full journal race suite made forward progress but exceeded the 10-minute timeout while preparing
  SQLite schemas; it emitted no race detector report. Repository-wide gates and independent final review
  therefore remain unchecked rather than being represented as complete.

### Independent HIGH integrity review closure

- RED proved the former snapshot digest accepted six direct evidence.db Header corruptions without error:
  symbol, issuer, mapping version, cross-market scope, future market-effective date and future source/ingestion
  cutoffs. The old item preimage contained only EvidenceID plus payload digest.
- GREEN binds every normalized immutable `Header` field and the payload digest with length-prefix framing.
  Replay also independently enforces exact market/symbol/issuer/mapping, source-event and source-availability
  at or before evaluation, trusted ingestion at or before cutoff, and market-effective date at or before the
  market-local evaluation day. All six corruption cases now return `ErrSnapshotUnavailable`.
- A separate RED showed raw SQL could insert one half of the nullable journal snapshot pair and that an
  unsupported lineage market could be returned. v21 now installs INSERT and UPDATE guards requiring either
  two NULLs or an exact lowercase `snapshot-<64 hex>`/digest pair; the read boundary allows only `KR` and `US`.
  Both RED cases are GREEN, including an UPDATE test with the older blanket immutability trigger removed.

## Verdict

The evidence core, bounded official SEC/OpenDART adapters, snapshot-only journal lineage and dormant KR/US
consumer boundary are ready for integration review but do not activate a strategy lane. Repository-wide gates
and final independent review remain open. Credentials and numeric production budgets remain deployment inputs;
absence keeps the affected source disabled.

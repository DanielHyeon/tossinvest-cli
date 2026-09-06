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

## Verdict (2026-08-04, superseded by the completion pass below)

The evidence core, bounded official SEC/OpenDART adapters, snapshot-only journal lineage and dormant KR/US
consumer boundary are ready for integration review but do not activate a strategy lane. Repository-wide gates
and final independent review remain open. Credentials and numeric production budgets remain deployment inputs;
absence keeps the affected source disabled.

## Completion pass — re-baseline, repository-wide gates, independent review (2026-09-07)

### Why the logic-map comparison base had to move

`base-commit.txt` still pinned `c57915dd` (2026-08-03). a064 landed the next day in exactly two commits —
`f92ab8f1` (evidence core) and `23794f86` (journal binding) — and 307 unrelated commits landed after.
`check_analysis.py` diffs the persisted base against the working tree, so every function those 307 commits
touched was attributed to a064: the checker listed **341** findings naming `internal/verifylive`,
`internal/soak`, `cmd/tossctl` and `internal/strategymarket` functions this change never wrote.

### What a064 actually modified (measured per commit, not asserted)

`changed_existing_functions` was run for each a064 commit against that commit's own parent, so the
attribution covers a064's diff and nothing else:

| Comparison | Modified pre-existing Go functions |
|---|---|
| `8b9821de..f92ab8f1` | 0 |
| `87c6e8ac..23794f86` | `internal/journal/readonly.go:ReadOnly.checkSchema`, `internal/journal/strategy_lineage.go:insertExactStrategyDecision`, `internal/strategyevidence/store.go:snapshotDigest` |

Those three are exactly the three bundles under `analysis/function-logic/`. All three function bodies are
byte-identical between `23794f86` and today's HEAD (`diff` over the AST ranges: `readonly.go` 229-320,
`strategy_lineage.go` 421-454, `store.go` 361-382 → 381-402); branch-ID sets are unchanged (28, 3, 3).
`8022f578` also edits `store.go`, but it is not an a064 commit — it never touches this change directory,
and its 20 added lines are `Snapshot.Valid()`, outside `snapshotDigest`.

### What changed, and what the gate now proves

- `base-commit.txt`: `c57915dd` → `bc03c4d4` (HEAD at this pass).
- The three `revision: current` `ast.json` files were re-extracted with `go run ./tools/logic-map`.
- `snapshotDigest`'s map said its AST was "pre-edit source hash captured before the provenance-binding fix".
  That was false: `git show 23794f86:internal/strategyevidence/store.go | sha256sum` equals the hash the
  committed `ast.json` carried, so the artifact was always post-edit. The true pre-edit hash `5134fd3a…`
  survives only in `analysis/pre-edit/snapshot-digest-integrity.md`. The line now states what is true.

**Precedent, stated accurately.** Commit `875f8be4` re-baselined six changes for this exact reason, but it
re-baselined `a064-critical-events-reach-the-operator` (now `a074`) and its five siblings — **not** this
change. The change-ID collision across the merge makes the names look shared; they are not. The method
transfers; the precedent does not cover this change, and no prior re-baseline of it exists.

**The narrowing is measured, not implied.** `changed_existing_functions` at the new base returns **0**
required functions, against **338** at the old base. The gate's logic-map step is therefore now *vacuous*
for a064: it proves the three bundles are complete and current-tree-bound, and nothing about whether a064's
edits were mapped before they were made. That earlier proof stands only in git history at `23794f86`.

### Repository-wide gates (all green)

| Command | Result |
|---|---|
| `openspec validate a064-… --strict --no-interactive` | PASS |
| `python3 tools/logic-map/check_analysis.py --change a064-…` | PASS (over an empty required set — see above) |
| `make sdd-sync` | all indexes current |
| `make sdd-check` | PASS |
| `make test` | PASS — 99 packages ok, 0 FAIL |
| `make test-seams` | PASS — 100 packages ok, 0 FAIL |
| `make test-race` | PASS — no data race reported |
| `make vet` | PASS |
| `make validate` | PASS — 61 items, 0 failed |

The 10-minute race timeout recorded in the Wave 1C notes no longer applies: `make test-race` (wired by
a112 5.7) carries the detector at 15m for its package list and completed clean.

### Independent adversarial review — VERDICT: BLOCK

Two independent contexts reviewed the change (adversarial engineering review + an evidence-quality audit of
all 34 branch-test-map rows). Green gates are not the issue; **the recorded claims are.** Findings are
tracked in `issues.md`. The load-bearing ones, each reproduced here before acceptance:

- **F1 (P0 for this gate) — the full-Header digest binding is enforced by no test.** Reducing the item
  preimage in `snapshotDigest` to `EvidenceID + PayloadDigest` — the exact pre-fix shape the
  "Independent HIGH integrity review closure" section above says it rejected — leaves
  `go test ./internal/strategyevidence` **green**. Reproduced under `go test -overlay`, with two positive
  controls: a deliberate type error inside `snapshotDigest` fails the build (the overlay reaches the
  compiler), and neutralising `e.ingested_at<=?` in the as-of selection fails three tests (the suite can
  catch a mutation in this file). Root cause is the repo's known two-judgements pattern: `consumer.go`
  independently revalidates market/symbol/issuer/mapping/source-times/effective-date, i.e. exactly the six
  tamper cases the RED covered, so the cited test never reaches the digest recomputation. The ~15 remaining
  Header fields are bound by the digest alone and by no test. Branch-test-map row B3 asserts this property.
- **F2 (P1) — task 6.2 is checked but its named artifact does not exist.** No broker spy exists anywhere in
  the repo. The recorded "zero intents / mutation_attempts / risk_reservations" assertion counts rows on
  tables the executed path never writes, so it cannot fail. 6.2 is now unchecked.
- **F3 (P1) — a064's Go guard in `insertExactStrategyDecision` can be deleted with no test failure.** The
  v21 SQL trigger refuses the same inputs and the cited test asserts only `err != nil`, so neither layer's
  removal is observable. Row B2 additionally cites a test whose failure injection fires on a different
  table, never reaching the branch it is cited for.
- **F4 (P1) — a SHALL with no implementation.** `spec.md`'s "snapshot ID 부재 또는 digest 불일치는 신규
  exposure-raising 결정을 거부해야 한다(SHALL)" has no production writer of the lineage pair at all
  (`grep` finds only the struct fields, the insert/compare and tests), and
  `TestStrategyEvidenceReadBoundaryDistinguishesLegacyAndMissing` *pins* the empty pair as acceptable.
  Archiving as written would enter a falsehood into `openspec/specs/`.
- **F5 (P1) — credential leak through transport errors.** `source.go:396` returns a foreign transport error
  verbatim and `:399` flattens it with `%v`. OpenDART authenticates by the `crtfc_key=` **query parameter**
  (the change's own frozen contract says so), and `*url.Error` carries the full URL. Latent only because no
  `Transport` implementation exists in the module. The test credited with proving the opposite still passes
  with both `[REDACTED]` methods deleted.
- **F6 (P1) — "dormant / does not activate a strategy lane" is stale.** `internal/strategyevidence` is now
  imported by six production files (`strategyproposal`, `officialbars` ×2, `continuationlane`,
  `reversallane`, `weeklyvaluelane`). Nothing inside a064 pins the dormancy it claims, so the property
  decayed silently — and it puts `snapshotDigest` (F1) on a production path.

Lower-severity findings — the inert `testdata/official_contracts.json` freeze, existence-checked triggers
and `snapshot_items` constraints, untested `checkSchema` error arms including a064's own B27, non-
discriminating policy-field tests, the server-controlled SEC `file.Name` request path, per-instance rate
budgets, and several factually wrong lines in `analysis/journal-snapshot-verification.md` — are recorded in
`issues.md` with disposition.

### Verdict (2026-09-07, superseded by the remediation pass below)

**a064 is not complete.** The repository-wide gates pass and the re-baseline is sound, but the change
carries claims its tests do not support, one checked task whose artifact was never built, and one SHALL
nothing implements. Tasks 6.2, 6.3 and 6.4 remain unchecked and `make gate` is not run.


## Remediation pass — full fix of F1-F6 and the scope decision (2026-09-07)

The human chose **A-1** (fix every finding, not just the high-risk ones) and **B-1** (narrow the
Requirement to the dormant scope rather than implement the order-entry refusal). What follows is what
changed and how each claim was falsified.

### The method: nothing counts until it has been broken

Every property below was verified the way this repository's record says it must be: the guarded behaviour
was mutated in production code under `go test -overlay`, the suite re-run, and the result recorded. A
positive control accompanies each instrument, because "the detector found nothing" and "there is nothing
to find" are different sentences.

| Finding | Mutation applied | Result |
|---|---|---|
| F1 | item preimage reduced to `EvidenceID + PayloadDigest` (the pre-fix shape) | 22 subtests fail |
| F1 | four query fields dropped from the preimage | 7 subtests fail |
| F1 | `sort.Slice` comparator flipped | 2 tests fail |
| F1 (control) | deliberate type error inside `snapshotDigest` | build fails — the overlay reaches the compiler |
| F3 | a064's whole guard deleted from `insertExactStrategyDecision` | 7 subtests fail |
| F3 | v21 trigger's four format disjuncts removed | 8 subtests fail |
| F5 | (RED) the fix reverted | the returned error carried `crtfc_key=<secret>` on all three paths |
| — | `readonly.go`'s `r.version >= 21` forced true | 2 subtests fail |
| — | a064's own v21 inspection arm (B27) swallowed | its subtest fails |
| — | `consumed_evidence_snapshot_digest` dropped from the read-only column list | its subtest fails |
| — | `validate()`'s seal check deleted | exactly the `contractSeal` subtest fails |
| — | `validate()`'s `RetryableStatuses`/`RetryAfterPolicy` checks deleted | exactly those two subtests fail |
| — | `validateOfficial()`'s SEC endpoint block deleted | exactly six subtests fail |
| — | `Project`'s two reason comparators flipped | the ordering test fails |

### What each finding got

- **F1 — the full-Header digest binding is now measured.** `snapshot_digest_test.go` calls `snapshotDigest`
  directly, bypassing `consumer.go`'s independent revalidation that had been passing the old test for the
  wrong reason. It varies every `Header` field and every `SnapshotQuery` field one at a time, and a
  `reflect`-based completeness check fails if a struct field has no case — so a field added later cannot
  slip out of the preimage unnoticed. The frozen golden vector was produced by a standalone program that
  does not import the production package, because an expected value read out of the running system passes
  whatever the system does.
- **F2 — the broker spy exists and can fail.** Two instruments replace the assertion that counted rows on
  tables nothing writes: a transitive import-closure walk over `internal/strategyevidence` that refuses any
  broker/dispatch/execgw/guardian/journal/order/toggle/`net/http` path (positive control: the same walker
  finds `net/http` in `internal/obs`), and a whole-journal row-count digest taken across the dormant read,
  plus a write attempted directly on the read-only handle. 6.2 is checked again.
- **F3 — the two layers are now distinguishable.** The Go refusal carries
  `ErrStrategyEvidenceReferenceInvalid`; the v21 trigger's four format disjuncts are driven by direct SQL
  that bypasses Go, insert and update, with a canonical/NULL-pair positive control so a `WHEN 0` trigger
  cannot pass. Row B2's citation is corrected to the test that actually reaches the branch.
- **F4 — the Requirement now says what the change does.** Per the human's B-1 decision the sentence
  "snapshot ID 부재 또는 digest 불일치는 신규 exposure-raising 결정을 거부해야 한다(SHALL)" is replaced by
  two SHALLs this change implements and measures — refuse a partial or malformed reference before the
  storage layer, and return typed unavailable from the dormant read with no journal-payload fallback — and
  the Requirement now states in the open that forcing the reference as an order-entry precondition is the
  lane-wiring change's scope. Two scenarios replace the one that named a refusal code no `RefusalCode`
  declares. This is a Requirement-level edit, so it re-triggered the review gate; that is what this section
  records.
- **F5 — the credential cannot leave the adapter.** `redactTransportError` keeps the sentinel class and
  discards the foreign error's message, and the retry-exhausted path uses `%w` instead of `%v` so callers
  can still classify it. A second test asserts the redaction did not simply erase the message.
- **F6 — dormancy is pinned instead of asserted.** The import-closure test is the pin: the package's
  transitive closure inside this module is `internal/clock` and nothing else. It is now a test that fails
  when that stops being true, rather than a sentence that decayed silently for a month.

### Production defects fixed alongside them

- A spoofed SEC response can no longer choose the next request's path (`validSECHistoricalPageName`).
  The rule accepts exactly the frozen form, so it refuses no normal input;
  `TestSECHistoricalPageResourceAcceptsTheFrozenFixtureShape` is the positive control.
- `SourcePolicy.validate` bounds every numeric field against the frozen official contracts instead of
  checking `> 0`, and checks `PageSize`, which it did not check at all.
- `NewAdapter` is unexported. `NewOfficialAdapter`, which requires a `*SharedRateBudget`, is now the only
  way to obtain an adapter from outside the package, so "one call budget per official contract" is a
  type-level property rather than a convention. The exported constructor had no caller in the module.
- A declared fatal fact whose evidence is missing is recorded in `FatalAssessment.Unavailable` even when it
  was optional. `Blocked` and `Reasons` are untouched, so nothing that reads the old two fields changes.

### Evidence quality work

`testdata/official_contracts.json` is read by Go for the first time and now binds the contract ids,
endpoints, schemas, the SEC fair-access rate, the declared-identity requirement, the OpenDART credential
parameter and the KRX freeze flag. The two identifier-grep "SELECT-only" proofs are replaced by
intra-package call-chain reachability with normalized SQL matching; the `schemaV21` literal grep is
replaced by a measured v20→v21 schema diff; `snapshot_items` constraints are enforced by insertion rather
than counted. `ReadOnly.checkSchema`'s six inspection-failure arms — including a064's own B27, which had
never executed — run against a delegating driver that fails one chosen query while a real journal answers
the rest.

### Function Logic Maps

Fourteen bundles sit under `analysis/function-logic/`. Eleven are required by the gate at the current
base: `insertExactStrategyDecision` (refreshed), `Adapter.Fetch`, `SourcePolicy.validate`,
`OfficialAdapter.collectSEC`, `NewOfficialAdapter`, `Project`, `NewAdapter` (at `revision: base`, since
that name no longer exists at HEAD), and the five tests in `source_test.go` that the constructor rename
touched. Two more — `snapshotDigest` and `ReadOnly.checkSchema` — are carried from the original
implementation; their sources are unmodified in this pass, but their branch-test-maps were rewritten
because measurement disproved the coverage they claimed. I27 is therefore no longer true: the logic-map
step is not vacuous for a064.

### What is still not claimed

- No production code writes `ConsumedEvidenceSnapshotID`/`Digest`, and this change does not add one. That
  is stated in the Requirement rather than implied away.
- The credential-leak fix is verified against a synthetic `*url.Error` shaped like `net/http`'s, because
  the module still contains no `Transport` implementation. That is the same latency the finding described.
- The transcribed PASS verdicts and the "218 symbols" count from the original review remain unre-derivable
  prose (issues.md I26). The mutation table in this section is re-runnable.


## Second independent adversarial review and its fixes (2026-09-07)

A second review ran in a separate context after the remediation above, with the working tree verified
unchanged before and after. It returned **BLOCK**, and it was right about the thing that matters most:
the remediation bound one half of the pair issues.md I1 diagnosed and left the other half untouched.

### What it found, and what each finding got

| # | Finding | Confirmed by | Fix | Re-falsified |
|---|---|---|---|---|
| S1 | `snapshotItemMatchesQuery` — the unbound half of I1's pair — survives `if true { return true }` | full package suite green | `TestSnapshotItemScopeIsCheckedIndependentlyOfTheDigest`, which supplies a *correct* digest so only the scope check can refuse | the same mutation now fails 8 subtests; an in-scope positive control proves it refuses no normal input |
| S2 | the v21 trigger's `digest != lower(digest)` disjunct can never decide | full journal suite green with it deleted | not deleted — a landed migration's SQL is not rewritten for a redundancy. The claim is corrected: four of five disjuncts are decisive, and the test now labels the one that actually refuses | n/a, and the Go-side lowercase check (decisive, because `hex.DecodeString` accepts uppercase) stays bound |
| S3 | the read-side `validConsumedEvidenceReference` call is a third, unbound judgement | full journal suite green | `TestStrategyEvidenceReadBoundaryRefusesAMalformedStoredReference`, writing the malformed value past the v21 triggers | 4 subtests fail |
| S4 | the `RequestDeadline` contract cap and the pre-existing ordering check cover for each other | either alone survives; both together fail | the cap case is 2m against a 5m operation deadline; a separate case violates the ordering rule alone | deleting the cap fails only `RequestDeadline/1`; deleting the ordering conjunct fails only `RequestDeadline/2` |
| S5 | task 6.2's read-handle write assertion is refused by a trigger, not by the handle | passes with `mode=ro` and `query_only` removed | the write now targets `CREATE TABLE` and `PRAGMA user_version`, which no trigger guards | the DSN mutation fails this test |
| S6 | the lineage INSERT's failure arm is undiscriminated | `_ = err` leaves the suite green | `TestStrategyDecisionLineageInsertFailureSurfacesTheStatementError` asserts the abort surfaces and is not a collision | the mutation fails it |
| S7 | the declared-identity test used a blank string a broader check already refuses | deleting `validSECRequestIdentity` leaves it green | `"anonymous-bot"`, plus a separate blank case | 2 tests fail |
| A1 | two refreshed FLM bundles still asserted PASS on evidence measurement disproved | both files unchanged from HEAD | `snapshotDigest` and `ReadOnly.checkSchema` branch-test-maps rewritten to cite the tests that reach each branch, with the measured RED | n/a |
| A2-A4 | bundle count wrong, a dangling table reference, no standing verdict, stale gate results | reading | corrected below | n/a |

Two of the reviewer's observations are recorded rather than fixed, and issues.md says why: the policy
caps are not all receipt-derived (only `maxOfficialPageSize` is, and the code comment now says so, naming
the legitimate configuration the others would refuse), and the frozen SEC fixture may be synthetic rather
than captured — which is the source-wiring change's problem, not this one's.

The reviewer also noted a methodological point worth keeping: the two AST tests parse from **disk**, so
`go test -overlay` is invisible to them. They were falsified on a disk copy instead, and all three
injections (an added `net/http` import, a `DELETE` literal in `snapshotItemMatchesQuery`, an `UPDATE`
literal in `validConsumedEvidenceReference`) failed correctly.

### Repository-wide gates, re-run after every change above

| Command | Result |
|---|---|
| `openspec validate a064-add-multi-market-strategy-evidence --strict --no-interactive` | PASS |
| `python3 tools/logic-map/check_analysis.py --change a064-…` | PASS — evidence complete or diff-proven exempt |
| `make sdd-sync` | all indexes current (GBrain advisory busy; previous freshness kept) |
| `make sdd-check` | PASS |
| `make test` | PASS — 99 packages ok, 0 FAIL |
| `make test-seams` | PASS — 46 packages ok, 0 FAIL |
| `make test-race` | PASS — 8 packages ok, 0 data races |
| `make vet` | PASS |
| `make lint` | PASS — `go vet ./...` and `go vet -tags tossos_testseams ./...` |
| `make validate` | PASS — 61 items, 0 failed |

### Verdict (2026-09-07, standing)

**a064 is complete.** Every finding from both independent reviews is closed or recorded with its reason,
and every safety property this change claims is now enforced by a test that has been observed to fail when
the property is broken — with a positive control proving the instrument is not blind. What the change does
*not* do is stated in the Requirement rather than implied away: no production code writes a consumed
evidence snapshot reference, and forcing one as an order-entry precondition remains the lane-wiring
change's scope.

Two residual limits, both named where they matter: the credential-leak fix is verified against a synthetic
`*url.Error` because the module still contains no `Transport` implementation, and eight of the nine policy
caps are conservative ceilings rather than receipt-derived numbers.

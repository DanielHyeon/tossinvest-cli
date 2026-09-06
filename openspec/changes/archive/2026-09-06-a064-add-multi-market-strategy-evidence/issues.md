# Issues — a064-add-multi-market-strategy-evidence

Raised by the 2026-09-07 completion pass: one adversarial engineering review and one evidence-quality audit
that read all 34 branch-test-map rows, both run in contexts separate from the implementation. Severities are
this gate's, not CVSS. `REPRODUCED` means this session re-ran the check itself rather than accepting the
report; `REPORTED` means it is recorded from a review and not independently re-run here.

Disposition vocabulary: **OPEN** (blocks completion), **FIXED** (done in this pass), **DECISION** (needs a
human scope call), **NOTE** (recorded, not blocking).

---

## Blocking — untested High-risk properties

### I1 · P0 · FIXED · REPRODUCED — the full-Header digest binding is enforced by no test
`internal/strategyevidence/store.go:381-402` (`snapshotDigest`)

Reducing the per-item preimage to `EvidenceID + PayloadDigest` — the exact pre-fix shape that review.md's
"Independent HIGH integrity review closure" says it rejected — leaves `go test -count=1 ./internal/strategyevidence`
**green**. Two further mutations survive: dropping `SupersedesRevisionIdentity`/`Unit`/`Availability`/`Confidence`,
and dropping `IssuerIdentity`/`IssuerMappingVersion`/`EvaluationAt`/`IngestionCutoff` from the query half.

Instrument verified before the conclusion was drawn: a deliberate type error inserted inside `snapshotDigest`
fails the build (the `-overlay` reaches the compiler), and neutralising `e.ingested_at<=?` in the as-of
selection fails `TestStoreStampsTrustedIngestionClockAndRejectsBackdating`,
`TestStoreDualCutoffAndAppendOnlyCorrectionSnapshot` and `TestStoreDualCutoffNanosecondBoundaries` (the suite
can catch a mutation in this file).

Root cause is this repo's recorded two-judgements pattern: `internal/strategyevidence/consumer.go:74-84`
(`snapshotItemMatchesQuery`) independently revalidates market, symbol, issuer, mapping version, source times
and effective date — exactly the six tamper cases the cited RED covered — and returns `ErrSnapshotUnavailable`
from inside the row loop, before `consumer.go:68` ever recomputes the digest. So
`TestDormantSnapshotReadRejectsTamperedHeaderScopeAndCutoffs` passes for the wrong reason. `SealSnapshot`,
`verifySnapshot` and `Replay` all call the same digest function, so a weakened preimage stays self-consistent.
There is no frozen digest vector either, unlike the sibling `SealBarSeries` (`breakout_series_test.go:938-945`).

Branch-test-map row B3 asserts precisely this property. It is not backed.

**Fix:** per-field RED assertions that a single Header field change moves the digest, plus a committed golden
digest vector. Note the golden must be paired with a test that calls the producer — see I11.

### I2 · P1 · FIXED · REPRODUCED — `snapshotDigest` ordering comparator never executes
`internal/strategyevidence/store.go:387`

No test anywhere seals a snapshot with two or more items, so `sort.Slice`'s comparator is never invoked;
flipping `<` to `>` is undetectable, and "does not mutate the caller slice" is never asserted. Every
in-package assertion passes 0 or 1 expected ID, and every out-of-package `SealSnapshot` caller
(`strategyproposal`, `officialbars`, `continuationlane`, `reversallane`, `weeklyvaluelane` tests) appends
exactly one envelope. Row B2 claims coverage.

### I3 · P1 · FIXED · REPRODUCED — `snapshotDigest` query-field binding unmeasured
The "deterministic snapshot tests" assert digest *stability* only — the trivially-passing direction. No test
holds the item set constant and varies `query.Symbol` / `IssuerIdentity` / `IssuerMappingVersion`. Row B1
claims coverage.

### I4 · P1 · FIXED · REPRODUCED — a064's Go guard in `insertExactStrategyDecision` is deletable in silence
`internal/journal/strategy_lineage.go:422-424`

Deleting a064's entire edit to that function and running the cited tests passes: the v21 SQL trigger refuses
the same six inputs, and `TestStrategyEvidenceLineageRejectsPartialOrMalformedReference`
(`strategy_evidence_test.go:63-91`) asserts only `err != nil` plus zero rows, never which layer refused. Row
B1 claims "rejection before SQL" and cites that test.

### I5 · P1 · FIXED · REPRODUCED — row B2 of the same map cites a test that cannot reach the branch
`TestStrategyProductionIssuanceFailureRollsBackReservationAndAllAuthority`
(`internal/journal/strategy_lineage_test.go:357`) injects failure with a trigger on
`strategy_attempt_lineage`. The decision-lineage `INSERT` at `strategy_lineage.go:426` succeeds; the branch
at `:431-433` is never taken. A test that would reach it exists and is not cited:
`internal/journal/strategy_first_leg_atomic_test.go:267`.

### I6 · P1 · FIXED · REPRODUCED — `ReadOnly.checkSchema` error arms have zero coverage, including a064's own B27
`internal/journal/readonly.go:230-232, 246-247, 263-264, 282-283, 293-294, 310-311`

All 16 `OpenReadOnly(` call sites in `internal/journal/*_test.go` supply a well-formed SQLite file, and
`OpenReadOnly` stats and pings before `checkSchema`, so no driver error can arise. Rows B1, B6, B11, B17, B21
and **B27** are marked PASS on "existing failure coverage". B27 is the `case err != nil` arm a064 itself
added.

### I7 · P1 · FIXED · REPRODUCED — the v21 version gate is never observed to skip
No test opens a complete v20-or-older journal read-only, so `readonly.go:302`'s false arm (row B23) is never
taken; deleting `if r.version >= 21` breaks nothing. Same shape at `readonly.go:274` (row B13).

### I8 · P2 · FIXED · REPRODUCED — "both columns are inspected" matches on either
`TestOpenReadOnlyRejectsDamagedV21EvidenceLineage` asserts
`strings.Contains(err.Error(), "consumed_evidence_snapshot")`, which matches either column name. Deleting
`consumed_evidence_snapshot_digest` from `internal/journal/strategy_evidence.go:38-39` still passes. Row B24.

### I9 · P2 · FIXED · REPRODUCED — three of four v21 trigger predicate disjuncts are untested
`internal/journal/strategy_evidence.go:20-23, 28-31`. The only behavioural trigger test inserts `(id, NULL)`,
exercising the NULL-parity disjunct alone. The format disjuncts (`'snapshot-'||digest`, `length!=64`,
`!=lower(...)`, `GLOB '*[^0-9a-f]*'`) are unreachable from Go because the guard at `strategy_lineage.go:422`
refuses first, and no test issues the direct SQL that would reach them. Deleting all four is undetectable.
Compounded by `strategy_evidence_migration_test.go:47-52`, which asserts the triggers exist **by name** — a
trigger gutted to `WHEN 0` still counts 1.

---

## Blocking — claims whose artifact does not exist

### I10 · P1 · FIXED · REPRODUCED — task 6.2's broker spy was never built
`tasks.md` 6.2 claimed "a broker spy test proving all a064 paths create zero order intents, zero live broker
requests and zero lane/automation toggle changes". No `brokerSpy`/`BrokerSpy` exists anywhere in the repo.
The evidence actually present is `internal/journal/strategy_evidence_test.go:33-38`, three `SELECT count(*)`
assertions on `intents`, `mutation_attempts` and `risk_reservations` — tables no path the test executes ever
writes, on a freshly created journal. The assertion holds for every possible mutation of the code under test.

The underlying property does hold today (no a064 file imports `net/http`, any broker, dispatch, execgw,
guardian, runtime or toggle package), but the recorded evidence names something never built. 6.2 is now
unchecked; it needs a spy that can actually fail.

### I11 · P2 · FIXED · REPORTED — the "frozen" official contract file is read by nothing
`internal/strategyevidence/testdata/official_contracts.json` has zero Go readers and no `//go:embed`
(`grep -rn official_contracts --include='*.go' .` → empty). The values it freezes are re-encoded as Go
constants at `source.go:72-73, 152, 156` and asserted against themselves. Changing the JSON breaks nothing;
changing the constant is not checked against the JSON. tasks.md 1.3 and review.md call this a freeze.

### I12 · P2 · FIXED · REPORTED — structural "SELECT-only" proofs are identifier greps of one method
`internal/strategyevidence/consumer_static_test.go:12-55` and
`internal/journal/strategy_evidence_static_test.go:12-55` each parse one file, walk one method body, and
reject import substrings, `Exec|ExecContext|Begin|BeginTx` selectors and SQL literals containing
`"INSERT "` etc. (each with a trailing space). A write reached through a same-package helper, a differently
named method, a `driver.Conn`, or `DELETE\nFROM` passes cleanly. These carry the change's strongest safety
claims.

### I13 · P2 · FIXED · REPORTED — `snapshot_items` constraints are counted, not enforced
`internal/strategyevidence/store_test.go:270-303` asserts the primary key has 2 columns and
`pragma_foreign_key_list` returns 2 rows. It never checks which columns, which referenced tables, that a
duplicate `(snapshot_id, evidence_id)` is rejected, or that an orphan `evidence_id` is rejected. review.md
promotes this to "Runtime PRAGMA tests prove …".

### I14 · P2 · FIXED · REPORTED — v21 migration isolation is a substring grep of a Go constant
`internal/journal/strategy_evidence_migration_test.go:53-58` greps the `schemaV21` string literal for
`payload`, `revision`, `credential`, `secret`, `source_response`, `create table`. Task 3.3 asks to test that
such columns/tables are *absent*; inspecting a literal in the test's own package cannot see a table
introduced by another migration, and `CREATE  TABLE` (double space) evades it.

---

## Blocking — production defects

### I15 · P1 · FIXED · REPRODUCED — credential leak through transport errors
`internal/strategyevidence/source.go:396` returns a foreign transport error verbatim; `:399` flattens it with
`%v` rather than `%w`, making it unrecoverable downstream. OpenDART authenticates by the `crtfc_key=`
**query parameter** — stated by the change's own `testdata/official_contracts.json`
(`"credential_parameter": "crtfc_key"`) — and `net/http` bakes the full URL, query included, into
`*url.Error` (`stripPassword` redacts userinfo only). Both the deadline path and the retry-exhausted path
leak; the second irreversibly. This violates the spec's MUST NOT on credentials in logs.

Latent today only because the module contains no `Transport` implementation. The test credited with proving
the opposite (`official_source_test.go:78-102`) still passes with both `[REDACTED]` `String()` methods
deleted — it asserts on `json.Marshal(batch)`, and `OfficialBatch` has no field a credential can reach.

### I16 · P1 · FIXED · REPRODUCED — server-controlled request path (existence check where a role check belongs)
`internal/strategyevidence/official_source.go:187` guards only `strings.TrimSpace(file.Name) == ""`, then
passes `file.Name` — read out of the remote SEC response body — as `Resource` into `fetchPage` →
`TransportRequest.Resource` → `transport.Do` (`source.go:392`, no validation). Real values match
`CIK\d{10}-submissions-\d{3}\.json`; nothing enforces that. A spoofed upstream response steers the next
request's path, which is the spec's "비공식 endpoint를 fallback으로 사용해서는 안 된다" prohibition.

### I17 · P2 · FIXED · REPRODUCED — policy bounds are positivity-only and the rate ceiling is per-instance
`source.go:181` checks only `> 0` for the eight numeric bounds and omits `PageSize` entirely;
`MintSourcePolicy` adds real caps for SEC only (`source.go:102`), the OpenDART branch checking `ContractID`
and `0 < PageSize <= 100`. Separately `NewAdapter` (`source.go:304`) is exported, validates nothing and
leaves `shared == nil`, so `acquire`/`consumeCall` fall back to per-`Adapter` counters — N adapters give
N × the policy rate. Nothing pins a process-wide budget.

### I18 · P2 · FIXED · REPRODUCED — an optional fatal fact that is missing leaves no trace
`internal/strategyevidence/projection.go:66-70`: a `policy.Fatal` requirement with `Required == false` whose
evidence is missing hits `continue` and is recorded in neither `Fatal.Reasons` nor any `Unavailable` map —
the lane loop has one, the fatal loop does not. A declared fatal fact disappears silently.

### I19 · P2 · FIXED · REPRODUCED — the policy zero-call test is non-discriminating for three of its four cases
`internal/strategyevidence/source_test.go:13-40` mutates an already-minted, sealed policy, so the mutation
breaks `contractSeal` and both the field check and the seal check return `ErrSourceDisabled`. Deleting the
`EndpointVersion`, `AccessContract == "official"` and `len(RetryableStatuses) != 0` checks from `validate()`
leaves all four subtests passing. No zero-call test exists for `AbsoluteCallWindow`, `MaxCalls`, `MaxPages`,
`MaxResponseBytes`, `MaxConcurrency`, `RequestDeadline`, `OperationDeadline`, `MaxRetries`,
`RetryAfterPolicy`, `Version` or `CredentialRequired`.

The KRX zero-call claim is partly backed — `source_test.go:23` does hold a spy and asserts
`transport.Calls() == 0` — but for `ContractVerified=false`, not for frozen-contract absence, which is what
`official_source_test.go:41-45` covers with no transport in scope. The mechanism itself is sound (four
independent hardcodes; `SourcePolicyConfig` has no endpoint fields, so no configuration can enable KRX).

---

## Requires a human scope decision

### I20 · P1 · DECIDED · REPRODUCED — a SHALL that nothing implements, and a test that pins the opposite
`specs/multi-market-strategy-evidence/spec.md`, Requirement "Evidence payload 권위는 trading journal과
분리된다": *"snapshot ID 부재 또는 digest 불일치는 신규 exposure-raising 결정을 거부해야 한다(SHALL)"*, with
the scenario "신규 진입은 `EVIDENCE_SNAPSHOT_UNAVAILABLE`로 거부".

Measured: no production code writes `ConsumedEvidenceSnapshotID`/`Digest` at all — `grep` over non-test Go
finds only the struct fields (`strategy_lineage.go:110-111`), the insert/read/compare (`:422, :428, :438,
:443`) and tests. Every production strategy decision written today carries NULL/NULL, which
`validConsumedEvidenceReference` (`strategy_evidence.go:88-91`) explicitly accepts and
`TestStrategyEvidenceReadBoundaryDistinguishesLegacyAndMissing` explicitly pins. No entry path consults an
evidence snapshot as a precondition.

Implementing the refusal would put a064 into the order-entry path, which its own proposal excludes as a
non-goal ("이 change는 증거 수집·정규화·재현 계약까지만 제공하며 주문, Guardian 승인, lane 활성화 또는
운영 토글을 수행하지 않는다"). Narrowing the requirement to the dormant scope keeps a falsehood out of
`openspec/specs/` but removes the sentence that made the lineage safety-meaningful, and — per
`docs/WORKFLOW.md`'s graded review gate — a Requirement-level edit re-triggers the gstack review.

Tasks 3.3, 5.2 and 5.3 are all marked `[x]`.

---

## Corrected in this pass

### I21 · FIXED — the re-baseline precedent was misattributed
Commit `875f8be4` re-baselined `a064-critical-events-reach-the-operator` (now `a074`) and five siblings, not
this change; the names collide across the merge. review.md now states this and records the measured
narrowing (338 required functions at the old base → 0 at the new one).

### I22 · FIXED — `analysis/journal-snapshot-verification.md` said "two" modified existing functions
There are three, and three bundles exist.

### I23 · FIXED — the same file said `schemaV21` "contains only two `ALTER TABLE … ADD COLUMN … TEXT`
statements"
It contains two `ALTER TABLE` **and two `CREATE TRIGGER`** statements
(`internal/journal/strategy_evidence.go:14-33`). review.md described the triggers, so the two documents
disagreed; the isolation argument rested on the wrong description.

### I24 · FIXED — `snapshotDigest`'s map called its AST a pre-edit capture
`git show 23794f86:internal/strategyevidence/store.go | sha256sum` equals the hash the committed `ast.json`
carried, so the artifact was always post-edit. The true pre-edit hash `5134fd3a…` survives only in
`analysis/pre-edit/snapshot-digest-integrity.md`.

### I25 · FIXED — "dormant … does not activate a strategy lane" was stale
`internal/strategyevidence` is imported by six production files today: `internal/strategyproposal/production.go:27`,
`internal/officialbars/producer.go:24`, `internal/officialbars/quote.go:31`,
`internal/continuationlane/production_proposal.go:10`, `internal/reversallane/production_proposal.go:10`,
`internal/weeklyvaluelane/production_proposal.go:10`. Nothing inside a064 pins the dormancy it claimed, so
the property decayed silently — and it places I1 on a production path. review.md and status.md now say so.

---

## Recorded, not blocking

### I26 · NOTE — PASS verdicts and the "218 symbols" impact count are transcribed prose
review.md and status.md carry hand-copied results (`ok internal/journal 7.068s`, "Final independent semantic
re-review: PASS", "a broad impact surface (218 symbols)") with no stored output file, no `-json` artifact and
no re-derivable command record. The 2026-09-07 gate table in review.md is re-runnable; these earlier lines
are not.

### I27 · NOTE — the gate's logic-map step is now vacuous for a064
A consequence of the re-baseline, stated rather than hidden: `changed_existing_functions` at the new base
returns 0. The step proves the three bundles are complete and current-tree-bound, nothing more. The
pre-edit proof stands only in git history at `23794f86`.

---

## Remediation, 2026-09-07

The human chose full remediation (A-1) and the narrowing branch of the scope decision (B-1). Every
diagnosis above stands as written; what follows is what closed it and how that was proven.

Each row's "falsified by" column names a mutation applied with `go test -overlay` — the guarded property
was broken in production code and the suite was re-run. A row with no mutation says so.

| # | What closed it | Falsified by |
|---|---|---|
| I1 | `snapshot_digest_test.go`: per-`Header`-field and per-`SnapshotQuery`-field binding with reflect completeness, plus a frozen golden authored by a program that does not import the production package | item preimage reduced to `EvidenceID + PayloadDigest` → 22 subtests fail |
| I2 | golden vector holds two items; a reversed input must produce the ascending digest and must not equal the frozen descending one; the caller slice is checked afterwards | `sort.Slice` comparator flipped → 2 tests fail |
| I3 | query half varied one field at a time against a constant item set | four query fields dropped from the preimage → 7 subtests fail |
| I4 | `ErrStrategyEvidenceReferenceInvalid`; `TestStrategyEvidenceGoGuardRefusesBeforeSQL` asserts the Go layer refused | a064's guard deleted → 7 subtests fail (the SQL trigger's error is a different type) |
| I5 | row B2 now cites `TestFirstLegAtomicAdmissionLateStatementFailuresRollbackEveryFamily`, which injects a trigger on `strategy_decision_lineage` and does reach the branch | n/a — citation correction |
| I6 | a delegating `driver.Connector` that passes queries to a real journal and fails one chosen query; all six inspection arms driven, with a no-injection positive control | a064's own v21 arm swallowed → its subtest fails |
| I7 | complete v19 and v20 journals opened read-only | `r.version >= 21` forced true → both subtests fail |
| I8 | half-migrated v21 journals with one column present; the refusal must name the missing one and not the present one | `consumed_evidence_snapshot_digest` dropped from the column list → its subtest fails |
| I9 | direct SQL that bypasses the Go guard, one case per trigger disjunct, insert and update, plus a canonical/NULL-pair positive control | four format disjuncts removed → 8 subtests fail |
| I10 | transitive import-closure walk over the package, and a whole-journal row-count digest across the dormant read plus a write attempt on the read handle; both with positive controls | n/a — the replaced assertion could not fail by construction |
| I11 | `frozen_contract_test.go` decodes the JSON and binds the contract ids, endpoints, schemas, the SEC fair-access rate, the declared-identity requirement, the OpenDART credential parameter and the KRX freeze flag | n/a — the file had no reader; it now has one |
| I12 | intra-package call-chain reachability from `Replay` / `ConsumedSnapshot`, normalized SQL matching, positive control on the sealing and insert paths | n/a — the widened scope is structural |
| I13 | primary-key column names and order, both foreign-key targets by name, and four rejected insertions (duplicate ordinal, duplicate evidence, orphan evidence, orphan snapshot) with a legitimate-insert positive control | n/a |
| I14 | measured v20→v21 schema diff: exactly two columns and two triggers added, nothing removed, no forbidden name, trigger bodies abort and do not write | n/a |
| I15 | `redactTransportError` + `transportFailure`; `%v` became `%w` over the redacted class | RED: the raw `*url.Error` carried `crtfc_key=<secret>` in all three paths |
| I16 | `validSECHistoricalPageName` pins `CIK<requested CIK>-submissions-<3 digits>.json` | RED: five spoofed names were all fetched |
| I17 | contract-derived caps in `validate()` including `PageSize`; `NewAdapter` unexported so `NewOfficialAdapter` (which requires a budget) is the only external door | RED: nine bound subtests; the rename is a structural closure with no behaviour change outside the package |
| I18 | `FatalAssessment.Unavailable`, filled for every fatal refusal; `Blocked`/`Reasons` semantics untouched | RED: the field did not exist |
| I19 | one case per `SourcePolicy` field with the seal re-stamped after each mutation, reflect completeness over the struct | seal check deleted → `contractSeal` subtest fails; `RetryableStatuses`/`RetryAfterPolicy` deleted → exactly those two fail; the SEC endpoint block deleted → exactly six fail |
| I20 | the Requirement now states what this change implements — refuse a partial/malformed reference before storage, and return typed unavailable from the dormant read with no payload fallback — and says in the open that forcing the reference as an order-entry precondition belongs to the lane-wiring change | n/a — human scope decision |
| I26 | still true. The 2026-09-07 and remediation results in review.md are re-runnable; the older transcribed lines are not, and are marked as such. | n/a |
| I27 | no longer true. The logic-map step now covers eleven modified functions, six of them production. | n/a |

Two branches nobody had asked about surfaced while writing I18's map and are closed with it: `Project`'s
scope-refusal early return (every existing test supplied a valid scope) and its two reason-ordering
comparators (no test ever produced two reasons). Both are now driven directly and both comparators fail
the suite when flipped.

---

## Second independent adversarial review, 2026-09-07 (post-remediation)

Run in a separate context after the remediation above, with the working tree verified unchanged before
and after. It returned **BLOCK** with four findings of the same shape the first review named, and it was
right: the remediation bound one half of the pair issues.md I1 diagnosed and left the other half exactly
as it was. Every finding below was confirmed by mutation by the reviewer, and the fix for each was
re-falsified here.

### S1 · P0 · FIXED — `snapshotItemMatchesQuery` was the unbound half of I1's pair
`internal/strategyevidence/consumer.go:74-84`. `if true { return true }` left the whole package green.
This is the only in-code enforcement of the second Requirement's dual-cutoff and effective-date rules, and
it appeared in no branch-test-map row. Fixed by `TestSnapshotItemScopeIsCheckedIndependentlyOfTheDigest`,
which computes the digest **over the tampered item set on purpose** so `Snapshot.Valid` can only refuse in
the scope check. Falsified: the same mutation now fails 8 subtests, and an in-scope positive control
proves the check does not refuse normal input.

### S2 · P2 · RECORDED — the v21 trigger's `digest != lower(digest)` disjunct can never decide
It is strictly implied by `GLOB '*[^0-9a-f]*'`: no ASCII character differs from its own lowercase while
lying inside `[0-9a-f]`. The `uppercase-digest` case is refused by the GLOB, so removing the lower()
disjunct from both triggers leaves the full journal suite green. Not fixed by deleting it — changing a
landed migration's SQL would make freshly-migrated journals differ from already-migrated ones for no
behavioural gain. The claim is corrected instead: four of five disjuncts are decisive, and the test's
`disjunct` label now names the one that actually refuses. The **Go-side** lowercase check is decisive
(`hex.DecodeString` accepts uppercase) and is bound.

### S3 · P1 · FIXED — the read-side reference guard was a third, unbound judgement
`internal/journal/strategy_evidence.go:88`. Dropping `validConsumedEvidenceReference` from
`ConsumedSnapshot` left the full journal suite green; the remaining `SnapshotID == ""` clause covers only
the legacy case. This call is the sole implementation of the narrowed Requirement's read-side SHALL for a
*malformed* stored reference. Fixed by
`TestStrategyEvidenceReadBoundaryRefusesAMalformedStoredReference`, which writes the malformed value past
the v21 triggers — the situation the branch actually guards, a row written before the triggers existed or
a hand-edited file. Falsified: 4 subtests now fail.

### S4 · P1 · FIXED — a contract cap and the pre-existing ordering check covered for each other
`TestEveryPolicyFieldHasAZeroCallRefusal`'s `RequestDeadline` over-cap case used one hour against a
five-second operation deadline, so `RequestDeadline > OperationDeadline` refused first and either check
could be deleted alone. Fixed: the cap case is now 2m against a 5m operation deadline (inside the ordering
rule, outside the cap), and a separate case violates the ordering rule alone. Falsified: deleting the cap
fails only `RequestDeadline/1`; deleting the ordering conjunct fails only `RequestDeadline/2`.

### S5 · P1 · FIXED — task 6.2's read-handle write assertion was non-discriminating
`UPDATE strategy_decision_lineage SET market='XX'` is refused by the `..._no_update` trigger, not by the
read-only handle, so removing `mode=ro` and `query_only(true)` from the DSN left the assertion green.
Fixed by writing where no trigger guards: `CREATE TABLE` and `PRAGMA user_version`. Falsified: the DSN
mutation now fails this test (and the pre-existing `TestOpenReadOnlyRefusesToWriteAnything`).

### S6 · P2 · FIXED — the lineage INSERT's failure arm was undiscriminated
Replacing `if err != nil { return 0, err }` with `_ = err` left the journal suite green: the following
read-back finds no row and returns `StrategyCollisionError`, which the rollback assertions accept. Fixed
by `TestStrategyDecisionLineageInsertFailureSurfacesTheStatementError`, which asserts the abort message
surfaces and that the error is *not* a collision.

### S7 · P2 · FIXED — the declared-identity test used an input a broader check already refused
`RequestIdentity = ""` is refused by the generic non-blank check, so deleting `validSECRequestIdentity`
left it green. Fixed: the case now uses `"anonymous-bot"`, and a separate case covers the blank string.

### A1 · P1 · FIXED — two of the refreshed FLM bundles still asserted PASS on absent evidence
`snapshotdigest` and `readonly-.checkschema` branch-test-maps were unchanged from HEAD. Rewritten: all
three `snapshotDigest` rows and the twelve `checkSchema` rows the reviewer measured now cite the tests
that reach them, and each RED column records what the mutation actually did.

### A2 · FIXED — review.md's bundle count was wrong (eleven claimed, twelve named, fourteen on disk).
### A3 · FIXED — review.md referred to a gate table that did not exist and carried no standing verdict.
### A4 · FIXED — the repository-wide gates are re-run after this pass, not carried over from before it.

### Also recorded, not defects

- The nine policy caps are not all from a receipt. Only `maxOfficialPageSize` is bound to
  `official_contracts.json`; the other eight are conservative ceilings above what this module's two
  minted policies use. That means they *do* refuse some legitimate configuration — OpenDART's documented
  daily quota is 20,000 requests while `maxOfficialCalls` allows 1000 per 24h window — which is latent
  because `MintSourcePolicy` has no production caller. The code comment now says exactly this instead of
  claiming the caps come from the frozen contract.
- The reviewer suspects the live SEC submissions API returns an unpadded top-level `cik` where the frozen
  fixture is padded. If so the pre-existing `root.CIK != request.EntityID` check already fails closed
  against the real endpoint and the fixture is synthetic rather than captured. Unverified here (no
  network) and outside this change's scope, but it weakens the word "frozen" and is recorded for the
  source-wiring change.

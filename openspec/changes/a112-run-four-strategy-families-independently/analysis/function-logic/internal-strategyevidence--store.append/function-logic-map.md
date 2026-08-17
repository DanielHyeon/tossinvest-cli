# Function Logic Map: `Store.Append`

- Source: `internal/strategyevidence/store.go`
- Source SHA-256: `53de9bb53980c7e3be959b1ab51b03667c91c232d04c50bb1724426e837910ca` (current worktree; verified with `sha256sum` 2026-08-17, equal to `source_sha256` in `ast.json`)
- Signature: `(s *Store) Append(ctx context.Context, evidence Envelope) (AppendResult, error)` (`ast.json`: `Store.Append(params=2, results=2)`)
- Source range: `194:1`–`263:2`
- AST counts: branches 17, returns 15, calls 40, defers 1, go statements 0 (`ast.json` generated 2026-08-17 by `go run ./tools/logic-map`).
- Risk scan: `risk-pattern-report.md`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

## Inputs and invariants

- The single write door of the evidence ledger. Everything a112 mints — including the closed 1m bars the L1b producer will build — reaches SQLite through this function and no other. The tables are append-only by trigger: `evidence_no_update`, `evidence_no_delete`, `conflict_no_update`, `conflict_no_delete` (store.go:181–184), pinned by `TestStoreSchemaTooNewAndAppendOnlyTriggers`.
- Construction guard: `evidence.PayloadDigest() == ""` is refused with `RefusalPayloadInvalid`, detail "use NewEnvelope" (195–197). An `Envelope` can only get a digest through `NewEnvelope`, so this rejects a zero value that skipped canonicalisation.
- Clock authority: `trustedIngestedAt := s.clk.Now().UTC()` (198) is the only accepted ingestion time. A caller-supplied `IngestedAt` is overwritten at 203 and the envelope is rebuilt through `NewEnvelope` at 204, so the row persisted is the re-stamped one. A store clock earlier than `ObservedAt` is refused with `RefusalTimestampInvalid` (200–202) rather than silently accepted.
- Identity: `readByIdentity(ctx, tx, h.Authority, h.SourceRecordID, h.RevisionIdentity)` at 215 is the uniqueness probe — authority plus source record plus revision, not the evidence id.
- Correction rule: when `SupersedesRevisionIdentity` is non-empty the superseded revision must already exist and must agree on market, symbol, issuer identity, issuer mapping version and kind (237–244). A revision may correct values; it may not move the evidence to another scope.
- Everything runs in one transaction opened at 209 with `defer func() { _ = tx.Rollback() }()` at 213, so every refusal path leaves the database untouched; the two success paths and the quarantine path commit explicitly.

## Branches and early returns

Exact AST return nodes: `196, 201, 206, 211, 217, 221, 230, 233, 235, 239, 241, 243, 257, 260, 262`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 195:2 | no payload digest → `RefusalPayloadInvalid` ("use NewEnvelope") | not-applicable: defensive constructor guard — every test builds envelopes through `NewEnvelope`, so a digest-less value never reaches this call |
| B2 | if | 200:2 | trusted store clock is before `ObservedAt` → `RefusalTimestampInvalid` | not-applicable for the taken side: no test sets the store clock behind a header's `ObservedAt`; the untaken side (clock ahead, caller backdating overwritten) is pinned by `TestStoreStampsTrustedIngestionClockAndRejectsBackdating` |
| B3 | if | 205:2 | rebuilding the envelope with the trusted `IngestedAt` failed → return the header refusal | not-applicable: unreachable in practice — the only field changed is `IngestedAt`, whose constraints are non-zero (model.go:172) and not before `ObservedAt` (model.go:174); the latter is exactly B2, so only a zero-valued store clock could reach it |
| B4 | if | 210:2 | `s.db.BeginTx` failed → return | not-applicable: driver failure. `TestOpenReadOnlyReplaysButCannotCreateSnapshotOrAppend` proves a read-only handle refuses the append, but asserts only that an error is returned and does not distinguish B4 from B16 |
| B5 | if | 216:2 | the identity probe failed → return | not-applicable: SQLite driver failure |
| B6 | if | 219:2 | the identity already has a row | taken: `TestStoreAppendIsIdempotentAndQuarantinesRevisionConflict`, `TestRevisionConflictUsesInjectedClock`, `TestStoreSameDigestWithChangedProvenanceIsConflict`, `TestClosedBarAppendQuarantinesSameRevisionWithADifferentDigest`; untaken (fresh identity): `TestStoreDualCutoffAndAppendOnlyCorrectionSnapshot` |
| B7 | if | 220:3 | same digest AND `sameImmutableProvenance` → return the existing evidence with `Idempotent: true` and commit | taken: `TestStoreAppendIsIdempotentAndQuarantinesRevisionConflict` (a whitespace-only re-encoding of the same payload returns the first digest, `Idempotent` true); untaken: `TestStoreSameDigestWithChangedProvenanceIsConflict` (identical digest, drifted `EvidenceID`/`SourceAvailableAt`/`ObservedAt` → conflict, not idempotence) |
| B8 | if | 229:3 | the `INSERT OR IGNORE INTO revision_conflicts` failed → return | not-applicable: driver failure; the statement is `OR IGNORE`, so a repeated quarantine is not an error |
| B9 | if | 232:3 | committing the quarantine failed → return | not-applicable: driver failure |
| B10 | if | 237:2 | this revision claims to supersede another | taken: `TestStoreDualCutoffAndAppendOnlyCorrectionSnapshot`, `TestStoreRejectsCorrectionThatChangesEvidenceScope`, `TestSealBarSeriesRefusesEveryHeaderPayloadDrift`; untaken: `TestStoreAppendIsIdempotentAndQuarantinesRevisionConflict` |
| B11 | if | 238:3 | the superseded-revision lookup failed → return | not-applicable: driver failure |
| B12 | else | 240:10 | the else arm of the supersedes lookup | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift`, case "correction references a revision that was never stored" |
| B13 | if | 240:10 | `!found` — the referenced revision is absent → `RefusalIdentityMismatch` ("referenced revision is absent") | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift`, case "correction references a revision that was never stored" (staged as `driftAtAppend`, field `supersedes`) |
| B14 | else | 242:10 | the else arm carrying the scope comparison | `TestStoreRejectsCorrectionThatChangesEvidenceScope` |
| B15 | if | 242:10 | market, symbol, issuer identity, issuer mapping version or kind differs from the superseded revision → `RefusalIdentityMismatch` ("revision cannot change evidence scope identity") | `TestStoreRejectsCorrectionThatChangesEvidenceScope` (symbol AAPL → MSFT; evidence count stays 1) |
| B16 | if | 256:2 | the `INSERT INTO evidence_records` failed → `strategy evidence: appending: %w` | `TestOpenReadOnlyReplaysButCannotCreateSnapshotOrAppend` (a read-only handle must not append; the test does not separate this from B4) |
| B17 | if | 259:2 | the final commit failed → return | not-applicable: driver failure |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `evidence.PayloadDigest()` | 195, 220, 228, 255 | the canonical digest, computed once in `NewEnvelope`; the comparison at 220 is what separates idempotence from conflict |
| `invalid(...)` | 196, 201, 241, 243 | typed refusals (`RefusalPayloadInvalid`, `RefusalTimestampInvalid`, `RefusalIdentityMismatch` twice) |
| `s.clk.Now().UTC()` | 198 | the trusted ingestion clock; injected in tests via `marketclock.NewFake` (`TestRevisionConflictUsesInjectedClock`) |
| `evidence.Header()` | 199, 214, 220, 227, 242 | header snapshots; re-read at 214 after the re-stamp so the persisted values are the trusted ones |
| `NewEnvelope(h, evidence.CanonicalPayload())` | 204 | rebuild with the trusted `IngestedAt`; re-runs the whole header validation |
| `s.db.BeginTx(ctx, nil)` | 209 | one transaction for probe, quarantine and insert |
| `tx.Rollback()` | 213 (deferred) | the only defer; makes every refusal path leave no row |
| `readByIdentity(ctx, tx, …)` | 215, 238 | identity probe, then the superseded-revision probe |
| `sameImmutableProvenance(existing.Header(), h)` | 220 | equality on the header with `EvidenceID` and `IngestedAt` blanked (store.go:265–269) |
| `tx.Commit()` | 221, 232, 259 | idempotent return, quarantine commit, and the accepted-append commit |
| `tx.ExecContext(…INSERT OR IGNORE INTO revision_conflicts…)` | 223 | quarantine row; the attempted payload is stored as the literal `{"redacted":true}` (228), never the real bytes |
| `stamp(...)` | 228, 254 (×4) | text timestamps for `quarantined_at` and the four evidence clocks |
| `fmt.Errorf("%w: %s/%s/%s", ErrRevisionConflict, …)` | 235 | the conflict error names authority/record/revision and is returned together with the accepted evidence |
| `tx.ExecContext(…INSERT INTO evidence_records…)` | 246 | the accepted append; `NULLIF(?, '')` keeps an absent supersedes as SQL NULL |
| `fmt.Errorf("strategy evidence: appending: %w", err)` | 257 | insert failure |

## State mutations and fallbacks

- SQLite writes, all inside one transaction: `revision_conflicts` on the conflict path (223) and `evidence_records` on the accepted path (246). Both tables refuse UPDATE and DELETE by trigger, so the only correction mechanism is another append that supersedes.
- The parameter `evidence` is rebound at 208 to the re-stamped envelope, so the value returned in `AppendResult` is the one that was persisted, not the one the caller handed in. On the idempotent and conflict paths the returned evidence is the *existing* row (221, 235) — a caller cannot mistake its rejected attempt for the accepted record.
- Deliberate data loss: the quarantine row stores `{"redacted":true}` in `attempted_payload`, asserted by `TestStoreAppendIsIdempotentAndQuarantinesRevisionConflict`, which checks both that the real payload is absent and that the column equals that exact literal.
- No fallback anywhere: every refusal returns `AppendResult{}` with a typed error, and the deferred rollback discards partial work. 15 AST assignments, all local apart from the two SQL statements.

## Safety conclusion

- Safe edit boundary: the clock authority (198/203), the identity probe key (215) and the supersedes scope comparison (242) are the three places where a widening would let unverified evidence become ledger truth. None may change without a fresh RED and a Branch Test Map.
- High-risk impact: yes — this is the evidence ledger. It does not itself place orders, but strategy projections read what it accepted, and an accepted-but-wrong revision is indistinguishable downstream from a correct one. The append-only triggers mean a bad row cannot be withdrawn, only superseded.
- Untested branches are nine: B1 (defensive), B3 (unreachable given B2), B2's taken side, and six driver-failure paths (B4, B5, B8, B9, B11, B17). Every decision branch that separates accept from refuse from quarantine — B6, B7, B10, B12/B13, B14/B15 — is covered, and B16 is covered by the read-only handle test. Package suite green (`go test ./internal/strategyevidence -count=1`, 2026-08-17).

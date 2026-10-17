# Branch Test Map: `Store.Append`

- Source: `internal/strategyevidence/store.go`, SHA-256 `53de9bb53980c7e3be959b1ab51b03667c91c232d04c50bb1724426e837910ca`; branch IDs follow `ast.json` (generated 2026-08-17).
- AST counts: branches 17, returns 15, calls 40, defers 1, go statements 0. Source range `194:1`–`263:2`. Signature `(s *Store) Append(ctx context.Context, evidence Envelope) (AppendResult, error)`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1b brief (official raw reader + bar producer). Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 195:2 — an envelope with no payload digest (never built by `NewEnvelope`) | not-applicable: defensive constructor guard; every test envelope is built through `NewEnvelope` | n/a (not edited) | not-applicable |
| B2 | if at 200:2 — the trusted store clock sits behind the header's `ObservedAt` | not-applicable for the taken side; the untaken side (caller-backdated `IngestedAt` overwritten by the trusted clock, then excluded by the ingestion cutoff) is `TestStoreStampsTrustedIngestionClockAndRejectsBackdating` | n/a (not edited) | existing suite green |
| B3 | if at 205:2 — rebuilding the envelope with the trusted `IngestedAt` is refused | not-applicable: unreachable given B2 — only a zero-valued store clock could fail the re-validation | n/a (not edited) | not-applicable |
| B4 | if at 210:2 — the transaction cannot be opened | not-applicable: driver failure; `TestOpenReadOnlyReplaysButCannotCreateSnapshotOrAppend` proves a read-only handle refuses but does not separate this from B16 | n/a (not edited) | not-applicable |
| B5 | if at 216:2 — the identity probe fails | not-applicable: driver failure | n/a (not edited) | not-applicable |
| B6 | if at 219:2 — the same authority/record/revision already has a row; and the fresh-identity path that must not enter it | `TestStoreAppendIsIdempotentAndQuarantinesRevisionConflict`, `TestRevisionConflictUsesInjectedClock`, `TestStoreSameDigestWithChangedProvenanceIsConflict`, `TestClosedBarAppendQuarantinesSameRevisionWithADifferentDigest`, `TestStoreDualCutoffAndAppendOnlyCorrectionSnapshot` (untaken side) | n/a (not edited) | existing suite green |
| B7 | if at 220:3 — a byte-identical re-append is idempotent, while an equal digest with drifted provenance is a conflict | `TestStoreAppendIsIdempotentAndQuarantinesRevisionConflict` (taken), `TestStoreSameDigestWithChangedProvenanceIsConflict` (untaken) | n/a (not edited) | existing suite green |
| B8 | if at 229:3 — the quarantine insert fails | not-applicable: driver failure; the statement is `INSERT OR IGNORE` | n/a (not edited) | not-applicable |
| B9 | if at 232:3 — committing the quarantine fails | not-applicable: driver failure | n/a (not edited) | not-applicable |
| B10 | if at 237:2 — the appended revision names a superseded one; and a first revision that does not | `TestStoreDualCutoffAndAppendOnlyCorrectionSnapshot`, `TestStoreRejectsCorrectionThatChangesEvidenceScope`, `TestSealBarSeriesRefusesEveryHeaderPayloadDrift`, `TestStoreAppendIsIdempotentAndQuarantinesRevisionConflict` (untaken side) | n/a (not edited) | existing suite green |
| B11 | if at 238:3 — the superseded-revision lookup fails | not-applicable: driver failure | n/a (not edited) | not-applicable |
| B12 | else at 240:10 — the else arm reached when the lookup itself succeeded | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` (case "correction references a revision that was never stored") | n/a (not edited) | existing suite green |
| B13 | if at 240:10 — the referenced revision was never stored, so the correction is refused before any row is written | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` (same case, staged at append with field `supersedes`) | n/a (not edited) | existing suite green |
| B14 | else at 242:10 — the else arm carrying the scope comparison | `TestStoreRejectsCorrectionThatChangesEvidenceScope` | n/a (not edited) | existing suite green |
| B15 | if at 242:10 — a correction that changes market, symbol, issuer identity, issuer mapping version or kind is refused and appends nothing | `TestStoreRejectsCorrectionThatChangesEvidenceScope` (symbol moved AAPL → MSFT, evidence count stays 1) | n/a (not edited) | existing suite green |
| B16 | if at 256:2 — the evidence insert fails on a store that cannot be written | `TestOpenReadOnlyReplaysButCannotCreateSnapshotOrAppend` | n/a (not edited) | existing suite green |
| B17 | if at 259:2 — the final commit fails | not-applicable: driver failure | n/a (not edited) | not-applicable |

Non-branch properties the L1b brief cites: the append-only triggers on `evidence_records` and `revision_conflicts` (`TestStoreSchemaTooNewAndAppendOnlyTriggers`), the redacted quarantine payload and the conflict returning the *accepted* evidence (`TestStoreAppendIsIdempotentAndQuarantinesRevisionConflict`), the injected quarantine clock (`TestRevisionConflictUsesInjectedClock`), and bar-shaped appends round-tripping through the constructor (`TestSealBarSeriesRoundTripsEnvelopesBuiltByTheConstructor`).

Verification: `go test ./internal/strategyevidence -count=1` green on 2026-08-17 (exit 0). No RED round applies — a112 does not edit this function.

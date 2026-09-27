# Function Logic Map: `Journal.CommitRiskBucketAdmission`

- Source: `internal/journal/risk_bucket.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| admission plan | canonical five buckets, immutable snapshots, exact owner/campaign/lane identity | `riskbucket.CalculateAdmission` plus journal rows | refusal or snapshot mismatch before mutation |
| transaction replay | one immutable `transaction_id` preimage | `risk_bucket_final_decisions` | byte-divergent replay fails closed |
| active owner reuse | same prospective generation/lane/campaign and clean replay digest | owner and state snapshot rows | conflict or replay mismatch |
| entry cleanliness | no exact-market active-owner latch, no exact-market scope latch, no applicable account/symbol reconcile | journal rows inside admission transaction, before owner lookup/insert | `ErrRiskBucketEntryBlocked`; no owner or exposure row written |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.5 post-edit) |
|---|---|---|---|---|
| B1 | if at 90:2 | `if decision.Refusal != nil {`; then `return RiskBucketAdmissionReceipt{}, decision.Refusal` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B2 | if at 93:2 | `if err := validateRiskBucketAdmission(plan, decision); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | covered |
| B3 | if at 97:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B4 | if at 101:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: begin risk bucket admission: %w", err)` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B5 | if at 109:2 | `if err == nil {`; then `if priorDigest != digest \|\| priorQ != decision.QFinal {` (line last changed by `4aee6853`) | a066 commit | covered |
| B6 | if at 110:3 | `if priorDigest != digest \|\| priorQ != decision.QFinal {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: transaction %s", ErrRiskBucketReplayMismatch, plan.TransactionID)` (line last changed by `4aee6853`) | a066 commit | covered |
| B7 | if at 114:3 | `if err := tx.Commit(); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B8 | if at 119:2 | `if !errors.Is(err, sql.ErrNoRows) {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: read risk bucket admission: %w", err)` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B9 | if at 124:2 | `if err := tx.QueryRowContext(ctx, `SELECT account_ref,state FROM risk_reservations WHERE id=?`, plan.ExistingReservationID).Scan(&reservationAccount, &reservationState); err != nil {`; then `if errors.Is(err, sql.ErrNoRows) {` (line last changed by `4aee6853`) | a066 commit | covered |
| B10 | if at 125:3 | `if errors.Is(err, sql.ErrNoRows) {`; then `return RiskBucketAdmissionReceipt{}, ErrReservationNotFound` (line last changed by `4aee6853`) | a066 commit | covered |
| B11 | if at 130:2 | `if reservationAccount != plan.Owner.Key.AccountID \|\| reservationState != ReservationHeld {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: existing reservation binding", ErrRiskBucketSnapshotMismatch)` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B12 | if at 133:2 | `if err := ensureRiskBucketEntryScopeClean(ctx, tx, plan.Owner.Key); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `9bf1a3a6`) | a066: exact-scope latch/reconcile gate before any owner lookup or insert | covered |
| B13 | if at 137:2 | `if err := refuseEntryUnderLossLock(ctx, tx, plan.Owner.Key.AccountID, plan.Owner.Key.Market, admissionHorizon(decision)); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `0004536c`) | a066: entry loss lock of the plan's account×market×horizon refuses inside the admission transaction (a066 5.5) | covered |
| B14 | switch at 144:2 | `switch {`; then `if prospective != plan.Owner.Key.ProspectiveGeneration \|\| lane != plan.Owner.LaneID \|\| campaign != plan.Owner.CampaignID {` (line last changed by `4aee6853`) | a066 commit | covered |
| B15 | case at 145:2 | `case err == nil:`; then `if prospective != plan.Owner.Key.ProspectiveGeneration \|\| lane != plan.Owner.LaneID \|\| campaign != plan.Owner.CampaignID {` (line last changed by `4aee6853`) | a066 commit | covered |
| B16 | if at 146:3 | `if prospective != plan.Owner.Key.ProspectiveGeneration \|\| lane != plan.Owner.LaneID \|\| campaign != plan.Owner.CampaignID {`; then `return RiskBucketAdmissionReceipt{}, ErrRiskBucketOwnerConflict` (line last changed by `4aee6853`) | a066 commit | covered |
| B17 | case at 150:2 | `case errors.Is(err, sql.ErrNoRows):`; then `_, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_owners(account_ref,market,symbol,prospective_generation,lane_id,campaign_id,acquired_at) VALUES(?,?,?,?,?,` (line last changed by `4aee6853`) | a066 commit | covered |
| B18 | if at 152:3 | `if err != nil {`; then `if strings.Contains(err.Error(), "UNIQUE constraint failed") {` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B19 | if at 153:4 | `if strings.Contains(err.Error(), "UNIQUE constraint failed") {`; then `return RiskBucketAdmissionReceipt{}, ErrRiskBucketOwnerConflict` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B20 | case at 158:2 | `default:`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B21 | if at 161:2 | `if ownerReused {`; then `if err := verifyRiskBucketStateDigest(ctx, tx, plan.Owner.Key); err != nil {` (line last changed by `4aee6853`) | a066 commit | covered |
| B22 | if at 162:3 | `if err := verifyRiskBucketStateDigest(ctx, tx, plan.Owner.Key); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B23 | if at 166:3 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B24 | for at 170:3 | `for rows.Next() {`; then `var d, v, pv string` (line last changed by `4aee6853`) | a066 commit | covered |
| B25 | if at 172:4 | `if err := rows.Scan(&d, &v, &pv); err != nil {`; then `rows.Close()` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B26 | if at 178:3 | `if err := rows.Close(); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B27 | if at 181:3 | `if len(existing) != len(decision.Caps) {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: scale-in bucket identity", ErrRiskBucketSnapshotMismatch)` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B28 | range at 184:3 | `for _, cap := range decision.Caps {`; then `if !existing[cap.Key] {` (line last changed by `4aee6853`) | a066 commit | covered |
| B29 | if at 185:4 | `if !existing[cap.Key] {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: scale-in bucket identity", ErrRiskBucketSnapshotMismatch)` (line last changed by `4aee6853`) | a066 commit | covered |
| B30 | if at 192:2 | `if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(owner_sequence),0)+1 FROM risk_bucket_final_decisions WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration).Scan(&ownerSequence); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B31 | if at 196:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B32 | if at 201:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: insert risk bucket decision: %w", err)` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B33 | range at 205:2 | `for i, cap := range decision.Caps {`; then `bucket := plan.Admission.Buckets[i]` (line last changed by `4aee6853`) | a066 commit | covered |
| B34 | if at 214:3 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B35 | if at 218:3 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B36 | if at 222:3 | `if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_policies WHERE bucket_dimension=? AND bucket_value=? AND policy_version=?`, string(cap.Key.Dimension), cap.Key.Value, cap.Key.PolicyVersion).Scan(&storedPolicyDigest); err != nil \|\| storedPolicyDigest != policyRecordDigest {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: immutable policy collision", ErrRiskBucketSnapshotMismatch)` (line last changed by `4aee6853`) | a066 commit | covered |
| B37 | if at 229:3 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B38 | if at 234:3 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B39 | if at 238:3 | `if err := tx.QueryRowContext(ctx, `SELECT record_digest FROM risk_bucket_snapshots WHERE snapshot_id=?`, ref.SnapshotID).Scan(&storedSnapshotDigest); err != nil \|\| storedSnapshotDigest != snapshotRecordDigest {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: immutable snapshot collision", ErrRiskBucketSnapshotMismatch)` (line last changed by `4aee6853`) | a066 commit | covered |
| B40 | if at 243:3 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: insert %s risk bucket reservation: %w", cap.Key.Dimension, err)` (line last changed by `4aee6853`) | a066 commit | covered |
| B41 | if at 248:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B42 | if at 252:2 | `if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_sequence),0)+1 FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration).Scan(&stateSequence); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B43 | if at 256:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B44 | if at 260:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `4aee6853`) | a066 commit | NOT covered |
| B45 | if at 263:2 | `if err := tx.Commit(); err != nil {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: commit risk bucket admission: %w", err)` (line last changed by `4aee6853`) | a066 commit | NOT covered |

5.5 post-edit (2026-09-27, HEAD `eecc5fe7` + working tree): AST 88–267, 44 → 45 branches. The new B13 (`refuseEntryUnderLossLock`) sits right after B12; old B13–B44 are now B14–B45 with identical bodies. The earlier grouped table was replaced by measured per-branch rows.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `refuseEntryUnderLossLock` + `admissionHorizon` | a066 5.5 entry loss lock for the plan's account×market×horizon, read inside this transaction (no TOCTOU) | lock → `*riskbucket.RefusalError{ENTRY_LOSS_LOCK_ACTIVE}` wrapping `ErrRiskBucketEntryLossLocked` (wraps `ErrRiskBucketEntryBlocked`); unknown scope / read error → `ErrRiskBucketEntryBlocked`; caller rolls back | AST + `analysis/mutation-5.5/ledger-run3.tsv` |
| `riskbucket.CalculateAdmission` | deterministic cap decision | refusal writes nothing | AST plus riskbucket unit tests |
| `verifyRiskBucketStateDigest` | reject tampered owner state before reuse | mismatch blocks admission | replay tests |
| `ensureRiskBucketEntryScopeClean` | enforce durable entry gate before both first-owner INSERT and reuse | exact-market latch/reconcile blocks; other-market evidence stays isolated | both late released-owner fill regressions |
| `recordRiskBucketStateTx` | seal committed admission state | shares admission transaction | crash/replay tests |

## State mutations and fallbacks

- All owner, decision, bucket reservation and state snapshot writes share one SQL transaction.
- The new gate is before active-owner lookup, first-owner INSERT, and every scale-in decision/reservation insert.
- There is no fallback that clears a latch or reconcile state.

## Wave 2A re-extraction (2026-09-25)

- AST re-extracted at HEAD (88–263, previously 72–246). Old/new branch alignment by `(kind, source line)` is
  identical B1–B44.
- The only body change since the bundle was written is commit `8022f578`: the immutable snapshot insert now
  stores `ref.snapshotWindow()` (snapshot-specific observed/fresh times, falling back to the shared
  `ObservedAt`/`FreshUntil`) instead of the shared window. No branch, return or refusal changed; the matching
  `validateRiskBucketAdmission` check compares policy and snapshot evidence against their own windows.

## Safety conclusion

- Safe edit boundary: add fail-closed journal cleanliness checks immediately after verifying the reused-owner digest.
- High-risk impact: yes — it prevents new exposure after a durable late-fill/reconcile signal without affecting exit or protection paths.

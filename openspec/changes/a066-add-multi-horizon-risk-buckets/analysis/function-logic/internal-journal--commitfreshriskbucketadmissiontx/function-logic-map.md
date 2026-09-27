# Function Logic Map: `commitFreshRiskBucketAdmissionTx`

- Source: `internal/journal/risk_bucket_issuance.go`
- AST evidence: `ast.json` (5.5 pre-edit, extracted 2026-09-27 at HEAD `02716357`; lines 397–481, 25 branches)
- Risk scan: `risk-pattern-report.md`
- Coverage: `analysis/harness/pertest_cover_5_5.sh` — 97 selected tests, one `-coverprofile` each; a subset, so
  `NOT covered` means "no selected test ran the body", not "no test in the repository does".

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | caller's open `BEGIN IMMEDIATE` transaction (q_final issuance `risk_bucket_issuance.go:140`, strategy first leg `strategy_first_leg_atomic.go:174`) | caller | any error returns; the caller's deferred rollback discards every row |
| plan / decision | admission already calculated and validated by the caller | `riskbucket.CalculateAdmission`, `validateRiskBucketAdmission` | not re-validated here except identity/binding checks |
| existing aggregate reservation | HELD, same account and decision | `risk_reservations` | not found / snapshot mismatch |
| entry scope | no exact-market owner latch, scope latch or applicable reconcile | journal rows inside the transaction | `ErrRiskBucketEntryBlocked` |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.6.1 post-edit) |
|---|---|---|---|---|
| B1 | if at 399:2 | `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_final_decisions WHERE transaction_id=? OR decision_id=?`, plan.TransactionID, plan.DecisionID).Scan(&existing); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B2 | if at 402:2 | `if existing != 0 {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: q_final issuance identity", ErrRiskBucketReplayMismatch)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B3 | if at 406:2 | `if err := tx.QueryRowContext(ctx, `SELECT account_ref,decision_id,state FROM risk_reservations WHERE id=?`, plan.ExistingReservationID).Scan(&reservationAccount, &reservationDecision, &reservationState); err != nil {`; then `if errors.Is(err, sql.ErrNoRows) {` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B4 | if at 407:3 | `if errors.Is(err, sql.ErrNoRows) {`; then `return RiskBucketAdmissionReceipt{}, ErrReservationNotFound` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B5 | if at 412:2 | `if reservationAccount != plan.Owner.Key.AccountID \|\| reservationDecision != plan.DecisionID \|\| reservationState != ReservationHeld {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("%w: existing reservation binding", ErrRiskBucketSnapshotMismatch)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B6 | if at 415:2 | `if err := ensureRiskBucketEntryScopeClean(ctx, tx, plan.Owner.Key); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B7 | if at 419:2 | `if err := refuseEntryUnderLossLock(ctx, tx, plan.Owner.Key.AccountID, plan.Owner.Key.Market, admissionHorizon(decision)); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `0004536c`) | not a066 | covered |
| B8 | switch at 426:2 | `switch {`; then `if prospective != plan.Owner.Key.ProspectiveGeneration \|\| lane != plan.Owner.LaneID \|\| campaign != plan.Owner.CampaignID {` (line last changed by `a37d97f5`) | a066 commit | covered |
| B9 | case at 427:2 | `case err == nil:`; then `if prospective != plan.Owner.Key.ProspectiveGeneration \|\| lane != plan.Owner.LaneID \|\| campaign != plan.Owner.CampaignID {` (line last changed by `a37d97f5`) | a066 commit | covered |
| B10 | if at 428:3 | `if prospective != plan.Owner.Key.ProspectiveGeneration \|\| lane != plan.Owner.LaneID \|\| campaign != plan.Owner.CampaignID {`; then `return RiskBucketAdmissionReceipt{}, ErrRiskBucketOwnerConflict` (line last changed by `a37d97f5`) | a066 commit | covered |
| B11 | case at 432:2 | `case errors.Is(err, sql.ErrNoRows):`; then `if _, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_owners(account_ref,market,symbol,prospective_generation,lane_id,campaign_id,acquired_at) VALUES(?,?,?,?` (line last changed by `a37d97f5`) | a066 commit | covered |
| B12 | if at 433:3 | `if _, err = tx.ExecContext(ctx, `INSERT INTO risk_bucket_owners(account_ref,market,symbol,prospective_generation,lane_id,campaign_id,acquired_at) VALUES(?,?,?,?,?,?,?)`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, plan.Owner.LaneID, plan.Owner.CampaignID, canonicalRiskTime(plan.CreatedAt)); err != nil {`; then `if strings.Contains(err.Error(), "UNIQUE constraint failed") {` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B13 | if at 434:4 | `if strings.Contains(err.Error(), "UNIQUE constraint failed") {`; then `return RiskBucketAdmissionReceipt{}, ErrRiskBucketOwnerConflict` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B14 | case at 439:2 | `default:`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B15 | if at 442:2 | `if ownerReused {`; then `if err := verifyRiskBucketStateDigest(ctx, tx, plan.Owner.Key); err != nil {` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B16 | if at 443:3 | `if err := verifyRiskBucketStateDigest(ctx, tx, plan.Owner.Key); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B17 | if at 446:3 | `if err := verifyRiskBucketScaleInIdentity(ctx, tx, plan.Owner.Key, decision.Caps); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B18 | if at 453:2 | `if err := refuseStaleBucketUsage(ctx, tx, plan.Owner.Key.AccountID, plan.Admission.Buckets); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `b8211926`) | a066: 5.6.1 F1: shared-bucket snapshot understates ledger usage → BUCKET_USAGE_STALE (retryable) | covered |
| B19 | if at 457:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B20 | if at 462:2 | `if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(owner_sequence),0)+1 FROM risk_bucket_final_decisions WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration).Scan(&ownerSequence); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B21 | if at 466:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B22 | if at 470:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, fmt.Errorf("journal: insert q_final risk bucket decision: %w", err)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B23 | if at 473:2 | `if err := insertFreshRiskBucketReservations(ctx, tx, plan, decision); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B24 | if at 477:2 | `if err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B25 | if at 481:2 | `if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_sequence),0)+1 FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration).Scan(&stateSequence); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B26 | if at 484:2 | `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_state_snapshots(snapshot_id,account_ref,market,symbol,prospective_generation,state_digest,event_sequence,created_at) VALUES(?,?,?,?,?,?,?,?)`, plan.TransactionID+":state:"+strconv.FormatInt(stateSequence, 10), plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, stateDigest, stateSequence, canonicalRiskTime(plan.CreatedAt)); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B27 | if at 487:2 | `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,event_digest,payload,created_at) VALUES(?,?,?,?,?,'ADMISSION_COMMITTED',?,?,?)`, plan.TransactionID+":event:1", plan.Owner.Key.AccountID, string(plan.Owner.Key.Market), plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, digest, storedPreimage, canonicalRiskTime(plan.CreatedAt)); err != nil {`; then `return RiskBucketAdmissionReceipt{}, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |

5.6.1 post-edit (HEAD `b8211926`): 26 → 27: new B18 `refuseStaleBucketUsage` after the owner/scale-in checks, before any write; the F2 helper replaced the policy block inside `insertFreshRiskBucketReservations`. Pre-edit table: `analysis/pre-edit/5.6.1/internal-journal--commitfreshriskbucketadmissiontx.md`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `refuseEntryUnderLossLock` + `admissionHorizon` | a066 5.5 entry loss lock for the plan's account×market×horizon, read inside this transaction (no TOCTOU) | lock → `*riskbucket.RefusalError{ENTRY_LOSS_LOCK_ACTIVE}` wrapping `ErrRiskBucketEntryLossLocked` (wraps `ErrRiskBucketEntryBlocked`); unknown scope / read error → `ErrRiskBucketEntryBlocked`; caller rolls back | AST + `analysis/mutation-5.5/ledger-run3.tsv` |
| `ensureRiskBucketEntryScopeClean` | durable entry gate before owner lookup/insert | exact-market latch/reconcile → `ErrRiskBucketEntryBlocked` | AST B6 |
| `verifyRiskBucketStateDigest`, `verifyRiskBucketScaleInIdentity` | reused owner must be untampered and same five bucket keys | mismatch errors | AST B15–B16 |
| `encodeQFinalStoredIssuance`, `riskBucketSnapshotSetDigest` | stored preimage / snapshot seal | error returns | AST B17, B19 |
| `insertFreshRiskBucketReservations` | five HELD monetary reservations | error returns | AST B21 |
| `loadRiskBucketState` | state digest after writes | error returns | AST B22 |

## State mutations and fallbacks

- Writes (all inside the caller's transaction): owner INSERT (B10 path), final decision INSERT, five reservations,
  state snapshot, `ADMISSION_COMMITTED` event. No fallback: every error returns and the caller rolls back.
- No broker, network or clock call.

## Safety conclusion

- Safe edit boundary (5.5): one read-only call to the entry-loss-lock rule immediately after B6
  (`ensureRiskBucketEntryScopeClean`), before the owner lookup and any INSERT. Its refusal returns like B6 and the
  caller's rollback removes the aggregate decision/reservation rows the caller already wrote in the same transaction.
  The horizon it passes is the plan's horizon bucket value (the value the five reservations are written from).
- High-risk impact: yes — this is the atomic q_final entry issuance used by Guardian and strategy first leg.

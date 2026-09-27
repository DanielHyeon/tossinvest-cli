# Function Logic Map: `Journal.RevalidateQFinalAdmission`

- Source: `internal/journal/risk_bucket_issuance.go`
- AST evidence: `ast.json` (5.5 pre-edit, extracted 2026-09-27 at HEAD `02716357`; lines 558–626, 17 branches)
- Risk scan: `risk-pattern-report.md`
- Coverage: `analysis/harness/pertest_cover_5_5.sh` — 97 selected tests; a subset (see the commitFresh map).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| decision ID | existing decision row | `LookupDecision` | lookup error returned as is (Gateway maps `ErrDecisionNotFound` to Guardian missing) |
| policy marker | `RiskIntent.PolicyVersion` with the q_final marker | durable preimage | unmarked or non-`RiskIntent` → `(false, nil)`: legacy decisions are not q_final-checked |
| q_final evidence | final decision, HELD aggregate, active exact owner, clean state digest, five HELD bucket reservations | a066 journal rows | any mismatch → `(true, ErrRiskBucketReplayMismatch…)` |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.5 post-edit) |
|---|---|---|---|---|
| B1 | if at 564:2 | `if err != nil {`; then `return false, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B2 | if at 568:2 | `if err != nil {`; then `return false, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B3 | if at 572:2 | `if !ok {`; then `return false, nil` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B4 | if at 576:2 | `if !required {`; then `return false, nil` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B5 | if at 582:2 | `if err != nil {`; then `return true, fmt.Errorf("%w: q_final decision evidence: %v", ErrRiskBucketReplayMismatch, err)` (line last changed by `a37d97f5`) | a066 commit | covered |
| B6 | if at 586:2 | `if !valid \|\| quantity != strconv.FormatUint(qFinal, 10) \|\| account != decision.AccountRef \|\| strings.EqualFold(market, intent.Market) == false \|\| symbol != strings.ToUpper(intent.Symbol) {`; then `return true, fmt.Errorf("%w: q_final preimage mismatch", ErrRiskBucketReplayMismatch)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B7 | if at 590:2 | `if err := j.db.QueryRowContext(ctx, `SELECT state,decision_id FROM risk_reservations WHERE id=?`, existingID).Scan(&legacyState, &legacyDecision); err != nil \|\| legacyState != ReservationHeld \|\| legacyDecision != decision.ID {`; then `return true, fmt.Errorf("%w: q_final aggregate reservation", ErrRiskBucketReplayMismatch)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B8 | if at 595:2 | `if err := j.db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND lane_id=? AND campaign_id=? AND released_at IS NULL`, account, market, symbol, prospective, lane, campaign).Scan(&activeOwner); err != nil \|\| activeOwner != 1 {`; then `return true, fmt.Errorf("%w: q_final owner", ErrRiskBucketReplayMismatch)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B9 | if at 598:2 | `if err := verifyRiskBucketStateDigest(ctx, j.db, key); err != nil {`; then `return true, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B10 | if at 602:2 | `if err != nil {`; then `return true, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B11 | for at 607:2 | `for rows.Next() {`; then `var dimension riskbucket.Dimension` (line last changed by `a37d97f5`) | a066 commit | covered |
| B12 | if at 610:3 | `if err := rows.Scan(&dimension, &state, &reserved, &held, &linkedExisting, &linkedOwner); err != nil {`; then `return true, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B13 | if at 613:3 | `if seen[dimension] \|\| state != "HELD" \|\| reserved == "" \|\| held != reserved \|\| linkedExisting != existingID \|\| linkedOwner != prospective {`; then `return true, fmt.Errorf("%w: q_final monetary reservation", ErrRiskBucketReplayMismatch)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B14 | if at 618:2 | `if err := rows.Err(); err != nil {`; then `return true, err` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B15 | range at 621:2 | `for _, dimension := range riskbucket.RequiredDimensionOrder() {`; then `if !seen[dimension] {` (line last changed by `a37d97f5`) | a066 commit | covered |
| B16 | if at 622:3 | `if !seen[dimension] {`; then `return true, fmt.Errorf("%w: missing %s q_final reservation", ErrRiskBucketReplayMismatch, dimension)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B17 | if at 626:2 | `if len(seen) != len(riskbucket.RequiredDimensionOrder()) {`; then `return true, fmt.Errorf("%w: q_final reservation dimension set", ErrRiskBucketReplayMismatch)` (line last changed by `a37d97f5`) | a066 commit | NOT covered |
| B18 | if at 632:2 | `if err != nil {`; then `return true, err` (line last changed by `00000000`) | a066: decision's horizon reservation cannot be read (storage error only: B16/B17 already proved all five HELD rows) — refuses | NOT covered |
| B19 | if at 635:2 | `if err := refuseEntryUnderLossLock(ctx, j.db, account, riskbucket.Market(market), horizon); err != nil {`; then `return true, err` (line last changed by `00000000`) | a066: user decision ⑤: a decision issued before the lock is refused at revalidation once its account×market×horizon lock is active (a066 5.5) | covered |

5.5 post-edit (2026-09-27, HEAD `eecc5fe7` + working tree): AST 562–639, 17 → 19 branches. B1–B17 unchanged; new B18 (horizon reservation read) and B19 (`refuseEntryUnderLossLock`) are the last checks before `return true, nil`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decisionHorizon` + `refuseEntryUnderLossLock` | user decision ⑤: a pre-lock decision is refused at submit time | `(true, err)`; Gateway maps `ErrRiskBucketEntryLossLocked` to `entry_loss_lock_active` | AST + `analysis/mutation-5.5/ledger-run3.tsv` |
| `LookupDecision`, `ParsePreimage`, `splitQFinalPolicyVersion` | find the durable q_final marker | lookup/parse errors returned | AST B1–B4 |
| `verifyRiskBucketStateDigest` | owner state untampered | mismatch error | AST B9 |
| direct `j.db` reads | final decision, aggregate hold, owner, five reservations | read error → error (fail closed at Gateway) | AST B5–B17 |

Callers: `Gateway.checkReservation` only (`gateway.go:901`), which is reached only for EXPOSURE_RAISING decisions
(checkReservation B1 returns nil for risk-reducing), both at the initial check and immediately before the broker call.

## State mutations and fallbacks

- Read-only; no transaction, no repair, no release.

## Safety conclusion

- Safe edit boundary (5.5): after the last existing check (B17) and before `return true, nil`, read the decision's
  horizon reservation value and call the same entry-loss-lock rule as the admission transaction. Placed last so every
  existing guard keeps its own tests (a new guard in front would shadow them). Refusal returns `(true, err)`, which
  the Gateway turns into a pre-broker refusal. Risk-reducing and unmarked legacy decisions never reach it.
- High-risk impact: yes — final exposure gate before broker transport.

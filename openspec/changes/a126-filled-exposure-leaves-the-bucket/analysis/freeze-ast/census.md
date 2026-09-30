# a126 freeze AST census (생성물 — 손으로 고치지 말 것)

- 기준 커밋(base-commit.txt): `989ab031a9177041da1ebb79fd1ed1209347ed30`
- 생성기: `analysis/freeze-ast/extract.py` → `go run ./tools/logic-map --file <f> --func <fn>`
- 각 분기 줄의 텍스트는 같은 base blob 에서 읽음. 열거는 AST 가 낸 전부임(선택 없음).

## `internal/journal/risk_bucket_owner.go` · `Journal.releaseRiskBucketOwner` (L814–1043, 분기 48)

AST: `ast/internal-journal--journal.releaseriskbucketowner.json` · sha256 `e0c9930d76928f3b…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 815 | `if err := validateRiskBucketOwnerKey(key, "release"); err != nil {` |
| B2 | if | 819 | `if err != nil {` |
| B3 | if | 828 | `if errors.Is(err, sql.ErrNoRows) {` |
| B4 | if | 831 | `if err != nil {` |
| B5 | if | 834 | `if released.Valid {` |
| B6 | if | 835 | `if !actual.Valid \|\| actual.String == "" {` |
| B7 | if | 838 | `if err := validateRiskBucketOwnerReleaseReceipt(ctx, tx, key, lane, campaign, actual.String, released.String); err != nil {` |
| B8 | if | 841 | `if err := tx.Commit(); err != nil {` |
| B9 | if | 846 | `if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {` |
| B10 | if | 850 | `if !actual.Valid \|\| parseErr != nil \|\| actualGeneration <= 0 {` |
| B11 | if | 856 | `if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_id,decision_id,prospective_token,` |
| B12 | if | 860 | `if errors.Is(err, sql.ErrNoRows) {` |
| B13 | if | 865 | `if campaignAccount != key.AccountID \|\| normaliseMarket(campaignMarket) != normaliseMarket(string(key.Market)) \|\|` |
| B14 | if | 870 | `if campaignState != "CLOSED" {` |
| B15 | if | 874 | `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM position_campaign_claims WHERE campaign_id=? OR` |
| B16 | if | 878 | `if activeClaims != 0 {` |
| B17 | if | 886 | `if err := tx.QueryRowContext(ctx, `SELECT p.id,p.state,p.quantity,COALESCE(p.closed_at,''),p.entry_decision_id,v.version FROM positions p` |
| B18 | if | 891 | `if errors.Is(err, sql.ErrNoRows) {` |
| B19 | if | 896 | `if !entryDecision.Valid \|\| positionVersion <= 0 {` |
| B20 | if | 899 | `if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(instance_seq),0) FROM positions WHERE account_ref=? AND market=? AND symbol=?`, key.AccountI...` |
| B21 | if | 902 | `if int64(latestGeneration) != actualGeneration {` |
| B22 | if | 905 | `if positionState != "CLOSED" \|\| closedAt == "" {` |
| B23 | if | 909 | `if err != nil \|\| zero != 0 {` |
| B24 | if | 913 | `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_final_decisions d JOIN risk_reservations r` |
| B25 | if | 919 | `if lineageCount == 0 \|\| decisionID != entryDecision.String {` |
| B26 | range | 938 | `for _, check := range checks {` |
| B27 | if | 940 | `if err := tx.QueryRowContext(ctx, check.query, check.args...).Scan(&count); err != nil {` |
| B28 | if | 943 | `if count != 0 {` |
| B29 | if | 947 | `if dirty, err := unresolvedScopedMutation(ctx, tx, key.AccountID, campaignMarket, campaignSymbol, "BUY"); err != nil {` |
| B30 | else | 949 | `} else if dirty {` |
| B31 | if | 949 | `} else if dirty {` |
| B32 | if | 952 | `if dirty, err := unresolvedScopedMutation(ctx, tx, key.AccountID, campaignMarket, campaignSymbol, "SELL"); err != nil {` |
| B33 | else | 954 | `} else if dirty {` |
| B34 | if | 954 | `} else if dirty {` |
| B35 | if | 959 | `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM reconcile_states WHERE account_ref=? AND released_at IS NULL` |
| B36 | if | 963 | `if activeReconcile != 0 {` |
| B37 | if | 967 | `if err != nil {` |
| B38 | range | 971 | `for _, latest := range []struct {` |
| B39 | if | 987 | `if err := tx.QueryRowContext(ctx, latest.query, latest.args...).Scan(&value); err != nil {` |
| B40 | if | 992 | `if fresh, err := timestampStrictlyAfter(authority.Record.BrokerAsOf, predecessors...); err != nil \|\| !fresh {` |
| B41 | if | 996 | `if err != nil {` |
| B42 | if | 1000 | `if err != nil {` |
| B43 | if | 1012 | `if err != nil {` |
| B44 | if | 1019 | `if err != nil {` |
| B45 | if | 1022 | `if affected, err := result.RowsAffected(); err != nil \|\| affected != 1 {` |
| B46 | if | 1026 | `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,event_digest,...` |
| B47 | if | 1029 | `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_owner_release_receipts(` |
| B48 | if | 1039 | `if err := tx.Commit(); err != nil {` |

## `internal/journal/risk_bucket_owner.go` · `Journal.applyRiskBucketOwnerBindingInTx` (L491–563, 분기 17)

AST: `ast/internal-journal--journal.applyriskbucketownerbindingintx.json` · sha256 `e0c9930d76928f3b…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 492 | `if strings.ToUpper(strings.TrimSpace(fill.Side)) != "BUY" {` |
| B2 | if | 499 | `if err != nil {` |
| B3 | if | 503 | `if err != nil {` |
| B4 | if | 507 | `if positive <= 0 {` |
| B5 | if | 516 | `if err != nil {` |
| B6 | for | 520 | `for rows.Next() {` |
| B7 | if | 522 | `if err := rows.Scan(&key.AccountID, &key.Market, &key.Symbol, &key.ProspectiveGeneration); err != nil {` |
| B8 | if | 528 | `if err := rows.Close(); err != nil {` |
| B9 | if | 531 | `if len(keys) == 0 {` |
| B10 | if | 534 | `if len(keys) != 1 {` |
| B11 | range | 535 | `for _, key := range keys {` |
| B12 | if | 536 | `if err := latchRiskBucketFillFailure(ctx, tx, key, fill, "REPLAY_MISMATCH", "owner bind matched multiple active scopes"); err != nil {` |
| B13 | if | 539 | `if err := j.recordRiskBucketStateTx(ctx, tx, key, "OWNER_BIND_REFUSED", "owner-bind-refused-"+fillObservationID(fill), "multiple-active-scopes", fi...` |
| B14 | if | 545 | `if _, err := j.bindRiskBucketOwnerActualInTx(ctx, tx, keys[0], fill.CommittedAt); err != nil {` |
| B15 | if | 547 | `if !errors.As(err, &lifecycle) && !isRiskBucketSemanticError(err) && !errors.Is(err, sql.ErrNoRows) {` |
| B16 | if | 550 | `if lifecycle == nil {` |
| B17 | if | 557 | `if err := latchRiskBucketFillFailure(ctx, tx, keys[0], fill, "REPLAY_MISMATCH", detail); err != nil {` |

## `internal/journal/risk_bucket_owner.go` · `Journal.latchReleasedOwnerLateFillInTx` (L565–631, 분기 15)

AST: `ast/internal-journal--journal.latchreleasedownerlatefillintx.json` · sha256 `e0c9930d76928f3b…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 572 | `if err != nil {` |
| B2 | for | 576 | `for rows.Next() {` |
| B3 | if | 578 | `if err := rows.Scan(&key.AccountID, &key.Market, &key.Symbol, &key.ProspectiveGeneration); err != nil {` |
| B4 | if | 584 | `if err := rows.Close(); err != nil {` |
| B5 | if | 587 | `if len(released) == 0 {` |
| B6 | range | 591 | `for _, old := range released {` |
| B7 | if | 592 | `if err := latchRiskBucketScope(ctx, tx, old, "ORPHAN_FILL", detail, fill.CommittedAt); err != nil {` |
| B8 | if | 595 | `if _, err := tx.ExecContext(ctx, `INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause,scope_market)` |
| B9 | if | 606 | `if err != nil {` |
| B10 | for | 610 | `for activeRows.Next() {` |
| B11 | if | 612 | `if err := activeRows.Scan(&key.AccountID, &key.Market, &key.Symbol, &key.ProspectiveGeneration); err != nil {` |
| B12 | if | 618 | `if err := activeRows.Close(); err != nil {` |
| B13 | range | 621 | `for _, key := range active {` |
| B14 | if | 622 | `if err := latchRiskBucketFillFailure(ctx, tx, key, fill, "ORPHAN_FILL", detail); err != nil {` |
| B15 | if | 625 | `if err := j.recordRiskBucketStateTx(ctx, tx, key, "LATE_RELEASED_OWNER_FILL", "late-released-owner-fill-"+fillObservationID(fill), detail, fill.Com...` |

## `internal/journal/risk_bucket_fill.go` · `Journal.applyRiskBucketFillInTx` (L202–248, 분기 13)

AST: `ast/internal-journal--journal.applyriskbucketfillintx.json` · sha256 `d8d3cefcb0667911…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 203 | `if strings.ToUpper(strings.TrimSpace(fill.Side)) != "BUY" \|\| fill.Delta == "0" {` |
| B2 | if | 207 | `if err != nil {` |
| B3 | if | 208 | `if errors.Is(err, ErrRiskBucketReplayMismatch) {` |
| B4 | if | 213 | `if !found {` |
| B5 | if | 217 | `if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {` |
| B6 | if | 218 | `if isRiskBucketSemanticError(err) {` |
| B7 | if | 224 | `if err != nil {` |
| B8 | if | 225 | `if isRiskBucketSemanticError(err) {` |
| B9 | if | 231 | `if err != nil {` |
| B10 | if | 234 | `if result.Duplicate {` |
| B11 | if | 237 | `if err := persistRiskBucketFillTransition(ctx, tx, order, event, state, next, result, nil, fill.CommittedAt); err != nil {` |
| B12 | if | 238 | `if errors.Is(err, ErrRiskBucketReplayMismatch) {` |
| B13 | if | 244 | `if err != nil {` |

## `internal/journal/risk_bucket_fill.go` · `Journal.completeRiskBucketFillActual` (L295–353, 분기 17)

AST: `ast/internal-journal--journal.completeriskbucketfillactual.json` · sha256 `d8d3cefcb0667911…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 296 | `if strings.TrimSpace(plan.OrderID) == "" \|\| strings.TrimSpace(plan.DecisionID) == "" \|\| plan.Owner.AccountID == "" \|\| plan.Owner.Symbol == ""...` |
| B2 | if | 300 | `if err != nil {` |
| B3 | if | 305 | `if err != nil {` |
| B4 | if | 308 | `if !found {` |
| B5 | if | 312 | `if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {` |
| B6 | if | 317 | `if err != nil {` |
| B7 | if | 321 | `if err := tx.QueryRowContext(ctx, `SELECT evidence_digest FROM risk_bucket_fill_actual_evidence WHERE fill_id=?`, fillID).Scan(&stored); err == nil {` |
| B8 | else | 329 | `} else if !errors.Is(err, sql.ErrNoRows) {` |
| B9 | if | 322 | `if stored != actualDigest {` |
| B10 | if | 325 | `if err := tx.Commit(); err != nil {` |
| B11 | if | 329 | `} else if !errors.Is(err, sql.ErrNoRows) {` |
| B12 | if | 333 | `if err != nil {` |
| B13 | if | 337 | `if err != nil {` |
| B14 | if | 340 | `if !result.ActualEvidenceCompleted {` |
| B15 | if | 343 | `if err := persistRiskBucketFillTransition(ctx, tx, order, event, state, next, result, plan.Actual, canonicalRiskTime(plan.ObservedAt)); err != nil {` |
| B16 | if | 346 | `if err := j.recordRiskBucketStateTx(ctx, tx, key, "ACTUAL_EVIDENCE_COMPLETED", fillID, actualDigest, canonicalRiskTime(plan.ObservedAt)); err != nil {` |
| B17 | if | 349 | `if err := tx.Commit(); err != nil {` |

## `internal/journal/risk_bucket_fill.go` · `persistRiskBucketFillTransition` (L989–1091, 분기 29)

AST: `ast/internal-journal--persistriskbucketfilltransition.json` · sha256 `d8d3cefcb0667911…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 991 | `if orderIdentity == "" {` |
| B2 | if | 995 | `if len(order.reservations) != len(riskbucket.RequiredDimensionOrder()) {` |
| B3 | range | 1003 | `for key, usage := range next.Buckets {` |
| B4 | if | 1006 | `if reservationID == "" \|\| !ok {` |
| B5 | if | 1010 | `if err != nil {` |
| B6 | if | 1014 | `if err != nil {` |
| B7 | if | 1018 | `if err != nil {` |
| B8 | if | 1022 | `if err := tx.QueryRowContext(ctx, `SELECT held_minor,filled_minor,overage_minor FROM risk_bucket_reservations WHERE reservation_id=?`, reservationI...` |
| B9 | if | 1023 | `if errors.Is(err, sql.ErrNoRows) {` |
| B10 | if | 1029 | `if err != nil {` |
| B11 | if | 1033 | `if err != nil {` |
| B12 | if | 1037 | `if err != nil {` |
| B13 | if | 1044 | `if result.ActualEvidenceCompleted {` |
| B14 | else | 1049 | `} else {` |
| B15 | if | 1046 | `if err != nil {` |
| B16 | if | 1054 | `if err != nil {` |
| B17 | if | 1058 | `if result.ActualEvidenceCompleted {` |
| B18 | else | 1062 | `} else {` |
| B19 | if | 1059 | `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_fill_actual_evidence(fill_id,evidence_digest,quote_currency,base_currency,price_quote,fx_...` |
| B20 | if | 1063 | `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_fills(fill_id,order_key,order_id,cumulative_fill,delta_quantity,actual_known,fill_digest,...` |
| B21 | range | 1067 | `for key, update := range updates {` |
| B22 | if | 1068 | `if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations SET held_minor=?,filled_minor=?,overage_minor=?,state=CASE WHEN ?='0' THEN 'FILLE...` |
| B23 | if | 1071 | `if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations SET risk_overage_latched=?,unknown_actual_latched=?,updated_at=? WHERE account_re...` |
| B24 | if | 1074 | `if result.ActualEvidenceCompleted {` |
| B25 | else | 1078 | `} else {` |
| B26 | if | 1075 | `if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_fill_allocations SET filled_minor=? WHERE fill_id=? AND reservation_id=?`, record.FilledMinor[...` |
| B27 | if | 1079 | `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_fill_allocations(fill_id,reservation_id,transfer_minor,filled_minor) VALUES(?,?,?,?)`, ev...` |
| B28 | if | 1084 | `if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_orders SET cumulative_fill=?,updated_at=? WHERE order_key=?`, event.NewCumulativeFill, observe...` |
| B29 | if | 1087 | `if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_owners SET risk_overage_latched=?,unknown_actual_latched=? WHERE account_ref=? AND market=? AN...` |

## `internal/journal/risk_bucket_fill.go` · `riskBucketSharedUsage` (L740–758, 분기 4)

AST: `ast/internal-journal--riskbucketsharedusage.json` · sha256 `d8d3cefcb0667911…`

| id | kind | line | source |
|---|---|---|---|
| B1 | range | 742 | `for key, usage := range owner {` |
| B2 | if | 744 | `if err != nil {` |
| B3 | if | 745 | `if errors.Is(err, riskbucket.ErrJournalUsageInvalid) {` |
| B4 | if | 752 | `if !okTotal \|\| !okOwn \|\| total.Cmp(own) < 0 {` |

## `internal/journal/risk_bucket_fill.go` · `queryRiskBucketOrder` (L579–621, 분기 10)

AST: `ast/internal-journal--queryriskbucketorder.json` · sha256 `d8d3cefcb0667911…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 582 | `if err != nil {` |
| B2 | for | 587 | `for rows.Next() {` |
| B3 | if | 589 | `if err := rows.Scan(&o.orderKey, &o.orderID, &o.decisionID, &o.account, &o.market, &o.symbol, &o.prospective, &o.state, &o.releaseReason, &o.quanti...` |
| B4 | if | 594 | `if err := rows.Err(); err != nil {` |
| B5 | if | 597 | `if len(matches) == 0 {` |
| B6 | if | 600 | `if len(matches) != 1 {` |
| B7 | if | 606 | `if err != nil {` |
| B8 | for | 609 | `for r.Next() {` |
| B9 | if | 611 | `if err := r.Scan(&d, &v, &pv, &id); err != nil {` |
| B10 | if | 617 | `if err := r.Close(); err != nil {` |

## `internal/journal/risk_bucket_usage.go` · `refuseStaleBucketUsage` (L30–84, 분기 12)

AST: `ast/internal-journal--refusestalebucketusage.json` · sha256 `8eacbf2fea3f1234…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 31 | `if len(caps) != len(buckets) {` |
| B2 | range | 34 | `for i, bucket := range buckets {` |
| B3 | if | 36 | `if err != nil {` |
| B4 | if | 43 | `if err := latchedUsageRefusal(bucket.Key.Dimension, bucket.Key.Value, usage); err != nil {` |
| B5 | if | 47 | `if !ok {` |
| B6 | if | 51 | `if !ok {` |
| B7 | if | 54 | `if claimed.Cmp(ledger) < 0 {` |
| B8 | if | 63 | `if caps[i].Key != bucket.Key {` |
| B9 | if | 67 | `if err != nil {` |
| B10 | if | 70 | `if !found {` |
| B11 | if | 74 | `if !ok {` |
| B12 | if | 77 | `if after.Cmp(recorded) > 0 {` |

## `internal/journal/risk_bucket_usage.go` · `smallestRecordedBucketLimit` (L88–114, 분기 6)

AST: `ast/internal-journal--smallestrecordedbucketlimit.json` · sha256 `8eacbf2fea3f1234…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 92 | `if err != nil {` |
| B2 | for | 97 | `for rows.Next() {` |
| B3 | if | 99 | `if err := rows.Scan(&raw); err != nil {` |
| B4 | if | 103 | `if !ok \|\| limit.Sign() < 0 {` |
| B5 | if | 106 | `if smallest == nil \|\| limit.Cmp(smallest) < 0 {` |
| B6 | if | 110 | `if err := rows.Err(); err != nil {` |

## `internal/journal/risk_bucket_usage.go` · `latchedUsageRefusal` (L129–141, 분기 3)

AST: `ast/internal-journal--latchedusagerefusal.json` · sha256 `8eacbf2fea3f1234…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 130 | `if !usage.Latched {` |
| B2 | if | 134 | `if usage.OverageLatched {` |
| B3 | if | 137 | `if usage.UnknownLatched {` |

## `internal/journal/risk_bucket_relaxation.go` · `Journal.ReleaseRiskOverageLatch` (L309–388, 분기 17)

AST: `ast/internal-journal--journal.releaseriskoveragelatch.json` · sha256 `090240a2a76db7e8…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 310 | `if j == nil \|\| j.db == nil {` |
| B2 | if | 313 | `if err := validRelaxation(req.Actor, req.Approval, req.Reason, req.Auditor, req.ReleasedAt); err != nil {` |
| B3 | if | 317 | `if strings.TrimSpace(key.AccountID) == "" \|\| key.AccountID != strings.TrimSpace(key.AccountID) \|\| strings.TrimSpace(key.Symbol) == "" \|\|` |
| B4 | if | 323 | `if err != nil {` |
| B5 | if | 330 | `if errors.Is(err, sql.ErrNoRows) {` |
| B6 | if | 333 | `if err != nil {` |
| B7 | if | 336 | `if overage == 0 {` |
| B8 | if | 341 | `if err := tx.QueryRowContext(ctx, `SELECT state_digest FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospecti...` |
| B9 | if | 345 | `if persisted != strings.TrimSpace(req.ExpectedStateDigest) {` |
| B10 | if | 348 | `if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {` |
| B11 | if | 352 | `if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_owners SET risk_overage_latched=0 WHERE account_ref=? AND market=? AND symbol=? AND prospectiv...` |
| B12 | if | 356 | `if _, err := tx.ExecContext(ctx, `UPDATE risk_bucket_reservations SET risk_overage_latched=0 WHERE account_ref=? AND market=? AND symbol=? AND owne...` |
| B13 | if | 363 | `if err != nil {` |
| B14 | if | 367 | `if err != nil {` |
| B15 | if | 372 | `if err := j.recordRiskBucketStateTx(ctx, tx, key, "OVERAGE_LATCH_RELEASED", "overage-latch-release-"+eventDigest[:24], eventDigest, releasedAt); er...` |
| B16 | if | 378 | `if err := req.Auditor.RecordAction(AuditActionOverageLatchRelease, setting, relaxationAuditAttempt, detail); err != nil {` |
| B17 | if | 381 | `if err := tx.Commit(); err != nil {` |

## `internal/riskbucket/fill.go` · `ApplyFill` (L100–277, 분기 38)

AST: `ast/internal-riskbucket--applyfill.json` · sha256 `845a67b3669ecda7…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 104 | `if event.FillID == "" \|\| event.OrderID == "" \|\| event.OrderQuantity == 0 \|\| event.NewCumulativeFill == 0 \|\| event.NewCumulativeFill > event...` |
| B2 | if | 108 | `if err := validateFillBuckets(next.Buckets, event.ReservedMinor); err != nil {` |
| B3 | if | 111 | `if len(event.TargetHeldMinor) != 0 {` |
| B4 | if | 112 | `if len(event.TargetHeldMinor) != len(event.ReservedMinor) {` |
| B5 | range | 115 | `for key := range event.ReservedMinor {` |
| B6 | if | 116 | `if _, err := parseMinor(event.TargetHeldMinor[key], 0); err != nil {` |
| B7 | if | 122 | `if strings.TrimSpace(event.OrderKey) != "" {` |
| B8 | if | 126 | `if !exists {` |
| B9 | else | 139 | `} else if order.OrderQuantity != event.OrderQuantity \|\| order.QuoteCurrency != event.QuoteCurrency \|\| order.BaseCurrency != event.BaseCurrency ...` |
| B10 | range | 136 | `for key := range event.ReservedMinor {` |
| B11 | if | 139 | `} else if order.OrderQuantity != event.OrderQuantity \|\| order.QuoteCurrency != event.QuoteCurrency \|\| order.BaseCurrency != event.BaseCurrency ...` |
| B12 | if | 143 | `if record, seen := order.Fills[event.FillID]; seen {` |
| B13 | if | 144 | `if record.CumulativeFill != event.NewCumulativeFill {` |
| B14 | if | 147 | `if record.ActualKnown {` |
| B15 | if | 152 | `if !known {` |
| B16 | range | 156 | `for key, transferRaw := range record.TransferMinor {` |
| B17 | if | 158 | `if err != nil {` |
| B18 | if | 162 | `if err != nil {` |
| B19 | if | 166 | `if actualMinor.Cmp(target) > 0 {` |
| B20 | if | 169 | `if target.Cmp(previousFilled) > 0 {` |
| B21 | if | 173 | `if err != nil {` |
| B22 | if | 177 | `if err != nil {` |
| B23 | if | 189 | `if err := recomputeOverageLatches(&next); err != nil {` |
| B24 | if | 197 | `if event.NewCumulativeFill <= order.CumulativeFill {` |
| B25 | if | 210 | `if actualKnown {` |
| B26 | range | 213 | `for key, reservedRaw := range event.ReservedMinor {` |
| B27 | if | 215 | `if err != nil {` |
| B28 | if | 220 | `if err != nil \|\| allocated.Cmp(previousTransferred) < 0 {` |
| B29 | if | 226 | `if err != nil {` |
| B30 | if | 230 | `if err != nil {` |
| B31 | if | 234 | `if len(event.TargetHeldMinor) != 0 {` |
| B32 | if | 236 | `if targetErr != nil {` |
| B33 | if | 239 | `if heldDeduction.Cmp(targetHeld) > 0 {` |
| B34 | if | 245 | `if heldDeduction.Cmp(held) > 0 {` |
| B35 | if | 252 | `if actualKnown && actualMinor.Cmp(filledDelta) > 0 {` |
| B36 | if | 256 | `if err != nil {` |
| B37 | if | 260 | `if !actualKnown {` |
| B38 | if | 272 | `if err := recomputeOverageLatches(&next); err != nil {` |

## `internal/riskbucket/fill.go` · `recomputeOverageLatches` (L354–406, 분기 13)

AST: `ast/internal-riskbucket--recomputeoveragelatches.json` · sha256 `845a67b3669ecda7…`

| id | kind | line | source |
|---|---|---|---|
| B1 | range | 356 | `for key, usage := range state.Buckets {` |
| B2 | if | 358 | `if err != nil {` |
| B3 | if | 362 | `if err != nil {` |
| B4 | if | 366 | `if err != nil {` |
| B5 | if | 370 | `if err != nil {` |
| B6 | if | 375 | `if shared, ok := state.SharedUsedMinor[key]; ok {` |
| B7 | if | 377 | `if err != nil {` |
| B8 | if | 380 | `if used, err = addMinor(used, others, 0); err != nil {` |
| B9 | if | 385 | `if overage.Sign() > 0 {` |
| B10 | if | 388 | `if err != nil {` |
| B11 | if | 391 | `if overage.Cmp(previous) > 0 {` |
| B12 | if | 397 | `if !anyOverage {` |
| B13 | range | 401 | `for key, usage := range state.Buckets {` |

## `internal/riskbucket/fill.go` · `clearResolvedUnknownLatches` (L420–433, 분기 4)

AST: `ast/internal-riskbucket--clearresolvedunknownlatches.json` · sha256 `845a67b3669ecda7…`

| id | kind | line | source |
|---|---|---|---|
| B1 | range | 421 | `for _, order := range state.Orders {` |
| B2 | range | 422 | `for _, fill := range order.Fills {` |
| B3 | if | 423 | `if !fill.ActualKnown {` |
| B4 | range | 429 | `for key, usage := range state.Buckets {` |

## `internal/riskbucket/production_snapshot_authority.go` · `ReadJournalBucketUsage` (L444–450, 분기 1)

AST: `ast/internal-riskbucket--readjournalbucketusage.json` · sha256 `5f26bf28bbf8207f…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 446 | `if err != nil {` |

## `internal/riskbucket/production_snapshot_authority.go` · `aggregateProductionRiskUsage` (L475–502, 분기 3)

AST: `ast/internal-riskbucket--aggregateproductionriskusage.json` · sha256 `5f26bf28bbf8207f…`

| id | kind | line | source |
|---|---|---|---|
| B1 | range | 479 | `for _, row := range rows {` |
| B2 | if | 482 | `if !filledOK \|\| !heldOK \|\| rowFilled.Sign() < 0 \|\| rowHeld.Sign() < 0 \|\| rowFilled.BitLen() > 256 \|\| rowHeld.BitLen() > 256 \|\|` |
| B3 | if | 494 | `if filled.BitLen() > 256 \|\| held.BitLen() > 256 {` |

## `internal/riskbucket/production_snapshot_authority.go` · `loadProductionRiskEntries` (L338–421, 분기 13)

AST: `ast/internal-riskbucket--loadproductionriskentries.json` · sha256 `5f26bf28bbf8207f…`

| id | kind | line | source |
|---|---|---|---|
| B1 | if | 340 | `if !ok {` |
| B2 | if | 343 | `if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {` |
| B3 | if | 352 | `if err != nil {` |
| B4 | if | 357 | `if err := db.PingContext(ctx); err != nil {` |
| B5 | if | 361 | `if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil \|\| version != productionRiskJournalSchema {` |
| B6 | if | 365 | `if err := db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`,` |
| B7 | if | 376 | `if authorityObserved.After(scope.AsOf) \|\| authorityFresh.Before(scope.AsOf) {` |
| B8 | range | 380 | `for _, dimension := range requiredDimensions {` |
| B9 | if | 382 | `if err != nil {` |
| B10 | if | 386 | `if usage.Latched {` |
| B11 | if | 396 | `if err != nil {` |
| B12 | if | 406 | `if err != nil {` |
| B13 | if | 415 | `if err != nil {` |


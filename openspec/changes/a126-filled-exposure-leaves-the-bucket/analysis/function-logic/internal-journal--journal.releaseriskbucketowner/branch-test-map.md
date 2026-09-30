# Branch Test Map: `Journal.releaseRiskBucketOwner`

분기 무편집 — 수리는 검사표 데이터(B26 순회 · B28 판정이 소비하는 `unresolved_fill` SQL). 그 효과는 a126 결속 시험이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:815` `if err := validateRiskBucketOwnerKey(key, "release"); err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B2 | `:819` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B3 | `:828` `if errors.Is(err, sql.ErrNoRows) {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B4 | `:831` `if err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B5 | `:834` `if released.Valid {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B6 | `:835` `if !actual.Valid \|\| actual.String == "" {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B7 | `:838` `if err := validateRiskBucketOwnerReleaseReceipt(ctx, tx, key, lane, campaign, actual.String, released.String); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B8 | `:841` `if err := tx.Commit(); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B9 | `:846` `if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B10 | `:850` `if !actual.Valid \|\| parseErr != nil \|\| actualGeneration <= 0 {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B11 | `:856` `if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_id,decision_id,prospective_token,` | — | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B12 | `:860` `if errors.Is(err, sql.ErrNoRows) {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B13 | `:865` `if campaignAccount != key.AccountID \|\| normaliseMarket(campaignMarket) != normaliseMarket(string(key.Market)) \|\|` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B14 | `:870` `if campaignState != "CLOSED" {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B15 | `:874` `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM position_campaign_claims WHERE campaign_id=? OR` | — | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B16 | `:878` `if activeClaims != 0 {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B17 | `:886` `if err := tx.QueryRowContext(ctx, `SELECT p.id,p.state,p.quantity,COALESCE(p.closed_at,''),p.entry_decision_id,v.version FROM positions p` | — | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B18 | `:891` `if errors.Is(err, sql.ErrNoRows) {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B19 | `:896` `if !entryDecision.Valid \|\| positionVersion <= 0 {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B20 | `:899` `if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(instance_seq),0) FROM positions WHERE account_ref=? AND market=? AND symbol=?`, ke…` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B21 | `:902` `if int64(latestGeneration) != actualGeneration {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B22 | `:905` `if positionState != "CLOSED" \|\| closedAt == "" {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B23 | `:909` `if err != nil \|\| zero != 0 {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B24 | `:913` `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_final_decisions d JOIN risk_reservations r` | — | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B25 | `:919` `if lineageCount == 0 \|\| decisionID != entryDecision.String {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B26 | `:938` `for _, check := range checks {` | 예 | a126 결속 시험 — evidence 로 해소된 체결은 `unresolved_fill` 로 막지 않음(수리) | no | no |
| B27 | `:940` `if err := tx.QueryRowContext(ctx, check.query, check.args...).Scan(&count); err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B28 | `:943` `if count != 0 {` | 예 | a126 결속 시험 — 해소되지 않은 체결은 여전히 `unresolved_fill` 로 거절 | no | no |
| B29 | `:947` `if dirty, err := unresolvedScopedMutation(ctx, tx, key.AccountID, campaignMarket, campaignSymbol, "BUY"); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B30 | `:949` `} else if dirty {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B31 | `:949` `} else if dirty {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B32 | `:952` `if dirty, err := unresolvedScopedMutation(ctx, tx, key.AccountID, campaignMarket, campaignSymbol, "SELL"); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B33 | `:954` `} else if dirty {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B34 | `:954` `} else if dirty {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B35 | `:959` `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM reconcile_states WHERE account_ref=? AND released_at IS NULL` | — | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B36 | `:963` `if activeReconcile != 0 {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B37 | `:967` `if err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B38 | `:971` `for _, latest := range []struct {` | — | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B39 | `:987` `if err := tx.QueryRowContext(ctx, latest.query, latest.args...).Scan(&value); err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B40 | `:992` `if fresh, err := timestampStrictlyAfter(authority.Record.BrokerAsOf, predecessors...); err != nil \|\| !fresh {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B41 | `:996` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B42 | `:1000` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B43 | `:1012` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B44 | `:1019` `if err != nil {` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B45 | `:1022` `if affected, err := result.RowsAffected(); err != nil \|\| affected != 1 {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B46 | `:1026` `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,eve…` | 아니오 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B47 | `:1029` `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_owner_release_receipts(` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |
| B48 | `:1039` `if err := tx.Commit(); err != nil {` | 예 | 기존 — a126 무변(분기 무편집) | n/a | n/a |

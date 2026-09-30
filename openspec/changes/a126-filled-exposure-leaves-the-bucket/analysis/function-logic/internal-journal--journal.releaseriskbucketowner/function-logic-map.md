# Function Logic Map: `Journal.releaseRiskBucketOwner`

- Source: `internal/journal/risk_bucket_owner.go` (`814`–`1043`)
- Qualified: `Journal.releaseRiskBucketOwner`
- AST evidence: `ast.json` (`source_sha256` e0c9930d76928f3b…) — **편집 전**(base `a189e74f`; freeze census `989ab031` 과 sha · 분기 일치)
- Risk scan: `risk-pattern-report.md`
- AST branches 48 · return 45 · 호출 96

**역할.** owner 하나를 해제한다 — 모든 청결 조건(종결 · 최신 generation · campaign · 검사표 · 미해소 mutation · 대사 · 공식 broker-zero 관측의 선행 순서)을
원장에서 유도한 뒤 영수증과 `released_at` 을 한 트랜잭션에 쓴다. 영수증 · `released_at` 의 유일한 작성자다(`:1017` · `:1029`, design Q1).
**생산 호출자 0**(CodeGraph callers 9 전부 `_test.go`, a066 잔여 #5).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `key` | 유효 owner 키 | 호출자 | B1 거절 |
| owner 행 · actual generation | 존재 · bind 됨 | `risk_bucket_owners` | B3~B6 거절 · 이미 해제면 B5 창 영수증 대조 |
| Position · campaign · claim | CLOSED · 수량 0 · 최신 · claim 0 | 원장 | B14 · B16 · B21~B23 거절 |
| 검사표 `checks`(10행) | 모두 count 0 | 원장 SQL | B28 → `ownerLifecycleBlocked(field)` |
| 미해소 BUY/SELL mutation | 없음 | `unresolvedScopedMutation` | B31 · B34 거절 |
| 대사 · broker-zero 관측 | 활성 0 · 권위 있는 관측이 모든 선행 사건 뒤 | 원장 | B36 · B37 · B40 거절 |

## Branches and early returns

> 표는 `analysis/harness/branch_table.py` 가 `ast.json` · 소스 · 커버리지(편집 전 `-coverpkg=./internal/riskbucket,./internal/journal`)로 만들었다.
> 조건은 소스 원문, 「창의 return」은 위치다(의미 아님).

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:815` `if err := validateRiskBucketOwnerKey(key, "release"); err != nil {` | :816 | 아니오 |
| B2 | if | `:819` `if err != nil {` | :820 | 아니오 |
| B3 | if | `:828` `if errors.Is(err, sql.ErrNoRows) {` | :829 | 아니오 |
| B4 | if | `:831` `if err != nil {` | :832 | 예 |
| B5 | if | `:834` `if released.Valid {` | — | 예 |
| B6 | if | `:835` `if !actual.Valid \|\| actual.String == "" {` | :836 | 아니오 |
| B7 | if | `:838` `if err := validateRiskBucketOwnerReleaseReceipt(ctx, tx, key, lane, campaign, actual.String, released.String); err != nil {` | :839 | 예 |
| B8 | if | `:841` `if err := tx.Commit(); err != nil {` | :842, :844 | 예 |
| B9 | if | `:846` `if err := verifyRiskBucketStateDigest(ctx, tx, key); err != nil {` | :847 | 예 |
| B10 | if | `:850` `if !actual.Valid \|\| parseErr != nil \|\| actualGeneration <= 0 {` | :851 | 아니오 |
| B11 | if | `:856` `if err := tx.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_id,decision_id,prospective_token,` | — | — |
| B12 | if | `:860` `if errors.Is(err, sql.ErrNoRows) {` | :861, :863 | 아니오 |
| B13 | if | `:865` `if campaignAccount != key.AccountID \|\| normaliseMarket(campaignMarket) != normaliseMarket(string(key.Market)) \|\|` | :868 | 예 |
| B14 | if | `:870` `if campaignState != "CLOSED" {` | :871 | 예 |
| B15 | if | `:874` `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM position_campaign_claims WHERE campaign_id=? OR` | :876 | — |
| B16 | if | `:878` `if activeClaims != 0 {` | :879 | 예 |
| B17 | if | `:886` `if err := tx.QueryRowContext(ctx, `SELECT p.id,p.state,p.quantity,COALESCE(p.closed_at,''),p.entry_decision_id,v.version FROM positions p` | — | — |
| B18 | if | `:891` `if errors.Is(err, sql.ErrNoRows) {` | :892, :894 | 아니오 |
| B19 | if | `:896` `if !entryDecision.Valid \|\| positionVersion <= 0 {` | :897 | 예 |
| B20 | if | `:899` `if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(instance_seq),0) FROM positions WHERE account_ref=? AND market=? AND symbol=?`, ke…` | :900 | 예 |
| B21 | if | `:902` `if int64(latestGeneration) != actualGeneration {` | :903 | 예 |
| B22 | if | `:905` `if positionState != "CLOSED" \|\| closedAt == "" {` | :906 | 예 |
| B23 | if | `:909` `if err != nil \|\| zero != 0 {` | :910 | 아니오 |
| B24 | if | `:913` `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_final_decisions d JOIN risk_reservations r` | :917 | — |
| B25 | if | `:919` `if lineageCount == 0 \|\| decisionID != entryDecision.String {` | :920 | 예 |
| B26 | range | `:938` `for _, check := range checks {` | — | 예 |
| B27 | if | `:940` `if err := tx.QueryRowContext(ctx, check.query, check.args...).Scan(&count); err != nil {` | :941 | 아니오 |
| B28 | if | `:943` `if count != 0 {` | :944 | 예 |
| B29 | if | `:947` `if dirty, err := unresolvedScopedMutation(ctx, tx, key.AccountID, campaignMarket, campaignSymbol, "BUY"); err != nil {` | :948 | 예 |
| B30 | else | `:949` `} else if dirty {` | — | 예 |
| B31 | if | `:949` `} else if dirty {` | :950 | 예 |
| B32 | if | `:952` `if dirty, err := unresolvedScopedMutation(ctx, tx, key.AccountID, campaignMarket, campaignSymbol, "SELL"); err != nil {` | :953 | 예 |
| B33 | else | `:954` `} else if dirty {` | — | 예 |
| B34 | if | `:954` `} else if dirty {` | :955 | 예 |
| B35 | if | `:959` `if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM reconcile_states WHERE account_ref=? AND released_at IS NULL` | :961 | — |
| B36 | if | `:963` `if activeReconcile != 0 {` | :964 | 예 |
| B37 | if | `:967` `if err != nil {` | :968 | 예 |
| B38 | range | `:971` `for _, latest := range []struct {` | — | — |
| B39 | if | `:987` `if err := tx.QueryRowContext(ctx, latest.query, latest.args...).Scan(&value); err != nil {` | :988 | 아니오 |
| B40 | if | `:992` `if fresh, err := timestampStrictlyAfter(authority.Record.BrokerAsOf, predecessors...); err != nil \|\| !fresh {` | :993 | 예 |
| B41 | if | `:996` `if err != nil {` | :997 | 아니오 |
| B42 | if | `:1000` `if err != nil {` | :1001 | 아니오 |
| B43 | if | `:1012` `if err != nil {` | :1013 | 아니오 |
| B44 | if | `:1019` `if err != nil {` | :1020 | 아니오 |
| B45 | if | `:1022` `if affected, err := result.RowsAffected(); err != nil \|\| affected != 1 {` | :1023 | 예 |
| B46 | if | `:1026` `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,eve…` | :1027 | 아니오 |
| B47 | if | `:1029` `if _, err := tx.ExecContext(ctx, `INSERT INTO risk_bucket_owner_release_receipts(` | :1037 | 예 |
| B48 | if | `:1039` `if err := tx.Commit(); err != nil {` | :1040, :1042 | 예 |


## Calls and live bindings

원장 질의 · 쓰기(`tx.QueryRowContext` · `tx.ExecContext` · `tx.Commit`), 청결 판정 도우미(`validateRiskBucketOwnerKey` ·
`validateRiskBucketOwnerReleaseReceipt` · `unresolvedScopedMutation` · 상태 봉인 대조), 오류 생성(`ownerLifecycleBlocked`). 브로커 호출 없음.
원장 오류는 되던지고, 청결 조건 실패는 `RiskBucketOwnerLifecycleError`(BlockingField 이름)로 거절한다.

## State mutations and fallbacks

- 성공 경로에서만: owner `released_at`(`:1017`) · 영수증 INSERT(`:1029`) · `OWNER_RELEASED` 사건 — 한 트랜잭션.
- 거절 경로는 쓰지 않는다(기존 시험 `riskBucketOwnerLifecycleWrites` 전후 대조).

## Safety conclusion

- **Safe edit boundary (a126 — a066 결함 수리, Manager 판정 (가) 2026-10-01)**: 검사표의 `unresolved_fill` 첫 행(`:932`) SQL 한 곳만. 오늘
  `f.actual_known=0 OR NOT EXISTS(evidence)` 는 fill 행의 `actual_known` 이 생산에서 늘 0(`risk_bucket_fill.go:1063` 의 유일한 INSERT)이라
  **모든** 체결 owner 를 영구히 막는다 — 해소 정의의 정본 `loadRiskBucketFillTransition`(`actual_known=1 OR EXISTS evidence`)과 갈라진 철자다.
  수리는 두 자리가 **한 SQL 조각**(`riskBucketFillActualResolvedSQL`)을 쓰게 한다. 분기(B1~B48) · 순서 · 쓰기 무변.
- **방향**: 해제를 여는 쪽이나 생산 호출자 0 — 생산 효과 0. 배선은 a126 tasks 3.1(면제 불가 의존) 뒤. 사용자 거부권 항목으로 보고.
- **High-risk impact**: yes — 원장 · 사이징(해제 = 사용량 떠남의 유일한 사건).

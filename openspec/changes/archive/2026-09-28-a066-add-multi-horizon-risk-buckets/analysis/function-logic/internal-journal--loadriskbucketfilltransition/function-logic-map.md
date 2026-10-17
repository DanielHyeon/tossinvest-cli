# Function Logic Map: `loadRiskBucketFillTransition`

- Source: `internal/journal/risk_bucket_fill.go`
- AST evidence: `ast.json` (5.7 pre-edit, extracted 2026-09-28 at HEAD `cf381447`)
- Risk scan: `risk-pattern-report.md`
- Coverage: pre-edit copy, `go test -coverpkg ./internal/journal,./internal/riskbucket` (journal untagged, riskbucket `tossos_testseams`); rows by `analysis/harness/preedit_rows.py`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| target order, cumulative, actual | a registered BUY order of one owner | journal rows of that owner | semantic → ErrRiskBucketReplayMismatch (fill kept + latch); storage → error (tx rolled back) |

## Branches and early returns

| Branch | Position | Condition and first body statement (AST source line) | a066 relevance | Coverage (5.7 post-edit) |
|---|---|---|---|---|
| B1 | if at 762:2 | `if err != nil \|\| cumulative == 0 {`; then `return riskbucket.FillState{}, riskbucket.FillEvent{}, fmt.Errorf("%w: non-integral cumulative fill", ErrRiskBucketReplayMismatch)` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B2 | if at 767:2 | `if err := tx.QueryRowContext(ctx, `SELECT risk_overage_latched,unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, target.account, target.market, target.symbol, target.prospective).Scan(&ownerOverage, &ownerUnknown); err != nil {`; then `return state, riskbucket.FillEvent{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B3 | if at 773:2 | `if err != nil {`; then `return state, riskbucket.FillEvent{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B4 | for at 777:2 | `for rows.Next() {`; then `var decisionID, d, v, pv, limit, held, filled, overage string` (line last changed by `c60fee07`) | a066 commit | covered |
| B5 | if at 780:3 | `if err := rows.Scan(&decisionID, &d, &v, &pv, &limit, &held, &filled, &overage, &ol, &ul); err != nil {`; then `rows.Close()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B6 | if at 785:3 | `if !isRiskBucketDimension(key.Dimension) {`; then `rows.Close()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B7 | if at 789:3 | `if decisionBuckets[decisionID] == nil {`; then `decisionBuckets[decisionID] = map[riskbucket.BucketKey]bool{}` (line last changed by `4a364caf`) | a066 commit | covered |
| B8 | if at 792:3 | `if decisionBuckets[decisionID][key] {`; then `rows.Close()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B9 | if at 798:3 | `if usage.LimitMinor == "" {`; then `usage.LimitMinor = limit` (line last changed by `4a364caf`) | a066 commit | covered |
| B10 | else at 800:10 | `} else {`; then `currentLimit, currentErr := parseRiskMinor(usage.LimitMinor)` (line last changed by `4a364caf`) | a066 commit | covered |
| B11 | if at 803:4 | `if currentErr != nil \|\| candidateErr != nil {`; then `rows.Close()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B12 | if at 807:4 | `if candidateLimit.Cmp(currentLimit) < 0 {`; then `usage.LimitMinor = candidateLimit.String()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B13 | if at 813:3 | `if addErr != nil {`; then `rows.Close()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B14 | if at 818:3 | `if addErr != nil {`; then `rows.Close()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B15 | if at 823:3 | `if addErr != nil {`; then `rows.Close()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B16 | if at 827:3 | `if usage.Latches == nil {`; then `usage.Latches = map[riskbucket.Latch]bool{}` (line last changed by `4a364caf`) | a066 commit | covered |
| B17 | if at 834:2 | `if err := rows.Close(); err != nil {`; then `return state, riskbucket.FillEvent{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B18 | range at 837:2 | `for _, seen := range decisionBuckets {`; then `if len(seen) != len(riskbucket.RequiredDimensionOrder()) {` (line last changed by `4a364caf`) | a066 commit | covered |
| B19 | if at 838:3 | `if len(seen) != len(riskbucket.RequiredDimensionOrder()) {`; then `return state, riskbucket.FillEvent{}, fmt.Errorf("%w: aggregate decision bucket count", ErrRiskBucketReplayMismatch)` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B20 | if at 842:2 | `if len(decisionBuckets) == 0 \|\| len(state.Buckets) != len(riskbucket.RequiredDimensionOrder()) {`; then `return state, riskbucket.FillEvent{}, fmt.Errorf("%w: fill bucket count", ErrRiskBucketReplayMismatch)` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B21 | if at 848:2 | `if err != nil {`; then `return state, riskbucket.FillEvent{}, err` (line last changed by `00000000`) | a066: 5.7: shared usage unreadable → semantic error (fill kept, REPLAY_MISMATCH) or storage error | covered |
| B22 | if at 853:2 | `if err != nil {`; then `return state, riskbucket.FillEvent{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B23 | for at 858:2 | `for orders.Next() {`; then `var orderKey, orderID, quote, base, digest string` (line last changed by `c60fee07`) | a066 commit | covered |
| B24 | if at 861:3 | `if err := orders.Scan(&orderKey, &orderID, &quantity, &watermark, &quote, &base, &digest); err != nil {`; then `orders.Close()` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B25 | if at 866:3 | `if err != nil {`; then `orders.Close()` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B26 | range at 871:3 | `for key := range reserved {`; then `transferred[key] = "0"` (line last changed by `c60fee07`) | a066 commit | covered |
| B27 | if at 874:3 | `if previousKey := brokerOrderIDs[orderID]; previousKey != "" && previousKey != orderKey {`; then `orders.Close()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B28 | if at 882:2 | `if err := orders.Close(); err != nil {`; then `return state, riskbucket.FillEvent{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B29 | if at 886:2 | `if err != nil {`; then `return state, riskbucket.FillEvent{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B30 | for at 889:2 | `for fills.Next() {`; then `var fillID, orderKey string` (line last changed by `c60fee07`) | a066 commit | covered |
| B31 | if at 893:3 | `if err := fills.Scan(&fillID, &orderKey, &cum, &delta, &actualKnown); err != nil {`; then `fills.Close()` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B32 | if at 898:3 | `if orderIdentity == "" {`; then `fills.Close()` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B33 | if at 905:3 | `if err != nil {`; then `fills.Close()` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B34 | for at 909:3 | `for alloc.Next() {`; then `var d, v, pv, transfer, filled string` (line last changed by `c60fee07`) | a066 commit | covered |
| B35 | if at 911:4 | `if err := alloc.Scan(&d, &v, &pv, &transfer, &filled); err != nil {`; then `alloc.Close()` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B36 | if at 920:4 | `if err != nil {`; then `alloc.Close()` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B37 | if at 927:3 | `if err := alloc.Close(); err != nil {`; then `fills.Close()` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B38 | if at 934:2 | `if err := fills.Close(); err != nil {`; then `return state, riskbucket.FillEvent{}, err` (line last changed by `c60fee07`) | a066 commit | NOT covered |
| B39 | if at 938:2 | `if len(reserved) != len(riskbucket.RequiredDimensionOrder()) {`; then `return state, riskbucket.FillEvent{}, fmt.Errorf("%w: target order reservation", ErrRiskBucketReplayMismatch)` (line last changed by `4a364caf`) | a066 commit | NOT covered |
| B40 | if at 942:2 | `if err != nil {`; then `return state, riskbucket.FillEvent{}, err` (line last changed by `4a364caf`) | a066 commit | NOT covered |

5.7 post-edit (HEAD `54e67495` + 5.7 working tree): 39 → 40: new B21 — `riskBucketSharedUsage` fills `state.SharedUsedMinor`; its error returns (semantic errors are already typed inside the helper). Pre-edit table: `analysis/pre-edit/5.7/internal-journal--loadriskbucketfilltransition.md`.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| owner/reservation/order/fill queries | rebuild the owner-wide fill state | as above | AST |

## State mutations and fallbacks

- No broker call; the caller's transaction owns every write.

## Safety conclusion

- Safe edit boundary (5.7): after the bucket-count check, fill `state.SharedUsedMinor` from `riskBucketSharedUsage` (other entries' usage via `riskbucket.ReadJournalBucketUsage`, the same function as the admission check and the production reader). Invalid ledger rows become a semantic error (fill kept, REPLAY_MISMATCH latch); read failures stay storage errors.
- High-risk impact: yes — fill accounting (체결 반영). The fill and Position are never rejected by this change.

# Pre-edit branch table: `internal-journal--loadriskbucketfilltransition` (before a066 5.7 shared overage)

| Branch | Position | Condition (AST source line at `cf381447`) | Coverage (pre-edit) |
|---|---|---|---|
| B1 | if at 740:2 | `if err != nil \|\| cumulative == 0 {` | NOT covered |
| B2 | if at 745:2 | `if err := tx.QueryRowContext(ctx, `SELECT risk_overage_latched,unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospectiv` | NOT covered |
| B3 | if at 751:2 | `if err != nil {` | NOT covered |
| B4 | for at 755:2 | `for rows.Next() {` | covered |
| B5 | if at 758:3 | `if err := rows.Scan(&decisionID, &d, &v, &pv, &limit, &held, &filled, &overage, &ol, &ul); err != nil {` | NOT covered |
| B6 | if at 763:3 | `if !isRiskBucketDimension(key.Dimension) {` | NOT covered |
| B7 | if at 767:3 | `if decisionBuckets[decisionID] == nil {` | covered |
| B8 | if at 770:3 | `if decisionBuckets[decisionID][key] {` | NOT covered |
| B9 | if at 776:3 | `if usage.LimitMinor == "" {` | covered |
| B10 | else at 778:10 | `} else {` | NOT covered |
| B11 | if at 781:4 | `if currentErr != nil \|\| candidateErr != nil {` | NOT covered |
| B12 | if at 785:4 | `if candidateLimit.Cmp(currentLimit) < 0 {` | NOT covered |
| B13 | if at 791:3 | `if addErr != nil {` | NOT covered |
| B14 | if at 796:3 | `if addErr != nil {` | NOT covered |
| B15 | if at 801:3 | `if addErr != nil {` | NOT covered |
| B16 | if at 805:3 | `if usage.Latches == nil {` | covered |
| B17 | if at 812:2 | `if err := rows.Close(); err != nil {` | NOT covered |
| B18 | range at 815:2 | `for _, seen := range decisionBuckets {` | covered |
| B19 | if at 816:3 | `if len(seen) != len(riskbucket.RequiredDimensionOrder()) {` | NOT covered |
| B20 | if at 820:2 | `if len(decisionBuckets) == 0 \|\| len(state.Buckets) != len(riskbucket.RequiredDimensionOrder()) {` | NOT covered |
| B21 | if at 824:2 | `if err != nil {` | NOT covered |
| B22 | for at 829:2 | `for orders.Next() {` | covered |
| B23 | if at 832:3 | `if err := orders.Scan(&orderKey, &orderID, &quantity, &watermark, &quote, &base, &digest); err != nil {` | NOT covered |
| B24 | if at 837:3 | `if err != nil {` | NOT covered |
| B25 | range at 842:3 | `for key := range reserved {` | covered |
| B26 | if at 845:3 | `if previousKey := brokerOrderIDs[orderID]; previousKey != "" && previousKey != orderKey {` | NOT covered |
| B27 | if at 853:2 | `if err := orders.Close(); err != nil {` | NOT covered |
| B28 | if at 857:2 | `if err != nil {` | NOT covered |
| B29 | for at 860:2 | `for fills.Next() {` | covered |
| B30 | if at 864:3 | `if err := fills.Scan(&fillID, &orderKey, &cum, &delta, &actualKnown); err != nil {` | NOT covered |
| B31 | if at 869:3 | `if orderIdentity == "" {` | NOT covered |
| B32 | if at 876:3 | `if err != nil {` | NOT covered |
| B33 | for at 880:3 | `for alloc.Next() {` | covered |
| B34 | if at 882:4 | `if err := alloc.Scan(&d, &v, &pv, &transfer, &filled); err != nil {` | NOT covered |
| B35 | if at 891:4 | `if err != nil {` | NOT covered |
| B36 | if at 898:3 | `if err := alloc.Close(); err != nil {` | NOT covered |
| B37 | if at 905:2 | `if err := fills.Close(); err != nil {` | NOT covered |
| B38 | if at 909:2 | `if len(reserved) != len(riskbucket.RequiredDimensionOrder()) {` | NOT covered |
| B39 | if at 913:2 | `if err != nil {` | NOT covered |

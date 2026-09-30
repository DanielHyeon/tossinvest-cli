# Function Logic Map: `loadRiskBucketFillTransition`

- Source: `internal/journal/risk_bucket_fill.go` (`766`–`953`)
- Qualified: `loadRiskBucketFillTransition`
- AST evidence: `ast.json` (`source_sha256` d154c6845d9f613b…) — **편집 뒤**(구현 로트, 격리 워크트리). 편집 전 판은 `467322df`
- Risk scan: `risk-pattern-report.md`
- AST branches 40 · return 30 · 호출 72

**역할.** 체결 하나를 적용하기 전 owner 의 fill 상태(주문 · 체결 기록 · 배분 · 예약 · 한도 · 공유 사용량)를 원장에서 재구성한다. 각 체결의
「actual 해소」를 `actual_known=1 OR evidence 존재` 로 판정한다 — 해소 정의의 **정본**.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `target` · `cumulativeRaw` · `actual` | 등록 주문 · 누적 수량 · actual 증거(선택) | 호출자 | 파싱 · 대조 실패는 ReplayMismatch 류로 되던짐 |
| 주문 · fill · 배분 · 예약 · snapshot | 원장 행 | `risk_bucket_*` | 불일치는 `ErrRiskBucketReplayMismatch` |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 편집 뒤 커버리지(`analysis/impl/coverage-post-edit.out`, `-coverpkg=./internal/riskbucket,./internal/journal`)로 만들었다. 「창의 return」은 위치다.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:768` `if err != nil \|\| cumulative == 0 {` | :769 | 아니오 |
| B2 | if | `:773` `if err := tx.QueryRowContext(ctx, `SELECT risk_overage_latched,unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND mar…` | :774 | 아니오 |
| B3 | if | `:779` `if err != nil {` | :780 | 아니오 |
| B4 | for | `:783` `for rows.Next() {` | — | 예 |
| B5 | if | `:786` `if err := rows.Scan(&decisionID, &d, &v, &pv, &limit, &held, &filled, &overage, &ol, &ul); err != nil {` | :788 | 아니오 |
| B6 | if | `:791` `if !isRiskBucketDimension(key.Dimension) {` | :793 | 아니오 |
| B7 | if | `:795` `if decisionBuckets[decisionID] == nil {` | — | 예 |
| B8 | if | `:798` `if decisionBuckets[decisionID][key] {` | :800 | 예 |
| B9 | if | `:804` `if usage.LimitMinor == "" {` | — | 예 |
| B10 | else | `:806` `} else {` | — | 예 |
| B11 | if | `:809` `if currentErr != nil \|\| candidateErr != nil {` | :811 | 아니오 |
| B12 | if | `:813` `if candidateLimit.Cmp(currentLimit) < 0 {` | — | 예 |
| B13 | if | `:819` `if addErr != nil {` | :821 | 아니오 |
| B14 | if | `:824` `if addErr != nil {` | :826 | 아니오 |
| B15 | if | `:829` `if addErr != nil {` | :831 | 아니오 |
| B16 | if | `:833` `if usage.Latches == nil {` | — | 예 |
| B17 | if | `:840` `if err := rows.Close(); err != nil {` | :841 | 예 |
| B18 | range | `:843` `for _, seen := range decisionBuckets {` | — | 예 |
| B19 | if | `:844` `if len(seen) != len(riskbucket.RequiredDimensionOrder()) {` | :845 | 아니오 |
| B20 | if | `:848` `if len(decisionBuckets) == 0 \|\| len(state.Buckets) != len(riskbucket.RequiredDimensionOrder()) {` | :849 | 예 |
| B21 | if | `:854` `if err != nil {` | :855 | 예 |
| B22 | if | `:859` `if err != nil {` | :860 | 아니오 |
| B23 | for | `:864` `for orders.Next() {` | — | 예 |
| B24 | if | `:867` `if err := orders.Scan(&orderKey, &orderID, &quantity, &watermark, &quote, &base, &digest); err != nil {` | :869 | 아니오 |
| B25 | if | `:872` `if err != nil {` | :874 | 아니오 |
| B26 | range | `:877` `for key := range reserved {` | — | 예 |
| B27 | if | `:880` `if previousKey := brokerOrderIDs[orderID]; previousKey != "" && previousKey != orderKey {` | :882 | 예 |
| B28 | if | `:888` `if err := orders.Close(); err != nil {` | :889 | 예 |
| B29 | if | `:892` `if err != nil {` | :893 | 아니오 |
| B30 | for | `:895` `for fills.Next() {` | — | 예 |
| B31 | if | `:899` `if err := fills.Scan(&fillID, &orderKey, &cum, &delta, &actualKnown); err != nil {` | :901 | 아니오 |
| B32 | if | `:904` `if orderIdentity == "" {` | :906 | 아니오 |
| B33 | if | `:911` `if err != nil {` | :913 | 아니오 |
| B34 | for | `:915` `for alloc.Next() {` | — | 예 |
| B35 | if | `:917` `if err := alloc.Scan(&d, &v, &pv, &transfer, &filled); err != nil {` | :920 | 아니오 |
| B36 | if | `:926` `if err != nil {` | :929 | 아니오 |
| B37 | if | `:933` `if err := alloc.Close(); err != nil {` | :935 | 예 |
| B38 | if | `:940` `if err := fills.Close(); err != nil {` | :941 | 예 |
| B39 | if | `:944` `if len(reserved) != len(riskbucket.RequiredDimensionOrder()) {` | :945 | 아니오 |
| B40 | if | `:948` `if err != nil {` | :949, :952 | 아니오 |

## Calls and live bindings

원장 질의(`tx.QueryContext` · `tx.QueryRowContext`) · 금액 도우미 · `riskBucketSharedUsage`(공유 사용량 → `ReadJournalBucketUsage`). 브로커 호출 없음.
원장 오류와 재구성 불일치를 되던진다.

## State mutations and fallbacks

없다 — 읽기 · 재구성.

## Safety conclusion

- **Safe edit boundary (a126 — a066 결함 수리, 편집 뒤 — 상수는 파일 머리(import 뒤)에 두어 다른 함수의 범위에 들지 않게 함)**: fill 조회 SQL(`:885`)의 해소 조각을 공유 상수 `riskBucketFillActualResolvedSQL` 로 바꾼다 —
  **문자열 동일**(`actual_known=1 OR EXISTS(… evidence …)`), 분기 · 동작 무변. 목적은 해제 검사와 같은 한 규칙을 구조로 묶는 것.
- **High-risk impact**: yes — 체결 · overage 재계산의 입력.

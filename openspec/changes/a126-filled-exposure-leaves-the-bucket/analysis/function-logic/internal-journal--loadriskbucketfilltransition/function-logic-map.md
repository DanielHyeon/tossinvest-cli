# Function Logic Map: `loadRiskBucketFillTransition`

- Source: `internal/journal/risk_bucket_fill.go` (`760`–`947`)
- Qualified: `loadRiskBucketFillTransition`
- AST evidence: `ast.json` (`source_sha256` d8d3cefcb0667911…) — **편집 전**(base `a189e74f`)
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

> 표는 `analysis/harness/branch_table.py` 가 만들었다(편집 전 커버리지).

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:762` `if err != nil \|\| cumulative == 0 {` | :763 | 아니오 |
| B2 | if | `:767` `if err := tx.QueryRowContext(ctx, `SELECT risk_overage_latched,unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND mar…` | :768 | 아니오 |
| B3 | if | `:773` `if err != nil {` | :774 | 아니오 |
| B4 | for | `:777` `for rows.Next() {` | — | 예 |
| B5 | if | `:780` `if err := rows.Scan(&decisionID, &d, &v, &pv, &limit, &held, &filled, &overage, &ol, &ul); err != nil {` | :782 | 아니오 |
| B6 | if | `:785` `if !isRiskBucketDimension(key.Dimension) {` | :787 | 아니오 |
| B7 | if | `:789` `if decisionBuckets[decisionID] == nil {` | — | 예 |
| B8 | if | `:792` `if decisionBuckets[decisionID][key] {` | :794 | 예 |
| B9 | if | `:798` `if usage.LimitMinor == "" {` | — | 예 |
| B10 | else | `:800` `} else {` | — | 예 |
| B11 | if | `:803` `if currentErr != nil \|\| candidateErr != nil {` | :805 | 아니오 |
| B12 | if | `:807` `if candidateLimit.Cmp(currentLimit) < 0 {` | — | 예 |
| B13 | if | `:813` `if addErr != nil {` | :815 | 아니오 |
| B14 | if | `:818` `if addErr != nil {` | :820 | 아니오 |
| B15 | if | `:823` `if addErr != nil {` | :825 | 아니오 |
| B16 | if | `:827` `if usage.Latches == nil {` | — | 예 |
| B17 | if | `:834` `if err := rows.Close(); err != nil {` | :835 | 예 |
| B18 | range | `:837` `for _, seen := range decisionBuckets {` | — | 예 |
| B19 | if | `:838` `if len(seen) != len(riskbucket.RequiredDimensionOrder()) {` | :839 | 아니오 |
| B20 | if | `:842` `if len(decisionBuckets) == 0 \|\| len(state.Buckets) != len(riskbucket.RequiredDimensionOrder()) {` | :843 | 예 |
| B21 | if | `:848` `if err != nil {` | :849 | 예 |
| B22 | if | `:853` `if err != nil {` | :854 | 아니오 |
| B23 | for | `:858` `for orders.Next() {` | — | 예 |
| B24 | if | `:861` `if err := orders.Scan(&orderKey, &orderID, &quantity, &watermark, &quote, &base, &digest); err != nil {` | :863 | 아니오 |
| B25 | if | `:866` `if err != nil {` | :868 | 아니오 |
| B26 | range | `:871` `for key := range reserved {` | — | 예 |
| B27 | if | `:874` `if previousKey := brokerOrderIDs[orderID]; previousKey != "" && previousKey != orderKey {` | :876 | 예 |
| B28 | if | `:882` `if err := orders.Close(); err != nil {` | :883 | 예 |
| B29 | if | `:886` `if err != nil {` | :887 | 아니오 |
| B30 | for | `:889` `for fills.Next() {` | — | 예 |
| B31 | if | `:893` `if err := fills.Scan(&fillID, &orderKey, &cum, &delta, &actualKnown); err != nil {` | :895 | 아니오 |
| B32 | if | `:898` `if orderIdentity == "" {` | :900 | 아니오 |
| B33 | if | `:905` `if err != nil {` | :907 | 아니오 |
| B34 | for | `:909` `for alloc.Next() {` | — | 예 |
| B35 | if | `:911` `if err := alloc.Scan(&d, &v, &pv, &transfer, &filled); err != nil {` | :914 | 아니오 |
| B36 | if | `:920` `if err != nil {` | :923 | 아니오 |
| B37 | if | `:927` `if err := alloc.Close(); err != nil {` | :929 | 예 |
| B38 | if | `:934` `if err := fills.Close(); err != nil {` | :935 | 예 |
| B39 | if | `:938` `if len(reserved) != len(riskbucket.RequiredDimensionOrder()) {` | :939 | 아니오 |
| B40 | if | `:942` `if err != nil {` | :943, :946 | 아니오 |


## Calls and live bindings

원장 질의(`tx.QueryContext` · `tx.QueryRowContext`) · 금액 도우미 · `riskBucketSharedUsage`(공유 사용량 → `ReadJournalBucketUsage`). 브로커 호출 없음.
원장 오류와 재구성 불일치를 되던진다.

## State mutations and fallbacks

없다 — 읽기 · 재구성.

## Safety conclusion

- **Safe edit boundary (a126 — a066 결함 수리)**: fill 조회 SQL(`:885`)의 해소 조각을 공유 상수 `riskBucketFillActualResolvedSQL` 로 바꾼다 —
  **문자열 동일**(`actual_known=1 OR EXISTS(… evidence …)`), 분기 · 동작 무변. 목적은 해제 검사와 같은 한 규칙을 구조로 묶는 것.
- **High-risk impact**: yes — 체결 · overage 재계산의 입력.

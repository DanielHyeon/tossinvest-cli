# Function Logic Map: `IssuePrecheckedQFinalEntry`

- Source: `internal/execgw/riskguardian_qfinal.go` (187-274, 편집 뒤)
- Qualified: `RiskGuardian.IssuePrecheckedQFinalEntry`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 3e5cc7b86303b458…, 편집 전 분기 수 11) — 현재 branches 12 · returns 13
- Risk scan: `risk-pattern-report.md`

**역할.** q_final 진입 발급 문(봉인 precheck → 원자 발급).

**a136 편집.** collect 클로저의 ReserveRequest 에 `DailyBuyLimit: g.dailyBuyLimit()` 한 칸. (재추출 2026-10-14: a137 a1cfe0c2 가 같은 파일을 고쳐 소스 해시만 바뀜 — 이 함수의 분기 · 반환 · 호출 모양은 그대로.)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `precheck` | 봉인 완료 | PrecheckQFinalEntry | qFinalRefusal |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g == nil \|\| precheck.decision.QFinal == 0 \|\| precheck.request.Collect == nil \|\| precheck.admission.TransactionID == "" {` (if 188) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B2 | `if (err != nil && !legacyTestIdentity) \|\| (!legacyTestIdentity && !sameAccountBaseFX(precheck.accountFX, accountFX)) {` (if 194) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B3 | `if err != nil {` (if 198) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B4 | `if !legacyTestIdentity {` (if 202) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B5 | `if err != nil {` (if 204) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B6 | `if err != nil {` (if 226) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B7 | `if !legacyTestIdentity && !sameQFinalFXAuthority(accountFX, policy) {` (if 229) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B8 | `if err != nil {` (if 240) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B9 | `if !verdict.Allowed {` (if 244) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B10 | `if err != nil {` (if 248) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B11 | `if err != nil {` (if 252) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |
| B12 | `if err != nil {` (if 266) | — | 분기 본문 참조 | `TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling` · 기존 발급 시험군 |

Exact AST return positions: 189 · 195 · 199 · 205 · 227 · 230 · 241 · 245 · 249 · 253 · 255 · 267 · 269

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.journal.RecordQFinalDecisionAndReserveWithRecollection` | 원자 발급 | issuanceRefusal | AST |

## State mutations and fallbacks

- journal tx.

## Safety conclusion

- Safe edit boundary: 요청 칸 하나.
- High-risk impact: yes

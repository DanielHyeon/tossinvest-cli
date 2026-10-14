# Function Logic Map: `NewRiskGuardian`

- Source: `internal/execgw/riskguardian.go` (259-314, 편집 뒤)
- Qualified: `NewRiskGuardian`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 88cf7c183018aae8…, 편집 전 분기 수 11) — 현재 branches 11 · returns 8
- Risk scan: `risk-pattern-report.md`

**역할.** Guardian 생성 · 검증.

**a136 편집.** 구조체에 `manualBuys: opts.ManualBuys` 한 줄.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch {` (switch 261) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B2 | `case opts.Journal == nil:` (case 262) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B3 | `case account == "":` (case 264) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B4 | `case strings.TrimSpace(opts.PolicyVersion) == "":` (case 266) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B5 | `if err := opts.Policy.Validate(); err != nil {` (if 270) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B6 | `if !opts.Costs.Configured() {` (if 273) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B7 | `if err != nil {` (if 278) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B8 | `if err != nil {` (if 282) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B9 | `if clk == nil {` (if 287) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B10 | `if ttl <= 0 {` (if 291) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |
| B11 | `if newID == nil {` (if 295) | — | 분기 본문 참조 | `TestAGuardianRefusesToExistWithoutWhatItNeeds` · `TestA136AnUnknownManualBuySnapshotRefusesTheBuy` |

Exact AST return positions: 263 · 265 · 267 · 271 · 274 · 279 · 283 · 298

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ExposureLimitsFor` · `EncodeLimits` | 스냅숏 | 오류 | AST |

## State mutations and fallbacks

- 없음(구성만).

## Safety conclusion

- Safe edit boundary: 칸 · 등록 하나 추가 — 기존 분기 무변경.
- High-risk impact: yes(Guardian)

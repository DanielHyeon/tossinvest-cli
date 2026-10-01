# Function Logic Map (편집 전): `validProductionRiskPolicyContents`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `38de0b7d846b0a1af1b01bc24eb94adc673ca953a3f125f3130adbc3bb4f58c2`
- Signature: `validProductionRiskPolicyContents(params=1, results=1)`
- Source range: `236:1`–`281:2`
- AST evidence: `ast.json` — 편집 **전**(a112 6.1, Manager 판정 (C) 2026-10-01).

## Inputs and invariants

- 서명 검증을 통과한 정책 본문 하나. false 면 `verifyProductionRiskPolicy` 가 거절 → `LoadProductionRiskSnapshotAuthority` B5 → `ErrProductionRiskSnapshotUnavailable`(시장 위험 미준비 — 결함).
- 편집 전에는 `Strategies[]` 항목의 risk_id 가 어느 family 의 레인을 가리키는지 보지 않는다 — 서로 다른 family 의 레인이 같은 risk_id 를 공유해도 수락.

## Branches and early returns

- Exact AST return nodes: `239:3, 243:4, 248:4, 252:3, 260:4, 263:4, 270:4, 273:4, 276:4, 280:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 238:2 | `if !canonicalIdentity(fee.Version) // !canonicalRiskDigest(fee.Digest) {` |
| B2 | range | 241:2 | `for _, raw := range []string{fee.FixedBaseMinor, fee.PerUnitBaseMinor, fee.MinimumBaseMinor} {` |
| B3 | if | 242:3 | `if _, err := parseDecimal(raw, true, 256); err != nil {` |
| B4 | range | 246:2 | `for _, raw := range []string{body.MarketLimitMinor, body.HorizonLimits[HorizonShort], body.HorizonLimits[HorizonMedium]}` |
| B5 | if | 247:3 | `if _, err := parseMinor(raw, 256); err != nil {` |
| B6 | if | 251:2 | `if len(body.HorizonLimits) != 2 // len(body.Strategies) == 0 // len(body.Symbols) == 0 {` |
| B7 | range | 255:2 | `for _, value := range body.Strategies {` |
| B8 | if | 257:3 | `if strategyKeys[key] // !canonicalIdentity(value.LaneID) // !canonicalIdentity(value.LaneVersion) //` |
| B9 | if | 262:3 | `if _, err := parseMinor(value.LimitMinor, 256); err != nil {` |
| B10 | range | 268:2 | `for _, value := range body.Symbols {` |
| B11 | if | 269:3 | `if symbols[value.Symbol] // value.Symbol == "" // value.Symbol != strings.ToUpper(strings.TrimSpace(value.Symbol)) // !c` |
| B12 | if | 272:3 | `if _, err := parseMinor(value.SectorLimitMinor, 256); err != nil {` |
| B13 | if | 275:3 | `if _, err := parseMinor(value.SymbolLimitMinor, 256); err != nil {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `canonicalIdentity` | 238:6 |
| `canonicalRiskDigest` | 238:41 |
| `parseDecimal` | 242:16 |
| `parseMinor` | 247:16 |
| `len` | 251:5 |
| `len` | 251:37 |
| `len` | 251:66 |
| `string` | 256:63 |
| `canonicalIdentity` | 257:28 |
| `canonicalIdentity` | 257:64 |
| `canonicalIdentity` | 258:74 |
| `canonicalIdentity` | 259:5 |
| `parseMinor` | 262:16 |
| `strings.ToUpper` | 269:69 |
| `strings.TrimSpace` | 269:85 |
| `canonicalIdentity` | 269:122 |
| `parseMinor` | 272:16 |
| `parseMinor` | 275:16 |

## State mutations and fallbacks

- 순수 함수.

## Safety conclusion

- High-risk(위험 정책). 6.1 편집은 거절만 더한다(family 해소 불가 레인 · 한 risk_id 를 두 family 가 공유).

# Function Logic Map: `validProductionRiskPolicyContents`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `3aa9b66c00cdcdedf09e0bea1b0eeeaf56d46d0ba40149f28edf70e73d7e26b4`
- Signature: `validProductionRiskPolicyContents(params=1, results=1)`
- Source range: `237:1`–`290:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 6.1).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- risk_id 는 한 family 를 함의한다(설계 문장 「family 가 strategy risk bucket key 에 포함」의 정정된 해석 — design 정정 사슬). family 정본은 strategyrouter 표 하나.
- 매니페스트 형식 불변(v1) — family 필드 추가(판정 (A))는 v2 가 다른 이유로 필요할 때 묶음.

## Branches and early returns

- Exact AST return nodes: `240:3, 244:4, 249:4, 253:3, 264:4, 268:4, 272:4, 279:4, 282:4, 285:4, 289:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 239:2 | 편집 전과 같은 분기 |
| B2 | range | 242:2 | 편집 전과 같은 분기 |
| B3 | if | 243:3 | 편집 전과 같은 분기 |
| B4 | range | 247:2 | 편집 전과 같은 분기 |
| B5 | if | 248:3 | 편집 전과 같은 분기 |
| B6 | if | 252:2 | 편집 전과 같은 분기 |
| B7 | range | 259:2 | 편집 전과 같은 분기 |
| B8 | if | 261:3 | 편집 전과 같은 분기 |
| B9 | if | 267:3 | **(새)** family 해소 불가 레인 · 한 risk_id 를 두 family 가 공유 → false(정책 거절 — `ErrProductionRiskSnapshotUnavailable`, 범위 국소 아님) |
| B10 | if | 271:3 | 편집 전과 같은 분기(번호 이동) |
| B11 | range | 277:2 | 편집 전과 같은 분기(번호 이동) |
| B12 | if | 278:3 | 편집 전과 같은 분기(번호 이동) |
| B13 | if | 281:3 | 편집 전과 같은 분기(번호 이동) |
| B14 | if | 284:3 | 편집 전과 같은 분기(번호 이동) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `canonicalIdentity` | 239:6 |
| `canonicalRiskDigest` | 239:41 |
| `parseDecimal` | 243:16 |
| `parseMinor` | 248:16 |
| `len` | 252:5 |
| `len` | 252:37 |
| `len` | 252:66 |
| `string` | 260:63 |
| `canonicalIdentity` | 261:28 |
| `canonicalIdentity` | 261:64 |
| `canonicalIdentity` | 262:74 |
| `canonicalIdentity` | 263:5 |
| `strategyrouter.ProductionLaneFamily` | 266:20 |
| `strategyrouter.Market` | 266:56 |
| `parseMinor` | 271:16 |
| `strings.ToUpper` | 278:69 |
| `strings.TrimSpace` | 278:85 |
| `canonicalIdentity` | 278:122 |
| `parseMinor` | 281:16 |
| `parseMinor` | 284:16 |

## State mutations and fallbacks

- 순수 함수 — 상태 없음.

## Safety conclusion

- High-risk(위험 정책). 거절만 더한다 — 새로 통과하는 입력 0. 생산 서명 위험 정책 0건(활성화 0)이라 생산 동작 변화 0.

# Function Logic Map: `validProductionRiskPolicyContents`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `9a74db4abd523da823e02e716258d428d3f385647c3d18121e744f9fac45119e`
- Signature: `validProductionRiskPolicyContents(params=1, results=1)`
- Source range: `266:1`–`319:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 6.1).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- risk_id 는 한 family 를 함의한다(설계 문장 「family 가 strategy risk bucket key 에 포함」의 정정된 해석 — design 정정 사슬). family 정본은 strategyrouter 표 하나.
- 매니페스트 형식 불변(v1) — family 필드 추가(판정 (A))는 v2 가 다른 이유로 필요할 때 묶음.

## Branches and early returns

- Exact AST return nodes: `269:3, 273:4, 278:4, 282:3, 293:4, 297:4, 301:4, 308:4, 311:4, 314:4, 318:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 268:2 | 편집 전과 같은 분기 |
| B2 | range | 271:2 | 편집 전과 같은 분기 |
| B3 | if | 272:3 | 편집 전과 같은 분기 |
| B4 | range | 276:2 | 편집 전과 같은 분기 |
| B5 | if | 277:3 | 편집 전과 같은 분기 |
| B6 | if | 281:2 | 편집 전과 같은 분기 |
| B7 | range | 288:2 | 편집 전과 같은 분기 |
| B8 | if | 290:3 | 편집 전과 같은 분기 |
| B9 | if | 296:3 | **(새)** family 해소 불가 레인 · 한 risk_id 를 두 family 가 공유 → false(정책 거절 — `ErrProductionRiskSnapshotUnavailable`, 범위 국소 아님) |
| B10 | if | 300:3 | 편집 전과 같은 분기(번호 이동) |
| B11 | range | 306:2 | 편집 전과 같은 분기(번호 이동) |
| B12 | if | 307:3 | 편집 전과 같은 분기(번호 이동) |
| B13 | if | 310:3 | 편집 전과 같은 분기(번호 이동) |
| B14 | if | 313:3 | 편집 전과 같은 분기(번호 이동) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `canonicalIdentity` | 268:6 |
| `canonicalRiskDigest` | 268:41 |
| `parseDecimal` | 272:16 |
| `parseMinor` | 277:16 |
| `len` | 281:5 |
| `len` | 281:37 |
| `len` | 281:66 |
| `string` | 289:63 |
| `canonicalIdentity` | 290:28 |
| `canonicalIdentity` | 290:64 |
| `canonicalIdentity` | 291:74 |
| `canonicalIdentity` | 292:5 |
| `strategyrouter.ProductionLaneFamily` | 295:20 |
| `strategyrouter.Market` | 295:56 |
| `parseMinor` | 300:16 |
| `strings.ToUpper` | 307:69 |
| `strings.TrimSpace` | 307:85 |
| `canonicalIdentity` | 307:122 |
| `parseMinor` | 310:16 |
| `parseMinor` | 313:16 |

## State mutations and fallbacks

- 순수 함수 — 상태 없음.

## Safety conclusion

- High-risk(위험 정책). 거절만 더한다 — 새로 통과하는 입력 0. 생산 서명 위험 정책 0건(활성화 0)이라 생산 동작 변화 0.

a112 게이트 준비(2026-10-04): a127 82080177(전략 권한 적재기가 현재 원장을 읽음)이 같은 파일을 편집해 이 번들이 낡았다 — 분기 · return · 호출 구조 동일

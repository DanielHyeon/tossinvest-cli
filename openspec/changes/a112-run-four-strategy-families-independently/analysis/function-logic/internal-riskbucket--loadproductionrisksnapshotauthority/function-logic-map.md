# Function Logic Map: `LoadProductionRiskSnapshotAuthority`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `3aa9b66c00cdcdedf09e0bea1b0eeeaf56d46d0ba40149f28edf70e73d7e26b4`
- Signature: `LoadProductionRiskSnapshotAuthority(params=3, results=2)`
- Source range: `145:1`–`180:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 판정 무변(Manager 조건 ①): 거절되는 입력 집합은 편집 전과 같다 — 감싸기 동사만 바뀌었다(편집 전 번들 `lot-5.2.2.2-fix/pre-edit/`).

## Branches and early returns

- Exact AST return nodes: `147:3, 150:3, 158:3, 162:3, 166:3, 171:3, 175:3, 179:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 146:2 | ctx · 관측 시각 부재 |
| B2 | if | 149:2 | ctx 종료 |
| B3 | if | 155:2 | 구성 · 소유자 · 경로 · digest 형식 |
| B4 | if | 161:2 | 매니페스트 파일 · digest 불일치 |
| B5 | if | 165:2 | 매니페스트 해석 · 서명 검증 실패 |
| B6 | if | 170:2 | bind 실패 → `ErrProductionRiskSnapshotUnavailable` + **원인 %w**(범위 국소 sentinel 보존) |
| B7 | if | 174:2 | 원장 항목 적재 실패 → `ErrProductionRiskSnapshotUnavailable` + **원인 %w**(latch sentinel · 원장 결함 원인 보존) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `config.ObservedAt.IsZero` | 146:19 |
| `ctx.Err` | 149:12 |
| `canonicalProductionRiskConfig` | 152:11 |
| `productionRiskOwnerUID` | 153:20 |
| `ProductionRiskPolicyFileName` | 154:10 |
| `filepath.IsAbs` | 155:32 |
| `filepath.IsAbs` | 155:69 |
| `canonicalRiskDigest` | 156:4 |
| `canonicalIdentity` | 156:51 |
| `len` | 156:93 |
| `canonicalIdentity` | 157:4 |
| `canonicalCurrency` | 157:44 |
| `readProductionRiskFile` | 160:15 |
| `filepath.Join` | 160:38 |
| `productionRiskDigest` | 161:19 |
| `decodeProductionRiskPolicy` | 164:19 |
| `verifyProductionRiskPolicy` | 165:20 |
| `bindProductionRiskInputs` | 169:33 |
| `fmt.Errorf` | 171:41 |
| `loadProductionRiskEntries` | 173:18 |
| `fmt.Errorf` | 175:41 |
| `newRiskSnapshotAuthorityService` | 178:13 |
| `service.Load` | 179:9 |

## State mutations and fallbacks

- 원장 읽기 전용(`mode=ro` · `query_only`), 쓰기 없음.

## Safety conclusion

- High-risk(위험 권한). 신원 배관만 — 새로 통과 · 새로 거절 0.

a112 6.1: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

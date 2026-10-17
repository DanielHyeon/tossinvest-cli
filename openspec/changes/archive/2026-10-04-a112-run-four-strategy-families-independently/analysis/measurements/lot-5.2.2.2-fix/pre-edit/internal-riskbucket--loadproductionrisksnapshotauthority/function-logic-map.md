# Function Logic Map (편집 전): `LoadProductionRiskSnapshotAuthority`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `c5e0c64b2de6524804c4b0fbab5fb8596191379d764c68c319905ee92aa87ed2`
- Signature: `LoadProductionRiskSnapshotAuthority(params=3, results=2)`
- Source range: `138:1`–`172:2`
- AST evidence: `ast.json` — 편집 **전**(a112 5.2.2.2 리뷰 수리 로트, J4 = (A) 판정 2026-10-01).

## Inputs and invariants

- 이 로트의 편집은 **오류 신원 배관**뿐이다 — 갈래 판정(어느 입력이 거절되는가)은 바꾸지 않는다(Manager 조건 ①).

## Branches and early returns

- Exact AST return nodes: `140:3, 143:3, 151:3, 155:3, 159:3, 163:3, 167:3, 171:2`.

| Branch | AST kind | Source location | Condition (source line) | Edit |
|---|---|---|---|---|
| B1 | if | 139:2 | `if ctx == nil \\|\\| config.ObservedAt.IsZero() {` | 판정 갈래 — 편집 안 함 |
| B2 | if | 142:2 | `if err := ctx.Err(); err != nil {` | 판정 갈래 — 편집 안 함 |
| B3 | if | 148:2 | `if !ownerOK \\|\\| name == "" \\|\\| !filepath.IsAbs(config.ConfigDir) \\|\\| !filepath.IsAbs(config.JournalPath) \\|\\|` | 판정 갈래 — 편집 안 함 |
| B4 | if | 154:2 | `if err != nil \\|\\| productionRiskDigest(data) != config.ManifestDigest {` | 판정 갈래 — 편집 안 함 |
| B5 | if | 158:2 | `if err != nil \\|\\| !verifyProductionRiskPolicy(manifest, config) {` | 판정 갈래 — 편집 안 함 |
| B6 | if | 162:2 | `if err != nil {` | **편집 대상** — bind 실패를 `fmt.Errorf("%w: %v", ErrProductionRiskSnapshotUnavailable, err)` 로 감쌈: 내부 원인의 신원(%v)이 평탄화되어 범위 국소 사유와 결함을 가를 수 없음 |
| B7 | if | 166:2 | `if err != nil {` | **편집 대상** — 원장 항목 적재 실패를 같은 `%w: %v` 로 감쌈(원장 결함의 신원 소실) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `config.ObservedAt.IsZero` | 139:19 |
| `ctx.Err` | 142:12 |
| `canonicalProductionRiskConfig` | 145:11 |
| `productionRiskOwnerUID` | 146:20 |
| `ProductionRiskPolicyFileName` | 147:10 |
| `filepath.IsAbs` | 148:32 |
| `filepath.IsAbs` | 148:69 |
| `canonicalRiskDigest` | 149:4 |
| `canonicalIdentity` | 149:51 |
| `len` | 149:93 |
| `canonicalIdentity` | 150:4 |
| `canonicalCurrency` | 150:44 |
| `readProductionRiskFile` | 153:15 |
| `filepath.Join` | 153:38 |
| `productionRiskDigest` | 154:19 |
| `decodeProductionRiskPolicy` | 157:19 |
| `verifyProductionRiskPolicy` | 158:20 |
| `bindProductionRiskInputs` | 161:33 |
| `fmt.Errorf` | 163:41 |
| `loadProductionRiskEntries` | 165:18 |
| `fmt.Errorf` | 167:41 |
| `newRiskSnapshotAuthorityService` | 170:13 |
| `service.Load` | 171:9 |

## State mutations and fallbacks

- 원장은 읽기 전용(`mode=ro`, `query_only`). 쓰기 없음.

## Safety conclusion

- High-risk(위험 권한). 편집 전 모든 실패는 `ErrProductionRiskSnapshotUnavailable` 하나로 접힌다 — 엔진이 범위 국소 거절과 원장 결함을 구별할 수 없는 근본 원인.

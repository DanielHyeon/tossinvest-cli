# Function Logic Map: `LoadProductionRiskSnapshotAuthority`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `38de0b7d846b0a1af1b01bc24eb94adc673ca953a3f125f3130adbc3bb4f58c2`
- Signature: `LoadProductionRiskSnapshotAuthority(params=3, results=2)`
- Source range: `144:1`–`179:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 판정 무변(Manager 조건 ①): 거절되는 입력 집합은 편집 전과 같다 — 감싸기 동사만 바뀌었다(편집 전 번들 `lot-5.2.2.2-fix/pre-edit/`).

## Branches and early returns

- Exact AST return nodes: `146:3, 149:3, 157:3, 161:3, 165:3, 170:3, 174:3, 178:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 145:2 | ctx · 관측 시각 부재 |
| B2 | if | 148:2 | ctx 종료 |
| B3 | if | 154:2 | 구성 · 소유자 · 경로 · digest 형식 |
| B4 | if | 160:2 | 매니페스트 파일 · digest 불일치 |
| B5 | if | 164:2 | 매니페스트 해석 · 서명 검증 실패 |
| B6 | if | 169:2 | bind 실패 → `ErrProductionRiskSnapshotUnavailable` + **원인 %w**(범위 국소 sentinel 보존) |
| B7 | if | 173:2 | 원장 항목 적재 실패 → `ErrProductionRiskSnapshotUnavailable` + **원인 %w**(latch sentinel · 원장 결함 원인 보존) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `config.ObservedAt.IsZero` | 145:19 |
| `ctx.Err` | 148:12 |
| `canonicalProductionRiskConfig` | 151:11 |
| `productionRiskOwnerUID` | 152:20 |
| `ProductionRiskPolicyFileName` | 153:10 |
| `filepath.IsAbs` | 154:32 |
| `filepath.IsAbs` | 154:69 |
| `canonicalRiskDigest` | 155:4 |
| `canonicalIdentity` | 155:51 |
| `len` | 155:93 |
| `canonicalIdentity` | 156:4 |
| `canonicalCurrency` | 156:44 |
| `readProductionRiskFile` | 159:15 |
| `filepath.Join` | 159:38 |
| `productionRiskDigest` | 160:19 |
| `decodeProductionRiskPolicy` | 163:19 |
| `verifyProductionRiskPolicy` | 164:20 |
| `bindProductionRiskInputs` | 168:33 |
| `fmt.Errorf` | 170:41 |
| `loadProductionRiskEntries` | 172:18 |
| `fmt.Errorf` | 174:41 |
| `newRiskSnapshotAuthorityService` | 177:13 |
| `service.Load` | 178:9 |

## State mutations and fallbacks

- 원장 읽기 전용(`mode=ro` · `query_only`), 쓰기 없음.

## Safety conclusion

- High-risk(위험 권한). 신원 배관만 — 새로 통과 · 새로 거절 0.

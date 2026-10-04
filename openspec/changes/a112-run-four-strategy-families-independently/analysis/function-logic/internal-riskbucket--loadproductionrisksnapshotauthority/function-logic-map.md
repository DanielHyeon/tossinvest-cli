# Function Logic Map: `LoadProductionRiskSnapshotAuthority`

- **a112 재추출(2026-10-04 게이트 준비).** a127 `82080177`(전략 권한 적재기가 현재 원장을 읽음)이 이 함수의 분기 구조를 바꿔 a112 번들이 낡았다 — 현재 AST 와 같은 a127 아카이브 번들(`openspec/changes/archive/2026-10-01-a127-strategy-authorities-read-the-current-ledger/analysis/function-logic/internal-riskbucket--loadproductionrisksnapshotauthority/`)을 옮겨 왔다(그 판의 RED · 변이 경로는 아카이브 좌표로 고쳐 씀). a127 이전 a112 판(a112 의 편집 기록)은 `git show cc79c887:openspec/changes/a112-run-four-strategy-families-independently/analysis/function-logic/internal-riskbucket--loadproductionrisksnapshotauthority/function-logic-map.md` · `branch-test-map.md`.

- Source: `internal/riskbucket/production_snapshot_authority.go` (`170`–`209`)
- Qualified: `LoadProductionRiskSnapshotAuthority`
- AST evidence: `ast.json` (`source_sha256` 9a74db4abd523da8…) — **편집 뒤**(구현 로트, 커버리지 `openspec/changes/archive/2026-10-01-a127-strategy-authorities-read-the-current-ledger/analysis/impl/coverage-post-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 8 · return 9 · 호출 24

**역할.** 위험 snapshot 권한의 공개 적재기 — config 정규화 · 정책 파일 digest · 서명 검증 · 입력 결속(`bindProductionRiskInputs`, 섹터 매핑 없음은 범위 국소 거절) 뒤 원장 적재(`loadProductionRiskEntries`)를 부르고 원인을 `%w: %w` 로 보존해 감쌈. a127: 원장 스키마 주입 누락(0 이하)을 **정책 결속보다 앞**에서 거절하는 가드를 둠(design D2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` · `config.ObservedAt` | nil 아님 · 0 아님 | 호출자 | B1 거절(Unavailable) |
| config 경로 · digest · 키 · 계좌 · 통화 | 절대 경로 · 정규 digest · ed25519 · 정규 식별자 | engine 로더 | B3 거절 |
| 정책 파일 | 소유자 · 0400 · 1 MiB · digest 일치 · 서명 | config dir | B4 · B5 거절 |
| (a127) 원장 스키마 주입 값 | > 0(= `journal.SchemaVersion`) | engine 호출 자리 | 정책 결속 앞 거절(편집 뒤) |

## Branches and early returns

> 분기 표는 `openspec/changes/archive/2026-10-01-a127-strategy-authorities-read-the-current-ledger/analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:171` `if ctx == nil \|\| config.ObservedAt.IsZero() {` | :172 | 아니오 |
| B2 | if | `:174` `if err := ctx.Err(); err != nil {` | :175 | 예 |
| B3 | if | `:178` `if config.JournalSchemaVersion <= 0 {` | :179 | 예 |
| B4 | if | `:184` `if !ownerOK \|\| name == "" \|\| !filepath.IsAbs(config.ConfigDir) \|\| !filepath.IsAbs(config.JournalPath) \|\|` | :187 | — |
| B5 | if | `:190` `if err != nil \|\| productionRiskDigest(data) != config.ManifestDigest {` | :191 | 예 |
| B6 | if | `:194` `if err != nil \|\| !verifyProductionRiskPolicy(manifest, config) {` | :195 | 예 |
| B7 | if | `:199` `if err != nil {` | :200 | 예 |
| B8 | if | `:203` `if err != nil {` | :204, :208 | 예 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `config.ObservedAt.IsZero` | 171:19 |
| `ctx.Err` | 174:12 |
| `fmt.Errorf` | 179:41 |
| `canonicalProductionRiskConfig` | 181:11 |
| `productionRiskOwnerUID` | 182:20 |
| `ProductionRiskPolicyFileName` | 183:10 |
| `filepath.IsAbs` | 184:32 |
| `filepath.IsAbs` | 184:69 |
| `canonicalRiskDigest` | 185:4 |
| `canonicalIdentity` | 185:51 |
| `len` | 185:93 |
| `canonicalIdentity` | 186:4 |
| `canonicalCurrency` | 186:44 |
| `readProductionRiskFile` | 189:15 |
| `filepath.Join` | 189:38 |
| `productionRiskDigest` | 190:19 |
| `decodeProductionRiskPolicy` | 193:19 |
| `verifyProductionRiskPolicy` | 194:20 |
| `bindProductionRiskInputs` | 198:33 |
| `fmt.Errorf` | 200:41 |
| `loadProductionRiskEntries` | 202:18 |
| `fmt.Errorf` | 204:41 |
| `newRiskSnapshotAuthorityService` | 207:13 |
| `service.Load` | 208:9 |

(a127 판 서술) `canonicalProductionRiskConfig` · `productionRiskOwnerUID` · `readProductionRiskFile` · `decodeProductionRiskPolicy` · `verifyProductionRiskPolicy` · `bindProductionRiskInputs` · `loadProductionRiskEntries` · `newRiskSnapshotAuthorityService(...).Load`. 원장 · 브로커 쓰기 없음.

## State mutations and fallbacks

없음 — 읽기 전용. 실패는 전부 거절(`ErrProductionRiskSnapshotUnavailable` 감쌈, 범위 국소면 `ErrProductionRiskScopeRefused` 도 보존).

## Safety conclusion

- **Safe edit boundary**: 주입 가드 하나 추가(정책 결속 앞). 나머지 분기 판정 · 순서 · 오류 신원 불변.
- **High-risk impact**: yes — 진입 경로 위험 권한의 입구.

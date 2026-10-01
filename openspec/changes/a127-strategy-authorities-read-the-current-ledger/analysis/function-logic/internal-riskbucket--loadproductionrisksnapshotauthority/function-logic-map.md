# Function Logic Map: `LoadProductionRiskSnapshotAuthority`

- Source: `internal/riskbucket/production_snapshot_authority.go` (`145`–`180`)
- Qualified: `LoadProductionRiskSnapshotAuthority`
- AST evidence: `ast.json` (`source_sha256` 3aa9b66c00cdcded…) — **편집 전**(base `de3b4f65` 의 바이트, 커버리지 `analysis/impl/coverage-pre-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 7 · return 8 · 호출 23

**역할.** 위험 snapshot 권한의 공개 적재기 — config 정규화 · 정책 파일 digest · 서명 검증 · 입력 결속(`bindProductionRiskInputs`, 섹터 매핑 없음은 범위 국소 거절) 뒤 원장 적재(`loadProductionRiskEntries`)를 부르고 원인을 `%w: %w` 로 보존해 감쌈. a127: 원장 스키마 주입 누락(0 이하)을 **정책 결속보다 앞**에서 거절하는 가드를 둠(design D2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` · `config.ObservedAt` | nil 아님 · 0 아님 | 호출자 | B1 거절(Unavailable) |
| config 경로 · digest · 키 · 계좌 · 통화 | 절대 경로 · 정규 digest · ed25519 · 정규 식별자 | engine 로더 | B3 거절 |
| 정책 파일 | 소유자 · 0400 · 1 MiB · digest 일치 · 서명 | config dir | B4 · B5 거절 |
| (a127) 원장 스키마 주입 값 | > 0(= `journal.SchemaVersion`) | engine 호출 자리 | 정책 결속 앞 거절(편집 뒤) |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:146` `if ctx == nil \|\| config.ObservedAt.IsZero() {` | :147 | 아니오 |
| B2 | if | `:149` `if err := ctx.Err(); err != nil {` | :150 | 예 |
| B3 | if | `:155` `if !ownerOK \|\| name == "" \|\| !filepath.IsAbs(config.ConfigDir) \|\| !filepath.IsAbs(config.JournalPath) \|\|` | :158 | — |
| B4 | if | `:161` `if err != nil \|\| productionRiskDigest(data) != config.ManifestDigest {` | :162 | 예 |
| B5 | if | `:165` `if err != nil \|\| !verifyProductionRiskPolicy(manifest, config) {` | :166 | 예 |
| B6 | if | `:170` `if err != nil {` | :171 | 예 |
| B7 | if | `:174` `if err != nil {` | :175, :179 | 예 |

## Calls and live bindings

`canonicalProductionRiskConfig` · `productionRiskOwnerUID` · `readProductionRiskFile` · `decodeProductionRiskPolicy` · `verifyProductionRiskPolicy` · `bindProductionRiskInputs` · `loadProductionRiskEntries` · `newRiskSnapshotAuthorityService(...).Load`. 원장 · 브로커 쓰기 없음.

## State mutations and fallbacks

없음 — 읽기 전용. 실패는 전부 거절(`ErrProductionRiskSnapshotUnavailable` 감쌈, 범위 국소면 `ErrProductionRiskScopeRefused` 도 보존).

## Safety conclusion

- **Safe edit boundary**: 편집 전 — a127 은 B3 앞(또는 B1 뒤)에 주입 가드 하나를 더할 예정. 기존 분기 B1~B7 의 판정 · 순서는 바꾸지 않음.
- **High-risk impact**: yes — 진입 경로 위험 권한의 입구.

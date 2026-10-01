# Function Logic Map: `LoadProductionRouteAuthorityBatch`

- Source: `internal/strategyrouter/production.go` (`318`–`411`)
- Qualified: `LoadProductionRouteAuthorityBatch`
- AST evidence: `ast.json` (`source_sha256` 7d60a867a87576ca…) — **편집 전**(base `de3b4f65` 의 바이트, 커버리지 `analysis/impl/coverage-pre-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 16 · return 14 · 호출 50

**역할.** route 권한의 공개 적재기 — 매니페스트 digest · 서명 검증 뒤 원장을 열고(`openProductionRouteSnapshot`) 서명 범위마다 owner snapshot 을 재구성해 봉인. **편집 전**: B7 `:351-352` 가 opener 의 어떤 오류든 `fmt.Errorf("%w: owner snapshot", ErrProductionRouteUnavailable)` 로 바꿔 원인을 지움. a127: 주입 가드(원장 열기 전) · opener 에 주입 값 전달 · 원인 보존(D2 · D3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` · `ObservedAt` | nil 아님 · 0 아님 | 호출자 | B1 거절 |
| config 경로 · 키 · 계좌 · 시장 | 절대 경로 · ed25519 · 정규 식별자 | engine 로더 | B3 거절 |
| 매니페스트 | digest · 서명 · 범위 ≥ 1 | config dir | B4~B6 거절 |
| 원장 snapshot | opener 성공 | `openProductionRouteSnapshot` | B7 거절(원인 지움 — 편집 전) |
| owner 재구성 | revision == 매니페스트 OwnerRevision | `loadProductionRouteOwnersFrom` | B13 거절 |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:319` `if ctx == nil \|\| config.ObservedAt.IsZero() {` | :320 | 아니오 |
| B2 | if | `:322` `if err := ctx.Err(); err != nil {` | :323 | 예 |
| B3 | if | `:329` `if !ownerOK \|\| name == "" \|\| !filepath.IsAbs(config.ConfigDir) \|\| !filepath.IsAbs(config.JournalPath) \|\|` | :334 | — |
| B4 | if | `:337` `if err != nil \|\| productionRouteDigest(data) != config.ManifestDigest {` | :338 | 예 |
| B5 | if | `:341` `if err != nil \|\| len(manifest.Scopes) == 0 {` | :342 | 아니오 |
| B6 | if | `:347` `if _, verified := verifyProductionRouteManifest(manifest, verificationConfig); !verified {` | :348 | 예 |
| B7 | if | `:351` `if err != nil {` | :352 | 아니오 |
| B8 | if | `:364` `if err != nil \|\| EvaluateMarketLifecycle(record, config.ObservedAt) != LifecycleReady {` | :365 | 아니오 |
| B9 | range | `:368` `for _, target := range targets {` | — | 예 |
| B10 | if | `:369` `if err := ctx.Err(); err != nil {` | :370 | 아니오 |
| B11 | if | `:373` `if !found {` | — | 예 |
| B12 | if | `:377` `if err != nil {` | :378 | 아니오 |
| B13 | if | `:381` `if err != nil \|\| revision != scope.OwnerRevision {` | :382 | 예 |
| B14 | if | `:385` `if err != nil {` | :386 | 아니오 |
| B15 | range | `:389` `for _, value := range scope.Candidates {` | — | 예 |
| B16 | if | `:407` `if err := tx.Commit(); err != nil {` | :408, :410 | 예 |

## Calls and live bindings

`canonicalProductionRouteConfig` · `productionRouteOwnerUID` · 매니페스트 읽기 · `decodeProductionRouteManifest` · `verifyProductionRouteManifest` · `openProductionRouteSnapshot` · `newMarketRecord` · `EvaluateMarketLifecycle` · `validProductionRouteScopes` · `NewOwnerKey` · `loadProductionRouteOwnersFrom` · `newOwnerSnapshot` · 봉인 함수 · `tx.Commit`. 원장 쓰기 없음(읽기 전용 tx).

## State mutations and fallbacks

없음 — 읽기 전용 tx 를 열고 닫음.

## Safety conclusion

- **Safe edit boundary**: 편집 전 — a127 은 주입 가드(열기 전) · opener 호출 인자 · B7 감싸기의 원인 보존만 바꿀 예정. 범위 · 재구성 · 봉인 판정 불변.
- **High-risk impact**: yes — 진입 경로 route 권한.

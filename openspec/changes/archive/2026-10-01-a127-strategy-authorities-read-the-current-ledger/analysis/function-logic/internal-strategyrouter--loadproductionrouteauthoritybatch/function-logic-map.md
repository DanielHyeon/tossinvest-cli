# Function Logic Map: `LoadProductionRouteAuthorityBatch`

- Source: `internal/strategyrouter/production.go` (`329`–`427`)
- Qualified: `LoadProductionRouteAuthorityBatch`
- AST evidence: `ast.json` (`source_sha256` 617163a508030dea…) — **편집 뒤**(구현 로트, 커버리지 `analysis/impl/coverage-post-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 17 · return 15 · 호출 51

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
| B1 | if | `:330` `if ctx == nil \|\| config.ObservedAt.IsZero() {` | :331 | 아니오 |
| B2 | if | `:333` `if err := ctx.Err(); err != nil {` | :334 | 예 |
| B3 | if | `:337` `if config.JournalSchemaVersion <= 0 {` | :338 | 예 |
| B4 | if | `:344` `if !ownerOK \|\| name == "" \|\| !filepath.IsAbs(config.ConfigDir) \|\| !filepath.IsAbs(config.JournalPath) \|\|` | :349 | — |
| B5 | if | `:352` `if err != nil \|\| productionRouteDigest(data) != config.ManifestDigest {` | :353 | 예 |
| B6 | if | `:356` `if err != nil \|\| len(manifest.Scopes) == 0 {` | :357 | 아니오 |
| B7 | if | `:362` `if _, verified := verifyProductionRouteManifest(manifest, verificationConfig); !verified {` | :363 | 예 |
| B8 | if | `:366` `if err != nil {` | :368 | 예 |
| B9 | if | `:380` `if err != nil \|\| EvaluateMarketLifecycle(record, config.ObservedAt) != LifecycleReady {` | :381 | 아니오 |
| B10 | range | `:384` `for _, target := range targets {` | — | 예 |
| B11 | if | `:385` `if err := ctx.Err(); err != nil {` | :386 | 아니오 |
| B12 | if | `:389` `if !found {` | — | 예 |
| B13 | if | `:393` `if err != nil {` | :394 | 아니오 |
| B14 | if | `:397` `if err != nil \|\| revision != scope.OwnerRevision {` | :398 | 예 |
| B15 | if | `:401` `if err != nil {` | :402 | 아니오 |
| B16 | range | `:405` `for _, value := range scope.Candidates {` | — | 예 |
| B17 | if | `:423` `if err := tx.Commit(); err != nil {` | :424, :426 | 예 |

## Calls and live bindings

`canonicalProductionRouteConfig` · `productionRouteOwnerUID` · 매니페스트 읽기 · `decodeProductionRouteManifest` · `verifyProductionRouteManifest` · `openProductionRouteSnapshot` · `newMarketRecord` · `EvaluateMarketLifecycle` · `validProductionRouteScopes` · `NewOwnerKey` · `loadProductionRouteOwnersFrom` · `newOwnerSnapshot` · 봉인 함수 · `tx.Commit`. 원장 쓰기 없음(읽기 전용 tx).

## State mutations and fallbacks

없음 — 읽기 전용 tx 를 열고 닫음.

## Safety conclusion

- **Safe edit boundary**: 주입 가드 추가, opener 에 주입 값 전달, B7 감싸기가 원인을 `%w` 로 보존. 나머지 불변.
- **High-risk impact**: yes — 진입 경로 route 권한.

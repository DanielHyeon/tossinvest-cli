# Function Logic Map: `strategyRouteAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_route_authority.go` (`141`–`221`)
- Qualified: `strategyRouteAuthorityLoader.collectMarket`
- AST evidence: `ast.json` (`source_sha256` 85e97cf96bd416ee…) — **편집 뒤**(구현 로트, 커버리지 `analysis/impl/coverage-post-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 13 · return 11 · 호출 64

**역할.** schedule · candidate 권한 위에서 route 매니페스트 환경값을 읽고 `loader.load`(= `strategyrouter.LoadProductionRouteAuthorityBatch`)를 부름. 오류는 버리고 `StrategyRouteAuthorityInvalid` 로 접음(B8). a127: config 리터럴에 원장 스키마 주입 값(상수 선택자)을 넣음(D2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| schedule · candidate | Ready · 활성화 | 상위 로더 | B1~B3 거절 |
| 로더 배선 · 키 환경값 | 존재 · ed25519 | 환경 | B4 · B5 거절 |
| Batch 결과 | 오류 없음 · digest 일치 | strategyrouter | B8 거절(오류 버림) |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:148` `if !schedule.snapshot.Ready \|\| schedule.restore.Activation == nil {` | :149 | 예 |
| B2 | if | `:151` `if !candidates.snapshot.Ready {` | :152 | 예 |
| B3 | if | `:154` `if candidates.approved.Len() == 0 {` | :155 | 예 |
| B4 | if | `:157` `if loader.getenv == nil \|\| loader.load == nil \|\| loader.configDir == "" \|\| loader.journalPath == "" \|\| loader.accountRef == "" {` | :158 | 예 |
| B5 | if | `:162` `if err != nil \|\| base64.StdEncoding.EncodeToString(key) != encoded \|\| len(key) != ed25519.PublicKeySize {` | :163 | 아니오 |
| B6 | for | `:171` `for index := 0; index < candidates.approved.Len(); index++ {` | — | 예 |
| B7 | if | `:173` `if !ok \|\| !approved.Valid() \|\| approved.Market() != string(market) \|\| seen[approved.Symbol()] {` | :174 | 아니오 |
| B8 | if | `:187` `if err != nil \|\| batch.ManifestDigest() != digest {` | :188 | 예 |
| B9 | range | `:192` `for _, approved := range approvedValues {` | — | 예 |
| B10 | if | `:194` `if !ok {` | — | 예 |
| B11 | if | `:202` `if routed.Code != strategyrouter.RefusalNone \|\| !routed.Valid() \|\| len(routed.Decisions) == 0 \|\|` | :209 | — |
| B12 | if | `:210` `if len(entries) == 0 {` | :211 | 예 |
| B13 | range | `:215` `for _, entry := range entries {` | :218 | 예 |

## Calls and live bindings

`loader.getenv` · base64 · `candidates.approved.At` · `loader.load`(config 리터럴) · `batch.For` · `strategyrouter.RouteSet` · sha256.

## State mutations and fallbacks

없음 — 시장 route 권한 값을 만듦.

## Safety conclusion

- **Safe edit boundary**: config 리터럴에 `JournalSchemaVersion: journal.SchemaVersion` 한 필드. 분기 불변.
- **High-risk impact**: yes — route 권한 배선.

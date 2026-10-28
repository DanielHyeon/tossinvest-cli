# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go` (319-364)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 뒤(브랜치 `a163-impl` GREEN 로트, 2026-10-22). 편집 전 번들 문서는 커밋 `78efe1951` · 생성 입력 `../../flm-gen/spec-base.json`
- AST 분기 6 · 반환 4 · defer 0 · go 0
- 위험 등급: High-risk(생산 청산 관측자 조립 — 호출자 값을 덮어쓰는 자리)
- 출처: `archive/2026-10-22-a163-quote-freshness-uses-trade-time/analysis/function-logic/internal-app-engine--context.exitobserver` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/app/engine/exitwiring.go` = `6020a40b2030…` 이 번들의 `source_sha256`(revision current)과 같아 번들을 그대로 복사함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c` | 엔진 컨텍스트 | `NewContext` | nil · 미검증 · 비발행 Guardian 이면 거부 |
| `opts` | 호출자 옵션 | `engineRuntime` | 일부 필드는 무조건 덮음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 320 `if c == nil {` | — | `ErrExitObserverUnavailable` | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` |
| B2 | if 323 `if !c.Automation.Verified {` | — | 게이트 미검증 거부 | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` |
| B3 | if 327 `if !ok {` | — | Guardian 비발행 거부 | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` |
| B4 | if 343 `if opts.Names == nil {` | `opts.Names` 기본 | — | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` |
| B5 | if 348 `if c.Notifier != nil {` | Alerts·Announcer·Critical 덮음 | — | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` |
| B6 | if 360 `if opts.Floor == nil {` | `opts.Floor` 기본 | — | `TestA091TheProductionAssemblyReadsTheLoadedSwitch` |

반환 좌표(AST): 321 · 324 · 328 · 363

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `NewExitObserver` | 검증·구성 | 오류 반환 | AST |
| `exitSideRetrier` · `exitSideFloor` | 기록 전용 재시도기·상한 | — | AST |

## State mutations and fallbacks

- `opts` 사본 필드 덮어쓰기(무조건: Journal·Prices·Retrier·Issuer·Submit·AccountRef·CommonPolicy·NotificationsEnabled·ZeroFloorLog).

## a163 편집

- a163 3.6: 무조건 덮어쓰기 줄 `opts.SourceAgeReport = c.Config.Engine.ExitSourceAgeReport` 하나(선례 `NotificationsEnabled`). 분기 6 · 반환 4 불변(design D8). — **적용됨**, 편집 전·뒤 AST 분기·반환·defer 수 대조는 `../../flm-gen/README.md`

## Safety conclusion

- Safe edit boundary: 어느 분기 안도 아닌 대입 한 줄.
- High-risk impact: yes(배선) — 2.15 · 변이 ⑳.

# Function Logic Map: `NewContext`

- Source: `internal/app/engine/engine.go` (437-607)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 기준: 편집 뒤(브랜치 `a163-impl` GREEN 로트, 2026-10-22). 편집 전 번들 문서는 커밋 `78efe1951` · 생성 입력 `../../flm-gen/spec-base.json`
- AST 분기 15 · 반환 11 · defer 0 · go 0
- 위험 등급: High-risk(기동 인터록) — a163 은 호출 인자 하나만 바꿈
- 출처: `archive/2026-10-22-a163-quote-freshness-uses-trade-time/analysis/function-logic/internal-app-engine--newcontext` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum internal/app/engine/engine.go` = `60e3a8d45487…` 이 번들의 `source_sha256`(revision current)과 같아 번들을 그대로 복사함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts` | 엔진 옵션 | 호출자 | — |
| `cfg` | 로드된 설정(기동 1회) | `config.NewService(...).Load` | 오류 B1 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 439 `if err != nil {` | — | 설정 실패 반환 | 기존 시험 |
| B2 | if 446 `if err != nil {` | — | audit 열기 실패 반환 | 기존 시험 |
| B3 | if 450 `if clk == nil {` | 시스템 시계 | — | 기존 시험 |
| B4 | if 464 `if opts.Publisher != nil {` | 주입 publisher 우선 | — | `TestAnInjectedPublisherWins` |
| B5 | if 472 `if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notifications,` | — | audit 기록 실패 → 기동 거부 | 기존 시험 |
| B6 | if 487 `if err != nil {` | — | 계좌 미해석 거부 | 기존 시험 |
| B7 | if 494 `if err != nil {` | — | 원장 열기 실패 거부 | 기존 시험 |
| B8 | if 503 `if err := bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger)); err != nil {` | — | apply hook 실패 거부 | 기존 시험 |
| B9 | if 522 `if err != nil {` | — | gateway 실패 거부 | 기존 시험 |
| B10 | if 537 `if gate.Enabled && guardian == nil && !opts.disableProductionGuardian {` | Guardian 생성 | — | 기존 시험 |
| B11 | if 539 `if factory == nil {` | 기본 factory | — | 기존 시험 |
| B12 | if 545 `if err != nil {` | — | Guardian 실패 거부 | 기존 시험 |
| B13 | if 565 `if err != nil {` | — | 인터록 실패 반환 | 기존 시험 |
| B14 | if 569 `if !automation.Verified {` | 미검증이면 Guardian 버림 | — | 기존 시험 |
| B15 | if 574 `if err != nil {` | — | 전략 투영 실패 반환 | `TestDormantSnapshotContainsExactPairedHonestMarkets` |

반환 좌표(AST): 440 · 447 · 477 · 488 · 495 · 505 · 524 · 547 · 567 · 576 · 578

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `recordGateSettings` | 기동 audit(B5) | 오류 → 기동 거부 | AST |
| `runInterlock` | 자동화 게이트 | — | AST |

## State mutations and fallbacks

- Context 조립.

## a163 편집

- a163 3.6: B5 의 `recordGateSettings(...)` 호출에 끝 인자 `cfg.Engine.ExitSourceAgeReport` 하나. 분기 15 · 반환 11 불변. 실패 모양(B5 → 기동 거부) 불변. — **적용됨**, 편집 전·뒤 AST 분기·반환·defer 수 대조는 `../../flm-gen/README.md`

## Safety conclusion

- Safe edit boundary: 호출 인자 하나.
- High-risk impact: no 새 분기 — 2.16 · 변이 ㉑ 가 값 전달을 잰다.

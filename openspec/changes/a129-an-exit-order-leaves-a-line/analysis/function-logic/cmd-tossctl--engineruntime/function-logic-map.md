# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (**편집 뒤** `649`–`741`; 편집 전 `649`–`739`, base `381e73da`)
- Qualified: `engineRuntime`
- AST evidence: `ast.json` (**편집 뒤** `source_sha256` de6bcd1329628b5e… — 편집 전 6844d37952bf3b9f…) — branches 6 · returns 7 · defers 0 · go 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(tasks 0.5). 선행 번들: a090 아카이브 `…/analysis/function-logic/cmd-tossctl--engineruntime/`(같은 분기 6 · return 7).

**역할.** 생산 엔진의 루프 집합(대사 · exit · 체결 감지 · 전략 진입 외곽)과 보조 실행자를 조립해 감독자에 넘긴다. 각 생성자의 첫 실패에서 거절한다.
호출자(CodeGraph `callers engineRuntime`): 생산은 `engineRuntimeFactory` 변수(`engine.go`) + 시험 4.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | 조립된 엔진 Context | `assembleEngine` | 각 생성자 오류로 거절 |
| `logger` | nil 가능(시험) | `runEngine` 의 엔진 로거(`engine.log`) | nil 로거는 무출력(`obs.Logger.emit` nil 가드) |
| `clk` · `ready` | 주입 | 호출자 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:652` `if err != nil` (`engineFillDetector`) | 없음 | `return nil, err` `:653` | 무편집 |
| B2 | `:661` `if err != nil` (`ectx.ReconcileDriver`) | 없음 | `:662` | 무편집 |
| B3 | `:674` `if err != nil` (`ectx.ExitObserver`) | 없음 | `:675` | 무편집 — a129 편집은 이 호출의 **인자 리터럴** 필드 하나 |
| B4 | `:679` `if err != nil` (`ectx.Recovery`) | 없음 | `:680` | 무편집 |
| B5 | `:683` `if err != nil` (전략 진입 외곽) | 없음 | `:684` | 무편집 |
| B6 | `:695` `if err != nil` (`ectx.AlertDeliverer`) | 없음 | `:696` | 무편집 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ectx.ExitObserver` | exit 관측자 조립 — 이 change 의 대상 | 오류 → B3 | AST `:665` |
| `engineFillDetector` · `ectx.ReconcileDriver` · `ectx.Recovery` · `ectx.NewRefreshingPairedStrategyEntrySupervisor` · `ectx.AlertDeliverer` | 다른 루프 | 각 B | AST |
| `engine.NewRuntime` | 감독자 | 그대로 반환 | AST `:699` |

Live config binding: 없음(설정은 `ectx` 가 이미 읽음).

## State mutations and fallbacks

- 없음(조립만). a129: `engine.ExitObserverOptions{…}` 리터럴에 `ExitOrderLog: logger` 를 더한다. `Log` 는 넣지 않는다.

## Safety conclusion

- Safe edit boundary: 리터럴 필드 하나. 분기·순서·반환 무변경.
- High-risk impact: yes(생산 exit 루프 조립) — 기록 싱크만 더함. 구조 핀(go/parser 키 census)이 `ExitOrderLog = logger` 있음 · `Log` 없음을 잰다.

## 편집 뒤 (a129)

분기 불변. B3 이후 좌표 +2(리터럴에 주석 한 줄 + `ExitOrderLog: logger,`): B1 `:652` · B2 `:661` · B3 `:676` · B4 `:681` · B5 `:685` · B6 `:697`.

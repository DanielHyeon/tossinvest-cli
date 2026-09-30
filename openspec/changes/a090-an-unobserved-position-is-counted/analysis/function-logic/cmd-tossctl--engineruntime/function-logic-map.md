# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (**편집 뒤** `647`–`737`(a094 착지 `d485a45f` 위); 편집 전 `636`–`724`)
- Qualified: `engineRuntime`
- AST evidence: `ast.json` (**편집 뒤** `source_sha256` 5aa437eb90d87a3a… — 편집 전 aeefd5dc…, base `2f698db6`) — branches 6 · returns 7
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **구현 로트 편집 전**(tasks 1.1, base 재고정 뒤 2026-09-30). 선행 번들: a092 아카이브
  `openspec/changes/archive/2026-09-30-a092-an-alert-does-not-hold-the-stop/analysis/function-logic/cmd-tossctl--engineruntime/` — **같은 source_sha256**
  (a092 착지 뒤 이 함수를 고친 커밋 없음).

**역할.** 생산 엔진의 루프 집합(대사 · exit · 체결 감지 · 전략 진입 외곽)과 보조 실행자(배달 · 일반 등급 이관)를 조립해 감독자에 넘긴다.
각 생성자가 실패하면 첫 실패에서 거절한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ectx` | 조립된 엔진 Context | `assembleEngine` | 각 생성자 오류로 거절 |
| `logger` | nil 가능(시험) | `runEngine` 의 엔진 로거 | nil 로거는 무출력(`obs.Logger.emit` nil 가드) |
| `clk` · `ready` | 주입 | 호출자 | — |

## Branches and early returns

| Branch | 종류 | 조건 (원문) | 결과 |
|---|---|---|---|
| B1 | if | `:639` `if err != nil {` (`engineFillDetector`) | return nil, err `:640` |
| B2 | if | `:648` `if err != nil {` (`ectx.ReconcileDriver`) | return nil, err `:649` |
| B3 | if | `:659` `if err != nil {` (`ectx.ExitObserver`) | return nil, err `:660` |
| B4 | if | `:664` `if err != nil {` (`ectx.Recovery`) | return nil, err `:665` |
| B5 | if | `:668` `if err != nil {` (`ectx.NewRefreshingPairedStrategyEntrySupervisor`) | return nil, err `:669` |
| B6 | if | `:680` `if err != nil {` (`ectx.AlertDeliverer`) | return nil, err `:681` |

마지막 return `:684` `engine.NewRuntime(...)`.

## Calls and live bindings

- `ectx.ExitObserver(engine.ExitObserverOptions{Clock, Costs, SLO, Escalate})`(`:652-658`) — **a090 의 편집 자리.** 관측자의 `Log` 는
  넘기지 않는다(생산에서 nil — 관측자의 기존 줄은 생산에 나가지 않는다). `Announcer`·`Alerts` 는 `Context.ExitObserver` 가 기록 전용으로 덮는다
  (`internal/app/engine/exitwiring.go:348-351`). 그 밖의 옵션(`UnobservedLog` 포함)은 **그대로 통과**한다 — `Context.ExitObserver` 는
  Journal·Prices·Retrier·Issuer·Submit·AccountRef·CommonPolicy·Names·Alerts·Announcer·Floor 만 채운다(`exitwiring.go:336-354`).
- 나머지: `engineFillDetector` · `ectx.ReconcileDriver` · `ectx.Recovery` · `ectx.NewRefreshingPairedStrategyEntrySupervisor` · `ectx.AlertDeliverer` ·
  `engine.NewRuntime` · `recoverThenReady` · `engineRecoverySequence` · `engineRecoveryObserver` · `ectx.NormalAlertRelayExecutor`.

## State mutations and fallbacks

- 없음(조립만). 로거는 공유 포인터.

## Safety conclusion

- **Safe edit boundary(a090 design D7 · D8)**: `ExitObserverOptions` 합성 리터럴에 **`UnobservedLog: logger` 한 줄**. `Log` 는 넣지 않는다(4판 R2-1 —
  관측자의 기존 줄은 `obs.FieldAccount` 로 계좌 원문을 싣는다, `exitloop.go:1735`). 분기·반환 무변화.
- **High-risk impact**: yes(조립) — 로그 싱크 하나를 새 normal 줄 전용으로 배선. 주문·판정 경로 무변화.

## 편집 뒤 (a090 구현 로트, 2026-09-30)

- 편집: `ExitObserverOptions` 리터럴에 `UnobservedLog: logger`(주석 1 + 줄 1). 분기 6 · 반환 7 동일 — difflib 재번호 불변. 좌표는 a094 의 파일 앞쪽 편집(+11)과 이 편집(+2, B3 이후)을 합쳐
  B1 639→650 · B2 648→659 · B3 659→672 · B4 664→677 · B5 668→681 · B6 680→693.

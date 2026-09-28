# Function Logic Map: `ExitObserver.ObserveOnce`

- Source: `internal/app/engine/exitloop.go` (`413`–`470`)
- Qualified: `ExitObserver.ObserveOnce`
- AST evidence: `ast.json` (`source_sha256` 522d5d81c4992c57…) — branches 8 · returns 5 · calls 16 · assignments 15
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **proposal 단계(구현 전)**. 이 함수의 분기를 근거로 삼는 a090 문서보다 먼저 만들었다.
- 선행 산출물: a092 의 같은 함수 번들(`openspec/changes/a092-an-alert-does-not-hold-the-stop/analysis/function-logic/
  internal-app-engine--exitobserver.observeonce/`)은 **7 분기** 판이다. 그 뒤 a111(`882a0b49`)이 `quoteUsable`
  `continue`(현 B7)를 더해 **8 분기**가 됐다. a092 의 B6(`:455`)는 이 번들의 **B6(`:453`)**, a092 의 B7(`:462` judge 오류)은
  이 번들의 **B8(`:465`)** 이다.

**역할.** 관측 한 주기: 체결 감지 양보 → 작업 집합 → 한 번의 가격 읽기 → 포지션마다 판정.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.opts.SLO.FillDetectionBehind()` | bool | 체결 감지 SLO 어댑터 | 참이면 주기 전체 양보(B1) — outage 시계는 계속 돈다 |
| `states` | 보유 포지션 중 관리 대상(열린 exit state) | `workingSet` → `Journal.Positions` · `OpenExitStateResults` | 오류면 주기 오류(B2), 비면 outage 시계 리셋(B3) |
| `quotes` | 응답한 종목 → `observedQuote`. `observe`(`:743-814`)가 `Last<=0`·NaN·Inf·요청 밖 종목·(FetchedAt 가 있을 때) 15초 초과·미래 시각을 **버린다**(`:765-775`) | `Retrier.Query(QueryPrice)` 한 번 | **하나도 안 남으면** 주기 오류(B4 — `:779-781` `NewEvidenceInvalidError`) → 계정 사다리 |
| `o.lastObserved` · `o.outageRaised` | 계정 단위 | 관측자 필드 | **하나라도 응답하면** `:447-448` 에서 리셋된다 — 나머지 종목의 미응답은 이 시계에 안 잡힌다 |
| `quoteUsable(quote)` | 시각 역행 없음 ∧ 사용 임대 ≤ `execgw.QueryPriceEvidenceDuration`(15초, `internal/execgw/retry.go:192`) | 관측자 시계 | 거짓이면 B7 무음 `continue` |

## Branches and early returns

> 조건은 소스의 그 줄 원문. 「진입 실측」 은 `analysis/harness/observeonce_entry.sh` 결과(`analysis/harness/observeonce.blocks`, commit `eac13df1`) —
> 그 줄에서 시작하는 블록의 count 가 1 이면 「예」. **무음 `continue` 두 개(`:457`·`:462`)는 AST 가 기록하지 않는다** — AST 는 분기·return·호출만 싣는다.
> 두 줄은 소스 스캔(`grep -n continue`, 413-470 범위)으로 찾았고, 그 분기 창(`:454-457`·`:460-462`)에 호출이 없음은 AST `calls` 로 확인했다.

| Branch | 종류 | 조건 (원문) | 부수 효과 / 이탈 | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:417` `if o.opts.SLO != nil && o.opts.SLO.FillDetectionBehind() {` | `cycle.Deferred=true` · `o.checkOutage` · **return** `:423` | 예 |
| B2 | if | `:427` `if err != nil {` (workingSet) | `cycle.Err=err` · **return** `:429` | **아니오** |
| B3 | if | `:431` `if len(states) == 0 {` | `lastObserved=now` · `outageRaised=false` · **return** `:439` | 예 |
| B4 | if | `:442` `if err != nil {` (observe) | `cycle.Err` · `o.checkOutage` · **return** `:445` — 전 종목 미응답의 계정 사다리 | 예 |
| B5 | range | `:451` `for _, state := range states {` | 포지션마다 정확히 1회 | 예 |
| **B6** | if | `:453` `if !ok {` | **무음 `continue` `:457`** — 로그·알림·원장 행 없음 | 예 |
| **B7** | if | `:459` `if !o.quoteUsable(quote) {` | **무음 `continue` `:462`** — 같음 | 예 |
| B8 | if | `:465` `if err := o.judge(...); err != nil && cycle.Err == nil {` | 첫 판정 오류만 `cycle.Err` | 예 |

마지막 **return** `:469`.

**B6·B7 이 a090 의 대상이다.** 두 `continue` 는 아무것도 남기지 않는다. 그런데 같은 주기에 다른 종목이 하나라도 응답했으면
`:447-448` 이 계정 outage 시계를 이미 리셋했다 — 그래서 **미응답 종목 하나는 무기한 판정되지 않고, 계정 사다리(B4)도,
주기 실패 로그(`reportCycle` `:382-388` — `cycle.Err != nil` 일 때만)도 그것을 보지 못한다.** 파일 `reportCycle` 주석(`:377-378`)는
"The conditions that mean a position is actually unprotected raise their own critical events from where they happen" 라고
약속하지만 B6·B7 은 그 자리이면서 올리지 않는다(a089 2차 리뷰 C2 의 결함, HEAD 에서 그대로).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.opts.SLO.FillDetectionBehind` | 양보 판단 | 로컬 | AST `:417` |
| `o.checkOutage` | 계정 사다리 2단(`:817-854`) — 60초(`DefaultExitObservationOutage` `:105`) 넘으면 critical `EventExitObservationOutage`(key `type|account` `:832`) + `EscalateOperatingMode(…ModeTriggerExitObservationOutage…)`(`:846-847`), 한 두절에 1회(`outageRaised`) | 로컬 + 원장 쓰기 | AST `:422`·`:444` |
| `o.workingSet` | 포지션·exit state 조정 | 원장 읽기 오류 → B2 | AST `:426` |
| `o.observe` | **유일한 브로커 호출**(`/api/v1/prices` 한 번, `Retrier` 경유) | Retrier 행렬·조회 예산 | AST `:441` |
| `o.quoteUsable` | 사용 임대 재검사(a111) | 로컬 | AST `:459` |
| `o.judge` | 판정·기록·발의 | 오류는 되던짐 → B8 | AST `:465` |

생산 호출자: `ExitObserver.Run`(`exitloop.go:354-364`)과 tracer `Run`(`internal/app/engine/tracer.go:273`) — CodeGraph 1.6.0
`codegraph callers ObserveOnce`(나머지 여섯은 시험). tracer 는 비시험 생성자 호출이 0 이고 단일 종목이라 B6 에는 닿지 않지만(한 종목 미스는 B4) **B7 에는 닿을 수 있다** — 가격 읽기 뒤 원장 작업(`:793`)이 임대를 태울 수 있다(5판 정정).

## State mutations and fallbacks

- 관측자 필드: `lastObserved`·`outageRaised`(B3·`:447-448`), `checkOutage` 안의 `outageRaised`.
- 원장: `judge` 경유만. B6·B7 은 **아무것도 쓰지 않는다**.
- 파일 머리 계약(`exitloop.go:67-68`): 관측자 필드는 경보 래치와 타이머뿐이며 재시작으로 잃으면 경보가 **다시 올라간다**(잃지 않는다).

## Safety conclusion

- **Safe edit boundary(2판)**: B6·B7 의 `continue` 직전에 원인 기록 1개씩, 판정 진입(`:464`) 표시 1개, **B3 조기 반환 앞** 순회-뒤 처리 1개(보유 대상이
  전부 `workingSet` 에서 탈락한 경우 — 그 주기도 미관측으로 센다), 순회 뒤(`:469` 앞) 처리 1개. 임계 판정·알림(enqueue-only)·강화는 **순회 뒤**에만
  — 루프 안에서는 기록만(design D4). **분기 조건·이탈·B1·B2·B4 무변화.** 새 판정 로직은 새 파일. 탈락 자리 다섯은 `workingSet` 번들에 있다.
- **High-risk impact**: yes — 손절 관측 경로. 추가는 관측·알림(과 Q1 결정에 따라 모드 강화)뿐이며 판정·발의·주문을 바꾸지 않는다.

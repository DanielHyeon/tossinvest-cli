# Function Logic Map: `ExitObserver.ObserveOnce`

- Source: `internal/app/engine/exitloop.go` (**편집 뒤** `437`–`502`(a094 착지 `b74875e7` 위); 편집 전 `413`–`470`)
- Qualified: `ExitObserver.ObserveOnce`
- AST evidence: `ast.json` (**편집 뒤** `source_sha256` 014cdcc7350d17ef… — 편집 전 2d34b5c5…, base `2f698db6`) — branches 8 · returns 5 · calls 16 · assignments 15
- Risk scan: `risk-pattern-report.md`
- **base 재고정(2026-09-30, `1ffe2295` → base `2f698db6`)**: `ast.json` 을 새 base 소스로 재추출했다. `exitloop.go` sha256 은 `2d34b5c5…`(a092 의 `checkOutage`·`alert` 편집)로 바뀌었으나 **이 함수의 AST 는 옛 base `d3bd1843` 판과 필드 단위로 같다**(start·end·branches·returns·calls·assignments 동일 — 비교 스크립트 결과) — 아래 줄 좌표·분기 번호는 그대로 유효하다.
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

## 편집 뒤 (a090 구현 로트, 2026-09-30)

- **재번호(difflib 정렬, 편집 전 base `2f698db6` ↔ 편집 뒤)**: 분기 8 · 반환 5 의 종류·순서가 같다 — 번호 불변, 좌표만 이동
  (B1 417→441 · B2 427→451 · B3 431→455 · B4 442→468 · B5 451→477 · B6 453→479 · B7 459→487 · B8 465→495). 줄 이동의 원인은 구조체
  필드 추가(`ExitObserverOptions.UnobservedLog` · `ExitCycle.Unobserved` · `ExitObserver` 기록 필드)와 이 함수 안의 호출 넷.
- **편집(분기 조건·이탈 무변화)**: B3 블록 첫 문장 `o.settleUnobserved(ctx, &cycle)` · B6 `continue` 직전 `o.noteUnobservedCause(…, unobservedNoQuote)` ·
  B7 `continue` 직전 `o.noteUnobservedCause(…, unobservedQuoteExpired)` · `cycle.Judged++` 직후 `o.noteJudged(&cycle, state)` · 순회 뒤
  `o.settleUnobserved(ctx, &cycle)`. B1·B2·B4 무편집(그 주기는 포지션 단위 처리를 하지 않는다 — design D11).
- 새 호출 셋은 전부 `exit_unobserved.go`(새 파일)의 새 함수다. 루프 안의 둘(`noteUnobservedCause` · `noteJudged`)은 관측자 메모리에만 쓴다 —
  원장·알림 없음. 순회 뒤 `settleUnobserved` 만 알림기 기록 입구(`RecordCritical`, 창 0)와 `EscalateOperatingMode` 를 부른다.

## 착지 리비전 refresh (2026-10-01, Manager 판정 D)

착지 기록 대상 = a090 의 마지막 자기 Go 커밋 `df3a6c69`. 그 사이 a094 가 같은 파일을 편집해 파일 sha 가 밀렸으므로 `ast.json` 을 `df3a6c69` 소스로 재추출했다. **함수 본문 무변** — 시작 줄 기준 상대 좌표로 정규화한 AST(분기·반환·호출·대입) 필드 단위 동일(비교 스크립트). 바뀐 것은 파일 sha(`014cdcc7…` → `aa184f13…`)뿐 — 좌표 동일.

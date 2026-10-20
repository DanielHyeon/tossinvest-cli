# a087 §0.5 — 청산 판정 상류의 세션·시각 입력 열거 (design D7 · 재리뷰 P1-4)

- 측정 순간: 2026-10-20, HEAD `00fed684c7bebbd359ab20949508a5fa11e6292a`(브랜치 `feat/a112-four-family-runtime`).
  `internal/app/engine/exitloop.go`·`internal/official/market_reads.go` 는 워킹트리 무편집(각 `ast.json` 의 `source_sha256`
  이 그 바이트에 묶여 있고 `enumerate.py` 가 실행 때 다시 대조한다).
- 도구: `tools/logic-map`(HEAD 에서 빌드) · **codegraph 1.6.0**(index 2,914 files · 50,973 nodes · 188,627 edges) · `rg`.
- 성격: read-only. production 편집 0. 이 문서는 D7 처분의 **입력**이고 처분은 사람이 정한다(tasks 0.5).

## 방법

1. 대상 함수 12개의 AST 를 `tools/logic-map` 으로 생성 — `session-gate-0.5/*.ast.json`.
   엔진 10: `ExitObserver.{Run, ObserveOnce, workingSet, observe, checkOutage, quoteUsable, judge, judgeRatchet, judgeLadder, record}`.
   시세 생산자 2: `official.Client.Prices` · `official.adaptPrices` (엔진 `opts.Prices = c.Official` — `exitwiring.go:337`).
2. `session-gate-0.5/enumerate.py` 가 **모든** 분기 좌표의 원문 줄과 **모든** 호출 텍스트를 두 정규식으로 분류한다 —
   `SESSION`(정규장·장 시간·시장 현지 시각·달력: `InRegularSession|RegularSession|TradingHours|TradingDay|Session|Location|Weekday|Hour(|Calendar|Holiday`)
   과 `CLOCK`(벽시계·경과·신선도). 출력 전문은 `session-gate-0.5/enumerate-output.md`(재생성: `python3 -I enumerate.py`).
   손으로 고른 줄이 아니라 AST 가 낸 좌표 전부다.
3. **양성 대조군**(정규식이 눈먼 것이 아님): `clock.MarketUS.InRegularSession(at)`(`verifylive/hours.go:130` 원문) →
   `SESSION`, `market.TradingDay(g.clk.Now())`(`execgw/gateway.go` 원문) → `SESSION`, `clock.MarketKR.Location()` → `SESSION`,
   `if m.identityErr != nil {` → 분류 없음. 2026-10-20 실행으로 확인.
4. 호출 사슬은 CodeGraph `callers` 로 단계별 확인(아래 표). 이름 충돌 엣지(로컬 클로저 `record`·`judge` 등)는 `gate-0.4-codegraph.md`
   의 「거짓 엣지」와 같은 종류라 제외했다.

## 호출 사슬 (CodeGraph 1.6.0, 비시험 호출자)

| 피호출 | 비시험 호출자 | 근거 |
| --- | --- | --- |
| `record` (`exitloop.go:1288`) | `judgeRatchet` (`:1025`) · `judgeLadder` (`:1079`) | `codegraph callers record` |
| `judgeRatchet` · `judgeLadder` | `judge` (`:961`) | `codegraph callers judgeRatchet` · `judgeLadder` |
| `judge` | `ObserveOnce` (`:457`) | `codegraph callers judge` |
| `ObserveOnce` | `ExitObserver.Run` (`:395`) · `tracer.Run` (`tracer.go:273`, 사람이 모는 tracer) | `codegraph callers ObserveOnce` |
| `ExitObserver.Run` | `cmd/tossctl/engine.go:733` — `engine.NewRuntime` 의 `Loops` 목록에 `{Name: "exit", Run: observer.Run}` 로 **조건 없이** 등록 | 합성 리터럴(분기 아님) — `rg` + 원문 |
| `quoteUsable` | `ObserveOnce` · `judge` · `judgeRatchet` · `judgeLadder` · `refreshObservation` · `record` | `codegraph callers quoteUsable` |

## 결과

**`SESSION` 0건 · `CLOCK` 40건 (AST 12개).** 청산 판정 상류(엔진 루프 기동 → 관측 → 판정 → 기록)의 어느 분기·호출도 정규장·장 시간·
시장 현지 시각·달력을 읽지 않는다. 시각 입력은 전부 **경과 시간**이며, 아래 다섯 갈래로 나뉜다(좌표는 `enumerate-output.md`).

| 갈래 | AST 좌표 | 읽는 것 | 세션 판정인가 |
| --- | --- | --- | --- |
| 루프 주기 | `Run` B3 `:401` `o.clk.Sleep(ctx, o.Interval())` | `DefaultExitObservationInterval = 5s`(`exitloop.go:98`) | 아니오 — 시각과 무관하게 5초마다 돈다 |
| 시세 신선도 | `observe` B5·B6/B7 `:868`·`:871-872`, `quoteUsable` B1 `:1184` + `:1187` `clock.LeaseElapsed`, `judge` B1 `:963`, `judgeRatchet` B5 `:1067`, `judgeLadder` B7 `:1138`, `record` B1 `:1291`, `ObserveOnce` B7 `:511` | `FetchedAt` 의 미래 여부·`QueryPriceEvidenceDuration = 15s`(`execgw/retry.go:192`) 경과 | 아니오 — 「가져온 지 얼마나 됐나」 |
| 관측 두절 | `checkOutage` B1·B2 `:918`·`:921` | `lastObserved`/`startedAt` 이후 `DefaultExitObservationOutage = 60s`(`:105`) | 아니오 |
| 체결 감지 양보 | `ObserveOnce` B1 `:465` `o.opts.SLO.FillDetectionBehind()` | 생산 구현 `cmd/tossctl/engine.go:795` = `detector.Health().EntryBlocked`(체결 감지 지연 SLO) | 아니오 — `internal/filldetect` 비시험 코드의 세션 이름 `rg` 0건 |
| 관측 라벨 | `ObserveOnce` `:463` `observation.at = clk.Now()`, `record` B3 `:1313` | 시세에 `FetchedAt` 이 없을 때 `ObservedAt`·`cycle:` 출처 라벨 | 아니오 — 기록 값, 분기 조건 아님 |

### 신선도가 세션 게이트 구실을 하지 않는 이유 (시세 생산자 AST)

`adaptPrices`(`official/market_reads.go:168-180`) 의 AST 는 분기 1(`range`) · 시각 호출 1(`:175` `time.Now`)이다 — 즉
`FetchedAt` 은 **로컬 수신 시각**이다. 브로커 응답의 `timestamp`(「data timestamp」, `apiPrice.Timestamp` — `market_reads.go:132·137`)는
`adaptPrices` 가 읽지 않는다(`rg -n '\.Timestamp\b' internal/official/market_reads.go` 의 적중 셋 `:311`·`:436`·`:489` 은 모두 다른 함수).
그래서 엔진의 15초 신선도 검사는 「시세가 체결된 지 15초」가 아니라 「응답을 받은 지 15초」를 잰다. 정규장 밖에 `/api/v1/prices` 가
마지막 체결가를 돌려주면 그 값은 신선한 관측으로 통과한다.

## D7 에 넘기는 결론 (처분 아님)

1. **엔진에는 세션 게이트가 없다.** design D7 의 마지막 줄 「엔진이 그 구간에 판정하지 않음이 확인되면 교집합은 공집합」은
   이 열거로는 **성립하지 않는다** — 엔진은 시각과 무관하게 5초마다 관측하고, 시세가 오면 판정한다.
2. 따라서 정규장 밖에 보호 청산이 **판정**되는지는 브로커 `/api/v1/prices` 가 그 구간에 유효 시세(`lastPrice > 0`)를 돌려주는지에만
   달린다 — `[미측정]`(이 작업은 브로커를 부르지 않았다). 시세가 오면 판정·제출 시도가 일어나고, 그 주문이 MARKET 으로 접수되는지는
   §5.2 의 측정 대상이다.
3. 교집합(MARKET 불가 ∩ LIMIT 금지)이 실재하는지는 이제 **§5.2 결과 하나**와 위 2 의 `[미측정]` 에 걸린다. spec Requirement 1 의
   SHALL NOT 유지·세션별 유형 분기 등의 처분은 사람 결정이다(design D7 「처분」).

## 한계 (본 것과 안 본 것)

- 열거 범위는 위 12 함수다. `submit` 이하(제출 경로 — `execgw` 게이트웨이)는 판정 **하류**라 AST 를 만들지 않았다. 참고로
  `execgw/gateway.go` `prepareRequest` 가 `market.TradingDay(g.clk.Now())` 를 부르지만 `clock/market.go:158-164` 는 현지 날짜 문자열을
  만들 뿐이다(오류 갈래는 `Location()` 실패뿐) — 이 문장은 손으로 읽은 것이며 이 문서의 결론 근거로 쓰지 않는다.
- 분류는 **이름 기반**이다. 세션 이름을 쓰지 않는 피호출자 내부(예: `o.opts.Journal.*`, `o.opts.Retrier.Query`)에 숨은 세션 판정은
  정규식이 못 본다. 보완으로 비시험 Go 전체에서 세션 술어(`InRegularSession|RegularSession(|TradingHours|GetTradingHours`)를 쓰는 파일을
  `rg` 로 셌다 — `internal/client/marketdata.go`·`internal/output/marketdata.go`·`cmd/tossctl/market.go`·`internal/domain/models.go`·
  `internal/clock/market.go`·`internal/verifylive/hours.go` 여섯이고, 엔진·`journal`·`execgw`·`filldetect`·`official` 은 없다.
  `exitloop.go`·`exitwiring.go` 의 `scheduler|Calendar|marketclock` 적중도 0 이다(진입 쪽 `strategy_schedule_authority.go` 의 달력은
  청산 루프에 배선되지 않는다).
- `tracer.Run` 은 사람이 모는 진단 경로라 세션 게이트 판단과 무관하다.
- 분기 행의 분류는 AST 분기 좌표의 **첫 줄**만 본다. 여러 줄에 걸친 조건의 둘째 줄 이하에 세션 **필드**(호출이 아닌 값 읽기)가
  있으면 놓칠 수 있다 — 호출은 `calls` 로 따로 잡히고(예: `observe` `:872` `validatedAt.Sub`), 위 파일 단위 `rg` 가 이 공백을 덮는다.

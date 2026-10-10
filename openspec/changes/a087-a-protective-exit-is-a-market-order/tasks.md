# a087 tasks

> **High-risk.** 손절 주문의 **유형**을 바꾼다. 구현 착수 전 Pre-Edit 선언과
> proposal-freeze 재리뷰(적대적 Eng 관점 포함)가 필요하다.
> 초안(호가 그리드)의 리뷰 결과는 `review.md`에 보존한다 — 이 tasks는 그 리뷰가
> 도달한 결론을 실행하는 것이지 우회하는 것이 아니다.
> **2026-10-10 개정**: 2차 proposal-freeze 재리뷰 13건 처분(0.6) — §1 에 셋째 관문(1.5)·종단 시험(1.6), §3 에 3.6~3.8,
> §4 에 4.5·4.6, §5 에 실행 수단과 결과 → 처분 표.
> **2026-10-10 3차 개정**: 3차 재리뷰 처분(0.8) + 사용자 결정 2차 — 1.5.x 를 **엔진 인스턴스 한정 개방**으로 다시 썼고
> (1.5.5 = upstream 동일성 고정 시험), §5 「수용」 기준에 주문 상태 확인·누락 결과 3행·실행 전제를 더했다.

## 0. 게이트 선행

- [x] 0.1 `capture_change_base.py --change a087-a-protective-exit-is-a-market-order`로 base commit 재고정 (디렉터리명이 바뀌었다) — `5491451b`(ec29dc72 → 102d4e99, 2026-09-30, 승인 기록 review.md)
- [x] 0.2 `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive` — 2026-10-10 rc=0
      ("is valid", 문서 수정 없음)
- [x] 0.3 **proposal-freeze 재리뷰** 실행 후 `review.md`에 2차 절 추가 (적대적 Eng 필수)
      — 2026-10-10 실행(Opus 독립 리뷰, `[subagent-only]`): **판정 = 수정 필요, freeze 불가.**
      P0 1·P1 4·P2 5·P3 3. P0-1(`trading/service.go` 셋째 관문 — tasks 대로면 보호 청산 100%
      로컬 거부)·P1-1(5.1 이 출하 도구로 실행 불가) 은 Manager 가 HEAD 실물 대조로 확인.
      P0-1·P1-1~4 를 문서로 닫고 Eng 재리뷰 재실행 전 §1~§3 착수 금지. P0-1 처분
      (internal/trading 수정 = 사람 CLI upstream 동작 동반 변경)은 **사람 결정**.
      → 이후 정정: 사용자 결정 2차(2026-10-10)로 ③ 개방은 **엔진 인스턴스 한정**, 사람 CLI 동반 변경은 철회(1.5).
- [x] 0.4 `make sdd-sync` 후 `sellIntent`·`checkOrderShape`·`isProtective`·`buildOrderCreate`의
      definition/callers/impact 확인 — 2026-10-10 HEAD `6e844e11`, codegraph 1.6.0. `make sdd-sync` rc=2(2회):
      CodeGraph sync 성공·fingerprint 기록, CodeGraphContext `update` 300초 타임아웃(advisory 미갱신).
      산출물 `analysis/gate-0.4-codegraph.md`
- [ ] 0.5 **엔진 세션 게이트 유무 확인 (design D7 — 재리뷰 P1-4)** — 청산 판정 상류(`ObserveOnce`·`judge`·`record`)를
      CodeGraph 로 따라가 세션·시각 입력이 있는지 열거하고 `analysis/` 에 남긴다. 결과와 §5.2 를 입력으로 D7 교집합의
      처분을 사람이 정한다. **§1~§3 착수 조건**. read-only 작업이고 그 답이 spec Requirement 1 의 SHALL NOT 을 무효화할 수
      있으므로 **0.9 freeze 판정 전에 수행하는 것을 권장**한다(3차 재리뷰 A-P3-4 — 차단 아님, Manager 판단)
- [x] 0.6 **2차 재리뷰 문서 수리** — P0-1·P1-1~4·P2-1~5·P3-1~3 처분(2026-10-10, 처분 대조표 `review.md` 끝).
      사용자 결정 2건(2026-10-10) 반영. production 편집 0
- [x] 0.7 **Eng 재리뷰 1회** (재리뷰 「freeze 재개 조건」) — 0.6 개정본 대상. 2026-10-10 실행(3차, `e93cd6a4`):
      **판정 = freeze 재개 불가**(문서 P1 2·P2 5·P3 4 / 도구 P1 1·P2 1·P3 6). 사용자 결정 2차(`77a8e119`) — I-R1 = 엔진 인스턴스 한정
- [x] 0.8 **3차 재리뷰 수리** — 3차 발견 전수(P1 3·P2 6·P3 10) 처분(처분 대조표 3차분 `review.md` 끝). 도구 변이 하네스는
      사본 실행으로 수리(`45167aa7`, 변이 13/13 CAUGHT). production 편집 0
- [ ] 0.9 **Manager 대조 + ③ 인스턴스 옵션 설계 문단 독립 확인 1회** (3차 「freeze 재개 조건」 4 — 전면 Eng 재리뷰 불요).
      통과 전 §1~§3 착수 금지

## P1. Phase 1 — 가격 사다리 KR 하한가 단 (선행 · 실측 불요 · design D2a)

> **종결 (2026-09-30, Manager 판정 선택지 1): 불구현 — 전제 반증.** P1.2 ① 열거에서 대상 모집단이 생산에서 0
> (B1·B2 도달 불가, 엔진 커버리지 0)으로 측정됐다. 영수증: `issues.md` I-P1 · design D2a 「반증」 ·
> `analysis/function-logic/internal-app-engine--exitobserver.sellintent/` · `analysis/p1-population/`.
> Phase 1 의 대체 산출물은 P1.7 핀 시험 하나다. production 편집 0.

- [x] P1.1 **Pre-Edit 선언 + FLM** — FLM/BTM 작성(`b10cefaf`). Pre-Edit 선언: not-applicable — production 편집 없음
- [x] P1.2 **RED — 모집단 열거 포함** — ① 열거 완료: 모집단 0(`issues.md` I-P1, `b10cefaf`). ②~④ not-applicable —
      세울 단이 없음
- [x] P1.3 **GREEN** — not-applicable (D2a 불구현, `PriceLimits` 배선 없음 · 새 브로커 호출 0)
- [x] P1.4 원장 기록 — not-applicable (하한가 제출 경로 없음)
- [x] P1.5 변이 + 리뷰 — P1.7 핀 시험에 대해 수행: 무변이 대조군 GREEN 후 변이 4/4 CAUGHT(`7bd8f197` 메시지).
      Manager 독립 대조로 모집단 판정 확인(2026-09-30)
- [x] P1.6 D2b 제거 예약 — 제거할 단이 없음으로 정정(design D2b 취소선 주석)
- [x] P1.7 **평가기 관측가 가드 핀 시험** — `internal/exitpolicy/a087_observed_price_pin_test.go`(`7bd8f197`):
      빈·공백·"0"·파싱 불가 관측가 → `EvaluateRatchet`·`EvaluateLadder` 모두 "observed price" 거부. B1·B2 가 닫혀 있는
      등식을 고정. FLM: not-applicable — 시험 전용(비례 원칙)

> **아래 §1~§3(Phase 2)의 착수 조건 (2026-10-10 개정)**: ① §5 실측 기록과 그 결과가 아래 §5 표의 「§1~§3 착수 가능」
> 행일 것(§0.7 사람 승인), ② 0.5 확인 + D7 처분, ③ 0.9 통과(0.7 은 재개 불가로 끝났다), ④ §5 표 아래 「US」 행 사람 확인.
> Phase 1 은 그 실측을 기다리지 않았다(종결).

## 1. 실행 관문 두 곳(②·③)을 축소 시장가에 연다 (Phase 2)

> **개정 (2026-10-10, 재리뷰 P0-1 · 사용자 결정)**: 3판의 §1 은 ②(`checkOrderShape`)만 다뤘다. ③(`placeIntentSupported`)을
> 열지 않으면 보호 청산이 100% 로컬 거부된다. 1.5 가 ③ 이다. ② 도 High-risk 기존 함수 내부 편집인데 3판에 Pre-Edit·FLM
> task 가 없었으므로 1.0·1.0.1 을 더한다(§2 와 같은 형식).

- [ ] 1.0 **Pre-Edit 선언** — `execgw.checkOrderShape` (WORKFLOW §Pre-Edit 형식)
- [ ] 1.0.1 **Function Logic Map** — `internal/execgw/failclosed.go` / `checkOrderShape`
      (`ast.json` + `function-logic-map.md` + `branch-test-map.md` + `risk-pattern-report.md`). HEAD 에서 생성.
      design D3 의 「새 분기 위치(통화·수량 검사 뒤)」·「정수 수량 검사 자리」를 이 열거로 확정
- [ ] 1.1 **RED** — `checkOrderShape` 표 테스트(②만 단독으로 — ③을 거치지 않는 직접 호출): `sell+market`(통과)·
      `buy+market`(거부)·`sell+market`에 가격 있음(거부)·`sell+market` + KR 비 KRW 통화(거부 유지)·`sell+market` + 수량 0(거부
      유지)·US fractional 기존 분기 무변화·`limit` 전 분기 무변화
- [ ] 1.1a **비분수 MARKET 정수 수량 처분** (재리뷰 P2-2) — 1.0.1·1.5.1 열거 뒤 정한다: 로컬 거부 추가 또는 브로커 400 위임.
      어느 쪽이든 시험 이름으로 고정하고 design D3 에 결과를 적는다
- [ ] 1.2 **GREEN** — [`internal/execgw/failclosed.go`](../../../internal/execgw/failclosed.go) `checkOrderShape`
      의 `orderType != "limit"` 거부에 축소 시장가 예외 추가(조기 통과 반환 금지 — design D3)
- [ ] 1.3 진입 경로가 여전히 시장가를 거부함을 **별도 테스트로 고정**(②·③ 각각) — riskcalc의 규칙은
      살아 있고 이 change가 그것을 건드리지 않았다는 증거
- [ ] 1.4 `ReasonUnsupportedOrderType` 메시지 문구 갱신 (지금 "only limit orders (and US
      fractional market orders)"라고 단언한다)
- [ ] 1.5 **셋째 관문 — `internal/trading` `placeIntentSupported`, 엔진 인스턴스 한정** (사용자 결정 2차 2026-10-10:
      sell+market·가격 없음 허용은 엔진이 만드는 `trading.Service` 에만 생성자 옵션으로. CLI·ops·MCP 기본 경로는 upstream 바이트
      동일 — 1차 결정의 「사람 CLI 동작 변경」은 철회. design D3 ③)
  - [ ] 1.5.0 **Pre-Edit 선언** — `trading.placeIntentSupported`·`Service` 생성(옵션 추가)·`engine.go` 엔진 인스턴스 생성
        (WORKFLOW §Pre-Edit 형식). upstream 상속 테스트 영향 = **기본 동작 무변화**(술어 시그니처에 인자 하나가 더해지는 것뿐 —
        upstream 시험의 기대값은 그대로여야 한다). 회귀 방지 = 1.5.5. 불변식 3 = 엔진 꺼짐 ⇒ 엔진 인스턴스 없음 ⇒ 흔적 0
        (`issues.md` I-R1 종결 인용). 엔진 인스턴스를 쓰는 사람 조작 표면이 1.5.1 에서 발견되면 그 사실을 여기 적는다
  - [ ] 1.5.1 **Function Logic Map** — `internal/trading/service.go` / `placeIntentSupported` (+ `PreviewPlace`·`Place` 의 술어 호출 자리)
        (`ast.json` + `function-logic-map.md` + `branch-test-map.md` + `risk-pattern-report.md`). HEAD 에서 생성.
        risk-pattern-report 에 비시험 호출자 둘(`PreviewPlace`·`Place`)과 `trading.NewService` 생성 자리 셋 — 엔진
        (`engine.go` `OrderPath.Trading` → `execgw` 게이트웨이·엔진 `tradingService`), CLI 앱(`app.go` → `tossctl order place`·`tossctl ops`),
        MCP 서버(`mcp.go` → ops `place_order`, MCP 카탈로그 노출) — 를 CodeGraph 로 적는다. design D3 ③ 의 「통화 검사가 비지정가 거부
        뒤」(조기 통과 금지 근거)를 이 열거로 확정
  - [ ] 1.5.2 **RED** — `placeIntentSupported`(③만 단독으로 — ② 를 거치지 않는 직접 호출) 표, **옵션 켬 / 옵션 끔** 두 열:
        옵션 켬 — `sell+market`·가격 없음(지원)·`buy+market`(비지원)·`sell+market`·가격 있음(비지원)·**`sell+market` + KR 비 KRW
        통화(비지원)**·**US `sell+market` + 통화 KRW/USD 외(비지원)**·fractional·limit 무변화. 옵션 끔 — 비분수 market 전부 비지원
        (오늘과 같음). 엔진 인스턴스의 `PreviewPlace(sell+market)` `LiveReady` 가 `Place` 의 지원 판정과 일치. **조기 통과 금지**
        (design D3 ③ — P2-2 를 ③ 에 적용): 통화 행이 GREEN 이 조기 `return true` 로 구현되면 실패해야 한다. `Place` 가 `buy+market` 에
        `ErrPlaceUnsupported` 를 계속 반환함을 고정
  - [ ] 1.5.3 **GREEN** — 생성자 옵션(모양은 design D3 ③ 제안 — 기존 빌더 관례) 추가, 술어가 그 값을 인자로 받아 비분수 비지정가
        거부에 `sell+market`·가격 없음 예외를 둔다(조기 통과 반환 금지 — 뒤의 통화 검사 생존). `engine.go` 엔진 인스턴스 생성에서만 켠다
  - [ ] 1.5.4 주석·문구 갱신 — `checkOrderShape` 의 「internal/trading is not ours to change — design D1」 주석을 「기본 동작은
        upstream 그대로, 엔진 인스턴스만 옵션」으로. `PreviewPlace` 경고 문구는 **기본 인스턴스에서 바이트 동일**; 엔진 인스턴스의
        문구 처리(그대로 / 한 줄 덧붙임)만 정한다. ops `place_order` Summary(「limit or US fractional market」)는 사실 그대로라 무변경
  - [ ] 1.5.5 **upstream 동일성 고정 시험** (3차 재리뷰 A-P1-1 — 사람·에이전트 경로 무변화) — 실주문 없음, 시험 더블:
        (a) 옵션 없는 기본 `trading.NewService` 가 비분수 market 매도를 `ErrPlaceUnsupported` 로 계속 거부하고 `PreviewPlace` 의
        `LiveReady=false`·경고 문구가 바이트 동일, (b) ops `place_order` 핸들러가 기본 인스턴스로 market 매도를 계속 거부(MCP 노출
        표면과 `tossctl ops` 가 같은 핸들러), (c) **구조 시험** — 옵션 호출 자리가 저장소 비시험 코드에서 `engine.go` 의 엔진 인스턴스
        생성 하나뿐(`app.go`·`mcp.go` 는 옵션을 켜지 않음). (c) 가 없으면 누가 CLI 생성에 옵션을 붙여도 (a)(b) 는 시험이 만든 인스턴스를
        재므로 초록이다
- [ ] 1.6 **종단 회귀 시험 (P0-1)** — 엔진 보호 청산 MARKET 이 ①`sellIntent`·②·③을 **모두** 지나 전송 계층(시험 더블)에
      닿는다. 이 시험이 없으면 P0-1 이 재발해도 단위 시험은 전부 초록이다

## 2. 보호 청산의 주문 유형 (High-risk 본체 · Phase 2)

- [ ] 2.0 **Pre-Edit 선언** — `ExitObserver.sellIntent` (WORKFLOW §Pre-Edit 형식)
- [ ] 2.1 **Function Logic Map** — `internal/app/engine/exitloop.go` / `ExitObserver.sellIntent`
      (`ast.json` + `function-logic-map.md` + `branch-test-map.md` + `risk-pattern-report.md`).
      `ast.json`의 `file`은 **저장소 상대 경로**. **HEAD 에서 재생성한다** — 기존
      `analysis/function-logic/internal-app-engine--exitobserver.sellintent/ast.json` 은 base `102d4e99`(start.line 1571)
      산출물이라 재사용 금지(재리뷰 P2-5; 본문 바이트 동일이라 분기 집합은 같을 것 — 검증 필요). risk-pattern-report 에
      `isProtective` 의 세 소비자(`record` 재판정 보류 예외·`submit` a091 floor·a087 유형 결정 — design D1)를 적는다
- [ ] 2.2 **RED** — `sellIntent` 분기: `BASELINE_BREACH`→market·`STOP_LOSS_LADDER`→market·
      `LADDER_TAKE_PROFIT`→limit(가격은 종전대로 관측가)·시장가에는 가격이 실리지 않음.
      그리고 `isProtective` **술어 자체가 바뀌지 않음**을 고정(재리뷰 P3-2 — 세 행동 집합 그대로)
- [ ] 2.3 **GREEN** — `sellIntent`가 `proposal`을 인자로 받아 `isProtective`로 유형을 정한다.
      현재 시그니처는 proposal을 안 받으므로 `submit`의 호출부도 함께 바꾼다
- [ ] 2.4 보호 제안일 때 관측가·기준선을 **읽지 않는지** 확인 — 시장가 경로에서 가격이
      필요 없어졌으므로 "가격 없음" 거부(`has no price to submit a liquidation at`)가
      보호 청산을 막지 않아야 한다. **이것이 §0.3의 핵심 개선분이다**
- [ ] 2.5 익절 경로는 가격 없음에 대해 종전 거부를 유지

## 3. 원장·관측·화면 (Phase 2)

- [ ] 3.1 **RED** — 시장가 청산의 `intents.price`가 비고 `order_type`이 market인지
- [ ] 3.2 **RED** — 관측가·기준선이 제출가로 기록되지 **않는지** (주문된 적 없는 가격)
- [ ] 3.3 **GREEN** — 원장 기록 경로
- [ ] 3.4 **RED/GREEN** — 운영자 화면·알림이 가격 없는 청산을 "시장가"로 표시(결측 아님).
      `operatorview`·콘솔 템플릿·`obs` 필드
- [ ] 3.5 체결가 확정이 체결 조회 경로로 이뤄지는지 확인 — 이미 `filldetect`가 하는 일이면
      변경 없음을 테스트로 고정, 아니면 갭을 `issues.md`에 기록
- [ ] 3.6 **부분체결 잔량** (재리뷰 P2-3 · 1차 리뷰 A14 의 행방, design D8) — MARKET 보호 청산의 잔량이 남는지·`filldetect`
      가 그 잔량을 어떻게 닫는지 확인. 갭이면 `issues.md` 에 기록
- [ ] 3.7 **`intents.price` NULL 소비부 전수 열거** (재리뷰 P3-3 — 표본 셋 `fills.go` coalesce·`exit_held_proposal.go`
      `workingOrderPrice`·`indoubt.go` 가격 와일드카드/notional 0 은 통과, 전수 아님). **방법 (3차 재리뷰 A-P3-3)**: CodeGraph 는
      심볼 단위라 SQL 열의 읽기 자리를 열거하지 못한다 — ① 비시험 Go 소스의 SQL 문자열에서 `intents` 와 `price` 를 함께 읽는 자리를
      `rg` 로 모으고, ② 그 질의를 감싼 저널 접근자 함수의 호출자를 CodeGraph 로 따라가 ①∪② 를 `analysis/` 에 남긴다. 각 자리의 NULL
      처분을 시험 또는 not-applicable 사유로 적는다
- [ ] 3.8 **청산 슬리피지 산출** (재리뷰 P1-3, design D6 · 3차 재리뷰 A-P2-4 정정) — **기존 칸 유도, schema·production 코드 추가
      없음**(proposal Impact 「Schema: 없음」과 일치). 결정 시 관측가 `exit_events.observed_price`(같은 행 `proposed_intent_id`) 와 그
      intent 의 브로커 `order_id` 로 찾은 체결 평균가 `fill_snapshots.average_price` 로 bps 를 계산하는 **읽기 전용 저널 질의**를
      `analysis/` 에 둔다(착지 뒤 보호 청산마다 사람이 돌림). 「원장에 남긴다」는 철회 — 새 열을 쓰지 않는다. 착수 시 intent → `order_id`
      결속이 기존 칸으로 성립하는지 먼저 확인하고, 성립하지 않으면 **멈추고** proposal Impact 를 schema 변경으로 고쳐 재리뷰를 받는다.
      비례 원칙: 보호 청산 판정을 바꾸지 않는 관측 산출이라 High-risk 무거운 규율(FLM·변이 원장) 대상 아님(`slippagePct` 는 진입 전용 —
      무변경)

## 4. 회귀 방지

- [ ] 4.1 `flatten` 무변화 — 이미 거래소 하한가 1순위이고 이 change의 범위 밖임을 테스트로 고정.
      `flatten.sell`/`Liquidate` 는 `checkOrderShape` 상류다(`analysis/gate-0.4-codegraph.md`) — ② 변경이 flatten 의 LIMIT 을 바꾸지 않음
- [ ] 4.2 `verifylive` 무변화 — 검증 도구의 "체결되면 안 되는 지정가" 성질 유지
- [ ] 4.3 조건주문 경로 무변화 (2c 범위)
- [ ] 4.4 upstream 상속 테스트 650 green 유지
- [ ] 4.5 **전송 표류 가드 KR market sell 골든 행** (재리뷰 P3-3) — `decision_test.go` 표류 가드 골든은 US market sell 만 있다.
      KR market sell 행을 더해 `buildOrderCreate` 와 사본 `PlaceWireBody`(`execgw/wirebody.go`)가 둘 다 가격·TIF 를 싣지 않음을 고정
- [ ] 4.6 **재생 경로** (재리뷰 P3-3) — `PlaceWireBody` 를 쓰는 재생 경로가 MARKET 보호 청산 intent 를 가격 없이 재구성하는지 시험

## 5. 실측 (사용자 승인 항목 — §0.7, 자동 실행 금지)

> **2026-10-05 큐 편성(사용자 승인 — a100 review.md 「사용자 결정 — 2026-10-05」 3항):**
> 5.1·5.2 를 a100 M-A 의 KR 장중 사람 실측 세션과 같은 큐에 올린다 — 같은 세션에서 측정해도
> 되고 별도 세션이어도 된다. 주문별 사람 즉시 승인·자동 실행 금지는 그대로다. a100 은 이
> 실측을 기다리지 않는다(a100 D8 계약 1 (다) 채택 — Phase 2 는 a100 선행 조건이 아니다).
>
> **2026-10-10 보류(2차 재리뷰 P1-1) → 같은 날 실행 수단 확정(사용자 결정 2):** 5.1 은 출하 도구로 실행 불가였다 —
> `tossctl order place` 도 엔진과 같은 `placeIntentSupported`(`internal/trading/service.go`, 비분수 `OrderType != "limit"`
> → `ErrPlaceUnsupported`)에서 **브로커에 닿기 전에** 로컬 거부되므로, 그 결과를 기록하면 우리 관문을 브로커 응답으로
> 적는 거짓 음성이 된다. 조건주문(MARKET 허용)은 다른 endpoint 라 `/api/v1/orders` 를 재지 못한다.
>
> **실행 수단 = 별도 실측 도구 `tools/a087-market-sell-probe/`** (사용자 결정 2026-10-10 — `internal/official` 직접 호출,
> 토큰 캐시 공유, 새 파일만, production 변경 **전에** 측정; 병행 팀메이트 구현 — 도구의 명령·출력 형식은 그 도구의 문서가
> 정본이며 이 tasks 는 선취하지 않는다). 주문 생성은 mutating — **실행은 사람 터미널 + 주문별 승인**, 에이전트는 준비까지.
> 대상 수량은 010170 매수 산식에 이미 포함(5.1 매도 1주·5.2 경계 1주). M-A 동승 큐(R2)는 이 도구가 준비되고 0.7 Eng
> 재리뷰가 끝난 뒤 재개한다. a100 `ma-runbook.md` R2 에 실행 명령이 없던 공백(재리뷰 P1-1)은 이 도구 경로로 메운다
> (runbook 쪽 반영은 이 디렉터리 범위 밖 — `issues.md` I-R5).

> **실행 전제 (2026-10-10, 3차 재리뷰 B-P1-1·A-P2-5·B-P2-1·B-P3-3·B-P3-4)** — 사람이 5.1·5.2 를 치기 직전에 확인한다:
>
> 1. **도구 소스가 착지 커밋 그대로일 것**: `git status --porcelain -- tools/a087-market-sell-probe` 가 비어 있고,
>    `git log -1 --format=%H -- tools/a087-market-sell-probe` 가 리뷰를 통과한 착지 커밋이다. 변이 하네스(`mutants.sh`)는
>    `45167aa7` 부터 저장소 밖 사본에서만 변이하므로 공유 트리를 바꾸지 않지만, 병행 세션의 미커밋 편집은 이 검사로만 잡힌다.
> 2. **매도가능수량 사전 스냅숏 (👁 read-only)**: 5.1 종목의 매도가능수량과 **같은 종목의 상주 SELL 조건주문**(a100 M-A 당일 등록분,
>    R0-3 기본 retain 잔여분)을 조회해 1주 여유를 확인한다. 도구는 이 표면을 일부러 읽지 않는다(`static_test.go`
>    `TestNoProductionOrderSurfaceIsNamed`). 실측 당일 a100 runbook §2 종목 확정 때 함께 본다 — 겹치면 5.1 종목을 바꾸거나
>    결과를 아래 표 「판정 불가(수량)」로 읽는다. 계좌 선택은 `accounts[0]`(도구는 계좌를 영수증에 적지 않음 — B-P3-6)이라
>    이 스냅숏이 「그 계좌에 그 종목이 있다」의 유일한 확인이다.
> 3. **결과 `unknown`(또는 `sending`/`no-answer` 영수증이 남음) 뒤에는 미리보기를 다시 돌리기 전에** `tossctl orders list` /
>    `tossctl orders completed` 로 그 종목 주문을 먼저 확인한다 — 새 미리보기는 새 멱등 키라 둘째 실주문이 된다(B-P3-4, 도구 잠금
>    대신 절차 — `issues.md` I-R7).
> 4. 확인 토큰의 「5분」은 검증 시점 기준이다 — 헤더 획득(토큰 교환·계좌 조회)이 끼어 실제 전송은 창 끝을 수십 초 넘을 수 있다.
>    브로커 멱등 창 10분 안이라 이중 주문 위험은 없다(B-P3-3).
>
> **영수증 디렉터리는 커밋 의도다 (B-P2-1 처분 — Manager 지시 2026-10-10)**: 기본 `--out` =
> `analysis/market-sell-receipts/` 의 영수증 JSON 을 측정 영수증으로 커밋한다. 담기는 것 — orderId·원 응답 본문·종목·수량·시각·
> 보낸 헤더 **이름** — 중 orderId 는 계좌 개인정보가 아니다(허용). 토큰·계좌 번호·API 키는 도구가 쓰지 않는다(헤더 값 미기록,
> sentinel 5종 시험). 원 응답 본문의 `error.data` 에 수량 값이 실릴 수 있으나 계좌 식별 정보가 아니다.

- [ ] 5.1 **KR MARKET 매도 1회.** 최소 수량(1주)·장중. 스키마는 지원하나 실접수 미측정.
      수단 `tools/a087-market-sell-probe/`. 성공·실패 모두 응답 코드·본문 code·주문 상태를 기록. 처분은 아래 표
- [ ] 5.2 **세션 경계.** 정규장 밖 KR MARKET 매도(1주)의 응답 코드 관측. **예상 응답 집합**(재리뷰 P1-4):
      `422 order-hours-closed` · **`422 order-type-not-allowed`**(openapi `POST /api/v1/orders` 422 예시 「현재 사용할 수 없는
      호가 유형」 — 저장소 Go 코드 참조 0, `refusal_code.go` 목록에도 없음) · 그 밖의 코드(있는 그대로 기록).
      시간외단일가에는 시장가가 없다
- [ ] 5.3 `openspec/changes/verify-execution-capability/measurements.md`에 M계열로 기록
      (`docs/trading/measurements.md`는 **존재하지 않는다** — 초안 리뷰 M7)

### 실측 결과 → 처분 (재리뷰 P1-1·P1-4, 사용자 결정 2026-10-10)

| 측정 | 결과 | 처분 |
| --- | --- | --- |
| 5.1 | **수용** — HTTP **200**(openapi `POST /api/v1/orders` 성공 응답은 200 하나 — 201 없음) **그리고** 사람이 `tossctl orders list` / `tossctl orders completed` 로 영수증 orderId 의 상태를 읽어 `PENDING`·`PARTIAL_FILLED`·`FILLED` 중 하나 | Phase 2 전제 성립. 나머지 착수 조건(위 §1 머리 ②~④) 충족 시 **§1~§3 착수 가능**. 체결가·관측가는 기록만(bps 판정 근거로 쓰지 않음 — n=1, design D6). 도구의 `accepted` 는 200 + orderId 일 뿐이고 이 행이 아니다(3차 재리뷰 A-P1-2) |
| 5.1 | **200 뒤 `REJECTED`** — 접수 응답은 200 이었으나 조회 상태가 `REJECTED`(openapi 주문 상태 어휘 「브로커가 주문을 거부한 상태」, MARKET·REJECTED 예시 존재) | 거부 사유가 MARKET 유형 자체이면 아래 「거부」 행, 종목 상태·수량 등 유형 무관 사유이면 「판정 불가」 행으로 읽는다. 사유를 못 읽으면 「판정 불가」 |
| 5.1 | **거부** — `422 order-type-not-allowed` 등 MARKET 매도 자체를 거부하는 4xx | **Phase 2 전제 반증.** production 무접촉(§1~§3 착수 안 함). change 처분(불구현 아카이브 / design D6 하한가 지정가 원안 재검토 등)은 **사람 결정** |
| 5.1 | **판정 불가 — 종목 상태 사유** — VI 발동 중 단일가·투자경고·단일가 매매 종목·거래정지 등 **그 종목의 그 시각** 상태를 말하는 거부(4xx 또는 200 뒤 `REJECTED`) | MARKET 유형 일반의 반증이 아니다. 구분 기준: 오류 code·message 가 종목·시장 상태를 가리키면 이 행, 주문 유형을 가리키면(`order-type-not-allowed` 류) 「거부」 행. 애매하면 이 행. 다른 종목·시각으로 재측정할지는 사람 결정 |
| 5.1 | **판정 불가 — 매도가능수량 사유** — 같은 종목의 상주 SELL 조건주문(a100 M-A 당일 등록분·R0-3 retain 잔여분)이나 다른 미체결이 매도가능수량을 묶어 수량 사유로 거부 | MARKET 반증 아님. 위 「실행 전제」 2 의 사전 스냅숏으로 피해야 하는 행이다 — 실측 당일 a100 runbook §2 종목 확정 때 겹침을 확인하고(M-A retain 결정과 교차: 잔여 조건주문은 취소하지 않으므로 5.1 종목을 바꾼다), 그래도 나오면 기록 후 사람 결정 |
| 5.1 | **판정 불가 — 그 밖** — 인증(401·403)·한도(429)·장 시간 등 MARKET 유형과 무관한 사유의 4xx, 5xx·타임아웃·409·in-doubt(도구 `unknown`) | 수용도 반증도 아니다. 원인·주문 상태(브로커 조회로 확정)를 기록하고 재측정 여부는 사람 결정. §1~§3 착수 안 함. 도구의 `refused` 는 4xx 라는 뜻뿐이라 이 행일 수 있다(B-P3-2) |
| 5.2 | 세션 밖 — 거부 응답(`422 order-hours-closed`·`order-type-not-allowed`·그 밖) | **응답 코드 기록만.** 자동 처분 없음 — design D7 교집합 처분의 입력(0.5 와 함께 사람 결정) |
| 5.2 | **세션 밖인데 200 으로 접수됨** (장전 동시호가 대기·익일 이월 등) | 실주문이 다음 세션에 체결될 수 있다. 즉시 `tossctl orders list` 로 상태를 읽고 **취소 여부는 사람이 그 자리에서 결정**(취소는 mutating — 사람 실행). 체결되면 010170 매수 산식의 5.2 몫(경계 1주)을 소진한 것으로 기록하고 추가 측정 수량을 다시 잡지 않는다. D7 표의 해당 구간은 「MARKET 접수됨(체결 시점 미확정)」으로 기록 |
| US | 측정 없음 | 자동 처분 없음. §1~§3 착수 전 사람이 「US 를 미측정으로 착지」(proposal 「US 한계」)를 확인한다 |

상태 어휘는 openapi 주문 상태 enum 10개(`PENDING`·`PENDING_CANCEL`·`PENDING_REPLACE`·`PARTIAL_FILLED`·`FILLED`·`CANCELED`·`REJECTED`·
`CANCEL_REJECTED`·`REPLACE_REJECTED`·`REPLACED`)다. 위 표가 이름 붙이지 않은 상태(예: 사람이 취소하지 않았는데 `CANCELED` — 잔량 소멸 등)는
**판정 불가**로 읽고 그 상태와 체결 수량을 그대로 기록한다(수용으로 올려 적지 않는다).

## 6. 게이트

- [ ] 6.1 `go test ./... -count=1 -race` 회귀 0
- [ ] 6.2 `make sdd-sync` 재실행 (마지막 파일 편집 후)
- [ ] 6.3 `make sdd-check`
- [ ] 6.4 **격리 worktree에서** `make gate CHANGE=a087-a-protective-exit-is-a-market-order`
- [ ] 6.5 독립 검증 (구현과 분리된 컨텍스트)
- [ ] 6.6 PM 동기화 → `openspec archive`

## 후속 change (이 change에서 하지 않는다)

> **정정 (2026-10-10, 재리뷰 P3-1)**: 3판까지 이 표는 a088·a089 를 소유자로 적었다. 「a088」은 `openspec/changes`·`archive`
> 어디에도 없고, a089 는 다른 내용(「나가지 못한 손절은 세어진다」)으로 2026-09-28 불구현 아카이브됐다. 아래는 **무소유** 장부다.

| 소유자 | 내용 | 근거 |
| --- | --- | --- |
| **무소유** (옛 표기 「a088」) | 호가 그리드 정본화 — StockOS 형태 이식 | 익절 지정가·`flatten` fallback이 필요로 한다. 초안 리뷰의 A1·A4·A5·A10·A11·A12 |
| **무소유** (옛 표기 「a089」) | 긴급 청산 재가격 에스컬레이션 + 게이트 우회 | 7분 42초 공백은 주문 가능한 제안이 없던 구간(proposal Why 정정) |
| **무소유** | 1차 리뷰 A1~A14 중 위 두 행 밖의 것, I2(400 분류) | a094 R1 이 가져간 것은 `opposite-pending-order-exists` 하나(`refusal_code.go`) |
| **무소유** | `order-type-not-allowed` 422 분류 | `issues.md` I-R2 |

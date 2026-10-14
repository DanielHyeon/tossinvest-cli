# a140 — 돌파 신호는 주문 전에 그림자에서 먼저 잰다 (breakout-signals-are-measured-in-shadow)

## Why

2026-10-14 StockOS 갭 분석의 P2 둘째 항목은 「진입 신호 생산자」다. 기억의 문장은 「4/8 레인은 있으나 진입에 살아 있는 신호
생산자가 없다」였고, 이 change 는 그것을 HEAD `050d092f` · CodeGraph 1.6.0 · `tools/logic-map` AST 로 다시 쟀다
(`analysis/ast-evidence/claims.md` · `analysis/code-context/`). 결과는 브리프보다 나쁘고 더 구체적이다.

1. **레인은 8/8 이 서 있다.** `strategyworker.ProductionWorkers`(worker.go:133~152)가 네 가족 × KR·US 여덟 인스턴스를 만들고
   생산 런타임이 그것을 쓴다(lane.go:110). 네 가족의 순수 평가기도 다 있다.
2. **살아 있는 신호 생산자는 0/8 이다.** 생산 제안은 엔진이 만들지 않고 **사람이 서명한 파일**
   (`strategy-proposal-input-<KR|US>.json`, Ed25519)에서 읽는다. 그 파일의 digest 는 프로세스 환경
   `TOSSOS_STRATEGY_PROPOSAL_<KR|US>_MANIFEST_SHA256` 에 핀되어 있어(collectMarket B8@370 · LoadProductionAuthorityBatch B3@321)
   새 신호는 재시작 없이는 통과하지 못하고, 범위마다 신선 창은 5분 이하다(continuationlane `productionEvidence`). 즉 지금의 계약은
   주기마다 생기는 신호를 실을 수 없다.
3. **증거를 쓰는 생산 코드가 0 이다.** 생산은 `evidence.db` 를 읽기 전용으로만 연다(B5@329). 쓰기 열기
   `strategyevidence.Open(` 의 생산 호출은 0, 닫힌 1분봉 생산자 `officialbars.PollClosedBars` 와 L1 호가 생산자 `PollQuoteL1` 은
   만들어져 시험 30 개가 지키지만 생산 호출이 0 이다(`raw/callers-*.txt`).
4. **가족마다 막힌 자리가 다르다.**
   - KR continuation · KR reversal: `kr_net_flow` 증거가 필요한데 **종목별 투자자 매매는 공식 표면에 없다**(a128 tasks 1.7 영수증).
   - US continuation · US reversal: `us_participation` 을 만드는 코드가 없고 「기준 거래량」의 정의가 어디에도 없다.
   - weekly value: 공시 증거 어댑터(OpenDART · SEC EDGAR) 정책과 주간 예약이 먼저다.
   - breakout-retest: **공식 1분봉 · 호가만으로 만들 수 있는 유일한 가족**이고 두 생산자가 이미 있다. 그런데도
     `buildLaneInput` B1@439 가 레인 ID 만 보고 무조건 `ErrBreakoutEvidenceUnavailable` 을 돌려준다 — ATR · RVOL · 윗꼬리 비율 ·
     거래량 확장 네 파생 지표를 만드는 코드가 없기 때문이다(production.go:57~61 이 스스로 적음).

그래서 P2 신호 생산은 한 change 로 할 수 없다. 첫 Story 는 「breakout 이 우리 시장에서 신호를 **실제로 내는가**, 낸다면 얼마나
자주 · 어떤 결과로」를 주문 경로 0 으로 재는 것이다. 이 숫자 없이 서명 · 활성화 · Jev 개입을 설계하면, 신호가 하루 0 건인 가족에
신뢰 앵커를 붙이는 일이 된다.

## What Changes

1. **새 순수 패키지 `internal/breakoutsignal/`** (새 파일만, 생산 호출자 0 — 다음 Story 가 엔진에 배선할 때 그대로 쓴다):
   저장된 공식 닫힌 1분봉 묶음(`strategyevidence.BarSeries`)에서 breakout v1 의 네 파생 지표를 결정적 정수 산술로 만든다.
   정의와 판본은 design D3~D6(사람 결정)이 정한다. 못 만드는 값은 0 이 아니라 typed 거절이다.
2. **새 측정 도구 `tools/a140-breakout-signal-shadow/`**: KR 정규장 동안
   - 세션 시작에 공식 랭킹으로 고정 종목 집합을 고른다(D8).
   - 종목마다 1분에 한 번 `officialbars.PollClosedBars` 로 **도구 전용** evidence DB 에 닫힌 1분봉을 적재한다.
   - `breakoutsignal` → `breakoutlane.Evaluate`(동결 v1 상수 그대로)로 상태기계를 굴리고, `ARMED` 에 닿은 setup 에만
     `PollQuoteL1`(GET 2)로 호가를 봉인해 최종 제안 여부를 판정한다.
   - 결과를 append-only JSONL 원장에 적는다 — 전이 · 거절 코드 · 「냈을 제안」(진입 · 손절 · 목표 · 수량, 가상 사이징 D7) ·
     수집 결손.
   - 라벨러: 같은 DB 의 뒤 봉으로 「냈을 제안」이 손절과 목표 중 무엇에 먼저 닿았는지, 장 마감까지 못 닿으면 마감가 R 배수를 적는다
     (추가 GET 0, D10).
   - 보고서: 세션별 setup 퍼널(DISCOVERED → … → PROPOSED) · 거절 분포 · 제안 수 · 결과 분포 · 표본 한계(n<30 판정 보류).
3. **spec 델타** — `breakout-retest-strategy-lane` 에 ADDED 셋: 파생 지표의 결정적 생산 · 그림자 측정의 무권위 · 원장의 모집단 복원.

## What Does NOT Change

- 엔진 · 주문 · Guardian · 원장(`journal`) · 대사 · 손절 · 콘솔 · 설정: **무접촉**. `internal/` 의 기존 파일 편집 0,
  `cmd/` 편집 0. a138 · a139 가 편집 중인 `risk.Policy` · `guardian_wiring.go` · `interlock.go` · config · 콘솔 `settings_limits` 와
  겹치는 파일 0.
- 생산 `evidence.db` 에 쓰지 않는다 — 도구는 그 경로를 거절한다.
- 서명 매니페스트를 읽지도 만들지도 않는다. 활성화 · 토글 · 레인 상태를 바꾸지 않는다.
- **Jev 를 부르지 않는다.** Jev 개입은 별도 Story(S2 · S3, design 「Story 분할」)다. 이 change 는 그 Story 들이 쓸 모집단
  (「엔진이 냈을 진입」)을 처음으로 만든다.
- breakout v1 동결 상수(opening range 15 · RVOL 1.5× · 윗꼬리 35% · timeout KR 8 봉 …)를 바꾸지 않는다.

## Human Gates

- design 「사람 결정」 표 D1~D12(첫 가족 · 지표 정의 · 가상 사이징 · 종목 집합 · 기간 · 라벨 · 호출 예산)를 사람이 정한 뒤에만 구현한다.
- 측정 실행은 read-only GET 뿐이라 mutating 이 아니다(에이전트 실행 가능). 호출 예산 기본값은 사람이 승인한다(D11).
- 엔진과 같은 Open API 키를 쓴다 — a128 프로브와 같은 날 돌릴지 사람이 정한다(D11).

## Impact

- Affected specs: `breakout-retest-strategy-lane`(ADDED 3).
- Affected code: `internal/breakoutsignal/`(신규) · `tools/a140-breakout-signal-shadow/`(신규)만.
- 위험도: **Low**(새 파일만 · High-risk 경로 무접촉 · 주문 side effect 0) — docs/WORKFLOW.md 「비례 원칙」으로 FLM · 변이 원장 ·
  다중 적대 리뷰는 `not-applicable`, 시험 + lint + 경량 gstack 리뷰. 단 새 패키지는 다음 Story 에서 High-risk 입력이 되므로 그 시험
  (경계값 · 결정성 · 거절)은 이 change 에서 완성한다.

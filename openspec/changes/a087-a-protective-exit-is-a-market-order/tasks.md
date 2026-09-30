# a087 tasks

> **High-risk.** 손절 주문의 **유형**을 바꾼다. 구현 착수 전 Pre-Edit 선언과
> proposal-freeze 재리뷰(적대적 Eng 관점 포함)가 필요하다.
> 초안(호가 그리드)의 리뷰 결과는 `review.md`에 보존한다 — 이 tasks는 그 리뷰가
> 도달한 결론을 실행하는 것이지 우회하는 것이 아니다.

## 0. 게이트 선행

- [x] 0.1 `capture_change_base.py --change a087-a-protective-exit-is-a-market-order`로 base commit 재고정 (디렉터리명이 바뀌었다) — `5491451b`(ec29dc72 → 102d4e99, 2026-09-30, 승인 기록 review.md)
- [ ] 0.2 `openspec validate a087-a-protective-exit-is-a-market-order --strict --no-interactive`
- [ ] 0.3 **proposal-freeze 재리뷰** 실행 후 `review.md`에 2차 절 추가 (적대적 Eng 필수)
- [ ] 0.4 `make sdd-sync` 후 `sellIntent`·`checkOrderShape`·`isProtective`·`buildOrderCreate`의
      definition/callers/impact 확인

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

> **아래 §1~§3(Phase 2)의 착수 조건은 §5 실측 기록이다(§0.7 사람 승인).** Phase 1 은
> 그 실측을 기다리지 않는다.

## 1. 실행 게이트를 축소 시장가에 연다 (Phase 2)

- [ ] 1.1 **RED** — `checkOrderShape` 표 테스트: `sell+market`(통과)·`buy+market`(거부)·
      `sell+market`에 가격 있음(거부)·US fractional 기존 분기 무변화·`limit` 전 분기 무변화
- [ ] 1.2 **GREEN** — [`internal/execgw/failclosed.go:84`](../../../internal/execgw/failclosed.go)
      의 `orderType != "limit"` 거부에 축소 시장가 분기 추가
- [ ] 1.3 진입 경로가 여전히 시장가를 거부함을 **별도 테스트로 고정** — riskcalc의 규칙은
      살아 있고 이 change가 그것을 건드리지 않았다는 증거
- [ ] 1.4 `ReasonUnsupportedOrderType` 메시지 문구 갱신 (지금 "only limit orders (and US
      fractional market orders)"라고 단언한다)

## 2. 보호 청산의 주문 유형 (High-risk 본체 · Phase 2)

- [ ] 2.0 **Pre-Edit 선언** — `ExitObserver.sellIntent` (WORKFLOW §Pre-Edit 형식)
- [ ] 2.1 **Function Logic Map** — `internal/app/engine/exitloop.go` / `ExitObserver.sellIntent`
      (`ast.json` + `function-logic-map.md` + `branch-test-map.md` + `risk-pattern-report.md`).
      `ast.json`의 `file`은 **저장소 상대 경로**
- [ ] 2.2 **RED** — `sellIntent` 분기: `BASELINE_BREACH`→market·`STOP_LOSS_LADDER`→market·
      `LADDER_TAKE_PROFIT`→limit(가격은 종전대로 관측가)·시장가에는 가격이 실리지 않음
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

## 4. 회귀 방지

- [ ] 4.1 `flatten` 무변화 — 이미 거래소 하한가 1순위이고 이 change의 범위 밖임을 테스트로 고정
- [ ] 4.2 `verifylive` 무변화 — 검증 도구의 "체결되면 안 되는 지정가" 성질 유지
- [ ] 4.3 조건주문 경로 무변화 (2c 범위)
- [ ] 4.4 upstream 상속 테스트 650 green 유지

## 5. 실측 (사용자 승인 항목 — §0.7, 자동 실행 금지)

- [ ] 5.1 **KR MARKET 매도 1회.** 최소 수량·장중. 스키마는 지원하나 실접수 미측정.
      성공·실패 모두 기록. 절차와 대상 종목은 사용자와 합의한 뒤 실행
- [ ] 5.2 **세션 경계.** 정규장 밖 KR MARKET 매도의 응답 코드 관측 (`422 order-hours-closed`
      인지 다른 코드인지). 시간외단일가에는 시장가가 없다
- [ ] 5.3 `openspec/changes/verify-execution-capability/measurements.md`에 M계열로 기록
      (`docs/trading/measurements.md`는 **존재하지 않는다** — 초안 리뷰 M7)

## 6. 게이트

- [ ] 6.1 `go test ./... -count=1 -race` 회귀 0
- [ ] 6.2 `make sdd-sync` 재실행 (마지막 파일 편집 후)
- [ ] 6.3 `make sdd-check`
- [ ] 6.4 **격리 worktree에서** `make gate CHANGE=a087-a-protective-exit-is-a-market-order`
- [ ] 6.5 독립 검증 (구현과 분리된 컨텍스트)
- [ ] 6.6 PM 동기화 → `openspec archive`

## 후속 change (이 change에서 하지 않는다)

| ID | 내용 | 근거 |
| --- | --- | --- |
| a088 | 호가 그리드 정본화 — StockOS 형태 이식 | 익절 지정가·`flatten` fallback이 필요로 한다. 초안 리뷰의 A1·A4·A5·A10·A11·A12 적용 |
| a089 | 긴급 청산 재가격 에스컬레이션 + 게이트 우회 | 7분 43초 공백, 지연 이벤트 0건. 이 change와 독립 |

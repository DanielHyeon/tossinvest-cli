# Tasks: verify-plans-the-object-it-mutates

> 이 change는 게이트를 느슨하게 하지 않는다. 계획 밖 요청이 실행을 멈추는 레일
> (`ErrOutsidePlan`), 승인 창, 노출 상한, 수량 상한은 전부 무변경이다. 고치는 것은
> 게이트에 주어지는 **목록의 정확성**이다.

- [x] 1.1 [T] `Step.ActsOnConditional` — 등록된 조건주문 위에서 동작하는 단계를 카탈로그가
  데이터로 선언한다. 가산·`omitempty`. `conditional-modify`·`conditional-cancel`에 선언하고,
  mutating 단계 중 `liveConditional`을 읽는 것과 이 선언이 일치함을 정적 AST 테스트로
  고정한다. RED/GREEN.
- [x] 1.2 [T] `mutationSymbol`이 `ActsOnConditional` 단계에 대해 살아 있는 조건주문의 종목을
  돌려준다. 조건주문이 아직 없으면 `holdingSymbol`(등록 단계가 쓸 종목). `NeedsHolding`과
  그 밖의 경로는 무변경. RED: probe 종목이 실린다 → GREEN: 조건주문의 종목이 실린다.
- [x] 1.3 [T] 계획 줄이 실행과 일치한다 — probe `005930`, 조건주문 `333430`인 오늘의 기록
  모양으로 계획을 만들면 `conditional-modify`·`conditional-cancel` 줄의 종목이 `333430`이다.
  RED/GREEN.
- [x] 1.4 [T] 신선한 전체 실행에서도 일치한다 — 같은 실행에서 `conditional-register`부터 도는
  계획의 정정·취소 줄이 보유 종목을 싣는다. RED/GREEN.
- [x] 1.5 [T] **경로 전체** — probe와 다른 종목에 조건주문이 살아 있는 기록으로 runner를
  돌려 `conditional-modify`·`conditional-cancel`이 인가를 통과해 fake broker에 도달하는지
  확인한다. 이 테스트가 이 change의 주장 자체다(design.md D5). RED: 실계좌와 **같은**
  `ErrOutsidePlan` → GREEN: 두 단계 통과, 계좌에 잔여물 0.
- [x] 1.6 [T] 대상을 이름할 수 없는 mutating 단계는 계획에서 제외되고 사유가 표시된다.
  RED: US·보유 0 계좌의 계획에 종목 없는 라이브 주문 줄 4개 → GREEN: 0개.
- [x] 1.7 [T] `Authorises`가 빈 종목을 와일드카드로 쓰지 않는다 — 종목 없는 계획 줄은 종목 있는
  요청을 인가하지 않고, 둘 다 빈 정리 줄은 계속 인가된다. RED/GREEN.
- [x] 1.8 [T] 게이트 무변경 회귀 — 종목이 다른 요청과 상한을 넘는 수량은 여전히 인가되지 않고,
  계획이 만들어진 그 요청은 인가된다. 이 change가 인가를 넓히지 않았다는 증거.
- [x] 2.1 Function Logic Map + Branch Test Map + risk-pattern-report,
  `python3 tools/logic-map/check_analysis.py --change verify-plans-the-object-it-mutates`
- [x] 2.2 Pre-Edit 선언을 review.md에 기록 (High-risk: 라이브 정정·취소의 승인 인가 표면)
- [x] 2.3 적대적 Eng 관점을 포함한 gstack proposal-freeze 리뷰 → review.md
- [x] 2.4 `make sdd-sync && make sdd-check`, `go test ./... -count=1`, `make vet`,
  `make validate`, PM registry allowlist + `tools/pm/test_generate_master_tracker.py` fixture
- [x] 2.5 `make gate CHANGE=verify-plans-the-object-it-mutates`

## 게이트 이후 — 사람이 실행한다

설치 후 콘솔에서 사용자가 `[재측정]`을 눌러 `conditional-modify`·`conditional-cancel`이
인가를 통과하는지 확인한다. **실계좌 실행이며 에이전트가 자동 실행하지 않는다(§0.1·§0.7).**
확인할 것은 두 가지다.

1. 실패한다면 그 자리가 **인가가 아니라 브로커**인가 — 그것이 이 change가 만드는 차이다.
2. `conditional-cancel`이 통과해 `p7hQz7HAXc…`가 계좌에서 사라지는가.

KR 장 마감 시간에 정정·취소가 접수되는지는 측정된 적이 없다. `order-hours-closed` 계열로
끝나면 그것도 실측이며, `verify-execution-capability` task 2.5의 정정 원자성은 장중 창에서
다시 잡는다. 결과는 `verify-execution-capability`의 measurements 기록에 남긴다.

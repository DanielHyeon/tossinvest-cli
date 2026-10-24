> 착수 2026-10-23(Terra). 1.3 1차·2차 FAIL → 사용자 3차 결정(F3=1)으로 범위 = R3·R4·R5. design·delta 3판 → 1.3 3차 리뷰.

## 0. 착수 준비

- [x] 0.1 proposal 의 열린 질문 1~4 에 사람 답을 받고 proposal·issues.md 에 기록한다. — 2026-10-23 「1~4번 권장대로」: 조건주문 제외 · (b) 8s 창 안에서만 · 401 재전송 「확실히 미접수」 인정(시험 고정 필요, 실측 근거 없음) · 409 재시도 a165 범위([issues.md](issues.md)).
- [x] 0.2 단일 writer 배정, `python3 tools/sdd/capture_change_base.py --change a165-safe-order-placement-retry` 로 base 고정. — Terra, base `34e12967e`([base-commit.txt](base-commit.txt)).
- [x] 0.3 `make sdd-sync` 후 CodeGraph hard evidence(`Gateway.Place`·`Cancel`·`Amend`·`submit`·`classifyMutation`·`Client.send` 의
  callers/callees/impact)와 `analysis/code-context/` 증거 조정 3파일. — codegraph 1.6.0, `raw/` 원문, CGC 는 다른 체크아웃(advisory 불가).
- [x] 0.4 FLM/BTM: 위 함수와 design 이 고르는 접수 조회 함수, 그리고 `withStrategyEntryGateAuthority`(클로저 호출 횟수) — High-risk, 면제 없음. — 9 번들(`analysis/function-logic/`, 생성기 `analysis/measure/gen_flm.py`), `check_analysis.py --change` rc=0.
- [x] 0.4.1 (1.3 S8/A10) 추가 번들: 조회 경로(`Resolver.scanBoth`·`ScanOrders`·`newMatcher`·`parseOrderFacts`), `unwrapAndDecode`, `tokenManager.refresh`, `refusalCodes`, `trading.Service.Cancel`, D7 의 `ExitObserver.ObserveOnce`·`clearTheSymbol`·`Saga` 루프·`resolveStep`·`Resolver.resolveCancel`. — 2차 결정으로 범위가 줄어 남는 범위가 기대는 14 번들을 만들었다(3차 결정 뒤 재조정 — issues.md 「3차」, 합계 23)(classifyStatus·doRequest·unwrapAndDecode·tokenManager.refresh·refusalCodes·parseOrderFacts·newMatcher·matcher.matches·roundTripFor·DispatchVerified·Service.Cancel·Saga.settleStep·classifyCancel·clearTheSymbol). ObserveOnce·Saga sell 루프·resolveCancel 은 청산 place·엔진 같은 주기 재시도가 범위 밖이 되어 not-applicable. 합계 23 번들, `check_analysis.py --change` rc=0.

## 1. 계약 (design freeze)

- [ ] 1.1 design.md — 갈래표(미접수 확정 / 결과 불명 / 확정 거절)를 FLM 반환 지점에 매핑, 같은 키 강제, 8s·최대 2회 시간 상한,
  재시도 중 재시작·ctx 취소·조회-재전송 경합, 손절·flatten 즉시성 영향, 401 기존 재전송(`client.go:344`)을 「확실히 미접수」로 넣고 근거를 시험으로 고정(사람 결정 3). — 첫 판(1.3 FAIL) → 2판 [design.md](design.md)(사용자 2차 결정 반영, 2026-10-23).
- [ ] 1.2 spec delta — order-execution 「Retry Matrix 산출물」·「취소의 일시적 거절 처리」 MODIFIED. — 2판: 「Retry Matrix 산출물」·「취소의 일시적 거절 처리」·「IN_DOUBT 해소」 MODIFIED(strict validate 통과).
- [ ] 1.3 proposal-freeze 리뷰(적대적 Eng 보이스 포함) → `review.md`. — 1차 2026-10-23 **FAIL(blocker 3: A1·S1·S2)**. D7 사람 결정(review.md 질문 1~7) 뒤 1.1·1.2 재작성 → 재리뷰. — 2차(2판) 2026-10-23 **FAIL(blocker F1·F2·F3)**. F3(R1·R2 생산 도달 0) 사람 결정 대기(review.md 「1.3 재리뷰(2판)」).

## 2. 구현 (RED → GREEN → REFACTOR → VERIFY)

- [ ] 2.1 Pre-Edit 선언.
- [ ] 2.2 RED: 취소 409 already-processing — 엔진 정책(확정 미수락·호출 1회·지연 0) / flatten 정책(재시도·누적 5s·창·소진 → IN_DOUBT·종목 보류·전송 전 실패 → IN_DOUBT) / code 불일치·모순·place·amend·정책 nil → 오늘과 같음 / 401 재전송 본문 동일·비 401 1회(design D10).
- [ ] 2.3 GREEN: 최소 구현 + 배선 2곳 + 정본 retry matrix §0.1·§1·§4 와 `execgw` 기본값을 같은 커밋에서 갱신.
- [ ] 2.4 attempt 기록에 재시도 횟수·갈래·확정 거절 분류(detail 꼬리표, 재시도 성공을 첫 시도 성공으로 기록하지 않음).

## 3. 검증

- [ ] 3.1 변이 원장 — 409 code 판정·kind·정책 구분·대기 상한·횟수·창·소진 되돌림·경계 재검사, 양성 대조군 포함(design D10).
- [ ] 3.2 다중 적대 리뷰(중복 주문 반례 레드팀) + gstack review/security.
- [ ] 3.3 영향 범위 시험(`internal/execgw`·`internal/official`·`internal/flatten`·엔진 clear 경로·배선, 필요 시 `-race`)과 미실행 범위 기록.
- [ ] 3.4 Manager 독립 검증.

## 4. 닫기

- [ ] 4.1 완료 게이트(사용자 명시 요청 시) → archive → Story 경로·PM sync.
- [ ] 4.2 검증된 교훈만 Hindsight retain.

> 착수 2026-10-23(Terra). 1.3 1차·2차 FAIL → 사용자 3차 결정(F3=1)으로 범위 = R3·R4·R5. design·delta 3판 → **1.3 2차 독립 리뷰(GPT-5.6 Terra) FAIL → 보완 완료(2026-10-28, design D5·D11) → 재리뷰 FAIL(P1 1·P2 1) → 보완 완료(design D11 r2·D12) → 3차 재리뷰 FAIL(P1 1·P2 1) → 보완 완료(r2 거래일 경계를 정규 범위로·D12 질의 2 범위 조인) → 4차 재리뷰(GPT) PASS, Manager 최종 승인 2026-10-28**. 다음 = 2.1 Pre-Edit.

## 0. 착수 준비

- [x] 0.1 proposal 의 열린 질문 1~4 에 사람 답을 받고 proposal·issues.md 에 기록한다. — 2026-10-23 「1~4번 권장대로」: 조건주문 제외 · (b) 8s 창 안에서만 · 401 재전송 「확실히 미접수」 인정(시험 고정 필요, 실측 근거 없음) · 409 재시도 a165 범위([issues.md](issues.md)).
- [x] 0.2 단일 writer 배정, `python3 tools/sdd/capture_change_base.py --change a165-safe-order-placement-retry` 로 base 고정. — Terra, base `34e12967e`([base-commit.txt](base-commit.txt)).
- [x] 0.3 `make sdd-sync` 후 CodeGraph hard evidence(`Gateway.Place`·`Cancel`·`Amend`·`submit`·`classifyMutation`·`Client.send` 의
  callers/callees/impact)와 `analysis/code-context/` 증거 조정 3파일. — codegraph 1.6.0, `raw/` 원문, CGC 는 다른 체크아웃(advisory 불가).
- [x] 0.4 FLM/BTM: 위 함수와 design 이 고르는 접수 조회 함수, 그리고 `withStrategyEntryGateAuthority`(클로저 호출 횟수) — High-risk, 면제 없음. — 9 번들(`analysis/function-logic/`, 생성기 `analysis/measure/gen_flm.py`), `check_analysis.py --change` rc=0.
- [x] 0.4.1 (1.3 S8/A10) 추가 번들 — 2차 결정 뒤 14 번들을 더했고(029e1d6f6), 3차 결정(F3=1) 뒤 진입 재전송 전용 10개(조회 경로·place·Amend·strategy 게이트 등)를 지우고
  남는 범위가 기대는 10개(New·buildGateway·buildFlattenWiring·resolveStep·planLiquidation·UnsettledCancelSymbols·classifyRefusalCode·PlaceOrder·CancelOrder·postWithHeaders)를 더했다(bc56522c7, issues.md 「3차」).
  **현재 24 번들**(2026-10-28 1.3 2차 처분으로 `Gateway.checkSymbolFree` 추가 — D11 latch 편집 대상, 실행 칸 미측정). 3차 결정 시점 23 번들: Gateway.submit·Gateway.Cancel·New·classifyMutation·statusOf·classifyRefusalCode·refusalCodes(execgw), buildGateway·ExitObserver.clearTheSymbol(engine), buildFlattenWiring(cmd/tossctl),
  Saga.settleStep·Saga.resolveStep·classifyCancel·Saga.planLiquidation(flatten), Attempt.DispatchVerified·ClassifyHTTPMutation·Journal.UnsettledCancelSymbols(journal), Client.send·Client.postWithHeaders·Client.PlaceOrder·Client.CancelOrder·tokenManager.refresh(official), Service.Cancel(trading).
  2026-10-28 HEAD `0c9bfd272` 에서 23 개 모두 `source_sha256` 일치(1.3 3차 (f)). 1판 문구의 조회 경로·`unwrapAndDecode`·`parseOrderFacts`·`newMatcher`·`roundTripFor`·`ObserveOnce` 번들은 이제 없다(범위 밖).

## 1. 계약 (design freeze)

- [x] 1.1 design.md 3판 — R3(엔진 취소 409 already-processing → FAILED_CONFIRMED, 지연 0)·R4(CLI flatten 취소 409 재시도, 누적 5s·2회·8s 창, 소진·중단·전송 전 실패 → 오늘 경로 AMBIGUOUS → `resolveStep`)·R5(401 재전송 근거 시험)의
  갈래를 FLM 반환 지점에 매핑, 즉시성 표(D4), 정책 nil = 오늘 동작(D5·D8), 기록(D6), 시험·변이 전략(D10). 진입 재전송·같은 키 재전송·조회-재전송 경합은 a171 로 이관(사용자 3차 결정).
  — 1판(1.3 FAIL) → 2판(재리뷰 FAIL F1~F10) → 3판 [design.md](design.md)(bc56522c7). 1.3 3차 처분으로 D10 에 배선 교차 변이·시험 추가(이번 커밋).
- [x] 1.2 spec delta 3판 — order-execution 「Retry Matrix 산출물」·「취소의 일시적 거절 처리」 MODIFIED(2판의 「IN_DOUBT 해소」 MODIFIED 는 진입 재전송과 함께 a171 로 빠짐).
  1.3 3차 처분으로 검증 도구 시나리오 둘의 적용 대상을 명시(이번 커밋). `openspec validate --strict` 통과.
- [x] 1.3 proposal-freeze 리뷰(적대적 Eng 보이스 포함) → `review.md`. — 1차 2026-10-23 **FAIL(blocker A1·S1·S2)** → 사용자 2차 결정 →
  2차(2판) 2026-10-23 **FAIL(blocker F1·F2·F3)** → 사용자 3차 결정 F3=1(R1·R2 를 a171 로; F2·F4·F6·A2·F8 이관, F1·F5·F7·F9·F10 은 3판에서 처분) →
  **3차(3판) 2026-10-28 PASS**(P0 0 · P1 2 는 이번 커밋 문서 수리 · P2 3 은 2.1 Pre-Edit 전 처분). 사람 결정 H1(R4 최악 지연 유지)·H2(gate ⑤ 창 — base 재고정 없음) 기록(review.md·issues.md).
  → **2차 독립 리뷰(GPT-5.6 Terra) 2026-10-28 FAIL(P1 2 · P2 2 · P3 1) → 보완 완료**(사람 결정 「1」): P2-2 바깥 래퍼를 D3·D5·submit FLM/BTM 에 일관 명세(`sendOnce` 폐기, B37 재적용·정착 ctx 분리·nil 동일성 AST 근거),
  P2-3 진입 전용 latch D11(schema 변경 없음, `checkSymbolFree` B3 뒤), D10 배선 4 시험 값+행동·ctx/latch 변이, 8s 창 시작 = dispatch 진입 직전. 새 번들 `checkSymbolFree`(24번째)·evidence AST 4.
  → **재리뷰(GPT-5.6 Terra) 2026-10-28 FAIL(P1 1·P2 1) → 보완 완료, 재리뷰 대기** — R3-1: D11 r2 를 latch 정착 transition(`attempt_transitions.id`) 뒤에 기록이 시작된 같은 대상 주문 CONFIRMED 취소로 한정, 새 질의 `ConfirmedCancelAfterLatch`(기존 `ConfirmedCancelOf` 불변), RED (f)~(h)·변이 ⑦~⑩, evidence AST 2(`Attempt.transition`·`Journal.Prepare`), schema 36 그대로 /
    R3-2: design D12 운영 절(읽기 전용 원장 조회·journalled 해제 r1/r2·`tossctl order cancel` 은 해제 아님·기존 `exit.liquidation_delayed` 경보), issues R-a165-5·6 처분 갱신.
  → **3차 재리뷰(GPT-5.6 Terra) 2026-10-28 FAIL(P1 1·P2 1) → 보완 완료, 재리뷰 대기** — R3-3: r2 비교에서 거래일을 빼고 대상 주문의 정규 범위(계좌·시장·종목·`target_order_id`)를 L 에서 읽어 비교, 번호 재사용은 주인 유일성(그 범위의 접수 확정 주인 intent 정확히 1)으로 막음, 순번 `r_C > s_L` 그대로,
    RED (h) 갱신(「다른 거래일 → 거절」 삭제)·(i)~(k) 추가, 변이 ⑩ 교체(거래일 같음 복원)·⑪(계좌·시장·종목 비교 제거)·⑫(주인 유일성 제거), evidence AST `execgw--gateway.preparerequest`, schema 36 그대로 /
    R3-4: D12 질의 1 에 계좌 열, 질의 2 를 r2 와 같은 범위 조인·주인 수·순번을 SQL 에서 판정하고 범위 열 출력(후보 목록 — 판정 아님). issues R-a165-5·6 갱신, 새 잔여 R-a165-7. **4차 재리뷰(GPT) PASS, Manager 최종 승인 2026-10-28.**

## 2. 구현 (RED → GREEN → REFACTOR → VERIFY)

- [x] 2.1 Pre-Edit 선언 — [pre-edit.md](analysis/pre-edit.md)에 1.3 3차 P2-1(gate ⑤ 창 형제 함수 처리 계획)·P2-2(바깥 래퍼 — design D5 로 확정)·P2-3(진입 전용 latch — design D11 로 확정) 처분을 기록함. `checkSymbolFree` coverprofile 66.7%를 측정했고, 기존 시험·fake를 고치지 않고 새 a165 시험 파일에 seam을 둠. 기준선 `go test ./internal/execgw`는 coverprofile 뒤 상위 프로세스가 25분 이상 종료하지 않아 TERM으로 중단했으며 PASS/FAIL 미판정으로 기록함.
- [x] 2.2 RED: 취소 409 already-processing — 엔진 정책(확정 미수락·호출 1회·지연 0) / flatten 정책(재시도·누적 5s·창·소진 → IN_DOUBT·종목 보류·전송 전 실패 → IN_DOUBT) / code 불일치·모순·place·amend·정책 nil → 오늘과 같음 / 401 재전송 본문 동일·비 401 1회 / 바깥 래퍼 (w1)~(w5) / 배선 W1~W4 / latch (a)~(k)(design D10·D11 — (h) 는 3차 재리뷰로 갱신, (i) 거래일을 넘긴 해제 · (j) 다른 계좌·시장·종목의 같은 번호 · (k) 번호 재사용·주인 0). — 원문 [red-output.txt](analysis/red-output.txt), D5 재전송 중 실시간 상한 RED 포함.
- [x] 2.3 GREEN: 최소 구현 + 배선 2곳 + 정본 retry matrix §0.1·§1·§4 와 `execgw` 기본값을 같은 커밋에서 갱신. design D12 를 `docs/operations.md` 운영 절로 옮긴다(r1 확인 SQL 확정 포함). — D5 4판은 `g.clk` 판정 및 `WithTimeout(dctx, Window)` 전송 상한으로 수정했고, retry dispatch timeout도 첫 409 AMBIGUOUS로 보존함.
- [x] 2.4 attempt 기록에 재시도 횟수·갈래·확정 거절 분류(detail 꼬리표, 재시도 성공을 첫 시도 성공으로 기록하지 않음). — retry 성공/미수락/중단 꼬리표 시험 및 변이로 고정함.

## 3. 검증

- [x] 3.1 변이 원장 — 409 code 판정·kind·정책 구분·대기 상한·횟수·창·소진 되돌림·경계 재검사, **배선 교차(flatten←엔진 정책 · 엔진←flatten 정책 · 배선 nil · 혼성 `{2, true}`)·`definitiveRefusalCodes` 에 already-processing 추가**(1.3 3차 P1-1·P3-3), **ctx(`winCtx` 부모 분리 · 정착 ctx 를 `winCtx` 로 · 재호출에 원 dctx · 창 시작점 이동)·latch(exit 도 latch 를 봄 · `unsettledFor` 에 넣음 · 해제 r1/r2 누락 · 종목 키 제거 · 메모리 latch · r2 순서 경계 제거 · r2 대상 id 비교 제거 · 순번을 정착 transition 으로 · r2 거래일 같음 조건 복원(옛 「거래일 비교 제거」 폐기) · r2 계좌·시장·종목 비교 제거 · r2 주인 유일성 제거)**(1.3 2차 독립 리뷰 P2·P1, 3차 재리뷰 R3-3), 양성 대조군 포함(design D10). — M01~M38 모두 잡힘, 살아남은 변이 없음([mutation-ledger.md](analysis/mutation-ledger.md)).
- [x] 3.2 다중 적대 리뷰(중복 주문 반례 레드팀) + gstack review/security. — gstack 적대 패스 P2(D5 retry timeout 결과 덮어쓰기)를 수리·M38로 고정했고, 보안/시험/유지보수 패스와 사후 독립 적대 패스는 발견 없음. 사후 Codex CLI 재호출은 저장소 훅의 `ai-memory query` 오류로 결론을 못 냈으며 review에 미완료 도구로 기록함.
- [x] 3.3 영향 범위 시험(`internal/execgw`·`internal/official`·`internal/flatten`·엔진 clear 경로·배선, 필요 시 `-race`)과 미실행 범위 기록. — 무태그·`tossos_testseams`·관련 `-race` 및 vet/gofmt 통과, 전체 시험과 `make gate`는 지시상 미실행([review.md](review.md)).
- [ ] 3.4 Manager 독립 검증.

## 4. 닫기

- [ ] 4.1 완료 게이트(사용자 명시 요청 시) → archive → Story 경로·PM sync.
- [ ] 4.2 검증된 교훈만 Hindsight retain.

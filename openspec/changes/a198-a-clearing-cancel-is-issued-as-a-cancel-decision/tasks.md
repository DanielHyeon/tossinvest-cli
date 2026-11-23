# a198 Tasks

> **High-risk(주문 취소·손절 경로).** 실계좌 주문·취소를 내는 자동 시험 금지. 픽스처는 합성 값만(실제 종목·수량·가격·주문번호·계좌 금지). 브로커는 가짜만 — **게이트웨이와 원장은 진짜를 쓴다**(이 change 의 요점).
> 전체 시험·전체 게이트는 종결 단계에서만. 원장 스키마·설정·토글 변경 없음. `RiskGuardian.IssueReduction` 본문·`submit` 의 제출 경로 무편집.
> 리뷰는 최대 2회차, 리뷰가 낸 새 가드·새 정책은 Manager 결정 전 구현 금지.

## 1. 계약과 증거 (production Go·시험 편집 0)

- [x] 1.1 Story `STORY-TOS-a198`(FEAT-TOS-009) 등록 확인, `base-commit.txt` 고정.
- [x] 1.2 CodeGraph hard evidence(버전 명기): `ExitObserver.clearTheSymbol`·`record`, `ReductionIssuer`, `Context.ExitObserver`, `RiskGuardian.IssueReduction`, `execgw.Issuer.IssueReduction`, `Gateway.Cancel`·`prepareRequest`·`checkPreimage`·`checkIdempotencyKey`, `journal.Prepare`, flatten 의 `decisionFor` — `analysis/code-context/`.
- [x] 1.3 `tools/logic-map` AST 를 `analysis/cited-ast/` 에 현재 HEAD 로 뽑고 proposal 표의 「조사 인용」「직접 읽음」 줄 전부와 design D1~D4·§2 의 주장을 대조한다. a197 브랜치(`feat/a197-external-opposite-buy-cancel`)의 probe 를 이 change 의 `analysis/probe/` 로 옮겨 **현재 HEAD 에서 다시 돌려** 원문을 남긴다(버리는 worktree 에서, 커밋하지 않는 시험 파일로). 반드시 답할 것: (a) 취소 결정에 운영 모드·킬 스위치·위험 체인 검사가 걸리는 자리가 있는가(design D2 의 전제 — 청산을 위한 취소는 어느 모드에서도 나가야 한다), (b) `checkPreimage` 가 취소 요청과 결정을 맞춰 보는 항목 전부(방향·수량·종목·시장·가격) — 미체결 수량이 부분 체결로 줄어든 주문을 취소할 때 결정의 최대 수량과 어떻게 비교되는가, (c) design §2-2·§2-3 의 기존 규칙(매도 취소의 종결 증거 대기, 미종결 attempt 가드, 지연 경보·연속 실패 계수)이 취소가 실제로 접수되는 흐름에서 어느 분기를 타는가. 틀리거나 설계를 바꾸는 사실은 문서를 고치지 말고 **Manager 에게 보고**.
- [x] 1.4 편집 표면 세기: `clearTheSymbol`·`Context.ExitObserver`·`ExitObserverOptions` 를 고정하는 구조 시험(a087·a090·a163 핀 등 grep 전수)과 a166 census(shapes.json·sites.json) 적중, design D2 의 두 안((나) Guardian 새 메서드 / (가) `execgw.Issuer` 주입) 각각의 편집 함수 수, 새 인터페이스 칸을 채워야 하는 기존 시험 하네스(시험 헬퍼·시험 함수) 수, 진행 중 change a087·a167 과의 겹침.
- [x] 1.5 rate limit 계상: retry matrix 문서의 취소 항목이 길 치우기의 취소를 이미 계상하는지 대조.
- [x] 1.6 편집 대상 기존 함수의 FLM/BTM 번들(pre-edit 판).
- [x] 1.7 `openspec validate a198-… --strict` 통과, proposal-freeze 적대 리뷰(다른 계열, 최대 2회, 적대적 Eng 관점 필수) — `review.md`.
- [ ] 1.8 Manager 1단계 검증·freeze.

## 2. Pre-Edit

- [ ] 2.1 Pre-Edit 선언을 `review.md` 에 확정.

## 3. RED → GREEN

- [ ] 3.1 RED: `internal/app/engine/a198_clearing_cancel_test.go` — design §4-1(현재 HEAD 에서 결함 재현, 원문 `analysis/tdd/red.txt`)과 2~8.
- [ ] 3.2 GREEN: 취소 결정 발행 새 함수 + `clearTheSymbol` 편집(발행 호출·실패 기록) + `Context.ExitObserver` 공급.
- [ ] 3.3 기존 a094 시험 하네스의 새 인터페이스 대응(design §4-9, 번들과 함께). 편집 뒤 번들 AST·FLM·BTM 재생성.

## 4. 검증

- [ ] 4.1 변이 원장(`analysis/mutation/`) — design §4 의 변이 6개, 닿음 확인·무변이 대조군 GREEN 뒤.
- [ ] 4.2 선택 시험 원문(design §4), `make lint`, `openspec validate --strict`, `check_analysis.py --change a198-…`. 미실행 범위 기록.
- [ ] 4.3 구현 적대 리뷰(다른 컨텍스트) → gstack review 1회 — `review.md`.

## 5. 종결

- [ ] 5.1 Manager 검증(diff·시험 원문 재실행).
- [ ] 5.2 완료 게이트(격리 worktree) → archive → Story 경로 갱신 → push.
- [ ] 5.3 (사람 승인) 운영 반영 — design §5.

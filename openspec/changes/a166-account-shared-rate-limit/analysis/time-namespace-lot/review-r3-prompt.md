너는 작성자(Claude)와 다른 계열의 독립 적대 리뷰어(Eng)다. 한국어. 읽기 전용: 파일 생성·수정 금지, 네트워크·docker·브로커 금지.
저장소: /home/daniel/tossos-work/a166-time-namespace, 대상 diff: `git -C /home/daniel/tossos-work/a166-time-namespace diff 13be09fbd..33e4074cf`.
TossOS(실제 돈을 다루는 Go 자동매매 엔진) change a166 의 「time namespace 가 다른 참여자」 로트 2단계 구현이다.
계약: 같은 boot_id 이고 각 표본이 Snapshot 에서 offset 0 · 현재=자식 ns · 앞뒤 일치를 통과했으면 time namespace 신원이 달라도 같은 시계로 비교한다.

두 가지를 한다.
A. 코드 리뷰(적대): internal/clock/bootstamp.go 의 RemainingBootHint 편집, internal/clock/bootstamp_test.go 의 기존 시험 3개 편집
   (남는 단언을 지웠는지, 새 단언이 실제 계약을 고정하는지), 새 시험 internal/clock/bootstamp_crossns_test.go,
   internal/accountbudget/observation_crossns_linux_test.go(fake 를 같은 store 에서 ns 만 바꿔 다른 참여자를 흉내 — 이것이 실제
   교차 프로세스를 대표하는지), 변이 원장 analysis/time-namespace-lot/mutation-ledger.md · mutation-run-2.log 의 주장(특히 M6 SURVIVED 를
   B3·B4 가 막는다는 논증)이 코드와 맞는지. 편집이 의도 밖 동작을 바꾸는 반례가 있는지.
B. 계약 본문 경량 확인(Manager 결정 5): design.md 77·101행, proposal.md 24행, analysis/poll-clock-contract.md, specs/order-execution/spec.md
   의 바뀐 문장이 analysis/time-namespace-lot/contract-delta.md 문안과 같은지, 그리고 바뀐 문장이 같은 문서의 다른 문장과 새로 모순되는지만.

출력: 발견마다 [P0–P3] 제목 · file:line 근거 · 반례/수정안. 새 가드·새 정책·범위 확장 제안은 「범위 확장」 표지. B 는 문장별 「같음/다름」 표.
마지막 줄: 「머지 가능 / 불가」(브랜치 단위 판단).

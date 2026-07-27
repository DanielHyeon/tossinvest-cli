# Tasks: verify-clears-leftovers

- [x] 1.1 [T] 정리 대상 선정 — 기록의 `Outstanding`에서 이 도구가 만든 객체만.
  주문은 항상 대상, 조건주문은 `conditional-cancel`이 settled인 경우에만 대상.
  RED: 잔여 주문이 있는 기록에서 대상이 비어 있음 → GREEN. 존속 대기 중인 조건주문은
  대상이 아님을 별도 테스트로 고정.
- [x] 1.2 [T] 계획에 정리 줄 — `Step: StepCleanup`, `cancel-order`/`cancel-conditional`,
  대상 심볼, 수량 없음. 대상이 없으면 줄도 없다(빈 계획에 정리 헤더만 남지 않는다).
  RED: 잔여물이 있어도 계획에 줄이 없음 → GREEN.
- [x] 1.3 [T] prologue 실행 — 승인 뒤 catalogue보다 먼저. `r.gate`를 통과하므로 계획
  밖 요청은 여전히 거절된다. RED: 잔여 주문이 취소되지 않고 남음 → GREEN(취소되고
  `Outstanding`이 비며 후속 mutating 단계가 상한에 걸리지 않는다).
- [x] 1.4 [T] 정리 실패는 실행을 죽이지 않는다 — fail로 기록하고 계속 간다. 뒤따르는
  mutating 단계는 기존 노출 상한이 거절한다. RED/GREEN.
- [x] 1.5 [T] 정리는 측정이 아니다 — `kind: "cleanup"`. `StepCount`·`RedoSet`·리포트·
  콘솔 단계 표가 능력 측정으로 세지 않는다. RED/GREEN.
- [x] 1.6 [T] 콘솔: 끝난 실행이 화면에 있어도 시작 섹션이 보인다. `Spent`·"이어할 단계가
  없다" 가드는 그대로 비활성화한다. 진행 중인 실행에서는 시작 섹션을 감춘다.
  RED: 0단계로 끝난 실행 뒤 폼이 사라짐 → GREEN.
- [x] 1.7 만료 문구 교정 — 콘솔 승인에는 확인 문자열이 없다.
- [x] 2.1 Function Logic Map + `check_analysis.py`
- [x] 2.2 PM registry allowlist + fixture, `make sdd-sync && make sdd-check && make gate`
- [x] 2.3 measurements.md에 3차 실행(M18~M21)과 도구 갭(M22 잔여물 교착, M23 승인 창
  만료 3회) 기록

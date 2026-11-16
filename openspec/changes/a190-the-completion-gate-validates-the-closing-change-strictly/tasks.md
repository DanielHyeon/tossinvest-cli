## 1. 계약과 증거

- [x] 1.1 Story `STORY-TOS-a190` 를 registry·`FEAT-TOS-001` 에 등록하고 proposal·design·spec delta·tasks 를 작성한다. `openspec validate <id> --strict` 통과.
- [x] 1.2 `base-commit.txt` 를 고정하고 CodeGraph·CodeGraphContext 증거와 기억 회고를 `analysis/code-context/` 에 기록한다.
- [x] 1.3 편집 전 `tools/gate.sh` 해당 구간을 `analysis/pre-edit/gate-before.txt` 에 원문으로 보존하고, 오늘의 strict 부채 목록을 `analysis/strict-debt/` 에 남긴다(D6).
- [x] 1.4 계약에 proposal-freeze 리뷰를 돌리고 `review.md` 에 기록한다. D1~D6 을 바꿔야 하는 발견이면 멈추고 Manager 에게 올린다.

## 2. 구현 (RED → GREEN)

- [ ] 2.1 RED: `tools/sdd/test_gate_validate_step.py` 를 쓰고 편집 전 트리에서 실패를 확인한다(원문 `analysis/tdd/red.txt`).
- [ ] 2.2 GREEN: `tools/sdd/check_openspec_validate.py` 를 쓰고 `tools/gate.sh` 11단계를 그것으로 바꾼다. `test_race_detector_actually_runs.py` 단계 수 단언을 고친다.
- [ ] 2.3 변이 원장: 사본에서 무변이 대조군 GREEN 확인 뒤 7개 이상 변이를 돌린다. 하네스·원문은 `analysis/mutation/`.
- [ ] 2.4 문서: `docs/WORKFLOW.md` ⑪ 서술 두 곳과 `tools/gate.sh` 머리 주석(D4 두 목록)을 맞춘다.

## 3. 검증과 리뷰

- [ ] 3.1 새 시험, `make sdd-test`, 실제 저장소에서 새 11단계 판정만 a189 id 로 돌린 결과를 `analysis/verify/` 에 원문으로 남긴다. `make gate`·`make test` 전체는 돌리지 않는다(Manager).
- [ ] 3.2 gstack `review` + 다른 계열 적대 보이스 1개를 최종 diff 에 돌리고 발견 표와 처리를 `review.md` 에 기록한다(최대 2회차).

## 4. 종결 (Manager · 사용자 지시 대기)

- [ ] 4.1 Manager 독립 재검증.
- [ ] 4.2 `make gate CHANGE=a190-the-completion-gate-validates-the-closing-change-strictly`.
- [ ] 4.3 archive 하고 Story 경로와 PM 파생 상태를 동기화한다.
- [ ] 4.4 검증된 교훈을 ai-memory·Hindsight 에 남긴다.

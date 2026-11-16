## 1. 계약과 증거

- [x] 1.1 Story `STORY-TOS-a190` 를 registry·`FEAT-TOS-001` 에 등록하고 proposal·design·spec delta·tasks 를 작성한다. `openspec validate <id> --strict` 통과.
- [x] 1.2 `base-commit.txt` 를 고정하고 CodeGraph·CodeGraphContext 증거와 기억 회고를 `analysis/code-context/` 에 기록한다.
- [x] 1.3 편집 전 `tools/gate.sh` 해당 구간을 `analysis/pre-edit/gate-before.txt` 에 원문으로 보존하고, 오늘의 strict 부채 목록을 `analysis/strict-debt/` 에 남긴다(D6).
- [x] 1.4 계약에 proposal-freeze 리뷰를 돌리고 `review.md` 에 기록한다. D1~D6 을 바꿔야 하는 발견이면 멈추고 Manager 에게 올린다.

## 2. 구현 (RED → GREEN)

- [x] 2.1 RED: `tools/sdd/test_gate_validate_step.py` 를 쓰고 편집 전 트리에서 실패를 확인한다(원문 `analysis/tdd/red.txt`).
- [x] 2.2 GREEN: `tools/sdd/check_openspec_validate.py` 를 쓰고 `tools/gate.sh` 11단계를 그것으로 바꾼다. `test_race_detector_actually_runs.py` 단계 수 단언을 고친다.
- [x] 2.3 변이 원장: 사본에서 무변이 대조군 GREEN 확인 뒤 7개 이상 변이를 돌린다. 하네스·원문은 `analysis/mutation/`.
- [x] 2.4 문서: `docs/WORKFLOW.md` ⑪ 서술 두 곳과 `tools/gate.sh` 머리 주석(D4 두 목록)을 맞춘다.

## 3. 검증과 리뷰

- [x] 3.1 새 시험, `make sdd-test`, 실제 저장소에서 새 11단계 판정만 a189 id 로 돌린 결과를 `analysis/verify/` 에 원문으로 남긴다. `make gate`·`make test` 전체는 돌리지 않는다(Manager).
      > 새 시험 39 OK(openspec 있음)·31 OK skip 8(없음, CI 모사) `analysis/tdd/green*.txt`. `make sdd-test` rc 0(기본 TMPDIR — `$HOME/tossos-work` 아래 TMPDIR 두 판은 기존 소켓 시험이 `AF_UNIX path too long`, `history-sdd-test-run1·2`). 새 11단계 a189·a187·a190 모두 rc 0 `analysis/verify/step11-*.txt`. gate ⑤ 단독 required 0 rc 0 `check-analysis.txt`.
- [x] 3.2 gstack `review` + 다른 계열 적대 보이스 1개를 최종 diff 에 돌리고 발견 표와 처리를 `review.md` 에 기록한다(최대 2회차).
      > 1회차 Codex CHANGES REQUESTED(P2 둘) + gstack 전문가 3·Claude 적대 → 고침 7. 2회차 Codex CHANGES REQUESTED(신규 P2 둘)·Claude 적대 → 고침 6. 2회차 수리는 3회차 리뷰를 받지 않았다(`review.md`).

## 4. 종결 (Manager · 사용자 지시 대기)

- [x] 4.1 Manager 독립 재검증.
      > 2026-11-16 Manager: 격리 worktree 새 시험 31 OK·단계 수 8 OK, 실제 트리 a189·a187·a190 rc 0·없는 id rc 1, 독립 행동 검사·변이 4종. 원문 `analysis/verify-manager/`, 기록 `review.md` 「Manager 재검증」.
- [x] 4.2 `make gate CHANGE=a190-the-completion-gate-validates-the-closing-change-strictly`.
      > 사용자 지시(2026-11-16, Manager 전달)로 종결 팀메이트 Opus 가 격리 worktree 에서 실행. gate ②의 미완료 0건 선행 조건으로 gate 전에 체크함(선례 a165 4.1) — 실제 11단계 결과는 PASS 뒤 `review.md` 「종결」·`analysis/gate/` 에 기록함.
- [x] 4.3 archive 하고 Story 경로와 PM 파생 상태를 동기화한다.
      > 같은 지시. gate ② 선행 조건으로 gate 전에 체크함 — archive 경로·Story 경로·PM check 영수증은 gate PASS 뒤 커밋에 남김.
- [x] 4.4 검증된 교훈을 ai-memory·Hindsight 에 남긴다.
      > 2026-11-16 Manager 가 Hindsight bank `tossos-dev-episodes` 에 retain 수행(a187·a189·a190 의 검증된 교훈 7항목, 시크릿·계좌·포지션·금액 없음 — Manager 전달 사실). ai-memory 는 훅 자동 기록이며 이 종결에서 durable page 는 쓰지 않음.

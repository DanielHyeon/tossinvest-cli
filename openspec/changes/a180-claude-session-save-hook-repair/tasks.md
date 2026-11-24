## 1. 계약과 증거

- [x] 1.1 Story `STORY-TOS-a180` 를 registry·Feature 에 등록하고 proposal·design·spec delta 를 작성한다.
      > 구현보다 나중에 작성했다(사후 계약). 역순은 `review.md` 에 기록했다.
- [x] 1.2 `base-commit.txt` 를 고정하고 CodeGraph·CodeGraphContext·GBrain 증거를 `analysis/code-context/` 에 기록한다.
      > 구현은 base 뒤 작업 트리에 미커밋 상태로 있어 비교 창에 그대로 잡힌다.
      > 2026-11-25 base 재고정 `232d6b86c` → 귀속 실측 커밋(Manager 승인 2026-11-25, 사람 승인 절차). 자기 Go 커밋 0, 옛 base required 55 는 전부 형제 착지 몫 — `review.md` 「base 재고정 — Manager 승인 2026-11-25」, `analysis/base-refix/attribution.txt`.

## 2. 구현과 검증

- [x] 2.1 RED 시험을 먼저 작성해 실패를 확인한다.
      > 초기 7개 중 6개 실패, 1개는 기존 동작 고정. gstack 리뷰 뒤 보강한 17개는 수리 전 7건 실패(FAIL 6, ERROR 1) 뒤 GREEN.
      > 그 7건 중 메모 시험 1건은 제 단언 오류였고 시험을 고쳤다. index 시험은 첫 수리 뒤에도 실패해 `git diff` 의 index 쓰기를 드러냈다.
- [x] 2.2 훅 설정과 `save-session.sh` 를 수리한다.
- [x] 2.3 사본 변이로 시험이 동작을 잡는지 확인한다.
      > `analysis/harness/mutate_save_session.py`: 대조군 통과, 변이 31개 전부 CAUGHT. 첫 판 12개는 리뷰가 추가한 변이 9개를 놓쳐 늘렸다.
      > 이 숫자는 하네스에 적은 변이만 말한다.
- [x] 2.4 영향 범위 시험과 동기화 검사를 실행하고 결과와 미실행 범위를 `review.md` 에 기록한다.

## 3. 리뷰

- [x] 3.1 gstack `review` 를 작성자와 분리된 컨텍스트에서 실행하고 판정을 `review.md` 에 기록한다.
      > 1차 CHANGES REQUESTED. 스킬 기준 INCOMPLETE(`review.md` 「gstack 리뷰」).
- [x] 3.2 리뷰 지적 수리분을 재리뷰한다.
      > 2차 CHANGES REQUESTED(흔들리는 시험 1건, 문서 4건). 지적대로 고쳤다. 고친 뒤 APPROVE 판정은 받지 않았다(`review.md` 「2차」).

## 4. 종결 (사용자 명시 요청 대기)

- [x] 4.1 새 세션에서 훅 실패 수, 도구 호출 지연, `.git/index.lock` 충돌, `/exit` 직후 저장 완료 여부, `save-failures.log` 를 잰다.
      > 2026-11-05 측정 결과는 `analysis/verification.md` 「tasks 4.1 실측」. 직전 세션 `/exit` 가 VS Code 확장에서 실행되지 않아, `/exit` 항목은 터미널 CLI 임시 저장소 재현 1건을 대체 증거로 삼았고 사용자가 2026-11-05 이를 인정했다. VS Code 패널 닫기와 실제 저장소 동시 `git add` 경합은 재지 않았다.
- [ ] 4.2 커밋한다.
- [ ] 4.3 `make gate CHANGE=a180-claude-session-save-hook-repair` 를 실행한다. 전체 시험을 포함하므로 사용자 명시 요청이 필요하다.
- [ ] 4.4 archive 하고 Story 경로와 PM 파생 상태를 동기화한다.
- [ ] 4.5 검증된 교훈을 ai-memory·Hindsight 에 남긴다.

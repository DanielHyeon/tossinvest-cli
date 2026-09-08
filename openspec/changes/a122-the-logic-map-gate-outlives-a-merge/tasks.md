## 1. Contract and hard evidence

- [ ] 1.1 `STORY-TOS-a122` 를 등록하고 구현 base 를 잡고 이 change 를 strict 로 검증한다.
- [ ] 1.2 CodeGraph 로 `check_analysis.py` 5단계 호출 사슬(`tools/gate.sh` → `main` →
      `check` → `changed_existing_functions` / `resolve_base`)과 그 호출자를 열거한다.

      > **Function Logic Map 은 not-applicable 이다 — 침묵한 생략이 아니다.**
      > 이 change 가 편집하는 함수는 Python(`tools/logic-map/check_analysis.py`)이고,
      > 이 저장소의 FLM 산출물은 `tools/logic-map/extract_go_ast.go` 가 **Go** 함수에서
      > 뽑는다(`docs/WORKFLOW.md` — "Python AST 분석기는 Go 코드에 재사용하지 않는다").
      > Go 파일을 편집하게 되면 그 시점에 해당 함수의 FLM 을 먼저 만든다.

- [ ] 1.3 착지 지점을 무엇으로 볼지 정한다 — 후보(merge 커밋 · 마지막 구현 커밋 ·
      아카이브 커밋)를 나열하고, 각 후보가 **위조 가능한 방식**을 먼저 적은 뒤 고른다.
- [ ] 1.4 proposal-freeze 적대 리뷰와 gstack 리뷰를 마치고 기록한다.

## 2. RED

- [ ] 2.1 착지 지점이 기록된 change 에 대해, base 와 착지 지점 사이에 없는 함수를
      요구하면 실패하는 테스트를 만든다. 픽스처는 실제 병합 모양을 재현해야 한다 —
      base → 착지 → 그 뒤 다른 change 의 커밋 셋.
- [ ] 2.2 착지 지점이 **없는** change 는 지금과 똑같이 워킹트리와 비교한다는 것을
      고정하는 테스트를 만든다. 이것이 없으면 이 change 가 진행 중 change 의 판정을
      조용히 바꾼다.
- [ ] 2.3 위조된 착지 지점(존재하지 않는 커밋 · base 보다 앞선 커밋 · 이 change 의
      작업을 담지 않은 커밋)이 통과하지 못하고 사유를 이름으로 말하는 테스트를 만든다.
- [ ] 2.4 착지 지점을 **넣기만 하면 5단계가 통과한다** 는 변이를 만들어 2.3 이 그것을
      잡는지 확인한다. 못 잡으면 2.3 이 존재 검사에 그친 것이다.

## 3. GREEN

- [ ] 3.1 착지 지점 기록의 읽기와 유효성 판정을 최소 구현한다.
- [ ] 3.2 `changed_existing_functions` 의 `target` 을 그 값으로 넘긴다. 기록이 없으면
      빈 문자열을 유지해 현재 동작을 보존한다.
- [ ] 3.3 5단계 실패 메시지가 "왜 이 함수가 요구되는가"를 말하도록 한다 — 오늘의
      메시지는 316개 함수 이름만 쏟아내고 그것이 남의 change 것이라는 사실을 말하지 않는다.

## 4. VERIFY and handoff

- [ ] 4.1 아카이브된 a075·a076 을 회귀 픽스처로 쓴다. 착지 지점을 주면 통과하고
      주지 않으면 오늘과 같이 실패해야 한다 — 이 change 가 실제로 그 다섯을 푸는지의 증거다.

      > **선결 조건 하나가 이미 드러났다.** 아카이브된 change 는 지금 id 로 재검사가
      > 되지 않는다 — `resolve_base` 가 `openspec/changes/<id>/base-commit.txt` 만 보고
      > `archive/<YYYY-MM-DD>-<id>/` 를 보지 않아서 `missing base-commit.txt` 로 떨어진다
      > (2026-09-08 실측). 커밋 `f6965ebb` 이 고친 것은 `resolve_referenced_change`(빌린
      > 증거 쪽)이고 이 경로가 아니다. 픽스처를 쓰려면 이것부터 닫아야 한다.
- [ ] 4.2 a074 · a077 · a079 에 대해 실행해 요구되는 함수 집합이 각 change 가 실제로
      고친 것으로 줄어드는지 확인하고 그 수를 기록한다.
- [ ] 4.3 focused 테스트와 `make test` · `make vet` · `make validate` · `make sdd-sync` ·
      `make sdd-check` 를 돌린다.
- [ ] 4.4 독립 적대 diff/테스트 리뷰와 gstack 리뷰를 마친다.
- [ ] 4.5 PM 동기화 후 `make gate CHANGE=a122-the-logic-map-gate-outlives-a-merge` 를
      돌리고 성공한 뒤에만 아카이브한다.

## 5. 이 change 가 열어 두는 것

- [ ] 5.1 a074 · a077 · a079 는 각자의 배포 후 실측이 남아 있어 이 change 로 닫히지
      않는다. 이 change 는 그 셋의 **게이트를** 답할 수 있게 만들 뿐이다.
- [ ] 5.2 2026-09-08 에 a075 · a076 이 5단계를 사유를 적고 면제한 기록(각 review 의
      §완료 게이트 5단계 면제)을 소급해 고치지 않는다. 그 면제는 당시 사실이다.

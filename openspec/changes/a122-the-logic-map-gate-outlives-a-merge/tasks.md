## 1. Contract and hard evidence

- [x] 1.1 `STORY-TOS-a122` 를 등록하고 구현 base 를 잡고 이 change 를 strict 로 검증한다.

      **확인 (2026-09-09).** `_registry.yaml:151` 에 등록, `base-commit.txt` =
      `7f8b1ae7`, `openspec validate --strict` 통과.
- [x] 1.2 CodeGraph 로 `check_analysis.py` 5단계 호출 사슬(`tools/gate.sh` → `main` →
      `check` → `changed_existing_functions` / `resolve_base`)과 그 호출자를 열거한다.

      > **Function Logic Map 은 not-applicable 이다 — 침묵한 생략이 아니다.**
      > 이 change 가 편집하는 함수는 Python(`tools/logic-map/check_analysis.py`)이고,
      > 이 저장소의 FLM 산출물은 `tools/logic-map/extract_go_ast.go` 가 **Go** 함수에서
      > 뽑는다(`docs/WORKFLOW.md` — "Python AST 분석기는 Go 코드에 재사용하지 않는다").
      > Go 파일을 편집하게 되면 그 시점에 해당 함수의 FLM 을 먼저 만든다.

      **산출물 (2026-09-09).** `analysis/gate-call-chain.md`. 사슬은
      `tools/gate.sh:251` → `check_analysis.py:660 main` → `:587 check` →
      `:259 resolve_base` · `:111 changed_existing_functions` 이고, 각각 호출자는
      하나뿐이다(파일 밖 사용자는 `execution_baseline.py:271` 의 지연 import 하나).

- [x] 1.3 착지 지점을 무엇으로 볼지 정한다 — 후보(merge 커밋 · 마지막 구현 커밋 ·
      아카이브 커밋)를 나열하고, 각 후보가 **위조 가능한 방식**을 먼저 적은 뒤 고른다.

      **결정 (2026-09-09).** `analysis/landing-point.md`. 파생 후보(디렉터리를 마지막
      으로 만진 커밋)·merge 커밋·아카이브 커밋을 **측정으로** 기각하고 선언 파일
      `landed-commit.txt` 를 고른다. 느슨해지는 위조 다섯을 표로 적고 각각의 판정을
      붙였다 — 엄해지는 방향은 일부러 안 막는다.

      > **측정이 전제 하나를 반증했다.** `448dfeb1..840b3377` 의 바뀐 Go 파일이 **0**
      > 이다. 2026-08-04 재기준화가 base 를 a074~a079 의 Go 작업 **뒤로** 옮겼으므로,
      > 이 다섯에는 "이 change 가 고친 함수"를 돌려주는 target 이 없다. 이 change 가
      > 주는 것은 올바른 질문이 아니라 **답할 수 있는 질문**이다. proposal 을 그렇게
      > 정정했다.
- [ ] 1.4 proposal-freeze 적대 리뷰와 gstack 리뷰를 마치고 기록한다.

      **1차 실행 (2026-09-09, 리뷰어 3 — 위조·재측정·파급).** review.md §독립 적대
      리뷰 1회차. 판정 (4) 가 존재 검사라 위조가 성립했고(High-risk 48개 은닉),
      증거 검사로 바꾸어 막았다. 아래 1.5~1.13 이 남은 지적이다.

- [x] 1.5 `landed-commit.txt` 를 워킹트리가 아니라 **커밋에서** 읽는다. 지금 모양은
      untracked·심볼릭 링크·게이트 통과 후 삭제가 전부 가능하다. a063 경로에는 이미
      `_regular_committed`(`execution_baseline.py:398-441`)가 있다.
- [x] 1.6 조상 판정을 `base-commit.txt` 의 글자가 아니라 `resolve_base` 가 **반환한**
      값(a063 은 E)에 건다. 지금 설계대로면 a063 에 846 파일 역방향 diff 가 난다.
- [x] 1.7 값을 40자리 hex 로 제한한다. `rev-parse` 는 `HEAD`·브랜치·태그를 받으므로
      `echo HEAD > landed-commit.txt` 한 줄이 커밋 안 된 Go 편집을 전부 요구에서 뺀다.
- [ ] 1.8 빌린 증거(`function-logic-reference.txt`)의 착지 공유 규칙을 정한다. base 는
      정확히 같아야 한다는 규칙이 있는데(`check_analysis.py:614`) 착지에는 없다.
- [x] 1.9 해소한 착지 SHA 와 요구 함수 수를 **항상** 출력한다. 5단계 기록에 어떤
      대상으로 통과했는지가 남아야 한다.
- [x] 1.10 착지 커밋에서의 change 디렉터리 경로를 정한다 — 아카이브·renumber 전의
      `openspec/changes/<id>/` 다. 경로가 아니라 내용으로 대조한다.
- [ ] 1.11 `tools/gate.sh:116` · `:184` 가 활성 경로를 하드코딩해 아카이브된 id 가
      5단계에 **도달하지 못한다**. 이것을 안 고치면 `check` 의 아카이브 인식은
      `make gate` 로는 못 쓴다.
      **2026-09-09 실물 증거 — a099 가 `:184` 에 막혔다.** `landed-commit.txt` 를
      `e6c4636a` 로 기록하니 5단계가 `required 32 / errors 0` 으로 통과하는데,
      `make gate CHANGE=a099-a-claim-excludes-the-second-sender` 는 거기까지 가지도
      못하고 **3단계**에서 죽는다: `짝 change 의 tasks.md 가 없습니다:
      openspec/changes/a098-nobody-sends-what-the-outbox-keeps/tasks.md`. a098 은
      2026-08-29 에 `archive/2026-08-29-a098-…` 로 옮겨졌고 `PAIR_DIR` 은 그것을
      보지 않는다. 즉 **짝이 먼저 아카이브되면 남은 쪽의 완료 게이트가 영구히
      막힌다** — [[borrowed-flm-evidence-goes-stale]] 가 `f6965ebb` 로 고친 것과
      같은 모양이 배포 짝 축에서 되풀이된 것이다. 고칠 때 `resolve_referenced_change`
      의 해결 규칙(날짜 접두사를 벗긴 나머지 일치, 중복이면 fail-closed)을 재사용할지
      결정한다.
      **거부할 정상 입력 열거** — 이 완화가 죽이면 안 되는 것: (a) 아직 활성인 짝
      (지금 경로로 찾힌다), (b) 존재하지 않는 오타 id 는 **계속 실패해야 한다**,
      (c) 활성과 아카이브 양쪽에 같은 id 가 있으면 [[stacked-changes-break-the-gate]]
      의 3.2.4 와 같은 그늘이 생기므로 fail-closed.
- [ ] 1.12 a063 의 `execution_baseline.validate` 가 착지 기록을 감사하지 않는다.
      키 집합 열거(`execution_baseline.py:410-412`)에 넣을지 정한다.
- [x] 1.13 ~~a077 의 증거가 자기 base 보다 낡았다~~ — **철회.** 근거였던 후보 탐색이
      `git log -- <file>` 의 이력 단순화 때문에 병합 커밋 `448dfeb1` 을 빠뜨렸다.
      블롭을 직접 해싱하니 a077 의 7개 번들 전부가 거기서 맞는다.

      대신 **a075** 에 진짜 결함이 있다 — `cmd-tossctl--runconsole` 의 branch-test-map 이
      트리에 없는 `TestContainerBuildsDoNotStageALocalUpdate` 를 인용한다. a075 의
      결함이고 a122 가 덮지 않는다.

## 2. RED

- [x] 2.1 착지 지점이 기록된 change 에 대해, base 와 착지 지점 사이에 없는 함수를
      요구하면 실패하는 테스트를 만든다. 픽스처는 실제 병합 모양을 재현해야 한다 —
      base → 착지 → 그 뒤 다른 change 의 커밋 셋.
- [x] 2.2 착지 지점이 **없는** change 는 지금과 똑같이 워킹트리와 비교한다는 것을
      고정하는 테스트를 만든다. 이것이 없으면 이 change 가 진행 중 change 의 판정을
      조용히 바꾼다.
- [x] 2.3 위조된 착지 지점(존재하지 않는 커밋 · base 보다 앞선 커밋 · 이 change 의
      작업을 담지 않은 커밋)이 통과하지 못하고 사유를 이름으로 말하는 테스트를 만든다.
- [x] 2.4 착지 지점을 **넣기만 하면 5단계가 통과한다** 는 변이를 만들어 2.3 이 그것을
      잡는지 확인한다. 못 잡으면 2.3 이 존재 검사에 그친 것이다.

      > **이 변이는 이미 한 번 통과했다.** 1차 설계의 판정 (4)(착지 커밋에 같은
      > `base-commit.txt` 가 있다)가 정확히 그 존재 검사였고, a112 에 `5c5efec7` 을
      > 주면 넷 다 통과하며 비테스트 Go 48개가 숨었다. 2.3 은 그 커밋을 **번들
      > 불일치 43건**으로 거절해야 한다.

- [x] 2.5 `changed_existing_functions` 호출에서 `target` 인자를 **빼는** 변이가 스위트를
      통과하는지 본다. 지금 여섯 patch 자리가 전부 `return_value` 라 인자를 안 본다
      (`test_check_analysis.py:83-90` 외 다섯). 통과하면 인자를 단언하는 시험을 만든다.
- [x] 2.6 요구 집합이 빈 사실의 출력은 `context` 로 낸다. `check` 의 반환 리스트에
      붙이면 `main()` 이 1을 돌려 5단계가 **실패**한다(`check_analysis.py:667-670`).

## 3. GREEN

- [x] 3.1 착지 지점 기록의 읽기와 유효성 판정을 최소 구현한다.
- [x] 3.2 `changed_existing_functions` 의 `target` 을 그 값으로 넘긴다. 기록이 없으면
      빈 문자열을 유지해 현재 동작을 보존한다.
- [x] 3.2.1 `validate_target` 의 `revision: current` 해싱 대상을 착지 커밋으로 바꾼다.
      비교 대상과 증거 대조 대상이 갈리면 정본이 둘이 된다.
- [x] 3.2.2 `check` 의 `change_dir`(`check_analysis.py:590`)을 아카이브 인식으로 바꾼다.
      `resolve_referenced_change`(`:231`)가 이미 그 문법을 안다 — 세 번째 사본을 만들지
      말고 공유한다.
- [ ] 3.2.3 인용된 테스트 이름의 탐색 범위를 정한다. 지금은 착지 지점을 줘도
      `test_index(root)` 가 **워킹트리**를 훑는다. 증거가 착지를 기술하면 그 인용도
      착지에서 찾는 것이 일관되지만 spec SHALL 은 source hash 만 말한다 — 넓힐지
      결정하고 사유를 적는다.
- [ ] 3.2.3.1 **번들이 0 인 change 는 착지 지점을 고정할 것이 없다.** 증거 판정은
      `revision: current` 번들을 훑는데, 번들이 없으면 그 순회가 빈 채로 통과해
      조상 판정만 남는다. 그 상태에서 면제 표식(`Function Logic Map: not-applicable`)
      까지 있으면 착지를 freeze 바닥으로 잡아 `required=0` 을 만들고 통과시킬 수 있다.
      2026-09-09 측정: `align-full-sdd-pm-contract` 은 표식이 **없어서** 안 뚫렸고
      (base 자신 → `missing analysis or marker`, base+1 → 53개 요구), 아카이브된
      a076 은 표식이 있어 그 경로로 통과했다 — 값은 정직했지만 기제는 못 가른다.
      막을 때 **거부할 정상 입력**을 먼저 열거할 것: base 가 이미 자기 작업을 담은
      change(a074·a075·a076·a079 가 전부 그 모양이다)를 같이 죽이면 안 된다.
- [ ] 3.2.4 열린 디렉터리가 아카이브본을 가린다. `resolve_referenced_change` 는
      `direct` 를 먼저 돌려주고 중복 판정은 **아카이브 안에서만** 센다(`:249-255`).
      아카이브 뒤 같은 id 로 디렉터리를 다시 만들면 조용히 그쪽이 이긴다(현재 0건).
- [ ] 3.3 5단계 실패 메시지가 "왜 이 함수가 요구되는가"를 말하도록 한다 — 오늘의
      메시지는 316개 함수 이름만 쏟아내고 그것이 남의 change 것이라는 사실을 말하지 않는다.
- [x] 3.4 착지 지점을 기록했는데 요구 집합이 비고 번들이 있으면 그 사실을 출력한다.
      조용한 통과는 증거로 통과한 것과 구분되지 않는다 — a074·a077·a079 가 정확히
      이 자리에 떨어진다(task 1.3 측정).

**RED·GREEN 실행 기록 (2026-09-09).**

RED 은 픽스처 둘로 갈랐다. `_merge_fixture` 는 실제 병합 모양(base P → 착지 L →
이웃 change N → 나중 리팩터 M)이고, `_clean_fixture` 는 **오늘 초록인** 자리
(R → P → L, 증거가 워킹트리와 일치)다. 위조 시험은 후자에서만 출발한다 — 전자에서는
착지와 무관한 이유로 이미 빨개서 판정을 재지 못한다. 첫 판본이 그 실수를 했고
셋이 엉뚱한 이유로 초록이었다. `test_the_fixture_passes_before_any_landing_record`
가 그 양성 대조군이다.

RED 결과 10개 중 8 FAIL · 2 ok(양성 대조군, 그리고 기록 없을 때의 오늘 동작 핀).
GREEN 뒤 10/10. `tools/logic-map` 129개 · `tools/sdd` 57개 · `make lint` 전부 통과.

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

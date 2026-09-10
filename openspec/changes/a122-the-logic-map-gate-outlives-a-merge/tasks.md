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
- [x] 1.8 빌린 증거(`function-logic-reference.txt`)의 착지 공유 규칙을 정한다. base 는
      정확히 같아야 한다는 규칙이 있는데(`check_analysis.py:614`) 착지에는 없다.

      (그 base 규칙은 편집 후 `check_analysis.py:785` 다 — 태스크가 적힌 시점의
      `:614` 에서 3.2.3.1 의 순서 변경과 이 편집으로 옮겨졌다.)

      **결정 (2026-09-10). 창은 양쪽 끝을 다 공유한다.** 빌리는 change 의
      `landed-commit.txt` 는 빌려주는 change 의 것과 정확히 같아야 하고, 한쪽만
      선언한 상태는 통과하지 않는다. 전체 근거는 review.md §Pre-Edit 1.8.

      **근거는 자유의 크기가 아니라 누가 고르는가다.** 3.2.3.1 의 고정("빌린 번들이
      착지를 고정한다")은 구간을 남긴다 — a072 실측으로 `revision: current` 번들
      **99개 파일**을 동시에 고정하는 커밋이 base..HEAD **326개 중 2개**이고, 그
      구간 안에서는 저자가 여전히 고른다. 선언 파일을 고른 근거는
      `analysis/landing-point.md` 의 한 문장뿐이다 — "번들이 고정하므로 저자가 고를 수
      없다". 빌리는 change 에서는 그 번들이 **남의 것**이라 그 문장이 끝까지 참이
      되지 않는다. 착지는 그것을 고정하는 증거가 사는 자리에 선언하고 빌리는 쪽은
      값을 **복사**한다.

      **거부할 정상 입력을 먼저 열거했고, 오늘 걸리는 것은 0건이다.** 저장소에
      빌리는 change 는 **1건**(a073 → a072), 착지를 선언한 change 는 **1건**(a099)
      이고 **두 집합이 안 겹친다**. a073·a072 는 둘 다 선언이 없어 새 규칙이 오류를
      하나도 더하지 않는다. 대신 a073 의 수리에 **순서**가 생긴다 — a072 가 먼저
      적어야 a073 이 같은 값을 적을 수 있다.

      **그 수리가 실제로 되는지 끝까지 돌려서 쟀다.** 오늘 a073 은 오류 **336개**
      (`missing evidence` 240 · `AST source hash is stale` 61 · `AST hash does not
      match` 35). 양쪽이 `171adda8` 을 선언한 상태를 선언을 **읽는 자리**에 주입해
      (그 뒤 판정 — 40-hex · 커밋 실재 · 조상 · 고정 · 번들 · 덮임 — 은 전부 진짜)
      `check` 를 부르면 **오류 0개**다. 336 → 0.

      | 검증 | 결과 |
      |---|---|
      | RED | 5건 중 **3건** 빨강(빌리는 쪽만 선언 · 창을 좁힘 · 빌려주는 쪽만 선언 — 셋 다 오늘 `[]` 로 **통과한다**). 나머지 2건은 **정상 입력 보호** 시험이라 오늘도 초록이어야 한다 |
      | 변이 M1 공유 판정 삭제 | **3건** FAIL |
      | 변이 M2 빈 선언을 "선언 없음"으로 | **1건** FAIL (이 편집이 만든 자리) |
      | 변이 M3 고정 0 거절 삭제 | **2건** FAIL (3.2.3.1 이 가려지지 않았다) |
      | 변이 M4 해시 대조 삭제 | **2건** FAIL — **빌린 쪽 위조 거절 포함** |
      | 변이 M5 한쪽만 선언한 창을 눈감기 | **2건** FAIL |
      | 실데이터 A/B | 실제 change id **126건** 옛·새 전수 — **rc·stdout·stderr 차이 0건** (`rc_diffs=0 out_diffs=0 stderr_diffs=0`, 글자 그대로 같다) |
      | 구조 | `check` 분기 37 → **40**, 반환 11 → **13**; `resolve_landing` 분기 20 → 16 · 반환 3 → 2 · raise 7 → 6 (**사라진 게 아니라 `_declared_landing` 으로 이동**, 보존 확인) |
      | 스위트 | `tools/logic-map` **145 OK** · `make sdd-test` 전부 OK · `openspec validate --strict` valid |

      **판정 둘이 서로를 가릴 자리를 편집 전에 찾았다.** 새 공유 판정은
      `resolve_landing` 앞에 선다. 그러면 3.2.3.1 회귀 시험
      (`test_borrowed_evidence_refuses_a_landing_it_does_not_describe`: 빌리는 쪽만
      바닥을 선언)이 공유 규칙에 먼저 걸려 고정 판정을 통째로 지워도 초록으로 남는다.
      그 시험의 픽스처를 "양쪽이 **같은** 바닥을 선언"으로 고쳐 고정 판정이 홀로
      판정하게 두었고, M4 가 그 뒤에도 2건을 빨갛게 하는지 **재서** 확인했다.

      **오류 문구는 형제 규칙과 같은 모양으로 뒀다** —
      `must share the exact landing point` 는 `must share the exact comparison base`
      의 짝이다. 값을 문구에 안 넣는 것도 그 짝을 따른다.

      기계 열거: `analysis/python-function-logic/tools-logic-map--check/ast.{before,after}-1.8.json` ·
      `tools-logic-map--resolve_landing/ast.after-1.8.json` ·
      `tools-logic-map--_declared_landing/ast.json` (**새 함수**).
      잔여는 5.5 에 적었다 — 빌리는 change 의 작업이 공유된 착지 **뒤에** 착지하면
      어떤 판정도 못 본다(빌리는 쪽은 그것을 고정할 증거를 소유하지 않는다).
- [x] 1.9 해소한 착지 SHA 와 요구 함수 수를 **항상** 출력한다. 5단계 기록에 어떤
      대상으로 통과했는지가 남아야 한다.
- [x] 1.10 착지 커밋에서의 change 디렉터리 경로를 정한다 — 아카이브·renumber 전의
      `openspec/changes/<id>/` 다. 경로가 아니라 내용으로 대조한다.
- [x] 1.11 `tools/gate.sh:116` · `:184` 가 활성 경로를 하드코딩해 아카이브된 id 가
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

      **결정 (2026-09-10): 규칙은 재사용하고 코드는 재사용하지 않는다.**
      `resolve_referenced_change` 는 Python 함수라 shell 에서 부를 수 없다. 그러므로
      해소 규칙(날짜 접두사를 벗긴 나머지 **전부** 일치 · 중복이면 fail-closed)만
      옮겨 적고, 정본이 둘이 되는 것을 [[transcribed-code-needs-both-sides-pinned]]
      대로 **양쪽 다** 시험으로 못 박는다. 한 지점에서 규칙이 갈리는데 고르지 않고
      적는다: 활성과 아카이브에 같은 id 가 동시에 있을 때 Python 은 `direct` 를 조용히
      고르고(3.2.4 가 연 결함) shell 은 **fail-closed** 한다. shell 을 Python 에 맞춰
      낮추지 않는다 — 3.2.4 가 Python 을 이쪽으로 올린다.
      구현은 2.7(RED) 과 3.4(GREEN) 로 내린다. 증거는
      `analysis/code-context/` 셋과 `review.md` 의 Pre-Edit Gate.
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
- [x] 2.7 gate 3단계가 **아카이브된 짝**을 넘어가는지 보는 테스트를 만든다. 픽스처는
      임시 저장소에 `tools/gate.sh` 를 복사해 세우고, 3단계까지만 관찰하도록 4단계
      (`review.md` 없음)에서 멈춘다. 네 축을 각각 고정한다 — (a) 활성 짝은 그대로
      통과, (b) 없는 오타 id 는 **계속 실패**, (c) 활성과 아카이브에 같은 id 가 있으면
      fail-closed, (d) 아카이브된 짝은 통과(지금은 여기서 죽는다). 더해서 gate 대상
      자신이 아카이브된 경우(`:116`)도 1단계에서 죽지 않는지 본다.

      `tools/sdd/test_gate_resolves_archived_changes.py` (8건). **첫 단언은 가짜였다** —
      픽스처에 `review.md` 가 없어 게이트는 어차피 4단계에서 죽으므로 거절 케이스의
      `returncode != 0` 은 해소기가 완전히 망가져도 초록이었다. 4단계에 **도달하지
      못했다**까지 보도록 고친 뒤에야 정직한 RED 3건이 나왔다(아카이브된 짝 ·
      활성이 아카이브를 가림 · gate 대상 자신이 아카이브).

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
- [x] 3.2.3.1 **고정할 증거가 없는 착지 선언을 거절한다 (2026-09-10).**
      `resolve_landing` 이 고정한 번들 수를 세고, 0 이면 선언을 거절한다. `check` 는
      빌린 증거를 착지 판정 **앞에서** 푼다. 설계(`analysis/landing-point.md`)가
      "착지가 base 와 같아도 안전하다 — **번들이 고정하므로**"로 세운 논거의 전제를
      코드에 세운 것이다. 전제가 없으면 그 문장은 공집합 위에서 참이 된다.

      **기계 열거가 기제를 가리켰다.** `analysis/python-function-logic/tools-logic-map--resolve_landing/`:
      B1~B10 은 전부 "이것이 어느 커밋인가"만 묻고, *고르지 못하게* 하는 것은
      `B11 L392` 순회 하나다. 그 순회가 0회 돌면 `mismatched` 가 빈 채 남고
      `B19 L403` 이 거짓이 되어 `L408 return candidate` 가 무조건 참이 된다.
      손으로 읽었으면 "거절이 여섯 개나 있다"에서 멈췄을 자리다.

      **거부할 정상 입력을 먼저 열거했고, 하나가 실제로 걸렸다.** 빌린 증거를 쓰는
      change 는 자기 번들이 0 이다(a073 이 a072 의 것을 빌린다). 지역 번들만 세는
      판정이면 **a073 의 유일한 수리 경로를 죽인다** — a073 은 오늘
      `AST source hash is stale` 로 빨갛고 그 빨강을 푸는 것이 착지 선언이다.
      실데이터로 쟀다: 착지 판정이 보는 고정 번들이 **옛 순서 0개 · 새 순서 209개**다.
      그래서 고정은 그 change 가 **실제로 딛는** 증거로 판정한다. 이 결정은 task 1.8
      (빌린 증거의 착지 **공유** 규칙)을 앞당겨 정하지 않는다 — "빌린 번들이 고정한다"는
      "착지가 같아야 한다"보다 약하다. 전체 표는 review.md §Pre-Edit 3.2.3.1.

      | 검증 | 결과 |
      |---|---|
      | RED | 3건 중 2건 빨강(번들 0 → `[]` 로 **통과했다**, 빌린 위조 → 다른 사유로만 빨감). 나머지 1건은 **정상 입력 보호** 시험이라 오늘도 초록이어야 한다 |
      | 변이 M1 고정 0 거절 제거 | **2건** FAIL |
      | 변이 M2 해소 순서 되돌리기 | **2건** FAIL — 그중 하나가 "정상 입력을 죽였다" |
      | 변이 M3 리비전 안 가리고 세기 | **1건** FAIL (`revision: base`) |
      | 변이 M4 해시 대조 삭제 | **2건** FAIL — 처음엔 1건이었다(아래) |
      | 실데이터 A/B | 실제 change id **126건** 옛·새 전수 — 출력 차이 **0건** |
      | 구조 | 반환 **3 → 3**(경로를 안 지운다), 분기 19 → 20, raise 6 → 7 |
      | 스위트 | `tools/logic-map` **135 OK** · `tools/sdd` 69 OK · `sdd-check` 0 · `validate` 58/58 |

      **판정 둘이 서로를 가려 주고 있었다.** M4 는 처음에 1건만 빨갛게 했다. 기존
      위조 시험(`test_a_landing_the_evidence_does_not_describe`)이 사유를
      `internal/own.go` 로 찾아서 `validate_target` 의 `AST source hash is stale` 도
      그 바늘을 만족했기 때문이다 — 착지 판정을 통째로 지워도 초록이었다.
      그 바늘을 착지 판정의 문장으로 좁혔고, 그 뒤 M4 는 2건을 빨갛게 한다
      ([[two-judgements-cover-for-each-other]]).

      **남은 것.** 번들이 **하나뿐**이고 그것이 change 가 만지지 않은 파일을 기술하면
      고정력이 약하다(그 파일을 안 건드린 모든 커밋이 통과한다). 이 태스크가 닫는 것은
      0 → 1 이고, "고정이 얼마나 좁히는가"는 아니다. §5 잔여에 적는다.
- [x] 3.2.4 **열린 디렉터리가 아카이브본을 가리던 것을 닫았다 (2026-09-10).**
      `resolve_referenced_change` 가 활성을 찾아도 아카이브를 마저 세고, 둘이면
      `AmbiguousChange` 로 멈춘다. 1.11 이 shell 을 fail-closed 로 두고 Python 을
      낮추지 않기로 한 갈림이 여기서 **Python 을 올려** 닫혔다.

      **좌표를 정정한다.** 위 문구의 `:249-255` 는 raise 두 줄의 대략 범위였다.
      기계 열거(`analysis/python-function-logic/`)가 준 정확한 자리는
      `B1 L239 if direct.is_dir():` → `L240 return direct`(early return)와
      `B6 L251 len(matches) > 1`(세는 대상 `matches` 는 아카이브 순회 산출)이다.

      **열거가 편집 범위를 하나에서 둘로 늘렸다.** CodeGraph 는 호출자를 `check`
      1건으로 답하는데, 그 안에 호출 자리가 **둘**이다 — `:690`(게이트 대상)과
      `:715`(빌린 증거). 앞쪽은 `except ValueError` 로 예외를 **삼키고**
      `openspec/changes/<id>` 로 되돌아간다. 해소기만 고쳤으면 게이트 대상 경로에서는
      여전히 활성이 조용히 이겼다. 그래서 호출 자리 하나를 같이 고쳤고, 가르는 근거는
      문구가 아니라 **타입**(`AmbiguousChange(ValueError)`)이다.

      | 검증 | 결과 |
      |---|---|
      | RED | 2건, 둘 다 `[]`(조용한 통과) → 기대 오류. **양성 대조**(충돌 전 통과)를 각 시험 안에 넣었다 |
      | 변이 M1 early return 복원 | 2건 FAIL |
      | 변이 M2 `except AmbiguousChange` 제거 | **1건** FAIL — 게이트 대상 시험만. 두 시험이 다른 것을 잰다는 증거 |
      | 변이 M3 중복을 2개 이상일 때만 | 2건 FAIL |
      | 원복 | 세 번 다 사본에서, sha256 대조 (`git checkout` 아님) |
      | 실데이터 A/B | 실제 id **126건**(활성 27 + 아카이브 99)을 옛·새 판본으로 각각 실행 — 출력 차이 **0건** |
      | 구조 확인 | 편집 후 열거에서 반환이 **2→1**(early return 소멸), `check` 의 `except AmbiguousChange` 가 `except ValueError` **앞** |
      | 스위트 | `tools/logic-map` 131 OK · `tools/sdd` 69 OK (shell 게이트 8건 포함) |

      **거부하게 될 정상 입력**(fail-closed 는 무엇을 죽이는지 말해야 한다):
      아카이브된 change 와 같은 id 로 활성 디렉터리를 다시 만드는 것 하나뿐.
      2026-09-10 측정으로 저장소에 0건(활성 27 · 아카이브 id 99 · 교집합 0 ·
      아카이브 내 중복 0). 대신 새 id 를 쓰는 것이 기존 renumber 관행이다.

      **아카이브 내 중복 메시지는 한 글자도 안 바꿨다** — 그 경우의 동작이 안 바뀌었기
      때문이다. 기존 시험 `test_real_archived_reference_rejects_two_copies` 가 그것을
      문구째로 잡고 있다.
- [x] 3.3 **실패 출력이 비교 창을 말한다 (2026-09-10).** 창 줄 둘을 `main` 이
      **실패 반환보다 앞에서** 찍고, 이름만 쏟아내던 메시지가 개수와 창을 담는다.

      **오늘의 출력을 먼저 쟀다** (HEAD `848b9ba3`): a074 는 324줄 중 316줄이
      `missing evidence for modified function` 이고 창을 말하는 줄이 **0**,
      a076 은 이름 316개를 쉼표로 이은 **21,838자짜리 한 줄**에 역시 **0** 이다.

      **기제는 early return 이었다.** 기계 열거(`analysis/python-function-logic/tools-logic-map--main/`)
      가 `main` HEAD = 분기 4 · 반환 2 이고 `B1 L824 if errors:` → `L827 return 1` 이
      `B4 L835 if landing:` 을 건너뛴다고 말한다 — 착지·요구 수를 찍는 자리가
      **실패 경로에서 도달 불가**다. 3.2.4 와 같은 모양이다. 그래서 task 1.9 가
      "항상 출력한다"고 적어 둔 것이 성공할 때만 참이었다. 이 태스크가 코드를
      그 기록 쪽으로 올려 그 갈림을 닫는다.

      새 출력(실측):

      ```
      [logic-map] a074-…: base 448dfeb1263d → working tree (no landed-commit.txt) required 319 function(s)
      [logic-map] a074-…: the target is the working tree, so this window also holds 283 commit(s)
                  that landed after the base and every existing function they changed is required
                  here too — record `landed-commit.txt` to narrow it to this change's own work
      ```

      **283** 이 이 태스크가 주는 숫자다. 319개 요구가 커밋 283개짜리 창에서 나왔다는
      말이 이제 출력에 있다.

      | 검증 | 결과 |
      |---|---|
      | RED | 4건 중 3건 빨강. 4번째는 **성공 줄 회귀 핀**이라 오늘도 초록이어야 한다(기록 스무 곳이 그 문자열을 인용한다) |
      | 변이 N1 창 출력을 실패 뒤로 | **2건** FAIL |
      | 변이 N2 창 크기 줄 삭제 | 1건 FAIL |
      | 변이 N3 메시지를 옛 문구로 | 1건 FAIL |
      | 변이 N4 커밋 수를 상수 `"3"` 으로 | 1건 FAIL — 첫 시험은 이것을 통과시켰다(아래) |
      | 실데이터 | 실제 change id **126건** 옛·새 전수: **rc 차이 0 · 판정 본문 차이 0 · stderr 차이 0**. 창 줄은 **116건**에 찍힌다 |
      | 구조 | 반환 **2 → 2**(경로를 안 더한다), 분기 4 → 6, `print` 가 `return 1` 앞뒤로 갈린다 |
      | 스위트 | `tools/logic-map` **139 OK** · `make sdd-test` 전부 OK · `sdd-check` 0 · `validate` 58/58 |

      **N4 가 시험 하나를 고치게 했다.** 픽스처의 정답이 마침 3 이라 `return "3"` 이
      단언을 통과했다 — 재서 쓴 값인지 상수인지 가르지 못한 것이다. 커밋을 하나 더
      얹어 **4가 되는지**까지 보게 고쳤다([[generated-evidence-must-be-measured]]).

      **불변식이 3.2.3.1 과 다르다.** 그쪽은 출력 차이 0 을 쟀지만 여기서는 출력이
      반드시 달라진다. 그래서 잰 것은 **판정 불변**이다 — rc 와, 창 줄·메시지 머리를
      정규화한 나머지가 126건 전부 글자 그대로 같다.

      **창 줄이 없는 10건의 정체를 셌다.** 116/126 의 나머지가 어디 갔는지 추측하지
      않고 갈랐다: `base-commit.txt` 가 아예 없는 **9건**(a119 · verify-execution-capability ·
      2026-07-26~27 아카이브 일곱)과 a063 **1건**이다. a063 은 파일은 있지만
      `resolve_base` 안의 실행 기준선 이관 검사가 `adoption requires detached HEAD` 로
      먼저 던져서 `effective_base` 가 채워지기 전에 빠져나간다. 즉 창 줄은 **비교 기준을
      해소한 change 에만** 찍힌다 — 해소 못 한 창을 지어내지 않는다(`if base:`).
      9 + 1 = 10 으로 116 과 맞아떨어진다.

      **안 바꾼 것.** `missing evidence for modified function …` 줄에는 접미사를 안
      붙였다. a074 에서 316번 반복될 자리라 창 줄이 한 번 말하는 편이 낫다. 성공 줄
      (`evidence complete or diff-proven exempt`)도 그대로 뒀다 — 회귀 핀 시험이 그것을
      고정한다. 이름 목록 자체도 자르지 않았다(증거를 줄이는 것은 별도 결정이다).
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
- [x] 3.4 `tools/gate.sh` 에 change 디렉터리 해소기를 하나 만들고 `CHANGE_DIR`(`:116`)
      과 `PAIR_DIR`(`:184`) 을 그것으로 바꾼다. 해소 순서는 활성 → 아카이브이며 둘 다
      있으면 멈춘다. 기존 검사 넷(id 형태, 자기 자신 선언 거부, 면제 줄 하나, 구성원
      집합 일치)은 손대지 않는다.

      **변이 검증** — 셋 다 정확한 테스트가 잡았다. 중복 판정 제거 → 활성/아카이브
      중복과 아카이브 내 중복 2건 FAIL. 전부 일치를 접미사 일치로 → 접미사 테스트 FAIL.
      아카이브 순회 제거 → 3건 FAIL. GREEN 원복은 `git checkout` 이 아니라 사본에서
      했고 sha256 으로 확인했다([[mutation-revert-needs-the-right-baseline]]).

      **실제 데이터 A/B** — 저장소의 실제
      `archive/2026-08-29-a098-nobody-sends-what-the-outbox-keeps` 와 a099 의 실제
      `deploy-pair.txt` 로 세운 픽스처에서, 새 gate 는
      `OK: a098-… — 완료 + 구성원 일치` 로 4단계까지 가고 옛 gate(`29c609bd`)는
      같은 픽스처에서 3단계에 죽는다.

      **회귀 전수** — 활성 28 · 아카이브 98 id 를 훑어 활성과 아카이브에 같은 id 가
      있는 경우 **0건**, 아카이브 내 중복 **0건**. 즉 더 엄격해진 규칙이 오늘 죽이는
      정상 입력은 없다. `deploy-pair.txt` 를 가진 활성 change 는 a099 하나뿐이다.

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
- [ ] 5.3 **고정의 세기는 이 change 가 답하지 않는다 (3.2.3.1 에서 열림).**
      3.2.3.1 은 고정 번들 수를 0 → 1 이상으로 요구할 뿐이다. 번들이 하나뿐이고 그것이
      이 change 가 만지지 않은 파일을 기술하면, 그 파일을 안 건드린 **모든** 커밋이
      고정을 통과한다. `analysis/landing-point.md` 의 a112 측정(132 번들 → 위조 후보에서
      43 불일치)이 강했던 것은 번들이 많아서다. "몇 개면 충분한가"는 숫자를 지어내지
      않고 두며, 정하려면 실데이터로 재야 한다.
- [ ] 5.4 **`revision: base` 번들만 있는 change 는 착지를 고정할 수 없다.**
      그 번들의 source hash 는 base 를 기술하고 `validate_target` 도 그것을 해싱하지
      않는다. 3.2.3.1 은 그런 change 의 착지 선언을 거절한다 — 저장소에 오늘 **0건**
      (번들>0 인 change 는 전부 고정 번들>0). 생기면 그때 무엇으로 고정할지 정한다.
- [ ] 5.5 **빌린 증거를 쓰는 change 의 자기 작업은 고정되지 않는다 (1.8 에서 열림).**
      1.8 은 창의 양쪽 끝을 빌려주는 change 와 공유하게 만든다. 그래도 빌리는 change 의
      작업이 그 공유된 착지 **뒤에** 착지하면 그 작업은 요구 집합 밖이고 어떤 판정도
      못 본다. 빌리는 change 는 정의상 자기 번들이 0 이라(공존 금지) 그것을 고정할
      증거를 **소유하지 않는다** — 공유 규칙이든 3.2.3.1 의 고정 규칙이든 이 잔여는
      같고 1.8 이 만드는 것이 아니다. 닫으려면 "빌림이 유효한 조건"(빌리는 쪽의 수정
      함수가 빌려주는 쪽 지도 안에 있다)을 신원이 아닌 증거로 물을 방법이 필요하다.

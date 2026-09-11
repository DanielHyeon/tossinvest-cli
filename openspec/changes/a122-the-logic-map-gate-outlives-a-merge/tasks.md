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
- [x] 1.12 a063 의 `execution_baseline.validate` 가 착지 기록을 감사하지 않는다.
      키 집합 열거(`execution_baseline.py:410-412`)에 넣을지 정한다.

      **결정 (2026-09-11). 키를 넣지 않는다 — 창의 끝은 이미 그 record 안에 있다.**
      이름이 `source_commit` 이고, a063 의 실물 record 는 오늘 14개 키에
      `source_commit = c727ad12` 를 담고 있다. 같은 값을 두 번째 이름으로 또 적으면
      재진술이고, 둘이 갈리면 어느 쪽이 계약인지 아무도 모른다. 게다가 키 집합은
      닫혀 있어(`set(record) != required` 면 거절) 키를 하나 넣으면 a063 의 커밋된
      record 와 ledger digest 를 다시 만들어야 한다.

      **대신 진짜 구멍 둘을 닫는다.** 전체 근거는 review.md §Pre-Edit 1.12.

      (1) **이관 경로의 대상이 감사된 값이 아니라 워킹트리였다.** `validate` 는
      `{"effective_base", "source", "ledger"}` 를 돌려주는데 `resolve_base` 는
      `effective_base` 만 읽는다 — 저장소 전수로 `adoption["source"]` 를 읽는 곳이
      **0곳**이었다. 그래서 `resolve_landing` 이 `landed-commit.txt` 를 찾고, 없으니
      대상이 워킹트리가 된다. 그 답이 오늘 맞는 이유는 `validate` 의 **다른** 판정
      (`source-to-evidence drift`) 때문이다 — 맞는 답을 우연으로 얻고 있었다.

      | a063 의 대상 | required |
      |---|---|
      | 감사된 `source_commit` c727ad12 | **9** |
      | 워킹트리 (2026-09-11 HEAD) | **36** |

      (`c727ad12..HEAD` = 커밋 32 · 파일 579, 그중 **.go 24**. 오늘 이관을 다시 돌리면
      drift 검사가 먼저 죽이므로 fail-closed 지만, 창의 계산은 남의 함수 27개를
      a063 의 것으로 센다.)

      (2) **3.3 이 넣은 안내가 이관 감사와 모순됐다.** 이관 픽스처로 5단계를 돌리면
      `… record \`landed-commit.txt\` to narrow it …` 을 찍는다 — 감사 어디에도 없는
      두 번째 손잡이를 만들라는 말이다. 그 파일은 `openspec/` 아래라 drift 검사가
      통과시키고, 추적 파일이라 untracked 감사도 못 보고, 닫힌 키 집합에도 없는데
      비교 대상을 고른다. 이관 경로는 그 기록을 **거절**하고, 대상을 감사된
      `source_commit` 으로 말한다.

      **거부할 정상 입력을 먼저 열거했고, 오늘 새로 거절하는 change 는 0건이다** —
      a063 에 `landed-commit.txt` 가 없고, 이관이 아닌 change 는 이 경로에 안 들어온다.
      판정이 바뀌는 change 도 0건이고, 바뀌는 것은 이관 실행의 **출력 문구**다.

      | 검증 | 결과 |
      |---|---|
      | RED | 3건 전부 빨강 (대상이 `''` · 안내가 `landed-commit.txt` 를 찍음 · 착지 기록이 통과) |
      | 변이 N1 이관의 착지 기록 거절 삭제 | **1건** FAIL |
      | 변이 N2 감사된 source 대신 `resolve_landing` 호출 | **2건** FAIL |
      | 변이 N3 대상 문구가 감사 여부를 안 가림 | **1건** FAIL |
      | 변이 N4 번들 경로 정규화 제거 | **1건** FAIL (아래 곁가지) |
      | 변이 N5 문맥이 없으면 이관을 못 봄 | **1건** FAIL |
      | 변이 N6 감사된 source 를 문맥에 안 적음 | **2건** FAIL |
      | 실데이터 A/B | 실제 change id **126건** 옛·새 전수 — **rc·stdout·stderr 차이 0건**. a063 은 옛·새 모두 rc=1 (`adoption requires detached HEAD`) — 이관 경로는 그 밖에서 도달하지 않으므로 재는 것은 픽스처 시험이다 |
      | 구조 | `check` 분기 40 → **43** · 반환 13 → **14**; `resolve_base` 분기 10 → **11** (반환·raise 불변); `resolve_landing` **16/2/6 그대로** |
      | 스위트 | `tools/logic-map` **149 OK** · `openspec validate --strict` valid |

      **`execution_baseline.py` 는 한 줄도 안 바꿨다.** 사람이 승인한 예외의 감사
      코드를 건드리지 않고, 그 감사가 **이미 돌려주는 값을 버리지 않게** 했을 뿐이다.
      이관 경로는 `resolve_landing` 을 부르지 않는다 — 그 판정들은 저자가 고른 값을
      위한 것이고, `validate` 가 `ancestry(P,E)` · `ancestry(E,source)` ·
      `ancestry(source,head,strict=True)` · tree 대조 · digest 셋으로 이미 묶는다.

      **곁가지로 3.2.3.1 의 결함 하나를 같이 고쳤다.** 고정 순회가 번들의 `file` 을
      정규화하지 않아 절대경로면 `git show <sha>:/abs/path` 가 언제나 실패한다 —
      **정상 입력이 위조로 몰린다**. 이관 픽스처의 번들이 절대경로를 쓰면서 드러났다.
      저장소 전수로 번들 `file` 은 **상대 3048 · 절대 0** 이라 실물 영향은 0 이고
      A/B 로는 영원히 안 보인다. 그래서 변이 N4 와
      `test_an_absolute_bundle_path_still_pins_a_landing` 이 이것을 재는 전부다.
      계약은 안 바뀐다 — spec 이 이미 요구하던 것을 코드가 못 지키고 있었다.

      기계 열거: `analysis/python-function-logic/tools-logic-map--check/ast.after-1.12.json` ·
      `tools-logic-map--resolve_base/ast.{before,after}-1.12.json` (**새 대상**) ·
      `tools-logic-map--resolve_landing/ast.after-1.12.json`.
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

- [x] 4.1 아카이브된 a075·a076 을 회귀 픽스처로 쓴다. 착지 지점을 주면 통과하고
      주지 않으면 오늘과 같이 실패해야 한다 — 이 change 가 실제로 그 다섯을 푸는지의 증거다.

      **선결 조건은 이미 닫혔다.** "아카이브된 change 는 id 로 재검사가 안 된다"는
      2026-09-08 실측이었고 task 3.2.4(`52c761de`)가 `check` 의 change 해소를
      `resolve_referenced_change` 로 바꾸면서 풀렸다. 오늘 두 change 모두 아카이브
      base `448dfeb1` 을 해소한다.

      **측정은 주입이 아니라 격리 worktree 에 실제로 커밋해서 했다**
      (`git worktree add --detach`, 커밋 `06b280e0`, 착지 `840b3377`). 선언을 읽는
      자리(HEAD 커밋에서 읽기·40-hex)까지 전부 진짜 경로다. 측정 뒤 worktree 는
      정리했고 아카이브본에는 아무것도 안 썼다.

      | | a075 착지 없음 | a075 착지 `840b3377` |
      |---|---|---|
      | required | **319** | **0** |
      | 오류 | **324** (`missing evidence` 318 · `stale` 4 · `hash does not match` 1 · 인용 테스트 1) | **1** |

      323개가 사라진다. 남은 1은 `cmd-tossctl--runconsole` 의 branch-test-map 이
      트리에 없는 테스트를 인용하는 것이고, task 1.13 이 "a075 의 결함이고 a122 가
      덮지 않는다"고 이미 적어 둔 것이다.

      **a076 은 해결되지 않는다 — 그리고 그 거절이 옳다.** a076 은 번들이 **0개**다
      (다섯 중 유일하다: a074 7 · a075 8 · a076 **0** · a077 7 · a079 4). 착지를
      선언하면 3.2.3.1 이 `pinned by no revision: current evidence` 로 거절한다.
      그 규칙이 막는 위조 경로가 정확히 이 모양이기 때문이다 — 번들 0 + 면제 표식 +
      구간 바닥 착지. 저장소는 "증거가 없어서 0"과 "고친 Go 가 없어서 0"을 가르지
      못한다. **proposal 이 다섯에 답할 수 있는 질문을 준다고 적은 것은 넷에 대해
      참이다.** 잔여 5.6 에 연다.

      **곁가지로 3.3 의 결함 하나가 실물로 나왔다 — 못 잰 창을 찍는다.** a076 에
      착지를 준 실행이 `base … → working tree (no landed-commit.txt) required 0
      function(s)` 를 찍었다. **거짓이다** — 대상은 워킹트리가 아니고(착지가 선언돼
      있다) 요구 수는 세지도 않았다(해소가 실패해 `landing`·`required_count` 가 아예
      안 채워지고 기본값이 찍혔다). 게다가 다음 줄이 이미 있는 파일을 만들라고 한다.
      3.3 이 창을 찍게 만든 이유가 "이름만 있고 이유가 없다"였는데 **지어낸 이유는
      그보다 나쁘다.** 창은 이제 **잰 것만** 찍고 사유는 `cannot derive …` 줄이 말한다.

      | 검증 | 결과 |
      |---|---|
      | a075 회귀 | 착지 없음 rc=1 오류 324 → 착지 있음 rc=1 오류 **1**(a122 밖 결함) |
      | a076 회귀 | 착지 없음 오류 1 → 착지 있음 **거절**(`pinned by no`) |
      | RED (창 거짓말) | 1건 빨강 — 유닛 픽스처가 a076 의 줄을 그대로 재현 |
      | 변이 O1 못 잰 창도 찍기(되돌리기) | **1건** FAIL — 새 시험 하나 |
      | 변이 O2 창 줄 통째로 삭제 | **4건** FAIL — 3.3 의 창 시험 둘 · 빈 요구 집합 · 1.12 이관 창. **O1 과 겹치지 않는다**(조건과 줄이 각각 묶였다) |
      | 실데이터 A/B | 실제 change id **126건** 옛·새 전수 — **rc·stdout·stderr 차이 0건** (오늘 이 줄을 보는 change 가 0건이므로 예상대로다) |
      | 스위트 | `tools/logic-map` **150 OK** |
- [x] 4.2 a074 · a077 · a079 에 대해 실행해 요구되는 함수 집합이 각 change 가 실제로
      고친 것으로 줄어드는지 확인하고 그 수를 기록한다.

      **셋 다 창을 되찾는다. 다만 이 task 의 전제는 틀렸다** — 요구 집합은 "각
      change 가 실제로 고친 것"으로 줄지 않고 **0 으로** 줄어든다. 그 0 은 산술이고
      아래에 근거를 적는다.

      측정은 4.1 과 같은 방법이다: `git worktree add --detach` 로 격리 worktree 를
      세우고 세 change 에 `landed-commit.txt` 를 **실제로 커밋**해서 잰다(커밋
      `886ab949`). 측정 뒤 worktree 는 제거했고 **세 change 디렉터리에는 아무것도
      쓰지 않았다**(아래 "선언은 아직 하지 않는다").

      | | 착지 없음 | 착지 `840b3377` |
      |---|---|---|
      | a074 | required **319** · rc=1 · 오류 **325** | required **0** · rc=0 · 오류 **0** |
      | a077 | required **319** · rc=1 · 오류 **322** | required **0** · rc=0 · 오류 **0** |
      | a079 | required **319** · rc=1 · 오류 **323** | required **0** · rc=0 · 오류 **0** |

      셋 다 base `448dfeb1`(2026-08-04 `origin/main` 병합)을 공유하고 셋의 착지가
      한 점 `840b3377` 로 모인다 — 그 커밋이 `chore(sdd): rebaseline a074-a079 after
      strategy merge` 다.

      **사라지는 것은 두 종류다.**

      | 오류 | a074 | a077 | a079 | 무엇인가 |
      |---|---|---|---|---|
      | `missing evidence for modified function` | 316 | 318 | 317 | **남의 함수**. 287 커밋이 창에 들어와 있었다 |
      | `AST source hash is stale` | 5 | 2 | 3 | **자기 증거**가 오늘의 트리와 안 맞는다 |
      | `AST hash does not match modified function revision` | 3 | 1 | 2 | 같음 |
      | 3.3 의 조언 줄(`record landed-commit.txt …`) | 1 | 1 | 1 | 오류가 아니라 안내 |

      두 번째 종류가 중요하다. stale 이 이름으로 부르는 파일은 정확히 그 change **자기**
      번들의 소스다(a074 `exitloop.go`·`engine.go`, a077 `portfolio_pages.go`, a079
      `position_policy_transport.go`·`position_policy.go`·`console.go`). 대상이
      워킹트리이면 증거 대조도 워킹트리에서 하므로, 남이 그 파일을 고치는 순간 자기
      증거가 썩는다. 착지를 선언하면 대조도 착지에서 하므로 같이 사라진다 — 이것이
      1.9 가 `validate_target(revision_ref=…)` 로 닫은 절반이고, 여기서 실물로 확인된다.

      **왜 required 가 0 인가 — 측정이다.** base 와 착지 사이 커밋은 **1개**이고 그
      커밋이 Go 파일을 **0개** 바꾼다. 그리고 세 change 의 `revision: current` 번들
      소스 **12/12 파일이 base 자체에서 이미 hash 일치**한다 — 셋의 Go 작업은 자기
      base **앞**에 착지했다. 2026-08-04 재기준화가 base 를 작업 **뒤**로 옮긴 것이다.
      spec 의 "착지 지점이 base 뒤에 있어 요구 집합이 비는 change" 시나리오가 정확히
      이 모양이고, 5단계는 요구 0 을 **출력**하면서 번들 자체는 계속 검사한다.

      **통과는 증거가 아니므로 변이로 확인했다.** rc=0 이 "검사를 안 했다"가 아님을
      가르는 것은 이것뿐이다.

      | 변이 | 결과 |
      |---|---|
      | P1 `exitobserver.run` 의 `source_sha256` 을 0 으로 (a074) | **거절** — `is not the revision this evidence describes: internal/app/engine/exitloop.go` |
      | P1' 같은 변이 (a077 `joinPositions`) | **거절** — `internal/console/portfolio.go` |
      | P1'' 같은 변이 (a079 `Console.routes`) | **거절** — `internal/console/console.go` |
      | P2 착지를 고정 구간 **밖**(worktree HEAD)으로 | **거절** — 두 파일을 이름으로 부른다 |
      | P3 번들 디렉터리 전체 삭제, 착지는 유지 | **거절** — `pinned by no revision: current evidence` (= a076 과 같은 문) |

      **고정은 구간을 남긴다 — 이번엔 그 구간이 결과를 바꾼다.** 1.8 은 "고정이 값을
      하나로 만들지 않고 구간을 남긴다"고 적었지만 당시 실측(a072 2/326, 둘 다
      required 147)에서 구간의 효과는 **0** 이었다. a074 는 구간이 **14/287** 이고
      그 안에서 결과가 갈린다.

      | a074 착지 후보 (14개 중) | required | rc |
      |---|---|---|
      | `840b3377` · `15d25f80` | **0** | 0 |
      | `df4407ed` · `359b1fe7` · `30d8bb93` | 4 | 1 |
      | `aaa7638d` … `f1aae509` (4개) | 9 | 1 |
      | `56e85c68` | 14 | 1 |
      | `c58b66c9` · `3dd077ae` · `53626032` | 17 | 1 |
      | `8dba0173` | **30** | 1 |

      **0 에서 30 까지, rc 는 0 과 1 사이에서 갈린다.** 저자가 고를 수 있다는 1.8 의
      주장은 이제 측정치를 가진다. 다만 이번 경우 통과하는 선택(구간의 **바닥**)이
      **옳은 선택과 같다** — 위로 갈수록 늘어나는 것이 a080·a081·a082·a083 의 함수,
      즉 a122 가 없애려는 바로 그 남의 작업이기 때문이다. 구간 폭 실측은 5.3 에
      옮겨 적는다.

      **선언은 아직 하지 않는다.** 셋 다 배포 후 실측 task 가 열려 있다(5.1). 지금
      착지를 선언하면 창이 그 자리에서 얼고, 그 실측이 Go 수정을 부르면 그 수정이
      창 **밖**으로 나가 어떤 판정도 못 본다 — 5.5 가 빌린 증거에 대해 적은 것과 같은
      모양이 빌리지 않은 change 에서도 생긴다. 선언은 셋이 각자 Go 작업을 끝냈다고
      판단할 때 그 change 가 한다. a122 는 그 선언이 **작동한다는 것**까지만 잰다.

      **회귀 전수** — 활성 change 중 `revision: current` 번들을 가진 것은 13건이고,
      그중 **8건**이 번들 소스가 자기 base 에서 이미 전부 일치한다(a074·a077·a079·
      a089·a091·a092·a094·a095). 즉 base 가 작업 뒤에 놓인 것은 이 셋만의 사정이
      아니다. 잔여 5.7 에 연다.
- [x] 4.3 focused 테스트와 `make test` · `make vet` · `make validate` · `make sdd-sync` ·
      `make sdd-check` 를 돌린다.

      | 게이트 | 결과 |
      |---|---|
      | focused `test_check_analysis.py` | **75 OK**(skipped 1) — a122 가 **31** 추가해 44 → 75 |
      | `tools/logic-map` python 전체 | **150 OK**(skipped 1) |
      | `make test` | rc=0 · 99 패키지 · FAIL 0 |
      | `go test -count=1 ./...` (무캐시 강제) | rc=0 · 99 패키지 · **cached 0** · FAIL 0 |
      | `make vet` | rc=0 |
      | `make lint` | rc=0 — 무태그 + `tossos_testseams` 태그 vet 둘 다 |
      | `make test-seams` | rc=0 · 100 패키지 · FAIL 0 |
      | `make test-race` | rc=0 · 8 패키지 · **DATA RACE 0** |
      | `make validate` | rc=0 · **58/58** |
      | `openspec validate --strict` 활성 전수 | **27/27** valid |
      | `make sdd-sync` | rc=0 · `all indexes current` |
      | `make sdd-check` | rc=0 · `CodeGraph hard-evidence index matches the worktree` |

      **`make test` 가 다수 `(cached)` 로 찍혀 무캐시로 다시 돌렸다.** 캐시된 결과는
      Go 가 입력 키로 검증하므로 틀린 값은 아니지만, VERIFY task 에서 "아까 봤다"와
      "지금 봤다"는 다른 문장이다 — [[missing-tool-reports-clean]] 의 "빈 출력·0 은
      위반 0 이 아니라 검사 0". `-count=1` 에서 cached 줄이 **0개**임을 확인했다.

      `make test-race` 는 4.3 의 목록에 없지만 돌렸다. a122 는 Go 를 0줄 바꾸므로
      면제 사유가 성립하지만(**not-applicable: Go 변경 0**), 사유를 적는 것보다 재는
      것이 싸서 쟀다. 8 패키지 전부 초록, DATA RACE 0.

      ### 실데이터 전수 — change id 126건

      활성 27 + 아카이브 99 에서 중복을 뺀 **126개 id** 에 대해 현재 도구를 돌렸다.

      | | 건수 |
      |---|---|
      | rc=2 (**도구 붕괴**) | **0** |
      | rc=0 (통과) | **2** — a099(착지를 선언했고 유효) · a122(Go 0줄) |
      | rc=1 (증거 부족) | 124 |

      124 중 **10건**이 한 줄로 끝난다. 그 사유를 전부 확인했다: **9건은 그 change 에
      `base-commit.txt` 가 아예 없다**(2026-07-26~27 의 관례 이전 아카이브 7건 +
      a119 · verify-execution-capability). 나머지 1건은 a063 으로,
      `adoption requires detached HEAD` 다 — 이관 경로가 전용 detached worktree 에서만
      돌게 되어 있어서이고, **a122 이전 도구도 글자 그대로 같은 줄을 찍는다**(확인함).

      **a122 는 저장소를 초록으로 만들지 않는다.** 124건의 rc=1 은 5단계 강제가
      아직 꺼져 있는 동안 쌓인 기존 부채이고, a122 가 주는 것은 각 change 가 자기
      창을 되찾을 **손잡이**다. 그 손잡이를 쓴 change 는 오늘 a099 하나다.

      ### 아카이브 재검사 A/B — 옛 도구 vs 지금 도구

      spec 의 "아카이브된 change 의 재검사" 요구가 실제로 닫혔는지 옛 도구
      (`a2d11fb2`, a122 가 `check_analysis.py` 를 만지기 직전)와 대조했다.

      | id | 옛 도구 | 지금 도구 |
      |---|---|---|
      | a075 (아카이브) | `missing base-commit.txt` | `base 448dfeb1 → working tree … required 319` |
      | a120 (아카이브) | `missing base-commit.txt` | `base e65e394b → … required 36` |
      | a073 (아카이브) | `missing base-commit.txt` | 창 줄 정상 |

      **계측기부터 검증했다.** 처음엔 옛 도구를 scratchpad 에 복사해 돌렸고 셋 다
      `ModuleNotFoundError: No module named 'role_check'` 로 죽었다 — 형제 모듈이
      import 되지 않은 **내 설정 오류**였지 옛 도구의 성질이 아니다. 도구를
      `tools/logic-map/` 안에 두고 **활성 id 로 양성 대조군**을 먼저 돌려(옛 도구가
      진짜 출력을 낸다) 계측기가 눈멀지 않았음을 확인한 뒤에 위 표를 쟀다 —
      [[mutation-must-reach-the-thing-under-test]].

      ### 1.12 이관 경로는 실데이터로 못 돌렸다 (not-applicable 아님, **차단**)

      a063 의 이관 경로를 실물로 돌리려 했으나 세 지점 모두 막힌다.

      | 시도 | 결과 |
      |---|---|
      | 메인 브랜치에서 | `adoption requires detached HEAD` |
      | a063 전용 worktree(`/tmp/tossos-a063-execution-adoption`) | `untracked/ignored input is not allowed: .codex-context/.save-session.lock` |
      | HEAD 에 새로 판 detached worktree | `required commit ancestry is absent` |

      마지막은 원인이 분명하다: a063 이 감사한 `source_commit c727ad12` 는 **HEAD 의
      조상이 아니고**(다른 갈래) 그 커밋에는 이관 기록 자체가 아직 없다. 남은 길은
      a063 의 worktree 를 청소하는 것인데 **그것은 a063 의 상태이지 a122 의 것이
      아니다** — 남의 change 작업 상태를 허락 없이 건드리지 않는다.

      **세 지점 전부에서 옛 도구와 지금 도구가 글자까지 같은 줄을 찍는다.** 즉 a122
      가 닿을 수 있는 이관 경로에 **관측 가능한 차이를 만들지 않았다**. 1.12 의 근거는
      유닛 픽스처 3건(1.12 표)과 이 동일성이고, **실데이터 확인은 아니다.** 이것을
      4.4 독립 리뷰의 입력으로 명시한다.
- [x] 4.4 독립 적대 diff/테스트 리뷰와 gstack 리뷰를 마친다.

      독립 원천 **다섯**으로 돌렸다 — gstack `/review` 전문가 4(Testing · Security ·
      Maintainability · Simplification+Performance, 각자 빈 문맥) · **codex
      `gpt-6-astra`**(외부 모델, read-only) · 이 세션이 직접 돌린 **뮤테이션 9개**.
      대상 HEAD `0cca39cb`, base `1687baac`, 리뷰한 실행 코드는 `tools/` 5개 파일
      (+1502/-23) 이다. 기록은 review.md `## VERIFY — task 4.4`.

      **리뷰의 결론은 "이대로 4.5 로 못 간다" 이다.** 막는 것은 하나다 — spec 이
      SHALL NOT 으로 금지한 "위조 가능한 착지 기록"이 실제로 위조된다. 서로 모르는
      세 원천이 각각 재현했고 이 세션이 **RED → GREEN 대조군**으로 못 박았다:
      증거를 base 상태로 써서 `landed-commit = base` 를 선언하면, a122 **이전에는**
      `AST source hash is stale` 로 빨갛던 같은 입력이 **이후에는** `[]` · required
      **0** 으로 통과한다. a122 는 못 막은 것이 아니라 **없던 길을 연다.**

      이것은 구현 버그가 아니라 **spec 의 틈**이다. 코드는 spec 이 적은 판정("모든
      `revision: current` 번들의 source hash 가 착지에서 일치")을 정확히 구현한다.
      그 판정이 같은 문단의 목표("위조 가능해서는 안 된다")에 못 미친다.

      뮤테이션 9개 중 **6개는 CAUGHT**(M1 빈 고정 · M5 해시 불일치 · M6 착지 해싱 ·
      M7 이관이 선언 거절 · M8 빌림 양끝 · M9 활성+아카이브). **3개는 SURVIVED** —
      `base ≤ landing` · `landing ≤ HEAD` · 40-hex 형태. 원복은 sha256 동일성으로
      확인했다.

      **생산 코드는 이 리뷰에서 한 줄도 바꾸지 않았다.** 리뷰가 잰 HEAD 와 기록이
      가리키는 HEAD 가 같아야 재현되기 때문이다. 나온 것은 전부 §6 으로 연다.
- [ ] 4.5 PM 동기화 후 `make gate CHANGE=a122-the-logic-map-gate-outlives-a-merge` 를
      돌리고 성공한 뒤에만 아카이브한다.

      **4.4 가 이 task 를 막았다.** §6 의 P0 둘이 닫히기 전에는 돌리지 않는다 —
      아카이브하는 순간 spec 본문이 저장소의 상설 규칙이 되고, 그 본문이 지금
      위조를 허용하는 판정을 SHALL 로 적고 있다.

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

      **4.2 가 구간 폭을 실측했다.** 고정이 남기는 후보 커밋 수는 이렇다 — a072
      **2**/326 · a075 **2**/286 · a077 **2**/287 · a079 **2**/287 · a074 **14**/287 ·
      a091 **26**/262. 그리고 폭이 결과를 바꾼다: a074 의 14 안에서 required 가
      **0 에서 30 까지** 가고 rc 가 0 과 1 사이에서 갈린다(4.2 의 표). 그러니 이
      잔여의 질문은 "번들이 몇 개면 충분한가"만이 아니라 **"구간이 몇 커밋이면
      좁은가"** 이고, 후자가 저자에게 실제로 남는 선택지다. 폭은 번들 수가 아니라
      번들이 기술하는 **파일들이 얼마나 자주 바뀌는가**의 함수다 — a091 은 번들이
      2 개인데 구간이 26 이고, a077 은 번들 7 개에 구간이 2 다.
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
- [ ] 5.6 **번들이 0 인 change 는 창을 좁힐 수 없다 (4.1 에서 실물로 나왔다).**
      a076 은 자기 창에서 기존 Go 함수를 하나도 안 고쳤지만(2026-08-04 재기준화가
      base 를 그 작업 뒤로 옮겼다) 그것을 증명할 증거를 소유하지 않는다. 착지를
      선언하면 3.2.3.1 이 거절하고, 그 거절은 옳다 — 저장소는 "증거가 없어서 0"과
      "고친 Go 가 없어서 0"을 가르지 못하고, 후자를 허용하면 전자가 같은 문으로
      들어온다. 닫으려면 "요구 집합이 빈 창"을 저자가 **고르지 못하게** 하는 방법이
      필요하다(예: 그 조건이 성립하는 가장 **늦은** 커밋으로 유도). 숫자를 지어내지
      않고 두며, 정하려면 실데이터로 재야 한다. 오늘 해당되는 change 는 a076 1건이다.
- [ ] 5.7 **base 가 작업 뒤에 놓인 change 는 5단계가 아무것도 요구하지 못한다
      (4.2 에서 실물로 나왔다).**
      a074 · a077 · a079 는 착지를 선언하면 요구 집합이 **0** 이 된다. 창이 비어서다 —
      base 와 착지 사이에 바뀐 Go 파일이 0 개이고, 세 change 의 번들 소스 12/12 가
      base 자체에서 이미 일치한다. 5단계는 그 0 을 **출력**하고 번들의 구조·소스
      해시는 계속 검사하므로(4.2 의 변이 P1·P2·P3 이 전부 빨갛다) 무판정은 아니다.
      그러나 "증거가 그 작업을 덮는가"는 묻지 못한다 — 요구가 0 이면 덮을 것이 없다.

      **이것은 a122 가 만드는 것이 아니다.** `base-commit.txt` 를 어디에 놓았는가의
      결과이고, 여기서는 2026-08-04 재기준화가 base 를 작업 뒤로 옮겼다. a122 는 그
      사실을 **보이게** 만들 뿐이다 — 전에는 319 개의 남의 함수에 묻혀 있었다.
      활성 change 중 `revision: current` 번들을 가진 13건 가운데 **8건**이 이 상태다
      (a074 · a077 · a079 · a089 · a091 · a092 · a094 · a095). 오늘 다수다.
      닫으려면 freeze 시점의 `base-commit.txt` 가 그 change 의 Go 작업 **앞**에
      있는지 물을 방법이 필요하다. 그 질문은 번들이 base 에서 일치하는지로 물을 수
      있다(4.2 의 census 가 그렇게 쟀다). 숫자·규칙을 지어내지 않고 둔다.

## 6. 독립 적대 리뷰가 연 것 (task 4.4)

- [ ] 6.1 **P0 — 착지 고정을 그 change 의 작업에 묶는다.** 오늘 `resolve_landing` 은
      "선언된 커밋에서 저자의 `revision: current` 번들 해시가 맞는가"만 묻고, 그 번들이
      **창 안에서 바뀐 파일을 기술하는가**는 묻지 않는다. 저자가 만든 값으로 저자가
      고른 값을 검증하는 순환이다. 실측 위조 셋:

      | 수단 | 대조군(선언 없음) | 선언 뒤 |
      |---|---|---|
      | 증거를 **base 상태**로 쓰고 `landed-commit = base` | `AST source hash is stale` → RED | `[]` · required **0** |
      | 안 건드린 파일의 **미끼 번들** 하나 + `landed-commit = base` | `missing evidence … Own` · required 1 | `[]` · required **0** |
      | 정직한 번들 둘 중 하나를 `revision: base` 로 **relabel** | `is not the revision this evidence describes` | `[]` · required 1, `Other()` 무분석 착지 |

      세 번째는 JSON 한 필드가 전부다 — `resolve_landing` 은 `base` 번들을 건너뛰고
      `validate_target` 은 `elif revision != "base"` 로 해시를 아예 안 본다.

      **5.3 의 축이 틀렸다.** 5.3 은 "번들이 몇 개면 충분한가"를 묻는데, a099 는 고정
      번들이 **37개**인데도 base..HEAD 239 커밋 중 **21개**가 그 고정을 통과한다.
      개수가 아니라 **번들과 창 사이의 결속**이 빠져 있다. 후보 방향(숫자·규칙은
      지어내지 않는다): 고정으로 세는 번들을 `git diff --name-only base landing -- '*.go'`
      안의 파일로 한정하고, 그 수가 0 이면 거절한다. 그러면 미끼도 base 상태 증거도
      고정에서 빠진다.

      **그 방향은 공짜가 아니다 — 5.7 과 정면으로 부딪친다.** a074 · a077 · a079 는
      base..착지 구간이 Go 파일 **0개**이므로(4.2 실측) 한정한 고정 수가 0 이 되어
      **선언 자체가 거절된다.** 그러면 대상이 워킹트리로 돌아가 required 319 가 되고,
      a122 가 풀어 주려던 셋이 다시 막힌다. 활성 change 13건 중 8건이 같은 모양이다.
      즉 6.1 은 "위조를 막는다"와 "base 가 작업 뒤에 놓인 change 를 푼다"를 **동시에**
      만족해야 하고, 그 둘이 한 판정 안에서 양립하는지가 이 task 의 진짜 질문이다.
      숫자·규칙을 지어내지 말고 13건 전체에 대해 실데이터로 가른 뒤 정한다.
- [x] 6.1.1 **측정 — 활성 13건 전수, 네 축을 실데이터로 걸었다.** HEAD `508f8b46`,
      생산 코드 변경 0. 결과는 review.md `VERIFY — task 6.1 (측정 단계)`.
      (1) 오늘의 고정은 **어느 change 에 대해서도** 착지를 결정하지 못한다 — 후보가
      하나인 것은 a066 한 건이고 나머지는 2~59개다. 후보 수는 번들 수와 무관하다
      (a112 번들 132 → 후보 39, a091 번들 2 → 후보 27). **5.3 의 "몇 개면 충분한가"
      에는 답이 없다.**
      (2) `landed-commit = base` 로 required 0 을 얻을 수 있는 change 가 **8건**이다
      (a074·a077·a079·a089·a091·a092·a094·a095). 위조가 아니라 오늘의 상태이고,
      정상과 위조가 **같은 모양**이라는 것이 문제다.
      (3) A축은 측정 이전에 **구조적으로** 그 8건을 거절한다 — 번들이 base 와 착지
      양쪽에서 맞으면 그 파일은 창 안에서 안 바뀐 것이므로 필터는 반드시 0 이다.
      4.4 가 3건의 실측으로 적은 충돌은 8건이고 동어반복이다.
      (4) B축(착지 ≥ 번들이 역사에 들어온 커밋)은 13건 전부에서 성립하는 정당한 축
      이지만(번들이 전부 자기 base 뒤에 커밋됐다) 후보를 하나로 좁히는 것은 여전히
      a066 한 건뿐이고, a089·a095 는 후보가 0 이 된다.
      (5) C축(가장 늦은 후보 강제)은 선택을 없애지만 a074 0→35 · a094 0→74 로
      **남의 작업을 다시 센다** — a122 가 없애려던 증상 그대로다.
      (6) "선택이 답을 바꾸지 않을 것"은 6/13 만 통과시키고 a074 를 거절한다.
      **결론: 가르는 정보가 change 디렉터리에도 창에도 없다.** a074 의 정답이 가장
      낮은 후보이고 위조자의 정답은 아닌데, 둘 다 증거가 base 에서 맞는다. 차이는
      "이 change 의 Go 작업이 실제로 어디 착지했는가"이고 저장소가 그것을 기록하지
      않는다. 그래서 6.1 은 "어느 규칙인가"가 아니라 **"무엇을 새로 기록할 것인가"**다.
- [x] 6.1.2 **결정 — 사람이 1번(게이트가 기록한다)을 골랐다(2026-09-11).**
      저자가 선언하는 대신 도구가 값을 **계산해서** 쓴다. 값을 만드는 주체가 바뀌는
      유일한 안이라는 것이 근거다. 손으로 쓴 값은 여전히 가능하지만 diff 에 보이는
      한 줄이 되고, 아래 하한이 그 한 줄이 고를 수 있는 범위를 없앤다.

      **무엇을 기록하는가 — "base 를 쓴다"는 내가 선택지에 적은 것이고 틀렸다.**
      base 는 위조와 같은 모양이다(6.1.1 소견 2). 6.1.1 이 실제로 준 값은 다르다:

      > **바닥(evidence floor)** = 그 change 의 고정 번들이 역사에 들어온 마지막 커밋.
      > **착지** = base..HEAD 안에서 바닥 이후이면서 고정을 통과하는 **가장 낮은** 커밋.

      a074 의 바닥은 `840b3377` 이고, 그것은 4.2 가 실제로 선언해 required 0 을 얻은
      바로 그 값이다(창 안 Go 파일 0). 저자는 바닥 아래를 고를 수 없다 — 오늘 만든
      번들을 과거 커밋에 넣을 수 없기 때문이다. 13건 전부 번들이 자기 base 뒤에
      커밋됐으므로(6.1.1 소견 4) 이 하한은 13건 전부에서 실재한다.

      **아카이브를 견딘다(실측).** `git log -M --diff-filter=MA -1 -- <아카이브 경로>
      <아카이브 전 경로>` 는 rename 커밋을 건너뛰고 내용 커밋을 찾는다. a099 의 바닥은
      `21a315d1` 이고 그 change 의 **기록된 착지 `e6c4636a` 이하**다 — 유일한 실물
      기록이 새 규칙에서 그대로 유효하다. 경로별 `--follow`(a112 61초)를 한 번의
      호출(0.02초)로 줄인 것도 이 형태다.
- [x] 6.1.2.1 **바닥을 계산하고 선언의 하한으로 세운다.** `resolve_landing` 에
      "선언된 착지는 바닥 이후여야 한다"를 더한다. RED 먼저 — 바닥 아래 선언이 오늘은
      통과하고 수정 뒤 거절되는 것을 시험으로 고정한다. 거절 문구는 그 가드의 **자기
      문장**이어야 한다(6.3 이 같은 파일에서 찾은 함정).
- [x] 6.1.2.2 **`--record-landing` — 도구가 값을 쓴다.** 기록이 이미 있으면 덮지
      않는다. 계산이 값을 못 내면 **쓰지 않고 이유를 이름으로 말한다**(a089·a095 가
      그 경우다 — 6.5). 워킹트리가 더러우면 기록하지 않는다: 기록은 커밋된 역사를
      가리키는데 통과는 워킹트리로 났으므로 둘이 다르면 거짓을 적는 것이다.
      **거부할 정상 입력을 먼저 열거한다** [[fail-closed-must-name-what-it-rejects]].
- [x] 6.1.2.3 **spec delta 를 고친다.** 오늘 본문은 "착지 지점에서 그 change 의 모든
      `revision: current` 증거 묶음의 source hash 가 일치해야 한다(SHALL)"만 적는데,
      6.1.1 이 그 판정만으로는 13건 중 12건에서 착지가 안 정해짐을 쟀다. 하한과
      "기록은 게이트가 쓴다"를 본문에 넣는다.
- [x] 6.1.2.4 **4.4 의 위조 셋을 새 코드로 재실행한다.** 무엇이 빨개지고 무엇이 안
      빨개지는지 **재서** 적는다. 셋 다 막힌다고 미리 쓰지 않는다 — 3번(`revision:
      base` relabel)은 `validate_target` 이 해시를 아예 안 보는 다른 결함이라 하한이
      닿는지 불명이다.
- [x] 6.1.2.5 **FLM/BTM 갱신.** `resolve_landing`·`check`·`main` 의
      `analysis/python-function-logic/` 열거를 편집 전후로 다시 뽑는다.
- [ ] 6.5 **P1 — `revision: current` 번들이 자기를 담은 커밋에서 이미 틀릴 수 있다.**
      6.1.1 측정 중 발견. a089 의 `internal/journal/outbox.go`, a095 의
      `internal/obs/notifier.go`(번들 2개)는 base 에서는 해시가 맞고 **그 번들을 담은
      커밋 `a30eb35ae` 에서는 안 맞는다** — 같은 커밋이 그 Go 파일을 바꿨기 때문이다.
      증거를 만든 뒤 같은 세션에서 Go 를 한 번 더 고치고 둘을 한 커밋에 담으면 이
      상태가 된다. 오늘은 대상이 워킹트리라 HEAD 하고만 비교해서 아무 게이트도 못 본다.
      착지를 선언하는 순간 드러나므로 6.1 결정과 함께 처리한다.
      [[generated-evidence-must-be-measured]]
- [ ] 6.1.2.6 **`make gate` 가 기록을 자동으로 쓸지 사람이 정한다.** 이번에 만든 것은
      명령(`--record-landing`)이고, 5단계 조언 줄이 그 명령을 이름으로 부른다.
      게이트가 **스스로 쓰게** 하지는 않았다 — 검사 게이트가 워킹트리를 바꾸면
      게이트의 뜻이 달라지고, 저장소 규칙이 mutating 단계를 사람 승인으로 묶는다.
      값을 만드는 주체가 도구로 바뀌는 결정은 이미 지켜졌다(저자는 계산된 값을
      복사할 뿐이다). 자동 쓰기를 원하면 그때 켠다.
- [ ] 6.6 **P1 — 하한을 지키면서도 창을 비우는 순서가 남는다 (실측).** 증거를
      **먼저** 커밋하고 그 지점을 착지로 기록한 뒤 Go 작업을 **그 뒤에** 붙이면,
      창 `base..기록` 이 그 작업을 안 담아 required 0 으로 초록이다. 픽스처로 재현:
      미끼 번들 하나 → `record_landing` rc=0 → 기록 커밋 → check `[]` required 0 →
      Go 작업 커밋 → check **여전히 `[]` required 0**.
      6.1.2 가 만든 것이 아니라 6.1.2 **뒤에 남은** 것이다(전에도 있었다). 기록이
      한 번 쓰이고 나면 뒤에 오는 작업과 다시 대조되지 않는 것이 기제다.
      규칙은 지어내지 않는다 — 후보 방향(그 change 의 디렉터리를 만진 커밋이 기록
      뒤에 있으면 경고/거절)이 1.12 의 "신원으로 판정 금지"와 어떻게 갈리는지부터
      가른 뒤 정한다.
- [x] 6.2 **P0 — 아카이브된 이관 change 가 자기 id 로 재검사되게 한다.** spec 은
      "아카이브된 change 의 함수 분석도 그 id 로 재검사할 수 있어야 한다(SHALL)"를
      요구하는데, 이관 경로만 못 간다. `execution_baseline.validate` 가
      `change_dir.name != CHANGE` 로 **날짜 없는 id** 를 요구하기 때문이다.
      a063 의 상태는 건드리지 않고 `execution-baseline.json` 만 날짜 붙은 이름의 임시
      디렉터리에 복사해 쟀다 — 활성 이름은 `adoption requires detached HEAD`(이름 검사
      통과), 아카이브 이름은 `execution-baseline adoption is not allowed for this
      change/base`. a063 이 아직 활성이라 **오늘은 잠복**이고 아카이브되는 날 터진다.
      **4.3 이 "1.12 를 실데이터로 못 돌렸다"고 넘긴 바로 그 사각지대다.**
      정본 신원을 디렉터리 basename 과 분리해 판정하고, 증거 경로·digest 검사는
      그대로 둔다.

      **닫힘 (2026-09-11). 처방의 뒤 절반이 틀렸다.** 가드를 하나씩 풀어 재 보니 아카이브
      뒤 막는 자리가 **넷**이다 — 이름 · 증거 경로 접두사 · 원장 읽기 · 리뷰 읽기(다섯째
      없음). 기록이 옮기기 **전** 경로를 적기 때문이고, 경로 검사를 그대로 두면 이름을
      고쳐도 둘째 줄에서 막힌다. 신원은 게이트가 요청받은 id 로(`validate` 의 필수 인자 —
      호출자 셋이 각자 해소한 id 를 넘긴다), 경로 판정은 **적힌** 자리로, 읽기는 **지금**
      자리로 하고 digest 비교는 한 글자도 안 바꿨다. `validate` 는 아카이브 문법을 새로
      배우지 않는다(집이 이미 둘이다, 6.4(f)). 변이 M1~M6 전부 CAUGHT.
      **실데이터 확인은 아니다** — a063 은 이 HEAD 에서 `adoption requires detached HEAD`
      앞을 못 지나고(4.3 과 같은 줄, 편집 전후 동일), 근거는 저장소 픽스처다.
- [x] 6.2.1 **P0 — 이관 경로의 깨진 착지 기록이 `check()` 를 터뜨린다.**
      `_declared_landing` 호출 다섯 자리 중 `check` 의 어댑션 probe(`if adopted and
      _declared_landing(...) is not None`) **한 곳만** try 밖에 있다. 비-UTF-8 기록이면
      `ValueError` 가 `check()` 를 뚫고 나가 `main()` 이 traceback 으로 죽고,
      **3.3 이 보장하기로 한 `[logic-map]` 창 줄이 하나도 안 찍힌다.** 저장소 자신의
      `_adoption_with_complete_bundle` 픽스처로 실측됐다. `adopted` 가 거짓이면
      단락되므로 오늘 닿는 change 는 a063 하나 — 6.2 와 **같은 사각지대**다.
      그 값을 위의 try 안에서 한 번 구해 재사용하고, 음성 경로 시험을 단다.

      **닫힘 (2026-09-11). 자리가 둘이었다.** 호출 자리를 AST 로 다시 세니 `record_landing`
      의 덮어쓰기 probe(6.1.2 가 만든 자리 — 4.4 **뒤**라 4.4 가 셀 수 없었다)도 try 밖이다.
      두 자리 다 기록이 **있는가**를 묻는데 해독하는 함수를 불렀다. try 로 감싸지 않고
      질문을 떼어 냈다: `_landing_record` 가 있는가만 답하고 `_declared_landing` 이 그 위에서
      해독한다(오류 문장은 그대로). spec 이 이관 경로의 착지 기록에 "받지 않는다고 이름으로
      말한다"를 요구하므로 못 읽는 기록도 **거절**로 답한다 — 4.4 처방(해독 오류로 돌려주기)
      과 다르다. 저장소에 비-UTF-8 기록을 재는 시험이 **0** 이었고 셋을 더했다. 변이 M7~M9
      CAUGHT.
- [x] 6.3 **P1 — 거절 시험이 실패 지점을 단언하게 한다.** 뮤테이션 M2·M3·M4 가
      살아남은 원인은 하나다. `AForgedLandingPointIsRefusedByName._refuse(value, needle)`
      의 바늘이 네 자리에서 `"landing"` 인데 그 단어는 착지 관련 **모든** 오류 문장에
      들어 있어서, 가드를 지워도 다른 가드가 거절하고 시험이 초록으로 남는다. M3
      (`landing ≤ HEAD`)은 시험 자체가 없다 — 곁가지 커밋을 만드는 픽스처가 0 이다.
      이 파일은 같은 함정을 **이미 한 번 발견하고 한 자리만 고쳤다**
      (`test_a_landing_the_evidence_does_not_describe` 의 주석이 "M4 로 실측"이라고
      적는다). 형제 넷에 같은 처방을 한다: 바늘을 그 가드의 **자기 문장**으로 바꾸고,
      곁가지 커밋 픽스처를 더해 `never landed on this history` 를 못 박는다.
      [[passing-test-is-not-evidence]] · "거절 테스트는 실패 **지점**을 단언할 것".

      **Testing 전문가가 이 세션과 따로 같은 실험을 하고 같은 결론에 닿았다**(가드
      넷을 각각 지워도 75 전부 초록). 그리고 이 세션이 안 건드린 셋을 더 찾았다 —
      `gate.sh` 1단계 `RESOLVE_ERROR` 블록(지워도 8/8 초록인데 hits>1 이면 fallback
      활성 디렉터리가 **실재**해서 변이가 4단계까지 간다), `gate.sh` 의
      `YYYY-MM-DD-` `case` 가드(지우면 `?` 와일드카드가 `archive/abcd-ef-gh-…` 를
      먹는다), `_target_text` 의 비이관 갈래(시험이 SHA 만 `assertIn` 해서 라벨을
      안 본다). 이 셋도 같이 못 박는다.

      **닫힘 (2026-09-12). 먼저 쟀다** — 지금 코드에서 착지 가드 넷(40-hex · 커밋 실재 ·
      `landing ≤ HEAD` · `base ≤ landing`) · 비이관 라벨 · gate.sh 두 자리까지 **일곱이 전부
      SURVIVED**. 넷째 가드(커밋 실재)는 4.4 표에 없었다. 바늘을 각 가드의 자기 문장으로
      바꾸고, 곁가지 커밋 픽스처로 `never landed on this history` 를, gate 대상이 활성·아카이브에
      동시에 있는 픽스처로 1단계의 `확정할 수 없는 change-id` 를, 날짜 모양이 아닌 아카이브
      이름 픽스처로 `case` 가드를 못 박았다. 라벨은 SHA 만 보던 자리를 **그 자리에서**
      `landed-commit <sha>` 로 좁혔다. 같은 스크립트로 다시 재니 **일곱 전부 CAUGHT** — 각
      변이를 빨갛게 한 것이 그 가드의 시험이다(40-hex 만 둘). 생산 코드 변경 0.
- [ ] 6.4 **P2 묶음.** 각각 작고 독립이다.
      (a) 착지를 선언하면 **커밋 안 된 Go 편집이 5단계에 안 보인다**. `make gate` 는
      `make sdd-check` fingerprint 로 일부 가리지만 `make sdd-sync` 를 다시 돌리면
      풀린다. 이관 경로가 이미 쓰는 `git diff --quiet` 를 착지 경로에도 둘지 결정한다 —
      **거부할 정상 입력을 먼저 열거할 것**([[fail-closed-must-name-what-it-rejects]]).
      (b) `base-commit.txt` 는 여전히 워킹트리에서 읽고 40-hex 검사도 없다. 같은 창의
      반대쪽 끝이 안 잠겨 있다(a122 이전부터 있던 것).
      (c) `ARCHIVED_CHANGE` 의 `\d` 는 유니코드 숫자를 먹고 `gate.sh` 의 `[0-9]` 는
      안 먹는다 → 두 해소기가 "날짜 접두사"의 뜻에 동의하지 않는다. `re.ASCII` 한 글자.
      (d) `check` 의 `except ValueError` 가 `archive holds N copies` 를 삼키고
      `missing base-commit.txt` 로 바꿔 말한다. `gate.sh` 는 같은 조건에 큰 소리로 멈춘다.
      (e) `gate.sh:110` 주석이 "Python 보다 엄격하다"고 하는데 **같은 diff 가** Python
      쪽에 `AmbiguousChange` 를 넣어 그 차이를 없앴다. `test_check_analysis.py` 의
      docstring 은 지워진 `return direct` 를 현재형으로 인용하고 `:689-692` 는 이제
      무관한 `except` 를 가리킨다([[frozen-census-needs-edit-stable-coordinates]]).
      (f) 규칙이 두 집(`resolve_change_dir` / `resolve_referenced_change`)에 사는데
      **동기화를 강제하는 것이 아무것도 없다** — 각 스위트가 자기 사본만 못 박아서
      한쪽만 고치면 둘 다 초록이다. 공유 표 하나로 두 해소기를 같이 돌리는 시험.
      (g) 성능(advisory): 같은 (ref, path) blob 을 `resolve_landing` 과
      `validate_target` 이 두 번 해싱한다(a099 실측 74 spawn → 서로 다른 blob 12개).
      (h) 픽스처가 개발자의 전역 git config 를 상속한다 — `commit.gpgsign = true` 가
      전역에 있으면 새 클래스 전부가 `gpg: signing failed` 로 **에러**가 된다(실측).
      `GIT_CONFIG_GLOBAL=/dev/null` 을 주거나 `_init_fixture` 에서 명시적으로 끈다.
      같은 노출: `core.hooksPath` · `gpg.format` · `core.autocrlf`.

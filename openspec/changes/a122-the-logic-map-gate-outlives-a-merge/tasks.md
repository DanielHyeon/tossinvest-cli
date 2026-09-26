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
- [x] 1.4 proposal-freeze 적대 리뷰와 gstack 리뷰를 마치고 기록한다.
      **7 절 로트 독립 리뷰 실행·수리 완료(2026-09-18).** 기록은 review.md
      `## VERIFY — task 1.4 독립 적대 리뷰 (7 절 로트) + 수리`.
      7 절은 전부 한 저자가 구현하고 그 저자가 검증했으므로 **다른 세션**(`tossos-be`)에
      맡겼다. 내 주장 넷과 내가 아는 의심 자리 넷을 **반증 대상으로 먼저 넘겼다**.
      판정 **CHANGES REQUESTED — P0 1 · P1 2 · P2 10**, 그리고 전부 수리했다.
      **P0 은 내 주장 3 이 거짓이었던 것**이다 — T6("동등 변이")은 동등이 아니라 **미도달**
      이었고, 마지막 머리가 멀쩡하고 내용만 잘린 응답에서 `None` 대신 `b""` 를 내어 판정을
      바꾼다(재현 확인). `end < 0` 을 판정으로 바꾸고 닿는 픽스처를 세웠다. 같은 성격의
      구멍 셋(남은 바이트 · 경로 NUL · 엄격 utf-8)도 같이 닫았다.
      P1 둘: `-Z` 최소 git 버전이 어디에도 없던 것 → `GIT_BATCH_MINIMUM` 한 곳 + 문서 둘 +
      일치 시험(rc≠0 을 결함으로 올리는 안은 **거부할 정상 입력 21건**이 나와 되돌렸다).
      편집 전 AST 10 중 **7 이 편집 중간 상태**에서 뽑힌 것 → 진짜 부모 `8091e6c4` 에서 재추출
      (내용은 열 개 다 동일 — 방법이 불건전하고 결과가 우연히 맞았다).
      P2 중 근거 정밀도 셋을 고쳤다: 가드 순서를 **multiset 수가 아니라 순서열**로 다시 재고
      ("여섯 전부 불변"은 과했다 — `_walk_floor` 반환 계약이 2→3 튜플), A/B 범위를
      `compute_landing` 으로 좁혀 적고, 배수는 **spawn 수로** 적는다(벽시계는 재현 안 된다).
      검증: 시험 187 · 변이 **22 전부 CAUGHT(생존 0, 양성 대조 포함)** · A/B **116/116 SAME** ·
      `sdd-test`(logic-map 262) · `lint` · `openspec validate` 58/58 · a122 rc=0.

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

      **→ 7.2.3(2026-09-14)이 이 결정을 뒤집었다**: 빌리는 change 는 창을 좁히지 않는다(리뷰 C4 — 복사한
      착지 뒤의 빌리는 쪽 작업이 창 밖으로 나갔다). 아래는 결정 당시의 기록이다.
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
      **→ 7.2.2(2026-09-14)로 그것도 못 한다.** 셋은 착지를 얻지 못하고(증거가 base 의 소스를 적어
      V1 과 같은 모양) 넓은 창으로 돌아간다. 셋의 5단계는 a075 · a076 이 2026-09-08 에 쓴 방식 —
      사유를 적은 사람의 면제 — 로 가야 한다. 도구에 그 경로는 없다.
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
- [x] 5.5 **빌린 증거를 쓰는 change 의 자기 작업은 고정되지 않는다 (1.8 에서 열림).**
      **→ 7.2.3(2026-09-14)으로 닫혔다**: 빌리는 change 는 착지를 갖지 못하므로 창이 워킹트리까지이고, 공유된 착지
      뒤의 자기 작업이 요구 집합에 들어온다(`test_a_lender_record_does_not_narrow_the_borrower`).
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
      (4.2 에서 실물로 나왔다).** **→ 7.2.2(2026-09-14)로 모양이 바뀌었다**: 그런 change 는 이제 착지를
      얻지 못하므로 "요구 0" 이 아니라 넓은 창(남의 함수까지 요구)으로 판정된다. 아래는 결정 전 기록이다.
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

- [x] 6.1 **P0 — 착지 고정을 그 change 의 작업에 묶는다.** 오늘 `resolve_landing` 은
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
      **닫음(2026-09-26) — 자식 여덟이 전부 `[x]`, 위조 표 셋을 지금 코드로 다시 돌렸다.** 기록은 review.md
      `## VERIFY — task 6.4 (b)~(f) · 6.1 부모`. 6.1.2.4 가 돌렸다는 "4.4 의 스크립트" 는 저장소에 없어서 같은 세 모양을
      스위트의 픽스처 헬퍼로 다시 만들었다(`analysis/harness/61_forgeries.py`, HEAD `5a54f78d`). **셋 다 빨갛다**: (a) · (b) 는
      6.1.2.4 와 같은 문장(`precedes the evidence that pins it`). (c) 는 **선언 값에 따라** 문장이 다르다 — base(P) 를 선언하면
      `is not the revision this evidence describes`(불일치 판정이 하한보다 앞에 서서 먼저 거절), Go 작업 커밋(L)을 선언하면 6.1.2.4 와
      **글자 그대로 같은** `precedes the evidence that pins it` 이다. **정정(6.4 보수, 주장정확성 리뷰 P1-F2)**: 첫 판은 "(c) 문장이 다른
      것은 원 픽스처를 못 되살린 탓" 이라 적었다 — 거짓이다. 같은 픽스처에 L 을 넘기면 같은 문장이 나오므로 다른 원인은 스크립트가 고른
      선언 값(P)이었다(`61_forgeries.py` 에 L 행을 더해 다시 돌렸다, HEAD `5a54f78d`). (c') 는 **정직한 증거 커밋**(E)을 선언한 추가 모양이고
      이제 `AST revision must be current` · required 2 로 빨갛다 — relabel 은 착지와 무관하게 `_verdict` 의 revision 기대값이 막는다.
      첫 판은 (c') 를 "4.4 가 `[]` 를 얻은 모양" 이라 적었는데 그것도 맞지 않는다: 4.4 는 required **1**, (c') 는 **2** 다 — 4.4 의 정확한
      선언 값은 스크립트가 저장소에 없어 되살릴 수 없고, (c') 는 재현이 아니라 더한 모양이다. 기록 명령은 (a)(b) 를 거절하고 (c) 는
      정직한 번들이 고정하는 커밋을 기록한다(그 뒤 판정이 relabel 로 빨갛다). 셋 다 빨갛다는 판단은 그대로다.
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
- [x] 6.5 **P1 — `revision: current` 번들이 자기를 담은 커밋에서 이미 틀릴 수 있다.**
      6.1.1 측정 중 발견. a089 의 `internal/journal/outbox.go`, a095 의
      `internal/obs/notifier.go`(번들 2개)는 base 에서는 해시가 맞고 **그 번들을 담은
      커밋 `a30eb35ae` 에서는 안 맞는다** — 같은 커밋이 그 Go 파일을 바꿨기 때문이다.
      증거를 만든 뒤 같은 세션에서 Go 를 한 번 더 고치고 둘을 한 커밋에 담으면 이
      상태가 된다. 오늘은 대상이 워킹트리라 HEAD 하고만 비교해서 아무 게이트도 못 본다.
      착지를 선언하는 순간 드러나므로 6.1 결정과 함께 처리한다.
      [[generated-evidence-must-be-measured]]
      **닫음(2026-09-14).** 기록은 review.md `## Pre-Edit Gate — task 6.5` · `## VERIFY — task 6.5`.
      **먼저 다시 쟀다**: "아무 게이트도 못 본다"는 오늘 틀렸다 — HEAD `483985dc` 에서 a089 · a095 는 워킹트리 대상 5단계가
      `AST source hash is stale` 로 빨갛고 7.1 조언이 base 를 적은 번들을 이름으로 대며, 착지를 선언하면 규칙이 틀린 파일을
      이름으로 거절한다. **남은 결함은 기록 명령의 사유 한 줄이었다** — "the evidence does not describe any revision on this
      history". 두 change 의 증거는 base 를 정확히 기술하므로 도구가 내는 유일한 사유가 사실과 반대였다. 걷기 실패 **17건
      전수**를 하한 아래까지 걸어 재니 **4건**(a089 · a095 · console-click-approval · verify-us-market)에서 실측으로 거짓이고
      나머지 13건도 순회가 안 걷는 범위를 주장했다. 수리: 사유는 걸은 것만 말하고, 무엇이 틀렸는지는 규칙이 **첫 후보**에 준
      문장을 그대로 인용한다(새 호출 0, 17/17 에서 첫 후보 = 하한). a089 는 이제 `… at the first commit walked, landing point
      a30eb35ae6da is not the revision this evidence describes: internal/journal/outbox.go` — 6.1.1 이 손으로 짚은 그 파일이다.
      기존 위조 기록 시험의 바늘이 **거짓 문장을 못 박고 있었다**(그 픽스처에서도 증거는 base 를 기술한다). 시험 141 → **143** ·
      변이 N1~N4 **CAUGHT**(N2 는 두 후보 픽스처 하나만 잡는다) · 옛 하네스 7.6 · 7.7 · 7.8 재실행 결과 불변(R14' 포함) ·
      실물 A/B 93건 **값·사유 같음 76 · 꼬리만 다름 17(편집 전 첫 후보 문장과 글자 그대로) · 그 밖 0** · `make sdd-test`
      (logic-map 218) · `lint` · `test-seams` · `validate --all` 58/58. 계측 사고: 사본 하네스가 형제 import 를 빠뜨려 무변이
      대조군부터 빨갰고 변이가 전부 "CAUGHT" 로 찍혔다 — 대조군이 잡았고, 이제 하네스가 대조군이 초록이 아니면 멈춘다.
- [x] 6.1.2.6 **`make gate` 가 기록을 자동으로 쓸지 사람이 정한다.** 이번에 만든 것은
      명령(`--record-landing`)이고, 5단계 조언 줄이 그 명령을 이름으로 부른다.
      게이트가 **스스로 쓰게** 하지는 않았다 — 검사 게이트가 워킹트리를 바꾸면
      게이트의 뜻이 달라지고, 저장소 규칙이 mutating 단계를 사람 승인으로 묶는다.
      값을 만드는 주체가 도구로 바뀌는 결정은 이미 지켜졌다(저자는 계산된 값을
      복사할 뿐이다). 자동 쓰기를 원하면 그때 켠다.
      **→ 7.2 전에는 정할 수 없다 (2026-09-12 리뷰).** 자동 기록은 C1(stale 증거를 편집 전
      커밋으로 세탁)을 자동화한다.
      **닫음 — 사람이 2026-09-16 에 "자동 기록 안 함"을 골랐다.** 게이트는 계속 조언만 하고
      `--record-landing` 은 사람이 친다. 이유 셋: (1) 검사 게이트가 워킹트리를 바꾸면 게이트의 뜻이
      달라지고 저장소 규칙이 mutating 단계를 사람 승인으로 묶는다, (2) 7.2.2 가 C1 을 규칙으로 막았어도
      자동 기록은 **언제** 기록하는지를 도구가 정하게 만든다 — 값은 이미 도구가 만들지만 시점은 저자의
      것이고 그 시점이 곧 창의 끝이다, (3) 7.2.6 뒤로 기록은 **낡을 수 있고** 복구는 기록을 지우는
      커밋을 요구한다. 자동으로 쓰면 그 삭제·재기록 고리를 사람 없이 돌게 된다.
- [x] 6.6 **P1 — 하한을 지키면서도 창을 비우는 순서가 남는다 (실측).** 증거를
      **먼저** 커밋하고 그 지점을 착지로 기록한 뒤 Go 작업을 **그 뒤에** 붙이면,
      창 `base..기록` 이 그 작업을 안 담아 required 0 으로 초록이다. 픽스처로 재현:
      미끼 번들 하나 → `record_landing` rc=0 → 기록 커밋 → check `[]` required 0 →
      Go 작업 커밋 → check **여전히 `[]` required 0**.
      6.1.2 가 만든 것이 아니라 6.1.2 **뒤에 남은** 것이다(전에도 있었다). 기록이
      한 번 쓰이고 나면 뒤에 오는 작업과 다시 대조되지 않는 것이 기제다.
      규칙은 지어내지 않는다 — 후보 방향(그 change 의 디렉터리를 만진 커밋이 기록
      뒤에 있으면 경고/거절)이 1.12 의 "신원으로 판정 금지"와 어떻게 갈리는지부터
      가른 뒤 정한다.
      **→ 7.2 로 흡수 (2026-09-12 리뷰).** 잔여가 아니라 기본 경로였다 — 저장소 규칙이 요구하는
      FLM-first 순서가 곧 "증거 먼저, 작업 나중"이다(review.md V1).
      **7.2.2 뒤에 다시 쟀다 (2026-09-14, `722_holes.py`).** 위 픽스처(미끼 번들 → 기록)는 이제 **막힌다** —
      미끼의 소스는 base 와 같아서 기록 명령이 rc 1 이고, 뒤의 Go 작업은 5단계에서 요구 1 · 증거 없음으로
      빨갛다. **남는 모양 둘**: (H2) 소스가 바뀐 정직한 번들로 기록한 **뒤에** 증거 없이 붙인 작업 — 5단계
      초록(요구 1, 뒤 작업은 창 밖); (H3) FLM 을 둘 먼저 쓰고 하나만 갱신한 뒤 다른 파일을 나중에 고친 모양 —
      기록 rc 0 · 초록. H2 는 창을 좁히는 한 본질이다(뒤의 작업이 누구 것인지 내용으로 못 가른다). H3 를 막는
      변형(착지에서 base 와 같은 고정 소스가 착지 뒤에 바뀌면 거절)은 편집 전 전수에서 상실이 8 → **14**
      (+a100 활성 · a047 · a049 · a050 · a052 · a060 아카이브)라 결정 범위 밖이다 — 사람 결정으로 남긴다.
      **→ 7.2.6 결정에 달림(2026-09-16).** 7.2.6 의 변형 B-ANYGO 가 한 커밋 모양의 H2 를 덮는다(픽스처 실측). 쪼갠 커밋 모양은 남는다.
      **닫음(2026-09-16) — 7.2.6 으로 흡수.** H2 의 한 커밋 모양(기록 뒤의 작업이 그 change 의 문서와 같은 커밋에 있는 경우)은
      이제 거절된다(`test_any_go_file_counts_not_only_the_pinned_one`). 쪼갠 커밋 모양은 spec 에 **알려진 한계**로 적었다 —
      내용으로 가르는 규칙은 이웃과 자기 수리를 못 갈라 68건 중 61/55 를 거절한다(7.2.6 전수). 후보 방향이 1.12 와 어떻게
      갈리는지도 갈랐다: **거절에만** 쓰므로 신원으로 착지를 받는 판정이 아니고, 틀리면 창이 넓어지는 쪽으로만 틀린다.
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
      **(h) 닫음(2026-09-27).** 기록은 review.md `## VERIFY — task 7.5.14 · 7.5.18 · 6.4(h)`. 시험 모듈이 import 때 `GIT_CONFIG_GLOBAL` ·
      `GIT_CONFIG_SYSTEM` 을 `os.devnull` 로 둔다(지역에서 따로 끄는 헬퍼는 그대로). 실측(HEAD `3bb11b2d`, 스위트 433): 전역에 `commit.gpgsign=true` +
      실패하는 `gpg.program` → **409** 에러, 실패하는 `pre-commit` 을 가진 `core.hooksPath` → **419**, `gpg.format=ssh` + 서명 → **409**, `core.autocrlf=true`
      → 0(픽스처 바이트에 CR 없음). 고친 뒤(439) 넷 다 **0**. 빈 전역 설정으로 돌려도 433 전부 초록 — 전역 `user.*` 에 기대는 픽스처 0.
      시험 `TheSuiteDoesNotReadTheDevelopersGitConfig`(자식 프로세스에 적대 설정을 주고 픽스처를 세운다), 변이 TM1 · TM2 CAUGHT.
      **보수(적대 리뷰 뒤, 2026-09-27)**: 막는 범위를 **설정 파일**에서 설정의 출처 셋으로 넓혔다 — 전역 · 시스템 파일과 환경 변수
      (`GIT_CONFIG_PARAMETERS` · `GIT_CONFIG_COUNT/KEY_<n>/VALUE_<n>` — 파일 고정은 이것을 못 막았다: `GIT_CONFIG_COUNT` 로 서명을 켜면 445 중
      실패 4 · 에러 448). 규칙은 `tools/logic-map/fixture_git_env.py` 한 곳이고 `test_execution_baseline.py` 도 부른다 — 그 모듈은 `discover` 에서
      이 모듈 뒤에 import 되어 **우연히** 보호받았다(단독 실행은 적대 설정에서 27 중 21 에러). 고친 뒤 둘 다 0. 변이 TM2b · TM2c CAUGHT.
      **(b)~(f) 닫음(2026-09-26)** — 기록은 review.md `## VERIFY — task 6.4 (b)~(f) · 6.1 부모`. 부모는 (a)(사람 결정) ·
      (g) · (h) 가 남아 열려 있다.
      (b) `resolve_base` 가 적힌 값이 **40자리 소문자 커밋 id 자신**인지(이름 · 짧은 id · 대문자 · 태그 객체 id 거절) 보고, 명령이 한 번
      푼 `head` 커밋에 같은 파일이 있으면 **그 값과 같아야** 받는다(커밋 안 한 편집이 창의 시작을 못 옮긴다). HEAD 에 없는 새 base 는
      디스크의 값 — freeze 직후 커밋 전 실행은 정상이다. 거절하는 정상 입력 0(HEAD `5a54f78d` 재측정: 저장소 `base-commit.txt` 119 전부
      40자리 + LF · 자기 id 가 커밋 · 자기 경로로 HEAD 와 바이트가 같다 — `analysis/harness/64r_census.py`).
      **보수(2026-09-26, 독립 리뷰 둘 — 적대 · 주장정확성).** 첫 판의 HEAD 대조는 **지금 디스크 경로**로 HEAD 를 물어서, 디렉터리를
      옮기면(커밋 안 한 `mv` · staged `git mv` · 아카이브 날짜 변경) "HEAD 에 없는 새 base" 로 받았다(P0-A). 이제 지금 경로가 HEAD 에
      없으면 같은 id 의 다른 자리(활성 `changes/<id>/` · HEAD 트리의 `changes/archive/<날짜>-<id>/`)를 찾아 값이 다르면 이동을 이름 대고
      거절한다(새 잎 `_committed_elsewhere`). 새 갈래가 거절하는 정상 입력 0(같은 측정 — 새 갈래에 닿는 디렉터리 0). **보장 범위를
      좁혀 적었다**(P1-C): 막는 것은 커밋 안 한 편집과 이동이다. base 를 고쳐 **커밋**하는 재기록은 어느 가드도 막지 않는다 — 이 저장소의
      재동결(HEAD 역사에서 `base-commit.txt` 제자리 재기록 커밋 16개 · id 13개, `--diff-filter=M`)은 "쓰고 커밋하기 전" 에
      `… an uncommitted edit does not move the comparison base; restore the committed value` 로 거절되고 커밋하면 받는다.
      **7.5.5 는 열린 채다(Manager 판정 2026-09-26)** — 이동 공격이 그 질문의 답을 "예" 로 유지했고, 이 보수 뒤에도 심링크 · `root`
      기준 · 커밋된 재기록이 남는다(아래 (i) · (k) · 사람 결정).
      (c) `ARCHIVED_CHANGE` 에 `re.ASCII`. **정정(보수, P0-F1)**: 첫 판은 shell 의 `[0-9]` 가 유니코드 숫자를 "안 먹는다" 고 적었지만
      bash 5.2.21 · `en_US.UTF-8` 에서 범위 `[0-9]` 는 콜레이션으로 `０…８` · `٠…٨` · `𝟎…𝟖` 를 먹는다(`９` 만 제외, `C.UTF-8` 에서는 안
      먹음). 9 없는 날짜면 그 로케일에서 두 해소기가 갈렸다. `gate.sh` 를 나열 `[0123456789]` 로(로케일 무관), 표에 9 없는 날짜 세 행,
      shell 을 `C.UTF-8` · `en_US.UTF-8` 둘에서. (d) **이미 닫혀 있었다** — 7.6(I4)이 fallback 을 지워 `_judged` · `record_landing` 이 해소기의
      문장을 그대로 낸다. 삼킴을 되살리는 변이 AL7 · AL8 이 CAUGHT(시험 `test_two_archived_copies_are_named_on_both_paths` 외).
      (e) `gate.sh` 해소기 주석과 `test_check_analysis.py` 의 docstring 둘을 과거형 + 함수 이름으로 고쳤다(줄 좌표 없음).
      (f) 시험 `TheTwoChangeResolversAgree` — 한 표(12 행)를 `resolve_referenced_change` 와 **실제 `gate.sh`** 에 같이 돌린다. (c) 는 그
      표의 첫 발견이다(RED: 전각 · 아라비아-인도 숫자 두 행). 시험 377 → **384** · Python 변이 AL1~AL8 **8/8 CAUGHT** · shell 변이 G1~G4
      (`64_gate_mut.py`) — 결과는 review.md. **보수 뒤**: 표 15 행을 로케일 둘(`C.UTF-8` · `en_US.UTF-8`)에서, 시험 **388**, Python 변이
      AL1~AL16 **16/16 CAUGHT**(새 AL9~AL16 — `fullmatch`→`match` 생존을 `<sha>~1` 행으로 닫았다), shell G1~G5 **5/5 CAUGHT**. 기록은 review.md
      `## 정정 · VERIFY — task 6.4 보수 (독립 리뷰 둘: 적대 · 주장정확성, 2026-09-26)`.
      **보수가 연 것 — 고치지 않고 열어 둔다:**
      (i) 심링크 change 디렉터리: untracked 대상을 가리키는 심링크를 커밋하면 `base-commit.txt` 가 HEAD 에 **영원히** 없는 모양이 되어 대조가
      안 선다([[a-silent-skip-is-a-door]] 계열). 적대 리뷰 보고, 이 세션 재현 안 함.
      (j) 해소기 표 밖 불일치: 아카이브 이름의 줄바꿈(Python missing · shell found), 아카이브 디렉터리를 못 읽을 때 Python 은 `PermissionError` 가
      `_judged` 의 `except ValueError` 밖으로 샌다. 적대 리뷰 보고, 이 세션 재현 안 함.
      (k) `_committed_bytes`(와 보수의 `_committed_elsewhere`)의 경로 기준이 git toplevel 이 아니라 `root` 다 — `root` ≠ toplevel 이면 가드 ②가
      조용히 꺼진다. 착지 기록 읽기도 같은 노출이고 a122 이전부터다.
      **사람 결정 — 커밋된 재기록 차단(역사 잠금).** base 를 고쳐 **커밋**하면 어느 가드도 막지 않는다(그 값이 곧 HEAD 의 값). 역사에 묶어
      잠글지는 사람이 정한다([[author-supplied-evidence-cannot-pin-an-author-choice]]). 부모 6.4 는 (a) · (g) · (h) · (i) · (j) · (k) · 이 결정이
      남아 열려 있다. (2026-09-27: (h) 닫음 — 남은 것은 (a) · (g) · (i) · (j) · (k) · 이 결정.)

## 7. §6 로트 독립 리뷰가 연 것 (gstack /review, 2026-09-12)

기록은 review.md `## VERIFY — §6 로트 독립 리뷰`. 원천 여덟, CRITICAL 둘은 이 세션이 손으로
재현했다(V1 · V2). **4.5 는 계속 막힌다.** 사람이 2026-09-12 에 B(모두 기록 + 조언 줄만 즉시
막기)를 골랐다.

- [x] 7.1 **P0 즉시 완화 — 5단계가 stale 증거 옆에서 `--record-landing` 을 권하지 않는다.**
      V1: FLM 을 먼저 커밋하고(저장소 규칙의 순서) 편집 뒤 번들을 안 갱신하면 5단계가
      `AST source hash is stale` 로 빨갛고, **같은 출력이** `--record-landing` 을 권하고, 도구가
      편집 전 커밋을 기록해 required 0 으로 통과한다. 조언 줄만 바꾼다 — 게이트 판정은 그대로다.
      그 줄이 바뀌는 **정상 입력**(나중에 이웃이 같은 파일을 고친 change 도 워킹트리 모드에선
      stale 로 보인다)을 활성 전수로 먼저 잰다. 명령이 편집 전 커밋을 계산하는 것 자체는 7.2 다.
      **닫음(2026-09-12).** 사람이 좁힌 판본을 골랐다 — stale 이라고 다 막지 않고
      **base 의 소스를 적은 번들**이 있을 때만 막는다. 그런 번들이 있으면 도구가 받아들일
      수 있는 착지는 전부 그 함수가 아직 base 와 같은 지점이라, 조언이 가리키는 곳에 얻을
      것이 없기 때문이다. 새 `_base_shaped_bundles` 가 blob 등식으로 가르고(신원이 아니다),
      `check` 가 재고 `main` 이 읽는다. **거절할 정상 입력을 먼저 쟀다**: 활성 전수에서
      11건이 stale 로 보이고 그중 9건만 base 를 기술한다(a066·a071 은 조언을 유지).
      구현 뒤 실측이 예측과 일치한다 — `SUPPRESSED: 9 KEPT_ADVICE: 15`.
      판정 불변은 구조(반환 14개 동일)가 아니라 실측으로 확인했다 —
      HEAD 사본과 A/B 대조 `IDENTICAL: 27 DIFFERENT: 0`. 변이 아홉 전부 CAUGHT
      (M5 는 처음 SURVIVED — 바늘이 오류 줄에 만족됐다, 6.3 과 같은 결함이 새 시험에서 재발).
      기록은 review.md `## VERIFY — task 7.1`. **명령 자체는 안 고쳤다 — C1 은 7.2 로 간다.**

- [x] 7.2 **P0 — 착지를 무엇에 묶을지 다시 정한다 (C1~C4 · H1 · H2 · H7).** **닫음(2026-09-14)** — C3 · H1 · H2 는 7.2.1,
      C1 · C2 는 7.2.2, C4 는 7.2.3, H7 은 7.2.4. 이 결정이 새로 연 사람 결정은 7.2.5(H3) · 6.6(H2) · 6.1.2.6. 6.1.2 의 하한
      ("증거가 역사에 들어온 지점, 저자는 그 아래를 못 고른다")은 선형 역사 · 정규 파일 · 병합
      없음에서만 참이고, 증거가 작업보다 먼저 들어오면 하한이 작업 앞에 온다. 드러난 길:
      (C1) stale FLM-first 번들 → 도구가 편집 전 커밋을 계산 · (C2) 곁가지 증거 병합, 병렬 가지
      (`-1` 이 committer date 로 고른다), 병합 안에서만 바뀐 증거 · (C3) 하한은 경로를, 판정은
      워킹트리 내용을 본다(심링크 `ast.json` · 자리표시 커밋 뒤 로컬 교체 · 대상 미커밋) ·
      (C4) 빌리는 쪽이 빌려주는 쪽의 가장 낮은 값을 복사해 자기 뒤 작업을 창 밖에 둔다 ·
      (H1) `--diff-filter=MA` 가 T · R<100 을 안 센다 · (H2) 사용자 `log.follow` 가 하한을 바꾼다 ·
      (H7) 병합 커밋에서 처음 들어온 정상 증거가 거절된다.
      리뷰가 낸 후보(**아직 규칙이 아니다**): 착지에서의 번들 blob 이 HEAD 의 것과 같다(정규 파일
      모드 포함) · 고정 번들의 함수가 base..착지에서 실제로 바뀌었다 · first-parent 에 묶는다 ·
      git 설정을 지운 호출. 규칙을 지어내지 않는다 — 각 후보가 활성 13건 · a099 · 아카이브에서
      **무엇을 거부하는지** 잰 뒤 정한다(6.1 의 순서). 6.6 · 6.1.2.6 은 이 결정 전에 닫을 수 없다.
      **측정 완료(2026-09-12) · 결정 대기.** 기록은 review.md `## MEASURE — task 7.2 후보 실측`.
      계측기는 생산 `compute_landing` 과 대조해 `AGREE 13 DIFFER 0`, 양성 대조군 통과. 번들을
      가진 94건 중 93건 측정, 오늘 착지를 얻는 것이 76건. 거부 실측: K1(번들 blob = 워킹트리,
      정규 파일 모드) **0** · K4(설정 지운 호출) **0** · H1(`MAT`·`-M100%`) **0** · K2 **8**
      (a074·a075·a077·a078·a079·a091·a092·a094) · K2-all **25** · K3 first-parent **26**(이동 19) ·
      K6(번들이 HEAD 를 기술) **69**. 구멍 실측: V1 은 K2 만, V2 는 K2·K3 만, C3 는 K1(워킹트리
      등식·모드)만 닫는다. **핵심**: V1 의 세탁과 a074 의 정당한 빈 창은 모든 내용 축에서 같은
      모양이라(둘 다 base..착지에서 안 바뀐 상태를 기술) K2 가 V1 을 닫으면 a122 가 존재하는
      이유인 change 들이 착지를 잃는다. 착지를 **증거 내용**에 묶는 축은 소진됐다 — 규칙 선택은
      사람 몫이고, 공짜인 셋(K1·K4·H1)만 지금 바로 취할 수 있다.
      **사람이 2026-09-12 에 "공짜 셋만 지금"을 골랐다 → 7.2.1 로 닫았다.** 남은 것은
      **C1 · C2 · C4 · H7** 이고, 넷 다 증거 **내용**으로는 안 갈린다(위 측정의 결론).
      **→ 사람이 2026-09-14 에 결정했다: C1 · C2 는 K2 로 막고(7.2.2, 대가 8건), C4 는 빌리는 change 가 좁히지 못하게
      (7.2.3), H7 은 한계로 기록(7.2.4).**
      **6.5 측정이 더한 사실(2026-09-14, 결정 아님)**: 기계적인 증거 재작성이 하한을 작업 뒤로 올린다. `3a2bc148`(ast.json
      `file` 을 절대→상대로, `source_sha256` 불변)이 console-click-approval · verify-us-market 의 번들 **전부**를 고쳐 하한이 됐고,
      두 change 의 증거는 그 아래 커밋을 전부 맞게 기술한다. 7.2.1 의 바이트 등식도 재작성 전 커밋을 막으므로 두 자물쇠가 같이
      잠근다. 무엇에 묶을지를 정할 때 같이 볼 입력이다.
      다음 축은 값을 만드는 주체를 다시 옮기는 것뿐이다 — 도구가 계산하되 사람이 그 값과
      근거를 review.md 에 적어야 받아들이는 형태(7.3 의 H3 와 같은 자리). 이 결정 전에는
      **6.6 · 6.1.2.6 이 닫히지 않고 4.5 도 계속 막힌다.**
- [x] 7.2.2 **사람이 2026-09-14 에 "기계로 막는다"(K2)를 골랐다 — C1 · C2 를 닫는다.** 착지는 그 change 의 고정
      번들 소스 중 **하나 이상**이 base 와 달라진 커밋이어야 한다. 선택지는 셋이었다(1 기계로 막는다 · 2 사람의 확인을
      조건으로 받는다 · 3 한계로 인정한다). 2 는 저자가 쓴 확인으로 저자의 선택을 검증하는 순환이고
      ([[author-supplied-evidence-cannot-pin-an-author-choice]]), 3 은 FLM-first 기본 경로에서 조용한 초록을 남긴다.
      **대가를 알고 골랐다**: 7.2 측정에서 착지를 얻는 76건 중 8건(활성 a074 · a077 · a079 · a091 · a092 · a094, 아카이브
      a075 · a078)이 착지를 잃는다 — base 가 작업 뒤로 옮겨진 change 들이고, 증거 내용으로는 V1 의 세탁과 같은 모양이다.
      그 change 들은 넓은 창으로 돌아간다. 도구에 면제 경로는 없고 a075 · a076 이 쓴 사람의 면제 기록이 선례다.
      **닫음(2026-09-14).** 기록은 review.md `## DECIDE — task 7.2.2` · `## Pre-Edit Gate — task 7.2.2` · `## VERIFY — task 7.2.2`.
      **편집 전에 사본으로 전수를 다시 쟀다**: 93건 같음 85 · 상실 **8**(예측과 같은 이름) · 이동 0 · 획득 0. 기존 시험 중 깨지는 것은
      둘이고 둘 다 a074 모양을 주석으로 적은 시험이었다. 규칙은 `_landing_refusal` **맨 뒤** 한 조건(선언 · 계산 두 경로가 같이
      묻는다)이고, 계산 실패의 머리 "matches every pinning bundle" 을 "is accepted as the landing" 으로 고쳤다 — K2 가 증거와 맞는데
      거절하는 후보를 만들어 옛 머리가 거짓이 된다. 리뷰 재현 V1 · V2 **둘 다 rc 1**, 5단계 required 1 로 빨강. 변이 K-A~K-G **일곱
      CAUGHT**(K-B · K-G 는 "하나 이상"을 재는 픽스처가 없어 살았고, 첫 보충 시험은 픽스처가 base 에 파일을 안 넣어 **도달 못 해**
      또 살았다 — 도달 단언을 세운 뒤 잡힘) · 옛 하네스 7.6 · 7.7 · 7.8 · 6.5 결과 불변(옮긴 자리 R14' · R20' CAUGHT) · 실물 A/B 착지 93:
      같음 68 · 머리만 17 · K2 상실 8/8 · 그 밖 0, 판정 a099 + 활성 6 **IDENTICAL** · 시험 143 → **148** · `make sdd-test`(logic-map 223)
      · `lint` · `test-seams` · `validate --all` 58/58. **안 닫은 것**: H2(기록 뒤 작업, 6.6) · H3(섞인 V1 — 막는 변형은 상실 14, 사람 결정).
- [x] 7.2.3 **빌리는 change 는 창을 좁히지 않는다 (C4, 같은 결정).** 1.8 의 "창의 양쪽 끝을 공유한다"를 뒤집는다 —
      빌려주는 쪽의 착지는 빌려주는 쪽 증거의 가장 낮은 값이라, 빌리는 쪽이 복사하면 그 뒤의 빌리는 쪽 Go 작업이 창
      밖으로 나간다. base 공유는 그대로 둔다. 거부할 정상 입력을 먼저 잰다(1.8 시점 빌리는 change 는 a073 하나).
      **닫음(2026-09-14).** 기록은 review.md `## Pre-Edit Gate — task 7.2.3` · `## VERIFY — task 7.2.3`. 빌리는 change 전수 **1**(a073,
      기록 없음 — 빌려주는 a072 도 없음). `check` 는 빌리는 쪽에 기록이 **있는가**만 묻고 있으면 `BORROWED_REFUSES_A_LANDING` 으로
      거절하며, 빌려주는 쪽 기록은 이 창에 안 쓴다. 기록 명령도 같은 문장. 핵심 시험은 C4 그 자체 — 빌려주는 쪽만 L 을 기록하고
      빌리는 쪽 작업이 L2 에 있으면 빌리는 쪽 창은 워킹트리이고 L2 의 작업이 요구된다(양성 대조 포함). 변이 C-A~C-E · R5' **첫 판
      전부 CAUGHT**(C4 구멍을 되살리는 C-E 는 그 시험 하나가 잡는다) · 옛 하네스 전부 재실행 결과 불변(7.7 하네스가 R10 에서 멈춘 것을
      R11~R22 필터 실행과 R10' 로 메웠다) · 실물 A/B a073 · a072 판정 **IDENTICAL**(오류 336 · required 388) · 시험 148 → **147** ·
      `make sdd-test`(logic-map 222) · `lint` · `test-seams` · `validate --all` 58/58.
- [x] 7.2.4 **H7 은 알려진 한계로 기록한다 (같은 결정).** 병합을 마치며 처음 커밋한 정상 증거가 거절되는 것은
      막는 쪽으로 틀리는 문제다. 지금 코드에서 재현되는지부터 재고 spec 에 한계로 적는다.
      **닫음(2026-09-14).** 기록은 review.md `## Pre-Edit Gate — task 7.2.4` · `## VERIFY — task 7.2.4`. **재현된다**(HEAD `f9811236`):
      하한 `''`, 기록 명령 rc 1, 기록 없는 5단계는 `[]` · required 1(막히지 않는다 — 좁히지 못할 뿐). 그런데 사유 문장 둘이
      "never entered this history (commit the bundles)" 로 **거짓**이었다(번들은 커밋돼 있다) — 조언 줄도 그 문장을 인용해 이미 한 커밋을
      또 하라고 권했다. 동작은 두고 **문장만** 한계를 사실대로 말하게 고쳤다("no ordinary commit … adds …, a merge commit's own changes are
      not read"). spec 에 한계 문단과 시나리오. 한계를 없애는 변이(H-C `git log -m`)를 시험 넷이 잡는다. 변이 H-A~H-C 와 옮긴 자리
      R4' · R18' · K-D' **전부 CAUGHT**, 옛 하네스 전부 재실행 결과 불변 · 실물 해당 id 0 · 시험 147 → **150** · `make sdd-test`(logic-map
      225) · `lint` · `test-seams` · `validate --all` 58/58.
- [x] 7.2.5 **H3 — 섞인 V1 을 막을지 사람이 정한다 (7.2.2 가 연 것).** FLM 을 여럿 먼저 커밋하고 일부만 갱신한 뒤 나머지 파일을
      나중에 고치면 K2(하나 이상)를 통과해 초록이다(`722_holes.py` H3). 막는 변형 — 착지에서 base 와 같은 고정 소스가 착지 **뒤에**
      바뀌면 거절 — 은 편집 전 전수에서 상실이 8 → **14**(+a100 활성 · a047 · a049 · a050 · a052 · a060 아카이브). 증거 내용으로는
      "이웃이 나중에 같은 파일을 고친 정상 change" 와 안 갈린다. 규칙을 지어내지 않는다.
      **→ 7.2.6(H4) 로 흡수 후보.** H4 를 막는 규칙은 이 변형이 거절하는 것을 전부 거절한다. 7.2.6 전수 뒤에 같이 정한다.
      **닫음(2026-09-16) — 7.2.6 의 B-ANYGO 로 흡수.** 사람이 고른 규칙은 내용(`base 와 같은 고정 소스`)이 아니라
      "같은 커밋이 이 change 의 디렉터리도 만졌는가"를 본다. 그래서 H3 의 **한 커밋 모양**(나머지 파일 수리가 그 change 의
      문서와 같은 커밋에 있는 경우)은 막히고, 쪼갠 커밋 모양은 spec 의 알려진 한계로 남는다. 편집 전 전수에서 상실 14 를
      부르던 내용 변형은 채택하지 않았다 — 원인이 거의 전부 이웃이었다(7.2.6 전수: FILE 61 · FUNC 55/68).
- [x] 7.2.6 **H4 — 착지 기록 뒤의 리뷰 수리가 고정 파일을 고치면 초록이다 (2026-09-16 실측, 사람 결정).** 7.2.5 · 6.6 의 추천을 쓰려고
      잰 픽스처에서 나왔다(scratchpad `rec_holes.py` · `rec_holes2.py`, 도구는 HEAD `0323e0ea` 사본, 5단계 `check` 만 잼). 정직하게
      기록한 뒤 **같은 고정 파일**을 고치고 FLM 을 안 고친 커밋: 기록 없음 → `AST source hash is stale` 빨강, 기록 있음 → `[]` ·
      required 1 **초록**. 성실한 쪽은 반대다 — 번들을 갱신하고 재기록을 안 하면 `landing point … is not the revision this evidence
      describes` 빨강인데 문장이 복구 경로를 안 말하고, `--record-landing` 은 "already exists — not overwritten" 으로 거절한다. 기록을
      지우고 다시 기록하면 착지가 갱신 커밋으로 옮겨 가 초록(정상). H2 · H3 보다 흔한 모양이다 — 로트마다 리뷰 수리 라운드를 돈다.
      후보 규칙: 고정 소스가 착지와 '지금' 에서 같아야 한다(파일 바이트 FILE · 함수 본문 FUNC). 7.2 측정의 K6(번들이 HEAD 를
      기술, 거부 **69**)이 FILE · now=HEAD 와 같은 축이다. 규칙을 지어내지 않는다 — 전수로 잰 뒤 사람이 정한다.
      **측정 완료(2026-09-16) · 결정 대기.** 기록은 review.md `## MEASURE — task 7.2.6`. 내용 규칙은 불가 — 착지 있는 68건 중
      FILE 61 · FUNC 55 거절(now=HEAD), 아카이브 시점 45 · 41, 원인은 거의 전부 이웃 커밋. 가르는 신호는 **같은 커밋이 그 change
      디렉터리를 만졌는가** 하나였다: 변형 B(착지 뒤에 Go 와 자기 디렉터리를 같이 만진 비병합 커밋이 있으면 그 후보를 안 받는다)는
      아카이브 시점 상실 **4**(실물 H4 셋 a083 · a084 · apply-us-measurement-fixes + 오탐 add-candidate-discovery), HEAD 에서 +3
      (a066 · a071 · a044 — 넓은 교차 커밋), ANYGO 는 이동 2(판정 변화 0). 오늘 판정 영향 0(기록 실물 a099 는 같음). 커밋을 쪼개면
      뚫리는 망각 가드이고, 거절에만 쓰므로 창을 좁히는 데 못 쓴다. ANYGO 는 한 커밋 모양의 H2 · H3 까지 덮는다.
      **닫음 — 사람이 2026-09-16 에 B-ANYGO 를 골랐다(옵션 1).** 기록은 review.md `## Pre-Edit Gate — task 7.2.6` ·
      `## VERIFY — task 7.2.6`. 새 `_self_repair_commits` 가 깃발 집합을 **한 곳**에서 만들고(디렉터리 + Go, 비병합,
      오래된 것부터), `_landing_refusal` 이 **맨 뒤에** 그 조건을 세우며, 두 호출자가 `floor` 와 같은 모양으로 한 번
      재서 넘긴다. 거절과 `--record-landing` 의 "이미 있다"가 **같은 상수**(`LANDING_RECOVERY`)로 복구 경로를 말한다.
      생산 판본이 사람이 비용을 본 그 규칙인지 먼저 쟀다 — 착지 있는 68건 전수에서 깃발 **224 · 불일치 0 · 순서 어긋남 0**.
      픽스처 실측: 고정 파일 수리와 다른 Go 수리는 거절, 이웃 편집 · 문서만 만진 커밋 · 쪼갠 커밋은 계속 초록.
      **독립 적대 리뷰에서 P0 하나가 나와 고쳤다(2026-09-16).** 경로 조회의 기본 역사 단순화가 곁가지의 자기 수리를
      통째로 버려, 이 change 의 제목이 가리키는 상황(병합)에서 H4 가 그대로 열려 있었다 — 직접 재현했다. 내 A/B 가
      **둘째 단계만** 비교해서 원리적으로 못 잡는 측정이었다([[falsification-must-vary-the-right-axis]]).
      `--full-history` 로 고쳤고 비용은 126건 전수에서 깃발 386 · 386 · 다른 change **0**. 같은 리뷰에서 일곱 건 더:
      git 실패가 가드를 조용히 끄던 것(→ 판정), 계산값 대조 거절만 빠뜨린 복구 경로, `diff.renames` · `core.quotePath`
      가 판정을 기계 설정의 함수로 만들던 것, 빌린 경로의 잠복 결합(spy 시험으로 못 박음), 공허한 단언과 어긋난 시험
      이름, 후보마다 `merge-base` 를 수리 개수만큼 돌던 것(→ `_repairs_after`).
      실물 A/B **126건 전수: 판정 다름 0** · 계산값 다름 9(상실 7 · 이동 2) · 기록 실물 a099 불변(required 32).
      시험 150 → **170**(새 20, 편집 전 도구에서 16 빨강) · 변이 **18개 중 17 CAUGHT**, S3(`--no-merges` 제거)는
      git 기본값 때문에 **동등 변이**임을 git 2.43 에서 재서 적었다. 비용은 한가한 기계에서 신호 0.02~0.04초,
      `check` 전체 차이는 잡음 안이다(7.5). 한계 셋(쪼갠 커밋 · 병합 커밋 자신의 수리 · 이웃의 나중 편집)은
      spec 에 적고 시험으로 못 박았다 — 놓치는 쪽이라 창이 넓어지지 않는다는 것도 spec 이 구분해 적는다.
- [x] 7.2.1 **거부 0 으로 실측된 셋을 규칙으로 넣는다 (C3 · H1 · H2).** 기록은 review.md
      `## DECIDE — task 7.2 규칙 선택`.
      (1) **판정이 읽은 번들을 착지 커밋이 들고 있어야 한다.** 새 `_unheld_bundles` 가
      `resolve_landing` 과 `compute_landing` **양쪽**에 선다 — 한 집에만 두면 도구가 자기가
      거절할 값을 계산한다([[two-judgements-cover-for-each-other]]). 등식은 HEAD 가 아니라
      **워킹트리**와 세운다: 판정이 읽는 것이 워킹트리이므로, HEAD 로 바꾸면 자리표시자
      교체가 안 보인다(뮤테이션 M-F 로 실측).
      (2) **옮기면서 고친 증거는 하한을 올린다** — `-M100%` · `--diff-filter=MAT`. 그대로
      옮긴 아카이브 이동은 계속 건너뛴다(a099 의 실물 기록이 그 조건에 달려 있다).
      (3) **하한은 개인 git 설정의 함수가 아니다** — `--no-follow`.
      **거절할 정상 입력을 먼저 쟀다**([[fail-closed-must-name-what-it-rejects]]): 등식 범위를
      번들 산문까지 넓히면 76건 중 3건이 착지를 잃어서(a092·a043·a096, 전부 이웃 change 가
      나중에 단 무효화 배너) `ast.json` 하나로 묶었다. 측정에 있던 모드 절과 `-c` 두 개는
      다른 절이 이미 덮어서 뺐다(M-C3 SURVIVED). 등식은 하한 **뒤**에 세웠다 — 앞에 두면
      기존 시험 둘의 거절 지점을 가로채서 그 가드가 못 박히지 않는다.
      뮤테이션 M-A~M-G **일곱 전부 CAUGHT**(첫 판에서 M-C·M-E 가 SURVIVED — 중복 절과
      계산 경로의 빈칸이었고 둘 다 없앴다).
      **실측**: 착지 전수 93건 — 같음 76 · 이동 0 · **상실 0** (양성 대조군 12/12), 판정 A/B
      28건 **IDENTICAL 28 · DIFFERENT 0**, 실물 기록 a099 는 착지 `e6c4636a` · required 32 ·
      8.5s→8.0s 로 불변. `make sdd-test` · 스위트 180개 · `openspec validate --strict` 통과.
- [x] 7.3 **P1 — 게이트가 "도구가 계산한 값"을 확인할지 사람이 정한다 (H3).** spec 은 기록이
      도구의 값이어야 한다(SHALL)고 적는데 게이트는 안 본다. a099 실측: 기록 `e6c4636a` ≠ 계산
      `21a315d1` — 같게 강제하면 유일한 실물 기록이 깨진다(소급 인정 / 다시 기록). 7.2 뒤에 정한다.
      **측정 완료(2026-09-13) · 결정 대기.** 기록은 review.md `## MEASURE — task 7.3 (H3)`.
      7.2.1 뒤에도 남은 선택은 크다 — 착지를 얻는 76건 중 **65건**이 수락값을 둘 이상 갖고
      (최대 516), 그중 **44건**은 그 선택이 판정 입력을 바꾼다(최대 279가지). 그런데 방향이
      한쪽이다: 수락값이 계산값의 자손이 **아닌** change 는 **0/76** 이고, 창을 넓혔을 때 바뀐
      Go 파일이 빠지는 자리도 **0/44**, 함수 단위 표본 6건에서 빠진 함수 **0**(늘어난 함수
      0~15). **그래서 오늘 이 역사에서 이탈은 자해뿐이다** — 구조적으로 배제된 것은 아니고
      (되돌려진 편집이 있으면 넓힌 창이 느슨해질 수 있다) 그 모양이 0건일 뿐이다.
      등식을 넣으면 거부되는 실물은 a099 하나이고, 계산값으로 재기록하면 **판정이 같다**
      (오류 0 · required 32 · 지금과 동일). 비용은 `check()` 15.85s 에 `compute_landing`
      **1.33s**(+8%). 후보 R2("계산값 이상")는 실측으로 **no-op** 이라 죽었다.
      **남은 선택은 R1(등식+a099 재기록) · R3(안 넣고 spec 을 실측에 맞게 정정) ·
      R4(계산값과 다르면 review.md 에 값과 근거 — 7.2 가 가리킨 축)이고 사람 몫이다.**
      **사람이 2026-09-13 에 R1 을 골랐다 → 7.3.1 로 닫았다.** (부모 체크박스는 그때 안 닫혔다 —
      2026-09-18 에 닫는다. 결정도 구현도 7.3.1 에 있고 이 줄이 남긴 것은 표시뿐이었다.)
- [x] 7.3.1 **기록이 계산값과 같은지 게이트가 확인한다 (H3, R1).** 기록은 `resolve_landing`
      의 **맨 뒤**에서 `compute_landing` 과 대조한다 — 앞에 두면 위 가드들의 거절 지점을
      이 등식이 가로챈다([[a-new-guard-unpins-the-guards-behind-it]], 7.2.1 이 같은 파일에서
      실측한 모양). 거절은 **두 값을 다 말한다**(적힌 값 · 계산된 값).
      **유일한 실물 기록 a099 를 계산값으로 재기록했다** — `e6c4636a` → `21a315d1`.
      판정은 안 바뀐다(둘 다 오류 0 · required 32, 실측). 옛 값은 손으로 쓴 것이고
      (2026-09-09 `5c848099`), 도구가 착지를 계산하는 `--record-landing` 은 그 이틀 뒤
      6.1.2 에서 생겼다. 정정 사유는 a099 의 아카이브 review.md 에도 한 문단 남겼다.
      spec 에 "게이트가 그 값과 같은지 확인해야 한다(SHALL)"와 "이 확인은 다른 유효
      조건들 **뒤에** 서야 한다(SHALL)", 시나리오 하나를 더했다.
      시험 4개 추가 · 뮤테이션 **N-A~N-D 넷 전부 CAUGHT**(N-C 는 처음 SURVIVED — 픽스처가
      전부 하한 = 계산값이라 스위트가 규칙의 **정체**를 못 갈랐다; 증거를 먼저 올리고 코드를
      나중에 올리는 흔한 순서로 둘을 갈라 놓고 잡았다).
      **비용**: 기록이 있는 change 마다 판정 때 walk 하나가 는다 — a099 실측 `check()`
      8.9s 중 `compute_landing` 1.3s. 등식이 맨 뒤라 walk 는 기록 지점에서 멈춘다.
      **이 결정이 뒤집는 것**: task 1.3 의 "엄해지는 방향(창을 넓히는 기록)은 일부러 안
      막는다". 이제 계산값 하나만 받는다.
- [x] 7.4 **P1 — `record_landing` 견고성 (H4 · H5 · H6).** 존재 확인 뒤 비원자적 덮어쓰기와 끊긴
      심링크를 따라가는 쓰기(`execution_baseline.py:336` 은 이미 `open(…, "xb")` · lstat 로 한다) ·
      `compute_landing` 의 `ValueError` 가 traceback 으로 죽는다 · 새 `timeout=60` 호출의
      `TimeoutExpired` 가 `check()` 를 빠져나가 창 줄이 안 찍힌다.
      **닫음(2026-09-13).** 기록은 review.md `## VERIFY — task 7.4`.
      (H4) 끊긴 심링크는 **실측으로 재현**했다 — `record_landing` 이 rc 0 으로 "recorded" 라고
      말하면서 저장소 **밖** 임시 디렉터리에 착지 값을 썼다. 링크 거절 + `open(…, "xb")`
      배타 생성으로 닫았고, 배타 생성이 존재 확인과 쓰기 사이의 창(리뷰 I1: walk 133.7s·219.6s)도
      같이 닫는다. (H5·H6) 결함을 **이름으로** 말하게 했다 — 목록은 새 `GATE_FAULTS` **한 곳**에
      살고 `subprocess.SubprocessError` 가 들어간다(`TimeoutExpired` 는 `OSError` 가 아니라서
      자리마다 적혀 있던 넷을 그대로 빠져나갔다). 가드는 자리마다가 아니라 **경계**에 세웠다 —
      이 도구를 부르는 생산 자리는 `tools/gate.sh:321` 의 CLI 하나뿐이고, 자리마다 목록을 베끼면
      새로 부르는 git 하나가 어느 목록에도 안 걸려 다시 스택이 된다.
      착지를 **잰 뒤** 터진 결함이면 창 줄은 그대로 찍힌다(시험이 단언한다).
      시험 6개 추가(112개 초록) · 뮤테이션 **M-A~M-G 일곱 전부 CAUGHT**(M-A 는 처음 SURVIVED —
      배타 생성이 링크 거절을 대신 막고 있었다, [[surviving-mutant-may-mean-accidental-safety]];
      거절 **지점**을 단언하게 고쳐서 잡았다).
- [x] 7.5 **P2 — 성능 (I1).** 못 찾는 walk 가 (후보+1)×(번들+2) spawn — a055 133.7s, 다른 아카이브
      219.6s. 파일 단위 fetch · 조기 종료 · 번들 목록 한 번 · `--ancestry-path`. 7.2 가 walk 를
      바꿀 수 있으므로 그 뒤에 한다.
      **닫음(2026-09-18). 생산 코드 변경 0.** 기록은 review.md `## MEASURE · Pre-Edit Gate — task 7.5` ·
      `## VERIFY — task 7.5`. **먼저 다시 쟀다**(리뷰 숫자는 리뷰 시점의 수다): 전수 126건 중 93이
      걷고 후보 합계 36,767, 그리고 비용은 **실패하는 walk** 에 있었다(a071 73.19s · a092 72.32s 대
      성공 1.64s · 0.78s). spawn 귀속으로 갈라 보니 **97.1~97.3% 가 blob fetch** 였다 — 근본 원인은
      알고리즘이 아니라 **fetch 단위**(`git show` 한 프로세스에 파일 하나)였다.
      고친 것 셋: 새 `_committed_many` 가 `git cat-file --batch -Z` **한 번**으로 읽고(`-Z` 는 개행 든
      경로와 tree 의 raw NUL 을 둘 다 견딘다, blob 만 내용으로 친다) · `_committed_bytes` 가 그 위의
      1-원소 호출이 되어 철자가 한 곳이 되고 · 번들 목록을 `floor`·`repairs` 와 같은 방식으로
      **한 번 재서 넘긴다**. `normalized_source` 의 `root.resolve()` 중복(호출당 2회)도 없앴다.
      첫 병목을 걷어내자 둘째가 프로파일에 드러나 셋을 차례로 쟀다: 73.19 → 27.84 → 10.97 → **3.69s**.
      **판정 전수 A/B: 비교 116 · SAME 116 · DIFFERENT 0**(예외는 타입·문장까지), 합계
      **2023.0s → 162.1s (12.5×)**, 최대 36.9×. 편집 전후 AST 대조로 `_landing_refusal`(11·9) 등
      **여섯 함수의 분기·반환·raise 가 전부 불변** — 가드도 순서도 안 움직였다.
      뮤테이션 18 중 **17 CAUGHT**. 첫 판 SURVIVED 넷은 전부 "시험이 그 갈래에 안 닿았다"였다
      (트리를 혼자 물음 · 빈 출력으로 죽는 git · 아카이브 뒤 커밋 없음 · 행동이 같은 두 철자).
      닿는 픽스처와 구조 시험으로 셋을 잡고, 끝까지 남은 T6 은 **동등 변이임을 T18(사전 채움 삭제 →
      59개 빨강)로 증명**했다 — 가설로 안 넘겼다.
      시험 175 → **182** · `make sdd-test`(logic-map 257) · `make lint` · `openspec validate` 58/58 ·
      실물 a122 rc=0.
      **안 한 것**: `--ancestry-path` 는 후보 **집합**을 바꾸는 규칙 변경이라 7.5 의 몫이 아니고
      (7.8 M6 이 곁가지 후보의 실재를 실측했다), 조기 종료는 거절 문장이 고칠 자리를 **전부**
      대야 해서 안 하며 배치 뒤엔 비용도 0 이다. 둘 다 review.md 에 사유를 적었다.
- [x] 7.5.1 **P0 — "못 물었다" 는 "없다" 가 아니다 (gstack 리뷰 2026-09-19).** 7.5 의 `_committed_many`
      는 rc≠0 에서 전부 `None` 을 돌려줬고, docstring 은 "부르는 쪽이 `None` 을 불일치로 센다" 고
      **전칭으로** 적었다. gstack `/review` 의 두 출처가 그 전칭을 두 자리에서 깼고 **둘 다 재현했다**:
      가드 7 은 한쪽만 실패한 `None != bytes` 를 "바뀌었다" 로 읽어 **편집 전 커밋을 착지로 기록했고**,
      `_landing_record` 는 `None` 을 "기록 없음" 으로 읽어 착지 검증을 건너뛰었다. 사람이 "지금 다
      고친다" 를 골랐다. **닫음(2026-09-19). 생산 Go 변경 0.** 기록은 review.md
      `## MEASURE · Pre-Edit Gate — task 7.5.1` · `## VERIFY — task 7.5.1`.
      근본 수리: rc≠0 을 결함으로(git 의 stderr 를 담아서) — `None` 은 이제 "git 이 물은 spec 그대로
      `missing` 이라 답했다" 하나뿐이다. 파서는 git 이 실제로 내는 두 모양만 받는다(실측으로 정했다,
      gitlink 포함). 곁들여: 입력을 **한 벌** 재서 선언·계산 경로가 공유(F2) · 수락 직전 증거 지문
      재확인(Codex P1-2 — 7.5 가 목록을 얼려서 생긴 회귀) · `validate_target`·가드 7 배치(F3) ·
      메시지 넷(F4~F7) · `GIT_BATCH_MINIMUM` 에 git 릴리스 노트 영수증(F9) · 손 복사 예외 목록
      넷 → `GATE_FAULTS`.
      **1.4 의 되돌림을 정정했다** — "시험 21개가 저장소 아닌 곳에서 돈다" 는 git 을 mock 하려다
      `_landing_record` 만 빠뜨린 픽스처였다. 빈 저장소로 바꾸니 전제는 그대로였다(`missing` rc 0).
      증거: `check()` **전체** A/B **126/126 SAME** · `compute_landing` 116/116 · 편집 전 AST 를
      revision `b29e1f4e` 로 **명시**해 뽑음 · `_landing_refusal` 가드 순서 불변(바뀐 것은 가드 7 조건 한 줄) ·
      변이 **31/31 CAUGHT**(첫 판 생존 V10·V13 은 양성 대조로 **닿았음**을 확인 → 시험이 그 갈래에 닿게
      고쳐 잡음) · 실물 a099 blob 프로세스 46 → 10 · 시험 187 → **203** · `sdd-test`(278) · `lint` ·
      `test-seams` · `validate` 58/58 · a122 rc=0.
      **남은 창 하나**를 기록했다: 수락 직전 재확인 뒤 `validate_target` 이 디스크를 다시 읽기까지.
- [x] 7.5.2 **P1 — 수리한 트리의 재리뷰가 연 것 (gstack 리뷰 2026-09-19, 두 번째).** 로트
      `a5c4bc77..e9f905bd` 재리뷰(Claude 전문가 다섯 · 적대 서브에이전트 · 레드팀 · Codex). 지난 수리
      다섯 중 넷은 확인됐고 **P1-2 는 반쯤이었다**: 지문과 판정 목록이 **따로 읽혀서**, 한 `ast.json`
      읽기가 잠깐 실패하면 그 번들이 판정에서 빠진 채 두 지문이 같다고 답한다(Codex 재현). 그리고
      7.5.1 이 실패 계약을 바꾸자 `test_placeholders_are_rejected` 가 **git 결함 줄로 통과**하게 됐다
      (자리표시자 가드를 지워도 초록 — 변이로 증명). 사람이 "이름 바뀐 파일 규칙(아래 7.5.3)을 뺀
      나머지 전부" 를 골랐다. 범위: 한 번 읽기(지문 = 판정 목록 = HEAD) · 수락 뒤 창(판정이 읽는
      `ast.json` 을 착지가 판정한 바이트에 묶음) · 이관 change 의 미리 읽기가 대상 이름을 잃는 회귀 ·
      조언 계산의 결함이 판정을 지우지 않게 · 선언 경로의 거절이 움직인 입력 위에서 복구 조언을 하지
      않게 · 새 git 의 `<oid> submodule` 답 · 모양이 틀린 `ast.json` 의 traceback · 시험 공백
      (지문 재작성 · 판정 경로 · 경계 `>=` · fallback · 거절 문장) · 문서·주석·메시지.
      **닫음(2026-09-19). 생산 Go 변경 0.** 기록은 review.md `## MEASURE · Pre-Edit Gate — task 7.5.2` ·
      `## VERIFY — task 7.5.2`. 근본 수리 한 줄: **증거를 읽는 자리를 `_read_evidence` 하나로** — 착지의
      지문(바이트 해시 + 고정 목록 + `HEAD`)과 판정 목록이 같은 읽기에서 나오고, walk 는 잰 바이트(`held`)만
      보고, `check()` 는 자기 읽기를 그 지문과 대조한 뒤 그 바이트만 쓴다(7.5.1 이 "설계 변경" 이라며 남긴
      수락 뒤의 창이 닫혔다). 곁들여: 이관 change 의 대상 이름 회귀 · 조언 결함이 판정을 지우지 않게 · 선언
      경로 거절 앞 재확인 · `<oid> submodule` · 버전 조언은 rc 129 에서만 · 모양 틀린 `ast.json` 은 판정 줄 ·
      수리 신호 `None`(잰 적 없음) · `LandingInputs`/`Fingerprint` 를 `NamedTuple` 로 · base 조언 한 프로세스 ·
      공허해진 `test_placeholders_are_rejected` 를 진짜로.
      증거: `check()` 전체 A/B **126/126 SAME**(기준 `e9f905bd`) · 변이 **52/52 CAUGHT**(스위트 전체, 첫 판
      생존 W23 은 안 닿음 → 시험 추가) · `_landing_refusal` 가드 순서열 차이 0 · a099 `ast.json` 읽기 446 → 114 ·
      시험 203 → **231** · `sdd-test` · `lint` · `test-seams` · `validate` 58/58 · a122 rc=0.
      **순서 이탈 하나를 적었다**: 넷(`_unheld_bundles` · `_bundle_text` · `_landing_refusal` · `_recording_refusal`)은
      GREEN 도중 편집 집합에 들어와 편집 전 AST 를 **편집 뒤에** revision 에서 뽑았다.
      **안 한 것**: 가드 7 의 base 반복 읽기(호출 사이 상태를 늘린다, ~1.2s) · 7.5.3(사람 결정).
      **정정(2026-09-20, 7.5.2.1)**: 위 "수락 뒤의 창이 닫혔다" 와 "선언값 ≠ 계산값 갈래는 도달 불가" 는
      **거짓이었다** — 재리뷰가 셋을 재현했다(아래 7.5.2.1). 시간 합계 2048.9s → 1934.5s 도 대부분 실행 순서
      표류였다. 판정 A/B 126/126 SAME 은 그대로 맞다.
- [x] 7.5.2.1 **P0 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰 2026-09-19, 출처 여덟).**
      7.5.2 는 `ast.json` 바이트만 묶고 "묶었다" 고 적었다. 재리뷰가 재현한 것: `HEAD` 는 지문에 **표본**만
      있고 뒤의 git 호출은 살아 있는 `HEAD` 를 읽어 가지 전환 한 번으로 rc 0 · 착지 기록은 그 표본 **전에**
      읽힘 · `analysis.exists()` 조기 반환이 묶기 앞 · `_bundle_text` 의 살아 있는 목록이 든 `ast.json` 을
      뺀다(Codex, fc35eb2d `[]`) · git 결함 넷이 거절 사유가 되어 "기록을 지워라" 로 · 모양 틀린 `ast.json`
      열 가지가 아직 traceback · 중심 주장의 변이 X1~X4 생존 · 검색만 되는 번들 디렉터리 · 심링크 고리 ·
      서로게이트 출력 · 앞선 실행의 사실 절반만 지움 · 시험 전용 `None` 기본값 · 하네스 `75_census.py` 고장.
      사람이 "전부, 같은 절차로" 를 골랐다(7.5.3 은 여전히 사람 결정). 이번엔 **입력 목록을 AST 로 먼저
      셌다**(`analysis/harness/7521_inputs.py`) — 기록은 review.md `## MEASURE · Pre-Edit Gate — task 7.5.2.1`.
      **닫음(2026-09-20). 생산 Go 변경 0.** 기록은 review.md `## VERIFY — task 7.5.2.1`. 근본 수리 한 줄: **명령이
      시작할 때 `HEAD` 를 sha 로 한 번 풀고 증거 디렉터리를 한 번 읽어(`Evidence` — 있었나 · 번들 목록 · 바이트),
      착지 판정 · 대상 판정 · 조언이 그 두 값만 쓴다** — 같은 값을 넘기니 대조할 것이 없고, 수락 · 거절 앞 재확인은
      대조만 한다. 곁들여: git 결함 넷(조상 · 하한 · 순회 · 깨끗한가)을 결함으로 · 모양 검사를 파싱하는 한 곳
      (`_parse_ast`)에 · 번들 파일 목록은 `iterdir`(못 열면 이름 댄 줄)이고 `ast.json` 은 목록과 무관 · 심링크 풀이를
      판본과 무관한 `realpath` 로(3.12 만 raise — 실측) · 출력 `backslashreplace` · 앞선 실행의 사실 전부 지움
      (`RUN_FACTS`) · 시험 전용 `None` 기본값 전부 필수로 · 지문(`Fingerprint`/`_digests`/`_evidence_fingerprint`) ·
      `_pinning_bundles` 삭제 · 하네스 여섯을 새 API 로 고치고 **전부 돌렸다**(`75_census` 는 다시 걷는다).
      증거: `main()` 출력 전체 A/B **126/126 SAME**(순서 번갈아) · `ast.json` 읽기 9,072 → 5,803 · 변이 **76/76 CAUGHT** ·
      상징 `HEAD` 읽기 10 → 1 · 시험 231 → **254** · `lint` · `sdd-test` · `test-seams` · `validate` 58/58 · a122 rc=0.
      **Pre-Edit 에서 바꾼 것 둘**(review 에 적음): 심링크 고리(예외 변환 → `realpath`) · 태어나지 않은 `HEAD` 가 mock
      픽스처 22 개에 있었다(픽스처에 빈 커밋, 가드 그대로). **시간 이득은 적지 않는다**(먼저 도는 쪽이 ~3% 느리다).
      **정정(2026-09-20, 7.5.2.2)**: "조언도 그 한 번 읽기" 는 거짓이었다 — `main` 의 기록 조언이 증거를 다시 읽는다
      (다음 명령을 예측하는 줄이라 그대로 둔다, 판정 줄은 그 읽기에 안 기댄다). "시험 전용 `None` 기본값 전부" 는
      "증거 · 착지 입력의" 로 좁힌다(`validate_target(index=None)` 는 캐시라 남았다). "실행 중 커밋은 판정을 멈추지
      않는다" 는 틀린 교환이었고 7.5.2.2 가 되돌린다. 그리고 `_bundle_text` 를 다시 쓰며 `is_file()` 거름을 빠뜨려
      FIFO 에 게이트가 멎는 회귀를 만들었다(재리뷰 출처 넷 재현) — 7.5.2.2.
- [x] 7.5.2.2 **P0 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰 2026-09-20, 출처 아홉).** 지난 수리는
      확인됐다(레드팀 주장별 변이 15 전부 죽음). 새로 재현된 것: 판정이 스냅숏으로만 서서 반환 전에 디스크와 안
      대조한다 — Go 변경을 찾는 동안 증거를 무효로 고쳐도 `[]`(Codex P1, 나 재현), 실행 중 `HEAD` 가 움직여도 옛
      `HEAD` 의 판정이 새 트리의 PASS 가 된다 · `_bundle_text` 의 FIFO 멎음 / `/dev/zero` traceback(내 회귀, 출처 넷) ·
      id 해소 실패 앞에서 사실이 안 지워짐 · 깊은 JSON 의 `RecursionError` · 빈 계산값 문장을 지운 것 · 시험 공백
      (못 푸는 바이트 · 명령 수준의 태어나지 않은 `HEAD` · 빌림/이관 갈래의 `HEAD`) · 주석 · 기록의 과장. 사람이
      "1 — 끝의 재확인은 `HEAD` 와 증거 둘 다, 판정한 sha 를 창 줄에" 를 골랐다(7.5.2.1 의 "실행 중 커밋은 안 멈춘다" 를
      되돌린다). 기록은 review.md `## MEASURE · Pre-Edit Gate — task 7.5.2.2`.
      **닫음(2026-09-20). 생산 Go 변경 0.** 기록은 review.md `## VERIFY — task 7.5.2.2`. 근본 수리 한 줄:
      **명령이 판정을 내놓기(기록을 쓰기) 직전에 판정한 `HEAD` 와 증거가 아직 그대로인지 다시 보고, 달라졌으면
      판정 대신 "다시 돌려라" 를 낸다** — 스냅숏은 판정을 **내적으로** 일관되게 하고, 끝의 대조는 그 판정이 **지금
      거기 있는 것**의 판정임을 보인다. 곁들여: 창 줄 끝에 판정한 sha(`judged at HEAD <sha12>`) · 번들 파일은
      `_read_regular` 한 자리로만(`O_NONBLOCK` + `fstat`, FIFO · 장치는 건너뛰고 **폴더 · 소켓은 이름 댄 줄**) ·
      사실을 **맨 앞**에서 비운다 · `RecursionError` 도 invalid · 빈 계산값이면 사유를 말한다 · git 의 말 첫 줄은
      `_first_line` 한 벌 · 기준점도 `realpath`.
      증거: `main()` 출력 전체 A/B **126/126 SAME**(순서 번갈아, `judged at HEAD` 꼬리는 떼고 비교하되 떼어 낸 sha 를
      따로 대조) · 변이 **95/95 CAUGHT**(생존 0 — Z14 는 도달한 채 살아남아 구조 시험을 더해 다시 잡았다) ·
      시험 254 → **273** · `lint` · `sdd-test`(348) · `test-seams` · `validate` 58/58 · a122 rc=0.
      **비용은 적는다**: 끝의 재확인이 증거를 한 번 더 읽어 `ast.json` 읽기 5,803 → **9,072**(번들 총수 한 벌) ·
      git 20,133 → **20,249**. 판정은 여전히 한 번 읽은 바이트로만 선다 — 둘째 읽기는 대조만 한다.
      **로트 도중 내가 만든 회귀를 내가 찾아 고쳤다**: `_read_regular` 이 종류 검사를 파일 객체로 감싼 **뒤에** 해서
      번들 안의 폴더가 `TypeError`(GATE_FAULTS 밖)로 게이트를 죽였다 — `open(fd)` 의 예외는 `filename` 에 정수 fd 를
      담는다. 종류 검사를 앞으로 옮기고, 폴더는 **건너뛰지 않고** 경로를 담은 `IsADirectoryError` 로(건너뛰면 폴더
      안에 열거형 호출 표를 넣어 감사를 끌 수 있다), 이름 대는 자리는 글자가 아니면 번들 이름으로, 어느 갈래로
      나가도 서술자를 닫는다. 시험 다섯 · 변이 넷(Z16~Z19)을 더했고 **A/B 와 변이를 새 소스 위에서 다시 돌렸다**.
      **정정**: Pre-Edit 표의 "정규 파일 아닌 번들 항목은 건너뛰기만 한다(거절 아님)" 는 폴더 · 소켓에서 거짓이다.
- [x] 7.5.2.3 **P0 — 재확인의 입력 집합은 판정의 입력 집합이어야 한다 (7.5.2.2 재리뷰 2026-09-20, 출처 여섯).**
      지난 수리의 대조는 `HEAD` + `Evidence`(디렉터리 · 번들 목록 · `ast.json` 바이트)만 다시 읽는데, 판정은 그 밖에
      번들 **산문** · Go **워킹트리** 소스 · `review.md`(면제 표지) · `base-commit.txt` · 트리 전체 `*_test.go` 색인을
      읽는다 — 실행 중 그것들을 건드리면 `[]` PASS 이고 같은 디스크 재실행은 빨갛다(출처 4, 산문 케이스는 나 재현).
      함께 재현된 것: `record_landing` 의 재확인이 더러운 트리 거절을 **다시 안 물어서** 133~219초 순회 도중 Go 편집이
      들어오면 `open("xb")` 로 **영구** 기록이 남는다(보수 불가, 사람 손) · `HEAD` 비교가 증거 스캔보다 앞이라 스캔 중
      커밋이 통과한다(나 재현) · 내 주장 다섯이 한 줄 변이에 273 초록(창 줄 sha · 재확인의 `present`/`targets` 절반 ·
      `_bundle_text` 배관 · "명령마다 `HEAD` 한 번" · "서술자는 언제나 닫힌다") · 조용한 건너뛰기가 타이밍 공격의
      문(게이트가 여는 **그 순간에만** FIFO 로 바꿨다 되돌리면 감사 OFF · 실측 6/14 · 3/10) · FIFO 가 `review.md` ·
      `function-logic-reference.txt` · `base-commit.txt` · 시험 파일 색인에서 **영원히 멎는다**(20s+) · 큰 정규 파일은
      `MemoryError`(GATE_FAULTS 밖) · 비UTF-8 필수 산문이 대상 이름을 지운다 · README·VERIFY 문장이 보장보다 넓다.
      사람이 **"2 — 1번 전부, 감사 스위치의 비교 규칙만 사람 결정으로 남긴다"** 를 골랐다(2026-09-20). 그 규칙은
      계약이고 바꾸면 이 change 밖의 번들까지 판정이 달라진다 → 7.5.6. 기록은 review.md
      `## MEASURE · Pre-Edit Gate — task 7.5.2.3`.
      **닫음(2026-09-20). 생산 Go 변경 0.** 기록은 review.md `## VERIFY — task 7.5.2.3`. 근본 수리 한 줄:
      **판정이 디스크를 읽는 자리를 깔때기 넷(파일 · 목록 · 순회 · 종류)으로 모으고 그 깔때기가 읽은 결과를
      원장에 적게 해서, 끝의 재확인이 다시 읽는 집합이 손으로 고른 목록이 아니라 판정이 읽은 그 집합이 되게
      했다** — 손으로 고른 목록이 낡아서 깨진 것이 7.5.2.1 · 7.5.2.2 두 번이다. 곁들여: 쓰기 직전에 거절 집합
      **전체**를 다시 묻는다(더러운 워킹트리는 역사에도 원장에도 없어서 영구 기록이 생겼다) · `HEAD` 를 재확인의
      **앞뒤로** 묻는다 · 조용한 건너뛰기 폐지(정규 파일 아님 · 목록 실패 · 16 MiB 초과 · UTF-8 아님은 전부 이름 댄
      판정 줄) · `review.md` · `function-logic-reference.txt` · `base-commit.txt` · 시험 색인의 FIFO 멎음 종결 ·
      못 여는 증거 디렉터리가 면제로 통과하던 permissive 구멍 종결.
      증거: `main()` 출력 전체 A/B **126/126 SAME**(DIFFERENT 0) · 변이 **113/113 CAUGHT**(생존 0 — AA19 는
      살아남아 목록의 **종류**를 재는 시험을 더한 뒤 다시 잡았다) · 시험 273 → **298** · `lint` · `sdd-test`(373) ·
      `test-seams` · `validate` 58/58 · a122 게이트 rc=0.
      **비용은 적는다**: 파일 열기 a112 2,407 → **3,854**(+60%, 5.94s → 6.69s) · a066 1,203 → **2,246**(20.83s →
      24.00s) · git 20,249 → **20,365**. `ast.json` 읽기는 9,072 → 9,072 로 같다(7.5.2.2 도 끝에서 증거는 다시
      읽었다) — 늘어난 것은 산문 · 워킹트리 소스 · `review.md` · `base-commit.txt` · 시험 색인이다.
      **하네스도 고쳤다**: 로트 도중 배경 판과 전경 창이 한 사본을 공유해 서로의 변이를 기준으로 삼았다(무변이
      대조군이 빨개져서야 알았고 그 창들의 결과를 버렸다) → 사본을 프로세스별로 가르고, 사본이 원본과 같은지
      단언하고, 도달 계측기가 변이가 앉을 자리에 표식을 심게 했다(옛 판본은 한 줄 통째 교체에서 언제나 "안 닿음").
      **정정**: 7.5.2.2 의 "FIFO · 장치는 건너뛴다" 는 이제 거짓이다 · 심링크 고리는 "없다" 가 아니라 "못 읽는다" 다.
      **이 기록의 정정 (2026-09-22, 독립 재리뷰 다섯).** 위 문장 셋이 거짓이었다 — 고침은 7.5.2.4:
      (1) "디스크를 읽는 자리가 깔때기 넷뿐 · 재확인의 집합은 **구조적으로** 판정의 집합" 은 **이 모듈의 파이썬
      읽기**에 대해서만 참이다(자식 프로세스 `git diff` · `git show` · `go run` 과 `execution_baseline` 은 원장 밖 →
      task 7.5.8). 그 "전수" 는 사실이 아니라 **방법의 한계**였다 — 인벤토리 하네스도 구조 시험도 모듈 AST 만 걸었다.
      (2) "도달 계측기가 변이가 앉을 자리에 표식을 심게 했다" 는 거짓이었다(원복 **전**에 불려 변이된 본문에서 옛 줄을
      찾았다 — 정적 재연 79/113 이 눈먼 상태). (3) "변이 113/113" 중 **Y13 · Y14 는 `NameError`** 라 정직한 값은
      **111 CAUGHT + 2 무효**였다. 기록된 CAUGHT 판정 자체는 안 바뀐다(생존 0 인 판에서는 눈먼 계측기가 답을 안 낸다).
      숫자·문장 정정 여섯(번들 최대의 모집단 · 재현 안 되는 벽시계 초 · 하네스 없는 6/14 · "지울 수 없고" 의 모순 ·
      "건너뛰는 모양 하나뿐" · 낡은 docstring 셋)도 7.5.2.4 가 했다. 기록은 review.md `## 정정 — task 7.5.2.3`.
- [x] 7.5.2.4 **P0 — 통합 diff 의 본문 줄은 파일 헤더가 아니다 (7.5.2.3 재리뷰 2026-09-21, 보안 · 출처 다섯).**
      `--unified=0` 이면 문맥 줄이 없어 본문은 전부 `-`·`+`·`\` 로 시작한다. 그래서 지워진 소스 줄 `-- x` 는
      `--- x` 로, 더한 소스 줄 `++ x` 는 `+++ x` 로 나오는데 `changed_existing_functions` 의 파서는 상태 없이
      그것을 **파일 헤더로 읽어** 파일 중간에서 이름을 바꿨다. 진짜 git 으로 재현(`analysis/harness/7524_diff.py`):
      `--- /dev/null` 이 지워지면 그 파일의 요구가 **0 건**(permissive) · `+++ /dev/null` 이 더해지면 요구가
      `revision: base` 로 내려앉아 **편집 전** 논리의 지도가 통과 · 평범한 SQL 주석 `-- name of the table` 이 지워지면
      `cannot load existing base file <주석 문장>` 으로 **거짓 차단**(이 모양은 오늘 추적 `*.go` 에 **111 줄 / 10 파일**).
      **이 결함은 7.5.2.x 가 만든 것이 아니고 원장이 원리상 못 본다** — 입력을 자식 프로세스가 읽는다. 경합도 아니다.
      사람이 **선택지 1**(P0-B + 문서·VERIFY 정정 + 하네스 계측기 수리 한 로트)을 골랐다(2026-09-22).
      **닫음(2026-09-22). 생산 Go 변경 0.** 기록은 review.md `## MEASURE · Pre-Edit Gate — task 7.5.2.4` ·
      `## VERIFY — task 7.5.2.4`. 근본 수리 한 줄: **파서가 diff 문법대로 상태를 갖는다** — `diff --git ` 이 파일을
      열고 첫 `@@` 가 본문을 열고 본문에서는 `@@` 만 읽는다. 이름은 **첫 훅 앞에서만** 읽으므로 **새 거절은 0 개**다.
      증거: 새 시험 클래스 여섯(진짜 저장소·진짜 `git diff`, 첫 시험이 **픽스처가 허구 아님**부터 못 박는다) ·
      편집 전 실패 2 + 오류 1 → 편집 후 6/6 · 변이 `AB1~AB5` **5/5 CAUGHT**(총 118, 앵커 전수 정확히 1회) ·
      **A/B 는 전수 논증으로 먼저 증명했다**: 수리는 본문에 헤더 모양 줄이 있을 때만 다르게 답하는데 `base-commit.txt`
      를 가진 change **117** 전부에서 그런 줄이 **0** 이다 → 오늘 어느 change 에서도 두 파서의 `required` 가 같다.
      실물 `main()` 출력 A/B **126/126 SAME · DIFFERENT 0** · 계수 불변(git 20,365 → 20,365 ·
      `ast.json` 읽기 9,072 → 9,072 — **이 로트의 비용은 0**) · 순서를 번갈아 잰 벽시계도 935.1→933.8s ·
      935.0→933.1s 로 같다. 시험 298 → **305** · `lint` · `sdd-test`(logic-map 380) · `test-seams`(**108 패키지** — 7.5.22 정정,
      앞 판본의 46 은 rtk 가 자른 로그를 센 것이다) ·
      `openspec validate --all --strict` 58/58 · a122 게이트 rc=0.
      **하네스 계측기 셋을 고쳤다**: 도달 계측기가 **원본** 본문에 표식을 심고(원복 뒤 호출) · 무변이 대조군이 창
      **끝에도** 돌아 환경이 무너진 창을 통째로 버리며 판마다 `Ran N` 을 대조군과 비교하고 · 동어반복 단언 둘을
      수리라고 적지 않는다(실제로 지켜 주는 것은 pid 별 `WORK` 이름 하나다). **Y13 · Y14 를 오늘 범위로 다시 썼고
      Y13 이 실제로 살아남아**(고친 계측기가 "도달함" 이라고 답했다 — 옛 판본이면 숨었다) 시험을 하나 더 낳았다.
      `revision: base` 도 **같이 쟀다**: 번들 266 중 base 커밋의 파일 해시와 **MATCH 248 · STALE 18** →
      파일 수준 대조를 넣으면 오늘 정상인 18개를 새로 거절한다 → **사람 결정 7.5.7**.
      (7.5.22 정정: 이 줄이 적었던 귀속 "a112 16" 은 **거짓**이다. 다시 세면
      **a112 13 · a063 2 · 아카이브 a060 2 · 아카이브 a044 1** 이다 — 활성 15 · 아카이브 3.)
- [x] 7.5.22 **P0 — git 이 본문을 안 내면 그 파일의 요구가 조용히 사라진다 (7.5.2.4 재리뷰 2026-09-22, 보안 · 출처 셋).**
      `changed_existing_functions` 는 훅(`@@`)으로만 "바뀐 기존 함수" 를 센다. 그래서 워킹트리에
      `.gitattributes` 한 줄(`*.go binary` 또는 `*.go -diff`)만 있으면 — **추적조차 안 해도** — git 이
      `Binary files … differ` 를 내고 훅이 **0 개**가 되어 그 change 의 Go 요구가 **판정 줄 없이 0 건**이 된다.
      앞단 가드는 `--name-only` 만 보므로 그 파일이 **평범하게** 보여서 아무것도 안 한다. 게이트의 중심 요구를
      저자가 커밋도 안 한 파일 하나로 끌 수 있었다(진짜 git 으로 재현).
      **닫음(2026-09-22). 생산 Go 변경 0.** 사람이 **선택지 2**(정정 로트 + P0)를 골랐다.
      기록은 review.md `## 정정 — task 7.5.2.4` · `## MEASURE · Pre-Edit Gate — task 7.5.22` · `## VERIFY — task 7.5.22`.
      근본 수리: **앞단 가드가 `--name-only` 대신 `--numstat` 을 읽는다**(같은 자리, 호출 수 그대로).
      "훅이 0 개면 거절" 은 답이 아니다 — 정상인 mode-only 변경도 훅이 0 개다. `--numstat` 이 정확히 가른다:
      본문 억제는 `-`/`-`, mode-only 는 `0`/`0`. 레코드 해독은 `_numstat_records` 로 뺐고, 새 거절은
      이름 검사를 **전부 마친 뒤**에 선다(앞에 세우면 이름 가드의 시험이 남의 가드를 잰다).
      덤: `--numstat` 은 rename 의 **양쪽 이름**을 내므로 파서가 `base_file` 에 넘기는 **옛** 이름이 처음으로
      가드 범위에 들어왔다 — **7.5.13 의 절반이 닫혔다**(`--name-only` 은 새 이름만 냈다).
      FLM 먼저: `analysis/python-function-logic/tools-logic-map--_safe_changed_go_paths/`
      (편집 전 `:104-125` 분기 **11** · raise 3 → 편집 후 `:135-179` 분기 **14** · raise **4**, 헬퍼 분기 6).
      **새 거절이 오늘 거절하는 정상 입력: 0 — 세어서 하는 말이다.** `base-commit.txt` 를 가진 change **117**
      전부에서 `*.go` 파일 줄 **97,235** 를 셌다(`-`/`-` **0** · `0`/`0` **0** · rename 쌍 **0** · 이름 거절 **0**,
      `analysis/harness/7522_numstat.py`). 가드 단독 A/B **234/234 SAME · DIFFERENT 0**(`7522_guard_ab.py`)이고
      실물 `main()` 출력 전체 A/B 는 **126/126 SAME · DIFFERENT 0** — 계수 불변(git **20,365 → 20,365** ·
      `ast.json` 읽기 **9,072 → 9,072**, **호출이 하나도 안 늘었다**). 벽시계는 두 순서 모두 after 가
      +0.3%(988.1→993.1s · 994.9→997.2s).
      RED: 새 시험 클래스 일곱 중 편집 전 **실패 4 · 통과 3**(통과 셋은 통과해야 하는 셋 — 픽스처가 허구가
      아님을 못 박는 양성 대조군, 거절의 경계인 mode-only, 가드 순서). 편집 후 7/7.
      **끝까지 이어지는 영수증**(`analysis/harness/7522_switch.py`, 격리 worktree 에서 재서 공유 워크트리를
      안 건드린다): a112 는 오늘 판정 줄이 **36개** 나오고 `required` 가 **64** 인데, 추적도 안 된
      `.gitattributes` 한 줄이면 편집 전 판본은 **판정 줄 0 · required 0 · `evidence complete`** 를 낸다.
      편집 후에는 같은 입력이 **이름 댄 거절 한 줄**이다. 평소 입력에서는 두 판본이 36·64 로 같다.
      변이 `AC1~AC5` **5/5 CAUGHT · 생존 0**(총 123, 대조군이 창 시작·끝 모두 GREEN `Ran 313`),
      다섯 중 넷은 **정확히 한 시험**만 빨갛게 한다.
      게이트: `lint` rc 0 · `sdd-test` rc 0(logic-map 380 → **388**) · `test-seams` rc 0
      (**108 패키지** — `ok` 100 · `[no test files]` 8 · FAIL 0, `rtk proxy` 로 셌다) ·
      `openspec validate --all --strict` **58/58** · a122 게이트 rc 0. 시험 305 → **313**.
      **정정을 같이 실었다** — 내가 쓴 거짓 **열넷**(review.md `## 정정` 의 표)과 **내가 만든 회귀 둘**:
      (C1) 도달 계측기가 표식으로 문법을 깨면 `SyntaxError` 트레이스백이 표식 줄을 인쇄해 **시험 0개 돈 판**을
      "도달함" 으로 기록했다 — 변이 **118** 중 **30**(`7522_marker.py`, 대상 `197a0355`). 그 30 은
      한 갈래가 아니다(7.5.23 정정): 거짓 "도달함" **26** + 거짓 "안 닿음" **4**. 이제 `compile()` 로 거르고
      `Ran N` 이 대조군과 같을 때만 표식을 읽으며, **"안 닿음" 과 "못 쟀다" 를 가른다**(앞 판본은 둘 다 `False`).
      (C2) 7.5.2.4 가 훅 적재를 `hunk()` 로 **옮기면서** 그 자리를 안 쟀다 — `@@ -a,b +c,d @@` 의 뒤 쌍을 앞 쌍으로
      읽는 변이가 시험 305개 **전부를 통과**했고 동등 변이가 아니다(함수 위에 줄을 끼우면 `current_hash` →
      `base_hash`, 이 change 가 없애려는 바로 그 내려앉음). 시험을 더해 못 박았다.
- [x] 7.5.23 **P0 — 가드가 판정이 읽는 것을 읽어야 한다 (7.5.22 재리뷰 2026-09-23, 보안 · 출처 셋).**
      7.5.22 가 닫았다고 한 문이 **안 닫혔다.** `--numstat` 이 가르는 것은 "git 이 본문을 냈는가" 가 아니라
      **"git 이 이것을 이진으로 다루는가"** 다. `diff=<드라이버>` 의 **textconv** 는 원본 blob 을 세는
      `--numstat` 에 `1`/`1` 을 내면서 판정 diff(textconv **출력**을 읽는다)의 본문을 통째로 지운다.
      추적 **안 된** `.gitattributes` 한 줄 + config 키 하나로 실물 게이트(a112)가 다시
      **required 0 · `evidence complete` · rc 0** 이었다(재현했다). 기전은 7.5.22 가 진단한 것 그대로다 —
      가드와 판정이 **다른 투영**을 본다. `--no-ext-diff` 도 실은 문을 닫고 있었는데 **지워도 313 시험이
      전부 초록**인 공짜 삭제였다.
      **닫음(2026-09-23). 생산 Go 변경 0.** 기록은 review.md `## 정정 — task 7.5.22` ·
      `## MEASURE · Pre-Edit Gate — task 7.5.23` · `## VERIFY — task 7.5.23`.
      근본 수리 둘: (1) 판정 diff 가 `--no-textconv` 를 준다 · (2) **두 투영을 맞춰 본다** — 가드가
      `--numstat` 레코드를 돌려주고, numstat 이 내용이 바뀌었다고 한 파일이 훅을 하나도 안 냈으면 거절한다.
      (2)가 문을 하나씩 세는 대신 **어긋남**을 보므로 깃발이 사라져도 런타임에 잡힌다 — 실측으로
      `--no-textconv`·`--no-ext-diff` 를 각각 지우면 거절이 나온다(그래서 두 깃발이 처음으로 못 박혔다).
      새 거절이 오늘 거절하는 정상 입력 **0**(저장소에 `.gitattributes` 0 개 · 117 change · 97,235 행).
      **정정을 같이 실었다** — 7.5.22 의 정정 커밋이 **자기 실수를 되풀이했다**(review.md 의 표 열둘):
      열거표가 또 **자기 번들 둘이 빠진** 값(12,412, 실제 12,414) · 새 좌표가 **전부 +2 만큼 낡음**
      (같은 커밋의 내 주석이 민 것) · 두 `ast.after` 의 지문이 **어디에도 없는 중간 상태** · RED 기록이
      같은 커밋의 다른 두 문서와 어긋남 · 자식 프로세스 열거를 고치며 또 틀림(아홉→**열**) ·
      살아 있는 좌표 하나(`_verdict:2159`) 누락 · 7.5.13 사유가 앞뒤 안 맞음.
      **그래서 규칙을 세는 것으로 만들었다**: `analysis/harness/7523_coords.py` 가 열린 task 의 살아 있는
      좌표 · 지문이 없는 `ast*.json` · 범위가 틀린 `ast*.json` 을 세고 하나라도 있으면 rc≠0 이다.
      그 하네스가 덤으로 **7.2.6 이 남긴 뜬 번들 둘**도 찾아 되살렸다. 산문 전체의 절대 인용 **106 건**은
      누적이라 세기만 한다.
      **계측기를 또 고쳤다**: 표식 head 가 유일하지 않으면 `replace` 가 첫 자리에 심는데 문법은 멀쩡해
      `compile()` 이 못 본다 — head **20 개**가 유일하지 않고 그중 **다섯**이 다른 문장에(넷은 다른 함수에)
      앉았다. 이제 유일할 때만 심고 · 일부만 심겼으면 `"?"` · 심을 자리가 **없으면** `"?"`(앞 판본은
      표식 없는 판을 돌려 놓고 "안 닿음" 이라 답했다 — 내 수리가 낸 새 구멍) · 원복이 `finally` 안 ·
      `control_ran` 기본값 제거. 그리고 계측기를 **처음으로 직접 시험했다**(`7523_reached.py` 6/6) —
      `reached()` 는 SURVIVED 인 판에서만 불려 생존 0 인 로트에서는 증거에 안 나타난다.
      증거: 시험 313 → **326** · 변이 **15/15 CAUGHT · 생존 0**(총 128, 대조군 창 양끝 GREEN `Ran 326`) ·
      실물 `main()` A/B **126/126 SAME · DIFFERENT 0**(계수 불변: git 20,365 → 20,365 · `ast.json`
      읽기 9,072 → 9,072 — **호출이 하나도 안 늘었다**) · `lint` rc 0 · `sdd-test` rc 0(logic-map **401**) ·
      `test-seams` rc 0(**108 패키지** FAIL 0) · `openspec validate --all --strict` 58/58 ·
      a122 게이트 rc 0 · `7523_coords.py` rc 0.
- [x] 7.5.24 **P0 — 내가 넣은 교차 검사가 정상 입력을 거절했다 (7.5.23 재리뷰 2026-09-23, 적대 · 주장정확성).**
      7.5.23 의 교차 검사는 가드의 `--numstat` 이름(날 바이트)과 파서의 이름(`removeprefix` 를 거친
      **유도값**, git 이 인용하면 `"a/we\"ird.go"`)을 **교집합**으로 견줬다. 두 이름 공간이 안 만나서
      부모에서 통과하던 정상 입력이 거절됐다(실측, 부모와 A/B): 새 파일 `we"ird.go` · `back\slash.go` ·
      `diff.noprefix` 아래의 `b/new.go`. 거절은 `cannot derive modified Go functions` 라 **판정 줄이
      하나도 안 나가고**, 메시지는 있지도 않은 필터를 탓했다. `--numstat` 열거로 "정상 입력 거절 0" 이라
      적은 것은 **오늘 저장소의 상태**를 잰 것이지 거절 가능한 **집합**을 잰 것이 아니었다
      ([[universal-check-passes-on-an-empty-sample]]).
      **닫음(2026-09-23). 생산 Go 변경 0.** 수리: 짝을 **이름이 아니라 순서로** 짓는다. 두 호출은 같은
      `git diff` 에 형식만 다르므로 파일을 같은 순서로 낸다(편집·mode-only·100% rename·rename+편집·
      삭제·추가 섞은 픽스처로 실측, 5/5 정렬). 이름 공간을 아예 안 비교하므로 인용·접두사 문제가 사라진다.
      두 시야의 **크기**가 다르면 그것도 거절한다(깃발이 사라지면 판정 diff 가 구역을 아예 안 내므로
      이쪽이 먼저 잡는다 — 실측).
      **경계 셋을 새로 못 박았다**(7.5.28 정정: "둘" 이라 적었다 — 크기 검사 시험은 AE2 가 살아남은 뒤 더했다): 건너뛰기 조건의 `and` 를 `or` 로 바꾸는 변이가 326 시험을 **전부
      통과**했다(더하기만 `N`/`0` · 지우기만 `0`/`N` 레코드가 통째로 빠진다) → `AE1` + 시험 ·
      git 이 인용하는 이름은 거절하면 **안 된다** → 시험. 가드를 mock 하던 옛 시험 셋도 레코드를 같이
      주게 고쳤다(빈 MagicMock 이 "바뀐 파일 0" 을 뜻하게 됐다).
      **좌표 게이트가 이제 산문의 수도 본다** — 7.5.23 의 거짓 셋이 전부 "하네스가 낸 수를 산문에 베껴
      적고 그 뒤 편집이 수를 바꿨다" 였다(인용 105→**106** · BTM 이 draft 의 `bodied.update` 두 자리와
      지워진 `AD5` 를 적음 · "앵커 셋 재겨눔" 이 실제로는 하나). `7523_coords.py` 가 자기가 내는 수를
      산문과 대조하고 다르면 rc≠0 이다.
      증거: 시험 326 → **329** · 변이 **19/19 CAUGHT · 생존 0**(총 132, 대조군 양끝 GREEN `Ran 329` —
      `AE2` 는 첫 판에서 살아남아 계측기가 "도달함" 이라 답했고 시험을 하나 더 낳았다) ·
      실물 `main()` A/B **126/126 SAME · DIFFERENT 0**(계수 불변: git 20,365 → 20,365 ·
      `ast.json` 읽기 9,072 → 9,072) · `lint` rc 0 · `sdd-test` rc 0(logic-map **404**) ·
      `test-seams` rc 0(**108 패키지**) · `openspec` 58/58 · a122 게이트 rc 0 · 좌표 게이트 rc 0.
- [x] 7.5.27 **P0 — 워킹트리를 다시 쓰는 알려진 문을 명령줄에서 닫는다 (7.5.25 의 싼 막기, 사람이 선택지 2 를 골랐다 2026-09-23).**
      7.5.25 의 **class**(git 의 투영을 안 믿고 판정 바이트를 직접 대조)는 그대로 열려 있다 — 이것은
      **알려진 문**을 닫는다. 실측으로 문을 셌다(진짜 git): `filter.<드라이버>.clean` · 진짜 v2 프로토콜을
      말하는 `filter.<드라이버>.process` · 거짓말하는 `core.fsmonitor` 는 편집을 **두 시야 모두**에서 감추고
      명령줄 `-c` 고정으로 되돌아온다. `assume-unchanged`(`ls-files -v` 의 소문자) · `skip-worktree`(`S`)
      는 설정이 아니라 인덱스 플래그라 고정으로 안 닫힌다. `core.checkStat=minimal` + `trustctime=false` +
      mtime·크기 되돌리기는 **재현되지 않았다**(git 이 잡았다) — 고정을 안 넣었다(YAGNI). (7.5.28 정정: **거짓이었다** — 픽스처가 mtime 을 과거로 안 돌려 racy-git 창에 걸려 있었다. 과거 시각으로 두면 재현되고, 7.5.28 이 고정을 넣었다)
      **측정 오류 하나를 스스로 찾았다**: 첫 전수에서 process 필터를 "감췄다" 로 읽었는데, 프로토콜을
      안 지키는 스크립트라 **git 이 rc 128 로 죽은 것**이었다(rc 를 안 봤다). 진짜 v2 프로토콜 필터로 다시
      재서 감추는 것을 확인했고, 시험 픽스처도 그 필터를 쓴다.
      **닫음(2026-09-23). 생산 Go 변경 0.** 수리 둘: (1) `_git_view_pins` 가 판정의 **두** `git diff` 에 같은
      고정을 준다 — `core.fsmonitor=false`, 그리고 설정된 **모든** 필터 드라이버의 `clean=`·`process=`·
      `required=false`(`clean` 만 비우면 `required=true` 드라이버에서 git 이 rc 128 로 죽는다 · 드라이버
      이름에 점이 들어갈 수 있어 키의 **마지막** 칸만 뗀다 · 이름에 `=`·공백이 있으면 `-c` 로 못 적으니
      거절). (2) `_hidden_by_index_flags` 가 워킹트리 대상에서 플래그 붙은 `*.go` 를 **같은 고정으로 해시**해
      인덱스 blob 과 다를 때만 이름 대고 거절한다 — sparse-checkout 도 `skip-worktree` 를 쓰므로 플래그만으로는
      안 거절한다(플래그만·편집 없음, 파일 없음(sparse 모양), 커밋 대상 셋이 경계 시험이다). 가장 뒤에 선다.
      **실물 게이트 영수증**(`analysis/harness/7527_switch.py`, 격리 worktree · 설정은 **환경 변수로만** 주입해
      공유 `.git/config` 에 아무것도 안 쓴다): a112 는 평소 판정 줄 **36** · required **64**. 추적 안 된
      `.gitattributes` 한 줄 + clean 필터 설정 하나면 편집 전 판본이 **판정 줄 0 · required 1 · `evidence complete`**,
      편집 후는 **36 · 64 그대로**(거절이 아니라 올바른 판정이다).
      **새 거절이 오늘 거절하는 정상 입력 0** — 추적 `*.go` 1,762 전부 플래그 없음(`H`) · 필터 드라이버 0 ·
      fsmonitor 없음 · sparse 꺼짐. **거절·변경 가능한 집합도 적는다**(7.5.24 의 교훈): `*.go` 에 **정상적인**
      clean 필터를 쓰는 저장소(git-lfs · git-crypt)는 중화 때문에 워킹트리가 raw 바이트로 비교된다 — 판정이
      달라지거나(lfs 포인터 대 실제 내용) base 파싱이 실패해 거절된다(git-crypt). 드라이버 이름에 `=`·공백이
      있으면 거절한다. 오늘 노출 0.
      변이에서 git 의 **우연** 하나: `process=` 하나만 비워도 clean 필터가 꺼져서(빈 `process` 값은 아무것도
      안 띄우고 clean 을 끈다 — 7.5.28 정정: "시작에 실패" 가 아니다, `GIT_TRACE` 로 `run_command` 0 줄) `clean=` 을 빼는 변이 `AF2` 가 살아남았다 — 동등이라 적지 않고 반환값을 직접 재는
      **계약 시험**으로 못 박았다. 옛 mock 시험 셋이 새 git 호출에서 먼저 실패를 받아 이름과 다른 이유로 통과할
      뻔한 것도 새 헬퍼를 고정해 막았다.
      증거: 시험 329 → **339** · 변이 첫 판 `113:143` **29/30**(AF2 생존) → 시험을 더해 `AF1~AF11` 재실행 **11/11**(총 143 · 생존 0 ·
      대조군 양끝 GREEN; 7.5.28 정정: 앞 판본의 "`113:130` 19/19" 는 창 이름이 틀렸다 — 그 창에는 17 개가 있다) · `lint` rc 0 · `sdd-test` rc 0(logic-map **414**) · `test-seams` rc 0(**108 패키지**) ·
      `openspec validate --all --strict` 58/58 · a122 게이트 rc 0 · 좌표 게이트 rc 0 · 실물 `main()` A/B
      **126/126 SAME · DIFFERENT 0** — git 호출은 20,365 → **20,596(+231, +1.1%)**, 판정마다 설정 읽기와
      (워킹트리 대상이면) 인덱스 조회가 한 번씩 붙는다.
- [x] 7.5.28 **P0 — 7.5.24·7.5.27 재리뷰가 연 것: 내 회귀 하나 · 내 거짓 하나 · 새 문 하나 · 옛 결함 하나 (2026-09-24, 적대).**
      **(1) 내 회귀.** 판정 diff 를 `text=True` 로 받아 `str.splitlines()` 로 잘랐는데, 그것은 `\n` 말고도
      U+2028 · U+2029 · U+0085 에서 자른다. `core.quotePath=false` 라 git 은 그 글자를 인용하지 않고 이름 가드는
      `\n\r\t` 만 거절한다. 경로 `a<U+2028>@@ -1 +1 @@.go` 가 `diff --git` 머리를 잘라 뒷조각이 **훅**으로 읽히고
      본문이 이름보다 먼저 열려 그 파일의 요구가 **통째로** 사라졌다. 이름 짝짓기(7.5.23)는 이 입력을 **거절**했으니
      순서 짝짓기(7.5.24)가 낸 회귀다(세 글자 모두 재현, 평범한 파일을 그 이름으로 rename 해도 같다). Go 는 이 이름을
      받는다. 수리: **바이트로 받아 `\n` 에서만 자른다**(해독은 엄격 — UTF-8 아닌 diff 는 전처럼 결함).
      **(2) 내 거짓.** 7.5.27 이 "`checkStat` 우회는 재현 안 됨" 이라 적었는데, 그 픽스처는 mtime 을 **과거로**
      안 돌려 racy-git 창에 걸려 있었다. 과거 시각으로 두면 `core.checkStat=minimal` 이 같은 크기 편집을 **두 시야
      모두**에서 감춘다(부모도 같다). 수리: `core.checkStat=default` · `core.trustctime=true` 고정.
      **(3) 새 문.** `.gitattributes` 의 **내장** 속성 `ident` 는 `$Id: …$` 를 `$Id$` 로 접은 뒤 비교해 그 안의 논리
      변경을 감춘다. 속성이라 `-c` 로 못 끈다 → 워킹트리 대상에서 추적 `*.go` 에 `ident` 가 켜져 있으면 **이름 대고
      거절**한다(`_ident_go_paths`, 오늘 노출 0).
      **(4) 옛 결함 — 재현하다 드러났다.** git 은 이름에 **공백**이 있으면 `---`·`+++` 줄 끝에 **탭**을 붙인다. 그 탭을
      이름으로 읽어 base 파일을 못 찾고 `cannot load existing base file` 로 **거짓 차단**했다 — 이 change 이전부터
      있던 결함이다(부모에서도 같다, 저장소 노출 0). 이름 가드가 `\t` 든 이름을 거절하므로 끝의 탭 하나를 떼는
      것은 모호하지 않다(`_header_name`).
      **덤**: 공백이 든 드라이버 이름을 거절하던 것은 헛거절이었다(`-c 'filter.my drv.clean='` 은 적힌다) — `=` 만
      거절한다. 플래그 검사의 `hash-object` 고정을 빼는 변이와 `100755` 모드를 빼는 변이가 시험 전부를 통과했다 —
      시험 둘로 못 박았다. mock 시험 일곱(7.5.25 정정: 경우 일곱 · 함수 여섯)은 판정 diff 가 바이트가 되면서 계약에 맞게 고쳤다.
      **흔들린 시험이 새 측정을 낳았다**: `checkStat` 시험이 부하 아래 12 판 중 3 판 실패했다 — inode 재사용 +
      같은 초의 ctime + 과거 mtime 은 **고정한 기본 stat 검사**도 속인다(설정 없이 20 판 중 17 판). 설정 문만 재도록
      픽스처를 고치고 남은 구멍은 7.5.25 에 적었다.
      **주장정확성 재리뷰의 정정도 실었다**(같은 실수 다섯 번째 — review.md 표): 가드 분기 "그대로" → 14→15 ·
      BTM 이 대던 시험이 그 갈래를 **못** 박음(둘) · 변이 창 이름 "`113:130` 19/19" 는 17 개 창 · 빈 `process=` 는
      "시작 실패" 가 아니라 **아무것도 안 띄움**(`GIT_TRACE`) · 하네스 주석의 "105건" · 7.5.24 "경계 둘" 은 셋.
      증거: 시험 339 → **348** · 변이 첫 판 `113:152` **38/39**(AG4 생존·"못 쟀다") → 경계 시험 뒤 `AG1~AG9` **9/9**
      (총 152 · 대조군 양끝 GREEN `Ran 348`) · `lint` rc 0 · `sdd-test` rc 0(logic-map **423**) ·
      `test-seams` rc 0(**108 패키지**) · a122 게이트 rc 0 · 실물 `main()` A/B **126/126 SAME** — git 호출
      20,596 → **20,826(+230, +1.1%)**, 워킹트리 대상마다 `ident` 조회 두 프로세스.
- [x] 7.5.31 **P0 — 7.5.25 재리뷰가 연 것: 내 회귀 셋 · 헛거절 둘 · 거짓 일곱 (2026-09-24, 사람이 선택지 1 을 골랐다).**
      **내 회귀 셋.** 스냅숏은 디스크 바이트를 바르게 읽지만 두 diff 는 그 blob 을 **oid 로** 다시 찾고 그때 저장소를
      믿는다. (F1) `refs/replace/<편집 blob>` 참조 **하나**면 편집 blob 이 base blob 으로 읽혀 요구가 빈다 — 편집 blob 이
      저장소에 없어도 된다(내가 재현: `git status` 는 `M`, 게이트 `[]`). (F2) 스테이지한 편집의 loose 객체를 base 바이트로
      위조 · (F3) pack 에 위조 — git 은 pack 을 먼저 보므로 스냅숏이 객체를 늘 써도 못 막는다. 부모(`cfe8d08b`)는 워킹트리
      쪽을 디스크에서 직접 읽어 셋 다 잡았다.
      **수리.** (1) 모듈이 `GIT_NO_REPLACE_OBJECTS=1` 을 한 번 둔다 — 게이트의 모든 git 이 교체 참조를 안 따른다(F1 과
      base 쪽의 옛 구멍 F4 가 같이 닫힌다). (2) `_snapshot_disagreement` — **git 밖의** 대조: 게이트가 Python 으로 해시한
      oid 가 base 트리(`ls-tree`)의 oid 와 다른 경로는 레코드의 새 이름이어야 하고 이름이 안 바뀐 `0`/`0` 이면 안 되며,
      지운 경로는 옛 이름이어야 한다. 가드가 읽은 **같은** 레코드로, 모든 가드 **뒤에** 선다. 위조 저장소 시험에서는 git
      이 그 경로를 레코드로 **아예 안 냈다** — 규칙 1 이 잡는다. (3) 대체 저장소 항목을 C 인용으로 적는다(`_c_quoted`) —
      7.5.25 의 `:`·`"` 경로 거절(헛거절, F10)을 없앴다. (4) 경로를 `os.fsdecode` 로 읽는다 — 이름이 UTF-8 이 아닌 **안
      바뀐** 파일로 멈추던 것(F9)을 없앴다; 바뀐 파일은 가드가 전처럼 거절한다.
      **거짓 일곱**(주장정확성 재리뷰, 같은 실수 **일곱 번째**): BTM 의 AH2 "21 시험" · AH15 "91 시험" 은 같은 커밋에서 뒤에
      더한 시험으로 22 · 92 가 됐다 · 7.5.26 의 "`diff.noprefix` 가 요구를 조용히 지운다" 는 **재지 않고 적었고 틀렸다**
      (`a/`·`b/` 로 시작하는 경로에서만 문제) · `color.ui=always` 는 base 에서 조용히 `[]` 였다("이전부터 거짓 차단" 아님) ·
      `mnemonicPrefix` 는 `--cached` 뒤로 `c/`·`i/` · 가드를 직접 부르는 시험은 셋 · 몸통 docstring "스냅숏이 열려 있는 동안"
      은 커밋 대상에서 거짓 · README 의 clean 필터 "늘" 은 과장(항등 필터면 같다). 전부 제자리 정정했다.
      **새 task 둘**: 7.5.32 파서를 판정받는 워킹트리에서 빌드한다(F6, P0, 이전부터) · 7.5.33 base 쪽은 여전히 저장소를
      믿는다(F5 · 부분 클론 fetch). `GIT_GLOB_PATHSPECS`(F7)는 7.5.26 에 실었다(7.5.34 가 닫았다 — 그 항목의 정정 참조).
      **변이가 코드를 지웠다**: `AI10`(ls-tree 의 blob 종류 거름을 뺌)이 살아남았다 — gitlink `*.go` 는 **인덱스에** 남아
      있으면 스냅숏이 모드로 거절하고(실측) 지우면 numstat 이 `0 1 sub.go` 로 낸다(실측). 닿을 수 없는 방어라 줄을 지웠다.
      (7.5.34 정정: 앞 판본은 "워킹트리에 남아 있으면" 이라 적었다 — 거절을 부르는 것은 인덱스 항목이고, 워킹트리에 디렉터리가
      없어도 거절됐다, 주장정확성 재리뷰 실측.)
      증거: 시험 357 → **362**(부모 코드에 붙이면 13 경우 빨강 — 새 동작을 재는 6 함수 · 8 경우: 실패 4 · 오류 4, 나머지 5
      오류는 스냅숏이 셋째 값을 돌려주게 된 계약 변경) · 변이 171 → 지운 AI10 을 빼고 **170**: 창 `0:43` · `43:86` · `86:129`
      각 **43/43**, `129:171` **41/42**(생존 AI10) — 대조군 모든 창 양끝 GREEN `Ran 362`. `lint` rc 0 · `sdd-test` rc 2 — **logic-map 437 통과**(357 → 362 에 스위트 밖 시험 포함 432 → 437); 멈춘 곳은 `tools/sdd` 의 `test_repository_contract_is_synchronized` 하나("Claude/Codex minimal safety bootstrap drift") — 다른 세션이 커밋하지 않은 `.claude/CLAUDE.md` · `.codex/agents.md` 편집 탓이다(깨끗한 클론 `9bc103fa` 에서는 통과, 이 로트는 두 파일을 안 건드리고 커밋에도 안 넣었다). 멈춘 뒤의 묶음은 따로 돌렸다: `sdd-history` 22 · `pm` 16 · `deploy` 18 · `go test ./tools/logic-map` 전부 통과 ·
      `test-seams` rc 0(**108 패키지**) · `openspec validate --all --strict` **58/58** · a122 게이트 **rc 0**(a112 `required 64` 그대로) · 좌표 게이트 **rc 0** · 실물 `main()`
      A/B **126/126 SAME · DIFFERENT 0**(기준 `9bc103fa`, 예산이 모자라 두 번에 이어 달림) — git 호출 20,710 → **20,825**: change 115 개가 정확히 +1(워킹트리 대상 판정마다 `ls-tree` 하나), 11 개는 그대로(커밋 대상이거나 이 함수 앞에서 멈춤) · `ast.json` 읽기 9,072 → 9,072 · 벽시계는 기계 부하로 흔들렸다(before 1,918.6 · 1,875.8s / after 1,779.0 · 1,824.8s — 참고만).
- [x] 7.5.34 **P0 — 7.5.31 재리뷰가 연 것: git 밖 대조는 내용이 아니라 있음만 봤다 (2026-09-24, 사람이 선택지 1 을 골랐다).**
      **결함.** `_snapshot_disagreement` 는 "oid 가 다른 경로가 레코드에 나왔는가" 만 본다 — git 이 그 blob 을 **oid 로**
      실제 저장소에서 찾는 한 저장소의 대답이 판정이다. (#1) 편집 blob 의 oid 에 **base 도 디스크도 아닌** 바이트를 심으면
      레코드는 나오고 요구는 빈다(또는 엉뚱한 함수로 옮겨 간다). (#2) `git mv` 뒤 새 이름의 blob 에 base 바이트를 심으면
      100% rename `0`/`0` 이라 규칙이 다 통과하고 요구가 빈다 — 삭제도 같이 숨는다. (#3) 새로 더한 **빈** `.go` 를 규칙 3
      이 헛거절한다. (#4) 연결 워크트리의 주 저장소 경로에 UTF-8 이 아닌 바이트가 있으면 `rev-parse` 출력의 엄격 해독이
      터진다. (#5) 그 경로에 줄바꿈이 있으면 대체 저장소 경로가 잘린다. 주장정확성 재리뷰는 base 쪽 구멍(7.5.33)이 미끼
      변형으로 ff19be01 에서도 `[]` 임을 쟀다.
      **수리(선택지 1 — class 를 닫는다).** 게이트의 git 이 판정 blob 을 **실제 저장소에서 oid 로 찾지 않는다.** base(와
      커밋 대상)는 커밋 → 트리 → blob 을 `cat-file --batch` 로 읽어 **Python 이 해시를 검증**하고, 워킹트리 쪽은 디스크
      바이트를 그대로 쓴다. 두 쪽의 `*.go` 만 담은 트리를 **대체 저장소가 없는** 임시 저장소에 세우고 두 diff 는 그 두
      트리를 견준다. 그래서 교체 참조 · 위조 loose · 위조 pack · 트리 · 커밋 위조가 모두 **해시 불일치로 이름 대고**
      멈추거나 판정에 닿지 않는다. git 밖 대조(`_snapshot_disagreement`)는 지킬 것이 없어져 지우고(#3 이 같이 없어진다),
      대체 저장소 경로(`--git-path objects` · `_c_quoted`)도 없어진다(#4 · #5). 두 diff 와 목록이 pathspec 을 안 받으므로
      `GIT_*_PATHSPECS` 도 판정에 안 닿는다. 판정의 옛 쪽은 `base_file` 의 `git show` 가 아니라 검증한 바이트를 읽는다.
      **작업 중 잰 것 셋.** (a) 첫 판은 `ls-tree -r -d` 의 항목을 전부 트리로 물었는데 `-d` 는 gitlink(`160000 commit`)도
      낸다 — 하위 모듈이 하나라도 있으면 헛거절이었다(실측 후 고침, 시험 `test_a_submodule_in_the_base_is_not_refused`).
      (b) `write-tree` 도 인덱스를 다시 쓴다 — `update-index` 에만 고정을 준 첫 판은 split index 에서 `sharedindex.*` 를
      실제 `.git` 에 썼다(기존 시험이 잡았다, 두 호출이 `INDEX_WRITE_PINS` 를 받는다). (c) git 은 뿌리 트리 · 커밋은 파싱할
      때 해시를 보지만(`hash mismatch`) **하위 트리는 해시 없이 읽는다** — 하위 트리를 빈 트리로 위조하면 `ls-tree -r` 에서도
      그 아래 파일이 조용히 사라진다(2.43 실측). 이 모양은 이 로트의 검증만 잡는다.
      **동작이 바뀐 것.** sparse 로 안 꺼낸 파일이 base 와 다르면 현재 쪽도 검증한 인덱스 blob 으로 본다(전에는 현재 쪽이
      비었다 — 요구가 늘 수만 있다). base 에서 지운 gitlink `*.go` 는 늘 `cannot load existing base file` 로 멈춘다(전에는
      그 커밋이 같은 저장소에 있으면 커밋 글을 Go 로 읽어 `{}`). 커밋 대상(착지 · 이관 감사)도 검증한 두 트리를 견준다.
      증거: 시험 362 → **372**(바꾼 것 10 을 17 로 · gitlink 둘 · 뿌리/커밋 위조 분리 하나). 시험 파일을 ff19be01 코드에
      붙이면 **53 경우 빨강**(실패 12 · 오류 41) — 행동이 틀린 것 22(실패 12: base 위조 셋 · 미끼 위조 둘 · rename · sparse
      둘 · 커밋 대상 둘 · gitlink 하나 · 구조 하나 / 오류 10: pathspec 셋 · 연결 워크트리 둘 · 빈 파일 둘 · 물려받은 대체
      저장소 · base 바이트 위조 둘 — 7.5.35 정정: 이 둘을 "헛거절" 이라 적었는데 위조 저장소를 거절한 것이지 정상 입력을
      거절한 것이 아니다; ff19be01 은 규칙 1 의 **이름 댄 거절**이었고 새 시험이 기대하는 문장과 달라 오류로 셌다, review.md
      의 표가 이미 그렇게 적었다), 나머지 31 은 새 API(`_verified_objects` · `_isolated_comparison` 결함 시험과
      대역). 뿌리 트리 · 커밋 위조와 하위 모듈 시험은 ff19be01 에서도 초록이다(앞은 git 자신이 거절, 뒤는 첫 판의 회귀 방지).
      변이 170 → **189**(AH · AI 31 을 코드와 함께 은퇴 · AJ 50 더함 · `_committed_many` 의 T6 · U1 · V6 앵커를 새 코드의 같은
      철자 때문에 넓힘): 첫 판 190 — 창 `0:48` 48/48 · `48:96` 48/48 · `96:143` 47/47 · `143:190` **45/47**(생존 AJ44 ·
      AJ47). AJ44(바이트 출처를 고르는 갈래)는 같은 oid 면 같은 바이트라 고를 것이 없었다 → oid → 바이트 한 표로 바꾸고 그
      표를 겨누는 변이로 교체. AJ47(두 diff 의 `GIT_INDEX_FILE`)은 트리 둘을 견주는 diff 가 인덱스를 안 쓴다 → 줄을
      지웠다. 그 함수를 겨누는 변이를 최종 코드에서 다시 돌렸다: 창 `174:187` **13/13** · `135:136`(AH21) **1/1**. 모든
      창 대조군 양끝 GREEN `Ran 372`(하네스 사본은 `docs/WORKFLOW.md` 가 없어 skip 2). `make lint` rc 0 · `make sdd-test`
      rc 2 — **logic-map 447 통과**(372 + 스위트 밖 75), 멈춘 곳은 전처럼 `tools/sdd` 의 bootstrap drift 하나(다른 세션의 커밋
      안 한 `.claude/CLAUDE.md` · `.codex/agents.md`; 깨끗한 클론에서 71 통과) · 멈춘 뒤의 `sdd-history` 22 · `pm` 16 ·
      `deploy` 18 · `go test ./tools/logic-map` 통과 · `make test-seams` rc 0(**108 패키지**, `rtk proxy` 로 셈 — `GOFLAGS=
      -trimpath` 로 돌린 첫 판은 자기 소스를 경로로 여는 Go 시험 셋이 FAIL 했다: 내 환경 탓, 코드 변경 0) · `openspec
      validate --all --strict` **58/58** · a122 게이트 **rc 0**(a112 `required 64` 그대로) · 좌표 게이트 **rc 0**(첫 판은 최종
      편집 전에 뽑은 `ast.after` 일곱이 떠서 rc 1 — 다시 뽑았다). 실물 `main()` A/B **126/126 SAME · DIFFERENT 0**(기준
      ff19be01) — git 호출 20,825 → **2,431**: ~~116 change 가 100~230 씩 줄었다~~ **115 change 가 8~228 씩 줄었고**(100
      이상 98 · 100 미만 17 — a121 · a064 −8, a120 −15, a099 −19, a112 −32, a118 −33, a117 −52, a113 · a114 · a115 −53, a116
      −55, a109 −66, a100 −84, a111 −96, a107 · a108 · a110 −99; 바뀐 파일마다 `git show` 하던 옛 쪽 읽기가 `cat-file` 하나가
      됐다), 10 은 그대로, **1 은 늘었다**(a122 14 → 18, +4). (7.5.35 정정 — 주장정확성 재리뷰, 같은 실수 아홉 번째: 분포의
      위쪽만 보고 적었다. A/B JSON 을 다시 세어 옮겼다.) · `ast.json` 읽기 9,072 → 9,072 · 벽시계 before 1,451.9 · 1,455.2s / after 1,434.1
      · 1,442.6s(순서를 번갈아, 참고만).
      **정정 — 주장정확성 재리뷰(ff19be01), 같은 실수 여덟 번째.** 틀린 문장 넷: 규칙 직접 시험의 docstring(규칙 1 · 2 가
      아니라 3 · 2 — 시험과 함께 지웠다) · 7.5.26 의 `GIT_GLOB_PATHSPECS` · `GIT_LITERAL_PATHSPECS` "조용히 지운다"(ff19be01
      에서는 규칙 2 의 헛거절) · 7.5.33 의 재현 문장(ff19be01 에서는 이름 댄 거절). 그리고 7.5.31 의 gitlink "워킹트리에 남아
      있으면" → 인덱스. 전부 제자리에 정정을 적었다.
- [x] 7.5.35 **P0 — 7.5.34 재리뷰가 연 것: 해시가 맞는 트리를 git 과 다르게 읽는다 · 판정 diff 의 형식 (2026-09-25, 사람이 선택지 1 을 골랐다).**
      **결함.** (P0-1, 7.5.34 가 연 회귀) 7.5.34 의 걷기는 정확히 `40000` 만 하위 트리로 따라갔다 — git 은 `040000` 도
      디렉터리로 읽는다(`canon_mode`, `fsck` 는 `zeroPaddedFilemode` 경고뿐). base 에 그런 트리를 두면 `sub/` 의 요구가 통째로
      빈다(이 로트에서 다시 쟀다: ff19be01 코드 `[('sub/x.go', 'F')]` · 54004f44(= 0ece1074) 코드 `[]` · 7.5.35 코드 이름 댄 거절 —
      워킹트리 · 커밋 대상 × sha1 · sha256 네 조합 모두 같다). (P1-2, 같은 회귀)
      같은 이름 둘이면 걷기는 뒤의 것만 남겨 한 순서에서 편집을 잃는다(git 도 순서에 따라 다르게 낸다 — `ls-tree -r` 은 둘 다
      내고 diff 는 (base, 편집) 순서에서 수정 + 삭제, (편집, base) 순서에서 삭제 하나; 7.5.35 의 두 재리뷰가 고친 문장이다) ·
      이름에 `/` 가 든 항목을 걷기는 하위 트리의 같은 경로와 하나로 모은다(git 은 둘 다 낸다) — 둘 다 한 순서에서
      `[]`(이 로트에서 다시 재현: `x.go` 둘(base, 편집) · `sub/` 에 편집 blob + 평평한 `sub/x.go` 에 base blob). (P0-3 · P2, 이
      change 이전부터 — 7.5.26) `diff.noprefix=true` 면 `b/x.go` 의 요구가 `x.go` 로 가고 `a/x.go` 는 `cannot load existing base
      file` 헛거절 · `color.ui` · `color.diff=always` 는 교차 검사 헛거절(이 change 의 **base 7f8b1ae7 에서는 요구를 조용히 비웠다** — 7.5.23 의
      교차 검사가 그것을 헛거절로 바꿨다; 이 로트에서 base 코드로 다시 쟀다) · `GIT_DIFF_OPTS=-u3` 과 `diff.interHunkContext=9` 는
      편집 안 한 함수까지 요구(뒤의 것은 이 로트에서 찾았다 — A 와 C 만 고치면 B 도 요구).
      **수리.** (1) 검증한 트리도 git 이 **쓰는** 모양이어야 받는다(`_verified_go_entries`): 기본 `git fsck` 가 받는 모드 여섯(`40000` ·
      `100644` · `100664` · `100755` · `120000` · `160000`) · git 의 정렬(하위 트리는 이름 뒤에 `/`) · 유일한 이름 · 비지 않고 `/` 가 들지 않고 `.` · `..` 가
      아닌 이름 — 어기면 `NOT_CANONICAL` 로 트리 · 자리 · 커밋 · 사유를 대고 멈춘다. 읽는 법을 git 에 맞추지 않은 까닭: 맞추면 fsck 가
      오류로 보는 모양(중복 · `/`)의 "git 의 읽기" 를 따라 적어야 하고 그 사본이 또 갈린다. (2) 판정 diff 에 `--src-prefix=a/
      --dst-prefix=b/ --no-color --inter-hunk-context=0`. (3) `_isolated_comparison` 이 자식 환경에서 `GIT_DIFF_OPTS` 를 지운다.
      **거절하는 정상 입력(먼저 셌다).** `git rev-list --objects --all` 로 닿는 트리 전부에 같은 규칙(측정 스크립트 `census.py` 가
      옮겨 적은 것 — `_verified_go_entries` 를 부르지 않는다)을 걸어 거절 **0**: 12 시 무렵 **13,749**(항목 218,464 — 모드 `100644`
      123,376 · `40000` 93,716 · `100755` 1,372), 20:47 에 다시 13,980(항목 222,112). 고정하지 않은 것: `diff.algorithm`(patience · histogram · minimal) ·
      `diff.indentHeuristic=false` 는 모호한 편집 셋(사이에 끼우기 · 가운데 지우기 · 본문 맞바꾸기)에서 판정이 같았다.
      `diff.mnemonicPrefix` 는 트리 둘을 견주는 diff 에서 접두사를 안 바꾼다(0ece1074 에서도 요구가 선다 — 시험에 남겼다).
      **변이가 시험을 바꿨다.** 첫 판 209 중 **208** CAUGHT — AK6(빈 이름 검사를 지움)이 살아남았다: 시험은 "canonical form"
      만 봤고, 빈 이름의 **파일**은 열쇠 `b""` 가 앞보다 클 수 없어 순서 검사가 같은 문장으로 대신 거절했다. 빈 이름의
      **트리**(열쇠 `b"/"`)는 이름 검사만 잡는다 — 그 모양을 더하고 모양마다 **사유**를 단언하게 했다. 이 시험 수정은 병행 세션
      (tossos-d6)의 팀메이트가 이 로트를 고아로 오판해 12:17 에 먼저 적었다(그 세션이 12:28 에 멈췄고 이 로트를 이 세션의 것으로
      확인했다 — 시각은 그 세션의 통보다). 내용을 읽고 돌려 보고 그대로 받았다. AK6 재실행 창 `192:193` **CAUGHT**(그 판은
      재리뷰 전 376 시험 파일 — 최종 377 시험의 증거는 창 `159:212`).
      **이 로트의 적대 재리뷰(첫 판, 코드 변경 전).** P0 · P1 없음. 고친 셋: (첫 판 코드는 따로 저장하지 않았다 — 아래 셋은 변이 AK21–AK23 이 그 모양이다.) (P2-1) 첫 판은 `100664` 를 거절했다 — 초기 git 이
      쓴 모드이고 기본 `git fsck` 가 받으며(`--strict` 만 `badFilemode`, 이 로트에서 모드 열 개로 실측) git 도 편집 전 게이트도
      일반 파일로 읽는다. 옛 저장소 헛거절이었다(이 저장소 노출 0) → 모드 표를 **기본 fsck 가 받는 여섯**으로. (P2-2) 붙어 있지
      않은 같은 이름(`x.go` · `x.go-` · 트리 `x.go` — 열쇠가 엄격히 오른다)을 시험이 안 못 박았다: 옆 항목만 보는 중복 검사(리뷰어
      M11)가 시험을 지나고 편집 전의 **조용한 삭제 손실**을 다시 연다 → 시험에 그 모양을 더했다. (P3-1) gitlink 의 정렬 열쇠(리뷰어
      M1 — `/` 를 붙이면 정상 하위 모듈 트리를 헛거절) → 정상 트리 시험에 gitlink `b` · `b-.go` 이웃. (P3-2) "같은 이름이면 git 은
      앞의 것을 읽는다" 는 부정확했다 — 제자리에 고쳤다(그 고침의 "git 의 diff 는 두 순서 모두 편집을 본다" 도 틀려 주장정확성
      재리뷰가 다시 고쳤다: (편집, base) 순서의 diff 는 삭제 하나다). 리뷰어가 지킨 공격: 정상 트리의 바이트 `0x7f` · `0xc3a9` · `0xff` 이름 · 디렉터리처럼 이름 붙은 gitlink · sha256 ·
      실제 이력의 걷기 출력 바이트 동일(8d9731c1 · 54004f44) · 리뷰어가 잰 설정들(`diff.relative` · `orderFile` · `renames=copies` ·
      `colorMoved` · `GIT_CONFIG_PARAMETERS` 로 준 `core.quotePath` 등)이 판정을 안 바꿨다(리뷰어 보고 — 이 세션은 다시 재지 않았다).
      증거: 시험 372 → **377**(트리 거절 둘 · 정상 트리 둘 · 형식 하나 · 구조 시험에 깃발 다섯). 최종 시험 파일을 HEAD(54004f44)
      코드에 붙이면 **20 경우 빨강**(실패 17 · 오류 3) — 트리 거절 13(뿌리 11 · 커밋 대상 2) · 형식 6(noprefix 둘 · 색 둘 ·
      interHunkContext · `-u3`) · 구조 하나; 같은 사본에 새 코드면 377 초록(skip 2 — 사본에 `docs/WORKFLOW.md` 가 없다). 정상 트리
      시험 둘(초기 git 의 `100664` 포함)과 `mnemonicPrefix` 는 HEAD 에서도 초록이다(거절이 정상 입력을 안 건드리는지 · 원래
      무해했는지를 잰다). 변이 189 → **212**(AK 23 더함 · AJ45 앵커를 새 조건으로 옮김). 첫 판(재리뷰 전 코드) 209 중 208 — AK6
      생존 → 시험 수정 → 재실행 CAUGHT. 최종 코드에서 212 전부 다시: 창 `0:53` · `53:106` · `106:159` · `159:212` 각 **53/53**,
      대조군 양끝 GREEN `Ran 377`. 그 판이 루트 파일시스템이 가득 찬 시간(병행 세션 통보, 약 13:30~15:30 KST)과 겹쳐 실패 수가
      첫 판보다 커진 24 변이 중 **1 넘게** 커진 15(정확히 +1 인 9 — AE4 · AH18 · AJ10 · AJ18 · AJ42 ·
      AJ44 · AJ46 · AJ48 · AJ49 — 는 새 시험 하나로 설명된다)와 새 AK21–AK23 을 디스크가 돌아온 뒤 다시 돌렸다 — **18/18 CAUGHT**, 실패 수는 첫 판과 같다(부풀린
      것은 디스크 오류). `make lint` rc 0(20:47 에 다시 — Go 만 본다: `gofmt` · `go vet`, 이 로트의 Python 은 밖) · `make sdd-test` **rc 0** — logic-map **452**(377 + 스위트 밖 75) · sdd 71(앞 로트의
      bootstrap drift 가 없어졌다 — `.claude/CLAUDE.md` · `.codex/agents.md` 가 54004f44 에서 맞춰졌다) · sdd-history 22 · pm 16 ·
      deploy 18 · scripts 15 · `go test ./tools/logic-map` 통과 · `make test-seams` rc 0(**108 패키지** — ok 100 · 시험 파일 없음 8,
      `-trimpath` 없이; 재리뷰 뒤의 Python 변경 전에 돌렸다 — Go 시험은 Python 을 안 읽는다) · `openspec validate --all --strict` **58/58** · a122 게이트: 54004f44 + 이 로트의 파일(연결 워크트리) **rc 0**(`required 0`); A/B 워크트리를 세운 8d9731c1 에서는 rc 1(브랜치 끝은 그 뒤로도 움직였다) —
      요구 5 가 전부 병행 세션의 a113 커밋 `ed3cb1d7`(`internal/strategyprojectionrpc/`)이다: 워킹트리 대상 창이 base 뒤에 착지한
      남의 커밋을 요구로 삼는 5.6/5.7 의 한계 그대로(이 로트와 무관 — A/B 에서 a122 도 SAME) · 좌표 게이트 **rc 0**(열린 task 의 좌표 0 · 뜬 번들 0 · 낡은 범위 0 · 낡은 수 0) ·
      실물 `main()` A/B **126/126 SAME · DIFFERENT 0**(기준 54004f44 = 0ece1074 코드, **고정한 연결 워크트리**에서 — 병행 세션이 같은
      브랜치에 커밋해 첫 판은 "HEAD moved" 로 어긋났다; 둘째 판의 DIFFERENT 5 는 전부 "No space left on device" 라 지우고 다시
      돌렸다; 나머지 121 중 116 은 git 호출 수가 7.5.34 판과 같고 5 는 **늘었다** — a063 2 → 10 · a109 21 → 22 · a113 19 → 21 ·
      a119 0 → 19 · a122 18 → 19, 그 뒤 편집된 change 들이다; 디스크 오류로 일찍 멈춘 판은 수가 **줄어드는** 모양이라 SAME 을
      공허하게 만든 것은 없다) — git 호출
      **2,462 → 2,462** · `ast.json` 읽기 9,087 → 9,087, change 마다 같다(새 프로세스 0). 벽시계는 디스크 사고와 겹쳐 적지 않는다.
      **남은 것.** 7.5.26 에서 이름에 `"` 나 `\` 가 든 `*.go` 의 헛거절(git 이 머리 줄을 인용한다 — `cannot load existing base
      file … "a/we\"ird.go"`, 이 로트에서 실측). 탭 든 이름은 설계대로 거절.
      **정정 — 주장정확성 재리뷰(0ece1074), 같은 실수 아홉 번째.** (README 의 4,093 정정은 이 로트의 주장정확성 재리뷰가 한
      번 더 고쳤다 — 아래 열 번째.) 7.5.34 의 A/B 분해("116 change 가 100~230 씩")를 JSON 에서
      다시 세어 제자리에 고쳤다(115 가 8~228 · 10 그대로 · 1 늘었다). README 의 "트리 4,093" 과 "`*.go` 마다 blob 을 쓴다" 를
      고쳤다. 7.5.34 의 "base 바이트 위조 둘의 헛거절" 은 이름 댄 거절이다(review.md 표와 맞췄다).
      **정정 — 이 로트의 주장정확성 재리뷰(최종 코드), 같은 실수 열 번째.** 거짓 둘: README · review 의 "4,093 은 어느 커밋의
      수도 아니다"(ff19be01 의 하위 트리 수였다 — 7.5.34 의 기준) · FLM B19 · B20 의 "git 은 하위 트리의 것만 본다"(둘 다 낸다).
      BTM 의 변이 시험 수 셋을 첫 판(376) 값으로 "최종" 이라 적었다(AJ42 · AJ46 35 → 36 · AJ49 150 → 151 · AJ18 19 → 20). 부정확
      열하나: 부풀린 변이 수("15" → 24 중 1 넘게 15) · 색 설정의 "이 change 이전부터 헛거절"(base 에서는 **조용한 `[]`** — 7.5.31 정정이
      이미 적은 것을 되풀이했다) · 트리 수에 걷기 방문과 검증하는 고유 객체를 섞음 · "브랜치 끝(8d9731c1)" · A/B 121 대조의 방법 ·
      트리 전수의 시각과 방법 · `/` "인" 이름 · AK6 재실행의 시험 판 · `make lint` 의 시각과 범위 · FLM B22 의 "같은 종류의 중복" ·
      "git 의 diff 는 두 순서 모두 편집을 본다". 전부 제자리에 고쳤고 새로 적은 값(고유 트리 1 + 3,821 · 1 + 3,716 · ff19be01 4,093 ·
      base 의 색 `[]` · 전수 13,980 · lint)은 이 세션이 다시 쟀다. 코드의 바뀐 것은 주석 · docstring 셋뿐이다 — 잰 판본과 AST
      (docstring 을 뺀 `ast.dump`)가 같다.
- [ ] 7.5.32 **P0(이 change 이전부터) — 함수 지도를 뽑는 파서를 판정받는 워킹트리에서 빌드한다 (7.5.25 적대 재리뷰 F6).**
      `go_functions` 는 `go run ./tools/logic-map` 으로 **워킹트리의** `extract_go_ast.go` 를 돌린다. 그 파일도 추적 `*.go`
      라서 게이트가 판정하는 대상인데, 그것을 고쳐 `--list` 가 `[]` 를 내게 하면 다른 파일의 편집과 **자기 편집까지**
      요구에서 사라진다(리뷰어 재현: mock 이 아니라 진짜 파서로, 새 코드와 부모 모두 `[]`). 7.5.8 은 원장 밖의 읽기이고
      이것은 판정 도구 자체의 신뢰 경계다 — `check_analysis.py` 를 고치는 것과 같은 경계지만 이것은 게이트가 지켜야 할
      `*.go` 다. 근본은 파서를 **base 커밋에서** 빌드하는 것(`git archive <base> tools/logic-map` → 임시 모듈)이고 판정
      비용과 기존 번들의 재현성이 달라지는 설계 변경이라 **사람 결정** 대기.
- [ ] 7.5.33 **P2(이 change 이전부터) — 부분 클론에서 게이트가 base blob 을 가져온다 (7.5.25 적대 재리뷰 F12). 위조 절반은 7.5.34 가 닫았다.**
      ~~base 쪽은 여전히 git 의 저장소를 믿는다(F5)~~ — 7.5.34 가 base 의 커밋 · 트리 · blob 을 해시 검증해 읽는다(하위 트리
      위조까지, 시험 `test_a_forged_base_object_is_refused`). (7.5.34 정정: 이 항목의 첫 판본은 "7.5.31 의 git 밖 대조는 워킹트리
      쪽만 본다 … pack 을 심으면 요구가 빈다(리뷰어 재현)" 라 적었는데, 그 재현은 리뷰어가 `9bc103fa` 에서 잰 것이고
      ff19be01 에서는 규칙 1 의 **이름 댄 거절**이었다 — 옮겨 적으며 새 코드에서 다시 재지 않았다. 구멍 자체는 실재했다:
      편집 내용 위에 미끼 줄을 얹은 위조는 ff19be01 에서도 `[]` 였다, 주장정확성 재리뷰 실측.)
      **남은 절반**: 부분 클론(`--filter=blob:none`)에서 base blob 이 없으면 git 이 promisor 로 **fetch** 하고 실제
      `.git/objects/pack` 에 pack 이 생긴다(2 → 3, 판정은 맞다). 7.5.34 뒤로 base blob 을 읽는 호출은 `_verified_objects` 의
      `cat-file --batch` 하나다. 이 기계의 git 2.43 은 `--no-lazy-fetch` 를 모른다(`git --no-lazy-fetch version` rc 129 실측) — 막는 방법을 따로 재야 한다.
- [x] 7.5.29 **P2 — `_self_repair_commits` 도 git 출력을 `str.splitlines()` 로 자른다 (7.5.28 에서 셈).**
      7.5.28 이 판정 diff 에서 고친 것과 같은 값이다: `git log` 의 커밋 덩이를 `block.splitlines()` 로 잘라
      커밋 sha 와 경로를 가른다. 경로에 U+2028 등이 있으면 한 경로가 여럿으로 갈린다. 그 결과가 자기 수리
      판정을 바꾸는지는 **재지 않았다** — 다른 판정이라 FLM 없이 고치면 절차 위반이므로 task 로 연다.
      **닫음(2026-09-27).** 먼저 쟀다 — 판정을 **바꾼다**, 두 방향: `.go` 가 아닌 `note.go<U+2028>txt` 에 깃발이 서고(거절 늘어남), git 이 인용하는
      `*.go`(`"` · `\` · 제어 문자)에는 깃발이 안 섰다(거절 사라짐 — 이 항목이 안 센 반대편). `git log -z` 로 받아 바이트로 읽고 모양이 다르면 결함.
      행동 시험 넷(RED 빨강) + 모양 시험. 전수 A/B(`7529_repairs_ab.py`) change 128 · 깃발 392 · SAME 128.
- [x] 7.5.26 **P2 — 파서가 `a/`·`b/` 접두사를 글자로 믿는다 (7.5.23 재리뷰 2026-09-23, 적대).**
      `changed_existing_functions` 가 `--- a/<이름>` 에서 `removeprefix("a/")` 로 이름을 정한다.
      `diff.noprefix=true` 면 헤더가 `--- <이름>` 이라 무해하지만, 경로가 **정말** `a/`·`b/` 로
      시작하면 실제 구성요소를 잘라 낸다(`--- a/a/hot.go` → `a/hot.go` 가 되어야 하는데 `noprefix`
      아래에서는 `--- a/hot.go` → `hot.go`). 미끼 파일이 있으면 요구가 조용히 사라지고, 없으면
      `cannot load existing base file` 로 거짓 차단이다. `diff.mnemonicPrefix=true` 는 워킹트리 대상에서
      `c/`·`w/` 를 내서 게이트가 **아예 못 돈다**(fail-closed, 이 change 이전부터). (7.5.31 정정: 7.5.25 뒤로 두 diff 가
      `--cached` 로 돌아 접두사는 `c/`·`i/` 다 — 거절 문장이 `c/x.go` 를 댄다, 주장정확성 재리뷰 실측.)
      저장소 노출 0(`git ls-files | grep -cE '^(a|b)/'` = 0). 근본은 접두사를 글자로 믿지 말고
      `--src-prefix=`·`--dst-prefix=` 로 **우리가 정하는 것**이다.
      **(2026-09-24, 7.5.28 재리뷰가 더 셌다 — 전부 이 change 이전부터, 7.5.25 의 스냅숏이 안 닫는다.)**
      출력 **형식**을 바꾸는 것은 워킹트리 바이트의 문제가 아니라 파서의 문제다: `diff.noprefix=true` 와
      `GIT_LITERAL_PATHSPECS=1`(`'*.go'` 가 글자 그대로의 이름이 되어 diff 가 빈다)은 요구를 **조용히** 지우고,
      `diff.mnemonicPrefix=true` · `color.ui=always` · 이름에 `"` 가 든 `*.go` 는 **거짓 차단**한다. 저장소 노출 0.
      (7.5.34 정정: `GIT_LITERAL_PATHSPECS=1` 의 "조용히 지운다" 는 base 와 `9bc103fa` 에서만 참이었다 — ff19be01 에서는
      7.5.31 의 규칙 2 가 **거절**했고 그 사유가 "the object store" 라 원인을 잘못 댔다, 주장정확성 재리뷰 실측. 7.5.34 의
      목록과 두 diff 는 pathspec 을 안 받아 이 변수가 판정에 안 닿는다 — 시험 `test_pathspec_variables_do_not_change_the_judgement`.)
      (7.5.31 정정 — 주장정확성 재리뷰가 쟀다. 앞 판본은 `diff.noprefix=true` 도 "요구를 조용히 지운다" 에 넣었는데 **재지
      않고 적은 것이었고 틀렸다**: 평범한 `x.go` · `pkg/x.go` 는 요구가 선다. 문제는 경로가 `a/`·`b/` 로 시작할 때다 —
      `a/x.go` 는 거짓 차단, `b/x.go` 는 요구를 **엉뚱한 경로** `x.go` 로 적고, 그 경로에 미끼가 있으면 요구가 빈다(7.5.25
      적대 재리뷰 F8). 이 항목 첫 문단이 이미 그렇게 적고 있었다. 그리고 `color.ui=always` 는 **이 change 의 base 에서는
      조용히 `[]`** 였고 지금 코드에서 거짓 차단이 됐다 — "이전부터 거짓 차단" 이 아니다. 7.5.25 적대 재리뷰가 하나를 더
      셌다: `GIT_GLOB_PATHSPECS=1` 이면 `'*.go'` 가 `sub/x.go` 에 안 맞아 요구가 조용히 빈다(F7, 이 change 이전부터).
      — 7.5.34 정정: 그 문장은 base 와 `9bc103fa` 에서 참이고 **이 문장을 적은 ff19be01 에서는 거짓**이었다 — 같은 커밋의
      규칙 2 가 "the object store answers for sub/x.go as if it were still there" 로 거절했다. 리뷰어가 앞 커밋에서 잰 것을
      새 코드에서 다시 재지 않고 옮겼다. 7.5.34 가 닫았다.)
      근본은 같다 — 판정 diff 의 형식(`--src-prefix=a/ --dst-prefix=b/ --no-color`)과 pathspec 해석
      (`--glob-pathspecs` 또는 환경 변수를 지운 자식 환경)을 **우리가 정한다**.
      **(7.5.35 — 대부분 닫혔다.)** 판정 diff 가 `--src-prefix=a/ --dst-prefix=b/ --no-color --inter-hunk-context=0` 을 받고
      자식 환경에서 `GIT_DIFF_OPTS` 를 지운다 — `diff.noprefix`(`a/` · `b/` 둘 다) · 색 둘 · 훅 합치기 · `-u3` 이 요구를 바꾸지
      않는다(시험 `test_the_judged_diff_format_does_not_follow_the_user_config`). pathspec 쪽은 7.5.34 가 닫았다.
      `diff.mnemonicPrefix` 는 트리 둘을 견주는 diff 에서 접두사를 안 바꾼다(이 로트에서 실측). **남은 것**: 이름에 `"` 나 `\`
      가 든 `*.go` 의 헛거절(git 이 머리 줄을 인용한다) — 저장소 노출 0.
      **닫음(2026-09-27, 잔여).** 7.5.13 과 같은 수리로 닫혔다 — `"` · `\` 가 든 `*.go` 는 이제 제 이름으로 판정된다(헛거절 없음). 닫힌 범위:
      git 이 인용하는 바이트 전부(`< 0x20` · `"` · `\` · `0x7f`, 단 `\n` · `\r` · `\t` 은 가드가 **이름 대고** 거절 — 7.5.41). 머리 줄의 접두사도 이제 글자로
      믿지 않고 렌더링과 통째로 대조한다.
- [ ] 7.5.30 **P2 — 시험 스위트가 공유 go 빌드 캐시를 판마다 약 375 MB 씩 불린다 (7.5.25 변이 중 실측, 2026-09-24).**
      `go_functions` 가 임시 저장소마다 `go run ./tools/logic-map` 을 부르고, go 의 캐시 키에 **디렉터리 경로**가
      들어가서 같은 코드가 경로마다 새 항목이 된다. 전용 `GOCACHE` 에서 잰 값: 스위트 한 판 +423 MB, 데운 뒤 둘째
      판 **+375 MB**; `GOFLAGS=-trimpath` 로는 둘째 판 **+1 MB**. 변이 160 개 판(대조군 포함 스위트 약 170 판)이면
      60 GB 가 넘는다 — 2026-09-24 에 디스크가 두 번 0 이 된 원인이고(주장정확성 리뷰어가 공유 캐시 약 110 GB 를
      지웠다), 7.5.25 의 첫 변이 판들도 공유 캐시를 188 → 203 GB 로 불렸다. 하네스는 이제 `-trimpath` + 전용
      `GOCACHE` 로 돈다(`analysis/harness/75_mut.py` 머리 주석). 게이트 쪽 수리(`go run -trimpath` 또는 시험의
      자식 환경)는 `go_functions` 내부 변경이라 FLM 이 먼저다 — task 로 연다. 생산 게이트는 저장소 루트 한 곳에서
      돌아 경로가 같으므로 이 증식은 시험·하네스에서만 생긴다(재지 않았다 — 추론이다).
- [x] 7.5.25 **P0 — 워킹트리를 다시 쓰는 것은 두 투영을 **함께** 거짓말하게 한다 (7.5.23 재리뷰 2026-09-23,
      적대).** 교차 검사는 **어긋남**만 본다. `filter.<x>.clean` 은 `git diff` 가 비교하는 워킹트리 바이트
      자체를 바꾸므로 `--numstat` 이 **아무 레코드도 안 낸다** — 어긋남이 없어 거절도 없고 `required` 가
      조용히 빈다. 실측(워킹트리 대상, 게이트의 **기본 모드**): 디스크에 `return 999` 가 그대로 있는데
      `numstat` 이 빈 글자이고 `required` 가 `[('x.go','Stop')]` → `[]`. 같은 계열로 `git update-index
      --skip-worktree` · `--assume-unchanged` 도 편집을 감춘다(커밋된 작업은 못 감추므로 그쪽은 낮다).
      공격 표면은 7.5.22 와 같다 — 추적 안 된 `.gitattributes` 한 줄 + 로컬 config 키 하나.
      **사람 결정 대기.** 닫으려면 게이트가 **git 의 어떤 투영도 안 믿고** 자기가 판정하는 바이트를
      직접 대조해야 한다(base blob 대 `go_functions` 가 읽는 워킹트리 바이트, 또는 판정 `git diff` 를
      `-c filter.<x>.clean=` 로 중화). 둘 다 이 change 밖의 판정을 바꿀 수 있고 비용도 다르다 —
      규칙 변경이라 7.5.4 · 7.5.6 · 7.5.7 과 같은 줄에 둔다.
      (2026-09-23: 사람이 **싼 막기**(선택지 2)를 골랐고 7.5.27 이 알려진 문 다섯을 닫았다 — clean·process
      필터 · fsmonitor · assume-unchanged · skip-worktree. **class 는 이 task 에 그대로 남는다**: 아직 모르는
      워킹트리 재작성 문은 여전히 두 시야를 함께 속일 수 있다.)
      **(2026-09-24, 7.5.28 이 잰 것 — 설정이 전혀 없어도 뚫린다.)** 같은 크기로 파일을 지우고 다시 쓰면 inode
      번호가 **재사용**되고, 마지막 인덱스 쓰기와 **같은 초** 안이면 ctime 도 같다. 여기에 mtime 을 과거로 돌리면
      git 의 **기본** stat 검사(`checkStat=default` · `trustctime=true` 를 고정한 상태)도 파일을 안 다시 읽는다 —
      실측 **20 판 중 17 판** 편집이 감춰졌고, 17 판 전부 inode 가 재사용된 판이었다. git 의 stat 캐시를 **믿는 것
      자체**가 문이라 명령줄 고정으로는 못 닫는다. 결정에 쓸 선택지 하나를 적어 둔다: 인덱스를 **임시 사본**으로
      복사해 `GIT_INDEX_FILE=<사본> git update-index -q --really-refresh` 로 내용을 다시 해시한 뒤 두 diff 를 그 사본
      위에서 돌리면 실제 인덱스를 안 건드리고 stat 캐시를 우회한다(비용: 판정마다 추적 `*.go` 전수 해시).
      **닫음(2026-09-24, 사람이 선택지 1 — class 수리 — 를 골랐다). 생산 Go 변경 0.** 7.5.27 · 7.5.28 이 문을 하나씩
      닫는 동안 재리뷰가 매번 하나를 더 찾았다(마지막: `working-tree-encoding=UTF-7` 의 이중 표현 — `+AHk-x` 가 `yx` 로
      해독돼 git 은 같다 하고 Go 는 `+AHk - x` 를 계산한다). 그래서 **투영을 안 쓴다**: `_worktree_snapshot` 이 인덱스가
      추적하는 `*.go` 마다 디스크 바이트를 깔때기(`_read_regular`)로 읽고, 그 blob 을 **임시** 객체 저장소에 loose 객체로
      쓰고(`git hash-object` 대신 — git 이 파일을 다시 읽지 않는다), 그것을 가리키는 **임시** 인덱스를 만든다. 가드와
      판정의 두 diff 는 `--cached` 로 그 인덱스를 base 와 견주고(`_compared` 하나가 쌍을 준다), 판정의 현재 쪽도 **같은
      바이트**로 함수 지도를 뽑는다. `_git_view_pins` · `_ident_go_paths` · `_hidden_by_index_flags` 는 지웠다.
      위의 선택지(인덱스 사본 + `--really-refresh`)를 안 쓴 까닭: 그것은 여전히 git 이 워킹트리를 **변환해** 해시하므로
      필터 · `ident` · `working-tree-encoding` 을 못 닫는다.
      **실측으로 잰 것**: 새 문 둘이 부모(`cfe8d08b`)에서 요구를 비운다 — UTF-7(재리뷰 F1) · **설정 없는** stat 캐시(같은
      inode · 같은 크기 · mtime 되돌림 · 같은 초의 ctime, **10/10** — 7.5.28 의 20 판 중 17 판은 inode 재사용이 우연이던
      판이고 이 픽스처는 결정적이다). 새 코드에서는 둘 다 요구가 선다. 원형을 일곱 모양(편집만 · clean · assume · skip · stat · UTF-7 ·
      FIFO)에 진짜 git 으로 돌려 실제 저장소의 `.git/objects` 파일 목록과 `.git/index` 바이트가 그대로인 것도 봤다(process
      필터는 원형에서 안 돌렸다 — 시험 `test_a_process_filter_does_not_hide_a_worktree_edit` 가 잰다).
      **실제 저장소에 새는 셋을 쟀다**: `update-index` 는 `post-index-change` 훅을 **띄운다**(실측) → `core.hooksPath=/dev/null`;
      split index 는 `sharedindex.*` 를 `.git` 에 쓴다 → `core.splitIndex=false`; 인덱스를 **읽는** 순간 fsmonitor 명령이 뜬다
      — 스냅숏이 아니라 **실제** 인덱스를 읽는 `ls-files` 에서 떴다(첫 시험 실패로 알았다) → `SNAPSHOT_PINS` 를 그 호출에도.
      **새 거절(이름 댐)과 그것이 오늘 거절하는 정상 입력**: 충돌 중인 `*.go` · 인덱스 모드가 `100644`/`100755` 가 아닌 `*.go`
      (저장소 전수 **0 / 1,762**, 전부 `100644`) · 실제 객체 디렉터리 경로에 `:`·`"`(0) · FIFO 등 정규 파일이 아닌 추적 `*.go`
      (깔때기의 `NotRegularFile`). **달라지는 판정**(보수 방향만): CRLF 체크아웃이면 손대지 않은 함수까지 요구된다(실측:
      편집 0 → 요구 2) · `*.go` clean 필터 저장소는 모든 추적 `*.go` 가 바뀐 것으로 읽힌다 — 이 저장소 노출 0(`core.autocrlf`
      미설정 · 필터 0 · 속성 파일 0 · `\r` 든 추적 `*.go` 0). **그대로 두는 것**: skip-worktree 인데 파일이 없으면 sparse 로
      안 꺼낸 것으로 읽는다 — 그 자리에서 **지운** 것은 구별되지 않는다(7.5.27 도 같았다, 잔여). 출력 **형식** 문은 7.5.26.
      **덤으로 좁힌 것**: 7.5.8 의 "`go_functions` 의 `go run` 이 워킹트리 Go 바이트를 원장 밖에서 읽는다" 중 **워킹트리 대상의
      현재 쪽**은 이제 스냅숏(깔때기 → 원장)의 바이트를 임시 파일로 읽는다. 커밋 대상 · base 쪽 · 나머지 자식 프로세스 읽기는
      그대로다 — 7.5.8 은 열린 채다.
      비용: 스냅숏 한 번에 추적 `*.go` **1,762 개 · 18.5 MB**, 0.39 초(부하 아래) · 프로세스 **3** 개(`rev-parse` · `ls-files` ·
      `update-index`, 편집 수와 무관).
      증거: 시험 348 → **357**(RED — 새 시험 파일을 부모 코드에 붙이면 두 클래스에서 **22 경우 · 함수 17** 이 빨갛다: 실패 9 ·
      오류 8; 문 둘은 **실패**=요구가 빈다) · 변이 **160**(AF1~AF11 · AG2~AG5 · AG7~AG9 는 겨눌 코드가 없어 뺐다, AH1~AH26 을
      더했다, 앵커 전부 1 회): `0:40` · `40:80` · `80:120` 각 **40/40** · `120:160` 첫 판 **37/40** — 생존 `AH16`(blob id 가
      아님) · `AH17`(모두 씀)은 판정이 그대로라 우연이 지키던 것(안 바뀐 1,762 파일이 "바뀐 파일" 로 오르고 판마다 18 MB 를
      쓴다), `AH19`(커밋 대상도 스냅숏)는 계측기 "못 쟀다" — 판정은 같고 원장만 다르다 → 시험 둘을 더해 `147:151` **4/4**.
      `AH7`(`read_bytes()` 로 비켜 읽기)은 FIFO 시험에서 스위트를 **멈췄다**(`wait_for_partner`) — 하네스가 "환경 의심" 으로
      표시해 CAUGHT 로 안 셌다. 따로 재니 깔때기 시험 둘이 잡고 FIFO 시험이 멈춘다(깔때기가 막던 그 멈춤). 하네스에는
      멈추지 않는 판본(`_opened_bytes` 로 원장만 비켜 감)을 넣고 `138:139` **1/1**. 대조군 모든 창 양끝 GREEN.
      `lint` rc 0 · `sdd-test` rc 0(logic-map 423 → **432**) · `test-seams` rc 0(**108 패키지**) · `openspec validate --all --strict` **58/58** ·
      a122 게이트 **rc 0**(a112 `required 64` 그대로) · 좌표 게이트 **rc 0**(머리말 범위 불일치 1 줄은 7.5.2.2 뒤로 안 건드린 `normalized_source` 의 기록 — 판정 밖) · 실물 `main()` A/B **126/126 SAME · DIFFERENT 0**(기준 `cfe8d08b`) — git 호출 20,826 → **20,710(−116)**, `ast.json` 읽기 9,072 → 9,072, 벽시계 1,055.5 → 1,080.3s · 1,053.8 → 1,090.7s(+2.4~3.5%, 워킹트리 대상마다 추적 `*.go` 전수를 읽고 해시하고 끝의 재확인이 다시 읽는 몫).
      **측정 사고 하나**: 변이 판을 `nohup &` 로 띄운 뒤 `ps | grep` 이 0 을 내서(rtk 가 걸렀다) 다섯 판을 더 띄웠고 **아홉 판**이
      겹쳤다. 그 판들이 공유 go 캐시를 188 → 203 GB 로 불려 `/` 가 8 GB 남짓까지 떨어졌다 — 원인을 재어 task 7.5.30 을 열었고,
      판들을 죽이고 결과를 버린 뒤 `-trimpath` + 전용 캐시로 다시 돌렸다. 위 숫자는 전부 그 뒤의 판이다.
- [ ] 7.5.4 **P0(앞 로트 전부터) — `ast.json` 의 구조가 소스에서 다시 유도되지 않는다 (7.5.2.1 재리뷰, 보안 전문가
      재현).** 게이트가 `ast.json` 을 소스에 묶는 것은 **파일 전체의 sha256** 하나다. 분기 · 반환 · 호출 · 시작/끝 ·
      서명은 저자가 적은 그대로 믿는다 — 분기 셋인 함수의 `ast.json` 에 `branches: null` 을 적고 B1 한 줄로 5단계가
      통과했다(기록 · 게이트 rc 0 재현). 추출기를 게이트 안에서 그 리비전에 돌려 구조 칸을 대조하는 것이 근본인데,
      비용(대상마다 `go run`)과 기존 번들 3,048 의 재추출 여부가 걸린 **사람 결정**이다.
- [ ] 7.5.5 **INVESTIGATE(앞 로트 전부터) — `base-commit.txt` 를 워킹트리에서 읽는다 (같은 재리뷰, 적대).** base 는
      proposal freeze 에 고정되고 spec 이 불변을 요구하지만, 게이트는 커밋이 아니라 디스크의 글자를 읽는다 — 커밋 안 한
      편집이 창의 시작을 옮길 수 있는지 재 볼 것.
- [ ] 7.5.3 **P1 — 이름 바뀐 고정 소스가 V1 을 다시 연다 (같은 재리뷰, 7.2.2 부터 있던 것).** 가드 7 은
      고정 소스를 **같은 경로**에서만 비교한다. 순수 `git mv` 뒤에는 base 쪽이 없어서(`None`) "바뀌었다"
      로 읽히고, FLM 을 먼저 커밋하고 번들을 안 갱신한 V1 이 기록되어 `required 0` 으로 통과한다
      (두 출처가 독립적으로 재현, `a5c4bc77` 에서도 같음). 규칙 변경이라 **사람 결정** 대기.
- [ ] 7.5.6 **P1 — 열거형 호출 감사는 쉬고 있는 파일로도 꺼진다 (7.5.2.2 재리뷰 2026-09-20, 보안).** 타이밍 없이,
      커밋된 · 사람 눈에 정상인 표 파일 하나로 감사가 꺼진다(전부 실측): UTF-8 **BOM** · `\xff` 한 바이트 · UTF-16 ·
      좌표 없는 행 하나. BOM 판은 순수 UTF-8 이라 git·GitHub·에디터가 아무것도 말하지 않는다. `role_check` 가
      표를 **무엇으로 세는가**(해독 규칙 · 행 판별)의 문제이고 7.5.2.x 보다 오래됐다. 고치면 이 change 밖의 번들
      3,000여 개의 판정이 달라질 수 있어 **사람 결정** 대기 — 7.5.2.3 은 이것을 손대지 않고 문장만 좁혔다.
- [ ] 7.5.7 **P1 — `revision: base` 는 파일 수준 소스 해시에 안 묶인다 (7.5.2.4 재리뷰 2026-09-21, 보안 · 내가 정정).**
      `validate_target` 의 `revision == "base"` 갈래는 소스 해시를 안 본다. **정정**: `required` 에 든 대상은
      `_verdict` 가 `ast.json` 의 `source_sha256` 을 `base_hash` 와 대조하므로 묶여 있다 — 안 묶이는 것은
      `required` **밖**의 `base` 번들이다. 묶으려면 base 커밋의 파일 바이트와 대조하면 되는데, 전수 실측에서
      **266 중 18** 이 오늘 안 맞는다 → 이 change 밖의 번들을 새로 거절한다. **사람 결정** 대기.
      귀속(7.5.22 재측정, `197a0355`): **a112 13 · a063 2 · 아카이브 a060 2 · 아카이브 a044 1** —
      활성 15 · 아카이브 3 이다. 앞 판본의 "a063 2 · a112 16" 과 "그중 16 은 이 브랜치에 살아 있는 a112" 는
      둘 다 거짓이었다. 재현: `analysis/harness/7522_stale.py`(7.5.22 에서 커밋했다).
- [ ] 7.5.8 **P0 — 판정 입력의 일부를 자식 프로세스가 읽어 원장 밖에 있다 (7.5.2.3 재리뷰 2026-09-21, 출처 둘 + 나).**
      `check_analysis.py` 의 `subprocess.run` 은 **열여섯 자리**이고 원장 항목은 **0** 이다(7.5.22 재측정).
      판정 입력을 읽는 것만 꼽아도 `changed_existing_functions` 의 `git diff` · `base_file` 의 `git show` ·
      `go_functions` 의 `go run`(**워킹트리 Go 바이트**) · `_safe_changed_go_paths` 의 `git diff --numstat` ·
      `_committed_many` 의 `git cat-file` · `_recording_refusal` 의 `git diff --quiet` · 역사를 걷는 **열** 자리 ·
      그리고 `execution_baseline.validate` 의 `read_bytes`·`scandir`·`lstat` 이다. 판정 중 추적 Go 파일을 고치면
      생산 CLI 가 **rc 0 `evidence complete`** 를 내고 재실행은 rc 1 이다(두 출처 독립 재현).
      노출(a112 실측): `required` 를 정하는 Go 파일 **32** 중 **13** 은 원장에 이름조차 없고,
      나머지 19 도 **인용·시험 색인 때문에** 있는 것이지 `go run` 이 읽어서가 아니다. 자식 프로세스의
      읽기로 원장에 남는 항목은 **0** 이다(원장의 `kind` 는 `file`·`dir`·`glob`·`kind` 넷뿐).
      원장의 총 항목 수는 안 적는다 — 하네스 사본이 든 모집단의 함수라 도는 동안 바뀐다(7.5.23 정정). (7.5.22 정정: 앞 판본의 "Go 138개 / 고정 소스 37개" 는
      **재현되지 않는다**. 좌표 `:132`·`:85`·`:56` 도 전부 편집 전 파일을 가리키고 있었다.)
      7.5.2.3 의 README·VERIFY 가 "깔때기 넷뿐" 이라고 적은 것이 이것 때문에 거짓이었다(7.5.2.4 가 문장은 고쳤다).
      (2026-09-24, 7.5.25 가 좁혔다: **워킹트리 대상의 현재 쪽** Go 바이트는 이제 스냅숏이 깔때기로 읽어 원장에 남고,
      `go run` 은 그 바이트를 쓴 임시 파일을 읽는다. base 쪽(`git show`) · 커밋 대상 · 나머지 자식 프로세스는 그대로다.)
- [x] 7.5.9 **P1 — `test_index` 가 추적되지 않는 파일을 색인한다 (같은 재리뷰, 정확성).** `test_index` 가
      `*_test.go` 전수에서 `.git` 만 거른다. 시험 인용이 **머지에 영원히 안 들어갈 파일**로 충족될 수 있다.
      덤: `resolve_test_file` 의 맨이름 갈래가 사본이 생기면 `len(matches)!=1` 이라 **없던 거절**을 만든다.
      수: 추적 **953** 만 안정이고 디스크 수는 **움직인다**. 차이는 전부
      `analysis/harness/_work/*/extract_go_ast_test.go`(gitignore 된 하네스 사본)이고, **하네스가 도는 동안
      늘었다 준다** — 같은 날 실측으로 정지 상태 **963**(차이 10), 변이 하네스가 도는 중 **964**(차이 11,
      열한째가 그 판의 `75_mut_work.<pid>/`). (7.5.22 정정: 앞 판본은 이 수를 고정값처럼 적었다.
      오염원이 **이 change 자신의 하네스**이고, 세는 순간이 답을 바꾼다.)
      (2026-09-26 확인: `1d1e5ca7` 에서 추적 **959** · 디스크 **971** · 차이 12 전부 `_work/**/extract_go_ast_test.go` (glob 표기 정정: `*` 는 한 단계라 한 단계 깊은 `75_mut_work.*/logic-map/` 둘을 놓친다 — 12 = 10 + 2).)
      **닫음(2026-09-26).** 색인 · 인용 해소가 `git ls-files -z`(새 깔때기 `_tracked`, 원장 `tracked`)의 추적 파일만 본다 —
      `check_analysis.py` `test_index` · `resolve_test_file`(세 갈래 모두) · 거절 문장 `in any tracked file`. `_globbed` · `_pattern_outcome` 은 호출자 0 이 되어 지웠다.
      덤 **정정**: 맨이름 갈래는 사본이 생기면 "없던 거절을 만드는" 것이 아니라 해소를 포기해 **있던 거절을 지웠다**(permissive).
      거부하는 정상 입력: 번들 3,084 · 인용 6,121 **0**(`759_census.py`) · 시험 픽스처 5(`run_check` 가 인덱스에 올리게 고침).
      보수(독립 리뷰 둘, 2026-09-26): 이름 붙은 좌표의 `..` · 심링크 디렉터리를 풀어서 대조(첫 판의 회귀) · 아카이브는 이름 먼저 · `ls-files` 에 `SNAPSHOT_PINS`. 기록 review.md `## VERIFY — task 7.5.9 · 7.5.10 · 7.5.11 · 7.5.15 · 7.5.16`.
- [x] 7.5.10 **P1 — 목록·순회 지문이 충돌한다 (같은 재리뷰, 정확성 — 충돌쌍 실제로 만듦).** `_listing_outcome` 이
      `name\t{d|f}` 를 `\n` 으로 잇고 `_pattern_outcome` 이 경로를 `\n` 으로 잇는다. 탭·개행은 POSIX 이름에 합법이라
      서로 다른 두 목록이 같은 지문을 낸다. 아이러니: 같은 파일의 `_safe_changed_go_paths` 가 그 문자들을 거절한다.
      **닫음(2026-09-26).** 목록 지문 = 항목마다 이름 길이(8바이트) · 이름 · 종류 — 단사(`_listing_outcome`). 순회 지문은 함수째 없어졌고
      추적 목록 지문은 git `-z` 바이트의 해시다. RED: 충돌쌍 둘이 `1d1e5ca7` 에서 같은 지문(`7510_collide.py` `COLLIDE`).
- [x] 7.5.11 **P1 — 깔때기 **안**에 조용한 건너뛰기가 둘 남았다 (같은 재리뷰, 정확성 · 시험품질).**
      `_listing_outcome` 의 `child.is_dir()` 가 `OSError` 를 삼켜 stat 안 되는 항목을 파일로 분류하고
      (`_read_evidence` 는 **이미 디렉터리로 분류된** 것만 `unlistable` 에 넣는다 — 그 함수 docstring 이 바로 그
      위험을 한 층 위에서만 고쳤다고 적고 있다), `_pattern_outcome` 의 `Path.rglob` 이 못 읽는 하위 트리의
      `OSError` 를 삼킨다. 뒤엣것은 결과가 보수적이라 재확인 구멍은 아니지만 시험이 0 이다.
      (`is_dir()` 가 삼키는 것은 판본 의존이다 — 이 저장소의 3.12 에서 실측할 것, [[python-behaviour-differs-by-version]].)
      **닫음(2026-09-26).** 3.12.3 실측(`7511_isdir.py`): `is_dir()` 이 끊긴 링크 · 고리 · 사라진 이름을 `False` 로 삼키고(3.14 는 권한까지),
      `rglob` 은 네 판본 모두 못 읽는 하위 트리를 건너뛴다. 종류를 `os.stat` 으로 묻고 못 물으면 `UnstatableEntry` 로 이름을 댄다
      (`FileNotFoundError` 가 아니라서 면제 경로로 안 샌다). `rglob` 절반은 7.5.9 가 순회를 없애 닫았다. 거부하는 정상 입력 0(디렉터리 3,183 · 항목 15,428).
- [ ] 7.5.12 **P1 — realpath ABA (같은 재리뷰, 적대).** 원장 키가 `normalized_source` 의 `os.path.realpath`
      뒤 경로라, 판정 중 심링크 디렉터리를 갈아끼우면 바이트가 같아 재확인이 통과한다. 저장소 노출 0.
- [x] 7.5.13 **P1 — `_safe_changed_go_paths` 가 git 이 인용하는 나머지를 놓친다 (7.5.2.3 재리뷰, 보안).**
      그 가드는 `\n\r\t` 만 거절하는데 git 은 `"` · `\` · 나머지 제어 문자도 인용한다(`core.quotePath=false`
      로도 안 꺼진다). `x"y.go` → `--- "a/x\"y.go"` → `removeprefix("a/")` 무효.
      **7.5.22 정정 — 앞 판본이 적은 사유가 틀렸다** (7.5.23 에서 다시 정정: 그 정정 자체가 앞뒤가 안 맞았다).
      정확히 적는다. `*.go` 이름에 git 이 인용하는 글자가 들어가면 결과는 **모양마다 다르다**:
      **탭이 든 이름을 편집**하면 `old_source` 가 인용된 글자가 되어 `cannot load existing base file …` 으로
      **거짓 차단**이고(7.5.22 의 rename 시험이 편집 전에 낸 메시지가 그것이다), 그 이름이 **rename 의 옛
      이름**이면 `--name-only` 이 그것을 아예 안 내서 가드가 **통과시켰다**. 즉 같은 결함이 한 모양에서는
      거짓 차단, 다른 모양에서는 가드 우회였다 — 앞 판본은 뒤엣것을 "permissive" 라고만 적고 앞엣것을
      그 증거로 들어서 두 문장이 서로를 부정했다. 7.5.22 가 rename 의 **양쪽 이름**을 가드 범위에 넣어
      우회 쪽을 닫았다 — 남은 것은 `"`·`\`·나머지 제어 문자이고, 그 결과는 **거짓 차단**이다.
      저장소 노출 0(117 change · 이름 97,235 실측).
      **닫음(2026-09-27).** 방향 (c): 판정의 이름을 머리 줄에서 유도하지 않고 `--numstat -z` 레코드(날 바이트)에서 고른다. 머리 줄은
      그 이름을 git 의 규칙으로 인용한 글자(`_git_header_path`)와 **대조**만 한다 — 틀리면 이름 댄 결함(막는 쪽). (a) 해독기는 틀리면 조용히 다른
      이름을 판정하고, (b) 거절은 7.5.24 가 받게 한 정상 입력(인용되는 이름의 새 파일)을 새로 거절해서 안 골랐다. RED: 편집 · 삭제 · 인용된 이름
      에서의 rename 은 헛거절, 인용된 이름**으로의** rename 은 새 쪽을 조용히 안 봤다(permissive — 이 로트에서 실측). 렌더러는 git 2.43.0 과 ASCII
      전수로 대조(`7513_quoted.py` · 시험). 센서스: 역사 전부 + 추적 `*.go` 1,816 중 인용되는 바이트 **0**(26e5bb3f). `\n\r\t` 거절은 그대로 — 7.5.41.
- [x] 7.5.14 **P2 — 구조 시험이 stat 계열을 안 본다 (같은 재리뷰, 정확성).** `test_check_analysis.py` 의 원시 집합 9개에
      `stat/lstat/exists/is_file/is_dir/is_symlink/access` 가 없어 네 번째 깔때기 `_kind` 를 아무것도 안 지킨다.
      살아 있는 우회: `changed_existing_functions` 의 `Path.exists` · `_recording_refusal` 셋. **7.5.2.4 에서 이 결함이 행동으로
      보였다** — Y13(`evidence.present` → `evidence.directory.exists()`)이 구조 시험을 그냥 지나갔다.
      덤: `owner.setdefault` 라 `("record_landing","open")` 면제가 그 함수 안의 **미래의 모든** `open` 을 함께 면제한다.
      `execution_baseline.py`(stat 16자리)는 아예 범위 밖이다.
      **닫음(2026-09-27).** 먼저 셌다(HEAD `3bb11b2d` AST): `changed_existing_functions` 의 `Path.exists` 는 7.5.34 가 이미 지웠고, 남은 우회는
      `_recording_refusal` 셋. 둘(`landing_file.exists()` → `_kind` · 빌림 표지 `.exists()` → 판정과 같은 `_read_regular` 갈래)은 깔때기로 옮기고, 하나
      (`landing_file.is_symlink()` — 쓰기 자리의 모양, `open("xb")` 가 다시 막는다)는 사유 달아 면제. 원시 집합에 stat 계열 · 경로 풀이를 더했고
      (이름 호출은 내장 `open` 만 — 지역 함수 `resolve` 가 걸리지 않게), 분류를 **깔때기 소속**(6 자리: `_opened_bytes` 둘 · `_listing_outcome` 둘 ·
      `_names_outcome` · `_kind_outcome`)과 **면제**(사유 · 자리 수)로 갈랐다. 덤: 면제는 (함수, **호출 형태**) · 자리 수 단위 — 같은 모양이 하나 더
      생기면 위반(시험). RED: 구조 시험 · 권한 없는 표지 · FIFO 표지가 편집 전 코드에서 빨강. 변이 MU1 · MU2 CAUGHT, 시험 쪽 TM3 · TM4 CAUGHT.
      **범위 밖(열린 채)**: `execution_baseline.py` 의 stat 계열 16 자리 — 이관 경로의 원장이 따로 없다. `normalized_source` · `resolve_test_file` 의
      `os.path.realpath` 는 면제로 적었다 — 7.5.12(realpath ABA) 의 부류.
      **수리(적대 리뷰 뒤, 2026-09-27).** (1) 계측기 대조 시험이 계측기를 안 부르고 규칙을 제 안에서 다시 구현했다 — `_disk_calls` 에서 `exists` 만
      빼는 변이 TM6 이 생존(Y13 모양 MU2 와 겹쳐도). `_disk_calls(source)` 로 갈라 탐침을 **그 함수로** 센다(탐침 이름은 글자 목록 — 원시 집합에서
      만든 첫 수리 판은 TM3 이 생존했다). (2) 원시 집합에 적대 리뷰의 이름 17 을 더하고(`group` 은 `re.Match.group` 16 자리와 못 갈라 뺐다),
      이름 호출을 `from … import` · 이름 대입 · 기본 인자 별칭까지(고정점) 푼다. 우회 탐침 31 모양 중 잡힘 24 · 한계 7(`Path.group` · `getattr` ·
      `eval` · `__builtins__[…]` · `map(open, …)` · 자식 프로세스 둘) — 한계는 `_disk_calls` docstring 에 적었다(`7514_probe_bypass.py`).
      (3) 깔때기 자리 수 대조 시험(TM5 생존). (4) second_site 를 "그 위반이 있다" 로 독립시켰다. TM3~TM7 · TM6+MU2 CAUGHT.
- [x] 7.5.15 **P2 — 판정 경로 subprocess 에 `timeout=` 이 없다 (같은 재리뷰, 정확성).** `check_analysis.py` 의
      `subprocess.run` **열여섯 자리 중 둘** — `base_file` 의 `git show` 와 `_safe_changed_go_paths` 의
      `git diff --numstat` — 에 `timeout=` 이 없다(7.5.22 재측정, 나머지 열넷은 있다). `execution_baseline.py` 는
      다섯 자리다. 멎으면 판정 줄이 없다 — `subprocess.SubprocessError` 를 `GATE_FAULTS` 에 넣은 그 파일의
      계약과 어긋난다.
      **닫음(2026-09-26).** `1d1e5ca7` 실측: `check_analysis.py` 22 자리 중 1(`_safe_changed_go_paths`; `base_file` 은 7.5.34 가 지웠다) · `execution_baseline.py` 5 전부.
      값은 같거나 비슷한 부류의 명령을 부르는 이웃의 값(30 · 120 · 10 · 60 · 10 · 30 — 같은 명령은 `merge-base` 와 두 `diff --quiet` 셋뿐, 보수 때 정정). 구조 시험이 두 파일의 **모든** `subprocess.` 호출에, 행동 시험이 판정 · 기록 한 판이 띄운 자식 전부에 시한을 요구한다.
- [x] 7.5.16 **P2 — `record_landing` 이 더러운 트리를 재확인 **앞**에서 묻는다 (같은 재리뷰, 정확성).**
      `record_landing` 이 head → refusal → replay 순서다. 순서를 바꾸면 공짜로 닫힌다.
      **닫음(2026-09-26).** `_recording_moved` 가 말하는 순서를 역사 → 원장 → 거절로(거절은 여전히 먼저 계산 — 그 읽기가 원장에 든다).
      재현 `7516_order.py`: `5a54f78d` 더러운 트리 · `1d1e5ca7` 6.4(b) base 거절 · 편집 뒤 `base-commit.txt changed …`. 이 수리가 맨 앞 역사 물음을 못에서 뽑아(AA8 생존) 구조 시험으로 못 박았다.
- [ ] 7.5.17 **P2 — 저자가 판정을 **판정 줄 0개로** 무한히 늘릴 수 있다 (같은 재리뷰, 보안).**
      `resolve_test_file` 의 맨이름 갈래가 전체 트리 `rglob` 을 인용마다 돌고(메모 없음) 해소 실패는 오류 줄을
      안 낸다. 원장이 그 패턴을 재확인에서 다시 돈다. 실측: 엉터리 맨이름 20개 → 판정 16.55s + 재확인 15.78s,
      **판정 줄 0**. `branch-test-map.md` 300 KB ≈ 4.5시간이고 `gate.sh:321` 에 timeout 이 없다.
      (2026-09-26, 7.5.9 뒤: 맨이름 해소는 더 이상 트리를 돌지 않는다 — 색인이 한 번 받은 추적 목록에서 고르고 원장의
       `glob` 항목도 없어졌다. 이 공격의 비용을 **다시 재지는 않았다**. 해소 실패가 판정 줄을 안 내는 절반은 그대로 열려 있다.)
      **(2026-09-27 재측정 — 비용 절반은 닫혔다, 이 항목은 잔여로 좁힌다.)** `analysis/harness/7517_cost.py`(HEAD `26e5bb3f`, 실제 저장소 읽기만,
      병행 A/B 가 도는 중): 기준 `1d1e5ca7` N=20 색인 16.2s · 판정 263.2s · 재확인 109.7s(병행 부하로 옛 16.55s 보다 크다) → 워킹트리 N=20 판정 0.00s ·
      재확인 0.26s, N=1,000(52 KB) 0.08s · 0.29s, N=10,000(538 KB) 0.69s · 0.69s. 판정 줄은 모든 판에서 **0**. **남은 것**: (i) 해소 못 한 좌표가
      판정 줄을 안 낸다(오늘의 정책 — `test_a_qualified_path_that_does_not_exist_is_not_silently_skipped` 가 못 박는다; 7.5.37 의 비대칭과 같은 결정),
      (ii) `gate.sh:330` 의 `check_analysis.py` 호출에 시한이 없다 — 한 줄이지만 값을 정해야 하고(오늘 change 한 판의 최대는 a071 37.0s, 1d1e5ca7 A/B)
      `timeout(1)` 이 macOS 기본에 없어 게이트를 그 기계에서 깬다. 비용 쪽 동기가 사라져 급하지 않으므로 사람 결정으로 둔다.
- [x] 7.5.18 **P2 — `_bundle_text` 배관 시험이 이름만큼 못 박지 않는다 (같은 재리뷰, 시험품질).**
      `test_the_bundle_text_is_built_from_the_bytes_the_command_read` 의 픽스처가 판정 바이트와 디스크 바이트가
      **다른 순간을 안 만든다** — 호출부를 "그 자리에서 새로 읽기" 로 바꿔도 그 시험은 통과한다(스위트는 다른 둘로 잡는다).
      **닫음(2026-09-27).** 증거를 읽은 **뒤**, 번들 산문을 모으기 **전**(`_verdict` 의 `test_index` 호출 자리)에 디스크의 `ast.json` 을 다른 바이트로
      바꾸는 픽스처로 강화했다. 반증 변이 MV1(호출부가 `_read_regular` 로 다시 읽음)을 그 시험 **하나만** 돌려 쟀다(`7514_test_mut.py`): `3bb11b2d` 의
      시험 판은 MV1 아래 **초록**, 새 판은 **빨강**. 스위트 전체로는 MV1 CAUGHT 3.
- [ ] 7.5.19 **P3 — `AA6`(재확인이 HEAD 를 먼저 묻지 않는다)이 구조로만 못 박혔다 (같은 재리뷰, 시험품질).**
      순서는 **관찰 가능**하다 — 원장 경로가 바뀌고 이웃 커밋이 같이 서면 HEAD 문장 대 경로 문장으로 갈린다.
      기록 경로는 이미 행동으로 못 박혀 있고 판정 경로만 없다.
- [ ] 7.5.20 **P3 — `READ_CAP` 값에 시험이 없다 (같은 재리뷰, 시험품질).** `16 << 20` → `1 << 20` 으로 바꿔도 전부 초록.
      값은 산문 넷과 코드 한 곳에 있는데 묶는 것이 없다. 해법은 한 화면 옆에 있다
      (`test_the_minimum_git_version_is_written_down_once`). 열거표는 이제 `7524_census.py` 가 다시 찍는다.
- [ ] 7.5.21 **P3 — `main()` 의 git 호출이 결함 경계 **밖**이고, 잔챙이 셋 (같은 재리뷰, 보안 · 정확성).**
      `main` 이 부르는 `_commits_after` 가 `try/except GATE_FAULTS` 앞이라 멎으면 **이미 계산된 판정 줄 36개가
      안 찍히고** traceback 만 나간다 — 바로 위 주석이 그 경계가 있는 이유를 적어 놨다. 잔챙이: 원장 divergence
      메시지가 `_raise_if_inputs_moved` 에서 glob 키를 뭉갠다 · `moved` 판정 아래에서 `main` 이 사라진 상태의
      창 줄과 권유를 찍는다 · 같은 거절이 두 문장으로 나간다.
- [ ] 7.5.36 **P2 — 재확인 때만 추적 목록을 못 물으면 사유 이름이 틀린다 (7.5.9 로트 독립 적대 리뷰 P2-1, 2026-09-26).**
      판정 중 `git ls-files -z` 는 성공하고 끝의 재확인에서만 실패하면 원장 대조가 실패 지문과 성공 지문을 견줘
      "the tracked file list changed" 라고 말한다 — 움직인 것이 아니라 못 물은 것이다. 막는 쪽이라 판정은 맞다.
- [ ] 7.5.37 **P2 — 이름 인용은 거절, 좌표 인용은 건너뜀 — 그리고 작성 중인 미추적 시험 파일 (같은 리뷰 P2-3).**
      추적 안 된 파일의 시험 **이름**은 "in any tracked file" 로 거절되고 **좌표**는 해소 못 함(오류 아님)으로 건너뛴다 — 비대칭이다.
      그리고 TDD 흐름에서 저자가 아직 `git add` 안 한 시험 파일은 **머지에 들어간다** — "미추적 = 머지에 없음" 전제가 깨지는
      정상 입력이고 7.5.9 의 센서스(저장소 번들 · 시험 픽스처) 밖이었다. 거절 문장이 `git add` 를 말하지 않는다.
- [ ] 7.5.38 **P2 — `UnstatableEntry` 의 실패 지문에 이름이 없다 · 편집기 임시 파일이 결함 문장이 된다 (같은 리뷰 P2-4 · P2-5).**
      `_failed` 는 `종류:errno` 만 적어서 판정 중 **다른** 항목이 stat 불가가 돼도 재확인은 같다고 본다(판정은 이미 빨갛다).
      목록 뒤 stat 전에 사라진 편집기 임시 파일(vim 의 `4913`)이 `cannot tell what \`4913\` is` 결함 문장이 된다 — 재실행으로 풀리지만 사유가 오해를 부른다.
- [ ] 7.5.39 **P2 — AA8b 생존: "HEAD moved" 판정이 두 집에 산다 (같은 리뷰 P2-6).** `_head_moved` 를 `if False` 로 죽여도
      스위트 408 이 초록이다 — `_recording_moved` 의 B1 과 `_judged_state_moved` 가 같은 판정을 두 자리에서 내서 서로의 시험을
      통과시킨다([[two-judgements-cover-for-each-other]]). 합칠지 · 갈라 시험할지는 측정 뒤 결정. (7.5.16 이 B1 을 구조 시험으로
      못 박았지만 그것은 **순서**이지 판정의 집이 아니다.)
- [ ] 7.5.40 **P3 — `_recording_refusal` 의 `git diff --quiet` 에 `SNAPSHOT_PINS` 가 없다 (7.5.9 로트 주장정확성 리뷰 F9 부수, 이 로트 전부터).**
      `core.fsmonitor` 에 적힌 프로그램이 기록 명령 · 조언 줄에서 뜰 수 있다. 7.5.9 보수는 새 `ls-files` 만 고정했다.
- [ ] 7.5.41 **P3 — `_safe_changed_go_paths` 의 `\n` · `\r` · `\t` 거절을 받을지 (7.5.13 에서 연다, 2026-09-27).**
      7.5.13 뒤로 판정의 이름은 머리 줄이 아니라 레코드에서 오고 렌더러가 탭 · 개행도 git 과 같이 인용한다 — 이 세 글자의 거절은 이제 파서 때문이
      아니다. 남은 사유는 출력이다: 이름 댄 판정 줄(`missing evidence for modified function …`)에 줄바꿈이 들어가면 한 판정이 여러 줄로 찍힌다.
      받으려면 판정 줄의 이름 표기(예: `repr`)를 먼저 정해야 한다. 저장소 노출 0(1,816 이름).
      **(마감 수리 — 적대 리뷰 실측, 2026-09-27) 같은 사유가 7.5.13 이 새로 받게 된 글자에도 이미 있다**: VT(0x0b) · FF(0x0c) · ESC(0x1b) ·
      RS(0x1e) 가 든 이름의 편집이 이제 `(이름, 'F')` 요구로 서고, 그 이름이 `_verdict` 의 `f"{source}:{function}"` 를 거쳐 가공 없이 찍힌다 — 단말에서
      판정 줄이 여러 줄로 갈리거나 지워져 보인다. U+0085 · U+2028 · U+2029(git 이 인용 안 함)도 같다. 표기를 정하면 이 글자들까지 한 번에.
- [x] 7.5.42 **P2 — `_self_repair_commits` 에 `log.showRoot` 고정이 없다 (7.5.13 로트 적대 리뷰, 2026-09-27).**
      `log.showRoot=false` 면 루트 커밋의 `-z` 목록이 `\0<sha>\0` 이 되어 모양 검사를 통과하고 그 커밋의 깃발이 조용히 빈다. 노출 0 — 이 저장소의
      루트 커밋은 change 디렉터리를 안 만든다. **닫음(2026-09-27, 같은 로트 마감 수리).** 목록 명령에 `-c log.showRoot=true`(`diff.renames` 명령줄 고정과
      같은 원칙). RED(편집 전 코드) `[]` → GREEN `[<root>]`(`test_the_root_commit_is_read_whatever_log_show_root_says`) · 변이 MX8 CAUGHT.
- [ ] 7.5.43 **P3 — `log.showSignature=true` 와 서명 커밋이면 자기 수리 목록이 멈춘다 (같은 리뷰, 이 로트 전부터).**
      `_self_repair_commits` 의 `touching`(`git log --format=%H -- <경로>`) 출력에 `No signature` 같은 서명 줄이 섞여 다음 명령이
      `fatal: bad revision 'No'` 로 멈춘다 — 막는 쪽(결함 문장)이다. 고정(`-c log.showSignature=false`)은 같은 원칙이나 이 로트는 재지 않았다.
- [ ] 7.5.44 **P3 — 기록 쓰기의 권한 오류를 `record_landing` 이 안 잡는다 (7.5.14 로트 적대 리뷰, 2026-09-27).**
      `record_landing` 의 쓰기(`open("xb")`)는 `FileExistsError` 만 잡는다. change 디렉터리에 쓰기 권한이 없으면 `PermissionError` 가 함수 밖으로
      나간다 — CLI(`main`)는 `GATE_FAULTS` 경계에서 `no landing recorded — [Errno 13] …` 로 이름 대지만(실측 rc 1), 함수를 직접 부르는 호출자
      (시험 · 하네스)에게는 예외다. 판정 줄의 원칙([[a-fault-must-become-a-verdict]])으로는 함수 안에서 `(1, [...])` 가 맞다. 막는 쪽이고 노출은 사람의
      권한 실수뿐이라 이 로트는 주석만 사실로 고쳤다.
- [x] 7.6 **P2 — 규칙 한 집 (I2 · I3 · I4 · I5 · I6 · I7).** `compute_landing` 과 `resolve_landing`
      의 수락 조건 두 벌 · `validate` 의 지역 `canonical` 이 모듈 함수를 가림 · 디렉터리 해소 두 벌과
      없는 id 의 엉뚱한 조언 · 이관 거절 문장 두 벌(이미 갈렸다) · `_pre_archive_path` 가 아카이브
      문법의 한 벌 더 · `_change_analysis_path` 문장이 "current" 라고 말한다.
      **닫음(2026-09-13).** 기록은 review.md `## Pre-Edit Gate — task 7.6` · `## VERIFY — task 7.6`.
      **먼저 정정했다: 7.2.1 · 7.3.1 · 7.4 는 FLM 없이 함수 내부를 바꿨다**(침묵한 생략). 7.2.1 직전과
      HEAD 를 같은 열거기로 뽑아 여섯 함수의 공백을 기록으로 메웠다.
      (I2) 새 `_landing_refusal` 하나에 여섯 조건을 **`resolve_landing` 의 순서로** 둔다 — 순서가 거절
      지점이라, 비용 0 인 `compute` 순서로 맞추면 불일치 가드가 못에서 빠진다. 그 선택의 비용을 **먼저**
      쟀다(93건 전수: 번들 대조 +328회, +5%, 거의 전부 이미 느린 아카이브 일곱) — 구현 뒤 실측 +4.1%.
      (I4) 없는 id 를 없는 경로로 바꿔 넘기던 fallback 을 지웠다 — 거부할 정상 입력을 먼저 쟀다(게이트가
      받을 수 있는 id 126개 중 그 갈래 0). 타입을 따로 둔 이유가 사라져 `AmbiguousChange` 도 걷었다.
      (I3·I5·I6·I7) 이름 · 상수 · 해독 함수 하나 · 문장.
      "한 집"은 행동 시험으로 못 박히지 않으므로(일치하는 두 사본은 초록) 구조 시험 셋을 세웠다.
      뮤테이션 R1~R16 에서 **R16 이 SURVIVED** — 계산 경로가 규칙에 하한을 넘기는지를 아무도 안 쟀다
      (7.6 전부터 있던 구멍). 하한 순서 가드 **혼자** 막아야 하는 모양(두 가지에 같은 바이트의 증거,
      줄기는 증거를 먼저)을 닿는지 확인한 뒤 시험으로 굳혀 **열여섯 전부 CAUGHT**, 규칙의 네 조건은
      계산·선언 **두 경로 모두**에서 빨개진다(나머지 둘은 계산 경로가 순회 전에 돌아가서 구조상 안 닿는다).
      판정 A/B **IDENTICAL 28** · `compute_landing` 전수 **SAME 93**(값·사유) · 시험 125 → **132** ·
      `make sdd-test`(logic-map 207) · `lint` · `test-seams` · `openspec validate` 58/58.
- [x] 7.7 **P2 — 창 줄이 스스로 모순된다 (I8).** 기록이 디스크에 있는데 커밋 전이거나 아카이브
      이동이 staged 면 "working tree (no landed-commit.txt)" + `--record-landing` 을 권하고, 그 명령은
      "already exists" 로 거절한다. 빌리는 쪽 · 번들 0 인 쪽에도 같은 조언이 나간다.
      **닫음(2026-09-13).** 기록은 review.md `## Pre-Edit Gate — task 7.7` · `## VERIFY — task 7.7`.
      **먼저 다시 쟀다**: 픽스처로 재현하니 리뷰의 넷에 **셋이 더** 있었다(더러운 트리 · 커밋 전 번들 ·
      끊긴 심링크 기록 — 전부 명령이 걷기 전에 멈추는 자리). 실물 id 126 전수(HEAD 사본 `check` + 기록 없이
      돌린 `record_landing`): 조언이 명령을 권하는 99 중 **38 이 거절**이고, **활성은 권유 15 중 12**(전부 번들 0).
      기제는 판정 두 벌이다 — 명령은 거절 조건을 들고 있고 조언 줄은 아무것도 안 묻고 권했다
      ([[two-judgements-cover-for-each-other]]). 수리: 걷기 전 거절을 `_recording_refusal` **한 함수**에 두고
      기록 명령과 조언 줄이 둘 다 묻는다(구조 시험이 못 박는다). 계산의 순회 전 사유 둘은 `_walk_floor` 로 모아
      같은 함수를 부른다. 조언은 사유가 있으면 권하지 않고 **명령의 문장**을 인용한다. 대상 텍스트는
      `(no landed-commit.txt in HEAD)`, 디스크에만 있는 기록은 경로 + `not in HEAD` 로 말한다.
      **걷지 않는다** — 걸어서 아는 거절 15(전부 아카이브, 명령 한 번 최대 225.1초)는 예측하지 않고 권유 문장이
      기록을 약속하지 않는다. 판정 호출은 `check` 가 아니라 `main` 의 권유 갈래에 둬서 조언의 결함이 판정을
      못 바꾼다. **변이 22 중 첫 판 SURVIVED 둘**: R6(base 해소 실패 갈래 — HEAD 에도 시험 0, 바깥 경계가 rc 를
      대신 지켰다)과 R10(순서가 못에 없음 — 그리고 틀려 있었다: 영원히 못 기록하는 change 가 "먼저 커밋하라"를
      먼저 듣는다). dirty 하나를 맨 뒤로 옮기고 두 조합으로 못 박아 **22 전부 CAUGHT**. GREEN 첫 판은 7.4 시험이
      잡았다(판정 호출이 결함 경계 밖). 실물 A/B: 판정 **IDENTICAL 28**(대상 텍스트 치환만 정규화) · 조언
      **99/99** 가 명령의 결과와 일치 · 판정 비용 번들 0 에서 0.01~0.20초, 번들 있는 셋에서 ~1.2초.
      시험 132 → **141** · `make sdd-test`(logic-map 216) · `make lint` · `make test-seams` · a122 spec delta valid.
- [x] 7.8 **P2 — 시험 (I9~I16).** 생존 변이: 다중 후보 walk(M1·M2·M6) · 위조 픽스처를
      `record_landing` 에(M7) · `--diff-filter` 의 `M`(M3) · 아카이브 뒤 착지 보존(M4) · 모호/빌림/이관
      거절(M11·M8·M9) · CLI 종료 코드(M5). 음성 대조의 바늘 · 사적 픽스처 결합 · 픽스처 다섯 벌.
      **닫음(2026-09-13). 생산 코드 변경 0** — 시험만 바꿨다(6.3 과 같은 모양). 기록은
      review.md `## VERIFY — task 7.8`.
      **먼저 다시 쟀다**: 리뷰가 센 열 중 **셋(M1·M2·M5)은 이미 죽어 있었다** — 7.3.1·7.4 가
      다중 후보 픽스처와 CLI 시험을 만들었기 때문이다([[caller-count-is-not-fix-site-count]]).
      변이 정의는 §6 리뷰가 남긴 `run_M*/` 사본을 diff 해서 그대로 썼다 — 지어내지 않았다.
      남은 일곱(M3·M4·M6·M7·M8·M9·M11)에 시험을 써서 **열둘 전부 CAUGHT**(대조군 PC1·PC2 포함).
      가장 큰 것은 **M6**: `git rev-list base..HEAD` 는 "base 에서 안 보이는 커밋 전부"라 병합이
      있으면 base 를 한 번도 보지 못한 곁가지가 후보에 들어오는데, 저장소 픽스처가 전부
      선형이라 그 절이 한 번도 안 걸렸다. 병합 픽스처를 만들어 **닿는지 먼저 확인**하고
      ([[mutation-must-reach-the-thing-under-test]]) 종단으로 쟀다 — M6 아래서는 도구가 곁가지
      S 를 기록하고 게이트가 곧바로 `precedes the comparison base` 로 **자기 기록을 거절한다**.
      M4 는 옛 시험이 `check(...) == []` 만 봐서 기록이 안 읽혀도 초록이었다(그 픽스처는
      워킹트리를 대상으로 삼아도 통과한다) — `facts["landing"]` 을 아카이브 전후로 단언한다.
      I15(음성 대조 바늘)는 **무엇을 통과시켰는지 실측**했다: 픽스처를 다른 이유로 깨뜨린 네
      모양에서 옛 단언은 전부 통과하고 새 단언은 전부 실패한다. I16 은 사적 픽스처 교차
      호출 2→0(모듈로 올림), **사본 합치기는 의도적으로 안 했다** — 겹침 최대 68%, 동일한
      사본 0, 갈리는 부분이 각 클래스가 재는 바로 그것이다.
      시험 116 → **125** · `make sdd-test`(logic-map 200) · `make lint` · `make test-seams` ·
      `openspec validate --all --strict` 58/58 · 실물 a099 rc=0(required 32 불변).
- [x] 7.9 **P2 — 문서 (I17).** `tools/logic-map/README.md` · `docs/WORKFLOW.md` 에 `--record-landing`
      · `landed-commit.txt` 가 없다. 7.2 가 규칙을 정한 뒤에 쓴다(지금 쓰면 곧 틀린다).
      **닫음(2026-09-18). 코드 변경 0.** 기록은 review.md `## VERIFY — task 7.9`. 두 파일에서 두
      낱말의 실측 출현은 **0회**였다. 7.2·7.3·7.7 이 규칙을 정한 뒤인 지금 썼다.
      `docs/WORKFLOW.md` 에 `### 착지 지점 — landed-commit.txt`(명령 · **값은 저자가 고르지 않는다** ·
      도구가 받는 일곱 조건 · 덮어쓰지 않는 규칙과 **복구 경로** · 받지 않는 세 경우), README 에 같은
      내용을 도구 쪽 말로(기록은 **커밋해야** 효력이 있다 — 게이트는 HEAD 에서 읽는다).
      조건과 거절 문장은 산문이 아니라 `_landing_refusal` · `_walk_floor` · `ADOPTION_REFUSES_A_LANDING` ·
      `BORROWED_REFUSES_A_LANDING` · `LANDING_RECOVERY` 를 읽어서 적었다
      ([[contract-numbers-from-the-receipt]]).

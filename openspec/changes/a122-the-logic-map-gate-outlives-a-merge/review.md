# a122 · Review

## Function Logic Map: not-applicable (proposal 시점)

**사유.** 이 change 가 편집하려는 함수는 Python 이다 —
`tools/logic-map/check_analysis.py` 의 `changed_existing_functions` 와 `resolve_base`,
그리고 그 CLI. 이 저장소의 FLM 산출물은 `tools/logic-map/extract_go_ast.go` 가 **Go**
함수에서 뽑고, `docs/WORKFLOW.md` 는 "Python AST 분석기는 Go 코드에 재사용하지 않으며
`tools/logic-map/extract_go_ast.go` 가 그 역할을 맡는다"고 적는다. Go 함수를 하나도
고치지 않는 change 에 Go AST 산출물을 요구할 자리가 없다.

**이 면제는 스스로 좁다.** 검사는 이 표식을 `required` 가 비었을 때만 받아들인다
(`check_analysis.py:618`). 즉 이 change 가 Go 파일을 건드리는 순간 표식은 아무것도
통과시키지 못하고 실제 FLM 을 요구한다. 그때 tasks 1.2 의 단서대로 해당 함수의 FLM 을
먼저 만든다.

**손으로 읽은 증거의 한계를 적어 둔다.** proposal 이 인용한 `changed_existing_functions`
의 동작(대상이 워킹트리라는 것, `target` 매개변수가 CLI 에 노출되지 않는다는 것)은
`check_analysis.py:111-140` 과 `main()` 을 읽어서 얻었다. Python 에는 이 저장소가 쓰는
AST 열거 도구가 없으므로 이것이 얻을 수 있는 증거이며, **볼 곳을 골랐을 수 있다.**
proposal 의 수치 주장은 그렇지 않다 — 아래는 실행한 명령이다.

## 문제의 측정 (2026-09-08, HEAD 62b35779)

```
for id in a074-critical-events-reach-the-operator \
          a077-screens-show-what-they-already-know \
          a079-operator-can-lift-a-quarantine; do
  python3 tools/logic-map/check_analysis.py --change "$id"
done
```

| change | 결과 |
|---|---|
| a074-critical-events-reach-the-operator | FAIL — missing evidence 316건 |
| a077-screens-show-what-they-already-know | FAIL — 318건 |
| a079-operator-can-lift-a-quarantine | FAIL — 317건 |

셋 다 이 세션이 손대지 않은 change 다. 같은 날 a075 · a076 도 같은 이유로 실패해
5단계를 사유를 적고 면제한 뒤 아카이브했다(각 review 의 §완료 게이트 5단계 면제).

## 결정: 착지 지점 (2026-09-09, tasks 1.3)

**착지 지점을 무엇으로 볼지가 이 change 의 전부다.** 후보를 고르기 전에 각 후보가
**위조되는 방식**을 먼저 적었다. 그 값이 5단계를 통과시키는 유일한 손잡이가 되므로,
"있으면 통과"로 구현되면 이 change 는 게이트를 고치는 것이 아니라 여는 것이 된다.
tasks 2.4 가 그 변이를 심는 자리다.

산출물은 `analysis/landing-point.md` 다. 선언 파일 `landed-commit.txt` 를 골랐고,
파생 후보·merge 커밋·아카이브 커밋은 **측정으로** 기각했다.

## 측정이 전제 하나를 반증했다 (2026-09-09)

proposal 은 착지 지점을 주면 질문이 "이 change 가 고친 함수에 증거가 있는가"로
되돌아온다고 적었다. **틀렸다.**

```
git rev-list --count 448dfeb1..840b3377        # 1
git diff --name-only 448dfeb1 840b3377 -- '*.go' | wc -l   # 0
```

`448dfeb1..840b3377` 사이의 유일한 커밋이 "chore(sdd): rebaseline a074-a079 after
strategy merge" 이고 `base-commit.txt` 들만 고쳤다. 즉 2026-08-04 재기준화가 base 를
a074~a079 의 Go 작업 **뒤로** 옮겼다. 확인:

```
git grep -c "func (o \*ExitObserver)" 448dfeb1 -- internal/app/engine/
# 448dfeb1:internal/app/engine/exit_quarantine_announce.go:2
```

a074 가 증거를 낸 `ExitObserver.record` 의 파일이 base 에 이미 있다. 그러므로 이
다섯에는 "이 change 가 고친 함수"를 돌려주는 target 이 **없다** — 착지 지점을 주면
요구 집합은 비고, 그 뒤를 주면 남의 작업이다.

`sdd-workflow` 가 base 불변을 SHALL 로 요구하므로 base 쪽으로는 못 고친다. 이 change
가 주는 것은 올바른 질문이 아니라 **답할 수 있는 질문**이고, proposal 을 그렇게
정정했다. 그리고 그 빈 요구 집합은 **조용하면 안 된다** — task 3.4 와 spec 의
새 SHALL 이 그것을 말하게 한다.

## 독립 적대 리뷰 1회차 (2026-09-09) — **이 설계는 진행 불가**

리뷰어 셋을 각각 위조 가능성·수치 재측정·파급 범위로 붙였다. 두 리뷰가 돌아왔고
지적 둘이 결정적이라 **직접 재측정**했다. 셋 다 재현됐다.

### R1. 요구 집합을 좁혀도 이 다섯은 안 풀린다

`validate_target` 은 `revision: current` 번들의 `source_sha256` 을 **워킹트리** 파일과
대조한다(`check_analysis.py:516` 이 `value["file"]` 을 `root` 기준으로 풀고 `:523` 이
`hashlib.sha256(source.read_bytes())` 를 비교). 이 검사는 `required` 와 무관하게
**모든 번들**에 대해 돈다(`:635-637`). 재측정:

| change | `revision: current` 번들 | 워킹트리 대비 stale |
|---|---|---|
| a074 | 7 | **5** |
| a077 | 7 | **2** |
| a079 | 4 | **3** |
| a075 (아카이브) | 8 | **4** |
| a076 (아카이브) | 0 — analysis 디렉터리 없음 | — |

즉 `required` 를 ∅ 으로 만들어도 셋은 `AST source hash is stale` 로 떨어진다. 착지
지점이 실제로 푸는 것은 **a076 하나뿐**이다(번들이 없어 `:614-618` 의 면제 경로로 가고,
그 경로는 `required` 가 비어야만 성립한다).

### R2. 선언한 착지 지점은 위조된다 — 재현했다

`base-commit.txt` 는 proposal freeze 에 HEAD 로 쓰이고 **불변**이다. 그래서
"착지 커밋에 이 change 의 `base-commit.txt` 가 같은 내용으로 있다" 는 판정은
*파일이 생긴 뒤*를 고를 뿐이고, 그 구간의 바닥은 착지가 아니라 **freeze** 다.
판정 (4) 는 존재 검사다 — [[existence-check-is-not-a-role-check]] 그대로다.

a112 로 재현했다. base `aeeb209e`, 고른 착지 `5c5efec7`:

| 판정 | 결과 |
|---|---|
| (1) 실재 커밋 | 통과 |
| (2) base 가 엄격한 조상 | 통과 |
| (3) HEAD 의 조상 | 통과 |
| (4) 착지에서 `base-commit.txt` 가 같다 | 통과 — `aeeb209e…` |

| 구간 | Go 파일 |
|---|---|
| base → 진짜 착지(`62b35779`) | 135 |
| base → 고른 착지(`5c5efec7`) | **12** |
| 숨는 것(`5c5efec7..HEAD`, 비테스트) | **51** |

숨는 51개에 `internal/journal/schema.go` · `strategy_lane_latch_v32.go` ·
`internal/app/engine/strategy_market_coordinator.go` · `internal/positioncampaign/*`
가 들어간다 — `.claude/CLAUDE.md` §5 의 High-risk 경로다.

### R3. 증거가 착지 지점을 고정한다 — **1차 측정이 틀렸다**

R2 의 정석 수리는 "착지 지점을 번들이 고정하게 한다" 였다 — `revision: current` 번들의
`source_sha256` 이 **착지 커밋에서** 맞아야 한다면 저자는 착지를 자유롭게 못 고른다.
그리고 그것이 R1 도 같이 고친다(워킹트리 대신 착지에서 해싱).

**1차 측정은 성립하지 않는다고 답했고 그것은 방법론 오류였다.** 매칭 후보를 각 파일을
**만진** 커밋으로만 셌는데, 파일 내용은 다음 수정까지 유지되므로 만지지 않은 커밋도
후보다. 경계 커밋 전체를 후보로 다시 재니 답이 뒤집힌다.

| change | 번들 | 후보 커밋 | **전 번들이 맞는 커밋** | 그 커밋 |
|---|---|---|---|---|
| a074 | 7 | 48 | **1** | `448dfeb1` (= 자기 base) |
| a077 | 7 | 32 | **1** | `6b47be2f` (= **이전** base) |
| a079 | 4 | 51 | **1** | `448dfeb1` |
| a075 | 8 | 61 | **1** | `448dfeb1` |

수십 개 후보 중 정확히 하나다. 그리고 R2 의 위조 시나리오에 걸어 보면:

| a112 후보 | 132개 `revision: current` 번들 중 불일치 |
|---|---|
| `5c5efec7` (위조로 고른 착지) | **43** — 거절 |
| `62b35779` (진짜 착지) | **0** — 통과 |
| `1687baac` (HEAD, 더 엄한 쪽) | 0 — 통과 |

**판정 (4) 를 신원 검사에서 증거 검사로 바꾸면 위조가 막힌다.**
[[existence-check-is-not-a-role-check]] 의 처방 그대로다.

### 이 수리가 다섯에 주는 것

`validate_target` 이 `revision: current` 를 워킹트리 대신 **착지 커밋**에서 해싱하면
R1 의 stale 도 같이 풀린다. 착지가 base 와 같아도 안전하다 — 번들이 고정하므로 저자가
고를 수 없다.

구현하고 실제로 재니 아래와 같다(GREEN 이후, 착지 = `448dfeb1`).

| change | 착지 없음 | 착지 있음 | 요구 집합 | `448dfeb1` 에서 번들 불일치 |
|---|---|---|---|---|
| a074 | FAIL 324 | **PASS** | ∅ | 0/7 |
| a077 | FAIL 321 | **PASS** | ∅ | 0/7 |
| a079 | FAIL 322 | **PASS** | ∅ | 0/4 |
| a076 | FAIL 1 | **PASS** | ∅ | 번들 없음 |
| a075 | FAIL 324 | FAIL **1** | ∅ | 0/8 |

**a077 이 거절된다고 쓴 것은 틀렸다.** 그 주장의 근거였던 "번들이 `6b47be2f` 를
기술한다"는 후보 탐색의 산물인데, `git log -- <file>` 이 **병합 커밋을 이력 단순화로
빼기 때문에** `448dfeb1`(병합)이 후보에 없었다. 블롭을 직접 해싱하니 a077 의 7개
번들 전부가 `448dfeb1` 에서 맞는다. [[history-search-needs-intervals-not-touches]] 의
같은 함정을 한 단계 더 들어가서 다시 밟았다.

a075 만 남는데 사유가 다르다 — `cmd-tossctl--runconsole` 의 branch-test-map 이
`TestContainerBuildsDoNotStageALocalUpdate` 를 덮는 테스트로 인용하는데 그 이름이
트리에 **없다**. 이것은 a075 의 결함이고([[borrowed-flm-evidence-goes-stale]] 의
두 번째 방향), 착지 지점과 무관하다.

**의도적으로 안 한 것 하나 — 침묵한 생략이 아니다.** 인용된 테스트 이름은 여전히
**워킹트리**에서 찾는다(`test_index(root)`). 증거가 착지 리비전을 기술한다면 그 인용도
착지에서 찾는 것이 일관되지만, 이 change 의 spec SHALL 은 *source hash* 의 대조 대상만
말한다. 범위를 넓히지 않고 task 3.2.3 으로 남긴다.

## 정정한 수치 (리뷰어 3, 재확인함)

| 자리 | 썼던 값 | 맞는 값 |
|---|---|---|
| `448dfeb1..워킹트리` Go 파일 | 452 | **449** — 452 는 `rtk` 가 `git diff` 출력에 덧붙인 것이다. 셀 때는 `rtk proxy` ([[missing-tool-reports-clean]]) |
| `840b3377` 이 고친 것 | "`base-commit.txt` 들만" | **33 파일** — `ast.json` 23 · `base-commit.txt` 6 · `function-logic-map.md` 2 · `branch-test-map.md` 2. AST 재추출과 분기 재번호를 했다 |
| `ExitObserver.record` 의 파일 | `exit_quarantine_announce.go` | **`exitloop.go:1018`** — 번들 자신의 `ast.json` 이 말한다. 내 grep 은 패키지 전체를 세어 파일을 못 고른다 |
| a112 위조가 숨기는 비테스트 Go | 51 | **48** (같은 `rtk` 원인) |
| a075·a076 의 실패 | "같은 실패" | **다른 실패** — `missing base-commit.txt` 1건, missing-evidence 0건 |
| 지적 함수의 패키지 | `internal/verifylive/*` · `internal/scheduler/*` | `internal/scheduler` 는 **0건**. 최대 묶음은 `internal/app/engine` 61건이고 그건 a074 가 만지는 패키지다 |
| 재기준화 | `840b3377` 하나 | **둘** — review 가 적은 `0c563c6c → 6b47be2f`, 그리고 문서에 없는 `6b47be2f → 448dfeb1`(`840b3377`) |
| 하드코딩한 주체 | `resolve_base` | **`check`** (`check_analysis.py:590`). `resolve_base` 는 `change_dir` 을 인자로 받는다 |

## 결론 (1차 리뷰)

착지 지점 기제는 **판정 (4) 를 증거 검사로 바꾸면** 성립한다. 그대로 두면 성립하지
않는다 — R2 의 위조가 High-risk 경로 48개를 숨긴다. 남은 지적은 tasks 로 옮긴다.

## Pre-Edit Gate — task 1.11 (2026-09-10)

```text
Pre-Edit Gate:
- change id / task id: a122-the-logic-map-gate-outlives-a-merge / 1.11 → 2.7 · 3.4
- 대상 심볼(패키지.함수): tools/gate.sh 의 CHANGE_DIR(:116) · PAIR_DIR(:184)
                          (shell 이라 패키지·함수가 없다)
- CodeGraph definition/callers/callees/impact: 대상이 shell 이라 인덱스에 없다.
  호출자는 rg 전수 열거로 확정 — Makefile:169 하나. 근거는
  analysis/code-context/codegraph-baseline.md
- CodeGraphContext 후보와 evidence reconciliation:
  analysis/code-context/{codegraphcontext-context,evidence-reconciliation}.md
  — 같은 결함을 f6965ebb 와 task 3.2.2 가 이미 두 번 고쳤고 이번이 세 번째 자리다
- 기존 동작 파악 근거: gate.sh 3단계 전문(:152-239)을 읽었다. 유지해야 하는 검사 넷 —
  id 형태 검사, 자기 자신 선언 거부, 면제 줄 정확히 하나, 배포 단위 구성원 집합 일치.
  실측 실패는 make gate CHANGE=a099-… 의 3단계(2026-09-09)
- Function Logic Map / Branch Test Map: not-applicable — Go 파일이 한 줄도 바뀌지
  않는다. 수정된 기존 Go 함수 0개이므로 5단계가 요구하는 산출물도 0개다
- upstream 상속 테스트 영향: no — 개발 게이트 스크립트이고 tossctl 바이너리에 없다
- 실패 테스트 선행 작성: yes — tools/sdd/test_gate_resolves_archived_changes.py
- 설정·DB·journal 변경과 rollback: 없음. 스크립트 한 파일이라 revert 가 곧 rollback
- 안전 불변식 §0 위반 여부 검토: 통과. 완화 방향이지만 거부할 정상 입력을 먼저
  열거했다(task 1.11 의 (a)(b)(c)) — 오타 id 와 활성/아카이브 중복은 계속 거절한다
```

## Pre-Edit Gate — task 3.2.4 (2026-09-10)

- 대상: `tools/logic-map/check_analysis.py` 의 `resolve_referenced_change` 와
  `check` 의 해소 호출 자리 하나(`:689-692`). 그 외 파일 없음.
- HEAD: `1f2a2d6d7907484929eb122c6218b8878b1458cd`
- High-risk 여부: **아니다.** 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결
  어디에도 닿지 않는다. 유일한 production 호출자는 사람이 직접 부르는
  `tools/gate.sh:321` 이고 CI 는 이 경로를 돌지 않는다(`.github/workflows/ci.yml:94`).
- CodeGraph definition/callers/callees/impact: `resolve_referenced_change` 정의
  `check_analysis.py:231`, 호출자 `check`(`:684`) **1건**. 함수 단위 답이라
  한 함수 안의 두 호출 자리는 가르지 못한다.
- CodeGraphContext 후보와 evidence reconciliation:
  `analysis/code-context/evidence-reconciliation.md` 의 「task 3.2.4」 절.
  CodeGraph 와 AST 열거의 해상도 차이를 거기서 해소했고, 그 결과 편집 범위가
  하나에서 **둘**로 늘었다.
- Function Logic Map / Branch Test Map: **Go 기준으로는 not-applicable** —
  Go 파일이 한 줄도 바뀌지 않는다(이 문서 머리의 사유가 그대로 적용된다).
  다만 이 편집은 **분기와 early return 을 근거로 삼으므로** 손으로 읽지 않고
  기계로 열거했다: `analysis/python-function-logic/` 의 `ast.json` 둘과
  거기서 유도한 지도·분기 시험표. 생성기(`enumerate.py`)를 같이 둬서 재현 가능하게 했다.
  **저장소 도구가 아니라 이 change 의 일회용 열거기다** — a120 선례.
- 설정·DB·journal 변경과 rollback: **없다.** 스키마·설정·저널 어디도 안 바뀐다.
  롤백은 커밋 되돌리기 하나이고, 되돌리면 오늘의 동작으로 정확히 돌아간다.
- 거부하게 될 정상 입력: 아카이브된 change 와 **같은 id 로 활성 디렉터리를 다시
  만드는 것** 하나뿐. 2026-09-10 측정으로 저장소에 **0건**(활성 27 · 아카이브 id 99 ·
  교집합 0 · 아카이브 내 중복 0). 활성만·아카이브만인 입력은 전부 지금과 같이 통과한다.
- 토글: 없다. 이 편집에 토글이 없으므로 "토글 OFF = upstream 동작" 조항은 해당 없다.
- 실패 방향: 게이트가 **안 열리는** 쪽이다. 보수적이다.

## Pre-Edit Gate — task 3.2.3.1 (2026-09-10)

- 대상: `tools/logic-map/check_analysis.py` 의 `resolve_landing`(고정 판정)과
  `check` 의 호출 순서 하나(빌린 증거 해소를 `resolve_landing` **앞**으로).
  그 외 파일 없음.
- HEAD: `52c761defec307b4535aa0c5a72c910df3d7152f`
- High-risk 여부: **아니다.** 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결
  어디에도 닿지 않는다. 유일한 production 호출자는 사람이 부르는 `tools/gate.sh:321`.
- CodeGraph definition/callers/callees: `resolve_landing` 정의 `:345`, 호출자
  `check` **1건**, 피호출 `_committed_bytes` · `_is_ancestor` · `_ast_value`.
  함수 단위 답이므로 호출 **자리**는 AST 로 따로 셌다 — `check` 안에서
  `resolve_landing` 은 `L728` **한 자리**뿐이다(3.2.4 때와 달리 여기서는 하나다).
- Function Logic Map / Branch Test Map: **Go 기준 not-applicable** — Go 파일이 한 줄도
  바뀌지 않는다. 분기를 근거로 삼으므로 손으로 읽지 않고 기계로 열거했다:
  `analysis/python-function-logic/tools-logic-map--resolve_landing/`.
  편집 전 HEAD 열거 = 분기 19 · 반환 3 · raise 6 · `L345-408`. 구멍의 자리는
  기계가 가리켰다: `B11 L392`(고정 순회)가 **0회 돌면** `B19 L403`(`if mismatched:`)
  가 거짓이고 `L408 return candidate` 가 무조건 참이다.
- 설정·DB·journal 변경과 rollback: **없다.** 롤백은 커밋 되돌리기 하나.
- **거부하게 될 정상 입력** (이 항목이 이 태스크의 본문이다):

  | 모양 | 오늘 저장소에서 | 판정 |
  |---|---|---|
  | 착지를 선언한 change | **1건**(a099, 고정 번들 37) | 통과 — 안 바뀐다 |
  | 착지 선언 + 자기 번들 0 | **0건** | **거절** — 이 태스크가 닫는 구멍 |
  | 착지 선언 + 빌린 증거(`function-logic-reference.txt`) | 0건(a073 은 선언이 없다) | 통과 — 빌린 쪽 번들로 고정한다 |
  | 착지 선언 + 번들이 전부 `revision: base` | 0건(번들>0 인 change 는 전부 고정 번들>0) | 거절 — 아래 사유 |
  | 착지 선언 없음 | 125건 | 통과 — 오늘과 동일 |

  세 번째 행이 이 태스크에서 **실제로 위험했던 자리**다. a073 은 a072 의 번들을
  빌리고 자기 번들이 0 이라, 지역 번들만 세는 판정이면 **a073 의 유일한 수리 경로를
  막는다**. 픽스처가 아니라 실데이터로 쟀다(2026-09-10):

  | 착지 판정이 보는 증거 | a073 의 고정 번들 | 판정 |
  |---|---|---|
  | 옛 순서 = a073 지역 디렉터리 | **0개** | 거절 — 정상 입력이 죽는다 |
  | 새 순서 = 빌린 a072 디렉터리 | **209개** | 통과 |

  (오늘 a073 은 `AST source hash is stale` 로 빨갛고, 그 빨강을 푸는 것이
  착지 선언이다). 그래서 고정 판정은 그 change 가 **실제로 딛는** 증거로 한다 —
  지역이면 지역, 빌렸으면 빌린 쪽. 이 결정은 task 1.8(빌린 증거의 착지 **공유**
  규칙)을 앞당겨 정하지 않는다. "빌린 번들이 고정한다"는 "착지가 같아야 한다"보다
  약하므로 1.8 이 어느 쪽으로 정해도 모순되지 않는다.

  네 번째 행(전부 `revision: base`)은 삭제된 함수만 고친 change 다. `revision: base`
  번들의 해시는 **base** 를 기술하므로 착지에 대해 아무 말도 못 한다 — 고정할 수
  없다는 것이 사실이라 거절이 맞다. 저장소에 오늘 0건이고, 생기면 그때
  "무엇으로 고정할 것인가"를 정해야 한다(§5 잔여에 적는다).
- 토글: 없다.
- 실패 방향: 게이트가 **안 열리는** 쪽이다. 보수적이다.

## Pre-Edit Gate — task 3.3 (2026-09-10)

- 대상: `tools/logic-map/check_analysis.py` 의 `main`(출력)과 `check` 의 실패 메시지
  하나(`missing Function Logic Map for …`). 그 외 파일 없음.
- HEAD: `848b9ba3` (3.2.3.1 커밋)
- High-risk 여부: **아니다.** 판정을 한 줄도 바꾸지 않는다 — 바꾸는 것은 **설명**이다.
  이 태스크의 불변식이 그것이다: 126개 id 의 **rc 가 하나도 안 바뀌어야 한다**.
- 오늘의 출력을 먼저 쟀다 (2026-09-10, HEAD `848b9ba3`):

  | change | 출력 줄 | 가장 긴 줄 | 비교 창을 말하는 줄 |
  |---|---|---|---|
  | a074 | 324 (그중 `missing evidence for modified function` **316**) | 183자 | **0** |
  | a076 | **1** | **21,838자** (이름 316개를 쉼표로 이은 한 줄) | **0** |

  둘 다 "이 함수들이 **어느 두 지점 사이**에서 바뀐 것인가"를 한 번도 말하지 않는다.
  a074 의 316줄은 전부 남의 change 함수인데 출력만 봐서는 알 길이 없다.
- 기계 열거가 기제를 가리켰다(`analysis/python-function-logic/tools-logic-map--main/`).
  `main` HEAD = L817-840 · 분기 4 · 반환 2. **`B1 L824 if errors:` 가 `L827 return 1`
  로 빠져나가 `B3`·`B4` 를 통째로 건너뛴다.** 착지·요구 수를 찍는 `B4 L835` 가 실패
  경로에서 **도달 불가**다. 그래서 task 1.9 가 적은 "항상 출력한다"는 성공할 때만
  참이었다 — 3.2.4 와 **같은 모양의 early return** 이다.
- CodeGraph / 소비자: 이 문구를 읽는 도구·CI 는 **없다**(저장소 전수 grep). 인용은
  전부 산문 기록이고 그것들이 인용하는 것은 **성공** 줄
  (`evidence complete or diff-proven exempt`)이다 — 그 줄은 손대지 않는다.
- Function Logic Map: **Go 기준 not-applicable** (Go 파일 0줄). 분기를 근거로 삼으므로
  기계 열거를 먼저 만들었다.
- 거부하게 될 정상 입력: **없다.** 이 편집은 어떤 입력도 새로 거부하지 않는다 —
  판정 경로에 조건을 더하지 않고 출력만 더한다. 그 주장은 126개 id 의 rc 대조로 잰다.
- 토글: 없다. 실패 방향: 판정 불변, 설명만 늘어난다.

## Pre-Edit Gate — task 1.8 (2026-09-10)

- 대상: `tools/logic-map/check_analysis.py` 의 `check`(빌린 증거 블록) 와
  `resolve_landing`(선언 읽기를 헬퍼로 분리). 그 외 파일 없음. Go 0줄.
- HEAD: `73cc6bdc` (3.3 커밋)
- High-risk 여부: **아니다.** 주문·손절·사이징 경로가 아니다. 다만 게이트를 **더
  엄하게** 만드는 편집이므로 아래 "거부할 정상 입력"을 먼저 열거하고 실물로 쟀다.

### 무엇이 비어 있는가

빌린 증거에는 창의 **시작**을 맞추는 규칙이 있다 — `referenced_base != base` 면
거절한다(기계 열거 `B16 L772` → `L773 return`). 창의 **끝**에는 그 규칙이 없다.
3.2.3.1 이 넣은 고정은 "빌린 번들이 착지를 고정한다"까지이고, 그것은 "착지가 같아야
한다"보다 약하다 — 그 태스크가 1.8 을 앞당겨 정하지 않겠다고 적은 자리다.

### 그 자유가 실제로 얼마나 큰지 쟀다 (2026-09-10)

a072 의 `revision: current` 번들은 **99개 파일 · 99개 digest** 다. base
`171739a4`..HEAD 의 **326 커밋** 중 그 99개를 **동시에** 고정하는 커밋은 **2개**다:
`171adda8` · `fb135d85`. 즉 빌리는 쪽이 고를 수 있는 값은 오늘 2개이고, 둘의
요구 집합은 **둘 다 147개로 같다**. 그래서 이 저장소에서 그 자유의 측정된 크기는
**0** 이다 — 원리로는 열려 있고 실물로는 안 갈린다.

같은 측정이 a073 의 수리 경로도 확정했다:

| a073 의 대상 | required | a072 번들이 안 덮는 것 |
|---|---|---|
| 워킹트리(오늘) | 388 | **240** |
| 착지 `171adda8` | 147 | **0** |
| 착지 `fb135d85` | 147 | **0** |

(오늘 a073 의 실제 실패는 `missing evidence for modified function` **240** ·
`AST source hash is stale` 61 · `AST hash does not match` 35 = **336개**다. 앞선
기록이 stale 쪽만 적었는데, 지배적인 것은 240 쪽이다.)

수리가 실제로 되는지는 **끝까지 돌려서** 쟀다. a072·a073 이 둘 다 `171adda8` 을
선언한 상태를 만들어(선언을 **읽는 자리** `_declared_landing` 에 값을 주입하고 그
뒤의 판정 — 40-hex · 커밋 실재 · 조상 · 고정 · 번들 · 덮임 — 은 전부 진짜로 돌렸다)
`check` 를 부르면 **오류 0개**다. 336 → 0. 덮임 숫자만 보고 "수리된다"고 적지 않았다.

### 결정

**창은 양쪽 끝을 다 공유해야 한다.** 빌리는 change 의 `landed-commit.txt` 는
빌려주는 change 의 것과 **정확히 같아야** 하고, 빌려주는 쪽에 선언이 없으면 빌리는
쪽도 선언할 수 없다.

이유는 크기가 아니라 **누가 고르는가**다. `analysis/landing-point.md` 가 선언 파일을
고른 근거는 "번들이 고정하므로 저자가 고를 수 없다" 하나다. 빌리는 change 에서는
그 번들이 **남의 것**이라, 고정 구간 안에 남은 선택이 저자에게 돌아온다. 착지는
그것을 고정하는 증거가 사는 자리에 선언하고, 빌리는 쪽은 그 값을 **복사**한다.

### 거부할 정상 입력 — 먼저 열거하고 쟀다

| 정상 입력 | 새 규칙 | 저장소 실물 |
|---|---|---|
| 양쪽 다 선언 없음 (a073·a072 의 **오늘 모양**) | 통과 — 판정 불변 | 1건, 영향 0 |
| 양쪽이 같은 값을 선언 | 통과 | 0건 (규칙이 여는 모양) |
| 빌리는 쪽만 선언 | **거절** | 0건 — 오늘 선언을 가진 change 는 a099 하나이고 빌리지 않는다 |
| 빌려주는 쪽만 선언 | **거절** | 0건 |
| 서로 다른 값을 선언 | **거절** | 0건 |

그러므로 이 규칙이 오늘 새로 거절하는 change 는 **0건**이다. 대신 a073 의 수리에
**순서**가 생긴다 — a072 에 `171adda8` 을 먼저 적어야 a073 이 같은 값을 적을 수
있다. 죽는 경로가 아니라 늘어나는 단계 하나이고, 그 단계가 값을 고정하는 증거가
사는 자리다.

### 이 규칙이 닫지 **않는** 것 (잔여)

빌리는 change 의 작업이 공유된 착지 **뒤에** 착지하면, 그 작업은 요구 집합 밖에
있고 어떤 판정도 그것을 못 본다. 빌리는 change 는 정의상 자기 번들이 0 이라 그것을
고정할 증거를 **소유하지 않는다**. 규칙 A(공유)든 오늘의 규칙 B(빌린 고정)든 이
잔여는 같다 — 새 규칙이 만드는 것이 아니다. §5 잔여에 적는다.

### 판정 둘이 서로를 가릴 자리를 먼저 찾았다

새 공유 판정은 `resolve_landing` **앞에** 선다. 그러면 기존 3.2.3.1 회귀 시험
(`test_borrowed_evidence_refuses_a_landing_it_does_not_describe`: 빌리는 쪽만 바닥을
선언)이 **공유 규칙**에 먼저 걸려서, 고정 판정을 통째로 지워도 초록으로 남는다.
그래서 그 시험의 픽스처를 "양쪽이 같은 바닥을 선언"으로 고쳐 고정 판정이 홀로
판정하게 둔다. 변이 M3 으로 잰다.

- Function Logic Map: **Go 기준 not-applicable** (Go 파일 0줄). 분기를 근거로 삼으므로
  `analysis/python-function-logic/` 기계 열거를 편집 **전에** 먼저 만들었다
  (`tools-logic-map--check/ast.before-1.8.json`: L737-833 · 분기 37 · 반환 11).
- CodeGraph: `resolve_landing` 의 호출자는 `check` **1건**. 호출 **자리**는 AST 로
  셌다 — `resolve_referenced_change` 가 `check` 안에 `L743`·`L768` **둘**이고 이
  편집이 닿는 것은 `L768`(빌린 쪽) 하나다.
- 토글: 없다.
- 실패 방향: 게이트가 **안 열리는** 쪽이다. 새 판정은 오류를 더할 뿐 지우지 않는다.

## Pre-Edit Gate — task 1.12 (2026-09-11)

- 대상: `tools/logic-map/check_analysis.py` 의 `resolve_base`(감사된 source 를 기록) ·
  `check`(이관 경로의 대상 결정) · `_target_text`(무엇이 끝을 고정했는지 말한다).
  `execution_baseline.py` 는 **안 바꾼다**. Go 0줄.
- HEAD: `52d5cb2c` (1.8 커밋)
- High-risk 여부: **아니다.** 다만 대상이 a063 **이관 예외** 경로다 — 사람이 승인한
  단 하나의 예외이므로 거부할 정상 입력을 먼저 열거하고 실물로 쟀다.

### 태스크의 물음과 답

물음은 "착지 기록을 record 의 키 집합 열거(`execution_baseline.py:410`)에 넣을지"다.
**답은 아니다 — 창의 끝은 이미 그 record 안에 있다.** 이름이 `source_commit` 이고,
a063 의 실제 record 는 오늘 14개 키에 `source_commit = c727ad12` 를 담고 있다.
같은 값을 두 번째 이름으로 또 적으면 재진술이고, 둘이 갈리면 어느 쪽이 계약인지
아무도 모른다 — [[two-judgements-cover-for-each-other]]. 게다가 키 집합은 닫혀 있어
(`set(record) != required` 면 거절) 키를 하나 넣으면 a063 의 커밋된 record 와 ledger
digest 를 다시 만들어야 한다.

### 진짜 구멍 둘 — 재서 확인했다

**(1) 이관 경로의 대상은 감사된 값이 아니라 워킹트리다.** `validate` 는
`{"effective_base": E, "source": source, "ledger": …}` 를 돌려주는데 `resolve_base` 는
`effective_base` 만 읽는다(저장소 전수: `adoption["source"]` 를 읽는 곳 **0곳**).
그래서 `resolve_landing` 이 `landed-commit.txt` 를 찾고, 없으니 대상이 워킹트리가 된다.

그 답이 오늘 맞는 이유는 `validate` 의 **다른** 판정 때문이다 — source..head 에
`openspec/`·`docs/pm/` 밖 파일이 있으면 `source-to-evidence drift` 로 죽는다. 즉
맞는 답을 **다른 판정의 우연**으로 얻고 있다([[uncovered-block-may-be-held-dead-by-a-coincidence]]).
그 우연이 얼마짜리인지 쟀다:

| a063 의 대상 | required |
|---|---|
| 감사된 `source_commit` c727ad12 | **9** |
| 워킹트리(2026-09-11 HEAD) | **36** |

(`c727ad12..HEAD` 는 커밋 32개 · 파일 579개, 그중 **.go 24개**. 오늘 이관을 다시
돌리면 drift 검사가 먼저 죽이므로 fail-closed 지만, 창의 계산은 남의 함수 27개를
a063 의 것으로 센다.)

**(2) 3.3 이 넣은 안내가 감사와 모순된다.** 이관 픽스처로 5단계를 실제로 돌려서 쟀다
(2026-09-11):

```
[logic-map] a063-…: base 145812c4c6e0 → working tree (no landed-commit.txt) required 1 function(s)
[logic-map] a063-…: … — record `landed-commit.txt` to narrow it to this change's own work
[logic-map] a063-…: execution-baseline adoption exception evidence complete
```

두 번째 줄이 저자에게 **감사되지 않는 두 번째 손잡이를 만들라고 시킨다.**
`landed-commit.txt` 는 `openspec/` 아래라 drift 검사가 통과시키고, 추적 파일이라
untracked 감사도 못 보고, 닫힌 키 집합에도 없다. 이관 예외의 정당성이
"판정에 들어가는 입력을 하나도 빠짐없이 열거하고 digest 로 묶었다"인데 그 열거에
없는 입력이 대상을 고른다.

### 결정

1. **키를 넣지 않는다.** 창의 끝은 `source_commit` 이다.
2. **이관 경로는 착지 기록을 거절한다.** 손잡이는 하나이고 그것은 감사된 것이다.
3. **감사된 값을 실제로 쓴다.** a063 의 대상은 `source_commit` 이고 출력이 그렇게
   말한다. 그러면 (2)가 거절하는 파일을 (2)가 시키지 않는다.

### 거부할 정상 입력 — 먼저 열거하고 쟀다

| 정상 입력 | 새 규칙 | 저장소 실물 |
|---|---|---|
| a063 이 착지를 선언하지 않는다 (**오늘 모양**) | 통과. 대상이 워킹트리 → **감사된 source** 로 바뀐다 | 1건. 판정은 이관 실행 시점에 같다(drift 가 보장) — 출력만 달라진다 |
| a063 이 착지를 선언한다 | **거절** | **0건** — a063 에 `landed-commit.txt` 가 없다 |
| 이관이 아닌 change 가 착지를 선언한다 | 영향 없음 | a099 1건, 이관 경로가 아니다 |
| 이관이 아닌 change 가 선언하지 않는다 | 영향 없음 | 나머지 전부 |

그러므로 이 규칙이 오늘 새로 거절하는 change 는 **0건**이고, 판정을 바꾸는 change 도
**0건**이다. 바뀌는 것은 a063 이관 실행의 **출력 문구**다.

### 이 편집이 지우는 우회로

이관 경로는 `resolve_landing` 을 부르지 않게 된다. 그 함수의 판정(40-hex · 커밋 실재 ·
HEAD 의 조상 · base 의 자손 · 증거 고정)은 **저자가 고른 값**을 위한 것이고, 감사된
source 는 고른 값이 아니다 — `validate` 가 이미 `ancestry(P,E)` · `ancestry(E,source)` ·
`ancestry(source,head,strict=True)` · `source^{tree}` 대조 · digest 셋으로 묶는다.
같은 판정을 두 번 하지 않는다.

- Function Logic Map: **Go 기준 not-applicable** (Go 0줄). 분기를 근거로 삼으므로
  기계 열거를 편집 **전에** 만들었다 — `tools-logic-map--resolve_base/ast.before-1.12.json`
  (L282-322 · 분기 10 · 반환 2 · raise 4). `check` 의 편집 전 상태는
  `tools-logic-map--check/ast.after-1.8.json` 이고 그 `source_sha256`(`a5ee9c6febc7`)이
  커밋된 `check_analysis.py` 의 해시와 같음을 대조했다.
- CodeGraph: `resolve_base` 의 호출자는 `check` 1건 + `execution_baseline` 의 지연
  import 경로. 호출 **자리**는 AST 로 셌다 — `check` 안에 `L768`·`L782` 둘이고
  이 편집이 닿는 것은 `L768`(게이트 대상) 하나다.
- 토글: 없다.
- 실패 방향: 게이트가 **안 열리는** 쪽이다. 창이 넓어지는 일은 없다 — 좁아진다.

## VERIFY — task 4.1: 아카이브된 a075·a076 회귀 픽스처 (2026-09-11)

- HEAD: `c91dc484` (1.12 커밋)
- 4.1 이 적어 둔 **선결 조건은 이미 닫혔다.** "아카이브된 change 는 id 로 재검사가
  안 된다"는 2026-09-08 실측이었고, task 3.2.4(`52c761de`)가 `check` 의 change 해소를
  `resolve_referenced_change` 로 바꾸면서 풀렸다. 오늘 두 change 모두 아카이브 base
  `448dfeb1` 을 해소한다.
- 측정은 주입이 아니라 **격리 worktree 에 실제로 커밋**해서 했다
  (`git worktree add --detach`, 커밋 `06b280e0`, 착지 = `840b3377`). 그래서 선언을
  읽는 자리(HEAD 커밋에서 읽기·40-hex)까지 전부 진짜 경로다.

### a075 — 해결된다

| | 착지 없음 | 착지 `840b3377` |
|---|---|---|
| required | **319** | **0** |
| 오류 | **324** (`missing evidence` 318 · `stale` 4 · `hash does not match` 1 · 인용 테스트 1) | **1** |

323개가 사라진다. 남은 **1**은
`cmd-tossctl--runconsole: branch test map cites TestContainerBuildsDoNotStageALocalUpdate`
이고, task 1.13 이 "**a075 의 결함이고 a122 가 덮지 않는다**"고 이미 적어 둔 것이다.
즉 이 change 는 자기가 고치겠다고 한 것만 정확히 고친다.

### a076 — **해결되지 않는다**

a076 은 `analysis/function-logic` 번들이 **0개**다(다섯 중 유일하다 — a074 7 · a075 8 ·
a077 7 · a079 4). 착지를 선언하면 task 3.2.3.1 의 규칙이 거절한다:

```
[logic-map] cannot derive modified Go functions: landing point 840b337725fa is pinned by
no `revision: current` evidence: a declared landing must be the revision some bundle describes
```

**그 거절은 옳다.** 3.2.3.1 이 막는 위조 경로가 정확히 이 모양이다 — 번들 0 + 면제
표식 + 구간 바닥 착지. 저장소는 "증거가 없어서 0"과 "고친 Go 가 없어서 0"을 가르지
못한다. a076 은 실제로 자기 창에서 기존 Go 함수를 하나도 안 고쳤지만(재기준화가
base 를 그 작업 뒤로 옮겼다), 그것을 증명할 증거를 **소유하지 않는다**.

그래서 proposal 이 "이 다섯에 답할 수 있는 질문을 준다"고 적은 것은 **다섯 중 넷**에
대해 참이다. a076 은 §5 잔여로 연다(5.6).

### 곁가지로 3.3 의 결함 하나가 실물로 나왔다 — 못 잰 창을 찍는다

a076 에 착지를 준 실행이 이렇게 찍었다:

```
[logic-map] a076-…: base 448dfeb1263d → working tree (no landed-commit.txt) required 0 function(s)
[logic-map] a076-…: … record `landed-commit.txt` to narrow it …
[logic-map] cannot derive modified Go functions: landing point … is pinned by no …
```

첫 줄은 **거짓**이다. 대상은 워킹트리가 아니고(착지가 선언돼 있다), 요구 수는 세지도
않았다(해소가 실패해 `landing`·`required_count` 가 아예 안 채워졌고 기본값이 찍혔다).
둘째 줄은 이미 선언된 파일을 만들라고 한다. 3.3 이 창을 찍게 만든 이유가 "이름만 있고
이유가 없다"였는데, **지어낸 이유는 그보다 나쁘다**.

고쳤다 — 창은 **잰 것만** 찍는다(`if base and "landing" in context:`). 사유는 아래
`cannot derive …` 줄이 말한다. 오늘 저장소에서 이 줄을 보는 change 는 **0건**이다
(착지를 선언한 change 는 a099 하나이고 유효하다). 시험
`test_a_window_that_could_not_be_derived_is_not_printed` 이 이것을 재고, 변이 O1 이
그 시험 하나를 빨갛게 한다.

- 측정 worktree 는 정리했다. 아카이브된 a075·a076 에 `landed-commit.txt` 를 **쓰지
  않았다** — 4.1 은 검증 태스크이고, 아카이브본 수리는 별개의 판단이다(§5).

## VERIFY — task 4.2

**대상** — a074 · a077 · a079 를 실행해 요구 집합이 줄어드는지 확인한다.
**HEAD** `a7d9045d`. **생산 코드 변경 0** — 이 task 는 순수 측정이고 `check_analysis.py`
는 손대지 않았다. 따라서 High-risk 경로 없음, 토글 없음.

**방법** — 4.1 과 같다. `git worktree add --detach` 로 격리 worktree 를 세우고 세
change 에 `landed-commit.txt` 를 **실제로 커밋**한다(`886ab949`). 주입이 아니라 선언을
읽는 자리(HEAD 커밋에서 읽기 · 40-hex · 조상 판정)까지 전부 진짜 경로다.

### 결과 — 셋 다 창을 되찾는다

| | 착지 없음 | 착지 `840b3377` |
|---|---|---|
| a074 | required 319 · rc=1 · 오류 **325** | required **0** · rc=0 · 오류 **0** |
| a077 | required 319 · rc=1 · 오류 **322** | required **0** · rc=0 · 오류 **0** |
| a079 | required 319 · rc=1 · 오류 **323** | required **0** · rc=0 · 오류 **0** |

셋 다 base `448dfeb1` 을 공유하고 착지가 한 점으로 모인다.

### 사라지는 것은 두 종류이고, 둘째가 새로 확인된 것이다

| 오류 | a074 | a077 | a079 |
|---|---|---|---|
| `missing evidence for modified function` (**남의 함수**) | 316 | 318 | 317 |
| `AST source hash is stale` (**자기 증거**) | 5 | 2 | 3 |
| `AST hash does not match modified function revision` | 3 | 1 | 2 |
| 3.3 의 조언 줄 | 1 | 1 | 1 |

stale 이 이름으로 부르는 파일은 정확히 그 change 자기 번들의 소스다 — a074
`exitloop.go`·`engine.go`, a077 `portfolio_pages.go`, a079
`position_policy_transport.go`·`position_policy.go`·`console.go`. 대상이 워킹트리이면
증거 대조도 워킹트리에서 하므로 **남이 그 파일을 고치는 순간 자기 증거가 썩는다**.
1.9 가 `validate_target(revision_ref=…)` 로 닫은 절반이 여기서 실물로 확인된다.

### 이 task 의 전제는 틀렸다 — 요구 집합은 0 으로 간다

"각 change 가 실제로 고친 것으로 줄어든다"가 아니다. **0** 이다. 산술이다.

- base 와 착지 사이 커밋은 **1개**이고 그 커밋(`chore(sdd): rebaseline a074-a079
  after strategy merge`)이 Go 파일을 **0개** 바꾼다.
- 세 change 의 `revision: current` 번들 소스 **12/12 파일이 base 자체에서 이미 hash
  일치**한다 → 셋의 Go 작업은 자기 base **앞**에 착지했다.

spec 의 "착지 지점이 base 뒤에 있어 요구 집합이 비는 change" 시나리오가 정확히 이
모양이고, 5단계는 그 0 을 출력하면서 번들 자체는 계속 검사한다. 이것을 잔여 5.7 에
연다 — 활성 change 중 번들을 가진 13건 가운데 **8건**이 같은 상태다.

### 거부할 정상 입력

이 task 는 코드를 안 바꾸므로 새로 거부되는 입력이 **없다**. 다만 아래 셋이 착지를
선언하면 무엇이 거부되는지는 재 두었다(P2·P3).

### 통과는 증거가 아니다 — 변이

rc=0 이 "검사를 안 했다"와 구분되는지는 이것으로만 갈린다.

| 변이 | 결과 |
|---|---|
| P1 `exitobserver.run` 의 `source_sha256` → 0 (a074) | 거절 · `is not the revision this evidence describes: internal/app/engine/exitloop.go` |
| P1' 같은 변이 (a077 `joinPositions`) | 거절 · `internal/console/portfolio.go` |
| P1'' 같은 변이 (a079 `Console.routes`) | 거절 · `internal/console/console.go` |
| P2 착지를 고정 구간 **밖**으로 | 거절 · 두 파일을 이름으로 부른다 |
| P3 번들 전체 삭제, 착지 유지 | 거절 · `pinned by no revision: current evidence` |

### 고정이 남기는 구간이 이번엔 결과를 바꾼다

1.8 은 "고정은 값이 아니라 **구간**을 남긴다"고 적었고 당시 실측(a072 2/326, 둘 다
required 147)에서 효과는 0 이었다. a074 는 구간이 **14/287** 이고 그 안에서 required
가 **0 → 30**, rc 가 **0 → 1** 로 갈린다.

| a074 착지 후보 | required |
|---|---|
| `840b3377` · `15d25f80` | **0** (rc=0) |
| `df4407ed` · `359b1fe7` · `30d8bb93` | 4 |
| `aaa7638d` · `8291c5e3` · `96c621d3` · `f1aae509` | 9 |
| `56e85c68` | 14 |
| `c58b66c9` · `3dd077ae` · `53626032` | 17 |
| `8dba0173` | **30** |

이번 경우 통과하는 선택(구간 **바닥**)이 **옳은 선택과 같다** — 위로 갈수록 늘어나는
것이 a080·a081·a082·a083 의 함수, 즉 a122 가 없애려는 남의 작업이기 때문이다. 구간
폭은 5.3 에 옮겨 적었다(a072 2 · a075 2 · a077 2 · a079 2 · a074 14 · a091 26).

### 선언은 하지 않았다

셋 다 배포 후 실측 task 가 열려 있다(5.1). 지금 착지를 선언하면 창이 그 자리에서
얼고, 그 실측이 Go 수정을 부르면 그 수정이 창 **밖**으로 나가 어떤 판정도 못 본다 —
5.5 가 빌린 증거에 대해 적은 모양이 빌리지 않은 change 에서도 생긴다. 선언은 셋이
각자 Go 작업을 끝냈다고 판단할 때 그 change 가 한다.

- 측정 worktree 는 제거했고(`git worktree remove --force` → `prune`) 세 change
  디렉터리에는 **아무것도 쓰지 않았다**. 확인: `git status` 에 tracked 변경 0.

## VERIFY — task 4.3

**대상** — 스위트 전체와 정적 게이트. **HEAD** `1bd132db`. **생산 코드 변경 0**
(4.2 와 같이 순수 검증 task). Go 0줄, High-risk 경로 없음, 토글 없음.

| 게이트 | 결과 |
|---|---|
| focused `test_check_analysis.py` | 75 OK · a122 가 31 추가(44→75) |
| `tools/logic-map` python 전체 | 150 OK |
| `make test` | rc=0 · 99 패키지 · FAIL 0 |
| `go test -count=1 ./...` | rc=0 · 99 패키지 · **cached 0** · FAIL 0 |
| `make vet` / `make lint` | rc=0 / rc=0 (태그 vet 포함) |
| `make test-seams` | rc=0 · 100 패키지 · FAIL 0 |
| `make test-race` | rc=0 · 8 패키지 · DATA RACE 0 |
| `make validate` | rc=0 · 58/58 |
| `openspec validate --strict` 활성 전수 | 27/27 |
| `make sdd-sync` / `make sdd-check` | rc=0 / rc=0 |

### 캐시된 초록은 "지금 봤다"가 아니다

`make test` 가 다수 `(cached)` 로 찍혀 `-count=1` 로 다시 돌렸고 cached 줄 **0개**를
확인했다. Go 의 캐시는 입력 키로 검증하므로 틀린 값은 아니지만, VERIFY 의 문장은
"검증된 값"이 아니라 "지금 돈 값"이어야 한다.

`make test-race` 는 4.3 목록 밖이다. a122 는 Go 0줄이라 면제 사유가 성립하지만
(**not-applicable: Go 변경 0**) 재는 쪽이 싸서 쟀다.

### 실데이터 전수 126건 — 그리고 a122 는 저장소를 초록으로 만들지 않는다

| | 건수 |
|---|---|
| rc=2 (도구 붕괴) | **0** |
| rc=0 | **2** — a099(유효한 착지 선언) · a122(Go 0줄) |
| rc=1 | 124 |

한 줄로 끝나는 10건의 사유를 전부 확인했다 — 9건은 `base-commit.txt` 부재(관례 이전
아카이브 7 + a119 + verify-execution-capability), 1건은 a063 의
`adoption requires detached HEAD` 이고 **옛 도구도 같은 줄을 찍는다**.

124건의 rc=1 을 a122 의 실패로 읽으면 안 된다. 5단계 강제가 꺼져 있는 동안 쌓인 기존
부채이고, a122 가 주는 것은 **손잡이**(`landed-commit.txt`)다. 오늘 그것을 쓴 change 는
a099 하나다.

### 아카이브 재검사 A/B

| id | 옛 도구(`a2d11fb2`) | 지금 도구 |
|---|---|---|
| a075 | `missing base-commit.txt` | `base 448dfeb1 → … required 319` |
| a120 | `missing base-commit.txt` | `base e65e394b → … required 36` |
| a073 | `missing base-commit.txt` | 창 줄 정상 |

**계측기를 먼저 검증했다.** 첫 시도는 옛 도구를 scratchpad 에 복사해 돌렸고 셋 다
`ModuleNotFoundError: role_check` 로 죽었다 — 형제 모듈 import 실패, 즉 **내 설정
오류**였다. 그대로 적었으면 "옛 도구는 아카이브에서 죽는다"를 **엉뚱한 이유로**
증명한 셈이 된다. 도구를 `tools/logic-map/` 안에 두고 **활성 id 양성 대조군**(옛 도구가
진짜 출력을 낸다)을 통과시킨 뒤 다시 쟀다.

### 1.12 는 실데이터로 확인되지 않았다 — 차단이고, 이것을 4.4 의 입력으로 넘긴다

| 시도 | 결과 |
|---|---|
| 메인 브랜치 | `adoption requires detached HEAD` |
| a063 전용 worktree | `untracked/ignored input …: .codex-context/.save-session.lock` |
| HEAD 의 새 detached worktree | `required commit ancestry is absent` |

마지막의 원인: a063 이 감사한 `source_commit c727ad12` 는 HEAD 의 **조상이 아니고**
그 커밋에 이관 기록이 아직 없다. 남은 길은 a063 worktree 청소인데 그것은 a063 의
상태다 — 남의 change 작업 상태를 허락 없이 건드리지 않는다.

**세 지점 전부에서 옛 도구와 지금 도구의 출력이 글자까지 같다.** a122 가 닿을 수 있는
이관 경로에 관측 가능한 차이는 없다. 1.12 의 근거는 유닛 픽스처 3건 + 이 동일성이며
**실데이터 확인은 아니다.** 적대 리뷰가 여기를 봐야 한다.

- 측정 worktree 는 제거했고(`wt43-a063`) a063 · a120 의 기존 worktree 는 손대지 않았다.

## VERIFY — task 4.4 (독립 적대 diff/테스트 리뷰 + gstack)

- 대상 HEAD: `0cca39cb` · 비교 base: `1687baac` (origin/main 과의 merge-base)
- 리뷰한 실행 코드: `tools/` 5개 파일 (+1502/-23). `openspec/` 아래는 문서·동결 AST
  산출물이라 코드로 리뷰하지 않았다.
- 독립 원천 **다섯**: gstack `/review` 전문가 4(Testing·Security·Maintainability·
  Simplification+Performance, 각자 빈 문맥) · **codex `gpt-6-astra`**(외부 모델,
  read-only, high reasoning) · 그리고 이 세션이 직접 돌린 **뮤테이션 9개**.

### 이 리뷰의 결론 — **이대로 4.5 로 못 간다**

a122 는 정직한 경로에서 측정 가능하게 낫다(아래 §통과한 것). 막는 것은 하나다:
**spec 이 SHALL NOT 으로 금지한 "위조 가능한 착지 기록"이 실제로 위조된다.**
서로 모르는 세 원천이 각각 재현했고, 이 세션이 **RED → GREEN 대조군**으로 못 박았다.

### P0-1 — 착지 고정이 그 change 의 작업에 묶여 있지 않다 (재현 3회)

`resolve_landing` 은 "선언된 커밋에서 저자의 `revision: current` 번들 해시가 맞는가"만
묻는다. 번들이 **그 change 가 창 안에서 바꾼 파일을 기술하는가**는 묻지 않는다.
그래서 고정은 저자가 만든 값으로 저자가 고른 값을 검증하는 순환이 된다.

| # | 위조 수단 | 대조군(선언 없음) | 선언 뒤 |
|---|---|---|---|
| a | 증거를 **base 상태**로 쓰고 `landed-commit = base` | `AST source hash is stale` + `hash does not match` → **RED** | `[]` · required **0** → **GREEN** |
| b | 안 건드린 파일의 **미끼 번들** 하나 + `landed-commit = base` | `missing evidence … Own` · required 1 | `[]` · required **0** |
| c | 정직한 번들 둘 중 하나만 `revision: current` → **`base` 로 relabel** | `landing point … is not the revision this evidence describes` | `[]` · required **1**, `Other()` 무분석 착지 |

(a) 와 (c) 는 이 세션이 실측했다. (a) 의 대조군이 핵심이다 — 같은 거짓 증거가 a122
**이전에는 빨갛고 이후에는 초록이다.** a122 는 못 막은 게 아니라 **없던 길을 연다.**
(c) 는 JSON 한 필드(`"current"`→`"base"`)를 바꾸는 것이 전부다: `resolve_landing` 은
`base` 번들을 건너뛰고 `validate_target` 은 `elif revision != "base"` 로 해시를 아예
안 본다.

**이것은 구현 버그가 아니라 spec 의 틈이다.** 코드는 spec 이 적은 판정
("모든 `revision: current` 번들의 source hash 가 착지에서 일치")을 **정확히** 구현한다.
그 판정이 spec 이 같은 문단에서 내건 목표("위조 가능해서는 안 된다")에 못 미친다.

5.3 이 이것을 절반 예고했지만 축이 틀렸다. 5.3 은 "번들이 몇 개면 충분한가"를 묻는다.
실측은 **개수가 축이 아니라고** 답한다 — a099 는 고정 번들이 **37개**인데도 base..HEAD
239 커밋 중 **21개**가 그 고정을 통과한다. (a)(b) 는 번들 수와 무관하게 성립한다.
빠진 것은 개수가 아니라 **번들과 창 사이의 결속**이다.

### P0-2 — 아카이브된 a063 은 자기 id 로 재검사가 안 된다 (실측)

spec: "완료 게이트는 아카이브된 change 의 함수 분석도 그 id 로 재검사할 수 있어야
한다(SHALL)". `check` 는 이제 아카이브를 해소한다 — 그런데 이관 경로만은 못 간다.
`execution_baseline.validate` 가 `change_dir.name != CHANGE` 로 **날짜 없는 id** 를
요구하기 때문이다.

| 디렉터리 이름 | `validate` 결과 |
|---|---|
| `a063-align-attestation-renewal-profile` (오늘) | `adoption requires detached HEAD` — 이름 검사 통과 |
| `2026-09-11-a063-…` (아카이브 뒤) | **`execution-baseline adoption is not allowed for this change/base`** |

a063 의 실제 상태는 건드리지 않고 `execution-baseline.json` 만 날짜 붙은 이름의 임시
디렉터리에 복사해 쟀다. a063 이 아직 활성이라 **오늘은 잠복**이고, 아카이브되는 날
터진다. **4.3 이 "1.12 를 실데이터로 못 돌렸다"고 넘긴 바로 그 사각지대다** —
외부 모델이 그 안에서 찾았다.

### P1 — a122 자신의 가드 셋이 지워져도 스위트가 초록이다 (뮤테이션 9/9 실측)

각 가드를 `if False:` 로 바꾸고 focused 스위트(75)를 돌렸다. 원복은 sha256 동일성으로
확인했다.

| 뮤테이션 | 결과 |
|---|---|
| M1 빈 고정 거절 · M5 해시 불일치 · M6 착지에서 해싱 · M7 이관이 선언 거절 · M8 빌림이 양끝 공유 · M9 활성+아카이브 충돌 | **CAUGHT** (1~3 시험 빨강) |
| **M2 `base ≤ landing` 조상 판정** | **SURVIVED** |
| **M3 `landing ≤ HEAD` 조상 판정** | **SURVIVED** |
| **M4 40-hex 형태 판정** | **SURVIVED** |

원인은 하나다. `AForgedLandingPointIsRefusedByName._refuse(value, needle)` 의 바늘이
네 곳에서 `"landing"` 인데, 그 단어는 착지 관련 **모든** 오류 문장에 들어 있다.
그래서 가드를 지워도 **다른** 가드가 거절하고 시험은 초록으로 남는다. M3 은 시험
자체가 없다 — 곁가지 커밋을 만드는 픽스처가 하나도 없다.

이 파일은 같은 함정을 **이미 한 번 발견하고 한 자리만 고쳤다**:

> `# 사유를 **착지 판정의 문장**으로 못 박는다. … 착지 판정을 통째로 지우는 변이가
> 이 시험을 초록으로 통과한다(M4 로 실측).`

`test_a_landing_the_evidence_does_not_describe` 만 구체 문장으로 바뀌었고 형제 넷은
느슨한 바늘에 남았다. 저장소의 기록된 교훈 그대로다 —
[[passing-test-is-not-evidence]] · "거절 테스트는 실패 **지점**을 단언할 것".

### P2 — 나머지 확인된 것

| # | 내용 | 성격 |
|---|---|---|
| P2-1 | 착지를 선언하면 **커밋 안 된 Go 편집이 5단계에 안 보인다**. 대상이 커밋이므로 워킹트리·index 를 아무도 안 본다. `make gate` 는 `make sdd-check`(fingerprint) 로 일부 가리지만 `make sdd-sync` 를 다시 돌리면 풀린다 | a122 가 만든 행동 변화 |
| P2-2 | `base-commit.txt` 는 여전히 **워킹트리에서** 읽고 40-hex 검사도 없다. `landed-commit.txt` 는 HEAD 에서 읽는데 **같은 창의 반대쪽 끝**이 안 잠겨 있다 | a122 **이전부터** 있던 것 |
| P2-3 | `ARCHIVED_CHANGE` 의 `\d` 는 유니코드 숫자를 먹고 `gate.sh` 의 `[0-9]` 는 안 먹는다 → 두 해소기가 "날짜 접두사"의 뜻에 동의하지 않는다. `re.ASCII` 한 글자 | a122 신규 |
| P2-4 | `check` 의 `except ValueError` 가 `archive holds N copies` 를 삼키고 `missing base-commit.txt` 로 바꿔 말한다. `gate.sh` 는 같은 조건에 큰 소리로 멈춘다 | a122 신규 |
| P2-5 | `gate.sh:110` 주석이 "Python 보다 엄격하다"고 하는데 **같은 diff 가** Python 쪽에 `AmbiguousChange` 를 넣어 그 차이를 없앴다 | 낡은 주석 |
| P2-6 | `test_check_analysis.py` docstring 이 지워진 코드(`return direct`)를 현재형으로 인용하고, `:689-692` 는 이제 `validate_target` 의 무관한 `except` 를 가리킨다 | 낡은 인용 |
| P2-7 | 규칙이 두 집(`resolve_change_dir` / `resolve_referenced_change`)에 사는데 **동기화를 강제하는 것이 아무것도 없다**. 각 스위트가 자기 사본만 못 박아서, 한쪽만 고치면 둘 다 초록이다 | 구조 |
| P2-8 | 성능: 같은 (ref, path) blob 을 `resolve_landing` 과 `validate_target` 이 **두 번** 해싱한다(a099 실측 74 spawn → 서로 다른 blob 12개). 게이트 1회엔 작지만 126-id 전수에서 곱해진다 | advisory |

### 통과한 것 — 이 리뷰가 확인한 a122 의 실제 이득

- 정직한 경로는 실물에서 작동한다: a099 의 고정 번들 37개가 기술하는 **12개 파일이 전부
  창 안**이고 창 밖은 **0개**다. 저자가 정직하면 이 기계는 맞는 답을 낸다.
- 가드 6/9 는 뮤테이션으로 **빨개진다** — M1·M5·M6·M7·M8·M9.
- 아카이브 재검사는 이관 경로를 빼면 열렸다(4.3 의 A/B).
- 빌린 증거의 창은 양끝이 잠겼다(M8).
- 생산 코드 diff 는 이 리뷰에서 **한 줄도 바꾸지 않았다** — 리뷰가 잰 HEAD 와 기록이
  가리키는 HEAD 가 같아야 재현되기 때문이다. P0·P1 은 새 task 로 연다.

### gstack 리뷰

`/review` 워크플로(gstack 1.72)를 정본대로 돌렸다 — 전문가 병렬 dispatch, codex
적대 pass(`CODEX_MODE: ready`), 교차 합성. codex 의 종결 한 줄:

> `Recommendation: Block merge because one valid base-era bundle lets an author declare
> the base as the landing and obtain zero required functions despite subsequent Go changes.`

교차 확인: P0-1 은 **세 원천**(codex · Security 전문가 · 이 세션의 대조군)이 독립으로
잡았다. P0-2 는 codex 단독, 이 세션이 실측으로 확증. P1 은 이 세션 단독(뮤테이션).

### P0-3 — 이관 change 의 깨진 착지 기록이 `check()` 를 **터뜨린다** (Testing 전문가 실측)

`check` 의 `_declared_landing` 호출 **다섯 자리 중 한 곳만** try 밖에 있다.

```
812:        try:            # 빌린 증거 경로 — 감싸져 있다
819:    adopted = bool(facts.get("execution_baseline_adoption"))
820:    if adopted and _declared_landing(change_dir, root) is not None:   # ← try 밖
829:    try:            # 착지 해소 경로 — 감싸져 있다
```

`_declared_landing` 은 비-UTF-8 기록에 `ValueError("landing point is not UTF-8: …")`
를 던진다(a122 신규). 820 에서 던지면 `check()` 를 뚫고 나가 `main()` 이 traceback
으로 죽고, **3.3 이 보장하기로 한 `[logic-map]` 창 줄이 하나도 안 찍힌다.**
저장소 자신의 `_adoption_with_complete_bundle` 픽스처로 실측됐다:

> `RAISED OUT OF check(): ValueError landing point is not UTF-8: …/landed-commit.txt`

`adopted` 가 거짓이면 단락되므로 **오늘 닿는 change 는 a063 하나**다. 6.2 와 같은 곳이다 —
이관 경로에 결함이 **둘** 있고 둘 다 4.3 이 "실데이터로 못 돌렸다"고 넘긴 사각지대 안이다.

### P1 — 뮤테이션: 두 원천이 독립으로 같은 곳을 짚었다

Testing 전문가가 이 세션과 **따로** 같은 실험을 하고 같은 결론에 닿았다(가드 넷을
각각 지워도 75 전부 초록). 번호만 다르다. 여기에 이 세션이 안 건드린 셋을 더 찾았다:

| 지운 것 | 스위트 | 무엇이 달라지나 |
|---|---|---|
| `gate.sh` 1단계 `RESOLVE_ERROR` 블록 | 8/8 초록 | hits>1 이면 fallback 활성 디렉터리가 **실재**해서 변이가 4단계까지 간다 — 해소기 주석이 막는다고 적은 "활성을 조용히 고르기" 바로 그것 |
| `gate.sh` 의 `YYYY-MM-DD-` `case` 가드 | 8/8 초록 | 남은 `${name#????-??-??-}` 의 `?` 는 아무 글자나 먹어서 `archive/abcd-ef-gh-a998-theirs` 가 `a998-theirs` 로 해소된다 |
| `_target_text` 의 비이관 갈래를 "audited" 로 고정 | 75/75 초록 | 평범한 선언도 `audited source-commit` 으로 찍힌다. 시험이 SHA 만 `assertIn` 해서 못 본다 — 이 함수의 docstring 이 "**무엇이** 그 끝을 고정했는지까지 말한다"고 적은 바로 그 구분 |

`gate.sh` 쪽 둘은 특히 나쁘다. **거절이 다른 이유로 일어나서** 시험이 초록인 것이
아니라, 한쪽(hits=0)은 다른 가드가 받고 다른 쪽(hits>1)은 **아무도 안 받는데**
시험이 hits=0 만 본다.

### P2 — 픽스처가 개발자의 전역 git config 를 상속한다 (실측)

`_init_fixture` 는 `user.email`·`user.name` 만 세운다. 전역 config 에
`commit.gpgsign = true` 가 있으면 새 클래스 전부가 `gpg: signing failed: No secret key`
로 **에러**가 된다(실측). 같은 노출: `core.hooksPath` · `gpg.format` · `core.autocrlf`.
`GIT_CONFIG_GLOBAL=/dev/null` 을 주거나 두 `git config` 줄 옆에 명시적으로 끈다.
나머지는 hermetic 하다 — 절대경로, CWD 비의존, 네트워크 0.

## VERIFY — task 6.1 (측정 단계)

**대상** — 6.1 이 "숫자·규칙을 지어내지 말고 활성 13건 전체에 대해 실데이터로 가른 뒤
정한다"고 스스로 적은 그 측정. **HEAD** `508f8b46`. **생산 코드 변경 0** —
`check_analysis.py` 도 `gate.sh` 도 안 건드렸다. 커밋된 역사만 읽는다.

**방법** — `git cat-file --batch-check` + `--batch` 로 (커밋 × 번들 소스) 쌍의 blob 을
한 번에 받아 sha256 을 낸다. 번들 판정 조건은 `resolve_landing` 의 순회와 **글자
그대로 같다**(`revision` 기본 `"current"`, `file` 과 `source_sha256` 이 둘 다 있을 것).
워킹트리는 한 번도 읽지 않는다.

### 실측 — 활성 13건 전수

`후보` = base..HEAD 안에서 그 change 의 고정을 **통과하는** 커밋 수(base 포함).
`A@경계` = 창 안에서 바뀐 Go 파일로 고정 번들을 한정했을 때 남는 수.
`B축` = 착지를 "그 change 의 고정 번들이 역사에 들어온 마지막 커밋" 뒤로 제한했을 때
남는 후보 수. `C: 최저→최고` = 가장 낮은 후보와 가장 높은 후보에서 창 안 Go 파일 수.

| change | base | 고정번들 | 소스 | 후보/창 | base도 후보? | A@경계 | B축 | C: 최저→최고 |
|---|---|---:|---:|---:|:---:|---:|---:|---:|
| a063 | `da80ce31` | 6 | 4 | **35**/353 | no | 6 | 35 | 855 → 866 |
| a066 | `23794f86` | 23 | 8 | **1**/340 | no | 23 | 1 | 149 → 149 |
| a071 | `775c37cb` | 35 | 20 | **2**/336 | no | 34 | 2 | 94 → 102 |
| a074 | `448dfeb1` | 7 | 4 | **15**/291 | **YES** | **0** | 14 | 0 → 35 |
| a077 | `448dfeb1` | 7 | 4 | **3**/291 | **YES** | **0** | 2 | 0 → 0 |
| a079 | `448dfeb1` | 4 | 4 | **3**/291 | **YES** | **0** | 2 | 0 → 0 |
| a089 | `ec29dc72` | 6 | 2 | **3**/266 | **YES** | **0** | **0** | 0 → 0 |
| a091 | `ec29dc72` | 2 | 2 | **27**/266 | **YES** | **0** | 24 | 0 → 23 |
| a092 | `285c7619` | 36 | 12 | **18**/258 | **YES** | **0** | 17 | 0 → 14 |
| a094 | `ec29dc72` | 15 | 11 | **59**/266 | **YES** | **0** | 56 | 0 → 74 |
| a095 | `ec29dc72` | 9 | 7 | **3**/266 | **YES** | **0** | **0** | 0 → 0 |
| a100 | `882a0b49` | 57 | 31 | **28**/158 | no | 39 | 28 | 34 → 77 |
| a112 | `aeeb209e` | 132 | 48 | **39**/73 | no | 47 | 39 | 99 → 135 |

### 1. 오늘의 고정은 **어느 change 에 대해서도** 착지를 결정하지 못한다

13건 중 후보가 하나인 것은 a066 **한 건**뿐이다. 나머지 12건은 2~59개 중에서 저자가
고른다. 그리고 후보 수는 번들 수와 **무관하다** — a112 는 번들 132개에 후보 39개,
a091 은 번들 2개에 후보 27개다. 잔여 5.3 이 묻는 "번들이 몇 개면 충분한가"에는
**답이 없다.** 개수는 축이 아니다. [[author-supplied-evidence-cannot-pin-an-author-choice]]

### 2. 8건은 base 자신이 후보다 — 위조가 아니라 오늘의 상태다

`landed-commit = base` 를 적으면 창이 비어 required 0 이 되는 change 가 **8건**이다.
4.4 가 위조 재현으로 보인 첫 번째 수단이 a074·a077·a079·a089·a091·a092·a094·a095
에서 **지금 그대로 쓸 수 있다.** 이것을 "8건이 위조다"로 읽으면 안 된다 — 그 8건은
2026-08-04 재기준화·freeze 순서 때문에 Go 작업이 자기 base **앞**에 착지한, 4.2 가
이미 잰 정상 상태다. 문제는 정상과 위조가 **같은 모양**이라는 것이다.

### 3. A축(창 안 Go 파일로 한정)은 측정 이전에 **구조적으로** 그 8건을 거절한다

4.4 는 이것을 a074·a077·a079 3건의 실측으로 적었는데, 실은 증명이다.

> 고정 번들이 base 에서도 착지에서도 해시가 맞으면, 그 파일은 창 안에서 안 바뀐
> 것이다. 따라서 "창 안에서 바뀐 Go 파일" 필터는 **반드시** 0 을 남긴다.

표의 `A@경계` 열이 base 가 후보인 8건 전부에서 0 인 것은 우연이 아니라 이 동어반복
이다. A축은 곧 "자기 창 안에 Go 작업이 있는 change 만 착지를 선언할 수 있다"이고,
없는 8건은 예외 없이 막힌다. 5.7 과의 충돌은 3건이 아니라 **8건, 그리고 구조적**이다.

### 4. B축(착지 ≥ 번들이 역사에 들어온 커밋)은 저자가 소급할 수 없지만 부족하다

13건 **전부** 번들이 자기 base 뒤에 커밋됐다. 그래서 이 축은 `landed-commit = base`
를 6건에서 걷어낸다(a074 15→14, a077 3→2, a079 3→2, a091 27→24, a092 18→17,
a094 59→56). 위조자는 오늘 만든 번들을 과거 커밋에 넣을 수 없으므로 축 자체는 정당
하다. 그러나 **후보를 하나로 좁히는 것은 여전히 a066 한 건뿐이고**, a089 와 a095 는
후보가 **0** 이 되어 완전히 막힌다.

그 0 의 이유가 그냥 넘길 모양이 아니다. 둘의 번들은 `a30eb35ae` 한 커밋에 들어왔는데
**같은 커밋이 그 소스 Go 파일을 바꾼다**:

| change | 번들이 안 맞는 파일 | @base | @번들커밋 `a30eb35a` | @HEAD |
|---|---|:---:|:---:|:---:|
| a089 | `internal/journal/outbox.go` | MATCH | **no** | no |
| a095 | `internal/obs/notifier.go` (번들 2개) | MATCH | **no** | no |

`revision: current` 라고 적혀 있지만 그 증거는 **자기를 담은 커밋에서 이미 틀렸다.**
생성 뒤 같은 세션에서 Go 를 한 번 더 고치고 둘을 한 커밋에 담으면 이 상태가 된다.
오늘 아무 게이트도 이것을 못 본다 — 대상이 워킹트리라 HEAD 하고만 비교하기 때문이다.

### 5. C축(고를 수 없게, 허용 후보 중 가장 늦은 것으로 강제)은 남의 작업을 되돌린다

선택이 완전히 사라지고 계산으로 정해진다는 점에서 가장 깨끗하다. 대가를 쟀다:
a074 **0 → 35**, a091 0 → 23, a092 0 → 14, a094 **0 → 74**, a100 34 → 77,
a112 99 → 135, a071 94 → 102, a063 855 → 866 (Go 파일 수). a074·a091·a092·a094 는
a122 가 없애려던 바로 그 증상 — **남의 change 의 함수를 자기 요구 집합에 담는 것** —
으로 돌아간다. a077·a079·a089·a095 넷만 0 을 유지한다.

### 6. "선택이 답을 바꾸지 않을 것"을 요구하면 6/13 이 통과한다

가장 낮은 후보와 가장 높은 후보에서 요구 집합이 **같아야** 한다고 요구하면
(= 저자가 무엇을 고르든 판정이 같다), a066·a077·a079·a089·a095 와 후보가 1인 경우만
남는다. 나머지 7건은 거절되는데, 거절 사유가 **실행 가능**하다는 점이 다르다 —
"당신의 증거는 착지를 정하지 못한다, 어느 파일의 번들이 없는지가 이것이다". 다만
a074 는 여기서도 거절되고, 그것을 풀려면 a074 가 **안 건드린 Go 파일 35개**의 번들을
요구하게 되므로 답이 아니다.

### 결론 — 두 요구는 **증거만으로는** 양립하지 않는다

네 축을 13건 전부에 걸어 본 결과다.

| 축 | 위조 차단 | 5.7(base 가 작업 뒤) | 비용 |
|---|---|---|---|
| A 창 안 Go 파일 한정 | 예 | **8건 전부 거절**(구조적) | a122 의 목적이 사라진다 |
| B 착지 ≥ 번들 커밋 | 부분(6건에서 base 제거) | a089·a095 거절 | 후보는 여전히 2~56 |
| C 가장 늦은 후보 강제 | 예(선택 없음) | 통과 | 4건이 남의 작업을 다시 센다 |
| D 선택 불변 요구 | 예 | a074 거절 | 안 건드린 파일의 번들을 요구 |

**가르는 정보가 change 디렉터리 안에도 창 안에도 없다.** a074 의 정답이 "가장 낮은
후보"이고 위조자의 정답이 "가장 낮은 후보가 아님"인데, 둘 다 증거가 base 에서 맞는다.
차이는 "이 change 의 Go 작업이 실제로 어디에 착지했는가"인데 그것은 커밋 귀속
정보이고, 저장소는 그것을 기록하지 않는다.

따라서 6.1 은 "어느 규칙을 쓸까"가 아니라 **"무엇을 새로 기록할 것인가"** 이고,
그 선택은 사람이 한다. 숫자는 지어내지 않았고 규칙도 고르지 않았다.

## VERIFY — task 6.1.2 (구현)

**결정** — 사람이 2026-09-11 에 1번(**게이트가 기록한다**)을 골랐다. 값을 만드는 주체가
저자에서 도구로 바뀌는 유일한 안이라는 것이 근거다.

**내가 선택지에 적은 "도구가 base 를 쓴다"는 틀렸고 그대로 구현하지 않았다.** base 는
위조와 같은 모양이다(6.1.1 소견 2). 6.1.1 이 실제로 준 값은 다르다.

> **바닥** = 그 change 의 고정 번들이 역사에 들어온 마지막 커밋.
> **착지** = base..HEAD 안에서 바닥 이후이면서 고정을 통과하는 **가장 낮은** 커밋.

저자는 바닥 아래를 못 고른다 — 오늘 만든 번들을 과거 커밋에 넣을 수 없기 때문이다.
13건 **전부** 번들이 자기 base 뒤에 커밋됐으므로(6.1.1 소견 4) 이 하한은 전수로 실재한다.

### 도구가 13건에 대해 실제로 계산하는 값

| change | 기록될 값 | 창 안 Go | 비고 |
|---|---|---:|---|
| a074 · a077 · a079 | `840b3377` | **0** | 4.2 가 실제로 선언해 required 0 을 얻은 그 값 |
| a091 · a094 | `a30eb35ae` | 6 | |
| a092 | `7f3cbb036` | 2 | |
| a100 | `016da6245` | 34 | |
| a112 | `4f49a8ebc` | 99 | |
| a066 | `a37d97f52` | 149 | |
| a071 | `6aec9791b` | 94 | |
| a063 | — | — | 이관 경로라 거부(spec 이 이 기록을 안 받는다) |
| a089 · a095 | — | — | 거부: 그 증거가 이 역사의 어느 리비전도 기술하지 않는다 (**6.5**) |

a122 가 풀어 주려던 셋이 값을 얻고, 증거가 깨진 둘은 **이름으로 거부된다.**

### 4.4 의 위조 셋을 새 코드로 재실행했다 — 셋 다 막힌다

4.4 가 쓴 **그 스크립트 그대로** 돌렸다(재구성이 아니다). 미끼 하나만 스크립트가 없어
같은 헬퍼로 만들었다.

| 위조 | 4.4 (a122 구현 직후) | 6.1.2 뒤 |
|---|---|---|
| 증거를 base 상태로 쓰고 `landed-commit = base` | `[]` · required **0** | **거절** · `precedes the evidence that pins it` |
| 안 건드린 파일의 미끼 번들 + `landed-commit = base` | `[]` · required **0** | **거절** · 같은 문장 |
| 번들 하나를 `revision: base` 로 relabel | `[]` · required 1, `Other()` 무분석 착지 | **거절** · 같은 문장 |

**대조군은 안 움직였다.** 선언 없는 같은 거짓 증거는 전과 같이 `AST source hash is
stale` 로 빨갛고(D), 정직한 증거 + 선언 없음은 전과 같이 초록이다(A). 정직한 증거로
base 를 선언하는 경우는 전과 같이 `is not the revision this evidence describes` 로
거절된다 — **기존 바늘이 안 바뀌었다**는 뜻이고, 그래서 하한을 불일치 판정 **뒤에**
세웠다(6.3 이 이미 한 번 고친 그 자리를 도로 무디게 만들지 않기 위해서다).

### 아직 열려 있는 것 — 재서 적는다

증거를 **먼저** 커밋하고 그 지점을 기록한 뒤 Go 작업을 **그 뒤에** 붙이면 창이 그
작업을 안 담는다. 픽스처 실측: `record_landing` rc=0 → 기록 커밋 → `[]` required 0 →
Go 작업 커밋 → **여전히 `[]` required 0**. 6.1.2 가 만든 구멍이 아니라 6.1.2 **뒤에
남은** 것이고 **6.6** 으로 열었다. 규칙은 지어내지 않았다.

### 통과는 증거가 아니다 — 변이 (원복은 sha256 동일성으로 확인)

| 변이 | 1차 | 시험 보강 후 |
|---|---|---|
| M1 하한 판정 삭제 | **CAUGHT** | CAUGHT |
| M2 `if not floor:` 삭제 | SURVIVED | **CAUGHT** |
| M3 `-M --diff-filter=MA` → 평범한 `git log` | SURVIVED | **CAUGHT** |

M2·M3 이 살아남아서 시험 둘을 더 썼다 — 커밋 안 된 증거는 고정하지 못한다는 것과,
**아카이브가 기록을 무효로 만들지 않는다**는 것. 후자는 spec 의 "아카이브된 change 도
그 id 로 재검사(SHALL)"와 같은 문장이고, 저장소의 유일한 실물 기록 a099 가 정확히 그
모양이다(바닥 `21a315d1` 은 기록된 착지 `e6c4636a` **앞**, rename 커밋은 뒤).

### 거부할 정상 입력 (먼저 열거했다)

`record_landing` 이 거부하는 것: 고정 번들이 없는 change(a122 자신·a067·a113 …) ·
증거를 빌리는 change · 이관 change(a063) · 기록이 이미 있는 change · **추적 파일이
수정된 워킹트리**. 마지막 하나가 유일하게 새로 거부하는 정상 상태다 — 기록은 커밋된
지점을 가리키는데 그 상태의 Go 편집은 어느 커밋에도 없으므로, 그대로 쓰면 5단계가 못
보는 편집이 생긴다. [[fail-closed-must-name-what-it-rejects]]

### 안 한 것

`make gate` 가 기록을 **자동으로 쓰게 하지 않았다**(6.1.2.6 으로 열었다). 검사 게이트가
워킹트리를 바꾸면 게이트의 뜻이 달라지고, 저장소 규칙이 mutating 단계를 사람 승인으로
묶는다. 5단계 조언 줄이 그 명령을 이름으로 부르고, 값은 여전히 도구가 계산한다.

## VERIFY — task 6.2 · 6.2.1 (이관 경로의 두 P0)

### 순서를 지켰다

6.1.2 는 열거를 편집 뒤에 뽑았다. 이번에는 다섯 함수(`validate` · `resolve_base` ·
`_declared_landing` · `check` · `record_landing`)의 `ast.before-6.2.json` 을 HEAD blob 에서
**먼저** 뽑고, 각 FLM 에 "편집 계획"을 코드보다 먼저 적었다. 계획의 예측(분기 수 42·11·43·11
그대로, `_declared_landing` 5→3·반환 3→2, "다섯째 막는 자리 없음")을 편집 뒤 같은 열거기로
대조했고 전부 맞았다. 계획 문장 하나는 편집 **전에** 틀린 것을 찾아 고쳤다 — "기본값 갈래를
부르는 시험은 `SDD_PYTHON` 이 건너뛰는 하나뿐"은 거짓이었다(`test_execution_baseline.py` 가
20 번 부른다). 결론(필수 인자)은 같고 사유가 다르다.

### 6.2 — 막는 자리는 넷이었다 (4.4 는 하나를 쟀다)

생산 코드를 건드리지 않고 `validate` 의 메모리 사본에서 가드를 하나씩 풀어 다음 실패를 쟀다.

| 푼 가드 | 다음 실패 |
|---|---|
| 없음 | `execution-baseline adoption is not allowed for this change/base` |
| 이름 | `adoption evidence path is outside current change analysis` |
| + 접두사 | `missing evidence: …/a063-…/analysis/execution-baseline-ledger.json` |
| + 원장 읽기 | `missing evidence: …/a063-…/analysis/adversary.md` |
| + 리뷰 읽기 | **통과**, `effective_base == E` |

기전: 이관 기록은 옮기기 **전** 경로를 적는다 — `draft` 가 그렇게 쓰고 a063 실물의 세 경로도
`openspec/changes/a063-…/analysis/…` 다. 4.4 처방 "신원만 분리하고 경로·digest 검사는
그대로"는 앞 절반만 맞다. 그대로 두면 이름을 고쳐도 둘째 줄에서 막힌다.

수리: 신원 = 게이트가 요청받은 id(`validate` 의 **필수** 인자 — 호출자 셋이 각자 해소한 id
를 넘긴다) · 경로 판정 = 적힌 자리 · 읽기 = 지금 자리. digest 비교는 한 글자도 안 바꿨다.
`validate` 는 아카이브 문법을 새로 배우지 않는다 — 그 문법의 집이 이미 둘이다(6.4(f)).

**수리가 없앤 벽 하나를 직접 쟀다.** 편집 전에는 a063 디렉터리를 다른 id 로 통째로
복사하면 이름에서 막히고, 이름을 지워도 접두사에서 한 번 더 막혔다. 편집 뒤 신원 판정을
지우면 복사본 둘(활성 이름·아카이브 이름)이 **통과**한다(`effective_base == E`). 복사를
막는 것은 이제 신원 하나이고 복사 시험이 그것을 못 박는다. 변이 M2 의 CAUGHT 만으로는 이
말을 할 수 없었다 — 바늘이 신원 판정의 문장이라 다른 검사가 막아도 빨개지기 때문이다.
[[surviving-mutant-may-mean-accidental-safety]]

### 6.2.1 — try 밖 자리는 둘이었다 (4.4 는 하나를 셌다)

`_declared_landing` 호출 자리를 AST 로 다시 셌다(편집 전 HEAD `fb4e8f92`).

| 줄 | 함수 | 묻는 것 | try |
|---|---|---|---|
| 486 | `resolve_landing` | 값 | 호출자 `check` 가 받는다 |
| 891 ×2 | `check` 빌린 증거 | 값 | `except ValueError` |
| **898** | `check` 이관 probe | **있는가** | **없음** — 4.4 가 센 자리 |
| **1027** | `record_landing` | **있는가** | **없음** — 6.1.2 가 만든 자리, 4.4 뒤라 셀 수 없었다 |

두 자리 다 "있는가"를 해독 함수에 물었다. try 를 두르지 않고 질문을 떼어 냈다 —
`_landing_record` 는 있는가만 답하고 `_declared_landing` 이 그 위에서 해독한다(오류 문장
그대로). spec 이 이관 경로의 착지 기록에 "받지 않는다고 이름으로 말한다"를 요구하므로 못
읽는 기록도 **거절**로 답한다 — 4.4 처방(해독 오류로 돌려주기)과 다르다.

저장소에 비-UTF-8 착지 기록을 재는 시험이 **0** 이었다. 셋을 더했다.

### RED → GREEN

| 시험 | 편집 전 | 편집 후 |
|---|---|---|
| `test_an_archived_adoption_is_rechecked_by_its_id` | **FAIL** — `not allowed for this change/base` | OK |
| `test_an_undecodable_landing_record_in_the_adoption_path_is_refused_not_raised` | **ERROR** — `ValueError` 가 `check()` 를 뚫음 | OK |
| `test_an_undecodable_committed_record_is_reported_not_raised` | **ERROR** — 같음, `record_landing` | OK |
| `test_a_copied_adoption_record_does_not_make_another_change_a063` | OK — 이름이 막음 | OK — 신원이 막음 |
| `test_the_value_path_still_names_an_undecodable_record` | OK | OK |

뒤의 둘은 RED 가 아니라 **회귀 고정**이다(편집 전에 초록인 것이 의도다). 직접 부르는 시험
22 자리(`test_execution_baseline.py` 20 · `test_check_analysis.py` 2)는 **디렉터리 이름**을
id 로 넘기게 바꿨다 — 각 시험이 재던 뜻(이름 = 신원)을 한 글자도 안 바꾸려는 것이다.
`resolve_base` 를 직접 부르는 시험 둘도 같다.

### 변이 9 — 전부 CAUGHT, 한 변이에 한 시험 (`62_mutations.py`, 원복 sha256 동일)

| 변이 | 빨개진 시험 |
|---|---|
| M1 신원을 다시 디렉터리 이름으로 | 아카이브 |
| M2 신원 판정 삭제 | 복사 |
| M3 접두사를 지금 자리로 판정 | 아카이브 |
| M4 원장을 적힌 자리에서 읽기 | 아카이브 |
| M5 리뷰를 적힌 자리에서 읽기 | 아카이브 |
| M6 `resolve_base` 가 id 대신 이름을 넘김 | 아카이브 |
| M7 이관 probe 가 다시 해독 | 이관 비-UTF-8 |
| M8 `record_landing` probe 가 다시 해독 | 덮어쓰기 비-UTF-8 |
| M9 값 경로가 관대하게 해독 | 값 경로 대조군 |

대조군(무변이)은 `-k adoption -k undecodable` 로 고른 16 시험 전부 초록.

### 실데이터 — 확인이 아니다

실물 a063 은 이 HEAD 에서 `adoption requires detached HEAD` 앞을 못 지난다 — 편집 전후 같은
줄이고 4.3 이 적은 줄과 같다. 4.3 이 쟀듯 detached worktree 에서도 감사된 `source_commit
c727ad12` 가 HEAD 의 조상이 아니라 막힌다. 이 수리의 근거는 **저장소 픽스처**다.

### 안 한 것

- 6.3(거절 시험 바늘) · 6.4 · 6.5 · 6.6 은 이 로트 밖이다.
- 6.4(c) `ARCHIVED_CHANGE` 의 `\d` 는 안 건드렸다 — `validate` 가 그 문법을 안 쓰므로 이
  수리와 무관하다.
- gstack 독립 리뷰는 §6 로트를 묶어 한 번 돈다. 아직 안 돌렸다.

### 게이트 (2026-09-11, 이 로트의 워킹트리)

| 게이트 | 결과 |
|---|---|
| `make sdd-test` | rc=0 — scripts 15 · logic-map **165**(skip 1) · sdd 69 · sdd-history 22 · pm 15 · deploy 18 |
| `openspec validate --all --strict` | 58/58 |
| a122 5단계 (`check_analysis --change a122-…`) | rc=0, required 0 |
| `make sdd-check` | rc=0 (advisory 경고 둘: codegraphcontext · gbrain) |
| `make sdd-sync` | **rc=2 — 두 번 돌려 두 번 같다.** `codegraphcontext` 가 `Could not set lock on file …/global/db/kuzudb` |

`sdd-sync` 실패는 재시도로 안 풀린다 — 시간 초과가 아니라 잠금이다. 잠금을 쥔 것은 pid 47230
`cgc mcp start`(CodeGraphContext MCP 서버, `fuser`·`lsof` 로 확인)이고 GBrain 은 `gbrain serve`
pid 55793 이 쥐고 있다. 둘 다 **advisory** 이고 hard CodeGraph fingerprint 는 신선하다
(`sdd-check` rc=0). 남의 세션일 수 있는 MCP 프로세스는 죽이지 않았다
([[tossos-parallel-session-gate-contention]]).

## VERIFY — task 6.3 (거절 시험이 실패 지점을 단언한다)

생산 코드 변경 **0** — 시험만 바꿨다. 함수의 갈래를 근거로 쓰므로 `resolve_landing`
(`ast.before-6.3.json`, 편집 없음)과 `_target_text`(`ast.json`)를 먼저 열거했다. gate.sh 는
shell 이라 저장소 열거기가 없다 — **not-applicable: shell 함수의 AST 도구가 없다.** 그 자리의
증거는 변이가 대신한다.

### 먼저 쟀다 — 일곱 전부 SURVIVED

| 변이 | 자리 | 6.3 전 | 6.3 후 | 빨개진 시험 |
|---|---|---|---|---|
| G1 40-hex 판정 삭제 | `resolve_landing` raise 510 | SURVIVED | **CAUGHT** | revision expression · empty record |
| G2 커밋 실재 판정 삭제 | raise 518 | SURVIVED | **CAUGHT** | not a commit |
| G3 `landing ≤ HEAD` 삭제 | raise 520 | SURVIVED | **CAUGHT** | never landed (신규) |
| G4 `base ≤ landing` 삭제 | raise 525 | SURVIVED | **CAUGHT** | before the base |
| T1 비이관 라벨을 감사 라벨로 | `_target_text` IfExp | SURVIVED | **CAUGHT** | empty required |
| GS1 1단계 `RESOLVE_ERROR` 블록 삭제 | `gate.sh` 1단계 | SURVIVED (8/8) | **CAUGHT** | target open+archived (신규) |
| GS2 `YYYY-MM-DD-` case 가드 삭제 | `gate.sh` 해소기 | SURVIVED (8/8) | **CAUGHT** | 날짜 없는 아카이브 이름 (신규) |

전은 HEAD 의 시험이고(`63_gate_mutations.py` 는 HEAD blob 의 gate 시험을 `tools/sdd/` 에
임시 이름으로 두고 돌린 뒤 지웠다), 후는 지금 시험이다. 같은 스크립트, 원복 sha256 동일,
대조군 초록(logic-map 91 · gate 10).

4.4 는 여섯을 셌다(착지 가드 셋 + gate 둘 + 라벨). **G2(커밋 실재)는 4.4 표에 없었다** —
같은 원인(옛 바늘 `"landing"` 이 raise 여덟 문장 전부에 들어 있다)이라 같이 드러났다.

### 무엇을 바꿨나

- `_refuse(value, needle)` 의 바늘 넷을 **그 가드의 자기 문장**으로. 이 파일이 한 자리
  (`test_a_landing_the_evidence_does_not_describe`)에서 이미 한 처방을 나머지에 했다.
- `_refuse` 에 `prepare` 를 더해 곁가지 커밋 픽스처를 만들었다 — `landing ≤ HEAD` 에는
  시험이 **아예 없었다**.
- 라벨: SHA 만 보던 자리를 **그 자리에서** `landed-commit <sha>` 로 좁혔다. 옆에 새 시험을
  세우지 않은 이유 — 느슨한 자리가 남으면 다음 사람이 그것을 베낀다.
- gate.sh 시험 둘: 대상이 활성·아카이브에 **동시에** 있으면 1단계가 `확정할 수 없는
  change-id` 로 멈춘다 · 날짜 모양이 아닌 이름(`abcd-ef-gh-<id>`)은 아카이브가 아니다.
  옛 스위트가 두 변이를 못 잡은 이유는 도우미가 아니라 **그 입력을 만드는 픽스처가 없었던
  것**이다 — 새 픽스처는 `assert_stopped_before_step_four` 만으로도 변이를 잡는다. 문장
  단언은 그 위에 얹었다: 다른 단계가 같은 입력을 4단계 전에 막아도 초록이 되지 않게.

### 정상 입력

바꾼 것이 시험뿐이라 새로 거부되는 정상 입력은 없다. 새 픽스처 셋은 각 가드가 **거절해야
할** 입력이고, 각 클래스의 양성 대조군(`test_the_fixture_passes_before_any_landing_record` ·
`test_active_pair_still_works` · `test_archived_pair_is_found`)은 그대로 초록이다.

### 게이트 (2026-09-12, 이 로트의 워킹트리)

| 게이트 | 결과 |
|---|---|
| `make sdd-test` | rc=0 — scripts 15 · logic-map **166**(skip 1) · sdd **71** · sdd-history 22 · pm 15 · deploy 18 |
| `openspec validate --all --strict` | 58/58 |
| a122 5단계 | rc=0, required 0 |
| `make sdd-check` | rc=0 (advisory 경고 둘) |
| `make sdd-sync` | rc=2 — 6.2 때와 같은 잠금. `fuser` 로 다시 보니 여전히 pid 47230(`cgc mcp start`) |
| 생산 코드 | `git diff --stat -- tools/` 에 시험 파일 둘뿐(+54 −6) |

## VERIFY — §6 로트 독립 리뷰 (gstack /review, 2026-09-12)

범위: `508f8b46..HEAD`(4.4 가 본 커밋 뒤) 의 코드 다섯 파일, +697 −76 — 6.1.2 `fb4e8f92` ·
6.2/6.2.1 `94702b5c` · 6.3 `007c015e`. 원천 여덟: Claude 구조 리뷰 · 전문가 다섯(testing ·
maintainability · security · performance · simplification) · Red Team · Claude 적대(시험 파일은
요약 모드) · Codex 적대 · Codex 구조 리뷰(`--base 508f8b46`, `[P1]` 0 → GATE PASS).
**이 리뷰는 코드를 한 줄도 바꾸지 않았다** — 4.4 와 같은 이유(리뷰가 잰 HEAD 와 기록이
가리키는 HEAD 가 같아야 재현된다)에 더해, 함수 내부를 바꾸려면 FLM/BTM 이 먼저다.

### 직접 재현 — CRITICAL 둘을 이 세션이 손으로 다시 쟀다 (`rv_verify.py`)

| 경우 | 기록 전 | `--record-landing` | 기록 뒤 `check` | base..HEAD Go |
|---|---|---|---|---|
| V1 FLM 먼저(저장소 규칙의 순서) · 편집 뒤 번들 미갱신 | 빨강(stale hash) | rc=0, **편집 전 증거 커밋 X** 기록 | `[]` **required 0** | `internal/own.go` |
| V2 Go 작업 먼저 · 작업 이전에서 갈라진 곁가지에 base 상태 증거 · 병합 | 빨강 | rc=0, **S1**(W 의 자손 아님) 기록 | `[]` **required 0** | `internal/own.go` |

### 6.1.2 의 결론을 정정한다

"하한 = 증거가 역사에 들어온 지점, 저자는 그 아래를 못 고른다"는 **선형 역사 · 정규 파일 ·
병합 없음**에서만 참이다. 그리고 그 셋이 다 참이어도 **증거가 작업보다 먼저 들어오면 하한이
작업 앞에 온다** — 이 저장소가 요구하는 순서(FLM 을 편집 **전에**)가 정확히 그 순서다. 6.6 을
"잔여"로 적었지만 그것은 가장자리가 아니라 **기본 경로**다. 게다가 5단계는 그 빨강 옆에
`--record-landing` 을 권하고, 도구가 그 값을 계산해 초록으로 만든다(V1). 도구 이전에도 손으로
X 를 적으면 같은 결과였지만(4.4 위조 1번의 모양), 이제는 도구가 **권하고 계산한다**.

### P0 — 5단계가 undocumented Go 편집을 required 0 으로 통과시킨다

| # | 무엇 | 자리 | 원천 |
|---|---|---|---|
| C1 | FLM-first 번들을 편집 뒤 안 갱신 → 조언대로 `--record-landing` → 편집 전 커밋 기록 → required 0 | `check_analysis.py:1011`(`compute_landing`) · `:1114-1122`(조언 줄) | testing · security · Codex 적대 P1#1 · **V1** |
| C2 | 하한이 DAG 의 하한이 아니다 — 곁가지 증거 병합 · 병렬 가지(`-1` 이 저자가 정하는 committer date 로 고름) · 병합 안에서만 바뀐 증거(`-m` 없음) | `:479` · `:555` | security · Claude 적대 2c/2d · Codex 적대 P1#4 · **V2** |
| C3 | 하한은 **경로**를 보고 판정은 **워킹트리 내용**을 본다 — `ast.json` 심링크(링크 텍스트만 추적) · `{}` 자리표시 커밋 뒤 로컬 교체 · 대상 파일을 한 번도 커밋 안 해도 통과 | `:466-481` · `:750`(`_ast_value`) | Claude 적대 #1(CLI 끝까지 재현) · Codex 적대 P1#2 |
| C4 | 빌리는 쪽 창: 빌려주는 쪽 기록은 가장 낮은 값이고 빌리는 쪽은 "복사하라"로 거절 → 빌리는 쪽의 뒤 Go 작업이 창 밖 | `:1046-1050` · `:907-913` | Red Team(픽스처) · Claude 적대 #7 |

### P1

| # | 무엇 | 자리 | 원천 |
|---|---|---|---|
| H1 | `--diff-filter=MA` 가 T(typechange)·R<100(편집 섞인 rename)을 안 센다 — 아카이브 커밋이 `source_sha256` 을 바꿔 실어도 하한이 안 움직인다 | `:479` | security · Claude 적대 2a/2b · Codex 적대 P1#3 |
| H2 | 사용자 `log.follow=true` 가 경로 하나짜리 pathspec 에 `--follow` 를 조용히 켜서 하한이 기계마다 다르다 | `:479` | Red Team(픽스처) |
| H3 | "도구가 계산한 값"을 게이트가 확인하지 않는다 — **a099 실측: 기록 `e6c4636a` ≠ 계산 `21a315d1`**. 같게 강제하면 유일한 실물 기록이 깨진다 → 사람 결정 | `:561` | Claude 적대 #3 · 이 세션 실측 |
| H4 | `record_landing` 이 존재 확인 뒤 긴 계산을 하고 `write_text` 로 덮는다(비원자) · 끊긴 심링크를 따라 저장소 밖에 쓴다 | `:1044` · `:1074` | Claude 적대 #6(재현) · Codex 적대 P2#5 |
| H5 | `compute_landing` 이 `ValueError`(저장소 밖 `file`·NUL)를 내면 `record_landing` 이 traceback 으로 죽는다 | `:1071` | Claude 적대 #4(재현) |
| H6 | 새 `timeout=60` 호출의 `TimeoutExpired` 가 `check()` 의 except 를 빠져나가 창 줄이 안 찍힌다 | `:480` · `:934` | Claude 적대 #5(코드) |
| H7 | 병합을 마치며 처음 커밋한 번들은 하한이 안 잡혀 **정상 증거가 거절된다**(C2 의 반대 방향) | `:478-481` | Codex 구조 P2(픽스처) |

### P2 / 정보

| # | 무엇 | 원천 |
|---|---|---|
| I1 | 성능: 못 찾는 walk 가 (후보+1)×(번들+2) spawn — 아카이브 a055 **133.7s**, enable-engine-autostart-menu **219.6s**, a089 10.3s, a095 14.1s. 번들마다 `git show`(파일 단위 아님)·조기 종료 없음·후보마다 재-glob·후보마다 `is_ancestor`(`--ancestry-path` 로 대체 가능) | performance(실측) · simplification(권고) |
| I2 | `compute_landing` 이 `resolve_landing` 의 수락 조건을 **따로 한 벌** 가진다 — 판정 둘([[two-judgements-cover-for-each-other]]) | maintainability |
| I3 | `execution_baseline.validate` 의 지역 `canonical` 이 모듈 함수 `canonical()` 을 가린다(6.2 가 만든 것) | maintainability |
| I4 | `record_landing` 이 `check` 의 디렉터리 해소를 복사했고, 없는 id 에 "base 를 capture 하라"는 엉뚱한 조언을 한다 | maintainability |
| I5 | 이관 거절 문장이 두 벌이고 이미 갈렸다("ends" / "already ends") | maintainability |
| I6 | `_pre_archive_path` 가 아카이브 문법의 또 한 벌이다 — 6.2 가 `validate` 에 "셋째를 만들지 않는다"고 적은 바로 그 모양 | maintainability |
| I7 | `_change_analysis_path` 의 "outside current change analysis" 가 이제 적힌 자리를 보는데 "current" 라고 말한다 | maintainability |
| I8 | 기록이 디스크에 있는데 아직 커밋 전이거나 아카이브 이동이 staged 면 창 줄이 "working tree (no landed-commit.txt)" + `--record-landing` 을 권하고, 그 명령은 "already exists" 로 거절 — 서로 모순된 두 문장 사이를 돈다. 빌리는 쪽·번들 0 인 쪽에도 같은 조언 | Red Team · Claude 적대 #7 |
| I9 | 시험: `compute_landing` 의 다중 후보 walk 가 한 번도 안 돈다(변이 M1·M2·M6 생존) | testing(변이 실측) |
| I10 | 시험: `test_it_does_not_record_the_base…` 는 주장보다 약하다 — 위조 픽스처를 `record_landing` 에 준 적이 없다(M7 생존) | testing |
| I11 | 시험: `--diff-filter` 의 `M` 이 없어져도 모른다(M3 생존) | testing |
| I12 | 시험: 아카이브 시험이 착지가 살아남았는지 안 본다(M4 생존) | testing |
| I13 | 시험: `record_landing` 의 모호·빌림·이관 거절 갈래(M11·M8·M9 생존) | testing |
| I14 | 시험: `--record-landing` CLI 종료 코드(M5 생존) | testing |
| I15 | 시험: 음성 대조가 아무 오류나 받는다(`AST source hash is stale` 를 안 봄) | testing |
| I16 | 시험 결합: 다른 TestCase 를 만들어 사적 `_clean_fixture` 를 부른다 · 픽스처 다섯 벌 중복 | maintainability |
| I17 | 문서: `tools/logic-map/README.md:16` · `docs/WORKFLOW.md:174` 가 `check_analysis.py --change` 만 적고 `--record-landing`·`landed-commit.txt` 가 없다 | 문서 staleness |

부록(신뢰 4): 번들 디렉터리 이름의 glob 문자가 pathspec 을 넓힌다 — 하한을 올리기만 해서 위조 경로는 아니다.
`execution_baseline` 의 id 신원 · 적힌 자리/지금 자리 분리에는 원천 여덟 모두 결함을 못 찾았다.

### 계획 대조 (tasks.md §6)

| 항목 | 판정 |
|---|---|
| 6.1.2.1 하한을 선언의 하한으로 | **PARTIAL** — 섰지만 C1~C3·H1·H2 로 하한이 아니다 |
| 6.1.2.2 `--record-landing` | DONE — 그러나 C1 을 **돕는다** |
| 6.1.2.3 spec "도구가 계산한다(SHALL)" | **PARTIAL** — 게이트가 확인하지 않는다(H3) |
| 6.1.2.4 · 6.1.2.5 | DONE (문서·스크립트, 코드 diff 밖) |
| 6.2 · 6.2.1 · 6.3 | DONE — 결함 없음, 다만 I3·I7 과 6.3 범위 밖 생존 변이(I9~I14) |

PR Quality Score(공식 `10 − 2×critical − 0.5×informational`, critical 7 = C1~C4·H1·H2·H4,
informational 22): **0/10**. 4.5 는 계속 막힌다.

## VERIFY — task 7.1 (조언 줄이 stale 증거를 세탁하지 않는다)

사람 결정 2026-09-12: §6 리뷰가 낸 B 를 **좁힌 판본**으로 간다 — stale 번들 중
**base 의 소스를 적은 것**이 있을 때만 `--record-landing` 조언을 막는다.

### 내가 D1 에 적은 전제가 틀렸다 (정정)

D1 에서 나는 B 가 "정상 입력을 하나도 거절하지 않는다 — 찍히는 조언 문구만 바꾼다"고
적었다. **틀렸다.** 판정(rc)은 그대로지만 **조언은 사라진다.** 활성 전수로 재 보니
11건이 워킹트리 모드에서 stale 로 보이고(`71_census.py`: `LOSE_HINT: 11`), 그중 9건의
stale 번들이 base 상태를 기술한다(`71_baseshape.py`:
`CHANGES_WITH_BASE_SHAPED_STALE: 9 of 11`). 나머지 둘(a066 · a071)만 base **뒤** 상태를
적었으므로 착지를 기록하면 창이 실제로 좁아진다. 이 실측이 옵션 1 을 고르게 한 근거이고,
사용자에게 결정을 다시 물은 이유다. [[fail-closed-must-name-what-it-rejects]] 가 말하는
"거절할 정상 입력을 먼저 열거하라"를 **내가 D1 에서 안 했다.**

### 순서

`ast.before-7.1.json`(`main` · `check` · `validate_target`, 커밋 `60150803` 시점)을
**먼저** 열거했다. 편집 뒤 `ast.after-7.1.json` 을 같은 `enumerate.py` 로 뽑았고,
FLM·BTM 의 분기 표는 두 열거를 스크립트가 대조해 만들었다 — 손으로 옮겨 적지 않았다.
`validate_target` 은 열거만 하고 **바꾸지 않았다**(stale 문구가 어느 자리에서 나오는지
확인하는 데 썼다: `validate_target` 둘 · `check` 하나).

### RED → GREEN

| 시험 | 구현 전 | 구현 후 |
|---|---|---|
| `test_a_bundle_that_records_the_base_is_not_told_to_record_a_landing` | **FAIL** | pass |
| `test_a_stale_bundle_that_does_not_record_the_base_still_hears_the_advice` | pass | pass |
| `test_a_fresh_bundle_for_a_file_unchanged_since_the_base_keeps_the_advice` | (구현 뒤 추가) | pass |
| `test_only_three_bundles_are_named_and_the_rest_are_counted` | (구현 뒤 추가) | pass |

둘째 시험은 **처음부터 초록이었다. RED 가 아니다** — 과잉 억제를 막는 핀이고, "stale
이면 전부 막는다"(변이 M1)를 넣어야 빨개진다. 통과했다는 사실 자체는 증거가 아니므로
그렇게 적는다 [[passing-test-is-not-evidence]].

### 무엇을 바꿨나

- 새 `_base_shaped_bundles(root, base, analysis)` — `revision: current` 인데 **base 의
  소스**를 적은 번들의 이름들. 신선한 번들은 먼저 건너뛰고, 남은 것만 base 의 blob 과
  대조한다. 판정은 신원이 아니라 **blob 등식**이다(1.12 가 신원 판정을 배제했다).
- `check` — `if not landing:` 일 때만 그 값을 `facts` 에 넣는다. 반환은 14개 그대로다.
- `main` — 창의 크기(**사실**)와 조언을 갈랐다. 두 갈래가 같은 `window` 문자열을
  공유하므로 3.3 이 세운 "실패해도 창을 찍는다"는 양쪽에서 글자 그대로 남는다.

`main` 에서 change 디렉터리를 다시 해소하지 않은 이유: 그러면 해소가 **세 벌**이 된다
(7.6 이 이미 두 벌을 결함으로 적었다). `check` 는 `change_dir`·`analysis` 를 이미 쥐고 있다.

### 게이트 판정은 그대로다 — 구조가 아니라 실측으로

구조 근거는 `check` 의 반환 14개가 편집 전후로 동일하다는 것이다. 그것만으로는 주장이라
HEAD 의 `check_analysis` 사본과 편집본을 **활성 change 전부**에 돌려 오류 목록·착지·
요구 수를 글자 단위로 대조했다(`71_rc_ab.py`).

```
IDENTICAL: 27  DIFFERENT: 0
```

사본은 `tools/logic-map/` 안에 두고(ROOT 가 `__file__` 에서 유도되므로 스크래치패드에
두면 다른 저장소를 본다) `finally` 로 지웠다 — `probe removed: True`.

### 조언이 실제로 어떻게 갈리나 (구현 뒤 활성 전수, `71_verify.py`)

```
SUPPRESSED: 9 KEPT_ADVICE: 15
```

억제된 아홉: a074 · a077 · a079 · a089 · a091 · a092 · a094 · a095 · a100.
a066 · a071 은 `base_shaped=0` 이라 조언을 **그대로 받는다** — 구현 전 예측
(9 of 11)과 실측이 정확히 일치한다. a092 는 base 를 기술하는 번들이 **29개**라
이름을 자르고 세는 갈래가 실물 경로다. a122 자신도 조언을 받는다(자기 검사 rc=0).

### 변이 아홉 — 전부 CAUGHT (`71_mutations.py`, 원복 sha256 동일)

원복은 `git checkout` 이 아니라 **저장한 바이트**로 했다. 구현이 아직 커밋 전이라
`git checkout` 은 GREEN 까지 지운다 [[mutation-revert-needs-the-right-baseline]].

| 변이 | 빨개진 시험 수 |
|---|---|
| M1 base 모양 판정 제거 (stale 이면 전부 억제) | 1 |
| M2 억제를 아예 안 함 | 2 |
| M3 신선 번들 건너뛰기(`continue`) 제거 | 1 |
| M4 base 가 아니라 HEAD 와 대조 | 2 |
| M5 어느 번들인지 안 말함 | 2 |
| M6 억제 갈래에서 창의 크기(사실)를 뺌 | 2 |
| M7 `check` 가 재지 않는다 | 2 |
| M8 이름을 자르지 않고 다 쏟아냄 | 1 |
| M9 못 적은 수를 안 셈 | 1 |

**M5 는 처음에 SURVIVED 였고 원인은 내 바늘이었다.** `assertIn("internal--own", printed)`
가 출력 **전체**를 보는데 같은 이름이 오류 줄
(`[logic-map] internal--own: AST source hash is stale`)에도 있어서, 조언 줄이 이름을
잃어도 통과했다. 6.3 이 `resolve_landing` 에서 고친 것과 **같은 결함**이
내가 이번에 새로 쓴 시험에서 재발했다 [[existence-check-is-not-a-role-check]].
`_advice_line` 로 조언 줄 하나를 집어내 바늘을 좁힌 뒤 CAUGHT 가 됐다.

### 거부할 정상 입력 (먼저 열거했다)

- **base 이후 안 바뀐 파일의 신선한 번들.** `sha256(오늘) == sha256(base)` 라 등식만
  보면 V1 과 똑같이 생겼다. 신선한 번들을 **먼저 건너뛰는** `continue` 가 그것을 가르고,
  `test_a_fresh_bundle_for_a_file_unchanged_since_the_base_keeps_the_advice` 가 그 자리를
  못 박는다(변이 M3).
- **a066 · a071 처럼 base 뒤 상태를 적은 stale 번들.** 조언을 그대로 받는다(변이 M1).

### 안 한 것

- **작업 중간 상태를 적은 FLM 은 안 잡힌다.** 번들이 base 도 오늘도 아닌 지점을
  기술하면 이 판정은 조용하다. blob 등식으로는 "V1" 과 "base 가 작업 뒤에 놓인 change"
  를 가를 수 없고, 그 차이는 **누구의 편집이었나** 하나인데 신원 판정은 1.12 가
  배제했다. 그것이 **7.2** 다 — 착지를 무엇에 묶을지 다시 정하는 일.
- **명령 자체는 그대로다.** `--record-landing` 은 여전히 편집 전 커밋을 계산할 수 있다.
  7.1 은 게이트가 그것을 **권하지 않게** 했을 뿐이고, 거절하지 않는다. C1 은 7.2 전까지
  닫히지 않는다. **4.5 는 계속 막힌다.**
- `validate_target`·`check` 의 stale 문구 자체는 안 건드렸다.

### 게이트 (2026-09-12, 이 로트의 워킹트리)

| 명령 | 결과 |
|---|---|
| `python3 -m unittest discover -s tools/logic-map` | **95 tests OK** (skipped 1) |
| `make lint` | rc=0 |
| `make test` | rc=0 |
| `make test-seams` | rc=0 |
| `make sdd-test` | rc=0 |
| `openspec validate … --strict` | valid |
| `check_analysis.py --change a122…` | rc=0 |
| `make sdd-sync` | **rc=2** — advisory 만 실패 |
| `make sdd-check` | rc=0 |

`sdd-sync` 의 rc=2 는 `codegraphcontext` 레그다:
`Could not set lock on file /home/daniel/.codegraphcontext/global/db/kuzudb`.
락을 쥔 것은 pid **47230** (`cgc`, 2026-09-11 23:16 기동) — **다른 세션의 프로세스**라
죽이지 않았다. 타임아웃이 아니므로 재시도는 소용없다(2026-09-11 과 같은 원인).
hard CodeGraph 레그(`codegraph sync .`)는 돌았고 `sdd-check` 가 rc=0 이므로 fingerprint 는
stale 이 아니다. CodeGraphContext 는 advisory 이며 현재 HEAD·테스트를 대체하지 않는다.

## MEASURE — task 7.2 후보 실측 (2026-09-12, 결정 전)

**이것은 결정이 아니라 측정이다.** tasks.md 7.2 가 "규칙을 지어내지 않는다 — 각 후보가
활성 13건 · a099 · 아카이브에서 무엇을 거부하는지 잰 뒤 정한다"고 적었으므로, 후보를
전수에 대고 재기만 했다. 생산 코드는 한 줄도 안 바꿨다.

### 계측기와 그 대조

`72_probe.py` 는 생산 `_pinning_bundles` · `_pre_archive_path` · `resolve_base` 를 그대로
부르고, 느린 자리(후보마다 번들마다 `git show`)만 `git cat-file --batch` 한 프로세스로
바꾼다. **계측기를 먼저 검증했다**: 번들이 있는 활성 13건 + a099 에 대해 생산
`compute_landing` 과 대조 → `AGREE 13 DIFFER 0`(a089 는 11.0s → 0.1s, 답은 같다).
양성 대조군으로 digest 한 글자를 망가뜨리면 빠른 경로는 **반드시** 착지를 잃어야 하고,
전 건에서 그랬다.

측정 대상: 번들을 가진 change **94건**(활성 13 · 아카이브 81). a063 은 base 를 못 정한다
(execution-baseline adoption 은 detached HEAD 를 요구) → 측정 93건. 그중 **오늘 착지를
얻는 것이 76건**이고, 17건은 오늘도 못 얻는다(고정을 통과하는 커밋이 없다).

### 무엇을 거부하는가 (오늘 착지를 얻는 76건 기준)

| 후보 | 착지 잃음 | 착지 이동 | 비고 |
|---|---|---|---|
| K1 번들 blob 이 HEAD 와 같다(정규 파일) | **0** | 0 | 공짜 |
| K1′ 번들 blob 이 **워킹트리**와 같다 | **0** | 0 | 공짜 |
| K2 고정 소스가 base..착지에서 바뀌었다(하나라도) | **8** | 0 | a074 · a077 · a079 · a091 · a092 · a094 · a078 · **a075** |
| K2-all 전부 바뀌었다 | **25** | 0 | a071 · a100 · a112 포함 |
| K3 first-parent 에 묶는다 | **26** | 19 | 바닥이 49건에서 달라진다 |
| K4 git 설정을 지운 호출 | **0** | 0 | 이 기계에서 차이 0 |
| H1 `--diff-filter=MAT` + `-M100%` | **0** | 0 | 바닥이 1건 이동 |
| H2 `log.follow=true` 인 기계 | 0 | 0 | 이 역사에서 차이 0 |
| K6 번들이 **HEAD** 의 소스를 기술한다(전부) | **69** | — | 후보 아님이 실측으로 확정 |
| K6′ 하나라도 HEAD 를 기술한다 | **17** | — | 〃 |
| 착지를 change id 를 부르는 마지막 커밋에 묶는다 | 13 적용불가 | 41 | 태그 0 인 change 13건 |

첫 판의 K1 은 47건을 거절했는데 **계측기 결함**이었다 — 아카이브는 번들을 통째로 옮기므로
착지 시점에는 오늘의 경로가 없다. `_pre_archive_path` 를 같이 주자 0 이 됐다. 고친 판의
첫 실행은 번들 0 개를 세어 `all()` 이 **표본 0으로 참**이 됐고(전칭 판정의 함정), 경로를
고치고 `assert n` 을 세운 뒤 다시 쟀다.

### 구멍을 닫는가 (`72_holes.py`, 임시 저장소 넷)

| 경우 | 오늘 | K1 | K1′ | K2 | K3 |
|---|---|---|---|---|---|
| V1 FLM-first 번들이 stale | 받는다 | 받는다 | 받는다 | **거절** | 받는다 |
| V2 곁가지 증거 병합 | 받는다 | 받는다 | 받는다 | **거절** | **거절** |
| C3 자리표시 커밋 + 로컬 교체 | 받는다 | 받는다 | **거절** | 거절 | 받는다 |
| C3 심링크 번들 | 받는다 | **거절** | **거절** | 거절 | 받는다 |

C3 픽스처는 처음에 `_commit_all` 이 워킹트리 번들까지 커밋해 버려 구멍이 사라졌다
(오늘도 "거절"로 나왔다). Go 편집만 커밋하도록 고친 뒤에야 대상에 닿았다.

### 실측이 말하는 것 — 후보들이 갈리지 않는 이유

**V1 의 세탁과 a074 의 정당한 빈 창은 모든 내용 축에서 같은 모양이다.** 둘 다 "번들이
base..착지에서 바뀌지 않은 상태를 기술한다"이다. 다른 것은 그 change 의 Go 작업이 base
**앞**에 착지했는가(a074 — 2026-08-04 재기준화) 아니면 착지 **뒤**에 있는가(V1)뿐이고,
증거 내용은 그 둘을 구분하지 못한다. 그래서 K2 가 V1 을 닫으면 a074 · a075 · a077 ·
a079 · a078 — **a122 가 존재하는 이유인 바로 그 change 들** — 이 착지를 잃고 넓은 창으로
돌아간다. "착지 > base 일 때만 적용"으로 좁혀 봤지만 8건 전부 착지 ≠ base 라 그대로다.

같은 이유로 "번들이 HEAD 를 기술해야 한다"(K6)도 못 쓴다. 이웃이 나중에 같은 파일을
고치면 번들은 stale 로 보이고, 그 stale 을 **착지가 정당하게 침묵시키는 것**이 a122 의
기능 자체다. V1 의 stale 은 저자 자신의 편집이 만든 것인데, 내용만으로는 이웃의 편집과
구분되지 않는다.

### 그래서 지금 살아남은 것

**공짜인 셋** — K1′(워킹트리 등식 + 정규 파일 모드) · K4(설정을 지운 호출) · H1(`MAT`,
`-M100%`). 거부하는 정상 입력 0건이고 각각 다른 구멍(C3 · H2 · H1)을 닫는다.
**못 쓰는 넷** — K2 · K2-all · K3 · K6 은 실측 거부가 8~69건이다.
**C1 과 C2 는 이 후보들로 닫히지 않는다.** 착지를 증거 내용에 묶는 축이 소진됐다는 것이
이 측정의 결론이고, 다음 규칙은 사람이 고른다.

## DECIDE — task 7.2 규칙 선택 (2026-09-12)

사람이 측정을 읽고 **공짜인 셋만 지금 취한다**를 골랐다. 위 `## MEASURE` 가 낸 후보 여섯 중
거부 0 으로 실측된 셋(K1′ · K4 · H1)이 규칙이 되고, K2 · K2-all · K3 · K6 은 채택하지 않는다.
**C1 과 C2 는 열린 채로 남는다** — 6.6 · 6.1.2.6 · 4.5 는 계속 막힌다.

### 규칙이 된 것 셋

| 규칙 | 무엇을 닫나 | 어디에 사나 |
|---|---|---|
| 판정이 **읽은** 번들을 착지 커밋이 들고 있어야 한다 | C3 (자리표시자 교체 · 심링크 · 미커밋) | `_unheld_bundles` — `resolve_landing` 과 `compute_landing` 둘 다 |
| 옮기면서 **고친** 증거는 하한을 올린다 (`-M100%` · `MAT`) | H1 (`R<100` · `T` 를 안 세던 하한) | `_evidence_floor` |
| 하한은 개인 git 설정의 함수가 아니다 (`--no-follow`) | H2 (`log.follow=true` 인 기계) | `_evidence_floor` |

첫째가 왜 **워킹트리와** 등식을 세우는가: 판정(`validate_target` · `_pinning_bundles`)이 읽는
것이 워킹트리이기 때문이다. HEAD 와 세우면 자리표시자를 커밋해 두고 워킹트리에서 갈아 끼운
자리가 안 보인다 — 뮤테이션 M-F 로 실측했다(HEAD 로 바꾸면 시험 셋이 초록으로 돌아간다).

### 측정이 구현을 세 자리에서 고쳤다

1. **범위는 `ast.json` 하나다.** 번들의 산문 파일까지 넓혀서 다시 쟀더니(번들 파일 9,092개)
   76건 중 **3건이 착지를 잃는다** — a092 · a043(2건) · a096 이고, 셋 다 **이웃 change 가
   나중에 단 무효화 배너**다(a099 가 a092 의 지도에 "이 문서의 절반이 무효" 주석을 달았다).
   정상 입력이므로 범위를 안 넓혔다. `ast.json` 만으로는 거부 0 이다.
2. **K1′ 의 "정규 파일 모드" 절은 안 넣었다.** 등식을 워킹트리와 세우면 도달할 수 없는 절이다
   — 심링크의 blob 은 **가리키는 경로 문자열**이라 등식에서 이미 갈리고, 그 문자열이 디스크
   내용과 같아지려면 대상 파일이 자기 경로를 담아야 하는데 그러면 JSON 파싱이 실패해서
   애초에 고정 번들이 아니다. 모드 절은 K1_head(HEAD 와 비교하는 판본)에서만 일하던 절이다.
3. **K4 의 "설정을 지운 호출"은 `--no-follow` 한 토큰으로 줄었다.** `-c log.follow=false` 는
   같은 일을 하고(둘 다 두면 서로를 가려 뮤테이션이 살아남는다), `-c diff.renames=true` 는
   `-M100%` 이 명령줄에서 이미 덮는다 — 지워도 죽는 시험이 없었다(M-C3 SURVIVED). 실측
   기준선에는 이 설정이 하나도 없으므로(local · global · system 전부 비었다) 답은 안 바뀐다.

### 거절 **지점**의 순서

등식은 하한 **뒤**에 세웠다. 앞에 두면 "증거가 역사에 없다"(M2)와 "착지가 증거보다
앞선다"(6.1.2.1)를 재던 기존 시험 둘의 거절 지점을 가로채서, 그 가드를 지워도 스위트가
초록으로 남는다. 처음 쓴 판본이 정확히 그랬고 기존 시험 둘이 빨개져서 잡혔다
([[first-failure-is-not-the-fix-scope]] · [[two-judgements-cover-for-each-other]]).

같은 이유로 H1 시험은 `check` 가 아니라 `_evidence_floor` 를 직접 부른다. 종단에서는 같은
입력을 C3 등식이 **먼저** 거절해서, 그 위에서 재면 하한이 아니라 다른 가드를 재게 된다.

### 뮤테이션 — 일곱 전부 CAUGHT

| 변이 | 결과 | 빨개진 시험 |
|---|---|---|
| M-A `-M100%` → `-M` | CAUGHT | `…archive_that_also_rewrites_the_bundle_moves_the_floor` |
| M-B `MAT` → `MA` | CAUGHT | `…turning_a_bundle_into_a_symlink_moves_the_floor` |
| M-C `--no-follow` 삭제 | CAUGHT | `…refused_on_a_log_follow_machine` · `…same_verdict_either_way` |
| M-D `resolve` 의 등식 삭제 | CAUGHT | `…symlinked_out_of_the_watched_path…` · `…placeholder_committed_early…` |
| M-E `compute` 의 등식 삭제 | CAUGHT | `…will_not_compute_a_landing_it_cannot_hold` |
| M-F 워킹트리 → HEAD | CAUGHT | 위 셋 전부 |
| M-G 아카이브 전 경로 fallback 삭제 | CAUGHT | `…archiving_the_change_does_not_invalidate_its_record` |

M-C 는 첫 판에서 SURVIVED 였다 — `--no-follow` 와 `-c log.follow=false` 를 둘 다 두어서
서로를 가렸다. M-E 도 첫 판에서 SURVIVED 였다(계산 경로에 닿는 시험이 0). 둘 다 규칙을
고친 것이 아니라 **중복과 빈칸**을 없앤 것이다.

### 실측 — 구현이 실저장소에서 무엇을 바꾸는가

**착지 전수 (93건).** 생산 `_evidence_floor` 와 `_unheld_bundles` 를 **그대로 부르고** 느린
자리만 배치로 바꿔 다시 셌다: **같음 76 · 이동 0 · 착지 상실 0 · 원래부터 없음 17**. 세 규칙을
다 넣어도 이 저장소의 착지는 하나도 안 움직인다. 후보를 따로따로 잰 값(각 0)과 조합의 값이
같다는 것을 따로 확인한 것이다.

계측기가 눈먼 경우를 가르는 **양성 대조군**: 같은 함수를 저장소의 **첫 커밋**에 대면 12건
전부가 "안 들고 있다"를 내고, 각자의 착지에 대면 12건 전부가 "들고 있다"를 낸다(12/12).

**판정 A/B (28건).** HEAD 판본과 편집 판본을 같은 입력에 대서 오류 목록 · 착지 · required 를
비교했다 — **IDENTICAL 28 · DIFFERENT 0**. 착지 기록을 가진 유일한 change a099 는 착지
`e6c4636a` 와 required 32 를 그대로 유지하고, 번들 32개를 추가로 대조하는데도 느려지지
않았다(8.5s → 8.0s).

`make sdd-test` 통과, `tools/logic-map` 스위트 105 + 형제 75 전부 초록,
`openspec validate --strict` valid, a122 자신의 `check_analysis` rc=0.

### 남은 것 (이 결정이 안 닫은 것)

- **C1** FLM-first stale 번들 → 도구가 편집 전 커밋을 계산한다. 등식은 "커밋된 것과 읽은 것이
  같은가"만 묻지 "그 번들이 **오늘의 소스**를 적는가"는 안 묻는다. 후자가 K2 이고 8건을 죽인다.
- **C2** 바닥이 DAG 의 바닥이 아니다(곁가지 병합 · 병렬 가지에서 `-1` 이 committer date 로 고른다).
- **C4** 빌리는 쪽이 빌려주는 쪽의 가장 낮은 값을 복사한다.
- **H7** 병합 커밋에서 처음 들어온 정상 증거가 거절된다.

넷 다 증거 **내용**으로는 안 갈린다. **6.6 · 6.1.2.6 은 이 결정 전에 닫을 수 없고 4.5 도 계속
막힌다.**

## MEASURE — task 7.3 (H3): 계산값과 다른 기록이 무엇을 열어 두는가 (2026-09-13, 결정 전)

spec 은 "착지 지점의 기록은 저자가 선언하는 값이 아니라 **도구가 그 change 의 증거로
계산해서 기록하는 값이어야 한다**(SHALL)"고 적는데, 게이트는 기록이 유효한지만 보고
**그 값인지는 안 본다**. 규칙을 지어내기 전에 무엇이 열려 있고 닫으면 무엇이 거부되는지
먼저 쟀다 (계측기 `73_choice.py` · `73_accept.py` · `73_monotone.py` · `73_sample.py`).

### 1. 남은 선택의 크기 — 7.2.1 뒤에도 크다

오늘의 가드(고정 해시 · 하한 · 판정이 읽은 번들 등식)를 **전부** 통과하는 커밋을
활성·아카이브 전수로 셌다.

| | 건수 |
|---|---|
| 번들을 가진 change | 94 (측정 93 — a063 은 이관 검증이 detached HEAD 를 요구해 이 실행에서 base 가 안 풀린다) |
| 오늘 착지를 얻는 change | 76 |
| 수락값이 **둘 이상**인 change | **65** (최대 516) |
| 그 선택이 **판정 입력을 바꾸는** change | **44** (최대 판정 279가지) |

판정 입력의 동일성은 정확한 충분조건으로 갈랐다 — (base..후보에서 바뀐 Go 파일 집합,
그 파일들의 후보 시점 blob oid)가 같으면 `changed_existing_functions` 의 입력이 바이트까지
같다. 즉 "선택이 답을 바꾼다"는 추정이 아니라 측정이다.

### 2. 그런데 그 선택은 **한 방향으로만** 열려 있다

- 수락값이 계산값의 **자손이 아닌** change: **0 / 76** (`73_accept.py`).
  손으로 쓸 수 있는 값은 전부 계산값보다 **뒤**다.
- 계산값보다 창을 넓혔을 때 바뀐 Go 파일이 **빠지는** 자리: **0 / 44** (`73_monotone.py`,
  전 후보 전수).
- 함수 단위 표본 6건(`73_sample.py`, 진짜 `changed_existing_functions`):
  빠진 함수 **0**, 늘어난 함수 0~15.

| change | 계산값 required | 최고값 required | 빠진 함수 |
|---|---|---|---|
| a056(아카이브) | 3 | 3 | 0 |
| a059(아카이브) | 7 | 7 | 0 |
| a092 | 0 | 2 | 0 |
| a108(아카이브) | 11 | 26 | 0 |
| verify-reopens-conditional-chain | 5 | 18 | 0 |
| a097(아카이브) | 5 | 7 | 0 |

**그래서 H3 이 남긴 자유는 오늘 이 역사에서 자해뿐이다.** 저자가 계산값에서 벗어나면
창이 넓어지고 요구가 늘어난다. 구조적으로 배제된 것은 아니다 — 같은 파일 안에서 편집이
되돌려지면 넓힌 창이 오히려 느슨해질 수 있고, 그 모양이 오늘 이 저장소에 0건일 뿐이다.

### 3. 등식을 넣으면 무엇이 거부되나

오늘 존재하는 실물 기록은 **하나**다 — a099(`e6c4636a`, 2026-09-09 커밋 `5c848099`,
`compute_landing` 이 생기기 **이틀 전**에 손으로 쓴 값).

- 계산값 `21a315d1` · 기록 `e6c4636a` — 기록은 수락값 21개 중 **index 1**(계산값 바로 위).
- 계산값으로 재기록하면: `check` 오류 **0건 · required 32** — 지금 기록과 **같은 판정**이다.
- 판정 비용: a099 의 `check()` 15.85s 에 `compute_landing` **1.33s** 가 더해진다(+8%).
  등식을 기존 가드 **뒤**에 두면 walk 는 기록 지점에서 멈추므로 I1 이 잰 133.7s·219.6s
  (아무 후보도 안 맞아 끝까지 도는 walk)는 이 경로에 안 온다.

### 4. 규칙 후보와 각각이 거부하는 것

| 후보 | 거부하는 것 | 여는 것 |
|---|---|---|
| R1 기록 == 계산값 | a099 의 현재 값 하나(재기록하면 판정 동일) | 정직하게 창을 **넓히는** 기록도 같이 막힌다 |
| R2 기록 ≥ 계산값(조상 판정) | **0건** — 실측으로 no-op | 아무것도 안 막는다 |
| R3 안 넣는다 | 0건 | 65건의 선택·44건의 판정 차이가 그대로 남는다 |
| R4 계산값과 다르면 review.md 에 값과 근거 | 근거 없는 이탈 | 새 기제 하나(7.2 가 가리킨 축) |

R2 는 측정으로 죽었다 — 수락값이 전부 계산값의 자손이므로 조상 판정은 아무것도 안 거른다.
**남은 것은 R1 · R3 · R4 이고 그 선택은 사람 몫이다.**

## VERIFY — task 7.4 (기록은 한 번만 쓰이고, 결함은 판정이 된다) (2026-09-13)

### 순서

RED 여섯 → GREEN → 뮤테이션 → 실측 A/B. 리뷰가 이름 붙인 셋(H4·H5·H6)을 그대로 열고,
각각이 **실제로 무엇을 하는지** 먼저 재현한 뒤에 고쳤다.

### H4 는 실측으로 재현했다 — 게이트가 저장소 **밖**에 썼다

`landed-commit.txt` 자리에 저장소 밖을 가리키는 **끊긴 심링크**를 두고
`record_landing` 을 부르면:

```
rc = 0
저장소 밖 파일이 생겼나: True
내용: 7bc48ebc51c4bf429487…
```

`exists()` 는 링크를 **따라가서** 답하므로 끊긴 링크에서는 거짓이고, 존재 확인을 그대로
통과한 뒤 `write_text` 가 링크를 따라간다. 도구는 rc 0 으로 "recorded" 라고 말하고, 그
값은 어느 커밋에도 없다.

### 무엇을 바꿨나

| 자리 | 전 | 후 |
|---|---|---|
| `record_landing` 존재 확인 | `exists()` 하나 | 심링크를 **먼저** 이름으로 거절 |
| `record_landing` 쓰기 | `write_text` | `open(…, "xb")` 배타 생성 + `FileExistsError` 거절 |
| `record_landing` 계산 | 맨몸 호출 | `except GATE_FAULTS` → 사유를 이름으로 |
| `main` | `check`·`record_landing` 맨몸 호출 | 경계 하나에 `except GATE_FAULTS` |
| 예외 목록 | 자리마다 넷을 베낌 | 새 `GATE_FAULTS` **한 곳** + `subprocess.SubprocessError` |

배타 생성이 존재 확인과 쓰기 **사이의 창**도 같이 닫는다 — 그 사이에는 walk 하나가
통째로 들어간다(리뷰 I1 실측 133.7s·219.6s). 보장의 자리는 확인이 아니라 **쓰기**다.

가드를 자리마다가 아니라 **경계**에 세운 이유: 이 도구를 부르는 생산 자리는
`tools/gate.sh:321` 의 CLI 하나뿐이고, 자리마다 목록을 베끼면 새로 부르는 git 하나가 어느
목록에도 안 걸려 다시 스택이 된다([[two-judgements-cover-for-each-other]] 의 반대 방향 —
판정이 둘이 아니라 **목록이 여럿**인 모양). `TimeoutExpired` 가 `OSError` 가 **아니라는**
것이 옛 목록 넷이 통째로 새던 이유다.

창 줄은 계속 찍힌다: `context` 는 참조로 채워지므로 착지를 **잰 뒤에** 터진 결함이면
`base … → working tree … required N` 이 그대로 남는다(시험이 단언한다).

### 뮤테이션 — 일곱 전부 CAUGHT (`74_mut.py`, 원복 sha256 동일)

| 변이 | 판정 | 죽인 시험 |
|---|---|---|
| M-A 심링크 거절 삭제 | CAUGHT¹ | `…symlink_record_is_refused_not_followed` |
| M-B 배타 생성 → `write_text` | CAUGHT | `…appears_while_computing_is_not_overwritten` |
| M-C `compute_landing` 가드 삭제 | CAUGHT | `…names_a_bundle_that_escapes_the_repository` |
| M-D `check` 경계 가드 삭제 | CAUGHT | 타임아웃·창 줄 시험 둘 |
| M-E `record_landing` 경계 가드 삭제 | CAUGHT | `…a_fault_the_record_path_cannot_answer` |
| M-F `GATE_FAULTS` 에서 `SubprocessError` 제거 | CAUGHT | 타임아웃 시험 둘 |
| M-G `GATE_FAULTS` 에서 `ValueError` 제거 | CAUGHT | 저장소 밖 번들 시험 둘 |

¹ **처음엔 SURVIVED.** 배타 생성이 심링크도 막으므로 "rc 1 · 밖에 안 씀"만 재는 시험은
링크 거절을 지워도 초록이었다([[surviving-mutant-may-mean-accidental-safety]]).
거절 **지점**을 단언하게 고쳐서 잡았다 — 6.3 이 이 파일에서 배운 것과 같은 수리다.

### 실측 — 판정은 안 바뀐다

- 판정 A/B(HEAD 사본 대 편집본, 활성 27 + 착지 기록 있는 아카이브 1):
  **IDENTICAL 28 · DIFFERENT 0**. a099 는 착지 `e6c4636a` · required 32 · 9.7s→9.0s.
- 스위트 112개 초록(추가 6) · `tools/logic-map` 전체 187개 · `make sdd-test` 15+187+71+22+16
  전부 초록 · `make lint` rc=0 · `openspec validate --strict` valid.

### 안 한 것 (침묵한 생략이 아니다)

같은 모양의 맨몸 호출이 `check` 안에 셋 더 있다 — `review.read_text`(`:963`) ·
`reference_file.read_text`(`:978`) · `test_index`(`:1045`). 셋 다 a122 **이전부터** 있던
자리이고, 이제 `main` 의 경계가 그 셋의 결함도 판정으로 바꾼다. 자리별 메시지는 안 붙였다 —
붙이려면 자리마다 시험이 필요하고, 그 시험이 재는 것은 이 change 의 범위 밖이다.

## DECIDE — task 7.3 규칙 선택 (2026-09-13)

사람이 **R1** 을 골랐다: 게이트가 기록을 계산값과 대조하고, 유일한 실물 기록 a099 를
계산값으로 재기록한다. 근거는 위 `## MEASURE — task 7.3` 의 표다.

### 규칙이 된 것

`resolve_landing` 의 **맨 뒤**에서 `compute_landing` 과 대조한다. 다르면 **두 값을 다**
이름으로 말한다 — 하나만 말하면 저자는 무엇을 적어야 하는지 모른 채 빨간 게이트만 본다.

자리가 맨 뒤인 것은 규칙의 일부다. 앞에 두면 위 가드들("증거가 역사에 없다" ·
"착지가 증거보다 앞선다" · "판정이 읽은 번들을 안 들고 있다")의 **거절 지점**을 이 등식이
가로채서, 그 가드를 지워도 스위트가 초록으로 남는다 — 7.2.1 이 같은 파일에서 실측한
모양이다([[a-new-guard-unpins-the-guards-behind-it]]).

### a099 재기록 — 판정은 안 바뀐다

| | 값 | 판정 |
|---|---|---|
| 옛 기록 (2026-09-09 `5c848099`, 손으로) | `e6c4636a` | 오류 0 · required 32 |
| 새 기록 (계산값) | `21a315d1` | 오류 0 · required 32 |

옛 값은 수락값 21개 중 **두 번째**였다 — 창이 한 커밋 더 넓었을 뿐 답을 바꾸지 않았다.
도구가 착지를 계산하는 `--record-landing` 은 그 기록 **이틀 뒤**(6.1.2)에 생겼으므로,
이 값은 규칙이 생기기 전의 손 기록이다. 정정 사유는 a099 의 아카이브 review.md 에도
한 문단 남겼다 — a099 만 읽는 사람이 값이 왜 바뀌었는지 알아야 한다.

### 뮤테이션 — 넷 전부 CAUGHT (`73_mut.py`, 원복 sha256 동일)

| 변이 | 판정 | 죽인 시험 |
|---|---|---|
| N-A 등식 삭제 | CAUGHT | `test_a_later_commit_that_also_matches_is_refused` |
| N-B 등식 뒤집기 | CAUGHT | 위 + `test_a_shared_landing_is_accepted` 외 1 |
| N-C 계산값 대신 **하한**과 비교 | CAUGHT¹ | `…first_matching_commit_not_where_evidence_entered` |
| N-D 계산값의 **자손이면** 수락(R2) | CAUGHT | `test_a_later_commit_that_also_matches_is_refused` |

¹ **처음엔 SURVIVED.** 저장소의 픽스처가 전부 하한 == 계산값이라 스위트가 규칙의
**정체**를 못 갈랐다 — "가장 낮은 수락 커밋"과 "증거가 들어온 커밋"이 같은 값이면 어느
쪽과 비교하든 초록이다([[falsification-must-vary-the-right-axis]]). 둘이 갈리는 모양은
흔한 커밋 순서다: 증거를 먼저 올리고 코드를 그 뒤에 올리면 하한에서는 아직 소스가
안 맞는다. 그 픽스처를 만들어 잡았다.

### 이 결정이 뒤집는 것 (명시한다)

task 1.3 은 "느슨해지는 위조 다섯을 막고 **엄해지는 방향은 일부러 안 막는다**"였다.
R1 은 그 방향도 막는다 — 정직하게 창을 넓힌 기록도 계산값이 아니면 거절된다. 그 대가를
받아들이는 근거는 두 가지다. (1) 넓힌 창이 요구하는 함수는 그 change 의 증거가 안 덮는
남의 함수이므로 게이트가 그것으로 얻는 것이 없다. (2) 넓힘이 **항상** 엄해지는 것은
오늘 이 역사의 성질일 뿐이고(되돌려진 편집 0건), 규칙은 역사보다 오래 산다.

### 실측

- 판정 A/B(HEAD `d6c16bb8` 대 편집본, 활성 27 + 착지 기록 있는 아카이브 1):
  **IDENTICAL 27 · DIFFERENT 1** — 다른 하나가 a099 이고, HEAD 의 **옛 기록**을 새 등식이
  이름으로 거절한 것이다(`landing point e6c4636adc43 is not the landing this change's
  evidence computes (21a315d17bb2)`). 재기록한 값으로는 오류 0 · required 32 다.
- 스위트 116개 초록(추가 4) · `make lint` rc=0 · `make sdd-test` 전부 초록 ·
  `openspec validate --strict` valid.
- 왕복: 76건 전부 계산값이 수락 집합의 원소이므로(§MEASURE 의 `73_accept.json`)
  `--record-landing` 이 쓰는 값은 이 등식을 통과한다. 픽스처에서도 왕복 시험이 재고 있다.

## VERIFY — task 7.8 (생존 변이 일곱과 느슨한 바늘 둘) (2026-09-13)

**생산 코드 변경 0** — `git diff --stat -- . ':!tools/logic-map/test_check_analysis.py'` 가
비었다. 6.3 과 같은 모양이라 판정 A/B 를 따로 돌리지 않는다: 판정 함수의 바이트가
같으므로 판정도 같다. 대신 실물 둘을 돌려 확인했다(아래 게이트 표).

### 먼저 다시 쟀다 — 리뷰가 센 열이 지금은 일곱이다

변이 정의는 지어내지 않았다. §6 리뷰(testing 전문가)가 남긴 `run_M*/` 사본을 기준선과
diff 해서 그대로 옮겼고, 7.2.1 이 자리를 바꾼 M1 만 같은 뜻으로 재표현했다.
[[caller-count-is-not-fix-site-count]] — 리뷰가 센 수는 리뷰 시점의 수다.

| 변이 | 무엇 | 리뷰(2026-09-12) | 7.8 **전** | 7.8 후 |
|---|---|---|---|---|
| M1 | `compute` 가 고정 판정을 건너뛴다 | SURVIVED | **CAUGHT** | CAUGHT (2) |
| M2 | `compute` 가 가장 높은 후보를 고른다 | SURVIVED | **CAUGHT** | CAUGHT (17) |
| M3 | 하한이 수정(`M`)을 안 센다 | SURVIVED | SURVIVED | **CAUGHT** |
| M4 | 아카이브된 기록이 안 보인다 | SURVIVED | SURVIVED | **CAUGHT** |
| M5 | `--record-landing` 이 항상 rc 0 | SURVIVED | **CAUGHT** | CAUGHT |
| M6 | `compute` 가 base 아래를 안 거른다 | SURVIVED | SURVIVED | **CAUGHT** (2) |
| M7 | `compute` 가 하한 없음을 안 막는다 | SURVIVED | SURVIVED | **CAUGHT** |
| M8 | `record` 가 빌림 거절을 안 한다 | SURVIVED | SURVIVED | **CAUGHT** |
| M9 | `record` 가 이관 거절을 안 한다 | SURVIVED | SURVIVED | **CAUGHT** |
| M11 | `record` 가 모호한 id 를 활성으로 떨어뜨린다 | SURVIVED | SURVIVED | **CAUGHT** |
| PC1 | `_pre_archive_path` 가 항상 빈 문자열 | (대조군) | CAUGHT | CAUGHT |
| PC2 | `resolve` 가 하한 순서를 안 본다 | (대조군) | CAUGHT | CAUGHT |

M1·M2·M5 를 죽인 것은 7.8 이 아니라 **7.3.1·7.4 가 쓴 시험**이다(다중 후보를 가진
픽스처가 그때 처음 생겼다). 리뷰의 열을 그대로 믿고 시험 열 개를 썼다면 셋은 이미
죽어 있는 것을 다시 죽이는 것이었다. 원복 확인 sha256 `36f72fc173da` (전·후 동일).

### 무엇을 어떻게 못 박았나

- **M3 — 그 자리에서 고친 번들.** 기존 시험 셋은 전부 **이동**(rename·심링크 전환)을
  잰다. `A`·`T` 만으로도 초록이라 `M` 이 빠져도 아무도 안 빨개졌다. 옮기지 않고
  `ast.json` 을 다시 쓰는 — 가장 흔한 — 모양을 하한 함수 층에서 잰다.
- **M4 — 아카이브 뒤에도 그 기록이 대상이다.** 옛 시험은 `check(...) == []` 만 봤는데
  이 픽스처는 **워킹트리를 대상으로 삼아도 통과한다**(base..worktree 의 변경 함수를
  증거가 그대로 기술하므로). 그래서 기록이 안 읽혀도 초록이었다. `facts["landing"]`
  을 아카이브 **전후로** 단언한다 — 재는 것은 판정이 아니라 대상이다.
- **M6 — 곁가지는 이 change 의 선 위에 없다.** 새 클래스
  `AComputedLandingIsAlwaysOneTheGateWillAccept`. `git rev-list base..HEAD` 는 "base
  에서 안 보이는 커밋 전부"라 병합이 있으면 base 를 한 번도 보지 못한 곁가지가 후보에
  들어온다. 저장소의 픽스처가 전부 선형이라 그 절이 한 번도 안 걸렸다.
  **먼저 닿는지 확인했다**([[mutation-must-reach-the-thing-under-test]]): 계측기
  대조군 시험 하나가 하한 = S · `base ≤ S` 거짓 · 순회 목록에 S 포함을 단언한다.
  변이 사본으로 실측한 종단 차이 — 정직: `record` 가 M 을 쓰고 `check` 초록(required 1),
  M6: `record` 가 **S** 를 쓰고 곧바로 `check` 가 `landing point precedes the
  comparison base` 로 **자기 기록을 거절**한다. 7.3.1 이 기록을 계산값과 묶은 뒤로
  그 값은 고칠 수도 없다([[two-judgements-cover-for-each-other]] 의 반대 방향).
- **M7 — 계산 경로의 하한 문장.** `resolve_landing` 에도 같은 뜻의 가드가 있어
  느슨한 바늘(`"never entered"`)로는 안 갈린다. 바늘을 `(commit the bundles)` 로
  좁혔다. 픽스처는 E 를 되돌려 번들을 **추적되지 않은** 파일로 만든다 — 추적 파일이
  안 바뀌므로 위의 dirty 거절에 안 걸린다(걸리면 하한이 아니라 그 가드를 재게 된다).
- **I10 — 위조 픽스처를 기록 경로에.** `TheGateRecordsTheLandingInsteadOfTheAuthor`
  는 "도구는 저자가 고를 값을 안 쓴다"를 주장하면서 위조 픽스처를 `record_landing`
  에 준 적이 없었다. 준다. 이 시험 하나가 M1 을 **독립으로 한 번 더** 잡는다
  (고정 판정을 지우면 도구가 위조 지점 E 를 기록한다).
- **M8 · M9 · M11 — 기록 경로의 거절 셋.** 빌림 · 이관 · 모호한 id. 셋 다 `check`
  쪽 거절만 시험이 있었다. 판정 둘이 서로를 덮는 모양이다 — 기록 경로가 값을 써
  버리면 5단계가 나중에 거절하지만 그때는 이미 감사 밖의 손잡이가 파일로 존재한다.
  셋 다 rc 와 문장에 더해 **파일이 안 생겼는지**까지 단언한다.

### I15 — 바늘 없는 음성 대조군이 무엇을 통과시켰나 (실측)

`self.assertTrue(check_analysis.check("mine", root))` 두 자리. 픽스처를 **다른 이유로**
깨뜨려서 쟀다.

| 픽스처 | 깨뜨린 방법 | 옛 단언 | 새 단언(`AST source hash is stale`) |
|---|---|---|---|
| 위조 | 그대로(대조군) | 통과 | 통과 |
| 위조 | 번들 통째 삭제 | **통과** | 실패 |
| 위조 | base 를 없는 커밋으로 | **통과** | 실패 |
| 자리표시자 | 그대로(대조군) | 통과 | 통과 |
| 자리표시자 | 번들 통째 삭제 | **통과** | 실패 |
| 자리표시자 | base 를 없는 커밋으로 | **통과** | 실패 |

음성 대조군이 재야 하는 것은 "오늘도 막힌다"가 아니라 "오늘은 **이 사유로** 막힌다"다.
번들 산문 하나만 지우는 방법으로는 안 갈렸다(두 오류가 같이 나온다) — 갈리는 것은
stale 판정 자체에 도달하지 못하게 만드는 둘이다.

### I16 — 결합과 사본 (실측)

| | HEAD | 지금 |
|---|---|---|
| 다른 TestCase 를 인스턴스로 만들어 사적 픽스처 호출 | **2회** | **0회** |
| 픽스처 빌더 (모듈 / 클래스) | 17 (3 / 14) | 19 (5 / 14) |

`_clean_fixture` · `_forged_fixture` 를 모듈 `_landed_work_fixture` ·
`_base_shaped_forgery_fixture` 로 올렸다. 픽스처를 한 클래스가 소유하는 한 그 결합은
없앨 수 없다 — 소유자를 옮기는 것이 수리다.

**안 한 것: 다섯 벌을 한 벌로 합치지 않았다.** 겹침을 재서 정했다 — 가장 높은 것이
`TheEvidenceFloorSurvives…._fixture` 68%, `_base_shaped_forgery_fixture` 67%,
`_own_work_fixture` 57%, `TheGateRecords…._fixture` 53%, `_placeholder_fixture` 51%,
`_merge_fixture` 46%, `_committed_change` 47%, `_renamed_bundle_fixture` 41%.
**동일한 사본은 하나도 없다.** 갈리는 부분이 각 클래스가 재는 바로 그것이다(base 앞의
R 커밋 · base 상태 해시 · `{}` 자리표시자 · `base_also_matches` 스위치 · 두 번째로
맞는 커밋). 스위치 다섯 개짜리 빌더 하나로 합치면 한 줄 편집이 다섯 클래스가 재는
것을 동시에 바꾼다 — [[two-judgements-cover-for-each-other]] 가 픽스처에서 나는 모양이다.

### 남은 것

I13 의 M10 은 `run_M10*` 사본이 없다(리뷰 당시 CAUGHT 라 안 남았다). I9~I16 중 **I16 의
사본 합치기만 의도적으로 안 했고** 위 표가 그 근거다. 7.5(성능)가 walk 를 바꾸면
M1·M2·M6 이 다시 열리므로 이 스위트가 그때의 안전망이다.

### 게이트 (2026-09-13)

| 게이트 | 결과 |
|---|---|
| `check_analysis` 스위트 | 116 → **125** (skip 1) |
| `make sdd-test` | rc=0 — scripts 15 · logic-map **200**(skip 1) · sdd 71 · sdd-history 22 · pm 16 · deploy 18 |
| `make lint` | rc=0 |
| `make test-seams` | rc=0 |
| `openspec validate --all --strict` | 58/58 |
| a099 5단계 (실물 기록) | rc=0 — `base af015dc9 → landed-commit 21a315d1 required 32` |
| a122 5단계 | rc=0 — `working tree required 0` |
| 생산 코드 | `git diff` 에 시험 파일 하나뿐 (+288 −58) |

## Pre-Edit Gate — task 7.6 (규칙 한 집) (2026-09-13)

### 먼저 정정 — 7.2.1 · 7.3.1 · 7.4 는 FLM 없이 편집했다

저장소 규칙은 기존 함수의 내부를 바꾸면 Function Logic Map 을 **먼저** 만들고, 생략하면
`not-applicable` 사유를 남기라고 한다. 7.1 은 `enumerate.py` 로 그렇게 했다. 그 뒤 세 task 는
`compute_landing` · `resolve_landing` · `record_landing` · `main` 의 내부를 바꾸면서 열거도 사유도
남기지 않았다 — **침묵한 생략**이다. 7.6 이 같은 함수를 다시 만지므로 7.2.1 직전(`eaf536d2`,
7.1 과 소스 동일)과 HEAD(`2b5b05c1`)를 같은 열거기로 뽑아 `ast.before-7.2.1.json` ·
`ast.before-7.6.json` 으로 남겼다. 늘어난 분기는 스크립트가 소스 한 줄로 대조했다:

| 함수 | 분기 | raise | 새로 생긴 것 (열거 그대로) | 만든 task |
|---|---|---|---|---|
| `compute_landing` | 8 → 10 | 0 → 0 | `if _pinning_at(…)[1]: continue`(부정 뒤집기) · `if not unheld:` · `if unheld:` | 7.2.1 |
| `resolve_landing` | 10 → 13 | 8 → 10 | `if unheld:` + raise · `if candidate != computed:` + raise · `computed[:12] if computed else …` | 7.2.1 · 7.3.1 |
| `record_landing` | 11 → 16 | 0 → 0 | `if landing_file.is_symlink():` · `try/except GATE_FAULTS` · `try/except FileExistsError` | 7.4 |
| `main` | 13 → 17 | 0 → 0 | `try/except GATE_FAULTS` 두 쌍 | 7.4 |
| `_evidence_floor` | 6 → 6 | 0 → 0 | (분기 없음 — git 인자만 `-M100% --no-follow --diff-filter=MAT`) | 7.2.1 |
| `_unheld_bundles` | — → 9 | — | 7.2.1 이 새로 만든 함수 | 7.2.1 |

세 task 의 **판정 근거**는 각 VERIFY 절의 뮤테이션·A/B 실측이었고 그것은 그대로 유효하다.
빠진 것은 "편집 전에 무엇이 있었나"의 기계 열거였고, 위 표가 그 자리를 채운다.

### 편집 전 열거 (HEAD `2b5b05c1`)

| 파일:함수 | 분기 | 반환 | raise | 줄 |
|---|---|---|---|---|
| `check_analysis.py:check` | 44 | 14 | 0 | L978-1107 |
| `check_analysis.py:resolve_landing` | 13 | 2 | 10 | L570-679 |
| `check_analysis.py:compute_landing` | 10 | 6 | 0 | L1110-1151 |
| `check_analysis.py:record_landing` | 16 | 11 | 0 | L1154-1236 |
| `check_analysis.py:_pre_archive_path` | 3 | 2 | 0 | L476-483 |
| `check_analysis.py:resolve_referenced_change` | 10 | 1 | 3 | L241-279 |
| `execution_baseline.py:validate` | 42 | 2 | 24 | L393-497 |
| `execution_baseline.py:_change_analysis_path` | 3 | 1 | 2 | L74-80 |

### I2 — 수락 규칙 두 벌. 열거가 보여 주는 것

`resolve_landing` 의 raise 열 줄 중 **여섯**(L614 base · L623 고정 0 · L628 불일치 · L640 하한 없음 ·
L645 하한 순서 · L655 미보유)과 `compute_landing` 의 분기 B1·B2·B7·B8·B10 이 **같은 여섯 조건**이다.
`_pinning_at` 의 첫 값은 후보와 무관하게 `len(_pinning_bundles)` 이므로 B1(순회 전)과 L623 은 같은
조건이다. 갈리는 것은 **순서**뿐이다 — `resolve` 는 base → 고정 → 불일치 → 하한 없음 → 하한 순서 →
미보유, `compute` 는 고정 → 하한 없음 → (base ∧ 하한 순서) → 불일치 → 미보유.

**순서는 거절 지점이라 못의 일부다.** `compute` 순서로 맞추면 비용은 0 인데
`test_a_landing_the_evidence_does_not_describe` 가 증거보다 **앞선** 커밋(P)을 선언해서 불일치
문장을 못 박고 있으므로, 그 입력이 하한 가드에 먼저 걸려 불일치 가드가 못에서 빠진다
([[a-new-guard-unpins-the-guards-behind-it]]). 그래서 규칙은 **`resolve` 의 순서로** 한 함수에 둔다.

**그 선택의 비용을 먼저 쟀다** (`76_order.py`, 93건 전수, 착지가 계산되는 지점까지):
오늘 `_pinning_at` 을 부르는 후보 **6,491** · 합친 순서에서 추가로 부르는 후보(base ≤ c 이지만
하한 뒤가 아닌 곁가지) **328**(+5%) · 싸게 건너뛰는 후보 32. 추가의 **322** 가 이미 90~294초
걸리는 아카이브 change 일곱(a055·a054·a042 …, 전부 같은 병합 구간)의 몫이고, 게이트가 실제로
도는 활성 change 에서는 a089·a095 각 **2** 뿐이다. 그 walk 는 7.5(성능 — `--ancestry-path`)가
맡는다. 계측기 결함 하나를 먼저 고쳤다: 저장된 표가 아카이브를 날짜 붙은 이름으로 들고 있어
첫 판은 93건 중 12건만 셌다.

### I4 — 없는 id 에 "base 를 capture 하라". 거부할 정상 입력을 먼저 쟀다

`check` 와 `record_landing` 이 `except ValueError: change_dir = changes/<id>` 를 한 벌씩 들고
있다. 해소기가 "없다"(또는 "아카이브에 사본이 둘")라고 말한 id 를 없는 경로로 바꿔 넘기므로
그 다음 `resolve_base` 가 **"`capture_change_base.py --change <id>` 를 돌려라"** 를 권한다 — 오타
난 id 에 새 change 의 base 를 만들라는 조언이다. 생산 호출자는 `tools/gate.sh:321` 하나이고
1단계가 id 를 이미 해소한 뒤다. **전수**(`76_fallback.py`): 게이트가 받을 수 있는 id 126개
(활성 + 아카이브에서 날짜를 벗긴 것) 중 해소 **126** · 모호 0 · fallback **0**, 빌림 참조 1개도
해소된다. fallback 을 이름 붙은 거절로 바꿔도 **거부되는 정상 입력은 0** 이다.

### I3 · I5 · I6 · I7

- **I3** `validate` 의 지역 `canonical = f"openspec/changes/{CHANGE}"` 가 모듈 함수 `canonical()` 을
  가린다. 오늘 그 함수 안에서 `canonical(...)` 호출은 0 이라 결함은 아니고 **덫**이다. 이름만 바꾼다
  — 열거의 분기·반환·raise 수가 같아야 한다.
- **I5** 이관 거절 문장: `check` "the window **ends** at" · `record_landing` "the window **already
  ends** at". 사용처는 코드 둘뿐(얼린 열거 사본 제외). 게이트가 찍는 쪽 문장을 상수 하나로.
- **I6** `_pre_archive_path` 가 `openspec/changes/archive/` 와 `ARCHIVED_CHANGE` 로 아카이브 이름을
  한 번 더 해독한다. 이름 → id 규칙을 함수 하나로 두고 해소기와 함께 쓴다. **id 를 호출 사슬로
  내려보내는 판본은 안 한다** — `_evidence_floor(root, analysis)` 를 직접 부르는 시험이 열 개가
  넘고, 서명을 바꿔도 판정은 그대로다. shell 쪽 사본(`tools/gate.sh`)은 언어가 달라 못 합치며
  `test_gate_resolves_archived_changes.py` 가 따로 못 박는다.
- **I7** `_change_analysis_path` 는 6.2 뒤로 **적힌 자리**(옮기기 전)를 보는데 "outside **current**
  change analysis" 라고 말한다.

### 생산 판정에 미칠 영향 (예측 — 편집 뒤 실측으로 확인한다)

I2·I3·I6 은 판정 불변, I4·I5·I7 은 오류 **문장**만 바뀐다(실물 입력에서 그 갈래에 닿는 것 0).
판정 A/B 28건 IDENTICAL 을 기대하고, 착지 전수 93건의 계산값이 76건 그대로여야 한다.

## VERIFY — task 7.6 (규칙 한 집) (2026-09-13)

### RED → GREEN

RED 아홉을 먼저 세웠고 **맞는 이유로** 빨갰다 — 옛 코드는 오타 id(`mnie`)와 아카이브 사본 둘 모두에
`missing base-commit.txt; run capture_change_base.py --change <change-id>` 를 권했다(실측 출력).
구현 뒤 `test_check_analysis` + `test_execution_baseline` **157** 초록.

| 리뷰 | 무엇을 바꿨나 |
|---|---|
| I2 | 새 `_landing_refusal(root, base, candidate, analysis, floor) -> (사유, 미보유 이름)`. 여섯 조건을 `resolve_landing` 의 순서 그대로. `resolve_landing` 은 선언 **글자** 판정 셋 + 규칙 + 7.3.1 등식, `compute_landing` 은 순회 전 조기 종료 둘 + 후보마다 규칙 |
| I3 | `validate` 의 지역 `canonical` → `recorded_dir` |
| I4 | `check` · `record_landing` 의 `except ValueError: change_dir = changes/<id>` 삭제 → 해소기의 문장으로 멈춘다. 해소기의 못 찾음 문장에서 "reference" 를 뺐다(대상 id 에도 나간다). `AmbiguousChange` 타입 제거 — 가를 호출자가 0 |
| I5 | 모듈 상수 `ADOPTION_REFUSES_A_LANDING`. 두 경로가 그것을 쓰고 시험이 **등식**으로 단언한다 |
| I6 | `_archived_change_id(name)` + `ARCHIVE_PREFIX`. 해소기와 `_pre_archive_path` 가 그것에 묻는다 |
| I7 | "outside current change analysis" → "outside this change's recorded analysis directory" |

"한 집"은 **행동 시험으로 못 박히지 않는다** — 일치하는 두 사본은 오늘 같은 답을 내므로 초록이다.
그래서 `TheLandingRuleLivesInOnePlace` 만 소스 구조를 본다: `resolve_landing` · `compute_landing` 둘 다
`_landing_refusal` 을 부르고 `_pinning_at` · `_unheld_bundles` 를 직접 부르지 않는다, 해소기와
`_pre_archive_path` 가 `_archived_change_id` 를 부르고 `ARCHIVED_CHANGE.fullmatch` 를 직접 부르지 않는다.

### 열거 대조 (편집 전 `ast.before-7.6.json` → 후 `ast.after-7.6.json`, 스크립트 대조)

| 함수 | 분기 | 반환 | raise | 요지 |
|---|---|---|---|---|
| `resolve_landing` | 13 → 8 | 2 → 2 | 10 → 5 | 가드 여섯이 빠지고 `if refusal: raise ValueError(refusal)` 하나 |
| `compute_landing` | 10 → 8 | 6 → 6 | 0 → 0 | 사본 네 줄(`base∧하한` BoolOp/If · 불일치 · `if not unheld`)이 `if not refusal` · `if names` 로 |
| `_landing_refusal` | — → 6 | — → 7 | — | 새 함수. 사유 문장은 편집 전 raise 문장과 글자 단위로 같다 |
| `check` | 44 → 43 | 14 → 14 | 0 | handler 둘 → 하나. 나머지 분기 42 는 소스 한 줄 단위로 같다 |
| `record_landing` | 16 → 15 | 11 → 11 | 0 | handler 둘 → 하나 |
| `resolve_referenced_change` | 10 → 10 | 1 → 1 | 3 → 3 | 해독을 `_archived_change_id` 에 · 문장 둘 |
| `_pre_archive_path` | 3 → 3 | 2 → 2 | 0 | 해독을 `_archived_change_id` 에 |
| `validate` | 42 → 42 | 2 → 2 | 24 → 24 | 차이 0 (이름만) |
| `_change_analysis_path` | 3 → 3 | 1 → 1 | 2 → 2 | 문장 하나 |

### 뮤테이션 — 규칙 한 집이 **두 경로 모두에서** 보이는가 (`76_mut.py`)

경로 분류는 이름으로 추측하지 않고, 빨개진 시험의 소스가 `record_landing`/`compute_landing`(계산)을
부르는지 `check`/`resolve_landing`(선언)을 부르는지를 AST 로 셌다.

| 변이 | 첫 판 | 7.6 후 | 빨개진 경로 |
|---|---|---|---|
| R1 규칙: base 조상 판정 삭제 | CAUGHT | CAUGHT | 계산 · 선언 |
| R2 규칙: 고정 0 판정 삭제 | CAUGHT | CAUGHT | 선언 (계산은 순회 전에 돌아가서 구조상 안 닿는다) |
| R3 규칙: 불일치 판정 삭제 | CAUGHT | CAUGHT | 계산 · 선언 |
| R4 규칙: 하한 없음 판정 삭제 | CAUGHT | CAUGHT | 선언 (계산은 구조상 안 닿는다) |
| R5 규칙: 하한 순서 판정 삭제 | CAUGHT | CAUGHT | 선언 → **계산 · 선언** |
| R6 규칙: 미보유 판정 삭제 | CAUGHT | CAUGHT | 계산 · 선언 |
| R7 compute 가 거절을 무시 | CAUGHT | CAUGHT | 계산 |
| R8 compute 가 미보유 이름을 안 모음 | CAUGHT | CAUGHT | 계산 |
| R9 resolve 가 거절을 무시 | CAUGHT | CAUGHT | 선언 |
| R10 아카이브 이름 해독이 항상 빈 문자열 | CAUGHT | CAUGHT | 12개 |
| R11 아카이브 이름 해독이 접미사로 고름 | CAUGHT | CAUGHT | 구조 시험 |
| R12 · R13 check · record 가 fallback 을 되살림 | CAUGHT | CAUGHT | 둘 다 |
| R14 record 의 이관 문장이 다시 갈림 | CAUGHT | CAUGHT | 계산 |
| R15 `_change_analysis_path` 문장 되돌림 | CAUGHT | CAUGHT | — |
| **R16 compute 가 규칙에 하한 대신 base 를 넘김** | **SURVIVED** | **CAUGHT** | 계산 |

**R16 은 7.6 이 만든 구멍이 아니라 드러낸 구멍이다.** 옛 `compute_landing` 의 하한 순서 절을 지우는
변이는 7.8 목록에 없었고, 규칙을 한 함수로 옮기며 "계산 경로가 규칙에 **하한을 넘기는가**"라는
배관이 처음 따로 보였다. 동등 변이로 넘기지 않았다([[surviving-mutant-may-mean-accidental-safety]]):
모든 픽스처에서 하한 앞의 후보가 **불일치로 먼저** 떨어져 가려졌을 뿐이다. 그 가드 **혼자** 막아야
하는 모양을 두 번 만에 찾았다 — 첫 판(곁가지에 증거, 줄기에 같은 증거를 나중에)은 계산 경로가
`start`(=하한)를 **먼저** 받아서 안 닿았고, 줄기가 저장소 규칙대로 **증거를 먼저** 커밋해 하한 커밋
자체가 불일치로 떨어지는 모양에서 닿았다. 닿는지를 대조군 시험으로 먼저 단언하고(하한 = F · F 는
B 의 조상이 아님 · F 는 불일치 · B 는 하한을 base 로 주면 받고 F 로 주면 거절) 왕복 시험으로 굳혔다.
커밋 시각을 고정했다 — `git log -1` 은 시각 순이라 같은 초에 겹치면 하한이 실행마다 달라진다.

7.8 의 변이도 새 코드에 다시 걸었다(`78_mut.py`): M2·M3·M4·M5·M7·M8·M9·PC2 **CAUGHT**, M1·M6·M11·PC1 은
치환 자리가 규칙 함수로 옮겨 가 SKIP — 같은 뜻의 R3·R1·R13·R10 이 잡는다. 원복 sha256 `59bfba56a695` 동일.

### 판정 A/B (`72_ab.py --refresh`, HEAD `2b5b05c1` 사본 대 워킹트리)

**IDENTICAL 28 · DIFFERENT 0** — 활성 27 + 착지 기록이 있는 아카이브 1(a099: `21a315d1` · required 32 · 오류 0).

### `compute_landing` 전수 비교 (`76_sweep.py`, HEAD 사본 대 워킹트리, 같은 change 에서 연달아)

판정 A/B 28건에는 착지를 **계산하는** 아카이브 대부분이 없다. 규칙의 순서가 바뀐 계산 경로(곁가지
후보에서 번들 대조를 먼저 한다)가 받는 값을 바꾸지 않는지는 여기서만 보인다.

| | 결과 |
|---|---|
| 착지값 + 사유 문장 | **SAME 93 · DIFF 0** |
| 착지가 나오는 change | 76 (7.3 측정과 같다) |
| 시간 합계 | 1220s → 1270s (**+4.1%**) — Pre-Edit 의 예측 "번들 대조 +5%" 와 맞는다 |
| 활성 change 12건 합계 | 44s → 45s |

change 별 시간은 ±4% 넘게 흔들린다 — 예측 목록(X > 0)에 없던 `enable-vpn-console-access` 가 +17% 로
나왔고, 그 구간에 이 세션이 `gh api` 호출을 겹쳐 돌렸다. 그래서 근거는 합계만 쓴다. 가장 큰 증가는
예측 목록의 a054(+34%)·a055(+10%)다.

### 게이트 (2026-09-13, 이 로트의 워킹트리)

| 게이트 | 결과 |
|---|---|
| `check_analysis` 스위트 | 125 → **132** (skip 1) — 구조 3 · I4 2 · R16 2 |
| `make sdd-test` | rc=0 — scripts 15 · logic-map **207**(skip 1) · sdd 71 · sdd-history 22 · pm 16 · deploy 18 |
| `make lint` · `make test-seams` | rc=0 · rc=0 |
| `openspec validate --all --strict` | 58/58 |
| a099 5단계 (실물 기록) | rc=0 — `landed-commit 21a315d1 required 32` |
| a122 5단계 | rc=0 (파이프 없이 잰 종료 코드) |
| 오타 id `mnie` — 판정 · 기록 | 둘 다 rc=1 `change is neither open nor archived: mnie` (옛: "capture_change_base 를 돌려라") |

### 안 한 것

- **id 를 호출 사슬로 내려보내 `_pre_archive_path` 를 없애는 판본** — 서명 넷이 바뀌고 그것을 직접
  부르는 시험이 열 개가 넘는데 판정은 그대로다. 대신 해독 규칙을 함수 하나로 모았다.
- **shell 해소기(`tools/gate.sh`)와의 통합** — 언어가 다르다. `test_gate_resolves_archived_changes.py` 가 따로 못 박는다.
- **곁가지 후보의 +328 번들 대조** — 7.5(성능)의 몫. `--ancestry-path` 로 순회 목록 자체가 하한의
  자손만 내면 규칙을 건드리지 않고 사라진다.
- `validate` 의 이름 덫(I3)에는 시험을 안 세웠다 — 변수 이름을 재는 시험은 이 저장소의 관례가 아니고,
  열거 대조(분기·반환·raise 전후 동일, 소스 차이 0)가 편집이 이름뿐임을 보인다.

## Pre-Edit Gate — task 7.7 (조언 줄이 기록 명령에 묻는다) (2026-09-13)

### 먼저 다시 쟀다 — 리뷰가 센 넷은 일곱이고, 실물에서는 99 권유 중 38 이 모순이다

리뷰 I8 이 적은 모양은 넷이다(디스크에만 있는 기록 · staged 아카이브 이동 · 빌리는 쪽 · 번들 0).
수리 전에 HEAD(`3da639a9`) 사본으로 **픽스처에서 하나씩** 재현했다(`77_shapes.py`). 조언 줄이
`--record-landing` 을 권하고 곧바로 그 명령을 돌린 결과다:

| 모양 | 조언 줄 | 명령 | 리뷰 |
|---|---|---|---|
| S0 대조: 기록 없음 · 번들 있음 · 깨끗한 트리 | 권함 | **rc 0 기록** | — |
| S1 기록이 디스크에만 있다(커밋 전) | `working tree (no landed-commit.txt)` + 권함 | `already exists — not overwritten` | I8 |
| S2 기록 커밋 뒤 아카이브 이동이 staged | 같음 | `already exists — not overwritten` | I8 |
| S3 빌리는 쪽(양쪽 다 기록 없음) | 권함 | `this change borrows its evidence …` | I8 |
| S4 번들 0 | 권함 | `no landing recorded — no revision: current evidence pins …` | I8 |
| S5 추적 파일 수정 | 권함 | `the working tree has uncommitted changes …` | **새로 셈** |
| S6 번들이 커밋 전 | 권함 | `… never entered this history (commit the bundles)` | **새로 셈** |
| S7 어느 커밋도 번들과 안 맞음 | 권함 | `… matches every pinning bundle …` | **새로 셈** |

S1~S6 과 끊긴 심링크 기록은 명령이 **후보를 걷기 전에** 멈추는 자리다. S7 만 걸어야 안다.

**실물 전수**(`77_census.py`) — 게이트가 받을 수 있는 id 126 전부에 HEAD 판본 `check` 를 돌려 `main` 이 어느
조언 갈래로 가는지 context 로 가르고, 권유 갈래면 `record_landing` 을 **기록 없이** 돌렸다(`landed-commit.txt`
를 여는 쓰기를 가로채 예외로 바꿨다 — 저장소에 안 썼다). 계측 동안 추적 파일은 건드리지 않았다(dirty 거절이 섞이지
않게 — 끝난 뒤 `git status` 로 추적 변경 0 확인).

| 조언 갈래 | 수 | 명령 결과 |
|---|---|---|
| 권함 | **99** | 기록 61 (활성 3: a066 · a071 · a112) · **거절 38** |
| — 걷기 전 거절: 번들 0 | 22 | **활성 12** (a067 · a068 · a070 · a087 · a107 · a113 · a114 · a115 · a121 · a122 · align-full-sdd-pm-contract · verify-observes-the-trigger) · 아카이브 10 |
| — 걷기 전 거절: 빌리는 쪽 | 1 | a073 |
| — 걸어서 거절: 어느 커밋도 안 맞음 | 15 | 전부 아카이브. 명령 한 번 5.6 ~ **225.1초**(a055) |
| base 모양 번들(7.1 갈래, 권하지 않음) | 16 | 활성 9 |
| 창을 안 찍음(해소 전 실패) | 10 | — |
| 착지 기록 있음 | 1 | a099 |

**활성 change 에서 권유 15 중 12 가 모순이었다.** 깨끗한 실물 트리에는 S1 · S2 · S5 · S6 이 0 건이다 — 작업 중에만
생기는 상태다.

### 편집 전 열거 (HEAD `3da639a9` blob → `ast.before-7.7.json`)

코드에 손대기 **전에** 뽑았다. 설계는 이 열거에서 나왔다 — `record_landing` 의 반환 열하나 중 여섯(심링크 ·
기록 존재 · 빌림 · base · 이관 · dirty)과 `compute_landing` 의 순회 전 반환 둘(번들 0 · 하한 없음)이 "걷기 전에
멈추는 자리"이고, `main` 의 권유 갈래(B13 `else`)는 그중 아무것도 부르지 않는다.

| 함수 | 분기 | 반환 | 줄 |
|---|---|---|---|
| `_target_text` | 2 | 2 | L373-381 |
| `record_landing` | 15 | 11 | L1186-1264 |
| `compute_landing` | 8 | 6 | L1144-1183 |
| `check` | 43 | 14 | L1015-1141 |
| `main` | 17 | 3 | L1267-1352 |

**순서를 적는다**: 열거(기계) → 설계 → RED → GREEN 을 **scratchpad 사본**에서 먼저 했고(전수 계측이 도는 동안
추적 파일을 못 고쳐서), FLM 산문 절과 Branch Test Map 은 GREEN 뒤·저장소 반영 뒤에 썼다. Branch Test Map 은
7.6 과 같이 변이가 **실제로 빨갛게 만든** 시험으로 채우므로 뮤테이션 뒤에만 쓸 수 있다.

### 설계 결정

- **판정 하나** `_recording_refusal(change, change_dir, root) -> (사유, base)` — 기록 명령의 걷기 전 거절 여섯 +
  `_walk_floor`(계산의 순회 전 사유 둘을 뺀 함수). 기록 명령과 조언 줄이 **둘 다** 이것을 부른다. 따로 두면 명령에
  거절이 늘 때 조언만 옛 조건으로 남는다([[two-judgements-cover-for-each-other]]).
- **걷지 않는다.** 걸어서 아는 거절(15건)을 예측하려면 워킹트리가 대상인 모든 5단계가 순회를 돈다 — 최대 225초,
  7.5(성능)와 반대 방향이다. 대신 권유 문장이 기록을 **약속하지 않는다**("if no commit on this history matches that
  evidence, the command says so instead of recording").
- **판정은 `check` 가 아니라 `main` 의 권유 갈래에서 부른다.** `check` 의 `facts` 에 넣는 안을 먼저 봤고 버렸다 —
  조언 때문에 부른 git 이 멎으면 **판정**이 결함 한 줄로 바뀐다. `main` 에서는 자기 `GATE_FAULTS` 로 받아 "cannot
  tell" 로 명령을 권하지 않고 판정 줄은 그대로 둔다.
- **대상 텍스트** `(no landed-commit.txt)` → `(no landed-commit.txt in HEAD)` — 게이트가 읽는 자리를 말한다.
- **디스크에만 있는 기록**은 HEAD 기록과 문장을 가른다: 경로 + `not in HEAD` + "commit it". 경로를 적는 이유 —
  staged 아카이브 이동은 HEAD 의 옛 자리에 기록이 있어서 "HEAD 에 없다"만으로는 틀리다.

## VERIFY — task 7.7 (2026-09-13)

### RED → GREEN

RED 12 가 **맞는 이유로** 빨갰다(`77_red.log`): 표 시험의 일곱 모양 전부 "명령의 거절 문장이 조언 줄에 없다",
구조 시험 "`record_landing` 이 `_recording_refusal` 을 안 부른다", 7.4 시험의 대상 텍스트. 양성 대조군
(`test_where_the_command_would_record_the_advice_still_names_it`)과 계산 경로 직접 시험은 HEAD 에서도 초록이다 — 뒤의
것은 기존 행동의 못이다.

GREEN 첫 판에서 **7.4 시험이 회귀를 잡았다**: 걷기 전 부분(`_walk_floor` → `_pinning_bundles`)이 저장소 밖을
가리키는 번들에서 `ValueError` 를 내는데, 판정 호출을 `record_landing` 의 결함 경계 밖에 뒀다. 경계 안으로 옮겼다
(변이 R21 이 그 자리를 다시 잰다).

| 시험 (새로 9) | 재는 것 |
|---|---|
| `test_the_advice_never_names_a_command_that_would_refuse` | 일곱 모양(subTest). 명령을 **실제로 돌려** 나온 거절 문장이 조언 줄 안에 있고, 권유 문장이 없고, 대상 텍스트가 `in HEAD` 다. 문장을 시험에 옮겨 적지 않는다 |
| `test_a_record_on_disk_is_named_as_not_committed` | 디스크 기록 = 경로 + `not in HEAD`; 커밋 뒤에는 옛 문장 그대로 |
| `test_where_the_command_would_record_the_advice_still_names_it` | 양성 대조 — 권하고, 명령이 rc 0 으로 기록 |
| `test_a_refusal_only_the_walk_finds_is_not_promised_away` | S7 — 권하되 약속 안 함, 명령은 걸어서 거절 |
| `test_a_fault_while_asking_the_recorder_does_not_become_advice` | git 이 멎으면 권하지 않고 판정 줄은 결함 없는 실행과 같다 |
| `test_a_refusal_that_outlives_a_commit_is_named_before_one_that_does_not` | 빌림+dirty · 번들 0+dirty — 영구 사유가 먼저 |
| `test_the_recorder_and_its_advice_ask_one_judge` | 구조: 두 호출자가 판정을 부르고, 명령이 거절 조각을 직접 안 부르고, 하한 사유는 `_walk_floor` 한 곳 |
| `test_the_computation_names_what_stops_it_before_the_walk` | `compute_landing` 직접 — 기록 명령이 판정에서 먼저 멈춰 이 두 반환이 그 경로로 안 닿는다 |
| `test_a_base_it_cannot_resolve_is_named_as_the_base` | 기록 명령의 base 해소 실패 문장 |

### 열거 대조 (편집 전 `ast.before-7.7.json` → 후 `ast.after-7.7.json`, 스크립트 대조 — `77_flm.py`)

| 함수 | 분기 | 반환 | 요지 |
|---|---|---|---|
| `record_landing` | 15 → 8 | 11 → 6 | 거절 여섯의 갈래·반환이 `if refusal:` 하나로 |
| `compute_landing` | 8 → 7 | 6 → 5 | 순회 전 둘이 `if why:` 하나로 |
| `main` | 17 → 20 | 3 → 3 | `try` · `except GATE_FAULTS` · `if refusal:` |
| `_target_text` | 2 → 2 | 2 → 2 | 반환 문장 하나 |
| `check` | 43 → 43 | 14 → 14 | **차이 0**(호출 집합 동일) — 편집 안 함 |
| `_recording_refusal` | — → 9 | — → 9 | 새 함수 |
| `_walk_floor` | — → 2 | — → 3 | 새 함수 |

### 뮤테이션 (`77_mut.py`, 저장소와 sha256 이 같은 사본에서 · 원복 sha256 확인)

| 변이 | 첫 판 | 최종 | 잡은 시험(셋까지) |
|---|---|---|---|
| R1 판정: 심링크 거절 삭제 | CAUGHT | CAUGHT | 심링크 기록 시험 · 표 |
| R2 판정: HEAD 기록 거절 삭제 | CAUGHT | CAUGHT | 디스크 기록 · 해독 못 하는 기록 |
| R3 판정: 디스크 기록 거절 삭제 | CAUGHT | CAUGHT | 디스크 기록 · 덮어쓰기 거절 · 표 |
| R4 판정: 디스크 문장을 HEAD 문장으로 합침 | CAUGHT | CAUGHT | 디스크 기록 (표는 못 잡는다 — 명령과 조언이 같은 판정이라 둘이 같이 바뀐다, 설계대로) |
| R5 판정: 빌림 거절 삭제 | CAUGHT | CAUGHT | 빌림 거절 · 순서 |
| **R6 판정: base 해소 실패를 안 받음** | **SURVIVED** | CAUGHT | `test_a_base_it_cannot_resolve_is_named_as_the_base` |
| R7 판정: 이관 거절 삭제 | CAUGHT | CAUGHT | 이관 기록 거절 |
| R8 판정: dirty 거절 삭제 | CAUGHT | CAUGHT | dirty 거절 · 표 |
| R9 판정: 걷기 전 하한 사유 무시 | CAUGHT | CAUGHT | 표 · 순서 (기록 명령은 계산이 대신 거절 — 조언 쪽에서만 드러나는 자리) |
| **R10 판정: 순서 — dirty 를 빌림 앞으로** | **SURVIVED** | CAUGHT | 순서 시험 |
| R11 판정: base 대신 빈칸(배관) | CAUGHT | CAUGHT | 9 |
| R12 기록: 판정 거절 무시 | CAUGHT | CAUGHT | 10 |
| R13 조언: 판정을 안 묻고 권함 | CAUGHT | CAUGHT | 표 · 결함 · 순서 |
| R14 조언: 결함이면 권함 | CAUGHT | CAUGHT | 결함 |
| R15 조언: 해소기 대신 활성 경로를 넘김(배관) | CAUGHT | CAUGHT | 표(staged 아카이브) |
| R16 계산: 걷기 전 사유 무시 | CAUGHT | CAUGHT | 계산 직접 시험 **하나** |
| R17 하한: 번들 0 판정 삭제 | CAUGHT | CAUGHT | 번들 0 거절 · 계산 직접 · 순서 |
| R18 하한: 하한 없음 판정 삭제 | CAUGHT | CAUGHT | 표 · 계산 직접 · 커밋 전 증거 |
| R19 대상 텍스트 되돌림 | CAUGHT | CAUGHT | 표 · 7.4 CLI |
| R20 권유 문장이 다시 약속함 | CAUGHT | CAUGHT | S7 |
| R21 기록: 판정을 결함 경계 밖에서 | CAUGHT | CAUGHT | 7.4 저장소 밖 번들 |
| R22 판정: 순서 — dirty 를 하한 사유 앞으로(옛 순서) | — | CAUGHT | 순서 시험 |

**R6 은 우연이 지키던 안전이었다**([[surviving-mutant-may-mean-accidental-safety]]) — HEAD 에도 이 갈래를 재는 시험이
0 이다. 변이 아래서도 rc 는 1 이었다(바깥 `GATE_FAULTS` 가 `no landing recorded — …` 로 받는다). 기록이 안 쓰이는 것은
경계가 지켰고, 빠진 것은 "**무엇을** 못 풀었나"였다.

**R10 은 순서가 못에 없다는 뜻이었고, 그 순서는 틀려 있었다.** 조언 줄은 사유를 **하나**만 말하므로 순서가 곧
조언이다. 기록 명령의 순서(dirty → 번들 0 → 하한 없음)를 그대로 쓰면 영원히 기록할 수 없는 change 가 "먼저 커밋하라"를
듣고 커밋한 뒤에야 진짜 사유를 듣는다 — 위 전수의 **활성 번들 0 열둘**이 tasks.md 한 줄만 고쳐도 그 상태다. dirty
하나를 맨 뒤로 옮겼다(기록 명령에서는 번들 0·하한 없음·dirty 가 겹친 입력의 **문장**만 바뀐다, rc 는 둘 다 1).
그 순서를 두 조합으로 못 박았고 R10 · R22 가 잡힌다. 실물에서도 드러난다: 편집으로 dirty 인 지금 트리에서 번들 0
활성 열둘은 dirty 가 아니라 번들 0 을 말한다(아래 C).

7.6 규칙 변이(`76_mut.py`, 저장소 대상)도 새 코드에 다시 걸었다: R1~R13 · R15 · R16 **CAUGHT**, R14(이관 문장 갈림)는
문장이 판정 함수로 옮겨 가 치환 자리가 없어 SKIP — 새 자리(`return ADOPTION_REFUSES_A_LANDING, ""`)에 같은 변이 R14' 를 걸어
**CAUGHT**(`test_the_recorder_refuses_the_adoption_path_too`). 7.8 변이(`78_mut.py` 경로만 사본으로 바꾼 `77_mut78.py`):
M2 · M3 · M4 · M5 · M7 · M8 · M9 · PC2 **CAUGHT**, M1 · M6 · M11 · PC1 SKIP — 7.6 때와 같은 넷이고 같은 뜻의 76 R3 · R1 · R13 · R10 이
잡는다. 원복 sha256 `1dfa432b0081` 동일.

**계측 사고 한 건**: R14' 를 처음 걸던 실행이 사용자 중단으로 원복 전에 멎어 **사본**(`77_work`)에 변이가 남았다. 다음
실행의 sha256 단언이 그것을 잡았고(`de86b65c…` ≠ `1dfa432b…`), 저장소 파일은 그대로였다(sha 대조). 사본을 저장소
바이트로 되돌린 뒤 다시 쟀다. 그 뒤 7.8 하네스는 저장소가 아니라 사본을 대상으로 돌렸다 — 중단돼도 저장소가 안 남게.

### 실물 A/B (`77_ab.py`)

| | 결과 |
|---|---|
| (A) 판정 — 활성 27 + 착지 기록 아카이브 1, HEAD 사본 대 편집 판본 `check` | **IDENTICAL 28 · DIFFERENT 0** (대상 텍스트 치환 하나만 정규화). 정규화 전 갈린 10 은 전부 번들 0 change 의 `missing Function Logic Map … and working tree (…)` 문자열 |
| (B) 조언 — 전수의 권유 99 에 새 판정, dirty 확인만 "깨끗함"으로 가로채 계측과 같은 상태 | **일치 99 · 불일치 0** — 걷기 전 거절 23 은 명령 문장과 **글자 그대로**, 기록 61 · 걸어서 거절 15 는 빈 사유(=권유) |
| (C) 판정 비용 — 활성 권유 15, 가로채지 않고 | 번들 0 열둘 **0.01~0.20초**(dirty 확인 전에 멈춘다) · 번들 있는 셋 1.14~1.29초(`git diff --quiet HEAD`) |

전수에서 `check` 한 번이 활성 0.1~23.7초 · 전체 0.1~34.9초였으므로 (C) 는 번들 있는 셋에서 +1.2초 안팎이고 번들 0 에서는 잴 수 없을 만큼 작다.

### 실물 CLI (편집 판본, 파이프 없이 잰 종료 코드)

| change | rc | 조언 줄 끝 |
|---|---|---|
| a122 (번들 0) | 0 | `--record-landing` cannot narrow it: no landing recorded — no `revision: current` evidence pins a landing for this change |
| a073 (빌리는 쪽 · 아카이브) | 1 (오류 336, HEAD 와 같음) | `--record-landing` cannot narrow it: this change borrows its evidence — copy the landing recorded on the change that owns the bundles … |
| a099 (착지 기록) | 0 | 조언 줄 없음 — `landed-commit 21a315d1… required 32` |

### 게이트 (2026-09-13, 이 로트의 워킹트리)

| 게이트 | 결과 |
|---|---|
| `check_analysis` + `execution_baseline` 스위트 | **168** (check_analysis 132 → **141**, skip 1) |
| `make sdd-test` | rc=0 — scripts 15 · logic-map **216**(skip 1) · sdd 71 · sdd-history 22 · pm 16 · deploy 18 |
| `make lint` · `make test-seams` | rc=0 · rc=0 |
| `openspec validate a122 --strict` | valid (spec delta: 조언 요구 + 시나리오 하나 추가, AND 절 수정) |

### 안 한 것

- **걸어야 아는 거절의 예측** — 위 설계 결정. 15건 전부 아카이브이고 명령이 스스로 이유를 말한다.
- **`_commits_after` 의 `TimeoutExpired`** — 조언 줄의 창 크기를 재는 git 호출이 `rc≠0` 만 받고 멎음은 안 받는다.
  7.7 이 만든 것이 아니고 I8 도 아니다. 이번에 넣은 판정 호출만 자기 경계를 갖는다.
- **빌리는 쪽 문장에 빌려주는 id 적기** — 문장은 옛 것 그대로다(`function-logic-reference.txt` 에 있다).
- `tools/logic-map/README.md` · `docs/WORKFLOW.md` — 7.9 의 몫(7.2 가 규칙을 정한 뒤).

## Pre-Edit Gate — task 6.5 (자기를 담은 커밋에서 이미 틀린 증거) (2026-09-14)

**High-risk 아님** — 게이트 도구(`tools/logic-map`)의 거절 **문장** 하나다. 주문·손절·사이징 경로와 무관하고 판정(rc·기록 여부·값)을 바꾸지 않는다(아래 A/B 로 확인).

### 먼저 다시 쟀다 — 6.5 가 연 "아무 게이트도 못 본다"는 이제 틀리고, 남은 것은 거짓 사유 한 줄이다

6.5 는 6.1.1 측정에서 열렸다: a089 · a095 의 번들이 base 에서는 맞고 번들을 담은 커밋 `a30eb35ae` 에서는 안 맞는다.
그 뒤 6.1.2 · 7.1 · 7.2.1 · 7.3.1 · 7.7 이 들어왔으므로 HEAD `483985dc` 에서 두 경로를 다시 돌렸다.

| 경로 | a089 · a095 에서 오늘 나오는 것 |
|---|---|
| 5단계, 기록 없음(워킹트리 대상) | rc 1 · `AST source hash is stale: internal/journal/outbox.go` 등 · 7.1 조언 줄이 "base 의 소스를 적은 번들(a089 6개 · a095 4개) — 그 번들을 갱신하라"고 **이름으로** 말한다 |
| 착지를 선언 | 규칙 `_landing_refusal` 의 `landing point a30eb35ae6da is not the revision this evidence describes: internal/journal/outbox.go` |
| `--record-landing` | rc 1 · `no commit at or after the evidence (a30eb35ae6da) matches every pinning bundle — **the evidence does not describe any revision on this history**` |

게이트는 이 모양을 **본다**. 남은 결함은 셋째 줄의 꼬리다 — 두 change 의 증거는 base `ec29dc72` 를 정확히 기술한다(base 에서
고정 번들 6/6 · 9/9 일치). 도구가 이 change 에 대해 내는 유일한 사유가 사실과 반대다. spec 은 "계산이 못 찾으면 사유를
이름으로 말해야 한다(SHALL)"를 요구한다.

### 걷기 실패 전수 — 꼬리가 참인가 (`65_census.py` · `65_prefloor.py`, HEAD 코드로 읽기만)

7.6 의 `compute_landing` 전수(93건)에서 값을 못 낸 것은 **17건**이고 17건 전부 같은 꼬리로 끝난다. 순회는 하한(`floor`)부터 HEAD
까지만 걷는다. 꼬리가 참이려면 그 **아래**(`base..floor`, base 포함)에서도 전부 맞는 커밋이 없어야 한다. 그 구간을 `git cat-file --batch`
한 프로세스로 전부 쟀다(바이트 등식만 — 규칙 함수는 안 부른다).

| change | 번들 | 하한 (커밋 제목) | base 에서 전부 맞음 | 하한 아래 전부 맞는 커밋 / 구간 | 하한에서 틀린 소스 |
|---|---:|---|:---:|---:|---|
| a089-an-unserved-stop-is-counted | 6 | `a30eb35ae6da` fix(safety): bound alerts and plan exit  | **예** | **1** / 1 | `internal/journal/outbox.go` |
| a095-a-stop-must-know-what-it-covers | 9 | `a30eb35ae6da` fix(safety): bound alerts and plan exit  | **예** | **1** / 1 | `internal/obs/notifier.go` |
| 2026-07-31-automate-soak-openapi-onboarding | 13 | `d42e42e91717` chore(openspec): renumber operator chang | 아니오 | 0 / 160 | `cmd/tossctl/console.go`, `cmd/tossctl/soakproc.go` 외 4 |
| 2026-07-31-console-adoption-controls | 16 | `d42e42e91717` chore(openspec): renumber operator chang | 아니오 | 0 / 226 | `cmd/tossctl/console.go`, `internal/app/engine/adoption.go` 외 7 |
| 2026-07-31-fix-console-referrer-origin | 4 | `d42e42e91717` chore(openspec): renumber operator chang | 아니오 | 0 / 162 | `internal/console/pages.go`, `internal/console/remote.go` 외 1 |
| 2026-07-31-streamline-trading-views | 4 | `d42e42e91717` chore(openspec): renumber operator chang | 아니오 | 0 / 156 | `internal/console/portfolio_label_test.go`, `internal/console/portfolio_test.go` 외 2 |
| 2026-08-01-a042-persist-exit-line-snapshots | 28 | `d42e42e91717` chore(openspec): renumber operator chang | 아니오 | 0 / 150 | `internal/app/engine/exitloop.go`, `internal/exitpolicy/recovery.go` 외 7 |
| 2026-08-02-a054-console-status-shell | 53 | `d42e42e91717` chore(openspec): renumber operator chang | 아니오 | 0 / 64 | `cmd/tossctl/console.go`, `internal/console/console.go` 외 11 |
| 2026-08-02-a055-console-settings-cadence | 88 | `d42e42e91717` chore(openspec): renumber operator chang | 아니오 | 0 / 59 | `cmd/tossctl/adoptionsettings_stop_percent_test.go`, `cmd/tossctl/console.go` 외 7 |
| 2026-08-08-console-system-update | 18 | `93165f969c40` feat(runtime): Guardian·복구·서명 업데이트·GBrai | 아니오 | 0 / 1 | `cmd/tossctl/console.go`, `internal/console/console.go` 외 2 |
| 2026-08-08-wire-production-risk-guardian | 18 | `93165f969c40` feat(runtime): Guardian·복구·서명 업데이트·GBrai | 아니오 | 0 / 1 | `cmd/tossctl/console.go`, `internal/console/console.go` 외 2 |
| 2026-08-29-add-common-exit-optimization | 36 | `c06192799fa9` feat(tossos): ship automated portfolio o | 아니오 | 0 / 1 | `cmd/tossctl/console.go`, `internal/config/engine.go` 외 2 |
| 2026-08-29-console-click-approval | 27 | `3a2bc1481999` chore(sdd): Function Logic Map ast.json  | 아니오 | **1** / 6 | `cmd/tossctl/console.go`, `internal/console/console_test.go` 외 3 |
| 2026-08-29-enable-engine-autostart-menu | 66 | `c06192799fa9` feat(tossos): ship automated portfolio o | 아니오 | 0 / 1 | `cmd/tossctl/console.go`, `cmd/tossctl/console_test.go` 외 1 |
| 2026-08-29-enable-vpn-console-access | 54 | `c06192799fa9` feat(tossos): ship automated portfolio o | 아니오 | 0 / 1 | `internal/config/engine.go` |
| 2026-08-29-verify-us-market | 54 | `3a2bc1481999` chore(sdd): Function Logic Map ast.json  | 아니오 | **2** / 5 | `internal/verifylive/fake_broker_test.go`, `internal/verifylive/mutate.go` |
| 2026-09-07-a065-add-position-campaign-leg-core | 12 | `b55fff17253a` fix(a065): 같은 규칙의 두 번째 사본을 없애고 6.4 리뷰 결함 | 아니오 | 0 / 320 | `internal/journal/apply_hook.go`, `internal/journal/readonly.go` 외 2 |

- **꼬리가 실측으로 거짓: 4/17** — a089 · a095(base 자체) · console-click-approval · verify-us-market(하한 아래 커밋).
- 나머지 13건은 `base..HEAD` 에서는 참이지만 base **아래**는 안 쟀다 — 문장은 "이 역사의 어느 리비전도"라고 말하므로 그 13건에서도 **증명되지 않은 주장**이다.
- 17/17 에서 순회의 첫 후보가 하한 자신이다(`start_is_floor`). 첫 후보의 거절은 17/17 이 규칙의 불일치 문장이고, a089 · a095 에서
  그 문장이 가리키는 파일이 6.1.1 표의 "번들 커밋에서 안 맞는 파일"과 **같다**(`internal/journal/outbox.go` · `internal/obs/notifier.go`).

**관찰 — 결정 아님 (7.2 의 몫).** console-click-approval · verify-us-market 의 하한은 `3a2bc148 chore(sdd): Function Logic Map ast.json
경로를 저장소 상대 경로로` 다. 그 커밋은 두 change 의 번들 27 · 54 개 **전부**에서 `file` 한 줄만 절대→상대로 바꿨고 `source_sha256` 은
그대로다(`git show` 실측). `--diff-filter=MAT` 는 이것을 증거가 새로 들어온 것으로 세어 하한을 작업 뒤로 올린다. 하한만 문제도 아니다 —
7.2.1 의 `_unheld_bundles` 는 착지 커밋의 `ast.json` 바이트가 오늘의 것과 같기를 요구하므로 경로를 고치기 전 커밋은 그쪽에서도 막힌다.
기계적인 증거 재작성이 이미 착지한 아카이브의 착지를 없앤다는 뜻이고, 무엇에 착지를 묶을지(7.2)가 답할 질문이다. d42e42e9(renumber)
일곱은 하한 아래에서도 전부 맞는 커밋이 0 이라 같은 모양인지 이 측정으로는 말할 수 없다.

### 편집 전 열거 (HEAD `483985dc` blob → `ast.before-6.5.json`)

`enumerate.py` 로 편집 **전에** 뽑았다 — `ast.after-7.7.json` 과 내용(분기·반환·raise·호출·sha)이 같음을 확인했다. 반환 5 중 마지막이 이 문장이다:
`('', f'no commit at or after the evidence ({floor[:12]}) matches every pinning bundle — the evidence does not describe any revision on this history')`.
호출자(AST 로 `tools/**/*.py` 전부): `resolve_landing`(7.3.1 등식 — 기록이 규칙을 통과하면 순회가 그 기록 이하에서 값을 내므로 이 문장에 닿지 않는다고 **추론**했다, 시험으로 재지 않았다) · `record_landing` · 시험 셋(호출 자리 넷).

### 설계 결정

1. **꼬리를 지우고 걸은 것만 말한다.** "no commit at or after the evidence (…) matches every pinning bundle" 는 순회가 실제로 잰 것이라 남긴다(7.7 의 `test_a_refusal_only_the_walk_finds_is_not_promised_away` 가 이 머리를 바늘로 쓴다).
2. **무엇이 틀렸는지는 규칙의 문장을 인용한다 — 새로 짓지 않는다.** 순회가 이미 받은 첫 거절 `refusal` 을 남겨 붙인다. 새 호출 0. 사유를 여기서 지으면 규칙이 문장을 바꿀 때 이 사유만 옛 말로 남는다([[two-judgements-cover-for-each-other]]).
3. **첫 후보다, 마지막이 아니다.** 첫 후보는 17/17 이 증거가 들어온 커밋이고 거기서 틀린 소스가 저자가 고칠 번들이다. 뒤 후보의 문장에는 이웃이 나중에 고친 파일이 붙는다 — 시험 픽스처가 그 차이를 만든다.
4. **"base 를 기술한다"를 새로 말하지 않는다.** 하한 아래를 걸어서 말할 수는 있지만 walk 가 늘고(7.5 가 이미 느리다고 연 자리) 게이트의 기록 없는 경로(7.1 조언)가 이미 그 사실을 이름으로 말한다.
5. `unheld` 갈래와 그 문장은 안 건드린다 — 이 결함이 아니다.

## VERIFY — task 6.5 (2026-09-14)

### RED → GREEN

시험 둘을 더하고 하나의 바늘을 고쳤다. 편집 전 코드에서 셋 중 **둘이 빨갛다**(`65_red.log`).

| 시험 | 편집 전 | 편집 후 | 재는 것 |
|---|---|---|---|
| `EvidenceAlreadyWrongInTheCommitThatHoldsItIsNamed.test_the_recorder_names_what_already_differs_where_the_evidence_entered` | **FAIL** (옛 꼬리) | OK | 6.5 모양 그대로 — 증거를 뽑은 뒤 Go 를 한 번 더 고치고 **한 커밋**에 담는다. 기록 명령의 출력 한 줄을 **글자 그대로** 단언한다. 두 번째 후보 X 에서만 틀리는 파일을 둬서 첫/마지막 후보의 문장이 갈린다 |
| `…test_the_gate_names_it_with_and_without_a_record` | OK | OK | 6.5 의 "아무 게이트도 못 본다"가 오늘 틀렸음을 못 박는다 — 워킹트리 대상은 `AST source hash is stale: internal/own.go`, 그 커밋을 선언하면 규칙의 불일치 문장 |
| `TheGateRecordsTheLandingInsteadOfTheAuthor.test_the_recorder_refuses_the_forgery_this_class_is_named_after` | **FAIL** | OK | 바늘이 옛 꼬리(`does not describe any revision`)였다 — **거짓 문장을 못 박고 있었다**. 이 픽스처에서 증거는 base P 를 정확히 기술한다(`_pinning_at(P) == (1, [])` 를 같이 단언) |

### 편집 전후 열거 (`ast.before-6.5.json` → `ast.after-6.5.json`)

바뀐 함수는 `compute_landing` **하나**다(HEAD 와 워킹트리의 모든 함수 AST 덤프를 대조). 분기 7 → 8(`first or refusal`) · 반환 5 → 5
(마지막 반환의 문장만) · raise 0 → 0 · 호출 변화 0. 표는 `tools-logic-map--compute_landing/function-logic-map.md` `## 편집 — task 6.5`.

### 뮤테이션 (`65_mut.py`, 사본 `65_work` — 시작 sha 단언, 무변이 대조군 먼저)

| 변이 | 결과 | 빨갛게 한 시험 |
|---|---|---|
| PC 무변이 | **GREEN** | — |
| N1 옛 꼬리로 되돌림 | **CAUGHT** | 6.5 기록 시험 · 위조 기록 시험 |
| N2 마지막 후보의 문장을 인용(`first = refusal`) | **CAUGHT** | 6.5 기록 시험 **하나** — 위조 픽스처는 후보가 하나라 못 가른다 |
| N3 인용을 뺌 | **CAUGHT** | 둘 다 |
| N4 넘긴 이름이 있을 때만 남김 | **CAUGHT** | 둘 다 |

**계측 사고 한 건 — 대조군이 잡았다.** 첫 실행에서 무변이 사본이 **빨갰고** 변이 넷이 전부 "CAUGHT" 로 찍혔다. 시험 파일이
`parents[1] / "sdd"` 에서 `sdd_doctor` 를 import 하는데 사본에 그 형제 디렉터리가 없어 ImportError 였다 — 변이가 아니라 import 가
빨갛게 한 것이다. `sdd_doctor.py` 를 사본에 넣고, 하네스가 **대조군이 GREEN 이 아니면 변이를 걸지 않고 멈추게** 고친 뒤 다시 쟀다
(위 표). 저장소 파일은 두 실행 내내 sha `f3fd18cae5ea` 로 그대로다.

옛 하네스를 새 코드에 **다시** 걸었다(전부 사본 대상 — 7.6 하네스는 저장소를 직접 겨누던 것을 경로만 바꾼 `65_mut76.py`):

| 하네스 | 결과 |
|---|---|
| 7.6 R1~R16 | R1~R13 · R15 · R16 **CAUGHT**, R14 SKIP(7.7 과 같다) — 옮긴 자리의 R14'(`65_r14p.py`) **CAUGHT** |
| 7.7 R1~R22 (`77_mut.py 65_work`) | **22 전부 CAUGHT** |
| 7.8 M·PC (`65_mut78.py`) | M2 · M3 · M4 · M5 · M7 · M8 · M9 · PC2 **CAUGHT**, M1 · M6 · M11 · PC1 SKIP(7.7 과 같은 넷) |

바늘을 바꾼 위조 기록 시험이 **혼자** 잡던 옛 변이는 없었다(7.8 의 M1 은 두 시험이 잡았고 7.6 에서 자리가 사라져 SKIP).

### 실물 A/B (`65_ab.py`) — 93건 전수, HEAD `483985dc` blob 대 워킹트리

| | 결과 |
|---|---|
| 값·사유 둘 다 같음 | **76** (착지를 얻는 76건 전부) |
| 사유 꼬리만 다름 | **17** — 새 꼬리가 `65_census.json`(편집 **전** 코드로 잰 첫 후보 거절)과 **글자 그대로** 같다 |
| 그 밖 | **0** |

a089 → `no commit at or after the evidence (a30eb35ae6da) matches every pinning bundle — at the first commit walked, landing point a30eb35ae6da is not the revision this evidence describes: internal/journal/outbox.go`
a095 → `… internal/obs/notifier.go`. 6.1.1 이 손으로 짚은 두 파일과 같다.

시간은 적지 않는다 — 6 프로세스가 동시에 걸어서 경합 아래 잰 값이다. 편집은 새 호출이 0 이다.

### 게이트 (2026-09-14, 이 로트의 워킹트리)

| 게이트 | 결과 |
|---|---|
| `test_check_analysis` | 141 → **143** (skip 1) |
| `make sdd-test` | rc=0 — scripts 15 · logic-map **218**(skip 1) · sdd 71 · sdd-history 22 · pm 16 · deploy 18 |
| `make lint` · `make test-seams` | rc=0 · rc=0 |
| `openspec validate --all --strict` | 58/58 (spec delta: 계산 실패 사유 SHALL 두 문장 + 시나리오 하나) |

### 안 한 것

- **"증거가 base 를 기술한다"고 말하기** — 하한 아래를 걸어야 한다(walk 가 는다, 7.5). 기록 없는 5단계가 이미 7.1 조언으로 그것을 이름으로 말한다.
- **기계적 증거 재작성이 하한을 올리는 것**(`3a2bc148` · 위 관찰) — 착지를 무엇에 묶을지의 질문이라 7.2 로 넘긴다. 7.2 에 한 줄 적었다.
- `unheld` 사유 문장 — 이 결함이 아니다.
- `tools/logic-map/README.md` · `docs/WORKFLOW.md` — 7.9.

## DECIDE — task 7.2.2 (C1 · C2 · C4 · H7) (2026-09-14)

사람에게 **정할 것 하나**를 물었다: 착지로 창을 좁히는 것을, 증거로 증명되지 않는 경우에도 허용할 것인가. 물을 때
V1 · V2 를 HEAD `9692b8d1` 에서 다시 재현했다(둘 다 기록 rc 0 → 5단계 `[]` required 0, `internal/own.go` 가 바뀌었는데).

| 선택지 | 무엇 | 비용 |
|---|---|---|
| **1 기계로 막는다** | 착지는 고정 소스 중 하나 이상이 base 와 다른 커밋이어야 한다(7.2 측정의 K2) | 7.2 측정으로 착지 76 중 **8 상실** — base 가 작업 뒤로 옮겨진 change 들 |
| 2 사람의 확인을 조건으로 | 도구가 계산하고 review.md 에 사람이 값·근거를 적어야 받는다 | 게이트는 문장의 **존재**만 본다 — 저자가 쓴 증거로 저자의 선택을 검증하는 순환 |
| 3 한계로 인정 | spec 에 알려진 구멍으로 적고 4.5 를 푼다 | FLM-first 기본 경로에서 번들 갱신을 잊으면 조용한 초록 |

**사람이 1 을 골랐다**(부수 결정 포함: C4 는 빌리는 change 가 좁히지 못하게 — 7.2.3, H7 은 한계로 기록 — 7.2.4).
이 절의 나머지는 1 의 구현(7.2.2)이다.

## Pre-Edit Gate — task 7.2.2 (착지는 고정 소스를 바꾼 커밋이어야 한다) (2026-09-14)

**High-risk 아님** — 게이트 도구의 수락 규칙이다. 거래 경로 무관. 다만 게이트가 **받는 것을 줄이는** 편집이라
거부할 정상 입력을 먼저 쟀다([[fail-closed-must-name-what-it-rejects]]).

### 편집 전 전수 — 규칙을 **사본**에 넣고 93건에 댔다 (`722_census.py`)

저장소 코드(HEAD = 워킹트리)를 기준으로, K2 를 넣은 사본 `722_k2` 의 `compute_landing` 을 같은 입력에 댔다. 같은 실행에서
더 엄격한 변형도 쟀다 — 착지에서 base 와 같은 고정 소스가 **착지 뒤에** 바뀌면 거절(아래 H3 를 막는 판본).

| 규칙 | 같음 | 착지 상실 | 이동 | 획득 |
|---|---:|---:|---:|---:|
| **K2 (하나 이상 바뀜)** — 사람이 고른 것 | 85 | **8** | 0 | 0 |
| K2 + 착지 뒤에 바뀌는 base 모양 소스 거절 | 79 | **14** | 0 | 0 |

- K2 의 상실 8 은 7.2 측정의 이름과 **같다**: 활성 a074 · a077 · a079 · a091 · a092 · a094, 아카이브 a078 · a075. 첫 후보(하한)에서
  고정 소스가 전부 base 와 같고 뒤 후보는 증거와 안 맞는다.
- 엄격한 변형의 추가 6: **a100(활성)** · a047 · a049 · a050 · a052 · a060(아카이브). 결정 범위 밖이라 넣지 않는다(아래 잔여).
- 착지를 원래 못 얻던 17건은 두 규칙에서 값이 그대로다(더 거절만 하는 규칙이므로).

### 기존 시험 중 깨지는 것 — 둘, 둘 다 대가를 **그대로** 적은 시험

K2 사본으로 스위트를 돌렸다(`722_suite_any.log`): 147 중 실패 **2**.

| 시험 | 주석이 적은 모양 | 처리 |
|---|---|---|
| `AnEmptyRequiredSetIsAnnouncedNotSwallowed.test_empty_required_with_bundles_is_reported_and_still_passes` | "작업이 base **앞**에 있다 — 2026-08-04 재기준화가 a074~a079 를 그 모양으로" | 시험의 뜻(요구 0 을 삼키지 않고 말한다)은 두고 픽스처를 K2 를 통과하는 모양으로 — 함수에서 떨어진 주석 한 줄만 바꾼 착지(요구 0). a074 모양은 새 클래스로 옮겼다. 첫 판(파일 끝에 새 함수)은 `--unified=0` 조각 경계가 `Own` 에 닿아 요구 1 이었다 |
| `TheGateRecordsTheLandingInsteadOfTheAuthor.test_it_does_not_record_the_base_even_when_the_base_matches` | 증거가 base 를 기술 | 이제 E 도 기록하지 않는다 — rc 1 · 파일 없음 · K2 문장 |

### 남는 구멍 — K2 사본으로 픽스처 셋 (`722_holes.py`)

| 모양 | K2 뒤 |
|---|---|
| H1 6.6 의 미끼 번들(안 건드린 파일) → 기록 → 뒤 작업 | **막힌다** — 기록 rc 1, 5단계 요구 1 · 증거 없음 |
| H2 소스가 바뀐 정직한 번들로 기록 → **뒤에** 증거 없이 다른 파일 작업 | **열림** — 5단계 초록, 뒤 작업은 창 밖 |
| H3 FLM 먼저 둘 → 하나만 편집·갱신 → 다른 파일은 나중에 편집(갱신 안 함) | **열림** — 기록 rc 0 · 초록 |

H2 는 창을 좁히는 한 본질이다 — 뒤의 작업이 이 change 의 것인지 이웃의 것인지 내용으로 못 가른다. H3 는 위 엄격한 변형이
막는다(상실 +6). 둘 다 이 결정이 닫지 않는다.

### 편집 전 열거 (HEAD `9692b8d1` blob → `ast.before-7.2.2.json`)

`_landing_refusal`(분기 6 · 반환 7) · `compute_landing`(분기 8 · 반환 5) · `main`(권유 문장) · `resolve_landing`(규칙을 부르는 자리,
내부 불변 확인용). 규칙의 거절 순서는 base 조상 → 고정 0 → 불일치 → 하한 없음 → 하한 순서 → 미보유이고, 수락 반환은 L652 `('', [])` 하나다.

### 설계 결정

1. **`_landing_refusal` 의 맨 뒤**(미보유 뒤, `resolve_landing` 의 계산값 대조 앞). 앞의 가드들은 각자 자기 문장으로 못 박혀 있다 —
   하한 앞으로 옮기면 그 가드들의 거절 지점을 가로챈다(변이 K-D 가 다섯 시험으로 확인).
2. **파일 단위 · 하나 이상.** 7.2 측정과 같은 정의다. 전부(K2-all)는 25 상실이라 사람이 고른 규칙이 아니다.
3. **계산 실패의 머리를 고친다** — "matches every pinning bundle" 는 K2 가 증거와 **맞는데** 거절한 후보를 만들므로 거짓이 된다
   ([[a-not-found-reason-claims-only-what-was-searched]]). "is accepted as the landing". 조언 줄의 권유 문장도 같은 조건으로.
4. 선언 경로와 계산 경로는 이미 한 규칙에 묻는다(7.6) — 새 사본을 만들지 않는다.

## VERIFY — task 7.2.2 (2026-09-14)

### RED → GREEN

새 클래스 `ALandingMustChangeWhatItsEvidencePins`(시험 5)를 쓰고 기존 시험 다섯의 기대를 바꿨다. 편집 전 코드에서 고른 23 중
**8 이 빨갛다**(`722_red.log`) — 새 넷(V1 기록 · V1 선언 · V2 기록 · a074 모양) · base 가 맞는 픽스처 · 계산 실패 머리를 보는 셋.
빈 요구 집합 알림 시험은 픽스처만 바꿔서 편집 전에도 초록이다(뜻이 규칙과 무관하므로 그래야 한다).
시험 143 → **148**. GREEN 은 먼저 사본에 넣어 전체 스위트를 돌린 뒤 같은 바이트를 저장소에 넣었다(`cmp` 동일).

### 편집 전후 열거

바뀐 함수는 `_landing_refusal` · `compute_landing` · `main` 셋(모든 함수 AST 덤프 대조). `_landing_refusal` 분기 6 → 10 · 반환 7 → 8,
`compute_landing` · `main` 은 문장 하나씩, `resolve_landing` 은 AST 동일. 맨 위 주석 "base 와 같아도 된다 — a074·a079·a075 의 정답"은
틀린 말이 되어 고쳤다(AST 동일 확인 뒤 열거를 다시 뽑았다). 표는 각 `function-logic-map.md` 의 `## 편집 — task 7.2.2`.

### 뮤테이션 (`722_mut.py`, 사본 `722_work`, 무변이 대조군 GREEN)

| 변이 | 1차 | 최종 | 잡은 시험 |
|---|---|---|---|
| K-A K2 가드 삭제 | CAUGHT | **CAUGHT** | 다섯 |
| K-B `all` → `any` | **SURVIVED** | **CAUGHT** | `test_one_changed_pinned_source_is_enough` 하나 |
| K-C base 대신 HEAD | CAUGHT | **CAUGHT** | 25 |
| K-D 가드를 하한 앞으로 | CAUGHT | **CAUGHT** | 다섯(하한 · 미보유 시험) |
| K-E 계산 실패 머리 되돌림 | CAUGHT | **CAUGHT** | 넷 |
| K-F 권유 문장 되돌림 | **SURVIVED** | **CAUGHT** | 7.7 의 walk 시험(바늘 추가) |
| K-G 고정 소스를 첫 번들 하나로 | **SURVIVED** | **CAUGHT** | `test_one_changed_pinned_source_is_enough` 하나 |

K-B · K-G 는 "하나 이상"을 재는 픽스처가 없어서 살았다 — 사람이 고른 K2(8 상실)와 K2-all(25 상실)을 가르는 바로 그 자리다.
시험을 더한 **둘째 판에도 둘 다 살았다**: 읽기 전용 파일을 base **뒤에** 만들어서 base 에 없었고, 그러면 "바뀐 소스"로 세어진다.
파일을 base 에 넣고 "그 소스가 base 와 착지에서 같다"는 도달 단언을 시험 안에 세운 셋째 판에서 잡혔다
([[mutation-must-reach-the-thing-under-test]]).

옛 하네스를 새 코드에 **다시** 걸었다(전부 사본): 7.6 R1~R16 은 이전과 같다(R14 SKIP, 옮긴 자리 R14' **CAUGHT**) ·
7.7 R1~R22 는 R20 만 SKIP(권유 문장을 이번에 바꿨다) — 새 자리에 건 R20' **CAUGHT** · 7.8 CAUGHT 8 · SKIP 4(이전과 같은 넷) ·
6.5 N1~N4 **CAUGHT**. 새 가드가 **뒤에서** 가려 살아남게 된 옛 변이는 0 이다. 사본 sha 는 전 구간 `87275a3a619b` 로 복원됐다.

### 실물 A/B (`722_ab.py`) — HEAD `9692b8d1` blob 대 워킹트리

| | 결과 |
|---|---|
| 착지 93 — 값 · 사유 같음 | **68** |
| 머리 문장만 다름(`matches every pinning bundle` → `is accepted as the landing`) | **17** (원래 착지를 못 얻던 건) |
| 예측한 K2 상실 | **8/8** — 사유가 K2 문장을 인용 |
| 그 밖 | **0** |
| 판정 — a099(착지 기록) | 오류 0 → 0 · 착지 `21a315d1` 그대로 · required 32 → 32 · **IDENTICAL** |
| 판정 — 상실 8 중 활성 6 | 전부 **IDENTICAL**(오류 257~324, required 255~319) — 기록을 쓴 적이 없어서 오늘의 판정은 안 바뀐다. 잃은 것은 **기록할 수 있는 가능성**이다 |

리뷰의 재현 스크립트(`rv_verify.py`)를 새 코드로: **V1 · V2 둘 다 기록 rc 1**, 5단계는 `required 1` 로 빨간 채 남는다(편집 전: rc 0 · `[]` · required 0).

### 게이트 (2026-09-14, 이 로트의 워킹트리)

| 게이트 | 결과 |
|---|---|
| `test_check_analysis` | 143 → **148** (skip 1) |
| `make sdd-test` | rc=0 — logic-map **223**(skip 1) · scripts 15 · sdd 71 · sdd-history 22 · pm 16 · deploy 18 |
| `make lint` · `make test-seams` | rc=0 · rc=0 |
| `openspec validate --all --strict` | 58/58 (spec: 요구 한 문단 + 시나리오 셋, 6.5 문장의 조건 정정) |

### 남은 것 — 이 결정이 닫지 않은 것

- **H2 기록 뒤의 작업** — 소스가 바뀐 정직한 번들로 기록한 뒤 증거 없이 붙인 작업은 창 밖이다(6.6 에 기록). 좁히는 한 본질.
- **H3 섞인 V1** — 번들 일부만 갱신하고 다른 파일을 나중에 고치면 초록. 막는 변형은 상실 8 → **14**(+a100 활성)라 **사람 결정**으로 남긴다.
- **a074 · a077 · a079 · a091 · a092 · a094 의 5단계** — 이제 착지로 풀 수 없다. 도구에 면제 경로는 없고, a075 · a076 이 2026-09-08 에 쓴 사람의 면제 기록이 선례다.
- C4(7.2.3) · H7(7.2.4) — 다음 task.
- 7.5(성능) · 7.9(문서) · 6.1.2.6(자동 기록 여부) — 이제 착수할 수 있다.

## Pre-Edit Gate — task 7.2.3 (빌리는 change 는 창을 좁히지 않는다) (2026-09-14)

**High-risk 아님** — 게이트 도구의 수락 규칙. 게이트가 받는 것을 줄이는 편집이라 거부할 정상 입력을 먼저 쟀다.

### 거부할 정상 입력

`function-logic-reference.txt` 전수: **1건** — a073(아카이브) → a072(아카이브). 둘 다 `landed-commit.txt` 가 없다. 새 규칙이 거절하는 것은
빌리는 쪽의 **기록**뿐이므로 오늘 저장소의 판정 변화는 0 으로 예측했다(아래 A/B 로 확인). 잃는 것은 1.8 이 a073 에 열어 둔 수리 경로
(양쪽이 같은 착지를 기록해 오류 336 → 0)이고 쓰인 적이 없다.

### 편집 전 열거 (HEAD `67d06bc9` blob → `ast.before-7.2.3.json`)

`check` 의 빌림 갈래: base 공유 판정 → **착지 공유 판정**(`_declared_landing` 두 번 비교, 해독 실패는 `cannot derive …`) → 빌린 analysis 로 교체.
`_recording_refusal` 의 빌림 갈래: "copy the landing recorded on the change that owns the bundles" — 리뷰 C4 의 구멍을 권하는 문장.

### 설계 결정

1. **빌리는 쪽 기록이 있는가만 묻는다**(`_landing_record`) — 해독하면 못 읽는 기록 앞에서 질문이 터진다(6.2.1).
2. **빌려주는 쪽 기록은 이 창에 안 쓴다** — 뒤의 `resolve_landing(change_dir, …)` 가 빌리는 쪽 디렉터리를 읽으므로 코드를 더하지 않아도 된다.
   그 배관이 곧 규칙이라 변이(C-E)로 못 박는다.
3. **문장 한 벌** — `BORROWED_REFUSES_A_LANDING` 을 5단계와 기록 명령이 같이 쓴다(이관 문장이 두 벌에서 갈린 선례, 리뷰 I5).
4. base 공유는 그대로.

## VERIFY — task 7.2.3 (2026-09-14)

### RED → GREEN

`ABorrowedWindowIsSharedAtBothEnds`(5) → `ABorrowedWindowIsNeverNarrowed`(5)로 바꾸고, `ADeclaredLandingMustBePinnedByEvidence` 의 빌림 시험 둘을
하나(`test_borrowed_evidence_no_longer_pins_a_landing`)로 줄였다 — "빌린 번들이 기술하지 않는 착지는 거절한다"는 이제 빌림 거절이 먼저 막아서
재는 것이 없다. 편집 전: C4 시험 **FAIL**(공유 판정 문장), 새 문장 상수를 쓰는 다섯 **ERROR**. 편집 뒤 스위트 148 → **147** 전부 초록.

### 편집 전후 열거

바뀐 함수 `check` · `_recording_refusal` 둘(모든 함수 AST 대조). `check`: `try` · `except ValueError` · `if not shared` 가 사라지고
`if _landing_record(change_dir, root) is not None` 하나, 호출 `_declared_landing` 사라짐. 표는 각 FLM 의 `## 편집 — task 7.2.3`.

### 뮤테이션 (`723_mut.py`, 사본 `723_work`, 대조군 GREEN) — 첫 판에 전부 CAUGHT

| 변이 | 결과 | 잡은 시험 |
|---|---|---|
| C-A 빌리는 쪽 기록 거절 삭제 | CAUGHT | 넷 |
| C-B 있는가 대신 해독(`_declared_landing`) | CAUGHT | `test_an_undecodable_borrower_record_is_refused_not_raised`(traceback) |
| C-C 빌려주는 쪽 기록을 봄 | CAUGHT | 셋 |
| C-D 기록 명령 문장 되돌림 | CAUGHT | `test_it_refuses_a_change_that_borrows_its_evidence` |
| **C-E 빌려주는 쪽 착지를 물려받음 (C4 구멍 자체)** | CAUGHT | `test_a_lender_record_does_not_narrow_the_borrower` |
| R5' 기록 명령의 빌림 거절 삭제(7.7 R5 의 옮긴 자리) | CAUGHT | 둘 |

옛 하네스를 새 코드에 다시 걸었다(전부 사본 `723_work`): 7.6 CAUGHT 15 · R14 SKIP(옮긴 자리 R14' **CAUGHT**) · 7.8 CAUGHT 8 · SKIP 4(이전과 같다) ·
6.5 N1~N4 **CAUGHT** · 7.2.2 K-A~K-G **CAUGHT** · 옮긴 자리 R20' **CAUGHT**. 7.7 하네스는 **R10 에서 멈췄다** — 순서 변이가 옛 빌림 문장을
치환 기준으로 삼아 단언이 깨졌다(쓰기 전이라 사본은 그대로, sha `5e4676c93884`). 그래서 R11~R22 는 필터로 따로 돌렸고(**R20 만 SKIP**, 나머지
CAUGHT), R5(빌림 거절 삭제)는 SKIP → 옮긴 자리 R5' **CAUGHT**, R10 은 새 문장 자리에 R10'(dirty 를 빌림 앞으로)을 걸어 **CAUGHT**.
새 규칙 뒤에 가려져 살아남게 된 옛 변이는 0 이다.

### 게이트 (2026-09-14)

| 게이트 | 결과 |
|---|---|
| `test_check_analysis` | 148 → **147** (skip 1) — 빌림 시험 둘을 하나로 줄였다 |
| `make sdd-test` | rc=0 — logic-map **222**(skip 1) · scripts 15 · sdd 71 · sdd-history 22 · pm 16 · deploy 18 |
| `make lint` · `make test-seams` | rc=0 · rc=0 |
| `openspec validate --all --strict` | 58/58 (spec: 공유 요구를 "좁히지 않는다"로, 시나리오 둘 교체, 3.2.3.1 의 빌린 고정 문장 제거) |

### 안 한 것

- review.md `## Pre-Edit Gate — task 1.8` 은 결정 당시 기록으로 두었다 — tasks.md 1.8 에 뒤집혔다는 표시를 달았다.
- 5.5(빌리는 쪽 자기 작업이 고정되지 않는다)는 이 규칙으로 닫혀 체크했다.

### 실물 A/B (`723_ab.py`) — HEAD `67d06bc9` 대 워킹트리

| | 결과 |
|---|---|
| a073(빌리는 쪽) 판정 | 오류 336 → 336 · 착지 없음 · required 388 → 388 · **IDENTICAL** |
| a072(빌려주는 쪽) 판정 | 오류 336 → 336 · required 388 → 388 · **IDENTICAL** |
| a073 기록 명령 사유 | "copy the landing recorded on …" → `BORROWED_REFUSES_A_LANDING` |

## Pre-Edit Gate — task 7.2.4 (H7: 병합 안에서만 들어온 증거) (2026-09-14)

**동작은 바꾸지 않는다** — 사람이 2026-09-14 에 H7 을 한계로 두기로 골랐다. 고친 것은 그 한계를 말하는 **문장** 둘이다.

### 지금 코드에서 재현했다 (`724_h7.py`, HEAD `f9811236`)

P(base) → 변경 디렉터리 → 곁가지(무관한 편집) · 줄기에 W(작업) → `git merge --no-commit` 뒤 번들을 추가해 병합 커밋 M 으로 커밋.

| 무엇 | 결과 |
|---|---|
| `_evidence_floor` | `''` — `git log --diff-filter=MAT`(`-m` 없음)가 병합 커밋의 변경을 안 읽는다 |
| `record_landing` | rc 1 · `no landing recorded — the pinning evidence never entered this history (commit the bundles)` — **거짓**(번들은 HEAD 에 커밋돼 있다) |
| `check`, 기록 없음 | `[]` · required 1 — 막히지 않는다. 한계의 크기는 "좁히지 못한다" |
| M 을 손으로 기록 | `… is pinned by evidence that never entered this history: commit …` — 같은 거짓 |

조언 줄(7.7)도 걷기 전 판정에서 같은 문장을 인용하므로, 이미 한 커밋을 또 하라고 권한다 — 7.7 이 닫은 I8(서로를 가리키는 두 문장)과 같은 결함.

### 실물 영향

7.2.2 A/B 의 착지 93 과 7.7 전수 126 에서 이 두 문장이 나온 id **0**. 판정 변화 0 을 예측한다.

### 편집 전 열거

`_walk_floor`(반환 3) · `_landing_refusal`(반환 8)의 하한 없음 반환 문장 둘. 분기 · 호출은 건드리지 않는다.

## VERIFY — task 7.2.4 (2026-09-14)

### RED → GREEN

새 클래스 `EvidenceFirstCommittedInsideAMergeIsAKnownLimit`(시험 3) + 옛 문장을 바늘로 쓰던 시험 셋의 바늘 교체. 편집 전 **5 FAIL**(새 둘 · 바늘
셋), `test_without_a_record_step_five_still_judges_the_working_tree` 는 편집 전에도 초록(한계의 크기를 못 박는 시험이라 그래야 한다).
기록 명령 시험은 **도달 단언을 먼저** 세웠다 — 번들이 HEAD 에 있고 M 의 첫 부모에는 없고 M 이 부모 둘인 병합이다. 실패 원인이 문장인지
확인했다(도달 단언은 통과). 시험 147 → **150**.

### 편집 전후 열거

바뀐 함수 `_walk_floor` · `_landing_refusal` 둘(모든 함수 AST 대조). 둘 다 분기 소스 · 반환 수 · 호출이 **같고** 반환 문장 하나씩만 다르다.

### 뮤테이션 (`724_mut.py`, 사본 `724_work`, 대조군 GREEN)

| 변이 | 결과 | 잡은 시험 |
|---|---|---|
| H-A `_walk_floor` 문장 되돌림 | **CAUGHT** | 셋(기록 명령 한계 시험 · 계산 · 미커밋 기록) |
| H-B 규칙의 하한 없음 문장 되돌림 | **CAUGHT** | 둘(병합 손 기록 · 미커밋 선언) |
| H-C **한계를 없앰**(`git log -m`) | **CAUGHT** | 넷 — 한계가 우연이 아니라 선택임을 시험이 못 박는다 |
| R4' 규칙의 하한 없음 판정 삭제(7.6 R4 의 옮긴 자리) | **CAUGHT** | 둘 |
| R18' `_walk_floor` 의 하한 없음 판정 삭제(7.7 R18 = 7.8 M7 의 옮긴 자리) | **CAUGHT** | 넷 |
| K-D' K2 를 하한 판정 앞으로(7.2.2 K-D 의 옮긴 자리) | **CAUGHT** | 다섯 |

옛 하네스 전부를 새 코드에 다시 걸었다: 7.6 CAUGHT 14 · SKIP R4 · R14(각각 R4' · R14' CAUGHT) · 7.7(R5 · R10 · R20 제외 실행) CAUGHT 18 · SKIP R18(R18'
CAUGHT) · 7.8 CAUGHT 7 · SKIP 5(옛 넷 + M7 → R18') · 6.5 N1~N4 · 7.2.2 K-A~K-G · 7.2.3 C-A~C-E · R5' · R20' · R10' **전부 CAUGHT**. 새로 생긴 SKIP 은 이번에
문장이 옮겨진 자리 넷뿐이고 모두 옮긴 자리 변이가 잡는다. 사본 sha `d9f9bda0a983` 로 복원.

### 게이트

| 게이트 | 결과 |
|---|---|
| `test_check_analysis` | 147 → **150** (skip 1) |
| `make sdd-test` | rc=0 — logic-map **225**(skip 1) · scripts 15 · sdd 71 · sdd-history 22 · pm 16 · deploy 18 |
| `make lint` · `make test-seams` | rc=0 · rc=0 |
| `openspec validate --all --strict` | 58/58 (spec: H7 한계 문단 + 시나리오 하나) |

실물 영향: 두 문장이 나오는 실물 id 0(위) — 판정 A/B 대상이 없다.

### 남은 것

- **H7 동작** — 한계 그대로. 없애는 변이(H-C, `git log -m`)를 시험이 잡으므로 없애려면 이 결정을 다시 해야 한다.
- 7.2 결정이 새로 연 사람 결정: **H3 섞인 V1**(막는 변형 상실 14) · **6.6 의 H2 기록 뒤 작업**(좁히는 한 본질 — 한계로 둘지) ·
  **6.1.2.6 자동 기록 여부**(이제 물을 수 있다).

## MEASURE — task 7.2.6 (H4: 착지 기록 뒤의 자기 수리) (2026-09-16)

**결정 전 측정이다 — 규칙은 아직 없다.** 7.2.5(H3) · 6.6(H2) · 6.1.2.6 의 추천을 쓰려고 픽스처를 재다가 H4 가 나왔다.
도구는 HEAD `0323e0ea` 사본(`cmp` 로 HEAD 와 같음 확인), 계측기는 scratchpad `rec_holes.py` · `rec_holes2.py` · `h4_census.py` ·
`h4_detail.py` · `h4b_census.py` · `h4b_fixture.py`. 저장소는 읽기만 했다.

### 픽스처 — 5단계 `check` 만 잼

| 모양 | 착지 기록 없음 | 착지 기록 있음 |
|---|---|---|
| H2 정직한 기록 뒤, 고정 안 된 파일 작업(증거 없음) | — | `[]` · required 1 **초록** |
| H3 섞인 V1 | — | 기록 rc 0 · `[]` **초록** |
| **H4 기록 뒤 같은 고정 파일을 수리, FLM 안 고침** | `AST source hash is stale` **빨강** | `[]` · required 1 **초록** |
| H4b 수리 + 번들 갱신, 재기록 안 함 | — | `landing point … is not the revision this evidence describes` 빨강 — 복구 경로를 안 말한다 |
| H4c 그 상태에서 `--record-landing` | — | rc 1 `already exists — not overwritten` |
| H4d 기록 삭제 → 재기록 | — | 착지 = 갱신 커밋, `[]` 초록(정상) |

기록이 없으면 잡히는 것을 기록이 가린다. 잊은 쪽은 초록, 성실한 쪽은 빨강이다.

### 내용 규칙 전수 — 고정 소스가 착지와 '지금' 에서 같아야 한다 (`h4_census.py`)

모집단은 오늘 착지를 계산할 수 있는 68건(활성 a066 · a071 · a100 · a112, 아카이브 64). 착지를 못 얻는 25건(K2 상실 8 + 걷기 실패 17)은
더 거절만 하는 규칙이라 값이 안 바뀐다. FILE 은 파일 바이트(도구의 결속 단위 — `source_sha256` 은 파일 해시다), FUNC 는 번들이
가리키는 함수의 본문 줄(`func` 선언~닫는 괄호, 추출기 `--list` 범위). now=HEAD 는 오늘 다시 돌리는 게이트, now=archive 는
아카이브 커밋(정상 흐름에서 게이트가 돈 자리의 근사).

| 규칙 | now | 대상 | 거절 | 활성 거절 |
|---|---|---:|---:|---|
| FILE | HEAD | 68 | **61** | a066 · a071 · a100 |
| FUNC | HEAD | 68 | **55** | a066 · a071 |
| FILE | archive | 63(+적용불가 1) | **45** | — |
| FUNC | archive | 63(+적용불가 1) | **41** | — |

- 7.2 측정의 K6(번들이 HEAD 를 기술, 76 중 69 거절)과 같은 축이고 같은 결론이다.
- **이동은 0 이다(유도).** 착지 후보는 번들 해시와 파일 바이트가 같아야 받으므로, L 을 거절한 규칙은 L 과 바이트가 같은 모든 후보를 같이 거절한다.
- 원인 커밋 분류: 좁은 태그 `[aNNN` 가 자기 id 인 커밋은 a083 · a084(두 now 모두), a061(HEAD 만 — `dd287dbd [a061]` 은 재번호 뒤 번호를
  다시 쓴 **다른** change). 나머지 거절의 원인은 전부 이웃 커밋이다. 넓은 정규식(본문 포함)은 병합 메시지 · 다른 change 인용을 잡아 폐기했다.

**결론: 내용 축으로는 이웃의 편집과 자기 수리를 못 가른다** — 7.2 가 V1 에서 얻은 결론이 H4 에도 그대로다.

### 가르는 신호 — 같은 커밋이 그 change 디렉터리를 만졌는가

실물 H4 셋을 커밋 내용으로 확인했다(`h4_detail.py`): a083 `52b5cb6d` · a084 `e1491071` · `0a072196` · `52b5cb6d` · apply-us-measurement-fixes
`f62457c3`. 셋 다 착지 뒤에 **고정 소스를 고쳤고, 자기 change 디렉터리를 같이 만졌고, 자기 FLM 번들은 안 만졌다**(e1491071 이 만진
FLM 149개는 a085 의 것). 기록이 없던 시절이라 실제 초록은 아니었지만 기록했다면 H4 였다.

변형 B: **착지 후보 C 뒤에(C..now) Go 를 고치면서 그 change 디렉터리(`openspec/changes/<id>/`, 아카이브면 그 경로도)를 같이 만진 비병합 커밋이
있으면 C 를 받지 않는다.** PINNED 는 그 커밋이 고정 소스를 고쳤을 때만, ANYGO 는 아무 `*.go` 나. 걷기는 생산 `compute_landing` 을 옮긴
것이고, B 를 끈 대조군이 생산 값과 **68/68 일치**(두 now 모두)한 뒤에 쟀다.

| 변형 | now | 같음 | 이동 | 상실 |
|---|---|---:|---:|---:|
| PINNED | archive | 59 | 0 | **4** |
| ANYGO | archive | 57 | 2 | **4** |
| PINNED | HEAD | 61 | 0 | **7** |
| ANYGO | HEAD | 59 | 2 | **7** |

- **아카이브 시점 상실 4** = 실물 H4 셋(a083 · a084 · apply-us-measurement-fixes) + 오탐 하나(add-candidate-discovery ← `515c6585` 이웃
  fix-chase-veto-measurement 가 남의 design.md 에 메모를 달면서 고정 소스를 고침).
- **HEAD 에서 더해지는 상실 3** — 전부 여러 change 를 한꺼번에 만진 넓은 커밋이다: 활성 a066 · a071 ← `8022f578`(1068 파일 반입, a066
  status.md · a071 FLM 산문 한 줄), 아카이브 a044 ← `d42e42e9`(419 파일 재번호).
- **이동 2**(ANYGO 만, a045 · a098) — 착지 뒤의 자기 작업이 창에 들어오지만 required 8→8 · 7→7 이라 **판정 변화 0**.
- **오늘의 판정 영향 0**: 기록을 가진 실물은 a099 하나이고 네 칸 모두 '같음'. 나머지는 기록이 없어 창이 워킹트리다.
- 병합 커밋은 안 센다(`--no-merges`) — H7 과 같은 한계.

### B 픽스처 대조 (`h4b_fixture.py`)

| 모양 | PINNED | ANYGO |
|---|---|---|
| H4 수리와 review.md 를 **한 커밋**에 | 착지 없음 → 워킹트리 창(stale 빨강) | 같음 |
| H4 수리 커밋과 문서 커밋을 **쪼갬** | 원래 착지 유지 — **뚫린다** | 뚫린다 |
| H2 고정 안 된 파일 작업 + review.md 한 커밋 | 원래 착지 유지 | 착지가 작업 뒤로 이동 → `Other` 요구(증거 없음 빨강) |
| 이웃이 고정 파일을 고침(디렉터리 안 만짐) | 거절 안 함 | 거절 안 함 |

### B 가 무엇인가 — 사람이 정할 때 볼 것

- **망각 가드이지 위조 가드가 아니다.** 커밋을 쪼개면 뚫린다([[existence-check-is-not-a-role-check]]). 이 저장소의 task 커밋은 코드와
  tasks.md · review.md 를 같이 싣는 관례라 흔한 모양을 잡는다.
- **거절에만 쓴다.** 1.12 가 금지한 것은 신원으로 착지를 *받는* 판정이다(freeze 이후 전부 참이라 아무것도 못 갈랐다). B 는 후보를 뒤로
  밀거나 없앨 뿐 창을 좁히지 못하므로, 틀리면 넓은 창(a122 이전 동작)으로 떨어진다 — 막는 쪽으로 틀린다. 그래도 1.12 문구의 적용 범위를
  바꾸는 결정이라 사람 몫이다.
- **오탐 모양은 넓은 교차 커밋**(반입 · 재번호 · 남의 문서 메모)이다. 오탐을 맞은 change 는 착지를 못 얻고 넓은 창으로 간다.
- 성능은 안 쟀다 — 걷기마다 그 디렉터리를 만진 커밋 목록 한 번 + 커밋별 `diff-tree` (7.5 와 같이 볼 것).

## Pre-Edit Gate — task 7.2.6 (H4: 착지 뒤의 자기 수리) (2026-09-16)

사람이 2026-09-16 에 **변형 B-ANYGO** 를 골랐다(옵션 1). 규칙: 착지 후보 **뒤에**, 그 change 의
디렉터리를 만지면서 Go 파일도 고친 **비병합** 커밋이 있으면 그 후보를 착지로 받지 않는다.
전수 비용은 위 `## MEASURE — task 7.2.6` 이 잰 그대로다.

### 무엇을 고치나

| 함수 | 편집 전 (HEAD `e9f86820`) | 무엇 |
|---|---|---|
| `_self_repair_commits` | **없다** | 새 함수 — 깃발 커밋 집합을 **한 곳에서** 만든다 |
| `_landing_refusal` | 분기 10 · 반환 8 · raise 0 (`ast.before-7.2.6.json`) | 맨 **뒤에** 조건 하나 · 인자 하나(`repairs`) |
| `compute_landing` | 분기 8 · 반환 5 | 깃발 집합을 한 번 재서 넘긴다 |
| `resolve_landing` | 분기 8 · raise 5 | 같은 집합을 넘기고, 거절에 **복구 경로**를 붙인다 |
| `_recording_refusal` | 분기 9 · 반환 9 | "이미 있다" 거절이 복구 경로를 말한다 |

`floor` 와 **같은 모양**이다 — 호출자가 한 번 재서 넘기고 판정은 한 집에 산다
([[two-judgements-cover-for-each-other]]). 자리는 **맨 뒤**다: 앞의 일곱 가드는 각자 자기 문장으로
못 박혀 있고, 앞에 세우면 그 못이 빠진다([[a-new-guard-unpins-the-guards-behind-it]]).

### 생산 판본이 측정한 규칙과 같은가 — 먼저 쟀다

측정은 커밋마다 `git diff-tree` 를 돌렸고 생산 후보는 `git log --no-walk --stdin --name-only` 한 번으로
같은 목록을 읽는다. 착지 있는 68건 전수 A/B(`h4b_ab_flagged.py`): 깃발 커밋 **224개**, 불일치 **0**.
두 판본이 다른 집합을 냈다면 생산 규칙은 사람이 비용을 본 그 규칙이 아니다.

### 이 편집이 거절할 정상 입력 (먼저 적는다, [[fail-closed-must-name-what-it-rejects]])

- **기록이 없는 change 의 판정은 안 바뀐다.** `resolve_landing` 은 기록이 없으면 계산을 안 부른다 —
  오늘 저장소에서 기록은 a099(아카이브) 하나뿐이고 그 값은 계산값과 같으며 네 시점 전부 "same" 이다
  (편집 전 실측: `check` rc 0 · required 32).
- **`--record-landing` 이 새로 거절하는 것**: 아카이브 시점 기준 4건(실물 H4 셋 + 오탐
  add-candidate-discovery), 오늘 HEAD 기준 +3(a066 · a071 · a044 — 넓은 교차 커밋). 그 change 들은
  창을 못 좁힐 뿐 빨개지지 않는다(기록이 없으면 대상은 워킹트리 = 가장 넓은 창).
- **기록을 가진 change 는 뒤에 자기 디렉터리를 만지는 Go 커밋이 서는 순간 빨개진다.** 이것이
  이 규칙의 **목적**이다(H4). 복구는 기록을 지우는 커밋 뒤 재기록이고, 그 경로를 문장에 넣는 것이
  같은 로트의 나머지 절반이다.

### 안 막는 것 — 한계로 적는다

- **커밋을 쪼개면 뚫린다** (Go 수리와 문서를 다른 커밋에). 망각 가드이지 위조 가드가 아니다.
- **병합 커밋 자신의 변경은 안 읽는다** (`--no-merges`) — 7.2.4 의 H7 과 같은 한계 부류.
- **신원을 거절에만 쓴다.** 1.12 가 막은 것은 신원으로 착지를 **받는** 판정이다. 틀리면 창이
  넓어지는 쪽으로만 틀린다.

## VERIFY — task 7.2.6 (착지 뒤의 자기 수리) (2026-09-16)

사람이 고른 **변형 B-ANYGO** 를 넣었다. 생산 Go 코드 변경 0 — 바뀐 것은 `tools/logic-map` 의
판정과 문장이다.

### 무엇이 들어갔나

- `_self_repair_commits` (새 함수) — 그 change 의 디렉터리를 만지면서 Go 도 고친 **비병합** 커밋을
  오래된 것부터. 조회 둘(디렉터리를 만진 커밋 → `--no-walk --stdin --name-only` 로 전체 목록).
- `_landing_refusal` — 인자 `repairs` 와 조건 하나를 **맨 뒤**에. 후보 뒤에 그런 커밋이 있으면 거절.
- `compute_landing` · `resolve_landing` — 같은 집합을 **한 번** 재서 넘긴다(`floor` 와 같은 모양).
- `LANDING_RECOVERY` (새 상수) — 거절당한 기록을 옮기는 길. 판정 경로와 기록 명령이 **같은 문장**을 쓴다.

### 규칙이 사람이 비용을 본 그 규칙인가 — 먼저 쟀다

생산 `_self_repair_commits` 대 측정 때 쓴 커밋별 `git diff-tree` 판본, 착지 있는 **68건 전수**
(`726_flagged_ab.py`): 깃발 커밋 **224개 · 불일치 0 · 순서(가장 오래된 것이 첫 항목) 어긋남 0**.

### H4 를 이 픽스처에서 직접 봤다 (`h4_probe726.py`)

| 수리 모양 | 편집 전 (HEAD `e9f86820`) | 편집 후 |
|---|---|---|
| 고정 파일 + change 디렉터리 (H4) | `[]` 초록 · required 1 | **거절** — `is followed by 1 later commit(s) …` |
| 다른 Go + change 디렉터리 (ANYGO) | `[]` 초록 | **거절** |
| 이웃의 Go 편집 | `[]` 초록 | `[]` 초록 (거절하면 안 되는 자리) |
| change 디렉터리만 (Go 없음) | `[]` 초록 | `[]` 초록 |
| 수리와 문서를 쪼갬 | `[]` 초록 | `[]` 초록 — **알려진 한계** |

기록을 지우면 같은 트리가 `AST source hash is stale` 로 빨갛다. 없앤 것은 그 **비대칭**이다.

### RED → GREEN

새 시험 **20개**(+ 기존 시험 하나에 단언 한 줄). 편집 전 도구(HEAD blob 사본)에 걸면 **16개**가
빨갛고(오류 13 · 실패 3), 초록으로 남는 넷은 **편집 전에도 참인** 것을 못 박는다 — 이웃의 나중
편집 · change 디렉터리만 만진 커밋 · 기록 없는 트리의 stale · 번들 갱신 뒤 재기록이 도는 복구 경로.
편집 후 전부 초록. 스위트 150 → **170** (`make sdd-test` 의 logic-map 부분).

### 변이 — `726_mut.py` (무변이 대조군 GREEN 확인 후)

| 변이 | 결과 | 잡은 시험 수 |
|---|---|---|
| S1 가드 삭제 | CAUGHT | 7 |
| S2 ANYGO → PINNED | CAUGHT | 1 |
| S3 `--no-merges` 제거 | **SURVIVED — 동등 변이(아래)** | 0 |
| S4 조상 판정 뒤집기 | CAUGHT | 17 |
| S5 후보 자신도 센다 | CAUGHT | 2 |
| S6 가드를 맨 앞으로 | CAUGHT | 1 |
| S7 가장 새 수리를 이름으로 | CAUGHT | 1 |
| S8 아카이브 이전 경로 무시 | CAUGHT | 1 |
| S9 복구 경로 문장 삭제 | CAUGHT | 2 |
| S10 계산 경로만 무장 해제 | CAUGHT | 1 |
| S11 선언 경로만 무장 해제 | CAUGHT | 1 |
| S9b 계산값 대조 거절의 복구 경로 삭제 | CAUGHT | 1 |
| S12 역사 단순화 (`--full-history` 제거) | CAUGHT | 2 |
| S13 첫 조회 실패를 빈 목록으로 | CAUGHT | 1 |
| S14 둘째 조회 실패를 빈 목록으로 | CAUGHT | 1 |
| S15 `core.quotePath` 기본값 | CAUGHT | 1 |
| S16 `diff.renames` 기본값 | CAUGHT | 1 |
| S17 `rev-list` 실패를 빈 목록으로 | CAUGHT | 1 |

**18개 중 17 CAUGHT · 1 SURVIVED(S3, 동등 변이).** 아래 S9b · S12~S17 은 적대 리뷰가 연 갈래다.

첫 판에서 S5 · S6 · S8 이 SURVIVED 였다. S6 · S8 은 재는 시험이 **없었고**(자리 · 아카이브),
S5 는 시험이 있었지만 **닿지 않았다** — 픽스처의 수리 커밋이 번들을 갱신하지 않아 불일치 가드에
먼저 걸렸다([[mutation-must-reach-the-thing-under-test]]). 수리와 번들 갱신을 **한 커밋**에 넣은
모양(이 저장소의 정직한 로트 모양)으로 바꾸니 잡힌다.

**S3 는 왜 동등 변이인가 — 재서 적는다.** `git log` 는 `--diff-merges` 를 명시해야만 병합 커밋의
diff 를 낸다. git 2.43 에서 직접 쟀다: 기본 · `log.diffMerges=first-parent` ·
`diff.merges=first-parent` 셋 다 병합의 이름 목록이 **비었고**, `--diff-merges=first-parent` 에서만
나왔다. 그래서 `--no-merges` 없이도 병합은 깃발이 안 서고, 이 한계는 저장소 설정의 함수가 아니다.
한계 자체는 `test_a_fix_made_inside_the_merge_commit_itself_is_the_known_limit` 이 못 박는다 —
누가 `--diff-merges` 를 더해 한계를 없애면 그 시험이 빨개진다.

### 안 고친 것 (알려진 한계로 spec 에 적었다)

- **쪼갠 커밋**: Go 수리와 그 change 의 문서를 다른 커밋으로 나누면 뚫린다. 망각 가드이지 위조
  가드가 아니다.
- **병합 커밋 자신의 수리**: 7.2.4 의 H7 과 같은 부류.
- **이웃의 나중 편집**: 내용으로는 자기 수리와 안 갈린다(68건 중 FILE 61 · FUNC 55 거절이 그 값).

### 읽기 하나를 사실대로 고쳤다

둘째 조회에 `-c core.quotePath=false` 를 준다. 기본값은 이름의 비ASCII 바이트를 따옴표로 인용하는데,
그러면 그런 이름의 `.go` 가 `.go` 로 안 끝나서 **안 보이고**, 안 보이면 거절을 안 해서 창이 좁아지는
쪽으로 틀린다. 오늘 이 저장소의 추적 경로 15,974개 중 비ASCII 는 **0** 이라 측정값은 그대로다(A/B 깃발 224 ·
불일치 0). 규칙을 바꾼 것이 아니라 읽기를 사실대로 만든 것이다.

## VERIFY — task 7.2.6 독립 적대 리뷰와 그 수리 (2026-09-16)

구현 뒤 독립 적대 리뷰를 돌렸다(읽기 전용, 별도 세션). **P0 하나**를 포함해 여덟 건이 나왔고,
모든 주장을 내가 직접 실행으로 재확인한 뒤 고쳤다.

### F1 (P0) — 경로 조회의 기본 단순화가 자기 수리 커밋을 통째로 버린다

경로를 제한한 `git log` 는 기본으로 역사를 단순화한다. 병합이 **그 경로에 대해** 한 부모와
TREESAME 이면 반대편 가지를 통째로 버린다 — 그래서 곁가지에서 한 자기 수리가 깃발 목록에서
사라진다. **직접 재현했다**: 곁가지에 (Go 수리 + change 디렉터리) 커밋을 두고 줄기가 같은 문서
편집을 한 뒤 병합하면, 깃발이 `[]` 이고 `check` 가 `[]` 초록인데 `git show <기록>:internal/own.go`
는 지금 소스와 **다르다**. H4 그 자체가, 이 change 의 제목이 가리키는 상황(병합)에서 그대로
열려 있었다.

**왜 내 측정이 이것을 못 잡았나.** `726_flagged_ab.py` 는 **둘째 단계만** 비교한다(커밋별
`git diff-tree` 대 `--no-walk`). 두 팔이 **같은(이미 가지쳐진) 후보 목록**을 입력으로 받으므로
원리적으로 이 결함을 반증할 수 없다 — [[falsification-must-vary-the-right-axis]] 그대로,
바꾼 축이 결함이 사는 축이 아니었다. 시험도 같은 구멍이었다: 유일한 병합 픽스처가 줄기가 change
디렉터리를 **안 만지는** 방향이라 가지치기가 일어나지 않았다
([[linear-fixtures-never-exercise-dag-guards]]).

**수리**: 첫 조회에 `--full-history`. **비용을 다시 쟀다** — 활성 · 아카이브 **126건 전수**에서
기본 판본과 `--full-history` 판본의 깃발이 **386 · 386 · 다른 change 0**. 오늘 넣으면 공짜다.
가지치기가 **실제로 일어나는지**를 먼저 단언하는 병합 픽스처를 시험에 넣었다(변이 S12 CAUGHT).

### F2 — git 이 실패하면 가드가 조용히 꺼졌다

`returncode` 와 "만진 커밋이 없다"를 한 줄에 섞어 둘 다 빈 목록으로 돌려줬다. 빈 목록은 이 규칙에서
"거절할 것이 없다"이므로, 조회 하나가 rc 1 을 내면 거절돼야 할 입력이 초록이 된다(주입 실험으로
재현). 옆의 `_evidence_floor` 는 실패하면 거절로 가는데 새 함수만 반대였다 —
[[missing-tool-reports-clean]] 의 "빈 출력·0은 위반 0이 아니라 **검사 0**".
**수리**: 세 실패 자리(첫 조회 · 둘째 조회 · `rev-list`)가 전부 `RuntimeError` 로 판정이 된다.
빈 목록은 이제 **하나**를 뜻한다 — 그 디렉터리를 만진 커밋이 정말 없다. 변이 S13 · S14 · S17 CAUGHT.

### F3 — 계산값 대조 거절만 복구 경로를 빠뜨렸다

내가 같은 로트에서 spec 에 쓴 SHALL("거절된 기록을 옮기는 경로를 사유 문장이 말해야 한다")을
그 문장 하나가 안 지켰다. 게다가 `--record-landing` 을 권하면서 그 명령은 "이미 있다"로 거절한다 —
7.7 이 닫은 **서로를 가리키는 두 문장**과 같은 모양이다. 기존 시험에 단언 한 줄(변이 S9b CAUGHT).

### F4 · F5 — 판정이 사람의 git 설정의 함수였다

`diff.renames` 가 켜져 있으면 `.go` 를 비-`.go` 이름으로 옮긴 커밋의 **옛 이름**이 사라져 깃발이
안 선다. `core.quotePath` 기본값은 비ASCII 이름을 인용해 `.go` 로 안 끝나게 만든다(내가 이미
껐지만 **시험이 0** 이었다 — 빼는 변이가 살아남았다). 둘 다 명령줄에서 못 박고 각각 시험을 넣었다
(S15 · S16 CAUGHT). `_evidence_floor` 가 `-M100% --no-follow` 로 지킨 원칙과 같은 자리다.

### F6 · F7 — 잠복 결합과 공허한 단언

`_self_repair_commits` 는 `analysis` 에서 change 디렉터리를 **유도**하는데, 빌린 증거 경로에서
`analysis` 는 빌려주는 쪽으로 재바인딩된다. 오늘은 도달하지 않는다(빌리는 쪽 기록은 이름으로
거절되고 선언이 없으면 먼저 돌아간다) — 그 사실을 **spy 시험으로 못 박았다**. 1.8 의 "빌려주는 쪽
착지를 복사" 규칙이 되살아나면 그 시험이 빨개져서, 누구의 디렉터리를 재야 하는지를 그때 사람이
정하게 된다. 그리고 `assertTrue(marks["record"])`(40자리 hex — **항상 참**)를 지우고, 재지 않는
것을 약속하던 시험 이름을 고쳤다(`test_merge_commits_are_not_read` → `…_a_fix_on_a_side_branch_is_read`).

### F8 — 후보마다 `merge-base` 를 수리 개수만큼 돌았다

docstring 은 "한 번 재서 넘긴다"고 적었는데 이 줄만 후보 하나에 프로세스를 `len(repairs)` 개 띄웠다
(a112 는 오늘 수리가 스물여섯). `_repairs_after` 로 옮겨 `rev-list <후보>..HEAD` 한 번으로 바꿨다 —
**같은 집합**이고(그것이 곧 "HEAD 에서 닿고 후보의 조상이 아닌" 커밋), 후보 자신은 여전히 빠진다.

### 수리 뒤 전수 재측정

| 무엇 | 값 |
|---|---|
| 실물 A/B, 활성 · 아카이브 **126건 전수** | **판정 다름 0** · 계산값 다름 9 · 예외 10(전부 양쪽 문구 동일 = 기존 조건) |
| 계산값이 달라진 9 | 착지 상실 7(a044 · a083 · a084 · add-candidate-discovery · apply-us-measurement-fixes · a066 · a071) · 이동 2(a045 · a098) |
| 기록으로 창이 좁아지는 실물 change | a099 하나, `21a315d1` · required 32 — **불변** |
| 시험 | 150 → **170** (새 20 · 편집 전 도구에서 16 빨강) |
| 변이 | 18개 중 **17 CAUGHT** · 1 SURVIVED(S3 동등 변이) |
| 비용 (한가한 기계, 판본을 번갈아 3회 중앙값) | 신호 자체 **0.02~0.04초**(조회 2회) · `check` 전체 차이 a099 −0.04초 · a094 −0.15초(잡음 안) |

착지 상실 7건은 **판정이 아니라 `--record-landing` 이 거절하는 것**이다 — 기록이 없으면 5단계는
워킹트리를 대상으로 삼으므로 창이 가장 넓다. 그래서 오늘 빨개지는 change 는 없다.

## MEASURE · Pre-Edit Gate — task 7.5 (성능, 리뷰 I1) (2026-09-18)

**리뷰가 적은 숫자는 리뷰 시점의 수다.** I1 은 `a055 133.7s · 다른 아카이브 219.6s` 로
적혀 있는데 그 뒤 7.2.1 · 7.2.2 · 7.2.4 · 7.2.6 · 7.6 · 7.7 이 같은 경로를 여섯 번 고쳤다.
그래서 수리 **직전에** 다시 쟀다 ([[caller-count-is-not-fix-site-count]]).

### 전수 census (오늘 HEAD, 126 change)

걷는 것 **93** · 안 걷는 것 33(번들 0 이 대부분, base 없음 셋). 후보 합계 **36,767**.
후보 수 × 번들 수로 예측한 spawn 은 최소 **1,153,251** ~ 최대 4,576,237 이고, 한 번의
게이트 실행이 무는 것은 그중 change **하나**의 몫이다. 가장 비싼 활성은 a071(후보 347 ·
번들 35), 가장 비싼 아카이브는 a062(후보 388 · 번들 220).

### 실측 — 비용은 **실패하는 walk** 에 있다

| change | 벽시계 | spawn | 결과 |
|---|---|---|---|
| a071-wire-kr-us-protection-readiness | **73.19s** | 12,155 | none(347 후보 전부 거절) |
| a092-an-alert-does-not-hold-the-stop | **72.32s** | 11,329 | none |
| a112-run-four-strategy-families-… | 1.64s | 278 | `4f49a8eb` |
| a100-wire-fill-to-broker-protection | 0.78s | 126 | `016da624` |

성공하는 walk 는 착지에서 멈추므로 이미 싸다. 느린 것은 **아무 후보도 안 받는** walk 다.

### spawn 이 어디서 나오는가 (스택 귀속, 잎이 아니라 호출자)

- a071: `_committed_bytes` **97.1%** — 호출자로 가르면 `_pinning_at` 이 압도, `_unheld_bundles` 70(0.6%).
- a092: `_committed_bytes` **97.3%**.
- a112(성공): `_pinning_at` **47.5%** · `_unheld_bundles` **47.5%** — 받는 후보가 두 함수를 다 통과하므로 반반이다.
- 어느 실행에서도 `_landing_refusal` 의 마지막 `all(...)` 는 4~6 spawn(0.0~2.2%)이다.

**근본 원인은 알고리즘이 아니라 fetch 단위다**: 도구가 blob 을 `git show` 로 **한 프로세스에
하나씩** 읽는다. 후보 하나에 번들 N 개면 프로세스 N 개다.

### 지렛대 넷을 재서 고른다 (a071, 고유 소스 20)

| 방식 | 후보당 | 347 후보 예측 |
|---|---|---|
| `git show` 하나씩 (오늘) | 83.4 ms | 28.9s |
| `git cat-file --batch` 한 번 | **8.5 ms** | **2.9s** |
| `git ls-tree` (blob id 만) | 4.9 ms | 1.7s — **쓸 수 없다**: sha256 을 못 만들어 판정이 안 선다 |
| 번들 재파싱 ×3 (오늘) | 25.2 ms | 8.8s |

두 방식이 **같은 바이트**를 낸다(20/20 확인). 한 개짜리도 batch 가 더 빠르다 —
`show` 4.05~4.80 ms 대 `cat-file --batch -z` **2.64~3.39 ms**. 그래서 `_committed_bytes` 를
배치 위에 올려도 단건 호출자가 느려지지 않는다.

### 편집 집합과 **안 하는 것**

바꾼다: 새 `_committed_many`(한 ref 의 여러 blob 을 `git cat-file --batch -z` **한 번**으로) ·
`_committed_bytes`(그 위의 1-원소 호출 — "그 ref 의 blob 을 읽는다"는 철자를 **한 곳**에 둔다,
[[two-judgements-cover-for-each-other]]) · `_pinning_at` · `_unheld_bundles`.

**안 바꾼다**, 그리고 사유를 적는다:
- `_landing_refusal` 의 마지막 `all(...)` — 측정값 4~6 spawn. 이 함수는 **순서가 못**이고
  (7.2.2 · 7.2.6 · 7.6 이 각 갈래를 그 순서로 못 박았다) 값이 0 인 편집으로 그 자리를
  건드리지 않는다 ([[a-new-guard-unpins-the-guards-behind-it]]).
- 번들 목록 hoist(예측 8.8s) — `compute_landing` 은 `_pinning_bundles` 를 **부르면 안 된다**
  (구조 시험 `test_the_recorder_and_its_advice_ask_one_judge` 가 단언한다). 뚫으려면
  `_walk_floor` 의 계약을 바꿔야 하는데, 배치 뒤 남는 몫이 작아 값이 그 대가에 못 미친다.
  **구현 뒤 실측으로 재확인한다.**
- `--ancestry-path` (I1 이 적은 넷째) — 후보 **집합**을 바꾼다. 7.8 의 M6 이 곁가지 후보가
  판정에 실제로 들어옴을 실측했으므로, 이것은 성능이 아니라 **규칙** 변경이고 7.5 의 몫이 아니다.

### Pre-Edit 선언

편집 전 AST 를 같은 열거기로 뽑아 뒀다 —
`analysis/python-function-logic/tools-logic-map--{_committed_bytes,_pinning_at,_unheld_bundles}/ast.before-7.5.json`.
`_committed_bytes` 는 이 change 에서 처음 편집하므로 디렉터리를 새로 만들었다.
분기 수(편집 전): `_committed_bytes` 1 · `_pinning_at` 3 · `_unheld_bundles` 9.
High-risk 아님 — 생산 거래 코드가 아니라 게이트 도구이고, 판정은 **바뀌면 안 된다**(A/B 로 증명한다).

## VERIFY — task 7.5 (성능: blob 을 한 프로세스에 하나씩 읽지 않는다) (2026-09-18)

**생산 Go 코드 변경 0.** 게이트 도구(`tools/logic-map/check_analysis.py`)만 바꿨고 판정은
전수 A/B 로 불변을 증명했다.

### 근본 원인은 알고리즘이 아니라 fetch 단위였다

측정 기록은 위 `## MEASURE · Pre-Edit Gate — task 7.5`. 두 가지가 겹쳐 있었다.

1. **blob 을 `git show` 로 한 프로세스에 하나씩** 읽었다 — 후보 하나에 번들 N 개면 프로세스 N 개.
2. 번들 목록(`_pinning_bundles`)이 **후보마다** 다시 계산됐다 — glob + JSON 파싱 +
   `Path.resolve()`. 그리고 `normalized_source` 는 `root.resolve()` 를 호출당 **두 번** 돌렸다.

첫째를 걷어내자 둘째가 드러났다(프로파일 두 번). 그래서 셋을 차례로 쟀다 —
73.19s → 27.84s → 10.97s → **3.69s** (a071).

### 고친 것

- 새 `_committed_many(root, ref, relatives)` — 한 ref 의 여러 blob 을 `git cat-file --batch -Z`
  **한 번**으로. `-Z` 는 입력과 출력을 둘 다 NUL 로 끊는다(개행 든 경로 · tree 의 raw NUL).
  크기는 **머리가 선언한 값**으로 자른다. **blob 만** 내용으로 친다.
- `_committed_bytes` 는 그 위의 1-원소 호출이 됐다 — "그 커밋의 blob 을 읽는다"는 철자가
  **한 곳**이다 ([[two-judgements-cover-for-each-other]]). 단건도 더 싸다(실측 4.05~4.80 ms → 2.64~3.39 ms).
- `_pinning_at` · `_unheld_bundles` 가 번들 목록을 **받는다**. `_walk_floor` 가 자기가 잰
  목록을 같이 돌려주고 `compute_landing` 이 `_landing_refusal` 에 넘긴다 — `floor` · `repairs` 가
  이미 그렇게 넘어오던 방식 그대로다. `_evidence_floor` 도 같은 목록을 받는다.
- `normalized_source` 의 `root.resolve()` 중복 제거(호출당 3회 → 2회). 판정 동일 —
  오히려 검사와 계산이 **같은 닻**을 쓴다.

**구조가 안 바뀐 것이 증거다.** 편집 전후 AST 를 같은 열거기로 뽑아 대조했다:
`_landing_refusal` 분기 11 · 반환 9, `compute_landing` 8 · 5, `resolve_landing` 8 · 2 · raise 5,
`_recording_refusal` 9 · 9, `_walk_floor` 2 · 3, `_evidence_floor` 6 · 2 — **여섯 전부 불변**.
가드도 그 **순서**도 안 움직였다 ([[a-new-guard-unpins-the-guards-behind-it]]).
바뀐 넷은 `_committed_many`(신규 9), `_committed_bytes`(1→0), `_pinning_at`(3→4),
`_unheld_bundles`(9→13)이고 늘어난 갈래는 전부 배치 입력 조립이다.

### 판정 A/B — 전수 126건

`compute_landing` 을 편집 전 판본(HEAD blob 에서 꺼낸 사본)과 지금 판본으로 각각 돌려
`(값, 사유)` 를 비교했다. 사본 대상 + 대조군 단언 ([[mutation-revert-needs-the-right-baseline]]).

**비교 116 · SAME 116 · DIFFERENT 0** (base 없음 10). 예외도 타입과 문장까지 비교했다.

합계 **2023.0s → 162.1s (12.5×)**. 가장 느렸던 것들:

| change | before | after | 배 |
|---|---|---|---|
| 2026-08-29-add-candidate-discovery | 249.49s | 8.47s | 29.4× |
| 2026-08-02-a055-console-settings-cadence | 193.71s | 5.25s | 36.9× |
| 2026-08-29-enable-engine-autostart-menu | 166.61s | 7.70s | 21.6× |
| 2026-08-29-verify-us-market | 147.27s | 7.23s | 20.4× |
| a071-wire-kr-us-protection-readiness | 73.19s | 3.69s | 19.8× |

spawn 도 같이 줄었다 — a071 12,155 → **697** (17.4×), a112 278 → **16**.
리뷰 I1 이 적은 `133.7s · 219.6s` 는 오늘 실측(a055 193.71s · add-candidate-discovery 249.49s)과
같은 자리이고 크기만 흘렀다.

### 뮤테이션 — 18 중 17 CAUGHT, 남은 하나는 **동등 변이임을 증명**했다

사본 대상 · 무변이 대조군 GREEN 선행. 첫 판에서 **넷이 SURVIVED**(T4 · T5 · T6 · T7 — 전부
파싱 루프)였고, 그것이 곧 시험이 그 갈래에 **안 닿았다**는 뜻이었다
([[mutation-must-reach-the-thing-under-test]]).

- **T4**(비-blob 내용을 안 건너뜀) · **T7**(크기를 선언값이 아니라 NUL 탐색으로) — 트리를
  **혼자** 물으면 응답이 하나라 경계가 밀려도 안 드러난다. 트리 뒤에 파일 둘을 더 묻고,
  NUL 이 든 blob 을 따로 세웠다 → 둘 다 CAUGHT.
- **T5**(rc≠0 인데 안 돌아감) · **T17**(T5+T6 동시) — 저장소 아닌 곳을 물으면 git 이 **빈 출력**
  으로 죽고 그때는 파서도 같은 답을 낸다(서로를 덮었다,
  [[surviving-mutant-may-mean-accidental-safety]]). **부분 출력을 남기고 죽는** git 을 세워
  가르니 파서에 맡긴 쪽은 "진짜 blob 을 읽었다"고 답한다 → 둘 다 CAUGHT.
- **T14**(`and` → `or`)는 아카이브된 번들을 **옮긴 뒤 커밋에서** 들고 있는 픽스처가 없어서
  살아남았다. 그 갈래에 닿는지 먼저 단언하고 시험을 세웠다 → CAUGHT.
- **T16**(`_committed_bytes` 를 자기 `git show` 로 되돌림)은 행동이 같아서 행동 시험이 원리적으로
  못 가른다. 구조 시험 둘(`test_a_commits_blob_is_read_by_one_function` ·
  `test_the_walk_measures_the_bundle_list_once`)로 의존을 못 박았다 → CAUGHT.
- **T6**(`if end < 0: break` → `end = len(data)`)는 **끝까지 SURVIVED**. 가설로 넘기지 않고
  변이로 확인했다: 답을 같게 만드는 것은 **사전 채움**(`found = {r: None for r in wanted}`)이고,
  그 사전 채움을 지우는 변이 **T18 은 59개 시험이 잡는다**. 즉 T6 는 동등 변이이고
  `break` 는 비용 선택이지 판정이 아니다.

CAUGHT 17: T1 · T2 · T3 · T4 · T5 · T7 · T8 · T9 · T10 · T11 · T12 · T13 · T14 · T15 · T16 · T17 · T18.

### 시험

`test_check_analysis` **175 → 182** (7.5 가 8개 추가: 배치 등식 · 트리 · NUL blob · 개행 경로 ·
실패 · 빈 요청 · spawn 수 둘 · 아카이브 보유, 그리고 구조 둘). spawn **수**를 세는 시험이
핵심이다 — 시간은 기계마다 다르지만 프로세스 수는 알고리즘의 함수다
([[passing-test-is-not-evidence]]).

시험 쪽 호출 자리도 같이 고쳤다: 사설 판정을 직접 부르는 자리(17곳)가 이제 생산 경로가
넘기는 것과 **같은 목록**을 `_bundles(root, analysis)` 로 만들어 넘긴다.

### 게이트

`make lint`(go vet ×2) · `make sdd-test`(logic-map **257** · 71 · 22 · 16 · 18 · 15, 전부 OK) ·
`openspec validate --all --strict` **58 passed, 0 failed** ·
`check_analysis.py --change a122-…` **rc=0**(required 0 · 창 줄 불변).

**`make lint` 은 Python 을 안 본다**(go vet 둘뿐) — 빈 출력은 "위반 0"이 아니라 "검사 0"이다
([[missing-tool-reports-clean]]). Python 쪽 판정은 위 스위트와 뮤테이션이 전부다.

### 안 한 것과 사유

- **`--ancestry-path`** (I1 의 넷째) — 후보 **집합**을 바꾼다. 7.8 의 M6 이 곁가지 후보가 판정에
  실제로 들어옴을 실측했으므로 이것은 성능이 아니라 **규칙** 변경이고 7.5 의 몫이 아니다.
- **조기 종료**(I1 의 둘째) — `_pinning_at` 은 안 맞는 소스를 **전부** 세어 거절 문장에 넣는다.
  첫 불일치에서 멈추면 저자가 고칠 자리 목록이 한 개로 줄어든다. 배치 뒤 그 순회의 비용은
  프로세스 0 이라 살 이유도 없어졌다.
- **`_landing_refusal` 의 마지막 `all(...)`** — 실측 4~6 spawn(0.0~2.2%). 이 함수는 순서가 못이라
  값 0 인 편집으로 건드리지 않는다.
- `_committed_many` 의 `timeout` 은 60 이다(옛 단건 `git show` 는 30). 결함이 판정이 되는
  시점이 30초 늦어질 뿐 판정 자체는 안 바뀐다 — `_evidence_floor` 가 쓰던 값과 같게 뒀다.

## VERIFY — task 7.9 (문서: 착지 기록이 어디에도 안 적혀 있었다) (2026-09-18)

**코드 변경 0.** `--record-landing` 과 `landed-commit.txt` 는 6.1.2 에서 생겨 7.2 · 7.3 · 7.7 이
규칙을 바꿔 왔는데 `tools/logic-map/README.md` 에도 `docs/WORKFLOW.md` 에도 한 글자가 없었다
(실측: 두 파일에 두 낱말 모두 0회). 7.2 가 규칙을 정한 뒤에 쓰기로 한 순서대로 지금 썼다.

- `docs/WORKFLOW.md` — `## Function Logic Map` 안에 `### 착지 지점 — landed-commit.txt`.
  명령, **값은 저자가 고르지 않는다**, 도구가 받는 일곱 조건, 덮어쓰지 않는 규칙과 **복구 경로**,
  받지 않는 세 경우(빌림 · a063 이관 · 번들 0).
- `tools/logic-map/README.md` — 같은 내용을 도구 쪽 말로. 기록을 **커밋해야** 효력이 있다는 것
  (게이트는 HEAD 에서 읽는다)을 명시했다.

일곱 조건과 세 거절은 산문에서 옮기지 않고 `_landing_refusal` · `_walk_floor` ·
`ADOPTION_REFUSES_A_LANDING` · `BORROWED_REFUSES_A_LANDING` · `LANDING_RECOVERY` 를 읽어서 적었다
([[contract-numbers-from-the-receipt]]).

## 인계 준비 (2026-09-18)

다음 세션이 이어받을 수 있게 셋을 만들었다. 코드·판정 변경 0.

- **`HANDOFF.md`** — 먼저 읽는 문서(a112 선례와 같은 자리). 상태·읽는 순서·남은 일·
  사람 결정 둘·밟기 쉬운 지뢰 여섯.
- **`analysis/harness/`** — 7.5 의 측정·A/B·변이 하네스 여섯을 저장소로 옮겼다.
  **이유는 재현성이다**: `_landing_refusal` · `_self_repair_commits` 의 branch-test-map 이
  `722_mut.py` · `726_mut.py` · `76_mut.py` 를 증거로 인용하는데 그 파일들은 세션
  스크래치패드와 함께 사라져서 **인용만 남고 재현이 불가능**했다
  ([[borrowed-flm-evidence-goes-stale]]). 두 map 에 그 사실을 각주로 적었다.
- 옮기면서 절대경로를 **유도**로 바꿨다 ([[renamed-checkout-strands-absolute-path-state]]).

**옮긴 것이 도는지 실제로 돌려서 확인했다** — 그리고 둘이 깨져 있었다.
`75_census.py` 는 `_walk_floor` 가 7.5 에서 값 셋을 돌려주게 되자 126건을 전부 "못 걸음"으로
찍었다(빈 결과가 발견처럼 보인다, [[missing-tool-reports-clean]]). `75_ab.py` 는 비교 기준이
`HEAD` 로 굳어 있어 7.5 가 랜딩한 뒤 **대조군이 오염**됐다 —
[[a-recorded-boundary-stops-being-rechecked]] 그대로다. 기준을 인자로 빼고 오염되면 멈추게
했다(기본값 `8091e6c4` = 7.5 직전). 고친 뒤 여섯 전부 새 자리에서 돈다:
census 93 걷음 · levers · attrib · time · ab 대조군 OK · mut 대조군 GREEN + T3 CAUGHT.

**병행 세션 실측**: 이 작업 중에 다른 세션이 같은 워크트리에서 `1764c833` 을 커밋했다
(`.claude/` 둘). 내 커밋에 담지 않았고 HANDOFF 에 적었다
([[tossos-parallel-session-gate-contention]]).

## VERIFY — task 1.4 독립 적대 리뷰 (7 절 로트) + 수리 (2026-09-18)

7 절(7.1~7.9)은 전부 한 저자가 구현하고 그 저자가 검증했다. `e34e9e2c` 는 어떤 독립 리뷰도
받지 않은 채 랜딩했다. **다른 세션(`tossos-be`)에 리뷰를 맡겼다** — 저자는 자기 작업을
리뷰할 수 없고, 그 독립성이 1.4 가 요구하는 것이다.

리뷰어에게 반증 대상으로 **내 주장 넷과 내가 아는 의심 자리 넷을 먼저 넘겼다.** 숨기면
리뷰가 아니라 승인 요청이 된다.

### 판정: CHANGES REQUESTED — P0 1 · P1 2 · P2 10

리뷰어는 하네스를 쓰지 않고 **직접 실측**했다(편집 전 사본 A/B 6/6 SAME · 번들 경로
3,555 × 3 ref = 10,665 비교 불일치 0 · a071 OLD 53.6s/12,263 → NEW 3.74s/703).

#### P0 — 주장 3 이 **거짓이었다**: T6 은 동등 변이가 아니다

내가 적은 근거는 "답을 같게 만드는 것은 사전 채움이고 그것은 T18 이 잡는다"였다.
**두 명제가 다르다.** 재현해서 확인했다:

```
응답: <oid> blob 3 NUL abc NUL <oid> blob 5   (마지막 머리는 멀쩡, 내용만 잘림)
원본   → {'a.go': b'abc', 'b.go': None}
T6 변이 → {'a.go': b'abc', 'b.go': b''}
```

`None` 과 `b""` 는 `_unheld_bundles` 의 아카이브 대체 갈래(`committed is None and before`)를
여닫으므로 **판정이 갈린다**. T18 이 잡는 것은 `missing`/`continue` 갈래(도달 153회)이고,
`end < 0` 갈래는 **182개 시험 · 116개 코퍼스에서 도달 0회**였다. 살아남은 이유는 동등이
아니라 **미도달**이다.

리뷰어의 지적 중 가장 아픈 것: 같은 로트의 T14 에는 정확한 규율을 썼다
(`assertTrue(_pre_archive_path(ast_path))` — 닿는지 **먼저 단언**). **T6 만 그 규율을 안 썼고,
살아남은 것도 그것이다.**

**수리**: `end < 0` 을 판정으로 바꿨다(부분 답을 쓰지 않는다). 곁들여 같은 성격의 구멍 셋을
같이 닫았다 — 남은 바이트(프레이밍 오독 → **다른 파일의 바이트**를 답에 넣는다) · 경로 안의
NUL(요청 프레이밍 문자다) · 엄격 `utf-8` 인코딩(고아 서로게이트 이름에서 판정 대신 traceback).

#### P1 — `-Z` 의 최소 git 버전이 어디에도 안 적혀 있었다

리뷰어 실측: `-Z` 를 거절하는 shim 아래에서 **182 중 89 실패**. fail-closed 는 맞지만
(required 32 → 269, rc 120) **오진 둘** — 저자의 증거를 탓하고, 이미 HEAD 에 있는
`landed-commit.txt` 를 커밋하라고 한다.

**rc≠0 을 결함으로 올려 오진을 없애려다 되돌렸다.** 거부할 정상 입력을 세어 보니
**저장소가 아닌 루트**가 이 함수의 정상 호출 모양이었다 — 번들 검증만 보는 시험 **21개**가
임시 디렉터리에서 돈다([[fail-closed-must-name-what-it-rejects]]: 열거가 설계를 죽였다).
대신 값을 `GIT_BATCH_MINIMUM = "2.42"` **한 곳**에 두고 README·WORKFLOW 가 인용하며,
시험이 셋의 일치를 잰다. **오진 자체는 열린 채로 남는다 — 알려진 한계로 적는다.**

#### P1 — 편집 전 AST 10개 중 **7개가 편집 중간 상태**에서 뽑혔다

`source_sha256` 값 `279cbe29…` · `145b3e11…` 이 `check_analysis.py` 의 **역사 어디에도
없다**(리뷰어 전수 확인, 내 재확인도 같다). `enumerate.py` 는 revision 을 인자로 받는데
기본값 `worktree` 로 돌렸고, Pre-Edit 에 선언한 셋만 진짜 부모 해시였다.

**방법이 불건전하고 결과가 우연히 맞은 경우다.** 진짜 부모 `8091e6c4` 에서 열 개를 다시
뽑았다 — **내용은 열 개 다 동일**하므로 결론은 서지만, 산출물은 건전한 것으로 교체했다.
게이트는 `*/ast.json` 만 글롭하므로 이것을 **영원히 못 잡는다**(6.4 후보로 적어 둔다).

#### P2 — 주장 1 · 2 · 4 는 결론이 서고 **근거 또는 정밀도**를 고쳤다

- **주장 2 (가드 순서)**: 결론은 참인데 근거가 못 받쳤다. `분기 11 · 반환 9` 는 multiset
  크기라 **순열에 불변**이다. 순서 있는 배열로 다시 재니: `_landing_refusal` 은 순서열 11개
  중 **1개만 다르고** 그것은 번들을 어디서 받는지(`_pinning_bundles(root, analysis)` →
  `bundles`)이며 **가드 8개와 그 순서는 바이트 동일**, 반환 9개도 동일. `compute_landing` ·
  `resolve_landing` · `_recording_refusal` 은 순서열 전체가 동일.
  그리고 **"여섯 전부 불변"은 과했다** — `_walk_floor` 의 반환 계약이 2-튜플 → 3-튜플로
  넓어졌는데 수는 `3 → 3` 이라 아무것도 못 봤다. FLM 여섯의 주석을 순서열 실측으로 바꿨다.
- **주장 1 (전수 A/B)**: 버틴다. 다만 덮는 범위가 문장보다 좁다 — `compute_landing` 만
  덮고 `resolve_landing` · `check()` 는 밖이다. 정확한 문장은 "**`compute_landing` 의 값과
  사유가** base 가 있는 116건 전부에서 불변"이다. (리뷰어 확인: 공허하지 않다 — 닿는 93건이
  전부 배치 ≥2, `missing` 갈래는 116 중 67 이 때린다.)
- **주장 4 (성능)**: 크기는 참, 유효숫자는 재현 안 된다. a071 기록 73.19s/19.8× 대 리뷰어
  53.6s/14.3× · 2차 54.9~56.2s/14.7~16.1×. **spawn 수는 거의 정확히 일치**(12,263 대 12,155,
  양쪽 17.4×). 즉 **알고리즘 주장은 단단하고 흔들리는 것은 벽시계뿐**이다. 앞으로 배수는
  spawn 으로 적고 벽시계는 자리수를 줄여 적는다.

#### P2 — 나머지

고쳤다: 경로 NUL desync(가드+시험) · 고아 서로게이트 회귀(`surrogateescape`) ·
WORKFLOW 가 거절 가드 **8개 중 7개**만 적던 것(병합 안에서만 증거가 들어온 경우를 추가) ·
`75_mut.py` 에 **양성 대조**(변이가 돌았는지까지 잰다 — 없으면 눈먼 계측기와 진짜 음성이
같게 기록된다, T6 이 그 결과다) · `75_ab_done.json` 을 baseline 으로 키 잡기.

적어만 둔다: `normalized_source` 가 서술에서 빠져 11개 수정·10개 서술이었다(이 절에서 센다) ·
tree→None 의 방향이 자리마다 다르다(`_pinning_at` 은 엄해지고 `_unheld_bundles` 는 `None` 이
대체를 열어 느슨해진다 — 둘 다 도달 희박) · `_landing_refusal` 가드 2("is pinned by no")의
문장을 단언하는 시험이 **0개**다(기존 결함, 순서 안전성 논증이 "가드마다 자기 문장"에 기대므로
값이 있다) · **커밋된 하네스는 기록된 수를 낸 적이 없다**(`75_ab.py` 는 인증 대상보다 한 커밋
뒤에 처음 커밋됐다).

### 수리 뒤 검증

`test_check_analysis` **182 → 187**(추가 6, 이름 바뀐 것 1) · 변이 **22 전부 CAUGHT · 생존 0**
(양성 대조 포함) · 판정 A/B **116/116 SAME · DIFFERENT 0**(기준 `e34e9e2c`, 그것이 이미
`8091e6c4` 대비 116/116 이므로 추이적으로 불변) · 성능 회귀 없음(합계 155.1s → 149.9s) ·
`openspec validate --all --strict` 58/58 · `check_analysis --change a122` rc=0.

**리뷰어는 수리하지 않았다** — findings 만 냈고 수리는 저자 몫이다. 이 절이 그 수리의 기록이다.

## MEASURE · Pre-Edit Gate — task 7.5.1 (gstack 리뷰가 찾은 permissive 결함) (2026-09-19)

gstack `/review` 를 로트 `8091e6c4..HEAD` 에 돌렸다(전체 브랜치 diff 는 380파일·46,978줄로
수개월치 a112 가 섞여 신호가 묻힌다). 세 출처 — Claude 구조 · Claude 적대 서브에이전트 · Codex.
**두 출처가 독립적으로 같은 근본 원인을 찾았고 둘 다 재현했다.** 사람이 2026-09-19 에
"지금 다 고친다"를 골랐다.

### 근본 원인 — `None` 이 "git 이 못 돌았다"와 "객체가 없다"를 섞는다

`_committed_many` 는 rc≠0 에서 전부 `None` 을 돌려준다. 옛 `git show` 에서는 "못 돌았다"가 거의
도달 불가였지만 `-Z` 의존이 그것을 **모든 읽기에서** 도달 가능하게 만들었다(`-Z` 를 모르는 git).
그리고 나는 docstring 에 "부르는 쪽은 `None` 을 불일치로 세므로 판정이 느슨해지지 않는다"고
**전칭으로** 적었다. 틀렸다 — blob 읽기 자리를 **상태까지** 세면 permissive 가 둘이다:

| 자리 | `None` 이 뜻하게 되는 것 | 방향 |
|---|---|---|
| 가드 7 `_landing_refusal` (base 쪽만 실패) | `None != bytes` → "소스가 바뀌었다" | **수락** ⚠ (서브에이전트 F1) |
| `_landing_record` | "기록 없음" → 착지 검증 전체 생략 | **생략** ⚠ (Codex P1-1) |
| `_pinning_at` · `_unheld_bundles` · `validate_target` | 불일치/누락 → 거절 | 보수적 ✓ |

**가드 7 은 내가 재현했다** (저장소 V1 픽스처, base 쪽 읽기만 실패시킴):

```
정상 git       → rc=1, 거절, 기록 안 씀
base 쪽만 실패 → rc=0, 편집 전 커밋 X 를 착지로 **기록하고 파일까지 쓴다**
```

그 창(`base..X`)에는 이 change 의 Go 작업이 없다. 가드 7 은 사람이 2026-09-14 에 V1 을 막으려고
고른 규칙인데 **자기가 막으려던 것을 통과시킨다.** 내 앞선 열거가 틀린 이유: 그 자리를
"`None == None` → 거절"로 셌고 **한쪽만** 실패하는 상태를 안 셌다.

### 어제의 되돌림이 이 구멍을 남겼다 — 그리고 그 근거를 다시 쟀다

1.4 수리 때 rc≠0 을 결함으로 올리려다 "시험 21개가 저장소 아닌 곳에서 돈다"는 이유로 되돌렸다.
다시 재니 그 시험들은 `run_check` 와 인라인 11개가 **git 을 mock 하려는** 픽스처였고
(`resolve_base`·`changed_existing_functions` 를 mock) `_landing_record` 만 빠뜨렸다 — 조용히
`None` 이라 무해해 보였을 뿐이다. 빈 저장소와 저장소 아닌 곳의 git 답을 쟀다:

```
git init 한 빈 저장소 → "HEAD:x.go missing"  rc=0   (= 기록 없음, 지금과 같은 답)
저장소 아닌 곳        → "fatal: not a git repository"  rc=128
```

사본에서 **rc≠0→raise + 그 픽스처들을 빈 저장소로**를 돌리니 깨지는 시험은 옛 동작을 단언하던
**내 시험 둘뿐**이다(23 → 2). 즉 거부할 정상 입력은 0 이고, 어제 센 21 은 픽스처의 누락이었다.

### F9 — `GIT_BATCH_MINIMUM = "2.42"` 는 **맞았지만 근거가 없었다**

기억으로 쓴 수였다. git 자신의 릴리스 노트로 확인했다 —
`/usr/share/doc/git/RelNotes/2.42.0.txt:150`:
*"git cat-file --batch" and friends learned "-Z" that uses NUL delimiter for both input and output.*

### 편집 집합

| 함수 | 무엇 | 출처 |
|---|---|---|
| `_committed_many` | rc≠0 → git 의 stderr 를 담아 raise · ref 의 NUL 도 검사 · 머리를 엄격히(아는 모양만, `missing` 은 **물은 spec 그대로**) · 페이로드 뒤 NUL 종단 확인 · 잘림은 **레코드**로 세고 초과는 "잘림"으로 | F1 · P1-1 · F6 · F7 · P2-3 · F4 · F5 |
| `_landing_refusal` | 가드 7 을 양쪽 한 프로세스씩으로(가드 **위치 불변**) | F3 |
| `validate_target` · `check` | 고정 소스를 착지에서 **한 번** 미리 읽어 넘긴다(`index` 와 같은 모양, 대상별 fallback 유지) | F3 |
| `resolve_landing` · `compute_landing` | 번들 목록·하한·수리 신호를 **한 번** 재서 공유 · 수락 직전에 목록이 그대로인지 재확인 | F2 · P1-2 |
| `check` · `_recording_refusal` | 손 복사 예외 목록 넷 → `GATE_FAULTS`(`SubprocessError` 가 빠져 타임아웃이 창 줄을 삼켰다) | 서브에이전트 |

**안 하는 것**: Codex P2-4(메모리 고갈) — 서브에이전트가 **실물 코퍼스 94건**을 재서 최대 배치
0.95MB `ast.json` + 1.95MB 소스라 비쟁점이다. Codex 는 인위적 160MiB 한도로 쟀다. 실측을 따른다.
F8(다른 세션의 `.claude/settings.json` 훅)은 **내 것이 아니라** 손대지 않고 사람에게 보고한다.

### Pre-Edit 선언

편집 전 AST 일곱을 **git revision `b29e1f4e` 를 명시해** 뽑았고 소스 해시가 그 커밋과 일치한다
(`ast.before-7.5.1.json`) — 7.5 에서 열 중 일곱이 편집 도중 상태로 뽑혔던 것의 수리다.
High-risk 아님(생산 거래 코드가 아니라 게이트 도구). 판정은 **permissive 두 자리만** 바뀌어야 하고
나머지는 A/B 로 불변을 증명한다.

## VERIFY — task 7.5.1 (gstack 리뷰의 permissive 결함 수리) (2026-09-19)

**생산 Go 코드 변경 0.** 게이트 도구만 바꿨다. 판정은 **정상 git 아래의 실물 코퍼스에서 한 글자도
안 바뀌고**, git 이 실패하거나 출력이 망가지거나 증거가 동시에 쓰일 때만 바뀐다 — 그리고 그때
바뀌는 방향은 전부 **거절 쪽**이다.

### 무엇을 고쳤나 (출처별)

| 출처 | 결함 | 수리 |
|---|---|---|
| 서브에이전트 F1 (**재현**) | 가드 7 이 한쪽 git 실패를 "바뀌었다" 로 읽어 **편집 전 커밋을 기록** | rc≠0 → 결함. `None` 은 이제 "git 이 물은 spec 그대로 `missing` 이라 답했다" 하나뿐 |
| Codex P1-1 (**재현**) | `_landing_record` 가 실패를 "기록 없음" 으로 읽어 착지 검증 생략 | 같은 수리 |
| Codex P1-2 (**재현**) | 걷는 도중 생긴 번들을 얼린 목록이 못 봄 — 7.5 의 회귀 | 입력 한 벌 + **수락 직전** 증거 지문 재확인 |
| 서브에이전트 F2 | "한 번 잰다" 주석이 거짓(실측 둘씩) | `_measure_landing_inputs` 한 곳 — 선언·계산 경로가 한 벌 |
| 서브에이전트 F3 | `validate_target`(본 판정 경로) · 가드 7 이 아직 소스마다 프로세스 | 착지에서 한 번 미리 읽기 · 가드 7 양쪽 한 번씩 |
| F4 · F5 | 잘림 문장이 blob 만 셈 · 넘침을 `left -6` 으로 거꾸로 말함 | 레코드를 셈 · 넘침은 "truncated" |
| F6 | stderr 를 받아 놓고 버림 | 결함 문장에 git 의 첫 줄 |
| F7 | NUL 검사가 경로만 봄 | 요청 문자열 전체 |
| Codex P2 | 모르는 머리를 "없음" 으로 · 내용 뒤 NUL 종단 미확인 | 아는 두 모양만(실측) · 종단 확인 |
| F9 | `GIT_BATCH_MINIMUM` 에 근거 없음 | git 릴리스 노트 `RelNotes/2.42.0.txt:150` 인용 |
| 서브에이전트 | 손 복사 예외 목록 넷이 `SubprocessError` 누락 → 타임아웃이 창 줄을 삼킴 | `GATE_FAULTS` 한 곳 |

**안 한 것**: Codex P2(메모리 고갈) — 서브에이전트의 **실물 코퍼스 94건 실측**(최대 0.95MB+1.95MB)이
비쟁점이라 답했고 Codex 는 인위적 160MiB 한도로 쟀다. 실측을 따른다. **F8**(다른 세션의
`.claude/settings.json` 훅) — 이 lot 의 것이 아니라 손대지 않고 사람에게 보고했다.

### 되돌렸던 판단의 정정

1.4 수리 때 rc≠0 을 결함으로 올리려다 "시험 21개가 저장소 아닌 곳에서 돈다" 는 이유로 되돌렸다.
**그 판단이 permissive 두 자리를 남겼다.** 다시 재니 21개는 `run_check` 와 인라인 11개가 git 을
mock 하려다 `_landing_record` 만 빠뜨린 픽스처였다. 빈 저장소로 만들면 전제("기록 없음")가 그대로다
— 실측 `HEAD:x.go missing` rc 0. 수리 후 깨진 시험은 옛 동작을 단언하던 **내 시험 둘뿐**이었다.
그때 내가 센 것은 "거부할 정상 입력"이었고, **거부하지 않아서 생기는 피해**는 안 셌다.

### 증거

- **엄격한 파서가 거부할 정상 입력 0** — git 머리를 실측했다(blob · tree · 심링크 · gitlink · 없는
  경로 · HEAD · 짧은 ref): 정확히 두 모양 `<hex oid> <type> <size>` · `<물은 spec> missing`.
- **판정 A/B 두 겹**, 기준 `b29e1f4e`: `compute_landing` **116/116 SAME** · `check()` **전체**
  **126/126 SAME**(활성+아카이브, 오류 줄 목록 단위 — 7.5 의 A/B 가 `compute_landing` 만 덮었다는
  서브에이전트 지적을 닫는다). 하네스 `analysis/harness/751_check_ab.py`.
- **편집 전후 AST**(before = revision `b29e1f4e` 명시): `_landing_refusal` 11→11 에서 **가드 7 의 조건
  한 줄만** 바뀌고 가드 여덟의 순서는 그대로 · `resolve_landing` 순서열 바이트 동일 · 나머지는 예외
  목록 교체와 미리 읽기·재확인 갈래 추가뿐.
- **변이 31/31 CAUGHT, 생존 0**(양성 대조 포함). 첫 판 생존 둘은 양성 대조가 **둘 다 닿았다**고 말했다
  — 계측기가 아니라 시험 공백이었다. **V13**(가드 7 소스별): `all()` 이 첫 번째로 다른 소스에서 멈추므로
  **받는** 픽스처에선 비용이 같다 → **거절하는** V1 모양으로 쟀다. **V10**(지문을 하한 뒤에): 첫 시험은
  B 를 `record_landing` 의 **걷기 전** 하한 측정에서 써 버려 두 판본이 다 B 를 봤다 → 주입을
  `_measure_landing_inputs` 가 스택에 있을 때로 좁혔다. 둘 다 **코드가 아니라 시험을 고쳐** 잡았다.
- 실물 a099(유일한 착지 기록): blob 읽기 프로세스 **46 → 10**, `check()` 5.64s → 5.27s, 오류 0 → 0.
- 시험 187 → **203** · `make sdd-test`(logic-map 278) · `make lint` · `make test-seams` ·
  `openspec validate --all --strict` 58/58 · a122 rc=0.

**남은 창 하나(기록)**: 지문 재확인은 `compute_landing` 의 수락 직전에 선다. 그 뒤 `check()` 의
`validate_target` 이 디스크를 다시 읽기까지의 짧은 창은 닫지 않았다 — 거기서 번들이 바뀌면 판정이
읽은 바이트와 착지가 든 바이트가 갈릴 수 있다. 7.5 이전에도 같은 모양으로 있던 창이고, 닫으려면
`check()` 전체를 한 지문에 묶어야 한다(설계 변경).

## MEASURE · Pre-Edit Gate — task 7.5.2 (수리한 트리의 재리뷰가 연 것) (2026-09-19)

gstack `/review` 를 **수리 로트** `a5c4bc77..e9f905bd` 에 다시 돌렸다. 출처 여덟 — Claude 전문가
다섯(testing · maintainability · security · performance · simplification) · 적대 서브에이전트 ·
레드팀 · Codex. 사람이 "이름 바뀐 파일 규칙을 뺀 나머지 전부" 를 골랐다(그것은 7.5.3, 사람 결정 대기).

### 지난 수리의 확인

| 지난 결과 | 판정 | 근거 |
|---|---|---|
| P1-1 · F1 (실패가 부재로) | **확인** | 새 시험을 옛 모듈에 돌리면 빨강, HEAD 에서 초록(적대 서브에이전트) · Codex 주입 |
| F2 (입력 두 번) | **확인** | `compute_landing(..., inputs)` |
| F3 (배치) | **확인** | 가드 7 · 미리 읽기 |
| 예외 목록 넷 | **확인** | `except GATE_FAULTS` 넷 |
| P1-2 (걷는 중 번들) | **반쯤** | 아래 C2 |

### 내가 직접 확인한 것 (리뷰어의 주장을 그대로 옮기지 않았다)

- **C3 — 시험이 공허하다.** `test_placeholders_are_rejected` 는 저장소 아닌 임시 디렉터리에서 돌고
  `len(errors) > 0` 만 본다. 7.5.1 뒤로 `check` → `resolve_landing` → `_landing_record` 의 git 이 rc 128
  로 죽어 `cannot derive … not a git repository` 한 줄이 나오고, 그 줄이 단언을 만족한다. 자리표시자
  가드 삭제가 `a5c4bc77` 에서는 잡히고 HEAD 에서는 산다(testing 전문가의 변이). 레드팀이 스위트 전체를
  두 판본에서 **시험별 출력으로** 대조했더니 통과 이유가 바뀐 시험은 이것과 무해한 하나(타임아웃 줄의
  출처만 바뀜)뿐이었다. 7.5.1 때 나는 **깨진** 시험을 셌고 **통과 이유가 바뀐** 시험을 안 셌다.
- **C1(→7.5.3) — 이름 바뀐 파일.** 보안 전문가의 픽스처를 그대로 돌렸다: 순수 `git mv` → FLM 먼저 →
  편집(번들 안 갱신) → `--record-landing` rc 0 → 게이트 rc 0 `required 0`, 대조군(이름 안 바꿈)은 rc 1.
  `a5c4bc77` 모듈로 돌려도 같다 — 이 로트가 만든 것이 아니라 7.2.2 규칙의 경로 비교가 연 것이다.
- **C2 — 두 번 읽기.** `_measure_landing_inputs` 는 `_evidence_fingerprint`(자기 `_pinning_bundles`)와
  `_walk_floor`(또 `_pinning_bundles`)가 목록을 **따로** 읽는다. `_ast_value` 는 읽기 실패를 `{}` 로 삼켜
  그 번들을 고정 목록에서 뺀다. 둘째 읽기에서만 B 가 한 번 실패하면 B 는 판정에서 빠지는데, 첫 지문과
  수락 직전 지문은 둘 다 B 를 담아 **같다** — 착지가 B 를 판정하지 않고 선다(Codex 재현). 성능
  전문가가 같은 자리를 다른 축으로 셌다: 착지 있는 `check()` 한 번에 증거 디렉터리 순회 4회(로트 전 2회).

### 편집 집합

| 함수 | 무엇 | 출처 |
|---|---|---|
| 새 `_read_evidence` · `_select_pinning` · `_parsed` · `_decoded` · `_head_commit` | 증거 읽기를 **한 곳**에서 한 번: 각 `ast.json` 의 바이트 → (해시, 파싱) 이 같은 바이트에서 나온다. 선별(`revision: current` · `file` · `source_sha256`)도 한 곳 | C2 |
| `_pinning_bundles` · `_ast_value` | 위의 둘에 위임(철자 한 곳) | C2 |
| `_evidence_fingerprint` | 지문 = (HEAD, 모든 `ast.json` 의 (경로, 바이트 해시), 고정 목록) — **판정 목록과 같은 읽기**에서 만들고 목록도 같이 돌려준다 | C2 · 레드팀(재작성) · F9(HEAD) |
| `_measure_landing_inputs` · `LandingInputs` | 한 번 읽기 · `NamedTuple` · 하한이 못 서면 수리 신호는 `None`(잰 적 없음) | C2 · M(위치 튜플) · M(`[]` 가 "없음"으로 읽힘) |
| `_walk_floor` | 이미 읽은 목록을 받을 수 있다(기록 명령의 사전 판정은 그대로 스스로 읽는다) | C2 |
| `compute_landing` | 필드 이름으로 읽기 · 움직인 입력 문장을 상수로 | C2 · 메시지 |
| `resolve_landing` | 거절을 내기 **전에** 지문을 다시 봐서 움직였으면 "다시 돌려라" · 받으면 판정한 지문을 `facts` 로 넘긴다 | 적대 F5 · 수락 뒤 창 |
| `_repairs_after` | 수리 신호가 `None`(잰 적 없음)이면 결함 | M |
| `check` | 착지가 판정한 바이트에 **묶는다**: `ast.json` 을 한 번 읽어 지문과 대조하고 그 바이트를 `validate_target` · 열거형 호출 판정에 준다 · 미리 읽기는 판정한 목록을 쓰고(없으면 선별을 `GATE_FAULTS` 안에서) · 조언 계산의 결함은 조언 줄로 | 수락 뒤 창 · C2(순회 수) · 이관 회귀(Codex P2 · 적대 F3) · 적대 F6 |
| `validate_target` | 미리 읽은 `ast.json` 바이트를 받는다 · `start`/`end` 가 사전이 아니거나 `branches` 가 목록이 아니면 "invalid"(traceback 대신) | 수락 뒤 창 · 레드팀(7.5 이전부터) |
| `_base_shaped_bundles` | base 읽기를 한 프로세스로 | 적대 F6 |
| `_committed_many` | `<oid> submodule` 을 blob 아님(`None`)으로 · git 버전 조언은 rc 129(사용법 오류)일 때만 · 파서 오류에 물은 경로 · 잘림 문장 한 벌 | Codex P2 · M · 적대 F7 |
| `main` | 조언 계산 결함을 조언 줄로 | 적대 F6 |

문서·주석(`GIT_BATCH_MINIMUM` 위 문장 · `_pinning_at` "번들은 안 변한다" · README 코드 스팬과 콜론 ·
WORKFLOW 인용 · 고아 주석 · `_walk_floor` 호출자 · `_committed_many` 의 `None` 계약(트리 안인데 못 읽는
객체도 git 은 `missing` 이라 답한다 — 보안 전문가 실측) · 빈 줄 · 7.5.1 문서의 과장: `_landing_record`
는 거절 하나를 건너뛰었지만 언제나 **가장 넓은 창**으로 갔다 — permissive 로 창을 좁힌 것은 가드 7
뿐이었다(적대 서브에이전트)) · 하네스 `75_time.py`·`75_attrib.py` 가 새 raise 에 traceback.

**`<oid> submodule` 의 영수증**: git 소스 `builtin/cat-file.c` —
`if (data->mode == S_IFGITLINK) report_object_status(opt, NULL, &data->oid, "submodule");` 이고
`report_object_status` 는 `printf("%s %s%c", obj_name ? obj_name : oid_to_hex(oid), status, …)` 다.
이 기계의 git 2.43.0 은 같은 경우에 `<spec> missing` 을 낸다(7.5.1 실측) — 두 모양 모두 blob 이 아니다.

### 거부할 정상 입력 ([[fail-closed-must-name-what-it-rejects]] — 그리고 거부 **안** 할 때 통과하는 것)

| 새 거절 | 정상 입력에서 | 거절 안 하면 통과하는 것 |
|---|---|---|
| `check` 의 바이트 묶음 | 0 — 동시 쓰기에서만 선다(A/B 로 확인할 것) | 착지가 판정 안 한 번들로 판정 |
| 지문의 HEAD | 0 — 실행 중 커밋에서만 | 잰 수리 신호보다 새 역사 위의 착지 |
| `ast.json` 모양 | **0 / 3,048**(저장소 전수: `start`·`end` 사전 아님 또는 `branches` 목록 아님) | traceback(판정 줄 없음) |
| 수리 신호 `None` | 도달 0 — 하한 없음이 그 앞에서 거절(시험으로 못 박음) | "잰 적 없음" 을 "없음" 으로 |

`<oid> submodule` 은 **느슨해지는** 쪽이다(결함 → `None`). `None` 의 소비자 일곱을 Codex 가 표로
셌다 — 전부 거절·누락·조언 쪽이고 창을 좁히는 자리는 없다.

### 안 하는 것

- **가드 7 의 base 쪽 반복 읽기**(후보마다 한 프로세스). 성능 전문가 실측: 활성 전수 합계 약 1.2s,
  어느 walk 에서든 최대 12%(a094). 없애려면 호출 사이에 상태(캐시)를 두거나 `_pinning_at` 의 반환을
  넓혀야 하는데, 판정이 기대는 상태를 늘리는 것이 이 로트의 방향(한 번 읽기)과 어긋난다. 후보 쪽 중복
  한 프로세스도 같은 이유로 둔다. 숫자는 review 에 남긴다.
- **7.5.3(이름 바뀐 파일)** — 규칙 변경이라 사람 결정.

### Pre-Edit 선언

편집 전 AST 열셋을 **revision `e9f905bd` 를 명시해** 뽑았고 소스 해시가 그 blob 과 같다
(`db2443db8f…`). High-risk 아님(게이트 도구). 정상 입력에서 판정은 **한 글자도** 안 바뀌어야 한다 —
`check()` 전체 A/B 를 `e9f905bd` 기준으로 전수 돌린다. RED 를 먼저 세운다.

## VERIFY — task 7.5.2 (판정은 한 번 읽은 바이트로 선다) (2026-09-19)

### 무엇이 바뀌었나 — 한 문장

**착지 판정과 그 뒤의 판정이 `ast.json` 을 같은 한 번의 읽기로 본다.** 7.5.1 은 지문과 판정 목록을
따로 읽었고(Codex P1), 후보마다 "드는가" 를 디스크에서 다시 읽었고(ABA), 수락 뒤 `check()` 가 대상·열거형
호출·번들 글자·요구 대조마다 또 읽었다(수락 뒤 창). 이제 읽는 자리는 `_read_evidence` 하나이고, 착지의
지문은 그 바이트 + 고정 목록 + `HEAD` 이며, `check()` 는 자기 읽기를 그 지문과 대조한 뒤 그 바이트만 쓴다.

### RED 가 빨갰던 이유 (구현 전, `e9f905bd` 코드 위에서)

새 시험 17 이 실패했다 — 각자 **자기 사유로**(오류 문장 인용은 세션 기록). 핵심 둘은 내가 직접 재현했다:
- Codex P1: 번들 B 의 **셋째** 읽기(하한의 목록)만 실패시키면 `record_landing` rc 0 으로 착지를 **기록했다**.
- 수락 뒤 창: 커밋된 증거(지도에 없는 분기 B2)로 착지가 판정된 뒤 통과하는 증거로 갈아 끼우면 `check` 가 `[]`.

### 편집 순서 이탈 — 적는다

Pre-Edit 에서 선언한 열셋 밖의 넷(`_unheld_bundles` · `_bundle_text` · `_landing_refusal` ·
`_recording_refusal`)을 **GREEN 도중에** 편집 집합에 넣었다. 첫 GREEN 뒤 내 코드를 다시 읽다가 `check()`
안에 `ast.json` 을 디스크에서 다시 읽는 자리가 셋 더 있고(`_bundle_text` · 열거형 호출 · 요구 대조),
walk 의 `_unheld_bundles` 도 후보마다 다시 읽는다(ABA)는 것을 찾았다. 넷의 편집 전 AST 는 편집 **뒤에**
revision `e9f905bd` 를 명시해 뽑았다 — 내용은 그 커밋의 blob 이라 정확하지만 **순서는 FLM-first 가
아니다.** `_landing_refusal` · `_recording_refusal` 은 분기·반환·raise 순서열 차이 0(인자·언팩만)이고,
`_unheld_bundles` · `_bundle_text` 는 새 갈래가 각 하나씩이며 둘 다 변이로 잡힌다(W9 · W11).

### 처음 넣었다 뺀 것

`resolve_landing` 의 "선언값 ≠ 계산값" 앞 재확인. 판정이 잰 바이트(`held`)와 불변 커밋만의 함수가 된
뒤로는, 선언값이 거절 가드를 통과하면 같은 입력의 계산이 그 값 이하에서 **받고** 받는 순간
`compute_landing` 이 재확인한다 — 그 갈래로 움직인 입력이 오는 길이 없다. 닿을 수 없는 갈래에 둔
가드는 시험도 변이도 못 잰다([[uncovered-block-may-be-held-dead-by-a-coincidence]]). 거절 앞 재확인은
남겼다: 그 거절은 **잴 때** 중간 상태였던 바이트로 날 수 있다(시험이 그 모양을 만든다).

### 증거

| 무엇 | 결과 |
|---|---|
| `check()` 전체 A/B (기준 `e9f905bd`, 활성+아카이브 전수) | **비교 126 · SAME 126 · DIFFERENT 0**(예외는 타입·문장까지), 합계 2048.9s → 1934.5s |
| 변이 (`75_mut.py`, 스위트 **전체**) | **52/52 CAUGHT** — 7.5.1 의 30(앵커를 현재 소스로) + 7.5.2 의 22. 첫 판 생존 W23(못 읽은 증거) 은 양성 대조 **안 닿음** → 닿는 시험을 더해 잡음 |
| `_landing_refusal` 가드 순서 | 편집 전후 분기·반환·raise **순서열 차이 0** |
| a099(실물 착지 기록) `check()` 한 번 | `ast.json` 읽기 **446 → 114**(번들당 11.7 → 3.0) · cat-file 10 → 10 · git 82 → 84(지문의 `HEAD` 둘) — `752_reads.py` |
| 시험 | 203 → **231** · `make sdd-test`(logic-map 306) · `make lint` · `make test-seams`(100 패키지) · `openspec validate` 58/58 · a122 rc=0 |
| 거부할 정상 입력 | `ast.json` 모양 검사 0/3,048 · 바이트 묶음 · `HEAD` 는 동시 쓰기에서만 선다 — A/B 가 확인 |

하네스 둘을 고쳤다: `751_check_ab.py` 의 이어 달리기 기록을 **양쪽 소스 해시**에 묶었다(기준만으로 묶던
판본에서 편집 도중 다시 돌렸더니 옛·새 after 로 잰 줄이 한 표에 섞일 뻔했다 — 멈추고 버렸다).
`75_mut.py` 는 스위트 전체를 돈다(부분집합이면 새 시험 클래스를 목록에 안 넣는 것만으로 변이가 산다).
`75_time.py` · `75_attrib.py` 는 새 raise 를 한 change 의 결과로 적고 다음으로 간다.

### 남은 것

- **가드 7 의 base 쪽 반복 읽기**: 안 했다(Pre-Edit 의 사유 — 호출 사이 상태를 늘린다, 활성 전수 ~1.2s).
- **7.5.3(이름 바뀐 고정 소스)**: 사람 결정 대기.
- 남은 창: 없다고 **주장하지 않는다.** `ast.json` 밖의 번들 파일(지도 · 표)은 착지가 판정하지 않으므로
  (7.2 가 실측으로 범위를 `ast.json` 으로 정했다) 판정 도중 바뀌면 그 바이트로 판정한다 — 이것은 창이
  아니라 착지 판정의 **범위**다. 그리고 수락 직전 재확인과 `check()` 의 대조 사이에 바뀌었다 돌아오는 것은
  둘 다 **같은** 지문과 대조하므로 `check()` 가 읽은 바이트가 지문과 다르면 멈춘다 — 돌아온 뒤 읽었다면
  판정한 바이트와 같다.

## MEASURE · Pre-Edit Gate — task 7.5.2.1 (판정이 읽는 입력을 전부 세고 하나씩 묶는다) (2026-09-20)

7.5.2 의 재리뷰(로트 `e9f905bd..fc35eb2d`, 출처 여덟)가 그 로트의 중심 주장 셋을 깼다. 사람이 "전부
같은 절차로 고친다"(7.5.3 은 여전히 사람 결정)를 골랐다. 이번엔 편집 전에 **판정이 읽는 입력의 목록**을
먼저 센다 — 7.5.2 는 바이트만 묶고 "묶었다" 고 적었고, 목록을 세지 않아서 역사(`HEAD`) · 디렉터리
존재 · 목록 열거 셋을 놓쳤다.

### 재리뷰가 깬 것 (재현된 것만)

| 주장 | 깬 입력 | 재현 |
|---|---|---|
| "`HEAD` 도 지문에 있다" | 지문은 `HEAD` 를 **표본**으로 한 번 잴 뿐이고, 착지 기록은 그 표본 **전에** 상징 `HEAD` 에서 읽혔으며 뒤의 git 호출은 전부 살아 있는 `HEAD` 를 다시 읽었다 | 나: 떠났다 돌아온 `HEAD`(ABA) rc 0 · 레드팀 `repro_b`: 가지 전환 **한 번** rc 0(전환 전 `HEAD` 도 후 `HEAD` 도 각자 rc 1) |
| "수락 뒤 창을 닫았다" | `analysis.exists()` 조기 반환이 묶기 **앞**에 있다 · `_bundle_text` 가 살아 있는 `glob("*")` 로 목록을 만든다 | Codex: 스냅숏 뒤 `ast.json` 을 지우면 fc35eb2d `[]`, e9f905bd 는 `missing ast.json` |
| "선언값 ≠ 계산값 갈래는 도달 불가" | `HEAD` 이동 · git 결함이 거절 사유가 된다 | 레드팀 `repro_a`: rev-list rc 128 · `reset --soft` → 복구 조언(기록을 지워라) |
| 모양 틀린 `ast.json` 은 판정 줄 | `calls`/`returns` 가 목록 아님 · 최상위가 사전 아님 — 열 가지 | 나: 10/27 이 아직 traceback |
| 중심 주장의 시험 | 묶은 뒤 둘째 읽기 · 추가/제거 절반 · 새 지문 | testing 전문가 변이 X1~X4 생존 |

### 입력 목록 — AST 로 셌다 (`analysis/harness/7521_inputs.py`)

진입점 `check` · `record_landing` · `main` 에서 닿는 모든 함수의 모든 I/O 호출(`subprocess.run` ·
파일시스템 메서드 · `os.environ` · 상징 `"HEAD"` 상수)을 기계로 열거했다. 첫 판은 `subprocess.run`
안의 `HEAD` 만 세어 `_landing_record` 의 `_committed_bytes(root, "HEAD", …)` 를 놓쳤다 — 상수를 세게
고쳤다. 아래 표의 "읽는 자리" 는 그 출력이다(fc35eb2d 줄 번호).

| 입력 | 읽는 자리 (fc35eb2d) | 오늘 | 7.5.2.1 |
|---|---|---|---|
| **역사(`HEAD`)** | `_landing_record` 572(빌리는 쪽 · 이관 · 선언 셋이 부른다) · `resolve_landing` 1102 · `_evidence_floor` 751 · `_self_repair_commits` 859 · `compute_landing` 1807 · `_repairs_after` 928 · `_head_commit` 1699(표본) · `_recording_refusal` 1911(깨끗한가) · `_commits_after` 553(창 줄) | 상징 `HEAD` 를 **열 자리**에서 따로, 표본 하나 | **고정**: 명령마다 `_head_commit` **한 번** → 그 sha 를 모든 역사 읽기에. 지문에서 `head` 는 빠진다(움직일 것이 없다) |
| 증거 디렉터리 **존재** | `_read_evidence` 604(`is_dir`) · `check` 1584(`exists`, 조기 반환) | 두 번, 조기 반환은 묶기 앞 | 한 번 읽기 `Evidence.present` — 조기 반환이 그것을 본다 |
| 증거 **목록** | `_read_evidence` 604(`glob("*/ast.json")`) · `check` 1595(`iterdir`) · `check` 1627(`glob("*/function-logic-map.md")`) | 세 번, 두 종류 — 검색만 되는 디렉터리를 `glob` 은 빼고 `iterdir` 는 넣는다 | `Evidence.targets` 한 번(`iterdir`) · `ast.json` 은 `target/ast.json` 을 **직접** 읽는다 |
| `ast.json` **바이트** | `_read_evidence` 606(재기 · 판정 · 재확인이 각자 부른다) · `_base_shaped_bundles` 670(→`_pinning_bundles`, 따로 읽기) · `_unheld_bundles` 803 · `_bundle_text` 1337 · `validate_target` 1397(뒤 셋은 인자가 없을 때 — 시험만) | 착지 경로에서 판정 읽기 둘(잴 때 · 판정할 때) + 재확인 | `check`/`record_landing` 이 **한 번** 읽어 착지와 판정에 **같은 값**을 넘긴다. 재확인은 비교만 한다. 시험 전용 디스크 fallback 은 없앤다(인자 필수) |
| 번들 **파일 목록**(산문) | `_bundle_text` 1325(`glob("*")` + `is_file`) | 살아 있는 목록 — `ast.json` 이 목록에서 빠지면 판정이 든 바이트도 빠진다 · 못 여는 디렉터리는 조용히 빈다 | `ast.json` 은 목록과 **무관하게** 한 번 읽은 바이트 · 산문 목록은 `iterdir()` 로 — 못 열면 **이름 댄 판정 줄**(조용히 비지 않는다) |
| 산문 파일(FLM · BTM · 위험 보고 · 그 밖) | `validate_target` 1397 · `_bundle_text` 1337 | 살아 있다 | **설계상 살아 있다 — 묶지 않는다.** 착지는 이것을 판정하지 않는다(7.2 실측: "드는가" 등식을 산문까지 넓히면 정상 착지 76건 중 3건을 잃는다). 창이 아니라 착지 판정의 **범위**다 |
| 소스 경로의 심링크 풀이 | `normalized_source` 1158(`resolve`) — 고정 목록(착지 · 재확인)과 `validate_target` 이 각자 | 두 번 | 고정 목록은 수락 직전 재확인이 이미 비교한다(그대로). 판정 쪽 풀이는 **살아 있다 — 적는다**: 둘이 갈려도 판정이 통과하려면 두 경로가 착지에서 **같은 digest** 를 들어야 하고, 그러면 번들은 둘 다를 사실대로 기술한다. 심링크 고리는 `RuntimeError` → `ValueError`(대상 이름으로) |
| 착지의 소스 | `_committed_many(landing)` | 불변 객체 | 그대로 |
| 워킹트리 소스 | `validate_target`(착지 없을 때) · `changed_existing_functions(target="")` · `_base_shaped_bundles`(조언) | 살아 있다 | 그대로 — 착지가 없으면 창의 끝이 **정의상** 워킹트리다 |
| 시험 파일 | `test_index` · `resolve_test_file` · `test_citation_errors` | 살아 있다 | 그대로(착지가 판정하지 않는다) |
| `base-commit.txt` · `execution-baseline.json` · `function-logic-reference.txt` · `review.md` · `SDD_BASE_REF` | 각 한 번 | 한 번 | 그대로. `execution_baseline.validate` 는 `HEAD` 를 **자기 안에서 한 번** 읽는다(a063 하나, 창 끝은 감사된 source) — 이 로트 밖 |
| **git 의 답(결함)** | `_is_ancestor` 530(rc 128 → "조상 아님") · `_evidence_floor` 756(rc≠0 → "") · `compute_landing` 1811(rc≠0 → 사유) · `_recording_refusal` 1915(rc 128 → "더럽다") | 결함이 **판정**으로 읽혀 복구 조언까지 붙는다 | 결함은 `RuntimeError`(`GATE_FAULTS`) — 7.5.1 이 `_committed_many` 에만 한 것을 나머지에 |

### 편집 집합 (편집 전 AST 30개 — revision `fc35eb2d` 를 **명시**, 소스 해시 `98845ce2dea1…` = 그 blob)

| 함수 | 무엇 | 출처 |
|---|---|---|
| `_head_commit` → 진입점 셋 | `check` · `record_landing` 이 **먼저** 푼다; `main` 은 `check` 가 푼 것을 쓴다 | 적대 · 레드팀 · 나 |
| `_landing_record` · `_declared_landing` · `_is_ancestor`(인자) · `_evidence_floor` · `_self_repair_commits` · `_repairs_after` · `compute_landing` · `_recording_refusal` · `_commits_after` | 상징 `HEAD` 대신 받은 sha | 같음 |
| `_is_ancestor` · `_evidence_floor` · `compute_landing` · `_recording_refusal` | rc 가 "예/아니오" 가 아니면 결함 | 레드팀 `repro_a`·`repro_a3` |
| 새 `Evidence` · `_read_evidence` | 존재 · 목록 · 바이트를 **한 값**으로. `iterdir` 로 번들을 세고 `target/ast.json` 을 직접 읽는다: 없음(`FileNotFoundError`) = 키 없음, 못 읽음 = `None` | 수락 뒤 창 · Codex P2 |
| `_digests` · `_pinning_bundles` · `_evidence_fingerprint` · `Fingerprint` | 지운다 — 지문은 `Evidence` 자체(바이트 비교, 해시 불필요)와 고정 목록 | 단순화 · 둘 이름 셋 |
| `_select_pinning` · `_base_shaped_bundles` | `Evidence` 를 받는다 — 조언도 같은 한 번 읽기 | 수락 뒤 창 |
| `LandingInputs` · `_measure_landing_inputs` · `_walk_floor` · `_landing_refusal` · `_unheld_bundles` | 입력 한 벌(`head` · `evidence` · 고정 목록 · 하한 · 사유 · 수리)을 **받는다** — 스스로 읽는 기본값 없음 | maintainability(시험 전용 `None`) |
| `_raise_if_inputs_moved` | 증거를 다시 읽어 **먼저** 바이트로 대조하고, 같을 때만 고정 목록 — 고르다 난 `ValueError` 도 "움직였다" | 레드팀(escapes 가 MOVED 앞) |
| `resolve_landing` | `head` · `evidence` 를 받고 `facts` 를 안 받는다. 선언값 ≠ 계산값 갈래의 주석을 **참인 이유**로 | 레드팀 |
| `normalized_source` | 심링크 고리를 `ValueError` 로 | 보안 전문가 |
| 새 `_ast_value` · `_parsed` · `validate_target` | 모양 검사를 **파싱하는 자리 한 곳**에: 사전 · `start`/`end` 사전 · `branches`/`calls`/`returns` 목록. `validate_target` 은 `held` 필수 | 모양 열 가지 |
| `_bundle_text` | `(대상, 한 번 읽은 ast 바이트)` — 목록과 무관 · 목록은 `iterdir` | Codex P1 |
| `check` | 고정 → 빌리는/이관 탐침 → **한 번 읽기** → 착지 · 판정 · 조언이 그 값. 앞선 실행의 사실은 **전부** 지운다 | 수락 뒤 창 · pop 절반 |
| `record_landing` | 고정 → 한 번 읽기 → 사전 판정과 걷기가 같은 값 | 같음 |
| `main` | 창 줄 · 조언이 `check` 가 푼 sha · 출력은 `backslashreplace` | 레드팀(서로게이트) |
| `_committed_many` | 주석 하나("물은 spec 그대로" — gitlink 는 oid 로 온다). AST 불변 | maintainability |

### 거부할 정상 입력 — 그리고 거부 **안** 할 때 통과하는 것

| 새 거절·결함 | 정상 입력에서 (실측) | 안 하면 통과하는 것 |
|---|---|---|
| 명령 시작에 `HEAD` 를 푼다 | 저장소 0 — 거절되는 모양은 **태어나지 않은 `HEAD`**(첫 커밋 전 고아 가지)뿐이다. 그 모양은 옛 판본에서 base → 워킹트리 창으로 판정됐고 이제 결함 줄이다(막는 쪽) | 두 역사를 섞은 판정(rc 0) |
| git 결함 넷 → 결함 | 0 (A/B 로 확인할 것) | 결함이 거절로 읽혀 **유효한 기록을 지우라**는 조언 |
| `ast.json` 모양 | **0 / 3,048**(`start`·`end` 사전 아님 · `branches`·`calls`·`returns` 목록 아님 · 최상위 사전 아님 · 글자 아님 전부) | traceback |
| 목록을 못 여는 번들 디렉터리 | **0 / 3,048** | 열거형 호출 판정이 그 번들의 산문을 조용히 잃는다(permissive) |

**느슨해지는 쪽도 적는다**: 실행 중 커밋은 이제 판정을 멈추지 않는다 — 판정은 **고정한 sha 의 역사**를
말한다. 7.5.2 는 그것을 `INPUTS_MOVED` 로 멈췄는데, 그 멈춤이 지키던 것(뒤섞인 역사)은 고정이 없앴다.

### Pre-Edit 선언

High-risk 아님(게이트 도구, 생산 Go 변경 0). 정상 입력에서 판정은 **한 글자도** 안 바뀌어야 한다 —
`main()` 출력 전체(판정 줄 + 창 줄 + 조언 줄)의 A/B 를 `fc35eb2d` 기준으로 전수, **순서를 번갈아** 돌리고
시간 대신 **결정적 계수**(git 프로세스 수 · `ast.json` 읽기 수)를 적는다(7.5.2 의 시간 이득은 대부분
실행 순서 표류였다 — performance 전문가). RED 를 먼저 세운다.

## VERIFY — task 7.5.2.1 (명령 하나는 역사 하나와 증거 읽기 하나 위에서 판정한다) (2026-09-20)

### 무엇이 바뀌었나 — 한 문장

**명령(`check` · `record_landing`)이 시작할 때 `HEAD` 를 sha 로 한 번 풀고 증거 디렉터리를 한 번 읽어,
착지 판정 · 대상 판정 · 조언이 전부 그 두 값만 쓴다.** 7.5.2 는 바이트만 한 번 읽어 착지에 묶고 `check()` 가
다시 읽어 대조했다 — 그 대조 **앞**의 조기 반환과 대조 **밖**의 목록 열거가 디스크를 다시 물었고, 역사는
표본 하나만 비교해서 표본 **전**의 읽기와 떠났다 돌아온 `HEAD` 가 뚫렸다. 같은 값을 넘기면 대조할 것이 없다.
수락 · 거절 앞 재확인은 남았고, 이제 **대조만** 한다(그 읽기의 바이트는 판정에 안 들어간다).

### 7.5.2 가 쓴 문장 중 거짓이었던 것 — 정정

- "수락 뒤의 창이 닫혔다"(tasks 7.5.2 · VERIFY 7.5.2): 거짓. `analysis.exists()` 조기 반환과 `_bundle_text` 의
  살아 있는 목록 둘이 남아 있었다(재리뷰 적대 · Codex 재현).
- "선언값 ≠ 계산값 갈래에는 움직인 입력이 못 온다(재확인 불필요)": 거짓. `HEAD` 이동과 git 결함이 둘 다
  왔다(레드팀 repro_a). 이번에 그 둘을 없앴고 주석은 **이제 참인 이유**를 적는다.
- "`HEAD` 도 지문에 넣었다": 표본이었지 고정이 아니었다(적대 · 레드팀 · 나 — 셋이 재현).
- "합계 2048.9s → 1934.5s": 대부분 실행 순서 표류(늘 before 먼저). 판정 SAME 126/126 은 그대로 맞다.

### RED 가 빨갰던 이유 (구현 전, `fc35eb2d` 위에서)

새 시험 25 중 **20 이 실패** — 각자 자기 사유로(출력: 세션 스크래치 `7521_red_on_fc35eb2d.txt`):
가지 전환 한 번 → `[]`(대조군은 "later commit(s) of this change's own Go work") · 떠났다 돌아온 `HEAD` → `[]` ·
상징 `HEAD` 를 읽는 git 호출이 명령마다 여럿 · 실행 중 커밋 → `INPUTS_MOVED`(판정/기록 모두) · git 결함 셋 →
"to move a record" 복구 조언 · `git diff --quiet` rc 128 → "uncommitted changes" · 모양 틀린 `ast.json` →
`TypeError` · 수락 뒤 디렉터리 삭제 → `[]` · 목록에서 빠진 `ast.json` → `[]`(Codex 모양) · 검색만 되는 번들 →
`missing ast.json` · 심링크 고리 → `RuntimeError` · 서로게이트 → `Traceback` · 판정 읽기 2회 · 수락 뒤 교체 →
`SWAPPED` 줄 · 걷는 중 생긴 빈 번들 디렉터리 → 기록됨 · 재사용 문맥의 `adoption_source` 잔존 · 새 API 셋(오류).
나머지 5 는 옛 코드에서도 초록이다 — 회귀를 막는 못(빠진 번들 · submodule `fullmatch` · rc 1 조언 · 멎은 git
한 번 · base 번들)이다.

### 편집 순서 — 이번엔 지켰다

편집 전 AST 30 을 **편집 전에** revision `fc35eb2d` 로 뽑았다(Pre-Edit 표). 편집 뒤 새 함수 둘(`_parse_ast` ·
`_first_line`)과 클래스 하나(`Evidence`)가 생겼고 셋(`_digests` · `_pinning_bundles` · `_evidence_fingerprint`)이
없어졌다. `_landing_refusal` 의 가드 순서는 편집 전후 **분기 순서열 · 반환 순서열 둘 다 같다**.

### 처음 설계에서 바꾼 것 둘

- **심링크 고리**: Pre-Edit 는 "`RuntimeError` → `ValueError`" 였다. 그렇게 쓰자 손 복사 예외 목록을 막는 구조
  시험(`test_every_fault_handler_uses_the_one_list`)이 그 `except RuntimeError` 를 잡았다. 원인을 재 보니 동작이
  **파이썬 판본의 함수**였다(2026-09-20 실측: `Path.resolve()` 가 고리에서 3.12.3 만 raise, 3.13.13 · 3.14.5 는 안 함,
  `os.path.realpath` 는 셋 다 같음). `realpath` 로 바꿨다 — 고리는 "AST source is missing" 이 된다.
- **태어나지 않은 `HEAD`**: Pre-Edit 는 저장소 전수 0 만 쟀다. 스위트를 돌리자 git 을 mock 하는 픽스처 22 개가
  커밋 없는 빈 저장소였다 — 픽스처에 빈 커밋 하나(`_empty_repo`)를 두었다. 픽스처의 전제("HEAD 에 기록 없음")는
  그대로다. 가드는 안 바꿨다.

### 증거

| 무엇 | 결과 |
|---|---|
| `main()` 출력 전체 A/B (`7521_main_ab.py`, 기준 `fc35eb2d`, 활성+아카이브 전수, 순서 번갈아) | **비교 126 · SAME 126 · DIFFERENT 0** — 찍은 줄 전부 + rc, 순서를 번갈아(before 먼저 63 · after 먼저 63) |
| 결정적 계수 (같은 A/B) | git 프로세스 20,019 → 20,133(+114 = 115 change 에서 +1 고정하는 자리 · a099 −1 — 7.5.2 의 표본 둘이 고정 하나로 · base 에서 멈추는 10 change 0) · `ast.json` 읽기 **9,072 → 5,803(−36%)** · 시간은 이득으로 적지 않는다: before 가 먼저 돈 63건 before 1134.2s · after 1103.0s, after 가 먼저 돈 63건 before 1088.7s · after 1115.4s — **먼저 도는 쪽이 ~3% 느리다**(7.5.2 의 2048.9s → 1934.5s 도 이것이었다) |
| 변이 (`75_mut.py`, 스위트 전체, 생존 시 양성 대조) | **76/76 CAUGHT** — 7.5.2 까지의 52 중 48 유지(앵커 14 갱신, 셋은 이름도) · 코드가 없어진 넷(W2~W4 지문 칸 · W7 착지 뒤 대조) 퇴역 · 새 Y1~Y28. 무변이 대조군 초록(skipped 2 — 하나는 `docs/` 가 없는 사본에서 문서 시험이 스스로 건너뛴다, 변이 76 중 그 대상을 건드리는 것 0) |
| a099 `check()` 한 번 (`752_reads.py`) | `ast.json` 읽기 446(7.5.1) → 114(7.5.2) → **76**(번들당 2.0 = 판정 1 + 수락 직전 대조 1) · git 82 → 83 |
| 입력 목록 (`7521_inputs.py`) | `check` · `record_landing` · `main` 에서 닿는 상징 `HEAD` 읽기 **1**(푸는 자리) — 7.5.2 는 10 |
| 시험 | 231 → **254**(+25 새 · +2 재확인 · −4 대체) · `make sdd-test`(logic-map **329**) · `make lint` · `openspec validate` 58/58 · a122 rc=0 · `make test-seams` rc 0 · `make sdd-check` rc 0(CodeGraph hard-evidence 가 워킹트리와 같다) |
| `make sdd-sync` | **rc 2 — advisory 둘만**: CodeGraphContext 는 kuzudb 잠금을 다른 세션의 `cgc mcp start`(가동 1일 6시간)가 쥐고 있어 색인을 못 했고(두 번 돌려 같음 — 시간초과가 아니라 잠금이라 재시도로 안 풀린다), GBrain 은 `gbrain serve` 가 바빠 앞 신선도를 유지했다. `codegraph sync` 는 됐다. 남의 서버라 죽이지 않았다 — 둘 다 advisory(WORKFLOW) |
| 하네스 | `75_census.py` 가 다시 걷는다(93/126) · `75_time` · `75_attrib` · `75_ab` · `751_check_ab` · `752_reads` · `75_levers` 전부 표본으로 돌렸다 · `75_ab` 이어 달리기를 after 소스에 묶었다 |
| 문서 | README 의 `-Z` 없는 git 예시를 실제로 찍어 본 줄로(접두어 · `at <sha12>`) |

### 남은 것

- **7.5.3(이름 바뀐 고정 소스)**: 사람 결정 대기(그대로).
- **살아 있는 입력 — 적었다, 묶지 않았다**: 번들 산문(착지 판정의 범위 밖, 7.2 실측) · 소스 경로의 심링크 풀이
  (판정 쪽 한 번 더 — 갈려도 판정이 통과하려면 두 경로가 착지에서 같은 digest 를 들어야 한다) · 시험 파일 ·
  워킹트리 소스(착지 없을 때 창의 끝이 정의상 워킹트리) · `execution_baseline.validate` 의 자기 `HEAD` 한 번(a063
  하나, 창 끝은 감사된 source).
- **못 읽는 `ast.json` 이 있는 채 기록**: 못 읽은 번들은 고정 목록에 안 들어가 착지가 그것 없이 계산된다(7.5.2
  부터 같다). `check` 는 그 대상을 "could not be read" 로 빨갛게 하고, 읽히게 되면 게이트가 기록을 다시 판정한다
  — 막는 쪽이다. 재리뷰가 짚지 않았고 이 로트에서 안 바꿨다.

## MEASURE · Pre-Edit Gate — task 7.5.2.2 (스냅숏은 끝에서 디스크와 다시 대조한다) (2026-09-20)

로트 `fc35eb2d..908a8a36`(7.5.2.1)을 gstack `/review` 로 다시 재리뷰했다 — 출처 아홉(Claude 전문가 다섯 · 레드팀 ·
적대 서브에이전트 · Codex 적대 · Codex 구조). 사람이 "1 — 7.5.2.2 로 같은 절차, 끝의 재확인은 `HEAD` 와 증거 둘 다,
판정한 sha 를 창 줄에, ast.json 구조 문제는 7.5.4 로 따로" 를 골랐다.

### 지난 수리의 확인

Codex(구조 · 적대) · 레드팀 · 보안 전문가가 각자 주입해 확인했다: `HEAD` 는 명령마다 한 번 풀린다 · 증거는 한 번
읽힌다 · git 결함 넷은 결함 · 모양 열 가지 · 목록에서 빠진 `ast.json` · 검색만 되는 디렉터리 · 심링크 고리 ·
서로게이트. 레드팀의 주장별 변이 15 가 전부 죽었다. **예외 둘**: 앞선 실행의 사실은 id 해소 실패 앞에서는 안
지워진다 · "조언도 그 한 번 읽기" 는 `main` 의 기록 조언이 다시 읽어서 거짓이다.

### 새 결함 — 내가 다시 재현한 것

| 결함 | 출처 | 재현(깨끗한 하위 프로세스) |
|---|---|---|
| **A.** `_bundle_text` 를 `iterdir` 로 다시 쓰며 `is_file()` 거름을 뺐다 — FIFO 에 영원히 멎고 `/dev/zero` 심링크면 `MemoryError`(GATE_FAULTS 밖 → traceback) | 레드팀 · 적대 · Codex 적대 · Codex 구조 | FIFO 하나 든 번들 → `_bundle_text` 8초 넘게 멎음 |
| **B.** 판정은 스냅숏으로만 서고 반환 전에 디스크와 대조하지 않는다 — Go 변경을 찾는 동안 증거를 무효로 고쳐도 `[]`, 디스크(커밋될 것)는 무효 · 실행 중 `HEAD` 가 움직여도 판정은 옛 `HEAD` 의 것이고 창 줄에 sha 가 없다 | Codex 적대 P1 · 적대(HEAD) | 워킹트리 경로: 스냅숏 판정 `[]`, 같은 디스크로 다시 돌리면 `missing AST branches ['B2']` |
| **C.** 앞선 실행의 사실이 id 해소 실패(`return [str(exc)]`)에서 안 지워진다 | 레드팀 · Codex | 오타 id 뒤 6 키 잔존 |
| **D.** 아주 깊은 JSON 은 `RecursionError` — 대상 이름 없이 판정 전체가 한 줄 | 적대 | `_parse_ast` 가 raise |
| **E.** 선언값 ≠ 계산값 갈래의 "진짜 불일치뿐" 은 너무 세다 — 같은 sha 라도 git 의 역사 해석(얕은 복제 경계 · 대체 ref · graft)이 실행 중 바뀌면 계산이 빈 값을 낸다 | Codex 적대 | (Codex 가 얕은 경계를 메모리로 바꿔 재현) |

### 입력 목록 — 무엇이 더 있었나

| 입력 | 7.5.2.1 | 7.5.2.2 |
|---|---|---|
| **끝의 디스크 상태**(판정한 증거 · `HEAD` 가 아직 거기 있나) | 수락 · 거절 앞에서만 대조, 판정 반환 전에는 안 봄 | `check` 의 판정 반환 **직전** 한 번 · `record_landing` 의 쓰기 직전 한 번 — 증거(`Evidence` 전체)와 `HEAD` 둘 다. 다르면 "다시 돌려라" |
| **번들 파일의 종류**(정규 파일인가) | `_bundle_text` 가 무엇이든 열었다 · `_read_evidence` · 대상 판정의 산문 읽기도 종류를 안 봤다(앞 로트부터) | 한 함수 `_read_regular`: `O_NONBLOCK` 으로 열고 `fstat` 이 정규 파일이 아니면 읽지 않는다 — 확인과 읽기 사이 틈이 없다 |
| **`execution_baseline.validate` 의 `HEAD`** | 목록에 없었다(열거기가 두 모듈만 걷고 별칭을 못 풀었다) | 열거기를 고쳐 보인다(409 · 411). 고정 **앞**에 읽힌다 — 끝의 `HEAD` 대조가 고정 뒤의 이동은 잡고, 고정 전에 갔다 돌아온 것(ABA)은 못 잡는다. 이관 창의 끝은 감사된 source 라 넓어지지 않는다(레드팀 확인) — **적는다** |
| 트리 전체의 `*_test.go` 읽기(`test_index`) | 종류를 안 본다 | **이 로트 밖 — 적는다**: 추적 파일만 있는 저장소 트리이고 번들이 아니다 |
| `base-commit.txt`(워킹트리에서 읽음) | — | **이 로트 밖 — 7.5.5 로 연다**(적대 INVESTIGATE, 앞 로트 전부터) |

### 편집 집합 (편집 전 AST 를 revision `908a8a36` 로 명시해 뽑는다)

| 함수 | 무엇 | 출처 |
|---|---|---|
| 새 `_read_regular` | 정규 파일만 읽는다(`O_NONBLOCK` + `fstat`) · 없으면 `FileNotFoundError` · 다른 종류는 `None` | A |
| `_read_evidence` · `_bundle_text` · `validate_target` | 번들 파일을 전부 `_read_regular` 로 — `_bundle_text` 는 정규 파일 아닌 것을 건너뛰고 못 읽는 정규 파일은 `OSError` 로 올린다(이름 댄 줄) · 대상 판정의 산문도 종류가 틀리면 "could not be read" | A · 적대(EACCES 조용히) |
| 새 `_judged_state_moved` · `check` · `record_landing` | 반환 · 쓰기 직전 `HEAD` 와 증거를 다시 본다 | B |
| `check` | 사실을 **맨 앞**에서 `facts.clear()` — 손으로 적은 목록 `RUN_FACTS` 를 없앤다 | C · 단순화 · maintainability |
| `main` | 창 줄에 판정한 `HEAD` 의 sha | B |
| `_parse_ast` | `RecursionError` 도 invalid | D |
| `resolve_landing` | 빈 계산값이면 `none — <사유>` 를 다시 말한다 · 주석을 참인 만큼으로 | E |
| `normalized_source` | 기준점도 `realpath`(판본과 무관 — 둘을 같은 도구로) | maintainability |
| `_self_repair_commits` · `_committed_many` | git 의 말 첫 줄은 `_first_line` 한 벌 · 낡은 주석(`_evidence_floor` 가 거절한다 → 결함이다) | maintainability |

### 거부할 정상 입력 — 그리고 거부 **안** 할 때 통과하는 것

| 새 거절 | 정상 입력에서 (실측) | 안 하면 통과하는 것 |
|---|---|---|
| 끝에서 `HEAD` 가 움직였으면 "다시 돌려라" | **실행 중 병행 세션이 커밋하면 걸린다** — 이 워크트리에서 실제로 일어난다(HANDOFF: 한 시간에 두 번). 값은 재실행 한 번. 정적 A/B 에서 0 이어야 한다 | 옛 `HEAD` 의 판정이 새 트리의 게이트 PASS 로 찍힌다(자기 수리가 끼면 거절될 상태) |
| 끝에서 증거가 달라졌으면 "다시 돌려라" | 실행 중 번들을 쓰면 걸린다 — 쓰는 사람에게는 맞는 말 | 무효 증거가 디스크에 있는 채 `[]` |
| 정규 파일 아닌 번들 항목 | **0 / 12,193**(저장소 전수) — 건너뛰기만 한다(거절 아님) | 멎음 · traceback |
| 못 읽는 정규 번들 파일 | **0 / 12,193** | 그 파일의 표가 열거형 호출 판정에서 조용히 빠진다 |
| `facts.clear()` | 호출자가 미리 넣은 키를 쓰는 곳 0 — `main` · 하네스 · 시험 전부 빈 사전 | 오타 id 뒤 앞선 실행의 창 |

**7.5.2.1 의 선택을 되돌린다**: "실행 중 커밋은 판정을 멈추지 않는다" 는 틀린 교환이었다 — 판정이 옛 역사의
것이어도 게이트의 나머지는 새 트리로 가고, 무엇을 판정했는지 아무 데도 안 남았다. 사람이 이 되돌림을 골랐다.

### 안 하는 것 — 사유

- **기록 명령의 하한 두 번 재기**(`_recording_refusal` + `_measure_landing_inputs`): 드문 수동 명령이고 `git log` 한
  번(ms)이다. 합치면 시험이 못 박은 거절 사유 순서(심링크 · 이미 있음 · 빌림 · 이관)가 바뀐다.
- **미리 읽기의 셋째 `_select_pinning`**: 착지 있는 change(오늘 a099 하나)에서 0.06s. 미리 읽기는 최적화이고 빠지면
  착지에서 읽는다 — `resolve_landing` 의 반환 모양을 바꿀 값이 아니다.
- **`validate_target(index=None)`**: 증거가 아니라 시험 함수 색인(캐시)이다 — 7.5.2.1 의 "시험 전용 `None` 기본값
  전부" 문장을 "증거 · 착지 입력의" 로 **좁힌다**(주장 정정).
- 확신 4 이하 항목(`_parse_ast` 사유 문자열 상수화 · `_empty_repo` 이름 · A/B 하네스 사이 중복): 부록.
- **7.5.3 · 7.5.4(ast.json 구조가 소스에서 다시 유도되지 않는다 — 보안 전문가 재현, 앞 로트 전부터) · 7.5.5**: 사람
  결정 · 별도 task.

### Pre-Edit 선언

High-risk 아님(게이트 도구, 생산 Go 변경 0). 정상 입력에서 판정 줄은 **한 글자도** 안 바뀌어야 한다 — 창 줄에
붙는 `HEAD` sha 만 예외이고, A/B 는 그 꼬리를 떼고 비교한 뒤 떼어 낸 sha 가 그때의 `HEAD` 인지 따로 본다.
RED 를 먼저 세운다.

## VERIFY — task 7.5.2.2 (판정은 내놓는 순간에도 거기 있는 것의 판정이다) (2026-09-20)

### 무엇이 바뀌었나 — 한 문장

**명령이 판정을 내놓기(기록을 쓰기) 직전에 판정한 `HEAD` 와 증거가 아직 그대로인지 다시 보고, 달라졌으면 판정
대신 "다시 돌려라" 를 낸다 · 창 줄 끝에 판정한 sha 를 적는다 · 번들 파일은 정규 파일만 읽는다.** 7.5.2.1 의
스냅숏은 판정을 **내적으로** 일관되게 했지만 그 판정이 **지금 거기 있는 것**의 판정인지는 안 봤다 — 둘 다 있어야
한다.

### 7.5.2.1 이 쓴 문장 중 거짓이었던 것 — 정정

- "착지 판정 · 대상 판정 · 조언이 그 두 값만 쓴다": 기록을 권하는 조언 줄은 다시 읽는다(다음 명령을 예측하는
  줄이라 둔다 — 판정 줄은 그 읽기에 안 기댄다). 주석 · docstring 을 "base 모양 조언" 으로 좁혔다.
- "실행 중 커밋은 판정을 멈추지 않는다": 틀린 교환이었다 — 되돌렸다(사람 결정).
- "선언값 ≠ 계산값 갈래는 진짜 불일치뿐, 빈 값은 못 온다": git 의 역사 해석이 바뀌면 온다(Codex) — 사유 문장을
  되살리고 주석을 "거의 언제나" 로.
- "시험 전용 `None` 기본값 전부 필수로": "증거 · 착지 입력의" 로 좁힌다(`validate_target(index=None)` 은 캐시).
- 편집 집합 표의 `_ast_value` 는 구현에서 `_parse_ast` 로 이름을 바꿨다(옛 디스크 읽기 함수와 이름이 겹쳐서) — 그
  표에 안 적었다.
- a099 행의 "git 82 → 83" 은 7.5.1 기준이었다 — 7.5.2 기준으로는 84 → 83 이다(performance 전문가).
- `_bundle_text` 를 `iterdir` 로 다시 쓰며 `is_file()` 거름을 빠뜨렸다 — FIFO 에 멎는 **회귀를 내가 만들었다**.
  시험 254 · 변이 76 이 못 봤다(번들에 정규 파일 아닌 것을 두는 픽스처가 0 이었다).

### RED 가 빨갰던 이유 (구현 전, `908a8a36` 위에서)

새 시험 15 중 **12 가 실패** — 각자 자기 사유로(출력: 세션 스크래치 `7522_red_on_908a8a36.txt`): FIFO 셋(번들 파일 ·
필수 산문 자리 · `ast.json` 자리) → 하위 프로세스가 30초 넘게 멎음 · `/dev/zero` 심링크 → `MemoryError`(주소 공간 1GiB
상한 아래) · Go 변경을 찾는 동안 증거를 무효로 → `[]` · 실행 중 커밋 → 판정 `[]` / 기록 rc 0 · 창 줄에 sha 없음 · 못 읽는
번들 파일 → 조용히 `[]` · 오타 id 뒤 사실 잔존 · 깊은 JSON → `RecursionError` · 빈 계산값 → `computes ()`. 나머지 3 은
옛 코드에서도 초록인 못(글자 아닌 `ast.json` · 명령 수준의 태어나지 않은 `HEAD` · 상징 `HEAD` ref 문자열이 푸는 자리 밖에
없다는 구조 시험 — 빌림 · 이관 갈래의 되돌림 변이를 잡으려고).

### 편집 순서 — 지켰다

편집 전 AST 11 을 **편집 전에** revision `908a8a36` 로 뽑았다. 새 함수 셋(`_read_regular` · `_verdict` ·
`_judged_state_moved`). `check` 를 `check` + `_verdict` 로 나눈 것은 **기계로** 대조했다: 옛 `check` 의 분기 소스 순서열 =
새 `check` 의 것 + `_verdict` 의 것, 차이는 셋뿐이고 전부 의도다(`RUN_FACTS` 순회 삭제 · 출구의 재확인 갈래 · 못 읽은 파일
이름 갈래). 반환은 13 = 9(창까지) + 4(`_verdict`), 새 출구 하나를 더해 10 + 4.

### 내가 내 로트에서 찾은 회귀 — 번들 안의 폴더

BTM 을 쓰다 "정규 파일이 아닌 항목" 의 종류를 **세어 보려고** 소켓 · 끊긴 링크 · 디렉터리를 실제로 넣어 봤다.
디렉터리가 `TypeError` 로 게이트를 죽였다 — 판정 줄 없이 traceback:

```text
IsADirectoryError: [Errno 21] Is a directory: 3          ← `open(fd)` 가 디렉터리에서 스스로 터진다
TypeError: argument should be a str or an os.PathLike …  ← 그 예외의 `filename` 은 경로가 아니라 **정수 fd**
```

`_read_regular` 이 종류 검사를 **파일 객체로 감싼 뒤에** 했기 때문이다(`with open(descriptor…)` 가 먼저 왔다).
`GATE_FAULTS` 밖이라 `main` 도 못 잡는다. 수리는 셋이다: 종류 검사를 감싸기 **앞으로**(`fstat` 먼저) ·
디렉터리는 **건너뛰지 않고** 경로를 담은 `IsADirectoryError`(건너뛰면 번들 안에 폴더를 만들어 그 안에 열거형 호출
표를 넣는 것으로 감사가 꺼진다 — `_bundle_text` 가 막으려는 바로 그것) · 이름 대는 자리는 `filename` 이 글자가
아니면 번들 이름으로 떨어진다. 그리고 어느 갈래로 나가도 서술자를 닫는다(`finally`) — 안 닫으면 번들 3,269 개를
도는 한 번의 게이트가 열린 파일 상한에 걸린다.

시험 넷(폴더 · 소켓 · 글자 아닌 이름 · 서술자 수)을 더했고, 되돌린 사본에서 **둘이 빨갛다**(폴더 · 이름;
소켓과 서술자 수는 새 못이다). 변이 Z16 · Z17 · Z18 · Z19 가 같은 것을 네 방향에서 잰다.

**이것이 A/B 와 변이를 두 번 돌린 까닭이다** — 첫 판은 수리 전 소스(`6060c2a2`) 위였고, 소스가 바뀌었으니
증거도 다시 만들었다(앵커 셋은 철자가 바뀌어 다시 걸었다).

### 계측기가 눈멀 뻔한 것 — 셋

증거 읽기가 `_read_regular`(`os.open`)로 옮기자 `Path.read_bytes` 만 걸던 계측기가 전부 눈멀었다:
시험의 `_reads`(주입 · 읽기 수 · 불변식 — 불변식 시험은 빈 표본 위에서 참이 될 뻔했다), `7521_main_ab.py` 의 읽기 계수
(첫 표본 `48 → 0`), `752_reads.py`. 셋 다 `_read_regular` 도 걸게 고쳤고, 표본으로 계수가 말이 되는지 보았다.

### 증거

| 무엇 | 결과 |
|---|---|
| `main()` 출력 전체 A/B (`7521_main_ab.py`, 기준 `908a8a36`, 전수, 순서 번갈아, 창 줄의 `judged at HEAD` 꼬리는 떼고 비교 · 떼어 낸 sha 는 그때의 `HEAD` 와 따로 대조) | **126 / 126 SAME · DIFFERENT 0** |
| 결정적 계수 (같은 A/B) | git 20,133 → **20,249**(+116 = 끝에서 `HEAD` 를 다시 푸는 한 번 × 거기까지 가는 change) · `ast.json` 읽기 5,803 → **9,072**(+3,269 = 저장소의 번들 총수 한 벌 — 끝의 재확인이 증거를 한 번 더 읽는다). **7.5.2.1 전 값(9,072)과 같은 수인데 의미가 다르다**: 판정은 여전히 한 번 읽은 바이트로만 서고, 둘째 읽기는 판정에 안 들어가고 대조만 한다. 시간은 순서 편향 안이다(먼저 돈 쪽이 ~5% 느리다: before 먼저 1730.4s → 1589.7s · after 먼저 1601.3s → 1679.2s). 이 판의 절대 시간은 Go 빌드 캐시를 비운 직후라 앞 판(1087.7s)보다 느리다 — 시간을 결론에 쓰지 않는 까닭이다 |
| 변이 (`75_mut.py`, 스위트 전체) | **95 / 95 CAUGHT · 생존 0**. 94 는 한 판(스위트 272 · 무변이 대조군 GREEN). Z14 는 **도달한 채 살아남았다** — `io.BytesIO(None)` 이 조용히 빈 버퍼라 "정규 파일이 아니면 건너뛴다" 를 지워도 이어 붙인 글자가 같다(행동으로는 안 보인다). 구조 시험(`_decoded` **앞**의 `raw is None` 갈래)을 더해 다시 재니 CAUGHT — 지키던 것이 우연이었다 ([[surviving-mutant-may-mean-accidental-safety]]) |
| a099 `check()` 한 번 (`752_reads.py 908a8a36`) | `ast.json` 읽기 76 → **114**(번들당 2.0 → 3.0 — 끝의 대조 하나) · git 83 → **84**(끝의 `HEAD`) — 이 로트가 더한 비용이고 적는다 |
| 시험 | 254 → **273**(+21 새 · −2 뒤집힘, 건너뜀 1 은 앞서부터 있던 외부 `SDD_PYTHON` 환경) · `make lint` · `make sdd-test`(348) · `make test-seams` · `openspec validate --all --strict` 58/58 · a122 게이트 rc=0 |
| 하네스 | `7521_inputs.py` 가 `execution_baseline.py` 와 import 별칭을 걷는다(이관 경로의 `HEAD` 읽기 409 · 411 이 보인다) · `7521_main_ab.py` 이어 달리기 기록을 저장소 입력 상태에도 묶었다(Codex 구조 P2) · 읽기 계수기 둘 |
| 문서 | README 에 "판정은 역사 하나 · 읽기 하나, 끝에서 다시 보고 달라졌으면 다시 돌려라, 창 줄의 sha, 정규 파일만 읽고 **폴더 · 소켓은 이름 댄 줄**" · 하네스 README 에 "앵커는 소스를 따라 낡는다 · 소스를 고치면 A/B 와 변이를 다시 돌린다" |

### 남은 것

- **7.5.3 · 7.5.4(ast.json 구조가 소스에서 유도되지 않는다) · 7.5.5(`base-commit.txt` 를 워킹트리에서)**: 사람 결정 · 별도.
- **이관 경로의 `HEAD`**: `validate` 가 고정 **앞**에서 스스로 읽는다 — 끝의 대조가 고정 뒤의 이동은 잡고, 고정 전에
  갔다 돌아온 것(ABA)은 못 잡는다. 이관 창의 끝은 감사된 source 라 넓어지지 않는다(레드팀 확인). 적는다.
- **번들 산문은 두 번 읽힌다**(`_bundle_text` · 대상 판정): 착지가 판정하지 않는 범위이고, 끝의 대조는 `ast.json` 만 본다.
  실행 중 산문을 고치면 두 읽기가 갈릴 수 있다 — 적대가 INVESTIGATE 로 남겼다. 적는다.
- **안 한 것**(Pre-Edit 의 사유): 기록 명령의 하한 두 번 재기 · 미리 읽기의 셋째 `_select_pinning` · 확신 4 이하 항목.
- **측정 환경**: 첫 재측정은 루트 파일시스템이 100% 차서(`~/.cache/go-build` 203GB — 게이트가 change 마다 `go run`
  으로 Go 함수를 뽑으므로 A/B 한 판이 캐시를 크게 키운다) `No space left on device` 로 오염됐다 — A/B 는
  DIFFERENT 36 이 전부 그 문장이었고, 변이는 **디스크 때문에 빨개진 것을 CAUGHT 로 적을 수 있어** 둘 다 버리고
  캐시를 비운 뒤 처음부터 다시 돌렸다. 위 표는 그 재측정이다. (`TMPDIR` 을 `/mnt/D` 로 돌려 본 것도 버렸다 —
  그 마운트는 fuseblk 라 권한 · FIFO 픽스처가 깨져 무변이 대조군이 22개 실패로 빨갛다.)

## MEASURE · Pre-Edit Gate — task 7.5.2.3 (재확인의 입력 집합은 판정의 입력 집합이어야 한다) (2026-09-20)

로트 `1d12520c`(7.5.2.2)를 다시 재리뷰했다 — 출처 여섯(Codex 구조 · Codex 적대 · 레드팀(변이 24) · 정확성/동시성 ·
시험 품질(변이 28) · 보안). 사람이 **"2 — 1번 전부, 감사 스위치의 비교 규칙만 사람 결정으로 남긴다"** 를 골랐다.
그 규칙(표를 무엇으로 세는가)은 계약이고 바꾸면 이 change 밖의 번들까지 판정이 달라진다 → task 7.5.6.

### 지난 수리의 확인

Codex 구조 리뷰는 발견 0(스위트 273 확인). 레드팀·시험 품질 리뷰가 주장별 변이 52 를 걸어 **47 이 죽었다** —
끝의 재확인 자체(`HEAD` · 증거) · `_read_regual` 의 종류 검사 · 폴더/소켓의 이름 댄 줄 · `facts.clear()` 의 자리 ·
`RecursionError` · `_first_line` 은 확인됐다. 살아남은 다섯이 아래 결함 4 다.

### 새 결함 — 출처와 재현

| 결함 | 출처 | 재현 |
|---|---|---|
| **A.** 끝 재확인의 입력 집합이 판정 입력 집합의 **진부분집합**이다 — 재확인은 `HEAD` + `Evidence` 만 다시 읽는데 판정은 번들 **산문** · Go **워킹트리** 소스 · `review.md` · `base-commit.txt` · 트리 전체 `*_test.go` 를 읽는다 | 정확성/동시성 | 아래 입력 목록(측정) · 산문 케이스 나 재현: 판정 중 `risk-pattern-report.md` 에 `TODO` 를 넣으면 `[]`, 같은 디스크 재실행은 `still contains TODO` |
| **B.** `record_landing` 의 쓰기 직전 재확인이 **더러운 트리 거절을 다시 안 묻는다** — 133~219초 순회 도중 추적 Go 편집이 들어오면 `landed-commit.txt` 가 `open("xb")` 로 **영구** 기록되고 그 뒤 게이트는 지도 없는 새 분기에 `[]` | 보안 · 적대 | 보수 불가(사람 손) |
| **C.** `HEAD` 비교가 증거 스캔보다 **앞**이라 스캔 도중 커밋이 통과한다 | Codex 적대 | 나 재현 |
| **D.** 내 주장 다섯이 한 줄 변이에 273 초록 — 창 줄 sha(`head = _head_commit(root)`) · 재확인의 `present`/`targets` 절반(`.held != .held`) · `_bundle_text` 배관(`_bundle_text(target, None)` → 감사 OFF) · "명령마다 `HEAD` 한 번"(`@` 철자) · "서술자는 언제나 닫힌다"(정규 갈래만 누수) | 레드팀 · 시험 품질(독립) | 변이 |
| **E.** 조용한 건너뛰기가 타이밍 공격의 문 — 커밋된 표 파일을 게이트가 **여는 그 순간에만** FIFO 로 바꿨다 되돌리면(또는 이름을 밖으로 옮겼다 되돌리면) 열거형 호출 감사가 꺼지고 `[]` | 보안 | 실측 6/14 · 3/10 (디스크에는 표가 든 정규 파일이 그대로) |
| **F.** FIFO 가 `review.md` · `function-logic-reference.txt` · `base-commit.txt` · 시험 파일 색인에서 **영원히 멎는다** · 큰 **정규** 파일은 `MemoryError`(GATE_FAULTS 밖) · 비UTF-8 필수 산문이 대상 이름을 지운다 | 보안 · 적대 | 20s+ · 두 자리(`validate_target` · `_base_shaped_bundles`) |

### 입력 목록 — 판정이 읽는 것 전수 (`analysis/harness/7523_inputs.py`, 손으로 세지 않았다)

**정적**(모듈 AST 열거): 내용을 읽는 자리 **19 · 함수 14** · 이름만 재는 자리 **19 · 함수 10**.
읽는 함수: `_read_regular` · `_read_evidence` · `_bundle_text` · `validate_target` · `_base_shaped_bundles` ·
`check` · `resolve_base` · `resolve_referenced_change` · `test_index` · `test_spans` · `resolve_test_file` ·
`test_citation_errors` · `record_landing` · `_decoded`(BytesIO).

**동적**(`check('a112-…')` 한 판, 번들 148):

| | 수 |
|---|---|
| 판정이 내용을 읽은 **서로 다른 경로** | **1,739** (`*_test.go` 962 · 번들 736 · Go 소스 37 · `review.md` 1 · `base-commit.txt` 1 · 그 밖 2) |
| 이름만 잰 경로 | 326 |
| 트리 전체 패턴 순회 | 1 (`rglob('*_test.go')`) |
| **끝의 재확인이 다시 읽는 것** | **149** (증거 디렉터리 1 + `ast.json` 148) |
| **재확인 밖** | **1,591 = 91.5%** (`*_test.go` 962 · 번들 산문·목록 589 · Go 소스 37 · `base-commit.txt` · `review.md` · 아카이브 목록) |

재확인이 안 보는 것의 표본: `cmd/tossctl/a053_httpapi_reader_test.go` · `internal/app/engine/protection_wiring.go` ·
`…/analysis/function-logic/internal-app-engine--allsixstrategyfirstlegresults`(번들 목록) · `…/base-commit.txt` ·
`…/review.md` · `openspec/changes/archive`.
면제 경로는 `present=False` 라 바이트 표본이 **0** 이다 — 전칭 대조가 공허하게 참
([[universal-check-passes-on-an-empty-sample]]).

**크기와 비용**(상한의 근거): 가장 큰 `*.go` **97,231 B**(`internal/journal/fills.go`) · 가장 큰 `ast.json`
**25,466 B** · 번들 파일 최대 ~26 KB · `*_test.go` 962 개 합 **10.0 MB**. 다시 읽는 비용 실측:
`rglob('*_test.go')` **0.586s** + 962 파일 읽기 **0.062s**. 한 판 전체: a122 0.27s(번들 0) · a112 **6.0s**(148) ·
a066 21.8s(24, 순회 있음). 그래서 재확인이 1,739 경로를 다시 읽는 비용은 a112 에서 **약 0.7s(+12%)** 로 예상하고
A/B 에서 실측한다.

### 설계 — 왜 목록을 손으로 들지 않는가

7.5.2.1 은 손으로 적은 키 목록(`RUN_FACTS`)이 낡아서 깨졌고, 7.5.2.2 는 손으로 고른 재확인 집합
(`HEAD` + `Evidence`)이 판정 집합보다 좁아서 깨졌다. **같은 실패가 두 번**이다. 그래서 이번 수리는 목록을
만들지 않는다: 디스크를 읽는 **깔때기 네 개**(파일 · 디렉터리 목록 · 패턴 순회 · 종류)가 읽은 것의 **결과**를
원장(`ReadLedger`)에 적고, 끝의 재확인은 그 원장을 그대로 다시 읽어 대조한다. 재확인의 입력 집합은 **구조적으로**
판정의 입력 집합이다. 새 읽기 자리는 깔때기를 쓰는 순간 저절로 참여하고, 깔때기를 비켜 가는 것은 AST 구조 시험이
막는다(원시 호출은 깔때기 안에만 있을 것).

원장은 **실패도 적는다**(없음 · 정규 파일 아님 · 권한 · 너무 큼). 그래서 E 의 두 공격이 같은 기전으로 닫힌다:
여는 순간에만 FIFO 였다가 되돌아온 파일은 원장에 "정규 파일 아님" 으로 적혀 있고 재확인에서 바이트가 읽히므로
**다르다** — 이름을 빼고 되돌린 경우도 원장의 "없음" 과 다르다. 같은 경로를 한 판에서 **두 번** 읽었고 결과가
갈리면 그 자리에서 움직임이다(원장이 본다).

### 편집 집합 (편집 전 AST 는 revision `1d12520c` 로 뽑는다)

| 함수 | 무엇 | 결함 |
|---|---|---|
| 새 `ReadLedger` · `_ledger` · `_remember` · `_reads_moved` | 읽은 것의 원장 · 명령 하나의 수명 · 끝의 재대조(재확인 중에는 적지 않는다) | A |
| 새 `_file_outcome` · `_listing_outcome` · `_pattern_outcome` · `_kind_outcome` | 지문을 만드는 **한 자리**씩 — 판정과 재확인이 같은 함수를 부른다([[two-judgements-cover-for-each-other]]) | A |
| `_read_regular` | 깔때기: 결과를 원장에 적는다 · 정규 파일이 아니면 **이름 댄 예외**(조용한 `None` 폐지) · 크기 상한(16 MiB)으로 `MemoryError` 대신 판정 줄 | A · E · F |
| 새 `_listed` · `_globbed` · `_is_regular` | 디렉터리 목록(이름+종류) · 패턴 순회 · 종류 검사도 같은 원장으로 | A |
| `_read_evidence` · `_bundle_text` · `_verdict` | 목록을 `_listed` 로 · 번들 목록을 한 번 읽어 `_bundle_text` 에 넘긴다(`lexists` 폐지) · 정규 파일 아닌 번들 파일은 **이름 댄 판정 줄** | A · E |
| `validate_target` · `_base_shaped_bundles` | 워킹트리 소스도 `_read_regular`(`is_file`+`read_bytes` 두 syscall → 하나) · 비UTF-8 필수 산문은 **대상 이름 댄 줄** | A · F |
| `check` · `resolve_base` · `resolve_referenced_change` | 원장을 열고 닫는다 · `review.md` · `function-logic-reference.txt` · `base-commit.txt` · 아카이브 목록도 깔때기로(FIFO 멎음 종결) | A · F |
| `test_index` · `test_spans` · `resolve_test_file` · `test_citation_errors` | 시험 색인의 순회 · 읽기 · 인용 해소도 깔때기로 | A · F |
| `_judged_state_moved` | `HEAD` → 원장 전체 → `HEAD` (앞뒤로 감싼다 — C) | A · C |
| `record_landing` | 쓰기 직전에 거절 집합 **전체**를 다시 묻는다(더러운 트리 포함) | B |
| 시험 | D 의 다섯 주장을 못 박는다: 창 줄 sha(`_head_commit` 연속 반환이 갈리는 픽스처) · `present`/`targets` 절반 · `_bundle_text` 배관 · `HEAD` 호출 수(구조) · 서술자 누수(`/proc/self/fd`, 없으면 skip) | D |

### 거부할 정상 입력 — 그리고 거부 **안** 할 때 통과하는 것

| 새 거절 | 정상 입력에서 | 안 하면 통과하는 것 |
|---|---|---|
| 판정 중 만진 **어느** 파일이든 달라졌으면 "다시 돌려라" | 병행 세션이 이 워크트리에서 커밋·편집하면 걸린다(HANDOFF: 한 시간에 두 번). 값은 재실행 한 번. 정적 A/B 에서 0 이어야 한다 | 판정이 지금 체크아웃된 상태를 기술하지 않는다 — `[]` 로 게이트 PASS |
| 정규 파일 아닌 번들 파일 = 이름 댄 판정 줄(건너뛰기 폐지) | **0 / 12,193**(저장소 전수, 7.5.2.2 에서 셌다) | 여는 순간 교체로 감사 OFF (실측 6/14) |
| 16 MiB 넘는 정규 파일 = 이름 댄 판정 줄 | 가장 큰 `*.go` 97 KB · 가장 큰 번들 파일 26 KB — **170배 여유**, 오늘 걸리는 파일 0 | `MemoryError` 로 판정 줄 없이 죽음 |
| 비UTF-8 필수 산문 = 대상 이름 댄 줄 | 저장소 전수 0 | 판정 전체가 대상 이름 없는 한 줄 |
| 기록 쓰기 직전의 더러운 트리 = 안 쓴다 | 저자가 순회(133~219s) 도중 편집하면 걸린다 — 다시 부르면 된다 | 지도 없는 분기를 영구히 창 밖으로 (사람 손 보수) |

### 안 하는 것 — 사유

- **감사 스위치의 비교 규칙**(BOM · `\xff` · UTF-16 · 좌표 없는 행): **사람이 남기라고 했다** → task 7.5.6.
  이 로트는 문장만 좁힌다.
- **7.5.3 · 7.5.4 · 7.5.5**: 사람 결정 · 별도 task(앞 로트 전부터).
- **재확인 뒤의 창**: 원장을 다시 읽은 **그 순간 뒤**의 편집은 어느 설계로도 못 본다. rc 를 소비하는
  `tools/gate.sh:321` 이 즉시 읽으므로 창은 거기까지다 — **적는다**(없앴다고 쓰지 않는다).
- **CI**: 이 게이트는 CI 에서 돌지 않는다(`ci.yml:94`). 타이밍 공격을 뒤에서 잡아 줄 것이 없다 — 적는다.

### Pre-Edit 선언

High-risk 아님(게이트 도구, 생산 Go 변경 0). 정상 입력에서 판정 줄은 **한 글자도** 안 바뀌어야 한다 — A/B 는
`main()` 출력 전체를 번갈아 비교한다. 비용은 늘어난다(재확인이 1,739 경로를 다시 읽는다) — 실측해서 적는다.
RED 를 먼저 세운다.

## VERIFY — task 7.5.2.3 (재확인의 입력 집합은 판정의 입력 집합이다) (2026-09-20)

**근본 수리 한 줄: 판정이 디스크를 읽는 자리를 깔때기 넷으로 모으고 그 깔때기가 읽은 **결과**(실패 포함)를
원장에 적게 해서, 끝의 재확인이 다시 읽는 집합이 손으로 고른 목록이 아니라 **판정이 읽은 그 집합**이 되게 했다.**

7.5.2.1 은 손으로 적은 키 목록이 낡아서, 7.5.2.2 는 손으로 고른 재확인 집합(`HEAD` + `Evidence`)이 좁아서
깨졌다 — 같은 실패가 두 번이다. 이제 목록이 없다.

곁들여: 쓰기 직전에 **거절 집합 전체**를 다시 묻는다(더러운 워킹트리는 역사에도 원장에도 없다) · `HEAD` 를
재확인의 **앞뒤로** 묻는다 · 조용한 건너뛰기를 없앴다(정규 파일 아님 · 목록 실패 · 16 MiB 초과 · UTF-8 아님은
전부 **이름 댄 판정 줄**) · `review.md` · `function-logic-reference.txt` · `base-commit.txt` · 시험 색인의 FIFO
멎음을 닫았다 · 못 여는 증거 디렉터리가 면제로 통과하던 permissive 구멍을 닫았다.

### 증거

| 잰 것 | 결과 |
|---|---|
| `main()` 출력 **전체** A/B (전수 126, 순서 번갈아) | **SAME 126 · DIFFERENT 0** — 판정 줄이 한 글자도 안 바뀌었다 |
| 변이 (`75_mut.py`, 전수 113 = 기존 95 중 재겨눔 10 + 새 18) | **113/113 CAUGHT · 생존 0** — 112 는 한 판에서(스위트 297), AA19 는 **살아남아** 목록의 종류를 재는 시험을 더한 뒤 다시 재서 CAUGHT(298) |
| 시험 | 273 → **298** (+26 · 옛 FIFO 건너뛰기 시험 1 은 새 계약의 시험이 대신한다) |
| `lint` · `test-seams` · `sdd-test` · `openspec validate` · a122 게이트 | 아래 "게이트" 표 |

**비용은 적는다.** 재확인이 판정이 읽은 것을 전부 다시 읽으므로 파일 열기가 늘어난다(`7523_reads.py`):

| change | 판정 줄 | 파일 열기 | 프로세스 | 시간 |
|---|---|---|---|---|
| a112 (번들 148) | 36 → 36 | 2,407 → **3,854** (+60%) | 116 → 117 | 5.94s → **6.69s** (+12.6%) |
| a066 (번들 24, 순회 있음) | 388 → 388 | 1,203 → **2,246** (+87%) | 615 → 616 | 20.83s → **24.00s** (+15.2%) |
| a122 (번들 0) | 0 → 0 | 2 → 6 | 6 → 7 | 0.12s → 0.12s |

A/B 전수의 git 프로세스는 20,249 → **20,365**(+116 = change 하나당 `HEAD` 한 번 더 — 재확인의 뒤쪽 물음).
`ast.json` 읽기는 9,072 → **9,072** 로 **같다**: 7.5.2.2 도 끝에서 증거를 통째로 다시 읽었기 때문이다. 늘어난
것은 그 밖의 전부다(번들 산문 · 워킹트리 소스 · `review.md` · `base-commit.txt` · 시험 색인).

### 거부하는 정상 입력 — 전수로 셌다

| 새 거절 | 오늘 걸리는 것 |
|---|---|
| 정규 파일 아닌 번들 파일 | **0 / 12,193** |
| 16 MiB 넘는 번들 파일 | **0 / 12,193** (가장 큰 `*.go` 97,231 B · 가장 큰 번들 파일 25,466 B — 170배 여유) |
| UTF-8 아닌 번들 파일 | **0 / 12,193** |
| 목록 못 여는 번들 · 못 여는 증거 디렉터리 | 0 |
| 실행 중 입력이 달라짐 | 병행 세션이 이 워크트리를 같이 쓰면 걸린다. 값은 재실행 한 번이고, 정적 A/B 126 판에서 **0** 이었다 |

### 앞 로트의 문장 정정

- **7.5.2.2 의 "FIFO · 장치는 건너뛴다"** 는 이제 거짓이다 — 조용한 건너뛰기가 타이밍 공격의 문이었다(보안
  리뷰 실측 6/14). README 문단을 그에 맞게 다시 썼다.
- **심링크 고리는 "없다" 가 아니라 "못 읽는다"** 다(`AST source could not be read`). 링크는 거기 있다 —
  없는 것과 못 읽는 것을 한 말로 하면 저자가 있는 파일을 다시 만들러 간다.
- **"명령마다 `HEAD` 한 번"** 은 이제 "판정 경로에 한 자리, 재확인이 앞뒤로 두 번" 이다.

### 하네스도 고쳤다 (계측기부터 못 박는다)

로트 도중 **배경 판과 전경 창이 한 사본을 공유**해서 서로의 변이를 기준으로 삼았다 — 무변이 대조군이 빨개져서야
알았고, 그 사이의 창 결과는 버렸다. 세 가지를 고쳤다: 사본을 **프로세스별**로 가른다 · 사본이 **원본과 같은지**
단언하고 변이마다 **원복을 세어서** 확인한다 · 도달 계측기가 **변이가 앉을 자리**(옛 줄)에 표식을 심는다(새 줄의
머리를 찾던 판본은 한 줄 통째 교체에서 언제나 "안 닿음" 이라고 답했다 — 눈먼 계측기였다).
`7521_main_ab.py` 의 읽기 계수기도 깔때기 안(`_opened_bytes`)으로 내렸다.

### 안 한 것 — 사유

- **감사 스위치의 비교 규칙**(BOM · UTF-16 · 좌표 없는 행): 사람이 남기라고 했다 → **task 7.5.6**. 이 로트는
  못 푸는 바이트를 **거절**할 뿐(건너뛰기 폐지), 표를 **무엇으로 세는가**는 손대지 않았다.
- **7.5.3 · 7.5.4 · 7.5.5**: 사람 결정 · 별도 task.
- **재확인 뒤의 창**: 원장을 다시 읽은 그 순간 뒤의 편집은 어느 설계로도 못 본다. rc 를 읽는
  `tools/gate.sh:321` 까지가 그 창이고, **CI 는 이 게이트를 돌지 않는다**(`ci.yml:94`).
- `test_citation_errors` 의 "고른 뒤 사라짐" 갈래는 **경합 전용**이라 주입 없이는 못 닿는다 — BTM 에 적었다.

---

## 정정 — task 7.5.2.3 의 VERIFY 에서 거짓이던 문장 (2026-09-22, 독립 재리뷰 다섯)

7.5.2.3 을 커밋(`845471a1`)한 뒤 독립 재리뷰 다섯(적대 · 문서/주장정확성 · 정확성/동시성 · 보안 · 시험품질)을
돌렸다. **아래 셋은 내가 적은 문장이 거짓이었다.** 고친 것은 7.5.2.4 가 하고, 여기엔 무엇이 거짓이었는지를 남긴다.

1. **"디스크를 읽는 자리가 깔때기 넷뿐 · 재확인의 집합은 구조적으로 판정의 집합"** — 거짓이다.
   판정은 자식 프로세스로도 읽는다: `changed_existing_functions` 의 `git diff`(:132) · `base_file` 의
   `git show`(:85) · `go_functions` 의 `go run`(:56, **워킹트리 Go 바이트**) · `execution_baseline.validate`.
   판정 도중 추적 Go 파일을 고치면 생산 CLI 가 **rc 0 `evidence complete`** 를 내고 같은 디스크 재실행은 rc 1 이다
   (내가 재현했고 정확성 리뷰어가 별도 합성 저장소로 독립 재현). 노출: a112 는 Go **138개**가 `required` 를 정하는데
   원장의 Go 는 고정 소스 **37개**뿐이고, 고정 번들이 0 인 change 는 원장의 Go 표본이 **0** 이다.
   **내가 못 본 까닭**: 입력 목록 하네스(`7523_inputs.py`)도 구조 시험(`test_check_analysis.py`)도
   `check_analysis.py` 의 **모듈 AST만** 걸었다 — 자식 프로세스 읽기는 애초에 모집단에 없었다.
   "전수" 는 사실의 진술이 아니라 **방법의 한계**였다. → **task 7.5.8**.
2. **"도달 계측기가 변이가 앉을 자리에 표식을 심는다"** — 거짓이었다. `75_mut.py` 는 `reached()` 를 **원복 전**에
   불러 *변이된* 본문에서 *옛 줄* 을 찾았다. 줄을 통째로 바꾼 변이는 그 줄이 이미 없으니 언제나 "안 닿음" 이다
   (리뷰어 정적 재연: **79/113**). 생존이 0 이라 기록된 판정은 안 바뀌지만 문장은 거짓이었다. → 7.5.2.4 에서 고쳤다.
3. **"변이 113/113"** — 그중 **Y13 · Y14 는 행동 변이가 아니라 `NameError`** 였다(`_verdict` 범위에 `analysis`
   라는 이름이 없다 — 내가 AST 로 확인했다). 정직한 값은 **111 CAUGHT + 2 무효**였다. → 7.5.2.4 에서 오늘의
   범위로 다시 써서 둘 다 CAUGHT 로 만들었고, **Y13 은 다시 쓰자 실제로 살아남아** 시험을 하나 더 낳았다.

곁들여 정정한 숫자·문장(전부 7.5.2.4 에서 고쳤다): 번들 최대 크기의 **모집단이 한 문장에 둘**이었다(25,466 B 는
활성 번들만의 최대) · 벽시계 초(5.94→6.69s 등)는 `7523_reads.py` 가 before 를 늘 먼저 한 프로세스에서 단일
표본으로 돌아 **재현되지 않는다**(두 리뷰어가 따로 재서 "after 가 더 빠름") · "실측 6/14 · 3/10" 에 커밋된
하네스가 없다 · "그 기록은 지울 수 없고" 가 같은 README 의 "지우는 커밋" 과 모순 · "건너뛰는 모양은 사라진 파일
하나뿐" 인데 같은 경로에 셋 더 · 낡은 docstring 셋(`_head_commit` · `_bundle_text` · `_verdict` 의 `lexists`).

**철회된 지적 하나**: 정확성 리뷰어의 "`_raise_if_inputs_moved` 가 다른 가드를 가린다" 는 사실이 아니다 —
시험품질 리뷰어가 양방향으로 반증했다(`_reads_moved` 만 죽이면 원장 겨냥 시험이 정확히 빨개지고,
`_raise_if_inputs_moved` 를 통째로 죽이면 그 클래스 26개가 **전부 초록**이라 아무것도 그것에 안 기댄다).
(7.5.22 정정: 여기 적었던 "**11개**" 는 틀렸다. 다시 재니 **16개**다 — 시험 313 기준, `_reads_moved` 의
첫 줄에 `return ""` 를 넣고 전수 실행. 반증의 방향은 그대로 서고 수만 틀렸다.)

**보안 리뷰어의 한 대목도 내가 재서 정정한다.** "`revision: base` 는 소스 해시를 **아예 안 본다**" 는
`validate_target` 안에서만 참이다. `_verdict` 의 `base_hash` 대조가 `ast.json` 의 `source_sha256` 을 base 소스에서 뽑은
`base_hash` 와 대조하므로, **`required` 에 든 대상은** 묶여 있다. 안 묶이는 것은 `required` 밖의
`revision: base` 번들이고 오늘 그것으로 만들 수 있는 잘못된 판정은 없다. 실제 손해는 해시가 아니라
**무엇을 요구하느냐**다 — 현재 쪽이 안 돌면 저자는 *편집 뒤* 가 아니라 *편집 전* 논리의 지도를 내고 통과한다.

## MEASURE · Pre-Edit Gate — task 7.5.2.4 (통합 diff 의 본문 줄은 파일 헤더가 아니다) (2026-09-22)

**무엇을 재기로 했나.** 사람이 선택지 1 을 골랐다 — P0-B(diff 파서) + 문서·VERIFY 정정 + 하네스 계측기 수리를
한 로트로. 근거는 둘이다: 기록이 거짓이면 그 위에 쌓는 판단이 오염되고, 계측기가 눈먼 채로는 다음 로트의
변이 증거도 못 믿는다.

**FLM 먼저.** 편집할 함수는 `changed_existing_functions` 하나다. `enumerate.py` 로 편집 **전** AST 를 뽑았다 —
`analysis/python-function-logic/tools-logic-map--changed_existing_functions/ast.before-7.5.2.4.json`
(`:124-235` · 분기 **34** · 반환 2 · raise 3 · 호출 47). 파서는 `B26~B34` 이고 결함은 `B28`·`B30` 이다.

**결함은 진짜 git 으로 재현했다** (`analysis/harness/7524_diff.py`, 커밋했다). 네 벌의 **진짜 저장소**를 만들고
진짜 `git diff --unified=0` 을 생산 파서에 흘린다(Go 추출기만 세운다 — 고정 저장소에 `./tools/logic-map` 이 없다).
편집 전:

| 벌 | git 이 낸 헤더 모양 줄 | required | 해시 칸 |
|---|---|---|---|
| `control` | 없음 | `x.go:F` | `current_hash` |
| `deleted-dev-null` | `--- /dev/null` | **0 건** | – |
| `added-dev-null` | `+++ /dev/null` | `x.go:F` | **`base_hash`** |
| `accidental-sql-comment` | `--- name of the table` | **RuntimeError**: `cannot load existing base file <주석 문장>` | – |

- 1번은 **permissive**: 그 파일의 요구가 통째로 사라진다.
- 2번은 요구가 `revision: base` 로 내려앉아 **편집 전** 논리의 지도가 통과한다.
- 3번은 **거짓 차단**이고 메시지가 말이 안 된다. 이 모양은 오늘 추적 `*.go` 에 **111 줄 / 10 파일** 있다
  (`git grep -P '^(--|\+\+) ' HEAD -- '*.go'` — 전부 raw string 안의 SQL 주석). `/dev/null` 정확형은 0 이라
  오늘 1·2번은 잠복이고 3번만 살아 있다.

**이 결함은 7.5.2.x 가 만든 것이 아니고 원장이 원리상 못 보는 것이다.** `changed_existing_functions` 는 세 로트가
한 번도 안 건드렸고, 입력을 자식 프로세스가 읽으므로 깔때기·원장에 안 남는다. 경합도 아니다 — 재실행해도 같다.

**`revision: base` 를 같이 쟀다** (선택지 1 의 "같이 본다"). 저장소의 `ast.json` **3,075** 중 `revision` 칸이
`base` 인 것 **266** · `current` **599**. 그 266 의 `source_sha256` 을 각 change 의 `base-commit.txt` 커밋에서
꺼낸 **파일 바이트**와 대조: **MATCH 248 · STALE 18**. 그러니까 `validate_target` 에
파일 수준 해시 대조를 넣으면 **오늘 정상인 번들 18개를 새로 거절**한다.
(**7.5.22 정정 — 여기 적었던 귀속이 거짓이다.** "a063 2 · a112 16" 이 아니라
**a112 13 · a063 2 · 아카이브 a060 2 · 아카이브 a044 1** 이다 — 활성 **15** · 아카이브 **3**.
"그중 16 은 지금 이 브랜치에 살아 있는 a112" 도 따라서 거짓이었다. 사람이 7.5.7 을 결정할 근거가
바로 이 내역이므로 틀린 채로 두면 결정이 오염된다.) 이 change 밖의 판정을 바꾸는 규칙 변경이라 7.5.4 · 7.5.6 과 같은 **사람 결정**으로 넘긴다 →
**task 7.5.7**. ([[fail-closed-must-name-what-it-rejects]]: 거부할 정상 입력을 먼저 셌다.)

**A/B 를 값싸게 증명할 수 있는 길을 먼저 쟀다.** 수리는 **본문에 헤더 모양 줄이 있을 때만** 다르게 답한다
(훅 적재도 이름 해석도 그 밖에서는 글자 그대로 같다). 그래서 `base-commit.txt` 를 가진 change **117** 전부에 대해
base→HEAD 의 진짜 diff 를 흘려 본문의 `^(--- |\+\+\+ )` 를 셌다 — **합계 0 · 그런 줄을 가진 change 0**.
따라서 오늘 이 저장소의 어느 change 에서도 두 파서의 `required` 는 같다. 실물 A/B 는 그 위의 확인이다.

## VERIFY — task 7.5.2.4 (통합 diff 의 본문 줄은 파일 헤더가 아니다) (2026-09-22)

**근본 수리 한 줄: 파서가 diff 문법대로 상태를 갖는다.** `diff --git ` 이 파일을 열고, 첫 `@@` 가 본문을 열고,
본문에서는 `@@` 만 읽는다. `--unified=0` 이면 본문은 전부 `-`·`+`·`\` 로 시작하므로 열 0 의 `diff --git ` 과
`@@` 는 모호하지 않고, 이름은 **첫 훅 앞에서만** 읽으면 된다. 훅 적재는 두 자리가 같은 말을 하지 않도록
`hunk()` 한 곳으로 모았다. **새 거절은 0 개다** — 본문에서 이름을 안 읽을 뿐이다.
(7.5.22 정정: 이 문장을 **유도**처럼 쓴 것이 틀렸다. `deleted-dev-null` 모양은 편집 전에 `required 0` 으로
**통과**했고 지금은 요구가 생긴다 — 그러니까 "아무것도 새로 거절하지 않는다" 는 수리의 성질이 아니라
**열거표 117/0 이라는 측정**이다. 측정과 유도를 가르는 것이 [[correction-unit-must-be-the-value]] 의 요점이다.)
판정이 느슨해지는 방향은 없다:
본문을 이름으로 읽어 **사라지던** 요구가 돌아오고 **엉뚱한 이름 아래 생기던** 요구는 제자리를 찾는다
(개수는 늘 수도 줄 수도 있다 — 없어지는 것은 틀린 요구뿐이다). 위 3번의 거짓 차단도 사라진다.

**RED → GREEN.** 새 시험 클래스 `TheDiffBodyDoesNotNameTheFileUnderJudgement` 여섯. 지어낸 diff 문자열이 아니라
진짜 저장소·진짜 `git diff` 다. 첫 시험은 **픽스처가 허구가 아님**부터 못 박는다(`--- /dev/null` 과
`+++ /dev/null` 이 본문에 실제로 나온다). 편집 전 결과: **실패 2 · 오류 1 · 통과 3**(통과 셋은 문법 사실 하나와
진짜 헤더의 뜻을 지키는 대조군 둘). 편집 후 **6/6 통과**.

**변이 — 이 로트의 줄마다.** `75_mut.py` 에 `AB1~AB5` 를 더했다(총 **118**, 앵커 전수 재확인 — 전부 정확히 1회).
`AB1~AB5` **5/5 CAUGHT**, 실패한 시험 이름이 변이 내용과 직접 이어진다:

| 변이 | 잡는 시험 |
|---|---|
| `AB1_body_lines_name_the_file_again` | 결함 시험 셋 전부 |
| `AB2_the_body_never_opens` | 결함 시험 셋 전부 |
| `AB3_a_new_file_does_not_close_the_body` | `test_a_real_deleted_file_header_still_means_there_is_no_current_logic` 외 2 |
| `AB4_only_the_first_hunk_of_a_file_counts` | 결함 시험 셋 전부 |
| `AB5_dev_null_is_an_ordinary_name` | `test_a_real_added_file_header_still_means_there_is_no_base_logic` |

`V4` 의 앵커가 이 수리로 **2회**가 되자 하네스가 `assert` 로 멈췄다(조용히 아무 자리나 고르지 않는다) — 앞 줄을
붙여 자리를 특정했다. [[mutation-must-reach-the-thing-under-test]] 의 "앵커는 소스를 따라 낡는다" 그대로다.

**하네스 계측기 수리 셋.**
1. **도달 계측기가 원본 본문에 표식을 심는다.** `reached()` 가 `pristine` 을 받고 호출은 **원복 뒤**로 옮겼다.
   재는 질문은 "스위트가 이 자리를 도는가" 이고 그 자리는 옛 줄의 자리다.
2. **무변이 대조군이 창 끝에도 돈다.** 시작에만 돌리면 도중에 환경이 무너진 판이 전부 CAUGHT 로 찍힌다
   (리뷰어 실측: `/` 가 0 인 창에서 `AA7 n=130` · `AA8 n=244` · `AA10 n=269`, 비운 뒤 같은 다섯은 `n=1..2`).
   끝 대조군이 빨가면 **그 창의 결과를 전부 버린다**. 판마다 `Ran N` 이 대조군과 다르면 그 줄에 **환경 의심**을 찍는다.
3. **동어반복 단언을 수리라고 적지 않는다.** `pristine == origin` 은 `setup()` 이 방금 무조건 rmtree+copytree 했으므로
   구조상 참이고, 원복 확인은 방금 쓴 것을 읽는다. 실제로 지켜 주는 것은 **pid 별 `WORK` 이름** 하나다 — 주석에 그렇게 적었다.

**Y13 · Y14 를 오늘의 범위로 다시 썼고, Y13 이 살아남았다.** 둘 다 `_verdict` 에 없는 이름 `analysis` 를 넣어
`NameError` 였다. `evidence.directory` 로 바꿔 다시 재니 **Y14 CAUGHT · Y13 SURVIVED · 도달함** — 옛 계측기였다면
"안 닿음" 으로 숨었을 것이다(고친 계측기가 바로 값을 했다). 원인은 둘이다: 행동 시험이 없었고, `exists` 같은
**stat 계열이 깔때기 밖 읽기를 보는 구조 시험의 감시 집합에 없다**(→ task 7.5.14). 7.5.2.3 이 AA19 에 한 것처럼
시험을 더해 닫았다 — `TheVerdictJudgesTheEvidenceItWasGiven.test_the_early_return_reads_the_snapshot_not_the_directory`
(`_verdict` 의 이른 반환이 `evidence.present` 스냅숏을 쓰는지 직접 잰다). 그 뒤 **Y13 · Y14 둘 다 CAUGHT**.

**A/B — 전수 논증 + 실물.** 논증: 수리가 답을 바꾸는 조건은 "본문에 헤더 모양 줄" 하나뿐인데
`base-commit.txt` 를 가진 change **117** 전부에서 그런 줄이 **0** 이다(MEASURE) → 오늘 이 저장소의 어느
change 에서도 두 파서의 `required` 가 같다. 실물 `main()` 출력 전체 A/B 로 확인: **126/126 SAME ·
DIFFERENT 0**, 결정적 계수도 그대로 — git **20,365 → 20,365** · `ast.json` 읽기 **9,072 → 9,072**
(7.5.2.3 이 적은 값과 같다 — 이 로트의 비용은 **0** 이다). 벽시계도 순서를 번갈아 재서 붙인다:
before 먼저 63건 **935.1s → 933.8s** · after 먼저 63건 **935.0s → 933.1s**. 7.5.2.3 의 초 단위가 재현이
안 됐던 까닭이 이것이다 — `7523_reads.py` 는 before 를 늘 먼저 한 프로세스에서 단일 표본으로 돌았다.

**게이트.** `make lint` rc 0 · `make sdd-test` rc 0(logic-map **380**) · `make test-seams` rc 0
(**7.5.22 정정: 여기 적은 "46 패키지" 는 rtk 가 자른 로그를 센 것이다.** `go list ./...` = **108** 이고
그 타깃은 `./...` 를 돈다 — 빈 출력·작은 수는 "위반 0" 이 아니라 "검사 0" 이라는
[[missing-tool-reports-clean]] 이 정확히 이 실수를 경고한다. 셀 때는 `rtk proxy`.) ·
`openspec validate --all --strict` **58/58** · a122 `check_analysis.py` **rc 0**(`required 0` — 이 change 는
Go 를 안 바꾼다).

**시험 298 → 305.**

### 안 한 것 — 사유

- **P0 (자식 프로세스 읽기가 원장 밖)**: 사람이 선택지 1 을 골랐다. 범위가 다르고(원장의 경계를 옮기는 일)
  이 로트의 A/B 와 변이 증거를 섞지 않으려고 뺐다 → **task 7.5.8**.
- **`revision: base` 의 파일 수준 해시**: 오늘 정상인 번들 **18개**를 새로 거절한다(위 MEASURE) → **task 7.5.7**, 사람 결정.
- **`_safe_changed_go_paths` 의 `"` · `\` · 제어 문자**: 같은 함수 옆이지만 이 로트가 안 만든 P1 이고, 저장소 노출 0
  이며, 결과가 permissive 가 아니라 거짓 차단이다 → **task 7.5.13**.
- **그 함수들의 `timeout=`**: 편집한 함수와 같은 파일이지만 다른 결함이다 → **task 7.5.15**.
- **`READ_CAP` 값의 시험 · AA6 의 행동 못 박기 · `main()` 의 결함 경계**: 전부 P3 → task 7.5.19~7.5.21.
- **113 → 111+2 의 소급 정정**: 기록된 CAUGHT 판정은 **안 바뀐다**. 눈먼 계측기는 SURVIVED 인 판에서만 답을 내고
  7.5.2.3 은 생존이 0 이었다. 바뀐 것은 문장이지 측정이 아니다 — 위 `## 정정` 에 적었다.

## 정정 — task 7.5.2.4 (독립 재리뷰 셋: 적대 · 정확성 · 주장정확성, 2026-09-22)

**리뷰가 확인한 것: 기전 주장은 참, 숫자와 좌표는 거짓.** 파서 결함의 재현(세 모양) · 수리 · 7.5.2.3 의
"깔때기 넷뿐" 철회 · 상태 기계에 남은 흉내 모양 없음 · RED 2실패+1오류 · `AB1~AB5` 5/5 와 그중 `AB1` 이
**새 시험 셋만** 잡는다는 것 · 대조군 둘이 단독으로 변이를 잡는다는 것 · 스펙 시나리오의 적법함은
셋이 각자 확인했다. 아래는 **내가 쓴 거짓**이고 전부 내 손으로 다시 재서 고쳤다.

| # | 내가 쓴 것 | 실제 (재측정, `197a0355`) | 어디 |
|---|---|---|---|
| 1 | `make test-seams` rc 0 (**46 패키지**) | `go list ./...` = **108**. 내 로그에 `... (59 lines truncated)` 가 찍혀 있었다 — **rtk 가 자른 로그를 셌다** | VERIFY |
| 2 | 번들 열거표 **12,411 / 1,633** | **12,412 / 1,634**. **편집 도중** 스냅숏이었다(그 로트가 커밋한 번들 둘이 아직 없을 때) — 커밋한 하네스가 내는 수와 갈렸다 | `check_analysis.py` · README |
| 3 | STALE 귀속 **a063 2 · a112 16** | **a112 13 · a063 2 · 아카이브 a060 2 · 아카이브 a044 1** (활성 15 · 아카이브 3) | VERIFY · tasks 7.5.7 · HANDOFF |
| 4 | "**새 거절은 0 개**" 를 유도로 씀 | `deleted-dev-null` 은 편집 전 `required 0` 으로 **통과**했다. 0 은 **측정**(117/0)이지 수리의 성질이 아니다 | VERIFY · FLM |
| 5 | 새 task 열다섯의 `check_analysis.py` **줄 번호 열셋** | 수가 틀렸다 — `197a0355` 기준 **스무 자리 · 값 열여덟**이다(7.5.23 재측정). "**전부** 편집 전" 의 범위도 적어야 한다: 그 task 들이 **실린** 커밋(`197a0355`)에서는 전부 낡았지만, **쓰인** 커밋(`845471a1`) 기준으로는 스물둘 중 **열여덟이 정확했다** — 같은 커밋의 내 수리가 민 것이 기전이다. 넷은 `845471a1` 에서도 이미 틀렸다. 전부 **함수 이름**으로 바꿨다 | tasks 7.5.7~7.5.21 |
| 6 | `_head_commit` 은 "`check` 한 번에 **셋** 불린다" | 조건부다. 활성 change 27건 실측 **3회 24건 · 0회 3건**(head 를 묻기 전 거절) | docstring |
| 7 | `_reads_moved` 만 죽이면 "**정확히 11개**" | **16개** (시험 313 기준 전수 실행) | VERIFY |
| 8 | 자식 프로세스는 `git diff`·`git show`·`go run`·`execution_baseline` **넷** | 열거가 아니라 **예시**였다. `subprocess.run` 은 **열여섯 자리**이고 원장 항목은 **0** | docstring · README · tasks 7.5.8 |
| 9 | 7.5.8 의 노출 "Go **138개** / 고정 소스 **37개**" | **재현 안 됨.** a112 실측: `required` 를 정하는 Go 파일 **32** 중 **13** 은 원장에 이름조차 없고, 나머지 19 도 인용·시험 색인 때문이지 `go run` 때문이 아니다 | tasks 7.5.8 |
| 10 | README 비용 "계측 시점의 값이라 **±1** 로 움직인다" | **거짓 헤지.** 이틀에 8·16 움직였다. 모집단(번들·시험 파일·인용)의 함수라 커밋마다 움직인다 → 커밋을 적기로 | README |
| 11 | 7.5.9 의 "디스크 963 · 추적 953 · 나머지 10" | 추적 **953** 만 안정. 차이는 전부 `_work/*/extract_go_ast_test.go` 이고 **하네스가 도는 동안 늘었다 준다** — 같은 날 정지 상태 **963**(10), 변이 하네스 도는 중 **964**(11, 열한째가 그 판의 `75_mut_work.<pid>/`) | tasks 7.5.9 |
| 12 | HANDOFF 의 굵게 표시 | **다섯 문단**이다(7.5.23 정정 — 여섯이라 적었다). 여섯째는 이미 균형이 맞았고 내 정규식이 거기 여는 표시를 **더** 넣었다가 되돌렸다 — 고친 것이 아니라 모양을 맞춘 것이다 | HANDOFF |
| 13 | 7.5.10 · 7.5.11 의 인용 | "한 줄씩" 이 아니다 — 7.5.10 은 **한 줄**, 7.5.11 은 **두 줄** 어긋났다(7.5.23 정정) | tasks |
| 14 | 7.5.13 의 사유 "permissive 가 아니라 **거짓 차단**" | **rename 모양에서는 permissive 다**(키에 `base_hash` 만 남아 편집 전 지도가 통과). 편집 모양에서만 참이었다 | tasks 7.5.13 |

**패턴 하나로 묶을 수 있다.** 열넷 중 아홉(1·2·5·6·7·9·10·11·13)이 같은 실수다 — **오염되거나
편집 도중인 측정을 값으로 적었다.** 자른 로그를 세고, 편집 전 좌표를 인용하고, 내 하네스가 만든 파일이 든
모집단을 셌다. 그래서 이 로트의 정정은 값만 바꾸지 않고 **값을 다시 내는 방법**을 같이 바꾼다:
줄 번호를 **함수 이름**으로, 수를 **커밋과 함께**, 세는 것은 `rtk proxy` 로.

**내가 만든 회귀 둘도 같은 로트에서 닫는다** — 아래 `## VERIFY — task 7.5.22` 의 C1 · C2.

**새로 연 task 하나**: 7.5.22 — 이 로트가 닫았다. STALE 귀속 하네스는 task 로 미루지 않고 바로 커밋했다
(`analysis/harness/7522_stale.py`) — 옮긴 자리에서 **실제로 돌려** 보고 절대경로를 유도로 바꿨다.

## MEASURE · Pre-Edit Gate — task 7.5.22 (git 이 본문을 안 내면 요구가 사라진다) (2026-09-22)

**무엇을 재기로 했나.** 사람이 **선택지 2** 를 골랐다 — 정정 로트 + P0(`.gitattributes` 가드) 한 로트.
근거는 둘이다: 기록이 거짓이면 그 위에 쌓는 **사람 결정**(7.5.7 이 바로 그 STALE 내역에 걸려 있다)이
오염되고, 게이트의 중심 요구가 저자의 파일 한 줄로 꺼지는 것은 오늘 노출이 0 이어도 스위치다.

**FLM 먼저.** 편집할 함수는 `_safe_changed_go_paths` 하나다. `enumerate.py` 로 편집 **전** AST 를 뽑았다 —
`analysis/python-function-logic/tools-logic-map--_safe_changed_go_paths/ast.before-7.5.22.json`
(`:104-125` · 분기 **11** · 반환 0 · raise 3 · 호출 11).

**결함은 진짜 git 으로 재현했다.** 빈 저장소에 `x.go` 하나를 넣고 편집한 뒤, 워킹트리에
`.gitattributes` 한 줄을 놓는다 — **추적하지 않는다**:

| 입력 | `git diff --unified=0` | `--name-only`(오늘의 가드) | `--numstat` | `required` |
|---|---|---|---|---|
| 평범한 편집 | `@@ -4 +4 @@` | `x.go` | `1\t1\tx.go` | `x.go:F` |
| `chmod +x` (mode-only) | 훅 없음 | `x.go` | **`0\t0\t`**`x.go` | 0 건 — **맞다** |
| `*.go binary` | `Binary files … differ` | `x.go` | **`-\t-\t`**`x.go` | **0 건 — 틀렸다** |
| `*.go -diff` | 같음 | `x.go` | `-\t-\tx.go` | **0 건 — 틀렸다** |

세 번째·네 번째가 결함이다. **판정 줄이 안 나간다** — 그 change 의 Go 요구가 통째로 사라지고 게이트는
`evidence complete` 를 낸다. 두 번째 줄이 "훅이 0 개면 거절" 을 못 쓰는 이유다: mode-only 는 정상 입력이고
요구가 0 건인 것이 맞다. `--numstat` 만이 그 둘을 가른다.

**끝까지 이어지는 영수증 — CLI 수준으로 쟀다** (`analysis/harness/7522_switch.py`). 공유 워크트리를
건드리지 않으려고 `git worktree add --detach` 로 **격리 worktree** 를 만들고 그 안에만 `.gitattributes`
한 줄을 놓았다(추적 안 함). 대상은 오늘 판정 줄이 실제로 **36개** 나오는 a112 다:

| | 편집 전(`197a0355`) | 편집 후 |
|---|---|---|
| 평소 | 판정 줄 **36** · `required` **64** | **같다** (36 · 64) |
| `.gitattributes` 한 줄 (추적 **안 함**) | 판정 줄 **0** · `required` **0** · **`evidence complete`** | 판정 줄 **1** — `cannot derive modified Go functions: modified Go file has no textual diff (binary or -diff attribute): …` |

**저자가 커밋도 안 한 한 줄이 판정 줄 36개를 0개로 만들고 게이트를 통과시켰다.** 이것이 이 P0 의
전부다. 수리 뒤에는 같은 입력이 **이름 댄 거절**이 된다.

**둘째 결함이 같은 자리에 있었다.** `--name-only` 은 rename 에서 **새** 이름 하나만 낸다. 그런데 파서가
`base_file` 에 넘기는 것은 `--- a/<옛 이름>` 에서 뽑은 **옛** 이름이다. 탭이 든 이름에서 `git mv` 하면
가드가 통과시키고 파서가 `cannot load existing base file <sha>:"a/a\tb.go"` 로 **거짓 차단**한다 —
7.5.13 이 적은 모양이 rename 갈래에서 살아 있었다(RED 에서 그 메시지가 그대로 나왔다).

**거부할 정상 입력을 먼저 셌다** ([[fail-closed-must-name-what-it-rejects]]). `base-commit.txt` 를 가진
change **117** 전부에서 base→HEAD 의 `*.go` 파일 줄 **97,235**:

| 모양 | 오늘 수 |
|---|---|
| 본문 억제 `-`/`-` — **새 거절이 걸 것** | **0** |
| mode-only `0`/`0` — 거절하면 **안** 되는 것 | 0 |
| rename 쌍 (이름 검사가 새로 보는 것) | **0** |
| 이름 거절(UTF-8 아님 · `\n\r\t`) | 0 |

**0 은 측정이다** — 이번엔 유도로 쓰지 않는다(7.5.2.4 에서 그렇게 틀렸다). 하네스는
`analysis/harness/7522_numstat.py` 로 커밋했다. 충돌 중인 워킹트리도 재 봤다: `numstat` 이 평범한 3칸
레코드를 낸다(`5\t1\tx.go`) — 새 거절도, 새 결함도 아니다.

**A/B 를 값싸게 증명할 수 있는 길을 먼저 쟀다.** 수리는 같은 자리에서 **형식만** 바꾼다. 그래서 가드
하나만 떼어 `--name-only` 판본과 `--numstat` 판본을 change 117 × target 둘로 돌려 **올라오는 것까지
글자 그대로** 비교했다 — **234/234 SAME · DIFFERENT 0**(`analysis/harness/7522_guard_ab.py`).
실물 `main()` A/B 는 그 위의 확인이다.

## VERIFY — task 7.5.22 (git 이 본문을 안 내면 요구가 사라진다) (2026-09-22)

**근본 수리: 앞단 가드가 `--name-only` 대신 `--numstat` 을 읽는다.** 같은 자리, 같은 호출 수 —
형식만 바꿨다. 레코드 해독은 `_numstat_records` 로 뺐고(`<더함>\t<지움>\t<경로>\0`, rename 은 경로 칸이
비고 다음 두 칸이 옛·새 이름), 모양이 다르면 지어내지 않고 결함으로 올린다. **새 거절은 가장 뒤에 선다** —
이름 검사를 **전부** 마친 뒤에 본문을 묻는다([[a-new-guard-unpins-the-guards-behind-it]]: 앞에 세우면
이름 가드의 시험이 남의 가드를 잰다). 판정이 느슨해지는 방향은 없다: 오늘 **조용히 사라지던** 요구가
**이름 댄 거절**이 된다(`cannot derive modified Go functions: modified Go file has no textual diff …`).

**RED → GREEN.** 새 시험 클래스 `AGoFileWithNoTextualDiffIsNotSilentlyEmpty` 일곱. 지어낸 문자열이 아니라
진짜 저장소·진짜 git 이다. 편집 전: **실패 4 · 통과 3**. 통과한 셋은 **통과해야 하는 셋**이다 —

| 시험 | 편집 전 | 무엇을 지키나 |
|---|---|---|
| `…git_really_suppresses_the_body_for_a_binary_marked_go_file` | 통과 | **양성 대조군** — git 이 정말 `Binary files` 를 내고 `@@` 를 안 내며 `--numstat` 이 `-\t-\t` 다 |
| `…a_mode_only_change_is_not_refused` | 통과 | **거절의 경계** — 훅 0 개지만 정상 입력이다 |
| `…the_name_guard_still_speaks_first_when_both_are_wrong` | 통과 | 새 거절이 이름 가드를 안 가린다 |
| `…a_suppressed_body_is_refused_instead_of_silently_requiring_nothing` | **실패** | 본문 억제 거절 |
| `…an_uncommitted_gitattributes_is_enough_to_suppress_the_body` | **실패** | **추적 안 된** 파일로도 된다는 것 |
| `…the_minus_diff_attribute_suppresses_the_body_too` | **실패** | `-diff` 속성도 같다 |
| `…the_guard_reads_both_names_of_a_rename` | **실패** | rename 의 **옛** 이름 |

마지막 시험의 편집 전 메시지가 `cannot load existing base file <sha>:"a/a\tb.go"` 였다 — **7.5.13 이 적은
거짓 차단이 rename 갈래에서 살아 있었다**는 증거다. 편집 후 **7/7 통과**.

**변이 — 이 로트의 줄마다.** `75_mut.py` 에 `AC1~AC5` 를 더했다(총 **123**, 앵커 전수 정확히 1회).
**5/5 CAUGHT · 생존 0**, 대조군은 창 시작과 끝 모두 GREEN(`Ran 313` 양쪽 같음):

| 변이 | 빨개진 시험 |
|---|---|
| `AC1_a_suppressed_body_is_ordinary` | 억제 시험 **셋 전부** |
| `AC2_mode_only_is_refused_too` | `…a_mode_only_change_is_not_refused` **하나만** |
| `AC3_a_rename_keeps_only_the_new_name` | `…the_guard_reads_both_names_of_a_rename` **하나만** |
| `AC4_the_new_guard_stands_first` | `…the_name_guard_still_speaks_first_when_both_are_wrong` **하나만** |
| `AC5_the_hunk_reads_the_base_side_twice` | `…the_hunk_keeps_the_new_side_coordinates…` **하나만** (C2) |

### 내가 만든 회귀 둘 — 같은 로트에서 닫았다

**C1 — 도달 계측기가 거짓말했다.** 7.5.2.4 가 표식 심기를 원복 뒤로 옮겨 **눈먼** 것은 고쳤지만,
표식을 **문장이 아닌 자리** 앞에 심으면(여러 줄 호출의 중간 줄 · `elif`/`except` 머리) 파이썬 문법이 깨진다.
그러면 인터프리터가 `SyntaxError` 트레이스백을 찍고 **그 트레이스백이 문제의 소스 줄을 그대로 인쇄한다** —
그 줄에 `REACHED` 가 들어 있으므로 `"REACHED" in output` 이 참이 된다. **시험을 0개 돌린 판이 "도달함"** 이다.
정적 전수(`analysis/harness/7522_marker.py`, 스위트를 안 돌리므로 싸다): 변이 **118** 중 **30**
(7.5.2.4 의 변이표 기준, 대상은 `197a0355`). **7.5.23 정정 — 그 30 은 한 갈래가 아니다.**
`SyntaxError` 가 어느 줄을 인쇄하느냐로 갈린다: 표식 줄을 인쇄하면 거짓 "도달함"(**26**),
다른 줄을 인쇄하면 거짓 "안 닿음"(**4** — `Y19`·`AA14`·`AB1`·`AB2`). 뒤엣것도 거짓말이고,
**눈먼 것을 진짜 음성과 같게 적는** [[mutation-must-reach-the-thing-under-test]] 의 그 실패다.
한 수로 뭉갠 것이 틀렸다.
수리는 셋이다 — (1) 쓰기 전에 `compile()` 하고 깨지는 자리에는 **안 심는다** · (2) `Ran N` 이 대조군과
같을 때만 표식을 읽는다 · (3) **"안 닿음" 과 "못 쟀다" 를 가른다**(앞 판본은 둘 다 `False` 였다).
셋째가 요점이다: 앵커가 아예 없는 변이 **셋**(`Y3` · `Y4` · **`AB4` — 내가 7.5.2.4 에서 더한 것**)이
"안 닿음" 으로 찍히고 있었는데, 그것은 측정이 아니라 침묵이었다.

**C2 — 옮긴 코드의 한쪽을 안 쟀다.** 7.5.2.4 는 훅 적재를 두 자리에서 `hunk()` **한 곳으로 모았는데**,
그 자리를 재는 변이를 안 넣었다. `@@ -a,b +c,d @@` 의 뒤 쌍(현재 쪽)을 앞 쌍(base 쪽)으로 읽는 변이가
**시험 305개 전부를 통과**했고 동등 변이가 아니다: 함수 **위에** 줄을 끼우면 두 쪽 좌표가 벌어져
`current_hash` → **`base_hash`** 로 내려앉는다 — 이 change 가 없애려는 바로 그 `revision: base` 내려앉음이
다른 경로로 살아 있었다. 시험 하나를 더해 못 박았고(`AC5`), 그 시험 **하나만** 빨개진다.
[[transcribed-code-needs-both-sides-pinned]] 가 정확히 이 모양이다 — 옮겼으면 옮긴 자리를 재야 한다.

**A/B — 전수 논증 + 가드 단독 + 실물.** 논증: 수리는 같은 자리에서 **형식만** 바꾸고, 새 거절이 걸
입력은 오늘 **0** 이다(MEASURE 의 97,235 전수) → 오늘 이 저장소의 어느 change 에서도 답이 같다.
가드 단독 A/B **234/234 SAME · DIFFERENT 0**. 실물 `main()` 출력 전체 A/B: **126/126 SAME ·
DIFFERENT 0** — 결정적 계수도 그대로다(git **20,365 → 20,365** · `ast.json` 읽기 **9,072 → 9,072**,
7.5.2.4 가 적은 값과 같다). **호출이 하나도 안 늘었다** — 같은 자리에서 형식만 바꿨다는 것의 증거다.
벽시계는 순서를 번갈아 재서 붙인다: before 먼저 63건 **988.1s → 993.1s** · after 먼저 63건
**994.9s → 997.2s**. 두 순서 모두 after 가 **+0.3%** 로 같은 방향이다 — `--numstat` 출력이 `--name-only`
보다 크고 레코드를 두 번 도는 값이다. 숨기지 않고 적는다.

**A/B 를 돌리는 법을 여기서 한 번 적어 둔다 (이 로트에서 두 번 헛돌렸다).** `7521_main_ab.py` 의
이어달리기 키는 `repo_state()` = `HEAD` + **워킹트리 전체 diff** + 추적 안 된 파일 목록이다(의도된
설계 — 판본이 섞이는 것을 막는다). 그래서 **문서를 한 줄이라도 고치면 기록이 무효**가 되고 처음부터
돈다. 나는 A/B 를 띄운 **뒤에** docstring 을 고쳤고(1차 104/104 SAME 이 다른 바이트 기준이 됐다),
그 뒤 정정 문서를 쓰는 동안 또 한 번 키가 갈렸다. **A/B 는 로트의 마지막 편집 뒤에, 한 번에** 돌린다
([[a-measurement-carries-its-moment]]).

**게이트.** `make lint` rc **0** · `make sdd-test` rc **0**(logic-map 380 → **388**) ·
`make test-seams` rc **0** — **108 패키지**(`ok` 100 · `[no test files]` 8 · FAIL **0**), `go list ./...` 과
같은 수다. 이번엔 `rtk proxy` 로 세었다([[missing-tool-reports-clean]]) ·
`openspec validate --all --strict` **58/58** · a122 `check_analysis.py` **rc 0**(`required 0`).

**시험 305 → 313** (본문 없는 `*.go` 일곱 + 훅 좌표 하나).

### 안 한 것 — 사유

- **`analysis/python-function-logic/*/function-logic-map.md` 머리말의 절대 줄 ~60개**: 안 고쳤다.
  그것들은 **살아 있는 인용이 아니라 추출 시점의 기록**이다 — 저마다 `head_commit` 과
  `source_sha256` 을 가진 `ast.json` 이 옆에 있고, 그 범위가 바로 "그 리비전에서 이 함수가 여기 있었다"
  라는 산출물이다. 고치면 기록이 아니라 주장이 된다. 이 로트가 고친 것은 **살아 있는 인용**
  (새 task 열다섯 · 산문 · docstring)뿐이고, 같은 파일 안에서도 살아 있던 하나(`_verdict:2162`)는 고쳤다.
- **7.5.13 의 나머지 절반**(`"` · `\` · 나머지 제어 문자): 이 로트는 rename 의 **옛 이름**을 가드 범위에
  넣어 그 모양 하나를 닫았을 뿐이다. 남은 인용 문자들은 같은 함수지만 다른 결함이고, 저장소 노출 0 이다.
- **`_numstat_records` 의 결함 갈래 둘**(`cannot read … record` · `… rename record`)에 **행동 시험이 없다.**
  오늘 그 모양을 내는 진짜 git 입력을 못 만들었다(충돌 중 워킹트리도 평범한 3칸이었다).
  **지어낸 바이트로 시험을 쓰면 그것은 git 의 증거가 아니라 내 상상의 증거다** — 안 썼다.
  branch-test-map 에 그대로 적었다.
- **`--find-renames` 를 가드 호출에 안 넣었다.** `diff.renames` 기본이 참이라 오늘은 파서와 같은 집합을
  보고, 거짓이면 가드가 rename 을 삭제+추가 **둘**로 보므로 이름을 **더** 본다(초집합이라 안전).
- **7.5.8(자식 프로세스 읽기가 원장 밖)**: 여전히 P0 로 열려 있다. 이 로트는 그 자리 하나
  (`_safe_changed_go_paths` 의 `git diff`)의 **읽는 내용**을 바꿨을 뿐 원장의 경계를 안 옮겼다.
- **7.5.15 의 `timeout=`**: 내가 고친 바로 그 함수의 `subprocess.run` 에 `timeout=` 이 없다 —
  **일부러 안 넣었다.** 그 결함은 자리 일곱(이 파일 둘 · `execution_baseline.py` 다섯)에 걸쳐 있고
  한 자리만 고치면 나머지 여섯의 시험이 그 한 자리로 초록이 된다([[correction-unit-must-be-the-value]]).
  7.5.15 에서 **값 단위로** 한 번에 닫는다.
- **사람 결정 7.5.3 · 7.5.4 · 7.5.5 · 7.5.6 · 7.5.7**: 그대로 대기. 7.5.7 의 근거 내역은 이 로트가
  다시 재서 고쳤다(`analysis/harness/7522_stale.py`).

## 정정 — task 7.5.22 (독립 재리뷰 셋: 적대 · 시험품질 · 주장정확성, 2026-09-23)

**리뷰가 확인한 것.** P0 자체와 무거운 증거는 전부 재현됐다 — `.gitattributes` 살인 스위치(a112
36/64 → 0/0/`evidence complete`), 가드 A/B 234/234, 97,235 행 열거, STALE 귀속, `test-seams` 108,
`_reads_moved` 16, `subprocess.run` 16(무 timeout 정확히 둘), `_head_commit` 27건 히스토그램,
7.5.9 오염 기전, 변이 5/5 와 그 매핑, C2 가 305 시험을 전부 통과했다는 것, 시험 305→313, 게이트 전부.

**그런데 정정 커밋이 자기 실수를 되풀이했다.** 아래는 전부 내 손으로 재측정해 확인했다.

| # | 내가 쓴 것 | 실제 (재측정) |
|---|---|---|
| G1 | 열거표 **12,412 / 1,634** | 커밋 시점에 **12,414 / 1,636**. 12,412 는 **부모**의 값이다 — **이 로트가 더한 번들 둘**이 빠진 채로 셌다. 앞 로트가 12,411 을 쓴 것과 **글자 그대로 같은 기전**이고, 같은 표의 2번 행이 그것을 꾸짖고 있었다 |
| G2 | `_safe_changed_go_paths` **`:135-179`** · `_numstat_records` **`:104-132`** | **`:137-181`** · **`:106-134`**. 내가 같은 커밋에서 `READ_CAP` 위에 주석 두 줄을 더해 민 것이다 |
| G3 | 새 `branch-test-map.md` 의 **줄 칸 전부** | 같은 +2. `:174`·`:175` 는 이제 **이름 가드**라 분기를 잘못 가리킨다 |
| G4 | 두 `ast.after-7.5.22*.json` | `source_sha256` 가 **역사의 어느 blob 과도 안 맞는다** — 커밋 안 한 중간 상태를 기술한다. "`ast.json` 이 옆에 있으니 기록이다" 라는 내 사유가 **내가 만든 산출물에서** 깨졌다 |
| G5 | RED `Ran 6 · 통과 2` (BTM) | **`Ran 7 · 실패 4 · 통과 3`**. 순서 시험을 더하기 **전**에 재고 안 고쳤다 — 같은 커밋의 review·tasks 와도 어긋났다 |
| G6 | 줄 번호 "**열셋** · **전부** 편집 전" | `197a0355` 기준 **스무 자리 · 값 열여덟**. "전부" 는 그 커밋 기준으로만 참이고, **쓰인** 커밋(`845471a1`) 기준으로는 스물둘 중 열여덟이 정확했다 |
| G7 | 매달린 굵게 **여섯 문단** | **다섯**. 여섯째는 이미 균형이 맞았다 |
| G8 | 원장 항목 **1,743** | 깨끗한 체크아웃에서 1,734 · 하네스가 돌면 1,745. **7.5.9 가 지적한 오염을 같은 문단에서 저질렀다** |
| G9 | 자식 프로세스 "6 + 역사 **아홉**" | **열**이다. 6 + 9 = 15 라 바로 앞에 적은 16 과 안 맞았다 — 열거를 고치면서 열거를 틀렸다 |
| G10 | "살아 있는 인용은 전부 함수 이름으로" | `review.md` 에 **`_verdict:2159` 가 살아남았다**. `197a0355` 에서도 이미 틀렸고 지금은 함수 **밖**이다 |
| G11 | 7.5.10·7.5.11 "**한 줄씩**" | 한 줄과 **두 줄** |
| G12 | 7.5.13 의 새 사유 | **앞뒤가 안 맞는다** — "permissive 다" 라고 적고 그 증거로 **거짓 차단** 메시지를 들었다. 두 모양(편집 / rename)이 반대 결과를 내는데 한 문장으로 적었다 |

**같은 실수가 세 로트 연속이다.** 기전은 매번 같다 — **수리와 인용이 한 커밋에 있으면 인용이 먼저
쓰이고 수리가 나중에 민다.** 그래서 이번에는 값만 고치지 않고 **규칙을 세는 것**으로 만들었다:
`analysis/harness/7523_coords.py` 가 (1) **열린** task 의 살아 있는 좌표, (2) 지문이 어디에도 없는
`ast*.json`, (3) 지문은 맞는데 범위가 틀린 `ast*.json` 셋을 세고 하나라도 있으면 rc≠0 이다.
그 하네스가 G4 를 바로 잡았고, 덤으로 **7.2.6 이 남긴 같은 종류의 뜬 번들 둘**도 찾아냈다
(지문 `6764d06a…`, 어느 커밋에도 없다 — `06e7eb74` 에서 다시 뽑아 되살렸다).
산문 전체의 절대 인용 **106 건**은 이 change 의 누적이라 세기만 하고 판정에는 안 넣었다 —
줄이는 것은 다음 로트다.

## MEASURE · Pre-Edit Gate — task 7.5.23 (가드와 판정이 같은 diff 를 읽는다) (2026-09-23)

**무엇을 재기로 했나.** 7.5.22 가 닫았다고 한 문이 **안 닫혔다**. 적대 리뷰어가 냈고 내가 재현했다.
사람이 고른 선택지 2("정정 + P0")의 P0 이 미완이므로 새 범위가 아니라 **같은 일의 마무리**다.

**FLM 먼저.** 편집할 함수는 `_safe_changed_go_paths` 와 `changed_existing_functions` 둘.
편집 전 AST 는 `ast.before-7.5.23.json`(각 디렉터리), 대상은 `d7d494ee`.

**결함은 진짜 git 으로 재현했다.** 추적 **안 된** `.gitattributes` 한 줄 + config 키 하나:

| 입력 | 가드(`--numstat`) | 판정(훅 수) | 실물 게이트 a112 |
|---|---|---|---|
| `*.go binary` | `-`/`-` → 거절 | 0 | 7.5.22 가 닫았다 |
| `*.go diff=nop` + `diff.nop.textconv=true` | **`1`/`1` — 정상** | **0** | **required 0 · `evidence complete` · rc 0** |
| `*.go diff=ext` + `diff.ext.command` | `1`/`1` | 1 | `--no-ext-diff` 가 막고 있었다 |

**기전은 7.5.22 가 진단한 것 그대로다.** `--numstat` 은 **원본 blob** 을 세고 판정 diff 는
**textconv 출력**을 읽는다 — 가드와 판정이 여전히 **다른 투영**을 봤다. 그래서 7.5.22 가 적은
"`--numstat` 이 정확히 가른다" 는 거짓이고, 스펙 시나리오도 WHEN 이 일반("본문을 안 낼 때")인데
THEN 이 `-`/`-` 로 좁혀져 **구현이 자기 WHEN 을 만족하지 않았다**(둘 다 이 로트에서 고쳤다).

**`--no-ext-diff` 는 지워도 되는 깃발이었다.** 실측: 있을 때 훅 1 · 없을 때 0. 그런데 그것을 지우는
변이가 **313 시험을 전부 통과**했다 — 문을 닫고 있던 것이 시험에 없었다.

**거부할 정상 입력을 먼저 셌다.** 저장소에 `.gitattributes` 는 **0 개**이고, 117 change · `*.go`
파일 줄 97,235 에서 새 거절이 걸 것은 **0** 이다. 교차 검사가 거는 것도 0 이다 — 깃발 둘이 살아
있으면 본문이 안 사라진다. textconv 저장소에서도 결과는 거절이 아니라 **올바른 판정**이다(실측).

## VERIFY — task 7.5.23 (가드와 판정이 같은 diff 를 읽는다) (2026-09-23)

**근본 수리는 둘이고, 둘째가 첫째를 못 박는다.**

1. 판정 diff 가 `--no-textconv` 를 준다(`--no-ext-diff` 옆).
2. **두 투영을 맞춰 본다.** 가드가 `--numstat` 레코드를 **돌려주고**, 파서가 훅을 낸 파일 집합
   (`bodied`)을 모은 뒤, numstat 이 내용이 바뀌었다고(`0`/`0` 도 `-`/`-` 도 아니라고) 한 파일이
   훅을 하나도 안 냈으면 거절한다. 가드도 `--find-renames` 를 줘서 두 호출이 같은 짝을 본다.

(2)는 **문을 하나씩 세지 않는다.** 어긋남을 보므로 textconv · 외부 diff · 아직 모르는 git 기능이
같은 그물에 걸린다. 실측 — 깃발을 지우면 런타임에 잡힌다:

| 코드 | 저장소 | 결과 |
|---|---|---|
| 온전 | textconv | `required [('x.go','F')]` — 올바른 판정 |
| `--no-textconv` 지움 | textconv | **`RuntimeError: git reported content changes but emitted no diff body for x.go`** |
| `--no-ext-diff` 지움 | 외부 diff | **같은 거절** |
| 교차 검사까지 지움 | textconv | `required []` — 조용한 구멍(그래서 교차 검사가 버티는 것이다) |

**RED → GREEN.** 새 시험 클래스 둘 + 기존 클래스에 둘. 편집 전 `TheGuardAndTheJudgementReadTheSameDiff`
의 행동 시험 넷 중 textconv 하나가 빨갛고 나머지는 이유가 각각이다. 편집 후 전부 초록.
`TheNumstatTableIsNeverInvented` 다섯은 **지어낸 바이트**를 쓰는 유일한 시험이다 — 거기서 재는 것이
git 의 행동이 아니라 **순수 함수의 결함 처리**이기 때문이고, 7.5.22 가 "진짜 git 으로 못 닿는다" 며
못 박지 못했던 두 `raise` 를 이것이 닫는다.

**변이.** `AD1~AD4` · `AD6` 을 더했고 `AB2` 의 앵커를 오늘 소스로 다시 겨눴다.
(7.5.24 정정: 이 문장은 "`AB1`·`AB2`·`AB4` 셋" 이라고 적었는데 커밋된 표에서 바뀐 것은 **AB2 하나**다.
로트 도중 셋을 다 옮겼다가 중복을 없애면서 `AB1`·`AB4` 가 원래 모양으로 **되돌아갔고**, 산문은
그 중간 상태를 적었다 — 이 로트가 고치려던 바로 그 실수다.)
`AD5`(본문 표시를 한 자리만)는 **동등 변이**였다: 파일의 첫 훅이 `elif hunk(line)` 로 들어와 이름을
이미 넣고, 본문에서는 이름이 안 바뀌므로(7.5.2.4) 둘째 자리는 같은 것을 다시 넣을 뿐이다.
동등 변이로 **적지 않고 중복을 없앴다** — 겨눌 자리가 사라졌다.
`AD6`(가드의 `--find-renames`)은 오늘 기본 설정에서 행동이 같아 행동 시험이 못 잡는다 →
**구조 시험**으로 못 박았다(`test_both_git_calls_see_the_same_files`: 두 git 호출이
`--no-ext-diff`·`--no-textconv`·`--find-renames` 를 다 갖는지 AST 로 본다).

**계측기를 또 고쳤다 — 이번엔 "엉뚱한 줄".** 7.5.22 는 표식이 **문법을 깨는** 경우를 막았지만,
표식 head 가 파일에서 **유일하지 않으면** `str.replace(…, 1)` 이 첫 자리에 심는다. 문법은 멀쩡하므로
`compile()` 이 못 본다 — 계측기가 **한 번도 안 본 줄**에 대해 자신 있게 답한다. 실측: 표식 head
**20 개**가 유일하지 않고 그중 **다섯**이 다른 문장에(**넷은 다른 함수**에) 앉았다. 넷을 더 막았다:
유일할 때만 심고 · **일부만** 심겼으면 `"?"` 이고 · 심을 자리가 **하나도 없으면**(주석만 바꾸는 변이)
`"?"` 이며(앞 판본은 표식 없는 판을 돌려 놓고 "안 닿음" 이라 답했다 — 내 수리가 낸 새 구멍이다) ·
원복이 `finally` 안에 있다. `control_ran` 의 기본값도 없앴다 — 0 이면 `Ran N` 대조가 조용히 꺼진다.

**계측기를 처음으로 직접 시험했다** (`analysis/harness/7523_reached.py`). `reached()` 는 **SURVIVED
인 판에서만** 불리므로 생존 0 인 로트에서는 증거에 안 나타난다 — 7.5.22 가 정확히 그랬다.
여섯 경우 **6/6**: 도달함 · 안 닿음 · 유일하지 않은 head · 문법 깨짐 · 심을 자리 없음 · `Ran N` 불일치.

**좌표를 규칙이 아니라 세는 것으로 만들었다** (`analysis/harness/7523_coords.py`, 위 `## 정정` 참조).

**A/B.** 실물 `main()` 출력 전체 **126/126 SAME · DIFFERENT 0**(기준 `d7d494ee`). 결정적 계수도
그대로다 — git **20,365 → 20,365** · `ast.json` 읽기 **9,072 → 9,072**. **호출이 하나도 안 늘었다**:
가드는 같은 자리에서 형식만 바꿨고, 교차 검사는 이미 읽은 두 결과를 맞대 볼 뿐이다.
벽시계는 순서를 번갈아: before 먼저 63건 **940.2 → 942.9s** · after 먼저 63건 **946.0 → 941.2s** —
두 순서가 반대 방향이라 차이는 잡음이다. A/B 는 **마지막 편집 뒤에 한 번에** 돌렸다(7.5.22 에서
두 번 헛돌린 뒤 배운 것 — 이어달리기 키가 워킹트리 전체에 묶여 있다).

**게이트.** `make lint` rc **0** · `make sdd-test` rc **0**(logic-map 388 → **401**) ·
`make test-seams` rc **0**(**108 패키지** · FAIL 0, `rtk proxy` 로 셌다) ·
`openspec validate --all --strict` **58/58** · a122 `check_analysis.py` **rc 0** ·
`7523_coords.py` **rc 0**(번들 332 전부 실재하는 소스 · 열린 task 의 살아 있는 좌표 0).

**시험 313 → 326.**

### 안 한 것 — 사유

- **산문 전체의 절대 인용 106 건**: 이 change 의 누적이고 대부분이 닫힌 task·옛 FLM 의 기록이다.
  하네스가 **세고 있으므로** 다음 로트가 줄인다. 판정에 넣으면 이 로트가 무관한 파일 수십 개를
  건드려야 하고, 그것이야말로 [[correction-unit-must-be-the-value]] 가 경고하는 모양이다.
- **가드의 `timeout=`**: 같은 결함이 일곱 자리(이 파일 둘 · `execution_baseline.py` 다섯)에 걸쳐
  있다. 한 자리만 고치면 나머지 여섯의 시험이 그 한 자리로 초록이 된다 → **값 단위**로 7.5.15.
- **7.5.13 의 나머지**(`"` · `\` · 제어 문자): rename 우회 쪽은 7.5.22 가 닫았고 남은 것은
  거짓 차단이며 저장소 노출 0 이다.
- **이진 rename 의 거절 문장이 어느 이름을 대는가**: 행동은 쟀지만(새 이름) 못 박지 않았다.
- **7.5.8**(자식 프로세스 읽기가 원장 밖): 여전히 P0 로 열려 있다. 이 로트는 그 자리들이 **읽는
  내용**을 바꿨을 뿐 원장의 경계를 안 옮겼다.
- **사람 결정 7.5.3 · 7.5.4 · 7.5.5 · 7.5.6 · 7.5.7**: 그대로 대기.

## 정정 — task 7.5.23 (독립 재리뷰 둘: 적대 · 주장정확성, 2026-09-23)

**적대 리뷰어가 내 수리에서 회귀를 찾았고, 주장정확성 리뷰어가 정정의 거짓 셋을 찾았다. 둘 다 재현했다.**

### 내가 넣은 회귀 — 교차 검사가 정상 입력을 거절했다 (P0, 7.5.24 에서 닫음)

짝을 **이름으로** 지었다: 가드의 `--numstat` 이름은 `-z` 라 **날 바이트**이고, 파서가 아는 이름은
`removeprefix("a/")` 를 거친 **유도값**이며 git 이 인용하면 `"a/we\"ird.go"` 다. 두 이름 공간을
교집합으로 견주니 만나지 않는다. 부모와 A/B 로 잰 결과:

| 입력 | `d7d494ee` | `de5a78b4` |
|---|---|---|
| 평범한 편집 (대조군) | `required [('seed.go','F')]` | 같음 |
| 새 파일 `we"ird.go` | `required []` | **RuntimeError** |
| 새 파일 `back\slash.go` | `required []` | **RuntimeError** |
| 새 파일 `b/new.go` + `diff.noprefix=true` | `required []` | **RuntimeError** |

거절은 `check()` 에서 `cannot derive modified Go functions` 로 잡히므로 **창 줄도 판정 줄도 하나도
안 나간다**. 메시지는 "an external diff or textconv filter is hiding it" 이라고 **없는 필터를 탓했다**.

**그리고 내가 쓴 "새 거절이 오늘 거절하는 정상 입력 0" 은 빈 표본이었다.** 그 수는 **오늘 저장소의
파일 이름**을 센 것이지 거절 가능한 **집합**을 센 것이 아니다 — `.gitattributes` 없이도 닿는다
([[universal-check-passes-on-an-empty-sample]] · [[fail-closed-must-name-what-it-rejects]]).

### 정정이 또 자기 실수를 되풀이했다 — 네 번째 세대

| # | 내가 쓴 것 | 실제 |
|---|---|---|
| H1 | 산문의 절대 인용 "**105 건**" (`review.md`·`tasks.md`) | **106.** 같은 커밋의 `HANDOFF.md`·커밋 메시지·하네스 출력은 106 이었다 — **한 커밋 안에서 갈렸다.** 105 는 **부모**의 값이고, 이 커밋이 더한 인용 하나(`_verdict:2159` 를 **정정하며 적은 것**)가 빠졌다. 12,411 → 12,412 → 105 로 **네 번째 세대** |
| H2 | BTM 의 "`bodied.update(...)` **두 자리**" + `AD5` | 코드에 **한 자리**이고 `AD5` 는 변이표에 **없다**(같은 커밋에서 중복을 없애고 변이를 지웠다). BTM 은 **draft** 를 적었다 |
| H3 | "`AB1`·`AB2`·`AB4` 의 앵커를 다시 겨눴다" | 커밋된 표에서 바뀐 것은 **`AB2` 하나**. 로트 도중 셋을 옮겼다가 중복 제거로 둘이 **되돌아갔고** 산문은 중간 상태를 적었다 |

**셋 다 같은 모양이다** — 하네스나 코드가 내는 값을 산문에 베껴 적고, 그 뒤의 편집이 값을 바꿨다.
`7523_coords.py` 는 지문·범위·열린 task 의 좌표를 셌지만 **산문의 수**는 안 봤다. 이제 본다:
자기가 내는 인용 수를 산문과 대조하고 다르면 rc≠0 이다(H1 을 바로 잡는다).

### 남은 더 큰 구멍 — 사람 결정 7.5.25

교차 검사는 **어긋남**만 본다. `filter.<x>.clean` 은 `git diff` 가 비교하는 **워킹트리 바이트 자체**를
바꾸므로 `--numstat` 이 레코드를 아예 안 낸다 — 어긋남이 없다. 실측(워킹트리 대상, 게이트의 기본):
디스크에 `return 999` 가 있는데 `numstat` 이 빈 글자이고 `required` 가 `[('x.go','Stop')] → []`.
같은 계열로 `git update-index --skip-worktree`·`--assume-unchanged` 도 편집을 감춘다.
닫으려면 게이트가 **git 의 어떤 투영도 안 믿고** 자기가 판정하는 바이트를 직접 대조해야 한다 —
이 change 밖의 판정을 바꿀 수 있는 규칙 변경이라 **사람 결정**으로 넘긴다.

**내가 공유 저장소에 남긴 것도 지웠다.** textconv 를 재현하려고 격리 worktree 에서
`git config diff.nop.textconv true` 를 했는데, worktree 는 `.git/config` 를 **공유한다** —
킬 스위치의 절반이 저장소에 남아 있었다. 지웠고 게이트가 `required 64` 인 것을 확인했다.
격리 worktree 는 파일은 격리하지만 **설정은 격리하지 않는다**.

## VERIFY — task 7.5.24 (짝은 이름이 아니라 순서로) (2026-09-23)

**근본 수리: 이름 공간을 아예 비교하지 않는다.** 두 호출은 같은 `git diff` 에 형식만 다르므로
파일을 **같은 순서**로 낸다. 그래서 `bodied` 를 이름 집합이 아니라 **구역마다 한 칸인 목록**으로 두고
`zip(records, bodied)` 로 짝짓는다. 순서가 정말 같은지 먼저 쟀다 — 여섯 모양을 섞은 픽스처:

| # | numstat | 본문 | `diff --git` |
|---|---|---|---|
| 0 | `1/1 aa.go` | 있음 | `a/aa.go b/aa.go` |
| 1 | `0/0 bb.go` (mode-only) | 없음 | `a/bb.go b/bb.go` |
| 2 | `0/0 cc.go→cc2.go` (100% rename) | 없음 | `a/cc.go b/cc2.go` |
| 3 | `1/1 dd.go→dd2.go` (rename+편집) | 있음 | `a/dd.go b/dd2.go` |
| 4 | `1/1 ee.go→new.go` (삭제+추가가 rename 으로) | 있음 | `a/ee.go b/new.go` |

레코드 5 · 구역 5 · 순서 일치. **크기가 다르면 그 자체가 어긋남**이라 따로 거절한다 — 깃발이 사라지면
판정 diff 가 훅이 아니라 **구역 자체**를 안 내므로 이쪽이 먼저 잡는다(실측).

**회귀가 사라졌다**: 위 표의 셋이 다시 부모와 같은 답을 낸다. **textconv 차단은 그대로**다 —
깃발을 각각 지우면 여전히 런타임에 거절한다(실측 셋).

**경계 셋을 새로 못 박았다.**

| 무엇 | 왜 필요했나 | 시험 |
|---|---|---|
| 건너뛰기 조건의 `and` | `or` 로 바꾸면 **한쪽 칸이 0 인 레코드가 전부** 빠진다(더하기만 `N`/`0` · 지우기만 `0`/`N`). 그 변이가 **326 시험을 전부 통과**했다 | `…an_append_only_change_that_lost_its_body_is_still_refused`(subTest 둘) |
| git 이 인용하는 이름 | 거절하면 **안 되는** 정상 입력이다 — 내 회귀가 정확히 그것이었다 | `…a_name_git_has_to_quote_is_not_called_a_vanished_body`(subTest 둘) |
| 두 시야의 **크기** | 깃발이 살아 있으면 어느 시험도 이 갈래를 안 돌아, 검사를 지우는 변이가 **328 시험을 전부 통과**했다 | `…a_judged_diff_that_lists_fewer_files_is_refused` |

**가드를 mock 하던 옛 시험 셋도 고쳤다.** `mock.patch("…_safe_changed_go_paths")` 가 `MagicMock` 을
돌려주어 `len(records)` 가 **0** 이 됐다 — "바뀐 파일 0" 이라는 뜻이 되어 길이 검사가 빨개졌다.
셋 다 레코드를 같이 주게 했다. 두 호출이 **같은 diff 를 봐야** 교차 검사가 성립하고, mock 픽스처도
그 계약을 지켜야 한다.

**변이.** `AE1~AE4` 를 더했고 `AB2`·`AD3` 의 앵커를 오늘 소스로 다시 겨눴다.
전수 **19/19 CAUGHT · 생존 0**(총 132, 대조군 창 양끝 GREEN `Ran 329`). `AE2`(크기 검사)는 첫 판에서
**살아남았고** 계측기가 "도달함" 이라 답해 — 안 닿은 것이 아니라 **시험이 없던 것**이라 — 시험을
하나 더 낳았다. 7.5.2.3 의 `AA19`, 7.5.2.4 의 `Y13` 과 같은 모양이다.

**A/B.** 실물 `main()` 출력 전체 **126/126 SAME · DIFFERENT 0**(기준 `de5a78b4`). 계수 불변 —
git **20,365 → 20,365** · `ast.json` 읽기 **9,072 → 9,072**. 벽시계는 순서를 번갈아:
991.4 → 961.6s · 962.3 → 993.2s — 두 순서가 반대라 차이는 잡음이다.

**게이트.** `make lint` rc **0** · `make sdd-test` rc **0**(logic-map 401 → **404**) ·
`make test-seams` rc **0**(**108 패키지**) · `openspec validate --all --strict` **58/58** ·
a122 게이트 **rc 0** · `7523_coords.py` **rc 0**(번들 335 · 열린 task 좌표 0 · 산문의 수 0 불일치).

**시험 326 → 329.**

### 안 한 것 — 사유

- **7.5.25(워킹트리를 다시 쓰는 것)**: **사람 결정**으로 넘겼다. 닫으려면 게이트가 git 의 투영이 아니라
  자기가 판정하는 바이트를 직접 대조해야 하고, 그것은 이 change 밖의 판정을 바꿀 수 있는 규칙 변경이다.
- **`diff.noprefix`·`diff.mnemonicPrefix` 아래의 `removeprefix`**: 이 로트의 교차 검사는 이름을 안 보게
  고쳤지만, **파서**는 여전히 `removeprefix("a/")` 로 이름을 정한다. `a/`·`b/` 로 시작하는 경로가
  있으면 요구가 엉뚱해질 수 있다(저장소 노출 0 — `git ls-files | grep -cE '^(a|b)/'` = 0).
  `mnemonicPrefix` 는 워킹트리 대상에서 `c/`·`w/` 를 내서 게이트가 아예 못 돈다(fail-closed, 이전부터).
  → 7.5.26 으로 연다.
- **가드의 `timeout=`**: 값 단위로 7.5.15.
- **산문의 절대 인용 106 건**: 하네스가 세고 있다. 다음 로트가 줄인다.

## MEASURE · Pre-Edit Gate — task 7.5.27 (워킹트리를 다시 쓰는 알려진 문) (2026-09-23)

**무엇을 하기로 했나.** 사람이 7.5.25 에 대해 **선택지 2 — 싸게 막기**를 골랐다. 근본(판정하는 바이트를
git 의 투영 없이 직접 대조)은 7.5.25 에 그대로 남고, 이 로트는 **알려진 문**을 닫는다. 문을 하나씩 세는
것이 class 를 못 닫는다는 것은 7.5.23 에서 배웠다 — 그래서 이 로트는 자기가 class 를 닫는다고 **적지 않는다**.

**FLM 먼저.** 편집 전 `ast.before-7.5.27.json`(가드 · 판정 각 디렉터리, 대상 `6c2ce774`).

**문을 진짜 git 으로 셌다.** 편집은 디스크에 있고(`return 9`), 같은 길이로 바꿨다:

| 문 | 기본 `git diff --numstat` | 명령줄 고정 뒤 | 닫는 법 |
|---|---|---|---|
| `filter.<d>.clean` | **빈 글자** (감춤) | `1 1 x.go` | `-c filter.<d>.clean=` |
| `filter.<d>.process` (진짜 v2 프로토콜) | **빈 글자** (감춤) | `1 1 x.go` | `-c filter.<d>.process=` |
| 위 + `required=true` · 이름에 점(`a.b`) | 감춤 | `clean=` 만: **rc 128**(git 이 죽는다) · `required=false` 까지: `1 1 x.go` | 셋을 같이 · 키의 마지막 칸만 뗀다 |
| 거짓말하는 `core.fsmonitor` | **빈 글자** (감춤) | `1 1 x.go` | `-c core.fsmonitor=false` |
| `update-index --assume-unchanged` | **빈 글자** (감춤) | 설정이 아니라 안 닫힌다 | `ls-files -v` 의 소문자 → 해시 대조 |
| `update-index --skip-worktree` | **빈 글자** (감춤) | 안 닫힌다 | `ls-files -v` 의 `S` → 해시 대조 |
| `core.checkStat=minimal` + `trustctime=false` + mtime·크기 되돌림 | `1 1 x.go` — **재현 안 됨** (7.5.28 정정: **거짓** — mtime 을 과거로 안 돌린 픽스처였다. 과거 시각이면 **빈 글자**) | – | 고정 안 넣었다 (YAGNI) |

**내 측정 오류 하나.** 첫 전수에서 process 필터를 "기본: 빈 글자 → 감췄다" 로 읽었는데, 프로토콜을
안 지키는 셸 스크립트라 **git 이 rc 128 로 죽은 것**이었다 — 나는 표준출력만 보고 rc 를 안 봤다. 시험이
`check_output` 으로 그것을 **오류**로 드러냈고, git 의 long-running 필터 프로토콜(v2, pkt-line)을 말하는
필터를 따로 써서 다시 쟀다 — 진짜로 rc 0 으로 감춘다. 시험 픽스처도 그 필터다.
([[missing-tool-reports-clean]]: 빈 출력은 "없음" 이 아니라 "못 물었다" 일 수 있다.)

**sparse-checkout 이 경계를 정했다.** `skip-worktree` 는 sparse-checkout 의 정상 도구다. 플래그만 보고
거절하면 정상 사용자를 막는다. 그래서 거절은 **플래그 + 디스크에 파일이 있음 + 판정과 같은 고정으로
해시한 바이트가 인덱스 blob 과 다름** 셋이 겹칠 때만이다. 실측 픽스처: 감춘 편집 둘(`a`·`b`)만 해시가
다르고, 플래그만 붙은 `c` 와 파일이 없는 `d`(sparse 모양)는 인덱스와 같거나 해시할 것이 없다.

**거부할 정상 입력을 셌다 — 오늘의 상태와 거부 가능한 집합 둘 다** ([[fail-closed-must-name-what-it-rejects]]
· 7.5.24 의 교훈): 오늘 이 저장소에서 추적 `*.go` **1,762** 전부 플래그 없음(`H`), 필터 드라이버 **0**,
fsmonitor 없음, sparse 꺼짐 → 새로 거절·변경되는 것 **0**. 거부·변경 가능한 **집합**: `*.go` 에 **정상적인**
clean 필터를 쓰는 저장소(git-lfs 는 포인터 대 실제 내용으로 판정이 달라지고, git-crypt 는 base 가 암호문이라
파싱이 실패해 거절된다) · 드라이버 이름에 `=`·공백 → 거절 · 설정/인덱스 읽기 실패 → 결함.

## VERIFY — task 7.5.27 (2026-09-23)

**수리 둘.** (1) `_git_view_pins` 가 고정을 만들고 `changed_existing_functions` 가 **두** `git diff`(가드의
`--numstat` · 판정의 `--unified=0`)에 **같은** 고정을 펼친다 — 한쪽만 고정하면 두 시야가 갈려 교차 검사가
거절하므로 행동 시험이 잡고, 구조 시험도 두 argv 가 `*pins` 를 펼치는지 본다. (2) `_hidden_by_index_flags`
가 워킹트리 대상에서만, **가장 뒤에서** 감춘 편집을 이름 대고 거절한다.

**RED → GREEN.** `TheWorktreeIsNotRewrittenUnderTheGate` 여덟. 편집 전 **실패 5**(subTest 포함 6) ·
**통과 3** — 통과 셋은 통과해야 하는 셋이다(양성 대조군 · 플래그만 있고 감춘 것이 없는 경계 · 커밋 대상 경계).
편집 후 8/8.

**옛 mock 시험 셋이 엉뚱한 이유로 통과할 뻔했다.** `subprocess.run` 을 통째로 mock 하는 시험들이 새 git
호출(설정 읽기)에서 먼저 실패를 받았다. 복호화만 고쳤다면 `test_git_diff_failure_is_not_treated_as_empty_change`
는 **판정 diff 가 아니라 설정 읽기의 실패로** 초록이 됐을 것이다 — 이름과 다른 이유다. 새 헬퍼 둘을 고정해서
mock 이 판정 diff 에만 닿게 했다 ([[passing-test-is-not-evidence]]: 실패 계약이 바뀌면 통과 **이유**를 센다).

**실물 게이트 영수증** (`analysis/harness/7527_switch.py`). 격리 worktree 에 `.gitattributes` 만 두고
필터 설정은 **환경 변수**(`GIT_CONFIG_COUNT`)로만 줬다 — 공유 `.git/config` 에 아무것도 안 쓴다(7.5.24 에서
그렇게 남긴 킬 스위치 절반의 교훈). 끝난 뒤 공유 설정과 프로브 워킹트리가 깨끗한 것을 확인했다.

| a112 | 편집 전 `6c2ce774` | 편집 후 |
|---|---|---|
| 평소 | 판정 줄 **36** · required **64** | 같다 |
| 추적 안 된 `.gitattributes` 한 줄 + clean 필터 | 판정 줄 **0** · required **1** · **`evidence complete`** | 판정 줄 **36** · required **64** — 거절이 아니라 **올바른 판정** |

**변이.** `AF1~AF11` 을 더했다(총 **143**, 앵커 전수 정확히 1회). 첫 판 `113:143` 은 **29/30 CAUGHT** —
`AF2`(드라이버의 `clean=` 고정을 뺀다)가 **살아남았고** 계측기가 "도달함" 이라 답했다. 원인을 재 보니
**git 2.43 에서는 `process=` 하나만 비워도 clean 필터가 꺼진다** — 빈 `process` 값은 **아무것도 띄우지 않고**
clean 명령을 끄는 **우연**이다(`고정 없음: 빈 글자` · `process= 만: 1 1 x.go` 실측). (7.5.28 정정: 앞 판본은 "빈
process 가 시작에 실패하고 `required=false` 라 통과" 라 적었는데, `GIT_TRACE` 로 보면 빈 값에서는 `run_command`
가 **0 줄**이고 `/nonexistent` 에서는 세 번 시도한다 — 실패가 아니라 꺼짐이다.)
그러니 `clean=` 은 이 git 에서 우연히 죽은 코드다. 동등 변이라고 **적지 않았고**, 그 우연에 기대지도 않았다 —
판본마다 다를 수 있다. `clean=` 을 남기고 **함수의 반환값을 직접 재는 계약 시험**을 더했다
(`test_every_configured_driver_gets_all_three_pins` — 드라이버마다 셋이 다 있는지, 점 든 이름이 안 잘리는지 ·
`test_a_driver_name_that_cannot_be_pinned_is_refused`). 세 고정 각각을 빼는 변이가 이제 전부 빨개진다 — `AF` 구간 재실행 **11/11 CAUGHT · 생존 0**(대조군 양끝 GREEN `Ran 339`), 첫 판 `113:143` 의 AF 밖 **19** 는 전부 CAUGHT (7.5.28 정정: "`113:130` 19/19" 라 적었는데 그 창에는 17 개가 있다 — AF 가 130 에 끼어 AD6·AC5 가 밀렸다)
([[surviving-mutant-may-mean-accidental-safety]]: 아래 층이 우연히 막고 있으면 의존을 직접 못 박는다).

**A/B.** 실물 `main()` 출력 전체 **126/126 SAME · DIFFERENT 0**(기준 `6c2ce774`). 이번엔 **호출이 늘었다** —
git **20,365 → 20,596 (+231, +1.1%)**. 판정마다 설정 읽기(`git config`) 한 번, 워킹트리 대상이면 인덱스 조회
(`git ls-files`) 한 번이 붙는 설계 그대로다(플래그 붙은 파일이 없으면 해시 프로세스는 안 뜬다). `ast.json`
읽기는 **9,072 → 9,072**. 벽시계는 순서를 번갈아 983.9 → 959.0s · 967.0 → 988.4s — 두 순서가 반대라 잡음이다.

**게이트.** `make lint` rc **0** · `make sdd-test` rc **0**(logic-map 404 → **414**) · `make test-seams` rc **0**
(**108 패키지**) · `openspec validate --all --strict` **58/58** · a122 게이트 **rc 0** · 좌표 게이트 **rc 0**
(번들 341 · 산문의 수 불일치 0).

**시험 329 → 339.**

### 안 한 것 — 사유

- **7.5.25(class)**: 사람이 싼 막기를 골랐다. 아직 모르는 워킹트리 재작성 문은 여전히 두 시야를 함께
  속일 수 있다 — 판정 바이트를 git 의 투영 없이 직접 대조하는 것이 근본이고 사람 결정으로 남는다.
- **`core.checkStat`·`trustctime` 고정**: 우회가 **재현되지 않았다**(git 이 잡았다). 재현 안 된 문을 닫는
  코드는 안 넣었다. (7.5.28 정정: **거짓이었다** — 픽스처가 mtime 을 과거로 안 돌려 racy-git 창에 걸려 있었다. 과거 시각으로 두면 재현되고, 7.5.28 이 고정을 넣었다)
- **정상적인 clean 필터(git-lfs · git-crypt)**: 중화 때문에 그런 저장소의 판정이 raw 바이트로 바뀐다.
  오늘 노출 0 이고, 막을 방법(정상 필터와 악의 필터를 가르는 것)이 없다 — README 에 적었다.
- **7.5.26 · 7.5.15**: 그대로.

## 정정 · MEASURE · VERIFY — task 7.5.28 (7.5.24·7.5.27 재리뷰, 2026-09-24)

**적대 리뷰어가 넷을 냈고 넷 다 부모 커밋과 A/B 로 재현했다.**

| # | 무엇 | 7.5.23 `de5a78b4` (7.5.25 정정 F4: 앞 판본은 "부모" — 부모는 `4de48502` 다) | `4de48502` | 누구의 것 |
|---|---|---|---|---|
| 1 | 경로에 U+2028 · U+2029 · U+0085 (`a<구분자>@@ -1 +1 @@.go`) | **거절** | **`required []`** | **내 회귀**(7.5.24 의 순서 짝짓기) |
| 2 | `core.checkStat=minimal` + 같은 크기 편집 + mtime 을 **과거로** | `required []` | `required []` | **내 거짓**(7.5.27 이 "재현 안 됨" 이라 적었다) |
| 3 | 추적 안 된 `.gitattributes` 의 `*.go ident` + `$Id: …$` 안의 논리 | `required []` | `required []` | 새 문 — 내장 속성이라 `-c` 로 못 끈다 |
| 4 | 이름에 **공백**이 든 평범한 `*.go` 편집 | **거짓 차단** | **거짓 차단** | 이 change **이전부터** — 1을 재현하다 드러났다 |

**1의 기전.** 판정 diff 를 `text=True` 로 받아 `str.splitlines()` 로 잘랐다. 그것은 `\n` 말고도 U+2028 등에서
자르고, `core.quotePath=false` 라 git 은 그 글자를 인용하지 않으며 이름 가드는 `\n\r\t` 만 거절한다. 그래서
`diff --git` 머리가 잘려 `@@ -1 +1 @@.go b/a…` 조각이 **훅**으로 읽히고, 본문이 이름(`---`/`+++`)보다 먼저
열려 이름이 빈 채 `flush()` 가 훅을 버렸다. 구역에는 본문이 있다고 표시되므로 순서 짝짓기의 크기·본문 검사가
통과한다. 이름으로 짝짓던 7.5.23 은 이 입력을 거절했다 — **내가 회귀를 넣었다**.

**4는 1을 고치다 드러났다.** 줄 경계를 고치자 파서가 이름을 제대로 읽었는데 `base_file` 이 그 이름을 못 열었다.
원인은 git 이 이름에 **공백이 있으면 `---`·`+++` 줄 끝에 탭을 붙이는** 관례였다(`… @@.go\t`). 공백 든 `*.go`
는 전부 `cannot load existing base file` 로 거짓 차단되고 있었다 — 부모에서도 같다. 저장소 노출 0.

**2는 내 측정이 틀렸다.** 7.5.27 의 픽스처는 편집한 파일의 mtime 을 과거로 안 돌려 git 의 racy 검사 창에
걸려 있었고, 그래서 git 이 파일을 다시 읽었다. 과거 시각으로 두면 `checkStat=minimal` 이 편집을 감춘다.
"재현 안 됨" 을 적은 네 자리에 정정을 붙였다.

**수리.** (1) 판정 diff 를 **바이트로** 받아 `\n` 에서만 자른다(해독은 엄격 — 비 UTF-8 은 전처럼 결함).
(2) `core.checkStat=default` · `core.trustctime=true` 고정. (3) `_ident_go_paths` — 워킹트리 대상에서 추적
`*.go` 에 `ident` 가 켜져 있으면 이름 대고 거절(가장 뒤, 인덱스 플래그보다 앞). (4) `_header_name` — 머리 줄
끝의 탭 하나를 뗀다(이름 가드가 탭 든 이름을 거절하므로 모호하지 않다). 덤: 공백 든 드라이버 이름 거절은
**헛거절**이었다(`-c 'filter.my drv.clean='` 은 적힌다, 명령줄 값이 이긴다) — `=` 만 거절한다.

**같은 값의 다른 사본을 셌다** ([[correction-unit-must-be-the-value]]). `str.splitlines()` 로 git 출력을 자르는
자리는 이 파일에 하나 더 있다 — `_self_repair_commits` 의 `block.splitlines()`. 다른 판정(착지 뒤 자기 수리)이고
FLM 없이 고치면 절차 위반이라 **task 7.5.29** 로 열었다. 그 결과가 판정을 바꾸는지는 **안 쟀다**.

**RED → GREEN.** 새 시험 **아홉**. 처음 일곱(subTest 포함 아홉)의 편집 전 결과는 실패 6 · 오류 1 이고, 통과 둘은
**이미 있던 행동을 못 박는** 시험이다(플래그 해시의 고정 · `100755` — 재리뷰가 그 둘을 빼는 변이가 시험 전부를
통과한다고 보였다). 나머지 둘은 로트 도중에 더했다: 공백 이름(수리하다 드러난 옛 결함, 편집 전 **오류** — 7.5.25 정정 F3: 앞 판본은 "실패" 라 적었는데 바로 앞 문장이 둘을 가른다) · 끈
`-ident`(경계, `AG4` 가 살아남은 뒤).
판정 diff 가 바이트가 되면서 문자열을 돌려주던 mock 시험 일곱(7.5.25 정정: 깨진 **경우** 일곱 · 시험 **함수** 여섯 — 하나가 subTest 둘)이 깨졌고, 계약(바이트)에 맞게 고쳤다 —
그중 셋은 새 헬퍼(`_ident_go_paths`)도 고정해 mock 이 판정 diff 에만 닿게 했다.

**흔들린 시험 하나 — 그것이 새 측정을 낳았다.** `checkStat` 시험이 혼자서는 6/6 통과하는데 부하 아래 12 판
중 3 판 실패했다. 픽스처가 파일을 `unlink` 뒤에 다시 써서 inode 가 **재사용**되고, 같은 초 안이면 ctime 도 같아
**고정한 기본 stat 검사**도 속은 것이다. 설정 없이 따로 재니 **20 판 중 17 판** 편집이 감춰졌고 17 판 전부 inode
재사용 판이었다. 그러니 `checkStat` 고정은 **설정 문**만 닫고, stat 캐시를 믿는 것 자체는 못 닫는다 → 7.5.25 에
측정과 함께 적었다(임시 인덱스 사본에 `--really-refresh` 를 거는 선택지도). 시험은 새 파일을 만들어 **옮겨 놓는**
방식으로 고쳐 inode 가 반드시 다르게 했다 — 설정 문만 결정적으로 잰다(동시 12 판 전부 통과 · 고정을 빼면 6 판
전부 실패).

**변이.** `AG1~AG9` 를 더했다(총 **152**, 앵커 전부 1 회; 이 로트가 옮긴 `AF1`·`AF7`·`AF11` 은 다시 겨눴고 `AG1` 은
하네스 안의 `"\n"` 이 진짜 줄바꿈으로 해석돼 한 번 빗나갔다). 첫 판 `113:152` 는 **38/39** — `AG4`(명시적으로 끈
`-ident` 를 켠 것으로 셈)가 살아남았고 계측기는 **"못 쟀다"**(표식 줄이 문장이 아니다)라고 답했다. 거절하면 안 되는
정상 입력이라 경계 시험을 더했고, 흔들리던 `checkStat` 픽스처를 고친 **뒤에** `AG` 구간을 다시 돌려 **9/9 CAUGHT ·
생존 0**(대조군 양끝 GREEN `Ran 348`)이다.

**게이트.** `make lint` rc **0** · `make sdd-test` rc **0**(logic-map 414 → **423**) · `make test-seams` rc **0**
(**108 패키지**) · a122 게이트 **rc 0** · a112 `required 64` 그대로.

**A/B.** 실물 `main()` 출력 전체 **126/126 SAME · DIFFERENT 0**(기준 `4de48502`). git 호출은 **20,596 → 20,826
(+230, +1.1%)** — 워킹트리 대상 판정마다 `ident` 조회(`git ls-files` + `git check-attr`) 두 프로세스가 붙는다.
`ast.json` 읽기 **9,072 → 9,072**. 벽시계 948.7 → 947.1s · 951.0 → 950.9s.

**시험 339 → 348.**

### 정정 — 7.5.24·7.5.27 주장정확성 재리뷰 (같은 실수 **다섯 번째**)

리뷰어가 커밋된 트리에서 잰 것이고 전부 내 손으로 다시 확인했다(가드 분기 수 · `GIT_TRACE` 는 직접 쟀다).

| # | 내가 쓴 것 | 실제 |
|---|---|---|
| R1 | 가드 FLM "분기 수는 그대로다" | **14 → 15** — `*(pins or [])` 의 `or` 가 늘었다 |
| R2 | BTM "두 시야의 크기" 행이 `…vanished_body…` · `…textconv…` 로 못 박힌다 | 검사를 지워도 **둘 다 초록**. 못 박는 것은 `…lists_fewer_files…` **하나** — 행을 AE2 시험을 더하기 **전**에 썼다 |
| R3 | BTM·FLM: 필터 고정을 행동 시험 셋이 못 박는다 · 시험 "넷" | `clean=` 을 빼도 행동 시험은 **초록**(git 의 우연). 못 박는 것은 **계약 시험** — 그리고 여섯이다 |
| R4 | "변이 `113:130` **19/19**" | 그 창에는 **17** 개다. AF 가 130 에 끼어 AD6·AC5 가 밀렸다. 19 는 첫 판 `113:143` 의 AF 밖 개수였다 |
| F6 | "빈 process 가 **시작에 실패**하고 `required=false` 라 통과" | `GIT_TRACE`: 빈 값에서 `run_command` **0 줄**, `/nonexistent` 에서 3 줄. 실패가 아니라 **꺼짐**이다 |
| F7 | `7523_coords.py` 주석 "산문 105건" | 106. 7.5.24 가 고쳤다고 한 **바로 그 값**의 사본이 하네스 안에 남았다 |
| F8 | 7.5.24 tasks "경계 **둘**" | review·커밋은 "셋". 크기 시험을 나중에 더하고 tasks 를 안 고쳤다 |

(F5 — `checkStat` "재현 안 됨" — 는 적대 리뷰도 냈고 위 표의 2번으로 고쳤다.)

**R1~R4 는 넷 다 같은 모양이다 — 같은 커밋의 **뒤 편집**이 앞에서 쓴 산문을 낡게 했다.** 좌표 게이트는
지문·범위·열린 task 의 좌표·자기 인용 수를 세지만, **FLM 의 분기 수 · BTM 행이 대는 시험 · 변이 창 이름**은
안 본다. 이번 로트는 도구를 더 만들지 않고 **순서**를 바꿨다: 산문(FLM·BTM·review·tasks)은 마지막 변이 판과
`ast.after` 추출 **뒤에** 쓰거나 다시 대조한다. BTM 의 "무엇이 못 박나" 칸은 변이 로그의 "빨개진 시험" 에서
옮겨 적는다 — 추측으로 쓰지 않는다.

## 정정 · MEASURE · VERIFY — task 7.5.25 (class 수리: git 의 워킹트리 투영을 안 쓴다, 2026-09-24)

**cfe8d08b 재리뷰 둘의 결과와 처리.**

| 출처 | 발견 | 처리 |
|---|---|---|
| 적대 F1 | `working-tree-encoding=UTF-7` 이 `ident` 처럼 편집을 두 시야에서 감춘다(내가 재현: `git status` 는 `M`, 게이트 `[]`) | **class 로 닫음** — 시험 `test_working_tree_encoding_does_not_hide_a_worktree_edit` |
| 적대 | `ident` 코드의 변이 아홉이 시험 전부를 통과(가장 무거운 것: "첫 파일만") | 코드가 **없어졌다**(`_ident_go_paths` 삭제) |
| 적대 | `trustctime` 고정을 빼도 아무 시험이 모른다 · 그 고정은 문 하나를 막는다 | 고정이 **필요 없어졌다** — stat 캐시를 안 쓴다. 설정 없는 stat 문도 닫힘(`test_the_stat_cache_hides_nothing_without_any_config`) |
| 적대 | `ident=foo` 처럼 git 이 적용 안 하는 값도 거절(헛거절) | 거절이 없어졌다 |
| 적대 | `diff.noprefix` · `GIT_LITERAL_PATHSPECS` 는 요구를 지우고 `diff.mnemonicPrefix` · `color.ui=always` · `"` 든 이름은 거짓 차단 | 형식·pathspec 의 문제라 스냅숏이 안 닫는다 → **7.5.26** 에 실었다 |
| 주장정확성 F1 | BTM·FLM: stat 시험이 `checkStat`·`trustctime` 둘을 못 박는다 | **거짓** — `checkStat` 만. `AG2` 가 둘을 한꺼번에 빼서 기록이 뭉쳐 있었다. 제자리 정정 |
| 주장정확성 F2 | FLM "`clean=` 을 못 박는 것은 계약 시험 **하나뿐**" | **둘** — 같은 커밋에서 뒤에 더한 공백 드라이버 시험이 문장을 낡게 했다. **같은 실수 여섯 번째.** 제자리 정정 |
| 주장정확성 F3 | review "공백 이름 … 편집 전 **실패**" | **오류**였다. 제자리 정정 |
| 주장정확성 F4 | review 표 머리 "부모 `de5a78b4`" | 부모는 `4de48502`; 그 열은 7.5.23. 제자리 정정 |
| 주장정확성 | "mock 시험 일곱" | 경우 일곱 · 함수 여섯. 제자리 정정 |
| 주장정확성(부수 효과) | 리뷰어가 디스크 0 을 두 번 만나 공유 go 캐시 약 110 GB 를 지웠다 | 원인을 이 로트에서 쟀다 → **7.5.30** |

**왜 문을 닫지 않고 class 를 닫았나.** 필터 → fsmonitor → 인덱스 플래그 → stat 캐시 → `ident` → `working-tree-encoding`.
로트마다 리뷰가 git 이 워킹트리를 **비교 전에 다시 쓰거나 안 읽는** 새 방법을 찾았다. 목록을 세는 한 끝나지 않으므로
게이트가 비교할 바이트를 **스스로 읽는다**. 사람이 이 설계 변경(선택지 1)을 골랐다.

**MEASURE — 바꾸기 전에 잰 것.** (1) 설정이 하나도 없는 stat 문: 같은 파일에 같은 크기로 쓰고 mtime 을 되돌리되 ctime 을
인덱스 기록과 같은 초에 두면 **10/10** 감춰진다 — 초 경계 바로 뒤에서 시작하는 결정적 픽스처. (2) 원형(임시 인덱스 +
임시 객체 저장소 + `--cached`)을 /tmp 저장소 일곱 모양에 돌렸다 — 편집만 · clean · assume · skip · stat · UTF-7 은 전부
`[('p.go','F')]`, FIFO 는 이름 댄 거절, 실제 `.git/objects` 목록과 `.git/index` 바이트는 그대로. (3) `update-index` 가
`post-index-change` 훅을 **띄운다**(실측) · git 2.43 에서 빈 blob 은 가상 객체가 **아니다**(새 저장소에서 `cat-file -e` rc 1;
`add -N` 이 써 둔다) — 그래서 "빈 파일이면 늘 쓴다" 갈래를 넣었다가 뺐다(손으로 만든 인덱스만 걸리고, 걸려도 크게 실패한다).
(4) 저장소 노출: 추적 `*.go` 1,762 개 전부 모드 `100644` · 플래그 없음 · `core.autocrlf` 미설정 · 필터 0 · 속성 파일 0 ·
`\r` 든 것 0.

**RED → GREEN.** 새 시험 파일을 부모 코드에 붙이면 `TheWorktreeIsNotRewrittenUnderTheGate` · `TheGuardAndTheJudgementReadTheSameDiff`
에서 **22 경우(함수 17)** 가 빨갛다 — 실패 9 · 오류 8(오류는 부모가 이름 대고 거절하던 모양 · 새 헬퍼가 없는 모양). 새 문 둘
(UTF-7 · 설정 없는 stat)은 **실패**, 즉 부모가 요구를 비운다. 편집 후 **357** 전부 초록. 지운 시험 둘: 필터 고정 계약 ·
`-ident` 경계(겨눌 코드가 없다). 거절이나 고정 목록을 단언하던 시험 **여섯**(`=` 드라이버 · `ident` · 필터 뒤 플래그 · 실행 파일 플래그 ·
공백 드라이버 · 인덱스 플래그)은 요구가 **서는지**를 단언하도록 바꿨다.

**변이.** 총 **160**(겨눌 코드가 없어진 18 개를 빼고 `AH1~AH26` 을 더했다; 앵커 전부 1 회). 창 네 개를 병렬로:
`0:40` · `40:80` · `80:120` 각 **40/40**, `120:160` **37/40**. 생존 셋을 **동등이라 적지 않았다**
([[surviving-mutant-may-mean-accidental-safety]]): `AH16`(blob id 대신 날 해시) · `AH17`(모든 blob 을 씀)은 git 이 내용을
견주어 훅을 안 내서 판정이 그대로였을 뿐, 안 바뀐 1,762 파일이 "바뀐 파일" 로 오르고 판마다 18 MB 를 쓴다 →
`test_an_untouched_worktree_is_an_empty_comparison`(목록 0 · 객체 0, 한 파일 편집이면 1 · 1). `AH19`(커밋 대상도 스냅숏)는
계측기가 "못 쟀다" — 판정은 같고 커밋 대상 판정이 워킹트리를 **원장에 남긴다** → 커밋 대상 시험에 원장 단언. 다시 돌려
`147:151` **4/4**. `AH7`(`read_bytes()` 로 깔때기를 비켜 감)은 FIFO 시험에서 스위트가 **멈췄다**(`wait_for_partner`) — 하네스가
`Ran -1` 을 "환경 의심" 으로 표시했고 CAUGHT 로 세지 않았다. 사본에서 따로 재니 깔때기 시험 둘이 잡고 FIFO 시험은 멈춘다.
하네스에는 멈추지 않는 판본(원장만 비켜 가는 `_opened_bytes`)을 넣어 `138:139` **1/1**. 모든 창 대조군 양끝 GREEN.
변이마다 **값 하나**만 겨눴다 — 고정 셋(`splitIndex` · `hooksPath` · fsmonitor)과 `ls-files` 의 고정은 각자 변이가 있다
(7.5.28 `AG2` 의 교훈).

**측정 사고.** 변이 판을 `nohup &` 로 띄우고 `ps | grep` 이 0 줄을 내서(rtk 가 걸렀다) 판을 다섯 더 띄웠다 — **아홉 판**이
겹쳤고, 스위트가 임시 저장소마다 `go run` 해서 공유 go 캐시가 20 분 새 188 → 203 GB, `/` 는 8 GB 남짓이 됐다. 판을 전부
죽이고 결과를 버렸다. 전용 캐시로 재니 스위트 한 판이 **+375 MB**(데운 뒤), `-trimpath` 면 **+1 MB** — task 7.5.30.
하네스가 이제 스스로 `-trimpath` + 전용 캐시(`_work/gocache`)로 돈다. 공유 캐시의 15 GB 는 지우지 않았다(사람의 것과 섞여 있다).

**게이트.** `make lint` rc 0 · `make sdd-test` rc 0(logic-map 423 → **432**)(이 판은 `GOFLAGS=-trimpath` · 전용 `GOCACHE` 로 돌렸다 — 7.5.30 의 증식을 피하려고. `-trimpath` 는 실행 파일의 소스 경로 표기만 바꾼다) · `make test-seams` rc 0(**108 패키지**) · `openspec validate --all --strict`
**58/58** · a122 게이트 **rc 0**(a112 `required 64` 그대로) · 좌표 게이트 **rc 0**(머리말 범위 불일치 1 줄은 7.5.2.2 뒤로 안 건드린 `normalized_source` 의 기록 — 판정 밖).

**A/B.** 실물 `main()` 출력 전체 **126/126 SAME · DIFFERENT 0**(기준 `cfe8d08b`) — git 호출 20,826 → **20,710(−116)**, `ast.json` 읽기 9,072 → 9,072, 벽시계 1,055.5 → 1,080.3s · 1,053.8 → 1,090.7s(+2.4~3.5%, 워킹트리 대상마다 추적 `*.go` 전수를 읽고 해시하고 끝의 재확인이 다시 읽는 몫). git 호출은 change 마다 **정확히 1** 줄었거나(116) 그대로다(10 — 이 함수에 닿기 전에 멈추는 change). 코드로 세면 워킹트리 대상 6 → 5(`git config` · 가드 · 판정 · `ls-files` 둘 · `check-attr` → `rev-parse` · `ls-files` · `update-index` · 가드 · 판정), 커밋 대상 3 → 2(`git config` 이 빠졌다). 파일 `_work/7521_main_ab_done.cfe8d08bf1f4.*.json`.

### 안 한 것 — 사유
- **7.5.26**(출력 형식 · pathspec 문): 워킹트리 바이트가 아니라 파서 문제다. 이 로트에서 같이 고치면 판정 diff 의 명령줄이 두
  이유로 동시에 바뀌어 A/B 가 어느 쪽 변화인지 못 가른다.
- **skip-worktree 뒤의 삭제**: sparse-checkout 과 구별할 신호가 인덱스에 없다(7.5.27 도 같았다). 잔여로 적었다.
- **7.5.30**(게이트의 `go run -trimpath`): `go_functions` 내부 변경이라 FLM 이 먼저다.


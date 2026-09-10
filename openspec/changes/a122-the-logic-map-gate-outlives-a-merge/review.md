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

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

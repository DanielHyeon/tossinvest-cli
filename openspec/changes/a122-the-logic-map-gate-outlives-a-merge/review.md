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

# Function Logic Map: `compute_landing`

`tools/logic-map/check_analysis.py:972-1002`(worktree) · Python · **task 6.1.2**

> **손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.worktree.json` 을
> `enumerate.py` 가 기계로 열거했고 아래는 그 열거를 옮긴 것이다. 이 함수는 6.1.2 가
> 새로 만든 것이라 편집 전 판본이 없다 — `ast.json`(HEAD) 이 없는 이유다.

## Inputs and invariants

게이트가 기록할 착지. **저자가 고르지 않는다.** 바닥 이후이면서 고정을 통과하는
**가장 낮은** 커밋. 가장 낮은 것을 고르는 이유는 창을 좁히는 것이 이 기능의 목적
이기 때문이고, 그것이 안전한 이유는 바닥 아래로 못 내려가기 때문이다.

## Branches and early returns — 열거 그대로 (분기 8)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 982 | If | `if not _pinning_bundles(root, analysis):` |
| B2 | 985 | If | `if not floor:` |
| B3 | 987 | IfExp | `floor if _is_ancestor(root, base, floor) else base` |
| B4 | 992 | If | `if process.returncode:` |
| B5 | 994 | For | `for candidate in [start, *process.stdout.split()]:` |
| B6 | 995 | BoolOp | `not _is_ancestor(root, base, candidate) or not _is_ancestor(root, floor, candidate)` |
| B7 | 995 | If | `if not _is_ancestor(root, base, candidate) or not _is_ancestor(root, floor, candidate):` |
| B8 | 997 | If | `if not _pinning_at(root, candidate, analysis)[1]:` |

반환 5개: `983` · `986` · `993` · `998` · `999`

## Calls and live bindings

`_pinning_bundles` · `_evidence_floor` · `_is_ancestor` · `git rev-list --reverse`
· `_pinning_at`.

## State mutations and fallbacks

없다. **아무것도 쓰지 않는다** — 쓰는 것은 `record_landing` 이다.

## Safety conclusion

값을 못 정하면 `("", 사유)` 를 돌려준다. 사유는 네 가지로 갈린다(번들 0 · 바닥
없음 · 역사 순회 실패 · 맞는 커밋 없음). 실측: 활성 13건 중 11건이 값을 얻고
a089·a095 가 마지막 사유로 거부된다 — 그 둘의 증거가 자기를 담은 커밋에서 이미
틀렸기 때문이다(6.5).

---

## 공백 기록 — 7.2.1 · 7.3.1 · 7.4 는 이 함수를 FLM 없이 바꿨다 (2026-09-13, 7.6 이 적음)

이 세 task 는 내부를 바꾸면서 열거도 `not-applicable` 사유도 남기지 않았다. 7.6 이 7.2.1 직전
(`eaf536d2`, 7.1 과 소스 동일)과 HEAD(`2b5b05c1`)를 같은 열거기로 뽑아 `ast.before-7.2.1.json` ·
`ast.before-7.6.json` 으로 남긴다. 판정 근거(뮤테이션·A/B)는 각 task 의 VERIFY 절에 있다.

| | 종류 | 소스 |
|---|---|---|
| 사라짐 | 분기 | `if not _pinning_at(root, candidate, analysis)[1]:` |
| 생김 | 분기 | `if _pinning_at(root, candidate, analysis)[1]:` |
| 생김 | 분기 | `if not unheld:` |
| 생김 | 분기 | `if unheld:` |

---

## 편집 — task 7.6 (규칙 한 집) · 분기 10 → 8 · 반환 6 → 6 · raise 0 → 0

편집 전 `ast.before-7.6.json`(HEAD `2b5b05c1`) · 편집 후 `ast.after-7.6.json`(워킹트리). 아래 표는 두 열거의
`source` 를 스크립트가 줄 단위로 대조한 것이다.

후보마다 들고 있던 규칙 사본(`base ∧ 하한 순서` · `불일치` · `미보유`)을 `_landing_refusal` 호출 하나로 바꿨다. 순회 전 조기 종료 둘(고정 번들 0 · 하한 없음)은 **계산 경로의 사유 문장** 때문에 남는다 — 규칙 안에도 같은 조건이 있어서 그 둘을 지워도 받는 후보는 안 바뀌고 사유 문장만 바뀐다(7.8 의 M7 시험이 그 문장을 잰다). 곁가지 후보에서 번들 대조가 +328회(93건 합계) 느는 비용은 review.md `## Pre-Edit Gate — task 7.6` 에 있다.

| | 종류 | 소스 |
|---|---|---|
| 생김 | 분기 | `if not refusal:` |
| 생김 | 분기 | `if names:` |
| 사라짐 | 분기 | `not _is_ancestor(root, base, candidate) or not _is_ancestor(root, floor, candidate)` |
| 사라짐 | 분기 | `if not _is_ancestor(root, base, candidate) or not _is_ancestor(root, floor, candidate):` |
| 사라짐 | 분기 | `if _pinning_at(root, candidate, analysis)[1]:` |
| 사라짐 | 분기 | `if not unheld:` |

## 편집 — task 7.7 (조언이 기록 명령의 판정에 묻는다) · 분기 8 → 7 · 반환 6 → 5 · raise 0 → 0

편집 전 `ast.before-7.7.json`(HEAD `3da639a9` blob) · 편집 후 `ast.after-7.7.json`(워킹트리, L1162-1199). 아래 표는 두 열거의 `source` 를 스크립트가 줄 단위로 대조한 것이다.

순회 전 조기 종료 둘(고정 번들 0 · 하한 없음)을 `_walk_floor` 로 옮겼다 — `_recording_refusal` 이 같은 두 사유를 **같은 함수**로 묻게 하려는 것이다. 사유 문장은 글자 그대로다. 기록 명령은 이제 판정 함수에서 먼저 멈추므로 그 경로로는 여기 `if why:` 가 안 닿는다 — 그래서 `test_the_computation_names_what_stops_it_before_the_walk` 가 이 함수를 **직접** 부른다(변이 R16 은 그 시험 하나만 잡는다).

| | 종류 | 소스 |
|---|---|---|
| 사라짐 | 분기 | `if not _pinning_bundles(root, analysis):` |
| 사라짐 | 분기 | `if not floor:` |
| 생김 | 분기 | `if why:` |
| 사라짐 | 반환 | `('', 'no `revision: current` evidence pins a landing for this change')` |
| 사라짐 | 반환 | `('', 'the pinning evidence never entered this history (commit the bundles)')` |
| 생김 | 반환 | `('', why)` |

호출 — 사라짐 ['_evidence_floor', '_pinning_bundles'] · 생김 ['_walk_floor']

## 편집 — task 6.5 (걷기 실패 사유는 걸은 것만 말한다) · 분기 7 → 8 · 반환 5 → 5 · raise 0 → 0

편집 전 `ast.before-6.5.json`(HEAD `483985dc` blob, L1162-1199) · 편집 후 `ast.after-6.5.json`(워킹트리, L1162-1207). 아래 표는 두 열거의 `source` 를 스크립트가 줄 단위로 대조한 것이다. 같은 파일의 다른 함수는 AST 덤프가 HEAD 와 같다(바뀐 함수 집합 = `compute_landing` 하나, 실측).

마지막 반환의 꼬리 "the evidence does not describe any revision on this history" 를 "at the first commit walked, {first}" 로 바꿨다. `first` 는 순회가 받은 **첫** 거절 문장이다(`first = first or refusal` — 열거기가 `BoolOp` 를 분기로 세서 분기가 하나 는다). 옛 꼬리는 순회가 걷지 않는 하한 아래까지 주장했고, 걷기 실패 17건 전수에서 4건(a089 · a095 · console-click-approval · verify-us-market)은 증거가 하한 아래 커밋을 전부 맞게 기술해 실측으로 거짓이었다. 첫 후보인 이유: 17건 전부 첫 후보가 하한 자신이고(`start_is_floor` 17/17), 거기서 이미 틀린 소스가 저자가 고칠 번들이다. 뒤 후보의 문장에는 이웃이 나중에 고친 파일이 붙는다. 새 호출 0 — 문장은 규칙 `_landing_refusal` 이 이미 만든 것을 인용한다([[two-judgements-cover-for-each-other]]: 사유 문장을 여기서 새로 지으면 규칙이 문장을 바꿀 때 이 사유만 옛 문장으로 남는다). `unheld` 갈래와 그 문장은 그대로다.

| | 종류 | 소스 |
|---|---|---|
| 생김 | 분기 | `first or refusal` |
| 사라짐 | 반환 | `('', f'no commit at or after the evidence ({floor[:12]}) matches every pinning bundle — the evidence does not describe any revision on this history')` |
| 생김 | 반환 | `('', f'no commit at or after the evidence ({floor[:12]}) matches every pinning bundle — at the first commit walked, {first}')` |

호출 — 사라짐 없음 · 생김 없음

# Function Logic Map: `main`

`tools/logic-map/check_analysis.py:817-840`(HEAD `848b9ba3`) → `:836-872`(편집 후) · Python

> **이 표는 손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.json`(HEAD)과
> `ast.worktree.json`(편집 후)을 `enumerate.py` 가 기계로 열거했다. Go 용
> `extract_go_ast.go` 는 Python 함수에 쓸 수 없으므로 a120 선례를 따른다.

## Inputs and invariants

`--change` 와 `--root` 를 받아 `check` 를 돌리고 결과를 출력하고 종료 코드를 준다.
불변식은 **판정과 설명이 같은 실행에서 나와야 한다**는 것이다 — 어떤 창으로 요구했는지가
없으면 요구된 함수 이름은 판정이 아니라 소음이다.

## Branches and early returns — 편집 전 (분기 4 · 반환 2)

| ID | 줄 | 소스 |
|---|---|---|
| B1 | 824 | `if errors:` → `825 for error in errors:` → **`827 return 1`** |
| B2 | 825 | `for error in errors:` |
| B3 | 828 | `if context.get('execution_baseline_adoption'):` |
| B4 | 835 | `if landing:` → `836 print(landed-commit … required N function(s))` |

## 결함 (task 3.3)

**`B1` 의 `827 return 1` 이 `B3`·`B4` 를 통째로 건너뛴다.** 착지·요구 수를 찍는
`B4` 는 **실패 경로에서 도달 불가**다. 그래서 task 1.9 가 적은 "항상 출력한다"는
성공할 때만 참이었고, 실측이 그것을 확인한다(2026-09-10, HEAD `848b9ba3`):

| change | 출력 줄 | 가장 긴 줄 | 창을 말하는 줄 |
|---|---|---|---|
| a074 | 324 (`missing evidence for modified function` **316**) | 183자 | **0** |
| a076 | **1** | **21,838자** | **0** |

3.2.4 와 **같은 모양**이다 — early return 이 뒤의 판정을 건너뛴다. 그때는 아카이브를
안 세게 했고, 여기서는 설명을 안 찍게 한다.

## 편집 — 분기 4 → 6, 반환 2 → 2

| ID | 줄 | 소스 |
|---|---|---|
| B1 | 848 | `if base:` → `850 print(base … → 대상 … required N function(s))` |
| B2 | 854 | `if not landing:` → `856 print(창에 커밋 K개가 더 있다)` |
| B3 | 858 | `landed_after or '?'` — 못 재면 숫자를 **지어내지 않는다** |
| B4 | 862 | `if errors:` → `865 return 1` |
| B5 | 863 | `for error in errors:` |
| B6 | 866 | `if context.get('execution_baseline_adoption'):` |

`print` 호출 줄이 `[850, 856, 864, 867, 869]` 이고 `return 1` 이 `865` 다. 즉 창 줄
둘은 **실패 반환보다 앞**에 있으므로 두 경로 모두에서 찍힌다 — 구조가 그것을 말한다.
반환은 **둘 그대로**(`865: 1`, `872: 0`)라 이 편집은 경로를 더하거나 지우지 않는다.

옛 `B4 if landing:` 는 없어졌다. 한 줄이 성공·실패 양쪽을 담당하므로 두 벌을 두지
않는다 — 두 벌이면 한쪽만 조용해지는 오늘의 상태로 되돌아간다.

## Calls and live bindings

`check` · `_target_text` · `_commits_after`(`git rev-list --count <base>..HEAD`) ·
`print` · `argparse`. 쓰기는 없다.

## State mutations and fallbacks

없다. `context` 는 `check` 가 채우고 여기서는 읽기만 한다. fallback 은 하나 —
커밋 수를 못 재면 `'?'` 를 찍는다(`B3`). **숫자를 지어내지 않는다.**

## Safety conclusion

판정을 한 줄도 바꾸지 않는다. 이 태스크의 불변식은 **126개 id 의 rc 가 하나도
안 바뀌는 것**이고 그것으로 잰다. 성공 줄
(`evidence complete or diff-proven exempt`)은 기록 스무 곳이 인용하므로 손대지 않았다.
Go 파일 변경 0줄.

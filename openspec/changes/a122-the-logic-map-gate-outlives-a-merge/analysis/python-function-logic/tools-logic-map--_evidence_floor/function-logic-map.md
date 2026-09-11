# Function Logic Map: `_evidence_floor`

`tools/logic-map/check_analysis.py:436-465`(worktree) · Python · **task 6.1.2**

> **손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.worktree.json` 을
> `enumerate.py` 가 기계로 열거했고 아래는 그 열거를 옮긴 것이다. 이 함수는 6.1.2 가
> 새로 만든 것이라 편집 전 판본이 없다 — `ast.json`(HEAD) 이 없는 이유다.

## Inputs and invariants

고정 번들이 역사에 들어온 **마지막** 커밋. 저자가 고를 수 없는 유일한 하한이다
— 오늘 만든 번들을 과거 커밋에 넣을 수 없기 때문이다. a122 6.1.1 이 활성 13건을
전수로 재서 13건 **전부** 고정 번들이 자기 base 뒤에 커밋됐음을 확인했다.

## Branches and early returns — 열거 그대로 (분기 6)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 450 | For | `for ast_path, _, _ in _pinning_bundles(root, analysis):` |
| B2 | 451 | Try | `try:` |
| B3 | 453 | ExceptHandler | `except ValueError:` |
| B4 | 457 | If | `if before:` |
| B5 | 459 | If | `if not paths:` |
| B6 | 465 | IfExp | `'' if process.returncode else process.stdout.strip()` |

반환 2개: `460` · `465`

## Calls and live bindings

`_pinning_bundles` · `git log -M --diff-filter=MA -1`.

## State mutations and fallbacks

없다.

## Safety conclusion

`-M --diff-filter=MA` 와 아카이브 전 경로를 같이 주는 것이 안전의 핵심이다.
rename 을 세면 아카이브하는 순간 이미 유효했던 기록이 무효가 된다 — a099 의 바닥
`21a315d1` 은 기록된 착지 `e6c4636a` 앞이고 rename 커밋은 뒤다. 못 찾으면 빈
문자열이고, 호출자가 그것을 거절로 다룬다(빈 하한을 통과로 다루면 구멍이 남는다).

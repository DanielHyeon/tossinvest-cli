# Function Logic Map: `_pinning_bundles`

`tools/logic-map/check_analysis.py:390-412`(worktree) · Python · **task 6.1.2**

> **손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.worktree.json` 을
> `enumerate.py` 가 기계로 열거했고 아래는 그 열거를 옮긴 것이다. 이 함수는 6.1.2 가
> 새로 만든 것이라 편집 전 판본이 없다 — `ast.json`(HEAD) 이 없는 이유다.

## Inputs and invariants

`analysis` 디렉터리에서 **착지를 고정하는 번들**만 고른다 — `revision:
`current` 이고 `file`·`source_sha256` 이 둘 다 있는 것. 이 선별이 **한 곳에만** 사는
것이 불변식이다. `resolve_landing` 과 `compute_landing` 이 각자 순회를 가지면 한쪽만
고쳐도 양쪽 시험이 초록이 된다 ([[two-judgements-cover-for-each-other]]).

## Branches and early returns — 열거 그대로 (분기 6)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 398 | For | `for ast_path in sorted(analysis.glob('*/ast.json')) if analysis.is_dir() else ():` |
| B2 | 398 | IfExp | `sorted(analysis.glob('*/ast.json')) if analysis.is_dir() else ()` |
| B3 | 400 | BoolOp | `not isinstance(value, dict) or value.get('revision', 'current') != 'current'` |
| B4 | 400 | If | `if not isinstance(value, dict) or value.get('revision', 'current') != 'current':` |
| B5 | 404 | BoolOp | `not raw_source or not digest` |
| B6 | 404 | If | `if not raw_source or not digest:` |

반환 1개: `412`

## Calls and live bindings

`_ast_value`(깨진 JSON 은 판정에서 뺀다) · `normalized_source`(절대경로 번들을
정규화한다 — 안 하면 정상 입력이 위조로 몰린다).

## State mutations and fallbacks

없다. 디스크를 읽기만 한다.

## Safety conclusion

빈 목록을 돌려줄 수 있고 그것은 오류가 아니다. **호출자가** 그 0 을 어떻게
다룰지 정한다 — `resolve_landing` 은 거절하고 `compute_landing` 은 기록을 거부한다.

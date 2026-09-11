# Function Logic Map: `_pre_archive_path`

`tools/logic-map/check_analysis.py:426-433`(worktree) · Python · **task 6.1.2**

> **손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.worktree.json` 을
> `enumerate.py` 가 기계로 열거했고 아래는 그 열거를 옮긴 것이다. 이 함수는 6.1.2 가
> 새로 만든 것이라 편집 전 판본이 없다 — `ast.json`(HEAD) 이 없는 이유다.

## Inputs and invariants

아카이브된 경로의 **옮기기 전** 이름. 아카이브가 아니면 빈 문자열이다.
`ARCHIVED_CHANGE` 정규식 하나에 의존하며, 그 `\\d` 가 유니코드 숫자를 먹는 것은
`gate.sh` 의 `[0-9]` 와 갈리는 기존 결함이다(6.4(c), 이 task 는 안 건드린다).

## Branches and early returns — 열거 그대로 (분기 3)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 429 | If | `if not relative.startswith(prefix):` |
| B2 | 433 | BoolOp | `match and tail` |
| B3 | 433 | IfExp | `f'openspec/changes/{match.group('change')}/{tail}' if match and tail else ''` |

반환 2개: `430` · `433`

## Calls and live bindings

`ARCHIVED_CHANGE.fullmatch`.

## State mutations and fallbacks

없다. 순수 문자열 변환.

## Safety conclusion

실패는 전부 빈 문자열이다. 빈 문자열이면 `_evidence_floor` 가 그 경로를 안
더하고, 그러면 바닥이 rename 커밋으로 올라갈 수 있다 — 그 경우를
`test_archiving_the_change_does_not_invalidate_its_record` 가 못 박는다.

# Function Logic Map: `_unheld_bundles`

`tools/logic-map/check_analysis.py:625-670` · Python · **task 7.2 → 7.5** ·
분기 **9 → 13** · 반환 1 · raise 0 (`ast.before-7.5.json` · `ast.after-7.5.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** `enumerate.py` 의 열거를 옮긴 것이다.

## Inputs and invariants

`candidate` 가 **들고 있지 않은** 번들. 하한은 번들의 **경로**를 보고 판정은 **워킹트리의
내용**을 읽으므로, 그 둘이 갈리는 자리를 이 함수가 막는다 (task 7.2, 리뷰 C3).

7.5 가 바꾼 것은 둘이다 — 입력이 `analysis: Path` 에서 `bundles` 로, 그리고 git 왕복이
번들마다 하나에서 **전체 한 번**으로. 받는 후보는 `_pinning_at` 과 이 함수를 **둘 다**
통과하므로 한쪽만 고치면 성공하는 walk 의 비용이 절반만 준다(2026-09-18 실측: a112 에서
spawn 47.5% · 47.5%).

## Branches and early returns — 열거 그대로 (after)

| id | 줄 | 종류 | 소스 | before 의 짝 |
|---|---|---|---|---|
| B1 | 648 | For | `for ast_path, _, _ in bundles:` | B1 (`… in _pinning_bundles(root, analysis)`) |
| B2 | 649 | Try | 경로 정규화 | B2 |
| B3 | 651 | ExceptHandler | `except ValueError:` — 저장소 밖 번들 | B3 |
| B4 | 655 | comprehension | `for _, relative, before in watched` | **7.5 신규** (배치 입력) |
| B5 | 655 | comprehension | `for path in (…)` | **7.5 신규** |
| B6 | 656 | IfExp | `(relative, before) if before else (relative,)` | **7.5 신규** |
| B7 | 659 | For | `for ast_path, relative, before in watched:` | **7.5 신규** (순회가 둘로 갈림) |
| B8 | 660 | Try | 디스크 읽기 | B4 |
| B9 | 662 | ExceptHandler | `except OSError:` | B5 |
| B10 | 666 | BoolOp | `committed is None and before` | B6·B7 이 **합쳐진 것** |
| B11 | 666 | If | 같은 줄의 `if` | B6·B7 |
| B12 | 668 | BoolOp | `committed is None or committed != judged` | B8 |
| B13 | 668 | If | 같은 줄의 `if` | B9 |

반환 1개(`670`): `sorted(set(unheld))`.

**B10·B11 은 before 의 중첩 둘을 하나로 합친 것이다.** `if committed is None:` 안의
`if before:` 는 `if committed is None and before:` 와 같은 집합을 고른다 — 순서도 같고
`before` 는 순수 문자열 유도라 side effect 가 없다.

## Calls and live bindings

`_committed_many` ×1 (before 는 `_committed_bytes` ×N~2N) · `_pre_archive_path` ·
`ast_path.read_bytes` · `sorted` · `set`. `_pinning_bundles` 는 더 이상 안 부른다.

## State mutations and fallbacks

없다. 디스크 읽기가 `OSError` 면 그 번들을 **안 들고 있는 것**으로 센다 — 7.2 가 정한
보수적 방향이고 7.5 가 안 바꿨다.

## Safety conclusion

읽는 순서가 바뀌었다: before 는 번들마다 (디스크 읽기 → git 읽기), after 는 (경로 다 모으기
→ git 한 번 → 번들마다 디스크 읽기). 디스크 읽기가 실패해도 git 을 이미 불렀다는 것이
유일한 차이이고 **답에는 안 들어간다**. 등식은 여전히 **워킹트리와** 세운다 — 그것이
갈아 끼운 자리를 보는 유일한 방법이기 때문이다(2026-09-12 K1 실측).

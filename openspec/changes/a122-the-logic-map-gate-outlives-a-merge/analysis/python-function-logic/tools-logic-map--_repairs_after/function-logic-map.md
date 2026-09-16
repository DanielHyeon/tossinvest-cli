# Function Logic Map: `_repairs_after`

`tools/logic-map/check_analysis.py:694-712` · Python · **task 7.2.6 (적대 리뷰 F8)** ·
분기 3 · 반환 2 · raise 1 (`ast.after-7.2.6.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** 아래 표는 `enumerate.py` 의 열거를 그대로 옮긴 것이다.
> 7.2.6 이 새로 만든 함수라 편집 전 판본이 없다.

## Inputs and invariants

`(root, candidate, repairs)` 를 받아 그 후보 **뒤에** 서는 수리 커밋만 골라 돌려준다.
순서는 `repairs` 의 순서(오래된 것부터)를 그대로 지킨다 — 거절 문장이 **가장 오래된** 것을
이름으로 대야 저자가 고칠 첫 자리를 가리킨다.

첫 판은 후보마다 `merge-base --is-ancestor` 를 수리 개수만큼 돌렸다(a112 는 오늘 수리가
스물여섯이라 후보 하나에 프로세스 스물여섯). `rev-list <후보>..HEAD` 한 번이 같은 집합이다 —
그것이 곧 "HEAD 에서 닿고 후보의 조상이 아닌" 커밋이기 때문이다.

**후보 자신은 안 들어간다.** 후보는 자기 조상이므로 `A..HEAD` 에서 빠진다. 그래서 수리 커밋
**자신**은 착지가 될 수 있고, 그것이 복구 경로다(수리와 번들 갱신을 한 커밋에 넣는 정직한 로트).

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L702 | If | `if not repairs:` |
| B2 | L708 | If | `if process.returncode:` |
| B3 | L712 | comprehension | ` for commit in repairs if commit in after` |

| 반환 줄 | 값 |
|---|---|
| L703 | `[]` — 수리가 하나도 없으면 git 을 안 부른다 |
| L712 | `[commit for commit in repairs if commit in after]` |

| raise 줄 | 무엇을 못 쟀나 |
|---|---|
| L710 | `rev-list <후보>..HEAD` 실패 — 빈 목록으로 물러나면 가드가 조용히 꺼진다 |

## Calls and live bindings

`subprocess.run` ×1 (`git rev-list`) · `set` · 리스트 조립. 시각·환경·워킹트리를 안 읽는다.

## State mutations and fallbacks

없다. fallback 도 없다 — 못 재면 `RuntimeError` 이고 경계가 그것을 오류 줄로 바꾼다
([[a-fault-must-become-a-verdict]]).

## Safety conclusion

생산 Go 코드 변경 0. 판정을 바꾸지 않는다 — 첫 판(`merge-base` 반복)과 **같은 집합**을 내고,
바뀐 것은 프로세스 수뿐이다. 그 동치는 변이로 못 박았다: 멤버십을 뒤집으면(S4) 착지 관련 시험
열일곱이, 후보 자신을 집합에 넣으면(S5) 둘이 빨개진다.

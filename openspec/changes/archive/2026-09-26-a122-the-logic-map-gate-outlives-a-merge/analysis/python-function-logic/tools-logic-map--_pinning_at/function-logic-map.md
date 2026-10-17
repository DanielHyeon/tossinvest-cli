# Function Logic Map: `_pinning_at`

`tools/logic-map/check_analysis.py:549-567` · Python · **task 6.1.2 → 7.5** ·
분기 **3 → 4** · 반환 1 · raise 0 (`ast.before-7.5.json` · `ast.after-7.5.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** 아래 표는 `enumerate.py` 의 열거를 그대로 옮긴 것이다.
> 6.1.2 가 만들 때의 판본은 `ast.worktree.json` 이다.

## Inputs and invariants

한 커밋에서 고정 번들이 몇 개이고 그중 어느 소스가 안 맞는가. **워킹트리를 안 읽는다.**

7.5 가 바꾼 것은 **입력의 모양**이다. `analysis: Path` 를 받아 스스로 순회하던 것이
`bundles: list[...]` 를 받는다 — `floor` · `repairs` 가 이미 그렇게 넘어오고 있었다.
번들은 걷는 동안 안 변한다(도구가 쓰지 않는다). 후보마다 다시 재던 비용은 실측으로
a071 walk 하나에서 `_pinning_bundles` 341회 · 13.07s 중 **9.30s** 였다.

## Branches and early returns — 열거 그대로 (after)

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 562 | comprehension | `for _, source, _ in bundles` — **7.5 이 더한 것** (배치 입력) |
| B2 | 563 | For | `for _, source, digest in bundles:` |
| B3 | 565 | BoolOp | `blob is None or hashlib.sha256(blob).hexdigest() != digest` |
| B4 | 565 | If | 같은 줄의 `if` |

반환 1개(`567`): `(len(bundles), mismatched)`.

**before 와의 대조**: B2·B3·B4 는 before 의 B1·B2·B3 과 같은 소스 줄이다. 더해진 것은
B1 하나뿐이고 그것은 판정이 아니라 fetch 입력이다 ([[positional-branch-ids-break-hand-renumbering]] —
번호가 아니라 소스 줄로 짝지었다).

## Calls and live bindings

`_committed_many` ×1 (before 는 `_committed_bytes` ×N) · `hashlib.sha256` · `len`.
`_pinning_bundles` 는 **더 이상 부르지 않는다** — 호출자가 재서 넘긴다.

## State mutations and fallbacks

없다.

## Safety conclusion

blob 이 없으면(`None`) 불일치로 센다 — 파일이 그 커밋에 없는 것과 내용이 다른 것을 같게
다루는 보수적인 방향이고 7.5 가 안 바꿨다. 같은 소스를 적은 번들이 여럿이면
`_committed_many` 가 한 번만 묻고 같은 바이트를 그 번들들에 나눠 준다 — 옛 판본은 같은
바이트를 N 번 읽었을 뿐 답은 같았다.

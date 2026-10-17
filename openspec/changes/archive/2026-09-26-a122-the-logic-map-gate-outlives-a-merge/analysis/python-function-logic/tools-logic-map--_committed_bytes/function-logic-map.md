# Function Logic Map: `_committed_bytes`

`tools/logic-map/check_analysis.py:425-432` · Python · **task 7.5 (리뷰 I1)** ·
분기 **1 → 0** · 반환 1 · raise 0 (`ast.before-7.5.json` · `ast.after-7.5.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** 두 표 다 `enumerate.py` 의 열거다.

## Inputs and invariants

`ref` 시점의 파일 하나. **워킹트리를 안 본다.** 없으면 `None` 이고, 부르는 쪽은 전부
`None` 을 불일치로 센다.

## Branches — 편집 전후

| 판본 | id | 줄 | 종류 | 소스 |
|---|---|---|---|---|
| before | B1 | 380 | IfExp | `None if process.returncode else process.stdout` |
| after | — | — | — | **분기 0** — 판정이 `_committed_many` 한 곳으로 갔다 |

반환 1개(`432`): `_committed_many(root, ref, [relative])[relative]`.

## Calls and live bindings

before: `subprocess.run` (`git show`). after: `_committed_many` 하나.

## State mutations and fallbacks

없다.

## Safety conclusion

"그 커밋의 blob 을 읽는다"는 철자를 **한 곳**에 둔다 (task 7.5). 두 벌이면 한쪽만 고쳐도
양쪽 시험이 초록이고, 갈리는 순간 같은 질문에 두 답이 생긴다
([[two-judgements-cover-for-each-other]]).

단건 호출자가 느려지지 않는다 — 실측 `git show` 4.05~4.80 ms 대 `cat-file --batch -Z`
2.64~3.39 ms. 남은 호출자는 셋(`_declared_landing` · `_base_shaped_bundles` ·
`validate_target`)이고 전부 파일 하나를 묻는다.

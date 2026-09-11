# Function Logic Map: `_pinning_at`

`tools/logic-map/check_analysis.py:415-423`(worktree) · Python · **task 6.1.2**

> **손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.worktree.json` 을
> `enumerate.py` 가 기계로 열거했고 아래는 그 열거를 옮긴 것이다. 이 함수는 6.1.2 가
> 새로 만든 것이라 편집 전 판본이 없다 — `ast.json`(HEAD) 이 없는 이유다.

## Inputs and invariants

한 커밋에서 고정 번들이 몇 개이고 그중 어느 소스가 안 맞는지. **워킹트리를
안 읽는다** — `_committed_bytes` 가 `git show` 로 그 커밋의 blob 만 본다.

## Branches and early returns — 열거 그대로 (분기 3)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 419 | For | `for _, source, digest in bundles:` |
| B2 | 421 | BoolOp | `blob is None or hashlib.sha256(blob).hexdigest() != digest` |
| B3 | 421 | If | `if blob is None or hashlib.sha256(blob).hexdigest() != digest:` |

반환 1개: `423`

## Calls and live bindings

`_pinning_bundles` · `_committed_bytes` · `hashlib.sha256`.

## State mutations and fallbacks

없다.

## Safety conclusion

blob 이 없으면(`None`) 불일치로 센다. 파일이 그 커밋에 없는 것과 내용이 다른
것을 같게 다루는 것은 보수적인 방향이다.

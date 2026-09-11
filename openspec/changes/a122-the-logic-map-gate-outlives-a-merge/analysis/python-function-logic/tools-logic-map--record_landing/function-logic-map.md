# Function Logic Map: `record_landing`

`tools/logic-map/check_analysis.py:1005-1061`(worktree) · Python · **task 6.1.2**

> **손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.worktree.json` 을
> `enumerate.py` 가 기계로 열거했고 아래는 그 열거를 옮긴 것이다. 이 함수는 6.1.2 가
> 새로 만든 것이라 편집 전 판본이 없다 — `ast.json`(HEAD) 이 없는 이유다.

## Inputs and invariants

계산한 값을 `landed-commit.txt` 에 쓴다. **거부하는 정상 입력을 먼저 열거했다**
([[fail-closed-must-name-what-it-rejects]]): 고정 번들이 없는 change · 증거를 빌리는
change · 이관 change(a063) · 기록이 이미 있는 change · 추적 파일이 수정된 워킹트리.

## Branches and early returns — 열거 그대로 (분기 11)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1020 | Try | `try:` |
| B2 | 1022 | ExceptHandler | `except AmbiguousChange as exc:` |
| B3 | 1024 | ExceptHandler | `except ValueError:` |
| B4 | 1027 | BoolOp | `landing_file.exists() or _declared_landing(change_dir, root) is not None` |
| B5 | 1027 | If | `if landing_file.exists() or _declared_landing(change_dir, root) is not None:` |
| B6 | 1029 | If | `if (change_dir / 'analysis' / 'function-logic-reference.txt').exists():` |
| B7 | 1035 | Try | `try:` |
| B8 | 1037 | ExceptHandler | `except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:` |
| B9 | 1039 | If | `if facts.get('execution_baseline_adoption'):` |
| B10 | 1048 | If | `if dirty.returncode:` |
| B11 | 1055 | If | `if not landing:` |

반환 8개: `1023` · `1028` · `1030` · `1038` · `1040` · `1049` · `1056` · `1058`

## Calls and live bindings

`resolve_referenced_change` · `_declared_landing` · `resolve_base` ·
`git diff --quiet HEAD` · `compute_landing` · `Path.write_text`.

## State mutations and fallbacks

**유일한 쓰기 경로다.** 성공할 때만 `landed-commit.txt` 하나를 만든다. 기존
기록은 절대 덮지 않는다 — 덮으면 그 값이 무엇이었는지가 아무 데도 안 남는다.

## Safety conclusion

더러운 워킹트리를 거부하는 이유: 기록은 **커밋된** 지점을 가리키는데 그 상태의
Go 편집은 어느 커밋에도 없다. 그대로 쓰면 5단계가 못 보는 편집이 생긴다.
`make gate` 는 이것을 자동 실행하지 않는다 — 검사 게이트가 워킹트리를 바꾸면
게이트의 뜻이 달라지고, 저장소 규칙이 mutating 단계를 사람 승인으로 묶는다.

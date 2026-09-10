# Function Logic Map: `resolve_landing`

`tools/logic-map/check_analysis.py:345-408`(HEAD `52c761de`) → `:345-424`(편집 후) · Python

> **이 표는 손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.json`(HEAD)과
> `ast.worktree.json`(편집 후)을 `enumerate.py` 가 기계로 열거했고, 아래는 그 열거를
> 옮긴 것이다. Go 용 `extract_go_ast.go` 는 Python 함수에 쓸 수 없으므로 a120 의
> `analysis/python-function-logic/` 선례를 따른다.

## Inputs and invariants

`change_dir` · `root` · `base`(해소된 비교 기준) · `analysis`(그 change 가 딛는 증거
디렉터리)를 받아 착지 커밋 SHA 하나를 돌려준다. 기록이 없으면 빈 문자열이다.

불변식은 `analysis/landing-point.md` 가 적은 그대로다: **저자는 자기 증거가 기술하는
리비전밖에 못 고른다.** 그 문장에는 전제가 있다 — 고정할 번들이 **있어야** 한다.

## Branches and early returns — 편집 전 (ast.json 열거 그대로, 분기 19)

| ID | 줄 | 종류 | 소스 | 무엇을 묻나 |
|---|---|---|---|---|
| B1 | 358 | Try | `try:` (경로가 root 밖) | — |
| B2 | 360 | ExceptHandler | `except ValueError:` | → `return ""` |
| B3 | 363 | If | `if raw is None:` | 기록이 커밋에 없다 → `return ""` |
| B4 | 365 | Try | `try:` (디코드) | — |
| B5 | 367 | ExceptHandler | `except UnicodeDecodeError as exc:` | → raise |
| B6 | 369 | If | `if not FULL_SHA.fullmatch(candidate):` | 40자리 hex 인가 |
| B7·B8 | 380 | BoolOp·If | `process.returncode or process.stdout.strip() != candidate` | 실재하는 커밋인가 |
| B9 | 382 | If | `if not _is_ancestor(root, candidate, "HEAD"):` | 착지한 적 있는 이력인가 |
| B10 | 387 | If | `if not _is_ancestor(root, base, candidate):` | base 보다 뒤인가 |
| B11 | 392 | For | `for ast_path in sorted(analysis.glob('*/ast.json')) …` | **고정 순회** |
| B12 | 392 | IfExp | `… if analysis.is_dir() else ()` | 증거 디렉터리가 있나 |
| B13·B14 | 394 | BoolOp·If | `revision != 'current'` | 이 번들이 착지를 말하나 |
| B15·B16 | 398 | BoolOp·If | `not source or not digest` | 값이 있나 |
| B17·B18 | 401 | BoolOp·If | 해시 대조 | 이 커밋의 내용인가 |
| B19 | 403 | If | `if mismatched:` | 어긋난 것이 있나 |

반환 셋 · raise 여섯: `361 return ''` · `364 return ''` · `368/373/381/383/388/404 raise` ·
`408 return candidate`.

## 결함 (task 3.2.3.1)

**B1~B10 은 전부 "이것이 어느 커밋인가"만 묻는다.** 어느 커밋인지를 *고르지 못하게*
하는 것은 B11 의 순회 하나뿐이다. 그리고 그 순회가 **0회 돌면**

1. `mismatched` 는 빈 채로 남고,
2. B19 가 거짓이 되고,
3. `408 return candidate` 가 **무조건** 참이 된다.

즉 번들이 0 인 change 에서는 착지 선언이 아무것에도 걸리지 않는다. 저자가 구간의
바닥(=base)을 고르면 `changed_existing_functions(base, base)` 가 ∅ 이고, 호출자
`check` 는 `analysis` 가 없으므로 면제 표식 경로로 통과한다. 이것이 3.2.3.1 이
적은 구멍이고, 위 세 줄이 그 기제다.

같은 이유로 **`revision: base` 번들만 있는 change 도 고정하지 못한다.** 그 번들의
`source_sha256` 은 base 를 기술하고 `validate_target` 도 그것을 해싱하지 않는다.

## 편집 — 분기 19 → 20, raise 6 → 7, 반환 3 → 3

`ast.worktree.json` 열거:

| 새 노드 | 줄 | 소스 |
|---|---|---|
| B19 | 410 | `if not pinning:` |
| raise | 415 | `landing point … is pinned by no `revision: current` evidence` |

반환은 **셋 그대로**다(`366 ''` · `369 ''` · `424 candidate`). 이 편집은 경로를
지우지 않고 마지막 반환 **앞에** 조건을 하나 더 세운다 — 그래서 "기록이 없으면
오늘과 똑같다"(task 2.2)는 두 early return 이 그대로 남아 구조로 보장된다.

## Calls and live bindings

`_committed_bytes`(git show) · `_is_ancestor`(git merge-base) · `_ast_value` ·
`FULL_SHA.fullmatch` · `hashlib.sha256` · `subprocess.run(git rev-parse)`.
**쓰기는 하나도 없다.**

## State mutations and fallbacks

없다. 로컬 `mismatched` · `pinning` 만 바꾼다. fallback 도 없다 — 실패는 전부 raise 이고
호출자 `check` 가 `cannot derive modified Go functions: …` 로 돌려준다.

## Safety conclusion

주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지 않는다. 유일한
production 호출자는 사람이 부르는 완료 게이트(`tools/gate.sh:321`)다. 실패 방향은
**게이트가 안 열리는 쪽**이므로 보수적이다. Go 파일 변경 0.

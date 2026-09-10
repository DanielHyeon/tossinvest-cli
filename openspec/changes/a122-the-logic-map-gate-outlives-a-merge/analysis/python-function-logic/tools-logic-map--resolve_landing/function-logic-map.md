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

## 편집 (task 1.8) — 분기 20 → 16, 반환 3 → 2, raise 7 → 6

`ast.after-1.8.json` 열거. **줄어든 것은 사라진 것이 아니라 옮겨간 것이다.**
선언을 읽는 앞부분(B1~B5)을 `_declared_landing` 으로 뽑았다 — 빌린 증거의 착지
공유 판정이 **같은 방법으로** 읽어야 비교가 뜻을 갖기 때문이다.

| 이 함수에서 사라진 노드 | 어디로 갔나 |
|---|---|
| `B1 Try` (경로가 root 밖) · `B2 except ValueError` | `_declared_landing` B1·B2 |
| `B3 if raw is None:` | `_declared_landing` B3 |
| `B4 Try` (디코드) · `B5 except UnicodeDecodeError` | `_declared_landing` B4·B5 |
| `raise ValueError('landing point is not UTF-8: …')` | `_declared_landing` 의 유일한 raise |
| `return ''` 둘 중 하나 | `_declared_landing` 의 `return None` 둘 |

| 이 함수에 새로 선 노드 | 줄 | 소스 |
|---|---|---|
| `if candidate is None:` | 391 | 선언 **파일이 없을** 때만 `return ''` |

`None`(파일 없음)과 `''`(빈 선언)을 가르는 것이 이 한 줄이다. 안 가르면 빈 파일을
커밋한 change 가 조용히 워킹트리를 대상으로 삼는다 — 변이 M2 로 쟀고
`test_an_empty_record_is_still_a_declaration` 이 그것을 막는다.

세는 것은 그대로다: 분기 20 − 5(이동) + 1(새) = 16, 반환 3 − 2 + 1 = 2,
raise 7 − 1 = 6. 판정은 하나도 안 없어졌다.

## 편집 (task 1.12) — 분기 16 → 16, 반환 2 → 2, raise 6 → 6

**구조가 하나도 안 바뀐다.** 고정 순회가 번들의 `file` 을 `normalized_source` 로
정규화하는 한 줄이 들어갔고, 그것은 분기가 아니다. 기계 열거에서 보이는 유일한
차이는 `B11·B12` 의 이름이다(`not source or not digest` → `not raw_source or not digest`).

| | 옛 | 새 |
|---|---|---|
| 순회가 `git show` 에 넘기는 경로 | `str(value["file"])` 그대로 | `normalized_source(...)` 의 상대경로 |

**왜 결함인가.** `validate_target` 은 같은 값을 `normalized_source` 로 정규화한다
(`check_analysis.py:454`). 번들의 `file` 은 절대경로일 수 있고, 그러면
`git show <sha>:/abs/path` 가 **언제나** 실패해서 `mismatched` 에 들어간다 — 즉
**정상 입력이 위조로 몰린다**. 3.2.3.1 이 이 순회를 넣을 때 놓친 자리이고,
a063 이관 픽스처의 번들이 절대경로를 쓰면서 드러났다.

저장소 전수(2026-09-11): 번들 `file` 필드 **상대 3048 · 절대 0**. 실물 영향은 0 이라
126건 A/B 에서 안 보인다. 그래서 변이 N4 와
`test_an_absolute_bundle_path_still_pins_a_landing` 이 이것을 재는 유일한 장치다 —
[[surviving-mutant-may-mean-accidental-safety]].

계약은 안 바뀐다. spec 은 이미 "착지 지점에서 그 change 의 모든 `revision: current`
증거 묶음의 source hash 가 일치해야 한다"고 적었고, 코드가 그 말을 못 지키고 있었다.

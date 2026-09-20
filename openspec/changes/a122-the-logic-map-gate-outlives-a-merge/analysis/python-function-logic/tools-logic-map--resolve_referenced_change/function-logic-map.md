# Function Logic Map: `resolve_referenced_change`

`tools/logic-map/check_analysis.py:231-256` · Python

> **이 표는 손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.json` 을
> `enumerate.py` 가 기계로 열거했고, 아래는 그 열거를 옮긴 것이다. Go 용
> `extract_go_ast.go` 는 Python 함수에 쓸 수 없으므로 a120 의
> `analysis/python-function-logic/` 선례를 따른다.

## Inputs and invariants

`root`(저장소 루트)와 `change`(id 한 개)를 받아 그 change 의 디렉터리 하나를 돌려준다.
불변식은 **돌려주는 디렉터리가 유일해야 한다**는 것이다 — 증거를 빌려주는 쪽이 둘이면
어느 증거로 게이트가 열렸는지 기록에 남지 않는다.

## Branches and early returns (ast.json 열거 그대로)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 239 | If | `if direct.is_dir():` |
| B2 | 242 | comprehension | `for path in (archive.iterdir() if archive.is_dir() else ())` (줄은 부모에서 상속) |
| B3 | 244 | IfExp | `archive.iterdir() if archive.is_dir() else ()` |
| B4 | 245 | BoolOp | `path.is_dir() and (matched := …) is not None and matched.group('change') == change` |
| B5 | 249 | If | `if not matches:` |
| B6 | 251 | If | `if len(matches) > 1:` |

반환 둘 · raise 둘:

| 줄 | 종류 | 내용 |
|---|---|---|
| 240 | return | `direct` — **early return** |
| 250 | raise | `ValueError("reference change is neither open nor archived: …")` |
| 253 | raise | `ValueError("archive holds N copies of …")` |
| 256 | return | `archive / matches[0]` |

## 결함 (task 3.2.4)

**B1 의 early return 이 B2~B6 을 통째로 건너뛴다.** 그러므로

1. 활성 디렉터리가 있으면 아카이브는 **한 번도 열리지 않는다**. 아카이브본이 있어도
   조용히 활성이 이긴다.
2. 중복 판정(B6)은 `matches` 위에서만 돈다. `matches` 는 B2 의 아카이브 순회가
   만든 것이므로 **중복 세기의 범위가 아카이브 안이다**. 활성+아카이브 쌍은 세지 않는다.

이 둘은 같은 하나의 원인 — 세는 범위가 반환 경로보다 좁다 —
[[existence-check-is-not-a-role-check]] 가 말하는 모양이다.

## Calls and live bindings

`Path.is_dir` · `Path.iterdir` · `ARCHIVED_CHANGE.fullmatch` · `re.Match.group` ·
`sorted` · `len` · `str.join` · `ValueError`. 파일을 **읽지 않는다** — 디렉터리 존재만 본다.

## State mutations and fallbacks

없다. 순수 함수이고 fallback 도 없다. 다만 호출자 `check`(`:689-692`)가
`except ValueError` 로 **되돌아간다** — 그 fallback 은 이 함수 밖에 있다.

## Safety conclusion

주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지 않는다. 유일한
production 호출자는 사람이 직접 부르는 완료 게이트(`tools/gate.sh:321`)이고
CI 는 이 경로를 돌지 않는다(`.github/workflows/ci.yml:94`). 실패 방향은
**게이트가 안 열리는 쪽**이므로 보수적이다.

---

## 편집 — task 7.6 (규칙 한 집) · 분기 10 → 10 · 반환 1 → 1 · raise 3 → 3

편집 전 `ast.before-7.6.json`(HEAD `2b5b05c1`) · 편집 후 `ast.after-7.6.json`(워킹트리). 아래 표는 두 열거의
`source` 를 스크립트가 줄 단위로 대조한 것이다.

아카이브 이름 해독을 `_archived_change_id` 에 묻는다(I6). 못 찾음 문장에서 "reference" 를 뺐다 — 이 해소기는 게이트 **대상**도 찾고, 7.6 뒤로 그 문장이 오타 난 대상 id 에 그대로 나간다(I4). `AmbiguousChange` 타입을 `ValueError` 로 되돌렸다: 타입을 따로 둔 이유(호출자 둘이 "못 찾음"만 fallback 으로 흘려야 했다)가 I4 로 없어져 가르는 호출자가 0 이다.

| | 종류 | 소스 |
|---|---|---|
| 사라짐 | 분기 | ` for path in (archive.iterdir() if archive.is_dir() else ()) if path.is_dir() and (matched := ARCHIVED_CHANGE.` |
| 생김 | 분기 | ` for path in (archive.iterdir() if archive.is_dir() else ()) if path.is_dir() and _archived_change_id(path.nam` |
| 사라짐 | 분기 | `path.is_dir() and (matched := ARCHIVED_CHANGE.fullmatch(path.name)) is not None and (matched.group('change') =` |
| 생김 | 분기 | `path.is_dir() and _archived_change_id(path.name) == change` |
| 사라짐 | raise | `raise ValueError(f'reference change is neither open nor archived: {change}')` |
| 생김 | raise | `raise ValueError(f'change is neither open nor archived: {change}')` |
| 사라짐 | raise | `raise AmbiguousChange(f'{change} is open and archived at once: ' + ', '.join((path.relative_to(root).as_posix(` |
| 생김 | raise | `raise ValueError(f'{change} is open and archived at once: ' + ', '.join((path.relative_to(root).as_posix() for` |

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:257-301` · 분기 11 · 반환 1 · raise 3 (편집 전 L249-290 · 분기 10 · 반환 1 · raise 3, `ast.before-7.5.2.3.json` = revision `1d12520c`).

활성/아카이브를 **고르는** 판정이 깔때기를 지난다. `is_dir()` 이 삼키던 `OSError` 도 이제 결함이다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 276 | Try | `try:` |
| B2 | 278 | ExceptHandler | `except (FileNotFoundError, NotADirectoryError):` |
| B3 | 280 | comprehension | ` for name, is_dir in entries if is_dir and _archived_change_id(name) == change` |
| B4 | 281 | BoolOp | `is_dir and _archived_change_id(name) == change` |
| B5 | 282 | IfExp | `[direct] if open_here else []` |
| B6 | 283 | If | `if not found:` |
| B7 | 287 | BoolOp | `open_here and archived` |
| B8 | 287 | If | `if open_here and archived:` |
| B9 | 293 | comprehension | ` for path in found` |
| B10 | 295 | If | `if len(archived) > 1:` |
| B11 | 299 | comprehension | ` for path in archived` |

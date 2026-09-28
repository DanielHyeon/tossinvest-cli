# Python Function Logic Map: `resolve_base` (편집 전)

- Source `tools/logic-map/check_analysis.py` @ base · sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · L910–1007
- 분기 18 · 반환 2 · raise 9 — 열거: a122 `enumerate.py`(기계, 손으로 고르지 않음)

## a125 편집 지점 (이관 관련 분기 — AST 원문 정규식 `(?i)adopt|audited|execution_baseline`)

- 분기: B13(L994), B14(L996), B16(L1000)
- 이관 값 반환: 없음
- 처분: 이관 분기·인자·문맥 키를 지운다. 나머지 분기는 불변 — 편집 후 재추출(3.3)에서 분기 수 차이가 이 목록과 같아야 한다.

## 분기 전수

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 934 | Try | `try:` |
| B2 | 938 | ExceptHandler | `except FileNotFoundError as exc:` |
| B3 | 944 | ExceptHandler | `except (OSError, UnicodeDecodeError) as exc:` |
| B4 | 946 | If | `if not FULL_SHA.fullmatch(candidate):` |
| B5 | 953 | IfExp | `[(relative, committed)] if committed is not None else _committed_elsewhere(root, head, change_id, re` |
| B6 | 954 | For | `for place, value in held:` |
| B7 | 956 | If | `if shown == candidate:` |
| B8 | 960 | If | `if place == relative:` |
| B9 | 979 | If | `if process.returncode:` |
| B10 | 981 | BoolOp | `process.stderr.strip() or f'invalid Function Logic Map base: {value}'` |
| B11 | 986 | If | `if persisted != candidate:` |
| B12 | 989 | Try | `try:` |
| B13 | 994 | ExceptHandler | `except AdoptionError as exc:` |
| B14 | 996 | IfExp | `str(adoption['effective_base']) if adoption else persisted` |
| B15 | 997 | If | `if context is not None:` |
| B16 | 1000 | If | `if adoption:` |
| B17 | 1005 | BoolOp | `override and resolve(override) != effective` |
| B18 | 1005 | If | `if override and resolve(override) != effective:` |

## 편집 후 재추출 (a125 3.3, `9e63b681`)

- `ast.after.json` — 분기 18→14 · 반환 2→2 · raise 9→8
- 분기 · 반환 원문 대조(`difflib.ndiff`, 위치가 아니라 원문으로 — 분기 번호는 위치다):

    - Try try:
    - ExceptHandler except AdoptionError as exc:
    - IfExp str(adoption['effective_base']) if adoption else persisted
    - If if adoption:
    - BoolOp override and resolve(override) != effective
    + BoolOp override and resolve(override) != persisted
    - If if override and resolve(override) != effective:
    + If if override and resolve(override) != persisted:
    - effective
    + persisted
- **편집 전 목록과의 차이(정정)**: 편집 전 「편집 지점」은 정규식으로 B13 · B14 · B16 셋을 골랐는데, 실제로 지운 분기는
  넷이다 — `IfExp …if adoption else persisted` 가 빠져 있었다(원문에 `adoption` 이 있는데 목록에서 누락된 것은 편집 전 목록을
  만든 뒤 정규식을 대소문자 무시로 다시 돌리며 고친 목록이 한 줄 어긋났기 때문이다). 그리고 `SDD_BASE_REF` 대조 대상이
  `effective` → `persisted` 로 **바뀐** 분기 둘(BoolOp · If)이 있다 — 기록이 없을 때 `effective == persisted` 였으므로 기록 없는
  change 의 판정은 같다(1.1 논증). 기록 있는 a063 에서만 대조 대상이 E → P 로 바뀐다(반전 시험 `test_sdd_base_ref_accepts_only_the_base_commit`).

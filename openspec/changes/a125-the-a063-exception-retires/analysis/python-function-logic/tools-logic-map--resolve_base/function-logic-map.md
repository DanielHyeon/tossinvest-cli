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

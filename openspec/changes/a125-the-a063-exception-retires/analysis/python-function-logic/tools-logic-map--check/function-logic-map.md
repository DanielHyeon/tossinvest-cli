# Python Function Logic Map: `check` (편집 전)

- Source `tools/logic-map/check_analysis.py` @ base · sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · L2727–2769
- 분기 3 · 반환 2 · raise 0 — 열거: a122 `enumerate.py`(기계, 손으로 고르지 않음)

## a125 편집 지점 (이관 관련 분기 — AST 원문 정규식 `(?i)adopt|audited|execution_baseline`)

- 분기: 없음(분기 없음 — 인자·문자열·호출만)
- 이관 값 반환: 없음
- 처분: 이관 분기·인자·문맥 키를 지운다. 나머지 분기는 불변 — 편집 후 재추출(3.3)에서 분기 수 차이가 이 목록과 같아야 한다.

## 분기 전수

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 2757 | IfExp | `{} if context is None else context` |
| B2 | 2765 | If | `if not judged:` |
| B3 | 2769 | IfExp | `[moved] if moved else verdict` |

## 편집 후 재추출 (a125 3.3, `9e63b681`)

- `ast.after.json` — 분기 3→3 · 반환 2→2 · raise 0→0
- 분기 · 반환 원문 대조(`difflib.ndiff`, 위치가 아니라 원문으로 — 분기 번호는 위치다):

    (차이 없음)
- 재추출 갱신: `5440edad`(리뷰 P2 의 docstring 두 자리 수정 뒤) — 분기 · 반환 원문이 `9e63b681` 판과 같다. `ast.after.json` 은 `5440edad` 판이다.

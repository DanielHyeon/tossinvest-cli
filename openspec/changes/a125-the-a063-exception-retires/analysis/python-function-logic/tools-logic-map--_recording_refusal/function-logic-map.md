# Python Function Logic Map: `_recording_refusal` (편집 전)

- Source `tools/logic-map/check_analysis.py` @ base · sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · L3201–3299
- 분기 12 · 반환 10 · raise 1 — 열거: a122 `enumerate.py`(기계, 손으로 고르지 않음)

## a125 편집 지점 (이관 관련 분기 — AST 원문 정규식 `(?i)adopt|audited|execution_baseline`)

- 분기: B9(L3273)
- 이관 값 반환: L3274
- 처분: 이관 분기·인자·문맥 키를 지운다. 나머지 분기는 불변 — 편집 후 재추출(3.3)에서 분기 수 차이가 이 목록과 같아야 한다.

## 분기 전수

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 3228 | If | `if landing_file.is_symlink():` |
| B2 | 3238 | If | `if _landing_record(change_dir, root, head) is not None:` |
| B3 | 3246 | If | `if _kind(landing_file):` |
| B4 | 3258 | Try | `try:` |
| B5 | 3260 | ExceptHandler | `except FileNotFoundError:` |
| B6 | 3262 | ExceptHandler | `except OSError as exc:` |
| B7 | 3269 | Try | `try:` |
| B8 | 3271 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B9 | 3273 | If | `if facts.get('execution_baseline_adoption'):` |
| B10 | 3276 | If | `if why:` |
| B11 | 3286 | If | `if dirty.returncode not in (0, 1):` |
| B12 | 3293 | If | `if dirty.returncode:` |

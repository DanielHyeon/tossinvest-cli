# Function Logic Map: `_verdict` (Python, a122 task 7.5.2.2)

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1756-1843` · 분기 26 · 반환 4 · raise 0 (새 함수).

새 함수(옛 `check` 의 뒷부분 그대로 — 대상 판정의 분기 · 반환은 옛 `check` 의 그 부분과 같다). 떼어 낸 까닭은 `check` 의 한 출구에서 대조하려는 것 하나. 안에서는 `evidence` 만 본다. 목록 · 파일을 못 읽은 번들은 `cannot read every file in the bundle (<사유>: <이름>)` — `filename` 이 글자가 아니면 번들 이름으로 떨어진다(이름을 대려다 판정 줄 없이 죽지 않게, 위 회귀). 경로를 담는 쪽은 `_read_regular` 가 못 박는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1765 | If | `if not evidence.present:` |
| B2 | 1766 | If | `if required:` |
| B3 | 1767 | comprehension | ` for source, function in required` |
| B4 | 1774 | IfExp | `[] if EXEMPTION in review_text else [f'missing analysis or `{EXEMPTION}` review marker']` |
| B5 | 1776 | If | `if not evidence.targets:` |
| B6 | 1790 | For | `for target in evidence.targets:` |
| B7 | 1791 | If | `if not os.path.lexists(target / 'function-logic-map.md'):` |
| B8 | 1793 | Try | `try:` |
| B9 | 1795 | ExceptHandler | `except OSError as exc:` |
| B10 | 1799 | IfExp | `exc.filename if isinstance(exc.filename, str) else ''` |
| B11 | 1800 | IfExp | `Path(named).name if named else target.name` |
| B12 | 1801 | BoolOp | `exc.strerror or exc` |
| B13 | 1802 | comprehension | ` for target, text in bundle_texts.items()` |
| B14 | 1815 | If | `if landing:` |
| B15 | 1816 | Try | `try:` |
| B16 | 1817 | comprehension | ` for _, source, _ in _select_pinning(root, evidence)` |
| B17 | 1818 | ExceptHandler | `except ValueError:` |
| B18 | 1821 | For | `for target in evidence.targets:` |
| B19 | 1826 | If | `if binding:` |
| B20 | 1827 | If | `if binding in covered:` |
| B21 | 1830 | For | `for binding, expected in required.items():` |
| B22 | 1832 | If | `if target is None:` |
| B23 | 1837 | BoolOp | `expected.get('current_hash') or expected.get('base_hash')` |
| B24 | 1838 | If | `if ast_value.get('source_sha256') != expected_hash:` |
| B25 | 1840 | IfExp | `'current' if expected.get('current_hash') else 'base'` |
| B26 | 1841 | If | `if ast_value.get('revision', 'current') != expected_revision:` |

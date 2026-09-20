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

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:2069-2165` · 분기 26 · 반환 4 · raise 0 (편집 전 L1756-1843 · 분기 26 · 반환 4 · raise 0, `ast.before-7.5.2.3.json` = revision `1d12520c`).

번들 목록을 `lexists` 로 다시 묻지 않고 한 번 읽은 증거의 목록을 쓴다. 목록을 **못 연** 번들은 이름 댄 판정 줄이고(조용히 빈 목록으로 두면 그 번들의 표 전부가 감사에서 빠진다), 번들 파일의 실패는 사유와 이름을 함께 댄다 — 권한 · 디렉터리 · 종류 · 크기 · UTF-8 아님이 모두 이 한 자리로 온다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 2078 | If | `if not evidence.present:` |
| B2 | 2079 | If | `if required:` |
| B3 | 2080 | comprehension | ` for source, function in required` |
| B4 | 2087 | IfExp | `[] if EXEMPTION in review_text else [f'missing analysis or `{EXEMPTION}` review marker']` |
| B5 | 2089 | If | `if not evidence.targets:` |
| B6 | 2103 | For | `for target in evidence.targets:` |
| B7 | 2106 | If | `if target in evidence.unlistable:` |
| B8 | 2111 | If | `if 'function-logic-map.md' not in names:` |
| B9 | 2113 | Try | `try:` |
| B10 | 2116 | ExceptHandler | `except OSError as exc:` |
| B11 | 2121 | IfExp | `exc.filename if isinstance(exc.filename, str) else ''` |
| B12 | 2122 | IfExp | `Path(named).name if named else target.name` |
| B13 | 2124 | comprehension | ` for target, text in bundle_texts.items()` |
| B14 | 2137 | If | `if landing:` |
| B15 | 2138 | Try | `try:` |
| B16 | 2139 | comprehension | ` for _, source, _ in _select_pinning(root, evidence)` |
| B17 | 2140 | ExceptHandler | `except ValueError:` |
| B18 | 2143 | For | `for target in evidence.targets:` |
| B19 | 2148 | If | `if binding:` |
| B20 | 2149 | If | `if binding in covered:` |
| B21 | 2152 | For | `for binding, expected in required.items():` |
| B22 | 2154 | If | `if target is None:` |
| B23 | 2159 | BoolOp | `expected.get('current_hash') or expected.get('base_hash')` |
| B24 | 2160 | If | `if ast_value.get('source_sha256') != expected_hash:` |
| B25 | 2162 | IfExp | `'current' if expected.get('current_hash') else 'base'` |
| B26 | 2163 | If | `if ast_value.get('revision', 'current') != expected_revision:` |

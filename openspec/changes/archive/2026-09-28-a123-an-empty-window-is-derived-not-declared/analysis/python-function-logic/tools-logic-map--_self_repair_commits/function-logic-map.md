# Python Function Logic Map: `_self_repair_commits` (편집 전)

- Source: `tools/logic-map/check_analysis.py` @ base `77e36cca22498ff5e022c0392f4166cec3260c7d`
- 열거: `python3 openspec/changes/archive/2026-09-26-a122-the-logic-map-gate-outlives-a-merge/analysis/python-function-logic/enumerate.py tools/logic-map/check_analysis.py _self_repair_commits <base>` → `ast.before.json`
- source_sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · 줄 1909–2009 · 분기 14 · 반환 2 · raise 4

## a123 관점

**편집 안 함 — 재사용.** R1 조건 2 와 R2 귀속이 이 함수의 결과(비병합 · 디렉터리+Go, 오래된 것부터)를 `_repairs_after(root, base, …, head)` 로 base 뒤 부분만 골라 쓴다. 알려진 구멍(디렉터리를 안 만진 Go 커밋 · 병합 자신의 변경)은 같은 방향으로 상속한다 — 픽스처 (c) 로 못 박고 고치지 않는다.

## 분기 (AST 열거 — 손으로 고르지 않음)

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 1936 | Try | `try:` |
| B2 | 1938 | ExceptHandler | `except ValueError as exc:` |
| B3 | 1945 | If | `if change_dir.startswith(ARCHIVE_PREFIX):` |
| B4 | 1949 | If | `if before:` |
| B5 | 1962 | If | `if touching.returncode:` |
| B6 | 1971 | If | `if not hashes:` |
| B7 | 1992 | If | `if listing.returncode:` |
| B8 | 2001 | For | `for block in listing.stdout.strip(b'\x00').split(b'\x00\x00'):` |
| B9 | 2003 | BoolOp | `names and (not names.startswith(b'\n'))` |
| B10 | 2003 | BoolOp | `not re.fullmatch(b'[0-9a-f]{40}\|[0-9a-f]{64}', commit) or (names and (not names.startswith` |
| B11 | 2003 | If | `if not re.fullmatch(b'[0-9a-f]{40}\|[0-9a-f]{64}', commit) or (names and (not names.startsw` |
| B12 | 2005 | If | `if any((name.endswith(b'.go') for name in names[1:].split(b'\x00'))):` |
| B13 | 2005 | comprehension | ` for name in names[1:].split(b'\x00')` |
| B14 | 2009 | comprehension | ` for commit in reversed(hashes) if commit in flagged` |

## 반환

- L1972: `[]`
- L2009: `[commit for commit in reversed(hashes) if commit in flagged]`

## raise

- L1943: `raise RuntimeError(f"cannot name this change's directory under the repository: {exc}") from exc`
- L1967: `raise RuntimeError(f'cannot list the commits that touched {change_dir}: {_first_line(touching.stderr`
- L1993: `raise RuntimeError('cannot read the files those commits changed: ' + _first_line(listing.stderr, 'gi`
- L2004: `raise RuntimeError('cannot read the files those commits changed: unexpected git log -z output')`

# Function Logic Map: `_bundle_text` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1312-1343` · 분기 7 · 반환 1 · raise 0 (편집 전 L1217-1240 · 분기 4 · 반환 1 · raise 0, `ast.before-7.5.2.json` = revision `e9f905bd`).

`known`(한 번 읽은 `ast.json`)이 있으면 `ast.json` 은 그 바이트만 쓴다 — 그 뒤에 생긴 것도 안 읽는다. **GREEN 도중 편집 집합에 들어왔다**(review.md).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1325 | For | `for path in sorted(map_path.parent.glob('*')):` |
| B2 | 1326 | If | `if not path.is_file():` |
| B3 | 1328 | Try | `try:` |
| B4 | 1329 | BoolOp | `known is not None and path.name == 'ast.json'` |
| B5 | 1329 | If | `if known is not None and path.name == 'ast.json':` |
| B6 | 1333 | If | `if raw is None:` |
| B7 | 1338 | ExceptHandler | `except (OSError, UnicodeDecodeError):` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1378-1409` · 분기 6 · 반환 1 · raise 0 (편집 전 L1312-1343 · 분기 7 · 반환 1 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`(대상, 한 번 읽은 ast 바이트)` 를 받는다. `ast.json` 은 목록과 **무관하게** 그 바이트다 — 7.5.2 는 살아 있는 목록에 있을 때만 썼고, 한 번 읽은 뒤 지워지면 판정한 바이트가 빠져 열거형 호출 판정이 꺼졌다(Codex P1 재현: fc35eb2d `[]`). 목록은 `iterdir()` — 못 열면 `OSError` 가 올라가 `check` 가 이름 댄 줄로 만든다(`glob` 은 그것을 삼켰다).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1396 | comprehension | ` for path in target.iterdir() if path.name != 'ast.json'` |
| B2 | 1397 | If | `if ast_raw is not None:` |
| B3 | 1400 | For | `for name in sorted(paths):` |
| B4 | 1401 | Try | `try:` |
| B5 | 1402 | IfExp | `_decoded(ast_raw) if name == 'ast.json' else paths[name].read_text(encoding='utf-8')` |
| B6 | 1404 | ExceptHandler | `except (OSError, UnicodeDecodeError):` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1412-1455` · 분기 9 · 반환 1 · raise 0 (편집 전 L1378-1409 · 분기 6 · 반환 1 · raise 0, `ast.before-7.5.2.2.json` = revision `908a8a36`).

번들 파일은 `_read_regular` 로만 연다 — 7.5.2.1 이 `_bundle_text` 를 다시 쓰며 `is_file()` 거름을 빠뜨려 FIFO 에 게이트가 멎고 `/dev/zero` 에 `MemoryError` 로 죽었다(재리뷰 출처 넷 재현). 정규 파일 아닌 것 · 사라진 것은 건너뛰고, **못 읽는 정규 파일은 `OSError` 로 올린다**(재리뷰 적대: 조용히 건너뛰면 그 파일의 열거형 호출 표가 판정에서 빠진다) — `check` 가 이름 댄 줄로 만든다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1435 | comprehension | ` for path in target.iterdir() if path.name != 'ast.json'` |
| B2 | 1436 | If | `if ast_raw is not None:` |
| B3 | 1439 | For | `for name in sorted(paths):` |
| B4 | 1440 | If | `if name == 'ast.json':` |
| B5 | 1443 | Try | `try:` |
| B6 | 1445 | ExceptHandler | `except FileNotFoundError:` |
| B7 | 1447 | If | `if raw is None:` |
| B8 | 1449 | Try | `try:` |
| B9 | 1451 | ExceptHandler | `except UnicodeDecodeError:` |

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1661-1710` · 분기 10 · 반환 1 · raise 1 (편집 전 L1412-1455 · 분기 9 · 반환 1 · raise 0, `ast.before-7.5.2.3.json` = revision `1d12520c`).

**조용히 건너뛰는 모양이 하나만 남았다**(사라진 파일). 종류가 틀린 것 · 못 푸는 바이트는 이름 댄 판정 줄이 된다 — 게이트가 여는 그 순간에만 FIFO 로 바꿨다 되돌리면 감사가 꺼졌고(실측 6/14), 못 푸는 바이트 한 개로도 같은 일이 일어났다. 목록은 호출자가 한 번 읽은 것을 받는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1691 | IfExp | `names if names is not None else tuple((name for name, _ in _listed(target)))` |
| B2 | 1691 | comprehension | ` for name, _ in _listed(target)` |
| B3 | 1692 | comprehension | ` for name in listed if name != 'ast.json'` |
| B4 | 1693 | If | `if ast_raw is not None:` |
| B5 | 1696 | For | `for name in sorted(paths):` |
| B6 | 1697 | If | `if name == 'ast.json':` |
| B7 | 1700 | Try | `try:` |
| B8 | 1702 | ExceptHandler | `except FileNotFoundError:` |
| B9 | 1704 | Try | `try:` |
| B10 | 1706 | ExceptHandler | `except UnicodeDecodeError as exc:` |

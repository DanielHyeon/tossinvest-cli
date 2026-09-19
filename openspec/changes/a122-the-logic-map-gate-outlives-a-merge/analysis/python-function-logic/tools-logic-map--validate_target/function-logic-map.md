# Function Logic Map: `validate_target`

> 손으로 읽어서 만들지 않았다 — `enumerate.py` 열거.

## Inputs and invariants

미리 읽은 것을 받는다. 빠진 경로면 예전처럼 한 번 읽는다 — 미리 읽기는 최적화이고 판정은 그것에 기대지 않는다.


## task 7.5.1 — gstack 리뷰의 permissive 결함 수리

`ast.before-7.5.1.json`(= git revision **`b29e1f4e`**, 소스 해시가 그 커밋과 일치)과
`ast.after-7.5.1.json` 을 같은 열거기로 뽑아 **순서 있는 배열**을 difflib 으로 정렬했다
(분기 37 · 반환 5 · raise 0 → 분기 39 · 반환 5 · raise 0).

미리 읽은 것을 받는다. 빠진 경로면 예전처럼 한 번 읽는다 — 미리 읽기는 최적화이고 판정은 그것에 기대지 않는다.

| | 종류 | 소스 |
|---|---|---|
| + | BoolOp | `prefetched is not None and relative in prefetched` |
| + | IfExp | `prefetched[relative] if prefetched is not None and relative in prefetched else _committed_` |

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1370-1496` · 분기 46 · 반환 6 · raise 0 (편집 전 L1251-1356 · 분기 39 · 반환 5 · raise 0, `ast.before-7.5.2.json` = revision `e9f905bd`).

`evidence`(한 번 읽은 `ast.json` 들)를 받으면 디스크를 다시 읽지 않는다: 그 읽기에 없던 것은 `missing ast.json`, 못 읽은 것은 `ast.json could not be read`. `start`·`end` 가 사전이 아니거나 `branches` 가 목록이 아니면 `ast.json is invalid`(옛 판본은 `AttributeError`/`TypeError` 로 판정 줄 없이 끝났다 — 레드팀, 7.5 이전부터; 저장소 3,048 번들 중 이 모양 0).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1385 | For | `for name in REQUIRED:` |
| B2 | 1387 | BoolOp | `name == 'ast.json' and evidence is not None` |
| B3 | 1387 | If | `if name == 'ast.json' and evidence is not None:` |
| B4 | 1388 | If | `if path not in evidence:` |
| B5 | 1392 | If | `if raw is None:` |
| B6 | 1396 | If | `if path.exists():` |
| B7 | 1401 | If | `if 'TODO' in texts[name]:` |
| B8 | 1403 | If | `if 'ast.json' not in texts:` |
| B9 | 1405 | Try | `try:` |
| B10 | 1407 | ExceptHandler | `except ValueError:` |
| B11 | 1410 | BoolOp | `not isinstance(value, dict) or any((not value.get(key) for key in keys))` |
| B12 | 1410 | If | `if not isinstance(value, dict) or any((not value.get(key) for key in keys)):` |
| B13 | 1410 | comprehension | ` for key in keys` |
| B14 | 1415 | BoolOp | `not isinstance(value['start'], dict) or not isinstance(value['end'], dict) or (not isinstance(value.get('branches') or [` |
| B15 | 1415 | If | `if not isinstance(value['start'], dict) or not isinstance(value['end'], dict) or (not isinstance(value.get('branches') o` |
| B16 | 1416 | BoolOp | `value.get('branches') or []` |
| B17 | 1418 | Try | `try:` |
| B18 | 1420 | ExceptHandler | `except ValueError as exc:` |
| B19 | 1423 | If | `if revision == 'current':` |
| B20 | 1424 | If | `if revision_ref:` |
| B21 | 1427 | BoolOp | `prefetched is not None and relative in prefetched` |
| B22 | 1427 | IfExp | `prefetched[relative] if prefetched is not None and relative in prefetched else _committed_bytes(root, revision_ref, rela` |
| B23 | 1429 | If | `if blob is None:` |
| B24 | 1431 | If | `if hashlib.sha256(blob).hexdigest() != value['source_sha256']:` |
| B25 | 1433 | If | `if not source.is_file():` |
| B26 | 1435 | If | `if hashlib.sha256(source.read_bytes()).hexdigest() != value['source_sha256']:` |
| B27 | 1437 | If | `if revision != 'base':` |
| B28 | 1441 | For | `for section in MAP_SECTIONS:` |
| B29 | 1442 | If | `if section not in logic:` |
| B30 | 1444 | BoolOp | `relative not in logic or function not in logic` |
| B31 | 1444 | If | `if relative not in logic or function not in logic:` |
| B32 | 1448 | BoolOp | `value.get('branches') or []` |
| B33 | 1450 | If | `if '# Branch Test Map' not in branch_map:` |
| B34 | 1453 | If | `if len(mapped) != len(set(mapped)):` |
| B35 | 1455 | comprehension | ` for branch in branches if isinstance(branch, dict) and branch.get('id')` |
| B36 | 1458 | BoolOp | `isinstance(branch, dict) and branch.get('id')` |
| B37 | 1460 | BoolOp | `branches and len(expected) != len(branches)` |
| B38 | 1460 | If | `if branches and len(expected) != len(branches):` |
| B39 | 1463 | If | `if missing:` |
| B40 | 1470 | IfExp | `expected if branches else {'B1'}` |
| B41 | 1471 | If | `if unexpected:` |
| B42 | 1488 | IfExp | `test_index(root) if index is None else index` |
| B43 | 1491 | BoolOp | `not branches and 'B1' not in mapped` |
| B44 | 1491 | If | `if not branches and 'B1' not in mapped:` |
| B45 | 1494 | BoolOp | `'# Risk Pattern Report' not in risk or relative not in risk` |
| B46 | 1494 | If | `if '# Risk Pattern Report' not in risk or relative not in risk:` |

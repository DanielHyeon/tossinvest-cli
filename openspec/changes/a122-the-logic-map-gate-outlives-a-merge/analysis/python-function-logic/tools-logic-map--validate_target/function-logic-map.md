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

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1460-1584` · 분기 43 · 반환 5 · raise 0 (편집 전 L1370-1496 · 분기 46 · 반환 6 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`held` 가 **필수 키워드**다(시험만 쓰던 디스크 fallback 삭제). 모양은 `_parse_ast` 한 곳이 본다 — 이 함수 안의 따로 된 모양 검사와 반환 하나를 지웠다(반환 6 → 5). 글자가 아닌 `ast.json` 은 이제 `invalid`(옛: `UnicodeDecodeError` 가 판정 전체를 `cannot judge` 한 줄로). 순서가 하나 바뀐다: 사전인데 필수 칸이 비고 **동시에** 모양이 틀리면 옛 판본은 placeholder, 새 판본은 invalid — 저장소 0 건.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1476 | For | `for name in REQUIRED:` |
| B2 | 1478 | If | `if name == 'ast.json':` |
| B3 | 1479 | If | `if path not in held:` |
| B4 | 1482 | If | `if held[path] is None:` |
| B5 | 1485 | Try | `try:` |
| B6 | 1487 | ExceptHandler | `except ValueError:` |
| B7 | 1489 | If | `if path.exists():` |
| B8 | 1494 | If | `if 'TODO' in texts[name]:` |
| B9 | 1496 | If | `if 'ast.json' not in texts:` |
| B10 | 1501 | If | `if problem == 'invalid':` |
| B11 | 1504 | BoolOp | `problem or any((not value.get(key) for key in keys))` |
| B12 | 1504 | If | `if problem or any((not value.get(key) for key in keys)):` |
| B13 | 1504 | comprehension | ` for key in keys` |
| B14 | 1506 | Try | `try:` |
| B15 | 1508 | ExceptHandler | `except ValueError as exc:` |
| B16 | 1511 | If | `if revision == 'current':` |
| B17 | 1512 | If | `if revision_ref:` |
| B18 | 1515 | BoolOp | `prefetched is not None and relative in prefetched` |
| B19 | 1515 | IfExp | `prefetched[relative] if prefetched is not None and relative in prefetched else _committed_bytes(root, revision_ref, rela` |
| B20 | 1517 | If | `if blob is None:` |
| B21 | 1519 | If | `if hashlib.sha256(blob).hexdigest() != value['source_sha256']:` |
| B22 | 1521 | If | `if not source.is_file():` |
| B23 | 1523 | If | `if hashlib.sha256(source.read_bytes()).hexdigest() != value['source_sha256']:` |
| B24 | 1525 | If | `if revision != 'base':` |
| B25 | 1529 | For | `for section in MAP_SECTIONS:` |
| B26 | 1530 | If | `if section not in logic:` |
| B27 | 1532 | BoolOp | `relative not in logic or function not in logic` |
| B28 | 1532 | If | `if relative not in logic or function not in logic:` |
| B29 | 1536 | BoolOp | `value.get('branches') or []` |
| B30 | 1538 | If | `if '# Branch Test Map' not in branch_map:` |
| B31 | 1541 | If | `if len(mapped) != len(set(mapped)):` |
| B32 | 1543 | comprehension | ` for branch in branches if isinstance(branch, dict) and branch.get('id')` |
| B33 | 1546 | BoolOp | `isinstance(branch, dict) and branch.get('id')` |
| B34 | 1548 | BoolOp | `branches and len(expected) != len(branches)` |
| B35 | 1548 | If | `if branches and len(expected) != len(branches):` |
| B36 | 1551 | If | `if missing:` |
| B37 | 1558 | IfExp | `expected if branches else {'B1'}` |
| B38 | 1559 | If | `if unexpected:` |
| B39 | 1576 | IfExp | `test_index(root) if index is None else index` |
| B40 | 1579 | BoolOp | `not branches and 'B1' not in mapped` |
| B41 | 1579 | If | `if not branches and 'B1' not in mapped:` |
| B42 | 1582 | BoolOp | `'# Risk Pattern Report' not in risk or relative not in risk` |
| B43 | 1582 | If | `if '# Risk Pattern Report' not in risk or relative not in risk:` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1508-1642` · 분기 46 · 반환 5 · raise 0 (편집 전 L1460-1584 · 분기 43 · 반환 5 · raise 0, `ast.before-7.5.2.2.json` = revision `908a8a36`).

번들 파일은 `_read_regular` 로만 연다 — 7.5.2.1 이 `_bundle_text` 를 다시 쓰며 `is_file()` 거름을 빠뜨려 FIFO 에 게이트가 멎고 `/dev/zero` 에 `MemoryError` 로 죽었다(재리뷰 출처 넷 재현). 필수 산문(`exists()` 뒤 `read_text`)도 앞 로트부터 종류를 안 봤다 — 없으면 `missing`, 정규 파일이 아니거나 못 읽으면 `could not be read`(ast.json 과 같은 말). 글자가 아니면 옛 판본처럼 올라간다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1524 | For | `for name in REQUIRED:` |
| B2 | 1526 | If | `if name == 'ast.json':` |
| B3 | 1527 | If | `if path not in held:` |
| B4 | 1530 | If | `if held[path] is None:` |
| B5 | 1533 | Try | `try:` |
| B6 | 1535 | ExceptHandler | `except ValueError:` |
| B7 | 1541 | Try | `try:` |
| B8 | 1543 | ExceptHandler | `except FileNotFoundError:` |
| B9 | 1546 | ExceptHandler | `except OSError:` |
| B10 | 1548 | If | `if raw is None:` |
| B11 | 1552 | If | `if 'TODO' in texts[name]:` |
| B12 | 1554 | If | `if 'ast.json' not in texts:` |
| B13 | 1559 | If | `if problem == 'invalid':` |
| B14 | 1562 | BoolOp | `problem or any((not value.get(key) for key in keys))` |
| B15 | 1562 | If | `if problem or any((not value.get(key) for key in keys)):` |
| B16 | 1562 | comprehension | ` for key in keys` |
| B17 | 1564 | Try | `try:` |
| B18 | 1566 | ExceptHandler | `except ValueError as exc:` |
| B19 | 1569 | If | `if revision == 'current':` |
| B20 | 1570 | If | `if revision_ref:` |
| B21 | 1573 | BoolOp | `prefetched is not None and relative in prefetched` |
| B22 | 1573 | IfExp | `prefetched[relative] if prefetched is not None and relative in prefetched else _committed_bytes(root, revision_ref, rela` |
| B23 | 1575 | If | `if blob is None:` |
| B24 | 1577 | If | `if hashlib.sha256(blob).hexdigest() != value['source_sha256']:` |
| B25 | 1579 | If | `if not source.is_file():` |
| B26 | 1581 | If | `if hashlib.sha256(source.read_bytes()).hexdigest() != value['source_sha256']:` |
| B27 | 1583 | If | `if revision != 'base':` |
| B28 | 1587 | For | `for section in MAP_SECTIONS:` |
| B29 | 1588 | If | `if section not in logic:` |
| B30 | 1590 | BoolOp | `relative not in logic or function not in logic` |
| B31 | 1590 | If | `if relative not in logic or function not in logic:` |
| B32 | 1594 | BoolOp | `value.get('branches') or []` |
| B33 | 1596 | If | `if '# Branch Test Map' not in branch_map:` |
| B34 | 1599 | If | `if len(mapped) != len(set(mapped)):` |
| B35 | 1601 | comprehension | ` for branch in branches if isinstance(branch, dict) and branch.get('id')` |
| B36 | 1604 | BoolOp | `isinstance(branch, dict) and branch.get('id')` |
| B37 | 1606 | BoolOp | `branches and len(expected) != len(branches)` |
| B38 | 1606 | If | `if branches and len(expected) != len(branches):` |
| B39 | 1609 | If | `if missing:` |
| B40 | 1616 | IfExp | `expected if branches else {'B1'}` |
| B41 | 1617 | If | `if unexpected:` |
| B42 | 1634 | IfExp | `test_index(root) if index is None else index` |
| B43 | 1637 | BoolOp | `not branches and 'B1' not in mapped` |
| B44 | 1637 | If | `if not branches and 'B1' not in mapped:` |
| B45 | 1640 | BoolOp | `'# Risk Pattern Report' not in risk or relative not in risk` |
| B46 | 1640 | If | `if '# Risk Pattern Report' not in risk or relative not in risk:` |

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1763-1910` · 분기 49 · 반환 5 · raise 0 (편집 전 L1508-1642 · 분기 46 · 반환 5 · raise 0, `ast.before-7.5.2.3.json` = revision `1d12520c`).

산문이 UTF-8 이 아니면 **그 대상의** 판정 줄이다(옛 판본은 올려 보내 판정 전체를 대상 이름 없는 한 줄로 만들었다). 워킹트리 소스도 깔때기로 읽고, **없는 것과 못 읽는 것을 가른다** — 심링크 고리는 "없다" 가 아니라 "못 읽는다" 다(7.5.2.2 의 문장을 정정한다).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1779 | For | `for name in REQUIRED:` |
| B2 | 1781 | If | `if name == 'ast.json':` |
| B3 | 1782 | If | `if path not in held:` |
| B4 | 1785 | If | `if held[path] is None:` |
| B5 | 1788 | Try | `try:` |
| B6 | 1790 | ExceptHandler | `except ValueError:` |
| B7 | 1796 | Try | `try:` |
| B8 | 1798 | ExceptHandler | `except FileNotFoundError:` |
| B9 | 1801 | ExceptHandler | `except OSError:` |
| B10 | 1804 | Try | `try:` |
| B11 | 1806 | ExceptHandler | `except UnicodeDecodeError:` |
| B12 | 1811 | If | `if 'TODO' in texts[name]:` |
| B13 | 1813 | If | `if 'ast.json' not in texts:` |
| B14 | 1818 | If | `if problem == 'invalid':` |
| B15 | 1821 | BoolOp | `problem or any((not value.get(key) for key in keys))` |
| B16 | 1821 | If | `if problem or any((not value.get(key) for key in keys)):` |
| B17 | 1821 | comprehension | ` for key in keys` |
| B18 | 1823 | Try | `try:` |
| B19 | 1825 | ExceptHandler | `except ValueError as exc:` |
| B20 | 1828 | If | `if revision == 'current':` |
| B21 | 1829 | If | `if revision_ref:` |
| B22 | 1832 | BoolOp | `prefetched is not None and relative in prefetched` |
| B23 | 1832 | IfExp | `prefetched[relative] if prefetched is not None and relative in prefetched else _committed_bytes(root, revision_ref, rela` |
| B24 | 1834 | If | `if blob is None:` |
| B25 | 1836 | If | `if hashlib.sha256(blob).hexdigest() != value['source_sha256']:` |
| B26 | 1842 | Try | `try:` |
| B27 | 1844 | ExceptHandler | `except (FileNotFoundError, NotADirectoryError):` |
| B28 | 1846 | ExceptHandler | `except OSError as exc:` |
| B29 | 1849 | If | `if hashlib.sha256(blob).hexdigest() != value['source_sha256']:` |
| B30 | 1851 | If | `if revision != 'base':` |
| B31 | 1855 | For | `for section in MAP_SECTIONS:` |
| B32 | 1856 | If | `if section not in logic:` |
| B33 | 1858 | BoolOp | `relative not in logic or function not in logic` |
| B34 | 1858 | If | `if relative not in logic or function not in logic:` |
| B35 | 1862 | BoolOp | `value.get('branches') or []` |
| B36 | 1864 | If | `if '# Branch Test Map' not in branch_map:` |
| B37 | 1867 | If | `if len(mapped) != len(set(mapped)):` |
| B38 | 1869 | comprehension | ` for branch in branches if isinstance(branch, dict) and branch.get('id')` |
| B39 | 1872 | BoolOp | `isinstance(branch, dict) and branch.get('id')` |
| B40 | 1874 | BoolOp | `branches and len(expected) != len(branches)` |
| B41 | 1874 | If | `if branches and len(expected) != len(branches):` |
| B42 | 1877 | If | `if missing:` |
| B43 | 1884 | IfExp | `expected if branches else {'B1'}` |
| B44 | 1885 | If | `if unexpected:` |
| B45 | 1902 | IfExp | `test_index(root) if index is None else index` |
| B46 | 1905 | BoolOp | `not branches and 'B1' not in mapped` |
| B47 | 1905 | If | `if not branches and 'B1' not in mapped:` |
| B48 | 1908 | BoolOp | `'# Risk Pattern Report' not in risk or relative not in risk` |
| B49 | 1908 | If | `if '# Risk Pattern Report' not in risk or relative not in risk:` |

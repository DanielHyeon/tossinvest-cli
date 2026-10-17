# Function Logic Map: `_listing_outcome` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:805-812` · 분기 5 · 반환 2 · raise 0 (새 함수).

새 함수. 디렉터리 목록의 지문 — 이름과 **종류**를 같이 넣는다(같은 이름의 파일↔폴더 교체도 변화다).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 807 | Try | `try:` |
| B2 | 808 | comprehension | ` for child in path.iterdir()` |
| B3 | 809 | ExceptHandler | `except OSError as exc:` |
| B4 | 811 | IfExp | `'d' if is_dir else 'f'` |
| B5 | 811 | comprehension | ` for name, is_dir in entries` |

## task 7.5.10 — 목록 지문은 단사다 (7.5.2.3 재리뷰 정확성 — 충돌쌍, 2026-09-26)

> 편집 전 `ast.before-7510.json`(revision `1d1e5ca7`) · 편집 후 `ast.after-7510.json`, `759_flm_rows.py` 로 정렬.

편집 후 `tools/logic-map/check_analysis.py:1392-1409` · 분기 6 · 반환 2 · raise 0 · 호출 10 (`ast.after-7510.json`, source sha `da96c341fb2a`)
편집 전 `tools/logic-map/check_analysis.py:1392-1399` · 분기 5 · 반환 2 · raise 0 · 호출 8 (`ast.before-7510.json`, revision `1d1e5ca7`, source sha `b28398e26f3d`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 1401 | Try | `try:` | 같음 |
| B2 | B2 | 1402 | comprehension | `for child in path.iterdir()` | 같음 |
| B3 | B3 | 1403 | ExceptHandler | `except OSError as exc:` | 같음 |
| B4 | — | — | IfExp | `'d' if is_dir else 'f'` (옛) | **빠짐** |
| — | B4 | 1405 | comprehension | `for raw, is_dir in ((name.encode('utf-8'), is_dir) for name, is_dir in entries)` | **새** |
| — | B5 | 1406 | IfExp | `b'd' if is_dir else b'f'` | **새** |
| B5 | B6 | 1407 | comprehension | `for name, is_dir in entries` | 번호만 |

**결함.** 옛 지문은 `이름\t{d|f}` 를 `\n` 으로 이었다. 탭 · 개행은 POSIX 이름에 합법이라 `{a(f), b(f)}` 와 `{"a\tf\nb"(f)}` ·
`{a(f), b(d)}` 와 `{"a\tf\nb"(d)}` 가 같은 글자를 냈다(RED: `test_the_known_collisions_have_different_fingerprints` 두 subTest ·
원장 수준 `test_a_listing_that_becomes_its_collision_while_judged_asks_for_a_rerun` 이 `''` — 재확인이 못 봤다. 영수증
`analysis/harness/7510_collide.py` 가 `1d1e5ca7` 에서 `COLLIDE`).

**수리.** 항목마다 `이름 길이(8바이트, big-endian) · 이름 바이트 · 종류 한 글자` 를 잇는다(B4 · B5). 길이를 먼저 읽으면 이름이
어디서 끝나는지가 정해지므로 푸는 길이 하나다 — 이름 안의 어떤 글자에도(NUL 이 못 든다는 사실에도) 기대지 않는다. 이름을 바이트로
바꾸는 규칙(엄격한 UTF-8)은 옛 판본 그대로다 — 홀로 선 서로게이트 이름이 결함(`UnicodeEncodeError`, `GATE_FAULTS` 안)이 되던 동작을
바꾸지 않았다. 셋째 충돌쌍(`{a, b}` 대 `{afb}`)은 **구분자 없이 잇는** 판본의 것이다 — 길이 접두가 빠지면 빨갛다(편집 전 인코딩에서는
초록 · 못 박기용).

## task 7.5.11 — 종류를 못 물은 항목은 이름 댄 결함이다 (7.5.2.3 재리뷰 정확성 · 시험품질, 2026-09-26)

> 편집 전 `ast.before-7511.json` = `ast.after-7510.json`(같은 바이트 — 7.5.10 직후의 워킹트리) · 편집 후 `ast.after-7511.json`.
> 중간판(`check_analysis.py` source sha `ff6142a55df0` — **최종이 아니다**, 보수 절 정정) 에서 `:1404-1434`(7.5.15 가 앞에 두 줄을 더했다).

편집 후 `tools/logic-map/check_analysis.py:1402-1432` · 분기 8 · 반환 2 · raise 1 · 호출 14 (`ast.after-7511.json`, source sha `b088352ac906`)
편집 전 `tools/logic-map/check_analysis.py:1392-1409` · 분기 6 · 반환 2 · raise 0 · 호출 10 (`ast.before-7511.json`, revision `worktree`, source sha `da96c341fb2a`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 1416 | Try | `try:` | 같음 |
| B2 | — | — | comprehension | `for child in path.iterdir()` (옛) | **빠짐** |
| — | B2 | 1418 | For | `for child in path.iterdir():` | **새** |
| — | B3 | 1419 | Try | `try:` | **새** |
| — | B4 | 1421 | ExceptHandler | `except OSError as exc:` | **새** |
| B3 | B5 | 1426 | ExceptHandler | `except OSError as exc:` | 번호만 |
| B4 | B6 | 1428 | comprehension | `for raw, is_dir in ((name.encode('utf-8'), is_dir) for name, is_dir in entries)` | 번호만 |
| B5 | B7 | 1429 | IfExp | `b'd' if is_dir else b'f'` | 번호만 |
| B6 | B8 | 1430 | comprehension | `for name, is_dir in entries` | 번호만 |

**판본 영수증 먼저**(`analysis/harness/7511_isdir.py`, 2026-09-26, euid 1000): `Path.is_dir()` 은 3.11.15 · **3.12.3(이 저장소의
`/usr/bin/python3`)** · 3.13.13 에서 끊긴 심링크 · 심링크 고리 · 사라진 이름을 `False` 로 삼키고 권한 오류(`EACCES`)는 올린다. **3.14.5 는
권한 오류까지 `False`** 다. `rglob` 은 네 판본 모두 못 읽는 하위 트리를 조용히 건너뛴다(그 절반은 7.5.9 가 순회를 없애며 닫았다).

**결함.** 증거 디렉터리의 끊긴 심링크 · 고리가 "파일" 로 분류되어 번들 목록에서 조용히 빠졌다(RED: `test_a_dangling_link_among_the_bundles_is_named`
· `…link_loop…` 이 `[]`). 번들 **안**에서는 목록이 "파일" 로 넣고 `_bundle_text` 가 못 여는 것을 사라진 파일로 건너뛰어 역시 `[]`
(`test_a_dangling_link_inside_a_bundle_names_that_bundle` — 조용한 건너뛰기 둘이 겹친 자리, 이 로트에서 재서 알았다).

**수리.** 종류를 `os.stat` 으로 직접 묻는다(B2~B4, 따라가서 디렉터리면 디렉터리 — `is_dir()` 과 같은 뜻, 양성 대조
`test_a_link_to_a_bundle_is_still_a_bundle`). 못 물으면 새 `UnstatableEntry(OSError)` — 이름과 사유를 담고 목록 **전체**가 실패다.
`FileNotFoundError` 로 두지 않는 까닭: `_read_evidence` 는 목록의 `FileNotFoundError` · `NotADirectoryError` 를 "증거 디렉터리가 없다" 로
읽는다 — 끊긴 링크 하나가 디렉터리 전체를 면제 경로로 만든다. 별도 타입이라 그 `except` 에 안 걸리고, 증거 디렉터리에서는
`cannot derive modified Go functions: … cannot tell what \`ghost\` is: …` 로, 번들 안에서는 그 번들의 `unlistable` 줄로 나온다.

**거부하는 정상 입력(편집 전에 셌다, HEAD `1d1e5ca7`).** 목록 깔때기가 여는 디렉터리(활성 · 아카이브의 `analysis/function-logic` 과 그
번들) **3,183** · 항목 **15,428** · 심링크 **0** · `os.stat` 실패 **0** → 새 거절 0.

## 보수 — 최종 파일로 다시 열거 (독립 주장정확성 리뷰 F11, 2026-09-26)

최종 `tools/logic-map/check_analysis.py:1416-1446` · 분기 8 · 반환 2 · raise 1 · 호출 14 · source sha `ff590b79db6e` (비교 기준 `ast.after-7511.json`, 그 판 `b088352ac906` `:1402-1432`).
위 절들이 "최종" 이라 적은 sha 는 **중간판**이었다(리뷰 지적 — 표기를 고쳤다). 최종 파일에서 `ast.after-759r.json` 을 다시 뽑아 앞 절의 편집 후 열거와 대조했다 — 분기(종류 · 소스) · 반환 · raise · 호출 수가 **같다**(구조 동일). 바뀐 것은 sha 와 줄 좌표다.

## task 7.5.38 덤 — 목록 뒤 사라진 항목 (2026-09-27)

> 편집 전 `ast.before-7538.json`(HEAD `5406cac1`) · 편집 후 `ast.after-7538.json`(최종 `check_analysis.py` `463982f42151`), `759_flm_rows.py` 로 정렬.

편집 후 `tools/logic-map/check_analysis.py:1474-1509` · 분기 10 · 반환 2 · raise 2 · 호출 17 (`ast.after-7538.json`, source sha `463982f42151`)
편집 전 `tools/logic-map/check_analysis.py:1465-1495` · 분기 8 · 반환 2 · raise 1 · 호출 14 (`ast.before-7538.json`, revision `5406cac1`, source sha `8d2a3338fdab`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 1488 | Try | `try:` | 같음 |
| B2 | B2 | 1490 | For | `for child in path.iterdir():` | 같음 |
| B3 | B3 | 1491 | Try | `try:` | 같음 |
| B4 | B4 | 1493 | ExceptHandler | `except OSError as exc:` | 같음 |
| — | B5 | 1495 | BoolOp | `exc.errno == errno.ENOENT and (not os.path.lexists(child))` | **새** |
| — | B6 | 1495 | If | `if exc.errno == errno.ENOENT and (not os.path.lexists(child)):` | **새** |
| B5 | B7 | 1503 | ExceptHandler | `except OSError as exc:` | 번호만 |
| B6 | B8 | 1505 | comprehension | `for raw, is_dir in ((name.encode('utf-8'), is_dir) for name, is_dir in entries)` | 번호만 |
| B7 | B9 | 1506 | IfExp | `b'd' if is_dir else b'f'` | 번호만 |
| B8 | B10 | 1507 | comprehension | `for name, is_dir in entries` | 번호만 |

새 B5 · B6: `stat` 이 ENOENT 이고 `lstat` 도 없으면(`os.path.lexists` 거짓) 목록 뒤에 **사라진** 것 — 새 `ListingMoved(OSError)` 로 "`<이름>` disappeared while the directory
was being listed — run it again". 끊긴 링크(있는데 못 묻는 것)는 그대로 `UnstatableEntry`(양성 대조 시험 · 변이 MW6). **빈도 논거(실측, 2026-09-27)**: 스크래치 디렉터리에서
vim(`-u NONE -es`)으로 파일을 2,176 번 쓰는 동안 옆 스레드가 목록 → stat 을 돌려 목록 뒤 stat 이 ENOENT 인 경합 **1,106** 회(이름 `.notes.md.swp` · `.swx` — 이 설정의
vim 은 `4913` 대신 스왑 파일을 만들었다). 게이트가 증거 디렉터리를 여는 순간과 겹칠 확률은 작지만 편집 중 흔한 흐름이다. **어디서 달라지나**: 증거 디렉터리(맨 위)의
목록이면 판정 **앞**이라 재확인이 안 돌아서 편집 전에는 `cannot tell what \`4913\` is` 결함이었다 → 이제 "disappeared … run it again". 번들 **안**의 같은 경합은
편집 전에도 끝의 재확인이 "changed … run it again" 으로 덮었다(이 세션 실측). 둘 다 rc 1 — 바뀐 것은 문장이다.

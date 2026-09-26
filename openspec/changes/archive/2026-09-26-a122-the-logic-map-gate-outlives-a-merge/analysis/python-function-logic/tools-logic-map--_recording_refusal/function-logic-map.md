# Function Logic Map: `_recording_refusal`

`tools/logic-map/check_analysis.py:1202-1273` · Python · **task 7.7 (리뷰 I8)** ·
분기 9 · 반환 9 · raise 0 (`ast.after-7.7.json`)

## Inputs and invariants

요청받은 id `change`, 해소기가 이미 찾은 `change_dir`, `root`. `(사유, base)` 를 돌려준다 — 멈추면 base 는
빈칸이고, 안 멈추면 사유가 빈칸이고 base 는 `resolve_base` 가 돌려준 **유효 base** 다(기록 명령이 그 값으로 걷는다).
사유는 `{change}: ` 를 **안 붙인** 문장이다. 기록 명령은 앞에 붙여 찍고, 조언 줄은 그대로 인용한다.

**걷지 않는다.** 후보 순회는 `compute_landing` 에 남는다. 조언 줄은 워킹트리가 대상인 모든 5단계 실행에서
이 함수를 부르는데, 7.7 전수 계측에서 기록 명령 한 번이 a055 225.1초였다 — 순회 없이 걷기 전
부분에서 멈추는 번들 0 change 들은 1.0~2.1초였으므로 차이는 거의 전부 순회다.

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L1222 | If | `if landing_file.is_symlink():` |
| B2 | L1232 | If | `if _landing_record(change_dir, root) is not None:` |
| B3 | L1234 | If | `if landing_file.exists():` |
| B4 | L1244 | If | `if (change_dir / 'analysis' / 'function-logic-reference.txt').exists():` |
| B5 | L1250 | Try | `try:` |
| B6 | L1252 | ExceptHandler | `except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:` |
| B7 | L1254 | If | `if facts.get('execution_baseline_adoption'):` |
| B8 | L1257 | If | `if why:` |
| B9 | L1267 | If | `if dirty.returncode:` |

| 반환 줄 | 값 |
|---|---|
| L1227 | `(f'`{LANDING_FILE}` is a symlink — not followed: the record must be a regular file in the change directory, or the value the gate reads back is not th` |
| L1233 | `(f'`{LANDING_FILE}` already exists — not overwritten', '')` |
| L1240 | `(f'`{relative}` already exists on disk but not in HEAD — not overwritten: the gate reads the record from the commit, so commit it', '')` |
| L1245 | `('this change borrows its evidence — copy the landing recorded on the change that owns the bundles instead of computing a second one', '')` |
| L1253 | `(f'cannot resolve the comparison base: {exc}', '')` |
| L1255 | `(ADOPTION_REFUSES_A_LANDING, '')` |
| L1258 | `(f'no landing recorded — {why}', '')` |
| L1268 | `('the working tree has uncommitted changes to tracked files — commit them first, because a recorded landing points at a commit and step 5 would then n` |
| L1273 | `('', base)` |

**순서가 조언이다.** 조언 줄은 사유를 하나만 말하므로 먼저 걸리는 사유가 사람이 듣는 사유다.
기록 명령이 쓰던 순서(심링크 → 기록 존재 → 빌림 → base → 이관 → dirty → 번들 0 → 하한 없음)에서
dirty(B9) **하나만** 맨 뒤로 옮겼다. dirty 는 커밋하면 사라지는 유일한 사유라서, 앞에 두면 영원히 기록할 수
없는 change 가 "먼저 커밋하라"를 듣고 커밋한 뒤에야 진짜 사유를 듣는다. 이 저장소의 활성 change 중 번들 0 이
열둘이다(7.7 전수 계측). 이 순서를 재는 시험이 첫 판에 0 이었다(변이 R10 SURVIVED) — 지금은
`test_a_refusal_that_outlives_a_commit_is_named_before_one_that_does_not` 가 R10·R22 를 잡는다.

B2·B3 을 가른 이유: 옛 판본은 `exists() or _landing_record(...)` 한 조건에 "already exists" 한 문장이었다.
디스크에만 있는 기록(커밋 전 · staged 아카이브 이동)에서 그 문장과 "(no landed-commit.txt)" 가 고리를 이뤘다.
B3 은 경로를 적는다 — staged 아카이브 이동에서는 HEAD 의 **옛 자리**에 기록이 있어 "HEAD 에 없다"만으로는 틀리다.

## Calls and live bindings

`landing_file.is_symlink` · `_landing_record` · `landing_file.exists` · `landing_file.relative_to` ·
`(change_dir / 'analysis' / 'function-logic-reference.txt').exists` · `resolve_base` · `_walk_floor` · `subprocess.run`
(`git diff --quiet HEAD`, timeout 30). 부르는 쪽 둘: `record_landing`(결함 경계 `GATE_FAULTS` 안) ·
`main` 의 조언 권유 갈래(자기 `GATE_FAULTS` 안). 구조 시험
`test_the_recorder_and_its_advice_ask_one_judge` 가 둘 다 이 함수를 부르고 기록 명령이 거절 조각을 직접 안 부르는지 본다.

## State mutations and fallbacks

쓰기 없음. `resolve_base` 가 넘긴 `facts` 는 이 함수 지역이다. 결함은 두 종류로 나간다: base 해소 실패는
B6 이 **사유 문장**으로 받고, 나머지(`_walk_floor` 의 `ValueError`, git 의 `TimeoutExpired`)는 호출자 경계로 올린다.

## Safety conclusion

기록 명령이 **무엇을 거절하는지**는 순서 하나 말고는 안 바뀌었다 — 옮긴 거절 여섯의 조건과 문장(`{change}: ` 뒤)은
글자 그대로이고, 기록이 쓰이는 입력은 그대로다(기록 쓰기는 여전히 `open(…, "xb")`). 바뀐 것: (1) dirty 가 걷기 전
하한 사유 뒤로 가서 번들 0·하한 없음·dirty 가 겹친 입력의 **문장**이 바뀐다(rc 는 둘 다 1), (2) 디스크에만 있는
기록의 문장이 경로와 `not in HEAD` 를 말한다, (3) 조언 줄이 이 함수의 사유를 인용한다. 판정(`check`)은 이 함수를
부르지 않는다 — 조언이 판정 줄을 바꿀 길이 없다.

## 편집 — task 7.2.3 (빌리는 change 는 창을 좁히지 않는다) · 분기 9 → 9 · 반환 9 → 9 · raise 0 → 0

편집 전 `ast.before-7.2.3.json`(HEAD `67d06bc9` blob, L1237-1308) · 편집 후 `ast.after-7.2.3.json`(워킹트리, L1242-1312). 표는 두 열거의 `source` 를 스크립트가 대조한 것이다.

빌림 갈래의 문장을 모듈 상수 `BORROWED_REFUSES_A_LANDING` 으로 — 5단계와 **같은 문장**. 예전 문장("copy the landing recorded on the change that owns the bundles")은 리뷰 C4 의 구멍을 권했다. 분기 · 순서 변화 없음.

| | 종류 | 소스 |
|---|---|---|
| 사라짐 | 반환 | `('this change borrows its evidence — copy the landing recorded on the change that owns the bundles instead of computing a second one', '')` |
| 생김 | 반환 | `(BORROWED_REFUSES_A_LANDING, '')` |

호출 — 사라짐 없음 · 생김 없음

## 편집 — task 7.2.6 (이미 있는 기록도 길을 말한다) · 분기 9 → 9 · 반환 9 → 9 · raise 0 → 0

편집 전 `ast.before-7.2.6.json`(HEAD `e9f86820` blob) · 편집 후 `ast.after-7.2.6.json`(워킹트리, L1353-1425).

"이미 있다 — 덮어쓰지 않았다" 뒤에 `LANDING_RECOVERY` 를 잇는다. 덮어쓰기 거절 **동작은 그대로다**. 디스크에만 있는 기록의 문장(`… already exists on disk but not in HEAD`)은 안 바꿨다 — 거기서 할 일은 커밋이지 재기록이 아니고, 두 문장이 갈려 있어야 한다는 것을 7.7 의 시험이 이미 못 박고 있다.

| | 종류 | 소스 |
|---|---|---|
| 사라짐 | 반환 | `(f'`{LANDING_FILE}` already exists — not overwritten', '')` |
| 생김 | 반환 | `(f'`{LANDING_FILE}` already exists — not overwritten; {LANDING_RECOVERY}', '')` |

## task 7.5 — 번들 목록을 한 번 재서 넘긴다

`ast.before-7.5.json`(= `8091e6c4` 의 소스)과 `ast.after-7.5.json` 을 같은 열거기로 뽑아
**순서 있는 배열로** 대조했다: 분기 **순서열 9개가 바이트 동일** · 반환 **순서열 9개가 바이트 동일**.

> **처음에는 개수로 적었다가 정정했다** (2026-09-18 독립 리뷰 P2). `분기 N · 반환 M` 은
> multiset 크기라 **순열에 불변**이다 — 가드를 서로 바꿔도 같은 수가 나오므로 "순서가 안
> 움직였다"를 못 받친다. 순서 있는 배열은 이미 같은 JSON 안에 있었고, 그것을 인용해야 한다
> ([[a-new-guard-unpins-the-guards-behind-it]] 가 지키려는 것이 바로 순서다).

바뀐 것은 고정 번들 목록이 **어디서 오는가** 하나다. 후보마다 디렉터리를 다시 순회하던
것을 호출자가 한 번 재서 넘긴다 — `floor` 와 `repairs` 가 이미 그렇게 넘어오고 있었고
(7.2.6 · 7.6), 7.5 는 셋째를 같은 방식으로 묶었다. 번들은 걷는 동안 안 변한다: 이 도구는
번들을 **읽기만** 한다.

값: 2026-09-18 프로파일에서 `_pinning_bundles` 가 a071 walk 하나에 341회 돌아
13.07s 중 9.30s 였다. 묶은 뒤 a071 이 10.97s → **3.69s**.

## task 7.5.1 — gstack 리뷰의 permissive 결함 수리

`ast.before-7.5.1.json`(= git revision **`b29e1f4e`**, 소스 해시가 그 커밋과 일치)과
`ast.after-7.5.1.json` 을 같은 열거기로 뽑아 **순서 있는 배열**을 difflib 으로 정렬했다
(분기 9 · 반환 9 · raise 0 → 분기 9 · 반환 9 · raise 0).

손 복사 예외 목록을 `GATE_FAULTS` 로(서브에이전트) — `SubprocessError` 가 빠져 타임아웃이 창 줄을 삼켰다.

| | 종류 | 소스 |
|---|---|---|
| − | ExceptHandler | `except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:` |
| + | ExceptHandler | `except GATE_FAULTS as exc:` |

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`_walk_floor` 의 반환이 `(하한, 사유)` 둘로 줄어 언팩 한 줄(`_, why, _ =` → `_, why =`)만 바뀌었다.
편집 전후 AST(`ast.before-7.5.2.json` = revision `e9f905bd` · `ast.after-7.5.2.json`)의 분기·반환·raise
**순서열 차이 0** — 거절 일곱의 순서와 문장은 그대로다. **GREEN 도중 편집 집합에 들어왔다**(review.md).

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1918-2002` · 분기 10 · 반환 9 · raise 1 (편집 전 L1849-1921 · 분기 9 · 반환 9 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`head` · `evidence` 를 받는다 — 기록 명령에서 걷기 전 판정과 걷기가 같은 한 벌. "깨끗한가" 는 `head` 와 비교하고, git 이 0 · 1 말고 답하면 결함(옛: rc 128 을 "먼저 커밋하라" 로).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1943 | If | `if landing_file.is_symlink():` |
| B2 | 1953 | If | `if _landing_record(change_dir, root, head) is not None:` |
| B3 | 1957 | If | `if landing_file.exists():` |
| B4 | 1967 | If | `if (change_dir / 'analysis' / 'function-logic-reference.txt').exists():` |
| B5 | 1972 | Try | `try:` |
| B6 | 1974 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B7 | 1976 | If | `if facts.get('execution_baseline_adoption'):` |
| B8 | 1979 | If | `if why:` |
| B9 | 1989 | If | `if dirty.returncode not in (0, 1):` |
| B10 | 1996 | If | `if dirty.returncode:` |

| raise 줄 | 소스 |
|---|---|
| 1992 | `raise RuntimeError(f'cannot tell whether the working tree matches {head[:12]}: ' + _first_` |

## task 6.4(b) — `resolve_base` 에 `head` 를 넘긴다 (2026-09-26)

`ast.before-6.4.json`(revision `0023fd12`) → `ast.after-6.4.json`. 분기 10 → 10 · 반환 9 → 9 · raise 1 → 1 · 호출 13 → 13, 갈래 변화 0.
바뀐 것은 `resolve_base(..., head=head)` 인자 하나다 — 이 함수가 이미 받은 sha 다. base 의 새 거절(모양 · HEAD 대조 · 태그)은 기존
`cannot resolve the comparison base: …` 사유로 나간다(시험 `test_an_uncommitted_edit_cannot_move_a_committed_base` 가 기록 경로에서 잰다).

## task 7.5.14 — 깔때기 밖 stat 둘을 깔때기로 (2026-09-27)

> 편집 전 `ast.before-7514.json`(HEAD `3bb11b2d`) · 편집 후 `ast.after-7514.json`, `759_flm_rows.py` 로 정렬.

편집 후 `tools/logic-map/check_analysis.py:3160-3256` · 분기 12 · 반환 10 · raise 1 · 호출 15 (`ast.after-7514.json`, source sha `c74f69b35ae6`)
편집 전 `tools/logic-map/check_analysis.py:3160-3244` · 분기 10 · 반환 9 · raise 1 · 호출 13 (`ast.before-7514.json`, revision `3bb11b2d`, source sha `12beb79433de`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 3187 | If | `if landing_file.is_symlink():` | 같음 |
| B2 | B2 | 3197 | If | `if _landing_record(change_dir, root, head) is not None:` | 같음 |
| B3 | — | — | If | `if landing_file.exists():` (옛) | **빠짐** |
| B4 | — | — | If | `if (change_dir / 'analysis' / 'function-logic-reference.txt').exists():` (옛) | **빠짐** |
| — | B3 | 3203 | If | `if _kind(landing_file):` | **새** |
| — | B4 | 3215 | Try | `try:` | **새** |
| — | B5 | 3217 | ExceptHandler | `except FileNotFoundError:` | **새** |
| — | B6 | 3219 | ExceptHandler | `except OSError as exc:` | **새** |
| B5 | B7 | 3226 | Try | `try:` | 번호만 |
| B6 | B8 | 3228 | ExceptHandler | `except GATE_FAULTS as exc:` | 번호만 |
| B7 | B9 | 3230 | If | `if facts.get('execution_baseline_adoption'):` | 번호만 |
| B8 | B10 | 3233 | If | `if why:` | 번호만 |
| B9 | B11 | 3243 | If | `if dirty.returncode not in (0, 1):` | 번호만 |
| B10 | B12 | 3250 | If | `if dirty.returncode:` | 번호만 |

**먼저 셌다(7.5.14 가 적은 "살아 있는 우회").** `changed_existing_functions` 의 `Path.exists` 는 **이미 없다**(7.5.34 가 그 함수를 다시 썼다 — HEAD `3bb11b2d`
AST 전수). 남은 것은 이 함수의 셋이다: `landing_file.is_symlink()` · `landing_file.exists()` · 빌림 표지의 `.exists()`.

- 옛 B3(`landing_file.exists()`) → 새 B3 `_kind(landing_file)` — 따라가서 묻는 뜻은 같고(끊긴 링크는 B1 이 먼저 멈춘다) 원장에 남는다. 못 물으면(권한)
  "" 로 "없다" 이고 쓰기(`open("xb")`)가 이름 대고 멈춘다.
- 옛 B4(빌림 표지 `.exists()`) → 새 B4~B6 `_read_regular` — 판정(`_judged`)과 **같은 깔때기 · 같은 갈래**: 없으면 빌림 아님, 못 읽으면(권한 · FIFO · 폴더)
  `UNREADABLE` 문장, 읽히면 빌림. RED(편집 전): 권한 없는 표지 · FIFO 표지에서 기록 명령이 **빌림 문장**을 냈다(판정은 `could not be read`).
- `landing_file.is_symlink()`(B1)는 **면제**로 남긴다 — 쓰기 자리의 모양(`lstat`)이고 판정 입력이 아니며, 쓰기 순간 `open("xb")` 가 같은 것을 다시 막는다.
  깔때기 `_kind` 는 따라가서 묻는 `os.stat` 이라 이 물음을 못 대신한다. 구조 시험이 이 호출 형태 **하나**만 면제한다.

**거부하는 정상 입력**: 저장소의 빌림 표지는 a073 하나이고 정규 파일이다(읽히면 옛 판본과 같은 문장). 새로 달라지는 것은 못 읽는 표지뿐 — 노출 0.

## task 7.5.14 수리 (적대 리뷰 뒤, 2026-09-27) — 주석만

`ast.after-7514r.json`(최종 `check_analysis.py` `8d2a3338fdab`, `:3160-3258`) — 분기 · 반환 · raise · 호출이 `ast.after-7514.json` 과 같다(스크립트 대조).
바뀐 것은 새 B3 옆 주석 하나: "못 물으면 쓰기(`open("xb")`)가 이름 대고 멈춘다" 는 **거짓**이었다(적대 리뷰 코드 읽기 · 이 세션 실측). 쓰기의 `PermissionError` 를
`record_landing` 은 잡지 않는다(`FileExistsError` 만) — CLI(`main`)의 `GATE_FAULTS` 경계가 `no landing recorded — [Errno 13] …` 로 이름 대고, 함수를 직접
부르면 예외다. 실측: change 디렉터리 `0555` → 직접 호출 `PermissionError`, CLI rc 1 + 그 문장. 열린 task 7.5.44. (x 없는 `0666` 이면 그 전에 증거 목록
`_listed(analysis)` 가 권한으로 실패해 이름 댄 줄이 된다.)

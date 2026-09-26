# Function Logic Map: `_self_repair_commits`

`tools/logic-map/check_analysis.py:601-691` · Python · **task 7.2.6 (H4)** ·
분기 15 · 반환 2 · raise 3 (`ast.after-7.2.6.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** 아래 표는 `enumerate.py` 의 열거를 그대로 옮긴 것이다.
> 7.2.6 이 새로 만든 함수라 편집 전 판본이 없다.

## Inputs and invariants

`(root, analysis)` 를 받아 **이 change 자신의 나중 Go 작업** 커밋들을 오래된 것부터 돌려준다.
세는 조건은 하나다: 그 change 의 디렉터리를 만지면서 Go 파일도 고친 **비병합** 커밋(HEAD 까지).
판정하지 않는다 — 어느 후보를 거절할지는 `_landing_refusal` 이 정한다. 예외를 만들지 않고,
못 세는 경우(저장소 밖 증거 · git 실패 · 만진 커밋 0)는 **빈 목록**이다.

`floor` 와 같은 모양으로 호출자가 **한 번** 재서 `_landing_refusal` 에 넘긴다. 두 호출자
(`resolve_landing` · `compute_landing`)가 각자 조건을 들고 있으면 갈린다
([[two-judgements-cover-for-each-other]]).

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L628 | Try | `try:` |
| B2 | L630 | ExceptHandler | `except ValueError as exc:` |
| B3 | L637 | If | `if change_dir.startswith(ARCHIVE_PREFIX):` |
| B4 | L641 | If | `if before:` |
| B5 | L654 | If | `if touching.returncode:` |
| B6 | L660 | IfExp | `touching.stderr.strip().splitlines()[0] if … else 'git log failed'` |
| B7 | L663 | If | `if not hashes:` |
| B8 | L679 | If | `if listing.returncode:` |
| B9 | L682 | IfExp | `listing.stderr.strip().splitlines()[0] if … else 'git log --no-walk failed'` |
| B10 | L685 | For | `for block in listing.stdout.split('\x00'):` |
| B11 | L686 | comprehension | ` for line in block.splitlines() if line` |
| B12 | L687 | BoolOp | `lines and any(name.endswith('.go') for name in lines[1:])` |
| B13 | L687 | If | `if lines and any(name.endswith('.go') for name in lines[1:]):` |
| B14 | L687 | comprehension | ` for name in lines[1:]` |
| B15 | L691 | comprehension | ` for commit in reversed(hashes) if commit in flagged` |

| 반환 줄 | 값 |
|---|---|
| L664 | `[]` — 그 디렉터리를 만진 비병합 커밋이 **없다**(정상) |
| L691 | `[commit for commit in reversed(hashes) if commit in flagged]` |

| raise 줄 | 무엇을 못 쟀나 |
|---|---|
| L635 | change 디렉터리를 저장소 안 경로로 못 적는다 (오늘 도달 불가) |
| L658 | 디렉터리를 만진 커밋 목록 조회 실패 |
| L680 | 그 커밋들의 파일 목록 조회 실패 |

## Calls and live bindings

`subprocess.run` ×2 (`git log`) · `_archived_change_id` · `ARCHIVE_PREFIX` ·
`analysis.parent.parent.relative_to` · 순수 문자열 처리. 시각·환경·워킹트리를 안 읽는다 —
HEAD 까지의 커밋 역사만 본다.

**판정은 사람의 git 설정의 함수가 아니다** — 세 가지를 명령줄에서 못 박는다
(2026-09-16 적대 리뷰 F1 · F4 · F5):

| 못 박은 것 | 안 박으면 |
|---|---|
| `--full-history` | 경로 조회의 기본 단순화가 병합 반대편 가지의 자기 수리를 **통째로 버린다**(실측 재현). 오늘 이 저장소 126건에서 더 세는 커밋은 **0** 이다 — 지금 넣으면 공짜다 |
| `-c diff.renames=false` | `.go` 를 비-`.go` 이름으로 옮긴 커밋의 옛 이름이 사라져 깃발이 안 선다 |
| `-c core.quotePath=false` | 비ASCII 이름이 인용돼 `.go` 로 안 끝난다 |

셋 다 **안 박으면 깃발이 덜 서는**(= 창이 안 넓어지는) 방향이다.

**왜 조회가 둘인가.** 경로 제한을 건 `git log` 는 **그 경로의 파일만** 적어 주므로
"이 커밋이 Go 도 고쳤나"를 같은 조회로 못 본다. 그래서 디렉터리를 만진 커밋을 먼저 고르고,
`--no-walk --stdin` 한 번으로 그 커밋들의 **전체** 목록을 읽는다. 커밋마다 `git diff-tree` 를
돌리는 측정 판본과 착지 있는 68건 전수에서 깃발 224개 · 불일치 0 이다(`h4b_ab_flagged.py`).

## State mutations and fallbacks

없다. 읽기만 한다. git 이 **실패하면 판정이다**(`RuntimeError` → 경계가 오류 줄로 바꾼다).
빈 목록으로 물러나지 않는다 — 빈 목록은 "거절할 것이 없다"라서 실패가 **위반 0** 으로 읽히고
가드가 조용히 꺼진다([[missing-tool-reports-clean]], 2026-09-16 적대 리뷰 F2: 주입 실험에서
rc 1 하나로 거절돼야 할 입력이 초록이었다). 옆의 `_evidence_floor` 도 실패하면 거절로 간다 —
두 자리가 같은 방향이어야 한다. **빈 목록은 하나뿐이다**: 그 디렉터리를 만진 커밋이 정말 없을 때.

## Safety conclusion

생산 Go 코드 변경 0. 신원(`openspec/changes/<id>/`)을 읽지만 **거절에만** 쓴다 — 1.12 가 막은
것은 신원으로 착지를 받아들이는 판정이다. 오늘 저장소에서 기록을 가진 change 는 a099 하나이고
네 시점 전부 값이 같다(2026-09-16 전수).

**한계 둘** [[fail-closed-must-name-what-it-rejects]]:
- Go 수리와 문서를 **다른 커밋**으로 쪼개면 깃발이 안 선다 — 망각 가드이지 위조 가드가 아니다.
- 병합 커밋 **자신의** 변경은 안 읽힌다(`--no-merges`, 그리고 `git log` 의 기본이 병합 diff 를
  안 낸다 — git 2.43 에서 어떤 설정으로도 안 바뀐다). 7.2.4 의 H7 과 같은 부류다.

두 한계는 **놓치는** 쪽이므로 창이 넓어지지 않는다 — 그 자리는 이 규칙 이전 상태로 남는다.
거절하는 쪽으로 틀리면(넓은 교차 커밋) 창이 넓어진다. 방향이 둘이고 다르다는 것을 spec 이 적는다.

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:868-958` · 분기 15 · 반환 2 · raise 3 (편집 전 L817-907 · 분기 15 · 반환 2 · raise 3, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

상징 `HEAD` 대신 명령이 **한 번** 푼 sha(`head`)를 받는다. 7.5.2 는 `HEAD` 를 열 자리에서 따로 읽고 지문에 표본 하나만 넣었다 — 기록은 표본 **전에** 읽혔고 뒤의 git 호출은 살아 있는 `HEAD` 를 다시 읽어서 가지 전환 한 번(레드팀 repro_b)도, 떠났다 돌아온 `HEAD`(이 세션이 재현)도 rc 0 이었다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 895 | Try | `try:` |
| B2 | 897 | ExceptHandler | `except ValueError as exc:` |
| B3 | 904 | If | `if change_dir.startswith(ARCHIVE_PREFIX):` |
| B4 | 908 | If | `if before:` |
| B5 | 921 | If | `if touching.returncode:` |
| B6 | 927 | IfExp | `touching.stderr.strip().splitlines()[0] if touching.stderr.strip() else 'git log failed'` |
| B7 | 930 | If | `if not hashes:` |
| B8 | 946 | If | `if listing.returncode:` |
| B9 | 949 | IfExp | `listing.stderr.strip().splitlines()[0] if listing.stderr.strip() else 'git log --no-walk failed'` |
| B10 | 952 | For | `for block in listing.stdout.split('\x00'):` |
| B11 | 953 | comprehension | ` for line in block.splitlines() if line` |
| B12 | 954 | BoolOp | `lines and any((name.endswith('.go') for name in lines[1:]))` |
| B13 | 954 | If | `if lines and any((name.endswith('.go') for name in lines[1:])):` |
| B14 | 954 | comprehension | ` for name in lines[1:]` |
| B15 | 958 | comprehension | ` for commit in reversed(hashes) if commit in flagged` |

| raise 줄 | 소스 |
|---|---|
| 902 | `raise RuntimeError(f"cannot name this change's directory under the repository: {exc}") fro` |
| 925 | `raise RuntimeError(f'cannot list the commits that touched {change_dir}: {(touching.stderr.` |
| 947 | `raise RuntimeError(f'cannot read the files those commits changed: {(listing.stderr.strip()` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:901-991` · 분기 13 · 반환 2 · raise 3 (편집 전 L868-958 · 분기 15 · 반환 2 · raise 3, `ast.before-7.5.2.2.json` = revision `908a8a36`).

git 의 말 첫 줄을 `_first_line` 한 벌로(손으로 적은 사본 둘 — 160자 자름이 없었다). 낡은 주석: "`_evidence_floor` 는 실패하면 거절로 간다" → 7.5.2.1 부터 결함이다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 928 | Try | `try:` |
| B2 | 930 | ExceptHandler | `except ValueError as exc:` |
| B3 | 937 | If | `if change_dir.startswith(ARCHIVE_PREFIX):` |
| B4 | 941 | If | `if before:` |
| B5 | 954 | If | `if touching.returncode:` |
| B6 | 963 | If | `if not hashes:` |
| B7 | 979 | If | `if listing.returncode:` |
| B8 | 985 | For | `for block in listing.stdout.split('\x00'):` |
| B9 | 986 | comprehension | ` for line in block.splitlines() if line` |
| B10 | 987 | BoolOp | `lines and any((name.endswith('.go') for name in lines[1:]))` |
| B11 | 987 | If | `if lines and any((name.endswith('.go') for name in lines[1:])):` |
| B12 | 987 | comprehension | ` for name in lines[1:]` |
| B13 | 991 | comprehension | ` for commit in reversed(hashes) if commit in flagged` |

| raise 줄 | 소스 |
|---|---|
| 935 | `raise RuntimeError(f"cannot name this change's directory under the repository: {exc}") fro` |
| 959 | `raise RuntimeError(f'cannot list the commits that touched {change_dir}: {_first_line(touch` |
| 980 | `raise RuntimeError('cannot read the files those commits changed: ' + _first_line(listing.s` |

## task 7.5.29 — `-z` 로 읽는다 (2026-09-27)

> 편집 전 `ast.before-7529.json`(HEAD `26e5bb3f`) · 편집 후 `ast.after-7529.json`, `759_flm_rows.py` 로 정렬.

편집 후 `tools/logic-map/check_analysis.py:1869-1967` · 분기 14 · 반환 2 · raise 4 · 호출 30 (`ast.after-7529.json`, source sha `c9a58a69af03`)
편집 전 `tools/logic-map/check_analysis.py:1820-1910` · 분기 13 · 반환 2 · raise 3 · 호출 23 (`ast.before-7529.json`, revision `26e5bb3f`, source sha `ff590b79db6e`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 1896 | Try | `try:` | 같음 |
| B2 | B2 | 1898 | ExceptHandler | `except ValueError as exc:` | 같음 |
| B3 | B3 | 1905 | If | `if change_dir.startswith(ARCHIVE_PREFIX):` | 같음 |
| B4 | B4 | 1909 | If | `if before:` | 같음 |
| B5 | B5 | 1922 | If | `if touching.returncode:` | 같음 |
| B6 | B6 | 1931 | If | `if not hashes:` | 같음 |
| B7 | B7 | 1950 | If | `if listing.returncode:` | 같음 |
| B8 | — | — | For | `for block in listing.stdout.split('\x00'):` (옛) | **빠짐** |
| B9 | — | — | comprehension | `for line in block.splitlines() if line` (옛) | **빠짐** |
| B10 | — | — | BoolOp | `lines and any((name.endswith('.go') for name in lines[1:]))` (옛) | **빠짐** |
| B11 | — | — | If | `if lines and any((name.endswith('.go') for name in lines[1:])):` (옛) | **빠짐** |
| B12 | — | — | comprehension | `for name in lines[1:]` (옛) | **빠짐** |
| — | B8 | 1959 | For | `for block in listing.stdout.strip(b'\x00').split(b'\x00\x00'):` | **새** |
| — | B9 | 1961 | BoolOp | `names and (not names.startswith(b'\n'))` | **새** |
| — | B10 | 1961 | BoolOp | `not re.fullmatch(b'[0-9a-f]{40}\|[0-9a-f]{64}', commit) or (names and (not names.startswith(b'\n')))` | **새** |
| — | B11 | 1961 | If | `if not re.fullmatch(b'[0-9a-f]{40}\|[0-9a-f]{64}', commit) or (names and (not names.startswith(b'\n'))):` | **새** |
| — | B12 | 1963 | If | `if any((name.endswith(b'.go') for name in names[1:].split(b'\x00'))):` | **새** |
| — | B13 | 1963 | comprehension | `for name in names[1:].split(b'\x00')` | **새** |
| B13 | B14 | 1967 | comprehension | `for commit in reversed(hashes) if commit in flagged` | 번호만 |

**먼저 쟀다(7.5.29 가 "재지 않았다" 고 적은 것).** 판정을 **바꾼다** — 두 방향으로(RED, 편집 전 코드):
- `.go` 가 **아닌** `note.go<U+2028|U+2029|U+0085>txt` 를 고친 커밋에 깃발이 선다 — `str.splitlines()` 가 그 글자에서 잘라 `note.go` 가 된다(git 은 비ASCII 를 인용
  안 한다). 거절이 **늘어나는** 쪽.
- git 이 **인용하는** `*.go`(`"` · `\\` · `\x01`)를 고친 커밋에 깃발이 **안** 선다 — 줄이 `"we\\"ird.go"` 라 `.go` 로 안 끝난다. 거절이 **사라지는** 쪽 —
  7.5.29 가 안 센 반대편이고 7.5.13 과 같은 뿌리다.

**수리.** `git log -z` 로 받아 바이트로 읽는다(`text=True` 를 뗐다). git 2.43.0 모양: 커밋마다 `\0<sha>\0\n<이름>\0…` — 형식 `%x00` 이 앞 커밋의 마지막 이름 끝
NUL 과 만나 경계가 NUL **둘**이고, 이름은 비지 않으므로 NUL 둘은 경계에서만 난다(새 B8). 머리가 40/64 hex 가 아니거나 이름 목록이 `\n` 으로 시작하지 않으면
지어내지 않고 결함(새 B9~B11). 이름은 인용되지 않으므로 `endswith(b".go")` 가 그대로 맞다(새 B12 · B13).

**전수 A/B**(`analysis/harness/7529_repairs_ab.py`, before `26e5bb3f` 세 모듈 · after 워킹트리): change 디렉터리 **128**, 깃발 합계 392, **SAME 128 · DIFFERENT 0**
— 이 저장소에는 그 모양의 이름이 없다(위 센서스 0 과 같은 사실).

## task 7.5.42 — `log.showRoot` 고정 (마감 수리, 적대 리뷰, 2026-09-27)

> 편집 전 `ast.before-7542.json`(워킹트리 `c9a58a69`) · 편집 후 `ast.after-7542.json` — 분기 · 반환 · raise · 호출 수 같음(정렬 결과 전부 같음).

편집 후 `tools/logic-map/check_analysis.py:1869-1969` · 분기 14 · 반환 2 · raise 4 · 호출 30 (`ast.after-7542.json`, source sha `12beb79433de`)
편집 전 `tools/logic-map/check_analysis.py:1869-1967` · 분기 14 · 반환 2 · raise 4 · 호출 30 (`ast.before-7542.json`, revision `worktree`, source sha `c9a58a69af03`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 1896 | Try | `try:` | 같음 |
| B2 | B2 | 1898 | ExceptHandler | `except ValueError as exc:` | 같음 |
| B3 | B3 | 1905 | If | `if change_dir.startswith(ARCHIVE_PREFIX):` | 같음 |
| B4 | B4 | 1909 | If | `if before:` | 같음 |
| B5 | B5 | 1922 | If | `if touching.returncode:` | 같음 |
| B6 | B6 | 1931 | If | `if not hashes:` | 같음 |
| B7 | B7 | 1952 | If | `if listing.returncode:` | 같음 |
| B8 | B8 | 1961 | For | `for block in listing.stdout.strip(b'\x00').split(b'\x00\x00'):` | 같음 |
| B9 | B9 | 1963 | BoolOp | `names and (not names.startswith(b'\n'))` | 같음 |
| B10 | B10 | 1963 | BoolOp | `not re.fullmatch(b'[0-9a-f]{40}\|[0-9a-f]{64}', commit) or (names and (not names.startswith(b'\n')))` | 같음 |
| B11 | B11 | 1963 | If | `if not re.fullmatch(b'[0-9a-f]{40}\|[0-9a-f]{64}', commit) or (names and (not names.startswith(b'\n'))):` | 같음 |
| B12 | B12 | 1965 | If | `if any((name.endswith(b'.go') for name in names[1:].split(b'\x00'))):` | 같음 |
| B13 | B13 | 1965 | comprehension | `for name in names[1:].split(b'\x00')` | 같음 |
| B14 | B14 | 1969 | comprehension | `for commit in reversed(hashes) if commit in flagged` | 같음 |

목록 명령에 `-c log.showRoot=true`. `false` 면 루트 커밋의 `-z` 목록이 `\0<sha>\0`(이름 없음)이라 모양 검사를 지나고 그 커밋의 깃발이 **조용히** 빈다 —
RED(편집 전, 스크래치): 루트 커밋이 change 디렉터리와 `a.go` 를 만들어도 `[]`. GREEN `[<root>]`. `diff.renames` 와 같은 원칙. 노출 0 — 이 저장소의 루트 커밋은
change 디렉터리를 안 만든다.

# Function Logic Map: `_worktree_snapshot` (Python, a122 task 7.5.25 — 새 함수)

`ast.after-7.5.25.json` — 분기 **17** · raise 8 · 반환 0(생성기, `yield` 하나). `enumerate.py` 가 기계로 열거했다.

## 답하는 질문

워킹트리가 대상일 때 게이트가 base 와 견줄 **바이트**는 무엇인가 — git 이 고른 것이 아니라 디스크에 실제로 있는 것.
돌려주는 것(`yield`): 두 diff 에 줄 환경(`GIT_INDEX_FILE` · `GIT_OBJECT_DIRECTORY` · `GIT_ALTERNATE_OBJECT_DIRECTORIES`)과
경로 → 읽은 바이트. 임시 디렉터리는 `with` 가 끝날 때 사라진다.

## 갈래

| 분기 | 줄 | 뜻 |
|---|---|---|
| B1 | `if described.returncode:` | `rev-parse` 실패 → 이름 댄 `RuntimeError` |
| B2 · B3 | `len(answer) < 2 or not answer[0] or not answer[1]` | 객체 형식 · 객체 디렉터리 둘을 못 받았다 → 결함 |
| B4 | `if listed.returncode:` | `ls-files` 실패 → 결함 |
| B5 · B6 | `for entry …` · `if not entry: continue` | `-z` 의 끝 빈 조각만 건너뛴다 |
| B7 · B8 | `len(fields) != 4 or not raw_path` | 레코드 모양이 틀림 → 결함(빈 스냅숏 = "바뀐 것 없음" 이 되지 않게) |
| B9 | `if stage != b"0":` | 충돌 중 → 이름 대고 거절 |
| B10 | `if mode not in GO_FILE_MODES:` | 인덱스가 일반 파일(`100644`·`100755`)이라 하지 않음 → 이름 대고 거절(저장소 전수 0 / 1,762) |
| B11 · B12 | `try: _read_regular` · `except (FileNotFoundError, NotADirectoryError)` | 디스크에 없다 — 아래 B13 으로 가른다. 그 밖의 실패(FIFO · 권한 · 너무 큼)는 **올린다** |
| B13 | `if tag.upper() == b"S":` | skip-worktree + 없음 = sparse 로 **안 꺼냄** → 인덱스 blob. 그 밖(플래그 없음 · assume-unchanged)의 없음 = **지웠다** → 인덱스에서 뺀다 |
| B14 | `if digest != oid:` | 인덱스와 다른 바이트만 임시 저장소에 쓴다(같은 것은 실제 저장소에 있다) |
| B15 · B16 | `os.pathsep in … or '"' in …` | 실제 객체 디렉터리를 대체 저장소 목록에 못 적는다 → 이름 대고 거절 |
| B17 | `if built.returncode:` | `update-index` 실패 → 결함 |

## 이것이 닫는 것과 안 닫는 것

닫는다(각각 양성 대조군 — 평범한 `git diff` 는 정말 못 본다 — 과 함께 시험): clean 필터 · process 필터 · `ident` ·
`working-tree-encoding`(UTF-7 이중 표현) · 거짓말하는 fsmonitor · `checkStat=minimal` · **설정 없는** stat 캐시 ·
assume-unchanged · skip-worktree(파일이 있을 때) · assume-unchanged 뒤의 삭제.

안 닫는다: diff 의 **출력 형식**을 바꾸는 설정(`diff.noprefix` · `diff.mnemonicPrefix` · `color.ui=always`)과
pathspec 환경 변수 — 그것은 워킹트리 바이트가 아니라 파서의 문제이고 7.5.26 이다. skip-worktree 뒤에서 **파일을
지우는** 것도 sparse 와 구별할 수 없어 인덱스 blob 으로 읽는다(그 삭제는 판정에 안 보인다 — 7.5.25 의 잔여로 적었다).

## 실제 저장소에 쓰지 않는다

임시 인덱스(`GIT_INDEX_FILE`)와 임시 객체 저장소(`GIT_OBJECT_DIRECTORY`)에만 쓴다. 실제 객체 저장소는 대체 저장소로
**읽기만** 한다. `update-index` 에는 `core.splitIndex=false`(안 그러면 `sharedindex.*` 를 `.git` 에 쓴다) ·
`core.hooksPath=/dev/null`(안 그러면 `post-index-change` 훅이 뜬다 — 실측) · `core.fsmonitor=false`(인덱스를 읽는
순간 fsmonitor 명령이 뜬다 — 실제 인덱스를 읽는 `ls-files` 에서 실측)를 준다. 시험이 셋을 다 켜 두고 `.git` 의
모든 파일 바이트와 표식 파일을 전후로 견준다.

## task 7.5.31 — 재리뷰 수리 (2026-09-24)

`ast.before-7.5.31.json`(분기 17 · raise 8) → `ast.after-7.5.31.json`(분기 **15** · raise **7**). difflib 정렬: 편집 전
B15 · B16(`os.pathsep in … or '"' in …` 거절)이 빠졌다 — 대체 저장소 항목을 `_c_quoted` 로 적는다. 나머지 분기는 순서
그대로다. 분기 밖의 변경 셋: 경로를 `os.fsdecode` 로 읽는다(이름이 UTF-8 이 아닌 **안 바뀐** 파일로 멈추지 않게, F9) ·
임시 인덱스에 실은 날 경로 → oid 를 `placed` 로 모아 셋째 값으로 돌려준다(sparse 로 안 꺼낸 경로도 인덱스 oid 로
싣는다) · 모듈이 `GIT_NO_REPLACE_OBJECTS=1` 을 둬 이 함수의 git 도 교체 참조를 안 따른다(F1 · F4).

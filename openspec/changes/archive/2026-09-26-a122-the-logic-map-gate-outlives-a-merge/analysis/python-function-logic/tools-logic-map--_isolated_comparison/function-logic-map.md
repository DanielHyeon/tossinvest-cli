# Function Logic Map: `_isolated_comparison` (Python, a122 task 7.5.34 — 새 함수)

`ast.after-7.5.34.json` — 분기 **24** · raise 3 · 반환 0(contextmanager — `Comparison` 을 `yield` 한다).

## 답하는 질문

두 diff 가 견줄 두 트리를 **대체 저장소가 없는** 임시 저장소에 세운다. 판정이 읽는 모든 바이트의 출처는 둘이다 —
**디스크**(워킹트리 쪽, 게이트가 해시한다)와 **검증한 객체**(base · 커밋 대상 · sparse 인덱스 blob). git 이 찾는 blob
은 게이트가 쓴 것뿐이고, 없으면 `unable to read` 로 멈춘다(실측).

## 갈래

| 분기 | 뜻 |
|---|---|
| B1 | 대상이 커밋이면 두 리비전, 워킹트리면 base 하나 |
| B2 · B3 | `-` 로 시작하는 리비전 → `ValueError`(옵션으로 읽히지 않게) |
| B4 · B5 | `rev-parse --show-object-format <r>^{commit} <r>^{tree} …` |
| B6 | `rev-parse` 실패 → 결함 |
| B7 · B8 · B9 | 답이 `1 + 2×리비전` 개가 아니거나 온전한 oid 가 아니다 → 결함 |
| B10 | oid 를 글자로 |
| B11 | 커밋 대상이면 새 쪽도 검증한 트리, 워킹트리면 `_worktree_entries` |
| B12 | **다른** 경로만(모드 · oid 가 다르거나 한쪽에만 있다) — 같은 oid 는 diff 가 안 읽는다. rename 짝짓기는 양쪽을 다 읽으므로 한쪽에만 있는 경로도 든다 |
| B13 · B14 · B15 | 다른 경로의 양쪽마다 |
| B16 · B17 · B18 | 물을 blob: 그 쪽에 있고 · gitlink 가 아니고 · 디스크에서 읽은 워킹트리 바이트가 아니면 — base, 커밋 대상, sparse 인덱스 blob |
| B19 | oid → 바이트 **한 표**: 검증한 객체에 디스크 바이트(게이트가 해시했다)를 더한다. oid 가 같으면 바이트도 같으므로 출처를 고를 갈래가 없다 — 고르던 첫 판본(`disk[path] if … else fetched[oid]`)은 변이 AJ44 가 살아남아 이렇게 바꿨다 |
| B20 · B21 | 다시 다른 경로의 양쪽마다 |
| B22 · B23 | 없거나 gitlink 면 넘어간다(gitlink 는 다른 저장소의 커밋이다). 나머지는 표의 바이트를 그 쪽에 싣고 임시 저장소에 쓴다 |
| B24 | 물려받은 `GIT_ALTERNATE_OBJECT_DIRECTORIES` 를 **지운다** — git 은 대체 저장소의 pack 을 loose 보다 먼저 보므로 부모 환경의 대체 저장소가 위조 pack 으로 격리를 이긴다 |

## 실제 저장소에 쓰지 않는다

읽기는 `rev-parse` · `ls-tree` · `cat-file` · `ls-files` 뿐이고, 쓰기는 임시 인덱스 둘과 임시 저장소뿐이다. 첫 판본은
두 diff 의 `GIT_INDEX_FILE` 도 새 쪽 임시 인덱스로 돌렸는데, 그 줄을 빼는 변이 AJ47 이 살아남았다 — 트리 둘을 견주는
diff 는 인덱스를 안 쓰고, 그 판에서도 "실제 저장소에 쓰지 않는다" 시험(`.git` 전 파일 바이트 · 훅 · fsmonitor 표식)이
초록이었다. 지킬 것이 없는 줄이라 지웠다.

## 이것이 닫는 것

7.5.25 의 워킹트리 투영 문 전부(스냅숏이 하던 것) + 교체 참조(F1 · F4) + 위조 loose · pack(F2 · F3 · 재리뷰 #1) +
rename 뒤의 위조(#2) + base 쪽 blob · 하위 트리 위조(7.5.33) + 물려받은 대체 저장소 + pathspec 환경 변수(F7) + 연결
워크트리 경로의 해독 · 줄바꿈(#4 · #5 — 실제 저장소의 경로를 묻지 않는다). 뿌리 트리 · 커밋 위조는 git 이 먼저
`hash mismatch` 로 거절한다(2.43 실측).

## 안 닫는 것

부분 클론(`--filter=blob:none`)에서 base blob 이 없으면 `cat-file` 이 promisor 로 **fetch** 한다 — 판정은 맞지만 실제
`.git/objects/pack` 에 쓴다(7.5.33 의 남은 절반). diff 의 **출력 형식**을 바꾸는 설정(`diff.noprefix` 등, 7.5.26).
판정 도구 자체를 워킹트리에서 빌드하는 것(7.5.32).

## task 7.5.35 — `GIT_DIFF_OPTS` 도 물려받지 않는다 (2026-09-25)

`ast.before-7.5.35.json` → `ast.after-7.5.35.json`: 분기 24 · raise 3 · 반환 0 그대로. 바뀐 것은 B24 의 조건 하나다 —
`key != "GIT_ALTERNATE_OBJECT_DIRECTORIES"` → `key not in ("GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_DIFF_OPTS")`.
`GIT_DIFF_OPTS=-u3` 은 판정 diff 의 `--unified=0` 을 **이겨** 문맥 줄이 편집 안 한 함수까지 요구하게 했다(2.43 실측, 이
change 이전부터). 앞의 "안 닫는 것" 의 출력 형식(7.5.26)은 7.5.35 가 판정 diff 의 명령줄에서 닫았다.

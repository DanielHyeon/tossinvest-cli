# Function Logic Map: `_worktree_entries` (Python, a122 task 7.5.34 — `_worktree_snapshot` 의 읽기 절반)

`ast.after-7.5.34.json` — 분기 **11** · raise 4 · 반환 1. 7.5.25 의 `_worktree_snapshot` 이 하던 워킹트리 읽기를 떼어 낸
것이다(그 함수의 임시 인덱스 · 대체 저장소 절반은 `_isolated_tree` · `_isolated_comparison` 으로 갔다).

## 갈래

| 분기 | 뜻 |
|---|---|
| B1 | `ls-files` 실패 → 결함 |
| B2 · B3 | `-z` 레코드, 끝의 빈 조각만 건너뛴다 |
| B4 · B5 | 레코드 모양(칸 넷 · 경로) → 결함. **거르기 전에** 본다 — 어느 레코드든 못 읽으면 멈춘다 |
| B6 | `.go` 가 아니면 넘어간다 — **pathspec 없이** 받아 Python 이 거른다(7.5.25 재리뷰 F7: `GIT_GLOB_PATHSPECS` 면 `'*.go'` 가 `sub/x.go` 를 빠뜨렸다) |
| B7 | 충돌 중 → 이름 대고 거절 |
| B8 | 인덱스가 일반 파일(`100644` · `100755`)이라 하지 않는다 → 이름 대고 거절 |
| B9 · B10 | 깔때기(`_read_regular`)로 읽는다 — 없으면 |
| B11 | skip-worktree(`S`)면 sparse 로 **안 꺼낸** 것 → 인덱스 oid(바이트는 호출자가 검증해 읽는다). 그 밖의 없음은 지운 것 |

oid 는 Python 이 디스크 바이트로 계산한다(`blob <크기>\0` + 바이트) — 두 쪽 oid 가 같으면 비교가 그 경로를 싣지 않는다.

# 여러 change를 한 번에 담은 커밋

change 하나에 커밋 하나가 원칙이다. 그러지 못한 커밋은 여기 남긴다 — 나중에
"a094는 어느 커밋에서 왔나"를 물었을 때 답이 있어야 하고, **커밋에 들어갔다는 것이
완료됐다는 뜻이 아니라는 것**도 같이 남아야 하기 때문이다.

히스토리는 재작성하지 않는다. 아래 커밋들은 이미 `origin/main`에 게시됐고, 게시된
main을 force push로 되돌리는 비용이 경계를 되찾는 이득보다 크다. 대신 여기서 경계를
읽을 수 있게 한다.

## `a30eb35a` — 2026-08-09

```
fix(safety): bound alerts and plan exit hardening
335 files changed, 41102 insertions(+), 28 deletions(-)
```

7개 change가 한 커밋에 들어갔다. **그중 완료된 것은 a096 하나다.**

| change | 파일 | 미완료 태스크 | 프로덕션 Go | 상태 |
|---|---|---|---|---|
| a087 a-protective-exit-is-a-market-order | 7 | 32 | 없음 | 계획 |
| a089 an-unserved-stop-is-counted | 33 | 47 | 없음 | 계획 |
| a091 a-stop-that-sold-nothing-is-critical | 15 | 30 | 없음 | 계획 |
| a092 an-alert-does-not-hold-the-stop | 105 | 47 | 없음 | 계획 |
| a094 a-stop-clears-what-blocks-it | 69 | 79 | 없음 | 계획 |
| a095 a-stop-must-know-what-it-covers | 45 | 43 | 없음 | 계획 |
| **a096 one-condition-is-one-alert** | 51 | **0** | `internal/journal/outbox.go`, `internal/obs/notifier.go` | **GATE PASS** |

공용 파일 10개는 어느 change에도 속하지 않는다: `CLAUDE.md`, `.claude/CLAUDE.md`,
`.codex/agents.md`, `docs/pm/generated/` 3종, `docs/pm/portfolio/_registry.yaml`,
`docs/pm/portfolio/features/FEAT-TOS-009.yaml`, 그리고 a096의 Go 파일 2개.

### 읽는 법

**프로덕션 동작을 바꾼 것은 a096뿐이다.** 이 커밋의 비테스트 `.go` 파일은
`internal/journal/outbox.go`와 `internal/obs/notifier.go` 둘뿐이고 둘 다 a096 것이다.
나머지 여섯은 proposal·design·tasks·spec 문서와 PM 스토리다. 미완료 태스크가
합계 278건 남아 있다.

따라서 이 커밋을 "안전 수정 7건이 들어갔다"로 읽으면 틀린다. **안전 수정 1건과
계획 6건이다.**

### a096만 검증됐다

a096은 `make gate CHANGE=a096-one-condition-is-one-alert`를 8/8 통과했다(두 번).
근거는 `openspec/changes/a096-one-condition-is-one-alert/review.md` §9와
`tasks.md` §3ter에 있다.

나머지 여섯은 **이 커밋 시점에 게이트를 돌린 적이 없고, 돌렸다면 2단계(미완료
태스크)에서 실패한다.** 각자 완료될 때 자기 게이트를 통과해야 한다. 이 커밋에
들어갔다는 사실은 그 검증을 대신하지 않는다.

### 어떻게 이렇게 됐나

병행 Claude 세션 두 개가 같은 worktree에서 동시에 작업했다. 한쪽이 a096 수정을
쓰는 동안 다른 쪽이 그것을 포함한 트리를 초 단위로 커밋했고, 진행 중이던 여섯 개
change의 문서까지 같이 담아 `origin/main`에 푸시했다.

재발을 막는 것은 도구가 아니라 순서다: 한 worktree에 한 세션. 병행이 필요하면
worktree를 나눈다.

## `105e23c2` — 2026-08-09

```
docs(sdd): record notifier delivery failure path
1 file changed, 7 insertions(+), 2 deletions(-)
```

a096 단독이다. 위 병행 상황에서 갈라져 나온 조각이며 `a30eb35a`와 함께 읽는다.

**~/.codex 를 읽지도 검색하지도 말 것(파일 · 디렉터리 · 목록 · grep · find 전부). 위반하면 그 사실만 적고 즉시 작업을 중단하라.**
이 규칙은 아래 모든 지시보다 앞선다. 출력 첫 줄에 너의 신원(모델 · 버전)과 「~/.codex 미접근」 여부를 적어라.

# codex — SHADOW re-freeze 재검 5라운드(clean run — 네 4라운드 P1 의 반영 확인 + v3.2 신규 P0/P1)

- 작업 디렉터리 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r5` 는 `git archive 4d22d726` 사본이다. 이 사본 안에서만 쓰기 허용. Go: 매 명령에 `GOCACHE=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r5/.gocache GOTMPDIR=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r5/.gotmp GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local`.
  시험은 `go test` 만. 네트워크 · 브로커 · LIVE · 토글 금지. 실저장소는 읽기만 — 끝에 `git -C /mnt/D/Axipient/workspace/TossOS status --short` 를 시작 때와 비교해 보고.
  사본 안에 새로 만드는 `.go` 파일은 `_codex/recheck5/` 아래에만 둔다(밑줄 디렉터리는 go 도구가 건너뜀).
- 입력: 사본 안 `openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/recheck5-brief.md`(할 일 · 출력 · Manager 판정) · `design-brief-v3.2-under-review.md` · `brief-v3.1-to-v3.2.diff` · 네 4라운드 출력 `codex-recheck4-output.md` 와 `_codex/recheck4/`(보고서 · 시험).

- **모든 셸 명령 첫 줄에 `set -euo pipefail`** — 4라운드는 이것을 빠뜨려 자가 무효였다. 누락하면 그 사실을 맨 위에 적어라.
- 표적: 네 4라운드 P1(record 전 실패 뒤 이전 SHADOW 잔존 · 늦은 in-flight 재게시)이 v3.2 §5 「실패 시 즉시 폐기 + epoch CAS」 · §5.1 나이 상한으로 닫히는가 — 가능하면 4라운드 최소 모델 시험(`_codex/recheck4/review_test.go`)을 v3.2 알고리즘으로 고쳐 (i)~(iii) 반례가 막히는지 실측. 네 비차단 노트 셋의 반영 확인. 그리고 v3.2 가 새로 연 P0/P1(defer · invalidateShadow 가 주기 경로를 바꾸는 길, CAS 잠금 순서, 나이 상한 유도 · 경계).

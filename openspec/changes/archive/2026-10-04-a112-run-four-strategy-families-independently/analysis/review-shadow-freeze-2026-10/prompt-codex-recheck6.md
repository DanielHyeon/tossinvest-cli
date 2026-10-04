**~/.codex 를 읽지도 검색하지도 말 것(파일 · 디렉터리 · 목록 · grep · find 전부). 위반하면 그 사실만 적고 즉시 작업을 중단하라.**
이 규칙은 아래 모든 지시보다 앞선다. 출력 첫 줄에 너의 신원(모델 · 버전)과 「~/.codex 미접근」 여부를 적어라.

# codex — SHADOW re-freeze 재검 6라운드(마지막 · clean run — v3.3 접기 다섯의 반영 검증)

- 작업 디렉터리 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r6` 는 `git archive 4d22d726` 사본이다. 이 사본 안에서만 쓰기 허용. Go: 매 명령에 `GOCACHE=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r6/.gocache GOTMPDIR=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r6/.gotmp GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local`.
  시험은 `go test` 만. 네트워크 · 브로커 · LIVE · 토글 금지. 실저장소는 읽기만 — 끝에 `git -C /mnt/D/Axipient/workspace/TossOS status --short` 를 시작 때와 비교해 보고.
  사본 안에 새로 만드는 `.go` 파일은 `_codex/recheck6/` 아래에만 둔다.
- **모든 셸 명령 첫 줄에 `set -euo pipefail`.** 누락하면 그 사실을 맨 위에 적어라.
- 입력: 사본 안 `openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/recheck6-brief.md`(종결 규칙 · 접기 다섯 · 출력 형식 — 먼저 전부 읽어라) · `design-brief-v3.3-under-review.md` · `brief-v3.2-to-v3.3.diff` · 네 5라운드 출력 `codex-recheck5-output.md` 와 `_codex/recheck5/`.
- 할 일: 접기 다섯 각각 반영됨 / 안 됨. 네 5라운드 반례 시험(`TestV32HealthyCycleContinuityCounterexample`)이 v3.3 §5.1 의 좁힌 주장과 「의도된 간극」 별도 시험 문면에 어떻게 맞는지 확인(필요하면 `_codex/recheck5/review_test.go` 를 v3.3 문면으로 고쳐 재실행). 새 사냥 축 금지 — P0 만 예외, 새 P1 급 추정은 「구현 단계 RED 후보」 로 따로.

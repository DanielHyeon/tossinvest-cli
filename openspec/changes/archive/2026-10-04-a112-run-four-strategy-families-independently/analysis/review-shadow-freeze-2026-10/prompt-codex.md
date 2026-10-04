**~/.codex 를 읽지도 검색하지도 말 것(파일 · 디렉터리 · 목록 · grep · find 전부). 위반하면 그 사실만 적고 즉시 작업을 중단하라.**
이 규칙은 아래 모든 지시보다 앞선다. 출력 첫 줄에 너의 신원(모델 · 버전)과 「~/.codex 미접근」 여부를 적어라.

# codex — a112 7.3.1 SHADOW freeze 급 독립 적대 리뷰(설계 + 계약 개정, 코드 0)

- 작업 디렉터리 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow` 는 `git archive c7219640` 사본이다(git 이력 없음, 실저장소 아님). 이 사본 **안에서만** 쓰기가 허용된다 — 실험 스케치 파일은 여기에.
- Go 실행: 매 명령에 `GOCACHE=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow/.gocache GOTMPDIR=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow/.gotmp GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local` 를 붙여라. 시험은 `go test` 만(태그 `tossos_testseams` 허용).
  네트워크 · 브로커 · LIVE · 토글 금지.
- 실저장소 `/mnt/D/Axipient/workspace/TossOS` 는 **읽기만** 하라. 끝에 `git -C /mnt/D/Axipient/workspace/TossOS status --short` 를 시작 때와 비교해 무변을 보고하라.
- 리뷰 대상 · 질문 · 출력: 사본 안 `openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/brief.md`(공통 브리프 — 질문 1~7 전부가 네 범위) · `design-brief-under-review.md` · `amendment-c7219640.patch`.
  특히 브리프 질문 1(MUST NOT 둘이 구조로 막히는가 — 사본에서 설계의 의도된 모양을 최소 스케치로 세워 길이 열리는지 실측) · 5(코드 사실 주장 실측)에 무게.

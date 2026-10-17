**~/.codex 를 읽지도 검색하지도 말 것(파일 · 디렉터리 · 목록 · grep · find 전부). 위반하면 그 사실만 적고 즉시 작업을 중단하라.**
이 규칙은 아래 모든 지시보다 앞선다. 출력 첫 줄에 너의 신원(모델 · 버전)과 「~/.codex 미접근」 여부를 적어라.

# codex — SHADOW re-freeze 재검 3라운드(네 2라운드 PARTIAL P1-5 · 새 P1 N1 의 반영 확인 + v3 신규 P0/P1)

- 작업 디렉터리 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r3` 는 `git archive 4d22d726` 사본이다. 이 사본 안에서만 쓰기 허용. Go: 매 명령에 `GOCACHE=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r3/.gocache GOTMPDIR=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r3/.gotmp GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local`.
  시험은 `go test` 만. 네트워크 · 브로커 · LIVE · 토글 금지. 실저장소는 읽기만 — 끝에 `git -C /mnt/D/Axipient/workspace/TossOS status --short` 를 시작 때와 비교해 보고.
- 입력: 사본 안 `openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/recheck3-brief.md`(할 일 · 출력) · `design-brief-v3-under-review.md` · `amendment-v3-4d22d726.patch` · 네 2라운드 출력 `codex-recheck-output.md` 와 스케치 `_codex/recheck/`.
- P1-5: amendment v3(결정 63 이 61 의 (a)·(b) 만 적용, 61(c) 비재채택 문장) · 브리프 v3 §1 이 닫는가. N1: v3 §4(운반은 authority 밖 별도 값 · opaque ShadowInput) · §2 ②(사용 기반 census)가 닫는가 — 가능하면
  네 2라운드 census 스케치를 v3 모양으로 고쳐 실측. 그리고 v3 가 새로 연 P0/P1(비동기 shadow 단계 · 운반 값 반환 모양 변경 · wrapper export).

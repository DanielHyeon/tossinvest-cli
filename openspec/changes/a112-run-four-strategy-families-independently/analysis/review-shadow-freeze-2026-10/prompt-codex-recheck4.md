**~/.codex 를 읽지도 검색하지도 말 것(파일 · 디렉터리 · 목록 · grep · find 전부). 위반하면 그 사실만 적고 즉시 작업을 중단하라.**
이 규칙은 아래 모든 지시보다 앞선다. 출력 첫 줄에 너의 신원(모델 · 버전)과 「~/.codex 미접근」 여부를 적어라.

# codex — SHADOW re-freeze 재검 4라운드(네 3라운드 N2 의 반영 확인 + v3.1 신규 P0/P1)

- 작업 디렉터리 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r4` 는 `git archive 4d22d726` 사본이다. 이 사본 안에서만 쓰기 허용. Go: 매 명령에 `GOCACHE=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r4/.gocache GOTMPDIR=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow-r4/.gotmp GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local`.
  시험은 `go test` 만. 네트워크 · 브로커 · LIVE · 토글 금지. 실저장소는 읽기만 — 끝에 `git -C /mnt/D/Axipient/workspace/TossOS status --short` 를 시작 때와 비교해 보고.
  사본 안에 새로 만드는 `.go` 파일은 `_codex/` 아래에만 둔다(밑줄 디렉터리는 go 도구가 건너뜀).
- 입력: 사본 안 `openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/recheck4-brief.md`(할 일 · 출력 · Manager 판정) · `design-brief-v3.1-under-review.md` · `brief-v3-to-v3.1.diff` · 네 3라운드 출력 `codex-recheck3-output.md` 와 스케치 `_codex/recheck3/`.
- N2: v3.1 §2(unsafe/reflect 「직접 import」 · 모듈 안 패키지 한정 · 표준 라이브러리 예외)가 닫는가 — 가능하면 strategyrouter · strategyworker 폐포의 모듈 패키지 import 목록을 사본에서 실측. 그리고 v3.1 이 새로 연 P0/P1(특히 §5 핀 ① 재진술 (A) · ④ · ⑤, `{wave, batch}` 보관과 evaluate 오류 갈래, §4 부재 값 대 빈 묶음, §5.1 투영 시점 만료).

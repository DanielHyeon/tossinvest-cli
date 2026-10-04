**~/.codex 를 읽지도 검색하지도 말 것(파일 · 디렉터리 · 목록 · grep · find 전부). 위반하면 그 사실만 적고 즉시 작업을 중단하라.**
이 규칙은 아래 모든 지시보다 앞선다. 출력 첫 줄에 너의 신원(모델 · 버전)과 「~/.codex 미접근」 여부를 적어라.

# codex 2차 — a112 8.5 독립 적대 리뷰(대상 A · B1 · B2 전부)

- 작업 디렉터리 `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-85-r2` 는 `git archive 556c3c1f` 사본이다(git 이력 없음, 실저장소 아님). 이 사본 **안에서만** 쓰기가 허용된다 — 실험용 시험 파일 · 변이 사본은 여기에.
- Go 실행: 매 명령에 `GOCACHE=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-85-r2/.gocache GOTMPDIR=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-85-r2/.gotmp GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local` 를 붙여라(사본 밖에 쓰지 않게). 시험은
  `go test` 만(태그 `tossos_testseams` 허용). 네트워크 · 브로커 · LIVE · 토글 금지.
- 실저장소 `/mnt/D/Axipient/workspace/TossOS` 는 **읽기만** 하라(전/후 커밋 비교가 필요하면 diff 파일을 쓰라 — 사본 안의
  `openspec/changes/a112-run-four-strategy-families-independently/analysis/review-8.5-2026-10/diff-*.patch`). 끝에
  `git -C /mnt/D/Axipient/workspace/TossOS status --short` 를 시작 때와 비교해 무변을 보고하라.
- 1차 판(무효)의 미실행 P2 셋을 실행으로 확인하라: (1) B2 — FX 미준비 + getenv nil 에서 닫힘 사유가 FX_NOT_READY → INTERNAL_FAILURE 로 바뀌는가(편집 전/후), (2) B2 — 조기
  활성화 적재의 ctx 취소 · 지연이 주기를 붙잡는가, (3) A — 같은 입력의 편집 전/후 결정 · seal 이 같은가.
- 대상 · 질문 · 출력 형식은 아래 공통 브리프를 따른다.


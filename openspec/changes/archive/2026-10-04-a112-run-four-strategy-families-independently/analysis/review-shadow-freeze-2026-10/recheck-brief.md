# SHADOW re-freeze — 표적 재검 브리프 (2026-10-04)

Manager 합본 판정: 설계 개정 → **같은 4판이 자기 P0/P1 종결만 표적 재검**(전면 재리뷰 아님). 재검 전원 PASS 뒤 Pre-Edit FLM.

- 안전 규칙: `brief.md`(같은 디렉터리)의 규칙 그대로 — `~/.codex` 금지(위반 시 맨 위에 적고 중단), 읽기 전용, 실험은 `/tmp` 아래 사본에서만(`git -C /mnt/D/Axipient/workspace/TossOS archive
  2817064c | tar -x -C <사본>`), `set -euo pipefail` · `git -C` · toplevel 단언, 네트워크 · 브로커 · LIVE · 토글 금지, 시작 · 끝 `git status --short` · `rev-parse HEAD` 동일 보고.
- **좌표: 커밋 `2817064c`**(amendment v2 착지). 입력:
  - 설계 브리프 v2 `design-brief-v2-under-review.md` — sha256 `7f1e840c85453fa0739afaca25db61472b5fce03cbd7f4f37e0506bb69ca37bd`(v1 은 `design-brief-under-review.md`).
  - amendment v2 diff `amendment-v2-2817064c.patch`(c7219640 → 2817064c, spec :38 · :91-94 · design 결정 63 문단).
  - 네 자신의 1차 출력(같은 디렉터리 `voice1-output.md` · `voice2-output.md` · `voice3-output.md` · `codex-output.md`).
- **할 일:** 네 1차 출력의 P0 · P1 **각각**에 대해: v2 의 어느 절 · amendment 의 어느 줄이 그것을 닫는가(좌표), 닫힘이 성립하는가(가능하면 사본 스케치로 실측), 판정
  **CLOSED / PARTIAL / OPEN**. 그리고 v2 가 **새로 연** P0/P1 이 있으면(예: 관문 앞 운반 타입이 dispatch 에 닿는 길, 별도 패키지 경계의 구멍, 차등 척도의 빈 표본) 그것만 추가로 적어라.
  P2 는 새로 쓰지 말고 1차 P2 중 v2 가 악화시킨 것만.
- 출력: 전체 판정 한 줄(PASS = 네 P0/P1 전부 CLOSED · 새 P0/P1 0 / FAIL) + 표(1차 ID · 판정 · v2 좌표 · 근거/재현) + 새 P0/P1 + 저장소 무변경 확인.

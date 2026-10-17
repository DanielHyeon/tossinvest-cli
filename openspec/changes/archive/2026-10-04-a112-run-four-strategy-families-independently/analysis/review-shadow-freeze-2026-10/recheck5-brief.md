# SHADOW re-freeze — 재검 5라운드 표적 브리프 (2026-10-05)

Manager 판정: 재검 4라운드(codex FAIL — 자가 무효(pipefail 누락) · 보이스 1 FAIL · 보이스 3 FAIL, P0 0)의 남은 항목만 표적으로 본다. 전면 재리뷰 아님. 보이스 2 는 PASS 유지.

- 안전 규칙: `brief.md` 그대로(`~/.codex` 금지 · 읽기 전용 · `/tmp` 사본만 `git -C /mnt/D/Axipient/workspace/TossOS archive 4d22d726 | tar -x -C <사본>` · **모든 셸 첫 줄 `set -euo pipefail`**(4라운드 codex 가 누락해 자가 무효) · `git -C` · toplevel 단언 ·
  네트워크/브로커/LIVE/토글 금지 · 시작/끝 `status --short` · `rev-parse HEAD` 동일 보고).
- **좌표: 커밋 `4d22d726`**(Go 변경 0). 입력: 브리프 v3.2 `design-brief-v3.2-under-review.md` — sha256 `68512bc36b58f0c184115016b377c17d0c3065535d136e1a41aa783c48b09af1`, v3.1 과의 차이 `brief-v3.1-to-v3.2.diff` ·
  4라운드 출력 `codex-recheck4-output.md`(+ `_codex/recheck4/report.md`) · `voice1-recheck4-output.md` · `voice3-recheck4-output.md`.
- Manager 판정(2026-10-05, v3.2 의 근거): 1 (a) cycle 클로저의 recover 없는 defer 가 비-nil 반환(오류 · panic)이면 `invalidateShadow(market)`(epoch++ · 관측 삭제) + shadow 게시는 epoch ∧ wave CAS
  (늦은 in-flight 재출현 차단) · (b) 관측 나이 상한 backstop(⚠ 값: 판정 문면 2×PollInterval 대신 v3.2 는 유도식 74s — 정상 입력 거부 회피, Manager 승인 2026-10-05). 2 칸 `{wave, batch, activation}` 같은
  임계 구역 + 금지 집합에 활성화 적재 셋. 3 `forMarket` shape 핀 + 핀 ① 타입 규칙 + 인자 index 5. 4 수집 helper · 생성자 shape 핀만(주입 seam 미채택 사유 기록). 5 문구 · 범위 정정(unsafe/reflect 핀은
  strategyshadow 루트 생산 그래프 둘만).
- **표적(리뷰어별):**
  - codex(clean run): 4라운드 P1(record 전 실패 뒤 이전 SHADOW 잔존 · 늦은 in-flight 재게시) → v3.2 §5 「실패 시 즉시 폐기 + epoch CAS」 · §5.1 나이 상한. 4라운드 비차단 노트 셋(forMarket 예외 · ④ 인자 전달 · §4 범위) 반영 확인.
  - 보이스 1: N2 PARTIAL(파도 정지 중 철회) → v3.2 §5 · §5.1(defer 폐기 · CAS · 나이 상한 · 핀 (f)) · NP1-1 → §5 핀 ① 타입 규칙 · forMarket shape · NP1-2 → §4 helper · 생성자 shape 핀(주입 핀 미채택 사유 판단 포함) · 활성화 출처 노트 → §5 칸.
  - 보이스 3: P1-C PARTIAL → §5 칸 `{wave, batch, activation}` · §2 ③ 활성화 적재 셋 · P1-D → §5 핀 ① · forMarket shape · index 5 · 좌표 정정(§0 :91 · 머리말).
  - 공통: v3.2 가 새로 연 P0/P1 만 — 특히 defer · invalidateShadow 가 주기 경로(panic 전파 · latch · 오류 반환)를 바꾸는 길, CAS 의 잠금 순서(런타임 mu 와 레인 잠금 — projection 의 「런타임 → 레인」 순서와 충돌 여부), 나이 상한 값의 유도와 경계 핀.
- 출력: 전체 판정(PASS = 자기 남은 항목 전부 CLOSED · 새 P0/P1 0 / FAIL) + 표(항목 · 판정 · v3.2 좌표 · 근거) + 새 P0/P1 + 저장소 무변경 확인.

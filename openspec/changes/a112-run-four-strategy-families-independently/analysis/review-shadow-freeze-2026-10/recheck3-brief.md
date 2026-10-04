# SHADOW re-freeze — 재검 3라운드 표적 브리프 (2026-10-04)

Manager 판정: 재검 2라운드(4판 FAIL, P0 0)의 남은 홀 8 항 + codex 신규 N1 만 표적으로 본다(N1 은 확정 3 이 닫으므로 반영 확인만). 전면 재리뷰 아님.

- 안전 규칙: `brief.md` 그대로(`~/.codex` 금지 · 읽기 전용 · `/tmp` 사본만 `git -C /mnt/D/Axipient/workspace/TossOS archive 4d22d726 | tar -x -C <사본>` · `set -euo pipefail` · `git -C` · toplevel 단언 ·
  네트워크/브로커/LIVE/토글 금지 · 시작/끝 `status --short` · `rev-parse HEAD` 동일 보고).
- **좌표: 커밋 `4d22d726`**(amendment v3 착지 — 결정 63 이 61 의 (a)·(b) 만 적용, 61(c) 비재채택). 입력: 브리프 v3 `design-brief-v3-under-review.md` — sha256 `f7883a47fb1f07dffb32e561e9e9ab00c8597f8cc6f108b97f3b5d65053230fc`(v2 와의 차이는 머리말 「v3 변경」 줄과 §1 · §2 · §3 · §4 · §5 · §5.1 · §8 · §10) ·
  `amendment-v3-4d22d726.patch` · 네 2라운드 출력(`voice{1,2,3}-recheck-output.md` · `codex-recheck-output.md`).
- 표적 8 항(Manager 확정): ① 조정 뒤 닫힘 여섯 운반 · 조정 앞 일곱 「관측 없음」(§4) ② 활성화 헬퍼는 새 파일 중립 wrapper export · 서술자 표 하나 · 읽기 오류 `%v`(§3) ③ authority 밖 별도 값 + opaque
  `ShadowInput` + census 는 types.Info.Uses · 본문 식 타입 + 양성 대조(§2 · §4) ④ shadow step 은 dispatch 뒤 · cycle 함수 밖 · 상수 마감 · AST 핀 · 마감 초과 fault(§5) ⑤ 프로세스 내 소거 +
  파도 신선도(§5.1) ⑥ 재시작 전제 WOULD_EMIT ≥ 1 전체 주기(§10) ⑦ FLM 목록(collectMarket 무조건, §4) ⑧ strategyshadow 폐포 unsafe/reflect 금지(§2). + codex N1 · P1-5(amendment v3) 반영 확인.
- **할 일:** 네 2라운드 PARTIAL · 새 P1 각각에 대해 CLOSED / PARTIAL / OPEN(v3 좌표 · 근거 — 가능하면 사본 스케치로). 그리고 v3 가 새로 연 P0/P1 만(예: 비동기 shadow 단계의 경합 · 운반 값
  변경이 order 경로 함수 서명을 바꾸는 자리 · wrapper export 가 strategyrouter API 를 넓히는 위험).
- 출력: 전체 판정(PASS = 네 남은 항목 전부 CLOSED · 새 P0/P1 0 / FAIL) + 표 + 새 P0/P1 + 저장소 무변경 확인.

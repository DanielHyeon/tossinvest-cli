# a090 · review

## 0. 초안 (2026-09-29) — 판정 아님

- 배정: Manager(2026-09-28, 사용자 결정 「a089 처분 추천안 수용」 의 둘째 항목). 정본 소스: a089 2차 리뷰 「a090(신설, 선행)」 원문 · C2 ·
  `exitloop.go:453-462` 의 두 무음 `continue`(base `d3bd1843` 에서 재확인) · a092 `ObserveOnce` FLM 「필요한 RED (a090 후보)」 R1~R6.
- 산출물 순서: AST(8 분기) → 진입 실측(`analysis/harness/observeonce_entry.sh`, commit `b0a202b8`) → FLM·BTM → proposal·design·spec·tasks.
- 사용자 결정 대기: Q1(포지션 단위 두절의 ENTRY_BLOCKED) · Q2(정지·0가격 종목 `/prices` 실측, 선택) · Q3(관측점 — 판정 진입 vs 기록 경계).
- 다음: proposal-freeze 적대 보이스 1(task 0.5) → codex(task 0.6, Manager 슬롯).

## Manager 판정 (2026-09-29) — Q1 · Q2 · Q3 (사용자행 아님)

- **Q1 = (a)** — 정본 준수(`openspec/specs/exit-policy/spec.md:62`·`:65`)이지 신규 정책이 아니다. spec delta 무변경.
- **Q2 = 구현 로트로 이연 + 사전 승인** — 읽기 전용 시세 GET 1회(정지 종목 포함), 장중, 결과로 D6 분기 확정. 문서 표기 [미측정 · 사전 승인된 실측 대기].
- **Q3 = 초안 그대로** — 관측점 판정 진입. 하류 무음 5자리는 명명된 잔여(후속 change 후보).
- 반영: design 「Q — 결정 기록」·D5·D6·D1 · proposal 표 · tasks 0.7·0.8·2.3a.


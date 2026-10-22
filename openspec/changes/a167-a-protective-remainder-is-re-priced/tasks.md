# a167 Tasks

> **High-risk.** 등록 시점(2026-10-22) 구현 0 — 아래는 범위 고정용 골격이고, freeze 가 세분한다.
> 착수 조건: a087 Phase 2 착지(구간 입력 a087 tasks 2.6 포함). 그 전에는 0.x 의 문서 작업만 한다.

## 0. 선행

- [ ] 0.1 a087 Phase 2 착지 확인(보호 청산 유형 + 구간·접수 입력이 정본에 있다).
- [ ] 0.2 열린 사람 결정 5건 회신(proposal 「열린 사람 결정」 — N·정정 대 재제출·가격 규칙·MARKET 잔량·토글 기본값).
- [ ] 0.3 CodeGraph hard evidence + `tools/logic-map` AST — proposal 「필수 규율」의 함수 전부를 HEAD 에서 생성(a087 3.6 산출물 재사용 금지 — 빌린 증거는 낡는다).
- [ ] 0.4 design.md 작성 + spec 델타 재작성(초안 대체) + base commit 고정.
- [ ] 0.5 a100 design D8 · a087 design D9 이중 매도 계약 재대조.
- [ ] 0.6 **proposal-freeze — 다중 적대 리뷰**(적대적 Eng 필수; 이중 매도·무보호 창·취소 불명 레드팀).

## 1. 실측 (사람 승인 — §0.7, 자동 실행 금지)

- [ ] 1.1 실측 설계(부분체결을 만드는 수량·종목·시각) — 사람 승인.
- [ ] 1.2 KR 부분체결 잔량의 브로커 거동(시장가 잔량 즉시 취소/잔존) — 사람 터미널.
- [ ] 1.3 선택한 방식(정정 또는 취소 후 재제출)의 왕복 1회 — 사람 터미널. 종결 확인 필드를 기록.

## 2. 구현 (freeze 뒤 세분)

- [ ] 2.0 Pre-Edit 선언 (WORKFLOW §Pre-Edit 형식).
- [ ] 2.1 Function Logic Map + Branch Test Map (0.3 대상 전부).
- [ ] 2.2 RED — proposal §1~§6 의 불변식(종결 확인 전 재제출 금지 · 남은 수량 · 가격 인상 금지 · 익절 제외 · 구간 규칙 · in-doubt 무재제출).
- [ ] 2.3 GREEN.
- [ ] 2.4 변이 원장 — proposal 「필수 규율」 변이 전부 CAUGHT.
- [ ] 2.5 엔진 종단 시험 — 「잔량 종결 → 재제출 수량 = 잔량」.

## 3. 게이트

- [ ] 3.1 독립 적대 리뷰(구현과 분리된 컨텍스트).
- [ ] 3.2 `make sdd-sync` · `make sdd-check`.
- [ ] 3.3 격리 worktree 에서 `make gate CHANGE=a167-a-protective-remainder-is-re-priced`.
- [ ] 3.4 PM 동기화 → `openspec archive`.

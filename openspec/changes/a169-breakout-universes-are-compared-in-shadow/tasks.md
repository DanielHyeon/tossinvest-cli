# a169 tasks

> 2026-10-23: 열린 질문 Q1~Q12 답 받음 — 전부 권장 채택(proposal 「사람 결정」 표 · `issues.md` §1). 지금 착수 = 2.1(Q11) 뿐.
> 2.2 이후는 Q6 에 따라 a140 archive 뒤. 출처: 사용자 결정 2026-10-23(종목군 셋 · ETF 별도 · `securityType` 노출 · Q1~Q12) 과 a140 review §4.

## 0. 착수 준비

- [x] 0.1 proposal 열린 질문 Q1~Q12 에 사람 답을 받고 proposal 에 기록한다. — 2026-10-23 전부 권장 채택, proposal 「사람 결정」 표 · `issues.md` §1.
- [x] 0.2 단일 writer 배정, `python3 tools/sdd/capture_change_base.py --change a169-breakout-universes-are-compared-in-shadow` 로 base 고정.
  — writer Terra(2026-10-23), base `2bbee345f`.
- [ ] 0.3 `make sdd-sync` 후 CodeGraph hard evidence(`adaptStocks` · `domain.Quote` 소비자 · Q6 에 따른 a140 도구 심볼)와 증거 조정.
  proposal 의 줄 번호를 AST 로 재확인.

## 1. 계약 (design freeze)

- [ ] 1.1 design.md — Q1~Q12 답 반영(선정 시각 · U2 식 · 예산 · 중복 · 실행 형태 · 판정 임계 · 유형 갈래).
- [ ] 1.2 spec delta 확정(현재 초안: `specs/breakout-retest-strategy-lane/spec.md` · `specs/official-read-compatibility/spec.md`).
- [ ] 1.3 proposal-freeze 경량 리뷰 → `review.md`.

## 2. 구현 (RED → GREEN → REFACTOR → VERIFY)

- [ ] 2.1 `securityType` 노출(Q11 답이 「먼저」면 0.1 · 1.1 을 기다리지 않음) — `adaptStocks` FLM/BTM 또는 `Function Logic Map: not-applicable` 사유, RED(유형 칸 · 다른 생산자 바이트 불변) → GREEN.
- [ ] 2.2 종목군 선정(U0 · U1 · U2) · 공통 필터 · 유형 영수증 · 원장 행 — RED → GREEN.
- [ ] 2.3 Q6 이 (a) 면 a140 판(lot) digest 불변 시험.
- [ ] 2.4 VERIFY — 관련 패키지 시험 · vet · `make lint`.

## 3. 측정

- [ ] 3.1 (Q4 답에 따라) 리허설 1 세션(측정에 안 셈) — 호출 예산 · 분 수집 시간 실측(Q4).
- [ ] 3.2 KR 5 세션 측정 · 라벨 완주 · 원장 커밋(계좌 데이터 0 확인).

## 4. 판정 · 종결

- [ ] 4.1 보고서 — 종목군 × 유형 갈래 퍼널 · 결과 분포(n < 30 보류) · (Q5 답에 따라) 겹침 표 · a140 비교(인용).
- [ ] 4.2 사람 결정 기록.
- [ ] 4.3 sdd-sync → gate → archive → Story 경로 갱신.

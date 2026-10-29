# a169 tasks

> 2026-10-23: 열린 질문 Q1~Q12 답 받음 — 전부 권장 채택(proposal 「사람 결정」 표 · `issues.md` §1). 지금 착수 = 2.1(Q11) 뿐.
> 2.2 이후는 Q6 에 따라 a140 archive 뒤. 출처: 사용자 결정 2026-10-23(종목군 셋 · ETF 별도 · `securityType` 노출 · Q1~Q12) 과 a140 review §4.

## 0. 착수 준비

- [x] 0.1 proposal 열린 질문 Q1~Q12 에 사람 답을 받고 proposal 에 기록한다. — 2026-10-23 전부 권장 채택, proposal 「사람 결정」 표 · `issues.md` §1.
- [x] 0.2 단일 writer 배정, `python3 tools/sdd/capture_change_base.py --change a169-breakout-universes-are-compared-in-shadow` 로 base 고정.
  — writer Terra(2026-10-23), base `2bbee345f`.
- [x] 0.3 `make sdd-sync` 후 CodeGraph hard evidence(`adaptStocks` · `domain.Quote` 소비자 · Q6 에 따른 a140 도구 심볼)와 증거 조정.
  proposal 의 줄 번호를 AST 로 재확인.
  — 2026-10-28: `make sdd-sync` 는 **미실행**(Manager 지시 — worktree 전진 부작용), 기존 codegraph 1.6.0 색인 사용(지문 stale 기록) ·
  호출 자리 표 `analysis/code-context/` §0.3 · AST 13 `analysis/evidence/` · 줄 번호 drift(adaptStocks 85~103 등) 기록.

## 1. 계약 (design freeze)

- [x] 1.1 design.md — Q1~Q12 답 반영(선정 시각 · U2 식 · 예산 · 중복 · 실행 형태 · 판정 임계 · 유형 갈래). — 2026-10-28, HEAD `0c9bfd272`.
- [x] 1.2 spec delta 확정(현재 초안: `specs/breakout-retest-strategy-lane/spec.md` · `specs/official-read-compatibility/spec.md`).
  — 2026-10-28: breakout delta 에 선정 시각 · U2 식 · 귀속 · 기본 모드 판 불변 시나리오 추가, `openspec validate --strict` rc=0.
- [ ] 1.3 proposal-freeze 경량 리뷰 → `review.md`. — 2026-10-28 초안 `review.md` §3(셀프리뷰 1, **Manager 판정 대기** · 보이스 미정).
  — 2026-10-28 독립 리뷰(Opus) `review.md` §4: **FAIL — P1 2(일봉 경계 · 시작 호가 생략 전제)**. 수정 + 사람 재확인(D2) 뒤 재리뷰.
  사람 결정 D1~D4 는 `design.md` §13 · `issues.md` §2 에 기록.
  — 2026-10-28 **반영 완료 — 재리뷰 대기**: D2 를 「생략 안 함」으로 변경(사용자 「권장대로」), P1 2 · P2 6 · P3 7 을 design · spec · issues 에 반영
  (위치 `review.md` §5), 남긴 것 `issues.md` §3. 새 AST 2(`analysis/evidence/claims.md` 합계 15). 재리뷰는 다른 보이스가 한다.
  — 2026-10-28 재리뷰(GPT) `review.md` §6: **FAIL — P1 1(공유 키 MARKET_DATA 예산) · P2 2(`--top` · U2 마감)**.
  — 2026-10-29 **재리뷰 FAIL → 보완 완료(별도 키), 재리뷰 대기**: 사람 결정 D5(별도 앱 키, 사용자 「4는 가능」)로 P1-3 을 design §7.5 · §14 runbook ·
  spec 에, P2-1 을 design §1 에, P2-2 를 design §4.5 에 반영(위치 `review.md` §7), 남긴 것 `issues.md` §4(한도 단위 [미측정] 등).
  새 AST 5(합계 20). `openspec validate --strict` rc=0(2026-10-29). 다음 재리뷰는 다른 보이스가 한다.
  — 2026-10-29 재리뷰(GPT) `review.md` §8: **FAIL — P1 4(토큰 바이트 비교 · preflight 순서 · 실시간 엔진 429 차단 · 엔진 경로) · P2 2(경로 정규화 · 엔진 비밀)**.
  — 2026-10-29 **재리뷰 FAIL → 보완 완료(Manager 결정: 토큰 동등 판정 · `--profile-preflight` · 엔진 로그 감시기 · `ENGINE_CONFIG_DIR`), 재리뷰 대기**:
  위치 `review.md` §9, 남긴 것 `issues.md` §5, 새 AST 6(합계 26). `openspec validate --strict` rc=0.
  — 2026-10-29 **§10 재리뷰 FAIL(P1-8~10) — a166 착지까지 보류(사람 결정 2026-10-29)**. 리허설 · 측정 보류, 재개 조건 `issues.md` §6, 처분 `review.md` §11.

## 2. 구현 (RED → GREEN → REFACTOR → VERIFY)

- [x] 2.1 `securityType` 노출(Q11 답이 「먼저」면 0.1 · 1.1 을 기다리지 않음) — `adaptStocks` FLM/BTM 또는 `Function Logic Map: not-applicable` 사유, RED(유형 칸 · 다른 생산자 바이트 불변) → GREEN.
  — 2026-10-23: FLM/BTM `analysis/function-logic/internal-official--adaptstocks/` · RED 3 FAIL → GREEN · 영향 범위 시험 두 구성 rc=0 ·
  경량 리뷰 P3 1 수용(review §2) · 실측 STOCK/ETF 확인, ETN 미실측(Q12 로 이월).
- [ ] 2.2 종목군 선정(U0 · U1 · U2) · 공통 필터 · 유형 영수증 · 원장 행 — RED → GREEN.
  같은 로트에 design §7.5 의 `profile.go`(G1~G6 · 경로 정규화 · 좁은 reader · `--profile-preflight`) · `enginewatch.go`(감시기 · `abort` 행 · 재시작 거부)와
  RED §10.3-13 · 16 이 든다 — 감시기 없이 리허설 · 측정을 돌리지 않는다.
- [ ] 2.3 Q6 이 (a) 면 a140 판(lot) digest 불변 시험.
- [ ] 2.4 VERIFY — 관련 패키지 시험 · vet · `make lint`.

## 3. 측정

- [ ] 3.1 (Q4 답에 따라) 리허설 1 세션(측정에 안 셈) — 호출 예산 · 분 수집 시간 실측(Q4).
  선행: 사람이 design §14 runbook 1~4 를 마침 — `ENGINE_CONFIG_DIR` 확인, 별도 키 측정 프로필 등록, **`--profile-preflight` 통과 뒤에만** `openapi test`.
  실행 중 감시기가 켜져 있어야 하고(엔진 로그를 못 열면 도구가 시작 거부), 끝나면 §14-5 의 `abort` 여부 · `engine.log` 가 자랐는지 · 사후 계수 ·
  도구 rate limit 수(`abort` 없음 · 0 이어야 측정)와 U2 `status` 를 기록.
- [ ] 3.2 KR 5 세션 측정 · 라벨 완주 · 원장 커밋(계좌 데이터 0 확인).

## 4. 판정 · 종결

- [ ] 4.1 보고서 — 종목군 × 유형 갈래 퍼널 · 결과 분포(n < 30 보류) · (Q5 답에 따라) 겹침 표 · a140 비교(인용).
- [ ] 4.2 사람 결정 기록.
- [ ] 4.3 sdd-sync → gate → archive → Story 경로 갱신.

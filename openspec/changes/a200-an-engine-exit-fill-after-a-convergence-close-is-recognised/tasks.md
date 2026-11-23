# a200 tasks — 수렴 조정이 닫은 포지션에 닿은 엔진 청산 체결의 인정

위험 등급 High-risk. 단계를 건너뛰지 않는다. 각 체크는 그 산출물과 같은 커밋에서 한다.
범위 개정 2026-11-24(Manager 결정 — 1단계 재현 뒤): 성과 · 운영 로그 · 수량 저장 제외, 편집할 기존 함수는 `ProjectPosition` 하나(`review.md`).

## 1. 등록 · 증거 (1단계)

- [x] 1.1 change 디렉터리 등록: `.openspec.yaml` · proposal · design · tasks · spec delta(아래 1.9 뒤에 조립) · Story `STORY-TOS-a200.yaml` · `_registry.yaml` · tracker 재생성
- [x] 1.2 `base-commit.txt`(change 루트 — `capture_change_base.py` 가 쓰는 자리). 1단계 값은 a198 착지 전 `884be2a74`. a198 착지 · `origin/main` merge 뒤 「사람 승인 base 재고정」(Manager 승인 2026-11-24 — 귀속 실측 · 승인 기록 · 단독 커밋 + 영수증) — `189367808`(884be2a74 → 28680af3d, 단독 커밋), 실측 `analysis/base-refix/attribution.txt`
- [x] 1.3 사전 증거를 현재 HEAD 로 재확인해 `analysis/evidence-phase0.md` 로(scratch `a200-evidence/` — AST 재추출, 달라진 곳 표) + 운영 집계 요약(Manager 제공 — 건수·종류·날짜만)
- [x] 1.4 **재현**: 조정 tx 뒤 엔진 매도 체결(W4)을 실제 원장·실제 훅으로 만드는 시험을 쓰고 편집 전 코드에서 돌려 `ATTRIBUTION_FAILED` 행 1 · `COMPLETED` 0 · 성과 0 을 확인(원문 `analysis/repro/w4.txt`). W5·W6 은 실제 `Tracker` 로
- [x] 1.5 **범위 밖 재현 시도**(결과만 보고, 수리 안 함): W2·W3 의 수량 불일치 고착, 부분 조정 뒤 초과 체결 — 재현됨/안 됨/판정 불가와 그 근거(`analysis/repro/out-of-scope.md`)
- [x] 1.6 design 전제 확인(하나라도 거짓이면 멈춤): (a) `SELL_ON_CLOSED` 는 `resolveFillOrigin` 통과 뒤에만 생긴다 (b) `exit_events.action` 에 스키마 제약이 없다(DDL 인용) (c) Project 훅 시점에 `pending_intent_id` 가 살아 있다(부분·종결 체결 모두) (d) `pending_intent_id` 를 쓰는 경로 전수(exit 제출 · flatten · 운영자 매도 · 편입) (e) 조정 `kind` 값 전수와 수렴이 쓰는 값 (f) 체결이 「주문이 결속된 인스턴스」를 찾는 방법 — 재매수로 새 인스턴스가 생긴 뒤에도 옛 인스턴스를 가리키는가
- [x] 1.7 CodeGraph hard evidence(버전 기재): `ProjectPosition` · `enterReconcileInTx` · `ApplyExitFill` · 성과 동결 함수 · `closeExitStateOnAdjustmentTx` 의 호출자와 호출 **자리 수**
- [x] 1.8 편집 표면 · 고정 장치 표(`analysis/edit-surface.md`): 편집할 기존 함수, 구조 시험·census·허용 목록 적중, `exit_events.action` 소비자와 표지 행의 영향, 진행 중 change 와의 파일 겹침(a197 · a087 · a167 · a171)
- [x] 1.9 정본 delta 조립: position-ledger ADDED(인정 규칙) + MODIFIED 「포지션 상태기계」(생성 스크립트 + 역치환 단언), reconciliation ADDED(인정된 체결은 귀속 실패 차단을 만들지 않는다 · 어긋나면 지금처럼). ADDED 본문 500자 이하. `openspec validate --strict` valid
- [x] 1.10 편집할 기존 함수마다 FLM/BTM pre-edit 번들(`ProjectPosition`)
- [x] 1.11 claim check(`analysis/claim-check.md`)
- [x] 1.12 제안 동결 적대 리뷰(codex, 최대 2회차) → `review.md`
- [x] 1.13 Manager freeze 2026-11-24 — (라) 수량 상한, codex 2회차 P1 2건은 출처 교체로 해소, 전제 (i)~(iii) 확인(`review.md`). a198 merge · base 재고정은 1.2

## 2. 구현

- [x] 2.1 Pre-Edit 선언(High-risk) — `analysis/tdd/pre-edit.md`
- [x] 2.2 RED — design §3 의 1~5(1.4 의 본 사고 재현 시험을 차단 0 기대로 뒤집는 것이 첫 줄, 「조정을 사이에 둔 부분 체결」 재현 시험은 그대로 PASS) — 원문 `analysis/tdd/red.txt`
- [x] 2.3 GREEN 최소 구현 — 새 파일(인정 판정 · 표지 쓰기) · `ProjectPosition` B5 1 자리. `ApplyExitFill` · 전이표 · 수렴 · 대사 해제 규칙 · 스키마 무편집. 주석은 한글로 「왜」 — 원문 `analysis/tdd/green.txt`
- [x] 2.4 post-edit 번들 · `check_analysis` — `ProjectPosition`(revision current) + 재현 시험 파일의 지운 시험 3(revision base) · 자리만 옮겨진 1(revision current), `analysis/verify/check-analysis.txt`(evidence complete)
- [x] 2.5 변이 원장(design §3-6) — 16/16 CAUGHT(2 회차), 1 회차 M2 생존은 범위 안 시험 보강으로 처분(`analysis/mutation/`)
- [x] 2.6 선택 시험 · `-race`(해당) · `make lint` · strict validate · tracker `--check` — 원문 `analysis/verify/`

## 3. 리뷰 · 검증

- [ ] 3.1 구현 적대 리뷰(별도 문맥, Opus + codex, 최대 2회차) — 초점: 인정이 통과시키면 안 되는 체결 · 누적 합의 출처 · 인정 뒤 수량/차단 상태
- [ ] 3.2 처분 반영
- [ ] 3.3 gstack `review` 1회
- [ ] 3.4 Manager 검증

## 4. 종결

- [ ] 4.1 착지 기록 → `origin/main` merge(rebase 금지) → 완료 게이트(격리 worktree) → archive → Story 경로 → push

## 5. 종결 뒤 (사람 승인 — 체크박스 아님)

- 운영 반영은 사람이 승인한다. 반영 뒤 확인은 관측으로: 다음에 같은 경합이 났을 때 귀속 실패 행 0 · 인정 표지 행.

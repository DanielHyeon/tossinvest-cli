# a127 tasks

> **High-risk (위험 권한 · 진입 경로의 적재기).** 판독 조건을 바꾸는 change 다 — 권한을 넓히지 않음(design D6)을 증명하기 전에는 GREEN 이 없다.
> freeze 전 구현 착수 금지.

## 0. Freeze (구현 착수 금지 게이트)

- [x] 0.1 base 고정 — `f9a25549`(신설 시점 HEAD)
- [x] 0.2 결함 실측 — 핀 두 자리 AST(`analysis/freeze-ast/`), 생산 진입점 · 읽기 전용 원장 열람 census(proposal), 읽기 집합의 현재 원장 성립 ·
      prepare 실패 대조(`analysis/measurements/readset-probe.log`), v28~v35 의 읽기 집합 변경 전수(design 「증거 기반」)
- [x] 0.3 범위 판정 — route 적재기 포함(Manager 2026-10-01), D1 = (b) 주입된 현재 버전과 정확 일치(Manager 재판정 2026-10-01, (a) 대체)
- [x] 0.4 design · proposal · delta(strategy-runtime ADDED 1) 작성
- [ ] 0.5 proposal-freeze 리뷰 — 독립 적대 리뷰(fail-open 축 · 증거 규율) + codex 교차 모델, 처분 반영
  - [x] 0.5.1 1라운드(review 0.5.1): 보이스 APPROVE-WITH-FIXES(P1 2 · P2 6 · P3 6) · codex FAIL(P1 1 · P2 3 · P3 1), 처분 → design 2판
  - [ ] 0.5.2 재리뷰(codex 슬롯은 Manager 요청제)
- [ ] 0.6 openspec validate --strict(리뷰 반영 뒤 재실행)
- [ ] 0.7 Manager freeze 승인

## 1. 구현 (freeze 승인 뒤)

- [ ] 1.0 착수 전
  - [ ] 1.0.1 base 재확인 — a112 6.1 (C) lint 가 같은 파일의 다른 함수(`validProductionRiskPolicyContents`)를 먼저 착지(Manager 통지 2026-10-01):
        그 해시 위에서 진행. 형제 착지가 창에 들어오면 사람 승인 재고정
  - [ ] 1.0.2 편집 대상 함수의 gate FLM 번들(편집 전) + Pre-Edit 선언 — `loadProductionRiskEntries` · `openProductionRouteSnapshot` ·
        두 config 검증 자리 · engine 의 두 호출 자리(`strategyRiskAuthorityLoader.collectMarket` · route 적재기의 config 리터럴 함수),
        그리고 AST 가 추가로 잡는 것 전부
- [ ] 1.1 RED — design 「반증 설계」 S1~S13 을 실패시키는 시험: 실제 원장 양성 risk(a112 트립와이어 반전) · route(**새 기반**: `tossos_testseams`
      서명 매니페스트 작성 seam + 외부 시험 패키지 `strategyrouter_test` 가 `journal.Open` 원장으로 Batch 호출), 더 새 · 더 옛 거절 + 방향 문구
      (route 는 Batch 경계), 주입 0 · 미설정 거절을 존재하지 않는 원장 경로로 관측, 열 삭제 원장 거절(route 는 active owner 없는 범위 포함 —
      표 재생성 픽스처), 스키마 거절의 신원(ScopeRefused 아님), engine 두 자리 · risk 판독 tx 의 구조 단언
- [ ] 1.2 GREEN — 최소 구현(D1~D3 · D7): config 필드 · 0 이하 거절(열기 전) · 정확 일치 비교 · 방향 문구, route `:352` 감싸기의 원인 보존,
      risk 판독 읽기 tx 하나, route 판독 전 prepare(같은 SQL 상수), 리터럴 상수 둘 삭제, engine 두 자리 주입
- [ ] 1.3 픽스처 — 축소 원장 셋(`riskbucket/production_snapshot_authority_test.go:219` · `strategyrouter/production_test.go:216` ·
      `app/engine/strategy_risk_authority_test.go:224`)이 주입 값과 같은 `user_version` 을 쓰게 바꿈 — riskbucket · strategyrouter 는 journal 버전처럼 보이지
      않는 값(예: 1), 행 · 단언 불변. `riskbucket/a126_snapshot_digest_replay_test.go:6` 주석 갱신. a112 시험 다리 `a112MirrorLedgerIntoRiskStub` 제거 ·
      거래 픽스처 위험 적재기 실제 원장 단일화, a112 동결 증거가 트립와이어를 인용하는 자리(`analysis/measurements/lot-5.2.2.2-fix/pre-edit/…collectmarket/branch-test-map.md:14` ·
      `analysis/harness/render_5222_bundles.py:215`)와 `collectMarket` 편집으로 밀리는 a112 FLM 번들의 재기준 — a112 소유자와 조율(Manager 통지)
- [ ] 1.4 변이 S1~S13 전수(무변이 대조군 GREEN · 시작 sha · 시험 파일 sha 판마다 단언)
- [ ] 1.5 회귀(riskbucket · strategyrouter · app/engine · journal, 무태그 + `tossos_testseams`, `-race`), `make lint`
- [ ] 1.6 독립 리뷰(보이스 + codex) · 처분

## 2. 게이트

- [ ] 2.1 격리 워크트리 `make gate CHANGE=a127-strategy-authorities-read-the-current-ledger`
- [ ] 2.2 아카이브(Manager 검증 뒤)

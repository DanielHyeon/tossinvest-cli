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
- [ ] 0.6 openspec validate --strict(리뷰 반영 뒤 재실행)
- [ ] 0.7 Manager freeze 승인

## 1. 구현 (freeze 승인 뒤)

- [ ] 1.0 착수 전
  - [ ] 1.0.1 base 재확인(형제 착지가 창에 들어오면 사람 승인 재고정)
  - [ ] 1.0.2 편집 대상 함수의 gate FLM 번들(편집 전) + Pre-Edit 선언 — `loadProductionRiskEntries` · `openProductionRouteSnapshot` ·
        두 config 검증 자리 · engine 의 두 호출 자리(`strategyRiskAuthorityLoader.collectMarket` · route 적재기의 config 리터럴 함수),
        그리고 AST 가 추가로 잡는 것 전부
- [ ] 1.1 RED — design 「반증 설계」 S1~S8 을 실패시키는 시험: 실제 원장 양성 두 적재기(a112 트립와이어를 양성 시험으로 뒤집음), 더 새 · 더 옛 ·
      주입 0 거절(두 적재기), 읽기 집합 열 삭제 원장 거절, engine 호출 자리가 `journal.SchemaVersion` 상수를 넘긴다는 구조 단언
- [ ] 1.2 GREEN — 최소 구현(D1 · D2 · D3): config 필드 · 0 이하 거절 · 정확 일치 비교 · 방향 문구, 리터럴 상수 둘 삭제, engine 두 자리 주입
- [ ] 1.3 픽스처 — 축소 원장 셋(`riskbucket/production_snapshot_authority_test.go` · `strategyrouter/production_test.go` · `app/engine/strategy_risk_authority_test.go`)이
      주입 값과 같은 `user_version` 을 쓰게 바꿈(행 · 단언 불변), a112 시험 다리 `a112MirrorLedgerIntoRiskStub` 제거와 a112 거래 픽스처의 위험 적재기
      실제 원장 단일화(a112 소유 파일 — Manager 통지 뒤)
- [ ] 1.4 변이 S1~S8 전수(무변이 대조군 GREEN · 시작 sha · 시험 파일 sha 판마다 단언)
- [ ] 1.5 회귀(riskbucket · strategyrouter · app/engine · journal, 무태그 + `tossos_testseams`, `-race`), `make lint`
- [ ] 1.6 독립 리뷰(보이스 + codex) · 처분

## 2. 게이트

- [ ] 2.1 격리 워크트리 `make gate CHANGE=a127-strategy-authorities-read-the-current-ledger`
- [ ] 2.2 아카이브(Manager 검증 뒤)

# a127 tasks

> **High-risk (위험 권한 · 진입 경로의 적재기).** 판독 조건을 바꾸는 change 다 — 권한을 넓히지 않음(design D6)을 증명하기 전에는 GREEN 이 없다.
> freeze 전 구현 착수 금지.

## 0. Freeze (구현 착수 금지 게이트)

- [x] 0.1 base 고정 — `f9a25549`(신설 시점 HEAD)
- [x] 0.2 결함 실측 — 핀 두 자리 AST(`analysis/freeze-ast/`), 생산 진입점 · 읽기 전용 원장 열람 census(proposal), 읽기 집합의 현재 원장 성립 ·
      prepare 실패 대조(`analysis/measurements/readset-probe.log`), v28~v35 의 읽기 집합 변경 전수(design 「증거 기반」)
- [x] 0.3 범위 판정 — route 적재기 포함(Manager 2026-10-01), D1 = (b) 주입된 현재 버전과 정확 일치(Manager 재판정 2026-10-01, (a) 대체)
- [x] 0.4 design · proposal · delta(strategy-runtime ADDED 1) 작성
- [x] 0.5 proposal-freeze 리뷰 — 독립 적대 리뷰(fail-open 축 · 증거 규율) + codex 교차 모델, 처분 반영
  - [x] 0.5.1 1라운드(review 0.5.1): 보이스 APPROVE-WITH-FIXES(P1 2 · P2 6 · P3 6) · codex FAIL(P1 1 · P2 3 · P3 1), 처분 → design 2판
  - [x] 0.5.2 2라운드(review 0.5.2): codex FAIL(P2 3 · P3 1) · 협대역 보이스 APPROVE-WITH-FIXES(P1 1 · P2 3 · P3 6), 처분 → design 3판
  - [x] 0.5.3 3라운드 협대역 codex(review 0.5.3): FAIL(P1 1 · P2 2 · P3 1), 처분 → design 4판
  - [x] 0.5.4 4판 — Manager 직접 판정(추가 codex 라운드 불요, review 0.7)
- [x] 0.6 openspec validate --strict — 4판 valid
- [x] 0.7 Manager freeze 승인(2026-10-01, review 0.7)

## 1. 구현 (freeze 승인 뒤)

- [x] 1.0 착수 전
  - [x] 1.0.1 **(review 0.7)** base 재고정 `f9a25549` → `3403be28`(a112 6.1 `40ec5aff` 포함). 원 지시: base 재확인 — a112 6.1 (C) lint 가 같은 파일의 다른 함수(`validProductionRiskPolicyContents`)를 먼저 착지(Manager 통지 2026-10-01):
        그 해시 위에서 진행. 형제 착지가 창에 들어오면 사람 승인 재고정
  - [x] 1.0.2 **(review 1.0.2)** 편집 대상 함수의 gate FLM 번들(편집 전) + Pre-Edit 선언 — `loadProductionRiskEntries` · `openProductionRouteSnapshot` ·
        두 config 검증 자리 · engine 의 두 호출 자리(`strategyRiskAuthorityLoader.collectMarket` · route 적재기의 config 리터럴 함수),
        그리고 AST 가 추가로 잡는 것 전부
- [x] **(review 1.1~1.4)** 1.1 RED — design 「반증 설계」 S1~S14 를 실패시키는 시험(픽스처 규율 — 버전만 바꾼 온전한 원장 · 조건부 질의 전용 열 삭제): 실제 원장 양성 risk(a112 트립와이어 반전) · route(**새 기반**: `tossos_testseams`
      서명 매니페스트 작성 seam + 외부 시험 패키지 `strategyrouter_test` 가 `journal.Open` 원장으로 Batch 호출), 더 새 · 더 옛 거절 + 방향 문구
      (route 는 Batch 경계), 주입 0 · 미설정 거절을 존재하지 않는 원장 경로로 관측, 열 삭제 원장 거절(route 는 active owner 없는 범위 포함 —
      표 재생성 픽스처), risk 는 latch 가 선 범위 + 사용량 전용 열 삭제, 스키마 거절의 신원(ScopeRefused 아님), engine 두 자리(import 해석 AST 선택자 — go/types 아님, review 1.6.1) ·
      risk 의 버전 · latch · 사용량 판독이 같은 tx 라는 구조 단언, 주입 음수
- [x] **(review 1.1~1.4)** 1.2 GREEN — 최소 구현(D1~D3 · D7): config 필드 · 0 이하 거절(열기 전) · 정확 일치 비교 · 방향 문구, route `:352` 감싸기의 원인 보존,
      risk 판독 읽기 tx 하나(버전 · latch · 사용량 전부), 두 적재기 판독 전 prepare(같은 SQL 상수, risk 는 latch early return 앞), 리터럴 상수 둘 삭제,
      engine 두 자리 주입
- [x] **(review 1.1~1.4)** 1.3 픽스처 — 축소 원장 셋(`riskbucket/production_snapshot_authority_test.go:219` · `strategyrouter/production_test.go:216` ·
      `app/engine/strategy_risk_authority_test.go:224`)이 주입 값과 같은 `user_version` 을 쓰게 바꿈 — riskbucket · strategyrouter 는 journal 버전처럼 보이지
      않는 값(예: 1), 행 · 단언 불변. `riskbucket/a126_snapshot_digest_replay_test.go:6` 주석 갱신. a112 시험 다리 `a112MirrorLedgerIntoRiskStub` 제거 ·
      거래 픽스처 위험 적재기 실제 원장 단일화, a112 동결 증거가 트립와이어를 인용하는 자리(`analysis/measurements/lot-5.2.2.2-fix/pre-edit/…collectmarket/branch-test-map.md:14` ·
      `analysis/harness/render_5222_bundles.py:215`)와 `collectMarket` 편집으로 밀리는 a112 FLM 번들의 재기준 — a112 소유자와 조율(Manager 통지)
- [x] **(review 1.1~1.4)** 1.4 변이 S1~S14 전수 — 하네스는 무태그와 `tossos_testseams` 스위트를 모두 돈다(S2 의 route 양성 시험이 태그 뒤)(무변이 대조군 GREEN · 시작 sha · 시험 파일 sha 판마다 단언)
- [x] **(review 1.1~1.4)** 1.5 회귀(riskbucket · strategyrouter · app/engine · journal, 무태그 + `tossos_testseams`, `-race`), `make lint`
- [x] 1.6 **(review 1.6.1 · 1.6.2 — 수리 묶음 변이 27/27)** 독립 리뷰(보이스 + codex) · 처분

## 2. 게이트

- [x] 2.1 **[처분 2026-10-01] 게이트 — 이 체크 커밋에서 격리 워크트리로 `make gate CHANGE=a127-strategy-authorities-read-the-current-ledger` 를 돌리고 결과를 `review.md` 에 적는다**(a095 6.8 관례: 게이트는 자기 줄도 미완료로 세므로 체크 → 실행 → 기록)
- [x] 2.2 **[처분 2026-10-01] 아카이브는 Manager 최종 검증 · 승인 뒤** — 아래 「아카이브 때 할 일」

## 사람 항목 (체크박스 아님 — 이 change 의 게이트 밖)

- **H1** — 배포 전 생산 설정 재실측(전략 매니페스트 digest 환경값 · scheduler 활성화 · automation gate 상태): design D6 「오늘 동작 변화 0」의 영수증.
  a127 착지 뒤에는 핀이라는 우연한 차단이 사라지므로, 서명 매니페스트 발급 · schedule 활성화 전에 사람이 확인한다.

## 아카이브 때 할 일 (Manager 승인 뒤 — 체크박스 아님)

- archive 커밋 메시지와 `review.md` 에 착지 `82080177`(+ 수리 묶음 `1e25b3a3`) · 착지 기록 `a8c0f978` · 게이트 실행 커밋을 인용한다.
- 델타 ADDED(strategy-runtime) 적용 확인, Story openspec.path → 아카이브 경로, tracker 재생성.
- 사용자 보고: D6 경보 수위(「핀이라는 우연 차단이 사라지고 설계된 조건 사슬만 남는다」, 사람 · 운영 조건 12 — 거래 정책 · LIVE 마스터 스위치 · 공식 자격 증명 포함)와 H1.

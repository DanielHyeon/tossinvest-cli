# a126 tasks

> **High-risk (사이징·원장).** 감소는 진입을 여는 방향이다 — 모든 감소가 원장 사실에
> 결속됨을 증명하기 전에는 GREEN 이 없다. freeze 전 구현 착수 금지.

## 0. Freeze (구현 착수 금지 게이트)

- [x] 0.1 base 고정 — `ed9d3685` → `989ab031`, 단독 커밋 `d8ed3de1`(자기 Go 커밋 0 실측, 승인 참조: 사용자 2026-09-28
      결정 ③ 분리 신설 + 2026-09-30 "남은것도 처리" 재개 지시 — review.md 「0.1」)
- [x] 0.2 원장 스키마 실측 — 귀속 열·쓰기 자리 census(design.md 「증거 기반」·Q1). persist 좌표 재확인: filled 쓰기
      `risk_bucket_fill.go:1068`, latch 플래그 `:1071`, owner latch `:1087`
- [x] 0.3 Q1~Q4 확정 — design.md. Q2 는 파생·소급 대상 0 건 실측, 활성화 직전 재실측 = 사람 승인 항목 H1(표기만)
- [x] 0.4 design 작성 + AST 선행 — `analysis/freeze-ast/`(18 함수, base blob 대조). gate FLM 번들은 편집 대상 함수에 대해
      1.1 에서 편집 **전**에 만든다(아래 1.0.2)
- [ ] 0.5 proposal-freeze 리뷰
  - [ ] 0.5.1 독립 적대 리뷰(다관점) — 공격 대상 필수: "감소가 원장 사실 없이 일어날 수 있는 경로 0" · overage latch
        재계산과의 순서 · D5 문장과의 정합
  - [ ] 0.5.2 codex 교차 모델 — Manager 슬롯(직렬화) 대기
- [ ] 0.6 openspec validate --strict
- [ ] 0.7 사람 확인 H2 — D8 "자동은 조이기만" 의 범위가 잠금·latch 해제이고 원장 사실에 결속된 사용량 감소는 포함하지
      않는다는 해석(design.md D7)

## 1. 구현 (freeze 뒤 — 별도 로트, High-risk 규율 전부)

- [ ] 1.0 착수 전
  - [ ] 1.0.1 사람 승인 base 재고정(형제 커밋 `55963f29` 이 창에 들어옴 — design.md R4)
  - [ ] 1.0.2 편집 대상 기존 함수의 gate FLM/BTM 번들(`readProductionRiskUsage`·`aggregateProductionRiskUsage`·
        `refuseStaleBucketUsage`·`smallestRecordedBucketLimit`, 그리고 AST 가 추가로 잡는 것 전부) + Pre-Edit 선언
- [ ] 1.1 RED — design.md 「반증 설계」 M1~M8 각각을 실패시키는 시험: 떠남(양성) · 활성 owner 불변 · 영수증 없는 해제 표식 ·
      late BUY 되돌림(RecordFill 경로와 전략 정산 경로 **각각**) · 한도 모집단 동행 · 손상 떠난 행 거절 · latch 집계 유지 ·
      부분 매도 불변
- [ ] 1.2 GREEN — 최소 구현(D1 판정 한 곳, D3 같은 루프)
- [ ] 1.3 변이(무변이 대조군 + 양성 대조군) — "감소를 무조건 수행"(M1) 변이가 반드시 잡혀야 한다(fail-open 축). M1~M8 전부
- [ ] 1.4 a066 owner-lifecycle 픽스처의 filled_minor 가림 주석 해소(`risk_bucket_owner_test.go:802–804`, 실값 픽스처로)
- [ ] 1.5 4보이스 리뷰 + gstack

## 2. 게이트

- [ ] 2.1 격리 워크트리 `make gate CHANGE=a126-filled-exposure-leaves-the-bucket`
- [ ] 2.2 아카이브 (Manager 검증 뒤)

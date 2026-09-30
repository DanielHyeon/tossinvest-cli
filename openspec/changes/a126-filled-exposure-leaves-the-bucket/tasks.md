# a126 tasks

> **High-risk (사이징·원장).** 감소는 진입을 여는 방향이다 — 모든 감소가 원장 사실에
> 결속됨을 증명하기 전에는 GREEN 이 없다. freeze 전 구현 착수 금지.

## 0. Freeze (구현 착수 금지 게이트)

- [x] 0.1 base 고정 — `ed9d3685` → `989ab031`, 단독 커밋 `d8ed3de1`(자기 Go 커밋 0 실측, 승인 참조: 사용자 2026-09-28
      결정 ③ 분리 신설 + 2026-09-30 "남은것도 처리" 재개 지시 — review.md 「0.1 base 재고정 승인」)
- [x] 0.2 원장 스키마 실측 — 귀속 열·쓰기 자리 census(design.md 「증거 기반」·Q1). persist 좌표 재확인: filled 쓰기
      `risk_bucket_fill.go:1068`, latch 플래그 `:1071`, owner latch `:1087`
- [x] 0.3 Q1~Q4 확정 — design.md. Q2 는 파생·소급 대상 0 건 실측, 활성화 직전 재실측 = 사람 승인 항목 H1(표기만)
- [x] 0.4 design 작성 + AST 선행 — `analysis/freeze-ast/`(24 함수, base blob 대조). gate FLM 번들은 편집 대상 함수에 대해
      1.1 에서 편집 **전**에 만든다(아래 1.0.2)
- [x] 0.5 proposal-freeze 리뷰 — codex C3 는 Manager 재판정(2026-09-30)으로 채택, 3.1 을 면제 불가 의존으로 바꿈
  - [x] 0.5.1 독립 적대 리뷰(R1 fail-open · R2 원장 순서 · R3 증거 규율) — 셋 다 APPROVE-WITH-FIXES, P0 0. 처분 반영
        design 2판(review.md 「0.5.1」)
  - [x] 0.5.2 codex 교차 모델 — 실행·기록 완료(review.md 「0.5.2」). 판정 **BLOCK**(P1 3 · P2 2). 다섯 전부 처분:
        C1·C2 설계·잔여·시험 핀, C3 는 Manager 재판정으로 채택(3.1), C4·C5 반영
- [x] 0.6 openspec validate --strict — 2판 델타 valid
- [x] 0.7 사람 확인 H2 — Manager 판정(2026-09-30): 사용자 결정 ③ 인용으로 확인, 새 사람 질의 불요, 사용자 최종 보고에
      해석 명시(거부권 유보) — design.md D7

## 1. 구현 (freeze 뒤 — 별도 로트, a092 뒤 큐, High-risk 규율 전부)

- [x] 1.0 착수 전
  - [x] 1.0.1 **(`b30318d6`, review 1.0.1)** 사람 승인 base 재고정(형제 커밋 `55963f29`·`f48e7865` 이 창에 들어옴 — design.md R4)
  - [x] 1.0.2 **(`467322df` 편집 전 5 번들 + Pre-Edit review 1.0.4. a066 결함 수리로 `releaseRiskBucketOwner` · `loadRiskBucketFillTransition` 추가 — Manager 판정 (가), review 1.0.3)** 편집 대상 기존 함수의 gate FLM/BTM 번들(`readProductionRiskUsage`·`aggregateProductionRiskUsage`, 주석만 고치는
        `refuseStaleBucketUsage`, 그리고 AST 가 추가로 잡는 것 전부) + Pre-Edit 선언
- [x] 1.1 **(review 1.1~1.4 대응표 · `analysis/impl/red-*.log`)** RED — design.md 「반증 설계」 M1·M2·M3·M3b·M4·M6·M6b·M7·M8 각각을 실패시키는 시험:
  - 떠남(양성) · 활성 owner 불변 · 영수증 없는 해제 표식
  - late BUY 되돌림 — RecordFill 경로와 전략 정산 경로 **각각**, Campaign hook 결선 전제를 시험 이름에 명시
  - 되돌림이 서지 않는 경로 (a) 소유 모호 · (b) 증분 판독 불가의 잔여 핀(해제 owner 에 scope latch 없음을 단언)
  - 떠난 행의 작은 한도가 큰 선언 한도의 진입을 계속 거절(D3)
  - 손상 떠난 행(HELD·held≠0) · released_at 불일치 · `r.*≠d.*` 거절, 무관한 활성 owner 의 체결은 커밋되고 scope latch 만
  - latch 집계 유지 · 부분 매도 불변
  - (codex #4) 영수증 × scope latch(되돌림) × HELD/held≠0 조합 — 되돌려진 행도 손상 검사를 받음
  - (codex #5) replay 결정성: 해제 전 · 해제 뒤 · 되돌림 뒤 · 공유 owner 체결과 교차한 순서에서 재시작 replay 후 사용량·
    snapshot digest·latch 가 같음, 건강한 활성 owner 체결이 `total < own` 을 만들지 않음
  - 잔여 핀(서지 않음/뚫림을 **단언**해 배선 로트가 바꾸면 깨지게): (f) 두 generation 이 같은 order id 를 쓸 때 옛 owner 가
    떠난 채 남음(codex #1), 해제 → 다른 종목 발급 → late 체결 되돌림 → 제출 재검증 통과(codex #2)
- [x] 1.2 **(review 1.1~1.4)** GREEN — 최소 구현(D1 판정 한 곳 — `readProductionRiskUsage` SQL + `aggregateProductionRiskUsage`). `RowDigest` 형식
      불변(D2). `risk_bucket_usage.go:59–62` 주석을 "한도 모집단은 떠난 행을 포함한다(a126 D3)" 로 정정(동작 불변)
- [x] 1.3 **(21/21 CAUGHT, 대조군 GREEN — review 표)** 변이(무변이 대조군 + 양성 대조군) — "감소를 무조건 수행"(M1) 변이가 반드시 잡혀야 한다(fail-open 축). 위 M 전부
- [x] 1.4 **(생산 작성자 경로로 전량 체결 — a066 수리 뒤 가능해짐)** a066 owner-lifecycle 픽스처의 filled_minor 가림 주석 해소(`risk_bucket_owner_test.go:802–804`, 실값 픽스처로)
- [x] 1.5 **(review 1.5 · 1.5.2 · 1.5.3)** 4보이스 리뷰 + gstack (gstack 은 별도 문맥 리뷰로 대체 — review 1.5 에 사유)
  - [x] 1.5.1 **(review 1.5)** 3 보이스 + codex 합본 — 생산 결함 0, codex FAIL(시험 공백 P2 셋), 처분안 R1~R16
  - [x] 1.5.2 **(review 1.5.2 · 27/27 CAUGHT)** 수리 로트(Manager 승인: R2=(가) · R1 · R3~R14 · R15/R16 기록) — 시험 · 문서만, 생산 코드 변경 0 · 변이 전수 재실행
  - [x] 1.5.3 **(review 1.5.3 · VERDICT PASS)** codex resume 재리뷰

## 2. 게이트

- [x] 2.1 **[처분 2026-10-01] 게이트 — 이 체크 커밋에서 격리 워크트리로 `make gate CHANGE=a126-filled-exposure-leaves-the-bucket` 를 돌리고 결과를 `review.md` 2.1 에 적는다**(a095 6.8 관례: 게이트는 자기 줄도 미완료로 세므로 체크 → 실행 → 기록)
  - [x] 2.1.1 **(review 2.1 — Manager 승인 2026-10-01, 형제 착지 개재)** 사람 승인 base 재고정 `a189e74f` → `97a6f717` + 시험 함수 경량 번들 다섯 + 착지 재기록
- [x] 2.2 **[처분 2026-10-01] 아카이브는 Manager 최종 검증 · 승인 뒤** — 아래 「아카이브 때 할 일」(a095 6.10 관례)

## 아카이브 때 할 일 (Manager 승인 뒤 — 체크박스 아님)

- archive 커밋 메시지와 `review.md` 에 착지 `958f8239` · 게이트 실행 커밋을 인용한다.
- Story openspec.path → 아카이브 경로, tracker 재생성.
- 사용자 거부권 항목(a066 해제 검사 수리 · H2 해석)과 H1(운영 원장 영수증 재실측) 표식을 완료 보고에 유지한다.

## 3. 다른 로트에 거는 조건 (이 change 가 체크하지 않음 — 기록용)

- 3.1 **owner 해제 배선 로트(a066 잔여 #5, 활성화 로트)는 R3 가족이 해소되기 전 착수할 수 없다 — 면제 불가 의존**
  (Manager C3 재판정 2026-09-30, R1 #1 (a) 의 "또는 사람 수용" 을 대체). 해소의 정의:
  - 되돌림이 서지 않는 경로 (a)–(f) **각각**이 떠남을 보수적으로 무효화하거나, 그 경로가 불가능함을 증명한다.
  - C2 구멍(떠남 뒤 발급된 주문이 되돌림 뒤 한도 초과 상태로 제출됨)을 닫는다.
  - 영수증~late 체결 창과 a066 #8(late 체결이 공유 bucket 을 latch 하지 않음)을 같은 기준으로 닫는다.

  허용되는 대안 형태: 떠남 판정을 **기본 OFF**(원장에 기록된 래치로 켜고 끔)로 싣고, R3 해소 뒤에 켠다. 이때 OFF 는 오늘의
  누적 합과 같아야 한다(§0.2). 사람 승인 항목 H1(운영 원장 영수증 재실측)은 어느 형태든 배선 전에 선행한다. 이 조건을 여는
  것은 이 문구가 아니라 사용자의 명시적 재결정이다.

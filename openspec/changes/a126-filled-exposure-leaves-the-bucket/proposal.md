# a126 · 체결 노출은 bucket 을 떠난다

- **Feature**: `FEAT-TOS-013` — Multi-horizon risk buckets
- **Story**: `STORY-TOS-a126`
- **Spec**: `multi-horizon-risk-buckets` (a066 아카이브가 신설한 정본)
- **위험 등급**: **High-risk** (사이징·원장 — bucket 사용량은 진입 cap 의 분모다)

> **출처**: a066 6.5 독립 적대 리뷰의 P1 기능 발견(2026-09-28, 4보이스)과 사용자 결정
> (2026-09-28 — "사용량 수명주기는 독립 후속 change 로 분리"). 이 초안은 Manager 가
> 작성했고, 분기 열거는 담지 않는다 — freeze 로트가 AST 산출물로 세운다.

## Why

**`filled_minor` 는 한 번 쌓이면 영영 줄지 않는다. 그래서 모든 bucket 이 평생 누적
cap 이 된다.**

a066 6.5 실측(아카이브 review.md:1124 부근):

- 설계 D5 는 사용량을 "**Position 에 귀속된** filled 노출"로 정의한다.
- 구현은 **누적 매수 합**이다 — 포지션이 닫혀도, owner 가 해제돼도, 그 매수의
  `filled_minor` 는 bucket 사용량에 남는다.
- owner-lifecycle 시험 픽스처는 `filled_minor='0'` 이라 이 갭을 가린다(:1161 — 픽스처에
  주석으로 명기됨, a066 처분).

귀결:

1. 시간이 갈수록 진입이 조여진다 — 오래 운용한 계정일수록 같은 한도에서 실제 진입
   가능량이 줄어든다(fail-closed 라 **돈 위험은 없다**, 방향은 과잉 차단).
2. 한도 상향이 이력 있는 bucket 에서 실효를 못 갖는다 — 누적분이 이미 새 한도를
   삼킨다.
3. D5 문장과 구현이 갈린 채 정본이 됐다 — a066 는 이 갭을 수리하지 않고 명명했다
   (사용자 결정으로 분리).

## What Changes

**사용량의 수명주기를 정의한다** (설계 축 — freeze 로트가 확정): 포지션 귀속 노출이 소멸하는 사건이 생기면 그 몫의
`filled_minor` 가 bucket 사용량에서 빠진다. 자동 회계 의미론이며 **운영자 해제
(a066 완화 가족)와 다른 축**이다 — 사람 승인이 아니라 원장 사실이 트리거다.

## 열린 질문 (freeze 로트 입력) — **2026-09-30 freeze 로트가 확정, 답과 영수증은 design.md**

- Q1 → ② owner release receipt(①·③ 을 선결 조건으로 포함). Q2 → 파생(저장값·스키마 불변), 운영 원장 영수증 0 건 실측,
  활성화 직전 재실측은 사람 승인 항목 H1. Q3 → latch 해제 → owner 해제(=떠남) → 다음 체결의 재계산, 떠남은 latch 를
  풀지 않음. Q4 → 부분 종결 감소 없음(SELL→예약 귀속 열 없음), 최소 귀속 단위 = owner generation.

아래는 초안의 원 질문이다.

- **Q1 — 떠나는 사건은 무엇인가.** 후보: ① 포지션 종결(매도 체결로 보유 0) ② owner
  해제(release — 청결 검사 통과 시점) ③ 대사 확인 0 보유. 어느 것이 D5 의 "Position
  귀속"과 일치하는지 원장 스키마(reservations·owners·fills)의 실제 귀속 열로 정한다.
- **Q2 — 소급인가 전방인가.** 기존 누적 이력의 정정(마이그레이션 재계산) vs 전방 적용만.
  소급은 운영 원장 재계산이라 사람 승인 항목이 될 수 있다.
- **Q3 — RISK_OVERAGE 해제와의 상호작용.** a066 완화 경로가 latch 를 풀 때 사용량은
  그대로다(의도된 분리). 수명주기 감소가 latch 재계산(recomputeOverageLatches)과 어떤
  순서로 만나는지.
- **Q4 — 부분 체결·부분 종결.** 수량 단위 감소의 근거 열(귀속 가능한 최소 단위).

## Impact (freeze 로트 확정 2026-09-30 — design.md)

- **Specs**: `multi-horizon-risk-buckets` (ADDED — 떠남 요구 1, latch 불가침 요구 1)
- **Code**: 사용량 reader 한 곳 — `internal/riskbucket/production_snapshot_authority.go`
  (`readProductionRiskUsage`·`aggregateProductionRiskUsage`·`JournalBucketUsage`). 한도 모집단(`smallestRecordedBucketLimit`)은
  **바꾸지 않는다**(freeze 리뷰 처분 D3 — 주석 한 줄만). persist 쓰기 자리(`risk_bucket_fill.go` :1068·:1071·:1087)와 owner
  해제·완화·재계산 경로는 **편집하지 않는다**(파생 설계).
- **Schema**: 없음(영수증 v24·scope latch v22 재사용). 저장값 불변.
- **생산 효과**: owner 해제가 배선될 때까지 0(a066 잔여 #5).
- **§0.3·손절**: 무접촉 — 진입 cap 만 다룬다. 감소는 진입을 **여는** 방향이므로 모든
  감소는 원장 사실에 결속돼야 한다(느슨한 감소 = fail-open)

## Non-goals

- 운영자 수동 사용량 조정(완화 가족과 혼합 금지)
- 한도 선언 체계 변경(공유 값당 단일 한도는 a066 잔여 3b — 상류 매니페스트 검증 몫)
- v35 사용량 인덱스(활성화 로트 실측 몫)

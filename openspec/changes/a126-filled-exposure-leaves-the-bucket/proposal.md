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

## 열린 질문 (freeze 로트 입력)

- **Q1 — 떠나는 사건은 무엇인가.** 후보: ① 포지션 종결(매도 체결로 보유 0) ② owner
  해제(release — 청결 검사 통과 시점) ③ 대사 확인 0 보유. 어느 것이 D5 의 "Position
  귀속"과 일치하는지 원장 스키마(reservations·owners·fills)의 실제 귀속 열로 정한다.
- **Q2 — 소급인가 전방인가.** 기존 누적 이력의 정정(마이그레이션 재계산) vs 전방 적용만.
  소급은 운영 원장 재계산이라 사람 승인 항목이 될 수 있다.
- **Q3 — RISK_OVERAGE 해제와의 상호작용.** a066 완화 경로가 latch 를 풀 때 사용량은
  그대로다(의도된 분리). 수명주기 감소가 latch 재계산(recomputeOverageLatches)과 어떤
  순서로 만나는지.
- **Q4 — 부분 체결·부분 종결.** 수량 단위 감소의 근거 열(귀속 가능한 최소 단위).

## Impact (예상 — freeze 로트가 확정)

- **Specs**: `multi-horizon-risk-buckets` (MODIFIED — 사용량 정의에 수명주기 절)
- **Code**: `internal/journal/risk_bucket_fill.go`(persist 쓰기 자리 :1071·:1087 계열),
  `internal/riskbucket/` 재계산 경로
- **Schema**: Q2 가 소급이면 마이그레이션(additive 원칙), 전방이면 없을 수 있음
- **§0.3·손절**: 무접촉 — 진입 cap 만 다룬다. 감소는 진입을 **여는** 방향이므로 모든
  감소는 원장 사실에 결속돼야 한다(느슨한 감소 = fail-open)

## Non-goals

- 운영자 수동 사용량 조정(완화 가족과 혼합 금지)
- 한도 선언 체계 변경(공유 값당 단일 한도는 a066 잔여 3b — 상류 매니페스트 검증 몫)
- v35 사용량 인덱스(활성화 로트 실측 몫)

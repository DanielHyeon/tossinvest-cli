# a126 tasks

> **High-risk (사이징·원장).** 감소는 진입을 여는 방향이다 — 모든 감소가 원장 사실에
> 결속됨을 증명하기 전에는 GREEN 이 없다. freeze 전 구현 착수 금지.

## 0. Freeze (구현 착수 금지 게이트)

- [ ] 0.1 base 고정 (`capture_change_base.py --change a126-filled-exposure-leaves-the-bucket`)
- [ ] 0.2 원장 스키마 실측 — fills·reservations·owners 의 포지션 귀속 열 열거(Q1·Q4 입력),
      쓰기 자리 census(persist :1071·:1087 계열 — 좌표는 현재 HEAD 로 재확인)
- [ ] 0.3 Q1~Q4 확정 — 각 답은 코드·스키마 영수증 인용. Q2 소급이면 사람 승인 항목으로 표기
- [ ] 0.4 design 작성 + 대상 함수 FLM(AST 먼저 — 분기 주장은 이 산출물 뒤에만)
- [ ] 0.5 proposal-freeze 리뷰: 적대 보이스 1 + codex 교차 모델. 공격 대상 필수 항목:
      "감소가 원장 사실 없이 일어날 수 있는 경로 0" · overage latch 재계산과의 순서 ·
      D5 문장과의 정합
- [ ] 0.6 openspec validate --strict

## 1. 구현 (freeze 뒤 — 별도 로트, High-risk 규율 전부)

- [ ] 1.1 RED — 종결 사건 → 사용량 감소, 감소 없는 경로의 거부, 부분 종결 수량 귀속
- [ ] 1.2 GREEN — 최소 구현
- [ ] 1.3 변이(무변이 대조군) — "감소를 무조건 수행" 변이가 반드시 잡혀야 한다(fail-open 축)
- [ ] 1.4 a066 owner-lifecycle 픽스처의 filled_minor 가림 주석 해소(실값 픽스처로)
- [ ] 1.5 4보이스 리뷰 + gstack

## 2. 게이트

- [ ] 2.1 격리 워크트리 `make gate CHANGE=a126-filled-exposure-leaves-the-bucket`
- [ ] 2.2 아카이브 (Manager 검증 뒤)

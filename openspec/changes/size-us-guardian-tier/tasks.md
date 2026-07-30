# Tasks: size-us-guardian-tier

위험도 **High-risk** — 사이징 수치를 완화 방향으로 옮긴다(§0.9). 1.x는 계약 확인,
2.x 이후는 RED 선행이다.

## 1. 계약 확인 (구현 전)

- [x] 1.1 §0.1 확인 — 티어 추가가 기동 인터록의 판정 집합을 바꾸지 않음을
  호출부 추적으로 확인하고 design §0.1에 적는다
- [x] 1.2 §0.9 확인 — 완화임을 인정하고, 보존되는 것과 완화하지 않을 때의
  비용(백스톱 우회 압력)을 design §0.9에 적는다
- [x] 1.3 다섯 수치의 유도를 design D1에 적는다 — 실측 인용(M49)과 승인된 KRW
  집합 등가 양쪽, 그리고 등가 논증이 깨지는 지점
- [x] 1.4 Pre-Edit 선언 기록 (review.md)

## 2. 티어 등재 (internal/config)

- [x] 2.1 RED — `us-single-name`이 다섯 값·통화와 함께 등록돼 있고 라벨이 있다
- [x] 2.2 RED — USD 상한이 새 티어의 값으로 올라가고 **KRW 상한은 다섯 필드
  전부 불변**이다
- [x] 2.3 RED — 수량 상한 100주와 비율 상한 1%는 USD에서도 **불변**이다
  (완화가 두 필드로 새지 않는다)
- [x] 2.4 RED — 새 티어가 자기 상한을 통과한다 (경계 포함)
- [x] 2.5 RED — 새 티어가 기동 인터록 규칙(`Validate`)을 통과한다
- [x] 2.6 RED — 다섯 수치가 design D1의 유도와 일치한다: 노출 = 주문×3,
  일일 손실은 `us-small-live` 값 유지(§0.9 — 리뷰 A2에서 $75를 버렸다),
  비율·수량은 전 티어 공통값, 주문 상한은 KRW 등가 임계 2,000 아래
- [x] 2.7 RED — 측정된 계측기 1주가 새 상한 아래에서 헤드룸을 갖는다
  (M49의 $300 관측을 테스트가 인용한다)
- [x] 2.8 GREEN — 티어 한 줄 추가 + 출처 주석 (StockOS 아님을 명시)

## 3. 상한의 사정거리를 고정한다

- [x] 3.1 RED — 상한은 콘솔 쓰기 경로 전용이다: 상한을 넘는 블록도
  `Validate`(인터록 동치)는 통과한다 — 두 판정이 다른 질문이라는 것을 고정
- [x] 3.2 RED — 전사 검사가 StockOS 전사 티어 넷을 그대로 지키고, 새 티어는
  StockOS 출처가 아닌 것으로 분리해 검사한다
- [x] 3.3 GREEN — 테스트 갱신 (전사 검사·상한 검사)

## 4. 스펙

- [x] 4.1 `risk-management` `정책 수치의 provenance` 개정 — TossOS 실측 출처
  범주 + 승인된 USD 집합 + 상한이 콘솔 전용이라는 사실
- [x] 4.2 `operator-console` 티어 provenance 문구를 인용 권위와 일치시킨다
- [x] 4.3 `openspec validate size-us-guardian-tier --strict --no-interactive`

## 5. 검증

- [x] 5.1 `make test` / `make vet` / `make validate` — 상속 테스트 회귀 0
- [x] 5.2 Function Logic Map — 수정된 기존 함수 대상 산출 또는 명시적 면제
- [x] 5.3 review.md — 적대적 Eng 리뷰 (High-risk 필수)
- [x] 5.4 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=size-us-guardian-tier`
- [x] 5.5 PM 등록

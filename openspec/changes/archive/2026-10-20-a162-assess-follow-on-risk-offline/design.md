## Context

a156 전체 완료와 분리한 설계다. 사용자 승인: 2026-10-20 “오프라인 판정 모듈부터 설계”.
기존 근거는 [a156 선별 검증](../a156-gate-live-strategy-policy-rollout/analysis/offline-follow-on-verification-2026-10-20.md)이다. 이 검증을 신규 모듈의 시험으로 재사용하거나 구현 완료로 표시하지 않는다.

## Goals / Non-Goals

세 전략의 KR·US 후속 진입 후보를 기존 평가기로 얻고, 명시적인 시나리오 위험 상한으로 잔여 예산을 계산한다. 동일 native 입력에서 기존 proposal 결과를 보존해야 한다.

비목표: a161 JSON/복원 자료에서 native 객체·FX seal 생성, 원장 읽기·쓰기, 실계좌 조회, 실행용 cap·qFinal·reservation·lease 발급, engine 연결, sellable 조회, 수수료·환율·슬리피지 자동 추정, 마감 청산, 운영 승급. 기존 a100 실측 선행조건을 이 change로 우회하지 않는다.

## Input and authority boundary

- continuation/reversal/weekly 각각의 typed request adapter를 둔다. 신규 다형성 프레임워크·전략 선택 registry·외부 입력 parser는 만들지 않는다.
- 입력은 caller가 이미 가진 native request다. 시장은 KR/US만 받는다. continuation/reversal leg 2~3, weekly leg 2~7만 받으며 계획·전략별 기존 추가 제약을 유지한다.
- plan, risk, evidence, stop, execution terms를 새로 생성하거나 보정하지 않고 기존 `ProposeKR/ProposeUS`에 전달한다. weekly의 별도 reservation·market-week 근거도 기존 request에 필요하다.
- 외부 artifact를 native request로 바꾸는 API는 없다. caller는 호출 중 native map·pointer를 변경하지 않아야 하며 adapter는 입력을 수정하지 않는다. 동시 변경을 안전하게 캡처하는 기능은 제공하지 않는다.
- native plan의 비공개 필드를 읽기 위해 새 getter를 만들지 않는다. 동일 호출의 동일 native plan/risk에 기존 lane별 `CaptureOfflineEvidence`를 적용해 내부 조회 projection을 얻는다. `Plan.Budget`, `Plan.AccountCurrency`, `Risk.Filled/Held`를 읽고 projection을 외부로 반환하지 않는다. caller의 별도 detached evidence 입력은 받지 않는다. capture 실패 시 거절한다. capture 시각은 요청의 명시적 평가 시각과 일치시키며 현재 시각으로 최신성을 새로 부여하지 않는다. 기존 capture의 전체 복사·codec 검증 비용 및 크기 제한을 수용하며 그 경로를 재구현하지 않는다.
- 위험 가정 입력은 family·binding digest·market·account currency와 결속한다. continuation은 `RiskBudgetDigest`, reversal/weekly는 native `PlanDigest`이며, 기존 capture의 `Risk.PlanDigest`를 공통 비교값으로 사용한다. family 간 digest의 이름이나 생성 의미를 통일하지 않는다.
- `per_unit_risk_upper_bound_minor`는 양의 정수이며 **해당 시나리오에서** 거래 비용·슬리피지·FX 영향까지 포함한다고 caller가 가정한 계좌 통화 기준 주당 상한이다. 모듈은 상한의 경제적 타당성을 인증하지 않는다.
- 이 값은 native plan의 기존 per-share risk와 자동으로 동일시하지 않는다. 통화 변환·FX 갱신·상한 추정은 하지 않는다. 통화·시장·digest가 다르거나 가정값이 없으면 거절한다.

## Calculation and result

기존 proposal의 성공 decision 이외 결과는 모두 수량 없이 반환한다. native kind·code·invalidation code를 구분해 그대로 보존하고 예산 계산으로 거절/무효화를 우회하지 않는다. proposal 성공 후에도 위험 미확인·latch·scope 불일치나 음수·손상 금액을 정상으로 보정하지 않는다.

`B`=native 초기 위험예산, `F`=native 체결 위험, `H`=native 예약 위험, `U`=시나리오 주당 위험 상한:

```text
R = B - F - H
q_offline = min(q_candidate, floor(R / U))
planned_risk_minor = q_offline * U
```

`F+H>B`는 거절한다. R=0 또는 floor(R/U)=0이면 `NO_BUDGET`이며 주문 후보를 내지 않는다. 이전 leg의 미사용 수량을 재배분하지 않는다. q_candidate를 늘리거나 입력 객체에 새 수량을 써 넣지 않는다.

H는 입력에 이미 반영된 모든 미결 금액 위험이며 q_offline은 **H에 포함되지 않은 추가분**이다. 기존 예약 주문의 retry/replacement 수량을 재평가하는 용도가 아니다. 대상 예약을 H에서 빼거나 해제하지 않는다. weekly ReservationState는 진입 주차/슬롯의 native 증거이며 별도의 금액으로 H에 다시 더하지 않는다. 결과는 이 보수적 시나리오의 상한이지 최대 실주문 가능량이 아니다.

금액은 parse 전에 ASCII canonical decimal(0 또는 앞자리 0이 없는 숫자), 최대 78자리로 제한하고, parse 후 BitLen<=256을 확인한다. 계산은 표준 `math/big`를 사용하며 중간 합·곱의 256-bit 한도를 넘으면 거절한다. 수량은 uint64 범위이고 division 결과를 q_candidate로 제한한 뒤 변환한다. float·반올림·음수 clamp는 사용하지 않는다. 이 상한은 신규 산술 파서의 한도이며 기존 native proposal/capture 전체 입력 비용의 한도를 보증하지 않는다.

결과는 `status`, native kind/code/invalidation code(해당 시), `q_candidate`, 양수인 경우의 `q_offline`, B/F/H/R/U 및 planned risk를 갖는 검사 값이다. 실패한 계산의 수량은 absence로 표현하며 0을 성공한 허가처럼 반환하지 않는다. B/F/H/R 값이 계산 불가하면 unknown을 0으로 채우지 않는다.

모든 결과에 `inspection_only=true`, `ledger_currentness=unverified`, `risk_bound_source=caller_assumption`, `execution_authority=false`를 명시한다. 결과 타입은 기존 admission/dispatch가 받는 타입이 아니며 native plan/cap/terms를 포함하지 않는다. 성공은 “이 가정에서 예산 내 수량이 있음”일 뿐 “매수 가능”이 아니다.

## Examples / acceptance anchors

- B=1000, F=200, H=300, U=100, q_candidate=8 → R=500, q_offline=5, planned risk=500.
- 같은 입력에서 q_candidate=3 → q_offline=3; 남은 예산 때문에 후보를 확대하지 않음.
- F+H=1000 → NO_BUDGET; F+H=1001 → 거절.
- U 누락/0, 다른 통화·digest·시장, unknown risk 또는 기존 proposal 거절 → 수량 없음.
- KR/KRW·US/USD 및 native frozen-FX를 갖춘 US/계좌통화 조합: U는 항상 account currency이며 이 모듈은 환율을 계산하지 않음.

## Verification and implementation handoff

구현 writer는 Terra, 독립 reviewer와 gstack reviewer는 별도 컨텍스트다. Manager가 diff·시험 원문을 최종 확인한다. 신규 코드 전 source/CodeGraph·관련 AST·입력 타입과 dependency guard를 고정하고, 영향 범위의 RED/GREEN만 수행한다.

2026-10-20 구현 준비 확인: 기존 `tossos_testseams`의 KR·US fixture를 재사용한다. weekly fixture는 첫 leg 예약만 포함하므로 동일 build tag의 `internal/weeklyvaluelane/strategyflow_testseam.go`에 후속 leg fixture를 최소 확장한다. 기존 native 예약 API로 구성하고 seal을 조작하지 않는다. 기본 빌드에는 이 파일이 포함되지 않아야 하며 production constructor/getter·권한 경로는 추가하지 않는다. 기존 태그 fixture의 영향 회귀와 기본 빌드 제외를 별도로 확인한다.

시험: 세 lane × KR/US의 후속 ordinal 양성, refusal/invalidation 보존, family/digest/currency 불일치, 동일 native snapshot/capture 결속, unknown·overbudget·0 budget, 78/79자리·2^256-1/2^256·합 overflow·uint64보다 큰 quotient의 min-before-convert, 입력 불변성, weekly reservation 누락과 슬롯/금액 이중 계상 방지, 결과가 실행용 타입/생성자를 노출하지 않는 의존성 경계. 기존 엔진 wiring은 변경하지 않는다.

전체 시험·전체 gate는 요청받지 않았으므로 실행하지 않는다. 이 change의 strict OpenSpec 검사와 PM 문서 동기화는 제품 전체 시험과 구분한다. 현재 단계는 **로컬 구현 및 영향 범위 검증 완료**다. [구현 근거](analysis/implementation-evidence.md)와 [리뷰](review.md)를 참조한다. 실측·운영 연결·커밋·푸시·archive 조건은 별도로 남는다.

## Research / skill applicability / risks

search-first: 기존 세 lane의 proposal·배분·시험을 재사용한다. 새 외부 의존성·설치·라이선스 변경 없음. 산술만 Go 표준 라이브러리로 한정한다. 외부 registry 검색은 기존 구현으로 충분해 실시하지 않았다. Hindsight의 관련 recall은 비어 있었고 현재 source를 우선했다.

grill-me: 대상은 KR·US 세 전략, 입력은 same-process native request, 출력은 비권한 가정 수량으로 결정했다. 사용자 승인 범위 밖인 영속 저장·dispatch·실측은 분리한다. 가장 큰 위험은 caller의 U를 검증된 실제 위험으로 오인하는 것이며 모든 결과의 가정·미검증 표시와 비실행 타입으로 제한한다.

## 추가 통합 검증 — 2026-10-20

사용자가 weekly 3–7차, non-nil FX, 계좌 최신성 관련 검증 보완을 승인했다.
이번 변경은 시험과 `tossos_testseams` fixture에 한정한다. 생산 평가기,
운영 설정, 기존 측정 증거는 변경하지 않는다.

- weekly KR·US 각각 3–7차의 모든 선행 순번을 native 예약·양수 체결
  전이로 생성한다. 순번 숫자만 바꾸거나 승인 seal을 위조하지 않는다.
  체결 위험의 누적, 예약 위험, 이전 슬롯 소비 및 현재 슬롯 요구를 확인한다.
- 기존 native FX 생성·검증·capture 경로를 재사용해 non-nil FX 입력을
  구성한다. 평가 전후 불변성과 독립 변경 탐지, invalid FX 거절을 검증한다.
- 실제 계좌 조회는 하지 않는다. stale 입력의 native 거절과, 양성 결과도
  `LedgerCurrentness=unverified`, `ExecutionAuthority=false`임을 시험한다.
  합성 시간이 유효하다는 사실은 실계좌 최신성 증명이 아니다.
- 변경 fixture의 기존 직접 소비자 시험과 새 통합시험만 선택한다.
  전체 suite·gate는 실행하지 않는다. 기존 로그는 보존하고 추가 검증
  로그·소스 해시는 별도 파일에 남긴다.

search-first: 현 HEAD `3415cf89f713ccdf61604963f29da9e0d5116294`와
현재 미커밋 a162 코드에서 native fixture/FX/capture를 우선 조사한다.
로컬 rg·Go 도구 및 스킬을 사용할 수 있고 Hindsight recall은 결과가
없었다. 외부 라이브러리 설치·라이선스 변경이 필요하지 않은 시험 확장이다.
grill-me: 온라인 최신성 인증과 오프라인 경계 시험을 분리했다.
사용자가 승인한 관련 통합시험 범위 안에서 진행하며 추가 운영 승인을
추정하지 않는다. impeccable: N/A — test/documentation-only.

impeccable: N/A — UI/CLI 없이 backend 설계만 포함한다. DB migration/배포 rollback: N/A — schema·운영 변경 없음. 구현 rollback은 미연결 신규 패키지 제거이며 기존 runtime에 영향이 없어야 한다.

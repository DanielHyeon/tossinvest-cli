## Why

전략마다 초기 손절의 출처가 다르더라도 승인된 가격·수량·정책이 주문·원장·청산까지 같은 계약으로 이어져야 한다. 계산과 전달 경계의 불일치부터 검사한다.

## What Changes

- Initial protection and sizing share a versioned campaign contract.
- Quantity respects all existing caps and risk bases.
- Existing campaigns preserve their protection snapshots.
- 권장 순서 3/9; 선행 완료 조건: a148, a149.
- 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `risk-management`: 이 change의 추가 요구사항과 검증 계약.

## Impact

대상: internal/strategyproposal, strategyflow, continuationlane, reversallane, weeklyvaluelane, breakoutlane, risk, riskbucket, execgw, app/engine exit initialization and journal bindings.
위험도: High: sizing, protection, RiskIntent, reservations, journal and exit initialization.
기존 운영 값·정책을 자동 이행하지 않는다. 라이브러리 추가는 기본적으로 없다.

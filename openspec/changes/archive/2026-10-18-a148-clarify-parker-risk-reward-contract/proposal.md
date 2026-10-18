## Why

Parker의 공간 RR, 목표 RR, 실제 실현 RR를 같은 숫자로 해석하면 손절·수량·익절의 검증 기준이 달라진다. frozen v1 동작을 보존하면서 세 의미를 먼저 구분한다.

## What Changes

- Parker RR diagnostics distinguish space target and realized results.
- Parker RR diagnostics reject undefined risk.
- 권장 순서 1/9; 선행 완료 조건: 없음.
- 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `strategy-engine`: 이 change의 추가 요구사항과 검증 계약.

## Impact

대상: internal/strategyengine/{lane.go,contracts.go}, existing Parker fixtures; no activation or dispatch changes.
위험도: Low for documentation/diagnostics; High if any entry predicate or RiskIntent changes.
기존 운영 값·정책을 자동 이행하지 않는다. 라이브러리 추가는 기본적으로 없다.

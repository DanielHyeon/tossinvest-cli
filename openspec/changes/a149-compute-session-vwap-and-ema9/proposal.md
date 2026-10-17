## Why

세션 누적 VWAP와 기간·시드가 명확한 EMA9를 계산해야 손절·진입 실험을 같은 입력으로 비교할 수 있다. rolling window 값을 세션 지표로 대체 사용하지 않는다.

## What Changes

- Session features have explicit reset seed and time semantics.
- Incomplete session inputs cannot become ready indicators.
- VWAP slope and EMA period are reproducible.
- 권장 순서 2/9; 선행 완료 조건: a148.
- 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `strategy-engine`: 이 change의 추가 요구사항과 검증 계약.

## Impact

대상: internal/strategycandle, internal/strategymarket, internal/strategyengine/contracts.go; new pure calculation leaf, no production caller wiring.
위험도: High before wiring into production; pure calculations first.
기존 운영 값·정책을 자동 이행하지 않는다. 라이브러리 추가는 기본적으로 없다.

## Why

LVN/HVN을 계산하기 전에 데이터가 가격대별 실제 체결량을 재현할 수 있는지 확인해야 한다. 5분 OHLCV 분배 근사로 진짜 Volume Profile이라고 주장하지 않는다.

## What Changes

- Volume profile feasibility separates executions from proxies.
- Unsupported volume profile does not block non-profile research.
- 권장 순서 5/9; 선행 완료 조건: a149.
- 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `candidate-discovery`: 이 change의 추가 요구사항과 검증 계약.

## Impact

대상: Offline input schema/quality report; immutable provided trade/price-bin data; existing LVN feature input boundary.
위험도: Read-only offline data feasibility.
기존 운영 값·정책을 자동 이행하지 않는다. 라이브러리 추가는 기본적으로 없다.

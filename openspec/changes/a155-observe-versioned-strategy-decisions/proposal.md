## Why

오프라인 계산이 런타임 입력에서도 재현되는지 관측할 필요가 있다. 관측 결과를 주문·예약·운영 제어에서 분리한다.

## What Changes

- Observation has no execution authority.
- Observation failure cannot delay protection.
- Actual observation requires a separate registered session approval.
- 권장 순서 8/9; 선행 완료 조건: a149, a150, a153, a154.
- 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `strategy-runtime`: 이 change의 추가 요구사항과 검증 계약.

## Impact

대상: Separate observation adapter consuming immutable market inputs; no gateway/reservation writer capability.
위험도: High isolation boundary; observation must have no order authority.
기존 운영 값·정책을 자동 이행하지 않는다. 라이브러리 추가는 기본적으로 없다.

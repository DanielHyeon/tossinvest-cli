## Why

개별 계산 정확성이 수익성 증명은 아니다. 비용과 의존 표본을 반영한 비교로 채택·보류를 결정한다.

## What Changes

- Strategy comparisons are preregistered and cost aware.
- Insufficient evidence never produces a promoted policy.
- Ablations preserve research boundaries.
- 권장 순서 6/9; 선행 완료 조건: a151, a152.
- 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `strategy-optimization`: 이 change의 추가 요구사항과 검증 계약.

## Impact

대상: Reuse offline evaluation artifacts from a151; immutable datasets and reports; no additional collection.
위험도: Offline evaluation; empirical claims require independent holdout evidence.
기존 운영 값·정책을 자동 이행하지 않는다. 라이브러리 추가는 기본적으로 없다.

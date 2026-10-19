## Why

0.7% 고정과 구조·변동성 손절을 비교할 때 진입·익절까지 동시에 바꾸면 원인을 구분할 수 없다. 손절만 바꾸는 사전 고정 비교부터 수행한다.

## What Changes

- Stop-policy comparisons vary one decision at a time.
- KR/KRW와 US/USD의 초기 손절만 비교한다. EMA9 청산은 후속 별도 실험으로 보류한다.
- Ambiguous intrabar paths are disclosed.
- 권장 순서 4/9; 선행 완료 조건: a150.
- 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `strategy-optimization`: 이 change의 추가 요구사항과 검증 계약.

## Impact

대상: Offline KR/US research tool consuming immutable fixtures and explicit session/tick/cost manifests; no live strategy imports with side effects. a149 KR-only 계산 및 운영 경로는 변경하지 않는다.
위험도: Offline research only; no trading code invocation.
기존 운영 값·정책을 자동 이행하지 않는다. 라이브러리 추가는 기본적으로 없다.

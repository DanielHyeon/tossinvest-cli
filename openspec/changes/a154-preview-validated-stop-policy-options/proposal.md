## Why

사용자는 초기 손절과 실제 청산 정책을 구분해서 선택해야 한다. 기존 설정 lifecycle을 확장하고 검증되지 않은 옵션이 운영 기본값이 되지 않게 한다.

## What Changes

- Validated initial-stop choices reuse settings lifecycle.
- Preview makes risk consequences visible without activation.
- Approved selection cannot widen existing positions.
- 권장 순서 7/9; 선행 완료 조건: a150, a153.
- 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `strategy-optimization`: 이 change의 추가 요구사항과 검증 계약.

## Impact

대상: Existing optimization descriptor registry, preview/CAS lifecycle and console; reuse common exit policy picker.
위험도: High at settings apply boundary; UI preview itself is read-only.
기존 운영 값·정책을 자동 이행하지 않는다. 라이브러리 추가는 기본적으로 없다.

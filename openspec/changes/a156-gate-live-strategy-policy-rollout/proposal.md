## Why

새 정책을 실거래에 연결하려면 최초 진입뿐 아니라 후속 leg·부분 체결·마감·복구까지 같은 안전 계약을 충족해야 한다. 개발 완료와 운영 승인을 분리한다.

## What Changes

- Live rollout is separately authorized and evidence bound.
- Later legs and restarts preserve campaign protection.
- Intraday flattening is scoped and fill verified.
- Rollback preserves exits and historical state.
- 권장 순서 9/9; 선행 완료 조건: a155.
- 기존 frozen v1, r1, a128/a147 증거, 인증·토큰 공유, 원장과 위험 통제는 보존한다. 이 문서는 운영 설정 변경·실주문·재시작·추가 세션 수집 승인이 아니다. Manager(Astra)가 설계를 작성·검토하고, Terra Teammate가 코드·테스트를 담당한다. 구현 writer와 독립 reviewer는 분리한다. gstack 리뷰, 해당 change 전체 gate, Manager 최종 확인 후에만 Terra가 범위 한정 커밋·푸시·정식 archive/PM sync를 수행한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `strategy-runtime`: 이 change의 추가 요구사항과 검증 계약.

## Impact

대상: Existing activation manifests, official gateway, Guardian, campaign/reservations, fill reconciliation and protective exit; scope refreshed before implementation.
위험도: High; NOT READY for live activation or weakening protection.
기존 운영 값·정책을 자동 이행하지 않는다. 라이브러리 추가는 기본적으로 없다.

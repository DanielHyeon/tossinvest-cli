## Why

후속 진입을 설계하려면 현재 campaign·position·leg·watermark가 같은 시점의 상태인지 먼저 확인해야 한다. 기존 정본의 분할·예산 규칙은 유지하며, 원장에 없는 전체 계획 원문을 0 위험 상태나 역산으로 복원하지 않는다.

## What Changes

- 기존 ReadOnly 핸들에 특정 계좌·시장·종목·campaign·실제 position generation의 일관된 상태 조회를 추가한다.
- 하나의 읽기 transaction에서 position/version, campaign stop/latch/version, ordered legs/watermarks를 읽는다.
- 전체 immutable plan/budget preimage는 복원 불가로 명시한다. 결과는 다음 진입의 수량이나 허가가 아니다.
- synthetic DB에서 범위/세대 격리, 동시 writer와 snapshot 일관성, 오류와 읽기 무변경을 검증한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `position-campaigns`: 권한을 만들지 않는 coherent continuation-state read 계약 추가.

## Impact

새 journal reader/test와 기존 ReadOnly 메서드 허용목록 시험의 이름 한 개 등록, OpenSpec·PM·검증 자료만 변경한다. migration·기존 writer·engine 연결·CLI/API/UI·브로커 호출·의존성 추가는 없다. a156의 live 승인/실증과 a157은 제외한다. 원본 checkout·기존 증거·실거래 및 운영 설정은 보존한다.

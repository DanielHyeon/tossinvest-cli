## ADDED Requirements

### Requirement: 종목별 정책 쓰기는 명확한 범위와 결과를 표시한다
콘솔은 a050의 `position-management` 카테고리에서 override/release/re-adopt 전 대상 계좌·시장·종목·현재 정책·기본 정책·실효 정책·다음 정책·적용 generation과 적용 시점/restart 필요 여부를 표시해야 한다 (SHALL).

#### Scenario: 정책 승인 화면
- **WHEN** 운영자가 종목 정책 변경을 연다
- **THEN** 기존 포지션 재해석 여부와 LIVE toggle이 바뀌지 않음을 승인 전에 표시한다

#### Scenario: version conflict
- **WHEN** write가 412로 거부된다
- **THEN** 현재 version을 다시 불러오도록 안내하고 자동 재시도하지 않는다

### Requirement: 종목 관리 설정은 기본값과 동작을 가까이 설명한다
`position-management`는 `종목별 정책`과 `외부 매수 자동편입`을 구분하고 모든 설정에 label, 쉬운 설명, 기본값, desired/effective 값, 단위·범위와 적용 시점을 표시해야 한다 (SHALL).

#### Scenario: 초기 자동편입 설정
- **WHEN** 저장 설정이 없는 새 설치가 화면을 연다
- **THEN** 자동편입 OFF, 합성 손절폭 5%, 허용범위 2~20%와 0.5% 조정 단위, 빈 include/exclude 및 exclude 우선을 표시한다

#### Scenario: 종목별 정책 미지정
- **WHEN** 한 종목에 override가 없다
- **THEN** 기본값과 실효값은 `공통 정책 상속`으로 표시되고 특정 preset을 복제 저장하지 않는다

#### Scenario: 1주 외부 보유
- **WHEN** 자동편입 preview 대상 보유가 1주다
- **THEN** `중간 익절 없음 · 보호선 승격 · 최종/손절 시 1주 전량` 설명을 표시한다

#### Scenario: release 요청
- **WHEN** 운영자가 자동관리 해제를 선택한다
- **THEN** 일반 설정 저장과 분리된 danger confirmation에서 보호 공백과 active exit 충돌 여부를 설명하고 3초 대기·명시 checkbox/button을 제공하되 문구 입력을 요구하지 않는다

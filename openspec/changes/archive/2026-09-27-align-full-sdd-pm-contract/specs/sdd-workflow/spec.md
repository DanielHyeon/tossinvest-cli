## ADDED Requirements

### Requirement: Full SDD 단계의 완전한 실행
비자명 변경은 다음 Full SDD 순서를 준수해야 한다(MUST): 기억 회고, Story와 승인된 OpenSpec 계약, CodeGraph hard evidence,
CodeGraphContext 보조 문맥, evidence reconciliation, 기존 함수 의존 시 Function Logic
Map·Branch Test Map, Pre-Edit Gate, RED→GREEN→REFACTOR→VERIFY, review/security/QA,
archive와 PM sync, 검증된 memory 승격 순서로 수행해야 한다(MUST). TossOS
`docs/WORKFLOW.md`는 각 단계의 입력·산출물·차단 조건을 명시해야 한다(MUST).

#### Scenario: evidence reconciliation 없이 편집 시도
- **WHEN** CodeGraph와 CodeGraphContext 결과가 다르지만 현재 HEAD 확인과 조정 기록 없이 구현을 시작하면
- **THEN** Pre-Edit Gate는 편집을 차단한다

#### Scenario: archive 이전 완료 선언
- **WHEN** gate만 통과하고 Story sync와 OpenSpec archive를 수행하지 않은 change를 완료로 보고하면
- **THEN** Full SDD 완료 조건 미충족으로 간주한다

### Requirement: TossOS 프로젝트 특성 보존
StockOS Full SDD 방법론을 동기화할 때 TossOS 프로젝트 특성을 보존해야 한다(MUST): Go AST·go test/race/vet 도구,
공식 Open API 주문/WTS 조회 경계, upstream OFF 회귀 계약, journal single-writer,
Guardian·reconciliation·인증 안전, TossOS 전용 graph/memory namespace와 Docker 배포
진입점을 유지해야 한다(MUST). StockOS의 Python/KIS 명령이나 STK portfolio 사실을
TossOS 정본으로 복사해서는 안 된다(MUST NOT).

#### Scenario: StockOS 방법론 동기화
- **WHEN** `docs/WORKFLOW.md`를 StockOS 정본에 맞춰 갱신하면
- **THEN** SDD 단계는 정렬되지만 TossOS 전용 도구·안전·배포 문단은 유지된다

### Requirement: PM 진행 상태의 증거 기반 파생
portfolio 원본은 계층, intent, Story 계약만 보관해야 하고(MUST), 진행 상태는 OpenSpec proposal,
tasks 체크 상태와 active/archive 위치에서 결정적으로 파생되어야 한다(MUST).
generated tracker는 수동 편집되어서는 안 된다(MUST NOT).

#### Scenario: tasks가 모두 완료된 활성 change
- **WHEN** Story의 change가 active이고 tasks에 미완료 항목이 없으며 완료 항목이 존재하면
- **THEN** generator는 Story 상태를 implemented로 파생한다

#### Scenario: archive된 change
- **WHEN** Story의 change가 archive 아래로 이동하면
- **THEN** generator는 Story 상태를 archived로 파생한다

## MODIFIED Requirements

### Requirement: PM 계층과 OpenSpec 역추적
모든 활성 OpenSpec change는 정확히 하나의 TossOS Delivery Story와 연결되어야 하고(MUST),
각 Story는 정확히 하나의 OpenSpec change ID와 경로를 가져야 한다(MUST). bootstrap
allowlist 또는 무기한 예외는 허용되지 않는다(MUST NOT). generator는
initiative→epic→feature→story 양방향 참조, Story→change와 change→Story 1:1,
고아·중복·잘못된 경로를 검사해야 한다(MUST). StockOS의 PM 데이터·기억·인덱스를
TossOS 사실로 복제해서는 안 된다(MUST NOT).

#### Scenario: Story 없는 활성 OpenSpec change
- **WHEN** 활성 change가 어떤 TossOS Story에도 연결되지 않으면
- **THEN** PM check가 실패하고 누락 change ID를 출력한다

#### Scenario: 하나의 change를 가리키는 복수 Story
- **WHEN** 두 Story가 같은 OpenSpec change ID를 가리키면
- **THEN** PM check가 1:1 위반으로 실패한다

#### Scenario: bootstrap 예외 등록 시도
- **WHEN** registry에 활성 change를 우회하는 bootstrap allowlist가 추가되면
- **THEN** PM check가 예외 계약을 거부한다

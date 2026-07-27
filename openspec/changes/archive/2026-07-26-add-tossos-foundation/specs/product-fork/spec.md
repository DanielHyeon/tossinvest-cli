# product-fork Specification (delta)

## ADDED Requirements

### Requirement: Upstream 히스토리 보존 Fork
TossOS 저장소는 upstream(JungHoonGhae/tossinvest-cli)의 전체 Git 히스토리를 보존한 clone에서 출발해야 하며(SHALL), GitHub 공개 fork network에는 포함하지 않는다.

#### Scenario: fork 저장소 검증
- **WHEN** `git log`로 히스토리를 조회하면
- **THEN** upstream의 원본 커밋 이력이 그대로 존재한다
- **AND** `git remote -v`에 `upstream`이 JungHoonGhae/tossinvest-cli를 가리킨다

### Requirement: 베이스라인 커밋 고정
fork 시점의 upstream 커밋 해시는 docs/baseline.md에 기록되어야 하며(SHALL), 이후 모든 회귀 판단의 기준점이 된다.

#### Scenario: 베이스라인 조회
- **WHEN** docs/baseline.md를 열면
- **THEN** 고정 커밋 해시, Go 버전, `go build`/`go vet`/`go test` 결과, 패키지별 커버리지가 기록되어 있다

### Requirement: Upstream 테스트 회귀 금지
TossOS의 모든 변경은 upstream에서 상속한 기존 테스트를 통과 상태로 유지해야 한다(SHALL).

#### Scenario: 변경 후 전체 테스트
- **WHEN** 임의의 change 구현 후 `go test ./...`를 실행하면
- **THEN** upstream 상속 테스트를 포함한 전체 테스트가 통과한다

### Requirement: 선별적 Upstream 동기화
지정 범주(보안 수정, 공식 API 스펙 변경, WTS endpoint·응답 파서 수정, 인증·세션 수정, 조건주문 수정, regression probe 개선) 밖의 upstream 변경을 일괄 merge해서는 안 된다(SHALL NOT). upstream 반영은 `upstream-sync` 브랜치에서만 수행하고, 반영·미반영 결정은 `docs/upstream-sync-log.md`에 기록되어야 한다(SHALL).

#### Scenario: upstream 반영 수행
- **WHEN** upstream 커밋을 선별 반영하면
- **THEN** upstream-sync 브랜치에서 merge·검증 후 main에 반영되고, docs/upstream-sync-log.md에 대상 커밋과 사유가 기록된다

### Requirement: Upstream push 차단
`upstream` remote로의 push는 기계적으로 차단되어야 한다(SHALL): push URL은 비활성 값(DISABLED)으로 설정된다.

#### Scenario: upstream push 시도
- **WHEN** `git push upstream`을 실행하면
- **THEN** 유효하지 않은 push URL로 인해 실패한다

### Requirement: 라이선스 준수
저장소는 upstream의 MIT LICENSE 파일과 원저작권 고지를 유지해야 한다(SHALL).

#### Scenario: 라이선스 확인
- **WHEN** 저장소 루트를 조회하면
- **THEN** 원저작권자가 명시된 MIT LICENSE가 존재한다

### Requirement: 시크릿 보호
세션 토큰·인증 정보·계좌 식별 정보가 담긴 파일은 커밋에서 제외되어야 한다(SHALL).

#### Scenario: 시크릿 파일 스테이징 시도
- **WHEN** 세션/설정 시크릿 경로를 `git add` 하면
- **THEN** .gitignore 규칙에 의해 무시된다

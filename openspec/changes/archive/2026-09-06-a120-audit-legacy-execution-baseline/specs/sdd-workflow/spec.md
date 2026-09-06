## MODIFIED Requirements

### Requirement: Full SDD 권위 계층
모든 신규 기능·동작 변경은 OpenSpec 계약, 현재 HEAD와 CodeGraph hard evidence, CodeGraphContext 보조 문맥, 기존 함수 내부 변경 시 Go AST·ast-grep Function Logic Map, Superpowers TDD, gstack 게이트 순서로 수행되어야 한다(SHALL). CodeGraphContext·GBrain·기억·관측 그래프는 advisory이며 OpenSpec, 현재 HEAD, 테스트, gstack을 대체해서는 안 된다(SHALL NOT).
일반 변경의 함수 분석 비교 기준은 불변 `base-commit.txt`여야 한다(SHALL).
아래 명시적 legacy 실행 기준선 이관의 모든 조건을 만족하는 a063만 고정된 실행
기준 E로 판정할 수 있으며, 원래 구현 전 증거 규칙을 소급 충족했다고 보고해서는
안 된다(SHALL NOT). 원래 계획 기준 P는 계속 보존해야 한다(SHALL).

#### Scenario: 보조 문맥과 현재 HEAD 충돌
- **WHEN** CodeGraphContext 또는 기억 결과가 현재 HEAD와 다르면
- **THEN** 현재 HEAD와 CodeGraph를 다시 확인·동기화한 뒤 그 결과를 구현 근거로 사용한다

#### Scenario: 기존 함수 내부 변경
- **WHEN** 기존 함수의 분기·early return·mutation·side effect를 변경하면
- **THEN** 구현 전에 Function Logic Map과 Branch Test Map을 만들고 변경 후 source hash·함수·분기와 묶인 증거로 최신화한다

#### Scenario: Function Logic Map 면제 시도
- **WHEN** `not-applicable` 면제를 기록했지만 비교 기준 대비 기존 Go 함수가 수정되었다
- **THEN** gate는 수정 함수를 diff에서 계산하고 해당 함수의 완전한 증거 묶음이 없으면 실패한다

#### Scenario: 실행 기준선 이관 기록이 없는 변경
- **WHEN** 변경에 execution-baseline 이관 기록이 없으면
- **THEN** 기존 불변 기준과 전체 수정 함수 분석 규칙을 그대로 적용한다

## ADDED Requirements

### Requirement: 고정된 legacy 실행 기준선 예외
The checker SHALL permit execution-baseline adoption only for
`a063-align-attestation-renewal-profile` with planning base
`da80ce31b6a1ab5d443016768f970a82bab102db` and execution base
`e65e394bf84b3c6e4559a219e816af96d341d75d`. It SHALL preserve the planning
base file, verify its regular committed bytes contain the full fixed P, and reject another change, base pair, malformed record or moving
snapshot reference. Invalid adoption SHALL fail closed rather than fall back.
The result SHALL be labeled `execution-baseline adoption exception` with
`retrospective-exception` provenance. The environment SHALL NOT select a base:
`SDD_BASE_REF` SHALL match only the valid record-derived effective base.

#### Scenario: 나중 기준으로 구현 변경을 숨기려는 시도
- **WHEN** 이관 기록이 허용된 E보다 뒤의 커밋이나 다른 변경 ID를 지정하면
- **THEN** 검사는 실패하고 원래 P를 덮어쓰거나 환경 변수로 우회하지 않는다

#### Scenario: 유효한 예외의 CI 기준
- **WHEN** 유효한 이관 기록의 E와 다른 커밋으로 SDD_BASE_REF를 설정하면
- **THEN** 검사는 기준 불일치로 실패한다

### Requirement: 이관 이력의 전수 회계와 독립 검토
An adoption SHALL retain a strict versioned ledger containing the complete
topologically ordered P..E reachable commit range, each commit's full parent
list and path/status differences against every parent (including merge parents
and root changes), and the net P-to-E modified-existing Go function
inventory, and the complete E..S Go path/function inventory. The validator SHALL
recompute these from immutable Git objects and reject missing, duplicate or
altered entries, ambiguous JSON, invalid schemas or digest mismatches. The
inherited range SHALL be identified as committed historical work with any
missing original analysis still outstanding; it SHALL NOT be labeled completed
FLM or a waiver. A committed adoption record SHALL bind the ledger and two
distinct, actually performed adversarial and subsequent gstack review documents
by repository-contained regular paths and SHA-256. The draft generator SHALL NOT
create approval claims, overwrite output or change a baseline or checkout.

#### Scenario: 과거 함수 한 개가 원장에서 빠짐
- **WHEN** 기록의 요약 건수와 해시가 있더라도 재계산한 P..E 함수가 원장에 없으면
- **THEN** 검사는 이력 누락으로 실패한다

#### Scenario: 소급 준수 주장
- **WHEN** 과거 누락 분석을 완료 또는 면제로 표시하거나 구현 전 작성된 증거라고 주장하면
- **THEN** 해당 이관 기록은 유효한 예외 증거로 인정되지 않는다

#### Scenario: 검토 자료 교체
- **WHEN** 원장이나 검토 문서가 지정 해시 또는 커밋된 내용과 다르거나 심볼릭 링크이면
- **THEN** 검사는 자료 교체로 실패한다

### Requirement: 커밋된 소스 스냅샷과 전체 함수 의무
Adoption acceptance SHALL run at clean detached HEAD H with full-commit source
snapshot S and verified P <= E <= S < H ancestry. H SHALL differ from S only
within `openspec/` and `docs/pm/`; staged or unstaged tracked changes SHALL fail.
All tracked Go paths, modes and bytes SHALL match S, and additional untracked or
ignored Go files and symlink substitutions SHALL fail. Other untracked/ignored
files SHALL also fail except fixed generated SDD/index/cache data locations
specified by the reviewed design. Such exceptions SHALL be regular,
non-executable, non-source files and SHALL NOT be actual Go/test/embed inputs
reported by successful untagged and `tossos_testseams` Go package enumeration;
enumeration failure SHALL block adoption. E..S Go changes SHALL be
limited to the a063 CLI soak, soak attestation/renewal diagnostic and console
files enumerated in the reviewed design. Every modified-existing E..H function
SHALL still satisfy the ordinary full bundle/hash/revision/branch/call/test
checks, with a recomputed inventory matching E..S. This exception SHALL NOT
waive operational evidence, other change tasks, full tests, or the final gate.

#### Scenario: 검토 이후 소스 추가 또는 수정
- **WHEN** H의 소스가 S와 다르거나 추적되지 않은 또는 ignored Go 파일이 존재하면
- **THEN** 이관 검사는 실패하고 검토 스냅샷의 테스트를 현재 소스의 성공으로 사용하지 않는다

#### Scenario: 실행 기준 이후 삭제된 기존 함수
- **WHEN** E에 있던 Go 함수가 S에서 삭제되고 해당 base-revision 번들이 없으면
- **THEN** 일반 함수 분석 검사가 실패한다

#### Scenario: 소스 확장자가 아닌 빌드 입력을 숨김
- **WHEN** 추적되지 않은 C/assembly/header 파일이나 embedded asset이 생기거나 허용 메타데이터 파일이 실제 Go 테스트 입력으로 사용되면
- **THEN** 확장자나 ignore 규칙으로 면제하지 않고 이관 검사를 실패시킨다

#### Scenario: 정상 SDD 인덱스가 존재함
- **WHEN** 고정된 허용 위치에 실행 불가능한 일반 인덱스 데이터만 있고 실제 빌드 입력과 겹치지 않으면
- **THEN** 그 생성 데이터 자체는 소스 오염으로 간주하지 않으며 다른 이관 조건은 모두 계속 검사한다

#### Scenario: 이관 증거만 뒤에 기록
- **WHEN** 소스 S 이후 원장과 검토 문서를 커밋한 깨끗한 detached H에서 모든 이관 조건과 함수 번들이 검증되면
- **THEN** E를 비교 기준으로 사용할 수 있지만 a063의 남은 운영 및 최종 수용 조건은 계속 검사한다

### Requirement: 이관 소스 밖의 명시적 SDD 인터프리터
The SDD doctor SHALL support an optional `SDD_PYTHON` absolute external
interpreter while preserving its existing local-venv behavior when the variable
is absent. An explicit interpreter SHALL be outside the repository both by its
path and resolved target, executable, and able to probe the repository-pinned
TypeDB driver version. Invalid explicit values or failed/mismatched dependency
probes SHALL fail without falling back to a local environment. An external
interpreter symlink to another external executable MAY be used. This option
SHALL NOT relax adoption source validation, choose an execution baseline or
create a repository-local environment. The doctor SHALL report the selected
interpreter and dependency result.

#### Scenario: 이관 작업 공간에서 외부 도구 환경 사용
- **WHEN** 소스 스냅샷 밖의 유효한 SDD_PYTHON으로 고정된 드라이버를 검사하면
- **THEN** 저장소 안에 가상환경을 만들지 않고 의존성을 확인하며 이관 소스 검사를 그대로 적용한다

#### Scenario: 잘못된 외부 인터프리터 지정
- **WHEN** SDD_PYTHON이 비어 있거나 상대 경로, 저장소 내부 경로, 실행 불가능한 파일 또는 잘못된 드라이버 환경을 가리키면
- **THEN** doctor는 실패하고 기존 로컬 가상환경으로 조용히 대체하지 않는다

#### Scenario: 기존 일반 작업 공간
- **WHEN** SDD_PYTHON을 지정하지 않으면
- **THEN** 기존 저장소 로컬 가상환경 검사와 일반 설정 흐름을 유지한다

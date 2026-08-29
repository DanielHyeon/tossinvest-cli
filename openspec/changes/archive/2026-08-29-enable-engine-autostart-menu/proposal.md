## Why

콘솔은 이미 LIVE 주문 정책, 자동화 게이트, 엔진 시작·정지를 제공하지만 부팅 후
엔진을 다시 시작할지는 설정으로 승인할 수 없다. 현재 준비된 autostart 스크립트도
설치 여부가 별도 수동 절차라서, 운영자가 메뉴에서 선택한 상태와 실제 부팅 동작이
일치하지 않는다.

## What Changes

- `engine.autostart`라는 별도 boolean 설정을 추가하고 기본값을 OFF로 둔다.
- `/settings` 운영 메뉴에 **엔진 자동 시작** 토글을 추가한다.
- 운영자가 토글을 ON으로 저장하면 기존 `StartEngine` seam으로 즉시 기동을
  시도하고, 실패하면 엔진 자신의 인터록 거부 사유를 화면에 표시한다.
- 토글을 OFF로 저장해도 실행 중 엔진을 암묵적으로 정지하지 않는다. 현재 실행
  상태는 기존 [엔진 정지] 버튼이 소유하고, OFF는 다음 프로세스/부팅의 자동
  기동만 막는다.
- 콘솔 프로세스가 시작될 때 autostart가 ON이면 기존 엔진 시작 경로를 한 번
  호출한다. 게이트·Guardian·attestation·거래 정책·단일 writer 인터록은 그대로
  최종 판단하며 어느 것도 우회하지 않는다.
- autostart 변경은 이전 값·새 값·시각·주체와 함께 audit에 기록한다.
- Docker 배포는 기존 TossOS 데이터 디렉터리를 정확히 마운트하고, 콘솔
  컨테이너의 `restart: unless-stopped`가 부팅 후 같은 autostart 판단을 다시
  수행하게 한다.

## Capabilities

### New Capabilities

없음.

### Modified Capabilities

- `operator-console`: 인증·CSRF 뒤에서 autostart 한 키를 편집하고 ON 저장 시 기존
  엔진 기동 seam을 호출하며 결과를 표시한다.
- `engine-safety`: autostart는 기본 OFF인 사람 승인 설정이고, ON이어도 기존
  startup interlock과 journal flock을 통과한 엔진만 기동한다.

## Impact

- Affected code: `internal/config`, `internal/console`, `cmd/tossctl`,
  `schemas/config.schema.json`, `compose.yaml`, 배포 문서와 테스트.
- Config API: `engine.autostart` boolean이 추가된다. 누락은 false로 해석되므로
  기존 설정과 하위 호환된다.
- Runtime: 콘솔 시작 시 조건부로 엔진 자식 프로세스가 기동될 수 있다. 기본값은
  OFF이고, 실제 주문 능력은 기존 자동화 게이트와 기동 인터록이 계속 통제한다.
- Deployment: Docker daemon과 콘솔 컨테이너의 기존 재시작 정책을 이용하며 새
  특권·호스트 네트워크·Docker socket은 추가하지 않는다.

PM bootstrap exception: 현재 TossOS portfolio에는 완료된 SDD 도입 story만
있으므로 `enable-engine-autostart-menu`는 product story hierarchy가 생길 때까지
`docs/pm/portfolio/_registry.yaml`에 임시 등록한다.

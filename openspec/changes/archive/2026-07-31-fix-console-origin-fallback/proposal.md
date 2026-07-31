## Why

설정·재시작 등 모든 콘솔 POST가 canonical HTTPS 주소에서 시작되어도 공통 origin
게이트에서 거부된다. 원격 응답이 `Referer`를 금지하는 동시에 일부 브라우저가
same-origin 폼 POST의 `Origin`을 생략할 수 있는데, 현재 구현은 이 합법적인 요청을
cross-origin 요청과 동일하게 취급한다.

## What Changes

- 명시적 `Origin`이 있으면 기존처럼 canonical HTTPS origin과 정확히 비교한다.
- `Origin`이 없고 유효한 `Referer`가 있으면 경로를 버리고 scheme·host·port만 비교한다.
- 상태 변경 gate에서 두 헤더 key가 모두 없으면 실제 TLS 연결과 요청 `Host`로 계산한
  origin이 canonical HTTPS origin과 정확히 일치하는 경우에만 origin 증거를 충족한
  것으로 인정한다.
- compatibility `/login` POST는 headerless fallback을 사용하지 않고 기존의 명시적
  Origin/Referer 계약을 유지한다.
- 기존 POST·CSRF·peer CIDR·exact Host·감사·운영 interlock은 그대로 유지한다.
- scheme·host·port가 다른 명시적 origin은 request-host fallback으로 덮지 않고 계속
  거부한다.

## Capabilities

### New Capabilities

- `console-request-origin`: 콘솔 상태 변경 요청의 canonical origin 증거 우선순위와
  privacy-header 환경의 TLS/Host fallback 계약.

### Modified Capabilities

- 없음.

## Impact

- `internal/console/remote.go`: strict header origin 판정과 mutation-only fallback 판정.
- `internal/console/console.go`: 상태 변경 gate가 mutation-only 판정을 호출한다.
- `internal/console/remote_test.go`: origin/referer 부재, path 무관성, explicit mismatch
  회귀 테스트.
- 모든 console POST가 이 공통 gate를 사용하지만 handler별 저장 로직, LIVE 주문,
  엔진 interlock, config schema와 배포 환경값은 변경하지 않는다.

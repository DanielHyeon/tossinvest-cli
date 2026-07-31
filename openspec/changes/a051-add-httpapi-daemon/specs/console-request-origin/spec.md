## ADDED Requirements

### Requirement: write origin identity는 scheme host port다
브라우저·API write guard는 configured HTTPS origin과 요청 시작점을 scheme, canonical host와 effective port로 비교해야 하며 path, query 또는 fragment를 비교해서는 안 된다 (MUST NOT).

#### Scenario: 다른 하위 경로
- **WHEN** configured origin이 `https://127.0.0.1:37085`이고 `/positions`에서 `/api/v1/settings`로 쓴다
- **THEN** scheme·host·port가 같으므로 origin 검사를 통과한다

#### Scenario: 다른 port
- **WHEN** 요청 origin이 `https://127.0.0.1:37086`이다
- **THEN** path가 같아도 origin 불일치로 거부한다

#### Scenario: forwarded HTTPS
- **WHEN** 신뢰된 proxy가 canonical HTTPS host/port를 전달한다
- **THEN** allowlisted proxy에서 온 헤더만 사용해 origin을 구성한다

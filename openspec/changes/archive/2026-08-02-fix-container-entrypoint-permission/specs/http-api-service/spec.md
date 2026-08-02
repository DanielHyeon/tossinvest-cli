## ADDED Requirements

### Requirement: 컨테이너 entrypoint mode는 checkout filesystem과 독립적이다
배포 image는 source checkout이나 Git executable bit에 의존하지 않고 entrypoint를 실행 가능한 mode로 설치해야 한다 (SHALL). non-root runtime identity는 Compose 재생성 후 entrypoint를 실행할 수 있어야 한다 (SHALL).

#### Scenario: NTFS checkout에서 image 재빌드
- **WHEN** entrypoint source가 `0644`인 checkout에서 image를 재빌드한다
- **THEN** image의 `/usr/local/bin/tossos-entrypoint`는 `0755`이고 service는 exit 126 없이 healthcheck까지 기동한다

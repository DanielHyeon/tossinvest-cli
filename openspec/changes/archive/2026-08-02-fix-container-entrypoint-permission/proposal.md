# 컨테이너 entrypoint 실행 권한 고정

## Why

a052 배포 재빌드에서 NTFS checkout의 `100644` mode가 image에 그대로 복사되어 두
서비스가 exit 126으로 재시작했다. 배포 결과가 host filesystem의 mode 보존 방식에
의존하면 안 된다.

## What Changes

- Dockerfile이 entrypoint를 명시적 `0755` mode로 복사한다.
- 정적 회귀 테스트가 executable copy 계약을 고정한다.
- 운영 문서가 NTFS를 포함한 재빌드 불변식을 설명한다.

## Impact

- image packaging과 운영 문서만 변경한다.
- 주문, LIVE, engine lifecycle 설정, journal 및 reconcile 상태는 변경하지 않는다.

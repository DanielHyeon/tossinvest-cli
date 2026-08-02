## Context

Git은 현재 저장소에서 `deploy/container-entrypoint.sh`를 `100644`로 추적한다. Docker
`COPY`는 source mode를 보존하므로 tini가 entrypoint를 실행하지 못하고 exit 126을
반환했다.

## Decision

Dockerfile의 copy instruction에 `--chmod=0755`를 사용한다. source checkout의
filesystem과 Git mode에 의존하지 않고 image layer가 실행 권한을 소유하게 한다.

## Verification

- 정적 Go 테스트가 exact executable copy instruction을 검증한다.
- image를 no-cache로 재빌드하고 non-root user로 entrypoint가 실행되는지 확인한다.
- Compose health와 HTTPS HTTP/2 canary를 재검증한다.

# HTTP/2 bodyless read 오판 수정

## Why

배포 카나리에서 body 없는 HTTP/2 GET/HEAD가 Go의 HTTP/2 전용 EOF body 객체 때문에
`BODY_NOT_SUPPORTED` 400으로 오판됐다. HTTP/1.1 healthcheck만으로는 이 실제 모바일
클라이언트 경로를 검증하지 못했다.

## What Changes

- inbound request의 실제/unknown body 존재를 `ContentLength != 0`으로 판정한다.
- body 없는 HTTP/2 GET/HEAD와 SSE를 허용한다.
- known-length 및 unknown-length body는 계속 fail-closed 400으로 거부한다.
- 실제 TLS HTTP/2 end-to-end 회귀 테스트를 추가한다.

## Impact

- `internal/httpapi`의 body 제한 middleware와 read/stream router만 변경한다.
- mutation, capability, 주문, LIVE, engine lifecycle 및 운영 toggle에는 영향이 없다.

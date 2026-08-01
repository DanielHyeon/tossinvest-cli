## Context

Go HTTP/1.1과 `httptest.NewRequest`는 body 없는 요청에 `http.NoBody`를 사용하지만,
HTTP/2 server는 END_STREAM 요청에도 EOF를 반환하는 전용 body 객체를 넣을 수 있다.
따라서 body 객체 identity는 wire body 존재 여부의 안정된 계약이 아니다.

## Decision

server-side request의 `ContentLength`를 사용한다. `0`은 known empty로 허용하고,
`-1` unknown/open body와 양수 declared body는 거부한다. 다만 HTTP/2 parser가 field를
0으로 정규화하면서 양수 또는 잘못된 `Content-Length` header를 보존한 경우도 fail-closed로
거부한다. body를 한 바이트 읽어 확인하지 않는다. 그런 probe는 body를 끝내지 않는
client에서 read timeout까지 handler를 막는다.

## Verification

HTTP/2가 실제 협상된 TLS test server에서 bodyless GET/HEAD/read stream을 허용하고,
known-length 및 unknown-length GET body를 stable `BODY_NOT_SUPPORTED`로 거부한다.
HTTP/1.1 계약과 256 KiB mutation ceiling도 기존 테스트로 보존한다.

## Safety

수정은 read/stream framing 판정뿐이다. signed capability mutation route, remote LIVE/gate/
kill-switch route 부재, broker 주문 경로와 autostart 상태는 바꾸지 않는다.

# Go transport template 수정

lot2 a112 통합에서 `NewTransport`의 `base.Clone()`이 caller 소유 transport의 HTTP/2 lazy 초기화를 수행해 원본 `TLSNextProto`/`TLSClientConfig.NextProtos`를 바꾸는 RED 확인. Go 1.25.0 `src/net/http/transport.go:329-372`에서 `Clone`은 원본 `nextProtoOnce.Do(onceSetNextProtoDefaults)`를 먼저 호출한다. 원본 호출 없이 공식 constructor 소유 base의 지원 설정을 새 HTTP/1 template에 명시적으로 옮긴다.

보존: `Proxy` 선택 함수, `DialContext`, TLS config의 검증 설정/RootCAs, TLSHandshakeTimeout, ResponseHeaderTimeout, ExpectContinueTimeout, IdleConnTimeout, MaxIdleConns/PerHost, MaxConnsPerHost, MaxResponseHeaderBytes, WriteBufferSize, ReadBufferSize, DisableCompression. HTTP client 자체 Timeout은 이 API 외부 client가 소유하며 여기서 바꾸지 않는다. enforcing transport에서는 DisableKeepAlives=true, ForceAttemptHTTP2=false, TLSNextProto 빈 map, ALPN http/1.1을 강제한다. 원본 설정은 수정하지 않는다.

임의 TLSNextProto, proxy CONNECT hooks/headers, HTTP2/Protocols 설정, legacy/custom TLS dial, TLS verification bypass 및 TLS callback은 거절한다. stdlib Clone이 만들어내는 정확한 ALPN `[h2, http/1.1]`만 빈 ALPN과 함께 입력으로 허용하고, 실제 HTTP 전송 전용 복사본에서는 이를 제거한 뒤 http/1.1을 설정한다. 임의 다른 ALPN은 거절한다. 이 검증은 전송 래퍼의 기능이며 caller transport의 거래 권한을 부여하지 않는다.

`template-red.log` 원본 변형 RED; `template-green.log`와 `template-callback-green.log`는 원본 불변·표준 clone 재포장·실제 검증 TLS HTTP/1·필드 mapping·지원 밖 설정 및 callback 거절 집중 race GREEN. `template-callback-red.log`는 미처 거절하지 않은 callback의 이전 실패. `template-clone-mutation.log`는 원본 Clone 복원 변이 RED, `template-hooks-mutation.log`는 임의 TLSNextProto 거절 제거 변이 RED. `template-final-go125-{normal,race}.log`, `template-final-go126-{normal,race}.log`는 최종 소스 대상 패키지 집중 검증이다. lot2 실제 official/a112 fixture 검증과 Manager 독립 검증은 별도다. 라이브·운영·전체 테스트/게이트 실행 없음.

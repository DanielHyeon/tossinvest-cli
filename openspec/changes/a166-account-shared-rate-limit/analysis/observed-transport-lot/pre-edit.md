# 사전 분석 — 미배선 관측 전송

기준 5f681f2a5. source AST: pre-roundtrip.json, pre-constructor.json, pre-open.json. 기존 store, clock, context/errors/http를 재사용하며 신규 의존성 없음. CodeGraph는 이 worktree에서 미초기화. impeccable N/A — backend-only.

`OpenContext`: 취소·설정 검증 → root/identity/lock 열기 → queue 유효성 및 release → 반환. 본문은 유지하고 신규 helper가 열린 Budget의 event.lock 검증 후 공개 전 private field를 한 번 설정; 실패 시 Close 오류를 보존한다. production root/queue/wire/event pin 동시 검증은 후속 v2 factory 범위다.

`NewTransport`: nil/지원불가 TLS·proxy hook 검증 → direct HTTP/1 template 복사 → Budget pointer 보관. 본문 및 호출자 원본 불변 유지; 관측 설정은 불변 Budget에서 읽는다.

`Transport.RoundTrip` FLM: priority/URL/proxy 거절(body close) → Acquire 거절(body close) → gate/TLS 설정 → 내부 RoundTrip → 무응답 gate.stop/drain 후 sent 판정·release → 응답 TLS 복구·body owner 이양·cancel cleanup. 변경은 Acquire와 TLS 사이 Begin/재검사, 양 반환 경로 Finish 추가로 한정. Begin 실패에서는 seq 없이 Finish 금지. pretransport body close·Finish·release 오류를 join하고 ErrNotSent. response가 있으면 상태/body를 보존하고 별도 collector/callback에 기록 실패를 보존. 기존 responseOwner/gate 함수 본문 유지.

BTM: unobserved 회귀; Begin 오류 및 취소 wire0/body close; poll non-GET 거절; Begin 이후 orphan poison 재검사; 응답 429·무효 hint·body 오류; Finish 실패 시 response 보존; 무응답 drain 뒤 sent 분류; 정상 응답 lease가 EOF/Close/cancel까지 지속; cleanup 오류 각각. 네트워크 시험은 loopback-only namespace와 endpoint pin한 local TLS만. mutation은 Begin, poll 재검사, Finish 및 응답 보존 경계. 전체 tests/gate·운영은 금지.

Begin은 TLS handshake 전 기록이다. pending은 HTTP 첫 Write 증거가 아니고 crash 시 보수적 false gap과 가용성 비용이 생긴다. 이 단계는 production factory/token/Retrier/전체 호출자 연결 및 deployment trust 입증이 아니다.

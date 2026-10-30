# Transport.RoundTrip FLM
기준 5f681f2a5: ast.json. priority/URL/proxy 검증 실패는 body close·ErrNotSent; Acquire 실패도 body close. 성공 시 fresh HTTP1 TLS gate; 응답 없음은 gate stop/drain 뒤 sent 판정·lease release; 응답은 TLS 복원과 body EOF/Close/cancel까지 lease 소유.
변경: Acquire 직후 observed Budget만 Begin; poll은 Begin 뒤 재검사; 전송 전 실패는 body close·성공 seq만 NoResponse Finish·lease release. 내부 전송 이후 무응답은 drain 뒤 Finish·release; 응답은 collector 429 기록과 Finish 후 본래 response/body 이양. unobserved는 이전 경로 그대로. Finish settle context는 요청 취소를 분리해 event-lock 2초만 제한. Error cleanup joins, sent 판정은 변경하지 않음.

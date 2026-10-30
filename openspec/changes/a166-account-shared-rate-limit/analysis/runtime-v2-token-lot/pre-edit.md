# 사전 FLM·BTM — runtime v2 / token
기준 e5c9d0256c4cbc8dafacc483c3fdf13bc928f3f4. pre-{open,manifest,duplicates,token}.json은 현재 Go AST. 검색: 기존 runtime manifest parser, observed Open, context scope, token 및 stdlib JSON/clock 재사용. 신규 의존성 없음. CodeGraph 이 worktree 미초기화. impeccable N/A backend-only.

- Open: nil/cancel·OFF·required 설정 검사 → 신뢰 manifest 읽기·root 외부 확인 → Budget OpenContext → root/queue/wire VerifyIdentity → 반환. v2 변경은 OpenObserved로 event pin 결속, 나머지 순서 및 오류/Close 보존. OpenObserved 실패는 내부가 Close하고 VerifyIdentity 실패만 caller Close.
- readManifest: 절대경로·부모 NOFOLLOW·0600 owner/link·bounded read·전후 stat → duplicate parser → DisallowUnknownFields·trailing 거부 → root ancestry 검사 → version/필수 검사. 변경은 v2 및 event_lock device/inode 필수. null·누락·0/shape/duplicate/unknown 거절.
- rejectDuplicates: token 재귀·객체별 allowed/seen·배열 탐색·종결 검사. event_lock 및 nested device/inode만 목록 추가. 의도하지 않은 필드/배열은 디코드와 필수값 검증에서 거절.
- tokenManager.exchange: POST form 생성 → NewRequest(ctx) → hc.Do → 응답 body 읽기/status 분류 → cache. enforcing Transport에서만 WithTokenObservation으로 kind token/key zero, 기존 collector/context/priority 전달. OFF 기존 방식 그대로.

BTM: v1/누락/0/copy/event dirty/root/queue/wire mismatch는 Open 거절·wire0; nested unknown/duplicate/type/trailing 거절; 초기화 실패 정리·OFF 정상; 복수 client·token/GET/401 각각 기록, cached token 교환0, poll scope에서 token 429 collector 유지·hint0, priority/deadline/cancel 유지. loopback-only/no-route namespace와 exact local TLS dial만; 기존 related fixture는 v2 provision·pinned event로만 갱신. source census와 미실행 통합 범위 구분. 전체 tests/gate·운영·브로커는 금지.

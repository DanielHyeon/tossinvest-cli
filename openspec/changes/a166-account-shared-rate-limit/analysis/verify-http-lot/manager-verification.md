# Manager 검증 — verify와 HTTP 조회 수명

후보 `2f4132510731c3eaf6d07df98c65b86cbf632b32`의 제한 범위 병합·푸시 승인. a166 전체 완료와 활성화 승인은 아니다.

Manager는 생산 diff와 `ee4d13de4` 수정을 손으로 읽었다. verify의 초기/계좌 선택 client가 같은 root factory를 거친다. Console의 Calendar/Holdings/Orders/Stocks 조회는 request와 서버 lifetime을 합친 context를 실제 read 종료까지 유지한다. 초기 client 생성 때만 lifetime을 적용하던 P1을 수정했다. HTTP 서버 오류와 Shutdown 오류는 보존하며 timeout 뒤 Close를 handler 완료로 간주하지 않는다. Console 재실행은 기존 별도 worker/handler 완료 조건을 유지한다.

독립 Terra가 고정 후보에서 Go 1.25 선택 normal 및 tagged race를 실행했다. cmd/tossctl 1.637초/2.778초, internal/console 0.147초/1.189초 PASS. Manager는 `/tmp/a166-verify-http-independent-review.md`와 normal/tag-race 원문을 확인했다. lifetime RED, read-entry 변이, request 취소 및 기존 종료 순서 fixture의 증거를 대조했다. 현재 AST와 원문 gzip SHA는 독립 검토자가 확인했으며 Manager의 제어 흐름 검토는 손읽기다.

네트워크 시험은 loopback-only/no-route namespace 및 로컬 TLS 경계다. 실제 runHTTPAPI 서버 선오류 전체 통합과 실제 exec, 브로커, 운영, 전체 tests/gate는 미실행. Shutdown timeout 뒤 handler quiescence는 증명하지 않았다. gstack은 범위 수동 체크리스트이며 자동 전체 pipeline 실행으로 확대하지 않는다. impeccable: N/A — backend context/lifecycle-only; 화면 변경 없음. stdlib context/http/errors 재사용, 신규 의존성 없음. main의 a128 변경을 보존하고 PM/OpenSpec/diff 검사 후 게시한다. 미완료 change는 archive하지 않는다.

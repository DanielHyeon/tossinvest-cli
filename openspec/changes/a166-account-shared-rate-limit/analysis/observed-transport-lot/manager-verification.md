# Manager 검증 — 미배선 관측 전송 계층

후보 `aee3b8df6f5958825a63d61b2d116b332f08c32d`의 제한 범위 병합·푸시 승인. runtime v2, token 분류, Retrier 또는 전체 운영 429 기록의 완료 판정은 아니다.

Manager는 생산 diff를 손으로 읽었다. observed-open은 store를 결속한 Budget만 반환한다. 실제 wire 이전 Begin 뒤 PollRemaining을 재확인해 복구된 orphan poison을 반영한다. 전송 전 오류는 body를 닫고 성공한 Begin만 별도 settlement context로 마감한 뒤 lease를 해제한다. 응답 429는 body를 읽기 전에 collector와 store에 기록하고, 기록 오류가 HTTP 응답/status/body를 비수용으로 바꾸지 않는다. 응답 없는 경로의 stop/drain과 기존 sent/NOT_SENT 구분을 유지한다. settlement 2초는 context 대기 상한이며 filesystem/callback 완료의 유한 시간 보장이 아니다.

독립 Terra는 고정 후보의 새 13개 및 기존 9개 선택 시험을 Go 1.25 normal 9.409초, tagged race 10.461초로 통과했다. Manager가 `/tmp/a166-observed-independent-go125-{normal,tag-race}.log` 원문을 확인했다. 사전 AST 소스 해시, 12개 원문 gzip SHA 및 5개 변이 RED는 독립 검토자가 확인했다. Manager의 소스 판단은 손읽기다. fixture는 loopback-only/no-route namespace와 로컬 TLS다.

운영 runtime/official/token/Retrier 배선은 변경하지 않았다. 실제 브로커·전원 손실·운영·전체 tests/gate는 미실행. gstack은 범위 수동 체크리스트다. impeccable: N/A — backend transport-only. 기존 stdlib와 저장소 store/lease를 확장했으며 새 의존성 없음. 원문 증거를 보존하고 최신 main과 순차 통합한다. a166 미완료 항목 때문에 archive하지 않는다.

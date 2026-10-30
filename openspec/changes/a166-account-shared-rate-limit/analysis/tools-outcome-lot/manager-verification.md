# Manager 검증 — outcome backfill와 d2 읽기 도구

후보 `48e736aeec24b2a0b126edf12592b628ff06fbf2`의 제한 기능 범위 병합·푸시 승인. 전체 a166 완료 또는 실제 도구 실행 승인이 아니다.

Manager는 생산 diff를 손으로 읽었다. a130 factory는 candle method closure를 반환한 뒤에도 run 끝까지 살아 있으며, 실패 시에도 close 오류를 결과에 합친다. lots가 없으면 factory를 열지 않지만 lots가 있는 preview는 기존처럼 조회한다. d2는 dry-run 뒤에 factory를 만들고 기존 5분 context·조회 반복·exit 판정을 유지한다. 두 도구의 read context는 Auxiliary다.

독립 Terra의 고정 후보 Go 1.25 normal은 a130 0.971초/d2 0.419초, tagged race는 8.463초/1.439초 PASS. Manager는 `/tmp/a166-tools-outcome-independent-review.md`와 normal/tag-race 원문을 확인했다. 임시 입력·fake credentials·loopback-only/no-route namespace·명시 로컬 TLS/DialContext를 사용했다. required 결측 거부, critical lease 중 wire0, token/candle closure, dry-run, 선택 Close 오류 및 두 auxiliary 변이를 검증했다. d2 원본 RED는 DNS 시도를 namespace가 차단했지만 고정 TLS 조건에 맞지 않아 유효 RED에서 제외했다. 대체 pinned RED를 보존한다.

절차 미준수: 상세 FLM/BTM을 구현 전에 완성하지 않았다. 사전 AST와 사후 분기 분석·독립 검증은 있으나 이를 사전 요건 준수로 소급하지 않는다. 이미 생긴 절차 위반을 지우거나 재작성하지 않고 기록한 상태로 현재 제한 결과를 수용한다. 다음 로트에서는 구현 전에 FLM/BTM을 고정한다.

d2 Close 실패 직접 주입과 lots 있는 a130 preview 전체 통합은 미검증이다. 실제 도구·실원장·브로커·운영·전체 tests/gate는 미실행. gstack은 범위 수동 검토이며 전체 자동 pipeline 인증이 아니다. impeccable: N/A — backend CLI wiring-only. 기존 factory/stdlib를 재사용하고 신규 의존성 없음. 타 작성자와 main a128 자료를 보존하고 제한 PM/OpenSpec/diff 검사 후 게시한다. 미완료 change archive는 하지 않는다.

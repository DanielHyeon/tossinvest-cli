# Manager 검증 — runtime v2와 token 관측

후보 `21b9bc85bdf1fbd7947b072a99f53554923b157c` 및 nil context 수정 `288358373b6453772ee0e4c4b8eb874a0a9f1c35`의 제한 병합·푸시 승인. Retrier poll 소비와 전체 a166 완료·운영 전환은 후속이다.

Manager는 생산 diff를 손으로 읽었다. required manifest는 v2와 event_lock pin을 요구하고 기존 root/queue/wire identity를 그대로 확인한다. OpenObserved 성공과 VerifyIdentity 뒤에만 factory를 공개한다. 자동 provisioning이나 v1 fallback이 없다. token scope helper는 private collector를 유지하며 Token/zero key만 덮고 부모 priority/deadline/cancel을 계승한다. OFF cache/HTTP body/오류 동작은 유지한다. 독립 검토에서 nil context 오류→panic 회귀를 찾아 helper 적용 조건에 ctx != nil을 추가했다. nil은 기존 NewRequestWithContext 오류 경로로 간다.

독립 Terra는 Go1.25의 승인된 정확 선택 목록을 11개 패키지에서 normal/tagged race로 검증했고 최신 tools 7개 시험을 별도로 통과했다. token429·401/cache·Finish fault, 실제 local TLS의 a112 clone/replay/probe, 두 process event 및 시작 fixture를 포함한다. delta에서는 Accounts(nil)/exchange OFF·required 비교와 token 회귀 세 시험의 normal/tagged race를 다시 통과했다. Manager는 `/tmp/a166-runtime-v2-token-independent-review.md`, independent normal/tag-race 및 nil/delta 원문을 확인했다. 사전 AST 4개 source SHA, 원문 gzip 14개 SHA와 다섯 변이 RED는 독립 검토자가 대조했다.

모든 네트워크 fixture는 loopback-only/no-route와 local TLS/exact DialContext다. a112의 test-only unsafe는 기존 constructor-owned transport의 RootCAs와 local DialContext만 바꾼다. base/flag/pointer 봉인을 바꾸지 않지만 이 fixture는 실제 운영 authority 증명이 아니다. 초기화 실패의 root fd 정리는 선택 fixture 근거이며 실제 powerloss 검증은 아니다.

최신 main candidate/soak 통합 후 같은 v2 root fixture를 쓰는 정확 시험을 추가 확인하고 source 차이를 기록한다. 현재 정적 송신 census는 `/tmp/a166-final-wire-census.md`이며 통합 SHA로 갱신한다. 저수준 OFF/legacy API와 모든 운영 참여자의 실제 namespace 검증은 구분한다. 전체 tests/gate·브로커·운영·활성화는 미실행. impeccable: N/A — backend runtime/token-only. 기존 API/stdlib 확장, 신규 의존성 없음. scoped 수동 gstack 및 독립 검증 범위이며 자동 전체 pipeline 인증은 아니다. a128 자료 보존·PM/OpenSpec/diff 확인 후 게시하며 archive하지 않는다.

# 사전 FLM / BTM: cached MB 및 사람 측정 probe

기준 e5c9d0256c4cbc8dafacc483c3fdf13bc928f3f4. 함께 기록한 pre AST의 source hash와 분기 목록이 기준이다. search-first: 기존 accountbudgetruntime.ParseLaunchConfig/Open/Factory.NewClient/Budget/Close, 공식 A112MBUS guard, probe 기존 별도 budget transport 재사용. 신규 의존성 없음. UI 없음(`impeccable: N/A — backend CLI`).

## 기존 경로와 변경

- a112 `newProductionDependencies`: 기존 무분기 생성자는 빈 Credentials·tokenCache로 official.New를 만든다. CLI `main` B1 명시 before 누락은 exit2, B2 run 실패 exit2. main의 os.Exit 전에 명시 run helper가 factory를 소유·닫아야 한다. required에서는 같은 factory client, OFF는 기존 client semantics 유지. reader의 Candle/Orderbook/Calendar는 기존 single GET과 15초 경계, 빈 credentials·캐시 토큰·no OAuth/no Accounts/no retry를 보존한다. 새로운 factory 오류는 HOLD·wire0, Close 오류도 HOLD·exit2다. 모든 GET은 auxiliary다.
- a087 `openOfficial` B1–B3는 경로·credential 오류/부재 거부. preview는 client/factory 접근0. `execute` B1–B8은 terminal·확인·주문조립·receipt 선행 가드다. B9–B15는 required budget 일치·transport 구성·OFF opener·header 및 재검증. B16–B18은 header·요청·pending receipt 거부. B19–B23은 최종 receipt 오류와 accepted/refused/unknown 판정. local guards 이전 factory/credentials 접근을 도입하지 않는다. required factory가 동일 Budget client와 POST transport를 공급하며 ctx priority Normal. preview/confirm, POST 1회, Sent와 receipt 결과는 기존대로다. run helper가 factory 수명을 관리하고 Close 오류를 기존 결과에 합친다.

## 사전 BTM / RED 계획

- a112 required missing manifest: explicit local TLS + exact DialContext, 기존 seam-only baseline에서 wire0 기대 실패 → guard GREEN; OFF cached token GET 성공 및 15초·no OAuth/no Accounts; held critical lease에서 auxiliary 즉시 거부와 wire0; Close 실패·기존 HOLD 확인.
- a087 preview/invalid confirmation은 env·factory 미접근, execute required missing manifest에서 인증 및 POST wire0. 정상 local TLS fixture에서 token/account + POST 각 1회, Normal priority, 기존 receipt/Sent 및 confirmed bytes 유지; held critical lease 중 Normal이 queue에서 대기하며 취소시 POST0; Close 실패 반환. 기존 test-only harness를 사용하며 실제 probe/실거래 실행 금지.
- 두 tool 모두 모든 network fixture는 격리 user/net namespace의 lo-only·route 없음 + local TLS + exact DialContext. 필요한 constructor option seam을 먼저 동작보존 형태로 추가해 pinned RED를 수행한다. auxiliary/normal, required guard 및 duplicate POST 변이의 실패를 확인한다. 선택 normal·tagged race, gstack 관련 수동 체크 후 독립 검토. 전체 tests/gate/운영/브로커 금지.

# Branch Test Map: `Dial`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | descriptor 파싱·검증 실패 | a108 `TestTheReadEndpointRefusesADescriptorItCannotTrust` 계열(a108 소유, 회귀로만 확인) | no | yes(회귀) |
| B2 | socket 모양·권한이 정확-0600이 아니다 | a108 클라이언트 검증 핀(a108 소유) | no | yes(회귀) |
| B3 | 아무도 수락하지 않는 socket | a109 `TestTheDaemonReattachesAfterTheEngineRestarts`가 이 거부를 지나 재부착으로 간다 | yes(§2.3) | yes |
| 종단 | 만들어진 transport가 요청 사이에 유휴 연결을 두지 않는다 | `TestTheDialedTransportKeepsNoIdleConnections` (a109 §2b.3 G5) | yes(§2b.3 — `DisableKeepAlives=false` 관측) | yes |

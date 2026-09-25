# Branch Test Map: `Dial`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | descriptor 실패 | a108 descriptor 핀(`TestDescriptorIdentityClausesRejectASwappedFile`) | no(회귀 핀) | yes |
| B2 | 정확-0600 아님 | a108 발행·거부 핀 | no(회귀 핀) | yes |
| B3 | 아무도 수락하지 않음 | `TestDialRefusesSocketWithNoListener` (a108) | no(회귀 핀) | yes |
| 종단 | 유휴 연결 없는 client | `TestTheDialedTransportKeepsNoIdleConnections` (a109) | no(회귀 핀) | yes |

# Branch Test Map: `StartPositionPolicyRuntimeServer`

구현 후 재최신화(§1.7 · §2b.3 G9). 분기 11개는 **지금 HEAD**의 AST다 — G9가 Mkdir →
회수 → 재Mkdir 세 절을 `openReclaimedControlDirectory`로 옮겨 그 자리가 B4 하나가 됐다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | reader 없이 기동 거부 | 커버 없음 | no | no |
| B2 | 빈 엔진 디렉터리 문자열 | 커버 없음 | no | no |
| B3 | 안전하지 않은 엔진 디렉터리 | 커버 없음(자매 endpoint의 `TestPositionPolicyControlRejectsInsecureControlFilesystem`이 같은 검사를 잰다) | no | no |
| B4 | 기존 control 디렉터리 위 기동 — ① pre-chmod 0700 잔재 회수 ② 산 주인 거부 ③ staging 잔재 회수 ④ 낯선 엔트리 거부 ⑤ 회수 후 재-Mkdir | `TestPositionPolicyRuntimeServerStartsOverALeftover`(a108 관용 핀) · `TestTheSiblingEndpointsStartOverAPreChmodSocket` · `TestTheSiblingEndpointsRefuseToTakeALiveOwnersSocket` · `TestTheSiblingEndpointsReclaimTheirStagingLeftovers` · `TestTheSiblingEndpointsRefuseAForeignEntry` (뮤테이션 M1b·M3·M5·M9·M17이 이 절을 죽인다) | yes(§1.1·1.2·1.3) | yes(§1.4 · §2b.3) |
| B5 | 0700 아닌 control 디렉터리 거부 | 커버 없음(회수 쪽 동형은 `TestReclaimRefusesAnUnsafeControlDirectory`) | no | no |
| B6 | staged bind → 0600 → rename 발행 | `TestRuntimeOnlyUnixControlCrossNamespaceTopologyAndRouteAbsence`(0600 socket 실측) · 뮤테이션 M21이 23건을 죽인다 | no | yes(§1.4) |
| B7 | 발행 후 정확-0600 확인 | 같음 + `TestTheClientSocketChecksStayExactlyZeroSixHundred`(완화 누출 감시) | no | no |
| B8 | 토큰 생성 실패 | 커버 없음 — `crypto/rand.Read`는 Go 1.24 이후 오류 대신 프로세스를 죽이므로 이 분기는 도달 불가다 | no | no |
| B9 | GET 아닌 메서드는 405 | `TestRuntimeOnlyUnixControlCrossNamespaceTopologyAndRouteAbsence`(POST /v1/runtime = 405) | no | no |
| B10 | reader 오류의 RPC 매핑 | 커버 없음 | no | no |
| B11 | descriptor 발행 실패 정리 | 커버 없음 | no | no |

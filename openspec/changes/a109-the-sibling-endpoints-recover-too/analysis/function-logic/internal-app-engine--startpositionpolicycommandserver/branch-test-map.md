# Branch Test Map: `StartPositionPolicyCommandServer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | commands 없이 기동 거부 | 커버 없음 | no | no |
| B2 | 빈 엔진 디렉터리 문자열 | 커버 없음 | no | no |
| B3 | group-writable 엔진 디렉터리 거부 | `TestPositionPolicyControlRejectsInsecureControlFilesystem` | no | no |
| B4 | 기존 control 디렉터리 위 기동 | `TestPositionPolicyCommandServerStartsOverALeftover` | no | no |
| B5 | 첫 기동 | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint` | no | no |
| B6 | 디렉터리 생성 불가 | 커버 없음 | no | no |
| B7 | 0750 control 디렉터리 거부 | `TestPositionPolicyControlRejectsInsecureControlFilesystem` ("wrong existing control directory mode") | no | no |
| B8 | 검증 실패 시 우리가 만든 디렉터리만 제거 | 커버 없음 | no | no |
| B9 | 그 밖의 실패 정리에서도 같음 | 커버 없음 | no | no |
| B10 | TCP bind 실패 | 커버 없음 | no | no |
| B11 | 토큰 생성 실패 | 커버 없음 | no | no |
| B12 | `/v1/health` 메서드 제한 | 커버 없음(인증은 `TestPositionPolicyControlRejectsMissingBearerBeforeCommand`가 잰다) | no | no |
| B13 | `/v1/positions` 메서드 제한 | 커버 없음 | no | no |
| B14 | List 오류의 RPC 매핑 | 커버 없음 | no | no |
| B15 | 격리 해제 능력 유무로 라우트가 갈린다 | `TestAServiceWithoutQuarantineCapabilityRegistersNoNewRoutes` · `TestTheQuarantineRoutesCarryAReleaseEndToEnd` | no | no |
| B16 | descriptor 발행 실패 정리 | 커버 없음 | no | no |
| (신규, B16 이후 위치) | staging 잔재가 회수되고 낯선 엔트리는 그대로 남는다 | a109 §1.3 RED가 현재 "잔재가 남는다"를 고정한다 | yes(§1.3) | yes(§1.4) |

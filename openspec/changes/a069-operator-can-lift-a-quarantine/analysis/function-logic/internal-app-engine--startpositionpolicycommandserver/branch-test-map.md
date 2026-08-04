# Branch Test Map: `StartPositionPolicyCommandServer`

번호는 구현 **후** 재생성한 AST 기준이다. a063이 추가한 분기는 **B16 하나**다 —
`commands`가 격리 capability를 함께 구현하는가. 새 라우트의 method·인증 검사는
`exit_quarantine_transport.go`의 별도 함수에 있으므로 이 함수의 분기 수를 늘리지
않는다. B1~B15는 조건과 순서가 그대로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 서비스 없이 시작하면 거부한다 | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint` (음성 경로) | no | pass |
| B2 | engine 디렉터리 없이 시작하면 거부한다 | 같음 | no | pass |
| B3 | engine 디렉터리 검증 실패는 거부한다 | `TestPositionPolicyControlRejectsInsecureControlFilesystem` | no | pass |
| B4 | control 디렉터리 생성 오류를 분류한다 | 같음 | no | pass |
| B5 | Mkdir 성공을 기억한다 | 같음 | no | pass |
| B6 | ErrExist가 아닌 생성 오류는 거부한다 | 같음 | no | pass |
| B7 | control 디렉터리 검증 실패는 거부한다 | 같음 | no | pass |
| B8 | 검증 실패 시 방금 만든 디렉터리를 되돌린다 | 같음 | no | pass |
| B9 | 이후 실패 경로도 같은 정리를 수행한다 | 같음 | no | pass |
| B10 | listener 바인드 실패는 정리 후 거부한다 | 같음 | no | pass |
| B11 | token 생성 실패는 listener를 닫고 거부한다 | 같음 | no | pass |
| B12 | `/v1/health`는 GET만 받는다 | `TestPositionPolicyCommandEndpointDoesNotExposeRuntime` | no | pass |
| B13 | `/v1/positions`는 GET만 받는다 | 같음 | no | pass |
| B14 | 목록 오류는 RPC 오류로 매핑된다 | `TestPositionPolicyControlStrictlyBoundsAndFramesJSON` | no | pass |
| B15 | descriptor 기록 실패는 전부 롤백한다 | `TestWritePositionPolicyDescriptorPreservesChmodAndWriteErrors` | no | pass |
| B16 | 격리 capability가 있을 때만 세 라우트를 등록한다 | `TestAServiceWithoutQuarantineCapabilityRegistersNoNewRoutes`, `TestTheQuarantineRoutesCarryAReleaseEndToEnd` | yes | pass |

B16의 RED는 a069 구현 전 코드에서 `TestTheQuarantineRoutesCarryAReleaseEndToEnd`가
실패하는 것으로 확인했고, 음성 방향(capability 없는 서비스에 새 라우트가 생기지
않는다)은 같은 커밋의 404 계약으로 고정한다. 새 라우트의 인증·method는
`TestTheQuarantineRoutesRequireTheEngineToken`과
`TestTheQuarantineRoutesRejectTheWrongMethod`가 덮는다.

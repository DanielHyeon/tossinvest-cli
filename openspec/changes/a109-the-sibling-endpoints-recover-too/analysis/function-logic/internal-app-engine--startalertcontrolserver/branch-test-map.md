# Branch Test Map: `StartAlertControlServer`

구현 후 재최신화(§1.7). 분기 13개는 편집 **후**의 AST다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | ops 없이 기동 거부 | 커버 없음 — 배선은 cmd 쪽 `TestTheAlertCommandsAreWiredUnderEngine`이 잰다 | no | no |
| B2 | 빈 엔진 디렉터리 문자열 | 커버 없음 | no | no |
| B3 | 안전하지 않은 엔진 디렉터리 | 커버 없음(자매 endpoint의 `TestPositionPolicyControlRejectsInsecureControlFilesystem`이 같은 검사를 잰다) | no | no |
| B4 | 기존 control 디렉터리 위에서 기동 | `TestAlertControlServerStartsOverALeftover`(a108 관용 핀) · `TestTheSiblingEndpointsStartOverAPreChmodSocket` | yes(§1.1) | yes(§1.4) |
| B5 | 디렉터리 생성 불가 | 커버 없음 | no | no |
| B6 | 회수: ① pre-chmod 0700 잔재 회수 ② 산 주인 거부 ③ staging 잔재 회수 ④ 낯선 엔트리 거부 | `TestTheSiblingEndpointsStartOverAPreChmodSocket` · `TestTheSiblingEndpointsRefuseToTakeALiveOwnersSocket` · `TestTheSiblingEndpointsReclaimTheirStagingLeftovers` · `TestTheSiblingEndpointsRefuseAForeignEntry` (뮤테이션 M1b·M3·M5·M9·M17이 이 절을 죽인다) | yes(§1.1·1.2·1.3) | yes(§1.4) |
| B7 | 회수 후 재-Mkdir 실패 | 커버 없음 — 회수가 디렉터리를 비웠으므로 도달하지 않는다. 회수가 비우지 않으면 이 절이 error가 되고, 그 경우를 뮤테이션 M16이 만든다 | no | no |
| B8 | 0700 아닌 control 디렉터리 거부 | 커버 없음(회수 쪽 동형은 `TestReclaimRefusesAnUnsafeControlDirectory`) | no | no |
| B9 | staged bind → 0600 → rename 발행 | `TestTheAlertSocketIsPrivateToThisUser`(0600 socket 실측) · 뮤테이션 M21(chmod 제거)이 23건을 죽인다 | no | yes(§1.4) |
| B10 | 발행 후 정확-0600 확인 | `TestTheAlertSocketIsPrivateToThisUser` · `TestTheClientSocketChecksStayExactlyZeroSixHundred`(완화 누출 감시) | no | no |
| B11 | 토큰 생성 실패 | 커버 없음 — `crypto/rand` 실패를 주입할 seam이 없다 | no | no |
| B12 | descriptor 인코딩 실패 | 커버 없음 — 고정 구조체라 도달 불가에 가깝다 | no | no |
| B13 | descriptor 발행 실패 정리 | 커버 없음 | no | no |

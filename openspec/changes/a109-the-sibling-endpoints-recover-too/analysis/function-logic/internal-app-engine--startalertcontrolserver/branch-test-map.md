# Branch Test Map: `StartAlertControlServer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | ops 없이 기동 거부 | 커버 없음 — 배선은 `TestTheAlertCommandsAreWiredUnderEngine`이 cmd 쪽에서 잰다 | no | no |
| B2 | 빈 엔진 디렉터리 문자열 | 커버 없음 | no | no |
| B3 | 안전하지 않은 엔진 디렉터리 | 커버 없음(자매 endpoint의 `TestPositionPolicyControlRejectsInsecureControlFilesystem`이 같은 검사를 잰다) | no | no |
| B4 | 기존 control 디렉터리 위에서 기동 | `TestAlertControlServerStartsOverALeftover` | no | no |
| B5 | 첫 기동(디렉터리 생성) | `TestTheAlertSocketIsPrivateToThisUser` | no | no |
| B6 | 디렉터리 생성 불가 | 커버 없음 | no | no |
| B7 | 실패 정리가 자기 디렉터리만 지움 | 커버 없음 | no | no |
| B8 | 0700 아닌 control 디렉터리 거부 | 커버 없음 | no | no |
| B9 | ① pre-chmod 0700 잔재에서 기동 — a109 §1.1 RED(현재 "preparing the alert control socket"으로 거부). ② 산 주인의 socket 탈취 — a109 §1.2 RED(현재 unlink 후 두 번째 서버가 올라선다) | 두 시나리오 모두 이 한 절이 만든다 | yes(§1.1·§1.2) | yes(§1.4) |
| B10 | 최종 경로 bind | `TestTheAlertSocketIsPrivateToThisUser`가 결과(0600 socket)를 잰다. staged bind 교체 후에도 같은 결과여야 한다 | no | yes(§1.4) |
| B11 | chmod 0600 | 같음 — 교체 후 chmod는 staged 경로에서 일어난다 | no | yes(§1.4) |
| B12 | 발행 후 정확-0600 확인 | `TestTheAlertSocketIsPrivateToThisUser` | no | no |
| B13 | 토큰 생성 실패 | 커버 없음 — `crypto/rand` 실패를 주입할 seam이 없다 | no | no |
| B14 | descriptor 인코딩 실패 | 커버 없음 — 고정 구조체라 도달 불가에 가깝다 | no | no |
| B15 | descriptor 발행 실패 정리 | 커버 없음 | no | no |

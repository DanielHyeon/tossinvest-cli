# a065 · Tasks

## 1. 편집 전 선언 (Pre-Edit Gate)

- [x] 1.1 편집할 기존 함수 셋(`runConsole`·`Console.routes`·`Console.settingsView`)의
      Function Logic Map과 Branch Test Map을 편집 전에 만든다.
- [x] 1.2 새 코드는 전부 **새 파일**에 둔다 — 기존 파일의 함수 본문을 늘리지 않는다.
- [x] 1.3 `review.md`에 선언 여섯을 적는다.

## 2. 채널 생성 (`internal/obs`)

- [x] 2.1 `NewTopic`이 `crypto/rand` 16바이트를 base32 소문자로 인코딩하고 접두어를 붙인다.
- [x] 2.2 실패 시 error를 반환한다 — 약한 난수로 대체하지 않는다.
- [x] 2.3 두 번 부르면 다른 값이 나오고, 접두어를 뺀 길이가 26자임을 테스트가 고정한다.
- [x] 2.4 `SubscribeURL`이 base URL과 topic으로 사람이 여는 주소를 만든다.

## 3. 설정 저장 (`internal/config`)

- [x] 3.1 `LoadRawNotifications`가 파일이 쓴 대로 블록을 반환한다. 없는 파일은 zero + no error.
- [x] 3.2 `SaveNotifications`가 `enabled`·`base_url`·`topic` **셋만** splice한다.
- [x] 3.3 `SaveNotificationsEnabled`가 `enabled` **한 키만** splice한다.
- [x] 3.4 닫힌 멤버 목록 둘이 서로소임을 테스트가 고정한다.
- [x] 3.5 저장이 다른 블록(`trading`·`automation_gate`·`adoption`)의 바이트를 건드리지
      않음을 테스트가 고정한다.

## 4. 콘솔 카드 (`internal/console`)

- [x] 4.1 `NotificationSettings` seam — `Load`·`Enable`·`Disable`·`Test` 넷.
- [x] 4.2 `settingsPage`에 알림 필드를 더한다 (구조체는 함수 밖이다).
- [x] 4.3 `NotificationGuard`가 미배선·읽기 실패를 block으로, 엔진 실행 중을 caution으로 낸다.
- [x] 4.4 세 handler: 켜기·테스트·끄기.
- [x] 4.5 카드 템플릿 — 현재 상태 · 적용 후 · 구독 주소 · 결과.
- [x] 4.6 notice 문구에 채널 식별자를 넣지 않는다.
- [x] 4.7 `settingsCardTab`에 `notifications: 도구`를 더한다.

## 5. 라우트와 정적 검사

- [x] 5.1 `Console.routes`에 세 라우트를 등록한다 (literal path).
- [x] 5.2 `consoleStateChanging`에 셋을 더한다.
- [x] 5.3 CSRF 게이트 목록에 셋을 더한다.
- [x] 5.4 라우트 수 하한을 실제 등록 수에 맞춘다.
- [x] 5.5 `screen_paths_test.go`가 요구하는 경로 상수를 더한다.

## 6. seam 구현과 audit (`cmd/tossctl`)

- [x] 6.1 `consoleNotificationSeam` — config service + audit log.
- [x] 6.2 `Enable`이 채널이 없을 때만 생성하고, 있으면 유지한다.
- [x] 6.3 audit 항목: `enabled`·`base_url`·`topic_configured`. 값은 남기지 않는다.
- [x] 6.4 `Test`가 `obs.Ntfy`로 한 통 보낸다 — outbox·gate·mode 무접촉.
- [x] 6.5 `Test`가 환경의 `TOSSCTL_NTFY_TOKEN`을 쓴다 (self-hosted 경로 유지).
- [x] 6.6 `runConsole`이 seam을 주입한다 (한 줄).

## 7. RED → GREEN

- [x] 7.1 채널 생성: 난수성·길이·중복.
- [x] 7.2 저장: 세 키만·한 키만·다른 블록 무변경.
- [x] 7.3 켜기: 채널 생성 + 저장 + audit + 반환.
- [x] 7.4 다시 켜기: 기존 채널 유지.
- [x] 7.5 끄기: `enabled`만 false, `topic` 바이트 유지.
- [x] 7.6 audit에 채널 식별자가 없다.
- [x] 7.7 notice URL에 채널 식별자가 없다.
- [x] 7.8 테스트 발송 실패가 설정을 되돌리지 않는다.
- [x] 7.9 화면에 토큰 입력란이 없다.
- [x] 7.10 미배선 빌드는 사유를 렌더하고 버튼을 렌더하지 않는다.

## 8. 변이 검증

- [x] 8.1 M1 — 채널 생성을 고정 문자열로 → RED
- [x] 8.2 M2 — 끄기가 topic도 지우게 → RED
- [x] 8.3 M3 — audit에 topic 값 기록 → RED
- [x] 8.4 M4 — notice에 채널 주소 포함 → RED
- [x] 8.5 M5 — 테스트 실패 시 설정 롤백 → RED
- [x] 8.6 M6 — 다시 켜기가 채널을 재생성 → RED
- [x] 8.7 변이 후 모든 파일이 바이트 동일하게 복원됨을 확인한다.

## 9. 게이트

- [x] 9.1 `go build ./...` · `go vet ./...` · `go test ./... -count=1`
- [x] 9.2 `openspec validate --all --strict`
- [x] 9.3 `check_analysis.py --change a065-alerts-turn-on-with-one-button`
- [x] 9.4 `make sdd-sync` → `make sdd-check`
- [x] 9.5 PM story/registry/feature/generated
- [x] 9.6 `make gate CHANGE=a065-alerts-turn-on-with-one-button`

## 10. 배포 후 실측

- [ ] 10.1 배포 후 콘솔에서 알림을 켜고 테스트 메시지가 실제로 도착하는지 실측한다.

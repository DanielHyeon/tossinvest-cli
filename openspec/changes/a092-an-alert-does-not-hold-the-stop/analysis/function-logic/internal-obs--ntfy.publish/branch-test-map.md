# Branch Test Map: `Ntfy.Publish`

Source: `internal/obs/ntfy.go` (85-139). AST 기준 분기 9 / 이탈 5 / defers 2.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:87` topic 없음 → `ErrNtfyNotConfigured` | `TestNtfyWithoutATopicIsNotConfigured` (`obs_test.go:281`) | no | yes |
| B2 | `:91` BaseURL 비면 기본 서버 | `TestNtfyPublicServiceIsFlagged` (`:290`) | no | yes |
| B3 | `:96` **`Timeout` 0이면 10s** | **없음** | no | no |
| B4 | `:104` 요청 생성 오류 | **없음** | no | no |
| B5 | `:110` 제목 헤더 | `TestNtfyPublishesToTheTopic` (`:208`) | no | yes |
| B6 | `:117` 베어러 토큰 | `TestNtfySendsTheBearerToken` (`:257`) | no | yes |
| B7 | `:122` **`HTTPClient` nil이면 `Timeout` 부여** | **없음** — 테스트는 `httptest` client를 주입하므로 이 분기를 지나지 않는다 | no | no |
| B8 | `:126` 전송 실패 | 간접 | no | yes |
| B9 | `:134` 2xx 아님 → 거부 오류 | `TestNtfyReportsARefusal` (`:269`) | no | yes |

## 예산을 만드는 두 분기가 무테스트다

**B3과 B7이 10초를 만드는데 둘 다 테스트가 없다.** 이유는 구조적이다 — 모든 ntfy
테스트가 `HTTPClient`에 `httptest` client를 주입하므로 B7이 거짓이 되고, `Timeout`을
명시하므로 B3도 거짓이 된다. **프로덕션 조립만 두 분기를 지난다**
(`notifications.go:101`, `cmd/tossctl/notificationsettings.go:151`).

이것이 "10초"가 코드에 있으면서 어떤 테스트도 그것을 고정하지 않는 이유다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R22 | 프로덕션 조립이 만드는 `*obs.Ntfy`의 유효 기한이 10초임을 고정 | B3·B7의 기본값이 소리 없이 바뀌면 실패한다 |

R22는 **이 change의 예산 계산을 미래에 대해 고정하는 가드**다. 이 change가
"34초"를 근거로 설계를 정하므로, 그 숫자가 조용히 바뀌면 설계 근거가 사라진다.
이 change는 `ntfy.go` 본문을 바꾸지 않지만 이 가드는 추가한다.

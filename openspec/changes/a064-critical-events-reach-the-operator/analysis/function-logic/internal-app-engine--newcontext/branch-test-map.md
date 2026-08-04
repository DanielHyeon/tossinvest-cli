# Branch Test Map: `NewContext`

편집 후 분기는 13개에서 14개가 되었다. 새 분기는 B4 — 주입된 publisher가 설정보다
우선한다 — 이며, 나머지 13개의 조건·순서·`jrn.Close()` 위치는 그대로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 설정·경로 조립 실패 → 반환 | 기존 engine 테스트 | no | pass |
| B2 | audit 로그 열기 실패 → 반환 | 기존 engine 테스트 | no | pass |
| B3 | `opts.Clock`이 nil → 시스템 시계 | 기존 engine 테스트 | no | pass |
| B4 | `opts.Publisher` 주입이 설정보다 우선한다 | `TestAnInjectedPublisherWins` | no | pass |
| B5 | 알림 설정 audit 실패 → 기동 거부 | 기존 engine 테스트 | no | pass |
| B6 | 계좌 미해석 → 거부 | 기존 engine 테스트 | no | pass |
| B7 | 원장 open 실패 → 거부 | 기존 engine 테스트 | no | pass |
| B8 | apply hook 바인딩 실패 → Close + 거부 | 기존 engine 테스트 | no | pass |
| B9 | gateway 조립 실패 → Close + 거부 | 기존 engine 테스트 | no | pass |
| B10 | 게이트 ON + Guardian 미주입 → 생성 | 기존 engine 테스트 | no | pass |
| B11 | factory 미지정 → 기본 factory | 기존 engine 테스트 | no | pass |
| B12 | Guardian 생성 실패 → Close + 거부 | 기존 engine 테스트 | no | pass |
| B13 | 인터록 실패 → Close + 거부 | 기존 engine 테스트 | no | pass |
| B14 | 미검증이면 Guardian을 버린다 | 기존 engine 테스트 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| §0.2 | 알림 off → publisher nil, 오늘 동작과 동일 | `TestStartupWiresNoPublisherWhenNotificationsAreOff` | **yes** (M7) | pass |
| 조립 | 알림 on → 설정대로 ntfy transport | `TestStartupWiresTheConfiguredPublisher` | no | pass |
| D7 | 거부된 알림 설정 → 기동은 하고 publisher는 없다 | `TestARefusedNotificationBlockDoesNotBlockStartup` | no | pass |
| §0.5 | 알림 설정이 audit에 남는다 | `TestNotificationSettingsAreAudited` | no | pass |
| §0.8 | audit에 topic 값도 token 값도 없다 | `TestTheAuditTrailCarriesNoNotificationSecret` | **yes** (M6) | pass |

**M7의 RED는 실제 관측이다** (2026-08-04): `!cfg.Enabled` 조기 반환을 지우자
`TestADisabledBlockWithAChannelStaysOff`가 조립된 `*obs.Ntfy`를 인용하며 실패했다.
